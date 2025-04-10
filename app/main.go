package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

type Response struct {
	statusLine StatusLine
	headers    string
	body       string
}

func (r *Response) String() string {
	response := r.statusLine.String()
	response += r.headers
	if r.headers != "" {
		response += "\r\n"
	}
	response += "\r\n" + r.body
	return response
}

func NewResponse(statusLine StatusLine, headers string, body string) *Response {
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

	statusLine := NewStatusLine("HTTP/1.1", StatusOK)
	if path != "/" {
		statusLine = NewStatusLine("HTTP/1.1", StatusNotFound)
	}

	response := NewResponse(*statusLine, "", "")
	_, err = conn.Write([]byte(response.String()))
	if err != nil {
		return fmt.Errorf("error writing to connection: %s", err.Error())
	}

	fmt.Println("Response sent: ", response)
	return nil
}
