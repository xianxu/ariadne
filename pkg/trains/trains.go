package trains

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// trains.go — the pure parser for a project's `## Trains` section (#329),
// shared by cmd/sdlc and cmd/tl-lead (#350). A
// train is a chain of issues in one code area, run in order by one slot; the
// TL authors the table and the cross-train edge list by hand, so the grammar
// is forgiving: a malformed row is skipped with a warning, never a panic.
//
//	| Train | Cars, in order | Slot |
//	|---|---|---|
//	| Main integration | ariadne#183 → #269 → #303 | ariadne:4 |
//
//	- ariadne#183 and ariadne#320 ⇒ ariadne#269/#303 (why …)
//	- ariadne#183 ~ ariadne#234: overlap, coordinated by the TL
//
// Cars: every ref in the cars cell, in order; a bare `#N` inherits the repo of
// the ref before it (the project's repo when there is none). Parenthesized text
// is an annotation and is stripped first. Every separator (→ ⇒ / ,) means "next
// car": one slot runs a train serially. Edges: a bullet line holding ⇒ splits
// into segments; every ref of a segment is a prerequisite of every ref of the
// next; prose after a colon in the last segment is the reason, not refs. `~`
// lines are overlap notes the scheduler ignores.

// IssueRef is a normalized `repo#N` (ID as an int, so #0329 == #329).
type IssueRef struct {
	Repo string
	ID   int
}

func (r IssueRef) String() string { return fmt.Sprintf("%s#%d", r.Repo, r.ID) }

// Train is one table row: its name, its cars in run order, and the slot
// allocation text.
type Train struct {
	Name, Slot string
	Cars       []IssueRef
}

// Edge is a hard cross-train dependency: From must land before To starts.
type Edge struct{ From, To IssueRef }

// TrainPlan is the parsed Trains section.
type TrainPlan struct {
	Trains   []Train
	Edges    []Edge
	Warnings []string
}

var (
	trainRefRE   = regexp.MustCompile(`(?:([A-Za-z][A-Za-z0-9_.\-]*))?#(\d+)\b`)
	trainParenRE = regexp.MustCompile(`\([^()]*\)`)
)

// stripParens removes (possibly nested) parenthesized annotations.
func stripParens(s string) string {
	for {
		next := trainParenRE.ReplaceAllString(s, "")
		if next == s {
			return s
		}
		s = next
	}
}

// parseRefs extracts the refs of text in order. last carries the inherited
// repo across calls (a bare `#N` takes it); it is updated in place.
func parseRefs(text string, last *string) []IssueRef {
	var out []IssueRef
	for _, m := range trainRefRE.FindAllStringSubmatch(stripParens(text), -1) {
		repo := m[1]
		if repo == "" {
			repo = *last
		}
		id, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		*last = repo
		out = append(out, IssueRef{Repo: repo, ID: id})
	}
	return out
}

// splitRow splits a markdown table row into trimmed cells (outer pipes dropped).
func splitRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	cells := strings.Split(line, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

func isSeparatorRow(cells []string) bool {
	for _, c := range cells {
		if strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}

// ParseTrains parses a `## Trains` section body. defaultRepo is the project's
// own repo, used for a bare `#N` with nothing before it to inherit from.
func ParseTrains(body, defaultRepo string) TrainPlan {
	var p TrainPlan
	nameCol, carsCol, slotCol := -1, -1, -1
	inTable := false
	for i, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "|"):
			cells := splitRow(line)
			if !inTable {
				nameCol, carsCol, slotCol = -1, -1, -1
				for j, c := range cells {
					switch {
					case strings.HasPrefix(c, "Train"):
						nameCol = j
					case strings.HasPrefix(c, "Cars"):
						carsCol = j
					case strings.HasPrefix(c, "Slot"):
						slotCol = j
					}
				}
				inTable = nameCol >= 0 && carsCol >= 0
				continue
			}
			if isSeparatorRow(cells) {
				continue
			}
			if carsCol >= len(cells) || nameCol >= len(cells) {
				p.Warnings = append(p.Warnings, fmt.Sprintf("trains line %d: row has %d cells, skipped", i+1, len(cells)))
				continue
			}
			last := defaultRepo
			t := Train{Name: cells[nameCol], Cars: parseRefs(cells[carsCol], &last)}
			if slotCol >= 0 && slotCol < len(cells) {
				t.Slot = cells[slotCol]
			}
			if t.Name == "" || len(t.Cars) == 0 {
				p.Warnings = append(p.Warnings, fmt.Sprintf("trains line %d: no train name or no cars, skipped", i+1))
				continue
			}
			p.Trains = append(p.Trains, t)
		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* "):
			inTable = false
			if !strings.Contains(line, "⇒") {
				continue
			}
			last := defaultRepo
			var prev []IssueRef
			segs := strings.Split(line[2:], "⇒")
			// Prose after a colon in the last segment is the edge's reason,
			// not more prerequisites ("⇒ ariadne#309: blocks pair#12 …").
			if c := strings.Index(segs[len(segs)-1], ":"); c >= 0 {
				segs[len(segs)-1] = segs[len(segs)-1][:c]
			}
			for k, seg := range segs {
				refs := parseRefs(seg, &last)
				if k > 0 {
					if len(prev) == 0 || len(refs) == 0 {
						p.Warnings = append(p.Warnings, fmt.Sprintf("trains line %d: ⇒ without a ref on both sides, skipped", i+1))
						break
					}
					for _, from := range prev {
						for _, to := range refs {
							p.Edges = append(p.Edges, Edge{From: from, To: to})
						}
					}
				}
				prev = refs
			}
		default:
			inTable = false
		}
	}
	return p
}

// Locate returns the train holding ref and the car's 1-based position, or
// ok=false when no train carries it.
func (p TrainPlan) Locate(ref IssueRef) (train Train, car int, ok bool) {
	for _, t := range p.Trains {
		for i, c := range t.Cars {
			if c == ref {
				return t, i + 1, true
			}
		}
	}
	return Train{}, 0, false
}
