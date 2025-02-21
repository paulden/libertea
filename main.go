package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var KEYS_MAPPING = map[string]rune{
	"up":    'u',
	"down":  'd',
	"right": 'r',
	"left":  'l',
}

var ARROWS_DISPLAY = map[rune]string{
	'u': "^",
	'd': "v",
	'r': ">",
	'l': "<",
}

const (
	autocraticRed   = lipgloss.Color("#BF1029")
	democraticGreen = lipgloss.Color("#3F8F29")
)

var (
	wrongInput = lipgloss.NewStyle().Foreground(autocraticRed)
	validInput = lipgloss.NewStyle().Foreground(democraticGreen)
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
