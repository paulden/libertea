package main

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

var KEYS_MAPPING = map[string]rune{
	"up":    'u',
	"down":  'd',
	"right": 'r',
	"left":  'l',
}

type model struct {
	currentStratagem    stratagem
	stratagemCompletion int
	successes           int
	errors              int
	streak              int
	blockedTimer        timer.Model
	styles              Styles
}

func NewModel(styles Styles) model {
	return model{
		currentStratagem:    GetRandomStratagem(),
		stratagemCompletion: 0,
		successes:           0,
		errors:              0,
		streak:              0,
		styles:              styles,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.blockedTimer, cmd = m.blockedTimer.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}

		if m.blockedTimer.Running() {
			m.blockedTimer.Start()
			return m, nil
		}

		expectedKey := m.currentStratagem.code[m.stratagemCompletion]

		if KEYS_MAPPING[key] == expectedKey {
			m.stratagemCompletion++
		} else {
			m.errors++
			m.streak = 0
			m.blockedTimer = timer.New(time.Second * 2)
			return m, m.blockedTimer.Init()
		}

		if m.stratagemCompletion == len(m.currentStratagem.code) {
			m.successes++
			m.streak++
			m.stratagemCompletion = 0
			m.currentStratagem = GetRandomStratagem()
		}
	}

	return m, nil
}

func (m model) View() string {
	var output string

	output += m.styles.FormatScoreTable(m.successes, m.errors, m.streak)
	output += "\n\n"
	output += m.styles.FormatStratagem(m.currentStratagem, m.stratagemCompletion, m.blockedTimer.Running())

	return m.styles.FormatScreen(output)
}
