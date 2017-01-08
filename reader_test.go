package ctxio_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/jackc/ctxio"
)

type mockReaderResponse struct {
	buf   []byte
	err   error
	delay time.Duration
}

type MockReader struct {
	responses []mockReaderResponse
}

func (mr *MockReader) Mock(buf []byte, err error, delay time.Duration) {
	mr.responses = append(mr.responses, mockReaderResponse{buf: buf, err: err, delay: delay})
}

func (mr *MockReader) Read(p []byte) (n int, err error) {
	if len(mr.responses) > 0 {
		resp := mr.responses[0]
		mr.responses = mr.responses[1:]

		if len(resp.buf) > len(p) {
			panic("mocked larger response than Read request can handle")
		}
		copy(resp.buf, p)

		if resp.delay > 0 {
			time.Sleep(resp.delay)
		}

		return len(resp.buf), resp.err
	}

	return 0, io.EOF
}

func TestReader(t *testing.T) {
	mr := &MockReader{}
	mr.Mock([]byte("Hello, world\n"), nil, 0)
	mr.Mock([]byte("This was slow\n"), nil, 2*time.Second)
	mr.Mock([]byte("Goodbye\n"), nil, 0)

	r := ctxio.NewReader(mr)
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	buf := make([]byte, 1024)

	n, err := r.ReadContext(ctx, buf)
	if n != 13 {
		t.Fatalf("expected 13 read, but was %d", n)
	}
	if err != nil {
		t.Fatal(err)
	}

	n, err = r.ReadContext(ctx, buf)
	if n != 0 {
		t.Fatalf("expected 0 read, but was %d", n)
	}
	if err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}
