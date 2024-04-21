package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	action      string
	confirmText string
	textInput   textinput.Model
	quitting    bool
	err         error
}

type errMsg error

func NewConfirmModel(action string, confirmText string) *confirmModel {
	ti := textinput.New()
	ti.Focus()
	return &confirmModel{
		action:      action,
		confirmText: confirmText,
		textInput:   ti,
		err:         nil,
	}
}

func (m confirmModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD, tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		default:
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m confirmModel) View() string {
	return fmt.Sprintf(
		"Do you really want to %s %s?\n\n%s\n\n%s\n%s",
		m.action,
		m.confirmText,
		m.textInput.View(),
		"Enter the name of the object you want to delete to confirm",
		"(ctrl+c or esc to quit)",
	) + "\n"
}

func Confirm(action string, confirmText string) error {
	// if o.IsAutoYes {
	// 	return nil
	// }
	program := tea.NewProgram(NewConfirmModel(action, confirmText))
	m, err := program.Run()
	if err != nil {
		return err
	}
	if m, ok := m.(confirmModel); ok {
		if m.err != nil {
			return m.err
		}
		if m.textInput.Value() == m.confirmText {
			return nil
		}
		if m.quitting || strings.Trim(m.textInput.Value(), " ") == "" {
			os.Exit(0)
		}
		return fmt.Errorf("answer did not match: %s", confirmText)
	}
	return fmt.Errorf("creating confirmation model")
}
