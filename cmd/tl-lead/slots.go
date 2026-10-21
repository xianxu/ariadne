package main

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var issueBranchRE = regexp.MustCompile(`^(\d{6})-(.*)$`)

// SlotLine is what the `## Slots` renderer knows about one numbered slot.
type SlotLine struct {
	Slot  FleetSlot
	ID    int    // the issue its branch carries; 0 when on no issue branch
	Slug  string // the branch's slug after the id
	Phase string // the issue's card status, or "?" when its observation failed
	Bar   string
	Agent string // its agent, from couch; "?" unread
	Hold  string // "until X" when the TL holds it to soak a risky change (ariadne#482)
}

// holdRE finds a soak hold on a rendered Slots line: `repo:N … hold: soak until X`.
var holdRE = regexp.MustCompile(`^(\S+)\s.*\bhold: soak (until .+?)\s*$`)

// parseHolds reads the holds a previous render recorded in the section, by
// slot address. The section is the hold's one home: every render carries it.
func parseHolds(section string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(section, "\n") {
		if m := holdRE.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			out[m[1]] = m[2]
		}
	}
	return out
}

// applyHolds sets each line's hold from holds, after adding set (addr → "until
// X") and dropping release; it refuses an address no line has.
func applyHolds(lines []SlotLine, holds map[string]string, set map[string]string, release []string) error {
	known := map[string]bool{}
	for _, l := range lines {
		known[l.Slot.Address] = true
	}
	for addr, until := range set {
		if !known[addr] {
			return fmt.Errorf("--hold %s: no such numbered slot", addr)
		}
		holds[addr] = until
	}
	for _, addr := range release {
		if _, ok := holds[addr]; !ok {
			return fmt.Errorf("--release %s: it holds nothing", addr)
		}
		delete(holds, addr)
	}
	for i := range lines {
		lines[i].Hold = holds[lines[i].Slot.Address]
	}
	return nil
}

// agentLabel renders a slot's agent, `none` when no agent runs there, or `?`
// when Couch was not read. Resting describes Git, not runtime readiness (#533).
func agentLabel(actors map[string]Actor, err error, addr string) string {
	if err != nil {
		return "?"
	}
	a, ok := actors[addr]
	switch {
	case !ok:
		return "none"
	}
	return a.Agent
}

// slotLines classifies every numbered slot (N ≥ 1): on an issue branch, idle
// on its resting branch, or on some other branch. :0 is the integration
// checkout, never a work slot.
func slotLines(fleet Fleet) []SlotLine {
	var out []SlotLine
	for _, s := range fleet.Slots {
		if s.Slot < 1 {
			continue
		}
		l := SlotLine{Slot: s}
		if m := issueBranchRE.FindStringSubmatch(s.HostBranch); m != nil {
			l.ID, _ = strconv.Atoi(m[1])
			l.Slug = m[2]
		}
		out = append(out, l)
	}
	return out
}

// renderSlots renders `## Slots`: one line per busy slot (its issue, phase,
// agent, bar and slug), then one line naming the free slots and their agents.
// A held slot is never free: it gets its own line, ending in its hold.
func renderSlots(lines []SlotLine) string {
	var busy, free []string
	for _, l := range lines {
		s := l.Slot
		switch {
		case l.Hold != "" && l.ID == 0:
			busy = append(busy, fmt.Sprintf("%-14s %-12s hold: soak %s", s.Address, l.Agent, l.Hold))
			continue
		case l.ID != 0:
			ref := fmt.Sprintf("%s#%d", s.Repo, l.ID)
			line := fmt.Sprintf("%-14s %-16s %-13s %-12s %s  %s", s.Address, ref, l.Phase, l.Agent, l.Bar, runes(l.Slug, 40))
			if l.Hold != "" {
				line += "  hold: soak " + l.Hold
			}
			busy = append(busy, line)
		case s.HostBranch == s.RestingBranch:
			free = append(free, s.Address+" ("+l.Agent+")")
		case s.HostBranch == "":
			busy = append(busy, fmt.Sprintf("%-14s detached  %s", s.Address, l.Agent))
		default:
			busy = append(busy, fmt.Sprintf("%-14s on %s  %s", s.Address, runes(s.HostBranch, 40), l.Agent))
		}
	}
	out := []string{"```"}
	out = append(out, busy...)
	if len(free) > 0 {
		out = append(out, "free: "+strings.Join(free, ", "))
	}
	if len(busy) == 0 && len(free) == 0 {
		out = append(out, "no numbered slots")
	}
	return strings.Join(append(out, "```"), "\n")
}

// gatherSlots reads `## Slots`' facts from the fleet inventory, couch's
// actors (each slot's agent), each busy slot's issue observation (its phase)
// and its checkout's issue file (its bar). A failed actors read renders `?`.
// The holds come from the section being replaced, at write time.
func gatherSlots(sd SDLC, ca CouchActors, vantage string) ([]SlotLine, error) {
	lines, _, _, err := gatherScopedSlots(sd, ca, vantage, false)
	return lines, err
}

// gatherScopedSlots selects before any per-issue observation or file read.
// Diagnostics are separate from rows so a failed actor read is reported once.
func gatherScopedSlots(sd SDLC, ca CouchActors, vantage string, all bool) ([]SlotLine, []string, hubScope, error) {
	scope, err := readHubScope(vantage)
	if err != nil {
		return nil, nil, scope, err
	}
	fleet, err := sd.FleetInventory(vantage)
	if err != nil {
		return nil, nil, scope, err
	}
	var lines []SlotLine
	var notes []string
	var unclaimed, foreign, tied int
	for _, l := range slotLines(fleet) {
		if scope.enabled && !scope.owns(l.Slot) {
			o := scope.owner(l.Slot)
			switch {
			case o.Unclaimed:
				unclaimed++
			case len(o.Owners) > 1:
				tied++
				for _, owner := range o.Owners {
					if owner == scope.self {
						notes = append(notes, fmt.Sprintf("ownership: %s tied between %s; not owned", l.Slot.Address, strings.Join(o.Owners, ", ")))
						break
					}
				}
			default:
				foreign++
			}
			if !all {
				continue
			}
		}
		lines = append(lines, l)
	}
	if scope.enabled && !all {
		notes = append(notes, fmt.Sprintf("excluded: %d unclaimed, %d foreign, %d tied (use --all to list)", unclaimed, foreign, tied))
	}
	actors, aerr := readActors(ca)
	if aerr != nil {
		notes = append(notes, "couch --actors unreadable: "+aerr.Error())
	}
	for i := range lines {
		lines[i].Agent = agentLabel(actors, aerr, lines[i].Slot.Address)
	}
	var mu sync.Mutex
	boundedEach(len(lines), func(i int) {
		l := lines[i]
		if l.ID == 0 {
			return
		}
		phase := "?"
		if o, err := sd.IssueShow(l.Slot.HostPath, l.ID); err == nil {
			phase = o.Status
		}
		bar := issueBar(l.Slot.HostPath, l.ID)
		mu.Lock()
		lines[i].Phase, lines[i].Bar = phase, bar
		mu.Unlock()
	})
	return lines, notes, scope, nil
}

// guardScopedHolds runs inside the guarded write's recomputation, against
// the latest section. A hidden hold must be explicitly released or retained.
func guardScopedHolds(lines []SlotLine, holds map[string]string, release []string) error {
	known := map[string]bool{}
	for _, l := range lines {
		known[l.Slot.Address] = true
	}
	for _, addr := range release {
		known[addr] = true
	}
	var hidden []string
	for addr := range holds {
		if !known[addr] {
			hidden = append(hidden, addr)
		}
	}
	sort.Strings(hidden)
	if len(hidden) > 0 {
		return fmt.Errorf("refused: scoped Slots would discard held %s; use --all or explicitly --release each hold", strings.Join(hidden, ", "))
	}
	return nil
}

func readActors(ca CouchActors) (map[string]Actor, error) {
	if ca == nil {
		return nil, errors.New("no couch actors seam")
	}
	return ca.Actors()
}
