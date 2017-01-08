package ctxio

import (
	"context"
	"net"
	"time"
)

type DeadlineReader struct {
	r               net.Conn
	readDoneChan    chan struct{}
	readDeadlineSet chan bool
}

func NewDeadlineReader(r net.Conn) *DeadlineReader {
	return &DeadlineReader{
		r:               r,
		readDoneChan:    make(chan struct{}),
		readDeadlineSet: make(chan bool),
	}
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

	go func() {
		select {
		case <-ctx.Done():
			dr.r.SetReadDeadline(time.Now())
			<-dr.readDoneChan
			dr.readDeadlineSet <- true
		case <-dr.readDoneChan:
			dr.readDeadlineSet <- false
		}
	}()

	n, err = dr.r.Read(p)

	// Cleanup cancelation goroutine
	dr.readDoneChan <- struct{}{}

	// Clear read deadline if set
	if <-dr.readDeadlineSet {
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
