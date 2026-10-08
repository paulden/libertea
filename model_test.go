package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var testStratagem = stratagem{"Test Stratagem", []rune{'u', 'd', 'l', 'r'}}

func newTestModel(t *testing.T) model {
	t.Helper()
	m := NewModel(NewStyles())
	m.currentStratagem = testStratagem
	return m
}

func press(m model, keys ...tea.KeyMsg) model {
	for _, key := range keys {
		updated, _ := m.Update(key)
		m = updated.(model)
	}
	return m
}

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

var (
	keyUp    = tea.KeyMsg{Type: tea.KeyUp}
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
	keyLeft  = tea.KeyMsg{Type: tea.KeyLeft}
	keyRight = tea.KeyMsg{Type: tea.KeyRight}
)

func TestCompleteStratagemWithArrows(t *testing.T) {
	m := press(newTestModel(t), keyUp, keyDown, keyLeft, keyRight)

	if m.successes != 1 || m.streak != 1 {
		t.Errorf("unexpected stats after success: successes=%d streak=%d", m.successes, m.streak)
	}
	if m.stratagemCompletion != 0 {
		t.Errorf("completion should be reset, got %d", m.stratagemCompletion)
	}
}

func TestErrorRestartsStratagemFromScratch(t *testing.T) {
	m := press(newTestModel(t), keyUp, keyDown, keyUp)

	if m.errors != 1 || m.streak != 0 {
		t.Errorf("unexpected stats after error: errors=%d streak=%d", m.errors, m.streak)
	}
	if m.stratagemCompletion != 0 {
		t.Errorf("completion should restart from scratch, got %d", m.stratagemCompletion)
	}
	if !m.blockedTimer.Running() {
		t.Error("input should be blocked after an error")
	}
	if m.currentStratagem.name != testStratagem.name {
		t.Error("the stratagem should not change after an error")
	}

	m = press(m, keyUp)
	if m.stratagemCompletion != 0 || m.errors != 1 {
		t.Error("input should be ignored while blocked")
	}
}

func TestUnmappedKeysAreIgnored(t *testing.T) {
	m := press(newTestModel(t), keyUp, runeKey('w'), runeKey('x'))

	if m.errors != 0 || m.stratagemCompletion != 1 {
		t.Errorf("unmapped keys should be ignored, got completion %d and %d errors", m.stratagemCompletion, m.errors)
	}
}
