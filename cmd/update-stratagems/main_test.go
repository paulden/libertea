package main

import (
	"strings"
	"testing"
)

func row(name string, arrows ...string) string {
	var code strings.Builder
	for _, arrow := range arrows {
		code.WriteString(`<span><img alt="Stratagem Arrow ` + arrow + `.svg" src="x" /></span>`)
	}
	return "<tr>\n<td>icon</td>\n<td><a href=\"/wiki/x\">" + name + "</a></td>\n<td>" + code.String() + "</td>\n<td>80s</td>\n</tr>\n"
}

func TestParsePage(t *testing.T) {
	page := `<details><summary>Orbital Strikes</summary><table><tr><th>Name</th></tr>` +
		row("Orbital Precision Strike", "Right", "Right", "Up") +
		`</table></details>
<details><summary>Mission Stratagems</summary>
<big><b>Ship</b></big><table>` + row("Reinforce", "Up", "Down", "Right", "Left", "Up") + `</table>
<big><b>Objective</b></big><table>` +
		row("SSSD Delivery", "Down", "Down", "Down", "Up", "Up") +
		row("SSSD Delivery", "Down", "Up") + `</table>
<big><b>Unavailable</b></big><table>` + row("Orbital Illumination Flare", "Right", "Right", "Left", "Left") + `</table>
</details>`

	stratagems, err := ParsePage(page)
	if err != nil {
		t.Fatal(err)
	}

	got := RenderYAML(stratagems)
	for _, want := range []string{
		"  - name: Orbital Precision Strike\n    category: offensive\n    type: Orbital\n    code: [right, right, up]\n",
		"  - name: Reinforce\n    category: mission\n    type: Ship\n    code: [up, down, right, left, up]\n",
		"  - name: SSSD Delivery\n    category: mission\n    type: Objective\n    code: [down, down, down, up, up]\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing entry:\n%s\nin:\n%s", want, got)
		}
	}
	if len(stratagems) != 3 {
		t.Errorf("duplicates and unavailable stratagems should be dropped, got %d stratagems", len(stratagems))
	}
}

func TestParsePageUnknownSection(t *testing.T) {
	page := `<details><summary>Boosters</summary><table>` + row("Hellpod Space Optimization", "Up") + `</table></details>`
	if _, err := ParsePage(page); err == nil {
		t.Error("an unknown section with stratagems should return an error")
	}
}

func TestScalarQuoting(t *testing.T) {
	if got := scalar("A/MG-43 Machine Gun Sentry"); got != "A/MG-43 Machine Gun Sentry" {
		t.Errorf("plain names should not be quoted, got %s", got)
	}
	if got := scalar("Name: with colon"); !strings.HasPrefix(got, `"`) && !strings.HasPrefix(got, `'`) {
		t.Errorf("names with special characters should be quoted, got %s", got)
	}
}
