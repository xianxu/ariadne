package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/trains"
)

// CarState is what the renderer knows about one car. Err set means the car's
// read failed: it renders `?`, never as waiting (a failed read is not evidence).
type CarState struct {
	Status   string
	Claimant string
	Bar      string
	Err      string
}

// carGlyph: ✓ done, ◑ codecomplete (landing), ◐ in flight (claimed and
// started), ○ waiting, ? unread.
func carGlyph(c CarState) string {
	switch {
	case c.Err != "":
		return "?"
	case c.Status == "done":
		return "✓"
	case c.Status == "codecomplete":
		return "◑"
	case c.Claimant != "" && c.Status != "open":
		return "◐"
	}
	return "○"
}

func runes(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		r = r[:n]
	}
	return string(r)
}

var slotCellRE = regexp.MustCompile(`[(,]`)

// renderTrains renders the `## Trains` block: one line per train, its glyphs,
// and its chain, with the first in-flight car carrying its bar. Every car is
// named in full repo#id form.
func renderTrains(plan trains.TrainPlan, cars map[trains.IssueRef]CarState) string {
	lines := []string{"```"}
	for _, t := range plan.Trains {
		var glyphs, chain []string
		current := false
		for _, ref := range t.Cars {
			c := cars[ref]
			g := carGlyph(c)
			glyphs = append(glyphs, g)
			label := ref.String()
			switch {
			case g == "◐" && !current:
				current = true
				claimant := c.Claimant
				chain = append(chain, fmt.Sprintf("→ %s %s %s @%s", label, c.Bar, c.Status, claimant))
			case g == "✓" || g == "◑" || g == "?":
				chain = append(chain, label+" "+g)
			default:
				chain = append(chain, label)
			}
		}
		slot := strings.TrimSpace(slotCellRE.Split(t.Slot, 2)[0])
		lines = append(lines, fmt.Sprintf("%-21s %-10s %-12s %s",
			runes(t.Name, 21), runes(slot, 10), strings.Join(glyphs, ""), strings.Join(chain, " ")))
	}
	for _, w := range plan.Warnings {
		lines = append(lines, "! "+w)
	}
	return strings.Join(append(lines, "```"), "\n")
}

