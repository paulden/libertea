package main

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	PENALTY_DURATION = 2 * time.Second
	PENALTY_TICK     = 100 * time.Millisecond
)

type stats struct {
	successes  int
	errors     int
	streak     int
	bestStreak int
	lastTime   time.Duration
	bestTime   time.Duration
}

type model struct {
	currentStratagem    stratagem
	stratagemCompletion int
	stratagemStart      time.Time
	stats               stats
	blockedTimer        timer.Model
	layout              keyLayout
	styles              Styles
}

func NewModel(styles Styles, layout keyLayout) model {
	return model{
		currentStratagem: GetRandomStratagem(""),
		layout:           layout,
		styles:           styles,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			m.stratagemStart = time.Now()
		}

		if direction != m.currentStratagem.code[m.stratagemCompletion] {
			m.stats.errors++
			m.stats.streak = 0
			m.stratagemCompletion = 0
			m.blockedTimer = timer.NewWithInterval(PENALTY_DURATION, PENALTY_TICK)
			return m, m.blockedTimer.Init()
		}

		m.stratagemCompletion++

		if m.stratagemCompletion == len(m.currentStratagem.code) {
			m.completeStratagem()
		}
	}

	return m, nil
}

func (m *model) completeStratagem() {
	elapsed := time.Since(m.stratagemStart)

	m.stats.successes++
	m.stats.streak++
	m.stats.bestStreak = max(m.stats.bestStreak, m.stats.streak)
	m.stats.lastTime = elapsed
	if m.stats.bestTime == 0 || elapsed < m.stats.bestTime {
		m.stats.bestTime = elapsed
	}

	m.stratagemCompletion = 0
	m.stratagemStart = time.Time{}
	m.currentStratagem = GetRandomStratagem(m.currentStratagem.name)
}

func (m model) View() string {
	var output string

	output += m.styles.FormatScoreTable(m.stats)
	output += "\n\n"
	output += m.styles.FormatStratagem(m.currentStratagem, m.stratagemCompletion, m.blockedTimer.Running(), m.blockedTimer.Timeout)
	output += "\n"
	output += m.styles.FormatTimes(m.stats)

	return m.styles.FormatScreen(output, m.layout.description)
}
