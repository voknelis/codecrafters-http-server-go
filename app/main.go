package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codecrafters-io/http-server-starter-go/app/server"
)

func main() {
	s := server.NewServer(4221)

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
		s.Close()
		isClosed = true
	}()

	err := s.Start()
	if !isClosed && err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
