package copyfrom

import (
	"context"
	"iter"

	"pg-bulk-flow/internal/inserter"
	"pg-bulk-flow/internal/model"

	"github.com/jackc/pgx/v5"
)

type source struct {
	next   func() (model.Name, bool)
	stop   func()
	values []any
}

func newSource(names iter.Seq[model.Name]) *source {
	next, stop := iter.Pull(names)
	return &source{
		next:   next,
		stop:   stop,
		values: make([]any, 0, 4),
	}
}

func (s *source) Next() bool {
	v, ok := s.next()
	if !ok {
		s.values = s.values[:0]
		return false
	}
	s.values = append(s.values[:0], v.Count, v.Type, v.Text, v.Gender)
	return true
}

func (s *source) Values() ([]any, error) {
	return s.values, nil
}

func (s *source) Err() error {
	return nil
}

func (s *source) close() {
	s.stop()
}

var _ pgx.CopyFromSource = &source{}

type Inserter struct {
	conn *pgx.Conn
}

func New(conn *pgx.Conn) *Inserter {
	return &Inserter{conn}
}

func (ins *Inserter) Insert(ctx context.Context, names iter.Seq[model.Name]) (int64, error) {
	src := newSource(names)
	defer src.close()

	return ins.conn.CopyFrom(ctx, pgx.Identifier{"names"},
		[]string{"count", "name_type", "name_text", "gender"}, src)
}

type asyncSource struct {
	ch     chan model.Name
	cancel chan struct{}
	values []any
}

func newAsyncSource(names iter.Seq[model.Name]) *asyncSource {
	ch := make(chan model.Name)
	cancel := make(chan struct{})

	go func() {
		defer close(ch)
		for v := range names {
			select {
			case ch <- v:
			case <-cancel:
				return
			}
		}
	}()

	return &asyncSource{
		ch:     ch,
		cancel: cancel,
		values: make([]any, 0, 4),
	}
}

// Err implements pgx.CopyFromSource.
func (a *asyncSource) Err() error {
	return nil
}

// Next implements pgx.CopyFromSource.
func (a *asyncSource) Next() bool {
	v, ok := <-a.ch
	if !ok {
		return false
	}
	a.values = append(a.values[:0], v.Count, v.Type, v.Text, v.Gender)
	return true
}

// Values implements pgx.CopyFromSource.
func (a *asyncSource) Values() ([]any, error) {
	return a.values, nil
}

func (a *asyncSource) close() {
	close(a.cancel)
}

var _ pgx.CopyFromSource = &asyncSource{}

func (ins *Inserter) InsertWithPipeline(ctx context.Context, names iter.Seq[model.Name]) (int64, error) {
	src := newAsyncSource(names)
	defer src.close()
	
	return ins.conn.CopyFrom(ctx, pgx.Identifier{"names"},
		[]string{"count", "name_type", "name_text", "gender"}, src)
}

var _ inserter.Inserter = &Inserter{}
