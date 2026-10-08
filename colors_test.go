package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func fakeEnv(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestResolveColorProfile(t *testing.T) {
	cases := []struct {
		name  string
		mode  string
		env   map[string]string
		isTTY bool
		want  termenv.Profile
	}{
		{"docker default TERM", "auto", map[string]string{"TERM": "xterm"}, true, termenv.ANSI},
		{"256 colors TERM", "auto", map[string]string{"TERM": "xterm-256color"}, true, termenv.ANSI256},
		{"truecolor", "auto", map[string]string{"TERM": "xterm", "COLORTERM": "truecolor"}, true, termenv.TrueColor},
		{"tmux", "auto", map[string]string{"TERM": "tmux"}, true, termenv.ANSI},
		{"unknown TERM", "auto", map[string]string{"TERM": "vt100"}, true, termenv.Ascii},
		{"no TERM", "auto", map[string]string{}, true, termenv.Ascii},
		{"NO_COLOR", "auto", map[string]string{"TERM": "xterm-256color", "NO_COLOR": "1"}, true, termenv.Ascii},
		{"not a TTY", "auto", map[string]string{"TERM": "xterm"}, false, termenv.Ascii},
		{"forced 256", "256", map[string]string{}, false, termenv.ANSI256},
		{"forced none", "NONE", map[string]string{"TERM": "xterm-256color"}, true, termenv.Ascii},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ResolveColorProfile(c.mode, fakeEnv(c.env), c.isTTY)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got profile %v, want %v", got, c.want)
			}
		})
	}
}

func TestUnknownColorMode(t *testing.T) {
	if _, err := ResolveColorProfile("rainbow", fakeEnv(nil), true); err == nil {
		t.Error("unknown color mode should return an error")
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

func column(line, substr string) int {
	return lipgloss.Width(line[:strings.Index(line, substr)])
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
	for _, category := range CATEGORIES {
		if _, ok := CATEGORY_COLORS[category]; !ok {
			t.Errorf("category %q has no color", category)
		}
	}
}
