package outline

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewFile(t *testing.T) {
	tests := []struct {
		s    string
		want []Pos
	}{
		{
			s:    "test\naaa\nテxスxト\n", // multibyte
			want: []Pos{5, 4, 6, 0},
		},
		{
			s:    "test\naaa\n1",
			want: []Pos{5, 4, 1},
		},
		{
			s:    "",
			want: []Pos{0},
		},
	}
	for _, tt := range tests {
		r := strings.NewReader(tt.s)
		f, err := NewFile(r)
		if err != nil {
			t.Fatalf("NewFile(%q): %v", tt.s, err)
		}
		if !reflect.DeepEqual(f.v, tt.want) {
			t.Errorf("NewFile(%q) = %v; want %v", tt.s, f.v, tt.want)
		}
	}
}

func TestFileEmpty(t *testing.T) {
	r := strings.NewReader("")
	f, err := NewFile(r)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	tests := []struct {
		pos  Pos
		addr Addr
	}{
		{pos: 0, addr: Addr{Line: 0, Col: 0}},
	}
	for _, tt := range tests {
		testMutualConversion(t, f, tt.pos, tt.addr)
	}
}

func TestFileLocation(t *testing.T) {
	r := strings.NewReader("test\naaa\n\nxxx\n")
	f, err := NewFile(r)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	tests := []struct {
		pos  Pos
		addr Addr
	}{
		{pos: 0, addr: Addr{Line: 0, Col: 0}},
		{pos: 1, addr: Addr{Line: 0, Col: 1}},
		{pos: 1, addr: Addr{Line: 0, Col: 1}},
		{pos: 4, addr: Addr{Line: 0, Col: 4}},
		{pos: 5, addr: Addr{Line: 1, Col: 0}},
		{pos: 6, addr: Addr{Line: 1, Col: 1}},
		{pos: 9, addr: Addr{Line: 2, Col: 0}},
		{pos: 13, addr: Addr{Line: 3, Col: 3}},
		{pos: 14, addr: Addr{Line: 4, Col: 0}},
	}
	for _, tt := range tests {
		testMutualConversion(t, f, tt.pos, tt.addr)
	}
}

func testMutualConversion(t *testing.T, f *File, pos Pos, wantAddr Addr) {
	t.Helper()
	addr, err := f.Addr(pos)
	if err != nil {
		t.Errorf("Addr(%v) = %v; want %v", pos, err, wantAddr)
		return
	}
	if addr != wantAddr {
		t.Errorf("Addr(%v) = %v; want %v", pos, addr, wantAddr)
	}

	p, err := f.Pos(wantAddr)
	if err != nil {
		t.Errorf("Pos(%v) = %v; want %v", wantAddr, err, pos)
		return
	}
	if p != pos {
		t.Errorf("Pos(%v) = %v; want %v", wantAddr, p, pos)
	}
}

func TestFileAddrErr(t *testing.T) {
	r := strings.NewReader("test\naaa\n\nxxx\n")
	f, err := NewFile(r)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	tests := []struct {
		pos Pos
		err error
	}{
		{pos: 15, err: errOutOfRange},
		{pos: 16, err: errOutOfRange},
		{pos: 100, err: errOutOfRange},
	}
	for _, tt := range tests {
		_, err := f.Addr(tt.pos)
		if err != tt.err {
			t.Errorf("Addr(%v) = %v; want %v", tt.pos, err, tt.err)
		}
	}
}

func TestFilePosErr(t *testing.T) {
	r := strings.NewReader("test\naaa\n\nxxx\n")
	f, err := NewFile(r)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	tests := []struct {
		addr Addr
		err  error
	}{
		{addr: Addr{Line: 0, Col: 5}, err: errOutOfRange},
		{addr: Addr{Line: 3, Col: 4}, err: errOutOfRange},
		{addr: Addr{Line: 4, Col: 1}, err: errOutOfRange},
		{addr: Addr{Line: 5, Col: 0}, err: errOutOfRange},
	}
	for _, tt := range tests {
		_, err := f.Pos(tt.addr)
		if err != tt.err {
			t.Errorf("Pos(%v) = %v; want %v", tt.addr, err, tt.err)
		}
	}
}

func TestFileUpdate(t *testing.T) {
	r := strings.NewReader("")
	f, err := NewFile(r)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	update := func(m, n Pos, s, want string) {
		err := f.Update(m, n, s)
		if err != nil {
			t.Fatalf("Update(%d, %d, %q): %v", m, n, s, err)
		}
		r := strings.NewReader(want)
		v, err := makeOutline(r)
		if err != nil {
			t.Fatalf("makeOutline(%s): %v", want, err)
		}
		t.Logf("%d %d %q => %v", m, n, s, f.v)
		if !reflect.DeepEqual(f.v, v) {
			t.Errorf("%v; want %v", f.v, v)
		}
	}
	update(0, 0, "hello\n", "hello\n")
	update(5, 5, " world", "hello world\n")
	update(12, 12, "aaaa\nbbbb\nccc", "hello world\naaaa\nbbbb\nccc")
	update(16, 18, "X", "hello world\naaaXbbbb\nccc")
	update(1, 1, "", "hello world\naaaXbbbb\nccc")
	update(0, 12, "", "aaaXbbbb\nccc")
}
