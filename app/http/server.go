package http

import (
	"fmt"
	"io"
	"net"
)

type Server struct {
	port     int
	listener net.Listener
	router   Router

	Protocol string
}

func NewServer(port int) *Server {
	router := NewRouter()
	return &Server{
		port:     port,
		router:   *router,
		Protocol: "HTTP/1.1",
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

func (s *Server) Route(method, pattern string, handler RouteHandler) {
	s.router.AddRoute(method, pattern, handler)
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

	responseWriter := NewResponseWriter(conn, s.Protocol)

	request := Request{}
	request.Parse(requestString)

	// Handle connection closage
	connection, ok := request.Headers["Connection"]
	if ok && connection == "close" {
		closeConnection = true
		responseWriter.Headers["Connection"] = "close"
	}

	// Handle encoding
	rawEncoding := request.Headers["Accept-Encoding"]

	encodings := Encodings{}
	encodings.Parse(rawEncoding)
	encoding := encodings.GetEncoding()

	if encoding != "" {
		responseWriter.Headers["Content-Encoding"] = encoding
	}

	// Match route
	handler, pattern := s.router.MatchRoute(request.RequestLine.Method, request.RequestLine.Target)
	if handler == nil {
		responseWriter.StatusCode = StatusNotFound
		_, err := responseWriter.Write([]byte{})

		if err != nil {
			return true, fmt.Errorf("error writing to connection: %s", err.Error())
		}

		return closeConnection, nil
	}
	request.pattern = &pattern

	// Run route handler
	err = handler(*responseWriter, &request)
	if err != nil {
		return true, fmt.Errorf("error writing to connection: %s", err.Error())
	}

	return closeConnection, nil
}

func readConnection(r io.Reader) ([]byte, error) {
	// currently request size is limited to 4kb
	// to simplify reading in chunks
	buffer := make([]byte, 4096)

	n, err := r.Read(buffer)
	return buffer[:n], err
}
