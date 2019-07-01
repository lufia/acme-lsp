package lsp

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestPLS(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	t.Logf("initialize")
	s, err := c.testSendRecv("initialize", `{
		"processId": null,
		"rootUri": "testdata/pkg1"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("body: %s\n", s)

	t.Logf("textDocument/didOpen")
	s, err = c.testSendRecv("textDocument/didOpen", `{
		"textDocument": {
			"uri": "pkg.go",
			"languageId": "go",
			"version": 1,
			"text": "\n",
			"text1": "package pkg1\n\ntype Language struct {\n\tName string\n}\n\nfunc (l *Language) String() string {\n\treturn l.Name\n}\n"
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("body: %s\n", s)
}

// textDocument/didOpen
// textDocument/didChange
// ->textDocument/publishDiagnostics
// textDocument/definition
// textDocument/didClose

func (c *Client) testSendRecv(method, body string) ([]byte, error) {
	r, err := c.NewRequest(method, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}

	/*
		{
		    "jsonrpc": "2.0",
		    "id" : 2,
		    "method": "textDocument/didOpen",
		    "params": {
		        "textDocument": {
		            "uri": "testdata/pkg1/pkg.go"
		        }
		    }
		}
	*/
	return ioutil.ReadAll(resp.Body)
}
