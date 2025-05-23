package model

import (
	"bufio"
	"cli/server"
	"cli/style"
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	ctx         context.Context
	server      *server.Server
	cancelFn    context.CancelFunc
	execModel   *ExecModel
	logs        *viewport.Model
	spinner     *spinner.Model
	tabs        []Tab
	currentTab  Tab
	locked      bool
	isLoading   bool
	initialized bool // true after BubbleTea loaded and gave us window size
	width       int
	heigth      int
	childWidth  int
	childHeight int
}

// will be given to each Tab, pointers to root model's fields
type TabCfg struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *viewport.Model
	width     *int
	height    *int
}

func NewModel(s *server.Server) Model {
	// locked, loading := false, false
	ctx, fn := context.WithCancel(context.Background())
	m := Model{
		currentTab:  Exec,
		server:      s,
		ctx:         ctx,
		cancelFn:    fn,
		locked:      false,
		isLoading:   false,
		tabs:        []Tab{Exec, Hide, Upload, Download},
		initialized: false,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return m.listenAndAcceptCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case ConnectionUpdateMsg:
		m.server.ConnState = server.ConnectionStatus(msg)
		if m.server.ConnState == server.Disconnected {
			m.server.Sock = nil
			return m, m.listenAndAcceptCmd()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.heigth = msg.Height
		m.childWidth = int(style.ChildModelWidthRatio * float32(msg.Width))
		m.childHeight = msg.Height - style.HeaderHeight

		// Instanciate tab components with pointers to root's fields
		if !m.initialized {
			tabCfg := TabCfg{
				srv:       m.server,
				locked:    &m.locked,
				isLoading: &m.isLoading,
				logs:      m.logs,
				width:     &m.childWidth,
				height:    &m.childHeight,
			}

			execModel := NewExecModel(tabCfg)
			m.execModel = &execModel

			m.initialized = true
		}

	// Handle important key presses and dispatch to currently focused child
	case tea.KeyMsg:

		if !m.initialized {
			return m, nil
		}
		switch msg.String() {

		case "ctrl+c", "q":
			if m.server.Sock != nil {
				(*m.server.Sock).Close()
				m.server.Sock = nil
			}
			m.server.ConnState = server.Disconnected
			m.cancelFn()
			return m, tea.Quit

		case "shift+tab", "right":
			if m.currentTab > 0 {
				m.currentTab--
			}

		case "tab", "left":
			if m.currentTab != Download {
				m.currentTab++
			}

		default:
			switch m.currentTab {
			case Exec:
				x, cmd := m.execModel.Update(msg)
				if idk, ok := x.(ExecModel); ok {
					m.execModel = &idk
					cmds = append(cmds, cmd)
				}
				// case Hide:
				// case Upload:
				// case Download:
			}
		}
	}

	if m.server.ConnState == server.Connected {
		cmds = append(cmds, m.readSocketCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.initialized {
		return "Loading..."
	}

	tabw := m.childWidth / 4
	selectedTabStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, false, true).
		Width(tabw).
		Height(2).
		Padding(0, 2).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#22AA55"))

	normalTabStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Padding(1, 2, 0, 2).
		Inherit(selectedTabStyle)

	row := strings.Builder{}

	tabs := []string{}
	for _, choice := range m.tabs {
		if choice == m.currentTab {
			tabs = append(tabs, selectedTabStyle.Render(choice.String()))
		} else {
			tabs = append(tabs, normalTabStyle.Render(choice.String()))
		}
	}

	row.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, tabs...))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).UnsetBorderTop().
		Padding(2).
		Width(tabw * 4).
		Height(m.childHeight)

	if t, err := m.getCurrentChild(); err == nil {
		row.WriteString(boxStyle.Render(t.View()))
	} else {
		row.WriteString(boxStyle.Render(fmt.Sprintf("\n\n\n\t\t [[ TODO %s ]]\n\n\n", m.currentTab.String())))
	}

	// s += "\nPress q to quit.\n"
	return row.String()
}

func (m Model) getCurrentChild() (tea.Model, error) {
	if m.initialized {
		switch m.currentTab {
		case Exec:
			return m.execModel, nil
		case Hide:
		case Upload:
		case Download:
		default:
			return nil, fmt.Errorf("TODO")
		}
	}
	return nil, fmt.Errorf("no init")
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
