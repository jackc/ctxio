package ctxio

import (
	"context"
	"io"
)

type GoroutineReader struct {
	r io.Reader

	busyChan chan struct{} // using this as a mutex because need to select on it
	buf      []byte
	err      error
}

func NewGoroutineReader(r io.Reader) *GoroutineReader {
	gr := &GoroutineReader{r: r, busyChan: make(chan struct{}, 1)}
	return gr
}

func (gr *GoroutineReader) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	// Check if ctx is already canceled
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	select {
	// If send is successful to busyChan then no read was already in progress
	case gr.busyChan <- struct{}{}:
		// If a previous read was completed after the context was canceled
		if gr.buf != nil {
			return gr.moveReadResult(p)
		}

		// Start a new read
		go func() {
			gr.buf = make([]byte, len(p))
			var n int
			n, gr.err = gr.r.Read(gr.buf)
			gr.buf = gr.buf[:n]

			<-gr.busyChan
		}()
	default:
	}

	select {
	// read has completed
	case gr.busyChan <- struct{}{}:
		n, err = gr.moveReadResult(p)
		<-gr.busyChan // clear busyChan
		return n, err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (gr *GoroutineReader) moveReadResult(p []byte) (n int, err error) {
	n = copy(gr.buf, p)
	if len(gr.buf) > n {
		gr.buf = gr.buf[:n]
	} else {
		gr.buf = nil
	}
	return n, gr.err
}
