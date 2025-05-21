package server

import "net"

type ConnectionStatus int

const (
	Connected ConnectionStatus = iota
	Disconnected
)

func (c ConnectionStatus) String() string {
	if c == Connected {
		return "CONNECTED"
	}
	return "DISCONNECTED"
}

type Server struct {
	Sock      *net.Conn
	Listener  *net.TCPListener
	Ip        string
	Port      int
	ConnState ConnectionStatus
}
