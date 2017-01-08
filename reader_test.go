package ctxio_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/jackc/ctxio"
)

func testInterruptableReader(t *testing.T, irFn func(io.Reader) ctxio.InterruptableReader) {
	mr := &MockReader{}
	mr.Mock([]byte("Hello, world\n"), nil, 0)
	mr.Mock([]byte("This was slow\n"), nil, 2*time.Second)
	mr.Mock([]byte("Goodbye\n"), nil, 0)

	ir := irFn(mr)
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	r := ctxio.WithContext(ctx, ir)

	buf := make([]byte, 1024)

	n, err := r.Read(buf)
	if n != 13 {
		t.Fatalf("expected 13 read, but was %d", n)
	}
	if err != nil {
		t.Fatal(err)
	}

	n, err = r.Read(buf)
	if n != 0 {
		t.Fatalf("expected 0 read, but was %d", n)
	}
	if err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}

func TestReader(t *testing.T) {
	testInterruptableReader(t, func(r io.Reader) ctxio.InterruptableReader { return ctxio.NewGoroutineReader(r) })
}
