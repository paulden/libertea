package ui

import (
	"strings"
	"testing"

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
		if !strings.Contains(render, "Test Stratagem") {
			t.Errorf("ids %v: the name should always be shown", ids)
		}
		if lipgloss.Width(render) > stratagemWidth+1 {
			t.Errorf("ids %v: render is too wide (%d)", ids, lipgloss.Width(render))
		}
	}
}

func TestCursorShownWithoutColors(t *testing.T) {
	render := NewStyles(termenv.Ascii, nil).FormatStratagem(testStratagem, 2, false, 0)
	lines := strings.Split(render, "\n")

	arrowsLine, cursorLine := -1, -1
	for i, line := range lines {
		if strings.Contains(line, "↑ ↓ ← →") {
			arrowsLine = i
		}
		if strings.Contains(line, "^") {
			cursorLine = i
		}
	}
	if arrowsLine == -1 || cursorLine != arrowsLine+1 {
		t.Fatalf("cursor should be right under the arrows:\n%s", render)
	}
	if column(lines[cursorLine], "^") != column(lines[arrowsLine], "←") {
		t.Errorf("cursor should be under the third arrow:\n%s", render)
	}
}

func TestNoCursorWithColors(t *testing.T) {
	render := NewStyles(termenv.ANSI, nil).FormatStratagem(testStratagem, 2, false, 0)
	if strings.Contains(render, "^") {
		t.Errorf("cursor should only be shown without colors:\n%s", render)
	}
}

func TestCategoryShownWithAndWithoutColors(t *testing.T) {
	for _, profile := range []termenv.Profile{termenv.Ascii, termenv.ANSI, termenv.ANSI256, termenv.TrueColor} {
		render := NewStyles(profile, nil).FormatStratagem(testStratagem, 0, false, 0)
		if !strings.Contains(render, "OFFENSIVE · ORBITAL") {
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
