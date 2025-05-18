package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {
	l, _ := zap.NewDevelopment()
	defer l.Sync() // flushes buffer, if any
	logger = l.Sugar()

	logger.Info("Starting..")

	port := flag.Int("port", 6667, "the port to listen on")
	flag.Parse()

	ln, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: *port,
	})

	if err != nil {
		logger.With("error", err).Fatal("couldn't create TCP listener")
	}
	defer ln.Close()

	// accept connection
	conn, err := ln.Accept()
	if err != nil {
		logger.With("error", err).Fatal("couldn't accept TCP connection")
	}

	defer conn.Close()
	dothestuff(conn)

	logger.Info("shutting down")
}

func dothestuff(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			logger.With("error", err).Info("couldn't read from socket")
			break
		}

		logger.With("message", message).Info("message received")

		_, err = fmt.Fprintf(conn, "you said: %v\n", message)
		if err != nil {
			logger.With("error", err).Info("couldn't write to socket")
			break
		}
	}
}
