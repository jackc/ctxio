package ctxio_test

import (
	"context"
	"net"
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

func BenchmarkGoroutineReaderTCPSource(b *testing.B) {
	dest := make([]byte, 250)
	tcpRandServer := NewTCPRandServer(
		randIntRange{min: b.N * 2, max: b.N*2 + 1},
		randIntRange{min: 1, max: 500},
		randIntRange{min: 0, max: 10},
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
		_, err := r.Read(dest)
		if err != nil {
			b.Fatal(err)
		}

		tcpRandServer.ReleaseSentBytes()
	}
}

func BenchmarkNormalReaderTCPSource(b *testing.B) {
	dest := make([]byte, 250)
	tcpRandServer := NewTCPRandServer(
		randIntRange{min: b.N * 2, max: b.N*2 + 1},
		randIntRange{min: 1, max: 500},
		randIntRange{min: 0, max: 10},
	)
	go tcpRandServer.AcceptOnce()
	defer tcpRandServer.Close()

	conn, err := net.Dial("tcp", tcpRandServer.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < b.N; i++ {
		_, err := conn.Read(dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}
