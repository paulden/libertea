package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/paulden/libertea/internal/stratagem"
)

func TestTransmitIconsAssignsOneIdPerIcon(t *testing.T) {
	stratagems := []stratagem.Stratagem{
		{Name: "A", Icon: "reinforce"},
		{Name: "B", Icon: "resupply"},
		{Name: "C", Icon: "reinforce"},
		{Name: "D"},
	}
	var out strings.Builder
	ids, err := TransmitIcons(&out, stratagems)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids["reinforce"] != 1 || ids["resupply"] != 2 {
		t.Errorf("unexpected ids: %v", ids)
	}
	if strings.Count(out.String(), "a=p,U=1") != 2 {
		t.Error("each icon should be transmitted once")
	}
}

func TestIconShownNextToTheStratagem(t *testing.T) {
	withIcon := testStratagem
	withIcon.Icon = "orbital-gatling-barrage"

	for _, ids := range []map[string]int{nil, {"orbital-gatling-barrage": 3}} {
		render := NewStyles(termenv.TrueColor, ids).FormatStratagem(withIcon, 0, false, 0)
		shown := strings.ContainsRune(render, '\U0010EEEE')
		if shown != (ids != nil) {
			t.Errorf("ids %v: icon shown = %v", ids, shown)
		}
		if !strings.Contains(render, "TEST STRATAGEM") {
			t.Errorf("ids %v: the name should always be shown", ids)
		}
		if lipgloss.Width(render) > innerWidth {
			t.Errorf("ids %v: render is too wide (%d)", ids, lipgloss.Width(render))
		}
	}
}

func TestCursorShownWithoutColors(t *testing.T) {
	render := NewStyles(termenv.Ascii, nil).FormatStratagem(testStratagem, 2, false, 0)
	lines := strings.Split(render, "\n")

	arrowsLine, cursorLine := -1, -1
	for i, line := range lines {
		if strings.Contains(line, "│ ↑ │ ↓ │ ← │ → │") {
			arrowsLine = i
		}
		if strings.Contains(line, "▲") {
			cursorLine = i
		}
	}
	// The bottom border of the cells sits between the arrows and the cursor.
	if arrowsLine == -1 || cursorLine != arrowsLine+2 {
		t.Fatalf("cursor should be right under the cells:\n%s", render)
	}
	if column(lines[cursorLine], "▲") != column(lines[arrowsLine], "←") {
		t.Errorf("cursor should be under the third arrow:\n%s", render)
	}
}

func TestNoCursorWithColors(t *testing.T) {
	render := NewStyles(termenv.ANSI, nil).FormatStratagem(testStratagem, 2, false, 0)
	if strings.Contains(render, "▲") {
		t.Errorf("cursor should only be shown without colors:\n%s", render)
	}
}

func TestCategoryShownWithAndWithoutColors(t *testing.T) {
	for _, profile := range []termenv.Profile{termenv.Ascii, termenv.ANSI, termenv.ANSI256, termenv.TrueColor} {
		render := NewStyles(profile, nil).FormatStratagem(testStratagem, 0, false, 0)
		if !strings.Contains(render, "Offensive · Orbital") {
			t.Errorf("profile %v: the category label should always be shown:\n%s", profile, render)
		}
	}
}

func TestEveryCategoryHasAColor(t *testing.T) {
	for _, category := range stratagem.Categories {
		if _, ok := categoryColors[category]; !ok {
			t.Errorf("category %q has no color", category)
		}
	}
}

func column(line, substr string) int {
	return lipgloss.Width(line[:strings.Index(line, substr)])
}

func TestStylesDoNotShareTheColorProfile(t *testing.T) {
	plain := NewStyles(termenv.Ascii, nil)
	NewStyles(termenv.TrueColor, nil)

	if render := plain.FormatStratagem(testStratagem, 1, false, 0); strings.Contains(render, "\x1b[") {
		t.Errorf("styles without colors should not render escape sequences after other styles were created:\n%q", render)
	}
}

func TestPenaltyBarDrains(t *testing.T) {
	styles := NewStyles(termenv.Ascii, nil)
	for _, c := range []struct {
		remaining time.Duration
		full      int
	}{
		{penaltyDuration, penaltyBarSize},
		{penaltyDuration / 2, penaltyBarSize / 2},
		{0, 0},
	} {
		render := styles.FormatStratagem(testStratagem, 0, true, c.remaining)
		if got := strings.Count(render, "█"); got != c.full {
			t.Errorf("%v left: got %d full cells, want %d:\n%s", c.remaining, got, c.full, render)
		}
		if got := strings.Count(render, "█") + strings.Count(render, "░"); got != penaltyBarSize {
			t.Errorf("%v left: the bar should keep %d cells, got %d", c.remaining, penaltyBarSize, got)
		}
	}
}

func TestTimesComparedToTheBest(t *testing.T) {
	styles := NewStyles(termenv.Ascii, nil)
	cases := []struct {
		name  string
		stats stats
		want  string
	}{
		{"no time yet", stats{}, "LAST -    BEST -"},
		{"slower than the best", stats{lastTime: 1420 * time.Millisecond, bestTime: 1180 * time.Millisecond}, "LAST 1.42s  ▲0.24    BEST 1.18s"},
		{"new best", stats{lastTime: time.Second, bestTime: time.Second, newBest: true}, "LAST 1.00s  NEW BEST    BEST 1.00s"},
		{"as fast as the best", stats{lastTime: 1003 * time.Millisecond, bestTime: time.Second}, "LAST 1.00s    BEST 1.00s"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := styles.FormatTimes(c.stats, false); !strings.Contains(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
	if got := styles.FormatTimes(stats{}, true); !strings.Contains(got, "WRONG INPUT") {
		t.Errorf("a wrong input should replace the times, got %q", got)
	}
}

func TestSparkline(t *testing.T) {
	ms := func(values ...int) []time.Duration {
		times := make([]time.Duration, len(values))
		for i, v := range values {
			times[i] = time.Duration(v) * time.Millisecond
		}
		return times
	}
	cases := []struct {
		times []time.Duration
		want  string
	}{
		{nil, ""},
		{ms(1000), ""},
		{ms(1000, 2000), "▁█"},
		{ms(1000, 1500, 2000), "▁▄█"},
		{ms(1200, 1200), "▅▅"},
	}
	for _, c := range cases {
		if got := sparkline(c.times); got != c.want {
			t.Errorf("sparkline(%v) = %q, want %q", c.times, got, c.want)
		}
	}
}

func TestFrameEmbedsTitleAndKeys(t *testing.T) {
	screen := NewStyles(termenv.Ascii, nil).FormatScreen("body", "←↑↓→ zqsd", false)
	lines := strings.Split(screen, "\n")

	if !strings.HasPrefix(lines[0], "╭─ LIBERTEA ") || !strings.HasSuffix(lines[0], " STRATAGEM DRILL ─╮") {
		t.Errorf("the title should be embedded in the top border: %q", lines[0])
	}
	if !strings.HasPrefix(lines[2], "╰─ ←↑↓→ zqsd move  ·  esc quit ") {
		t.Errorf("the keys should be embedded in the bottom border: %q", lines[2])
	}
	for i, line := range lines {
		if lipgloss.Width(line) != screenWidth {
			t.Errorf("line %d should be %d cells wide, got %d: %q", i, screenWidth, lipgloss.Width(line), line)
		}
	}
}

func TestFrameTurnsRedAfterAWrongInput(t *testing.T) {
	styles := NewStyles(termenv.TrueColor, nil)
	red := "38;2;224;67;79"
	if screen := styles.FormatScreen("body", "", false); strings.Contains(screen, red) {
		t.Error("the frame should not be red while playing")
	}
	if screen := styles.FormatScreen("body", "", true); !strings.Contains(strings.Split(screen, "\n")[0], red) {
		t.Errorf("the frame should be red after a wrong input: %q", screen)
	}
}

func TestNoticeWithoutColors(t *testing.T) {
	if notice := NewStyles(termenv.Ascii, nil).FormatNotice(); !strings.Contains(notice, "-color 256") {
		t.Errorf("a notice should suggest forcing colors, got %q", notice)
	}
	if notice := NewStyles(termenv.ANSI, nil).FormatNotice(); notice != "" {
		t.Errorf("no notice expected with colors, got %q", notice)
	}
}

func TestStratagemDoesNotMoveDuringThePenalty(t *testing.T) {
	for _, ids := range []map[string]int{nil, {"orbital-gatling-barrage": 3}} {
		styles := NewStyles(termenv.Ascii, ids)
		strat := testStratagem
		strat.Icon = "orbital-gatling-barrage"

		playing := strings.Split(styles.FormatStratagem(strat, 0, false, 0), "\n")
		blocked := strings.Split(styles.FormatStratagem(strat, 0, true, penaltyDuration), "\n")
		if column(playing[2], "┌") != column(blocked[2], "┌") {
			t.Errorf("ids %v: the strip moved from column %d to %d", ids, column(playing[2], "┌"), column(blocked[2], "┌"))
		}
	}
}
