package http

import (
	"fmt"
	"io"
)

type ResponseWriter struct {
	w        io.Writer
	protocol string

	Headers    map[string]string
	StatusCode int
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	statusLine := NewStatusLine(r.protocol, r.StatusCode)

	response := NewResponse(*statusLine, r.Headers, string(b))
	responseString := response.String()

	fmt.Println("Response sent:\n", responseString)
	fmt.Println("——————————————")

	return r.w.Write([]byte(responseString))
}

func NewResponseWriter(w io.Writer, protocol string) *ResponseWriter {
	return &ResponseWriter{
		w:        w,
		protocol: protocol,

		Headers:    make(map[string]string),
		StatusCode: StatusOK,
	}
}
