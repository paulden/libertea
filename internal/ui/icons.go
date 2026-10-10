package ui

import (
	"io"

	"github.com/paulden/libertea/internal/stratagem"
	"github.com/paulden/libertea/internal/terminal"
)

// TransmitIcons sends the icons of the stratagems to the terminal, and
// returns the image id assigned to each icon.
func TransmitIcons(w io.Writer, stratagems []stratagem.Stratagem) (map[string]int, error) {
	ids := map[string]int{}
	for _, s := range stratagems {
		if s.Icon == "" || ids[s.Icon] != 0 {
			continue
		}
		data, err := stratagem.Icon(s.Icon)
		if err != nil {
			return nil, err
		}

		id := len(ids) + 1
		if _, err := io.WriteString(w, terminal.TransmitCommands(id, data)); err != nil {
			return nil, err
		}
		ids[s.Icon] = id
	}
	return ids, nil
}
