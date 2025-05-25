package model

import (
	"fmt"
	"strings"
	"time"

	"cli/server"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type HideModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *viewport.Model
	isHidden  *bool
}

type (
	HideStartCmd int
	HideDoneCmd  int
)

func NewHideModel(cfg TabCfg, isHidden *bool) *HideModel {
	h := HideModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		logs:      cfg.logs,
		isHidden:  isHidden,
	}

	return &h
}

func (h HideModel) Init() tea.Cmd {
	return nil
}

func (h HideModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "h" {
			cmds = append(cmds, tea.Sequence(h.startHideCmd, h.waitForHideResultCmd))
		}
	}
	return h, tea.Batch(cmds...)
}

func (h HideModel) View() string {
	if h.srv.ConnState != server.Connected {
		return "Status: Unknown\n"
	}
	s := strings.Builder{}

	status := "Not Hidden"
	toggle := "hide"
	if *h.isHidden {
		status = "Hidden"
		toggle = "reveal"
	}

	s.WriteString(fmt.Sprintf("Status: %s\n", status))
	s.WriteString(fmt.Sprintf("Press 'h' to %s.\n", toggle))

	return s.String()
}

func (h HideModel) startHideCmd() tea.Msg {
	if *h.isLoading || h.srv.ConnState != server.Connected {
		h.logs.SetContent("Hide: not connected\n")
		return nil
	}
	*h.isLoading = true

	conn := *h.srv.Sock
	_, err := fmt.Fprintf(conn, "2\n")
	if err != nil {
		return ConnectionUpdateMsg(server.Disconnected)
	}

	return HideStartCmd(1)
}

func (h *HideModel) waitForHideResultCmd() tea.Msg {
	ch := h.srv.Channel
	timer := time.NewTimer(10 * time.Second)
loop:
	for {
		select {
		case msg, ok := <-ch:
			if ok {
				h.logs.SetContent(msg)
				*h.isHidden = !*h.isHidden
			} else {
				h.logs.SetContent("hide not ok")
			}
			break loop
		case <-timer.C:
			h.logs.SetContent("hide timed out")
			break loop
		}
	}
	*h.isLoading = false
	return HideDoneCmd(1)
}
