package server

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

type Server struct {
	port     int
	listener net.Listener
}

func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) Start() error {
	fmt.Printf("Starting TCP server at port %d\n", s.port)

	address := fmt.Sprintf("%s:%d", "0.0.0.0", s.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to bind to port %d", s.port)
	}

	s.listener = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %s", err.Error())
		}

		go func() {
			err = s.handleConnection(conn)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}
		}()
	}
}

func (s *Server) Close() error {
	fmt.Println("Closing TCP server")
	return s.listener.Close()
}

func (s *Server) handleConnection(conn net.Conn) error {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading connection: %s", err.Error())
	}
	fmt.Println("Response read:\n", string(buffer))
	fmt.Println("——————————————")

	request := http.Request{}
	request.Parse(string(buffer))
	path := request.RequestLine.Target

	var statusLine *http.StatusLine
	headers := make(map[string]string)
	body := ""

	if path == "/" {
		statusLine = http.NewStatusLine("HTTP/1.1", http.StatusOK)
	} else if strings.HasPrefix(path, "/echo/") {
		statusLine = http.NewStatusLine("HTTP/1.1", http.StatusOK)

		echoValue, _ := strings.CutPrefix(path, "/echo/")

		headers["Content-Type"] = "text/plain"

		rawEncoding := request.Headers["Accept-Encoding"]
		encodings := http.Encodings{}
		encodings.Parse(rawEncoding)
		encoding := encodings.GetEncoding()

		if encoding != "" {
			headers["Content-Encoding"] = encoding
		}

		body = echoValue
	} else if strings.HasPrefix(path, "/user-agent") {
		statusLine = http.NewStatusLine("HTTP/1.1", http.StatusOK)

		userAgent := request.Headers["User-Agent"]
		headers["Content-Type"] = "text/plain"

		body = userAgent
	} else if strings.HasPrefix(path, "/files/") {
		dir := os.Args[2]
		filename, _ := strings.CutPrefix(path, "/files/")
		filePath := filepath.Join(dir, filename)

		httpMethod := request.RequestLine.Method
		if httpMethod == "GET" {
			file, err := os.ReadFile(filePath)

			if err != nil {
				statusLine = http.NewStatusLine("HTTP/1.1", http.StatusNotFound)
			} else {
				statusLine = http.NewStatusLine("HTTP/1.1", http.StatusOK)
				headers["Content-Type"] = "application/octet-stream"

				body = string(file)
			}
		} else if httpMethod == "POST" {
			data := []byte(request.Body)
			err := os.WriteFile(filePath, data, 0777)

			if err != nil {
				statusLine = http.NewStatusLine("HTTP/1.1", http.StatusBadRequest)
				body = err.Error()
			} else {
				statusLine = http.NewStatusLine("HTTP/1.1", http.StatusCreated)
			}
		} else {
			statusLine = http.NewStatusLine("HTTP/1.1", http.StatusNotFound)
		}
	} else {
		statusLine = http.NewStatusLine("HTTP/1.1", http.StatusNotFound)
	}

	response := http.NewResponse(*statusLine, headers, body)
	responseString := response.String()
	_, err = conn.Write([]byte(responseString))
	if err != nil {
		return fmt.Errorf("error writing to connection: %s", err.Error())
	}

	fmt.Println("Response sent:\n", responseString)
	fmt.Println("——————————————")
	return nil
}
