package server

import "net"

type ConnectionStatus int

const (
	Connected ConnectionStatus = iota
	Disconnected
	Listening
)

func (c ConnectionStatus) String() string {
	switch c {
	case Connected:
		return "✅ CONNECTED"
	case Listening:
		return "📡 LISTENING"
	case Disconnected:
	default:
		return "❌ DISCONNECTED"
	}
	return "❌ DISCONNECTED"
}

type Server struct {
	Sock      *net.Conn
	Listener  *net.TCPListener
	Channel   chan string
	Ip        string
	Port      int
	ConnState ConnectionStatus
}
