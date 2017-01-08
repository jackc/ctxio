package ctxio_test

import (
	"io"
	"time"
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
