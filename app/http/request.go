package http

import (
	"errors"
	"fmt"
	"strings"
)

type RequestLine struct {
	Method      string
	Target      string
	HttpVersion string
}

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        string

	pattern *string
}

func (r *Request) DecodedBody(content string) (string, error) {
	encoding := r.Headers["Content-Encoding"]
	if encoding == "" {
		return content, nil
	}

	decodedBody, err := GetDecodedContent(encoding, content)
	if err != nil {
		return "", err
	}
	return decodedBody, nil
}

func (r *Request) Parse(buffer string) error {
	// Read request line
	requestLineIndex := strings.Index(buffer, "\r\n")
	if requestLineIndex == -1 {
		return errors.New("failed to parse request line")
	}

	requestLineString := buffer[:requestLineIndex]
	requestLine := strings.SplitN(requestLineString, " ", 3)
	if len(requestLine) != 3 {
		return errors.New("failed to parse request line")
	}

	r.RequestLine.Method = requestLine[0]
	r.RequestLine.Target = requestLine[1]
	r.RequestLine.HttpVersion = requestLine[2]

	// Read headers
	headers := make(map[string]string)
	headersIndex := requestLineIndex + 2
	for {
		headersString := buffer[headersIndex:]
		nextHeadersIndex := strings.Index(headersString, "\r\n")
		if nextHeadersIndex == -1 {
			return errors.New("failed to parse request headers")
		}

		headerString := buffer[headersIndex : nextHeadersIndex+headersIndex]
		header := strings.SplitN(headerString, ":", 2)

		// headers is empty
		if len(header) != 2 {
			headersIndex += 2
			break
		}

		headers[header[0]] = strings.TrimSpace(header[1])
		headersIndex += nextHeadersIndex + 2
	}
	r.Headers = headers

	// Read body
	body := strings.Trim(buffer[headersIndex:], "\x00")
	body, err := r.DecodedBody(body)
	if err != nil {
		return fmt.Errorf("failed to decode request body: %s", err)
	}

	r.Body = body

	return nil
}

func (r *Request) PathValue(name string) string {
	segments := strings.Split(*r.pattern, "/")
	routeSegments := strings.Split(r.RequestLine.Target, "/")

	for i, targetSerment := range segments {
		routeSegment := routeSegments[i]

		// check if template variable
		if strings.HasPrefix(targetSerment, "{") && strings.HasSuffix(targetSerment, "}") {
			paramName := targetSerment[1 : len(targetSerment)-1]
			if paramName == name {
				return routeSegment
			}
		}
	}

	return ""
}
