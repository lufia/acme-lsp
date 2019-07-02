package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

type PipeConn struct {
	cmd *exec.Cmd
	r   io.ReadCloser
	w   io.WriteCloser
}

func OpenCommand(name string, args ...string) (*PipeConn, error) {
	cmd := exec.Command(name, args...)
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
	return &PipeConn{cmd: cmd, r: r, w: w}, nil
}

func (c *PipeConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *PipeConn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func (c *PipeConn) Close() error {
	var err error
	catch := func(e error) {
		if err == nil {
			err = e
		}
	}
	if err := c.w.Close(); err != nil {
		catch(err)
	}
	if err := c.r.Close(); err != nil {
		catch(err)
	}
	if err := c.cmd.Process.Kill(); err != nil {
		catch(err)
	}
	if err := c.cmd.Wait(); err != nil {
		catch(err)
	}
	return xerrors.Errorf("can't kill gopls: %w", err)
}

type Client struct {
	BaseURL *url.URL
	Debug   bool

	nextID int
	conn   io.ReadWriteCloser
}

func NewClient(conn io.ReadWriteCloser) *Client {
	return &Client{conn: conn}
}

func (c *Client) debugf(format string, args ...interface{}) {
	if c.Debug {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

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

func (c *Client) Close() error {
	return c.conn.Close()
}

type Notification struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

func (c *Client) NewNotification(method string, p interface{}) (*Notification, error) {
	r := &Notification{
		Version: "2.0",
		Method:  method,
		Params:  p,
	}
	return r, nil
}

type Request struct {
	Version string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

func (c *Client) NewRequest(method string, p interface{}) (*Request, error) {
	c.nextID++
	r := &Request{
		Version: "2.0",
		ID:      c.nextID,
		Method:  method,
		Params:  p,
	}
	return r, nil
}

type Response struct {
	Version string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Result  interface{}    `json:"result,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (c *Client) Call(args, reply interface{}) error {
	if err := c.writeJSON(args); err != nil {
		return err
	}
	if reply == nil {
		return nil
	}

	rbuf := bufio.NewReader(c.conn)
	var contentLen int64
	for {
		s, err := rbuf.ReadString('\n')
		if err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			break
		}
		if n := len(s); s[n-1] == '\r' {
			s = s[:n-1]
		}
		a := strings.SplitN(s, ":", 2)
		if len(a) < 2 {
			continue
		}
		switch strings.TrimSpace(a[0]) {
		case "Content-Length":
			contentLen, _ = strconv.ParseInt(strings.TrimSpace(a[1]), 10, 64)
		}
	}

	var resp Response
	resp.Result = reply
	d := json.NewDecoder(io.LimitReader(rbuf, contentLen))
	if err := d.Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (c *Client) writeJSON(args interface{}) error {
	p, err := json.Marshal(args)
	if err != nil {
		return xerrors.Errorf("can't marshal: %w", err)
	}
	c.debugf("-> '%s'\n", p)
	_, err = fmt.Fprintf(c.conn, "Content-Length: %d\r\n\r\n", len(p))
	if err != nil {
		return xerrors.Errorf("can't write: %w", err)
	}
	_, err = c.conn.Write(p)
	if err != nil {
		return xerrors.Errorf("can't write: %w", err)
	}
	return nil
}
