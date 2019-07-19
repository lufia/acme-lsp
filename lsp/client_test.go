package lsp

import (
	"encoding/json"
	"io/ioutil"
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
			t.Errorf("Initialize: %v", err)
		}
		t.Logf("body: %v\n", *result)
	})

	t.Run("initialized", func(t *testing.T) {
		if err := c.Initialized(&InitializedParams{}); err != nil {
			t.Errorf("Initialized: %v", err)
		}
	})

	t.Run("textDocument/didOpen", func(t *testing.T) {
		u := c.URL("pkg.go")
		b, err := ioutil.ReadFile(u.String())
		if err != nil {
			t.Fatal(err)
		}
		err = c.DidOpenTextDocument(&DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{
				URI:        u,
				LanguageID: "go",
				Version:    1,
				Text:       string(b),
			},
		})
		if err != nil {
			t.Errorf("DidOpenTextDocument: %v", err)
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
			t.Errorf("GotoDefinition: %v", err)
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

	t.Run("textDocument/willSave", func(t *testing.T) {
		err := c.WillSaveTextDocument(&WillSaveTextDocumentParams{
			TextDocument: TextDocumentIdentifier{
				URI: c.URL("pkg.go"),
			},
			Reason: TextDocumentSaveReasonManual,
		})
		if err != nil {
			t.Errorf("WillSaveTextDocument: %v", err)
		}
	})

	t.Run("textDocument/willSaveWaitUntil", func(t *testing.T) {
		if !c.cap.TextDocumentSync.WillSaveWaitUntil {
			t.Skip()
			return
		}
		result := c.WillSaveWaitUntilTextDocument(&WillSaveTextDocumentParams{
			TextDocument: TextDocumentIdentifier{
				URI: c.URL("pkg.go"),
			},
			Reason: TextDocumentSaveReasonAfterDelay,
		})
		if err := result.Wait(); err != nil {
			t.Errorf("WillSaveWaitUntilTextDocument: %v", err)
		}
		t.Logf("body: %v\n", *result)
	})

	t.Run("shutdown", func(t *testing.T) {
		result := c.Shutdown()
		if err := result.Wait(); err != nil {
			t.Errorf("Shutdown: %v", err)
		}
		t.Logf("body: %v\n", *result)
	})

	t.Run("exit", func(t *testing.T) {
		if err = c.Exit(); err != nil {
			t.Errorf("Exit: %v", err)
		}
	})

	c.Close()
}

// textDocument/didChange
// ->textDocument/publishDiagnostics
// textDocument/didClose

func TestClientEventOverflow(t *testing.T) {
	conn, err := OpenCommand("gopls", "-v", "serve")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	c := NewClient(conn)

	m := &Message{
		Version: "2.0",
		Method:  "test/didReceiveMessage",
		Params:  json.RawMessage(`{}`),
	}

	// fulfill c.Event buffers
	for i := 0; i < cap(c.Event); i++ {
		c.Event <- m
	}

	result1 := c.Initialize(&InitializeParams{
		RootURI: c.URL("."),
	})
	if err := result1.Wait(); err != nil {
		t.Errorf("Wait(): %v", err)
	}
	if err := c.Initialized(&InitializedParams{}); err != nil {
		t.Errorf("Initialized(): %v", err)
	}

	result3 := c.Shutdown()
	if err := result3.Wait(); err != nil {
		t.Errorf("Wait(): %v", err)
	}
}
