package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type spn struct {
	spinner   spinner.Model
	prefix    string
	suffix    string
	quitting  bool
	cancelled bool
	done      chan bool
}

func newSpinner(prefix, suffix string) *spn {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &spn{spinner: s, prefix: prefix, suffix: suffix}
}

func (m *spn) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *spn) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting, m.cancelled = true, true
			return m, tea.Quit
		}
		return m, nil
	case error:
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *spn) View() string {
	str := fmt.Sprintf("%s%s %s", m.prefix, m.spinner.View(), m.suffix)
	if m.quitting {
		return ""
	}
	return str
}

func (m *spn) Stop() {
	m.quitting = true
	if m.done != nil {
		<-m.done
	}
}

func (m *spn) Text(t string) {
	m.suffix = t
}

func (m *spn) Start() {
	if !isInteractive {
		fmt.Println(m.View())
		return
	}

	ch := make(chan bool)
	m.done = ch
	m.quitting = false
	m.cancelled = false
	go func() {
		defer close(ch)
		_, _ = tea.NewProgram(m).Run()
		if m.cancelled {
			os.Exit(130)
		}
	}()
}

func StoppedSpinner(text string) *spn {
	spinner := newSpinner("", text)
	return spinner
}

func Spinner(text string) *spn {
	spinner := StoppedSpinner(text)
	spinner.Start()
	return spinner
}
