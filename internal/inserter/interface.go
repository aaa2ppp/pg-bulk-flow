package inserter

import (
	"context"
	"iter"

	"pg-bulk-flow/internal/model"
)

type Inserter interface {
	Insert(ctx context.Context, names iter.Seq[model.Name]) (int64, error)
	InsertWithPipeline(ctx context.Context, names iter.Seq[model.Name]) (int64, error)
}
