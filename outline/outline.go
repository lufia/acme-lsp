// Package outline implements file address converter.
package outline

import (
	"bufio"
	"errors"
	"io"
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

// NewFile returns File initialized with contents of r.
func NewFile(r io.Reader) (*File, error) {
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
	return &File{v: v}, nil
}

var errOutOfRange = errors.New("out of range")

// Pos returns the offset pointing to addr.
func (f *File) Pos(addr Addr) (Pos, error) {
	if addr.Line >= uint(len(f.v)) {
		return 0, errOutOfRange
	}
	if addr.Col >= f.v[addr.Line] {
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
		if p < v {
			return Addr{Line: uint(i), Col: p}, nil
		}
		p -= v
	}
	return Addr{}, errOutOfRange
}

// Update updates the contents of f.
func (f *File) Update(m, n Pos, text string) error {
	return nil
}
