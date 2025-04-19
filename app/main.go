package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codecrafters-io/http-server-starter-go/app/handlers"
	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

func main() {
	server := http.NewServer(4221)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	isClosed := false
	go func() {
		<-signalChannel
		isClosed = true
		server.Close()
	}()

	server.Route("GET", "/", handlers.WildcardHandler)
	server.Route("GET", "/echo/{value}", handlers.EchoHandler)
	server.Route("GET", "/user-agent", handlers.UserAgentHandler)
	server.Route("GET", "/files/{filename}", handlers.GetFileHandler)
	server.Route("POST", "/files/{filename}", handlers.CreateFileHandler)

	err := server.Start()
	if !isClosed && err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
