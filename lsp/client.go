package lsp

import (
	"bufio"
	"bytes"
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

// PipeConn represents a connection to a process.
type PipeConn struct {
	cmd *exec.Cmd
	r   io.ReadCloser
	w   io.WriteCloser
}

// OpenCommand returns a connection to executing command.
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

// Read reads bytes from stdout of c.
func (c *PipeConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

// Write writes b to stdin of c.
func (c *PipeConn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

// Close exits c.
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
	return err
}

type Message struct {
	Version string `json:"jsonrpc"`
	ID      int    `json:"id,omitempty"`
	Method  string `json:"method"`

	// This appears request or notification.
	Params json.RawMessage `json:"params,omitempty"`

	// These appears response only.
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents an error.
type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Error implements error interface.
func (e *ResponseError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type Call struct {
	Method string
	Args   interface{}
	Reply  interface{}
	Error  error

	msg  *Message
	done chan *Call
}

type Client struct {
	BaseURL *url.URL
	Event   chan *Message
	Debug   bool

	lastID int
	conn   io.ReadWriteCloser
	c      chan *Call
}

func NewClient(conn io.ReadWriteCloser) *Client {
	c := &Client{
		Event: make(chan *Message, 10),
		conn:  conn,
		c:     make(chan *Call),
	}
	go c.run()
	return c
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

// Call calls the method with args. If reply is nil,
// this don't wait for reply. Therefore it is notification.
func (c *Client) Call(method string, args, reply interface{}) error {
	r, err := c.makeRequest(method, args, reply)
	if err != nil {
		return err
	}
	call := &Call{
		Method: method,
		Args:   args,
		Reply:  reply,
		msg:    r,
		done:   make(chan *Call, 1),
	}
	c.c <- call
	call = <-call.done
	if call.Error != nil {
		return call.Error
	}
	return nil
}

func (c *Client) reader(replyc chan<- *Message) {
	defer close(replyc)
	r := bufio.NewReader(c.conn)
	for {
		msg, err := c.readMessage(r)
		if err == io.EOF {
			return
		}
		if err != nil {
			// TODO(lufia): where do we pass an error?
			return
		}
		replyc <- msg
	}
}

func (c *Client) run() {
	callc := c.c
	replyc := make(chan *Message, 1)
	go c.reader(replyc)

	cache := make(map[int]*Call)
	for callc != nil || replyc != nil {
		select {
		case msg, ok := <-replyc:
			if !ok {
				replyc = nil
				continue
			}
			if msg.Params != nil { // request from the server
				c.Event <- msg
				continue
			}

			call := cache[msg.ID]
			if call == nil {
				continue
			}
			delete(cache, msg.ID)
			if msg.Error != nil {
				call.Error = msg.Error
				call.done <- call
				continue
			}
			err := json.Unmarshal([]byte(msg.Result), call.Reply)
			if err != nil {
				call.Error = err
				call.done <- call
				continue
			}
			call.done <- call
		case call, ok := <-callc:
			if !ok {
				callc = nil
				continue
			}
			if err := c.writeJSON(call.msg); err != nil {
				call.Error = err
				call.done <- call
				continue
			}
			if call.msg.ID == 0 {
				call.done <- call
				continue
			}
			cache[call.msg.ID] = call
		}
	}
}

func (c *Client) readMessage(r *bufio.Reader) (*Message, error) {
	var contentLen int64
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			break
		}
		a := strings.SplitN(s, ":", 2)
		if len(a) < 2 {
			continue
		}
		switch strings.TrimSpace(a[0]) {
		case "Content-Length":
			v := strings.TrimSpace(a[1])
			contentLen, _ = strconv.ParseInt(v, 10, 64)
		}
	}

	buf := bytes.NewBuffer(make([]byte, 0, contentLen))
	if _, err := io.CopyN(buf, r, contentLen); err != nil {
		return nil, err
	}
	c.debugf("<- '%s'\n", buf.Bytes())
	var msg Message
	if err := json.Unmarshal(buf.Bytes(), &msg); err != nil {
		return nil, err
	}
	return &msg, nil
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

func (c *Client) Close() error {
	close(c.c)
	return c.conn.Close()
}

func (c *Client) makeRequest(method string, args, reply interface{}) (*Message, error) {
	params, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	var id int
	if reply != nil {
		c.lastID++
		id = c.lastID
	}
	return &Message{
		Version: "2.0",
		ID:      id,
		Method:  method,
		Params:  json.RawMessage(params),
	}, nil
}
