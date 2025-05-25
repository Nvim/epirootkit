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
	logs      *viewport.Model

	//
	pager *viewport.Model
	input *textinput.Model
}

func NewExecModel(cfg TabCfg) *ExecModel {
	ti := textinput.New()
	ti.Placeholder = "Run something"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = *cfg.width - 12
	ti.PromptStyle.Height(1)

	vp := viewport.New(*cfg.width-8, *cfg.height-10)
	e := ExecModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		logs:      cfg.logs,
		pager:     &vp,
		input:     &ti,
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
	case tea.KeyMsg:
		if msg.String() == "enter" {
			cmds = append(cmds, tea.Sequence(m.startExecCmd(m.input.Value()), m.waitForExecResultCmd))
			m.input.Reset()
		}
	}
	*m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ExecModel) View() string {
	row := strings.Builder{}
	style := lipgloss.NewStyle().
		// Margin(1, 2, 1, 2).
		Border(lipgloss.RoundedBorder())

	row.WriteString(
		lipgloss.JoinVertical(
			lipgloss.Center,
			style.Render(m.pager.View()),
			style.Render(m.input.View()),
		))

	return row.String()
}

func (m *ExecModel) startExecCmd(command string) tea.Cmd {
	return func() tea.Msg {
		if *m.isLoading || m.srv.ConnState != server.Connected {
			fmt.Println("NOT EXECING")
			return nil
		}
		*m.isLoading = true
		m.pager.SetContent("doing exec")

		conn := *m.srv.Sock
		_, err := fmt.Fprintf(conn, "0 %s\n", command)
		if err != nil {
			return ConnectionUpdateMsg(server.Disconnected)
		}

		return ExecStartedMsg(1)
	}
}

// Scheduled after we wrote to socket that we want to exec
// TODO: error handling, append to logs
func (m ExecModel) waitForExecResultCmd() tea.Msg {
	var content string
	ch := m.srv.Channel
	timer := time.NewTimer(15 * time.Second)
loop:
	for {
		select {
		case msg, ok := <-ch:
			if ok {
				if msg == "DONE" || msg == "DONE\n" {
					m.pager.SetContent(content)
					break loop
				}
				content += msg
			} else {
				fmt.Println("not ok")
				m.pager.SetContent("not ok, sadge")
				break loop
			}
		case <-timer.C:
			m.pager.SetContent("timed out sadge")
			break loop
		}
	}
	*m.isLoading = false
	return ExecDoneMsg(1)
}
