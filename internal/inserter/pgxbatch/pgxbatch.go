package pgxbatch

import (
	"context"
	"iter"

	"pg-bulk-flow/internal/inserter"
	"pg-bulk-flow/internal/model"

	"github.com/jackc/pgx/v5"
)

type Inserter struct {
	conn      *pgx.Conn
	batchSize int
}

func New(conn *pgx.Conn, batchSize int) *Inserter {
	return &Inserter{
		conn:      conn,
		batchSize: batchSize,
	}
}

func (i *Inserter) prepareInsert(ctx context.Context) error {
	_, err := i.conn.Prepare(ctx, "insert_name",
		`INSERT INTO names (count, name_type, name_text, gender) VALUES ($1, $2, $3, $4)`)
	return err
}

func (i *Inserter) deallocate(ctx context.Context) error {
	return i.conn.Deallocate(ctx, "insert_name")
}

func (i *Inserter) sendBatch(ctx context.Context, b *pgx.Batch) error {
	return i.conn.SendBatch(ctx, b).Close()
}

func (i *Inserter) Insert(ctx context.Context, names iter.Seq[model.Name]) (int64, error) {
	if err := i.prepareInsert(ctx); err != nil {
		return 0, err
	}
	defer i.deallocate(ctx)

	var count int64
	b := &pgx.Batch{}
	reset := func(b *pgx.Batch) {
		clear(b.QueuedQueries)
		b.QueuedQueries = b.QueuedQueries[:0]
	}

	for v := range names {
		b.Queue("insert_name", v.Count, v.Type, v.Text, v.Gender)
		if b.Len() >= i.batchSize {
			if err := i.sendBatch(ctx, b); err != nil {
				return count, err
			}
			count += int64(b.Len())
			reset(b)
		}
	}

	if b.Len() > 0 {
		if err := i.sendBatch(ctx, b); err != nil {
			return count, err
		}
		count += int64(b.Len())
	}

	return count, nil
}

func (i *Inserter) InsertWithPipeline(ctx context.Context, names iter.Seq[model.Name]) (int64, error) {
	if err := i.prepareInsert(ctx); err != nil {
		return 0, err
	}
	defer i.deallocate(ctx)

	ch := make(chan *pgx.Batch)
	done := make(chan struct{})

	b1, b2 := &pgx.Batch{}, &pgx.Batch{}
	reset := func(b *pgx.Batch) {
		clear(b.QueuedQueries)
		b.QueuedQueries = b.QueuedQueries[:0]
	}

	var (
		count int64
		err   error
	)

	go func() {
		defer close(done)
		for b := range ch {
			if err := i.sendBatch(ctx, b); err != nil {
				return
			}
			count += int64(b.Len())
		}
	}()

	for v := range names {
		b1.Queue("insert_name", v.Count, v.Type, v.Text, v.Gender)
		if b1.Len() >= i.batchSize {
			select {
			case ch <- b1:
				b1, b2 = b2, b1
				reset(b1)
			case <-done:
				return count, err
			}
		}
	}

	if b1.Len() > 0 {
		select {
		case ch <- b1:
		case <-done:
			return count, err
		}
	}

	close(ch)
	<-done
	return count, err
}

var _ inserter.Inserter = &Inserter{}
