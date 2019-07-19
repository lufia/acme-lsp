// Package pkg1 implements test program.
package pkg1

import _ "io"

// Language represents language.
type Language struct {
	Name string
}

// String implements Stringer.
func (l *Language) String() string {
	return l.Name
}
