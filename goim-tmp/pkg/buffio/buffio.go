package buffio

import (
	"errors"
	"io"
)

const (
	// 默认 4 × 1024
	defaultBufSize = 4096

	// 最小缓冲区读大小
	minReadBufferSize = 16

	// 最大连续空读
	maxConsecutiveEmptyReads = 100
)

var (
	ErrInvalidUnreadByte = errors.New("buffio: invalid use UnreadByte")
	ErrInvalidUnreadRune = errors.New("buffio: invalid use of UnreadRune")

	ErrBufferFull    = errors.New("buffio: buffer full")
	ErrNegativeCount = errors.New("buffio: negative count")

	errNegativeRead = errors.New("bufio: reader returned negative count from Read")
)

type Reader struct {
	buf  []byte
	rd   io.Reader
	r, w int
	err  error
}

func NewReaderSize(rd io.Reader, size int) *Reader {
	b, ok := rd.(*Reader)
	if ok && len(b.buf) >= size {
		return b
	}

	if size < minReadBufferSize {
		size = minReadBufferSize
	}

	r := new(Reader)
	r.reset(make([]byte, size), rd)
	return r
}

func (b *Reader) Reset(r io.Reader) {
	b.reset(b.buf, r)
}

func (b *Reader) reset(buf []byte, r io.Reader) {
	*b = Reader{
		buf: buf,
		rd:  r,
	}
}

func (b *Reader) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, b.readErr()
	}
	if b.r == b.w {
		if b.err != nil {
			return 0, b.readErr()
		}

		if len(p) >= len(b.buf) {
			n, b.err = b.rd.Read(p)
			if n < 0 {
				panic(errNegativeRead)
			}
			return n, b.readErr()
		}
		b.fill()
		if b.r == b.w {
			return 0, b.readErr()
		}
	}

	n = copy(p, b.buf[b.r:b.w])
	b.r += n
	return n, nil
}

func (b *Reader) fill() {
	if b.r > 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}

	if b.w >= len(b.buf) {
		panic("buffio: tried to fill full buffer")
	}

	for i := maxConsecutiveEmptyReads; i > 0; i-- {
		n, err := b.rd.Read(b.buf[b.w:])
		if n < 0 {
			panic(errNegativeRead)
		}
		b.w += n
		if err != nil {
			b.err = err
			return
		}
		if n > 0 {
			return
		}
	}
	b.err = io.ErrNoProgress
}

func (b *Reader) ReadByte() (c byte, err error) {
	for b.r == b.w {
		if b.err != nil {
			return 0, b.readErr()
		}
		b.fill()
	}
	c = b.buf[b.r]
	b.r++
	return c, nil
}

func (b *Reader) readErr() (err error) {
	err = b.err
	b.err = nil
	return err
}
