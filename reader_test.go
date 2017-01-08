package ctxio_test

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/jackc/ctxio"
)

func testInterruptableReaderDeadline(irFn func(io.Reader) ctxio.InterruptableReader) func(t *testing.T) {
	return func(t *testing.T) {
		mr := &MockReader{}
		mr.Mock([]byte("Hello, world\n"), nil, 0)
		mr.Mock([]byte("This was slow\n"), nil, time.Second)
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
}

func testInterruptableReaderCancel(irFn func(io.Reader) ctxio.InterruptableReader) func(t *testing.T) {
	return func(t *testing.T) {
		mr := &MockReader{}
		mr.Mock([]byte("Hello, world\n"), nil, 0)

		ir := irFn(mr)
		ctx, cancelFn := context.WithCancel(context.Background())
		r := ctxio.WithContext(ctx, ir)

		buf := make([]byte, 1024)

		cancelFn()
		n, err := r.Read(buf)
		if n != 0 {
			t.Fatalf("expected 0 read, but was %d", n)
		}
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}

func testInterruptableReaderResume(irFn func(io.Reader) ctxio.InterruptableReader) func(t *testing.T) {
	return func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())

		mr := &MockReader{}
		source := &bytes.Buffer{}

		for i := 0; i < 1000; i++ {
			buf := make([]byte, 1+rand.Intn(100))
			rand.Read(buf)
			mr.Mock(buf, nil, time.Duration(rand.Intn(5))*time.Millisecond)
			source.Write(buf)
		}

		ir := irFn(mr)

		dest := &bytes.Buffer{}

		for {
			ctx, _ := context.WithTimeout(context.Background(), 25*time.Millisecond)
			r := ctxio.WithContext(ctx, ir)
			_, err := dest.ReadFrom(r)
			if err == nil {
				break
			} else if err != context.Canceled && err != context.DeadlineExceeded {
				t.Fatal(err)
			}
		}

		if bytes.Compare(source.Bytes(), dest.Bytes()) != 0 {
			t.Fatalf("source and dest bytes did not match: %v, %v", source.Bytes(), dest.Bytes())
		}
	}
}

func testInterruptableReaderEOF(irFn func(io.Reader) ctxio.InterruptableReader) func(t *testing.T) {
	return func(t *testing.T) {
		mr := &MockReader{}
		mr.Mock([]byte("Hello, world\n"), nil, 0)

		ir := irFn(mr)
		r := ctxio.WithContext(context.Background(), ir)

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
		if err != io.EOF {
			t.Fatal(err)
		}
	}
}

func newGoroutineReaderFn(r io.Reader) ctxio.InterruptableReader {
	return ctxio.NewGoroutineReader(r)
}

func TestInterruptableReaderDeadline(t *testing.T) {
	t.Run("GoroutineReader", testInterruptableReaderDeadline(newGoroutineReaderFn))
}

func TestInterruptableReaderCancel(t *testing.T) {
	t.Run("GoroutineReader", testInterruptableReaderCancel(newGoroutineReaderFn))
}

func TestInterruptableReaderResume(t *testing.T) {
	t.Run("GoroutineReader", testInterruptableReaderResume(newGoroutineReaderFn))
}

func TestInterruptableReaderEOF(t *testing.T) {
	t.Run("GoroutineReader", testInterruptableReaderEOF(newGoroutineReaderFn))
}
