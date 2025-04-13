package http

import (
	"fmt"
	"strconv"
)

type Response struct {
	statusLine StatusLine
	headers    map[string]string
	body       string
}

func (r *Response) String() string {
	response := r.statusLine.String()

	encodedBody := r.EncodedBody()
	r.headers["Content-Length"] = strconv.Itoa(len(encodedBody))

	for header, value := range r.headers {
		response += fmt.Sprintf("%s: %s\r\n", header, value)
	}
	response += "\r\n"

	response += encodedBody
	return response
}

func (r *Response) EncodedBody() string {
	encoding := r.headers["Content-Encoding"]
	if encoding == "" {
		return r.body
	}

	encodedBody, err := GetEncodedContent(encoding, r.body)
	if err != nil {
		fmt.Println("warning: ", err)
		return r.body
	}
	return encodedBody
}

func NewResponse(statusLine StatusLine, headers map[string]string, body string) *Response {
	return &Response{
		statusLine: statusLine,
		headers:    headers,
		body:       body,
	}
}
