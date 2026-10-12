package terminal

import (
	"testing"

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
