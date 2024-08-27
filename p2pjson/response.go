package p2pjson

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	neturl "net/url"
	"strconv"
	"strings"
)

type Response struct {
	Request    *Request
	Identifier uint
	URL        *neturl.URL
	Header     textproto.MIMEHeader
	Body       io.Reader
	StatusCode int
	Status     string

	w writer
}

func (r *Response) Read(b []byte) (int, error) {
	buf := bytes.NewBuffer([]byte{})
	r.w.writeString(buf, fmt.Sprintf("%s %d %s\r\n", Version, r.StatusCode, r.Status))

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

func (resp *Response) ReadFrom(ir io.Reader) (int64, error) {
	buf := bytes.NewBuffer([]byte{})
	r := textproto.NewReader(bufio.NewReader(io.TeeReader(ir, buf)))

	p2p, err := r.ReadLine()
	if err != nil {
		return int64(buf.Len()), err
	}
	split := strings.Split(p2p, " ")
	if len(split) < 3 {
		return int64(buf.Len()), errors.New("malformed request")
	}
	resp.StatusCode, err = strconv.Atoi(split[1])
	if err != nil {
		return int64(buf.Len()), err
	}
	resp.Status = split[2]

	resp.Header, err = r.ReadMIMEHeader()
	if err != nil {
		return int64(buf.Len()), err
	}

	identifier, err := extractInt(resp.Header, "Identifier")
	if err != nil {
		return int64(buf.Len()), err
	}
	resp.Identifier = uint(identifier)

	contentLength, err := extractInt(resp.Header, "Content-Length")
	if err != nil {
		return int64(buf.Len()), err
	}

	resp.Body = io.LimitReader(ir, int64(contentLength))
	return int64(buf.Len()), nil
}

func NewResponse(r *Request, code int, body io.Reader) *Response {
	var id uint = 0
	var url *neturl.URL
	if r != nil {
		id = r.Identifier
		url = r.URL
	} else {
		url, _ = neturl.Parse("p2pjson://response.default/notification")
	}

	return &Response{
		Identifier: id,
		URL:        url,
		Request:    r,
		Header:     make(textproto.MIMEHeader),
		Body:       body,
		Status:     StatusText(code),
		StatusCode: code,

		w: writer{},
	}
}

func StatusText(code int) string {
	switch code {
	case 105:
		return "NOTIFICATION"
	default:
		return http.StatusText(code)
	}
}
