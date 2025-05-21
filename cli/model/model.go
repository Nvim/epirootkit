package model

import (
	"bufio"
	"cli/server"
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	Tab                 int
	ConnectionUpdateMsg server.ConnectionStatus
	NoopMsg             int
)

const (
	Exec Tab = iota
	Hide
	Upload
	Download
)

func (c Tab) String() string {
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

type Model struct {
	ctx        context.Context
	server     *server.Server
	cancelFn   context.CancelFunc
	tabs       []Tab
	cursor     Tab
	currentTab Tab
}

func NewModel(s *server.Server) Model {
	ctx, fn := context.WithCancel(context.Background())
	return Model{
		cursor:     Exec,
		currentTab: Exec,
		server:     s,
		ctx:        ctx,
		cancelFn:   fn,
		tabs:       []Tab{Exec, Hide, Upload, Download},
	}
}

func (m Model) Init() tea.Cmd {
	return m.listenAndAcceptCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ConnectionUpdateMsg:
		m.server.ConnState = server.ConnectionStatus(msg)
		if m.server.ConnState == server.Disconnected {
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
			m.server.ConnState = server.Disconnected
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
			m.currentTab = Tab(m.cursor)
		}
	}

	if m.server.ConnState == server.Connected {
		return m, m.readSocketCmd()
	}
	return m, nil
}

func (m Model) View() string {
	s := fmt.Sprintf("Status: %s\n\n", m.server.ConnState.String())
	s += "What do you want to do?\n"

	for _, choice := range m.tabs {
		cursor := " "
		if m.cursor == choice {
			cursor = ">"
		}

		checked := " "
		if choice == m.currentTab {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.String())
	}
	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func (m Model) listenAndAcceptCmd() tea.Cmd {
	s := m.server
	return func() tea.Msg {
		for {
			select {
			case <-m.ctx.Done():
				// logger.Info("stopping accepter routine")
				return NoopMsg(0)
			default:
				conn, err := (*s.Listener).Accept()
				if err != nil {
					if m.ctx.Err() != nil {
						// logger.Info("couldn't accept: context stopped")
						return NoopMsg(0)
					}
					// logger.With("error", err).Fatal("couldn't accept TCP connection")
				}
				s.ConnState = server.Connected
				s.Sock = &conn
				return ConnectionUpdateMsg(server.Connected)
			}
		}
	}
}

func (m Model) readSocketCmd() tea.Cmd {
	conn := m.server.Sock
	return func() tea.Msg {
		if m.server.ConnState == server.Disconnected {
			return NoopMsg(0)
		}

		_, err := bufio.NewReader(*conn).ReadString('\n')
		if err != nil {
			// logger.With("error", err).Info("couldn't read from socket")
			return ConnectionUpdateMsg(server.Disconnected)
		}

		// logger.With("message", message).Info("message received")

		// _, err = fmt.Fprintf(*conn, "you said: %v\n", message)
		// if err != nil {
		// logger.With("error", err).Info("couldn't write to socket")
		// return ConnectionUpdateMsg(server.Disconnected)
		// }
		return ConnectionUpdateMsg(server.Connected)
	}
}
