package main

import (
	"bytes"
	"os"
	"strings"
	"time"

	"9fans.net/go/acme"
	"github.com/lufia/acme-lsp/lsp"
	"github.com/lufia/acme-lsp/outline"
)

type Win struct {
	file string
	acme *acme.Win
	tag  string
	c    *lsp.Client
	f    *outline.File
}

func OpenFile(id int, file string, c *lsp.Client) (*Win, error) {
	p, err := acme.Open(id, nil)
	if err != nil {
		time.Sleep(10 * time.Millisecond)
		p, err = acme.Open(id, nil)
		if err != nil {
			return nil, err
		}
	}
	w := Win{
		file: file,
		acme: p,
		tag:  "Def",
		c:    c,
	}

	body, err := p.ReadAll("body")
	if err != nil {
		w.Close()
		return nil, err
	}
	f, err := outline.NewFile(bytes.NewReader(body))
	if err != nil {
		w.Close()
		return nil, err
	}
	w.f = f
	w.acme.Fprintf("tag", " %s ", w.tag)
	if err := w.didOpenFile(); err != nil {
		w.Close()
		return nil, err
	}
	return &w, nil
}

func (w *Win) didOpenFile() error {
	return w.c.DidOpenTextDocument(w.file, "go")
}

func (w *Win) closeFile() error {
	// TODO(lufia): willClose
	return nil
}

func (w *Win) watch() {
	for e := range w.acme.EventChan() {
		ok, err := w.execute(e)
		if err != nil {
			w.acme.Errf("%v", err)
			continue
		}
		if !ok {
			// TODO(lufia): kbd event will become an error.
			if err := w.acme.WriteEvent(e); err != nil {
				w.acme.Errf("%v", err)
			}
		}
	}
}

func (w *Win) execute(e *acme.Event) (bool, error) {
	//w.acme.Errf("%c: Q={%d %d} %b Nb=%d Nr=%d %q %q %q", e.C2, e.Q0, e.Q1, e.Flag, e.Nb, e.Nr, e.Text, e.Arg, e.Loc)
	p0 := outline.Pos(e.Q0)
	p1 := outline.Pos(e.Q1)
	s := string(e.Text)
	switch e.C2 {
	case 'I':
		params, err := w.makeContentChangeEvent(p0, p1, s)
		if err != nil {
			return true, err
		}
		w.c.DidChangeTextDocument(params)
		return true, w.f.Update(p0, p1, string(e.Text))
	case 'D':
		params, err := w.makeContentChangeEvent(p0, p1, s)
		if err != nil {
			return true, err
		}
		w.c.DidChangeTextDocument(params)
		return true, w.f.Update(p0, p1, "")
	case 'x', 'X':
		if s == "Def" {
			return true, w.ExecDef()
		}
	case 'l', 'L':
	}
	return false, nil
}

func (w *Win) makeContentChangeEvent(p0, p1 outline.Pos, s string) (*lsp.DidChangeTextDocumentParams, error) {
	a0, err := w.f.Addr(p0)
	if err != nil {
		return nil, err
	}
	a1, err := w.f.Addr(p1)
	if err != nil {
		return nil, err
	}
	return &lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{
				URI: w.c.URL(w.file),
			},
			Version: nil, // TODO(lufia): implement
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      int(a0.Line),
						Character: int(a0.Col),
					},
					End: lsp.Position{
						Line:      int(a1.Line),
						Character: int(a1.Col),
					},
				},
				RangeLength: int(p1 - p0),
				Text:        s,
			},
		},
	}, nil
}

func (w *Win) ExecDef() error {
	if err := w.acme.Ctl("addr=dot"); err != nil {
		return err
	}
	q0, q1, err := w.acme.ReadAddr()
	if err != nil {
		return err
	}
	w.acme.Errf("Def: %d %d %v", q0, q1, err)
	p0 := outline.Pos(q0)
	addr, err := w.f.Addr(p0)
	if err != nil {
		return err
	}
	r := w.c.GotoDefinition(&lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: w.c.URL(w.file),
		},
		Position: lsp.Position{
			Line:      int(addr.Line),
			Character: int(addr.Col),
		},
	})
	if err := r.Wait(); err != nil {
		return err
	}

	l := r.Locations[0]
	file := l.URI.String()
	q0, q1, err = rangeToPos(l.URI.String(), &l.Range)
	if err != nil {
		return err
	}
	w.acme.Errf("%s:#%d,#%d", file, q0, q1)
	return nil
}

func rangeToPos(file string, r *lsp.Range) (q0, q1 int, err error) {
	fin, err := os.Open(file)
	if err != nil {
		return
	}
	defer fin.Close()

	f, err := outline.NewFile(fin)
	if err != nil {
		return
	}
	pos := func(p lsp.Position) (int, error) {
		v, err := f.Pos(outline.Addr{
			Line: uint(p.Line),
			Col:  outline.Pos(p.Character),
		})
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
	q0, err = pos(r.Start)
	if err != nil {
		return
	}
	q1, err = pos(r.End)
	if err != nil {
		return
	}
	return
}

func (w *Win) Close() {
	w.acme.CloseFiles()
}

func start(c *lsp.Client) error {
	r, err := acme.Log()
	if err != nil {
		return err
	}
	defer r.Close()

	wins := make(map[int]*Win)
	for {
		ev, err := r.Read()
		if err != nil {
			return err
		}
		if !strings.HasSuffix(ev.Name, ".go") {
			continue
		}
		switch ev.Op {
		case "new":
			w, err := OpenFile(ev.ID, ev.Name, c)
			if err != nil {
				acme.Errf("./log", "can't watch: %v", err)
				continue
			}
			wins[ev.ID] = w
			go w.watch()
		case "del":
			if w, ok := wins[ev.ID]; ok {
				w.Close()
			}
			delete(wins, ev.ID)
		}
	}
}
