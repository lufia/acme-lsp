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
	Debug bool

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
	n, err := c.r.Read(b)
	if c.Debug {
		fmt.Fprintf(os.Stderr, "<- '%s'\n", b)
	}
	return n, err
}

func (c *PipeConn) Write(b []byte) (int, error) {
	if c.Debug {
		fmt.Fprintf(os.Stderr, "-> '%s'\n", b)
	}
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

	nextID int
	conn   io.ReadWriteCloser
}

func NewClient(conn io.ReadWriteCloser) *Client {
	return &Client{conn: conn}
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

func (r *Notification) WriteTo(w io.Writer) (int64, error) {
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

func (c *Client) Do(r io.WriterTo, p interface{}) error {
	if _, err := r.WriteTo(c.conn); err != nil {
		return xerrors.Errorf("can't write a request: %w", err)
	}
	if p == nil {
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
	resp.Result = p
	d := json.NewDecoder(io.LimitReader(rbuf, contentLen))
	if err := d.Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}
