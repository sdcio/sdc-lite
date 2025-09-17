package utils

import (
	"context"
	"io"
)

type ctxReader struct {
	ctx context.Context
	r   io.ReadCloser
}

func (cr *ctxReader) Read(p []byte) (int, error) {
	type readResult struct {
		n   int
		err error
	}
	ch := make(chan readResult, 1)

	go func() {
		n, err := cr.r.Read(p)
		ch <- readResult{n, err}
	}()

	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	case res := <-ch:
		return res.n, res.err
	}
}

func (cr *ctxReader) Close() error {
	return cr.r.Close()
}

func NewCtxReader(ctx context.Context, r io.ReadCloser) *ctxReader {
	return &ctxReader{
		ctx: ctx,
		r:   r,
	}
}
