package activetime

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BranchInterval describes a witnessed half-open worktree branch interval.
// Known false includes detached, unavailable and unsupported history.
type BranchInterval struct {
	Start, End time.Time
	Branch     string
	Known      bool
}

// BranchClaim is one pinned issue engagement for the selected slot.
type BranchClaim struct {
	Start, End time.Time
	Issue      string
	// Terminal keeps attribution identity but holds no reservation, as defined
	// by the issue vocabulary. Codecomplete is not terminal.
	Terminal bool
}

type BranchNote struct {
	Start, End time.Time
	Kind       string
	Branch     string
}

type BranchSelection struct {
	Windows []ActivityWindow
	Notes   []BranchNote
	steps   int
}

// BranchIssue recognizes the existing canonical NNNNNN-slug branch convention.
// A citation or an arbitrary branch suffix is never an issue identifier.
func BranchIssue(branch string) string {
	if len(branch) < 8 || branch[6] != '-' || strings.ContainsAny(branch, " \t\r\n\x00") {
		return ""
	}
	for _, c := range branch[:6] {
		if c < '0' || c > '9' {
			return ""
		}
	}
	n, err := strconv.Atoi(branch[:6])
	if err != nil || n <= 0 {
		return ""
	}
	return strconv.Itoa(n)
}

type branchEdge struct {
	at    time.Time
	issue string
	delta int
	held  bool
}

// SelectBranchActivity never widens prior source admission. Claims are
// reservations; a witnessed branch wins among concurrent reservations. Unknown
// branch history can retain sole-claim planning, but cannot choose two owners.
// Sorting once and sweeping all edges avoids a scan of every claim per event.
func SelectBranchActivity(issue string, windows []ActivityWindow, history []BranchInterval, claims []BranchClaim) (BranchSelection, error) {
	var out BranchSelection
	windows = slices.Clone(windows)
	history = slices.Clone(history)
	sort.Slice(windows, func(i, j int) bool { return windows[i].Start.Before(windows[j].Start) })
	sort.Slice(history, func(i, j int) bool { return history[i].Start.Before(history[j].Start) })
	var edges []branchEdge
	for i, w := range windows {
		if w.Start.IsZero() || !w.End.After(w.Start) || i > 0 && w.Start.Before(windows[i-1].End) {
			return out, fmt.Errorf("invalid or overlapping branch activity windows")
		}
		edges = append(edges, branchEdge{at: w.Start}, branchEdge{at: w.End})
	}
	for i, b := range history {
		if b.Start.IsZero() || !b.End.After(b.Start) || i > 0 && b.Start.Before(history[i-1].End) {
			return out, fmt.Errorf("invalid or overlapping branch history")
		}
		edges = append(edges, branchEdge{at: b.Start}, branchEdge{at: b.End})
	}
	for _, c := range claims {
		if c.Start.IsZero() || !c.End.After(c.Start) || c.Issue == "" {
			return out, fmt.Errorf("invalid branch claim interval")
		}
		edges = append(edges, branchEdge{at: c.Start, issue: c.Issue, delta: 1, held: !c.Terminal}, branchEdge{at: c.End, issue: c.Issue, delta: -1, held: !c.Terminal})
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].at.Before(edges[j].at) })
	active := map[string]int{}
	reserved := map[string]int{}
	wi, bi := 0, 0
	for i := 0; i < len(edges); {
		at := edges[i].at
		j := i
		for j < len(edges) && edges[j].at.Equal(at) {
			out.steps++
			if e := edges[j]; e.issue != "" {
				active[e.issue] += e.delta
				if active[e.issue] == 0 {
					delete(active, e.issue)
				}
				if e.held {
					reserved[e.issue] += e.delta
					if reserved[e.issue] == 0 {
						delete(reserved, e.issue)
					}
				}
			}
			j++
		}
		i = j
		if j == len(edges) {
			break
		}
		end := edges[j].at
		for wi < len(windows) && !windows[wi].End.After(at) {
			out.steps++
			wi++
		}
		for bi < len(history) && !history[bi].End.After(at) {
			out.steps++
			bi++
		}
		if wi == len(windows) || at.Before(windows[wi].Start) {
			continue
		}
		if active[issue] <= 0 {
			return out, fmt.Errorf("branch selection has no pinned claim for issue %s during %s → %s", issue, at.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))
		}
		branch, branchID := "", ""
		if bi < len(history) && !at.Before(history[bi].Start) && history[bi].Known {
			branch = history[bi].Branch
			branchID = BranchIssue(branch)
		}
		kind := "branch"
		switch {
		case branchID == issue:
		case branchID != "" && active[branchID] > 0:
			out.Notes = appendBranchNote(out.Notes, BranchNote{Start: at, End: end, Kind: "foreign-branch", Branch: branch})
			continue
		case len(reserved) == 1 && reserved[issue] > 0:
			kind = "ownership-only"
		default:
			return out, fmt.Errorf("concurrent claims have unknown, resting or ambiguous active branch during %s → %s", at.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))
		}
		part := windows[wi]
		part.Start, part.End = at, end
		last := len(out.Windows) - 1
		if last >= 0 && out.Windows[last].End.Equal(at) && out.Windows[last].Began.Equal(part.Began) {
			out.Windows[last].End = end
		} else {
			out.Windows = append(out.Windows, part)
		}
		out.Notes = appendBranchNote(out.Notes, BranchNote{Start: at, End: end, Kind: kind, Branch: branch})
	}
	return out, nil
}

func appendBranchNote(notes []BranchNote, note BranchNote) []BranchNote {
	last := len(notes) - 1
	if last >= 0 && notes[last].End.Equal(note.Start) && notes[last].Kind == note.Kind && notes[last].Branch == note.Branch {
		notes[last].End = note.End
		return notes
	}
	return append(notes, note)
}
