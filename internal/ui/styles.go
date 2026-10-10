package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"

	"github.com/paulden/libertea/internal/stratagem"
	"github.com/paulden/libertea/internal/terminal"
)

const (
	stratagemWidth = 55
	iconGap        = 2
)

// Colors are given explicitly for each profile: the automatic downgrade of hex
// colors to the 16 ANSI colors gives poor results.
var (
	borderForeground = color("#222323", "0")
	borderBackground = color("#FFE710", "3")
	wrongColor       = color("#BF1029", "1")
	validColor       = color("#3F8F29", "2")
	headerColor      = color("#0092A6", "6")
)

// In-game permit colors of each stratagem category, with bright ANSI variants
// that stay readable on dark backgrounds.
var categoryColors = map[string]lipgloss.CompleteColor{
	"offensive": color("#DC6455", "9"),
	"supply":    color("#55B9D2", "12"),
	"defensive": color("#699655", "10"),
	"mission":   color("#C9B269", "11"),
}

func color(hex, ansi string) lipgloss.CompleteColor {
	return lipgloss.CompleteColor{TrueColor: hex, ANSI256: hex, ANSI: ansi}
}

// Plain arrows from the Arrows block render as a single cell in most fonts,
// unlike the Supplemental Arrows-C ones which caused rendering artifacts.
var arrowSymbols = map[rune]string{
	'u': "↑",
	'd': "↓",
	'r': "→",
	'l': "←",
}

type Styles interface {
	FormatScoreTable(stats stats) string
	FormatStratagem(strat stratagem.Stratagem, completion int, isBlocked bool, remaining time.Duration) string
	FormatTimes(stats stats) string
	FormatScreen(render string, layoutDescription string) string
}

type styles struct {
	// Without colors, progress is shown with a cursor under the next arrow.
	hasColors bool
	// Image ids of the icons transmitted to the terminal, nil when the
	// terminal cannot display images.
	iconIDs map[string]int

	screen      lipgloss.Style
	stratagem   lipgloss.Style
	wrongInput  lipgloss.Style
	validInput  lipgloss.Style
	header      lipgloss.Style
	cell        lipgloss.Style
	tableBorder lipgloss.Style
	times       lipgloss.Style
	category    lipgloss.Style
}

// NewStyles builds the styles with their own renderer, so that the color
// profile does not leak to other instances through the lipgloss default one.
func NewStyles(profile termenv.Profile, iconIDs map[string]int) Styles {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(profile)

	return &styles{
		hasColors: profile != termenv.Ascii,
		iconIDs:   iconIDs,

		screen: r.NewStyle().
			Padding(2).
			Width(64).
			Height(20).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderForeground).
			BorderBackground(borderBackground),
		stratagem: r.NewStyle().
			Width(stratagemWidth).
			Bold(true).
			Align(lipgloss.Center),
		wrongInput: r.NewStyle().
			Foreground(wrongColor).
			Blink(true),
		validInput: r.NewStyle().
			Foreground(validColor),
		header: r.NewStyle().
			Foreground(headerColor).
			Bold(true).
			Align(lipgloss.Center),
		cell: r.NewStyle().
			Padding(0, 1),
		tableBorder: r.NewStyle().
			Foreground(borderBackground),
		times: r.NewStyle().
			Width(55).
			Faint(true).
			Align(lipgloss.Center),
		category: r.NewStyle(),
	}
}

// icon returns the placeholder of a stratagem icon, or an empty string when
// it was not transmitted to the terminal.
func (s styles) icon(strat stratagem.Stratagem) string {
	if id := s.iconIDs[strat.Icon]; id != 0 {
		return terminal.IconPlaceholder(id)
	}
	return ""
}

func (s styles) FormatScoreTable(stats stats) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(s.tableBorder).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return s.header
			default:
				return s.cell
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

func (s styles) FormatStratagem(strat stratagem.Stratagem, completion int, isBlocked bool, remaining time.Duration) string {
	icon := s.icon(strat)
	style := s.stratagem
	if icon != "" {
		style = style.Width(stratagemWidth - terminal.IconColumns - iconGap)
	}

	name, label := strat.Name, categoryLabel(strat)
	if !isBlocked {
		categoryStyle := s.category.Foreground(categoryColors[strat.Category])
		name, label = categoryStyle.Bold(true).Render(name), categoryStyle.Render(label)
	}
	rendering := fmt.Sprintf("%s\n%s\n\n", name, label)

	for i, arrow := range strat.Code {
		if i < completion {
			rendering += s.validInput.Render(arrowSymbols[arrow])
		} else {
			rendering += arrowSymbols[arrow]
		}
		rendering += " "
	}

	if !s.hasColors && !isBlocked {
		rendering += "\n" + strings.Repeat("  ", completion) + "^" + strings.Repeat("  ", len(strat.Code)-completion-1) + " "
	} else {
		rendering += "\n"
	}

	if isBlocked {
		rendering += fmt.Sprintf("\nWrong input! Start over in %.1fs", remaining.Seconds())
		return withIcon(icon, s.wrongInput.Inherit(style).Render(rendering))
	}

	return withIcon(icon, style.Render(rendering+"\n "))
}

func withIcon(icon, text string) string {
	if icon != "" {
		text = lipgloss.JoinHorizontal(lipgloss.Center, icon, strings.Repeat(" ", iconGap), text)
	}
	return fmt.Sprintf("%s \n", text)
}

func categoryLabel(strat stratagem.Stratagem) string {
	label := strings.ToUpper(strat.Category)
	if strat.Kind != "" {
		label += " · " + strings.ToUpper(strat.Kind)
	}
	return label
}

func (s styles) FormatTimes(stats stats) string {
	return s.times.Render(fmt.Sprintf("Last: %s   Best: %s", formatDuration(stats.lastTime), formatDuration(stats.bestTime)))
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
	if !s.hasColors {
		footer += "\nNo colors detected, try -color 256."
	}

	render = header + "\n" + output + "\n" + footer

	return s.screen.Render(render)
}
