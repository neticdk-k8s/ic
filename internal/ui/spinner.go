package ui

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neticdk-k8s/ic/internal/logger"
)

type model struct {
	spinner  spinner.Model
	message  string
	quitting bool
	finished bool
}

func newSpinner() *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return &model{
		spinner: s,
	}
}

func (m *model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting || m.finished {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case error:
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m *model) View() string {
	if m.quitting || m.finished {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

type Spinner struct {
	model   *model
	program *tea.Program
	writer  io.Writer
	logger  logger.Logger
	running bool
}

// NewSpinner creates a new Spinner
func NewSpinner(w io.Writer, logger logger.Logger) *Spinner {
	model := newSpinner()
	return &Spinner{
		model:   model,
		program: tea.NewProgram(model, tea.WithOutput(w)),
		writer:  w,
		logger:  logger,
		running: false,
	}
}

// Run starts the spinner
func (s *Spinner) Run(text string) {
	s.model.message = text

	if s.running {
		s.logger.Warn("spinner already running")
		return
	}

	s.running = true

	if !isInteractive {
		fmt.Fprintln(s.writer, s.model.View())
		return
	}

	s.model.quitting = false

	go func() {
		_, err := s.program.Run()
		if err != nil {
			s.logger.Error("spinner", "err", err)
			return
		}

		if s.model.quitting {
			s.logger.Warn("ctrl + c -> quitting")
			os.Exit(0)
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !isInteractive {
		return
	}

	s.model.finished = true

	// prevent race condition
	time.Sleep(50 * time.Millisecond)
	if err := s.program.ReleaseTerminal(); err != nil {
		s.logger.Error("Failed to release terminal", "err", err)
	}
}

// Text sets the text of the spinner
func (s *Spinner) Text(t string) {
	s.model.message = t
	if !isInteractive {
		fmt.Fprintln(s.writer, s.model.View())
		return
	}

	if !s.running {
		s.logger.Warn("spinner not running")
	}
}

// Running returns the running status of the spinner
func (s *Spinner) Running() bool {
	return s.running
}
