package lsp

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

/*
 * field:  type | null => pointer to type
 * field?: type        => omitempty
 */

const fileSchema = "file://"

// DocumentURI represents the interface described in the specification.
type DocumentURI string

// MarshalJSON implements json.Marshaler interface.
func (u DocumentURI) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(u) + `"`), nil
}

// String returns absolute URI.
func (u DocumentURI) String() string {
	n := len(fileSchema)
	return string(u[n:])
}

// SetRootURI updates c.BaseURL with s.
func (c *Client) SetRootURI(s string) error {
	if !path.IsAbs(s) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		s = path.Join(cwd, s)
	}
	var u url.URL
	u.Scheme = "file"
	u.Path = s
	c.BaseURL = &u
	return nil
}

// URL returns a document URI representation of s with c.BaseURL.
// If c.BaseURL is nil, client will assume it to be current directory.
func (c *Client) URL(s string) DocumentURI {
	if c.BaseURL == nil {
		c.SetRootURI(".")
	}
	if !path.IsAbs(s) {
		s = path.Join(c.BaseURL.Path, s)
	}
	return DocumentURI(fileSchema + path.Clean(s))
}

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
	TargetSelectionRange Range       `json:"targetSelectionRange"`
}

// TextEdit represents the interface described in the specification.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// TextDocumentIdentifier represents the interface described in the specification.
type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri"`
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
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
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
	// gopls don't support []LocationLink yet
	params.Capabilities.TextDocument.Definition.LinkSupport = false

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
	URI        DocumentURI `json:"uri"`
	LanguageID string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

// DidOpenTextDocument sends the document open notification to the server.
func (c *Client) DidOpenTextDocument(file, lang string) error {
	u := c.URL(file)
	b, err := ioutil.ReadFile(u.String())
	if err != nil {
		return err
	}
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        u,
			LanguageID: lang,
			Version:    1,
			Text:       string(b),
		},
	}
	call := c.Call("textDocument/didOpen", &params, nil)
	return c.Wait(call)
}

// DidChangeTextDocumentParams represents the interface described in the specification.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// VersionedTextDocumentIdentifier represents the interface described in the specification.
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version *int `json:"version"`
}

// TextDocumentContentChangeEvent represents the interface described in the specification.
type TextDocumentContentChangeEvent struct {
	Range       Range  `json:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// DidChangeTextDocument sends the document change notification to the server.
func (c *Client) DidChangeTextDocument(params *DidChangeTextDocumentParams) error {
	call := c.Call("textDocument/didChange", params, nil)
	return c.Wait(call)
}

// TextDocumentSaveReason represents reasons why a text document is saved.
const (
	TextDocumentSaveReasonManual     = 1
	TextDocumentSaveReasonAfterDelay = 2
	TextDocumentSaveReasonFocusOut   = 3
)

// WillSaveTextDocumentParams represents the interface described in the specification.
type WillSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Reason       int                    `json:"reason"`
}

// WillSaveTextDocument sends the document will save notification to the server.
func (c *Client) WillSaveTextDocument(params *WillSaveTextDocumentParams) error {
	call := c.Call("textDocument/willSave", params, nil)
	return c.Wait(call)
}

// TextEditsResult represents a result object for methods returning an array of TextEdit.
type TextEditsResult struct {
	TextEdits []TextEdit

	c    *Client
	call *Call
}

// WillSaveWaitUntilTextDocument sends the document will save request to the server.
func (c *Client) WillSaveWaitUntilTextDocument(params *WillSaveTextDocumentParams) *TextEditsResult {
	var result TextEditsResult
	result.c = c
	result.call = c.Call("textDocument/willSaveWaitUntil", params, &result.TextEdits)
	return &result
}

// DidSaveTextDocumentParams represents the interface described in the specification.
type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         string                 `json:"text,omitempty"`
}

// DidSaveTextDocument sends the document save notification to the server.
func (c *Client) DidSaveTextDocument(params *DidSaveTextDocumentParams) error {
	call := c.Call("textDocument/didSave", params, nil)
	return c.Wait(call)
}

// TextDocumentPositionParams represents the interface described in the specification.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// LocationsResult represents a result object for methods returning an array of Location.
type LocationsResult struct {
	Locations []Location

	c    *Client
	call *Call
}

// GotoDefinition sends the go to definition request to the server.
func (c *Client) GotoDefinition(params *TextDocumentPositionParams) *LocationsResult {
	var result LocationsResult
	result.c = c
	result.call = c.Call("textDocument/definition", params, &result.Locations)
	return &result
}

// Wait waits for a response of any request.
func (r *LocationsResult) Wait() error {
	return r.c.Wait(r.call)
}

// PublishDiagnosticsParams represents the interface described in the specification.
type PublishDiagnosticsParams struct {
	URI         DocumentURI  `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Diagnostics represents the interface described in the specification.
type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity,omitempty"`
	Code               string                         `json:"code,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

// DiagnosticRelatedInformation represents the interface described in the specification.
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}
