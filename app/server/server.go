package server

import (
	"fmt"
	"io"
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
			connection := conn
			defer func() {
				fmt.Println("closing connection")
				err := connection.Close()
				if err != nil {
					fmt.Println("failed to close connection: ", err.Error())
				}
			}()

			for {
				close, err := s.handleConnection(connection)

				if err != nil {
					fmt.Println("connection error: ", err.Error())
				}

				if close {
					break
				}
			}
		}()
	}
}

func (s *Server) Close() error {
	fmt.Println("Closing TCP server")
	return s.listener.Close()
}

func (s *Server) handleConnection(conn net.Conn) (bool, error) {
	closeConnection := false

	buffer, err := readConnection(conn)
	if err != nil {
		if err == io.EOF {
			// return and close the connection and latest package is received
			return true, nil
		}
		return true, fmt.Errorf("error reading connection: %s", err.Error())
	}

	requestString := string(buffer)
	fmt.Println("Response read:\n", requestString)
	fmt.Println("——————————————")

	request := http.Request{}
	request.Parse(requestString)

	path := request.RequestLine.Target

	var statusLine *http.StatusLine
	headers := make(map[string]string)
	body := ""

	connection, ok := request.Headers["Connection"]
	if ok && connection == "close" {
		closeConnection = true
		headers["Connection"] = "close"
	}

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
		return closeConnection, fmt.Errorf("error writing to connection: %s", err.Error())
	}

	fmt.Println("Response sent:\n", responseString)
	fmt.Println("——————————————")
	return closeConnection, nil
}

func readConnection(r io.Reader) ([]byte, error) {
	// currently request size is limited to 4kb
	// to simplify reading in chunks
	buffer := make([]byte, 4096)

	n, err := r.Read(buffer)
	return buffer[:n], err
}
