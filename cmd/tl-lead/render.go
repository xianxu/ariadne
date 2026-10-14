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

var issueBranchRE = regexp.MustCompile(`^(\d{6})-(.*)$`)

// OffTrainSlot is a numbered slot's state for `## Slots (off-train)`.
type OffTrainSlot struct {
	Slot FleetSlot
	ID   int    // the issue its branch carries; 0 when on no issue branch
	Slug string // the branch's slug after the id
	Bar  string
}

// offTrainSlots picks the slots to list: numbered slots (N ≥ 1) on an issue
// branch that no train lists, and numbered slots idle on their resting branch.
// bars maps a slot address to its issue's bar.
func offTrainSlots(fleet Fleet, listed map[trains.IssueRef]bool, bars map[string]string) []OffTrainSlot {
	var out []OffTrainSlot
	for _, s := range fleet.Slots {
		if s.Slot < 1 {
			continue
		}
		if m := issueBranchRE.FindStringSubmatch(s.HostBranch); m != nil {
			var id int
			fmt.Sscanf(m[1], "%d", &id)
			if listed[trains.IssueRef{Repo: s.Repo, ID: id}] {
				continue
			}
			out = append(out, OffTrainSlot{Slot: s, ID: id, Slug: m[2], Bar: bars[s.Address]})
			continue
		}
		if s.HostBranch == s.RestingBranch {
			out = append(out, OffTrainSlot{Slot: s})
		}
	}
	return out
}

// renderOffTrain renders `## Slots (off-train)`, or `_None._` when empty.
func renderOffTrain(slots []OffTrainSlot) string {
	if len(slots) == 0 {
		return "_None._"
	}
	lines := []string{"```"}
	for _, o := range slots {
		if o.ID == 0 {
			lines = append(lines, fmt.Sprintf("%-14s idle", o.Slot.Address))
			continue
		}
		ref := fmt.Sprintf("%s#%d", o.Slot.Repo, o.ID)
		lines = append(lines, fmt.Sprintf("%-14s %-16s %s  %s", o.Slot.Address, ref, o.Bar, runes(o.Slug, 40)))
	}
	return strings.Join(append(lines, "```"), "\n")
}
