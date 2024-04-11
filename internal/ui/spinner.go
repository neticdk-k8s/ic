package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
)

type spinnerModel struct {
	spinner  spinner.Model
	message  string
	quitting bool
	finished bool
}

func newSpinnerModel() spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{spinner: s}
}

func (m *spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
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

func (m *spinnerModel) View() string {
	str := fmt.Sprintf("%s %s", m.spinner.View(), m.message)
	if m.quitting {
		return ""
	}
	return str
}

type Spinner struct {
	model   *spinnerModel
	program *tea.Program
	writer  io.Writer
	logger  logger.Logger
	running bool
}

func NewSpinner(w io.Writer, logger logger.Logger) *Spinner {
	model := newSpinnerModel()
	return &Spinner{
		model:   &model,
		program: tea.NewProgram(&model),
		writer:  w,
		logger:  logger,
		running: false,
	}
}

func (s *Spinner) Run() {
	if !isInteractive {
		fmt.Fprintln(s.writer, s.model.View())
		return
	}

	if s.running {
		s.logger.Warn("spinner already running")
		return
	}

	s.running = true
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

func (s *Spinner) Stop() {
	s.model.finished = true
	if err := s.program.ReleaseTerminal(); err != nil {
		s.logger.Error("Failed to release terminal", "err", err)
	}
	s.program.Quit()
}

func (s *Spinner) Text(t string) {
	if !isInteractive {
		s.model.message = t
		fmt.Fprintln(s.writer, s.model.View())
		return
	}

	if !s.running {
		s.logger.Warn("spinner not running")
		return
	}

	s.model.message = t
}

func (s *Spinner) Running() bool {
	return s.running
}
