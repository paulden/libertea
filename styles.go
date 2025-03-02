package main

import (
	"fmt"

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
)

var ARROWS_DISPLAY = map[rune]string{
	'u': "🢁",
	'd': "🢃",
	'r': "🢂",
	'l': "🢀",
}


type Styles interface {
	FormatScoreTable(successes, errors, streak int) string
	FormatStratagem(strategem stratagem, completion int, isBlocked bool) string
	FormatScreen(render string) string
}

type styles struct{}

func NewStyles() Styles {
	return &styles{}
}

func (s styles) FormatScoreTable(successes, errors, streak int) string {
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
		Headers("SUCCESSES", "ERRORS", "STREAK").
		Rows([]string{fmt.Sprintf("%d", successes), fmt.Sprintf("%d", errors), fmt.Sprintf("%d", streak)})

	return t.Render()
}

func (s styles) FormatStratagem(stratagem stratagem, completion int, isBlocked bool) string {
	rendering := fmt.Sprintf("%s\n\n", stratagem.name)

	for i, arrow := range stratagem.code {
		if i < completion && !isBlocked {
			rendering += validInput.Render(ARROWS_DISPLAY[arrow])
		} else {
			rendering += ARROWS_DISPLAY[arrow]
		}
		rendering += " "
	}

	if isBlocked {
		return fmt.Sprintf("%s \n", wrongInput.Inherit(strategemStyle).Render(rendering))
	}

	return fmt.Sprintf("%s \n", strategemStyle.Render(rendering))
}

func (s styles) FormatScreen(output string) string {
	var render string

	header := "Call for your next stratagem and save democracy!\n"
	footer := "Press q to quit."

	render = header + "\n" + output + "\n" + footer

	return globalStyle.Render(render)
}
