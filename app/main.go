package main

import (
	"fmt"
	"net"
	"os"
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

func NewStatusLine(httpVersion string, statusCode int, statusCodePhrase string) *StatusLine {
	return &StatusLine{
		httpVersion:      httpVersion,
		statusCode:       statusCode,
		statusCodePhrase: statusCodePhrase,
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
	fmt.Println("Starting TCP server at port 4221")
	defer fmt.Println("Closing TCP server")

	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	handleConnection(conn)

	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	statusLine := NewStatusLine("HTTP/1.1", 200, "OK")
	response := NewResponse(*statusLine, "", "")
	_, err := conn.Write([]byte(response.String()))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Response sent: ", response)
}
