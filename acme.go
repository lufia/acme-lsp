package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"
	"time"

	"9fans.net/go/acme"
	"github.com/lufia/acme-lsp/lsp"
	"github.com/lufia/acme-lsp/outline"
	"golang.org/x/xerrors"
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
		tag:  "Ref Doc",
		c:    c,
	}

	body, err := w.acme.ReadAll("body")
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
	if err := w.didOpenFile(body); err != nil {
		w.Close()
		return nil, err
	}
	return &w, nil
}

func (w *Win) DocumentID() lsp.TextDocumentIdentifier {
	return lsp.TextDocumentIdentifier{
		URI: w.c.URL(w.file),
	}
}

func (w *Win) didOpenFile(body []byte) error {
	return w.c.DidOpenTextDocument(&lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        w.c.URL(w.file),
			LanguageID: "go",
			Version:    1,
			Text:       string(body),
		},
	})
}

func (w *Win) didSave() error {
	return w.c.DidSaveTextDocument(&lsp.DidSaveTextDocumentParams{
		TextDocument: w.DocumentID(),
	})
}

func (w *Win) watch() {
	for e := range w.acme.EventChan() {
		if err := w.handleEvent(e); err != nil {
			w.acme.Errf("%v", err)
			continue
		}
	}
}

func (w *Win) handleEvent(e *acme.Event) error {
	p0 := outline.Pos(e.Q0)
	p1 := outline.Pos(e.Q1)
	s := string(e.Text)
	switch e.C2 {
	case 'I':
		off := p1 - p0
		return w.updateBody(p0, p1-off, s)
	case 'D':
		return w.updateBody(p0, p1, "")
	case 'x', 'X':
		return w.execute(e)
	case 'l', 'L':
		return w.look(e)
	}
	return nil
}

func (w *Win) updateBody(p0, p1 outline.Pos, s string) error {
	params, err := w.makeContentChangeEvent(p0, p1, s)
	if err != nil {
		return err
	}
	if err := w.c.DidChangeTextDocument(params); err != nil {
		return err
	}
	return w.f.Update(p0, p1, s)
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
			TextDocumentIdentifier: w.DocumentID(),
			Version:                nil, // TODO(lufia): implement
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

func (w *Win) execute(e *acme.Event) error {
	switch string(e.Text) {
	case "Put":
		return w.ExecPut()
	case "Ref":
		return w.ExecRef()
	case "Doc":
		return w.ExecDoc()
	case "Test":
		return xerrors.New("not implement")
	default:
		// TODO(lufia): kbd event will become an error.
		return w.acme.WriteEvent(e)
	}
}

// readCursor returns a beginning address pointed by cursor.
func (w *Win) readCursor() (int, error) {
	// Acme can't set addr to dot at only once
	// from a window is opened if addr isn't reset by 0.
	w.acme.Addr("0")

	if err := w.acme.Ctl("addr=dot"); err != nil {
		return 0, err
	}
	q0, _, err := w.acme.ReadAddr()
	if err != nil {
		return 0, err
	}
	return q0, nil
}

func (w *Win) look(e *acme.Event) error {
	addr, err := w.f.Addr(outline.Pos(e.Q0))
	if err != nil {
		return err
	}
	r := w.c.GotoDefinition(&lsp.TextDocumentPositionParams{
		TextDocument: w.DocumentID(),
		Position: lsp.Position{
			Line:      int(addr.Line),
			Character: int(addr.Col),
		},
	})
	if err := r.Wait(); err != nil {
		return w.acme.WriteEvent(e)
	}

	l := r.Locations[0]
	file := l.URI.String()
	q0, q1, err := rangeToPos(l.URI.String(), &l.Range)
	if err != nil {
		return err
	}
	w.printResult(file, q0, q1)
	return nil
}

func (w *Win) printResult(file string, q0, q1 int) {
	r, err := os.Open(file)
	if err != nil {
		w.acme.Errf("can't open %s: %v", file, err)
		return
	}
	defer r.Close()

	f, err := outline.NewFile(r)
	if err != nil {
		w.acme.Errf("can't read %s: %v", file, err)
		return
	}
	data, err := w.readRange(r, q0, q1)
	if err != nil {
		w.acme.Errf("can't read %s: %v", file, err)
		return
	}
	addr0, err := f.Addr(outline.Pos(q0))
	if err != nil {
		w.acme.Errf("%s:#%d: %v", file, q0, err)
		return
	}
	w.acme.Errf("%s:%d %s", file, addr0.Line+1, data)
}

func (w *Win) readRange(f io.ReadSeeker, q0, q1 int) ([]byte, error) {
	if _, err := f.Seek(int64(q0), 0); err != nil {
		return nil, err
	}
	buf := make([]byte, q1-q0)
	if _, err := f.Read(buf[:]); err != nil {
		return nil, err
	}
	return buf, nil
}

func (w *Win) ExecPut() error {
	defer w.acme.Ctl("put")
	return w.c.WillSave(&lsp.WillSaveTextDocumentParams{
		TextDocument: w.DocumentID(),
		Reason:       lsp.TextDocumentSaveReasonManual,
	})
}

func (w *Win) ExecRef() error {
	q, err := w.readCursor()
	if err != nil {
		return err
	}
	addr, err := w.f.Addr(outline.Pos(q))
	if err != nil {
		return err
	}
	result := w.c.References(&lsp.ReferenceParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: w.DocumentID(),
			Position: lsp.Position{
				Line:      int(addr.Line),
				Character: int(addr.Col),
			},
		},
		Context: lsp.ReferenceContext{
			IncludeDeclaration: false,
		},
	})
	if err := result.Wait(); err != nil {
		return err
	}
	for _, loc := range result.Locations {
		file := loc.URI.String()
		w.acme.Errf("%s:%d", file, loc.Range.Start.Line+1)
	}
	return nil
}

func (w *Win) ExecDoc() error {
	result := w.c.DocumentLink(&lsp.DocumentLinkParams{
		TextDocument: w.DocumentID(),
	})
	if err := result.Wait(); err != nil {
		return err
	}
	for _, link := range result.DocumentLinks {
		if link.Target != "" {
			w.acme.Errf("%s", string(link.Target))
		}
	}
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
	go func() {
		for msg := range c.Event {
			switch msg.Method {
			case "textDocument/publishDiagnostics":
				if !*debugFlag {
					continue
				}
				var params lsp.PublishDiagnosticsParams
				err := json.Unmarshal([]byte(msg.Params), &params)
				if err != nil {
					acme.Errf(".", "lsp: %s: %s", msg.Method, msg.Params)
					continue
				}
				file := params.URI.String()
				for _, v := range params.Diagnostics {
					q0, q1, err := rangeToPos(file, &v.Range)
					if err != nil {
						acme.Errf(file, "lsp: %s: %s", msg.Method, msg.Params)
						continue
					}
					acme.Errf(file, "%s:#%d,#%d %s", path.Base(file), q0, q1, v.Message)
				}
			default:
				acme.Errf(".", "lsp: %s: %s", msg.Method, msg.Params)
			}
		}
	}()

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
		if path.Ext(ev.Name) != ".go" {
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
		case "put":
			if w, ok := wins[ev.ID]; ok {
				w.didSave()
			}
		case "del":
			if w, ok := wins[ev.ID]; ok {
				w.Close()
			}
			delete(wins, ev.ID)
		}
	}
}
