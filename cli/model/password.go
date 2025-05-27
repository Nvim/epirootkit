package model

import (
	"fmt"
	"strings"
	"time"

	"cli/server"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type PasswordModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *LogsModel
	isDoing   *bool
	prompt    *textinput.Model
}

type (
	PasswordStartCmd int
	PasswordDoneCmd  int
)

func NewPasswordModel(cfg TabCfg) *PasswordModel {
	b := false
	ti := textinput.New()
	ti.Placeholder = "Input password"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = *cfg.width - 12
	ti.PromptStyle.Height(1)
	h := PasswordModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		isDoing:   &b,
		logs:      cfg.logs,
		prompt:    &ti,
	}

	return &h
}

func (m PasswordModel) Init() tea.Cmd {
	return nil
}

func (m PasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if !*m.isLoading && !*m.isDoing {
				cmds = append(cmds, tea.Sequence(m.startPasswordCmd, m.waitForPasswordResultCmd))
				m.prompt.Reset()
			} else {
				m.logs.Append("password: another command is already running\n")
			}
		} else {
			var cmd tea.Cmd
			*m.prompt, cmd = m.prompt.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m PasswordModel) View() string {
	if m.srv.ConnState != server.Connected {
		return "Not connected.\n"
	}
	s := strings.Builder{}

	s.WriteString("\n\n\t\t\tRootkit is locked. A password is required\n\n\n")
	s.WriteString(m.prompt.View())

	return s.String()
}

func (m PasswordModel) startPasswordCmd() tea.Msg {
	if *m.isLoading || m.srv.ConnState != server.Connected {
		m.logs.Append("Password: not connected\n")
		return nil
	}

	conn := *m.srv.Sock
	input := m.prompt.Value()
	_, err := fmt.Fprintf(conn, "5 %s\n", input)
	if err != nil {
		return ConnectionUpdateMsg(server.Disconnected)
	}

	*m.isDoing = true
	*m.isLoading = true
	return PasswordStartCmd(1)
}

func (m *PasswordModel) waitForPasswordResultCmd() tea.Msg {
	if !*m.isDoing {
		return nil
	}
	ch := m.srv.Channel
	timer := time.NewTimer(10 * time.Second)
loop:
	for {
		select {
		case msg, ok := <-ch:
			if ok {
				l := msg[0]
				switch l {
				case '0':
					*m.locked = false
					m.logs.Append("password: welcome, hacker😈\n")
				case '1':
					*m.locked = true
					m.logs.Append("password: incorrect password❌\n")
				default:
					*m.locked = true
					m.logs.Append("password: couldn't determine lock status\n")
				}
			} else {
				m.logs.Append("password not ok\n")
			}
			break loop
		case <-timer.C:
			m.logs.Append("password timed out\n")
			break loop
		}
	}
	*m.isLoading = false
	*m.isDoing = false
	return PasswordDoneCmd(1)
}
