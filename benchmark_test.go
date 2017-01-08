package ctxio_test

import (
	"context"
	"testing"

	"github.com/jackc/ctxio"
)

func BenchmarkGoroutineReaderFastSource(b *testing.B) {
	dest := make([]byte, 1000)
	ir := newGoroutineReaderFn(ByteFillReader('x'))
	r := ctxio.WithContext(context.Background(), ir)

	for i := 0; i < b.N; i++ {
		_, err := r.Read(dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNormalReaderFastSource(b *testing.B) {
	dest := make([]byte, 1000)
	r := ByteFillReader('x')

	for i := 0; i < b.N; i++ {
		_, err := r.Read(dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}
