package lsp

import (
	"io/ioutil"
	"path"
)

/*
 * field:  type | null => pointer to type
 * field?: type        => omitempty
 */

type DocumentURI string

// Position represents the interface described in the specification.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents the interface described in the specification.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents the interface described in the specification.
type Location struct {
	URI   DocumentURI `json:"uri"`
	Range Range       `json:"range"`
}

// LocationLink represents the interface described in the specification.
type LocationLink struct {
	OriginSelectionRange *Range      `json:"originSelectionRange,omitempty"`
	TargetURI            DocumentURI `json:"targetUri"`
	TargetRange          Range       `json:"targetRange"`
	TargetSelectionRange `json:"targetSelectionRange"`
}

// InitializeParams represents the interface described in the specification.
type InitializeParams struct {
	ProcessID *int        `json:"processId"`
	RootURI   DocumentURI `json:"rootUri"`

	Capabilities ClientCapabilities `json:"capabilities"`

	Trace string `json:"trace,omitempty"` // off, message, verbose
}

// ClientCapabilities represents the interface described in the specification.
type ClientCapabilities struct {
	// TODO(lufia): implement
}

// TextDocumentClientCapabilities represents the interface described in the specification.
type TextDocumentClientCapabilities struct {
	Declaration struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
		LinkSupport         bool `json:"linkSupport,omitempty"`
	} `json:"declaration,omitempty"`
	Definition struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
		LinkSupport         bool `json:"linkSupport,omitempty"`
	} `json:"definition,omitempty"`
	TypeDefinition struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
		LinkSupport         bool `json:"linkSupport,omitempty"`
	} `json:"typeDefinition,omitempty"`
	Implementation struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
		LinkSupport         bool `json:"linkSupport,omitempty"`
	} `json:"implementation,omitempty"`
}

// InitializeResult represents the interface described in the specification.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`

	c    *Client
	call *Call
}

// ServerCapabilities represents the interface described in the specification.
type ServerCapabilities struct {
	// TODO(lufia): textDocumentSync is: TextDocumentSyncOption | number
	// TODO(lufia): missing
	// typeDefinitionProvider
	// implementationProvider
	// codeActionProvider
	// codeLensProvider
	// documentOnTypeFormattingProvider
	// renameProvider
	// documentLinkProvider
	// colorProvider
	// foldingRangeProvider
	// declarationProvider
	// workspace
	// experimental

	TextDocumentSync                TextDocumentSyncOptions `json:"textDocumentSync"`
	HoverProvider                   bool                    `json:"hoverProvider,omitempty"`
	CompletionProvider              CompletionOptions       `json:"completionProvider,omitempty"`
	SignatureHelpProvider           SignatureHelpOptions    `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider              bool                    `json:"definitionProvider,omitempty"`
	ReferencesProvider              bool                    `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider       bool                    `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider          bool                    `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider         bool                    `json:"workspaceSymbolProvider,omitempty"`
	DocumentFormattingProvider      bool                    `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider bool                    `json:"documentRangeFormattingProvider,omitempty"`
	ExecuteCommandProvider          ExecuteCommandOptions   `json:"executeCommandProvider,omitempty"`
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

// ExecuteCommandOptions represents the interface described in the specification.
type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

// Initialize sends the initialize request to the server.
func (c *Client) Initialize(params *InitializeParams) *InitializeResult {
	params.ClientCapabilities.Definition.LinkSupport = true

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
