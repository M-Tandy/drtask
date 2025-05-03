package main

import (
	"github.com/charmbracelet/lipgloss"
)


var (
	unselected_style = lipgloss.NewStyle().
				Bold(false).
				Foreground(lipgloss.Color("240")).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color("240"))

	selected_style = unselected_style.
			Bold(true).
			Foreground(lipgloss.Color("6")).
			BorderForeground(lipgloss.Color("6"))

	itemStyle         = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("240"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("6"))
)
