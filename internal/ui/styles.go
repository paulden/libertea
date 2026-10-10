package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/paulden/libertea/internal/stratagem"
	"github.com/paulden/libertea/internal/terminal"
)

// The screen is a ship console: a frame with the title and the key hints
// embedded in its borders, the stats on one line, the stratagem as a strip of
// cells that fill up, and the times with a sparkline.
const (
	screenWidth = 66
	innerWidth  = screenWidth - 2
	margin      = 3
	// The text column starts after the icon and a gap, and keeps a margin on
	// the right. It fits the longest name and a strip of 9 arrows.
	iconGap        = 3
	textColumn     = margin + terminal.IconColumns + iconGap
	textWidth      = innerWidth - textColumn - margin
	penaltyBarSize = 24
	// The penalty line is the bar followed by the remaining time, "  1.4s".
	penaltyLineWidth = penaltyBarSize + 6
	// Differences with the best time are shown with two decimals: smaller
	// ones would be displayed as ▲0.00.
	shownDifference = 10 * time.Millisecond
	sparklineSize   = 10
)

// Colors are given explicitly for each profile: the automatic downgrade of hex
// colors to the 16 ANSI colors gives poor results.
var (
	frameColor    = color("#FFE710", "11")
	errorColor    = color("#E0434F", "9")
	dimErrorColor = color("#7D3A3C", "1")
	textColor     = color("#D8D4C4", "7")
	dimColor      = color("#8C8878", "8")
	cellColor     = color("#6B6655", "8")
	ruleColor     = color("#3A382F", "8")
	valueColor    = color("#FFE710", "11")
	fillColor     = color("#191813", "0")
	validColor    = color("#3F8F29", "10")
	sparkColor    = color("#B8A640", "3")
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

var sparkLevels = []rune("▁▂▃▄▅▆▇█")

type Styles interface {
	FormatStats(stats stats) string
	FormatStratagem(strat stratagem.Stratagem, completion int, isBlocked bool, remaining time.Duration) string
	FormatTimes(stats stats, isBlocked bool) string
	FormatNotice() string
	FormatScreen(body string, keysHint string, isBlocked bool) string
}

type styles struct {
	// Without colors, progress is shown with a cursor under the next arrow.
	hasColors bool
	// Image ids of the icons transmitted to the terminal, nil when the
	// terminal cannot display images.
	iconIDs map[string]int

	base       lipgloss.Style
	frame      lipgloss.Style
	errorFrame lipgloss.Style
	title      lipgloss.Style
	errorTitle lipgloss.Style
	text       lipgloss.Style
	dim        lipgloss.Style
	value      lipgloss.Style
	rule       lipgloss.Style
	cell       lipgloss.Style
	doneArrow  lipgloss.Style
	nextArrow  lipgloss.Style
	errorText  lipgloss.Style
	errorBold  lipgloss.Style
	dimError   lipgloss.Style
	valid      lipgloss.Style
	spark      lipgloss.Style
}

// NewStyles builds the styles with their own renderer, so that the color
// profile does not leak to other instances through the lipgloss default one.
func NewStyles(profile termenv.Profile, iconIDs map[string]int) Styles {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(profile)
	fg := func(c lipgloss.CompleteColor) lipgloss.Style { return r.NewStyle().Foreground(c) }

	return &styles{
		hasColors: profile != termenv.Ascii,
		iconIDs:   iconIDs,

		base:       r.NewStyle(),
		frame:      fg(frameColor),
		errorFrame: fg(errorColor),
		title:      fg(frameColor).Bold(true),
		errorTitle: fg(errorColor).Bold(true),
		text:       fg(textColor),
		dim:        fg(dimColor),
		value:      fg(valueColor).Bold(true),
		rule:       fg(ruleColor),
		cell:       fg(cellColor),
		doneArrow:  fg(fillColor).Background(frameColor).Bold(true),
		nextArrow:  fg(valueColor).Bold(true).Underline(true),
		errorText:  fg(errorColor),
		errorBold:  fg(errorColor).Bold(true),
		dimError:   fg(dimErrorColor),
		valid:      fg(validColor),
		spark:      fg(sparkColor),
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

func (s styles) FormatStats(stats stats) string {
	item := func(label string, value int) string {
		return s.dim.Render(label+" ") + s.value.Render(fmt.Sprint(value))
	}
	line := strings.Join([]string{
		item("SUCCESS", stats.successes),
		item("ERRORS", stats.errors),
		item("STREAK", stats.streak),
		item("BEST", stats.bestStreak),
	}, "    ")
	return indent(line) + "\n" + indent(s.rule.Render(strings.Repeat("─", innerWidth-2*margin)))
}

// FormatStratagem renders the name, the category and the code of the
// stratagem next to its icon, always on as many lines as the icon.
func (s styles) FormatStratagem(strat stratagem.Stratagem, completion int, isBlocked bool, remaining time.Duration) string {
	nameStyle, labelStyle := s.errorText, s.errorText
	if !isBlocked {
		category := categoryColors[strat.Category]
		nameStyle, labelStyle = s.base.Foreground(category).Bold(true), s.base.Foreground(category)
	}

	top, middle, bottom := s.codeStrip(strat.Code, completion, isBlocked)
	lines := []string{
		nameStyle.Render(strings.ToUpper(strat.Name)),
		labelStyle.Render(categoryLabel(strat)),
		top, middle, bottom,
		s.underStrip(completion, isBlocked, remaining),
	}
	text := strings.Join(lines, "\n")

	icon := s.icon(strat)
	if icon == "" {
		// Without icon, the text block is centered in the screen. Its lines
		// are padded to the same width first, so that the cursor stays under
		// its arrow: PlaceHorizontal centers each line on its own. The width
		// includes the penalty line, so that the block does not move when it
		// appears.
		block := s.base.Width(max(lipgloss.Width(text), penaltyLineWidth)).Render(text)
		return lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, block)
	}
	return indent(lipgloss.JoinHorizontal(lipgloss.Top, icon, strings.Repeat(" ", iconGap), s.base.Width(textWidth).Render(text)))
}

// codeStrip draws the code as connected cells: done arrows are filled, the
// next one is underlined, and every arrow turns red after a wrong input.
func (s styles) codeStrip(code []rune, completion int, isBlocked bool) (string, string, string) {
	border := s.cell
	if isBlocked {
		border = s.errorText
	}

	var top, middle, bottom strings.Builder
	top.WriteString(s.cell.Render("┌"))
	middle.WriteString(s.cell.Render("│"))
	bottom.WriteString(s.cell.Render("└"))
	for i, direction := range code {
		arrow := arrowSymbols[direction]
		switch {
		case isBlocked:
			arrow = s.errorBold.Render(" " + arrow + " ")
		case i < completion:
			arrow = s.doneArrow.Render(" " + arrow + " ")
		case i == completion:
			arrow = " " + s.nextArrow.Render(arrow) + " "
		default:
			arrow = s.dim.Render(" " + arrow + " ")
		}

		topJoint, bottomJoint := "┬", "┴"
		if i == len(code)-1 {
			topJoint, bottomJoint = "┐", "┘"
		}
		top.WriteString(border.Render("───") + s.cell.Render(topJoint))
		middle.WriteString(arrow + s.cell.Render("│"))
		bottom.WriteString(border.Render("───") + s.cell.Render(bottomJoint))
	}
	return top.String(), middle.String(), bottom.String()
}

// underStrip shows the penalty as a draining bar after a wrong input, and a
// cursor under the next arrow when colors cannot show the progress.
func (s styles) underStrip(completion int, isBlocked bool, remaining time.Duration) string {
	if isBlocked {
		full := min(penaltyBarSize, int(float64(penaltyBarSize)*remaining.Seconds()/penaltyDuration.Seconds()+0.5))
		return s.errorText.Render(strings.Repeat("█", full)) +
			s.dimError.Render(strings.Repeat("░", penaltyBarSize-full)) +
			s.errorText.Render(fmt.Sprintf("  %.1fs", remaining.Seconds()))
	}
	if !s.hasColors {
		return strings.Repeat(" ", 2+4*completion) + "▲"
	}
	return ""
}

func categoryLabel(strat stratagem.Stratagem) string {
	label := strings.ToUpper(strat.Category[:1]) + strat.Category[1:]
	if strat.Kind != "" {
		label += " · " + strat.Kind
	}
	return label
}

func (s styles) FormatTimes(stats stats, isBlocked bool) string {
	if isBlocked {
		return indent(s.errorBold.Render("WRONG INPUT") + s.dim.Render("  start over when the bar is empty"))
	}

	line := s.dim.Render("LAST ") + s.text.Render(formatDuration(stats.lastTime))
	switch {
	case stats.newBest:
		line += "  " + s.valid.Render("NEW BEST")
	case stats.lastTime-stats.bestTime >= shownDifference:
		line += "  " + s.dimError.Render(fmt.Sprintf("▲%.2f", (stats.lastTime-stats.bestTime).Seconds()))
	}
	line += s.dim.Render("    BEST ") + s.text.Render(formatDuration(stats.bestTime))
	if spark := sparkline(stats.recentTimes); spark != "" {
		line += "    " + s.spark.Render(spark)
	}
	return indent(line)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// sparkline draws the times between the fastest and the slowest one.
func sparkline(times []time.Duration) string {
	if len(times) < 2 {
		return ""
	}
	fastest, slowest := times[0], times[0]
	for _, t := range times {
		fastest, slowest = min(fastest, t), max(slowest, t)
	}

	var b strings.Builder
	for _, t := range times {
		level := len(sparkLevels) / 2
		if slowest > fastest {
			level = int(float64(len(sparkLevels)-1) * float64(t-fastest) / float64(slowest-fastest))
		}
		b.WriteRune(sparkLevels[level])
	}
	return b.String()
}

// FormatNotice suggests forcing a color mode when none was detected.
func (s styles) FormatNotice() string {
	if s.hasColors {
		return ""
	}
	return indent("No colors detected, try -color 256.")
}

// FormatScreen frames the body, with the title and the key hints embedded in
// the borders. The frame turns red after a wrong input.
func (s styles) FormatScreen(body string, keysHint string, isBlocked bool) string {
	frame, title := s.frame, s.title
	if isBlocked {
		frame, title = s.errorFrame, s.errorTitle
	}

	left, right := "LIBERTEA", "STRATAGEM DRILL"
	top := frame.Render("╭─ ") + title.Render(left) + frame.Render(" "+strings.Repeat("─", innerWidth-6-len(left)-len(right))+" ") +
		s.dim.Render(right) + frame.Render(" ─╮")

	hint := keysHint + " move  ·  esc quit"
	bottom := frame.Render("╰─ ") + s.dim.Render(hint) + frame.Render(" "+strings.Repeat("─", innerWidth-3-lipgloss.Width(hint))+"╯")

	lines := []string{top}
	for _, line := range strings.Split(body, "\n") {
		padding := max(0, innerWidth-lipgloss.Width(line))
		lines = append(lines, frame.Render("│")+line+strings.Repeat(" ", padding)+frame.Render("│"))
	}
	lines = append(lines, bottom)
	return strings.Join(lines, "\n")
}

func indent(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", margin) + line
	}
	return strings.Join(lines, "\n")
}
