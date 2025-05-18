package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

type (
	ConnectionStatus int
	Choice           int
)

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

const (
	Exec Choice = iota
	Hide
	Upload
	Download
	None
)

func (c Choice) String() string {
	switch c {
	case Exec:
		return "Run commands"
	case Hide:
		return "Hide"
	case Upload:
		return "Upload a file"
	case Download:
		return "Download a file"
	default:
		return ""
	}
}

type Server struct {
	Sock      *net.Conn
	Listener  *net.TCPListener
	Ip        string
	Port      int
	ConnState ConnectionStatus
}

type model struct {
	ctx      context.Context
	server   *Server
	cancelFn context.CancelFunc
	choices  []Choice
	cursor   Choice
	selected Choice
}

func mainMenu(s *Server) model {
	ctx, fn := context.WithCancel(context.Background())
	return model{
		choices:  []Choice{Exec, Hide, Upload, Download},
		cursor:   Exec,
		selected: None,
		server:   s,
		ctx:      ctx,
		cancelFn: fn,
	}
}

type (
	ConnectionUpdateMsg ConnectionStatus
	NoopMsg             int
)

func (m model) listenAndAcceptCmd() tea.Cmd {
	s := m.server
	return func() tea.Msg {
		for {
			select {
			case <-m.ctx.Done():
				logger.Info("stopping accepter routine")
				return NoopMsg(0)
			default:
				conn, err := (*s.Listener).Accept()
				if err != nil {
					if m.ctx.Err() != nil {
						logger.Info("couldn't accept: context stopped")
						return NoopMsg(0)
					}
					logger.With("error", err).Fatal("couldn't accept TCP connection")
				}
				s.ConnState = Connected
				s.Sock = &conn
				return ConnectionUpdateMsg(Connected)
			}
		}
	}
}

func (m model) readSocketCmd() tea.Cmd {
	conn := m.server.Sock
	return func() tea.Msg {
		if m.server.ConnState == Disconnected {
			return NoopMsg(0)
		}
		message, err := bufio.NewReader(*conn).ReadString('\n')
		if err != nil {
			// logger.With("error", err).Info("couldn't read from socket")
			return ConnectionUpdateMsg(Disconnected)
		}

		// logger.With("message", message).Info("message received")

		_, err = fmt.Fprintf(*conn, "you said: %v\n", message)
		if err != nil {
			// logger.With("error", err).Info("couldn't write to socket")
			return ConnectionUpdateMsg(Disconnected)
		}
		return ConnectionUpdateMsg(Connected)
	}
}

func (m model) Init() tea.Cmd {
	return m.listenAndAcceptCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ConnectionUpdateMsg:
		m.server.ConnState = ConnectionStatus(msg)
		if m.server.ConnState == Disconnected {
			m.server.Sock = nil
			return m, m.listenAndAcceptCmd()
		}

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			if m.server.Sock != nil {
				(*m.server.Sock).Close()
				m.server.Sock = nil
			}
			m.server.ConnState = Disconnected
			m.cancelFn()
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor != Download {
				m.cursor++
			}

		case "enter":
			m.selected = Choice(m.cursor)
		}
	}

	if m.server.ConnState == Connected {
		return m, m.readSocketCmd()
	}
	return m, nil
}

func (m model) View() string {
	// The header
	s := fmt.Sprintf("Status: %s\n\n", m.server.ConnState.String())
	s += "What do you want to do?\n"

	for _, choice := range m.choices {
		cursor := " "
		if m.cursor == choice {
			cursor = ">"
		}

		checked := " "
		if choice == m.selected {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.String())
	}
	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

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

	s := &Server{
		Sock:      nil,
		Ip:        "",
		Port:      *port,
		ConnState: Disconnected,
		Listener:  ln,
	}

	p := tea.NewProgram(mainMenu(s))
	if _, err := p.Run(); err != nil {
		logger.With("error", err).Fatal("error running BubbleTea program")
	}

	logger.Info("shutting down")
}
