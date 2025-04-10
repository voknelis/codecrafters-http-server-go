package main

import (
	"errors"
	"fmt"
	"net"
	"os"
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

const (
	StatusOK       = 200
	StatusNotFound = 404
)

func StatusText(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusNotFound:
		return "Not Found"
	default:
		return ""
	}
}

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

	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("error accepting connection: %s", err.Error())
	}

	err = handleConnection(conn)
	if err != nil {
		return err
	}

	return nil
}

func handleConnection(conn net.Conn) error {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading connection: %s", err.Error())
	}
	fmt.Println("Response read: ", string(buffer))

	requestParts := strings.Split(string(buffer), "\r\n")
	if len(requestParts) == 0 {
		return errors.New("failed to parse http request")
	}

	requestStatusLine := strings.Split(requestParts[0], " ")
	path := requestStatusLine[1]

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

		body = echoValue
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
