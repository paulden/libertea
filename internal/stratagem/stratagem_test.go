package stratagem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedStratagemsAreValid(t *testing.T) {
	stratagems, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if len(stratagems) < 100 {
		t.Errorf("expected the full wiki list, got %d stratagems", len(stratagems))
	}

	counts := map[string]int{}
	for _, s := range stratagems {
		counts[s.Category]++
	}
	for _, category := range Categories {
		if counts[category] == 0 {
			t.Errorf("no stratagem in category %q", category)
		}
	}
}

func TestParse(t *testing.T) {
	stratagems, err := Parse([]byte(`
stratagems:
  - name: Reinforce
    category: mission
    type: Ship
    code: [up, down, right, left, up]
`))
	if err != nil {
		t.Fatal(err)
	}
	want := Stratagem{Name: "Reinforce", Category: "mission", Kind: "Ship", Code: []rune{'u', 'd', 'r', 'l', 'u'}}
	if len(stratagems) != 1 || stratagems[0].Name != want.Name || stratagems[0].Category != want.Category ||
		stratagems[0].Kind != want.Kind || string(stratagems[0].Code) != string(want.Code) {
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
			if _, err := Parse([]byte(data)); err == nil {
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

	stratagems, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(stratagems) != 1 || stratagems[0].Name != "Custom" {
		t.Errorf("unexpected stratagems: %+v", stratagems)
	}

	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("a missing file should return an error")
	}
	if err := os.WriteFile(path, []byte("stratagems: []"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), path) {
		t.Errorf("errors should mention the file path, got %v", err)
	}
}

func TestEveryEmbeddedStratagemHasAnIcon(t *testing.T) {
	stratagems, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range stratagems {
		if !HasIcon(s.Icon) {
			t.Errorf("%s: icon %q is missing", s.Name, s.Icon)
		}
		if _, err := Icon(s.Icon); err != nil {
			t.Errorf("%s: %v", s.Name, err)
		}
	}
}

func TestUnknownIcon(t *testing.T) {
	_, err := Parse([]byte("stratagems:\n  - name: A\n    category: mission\n    code: [up]\n    icon: ../../etc/passwd\n"))
	if err == nil {
		t.Error("an unknown icon should return an error")
	}
}

func TestRandomNeverRepeatsTheExcludedStratagem(t *testing.T) {
	stratagems := []Stratagem{{Name: "A"}, {Name: "B"}}
	for range 100 {
		if Random(stratagems, "A").Name != "B" {
			t.Fatal("the excluded stratagem should not be picked")
		}
	}
	if Random(stratagems[:1], "A").Name != "A" {
		t.Error("a single stratagem should be picked even when excluded")
	}
}
