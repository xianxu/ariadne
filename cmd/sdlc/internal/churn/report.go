package churn

import (
	"strconv"
	"strings"
)

// FileStat is one `git diff --numstat` row reduced to what the metric needs: the path
// and its insertions. Deletions are parsed and dropped deliberately — rework is defined
// as lines WRITTEN over lines KEPT, and folding deletions in would make a clean deletion
// of dead code read as churn.
type FileStat struct {
	Path       string
	Insertions int
}

// Buckets is the four-way insertion split of one diff.
type Buckets struct {
	CodeProd int
	CodeTest int
	Atlas    int
	Workshop int
}

// Report is the churn picture of one work window.
type Report struct {
	// Final is the four-bucket split of the window's NET diff (base..HEAD) — the lines
	// that survived to be reviewed.
	Final Buckets
	// FinalTotal is Final summed: the lines the window actually landed.
	FinalTotal int
	// Rework is insertions-across-the-window's-commits over FinalTotal. 1.0 means every
	// line written survived; 3.0 means the window wrote three lines for every one it
	// kept. 0 means undefined (no lines landed) — never Inf or NaN.
	Rework float64
}

// Summarize buckets the final diff and derives the rework multiple.
//
// commitInsertions is the sum across the window's COMMITS (the same file counted once
// per commit that touched it), which is what makes the ratio meaningful: rewriting one
// file five times is invisible in the final diff and is the waste this measures.
func Summarize(final []FileStat, commitInsertions int) Report {
	var r Report
	for _, s := range final {
		switch ClassifyPath(s.Path) {
		case Atlas:
			r.Final.Atlas += s.Insertions
		case Workshop:
			r.Final.Workshop += s.Insertions
		case CodeTest:
			r.Final.CodeTest += s.Insertions
		default:
			r.Final.CodeProd += s.Insertions
		}
	}
	r.FinalTotal = r.Final.CodeProd + r.Final.CodeTest + r.Final.Atlas + r.Final.Workshop
	// Guard the divisor, not the dividend: a pure-deletion or empty window has a real
	// commitInsertions and a zero FinalTotal, and +Inf (or 0/0's NaN) in a TSV column
	// poisons every downstream reader of the calibration ledger.
	if r.FinalTotal > 0 {
		r.Rework = float64(commitInsertions) / float64(r.FinalTotal)
	}
	return r
}

// ParseNumstat parses `git diff --numstat` / `git log --numstat --format=` output.
//
// Tolerant by design, because it consumes both shapes: log output interleaves blank
// lines between commits, and binary files render as `-\t-\tpath`. A row it cannot read
// is SKIPPED rather than erroring — but note the asymmetry that makes tolerance the
// right call here and dishonest elsewhere: this feeds a descriptive ratio, so a dropped
// row understates rework, whereas erroring out would cost a close its whole report.
func ParseNumstat(out string) []FileStat {
	var stats []FileStat
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}
		ins, err := strconv.Atoi(fields[0])
		if err != nil {
			continue // binary ("-"), or a header line that is not a numstat row
		}
		path := resolveRenamePath(strings.TrimSpace(fields[len(fields)-1]))
		if path == "" {
			continue
		}
		stats = append(stats, FileStat{Path: path, Insertions: ins})
	}
	return stats
}

// TotalInsertions sums insertions across stats WITHOUT collapsing by path. The
// no-collapse part is the point: on the commit side a file rewritten three times appears
// three times, and deduplicating would erase exactly the rework this package measures.
func TotalInsertions(stats []FileStat) int {
	total := 0
	for _, s := range stats {
		total += s.Insertions
	}
	return total
}

// resolveRenamePath rewrites numstat's rename notation to the DESTINATION path. Git renders
// a rename as `prefix{old => new}suffix` (or bare `old => new` when the paths share no
// prefix), so the raw field is not a path at all.
//
// It matters for bucketing, not tidiness: a cross-top-level rename renders as
// `{atlas => docs}/x.md`, whose first path segment is the literal `{atlas` — so
// ClassifyPath's segment rule, which is otherwise exactly right, would miss it and bucket
// map churn as production code. Resolving here rather than in ClassifyPath means every
// consumer of FileStat.Path gets a real path.
func resolveRenamePath(field string) string {
	const arrow = " => "
	if !strings.Contains(field, arrow) {
		return field
	}
	// openBrace/closeBrace, not open/close — `close` is a builtin, and the review that
	// found this function's bug also flagged a shadowed builtin two files over.
	openBrace := strings.Index(field, "{")
	closeBrace := strings.Index(field, "}")
	if openBrace < 0 || closeBrace < openBrace {
		// Bare form: `old.go => new.go` — the destination is everything after the arrow.
		return strings.TrimSpace(field[strings.Index(field, arrow)+len(arrow):])
	}
	inner := field[openBrace+1 : closeBrace]
	i := strings.Index(inner, arrow)
	if i < 0 {
		return field
	}
	// `prefix` + `new` + `suffix`, where an empty `new` collapses a doubled separator
	// (`workshop/{ => history}/x` vs `workshop/{history => }/x`).
	rebuilt := field[:openBrace] + strings.TrimSpace(inner[i+len(arrow):]) + field[closeBrace+1:]
	return strings.ReplaceAll(rebuilt, "//", "/")
}
