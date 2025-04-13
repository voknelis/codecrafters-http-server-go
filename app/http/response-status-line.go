package http

import "fmt"

type StatusLine struct {
	httpVersion      string
	statusCode       int
	statusCodePhrase string
}

func (s *StatusLine) String() string {
	return fmt.Sprintf("%s %d %s\r\n", s.httpVersion, s.statusCode, s.statusCodePhrase)
}

func NewStatusLine(httpVersion string, statusCode int) *StatusLine {
	statusText := StatusText(statusCode)
	return &StatusLine{
		httpVersion:      httpVersion,
		statusCode:       statusCode,
		statusCodePhrase: statusText,
	}
}
