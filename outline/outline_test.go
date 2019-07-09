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
	}
	for _, tt := range tests {
		addr, err := f.Addr(tt.pos)
		if err != nil {
			t.Errorf("Addr(%v) = %v; want %v", tt.pos, err, tt.addr)
			continue
		}
		if addr != tt.addr {
			t.Errorf("Addr(%v) = %v; want %v", tt.pos, addr, tt.addr)
		}

		pos, err := f.Pos(tt.addr)
		if err != nil {
			t.Errorf("Pos(%v) = %v; want %v", tt.addr, err, tt.pos)
			continue
		}
		if pos != tt.pos {
			t.Errorf("Pos(%v) = %v; want %v", tt.addr, pos, tt.pos)
		}
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
		{pos: 14, err: errOutOfRange},
		{pos: 15, err: errOutOfRange},
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
		{addr: Addr{Line: 4, Col: 0}, err: errOutOfRange},
	}
	for _, tt := range tests {
		_, err := f.Pos(tt.addr)
		if err != tt.err {
			t.Errorf("Pos(%v) = %v; want %v", tt.addr, err, tt.err)
		}
	}
}
