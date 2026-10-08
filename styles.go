package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

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

	timesStyle = lipgloss.NewStyle().
			Width(55).
			Faint(true).
			Align(lipgloss.Center)
)

// Plain arrows from the Arrows block render as a single cell in most fonts,
// unlike the Supplemental Arrows-C ones which caused rendering artifacts.
var ARROWS_DISPLAY = map[rune]string{
	'u': "↑",
	'd': "↓",
	'r': "→",
	'l': "←",
}

type Styles interface {
	FormatScoreTable(stats stats) string
	FormatStratagem(strategem stratagem, completion int, isBlocked bool, remaining time.Duration) string
	FormatTimes(stats stats) string
	FormatScreen(render string, layoutDescription string) string
}

type styles struct{}

func NewStyles() Styles {
	return &styles{}
}

func (s styles) FormatScoreTable(stats stats) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFE710"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			default:
				return cellStyle
			}
		}).
		Headers("SUCCESSES", "ERRORS", "STREAK", "BEST STREAK").
		Rows([]string{
			fmt.Sprintf("%d", stats.successes),
			fmt.Sprintf("%d", stats.errors),
			fmt.Sprintf("%d", stats.streak),
			fmt.Sprintf("%d", stats.bestStreak),
		})

	return t.Render()
}

func (s styles) FormatStratagem(stratagem stratagem, completion int, isBlocked bool, remaining time.Duration) string {
	rendering := fmt.Sprintf("%s\n\n", stratagem.name)

	for i, arrow := range stratagem.code {
		if i < completion {
			rendering += validInput.Render(ARROWS_DISPLAY[arrow])
		} else {
			rendering += ARROWS_DISPLAY[arrow]
		}
		rendering += " "
	}

	rendering += "\n"

	if isBlocked {
		rendering += fmt.Sprintf("\nWrong input! Start over in %.1fs", remaining.Seconds())
		return fmt.Sprintf("%s \n", wrongInput.Inherit(strategemStyle).Render(rendering))
	}

	return fmt.Sprintf("%s \n", strategemStyle.Render(rendering+"\n "))
}

func (s styles) FormatTimes(stats stats) string {
	return timesStyle.Render(fmt.Sprintf("Last: %s   Best: %s", formatDuration(stats.lastTime), formatDuration(stats.bestTime)))
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func (s styles) FormatScreen(output string, layoutDescription string) string {
	var render string

	header := "Call for your next stratagem and save democracy!\n"
	footer := fmt.Sprintf("Keys: %s. Press Esc to quit.", layoutDescription)

	render = header + "\n" + output + "\n" + footer

	return globalStyle.Render(render)
}
