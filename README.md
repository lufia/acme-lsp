# acme-lsp

## Features

### Jump to definition or declaration
When 3 button is clicked on the top of token in the Go source file, acme-lsp searches definition or declaration of that token, then prints filename and line number onto other window in Acme.

If acme-lsp couldn't find definition or declaration of the token, will search the token as simple text within same file.

### Document

## TODO
- go-fmt before saving (textDocument/formatting)
- run go-test
