// Shared noun-model helpers (ariadne#180). Every noun binding (issue, verdict,
// project) parses the same JSON shapes — categories, when-glosses, lifecycle
// edges — and needs the same predicates over them. Extracted at the third noun:
// vocab.go and verdict.go had already duplicated inCategory once, and the
// project model would have triplicated the lifecycle predicates too.
package vocab

import (
	"fmt"
	"strings"
)

// inCat reports whether s is a member of cats[cat]. Shared by every noun model.
func inCat(cats map[string][]string, cat, s string) bool {
	for _, v := range cats[cat] {
		if v == s {
			return true
		}
	}
	return false
}

// canTransition reports whether l declares a from→to edge.
func canTransition(l []Transition, from, to string) bool {
	for _, t := range l {
		if t.From == from && t.To == to {
			return true
		}
	}
	return false
}

func transitionForEvent(l []Transition, from, event string) *Transition {
	for i := range l {
		if l[i].From == from && l[i].Event == event {
			return &l[i]
		}
	}
	return nil
}

func firstTransitionForEvent(l []Transition, event string) *Transition {
	for i := range l {
		if l[i].Event == event {
			return &l[i]
		}
	}
	return nil
}

func isEventTarget(l []Transition, status, event string) bool {
	for _, tr := range l {
		if tr.To == status && tr.Event == event {
			return true
		}
	}
	return false
}

// legalTransitions returns from's legal targets in lifecycle order, de-duplicated.
// Empty when from is unknown or a true dead-end.
func legalTransitions(l []Transition, from string) []string {
	var out []string
	seen := map[string]bool{}
	for _, t := range l {
		if t.From == from && !seen[t.To] {
			out = append(out, t.To)
			seen[t.To] = true
		}
	}
	return out
}

// allStatuses concatenates cats in the given category order.
func allStatuses(cats map[string][]string, order []string) []string {
	var out []string
	for _, cat := range order {
		out = append(out, cats[cat]...)
	}
	return out
}

// renderLifecycleHelp renders the full lifecycle reference for any noun: a
// STATUSES section (each status + its one-line `when` semantics) and a LEGAL
// TRANSITIONS section (each status → its legal targets). 2-space indented to
// match the help style. Pure.
func renderLifecycleHelp(statuses []string, when map[string]string, l []Transition) string {
	width := 0
	for _, s := range statuses {
		if len(s) > width {
			width = len(s)
		}
	}
	var b strings.Builder
	b.WriteString("STATUSES\n\n")
	for _, s := range statuses {
		b.WriteString(fmt.Sprintf("  %-*s  %s\n", width, s, when[s]))
	}
	b.WriteString("\nLEGAL TRANSITIONS\n\n")
	for _, s := range statuses {
		targets := legalTransitions(l, s)
		if len(targets) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-*s  → %s\n", width, s, strings.Join(targets, ", ")))
	}
	return strings.TrimRight(b.String(), "\n")
}
