package main

import (
	"log"

	"9fans.net/go/acme"
	"github.com/lufia/acme-lsp/lsp"
)

func main() {
	acme.AutoExit(true)
	conn, err := lsp.OpenCommand("gopls", "-v", "serve")
	if err != nil {
		log.Fatal(err)
	}
	c := lsp.NewClient(conn)
	if err := initialize(c); err != nil {
		log.Fatal(err)
	}
	log.Fatal(start(c))
}

func initialize(c *lsp.Client) error {
	r := c.Initialize(&lsp.InitializeParams{
		RootURI: c.URL("."),
	})
	if err := r.Wait(); err != nil {
		return err
	}
	if err := c.Initialized(&lsp.InitializedParams{}); err != nil {
		return err
	}
	return nil
}
