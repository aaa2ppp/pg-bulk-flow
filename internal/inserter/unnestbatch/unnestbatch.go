package unnestbatch

import (
	"context"
	"iter"

	"pg-bulk-flow/internal/inserter"
	"pg-bulk-flow/internal/model"

	"github.com/jackc/pgx/v5"
)

type insertBatch struct {
	Count  []int32
	Type   []model.NameType
	Text   []string
	Gender []model.Gender
}

func newInsertBatch(batchSize int) *insertBatch {
	return &insertBatch{
		Count:  make([]int32, 0, batchSize),
		Type:   make([]model.NameType, 0, batchSize),
		Text:   make([]string, 0, batchSize),
		Gender: make([]model.Gender, 0, batchSize),
	}
}

func (b *insertBatch) Len() int {
	return len(b.Count)
}

func (b *insertBatch) Add(v model.Name) {
	b.Count = append(b.Count, v.Count)
	b.Type = append(b.Type, v.Type)
	b.Text = append(b.Text, v.Text)
	b.Gender = append(b.Gender, v.Gender)
}

func (b *insertBatch) Reset() {
	b.Count = b.Count[:0]
	b.Type = b.Type[:0]
	b.Text = b.Text[:0]
	b.Gender = b.Gender[:0]
}

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
	_, err := i.conn.Prepare(ctx, "insert_names",
		`INSERT INTO names (count, name_type, name_text, gender)
        SELECT UNNEST($1::int[]), UNNEST($2::name_type_enum[]), 
               UNNEST($3::text[]), UNNEST($4::gender_enum[])`)
	return err
}

func (i *Inserter) sendBatch(ctx context.Context, b *insertBatch) error {
	_, err := i.conn.Exec(ctx, "insert_names", b.Count, b.Type, b.Text, b.Gender)
	return err
}

func (i *Inserter) Insert(ctx context.Context, names iter.Seq[model.Name]) (int64, error) {
	if err := i.prepareInsert(ctx); err != nil {
		return 0, err
	}
	defer i.conn.Deallocate(ctx, "insert_names")

	var count int64
	b := newInsertBatch(i.batchSize)

	for v := range names {
		b.Add(v)
		if b.Len() >= i.batchSize {
			if err := i.sendBatch(ctx, b); err != nil {
				return count, err
			}
			count += int64(b.Len())
			b.Reset()
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
	defer i.conn.Deallocate(ctx, "insert_names")

	ch := make(chan *insertBatch)
	done := make(chan struct{})
	b1 := newInsertBatch(i.batchSize)
	b2 := newInsertBatch(i.batchSize)

	var (
		count int64
		err   error
	)

	go func() {
		defer close(done)
		for b := range ch {
			if err = i.sendBatch(ctx, b); err != nil {
				return
			}
			count += int64(b.Len())
		}
	}()

	for v := range names {
		b1.Add(v)
		if b1.Len() >= i.batchSize {
			select {
			case ch <- b1:
				b1, b2 = b2, b1
				b1.Reset()
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
