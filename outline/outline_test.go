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
