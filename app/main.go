package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Response struct {
	statusLine StatusLine
	headers    map[string]string
	body       string
}

func (r *Response) String() string {
	response := r.statusLine.String()

	for header, value := range r.headers {
		response += fmt.Sprintf("%s: %s\r\n", header, value)
	}
	response += "\r\n"

	response += r.body
	return response
}

func NewResponse(statusLine StatusLine, headers map[string]string, body string) *Response {
	return &Response{
		statusLine: statusLine,
		headers:    headers,
		body:       body,
	}
}

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

type RequestLine struct {
	Method      string
	Target      string
	HttpVersion string
}

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        string
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
	r.Body = body

	return nil
}

const (
	StatusOK         = 200
	StatusCreated    = 201
	StatusBadRequest = 400
	StatusNotFound   = 404
)

func StatusText(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusBadRequest:
		return "BadRequest"
	case StatusNotFound:
		return "Not Found"
	default:
		return ""
	}
}

var AcceptEncoding = []string{"gzip"}

func main() {
	err := startTCPServer(":4221")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startTCPServer(port string) error {
	fmt.Printf("Starting TCP server at port %s\n", port)
	defer fmt.Println("Closing TCP server")

	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to bind to port %s", port)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %s", err.Error())
		}

		go func() {
			err = handleConnection(conn)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}
		}()
	}
}

func handleConnection(conn net.Conn) error {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading connection: %s", err.Error())
	}
	fmt.Println("Response read: ", string(buffer))

	request := Request{}
	request.Parse(string(buffer))
	path := request.RequestLine.Target

	var statusLine *StatusLine
	headers := make(map[string]string)
	body := ""

	if path == "/" {
		statusLine = NewStatusLine("HTTP/1.1", StatusOK)
	} else if strings.HasPrefix(path, "/echo/") {
		statusLine = NewStatusLine("HTTP/1.1", StatusOK)

		echoValue, _ := strings.CutPrefix(path, "/echo/")

		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = strconv.Itoa(len(echoValue))

		encoding := request.Headers["Accept-Encoding"]
		if slices.Contains(AcceptEncoding, encoding) {
			headers["Content-Encoding"] = encoding
		}

		body = echoValue
	} else if strings.HasPrefix(path, "/user-agent") {
		statusLine = NewStatusLine("HTTP/1.1", StatusOK)

		userAgent := request.Headers["User-Agent"]
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = strconv.Itoa(len(userAgent))

		body = userAgent
	} else if strings.HasPrefix(path, "/files/") {
		dir := os.Args[2]
		filename, _ := strings.CutPrefix(path, "/files/")
		filePath := filepath.Join(dir, filename)

		httpMethod := request.RequestLine.Method
		if httpMethod == "GET" {
			file, err := os.ReadFile(filePath)

			if err != nil {
				statusLine = NewStatusLine("HTTP/1.1", StatusNotFound)
			} else {
				statusLine = NewStatusLine("HTTP/1.1", StatusOK)
				headers["Content-Type"] = "application/octet-stream"
				headers["Content-Length"] = strconv.Itoa(len(file))

				body = string(file)
			}
		} else if httpMethod == "POST" {
			data := []byte(request.Body)
			err := os.WriteFile(filePath, data, 0777)

			if err != nil {
				statusLine = NewStatusLine("HTTP/1.1", StatusBadRequest)
				body = err.Error()
			} else {
				statusLine = NewStatusLine("HTTP/1.1", StatusCreated)
			}
		} else {
			statusLine = NewStatusLine("HTTP/1.1", StatusNotFound)
		}
	} else {
		statusLine = NewStatusLine("HTTP/1.1", StatusNotFound)
	}

	response := NewResponse(*statusLine, headers, body)
	_, err = conn.Write([]byte(response.String()))
	if err != nil {
		return fmt.Errorf("error writing to connection: %s", err.Error())
	}

	fmt.Println("Response sent: ", response)
	return nil
}
