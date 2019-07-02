package lsp

type InitializeParams struct {
	ProcessID int    `json:"processId,omitempty"`
	RootURI   string `json:"rootUri"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	//custom:null
}

type ServerCapabilities struct {
}

/*
{"textDocumentSync":{"openClose":true,"change":2,"save":{}},"hoverProvider":true,"completionProvider":{"triggerCharacters":["."]},"signatureHelpProvider":{"triggerCharacters":["(",","]},"definitionProvider":true,"referencesProvider":true,"documentHighlightProvider":true,"documentSymbolProvider":true,"codeActionProvider":true,"documentFormattingProvider":true,"renameProvider":true,"documentLinkProvider":{},"typeDefinitionProvider":true,"workspace":{"workspaceFolders":{"supported":true,"changeNotifications":"workspace/didChangeWorkspaceFolders"}}}
*/

type InitializedParams struct {
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}
