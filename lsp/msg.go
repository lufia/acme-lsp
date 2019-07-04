package lsp

import (
	"io/ioutil"
	"path"
)

// InitializeParams represents the interface described in the specification.
type InitializeParams struct {
	ProcessID int    `json:"processId,omitempty"`
	RootURI   string `json:"rootUri,omitempty"`

	Capabilities ClientCapabilities `json:"capabilities"`

	Trace string `json:"trace,omitempty"` // off, message, verbose
}

// ClientCapabilities represents the interface described in the specification.
type ClientCapabilities struct {
	// TODO(lufia): implement
}

// InitializeResult represents the interface described in the specification.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`

	c    *Client
	call *Call
}

// ServerCapabilities represents the interface described in the specification.
type ServerCapabilities struct {
	TextDocumentSync          TextDocumentSyncOptions `json:"textDocumentSync"`
	HoverProvider             bool                    `json:"hoverProvider"`
	CompletionProvider        CompletionOptions       `json:"completionProvider"`
	SignatureHelpProvider     SignatureHelpOptions    `json:"signatureHelpProvider"`
	DefinitionProvider        bool                    `json:"definitionProvider"`
	ReferencesProvider        bool                    `json:"referencesProvider"`
	DocumentHighlightProvider bool                    `json:"documentHighlightProvider"`
	DocumentSymbolProvider    bool                    `json:"documentSymbolProvider"`
	WorkspaceSymbolProvider   bool                    `json:"workspaceSymbolProvider"`
	//CodeActionProvider bool or CodeActionOptions
	DocumentFormattingProvider bool `json:"documentFormattingProvider"`
	RenameProvider             bool `json:"renameProvider"`
}

//"documentLinkProvider"
//"typeDefinitionProvider"
//"workspace"

// TextDocumentSyncOptions represents the interface described in the specification.
type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"`
}

// CompletionOptions represents the interface described in the specification.
type CompletionOptions struct {
	ResolveProvider   bool     `json:"resolveProvider"`
	TriggerCharacters []string `json:"triggerCharacters"`
}

// SignatureHelpOptions represents the interface described in the specification.
type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
}

// Initialize sends the initialize request to the server.
func (c *Client) Initialize(params *InitializeParams) *InitializeResult {
	var result InitializeResult
	result.c = c
	result.call = c.Call("initialize", params, &result)
	return &result
}

// Wait waits for a response of initialize request.
func (r *InitializeResult) Wait() error {
	return r.c.Wait(r.call)
}

// InitializedParams represents the interface described in the specification.
type InitializedParams struct {
}

// Initialized sends the initialized notification to the server.
func (c *Client) Initialized(params *InitializedParams) error {
	return c.Wait(c.Call("initialized", params, nil))
}

// ShutdownResult represents result of shutdown response.
type ShutdownResult struct {
	c    *Client
	call *Call
}

// Shutdown sends the shutdown request to the server.
func (c *Client) Shutdown() *ShutdownResult {
	var result ShutdownResult
	result.c = c
	result.call = c.Call("shutdown", nil, &result)
	return &result
}

// Wait waits for a response of shutdown request.
func (r *ShutdownResult) Wait() error {
	return r.c.Wait(r.call)
}

// Exit sends the exit notification to the server.
func (c *Client) Exit() error {
	return c.Wait(c.Call("exit", nil, nil))
}

// DidOpenTextDocumentParams represents the interface described in the specification.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentItem represents the interface described in the specification.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenTextDocument sends the document open notification to the server.
func (c *Client) DidOpenTextDocument(file, lang string) error {
	f := path.Join(c.BaseURL.Path, file)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        path.Join(c.BaseURL.String(), file),
			LanguageID: lang,
			Version:    1,
			Text:       string(b),
		},
	}
	call := c.Call("textDocument/didOpen", &params, nil)
	return c.Wait(call)
}
