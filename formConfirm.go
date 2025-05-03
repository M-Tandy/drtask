package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type FormConfirmModel struct {
	Active  bool
	focused int
	Width   int

	OnSubmit func(m *FormConfirmModel)
	OnCancel func(m *FormConfirmModel)
}

func (m *FormConfirmModel) Init() tea.Cmd {
	return nil
}

func (m FormConfirmModel) Update(msg tea.Msg) (FormConfirmModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case CloseFormMsg:
		m.CleanUp()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "y":
			m.OnSubmit(&m)
			return m, Close
		case "esc", "n":
			m.OnCancel(&m)
			return m, Close
		}
	}
	return m, cmd
}

func (m FormConfirmModel) View() string {
	form_string := "  CONFIRM OPERATION   \n y - YES | n - CANCEL "
	return form_string
}

func (m *FormConfirmModel) Activate(
	onSubmit func(m *FormConfirmModel),
	onCancel func(m *FormConfirmModel),
) {
	m.Active = true

	m.OnSubmit = onSubmit
	m.OnCancel = onCancel
}

func (m *FormConfirmModel) CleanUp() {
	m.Active = false
}
