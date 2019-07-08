// Package outline implements file address converter.
package outline

import (
	"io"
)

// Pos is a offset from top of a file.
type Pos uint

// Addr is a combination of line number and column number.
type Addr struct {
	Line uint // 0 origin
	Col  uint // 0 origin
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
	return nil, nil
}

// Pos returns the offset pointing to addr.
func (f *File) Pos(addr Addr) (Pos, error) {
	return 0, nil
}

// Addr returns the address pointing to p.
func (f *File) Addr(p Pos) (Addr, error) {
	return Addr{}, nil
}

// Update updates the contents of f.
func (f *File) Update(m, n Pos, text string) error {
	return nil
}
