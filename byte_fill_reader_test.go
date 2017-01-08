package ctxio_test

type ByteFillReader byte

func (bfr ByteFillReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(bfr)
	}
	return len(p), nil
}
