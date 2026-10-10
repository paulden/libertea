// Package stratagem loads the stratagems to train on, with their icons.
package stratagem

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"slices"

	"go.yaml.in/yaml/v3"
)

// Generated from the Helldivers Wiki with `go run ./cmd/update-stratagems`.
//
//go:embed stratagems.yaml
var embeddedStratagems []byte

// Stratagem icons are hand traced from the game assets by Dogo314 for the
// Helldivers Wiki, see icons/README.md. They are generated with
// `go run ./cmd/update-stratagems`.
//
//go:embed icons/*.png
var iconFiles embed.FS

var Categories = []string{"offensive", "supply", "defensive", "mission"}

var directions = map[string]rune{
	"up":    'u',
	"down":  'd',
	"left":  'l',
	"right": 'r',
}

type Stratagem struct {
	Name     string
	Category string
	// Kind is the type of the stratagem in the wiki, such as Orbital or Eagle.
	Kind string
	// Code is made of the directions 'u', 'd', 'l' and 'r'.
	Code []rune
	// Icon is the name of an embedded icon, empty when there is none.
	Icon string
}

type stratagemsFile struct {
	Stratagems []struct {
		Name     string   `yaml:"name"`
		Category string   `yaml:"category"`
		Type     string   `yaml:"type"`
		Code     []string `yaml:"code"`
		Icon     string   `yaml:"icon"`
	} `yaml:"stratagems"`
}

// Load reads stratagems from a YAML file, or from the embedded list when no
// path is given.
func Load(path string) ([]Stratagem, error) {
	if path == "" {
		return Parse(embeddedStratagems)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	stratagems, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return stratagems, nil
}

func Parse(data []byte) ([]Stratagem, error) {
	var file stratagemsFile
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("invalid stratagems file: %w", err)
	}

	stratagems := make([]Stratagem, 0, len(file.Stratagems))
	seen := map[string]bool{}

	for i, entry := range file.Stratagems {
		if entry.Name == "" {
			return nil, fmt.Errorf("stratagem #%d has no name", i+1)
		}
		if seen[entry.Name] {
			return nil, fmt.Errorf("stratagem %q is defined twice", entry.Name)
		}
		seen[entry.Name] = true

		if !slices.Contains(Categories, entry.Category) {
			return nil, fmt.Errorf("stratagem %q has an unknown category %q, expected one of: %v", entry.Name, entry.Category, Categories)
		}
		if len(entry.Code) == 0 {
			return nil, fmt.Errorf("stratagem %q has no code", entry.Name)
		}

		code := make([]rune, 0, len(entry.Code))
		for _, word := range entry.Code {
			direction, ok := directions[word]
			if !ok {
				return nil, fmt.Errorf("stratagem %q has an unknown direction %q, expected up, down, left or right", entry.Name, word)
			}
			code = append(code, direction)
		}

		if entry.Icon != "" && !HasIcon(entry.Icon) {
			return nil, fmt.Errorf("stratagem %q has an unknown icon %q", entry.Name, entry.Icon)
		}

		stratagems = append(stratagems, Stratagem{entry.Name, entry.Category, entry.Type, code, entry.Icon})
	}

	if len(stratagems) == 0 {
		return nil, fmt.Errorf("no stratagem defined")
	}
	return stratagems, nil
}

// Random picks a random stratagem, avoiding the excluded one so that the same
// stratagem is never asked twice in a row.
func Random(stratagems []Stratagem, excludedName string) Stratagem {
	for {
		candidate := stratagems[rand.Intn(len(stratagems))]
		if candidate.Name != excludedName || len(stratagems) == 1 {
			return candidate
		}
	}
}

func iconPath(icon string) string {
	return "icons/" + icon + ".png"
}

// HasIcon tells whether an icon is embedded.
func HasIcon(icon string) bool {
	_, err := fs.Stat(iconFiles, iconPath(icon))
	return err == nil
}

// Icon returns the PNG data of an embedded icon.
func Icon(icon string) ([]byte, error) {
	return iconFiles.ReadFile(iconPath(icon))
}
