package lsp

import (
	"encoding/json"
	"path"
	"testing"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		body   string
		params bool
		result bool
	}{
		{
			body:   `{"id":1,"method":"test","params":{}}`,
			params: true,
		},
		{
			body:   `{"id":1,"method":"test","params":null}`,
			params: true,
		},
		{
			body:   `{"id":1,"method":"test","result":{}}`,
			result: true,
		},
		{
			body:   `{"id":1,"method":"test","result":null}`,
			result: true,
		},
	}
	for _, tt := range tests {
		var msg Message
		err := json.Unmarshal([]byte(tt.body), &msg)
		if err != nil {
			t.Fatalf("can't marshal: '%v'", tt.body)
		}
		if tt.params {
			if msg.Params == nil {
				t.Errorf("Marshal('%v') should have Params", tt.body)
			}
		} else {
			if msg.Params != nil {
				t.Errorf("Marshal('%v') shouldn't have Params", tt.body)
			}
		}
		if tt.result {
			if msg.Result == nil {
				t.Errorf("Marshal('%v') should have Result", tt.body)
			}
		} else {
			if msg.Result != nil {
				t.Errorf("Marshal('%v') shouldn't have Result", tt.body)
			}
		}
	}
}

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
	err = c.testNotify("textDocument/didOpen", &DidOpenTextDocumentParams{
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

	t.Logf("shutdown")
	s, err = c.testSendRecv("shutdown", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("body: %s\n", s)

	t.Logf("exit")
	err = c.testNotify("exit", nil)
	if err != nil {
		t.Fatal(err)
	}

	c.Close()
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
