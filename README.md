# acme-lsp

Acme-lsp is Language Server Protocol integration for Acme.

This is in development so it isn't stable yet. Please file a bug if you have find a bug.

## Installation

This app requiers [gopls](https://github.com/golang/go/wiki/gopls).

To install:

```console
$ go get golang.org/x/tools/gopls@latest
$ go get github.com/lufia/acme-lsp
```

## Usage

You can run `Local acme-lsp` by 3 button of mouse in Acme window anywhere, usually tag line. Then app starts watching events that Go source files is opened.

## Features

### Jump to definition or declaration
When 3 button is clicked on the top of token in the Go source file, acme-lsp searches definition or declaration of that token, then prints filename and line number onto other window in Acme.

If acme-lsp couldn't find definition or declaration of the token, will search the token as simple text within same file.

### Document

## TODO
- go-fmt before saving (textDocument/formatting)
- run go-test
