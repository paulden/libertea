// Package keys maps keyboard keys to stratagem directions.
package keys

import (
	"fmt"
	"sort"
	"strings"
)

const Default = "all"

type Layout struct {
	// Description lists the keys of the layout, for the help line.
	Description string
	keys        map[string]rune
}

var arrowKeys = map[string]rune{
	"up":    'u',
	"down":  'd',
	"left":  'l',
	"right": 'r',
}

var wasdKeys = map[string]rune{
	"w": 'u',
	"s": 'd',
	"a": 'l',
	"d": 'r',
}

var zqsdKeys = map[string]rune{
	"z": 'u',
	"s": 'd',
	"q": 'l',
	"d": 'r',
}

var vimKeys = map[string]rune{
	"k": 'u',
	"j": 'd',
	"h": 'l',
	"l": 'r',
}

// None of the letter layouts overlap in a conflicting way (S and D mean the
// same direction in WASD and ZQSD), so they can all be enabled at once.
var layouts = map[string]Layout{
	"arrows": {"arrows", mergeKeys(arrowKeys)},
	"wasd":   {"arrows / WASD", mergeKeys(arrowKeys, wasdKeys)},
	"zqsd":   {"arrows / ZQSD", mergeKeys(arrowKeys, zqsdKeys)},
	"vim":    {"arrows / HJKL", mergeKeys(arrowKeys, vimKeys)},
	"all":    {"arrows / WASD / ZQSD / HJKL", mergeKeys(arrowKeys, wasdKeys, zqsdKeys, vimKeys)},
}

func mergeKeys(mappings ...map[string]rune) map[string]rune {
	merged := map[string]rune{}
	for _, mapping := range mappings {
		for key, direction := range mapping {
			merged[key] = direction
		}
	}
	return merged
}

func Names() []string {
	names := make([]string, 0, len(layouts))
	for name := range layouts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Get(name string) (Layout, error) {
	layout, ok := layouts[strings.ToLower(name)]
	if !ok {
		return Layout{}, fmt.Errorf("unknown layout %q, expected one of: %s", name, strings.Join(Names(), ", "))
	}
	return layout, nil
}

// Direction returns the direction bound to a key, ignoring case so that
// Caps Lock does not break letter layouts.
func (l Layout) Direction(key string) (rune, bool) {
	direction, ok := l.keys[strings.ToLower(key)]
	return direction, ok
}
