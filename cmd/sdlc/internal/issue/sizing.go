package issue

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Sizing is the plan-derived complexity hint that drives the
// branching-strategy ask in `sdlc change-code`. All fields are cheap
// to compute from the issue file alone; no agent or git inspection.
//
// The Bucket label is mechanical (small / medium / large) — operator
// reads the numbers and decides whether worktree isolation is worth
// the setup cost.
type Sizing struct {
	EstimateHours float64
	PlanItems     int
	Milestones    int
	SpecWords     int
	RelatedFiles  int
	Bucket        Bucket
}

// Bucket is the coarse complexity label derived from Sizing's numeric
// fields. Thresholds are intentionally crude — the goal is "small
// enough to read at a glance," not nuanced sizing.
type Bucket string

const (
	BucketSmall  Bucket = "small"
	BucketMedium Bucket = "medium"
	BucketLarge  Bucket = "large"
)

// ComputeSizingFromContent reads an issue file's raw text and returns
// a Sizing. Missing frontmatter is treated as all-zeros (bucket falls
// to small) rather than an error — the structural-checks gate is the
// place to refuse on missing structure; sizing just describes.
//
// Pure: deterministic on input, no IO.
func ComputeSizingFromContent(text string) Sizing {
	s := Sizing{}

	fm, body, err := Parse(text)
	if err != nil {
		// No frontmatter — body is the whole text; estimate / related
		// fields stay zero. Plan + Spec extraction still works against
		// the raw body.
		body = text
	} else {
		if v, ok := GetField(fm, "estimate_hours"); ok {
			if n, err := strconv.ParseFloat(v, 64); err == nil {
				s.EstimateHours = n
			}
		}
		if v, ok := GetField(fm, "related"); ok {
			s.RelatedFiles = countFrontmatterList(v)
		}
	}

	if m := PlanSectionRE.FindStringSubmatchIndex(body); m != nil {
		section := body[m[2]:m[3]]
		s.PlanItems = len(PlanItemRE.FindAllStringIndex(section, -1))
		s.Milestones = len(milestoneLabelRE.FindAllStringIndex(section, -1))
	}

	if sec, ok := SectionBody(body, "Spec"); ok {
		s.SpecWords = len(strings.Fields(stripCodeFences(sec)))
	}

	s.Bucket = bucketFor(s)
	return s
}

// bucketFor classifies a Sizing into small / medium / large.
//
//	small  — estimate < 2h AND plan items ≤ 5 AND milestones == 0
//	large  — estimate ≥ 6h OR milestones ≥ 3
//	medium — everything else
//
// Boundaries are intentionally crisp; tied to the "operator decides"
// posture — the bucket is a hint, not a verdict.
func bucketFor(s Sizing) Bucket {
	if s.EstimateHours >= 6 || s.Milestones >= 3 {
		return BucketLarge
	}
	if s.EstimateHours < 2 && s.PlanItems <= 5 && s.Milestones == 0 {
		return BucketSmall
	}
	return BucketMedium
}

// milestoneLabelRE matches plan items that start with a milestone tag
// like `M1:`, `M4b:`, `M12c:`. Anchored against the plan-item bullet
// shape so a stray "M5" elsewhere doesn't count.
var milestoneLabelRE = regexp.MustCompile(`(?m)^- \[[ x.]\] (?:\*\*)?M\d+[a-z]?:`)

// countFrontmatterList counts comma-separated items inside a YAML
// inline list like `[a, b, c]`. Returns 0 for `[]` or unparseable
// input. Naive — doesn't handle nested brackets or quoted commas, but
// matches the project's flat-list convention.
func countFrontmatterList(v string) int {
	v = strings.TrimSpace(v)
	if v == "" || v == "[]" {
		return 0
	}
	// Strip surrounding [] if present.
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		v = v[1 : len(v)-1]
	}
	count := 0
	for _, p := range strings.Split(v, ",") {
		if strings.TrimSpace(p) != "" {
			count++
		}
	}
	return count
}

// Format renders a human-readable sizing summary, suitable for
// printing on stderr ahead of the branching ask.
func (s Sizing) Format(issueID, title string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Issue %s — %s\n", issueID, title)
	fmt.Fprintf(&b, "  estimate:      %sh\n", trimZero(s.EstimateHours))
	fmt.Fprintf(&b, "  plan items:    %d\n", s.PlanItems)
	fmt.Fprintf(&b, "  milestones:    %d\n", s.Milestones)
	fmt.Fprintf(&b, "  spec words:    %d\n", s.SpecWords)
	fmt.Fprintf(&b, "  related files: %d\n", s.RelatedFiles)
	fmt.Fprintf(&b, "  → %s\n", s.Bucket)
	return b.String()
}

// trimZero formats a float without trailing zeros so 1.5h prints as
// "1.5" and 2.0h prints as "2".
func trimZero(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
