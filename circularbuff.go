// circularbuff project circularbuff.go
package circularbuff

// The circular buffer connects a Writer with one or more Readers. When data
// is inserted into the buffer, readers can read it until the buffer wraps
// around.

import (
	"container/list"
	"errors"
	"runtime"
	"sync"
)

const (
	defaultBufSize = 4096
)

type circularBuffer struct {
	err  error
	buf  []byte
	wpos uint
	rds  *list.List
	l    sync.Cond
	m    sync.Mutex
}

type circularReader struct {
	err  error
	cbuf *circularBuffer
	el   *list.Element
	rpos uint
}

type CircularBuffer interface {
	NewReader() *circularReader
}

type CircularReader interface{}

// Creates a new CircularWriter with the default buffer size.
func NewCircularWriter() *circularBuffer {
	return NewCircularWriterSize(defaultBufSize)
}

func NewCircularWriterSize(size int) *circularBuffer {

	if size < 2 {
		panic("size of the buffer cannot be less than two")
	} else {
		// This buffer uses MSB to detect buffer  wraps and thus needs to be a
		// power of two.
		if size&(^size+1) != size {
			n := 0
			for ; size > 0; size = size >> 1 {
				n++
			}
			size = 1 << uint(n)
		}
	}

	b := &circularBuffer{
		buf:  make([]byte, size),
		wpos: 0,
	}
	b.l.L = &b.m
	b.rds = list.New()
	return b
}

func wrapped(pos uint, lim uint) bool {
	return pos&uint(lim) != 0
}

func pos(pos uint, lim uint) int {
	return int(pos & uint(lim-1))
}

func (b *circularBuffer) wrapped() bool {
	// The MSB of the counter is used as a flag to indicate if the
	// counter wrapped.
	return wrapped(b.wpos, uint(len(b.buf)))
}

func (b *circularBuffer) pos() int {
	// Return only the uint value of the buffer lenght value.
	return pos(b.wpos, uint(len(b.buf)))
}

func (b *circularBuffer) Write(p []byte) (n int, err error) {

	n = 0
	s := 0
	if len(p) < len(b.buf) {
		s = len(p)
	} else {
		s = len(b.buf)
	}

	// Copy the data
	for n < s {
		n += copy(b.buf[b.pos():], p[n:])
	}

	b.wpos += uint(n)

	// Update the readers' pointers
	for e := b.rds.Front(); e != nil; e = e.Next() {
		// If the write pointer wrapped
		r, ok := e.Value.(circularReader)
		if ok && b.wrapped() != r.wrapped() && r.pos() < b.pos() {
			r.rpos += uint(b.pos() - r.pos())
		}
	}

	b.l.L.Lock()
	b.l.Broadcast()

	// Allow other threads to execute in a single core
	// environement.
	runtime.Gosched()

	b.l.L.Unlock()

	return
}

func (b *circularBuffer) NewReader() *circularReader {
	rd := &circularReader{
		cbuf: b,
		rpos: b.wpos,
	}
	rd.el = b.rds.PushFront(&rd)
	return rd
}

func (r *circularReader) Close() error {
	r.cbuf.rds.Remove(r.el)
	r.err = errors.New("the reader has been closed")
	return nil
}

func (r *circularReader) wrapped() bool {
	// The MSB of the counter is used as a flag to indicate if the
	// counter wrapped.
	return wrapped(r.rpos, uint(len(r.cbuf.buf)))
}

func (r *circularReader) pos() int {
	return pos(r.rpos, uint(len(r.cbuf.buf)))
}

func (r *circularReader) Read(p []byte) (n int, err error) {

	if r.err != nil {
		return 0, r.err
	}

	n = 0

	// Wait for data
	r.cbuf.l.L.Lock()
	for r.wrapped() == r.cbuf.wrapped() && r.pos() == r.cbuf.pos() {
		r.cbuf.l.Wait()
	}
	r.cbuf.l.L.Unlock()

	// Loop until both pointers are equal and in the same state.
	for r.wrapped() != r.cbuf.wrapped() || r.pos() != r.cbuf.pos() {
		var a int
		if r.wrapped() == r.cbuf.wrapped() {
			a += copy(p, r.cbuf.buf[r.pos():r.cbuf.pos()])
		} else {
			a += copy(p[r.cbuf.pos():], r.cbuf.buf[r.pos():])
		}
		n += a
		r.rpos += uint(a)
	}

	return
}
