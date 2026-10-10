package terminal

import (
	"encoding/base64"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestIconPlaceholder(t *testing.T) {
	placeholder := IconPlaceholder(0x010203)

	if lipgloss.Width(placeholder) != IconColumns || lipgloss.Height(placeholder) != IconRows {
		t.Errorf("placeholder should be %dx%d cells, got %dx%d", IconColumns, IconRows, lipgloss.Width(placeholder), lipgloss.Height(placeholder))
	}

	lines := strings.Split(placeholder, "\n")
	for row, line := range lines {
		if !strings.HasPrefix(line, "\x1b[38;2;1;2;3m") {
			t.Errorf("row %d: the image id should be encoded in the foreground color: %q", row, line)
		}
		cells := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "\x1b[38;2;1;2;3m"), "\x1b[39m"), string(kittyPlaceholder))[1:]
		for column, cell := range cells {
			want := string([]rune{kittyDiacritics[row], kittyDiacritics[column]})
			if cell != want {
				t.Errorf("cell %d,%d: got diacritics %q, want %q", row, column, cell, want)
			}
		}
	}
}

func TestTransmitCommands(t *testing.T) {
	png := []byte(strings.Repeat("x", 7000))
	commands := TransmitCommands(7, png)

	chunks := regexp.MustCompile(`\x1b_G([^;\x1b]*)(?:;([^\x1b]*))?\x1b\\`).FindAllStringSubmatch(commands, -1)
	if len(chunks) != 4 {
		t.Fatalf("expected 3 data chunks and a placement, got %d commands", len(chunks))
	}

	if chunks[0][1] != "a=t,t=d,f=100,i=7,q=2,m=1" || chunks[1][1] != "m=1" || chunks[2][1] != "m=0" {
		t.Errorf("unexpected chunk controls: %q, %q, %q", chunks[0][1], chunks[1][1], chunks[2][1])
	}
	var payload string
	for _, chunk := range chunks[:3] {
		if len(chunk[2]) > kittyChunkSize {
			t.Errorf("chunk larger than %d bytes: %d", kittyChunkSize, len(chunk[2]))
		}
		payload += chunk[2]
	}
	if len(chunks[0][2])%4 != 0 || len(chunks[1][2])%4 != 0 {
		t.Error("every chunk but the last one should have a size multiple of 4")
	}
	if decoded, err := base64.StdEncoding.DecodeString(payload); err != nil || string(decoded) != string(png) {
		t.Errorf("payload does not decode to the image: %v", err)
	}

	if chunks[3][1] != "a=p,U=1,i=7,c=12,r=6,q=2" {
		t.Errorf("unexpected placement: %q", chunks[3][1])
	}
}

func TestParseQueryResponse(t *testing.T) {
	cases := []struct {
		name      string
		response  string
		supported bool
		done      bool
	}{
		{"supported", "\x1b_Gi=31;OK\x1b\\\x1b[?62;22c", true, true},
		{"unsupported", "\x1b[?62;22c", false, true},
		{"error from the terminal", "\x1b_Gi=31;EINVAL:bad\x1b\\\x1b[?62;22c", false, true},
		{"partial graphics answer", "\x1b_Gi=31;OK\x1b\\", false, false},
		{"partial attributes", "\x1b_Gi=31;OK\x1b\\\x1b[?62;2", false, false},
		{"empty", "", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			supported, done := ParseQueryResponse(c.response)
			if supported != c.supported || done != c.done {
				t.Errorf("got supported=%v done=%v, want supported=%v done=%v", supported, done, c.supported, c.done)
			}
		})
	}
}

func TestResolveIconsMode(t *testing.T) {
	if show, err := ResolveIconsMode("kitty", false); err != nil || !show {
		t.Errorf("kitty mode should force icons, got %v, %v", show, err)
	}
	if show, err := ResolveIconsMode("NONE", true); err != nil || show {
		t.Errorf("none mode should disable icons, got %v, %v", show, err)
	}
	if show, err := ResolveIconsMode("auto", false); err != nil || show {
		t.Errorf("auto mode should disable icons without a terminal, got %v, %v", show, err)
	}
	if _, err := ResolveIconsMode("sixel", true); err == nil {
		t.Error("unknown mode should return an error")
	}
}
