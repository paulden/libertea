package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var testStratagem = stratagem{name: "Test Stratagem", category: "offensive", kind: "Orbital", code: []rune{'u', 'd', 'l', 'r'}}

func newTestModel(t *testing.T, layoutName string) model {
	t.Helper()
	layout, err := GetLayout(layoutName)
	if err != nil {
		t.Fatal(err)
	}
	stratagems, err := LoadStratagems("")
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(NewStyles(termenv.ANSI256, nil), layout, stratagems)
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
	m := press(newTestModel(t, "arrows"), keyUp, keyDown, keyLeft, keyRight)

	if m.stats.successes != 1 || m.stats.streak != 1 || m.stats.bestStreak != 1 {
		t.Errorf("unexpected stats after success: %+v", m.stats)
	}
	if m.stratagemCompletion != 0 {
		t.Errorf("completion should be reset, got %d", m.stratagemCompletion)
	}
	if m.currentStratagem.name == testStratagem.name {
		t.Error("the same stratagem should not be asked twice in a row")
	}
	if m.stats.lastTime == 0 || m.stats.bestTime == 0 {
		t.Errorf("times should be recorded: %+v", m.stats)
	}
}

func TestLetterLayouts(t *testing.T) {
	cases := map[string][]rune{
		"wasd": {'w', 's', 'a', 'd'},
		"zqsd": {'z', 's', 'q', 'd'},
		"vim":  {'k', 'j', 'h', 'l'},
		"all":  {'Z', 'S', 'A', 'L'},
	}
	for layout, letters := range cases {
		t.Run(layout, func(t *testing.T) {
			m := newTestModel(t, layout)
			for _, letter := range letters {
				m = press(m, runeKey(letter))
			}
			if m.stats.successes != 1 || m.stats.errors != 0 {
				t.Errorf("unexpected stats: %+v", m.stats)
			}
		})
	}
}

func TestErrorRestartsStratagemFromScratch(t *testing.T) {
	m := press(newTestModel(t, "arrows"), keyUp, keyDown, keyUp)

	if m.stats.errors != 1 || m.stats.streak != 0 {
		t.Errorf("unexpected stats after error: %+v", m.stats)
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
	if m.stratagemCompletion != 0 || m.stats.errors != 1 {
		t.Error("input should be ignored while blocked")
	}
}

func TestUnmappedKeysAreIgnored(t *testing.T) {
	m := press(newTestModel(t, "arrows"), keyUp, runeKey('w'), runeKey('x'))

	if m.stats.errors != 0 || m.stratagemCompletion != 1 {
		t.Errorf("unmapped keys should be ignored, got completion %d and stats %+v", m.stratagemCompletion, m.stats)
	}
}

func TestQuitKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyEsc}, {Type: tea.KeyCtrlC}} {
		_, cmd := newTestModel(t, "all").Update(key)
		if cmd == nil {
			t.Fatalf("%s should quit", key)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("%s should quit", key)
		}
	}
}

func TestUnknownLayout(t *testing.T) {
	if _, err := GetLayout("dvorak"); err == nil {
		t.Error("unknown layout should return an error")
	}
	if _, err := GetLayout("ZQSD"); err != nil {
		t.Errorf("layout names should be case insensitive: %v", err)
	}
}

func TestViewIsCenteredInTheTerminal(t *testing.T) {
	m := newTestModel(t, "all")
	screen := m.View()
	screenWidth, screenHeight := lipgloss.Width(screen), lipgloss.Height(screen)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: screenWidth + 40, Height: screenHeight + 10})
	view := updated.(model).View()

	if lipgloss.Width(view) != screenWidth+40 || lipgloss.Height(view) != screenHeight+10 {
		t.Fatalf("view should fill the terminal, got %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
	lines := strings.Split(view, "\n")
	if strings.TrimSpace(lines[4]) != "" || strings.TrimSpace(lines[5]) == "" {
		t.Errorf("the screen should start after 5 blank lines")
	}
	if !strings.HasPrefix(lines[5], strings.Repeat(" ", 20)) || strings.HasPrefix(lines[5], strings.Repeat(" ", 21)) {
		t.Errorf("the screen should be indented by 20 columns: %q", lines[5])
	}
}

func TestViewInASmallTerminal(t *testing.T) {
	m := newTestModel(t, "all")
	screen := m.View()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	if view := updated.(model).View(); view != screen {
		t.Error("the screen should be left untouched when the terminal is too small")
	}
}
