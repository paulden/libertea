package main

import (
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/muesli/cancelreader"
)

// Stratagem icons are hand traced from the game assets by Dogo314 for the
// Helldivers Wiki, see icons/README.md. They are generated with
// `go run ./cmd/update-stratagems`.
//
//go:embed icons/*.png
var iconFiles embed.FS

// Icons are displayed with the kitty graphics protocol and its Unicode
// placeholders: images are transmitted once, then drawn with regular text
// cells that bubbletea can render and redraw like any other character.
// See https://sw.kovidgoyal.net/kitty/graphics-protocol/#unicode-placeholders
const (
	ICON_COLUMNS       = 12
	ICON_ROWS          = 6
	ICONS_ENV_VAR      = "LIBERTEA_ICONS"
	DEFAULT_ICONS_MODE = "auto"

	KITTY_PLACEHOLDER = '\U0010EEEE'
	KITTY_CHUNK_SIZE  = 4096
	KITTY_QUERY       = "\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\"
	KITTY_QUERY_OK    = "\x1b_Gi=31;OK\x1b\\"
	DEVICE_ATTRIBUTES = "\x1b[c"
	QUERY_TIMEOUT     = time.Second

	ENTER_ALT_SCREEN = "\x1b[?1049h\x1b[H"
	EXIT_ALT_SCREEN  = "\x1b[?1049l"
	// Delete every image and free its data.
	DELETE_IMAGES = "\x1b_Ga=d,d=A,q=2\x1b\\"
)

// Combining characters encoding row and column numbers, from
// https://sw.kovidgoyal.net/kitty/_downloads/f0a0de9ec8d9ff4456206db8e0814937/rowcolumn-diacritics.txt
var KITTY_DIACRITICS = []rune{
	'̅', '̍', '̎', '̐', '̒', '̽', '̾', '̿',
	'͆', '͊', '͋', '͌', '͐', '͑', '͒', '͗',
}

func IconsModeNames() []string {
	return []string{DEFAULT_ICONS_MODE, "kitty", "none"}
}

func iconPath(icon string) string {
	return "icons/" + icon + ".png"
}

func HasIcon(icon string) bool {
	_, err := fs.Stat(iconFiles, iconPath(icon))
	return err == nil
}

// ResolveIconsMode tells whether icons can be displayed. The "auto" mode asks
// the terminal whether it supports the kitty graphics protocol.
func ResolveIconsMode(mode string, isTTY bool) (bool, error) {
	switch strings.ToLower(mode) {
	case "none":
		return false, nil
	case "kitty":
		return true, nil
	case DEFAULT_ICONS_MODE:
		if !isTTY || !term.IsTerminal(os.Stdin.Fd()) {
			return false, nil
		}
		return queryKittyGraphics(os.Stdin, os.Stdout, QUERY_TIMEOUT), nil
	default:
		return false, fmt.Errorf("unknown icons mode %q, expected one of: %s", mode, strings.Join(IconsModeNames(), ", "))
	}
}

// queryKittyGraphics sends a graphics query followed by a device attributes
// request, which every terminal answers. Supporting terminals answer the
// graphics query first.
func queryKittyGraphics(in, out *os.File, timeout time.Duration) bool {
	state, err := term.MakeRaw(in.Fd())
	if err != nil {
		return false
	}
	defer term.Restore(in.Fd(), state)

	reader, err := cancelreader.NewReader(in)
	if err != nil {
		return false
	}
	defer reader.Close()

	if _, err := io.WriteString(out, KITTY_QUERY+DEVICE_ATTRIBUTES); err != nil {
		return false
	}

	result := make(chan bool, 1)
	go func() {
		var response []byte
		buf := make([]byte, 256)
		for {
			n, err := reader.Read(buf)
			response = append(response, buf[:n]...)
			if supported, done := ParseQueryResponse(string(response)); done || err != nil {
				result <- supported
				return
			}
		}
	}()

	select {
	case supported := <-result:
		return supported
	case <-time.After(timeout):
		reader.Cancel()
		return false
	}
}

// ParseQueryResponse tells whether the terminal supports the kitty graphics
// protocol, and whether the response is complete, i.e. the device attributes
// answer was received.
func ParseQueryResponse(response string) (supported bool, done bool) {
	start := strings.Index(response, "\x1b[?")
	if start == -1 || !strings.Contains(response[start:], "c") {
		return false, false
	}
	return strings.Contains(response[:start], KITTY_QUERY_OK), true
}

// TransmitIcons sends the icons of the stratagems to the terminal, and
// returns the image id assigned to each icon.
func TransmitIcons(w io.Writer, stratagems []stratagem) (map[string]int, error) {
	ids := map[string]int{}
	for _, s := range stratagems {
		if s.icon == "" || ids[s.icon] != 0 {
			continue
		}
		data, err := iconFiles.ReadFile(iconPath(s.icon))
		if err != nil {
			return nil, err
		}

		id := len(ids) + 1
		if _, err := io.WriteString(w, TransmitCommands(id, data)); err != nil {
			return nil, err
		}
		ids[s.icon] = id
	}
	return ids, nil
}

// TransmitCommands builds the escape sequences that upload a PNG image and
// create a virtual placement for Unicode placeholders. Responses are
// suppressed so that they do not reach bubbletea as key presses.
func TransmitCommands(id int, png []byte) string {
	var b strings.Builder
	payload := base64.StdEncoding.EncodeToString(png)

	for offset := 0; offset < len(payload); offset += KITTY_CHUNK_SIZE {
		end := min(offset+KITTY_CHUNK_SIZE, len(payload))
		more := 0
		if end < len(payload) {
			more = 1
		}
		if offset == 0 {
			fmt.Fprintf(&b, "\x1b_Ga=t,t=d,f=100,i=%d,q=2,m=%d;%s\x1b\\", id, more, payload[offset:end])
		} else {
			fmt.Fprintf(&b, "\x1b_Gm=%d;%s\x1b\\", more, payload[offset:end])
		}
	}

	fmt.Fprintf(&b, "\x1b_Ga=p,U=1,i=%d,c=%d,r=%d,q=2\x1b\\", id, ICON_COLUMNS, ICON_ROWS)
	return b.String()
}

// IconPlaceholder draws an image with Unicode placeholders. The image id is
// encoded in the 24-bit foreground color.
func IconPlaceholder(id int) string {
	color := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", (id>>16)&0xFF, (id>>8)&0xFF, id&0xFF)

	lines := make([]string, ICON_ROWS)
	for row := range ICON_ROWS {
		var line strings.Builder
		line.WriteString(color)
		for column := range ICON_COLUMNS {
			line.WriteRune(KITTY_PLACEHOLDER)
			line.WriteRune(KITTY_DIACRITICS[row])
			line.WriteRune(KITTY_DIACRITICS[column])
		}
		line.WriteString("\x1b[39m")
		lines[row] = line.String()
	}
	return strings.Join(lines, "\n")
}
