package ctxio_test

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/jackc/ctxio"
)

const bufSize = 2048

func BenchmarkGoroutineReaderFastSource(b *testing.B) {
	dest := make([]byte, bufSize)
	ir := newGoroutineReaderFn(ByteFillReader('x'))
	r := ctxio.WithContext(context.Background(), ir)

	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(r, dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNormalReaderFastSource(b *testing.B) {
	dest := make([]byte, bufSize)
	r := ByteFillReader('x')

	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(r, dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func sumBytes(buf []byte) int {
	var sum int
	for _, b := range buf {
		sum = sum + int(b)
	}
	return sum
}

func BenchmarkGoroutineReaderTCPSource(b *testing.B) {
	dest := make([]byte, bufSize)
	tcpRandServer := NewTCPRandServer(
		randIntRange{min: b.N * 10, max: b.N * 10},
		randIntRange{min: 1, max: bufSize * 2},
		randIntRange{min: 0, max: 0},
		false,
	)
	go tcpRandServer.AcceptOnce()
	defer tcpRandServer.Close()

	conn, err := net.Dial("tcp", tcpRandServer.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	ir := newGoroutineReaderFn(conn)
	r := ctxio.WithContext(context.Background(), ir)

	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(r, dest)
		if err != nil {
			b.Fatal(err)
		}

		if sum := sumBytes(dest); sum < 1000 {
			b.Fatalf("unusually low sum: %d", sum)
		}
	}
}

func BenchmarkDeadlineReaderCancelableTCPSource(b *testing.B) {
	dest := make([]byte, bufSize)
	tcpRandServer := NewTCPRandServer(
		randIntRange{min: b.N * 10, max: b.N * 10},
		randIntRange{min: 1, max: bufSize * 2},
		randIntRange{min: 0, max: 0},
		false,
	)
	go tcpRandServer.AcceptOnce()
	defer tcpRandServer.Close()

	conn, err := net.Dial("tcp", tcpRandServer.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	ir := ctxio.NewDeadlineReader(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	r := ctxio.WithContext(ctx, ir)

	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(r, dest)
		if err != nil {
			b.Fatal(err)
		}

		if sum := sumBytes(dest); sum < 1000 {
			b.Fatalf("unusually low sum: %d", sum)
		}
	}
}

func BenchmarkDeadlineReaderUncancelableTCPSource(b *testing.B) {
	dest := make([]byte, bufSize)
	tcpRandServer := NewTCPRandServer(
		randIntRange{min: b.N * 10, max: b.N * 10},
		randIntRange{min: 1, max: bufSize * 2},
		randIntRange{min: 0, max: 0},
		false,
	)
	go tcpRandServer.AcceptOnce()
	defer tcpRandServer.Close()

	conn, err := net.Dial("tcp", tcpRandServer.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	ir := ctxio.NewDeadlineReader(conn)
	r := ctxio.WithContext(context.Background(), ir)

	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(r, dest)
		if err != nil {
			b.Fatal(err)
		}

		if sum := sumBytes(dest); sum < 1000 {
			b.Fatalf("unusually low sum: %d", sum)
		}
	}
}

func BenchmarkNormalReaderTCPSource(b *testing.B) {
	dest := make([]byte, bufSize)
	tcpRandServer := NewTCPRandServer(
		randIntRange{min: b.N * 10, max: b.N * 10},
		randIntRange{min: 1, max: bufSize * 2},
		randIntRange{min: 0, max: 0},
		false,
	)
	go tcpRandServer.AcceptOnce()
	defer tcpRandServer.Close()

	conn, err := net.Dial("tcp", tcpRandServer.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(conn, dest)
		if err != nil {
			b.Fatal(err)
		}

		if sum := sumBytes(dest); sum < 1000 {
			b.Fatalf("unusually low sum: %d", sum)
		}
	}
}
