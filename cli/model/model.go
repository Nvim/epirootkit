package model

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"cli/server"
	"cli/style"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	Tab                 int
	ConnectionUpdateMsg server.ConnectionStatus
)

const (
	Exec Tab = iota
	HideLock
	Upload
	Download
)

var (
	OK_BYTES   = []byte{0x04, 0x03, 0x02, 0x01, 'O', 'K', '\n'}
	KO_BYTES   = []byte{0x04, 0x03, 0x02, 0x01, 'K', 'O', '\n'}
	DONE_BYTES = []byte{0x04, 0x03, 0x02, 0x01, 'F', 'I', 'N', 'I', 'S', 'H', 'E', 'D', '\n'}
)

func (c Tab) String() string {
	switch c {
	case Exec:
		return "Run commands"
	case HideLock:
		return "Hide/Lock"
	case Upload:
		return "Upload a file"
	case Download:
		return "Download a file"
	default:
		return ""
	}
}

type Model struct {
	ctx           context.Context
	isHidden      *bool
	server        *server.Server
	cancelFn      context.CancelFunc
	execModel     *ExecModel
	hideModel     *HideModel
	downloadModel *DownloadModel
	uploadModel   *UploadModel
	logs          *LogsModel
	pwdModel      *PasswordModel
	spinner       *spinner.Model
	picker        *filepicker.Model
	locked        *bool
	isLoading     *bool
	tabs          []Tab
	currentTab    Tab
	width         int
	height        int
	childWidth    int
	childHeight   int
	initialized   bool
}

// will be given to each Tab, pointers to root model's fields
type TabCfg struct {
	srv       *server.Server
	locked    *bool
	isLoading *bool
	logs      *LogsModel
	width     *int
	height    *int
}

func NewModel(s *server.Server) Model {
	// locked, loading := false, false
	ctx, fn := context.WithCancel(context.Background())
	lock, load, hidden := false, false, false
	sp := spinner.New(spinner.WithSpinner(spinner.Points))
	p := filepicker.New()
	p.CurrentDirectory, _ = os.UserHomeDir()
	p.Styles.Cursor = p.Styles.Cursor.Foreground(style.LightGreen)
	p.Styles.Selected = p.Styles.Cursor.Foreground(style.LightGreen)
	p.Styles.File = p.Styles.File.Foreground(style.DarkGreen)
	p.Styles.Directory = p.Styles.Directory.Foreground(style.Blue)
	p.SetHeight(8)
	m := Model{
		currentTab:  Exec,
		server:      s,
		ctx:         ctx,
		cancelFn:    fn,
		locked:      &lock,
		isLoading:   &load,
		isHidden:    &hidden,
		tabs:        []Tab{Exec, HideLock, Upload, Download},
		initialized: false,
		spinner:     &sp,
		picker:      &p,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Sequence(m.spinner.Tick, m.picker.Init(), m.listenAndAcceptCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case ConnectionUpdateMsg:
		m.server.ConnState = server.ConnectionStatus(msg)
		if m.server.ConnState == server.Disconnected {
			m.server.Sock = nil
			return m, m.listenAndAcceptCmd()
		} else if m.server.ConnState == server.Connected {
			return m, tea.Sequence(m.setHiddenLocked, m.readSocketCmd())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.childWidth = int(style.ChildModelWidthRatio * float32(msg.Width))
		m.childHeight = msg.Height - 8

		// Instanciate tab components with pointers to root's fields
		if !m.initialized {
			// vp := viewport.New(m.width-(m.childWidth)-10, int(float32(m.height)*0.4))
			l := NewLogsModel(m.width-(m.childWidth)-10, int(float32(m.height)*0.4), 1024)
			m.logs = l

			tabCfg := TabCfg{
				srv:       m.server,
				locked:    m.locked,
				isLoading: m.isLoading,
				logs:      m.logs,
				width:     &m.childWidth,
				height:    &m.childHeight,
			}

			m.execModel = NewExecModel(tabCfg)
			m.hideModel = NewHideModel(tabCfg, m.isHidden)
			m.pwdModel = NewPasswordModel(tabCfg)
			m.downloadModel = NewDownloadModel(tabCfg)
			m.uploadModel = NewUploadModel(tabCfg, m.picker)

			m.initialized = true
		} else {
			m.logs.pager.Width = m.width - (m.childWidth) - 10
			m.logs.pager.Height = int(float32(m.height) * 0.4)
			msg.Width = m.childWidth
			msg.Height = m.childHeight
			c := m.dispatchToCurrentChild(msg)
			cmds = append(cmds, c)
		}

	// Handle important key presses and dispatch to currently focused child
	case tea.KeyMsg:

		if !m.initialized {
			return m, nil
		}
		switch msg.String() {

		case "ctrl+c", "q":
			if m.server.Sock != nil {
				(*m.server.Sock).Close()
				m.server.Sock = nil
			}
			close(m.server.Channel)
			m.server.ConnState = server.Disconnected
			m.cancelFn()
			return m, tea.Quit

		case "shift+tab", "right":
			if m.currentTab > 0 {
				m.currentTab--
			}

		case "tab", "left":
			if m.currentTab != Download {
				m.currentTab++
			}

		default:
			if c := m.dispatchToCurrentChild(msg); c != nil {
				cmds = append(cmds, c)
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		*m.spinner, cmd = m.spinner.Update(msg)
		// cmds = append(cmds, cmd)
		return m, cmd
	}

	if m.initialized {
		cmds = append(cmds, m.dispatchToLogs(msg))
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.initialized {
		return "Loading..."
	}

	tabw := m.childWidth / 4
	boxStyle := style.MainBoxStyle(tabw, m.childHeight)
	right := strings.Builder{}

	tabs := []string{}
	for _, choice := range m.tabs {
		if choice == m.currentTab {
			tabs = append(tabs, style.SelectedTab(tabw, int(choice)).Render(choice.String()))
		} else {
			tabs = append(tabs, style.NormalTab(tabw, int(choice)).Render(choice.String()))
		}
	}

	right.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, tabs...))
	right.WriteString("\n")

	if m.server.ConnState != server.Connected {
		right.WriteString(boxStyle.Render("\n\t\t\tNot connected.\n\n"))
	} else {
		if t, err := m.getCurrentChild(); err == nil {
			right.WriteString(boxStyle.Render(t.View()))
		} else {
			right.WriteString(boxStyle.Render(fmt.Sprintf("\n\n\n\t\t [[ TODO %s ]]\n\n\n", m.currentTab.String())))
		}
	}

	appStyle := lipgloss.NewStyle()
	screen := strings.Builder{}

	leftStyle := style.LeftBoxStyle(m.width-lipgloss.Width(right.String())-2, m.childHeight)
	left := strings.Builder{}

	load := ""
	if *m.isLoading {
		load = fmt.Sprintf("%s  Loading...", m.spinner.View())
	}

	listen := ""
	if m.server.ConnState == server.Listening {
		listen = m.spinner.View()
	}

	left.WriteString(leftStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Bold(true).Render("Status:", m.server.ConnState.String(), listen),
			"\n",
			lipgloss.NewStyle().Italic(true).Foreground(style.Purple).Render(load),
			"\n",
			topBorder("Logs", m.logs.pager.Width),
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), false, true, true, true).
				Render(m.logs.View()),
		)))

	screen.WriteString(appStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Bottom,
			right.String(),
			left.String(),
		)))
	return screen.String()
}

func topBorder(title string, width int) string {
	char := "─"
	return fmt.Sprintf("┌%s┤ %s ├%s┐",
		strings.Repeat(char, (width/2)-len(title)),
		title,
		strings.Repeat(char, (width/2)-len(title)),
	)
}

func (m *Model) dispatchToCurrentChild(msg tea.Msg) tea.Cmd {
	// var cmds []tea.Cmd
	if *m.locked {
		x, cmd := m.pwdModel.Update(msg)
		if idk, ok := x.(PasswordModel); ok {
			m.pwdModel = &idk
		}
		return cmd
	}
	switch m.currentTab {
	case Exec:
		x, cmd := m.execModel.Update(msg)
		if idk, ok := x.(ExecModel); ok {
			m.execModel = &idk
		}
		return cmd
	case HideLock:
		x, cmd := m.hideModel.Update(msg)
		if idk, ok := x.(HideModel); ok {
			m.hideModel = &idk
		}
		return cmd
	case Download:
		x, cmd := m.downloadModel.Update(msg)
		if idk, ok := x.(DownloadModel); ok {
			m.downloadModel = &idk
		}
		return cmd
	case Upload:
		x, cmd := m.uploadModel.Update(msg)
		if idk, ok := x.(UploadModel); ok {
			m.uploadModel = &idk
		}
		return cmd
	}
	return nil
}

func (m *Model) dispatchToLogs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	x, cmd := m.logs.Update(msg)
	if idk, ok := x.(LogsModel); ok {
		m.logs = &idk
	}
	cmds = append(cmds, cmd)

	// picker's Init method triggers a `readDirMsg` that needs to be handled,
	// even though cli is locked or another tab is focused
	_, isKey := msg.(tea.KeyMsg)
	if !isKey {
		*m.picker, cmd = m.picker.Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m Model) getCurrentChild() (tea.Model, error) {
	if m.initialized {
		if *m.locked {
			return m.pwdModel, nil
		}
		switch m.currentTab {
		case Exec:
			return m.execModel, nil
		case HideLock:
			return m.hideModel, nil
		case Upload:
			return m.uploadModel, nil
		case Download:
			return m.downloadModel, nil
		default:
			return nil, fmt.Errorf("TODO")
		}
	}
	return nil, fmt.Errorf("no init")
}

func (m Model) listenAndAcceptCmd() tea.Cmd {
	s := m.server
	if s.ConnState != server.Disconnected {
		return nil
	}
	s.ConnState = server.Listening
	return func() tea.Msg {
		for {
			select {
			case <-m.ctx.Done():
				// logger.Info("stopping accepter routine")
				return nil
			default:
				conn, err := (*s.Listener).Accept()
				if err != nil {
					if m.ctx.Err() != nil {
						// logger.Info("couldn't accept: context stopped")
						return nil
					}
					// logger.With("error", err).Fatal("couldn't accept TCP connection")
				}
				s.ConnState = server.Connected
				s.Sock = &conn
				return ConnectionUpdateMsg(server.Connected)
			}
		}
	}
}

// Read socket every frame, and push bytes to channel.
// When running a command, the handler will poll channel for result
// until it's done
func (m Model) readSocketCmd() tea.Cmd {
	conn := m.server.Sock
	return func() tea.Msg {
		if !m.initialized || conn == nil || m.server.ConnState != server.Connected {
			return nil
		}

		r := bufio.NewReader(*conn)
		for {
			msg, err := r.ReadString('\n')
			if err != nil {
				return ConnectionUpdateMsg(server.Disconnected)
			}

			select {
			case <-m.ctx.Done():
				return nil
			default:
				m.server.Channel <- msg
			}
		}
	}
}

// Run only when we get a connection:
func (m *Model) setHiddenLocked() tea.Msg {
	msg, err := bufio.NewReader(*m.server.Sock).ReadString('\n')
	if err != nil {
		if m.logs != nil {
			m.logs.Append("couldn't read hidden status\n")
		}
		return ConnectionUpdateMsg(server.Disconnected)
	}
	switch msg[0] {
	case '1':
		*m.isHidden = true
	case '0':
		*m.isHidden = false
	default:
		if m.logs != nil {
			m.logs.Append(fmt.Sprintf("couldn't parse hidden status: %s\n", msg))
		}
	}
	switch msg[1] {
	case LOCKED:
		*m.locked = true
	case UNLOCKED:
		*m.locked = false
	default:
		if m.logs != nil {
			m.logs.Append(fmt.Sprintf("couldn't parse locked status: %s\n", msg))
			*m.locked = false
		}
	}
	return nil
}
