package ctxio

import (
	"context"
)

type InterruptableReader interface {
	ReadContext(ctx context.Context, p []byte) (n int, err error)
}

func WithContext(ctx context.Context, r InterruptableReader) *ReaderWithContext {
	return &ReaderWithContext{r: r, ctx: ctx}
}

type ReaderWithContext struct {
	r   InterruptableReader
	ctx context.Context
}

func (rwc *ReaderWithContext) Read(p []byte) (n int, err error) {
	return rwc.r.ReadContext(rwc.ctx, p)
}
