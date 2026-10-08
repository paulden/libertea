package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedStratagemsAreValid(t *testing.T) {
	stratagems, err := LoadStratagems("")
	if err != nil {
		t.Fatal(err)
	}
	if len(stratagems) < 100 {
		t.Errorf("expected the full wiki list, got %d stratagems", len(stratagems))
	}

	categories := map[string]int{}
	for _, s := range stratagems {
		categories[s.category]++
	}
	for _, category := range CATEGORIES {
		if categories[category] == 0 {
			t.Errorf("no stratagem in category %q", category)
		}
	}
}

func TestParseStratagems(t *testing.T) {
	stratagems, err := ParseStratagems([]byte(`
stratagems:
  - name: Reinforce
    category: mission
    type: Ship
    code: [up, down, right, left, up]
`))
	if err != nil {
		t.Fatal(err)
	}
	want := stratagem{name: "Reinforce", category: "mission", kind: "Ship", code: []rune{'u', 'd', 'r', 'l', 'u'}}
	if len(stratagems) != 1 || stratagems[0].name != want.name || stratagems[0].category != want.category ||
		stratagems[0].kind != want.kind || string(stratagems[0].code) != string(want.code) {
		t.Errorf("got %+v, want %+v", stratagems, want)
	}
}

func TestParseInvalidStratagems(t *testing.T) {
	cases := map[string]string{
		"no stratagem":      `stratagems: []`,
		"no name":           "stratagems:\n  - category: mission\n    code: [up]",
		"unknown category":  "stratagems:\n  - name: A\n    category: fun\n    code: [up]",
		"no code":           "stratagems:\n  - name: A\n    category: mission\n    code: []",
		"unknown direction": "stratagems:\n  - name: A\n    category: mission\n    code: [north]",
		"duplicated name":   "stratagems:\n  - name: A\n    category: mission\n    code: [up]\n  - name: A\n    category: mission\n    code: [down]",
		"unknown field":     "stratagems:\n  - name: A\n    category: mission\n    code: [up]\n    cooldown: 10",
		"not YAML":          "{{{",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseStratagems([]byte(data)); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestLoadStratagemsFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.yaml")
	content := "stratagems:\n  - name: Custom\n    category: supply\n    code: [left, right]\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	stratagems, err := LoadStratagems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(stratagems) != 1 || stratagems[0].name != "Custom" {
		t.Errorf("unexpected stratagems: %+v", stratagems)
	}

	if _, err := LoadStratagems(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("a missing file should return an error")
	}
	if err := os.WriteFile(path, []byte("stratagems: []"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadStratagems(path); err == nil || !strings.Contains(err.Error(), path) {
		t.Errorf("errors should mention the file path, got %v", err)
	}
}
