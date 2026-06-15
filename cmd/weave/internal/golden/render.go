package golden

import (
	"fmt"
	"sort"
	"strings"
)

// Render formats one repo's divergence ledger as readable text: a header naming
// the repo, the divergence lines grouped by class (UNEXPECTED first — it's the
// failure evidence — then MATCH, then EXPECTED), and a one-line per-class
// summary. Pure (string in/out), so it's unit-tested directly; the subcommand
// just prints what this returns.
func Render(repo string, divs []Divergence) string {
	var counts [3]int
	for _, d := range divs {
		counts[d.Class]++
	}

	var b strings.Builder
	fmt.Fprintf(&b, "== golden-diff: %s ==\n", repo)

	// Order classes UNEXPECTED, MATCH, EXPECTED so failures are read first.
	for _, cls := range []Class{Unexpected, Match, Expected} {
		lines := linesFor(divs, cls)
		for _, line := range lines {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(&b, "summary: MATCH %d  EXPECTED %d  UNEXPECTED %d\n",
		counts[Match], counts[Expected], counts[Unexpected])
	return b.String()
}

// linesFor renders the divergences of one class as sorted "CLASS verb path —
// detail" lines (sorted by path for a deterministic ledger).
func linesFor(divs []Divergence, cls Class) []string {
	var out []string
	for _, d := range divs {
		if d.Class != cls {
			continue
		}
		line := fmt.Sprintf("  %-10s %-8s %s", d.Class, d.Verb, d.Path)
		if d.Detail != "" {
			line += " — " + d.Detail
		}
		out = append(out, line)
	}
	sort.Strings(out)
	return out
}
