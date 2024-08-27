package p2pjson

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	neturl "net/url"
	"strconv"
	"strings"
	"sync"
)

type Request struct {
	Identifier uint
	URL        *neturl.URL
	Header     textproto.MIMEHeader
	Body       io.Reader

	ctx context.Context

	w writer
}

func (r *Request) Read(b []byte) (int, error) {
	buf := bytes.NewBuffer([]byte{})
	r.w.writeString(buf, fmt.Sprintf("%s %s\r\n", r.URL.String(), Version))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r.w.n, err
	}

	r.Header.Set("Identifier", fmt.Sprintf("%d", r.Identifier))
	r.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	for key, value := range r.Header {
		header := fmt.Sprintf("%s: %s\r\n", key, strings.Join(value, " "))
		r.w.writeString(buf, header)
	}

	r.w.writeString(buf, "\r\n")
	r.w.write(buf, body)

	if r.w.err != nil {
		return r.w.n, r.w.err
	}
	r.w.n = copy(b, buf.Bytes())
	return r.w.n, io.EOF
}

func (req *Request) ReadFrom(ir io.Reader) (int64, error) {
	buf := bytes.NewBuffer([]byte{})
	r := textproto.NewReader(bufio.NewReader(io.TeeReader(ir, buf)))

	p2p, err := r.ReadLine()
	if err != nil {
		return int64(buf.Len()), err
	}
	split := strings.Split(p2p, " ")
	if len(split) < 2 {
		return int64(buf.Len()), errors.New("malformed request")
	}
	req.URL, err = neturl.Parse(split[0])
	if err != nil {
		return int64(buf.Len()), err
	}
	if req.URL.Scheme != P2PJSONScheme {
		return int64(buf.Len()), errors.New("malformed request, invalid url host (should use 'p2pjson')")
	}

	req.Header, err = r.ReadMIMEHeader()
	if err != nil {
		return int64(buf.Len()), err
	}

	identifier, err := extractInt(req.Header, "Identifier")
	if err != nil {
		return int64(buf.Len()), err
	}
	req.Identifier = uint(identifier)

	contentLength, err := extractInt(req.Header, "Content-Length")
	if err != nil {
		return int64(buf.Len()), err
	}

	req.Body = io.LimitReader(ir, int64(contentLength))
	return int64(buf.Len()), nil
}

type reqCtxKeyType string

const reqCtxKey reqCtxKeyType = "req-ctx-key"

func (req *Request) Context() context.Context {
	if req.ctx == nil {
		req.ctx = context.WithValue(context.Background(), reqCtxKey, 0)
	}

	return req.ctx
}

func (req *Request) Set(key string, value any) {
	req.ctx = context.WithValue(req.Context(), reqCtxKeyType(key), value)
}

func (req *Request) Get(key string) any {
	return req.Context().Value(reqCtxKeyType(key))
}

func NewRequest(url string, body io.Reader) *Request {
	u, _ := neturl.Parse(url)

	return &Request{
		Identifier: idGen.next(),
		URL:        u,
		Header:     make(textproto.MIMEHeader),
		Body:       body,

		w: writer{},
	}
}

func extractInt(h textproto.MIMEHeader, key string) (int, error) {
	raw := h.Get(key)
	if len(raw) == 0 {
		return 0, errors.New("malformed header value, expected int")
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("malformed header value, expected int: %w", err)
	}

	return value, nil
}

type idGenerator struct {
	val uint
	mu  sync.Mutex
}

func (g *idGenerator) next() uint {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := g.val
	g.val++
	return id
}

var idGen = idGenerator{val: 1}
