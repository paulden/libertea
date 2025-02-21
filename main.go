package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var KEYS_MAPPING = map[string]rune{
	"up":    'u',
	"down":  'd',
	"right": 'r',
	"left":  'l',
}

var ARROWS_DISPLAY = map[rune]string{
	'u': "🢁",
	'd': "🢃",
	'r': "🢂",
	'l': "🢀",
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
