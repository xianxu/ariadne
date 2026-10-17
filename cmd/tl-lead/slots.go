package main

import (
	"errors"
	"fmt"
	"regexp"
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
	Agent string // its agent and whether it is idle, from couch; "?" unread
}

// agentLabel renders a slot's actor: `claude`, `codex·idle`, `none` when no
// agent runs there, `?` when couch was not read.
func agentLabel(actors map[string]Actor, err error, addr string) string {
	if err != nil {
		return "?"
	}
	a, ok := actors[addr]
	switch {
	case !ok:
		return "none"
	case a.Resting:
		return a.Agent + "·idle"
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
func renderSlots(lines []SlotLine) string {
	var busy, free []string
	for _, l := range lines {
		s := l.Slot
		switch {
		case l.ID != 0:
			ref := fmt.Sprintf("%s#%d", s.Repo, l.ID)
			busy = append(busy, fmt.Sprintf("%-14s %-16s %-13s %-12s %s  %s", s.Address, ref, l.Phase, l.Agent, l.Bar, runes(l.Slug, 40)))
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

// gatherSlots renders `## Slots` from the fleet inventory, couch's actors
// (each slot's agent), each busy slot's issue observation (its phase) and its
// checkout's issue file (its bar). A failed actors read renders `?`.
func gatherSlots(sd SDLC, ca CouchActors, vantage string) (string, error) {
	fleet, err := sd.FleetInventory(vantage)
	if err != nil {
		return "", err
	}
	lines := slotLines(fleet)
	actors, aerr := readActors(ca)
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
	return renderSlots(lines), nil
}

func readActors(ca CouchActors) (map[string]Actor, error) {
	if ca == nil {
		return nil, errors.New("no couch actors seam")
	}
	return ca.Actors()
}
