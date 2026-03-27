package interfaces

import (
	"context"
	"io"
)

type Output interface {
	ToString(ctx context.Context) (string, error)
	ToStringDetails(ctx context.Context) (string, error)
	ToStruct(ctx context.Context) (any, error)
	WriteToJson(ctx context.Context, w io.Writer) error
}
