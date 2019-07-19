// Package outline implements file address converter.
package outline

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// Pos is a offset from top of a file.
type Pos uint

// Addr is a combination of line number and column number.
type Addr struct {
	Line uint // 0 origin
	Col  Pos  // 0 origin
}

// File is a mapper to convert two kind addresses.
type File struct {
	// layout: v[lineno] = number of runes in this line (including \n).
	// v[0] = 13	# package main\n
	// v[1] =  1	# \n
	// v[2] =  9	# import (\n
	// v[3] =  0
	v []Pos
}

// Open returns a File initialized with contents of file.
func Open(file string) (*File, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return NewFile(r)
}

// NewFile returns a File initialized with contents of r.
func NewFile(r io.Reader) (*File, error) {
	v, err := makeOutline(r)
	if err != nil {
		return nil, err
	}
	return &File{v: v}, nil
}

func makeOutline(r io.Reader) ([]Pos, error) {
	b := bufio.NewReader(r)
	var v []Pos
	var c Pos
	for {
		r, _, err := b.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		c++
		if r == '\n' {
			v = append(v, c)
			c = 0
		}
	}
	v = append(v, c)
	return v, nil
}

var errOutOfRange = errors.New("out of range")

// Pos returns the offset pointing to addr.
func (f *File) Pos(addr Addr) (Pos, error) {
	if addr.Line >= uint(len(f.v)) {
		return 0, errOutOfRange
	}
	col := f.maxCol(addr.Line)
	if addr.Col > col {
		return 0, errOutOfRange
	}
	var p Pos
	for _, v := range f.v[:addr.Line] {
		p += v
	}
	return p + addr.Col, nil
}

// Addr returns the address pointing to p.
func (f *File) Addr(p Pos) (Addr, error) {
	for i, v := range f.v {
		col := f.maxCol(uint(i))
		if p <= col {
			return Addr{Line: uint(i), Col: p}, nil
		}
		p -= v
	}
	return Addr{}, errOutOfRange
}

func (f *File) maxCol(lineno uint) Pos {
	n := f.v[lineno]
	if n > 0 {
		n--
	}
	return n
}

// Update updates the contents of f.
func (f *File) Update(m, n Pos, text string) error {
	bp, err := f.Addr(m)
	if err != nil {
		return err
	}
	ep, err := f.Addr(n)
	if err != nil {
		return err
	}
	t, err := makeOutline(strings.NewReader(text))
	if err != nil {
		return err
	}

	v := make([]Pos, len(f.v)+len(t))
	p := copy(v, f.v[:bp.Line])
	if bp.Col > 0 || t[0] > 0 {
		v[p] += bp.Col + t[0]
	}
	if len(t) > 1 {
		p++
		p += copy(v[p:], t[1:]) - 1
	}
	v[p] += f.v[ep.Line] - ep.Col
	p++
	p += copy(v[p:], f.v[ep.Line+1:])
	f.v = v[:p]
	return nil
}
