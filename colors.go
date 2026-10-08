package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/muesli/termenv"
)

const (
	DEFAULT_COLOR_MODE = "auto"
	COLOR_ENV_VAR      = "LIBERTEA_COLOR"
)

var COLOR_MODES = map[string]termenv.Profile{
	"none":      termenv.Ascii,
	"16":        termenv.ANSI,
	"256":       termenv.ANSI256,
	"truecolor": termenv.TrueColor,
}

// Terminal families whose terminfo entries declare colors even when TERM does
// not say so explicitly (e.g. Docker sets TERM=xterm when a TTY is attached).
var COLOR_CAPABLE_TERMS = []string{"xterm", "screen", "tmux", "rxvt"}

func ColorModeNames() []string {
	return []string{DEFAULT_COLOR_MODE, "none", "16", "256", "truecolor"}
}

// ResolveColorProfile turns a color mode into a termenv profile. The "auto"
// mode relies on termenv detection, which honors NO_COLOR and CLICOLOR_FORCE.
func ResolveColorProfile(mode string, getenv func(string) string, isTTY bool) (termenv.Profile, error) {
	mode = strings.ToLower(mode)
	if mode != DEFAULT_COLOR_MODE {
		profile, ok := COLOR_MODES[mode]
		if !ok {
			return termenv.Ascii, fmt.Errorf("unknown color mode %q, expected one of: %s", mode, strings.Join(ColorModeNames(), ", "))
		}
		return profile, nil
	}

	output := termenv.NewOutput(os.Stdout, termenv.WithEnvironment(environ(getenv)), termenv.WithTTY(isTTY))
	profile := output.EnvColorProfile()

	if profile == termenv.Ascii && isTTY && !output.EnvNoColor() && isColorCapableTerm(getenv("TERM")) {
		return termenv.ANSI, nil
	}
	return profile, nil
}

func isColorCapableTerm(termName string) bool {
	for _, family := range COLOR_CAPABLE_TERMS {
		if termName == family || strings.HasPrefix(termName, family+"-") {
			return true
		}
	}
	return false
}

func StdoutIsTTY() bool {
	return term.IsTerminal(os.Stdout.Fd())
}

type environ func(string) string

func (e environ) Getenv(key string) string {
	return e(key)
}

func (e environ) Environ() []string {
	return os.Environ()
}
