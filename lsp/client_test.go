package lsp

import (
	"encoding/json"
	"path"
	"testing"
)

func TestPLS(t *testing.T) {
	conn, err := OpenCommand("gopls", "-v", "serve")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := NewClient(conn)
	c.Debug = true
	c.SetRootURI("testdata/pkg1")

	t.Logf("initialize")
	s, err := c.testSendRecv("initialize", &InitializeParams{
		RootURI: c.BaseURL.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("body: %s\n", s)

	t.Logf("initialized")
	err = c.testNotify("initialized", &InitializedParams{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("textDocument/didOpen")
	s, err = c.testSendRecv("textDocument/didOpen", &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        path.Join(c.BaseURL.String(), "pkg.go"),
			LanguageID: "go",
			Version:    1,
			Text:       "package pkg1\n\ntype Language struct {\n\tName string\n}\n\nfunc (l *Language) String() string {\n\treturn l.Name\n}\n",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("body: %s\n", s)
}

// textDocument/didChange
// ->textDocument/publishDiagnostics
// textDocument/definition
// textDocument/didClose

func (c *Client) testSendRecv(method string, p interface{}) ([]byte, error) {
	var s json.RawMessage
	if err := c.Call(method, p, &s); err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (c *Client) testNotify(method string, p interface{}) error {
	return c.Call(method, p, nil)
}
