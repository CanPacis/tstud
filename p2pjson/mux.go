package p2pjson

import (
	"errors"
)

type Handler interface {
	ServeP2PJSON(*Request) *Response
}

type HandlerFunc func(r *Request) *Response

type Mux struct {
	handlers map[string]HandlerFunc
}

func (m *Mux) HandleFunc(path string, fn HandlerFunc) {
	m.handlers[path] = fn
}

func (m *Mux) ServeP2PJSON(r *Request) *Response {
	fn, ok := m.handlers[r.URL.Path]
	if !ok {
		return ErrorResponse(nil, StatusNotFound, errors.New("not found"))
	}

	return fn(r)
}

func NewMux() *Mux {
	return &Mux{
		handlers: map[string]HandlerFunc{},
	}
}
