package model

import (
	"fmt"
	"strings"
	"time"

	"cli/server"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ExecModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *LogsModel

	//
	isExecing *bool
	pager     *viewport.Model
	input     *textinput.Model
	width     int
	height    int
}

func NewExecModel(cfg TabCfg) *ExecModel {
	ti := textinput.New()
	ti.Placeholder = "Run something"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = *cfg.width - 12
	ti.PromptStyle.Height(1)

	vp := viewport.New(*cfg.width-8, *cfg.height-12)
	b := false
	e := ExecModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		isExecing: &b,
		logs:      cfg.logs,
		pager:     &vp,
		input:     &ti,
		width:     *cfg.width - 4,
		height:    *cfg.height - 4,
	}

	return &e
}

type (
	ExecDoneMsg    int
	ExecStartedMsg int
)

func (m ExecModel) Init() tea.Cmd {
	return nil
}

func (m ExecModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	*m.pager, cmd = m.pager.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height - 4
		m.width = msg.Width - 4
		m.pager.Width = msg.Width - 8
		m.pager.Height = msg.Height - 12
		m.input.Width = msg.Width - 12
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if !*m.isLoading && !*m.isExecing {
				cmds = append(cmds, tea.Sequence(m.startExecCmd(m.input.Value()), m.waitForExecResultCmd))
				m.input.Reset()
			} else {
				m.logs.Append("exec: another command is already running\n")
			}
		}
	}
	*m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ExecModel) View() string {
	row := strings.Builder{}
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder())

	a := lipgloss.NewStyle().Width(m.width).Height(m.height)

	row.WriteString(
		a.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				style.Render(m.pager.View()),
				style.Render(m.input.View()),
			)),
	)

	return row.String()
}

func (m *ExecModel) startExecCmd(command string) tea.Cmd {
	return func() tea.Msg {
		if *m.isLoading || m.srv.ConnState != server.Connected {
			m.logs.Append("exec: not execing\n")
			return nil
		}
		m.logs.Append(fmt.Sprintf("exec: running %s...\n", command))

		conn := *m.srv.Sock
		_, err := fmt.Fprintf(conn, "0 %s\n", command)
		if err != nil {
			m.logs.Append("exec: failed to send command\n")
			return ConnectionUpdateMsg(server.Disconnected)
		}

		*m.isLoading = true
		*m.isExecing = true
		return ExecStartedMsg(1)
	}
}

// Scheduled after we wrote to socket that we want to exec
// TODO: error handling, append to logs
func (m ExecModel) waitForExecResultCmd() tea.Msg {
	if !*m.isExecing {
		return nil
	}
	var content string
	ch := m.srv.Channel
	timer := time.NewTimer(15 * time.Second)
	started := false
loop:
	for {
		select {
		case msg, ok := <-ch:
			if ok {
				switch msg {
				case string(DONE_BYTES):
					if started {
						m.pager.SetContent(content)
					} else {
						m.logs.Append("exec: no output\n")
					}
					break loop
				case string(OK_BYTES):
					m.logs.Append("exec: started\n")
					started = true
					continue
				default:
					content += msg
				}
			} else {
				m.logs.Append("exec: channel closed\n")
				break loop
			}
		case <-timer.C:
			m.logs.Append("exec: timed out\n")
			break loop
		}
	}

	*m.isExecing = false
	*m.isLoading = false
	return ExecDoneMsg(1)
}
