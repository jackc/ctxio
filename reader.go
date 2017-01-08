package ctxio

import (
	"context"
	"io"
)

type Reader struct {
	r io.Reader

	busyChan chan struct{} // using this as a mutex because need to select on it
	buf      []byte
	err      error
}

func NewReader(r io.Reader) *Reader {
	cr := &Reader{r: r, busyChan: make(chan struct{}, 1)}
	return cr
}

func (r *Reader) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	// Check if ctx is already canceled
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	select {
	// If send is successful to busyChan then no read was already in progress
	case r.busyChan <- struct{}{}:
		// If a previous read was completed after the context was canceled
		if r.buf != nil {
			return r.moveReadResult(p)
		}

		// Start a new read
		go func() {
			r.buf = make([]byte, len(p))
			var n int
			n, r.err = r.r.Read(r.buf)
			r.buf = r.buf[:n]

			<-r.busyChan
		}()
	default:
	}

	select {
	// read has completed
	case r.busyChan <- struct{}{}:
		n, err = r.moveReadResult(p)
		<-r.busyChan // clear busyChan
		return n, err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (r *Reader) moveReadResult(p []byte) (n int, err error) {
	n = copy(r.buf, p)
	if len(r.buf) > n {
		r.buf = r.buf[:n]
	} else {
		r.buf = nil
	}
	return n, r.err
}
