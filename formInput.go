package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FormInputModel struct {
	dump         io.Writer // For Debuggging
	Active       bool
	fieldNames   []string
	destinations []*string
	inputs       []textinput.Model
	focused      int
	Width        int

	OnSubmit func(m *FormInputModel) tea.Msg
	OnCancel func(m *FormInputModel)
}

func (m *FormInputModel) Init() tea.Cmd {
	return nil
}

func (m FormInputModel) Update(msg tea.Msg) (FormInputModel, tea.Cmd) {
	switch msg := msg.(type) {

	case CloseFormMsg:
		m.CleanUp()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.focused == len(m.inputs)-1 {
				return m, tea.Sequence(
					func () tea.Msg {return m.OnSubmit(&m)},
					Close,
				)
			}
			m.nextInput()
		case "esc":
			m.OnCancel(&m)
			return m, Close
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()
	}

	var cmds []tea.Cmd = make([]tea.Cmd, 3)
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m FormInputModel) View() string {
	form_string := ""
	for i := range m.inputs {
		form_string += fmt.Sprintf("%s: %s\n", m.fieldNames[i], m.inputs[i].View())
	}
	return form_string
}

func (
	m *FormInputModel) Activate(fieldNames []string,
	onSubmit func(m *FormInputModel) tea.Msg,
	onCancel func(m *FormInputModel),
) {
	l := len(fieldNames)
	m.inputs = make([]textinput.Model, l)
	for i := range l {
		m.inputs[i] = textinput.New()
		m.inputs[i].Prompt = ""
		m.inputs[i].CharLimit = 156
		m.inputs[i].Width = m.Width - len(fieldNames[i])
	}
	m.fieldNames = fieldNames

	m.inputs[0].Focus()
	m.focused = 0
	m.Active = true

	m.OnSubmit = onSubmit
	m.OnCancel = onCancel
}


type CloseFormMsg struct{}

func Close() tea.Msg {
	return CloseFormMsg{}
}

func (m *FormInputModel) CleanUp() {
	m.Active = false
	m.destinations = nil
	m.inputs = nil
	m.fieldNames = nil
	m.focused = 0
}

// nextInput focuses the next input field
func (m *FormInputModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *FormInputModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
