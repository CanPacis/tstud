package p2pjson

import "bytes"

type writer struct {
	n   int
	err error
}

func (r *writer) writeString(buf *bytes.Buffer, s string) {
	if r.err != nil {
		return
	}

	n, err := buf.WriteString(s)
	r.n += n
	r.err = err
}

func (r *writer) write(buf *bytes.Buffer, b []byte) {
	if r.err != nil {
		return
	}

	n, err := buf.Write(b)
	r.n += n
	r.err = err
}
