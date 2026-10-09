package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"math/rand"
	"os"
	"slices"

	"go.yaml.in/yaml/v3"
)

const STRATAGEMS_ENV_VAR = "LIBERTEA_STRATAGEMS"

// Generated from the Helldivers Wiki with `go run ./cmd/update-stratagems`.
//
//go:embed stratagems.yaml
var embeddedStratagems []byte

var CATEGORIES = []string{"offensive", "supply", "defensive", "mission"}

var DIRECTIONS = map[string]rune{
	"up":    'u',
	"down":  'd',
	"left":  'l',
	"right": 'r',
}

type stratagem struct {
	name     string
	category string
	kind     string
	code     []rune
	icon     string
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

// LoadStratagems reads stratagems from a YAML file, or from the embedded list
// when no path is given.
func LoadStratagems(path string) ([]stratagem, error) {
	if path == "" {
		return ParseStratagems(embeddedStratagems)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	stratagems, err := ParseStratagems(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return stratagems, nil
}

func ParseStratagems(data []byte) ([]stratagem, error) {
	var file stratagemsFile
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("invalid stratagems file: %w", err)
	}

	stratagems := make([]stratagem, 0, len(file.Stratagems))
	seen := map[string]bool{}

	for i, entry := range file.Stratagems {
		if entry.Name == "" {
			return nil, fmt.Errorf("stratagem #%d has no name", i+1)
		}
		if seen[entry.Name] {
			return nil, fmt.Errorf("stratagem %q is defined twice", entry.Name)
		}
		seen[entry.Name] = true

		if !slices.Contains(CATEGORIES, entry.Category) {
			return nil, fmt.Errorf("stratagem %q has an unknown category %q, expected one of: %v", entry.Name, entry.Category, CATEGORIES)
		}
		if len(entry.Code) == 0 {
			return nil, fmt.Errorf("stratagem %q has no code", entry.Name)
		}

		code := make([]rune, 0, len(entry.Code))
		for _, word := range entry.Code {
			direction, ok := DIRECTIONS[word]
			if !ok {
				return nil, fmt.Errorf("stratagem %q has an unknown direction %q, expected up, down, left or right", entry.Name, word)
			}
			code = append(code, direction)
		}

		if entry.Icon != "" && !HasIcon(entry.Icon) {
			return nil, fmt.Errorf("stratagem %q has an unknown icon %q", entry.Name, entry.Icon)
		}

		stratagems = append(stratagems, stratagem{entry.Name, entry.Category, entry.Type, code, entry.Icon})
	}

	if len(stratagems) == 0 {
		return nil, fmt.Errorf("no stratagem defined")
	}
	return stratagems, nil
}

// GetRandomStratagem picks a random stratagem, avoiding the excluded one so
// that the same stratagem is never asked twice in a row.
func GetRandomStratagem(stratagems []stratagem, excludedName string) stratagem {
	for {
		candidate := stratagems[rand.Intn(len(stratagems))]
		if candidate.name != excludedName || len(stratagems) == 1 {
			return candidate
		}
	}
}
