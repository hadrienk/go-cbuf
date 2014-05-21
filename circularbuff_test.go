// circularbuff_test
package circularbuff

import (
	"testing"
)

func Test_SizeNotPowTwo(t *testing.T) {

	cb := NewCircularWriterSize(7)
	if len(cb.buf) != 8 {
		t.Errorf("Buffer size 7 should become 8, got %d", len(cb.buf))
	}

	cb = NewCircularWriterSize(8)
	if len(cb.buf) != 8 {
		t.Errorf("Buffer size 8 should stay 8, got %d", len(cb.buf))
	}

	cb = NewCircularWriterSize(9)
	if len(cb.buf) != 16 {
		t.Errorf("Buffer size 9 should become 16, got %d", len(cb.buf))
	}

}

func Test_wrapped(t *testing.T) {
	if wrapped(0, 8) {
		t.Errorf("0 should not be wrapped with 8")
	}
	if !wrapped(8, 8) {
		t.Errorf("8 should be wrapped with 8")
	}
	if wrapped(16, 8) {
		t.Errorf("16 should not be wrapped with 8")
	}
	if !wrapped(24, 8) {
		t.Errorf("24 should be wrapped with 8")
	}
}

func Test_pos(t *testing.T) {
	if pos(0, 8) != 0 {
		t.Errorf("0 should be 0 with 8, got %d", pos(0, 8))
	}
	if pos(8, 8) != 0 {
		t.Errorf("8 should be 0 with 8, got %d", pos(8, 8))
	}
	if pos(9, 8) != 1 {
		t.Errorf("9 should be 1 with 8, got %d", pos(9, 8))
	}
}

func Test_WriterWraps(t *testing.T) {

	var n int

	cb := NewCircularWriterSize(8)
	r := cb.NewReader()
	buf := make([]byte, 12)

	val := []byte("1234")

	// Will write everything.
	n, _ = cb.Write(val)
	if n != len(val) {
		t.Errorf("Could't write all bytes. %d writen, wanted %d", n, len(val))
	}

	if cb.wpos != uint(len(val)) {
		t.Errorf("Write pointer incorrect. Got %d, wanted %d", cb.wpos, len(val))
	}

	// Shoul read all.
	n, _ = r.Read(buf)
	if string(buf[:n]) != string(val) {
		t.Errorf("Buffer received %q, wanted %q", string(buf[:n]), val)
	}

	val2 := []byte("56789a")

	// This should wrap.
	n, _ = cb.Write(val2)
	if n != len(val2) {
		t.Errorf("Could't write all bytes. %d writen, wanted %d", n, len(val2))
	}

	if cb.wpos != uint(len(val)+len(val2)) {
		t.Errorf("Write pointer incorrect. Got %d, wanted %d", cb.wpos, len(val)+len(val2))
	}

	if !cb.wrapped() {
		t.Errorf("Buffer should have wrapped. %t received, wanted %t", cb.wrapped(), true)
	}

	// Shoul read all.
	n, _ = r.Read(buf)
	if string(buf[:n]) != string(val2) {
		t.Errorf("Buffer received %q, wanted %q", string(buf[:n]), val2)
	}

}
