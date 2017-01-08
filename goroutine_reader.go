package ctxio

import (
	"context"
	"io"
)

const maxGoroutineReaderBufSize = 4096

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
		if len(gr.buf) > 0 {
			<-gr.busyChan // release busyChan
			return gr.moveReadResult(p)
		}

		// Start a new read
		go func() {
			if len(p) <= cap(gr.buf) {
				gr.buf = gr.buf[:len(p)]
			} else {
				gr.buf = make([]byte, len(p))
			}
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
		<-gr.busyChan // release busyChan
		return n, err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (gr *GoroutineReader) moveReadResult(p []byte) (n int, err error) {
	n = copy(p, gr.buf)
	if len(gr.buf) > n {
		gr.buf = gr.buf[:n]
	} else if cap(gr.buf) <= maxGoroutineReaderBufSize {
		gr.buf = gr.buf[0:0]
	} else {
		gr.buf = nil
	}
	return n, gr.err
}
