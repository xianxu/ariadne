// family.go — finding FAMILIES and the convergence signal (ariadne#194 M3).
//
// A stateless reviewer cannot see that it is describing the same missing rule for the
// third time, so the author keeps being told to fix instances. Measured on tools#1: one
// missing rule — "a part-of-speech word only opens a block when structurally placed" —
// surfaced in four shapes across four rounds, and only the third round wrote the rule.
//
// A `family:` slug names the underlying RULE rather than the symptom. When a family
// already has a finding in a prior round, the next round's prompt escalates the
// recommendation from "fix this instance" to "state the rule that covers all of them".
//
// THREE MECHANISMS ANCHOR THE SLUG, because no one of them is sufficient (#194 D3):
//  1. the in-play vocabulary is rendered into the prompt with an instruction to REUSE an
//     existing slug — this is what catches synonyms;
//  2. NormalizeFamily folds casing and punctuation on ingest — and NOTHING else;
//  3. tests cover both, including a near-miss fixture and a true-synonym fixture whose
//     name records the residual risk rather than pretending it is solved.
package gatestate

import (
	"fmt"
	"sort"
	"strings"
)

// NormalizeFamily folds a family slug to its canonical form: lowercase, non-alphanumeric
// runs collapsed to single hyphens, edges trimmed. Pure.
//
// It catches `Block Opener Rule` and `block_opener_rule`. It does NOT catch
// `block-opener` vs `block-opener-rule` — a true synonym is a judgement, not a spelling,
// and mechanism 1 is what addresses those.
func NormalizeFamily(s string) string {
	var b strings.Builder
	lastHyphen := true // leading hyphens are trimmed by never emitting one first
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// FamilyCounts tallies findings per normalized family across the WHOLE ledger — not a
// boundary-filtered view. A family recurring across milestones is precisely the signal
// this exists to surface: tools#1's rule spanned M1's rounds and the close review, and a
// per-boundary count could not have seen it. Pure.
func FamilyCounts(l Ledger) map[string]int {
	counts := map[string]int{}
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if fam := NormalizeFamily(f.Family); fam != "" {
				counts[fam]++
			}
		}
	}
	return counts
}

// sortedFamilies returns the in-play families, most-recurrent first then alphabetical, so
// the rendered vocabulary is deterministic (a golden-pinned prompt cannot tolerate map
// iteration order). Pure.
func sortedFamilies(counts map[string]int) []string {
	out := make([]string, 0, len(counts))
	for fam := range counts {
		out = append(out, fam)
	}
	sort.Slice(out, func(i, j int) bool {
		if counts[out[i]] != counts[out[j]] {
			return counts[out[i]] > counts[out[j]]
		}
		return out[i] < out[j]
	})
	return out
}

// renderFamilyVocabulary is D3 mechanism 1: show the reviewer which families are already
// in play and tell it to reuse one. Without this the reviewer coins a fresh slug, every
// count stays 1, and the escalation below never fires — the feature would be inert while
// looking implemented. Pure; "" when no family has been named yet.
func renderFamilyVocabulary(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Families already in play on this issue — REUSE one of these slugs when a\n")
	b.WriteString("finding belongs to it, and coin a new slug only when it genuinely does not:\n\n")
	for _, fam := range sortedFamilies(counts) {
		fmt.Fprintf(&b, "  %-36s %s\n", fam, pluralFindings(counts[fam]))
	}
	return b.String()
}

// renderFamilyEscalation is the substance of #194 M3: on a family that already has prior
// findings, change what the reviewer is asked to do. Pure; "" when nothing repeats.
func renderFamilyEscalation(counts map[string]int) string {
	var repeats []string
	for _, fam := range sortedFamilies(counts) {
		if counts[fam] >= 1 {
			repeats = append(repeats, fam)
		}
	}
	if len(repeats) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nIf a finding you are about to raise belongs to one of those families, it is the\n")
	b.WriteString("Nth instance of a rule that has already been patched at least once. Say so\n")
	b.WriteString("explicitly and change your recommendation:\n\n")
	fmt.Fprintf(&b, "  > **This is the %s finding in family `%s`.** Earlier rounds fixed instances.\n",
		ordinal(counts[repeats[0]]+1), repeats[0])
	b.WriteString("  > Do NOT fix this instance — state the rule that covers all of them, and fix\n")
	b.WriteString("  > that. If the rule cannot be stated, say why, and record the family in\n")
	b.WriteString("  > `Limits` with its measured prevalence.\n")
	return b.String()
}

// ConvergenceLine is the one-line answer to "is this converging, or am I patching cases?"
// — the signal that was missing when tools#1 ran four rounds without anyone being able to
// tell whether round five would find more. Capping on finding COUNT is arbitrary; capping
// when families stop repeating is not. Pure.
func ConvergenceLine(l Ledger, round int) string {
	priorFamilies := map[string]bool{}
	newCount, repeats, disposed := 0, 0, 0
	for _, r := range l.Rounds {
		if r.N == round {
			continue
		}
		for _, f := range r.New {
			if fam := NormalizeFamily(f.Family); fam != "" {
				priorFamilies[fam] = true
			}
		}
	}
	for _, r := range l.Rounds {
		if r.N != round {
			continue
		}
		disposed += len(r.Dispositions)
		for _, f := range r.New {
			newCount++
			if fam := NormalizeFamily(f.Family); fam != "" && priorFamilies[fam] {
				repeats++
			}
		}
	}
	verdict := "**Converging.**"
	if repeats > 0 {
		verdict = "**Not converging: fix rules, not instances.**"
	}
	return fmt.Sprintf("round %d — %s, %s, %d disposed. %s",
		round, pluralFindings(newCount), pluralFamilies(repeats), disposed, verdict)
}

func pluralFindings(n int) string {
	if n == 1 {
		return "1 new finding"
	}
	return fmt.Sprintf("%d new findings", n)
}

func pluralFamilies(n int) string {
	if n == 1 {
		return "1 repeat family"
	}
	return fmt.Sprintf("%d repeat families", n)
}

func ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}
