// Package ui is the game screen, built with bubbletea and lipgloss.
package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/paulden/libertea/internal/keys"
	"github.com/paulden/libertea/internal/stratagem"
)

const (
	penaltyDuration = 2 * time.Second
	penaltyTick     = 100 * time.Millisecond
)

type stats struct {
	successes  int
	errors     int
	streak     int
	bestStreak int
	lastTime   time.Duration
	bestTime   time.Duration
}

type Model struct {
	stratagems          []stratagem.Stratagem
	currentStratagem    stratagem.Stratagem
	stratagemCompletion int
	stratagemStart      time.Time
	stats               stats
	blockedTimer        timer.Model
	layout              keys.Layout
	styles              Styles
	// Terminal size, used to center the game. Zero until bubbletea reports it.
	width  int
	height int
	// now is time.Now, replaced in tests to control the measured times.
	now func() time.Time
}

func NewModel(styles Styles, layout keys.Layout, stratagems []stratagem.Stratagem) Model {
	return Model{
		stratagems:       stratagems,
		currentStratagem: stratagem.Random(stratagems, ""),
		layout:           layout,
		styles:           styles,
		now:              time.Now,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case timer.TickMsg, timer.StartStopMsg, timer.TimeoutMsg:
		var cmd tea.Cmd
		m.blockedTimer, cmd = m.blockedTimer.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}

		if m.blockedTimer.Running() {
			return m, nil
		}

		direction, ok := m.layout.Direction(msg.String())
		if !ok {
			// Keys outside of the layout are not counted as errors.
			return m, nil
		}

		if m.stratagemStart.IsZero() {
			m.stratagemStart = m.now()
		}

		if direction != m.currentStratagem.Code[m.stratagemCompletion] {
			m.stats.errors++
			m.stats.streak = 0
			m.stratagemCompletion = 0
			m.blockedTimer = timer.NewWithInterval(penaltyDuration, penaltyTick)
			return m, m.blockedTimer.Init()
		}

		m.stratagemCompletion++

		if m.stratagemCompletion == len(m.currentStratagem.Code) {
			m.completeStratagem()
		}
	}

	return m, nil
}

func (m *Model) completeStratagem() {
	elapsed := m.now().Sub(m.stratagemStart)

	m.stats.successes++
	m.stats.streak++
	m.stats.bestStreak = max(m.stats.bestStreak, m.stats.streak)
	m.stats.lastTime = elapsed
	if m.stats.bestTime == 0 || elapsed < m.stats.bestTime {
		m.stats.bestTime = elapsed
	}

	m.stratagemCompletion = 0
	m.stratagemStart = time.Time{}
	m.currentStratagem = stratagem.Random(m.stratagems, m.currentStratagem.Name)
}

func (m Model) View() string {
	var output string

	output += m.styles.FormatScoreTable(m.stats)
	output += "\n\n"
	output += m.styles.FormatStratagem(m.currentStratagem, m.stratagemCompletion, m.blockedTimer.Running(), m.blockedTimer.Timeout)
	output += "\n"
	output += m.styles.FormatTimes(m.stats)

	screen := m.styles.FormatScreen(output, m.layout.Description)

	// Place leaves the screen untouched when the terminal is smaller than it.
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, screen)
}
