package main

import (
	"fmt"
	"sort"
	"strings"
)

const (
	DEFAULT_LAYOUT = "all"
	LAYOUT_ENV_VAR = "LIBERTEA_LAYOUT"
)

type keyLayout struct {
	description string
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
var LAYOUTS = map[string]keyLayout{
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

func LayoutNames() []string {
	names := make([]string, 0, len(LAYOUTS))
	for name := range LAYOUTS {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func GetLayout(name string) (keyLayout, error) {
	layout, ok := LAYOUTS[strings.ToLower(name)]
	if !ok {
		return keyLayout{}, fmt.Errorf("unknown layout %q, expected one of: %s", name, strings.Join(LayoutNames(), ", "))
	}
	return layout, nil
}

// Direction returns the direction bound to a key, ignoring case so that
// Caps Lock does not break letter layouts.
func (l keyLayout) Direction(key string) (rune, bool) {
	direction, ok := l.keys[strings.ToLower(key)]
	return direction, ok
}
