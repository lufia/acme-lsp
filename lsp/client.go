package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

type Client struct {
	nextID int

	cmd *exec.Cmd
	r   io.ReadCloser
	w   io.WriteCloser
}

func NewClient() (*Client, error) {
	cmd := exec.Command("gopls", "-v", "serve")
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, xerrors.Errorf("can't pipe: %w", err)
	}
	r, err := cmd.StdoutPipe()
	if err != nil {
		w.Close()
		return nil, xerrors.Errorf("can't pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		r.Close()
		w.Close()
		return nil, xerrors.Errorf("can't start gopls: %w", err)
	}
	return &Client{nextID: 1, cmd: cmd, r: r, w: w}, nil
}

func (c *Client) Close() error {
	if err := c.cmd.Process.Kill(); err != nil {
		return xerrors.Errorf("can't kill gopls: %w", err)
	}
	return c.cmd.Wait()
}

type Request struct {
	Version string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func (c *Client) NewRequest(method string, p io.Reader) (*Request, error) {
	b, err := ioutil.ReadAll(p)
	if err != nil {
		return nil, xerrors.Errorf("can't read request body: %w", err)
	}
	c.nextID++
	r := &Request{
		Version: "2.0",
		ID:      c.nextID,
		Method:  method,
		Params:  json.RawMessage(b),
	}
	return r, nil
}

func (r *Request) WriteTo(w io.Writer) (int64, error) {
	p, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}
	written, _ := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(p))
	n, err := w.Write(p)
	if err != nil {
		return 0, err
	}
	written += n
	return int64(written), nil
}

type Response struct {
	Body io.Reader
}

func (c *Client) Do(r *Request) (*Response, error) {
	if _, err := r.WriteTo(c.w); err != nil {
		return nil, err
	}

	rbuf := bufio.NewReader(c.r)
	var contentLen int64
	for {
		s, err := rbuf.ReadString('\n')
		if err != nil {
			return nil, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			break
		}
		if n := len(s); s[n-1] == '\r' {
			s = s[:n-1]
		}
		kv := strings.SplitN(s, ":", 2)
		if len(kv) < 2 {
			continue
		}
		switch strings.TrimSpace(kv[0]) {
		case "Content-Length":
			contentLen, _ = strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
		}
	}

	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, rbuf, contentLen); err != nil {
		return nil, xerrors.Errorf("can't read respnse: %w", err)
	}
	return &Response{Body: &buf}, nil
}
