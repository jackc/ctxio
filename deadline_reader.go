package ctxio

import (
	"context"
	"net"
	"time"
)

type DeadlineReader struct {
	r net.Conn
}

func NewDeadlineReader(r net.Conn) *DeadlineReader {
	dr := &DeadlineReader{r: r}
	return dr
}

func (dr *DeadlineReader) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	// Avoid cancelation setup if cancelation is impossible
	if ctx.Done() == nil {
		return dr.r.Read(p)
	}
	return dr.readContext(ctx, p)
}

func (dr *DeadlineReader) readContext(ctx context.Context, p []byte) (n int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	readDoneChan := make(chan struct{})
	readDeadlineSet := make(chan bool)

	go func() {
		select {
		case <-ctx.Done():
			dr.r.SetReadDeadline(time.Now())
			readDeadlineSet <- true
		case <-readDoneChan:
			readDeadlineSet <- false
		}
	}()

	n, err = dr.r.Read(p)

	// Cleanup cancelation goroutine
	close(readDoneChan)

	// Clear read deadline if set
	if <-readDeadlineSet {
		dr.r.SetReadDeadline(time.Time{})
	}

	// If the Read completed without error then return success even if ctx has
	// since been canceled.
	if err == nil {
		return n, nil
	}

	// Prefer ctx err over Read err
	select {
	case <-ctx.Done():
		return n, ctx.Err()
	default:
		return n, err
	}
}
