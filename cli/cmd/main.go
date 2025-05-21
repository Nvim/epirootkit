package main

import (
	"cli/model"
	"cli/server"
	"flag"
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {
	l, _ := zap.NewDevelopment()
	defer func() {
		_ = l.Sync() // flushes buffer, if any
	}()

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

	s := &server.Server{
		Sock:      nil,
		Ip:        "",
		Port:      *port,
		ConnState: server.Disconnected,
		Listener:  ln,
	}

	p := tea.NewProgram(model.NewModel(s))
	if _, err := p.Run(); err != nil {
		logger.With("error", err).Fatal("error running BubbleTea program")
	}

	logger.Info("shutting down")
}
