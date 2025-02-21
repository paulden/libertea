package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	currentStratagem    stratagem
	stratagemCompletion int
	successes           int
	errors              int
	streak              int
	blockedTimer        timer.Model
}

func initialModel() model {
	return model{
		currentStratagem:    GetRandomStratagem(),
		stratagemCompletion: 0,
		successes:           0,
		errors:              0,
		streak:              0,
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
			m.blockedTimer = timer.New(time.Second * 3)
			return m, m.blockedTimer.Init()
		}

		if m.stratagemCompletion == len(m.currentStratagem.code) {
			m.successes++
			m.stratagemCompletion = 0
			m.currentStratagem = GetRandomStratagem()
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Call for your next stratagem!\n\n"

	strategemRendering := ""

	s += fmt.Sprintf("Current stratagem: %s\n\n", m.currentStratagem.name)

	for i, arrow := range m.currentStratagem.code {
		if i < m.stratagemCompletion && !m.blockedTimer.Running() {
			strategemRendering += validInput.Render(ARROWS_DISPLAY[arrow])
		} else if m.blockedTimer.Running() {
			strategemRendering += wrongInput.Render(ARROWS_DISPLAY[arrow])
		} else {
			strategemRendering += ARROWS_DISPLAY[arrow]
		}
	}

	s += fmt.Sprintf("%s \n", strategemRendering)

	if m.blockedTimer.Running() {
		s += fmt.Sprintf("Blocked for %s!", m.blockedTimer.View())
	}

	s += "\nPress q to quit.\n"

	return s
}
