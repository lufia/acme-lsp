package lsp

import (
	"encoding/json"
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
	result := c.Initialize(&InitializeParams{
		RootURI: c.BaseURL.String(),
		Trace:   "verbose",
	})
	if err := result.Wait(); err != nil {
		t.Fatal(err)
	}
	t.Logf("body: %v\n", *result)

	t.Logf("initialized")
	if err := c.Initialized(&InitializedParams{}); err != nil {
		t.Fatal(err)
	}

	t.Logf("textDocument/didOpen")
	if err := c.DidOpenTextDocument("pkg.go", "go"); err != nil {
		t.Fatal(err)
	}

	t.Logf("shutdown")
	result1 := c.Shutdown()
	if err := result1.Wait(); err != nil {
		t.Fatal(err)
	}
	t.Logf("body: %v\n", *result1)

	t.Logf("exit")
	if err = c.Exit(); err != nil {
		t.Fatal(err)
	}

	c.Close()
}

// textDocument/didChange
// ->textDocument/publishDiagnostics
// textDocument/definition
// textDocument/didClose
