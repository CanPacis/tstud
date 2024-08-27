package p2pjson

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"sync"
)

const Version = "P2PJSON/0.1"
const P2PJSONScheme = "p2pjson"
const RequestMessageType = "REQUEST"
const ResponseMessageType = "RESPONSE"
const ExitMessageType = "EXIT"

type Peer struct {
	rwc io.ReadWriteCloser
	mu  sync.Mutex

	sent map[uint]chan *Response
}

func (c *Peer) Request(r *Request) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.rwc.Write([]byte(fmt.Sprintf("%s\r\n", RequestMessageType))); err != nil {
		return nil, err
	}

	_, err := io.Copy(c.rwc, r)
	if err != nil {
		return nil, err
	}

	ch := make(chan *Response)
	c.sent[r.Identifier] = ch

	resp := <-ch
	return resp, nil
}

func (c *Peer) Respond(r *Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.rwc.Write([]byte(fmt.Sprintf("%s\r\n", ResponseMessageType))); err != nil {
		return err
	}

	_, err := io.Copy(c.rwc, r)
	return err
}

func (c *Peer) Listen(handler Handler) {
	r := textproto.NewReader(bufio.NewReader(c.rwc))

	for {
		typ, err := r.ReadLine()
		if err != nil {
			c.Respond(ErrorResponse(nil, StatusInternalServerError, err))
			continue
		}

		switch typ {
		case RequestMessageType:
			req := &Request{}
			_, err = req.ReadFrom(c.rwc)
			if err != nil {
				c.Respond(ErrorResponse(nil, StatusInternalServerError, err))
				continue
			}

			if resp := handler.ServeP2PJSON(req); resp != nil {
				c.Respond(resp)
			}
		case ResponseMessageType:
			resp := &Response{}
			_, err = resp.ReadFrom(c.rwc)
			if err != nil {
				c.Respond(ErrorResponse(nil, StatusInternalServerError, err))
				continue
			}

			reqCh, ok := c.sent[resp.Identifier]
			if ok {
				reqCh <- resp
				delete(c.sent, resp.Identifier)
			}
		case ExitMessageType:
			c.rwc.Close()
			return
		default:
			c.Respond(ErrorResponse(nil, StatusBadRequest, errors.New("unknown message type")))
			continue
		}
	}
}

func New(rwc io.ReadWriteCloser) *Peer {
	return &Peer{
		rwc:  rwc,
		mu:   sync.Mutex{},
		sent: map[uint]chan *Response{},
	}
}

type StdIOPeer struct {
	closed bool
}

func (p *StdIOPeer) Read(b []byte) (int, error) {
	if p.closed {
		return 0, os.ErrClosed
	}
	return os.Stdin.Read(b)
}

func (p *StdIOPeer) Write(b []byte) (int, error) {
	if p.closed {
		return 0, os.ErrClosed
	}
	return os.Stdout.Write(b)
}

func (p *StdIOPeer) Close() error {
	if p.closed {
		return os.ErrClosed
	}
	p.closed = true
	return nil
}

func NewStdIOPeer() *StdIOPeer {
	return &StdIOPeer{}
}

func ErrorResponse(r *Request, code int, err error) *Response {
	data := map[string]any{"error": err.Error()}
	encoded, _ := json.Marshal(data)

	return NewResponse(r, code, bytes.NewBuffer(encoded))
}
