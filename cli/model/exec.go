package model

import (
	"cli/server"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type ExecModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *viewport.Model

	//
	pager viewport.Model
	input textinput.Model
}

func NewExecModel(cfg TabCfg) ExecModel {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	e := ExecModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		logs:      cfg.logs,
		pager:     viewport.New(*cfg.width, *cfg.height),
		input:     ti,
	}

	return e
}

func (m ExecModel) Init() tea.Cmd {
	return nil
}

func (m ExecModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmd2 tea.Cmd
	m.pager, cmd = m.pager.Update(msg)
	m.input, cmd2 = m.input.Update(msg)

	return m, tea.Batch(cmd, cmd2)
}

func (m ExecModel) View() string {
	return fmt.Sprintf("\n%s\n\n%s\n", m.pager.View(), m.input.View())
}
