package outline_test

import (
	"fmt"
	"log"

	"github.com/lufia/acme-lsp/outline"
)

func ExampleOpen() {
	f, err := outline.Open("testdata/open.txt")
	if err != nil {
		log.Fatal(err)
	}

	pos, err := f.Pos(outline.Addr{Line: 1, Col: 1})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(pos)
	// Output: 5
}
