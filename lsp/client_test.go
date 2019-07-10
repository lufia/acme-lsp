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
	c.Debug = testing.Verbose()
	c.SetRootURI("testdata/pkg1")

	t.Run("initialize", func(t *testing.T) {
		result := c.Initialize(&InitializeParams{
			RootURI: c.URL("."),
			Trace:   "verbose",
		})
		if err := result.Wait(); err != nil {
			t.Errorf("Wait(): %v", err)
		}
		t.Logf("body: %v\n", *result)
	})

	t.Run("initialized", func(t *testing.T) {
		if err := c.Initialized(&InitializedParams{}); err != nil {
			t.Errorf("Initialized(): %v", err)
		}
	})

	t.Run("textDocument/didOpen", func(t *testing.T) {
		if err := c.DidOpenTextDocument("pkg.go", "go"); err != nil {
			t.Errorf("DidOpenTextDocument(): %v", err)
		}
	})

	t.Run("textDocument/definition", func(t *testing.T) {
		result := c.GotoDefinition(&TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: DocumentURI(c.URL("pkg.go").String()),
			},
			Position: Position{
				Line:      6,
				Character: 10,
			},
		})
		if err := result.Wait(); err != nil {
			t.Errorf("Wait(): %v", err)
			return
		}
		if n := len(result.Locations); n != 1 {
			t.Errorf("len(Locations) = %d; want 1", n)
			return
		}
		want := Range{
			Start: Position{Line: 2, Character: 5},
			End:   Position{Line: 2, Character: 13},
		}
		if loc := result.Locations[0]; loc.Range != want {
			t.Errorf("Location.Range = %v; want %v", loc.Range, want)
		}
	})

	t.Run("shutdown", func(t *testing.T) {
		result := c.Shutdown()
		if err := result.Wait(); err != nil {
			t.Errorf("Wait(): %v", err)
		}
		t.Logf("body: %v\n", *result)
	})

	t.Run("exit", func(t *testing.T) {
		if err = c.Exit(); err != nil {
			t.Errorf("Exit(): %v", err)
		}
	})

	c.Close()
}

// textDocument/didChange
// ->textDocument/publishDiagnostics
// textDocument/didClose
