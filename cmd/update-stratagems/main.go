// Command update-stratagems regenerates stratagems.yaml from the Helldivers Wiki.
//
// Usage: go run ./cmd/update-stratagems [-o stratagems.yaml]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	WIKI_PAGE_URL = "https://helldivers.wiki.gg/wiki/Stratagems"
	WIKI_API_URL  = "https://helldivers.wiki.gg/api.php?action=parse&page=Stratagems&prop=text&format=json&formatversion=2"
	USER_AGENT    = "libertea-update-stratagems (+https://github.com/paulden/libertea)"
)

type section struct {
	category string
	kind     string
}

// Wiki sections, keyed by their <summary> title, optionally followed by the
// <big> sub-title used in the mission section. Unknown sections are an error
// so that wiki layout changes are noticed.
var SECTIONS = map[string]*section{
	"Orbital Strikes":                {"offensive", "Orbital"},
	"Eagle Strikes":                  {"offensive", "Eagle"},
	"Support Weapons":                {"supply", "Support Weapon"},
	"Backpacks":                      {"supply", "Backpack"},
	"Vehicles":                       {"supply", "Vehicle"},
	"Sentries":                       {"defensive", "Sentry"},
	"Emplacements":                   {"defensive", "Emplacement"},
	"Mission Stratagems/Ship":        {"mission", "Ship"},
	"Mission Stratagems/Objective":   {"mission", "Objective"},
	"Mission Stratagems/Unavailable": nil,
}

var ARROWS = map[string]string{"Up": "up", "Down": "down", "Left": "left", "Right": "right"}

var (
	detailsRegexp  = regexp.MustCompile(`(?s)<summary>(.*?)</summary>(.*?)</details>`)
	subtitleRegexp = regexp.MustCompile(`<big><b>(.*?)</b></big>`)
	rowRegexp      = regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)
	cellRegexp     = regexp.MustCompile(`(?s)<td>(.*?)</td>`)
	tagRegexp      = regexp.MustCompile(`<.*?>`)
	arrowRegexp    = regexp.MustCompile(`alt="Stratagem Arrow (\w+)\.svg"`)
)

type stratagem struct {
	Name     string
	Category string
	Kind     string
	Code     []string
}

func main() {
	output := flag.String("o", "stratagems.yaml", "output file")
	flag.Parse()

	page, err := fetchPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	stratagems, err := ParsePage(page)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, []byte(RenderYAML(stratagems)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %d stratagems to %s\n", len(stratagems), *output)
}

func fetchPage() (string, error) {
	req, err := http.NewRequest(http.MethodGet, WIKI_API_URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", USER_AGENT)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status from the wiki: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Parse struct {
			Text string `json:"text"`
		} `json:"parse"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("cannot decode the wiki response: %w", err)
	}
	return parsed.Parse.Text, nil
}

// ParsePage extracts the stratagems from the rendered HTML of the wiki page.
// Duplicated names are dropped, keeping the first occurrence.
func ParsePage(page string) ([]stratagem, error) {
	var stratagems []stratagem
	seen := map[string]bool{}

	for _, details := range detailsRegexp.FindAllStringSubmatch(page, -1) {
		title, body := strings.TrimSpace(details[1]), details[2]

		// The mission section holds several tables, each one under a <big> sub-title.
		parts := subtitleRegexp.Split(body, -1)
		subtitles := subtitleRegexp.FindAllStringSubmatch(body, -1)
		keys := []string{title}
		for _, subtitle := range subtitles {
			keys = append(keys, title+"/"+subtitle[1])
		}

		for i, key := range keys {
			sec, known := SECTIONS[key]
			hasRows := rowRegexp.MatchString(parts[i])
			if !known {
				if hasRows {
					return nil, fmt.Errorf("unknown wiki section %q, update SECTIONS", key)
				}
				continue
			}
			if sec == nil {
				continue
			}

			for _, row := range rowRegexp.FindAllStringSubmatch(parts[i], -1) {
				cells := cellRegexp.FindAllStringSubmatch(row[1], -1)
				if len(cells) < 3 {
					continue
				}
				name := strings.TrimSpace(html.UnescapeString(tagRegexp.ReplaceAllString(cells[1][1], "")))
				var code []string
				for _, arrow := range arrowRegexp.FindAllStringSubmatch(cells[2][1], -1) {
					code = append(code, ARROWS[arrow[1]])
				}
				if name == "" || len(code) == 0 {
					return nil, fmt.Errorf("cannot parse a row of section %q", key)
				}
				if seen[name] {
					continue
				}
				seen[name] = true
				stratagems = append(stratagems, stratagem{name, sec.category, sec.kind, code})
			}
		}
	}

	if len(stratagems) == 0 {
		return nil, fmt.Errorf("no stratagem found, the wiki layout may have changed")
	}
	return stratagems, nil
}

func RenderYAML(stratagems []stratagem) string {
	var b strings.Builder
	b.WriteString("# Generated by `go run ./cmd/update-stratagems`, do not edit by hand.\n")
	fmt.Fprintf(&b, "# Source: %s\n", WIKI_PAGE_URL)
	b.WriteString("# Wiki content is licensed under CC BY-NC-SA 4.0: https://creativecommons.org/licenses/by-nc-sa/4.0\n")
	b.WriteString("#\n")
	b.WriteString("# Categories match the in-game permit colors:\n")
	b.WriteString("# offensive (red), supply (blue), defensive (green) and mission (yellow).\n")
	b.WriteString("stratagems:\n")
	for _, s := range stratagems {
		fmt.Fprintf(&b, "  - name: %s\n", scalar(s.Name))
		fmt.Fprintf(&b, "    category: %s\n", s.Category)
		fmt.Fprintf(&b, "    type: %s\n", scalar(s.Kind))
		fmt.Fprintf(&b, "    code: [%s]\n", strings.Join(s.Code, ", "))
	}
	return b.String()
}

// scalar renders a string as a YAML scalar, quoting it only when needed.
func scalar(value string) string {
	out, err := yaml.Marshal(value)
	if err != nil {
		panic(err)
	}
	return strings.TrimSuffix(string(out), "\n")
}
