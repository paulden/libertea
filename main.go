package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	layoutName := flag.String("layout", envOrDefault(LAYOUT_ENV_VAR, DEFAULT_LAYOUT), fmt.Sprintf(
		"keyboard layout, one of: %s (can also be set with %s)",
		strings.Join(LayoutNames(), ", "), LAYOUT_ENV_VAR,
	))
	flag.Parse()

	layout, err := GetLayout(*layoutName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	styles := NewStyles()
	model := NewModel(styles, layout)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
