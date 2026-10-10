package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/paulden/libertea/internal/keys"
	"github.com/paulden/libertea/internal/stratagem"
)

var testStratagem = stratagem.Stratagem{Name: "Test Stratagem", Category: "offensive", Kind: "Orbital", Code: []rune{'u', 'd', 'l', 'r'}}

func newTestModel(t *testing.T, layoutName string) Model {
	t.Helper()
	layout, err := keys.Get(layoutName)
	if err != nil {
		t.Fatal(err)
	}
	stratagems, err := stratagem.Load("")
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(NewStyles(termenv.ANSI256, nil), layout, stratagems)
	m.currentStratagem = testStratagem
	m.now = newFakeClock().now
	return m
}

// fakeClock moves forward by step each time it is read.
type fakeClock struct {
	current time.Time
	step    time.Duration
}

func newFakeClock() *fakeClock {
	return &fakeClock{current: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), step: 100 * time.Millisecond}
}

func (c *fakeClock) now() time.Time {
	c.current = c.current.Add(c.step)
	return c.current
}

func press(m Model, keys ...tea.KeyMsg) Model {
	for _, key := range keys {
		updated, _ := m.Update(key)
		m = updated.(Model)
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
	if m.currentStratagem.Name == testStratagem.Name {
		t.Error("the same stratagem should not be asked twice in a row")
	}
	if m.stats.lastTime == 0 || m.stats.bestTime == 0 {
		t.Errorf("times should be recorded: %+v", m.stats)
	}
}

func TestTimesAreMeasuredFromTheFirstKey(t *testing.T) {
	m := newTestModel(t, "arrows")
	clock := newFakeClock()
	m.now = clock.now

	// The clock is read on the first and the last key.
	m = press(m, keyUp, keyDown, keyLeft, keyRight)
	if m.stats.lastTime != clock.step || m.stats.bestTime != clock.step {
		t.Errorf("got last %v and best %v, want %v", m.stats.lastTime, m.stats.bestTime, clock.step)
	}

	clock.step = 300 * time.Millisecond
	m.currentStratagem = testStratagem
	m = press(m, keyUp, keyDown, keyLeft, keyRight)
	if m.stats.lastTime != 300*time.Millisecond || m.stats.bestTime != 100*time.Millisecond {
		t.Errorf("a slower stratagem should not change the best time, got last %v and best %v", m.stats.lastTime, m.stats.bestTime)
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
	if m.currentStratagem.Name != testStratagem.Name {
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

func TestViewIsCenteredInTheTerminal(t *testing.T) {
	m := newTestModel(t, "all")
	screen := m.View()
	screenWidth, screenHeight := lipgloss.Width(screen), lipgloss.Height(screen)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: screenWidth + 40, Height: screenHeight + 10})
	view := updated.(Model).View()

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
	if view := updated.(Model).View(); view != screen {
		t.Error("the screen should be left untouched when the terminal is too small")
	}
}

func TestNewBestAndRecentTimes(t *testing.T) {
	m := newTestModel(t, "arrows")
	clock := newFakeClock()
	m.now = clock.now

	for i := range sparklineSize + 2 {
		m.currentStratagem = testStratagem
		m = press(m, keyUp, keyDown, keyLeft, keyRight)
		if m.stats.newBest {
			t.Errorf("success %d: a time equal to the best is not a new best", i+1)
		}
	}
	if len(m.stats.recentTimes) != sparklineSize {
		t.Errorf("the sparkline should keep the last %d times, got %d", sparklineSize, len(m.stats.recentTimes))
	}

	clock.step = 50 * time.Millisecond
	m.currentStratagem = testStratagem
	m = press(m, keyUp, keyDown, keyLeft, keyRight)
	if !m.stats.newBest || m.stats.bestTime != 50*time.Millisecond {
		t.Errorf("a faster time should be a new best: %+v", m.stats)
	}
	if last := m.stats.recentTimes[len(m.stats.recentTimes)-1]; last != 50*time.Millisecond {
		t.Errorf("the last time should be at the end of the sparkline, got %v", last)
	}
}

func TestScreenKeepsItsSizeInEveryState(t *testing.T) {
	for _, profile := range []termenv.Profile{termenv.Ascii, termenv.TrueColor} {
		m := newTestModel(t, "all")
		m.styles = NewStyles(profile, nil)
		playing := m.View()
		blocked := press(m, keyDown).View()

		if lipgloss.Height(playing) != lipgloss.Height(blocked) {
			t.Errorf("profile %v: the height changed after a wrong input: %d, then %d", profile, lipgloss.Height(playing), lipgloss.Height(blocked))
		}
		for _, view := range []string{playing, blocked} {
			if lipgloss.Width(view) != screenWidth {
				t.Errorf("profile %v: the screen should be %d cells wide, got %d", profile, screenWidth, lipgloss.Width(view))
			}
		}
	}
}

func TestTimeStartsOverAfterAnError(t *testing.T) {
	m := newTestModel(t, "arrows")
	clock := newFakeClock()
	m.now = clock.now

	m = press(m, keyUp, keyUp)
	// Let the penalty pass, without waiting for the timer.
	clock.current = clock.current.Add(penaltyDuration)
	m.blockedTimer.Timeout = 0
	m = press(m, keyUp, keyDown, keyLeft, keyRight)

	if m.stats.lastTime != clock.step {
		t.Errorf("the time should be measured from the first key after the penalty, got %v, want %v", m.stats.lastTime, clock.step)
	}
}
