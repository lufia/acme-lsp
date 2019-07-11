package main

import (
	"log"
	"strings"
	"time"

	"9fans.net/go/acme"
	"github.com/lufia/acme-lsp/outline"
)

type Win struct {
	acme *acme.Win
	tag  string
	f    *outline.File
}

func (w *Win) Execute(line string) bool {
	return false
}

func (w *Win) Look(text string) bool {
	return false
}

func start() error {
	acme.AutoExit(true)
	r, err := acme.Log()
	if err != nil {
		return err
	}
	defer r.Close()
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
			log.Println("NEW:", ev.ID, ev.Name)
			// acme.Open(ev.ID, ev.Name)
		case "del":
			log.Println("DEL:", ev.ID, ev.Name)
			// close(ev.ID)
		}
	}
}

func open(file string) (*Win, error) {
	var err error
	var w Win
	w.tag = "Def"
	w.acme, err = acme.New()
	if err != nil {
		time.Sleep(10 * time.Millisecond)
		w.acme, err = acme.New()
		if err != nil {
			return nil, err
		}
	}
	w.acme.Fprintf("tag", " %s ", w.tag)
	go w.acme.EventLoop(&w)
	return &w, nil
}
