package model

import (
	"container/list"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Wraps a viewport with a list of strings to display
// List is used as a FIFO which can grow to a configurable size
type LogsModel struct {
	pager   *viewport.Model
	Logs    *list.List
	maxSize int
}

type LogAddedMsg string

func NewLogsModel(width, height, maxSize int) *LogsModel {
	vp := viewport.New(width, height)
	m := LogsModel{
		pager:   &vp,
		Logs:    list.New(),
		maxSize: maxSize,
	}

	return &m
}

func (m LogsModel) Init() tea.Cmd {
	return nil
}

func (m LogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var c tea.Cmd
	*m.pager, c = m.pager.Update(msg)
	if m.Logs.Len() > m.maxSize {
		m.Logs.Remove(m.Logs.Front())
	}
	return m, c
}

func (m LogsModel) View() string {
	l := m.Logs
	content := strings.Builder{}
	for e := l.Front(); e != nil; e = e.Next() {
		s, ok := e.Value.(string)
		if ok {
			content.WriteString(s)
		}
	}
	m.pager.SetContent(content.String())
	return m.pager.View()
}

func (m *LogsModel) Append(s string) {
	m.Logs.PushBack(s)
}
