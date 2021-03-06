package ctxio_test

import (
	"bytes"
	"math/rand"
	"net"
	"sync"
	"time"
)

type randIntRange struct {
	min int
	max int
}

func (rir randIntRange) Pick() int {
	if rir.min == rir.max {
		return rir.min
	}
	return rir.min + rand.Intn(rir.max-rir.min)
}

type TCPRandServer struct {
	ln          net.Listener
	packetCount randIntRange
	packetLen   randIntRange
	packetDelay randIntRange

	saveBytes   bool
	sentBufLock sync.Mutex
	sentBuf     *bytes.Buffer
}

func NewTCPRandServer(packetCount, packetLen, packetDelay randIntRange, saveBytes bool) *TCPRandServer {
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
		saveBytes:   saveBytes,
	}
}

func (s *TCPRandServer) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *TCPRandServer) SentBytes() []byte {
	s.sentBufLock.Lock()
	buf := make([]byte, len(s.sentBuf.Bytes()))
	copy(buf, s.sentBuf.Bytes())
	s.sentBufLock.Unlock()
	return buf
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

		if s.saveBytes {
			s.sentBufLock.Lock()
			s.sentBuf.Write(buf[:packetLen])
			s.sentBufLock.Unlock()
		}

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
