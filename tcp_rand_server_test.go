package ctxio_test

import (
	"bytes"
	"math/rand"
	"net"
	"time"
)

type randIntRange struct {
	min int
	max int
}

func (rir randIntRange) Pick() int {
	return rir.min + rand.Intn(rir.max-rir.min)
}

type TCPRandServer struct {
	ln          net.Listener
	sentBuf     *bytes.Buffer
	packetCount randIntRange
	packetLen   randIntRange
	packetDelay randIntRange
}

func NewTCPRandServer(packetCount, packetLen, packetDelay randIntRange) *TCPRandServer {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	return &TCPRandServer{
		ln:          ln,
		sentBuf:     &bytes.Buffer{},
		packetCount: packetCount,
		packetLen:   packetLen,
		packetDelay: packetDelay,
	}
}

func (s *TCPRandServer) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *TCPRandServer) SentBytes() []byte {
	return s.sentBuf.Bytes()
}

func (s *TCPRandServer) ReleaseSentBytes() {
	s.sentBuf.Reset()
}

func (s *TCPRandServer) Close() error {
	return s.ln.Close()
}

func (s *TCPRandServer) AcceptOnce() {
	conn, err := s.ln.Accept()
	if err != nil {
		panic(err)
	}
	s.handleConn(conn)
}

func (s *TCPRandServer) handleConn(conn net.Conn) {
	packetCount := s.packetCount.Pick()

	buf := make([]byte, s.packetLen.max)

	for i := 0; i < packetCount; i++ {
		delay := s.packetDelay.Pick()
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		packetLen := s.packetLen.Pick()

		rand.Read(buf[:packetLen])
		s.sentBuf.Write(buf[:packetLen])

		n, err := conn.Write(buf[:packetLen])
		if n != packetLen {
			break // did not write entire buffer
		}
		if err != nil {
			break
		}

	}

	conn.Close()
}