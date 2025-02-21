package main

import "github.com/charmbracelet/lipgloss"

var (
	globalStyle = lipgloss.NewStyle().
			Padding(2).
			Width(64).
			Height(20).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#222323")).
			BorderBackground(lipgloss.Color("#FFE710"))

	strategemStyle = lipgloss.NewStyle().
			Width(55).
			Bold(true).
			Align(lipgloss.Center)

	wrongInput = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BF1029")).
			Blink(true)

	validInput = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3F8F29"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0092A6")).
			Bold(true).
			Align(lipgloss.Center)

	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)
)
