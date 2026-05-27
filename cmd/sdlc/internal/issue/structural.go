package issue

import (
	"regexp"
	"strconv"
	"strings"
)

// StructuralFailure is one gate's verdict against an issue file's
// shape. Returned by CheckStructural; callers render the slice and
// refuse to proceed unless --force is set.
//
// Name is a short stable token (`spec-present`, `plan-present`, …)
// suitable for log/test pins; Message is the human-readable
// explanation including any specific numbers ("only 12 words; need
// ≥ 50").
type StructuralFailure struct {
	Name    string
	Message string
}

// CheckStructural runs the deterministic pre-implementation gates
// against an issue file's text. Empty return = all gates pass.
//
// Gates (each refusable with --force <reason>):
//
//	frontmatter-present — issue has YAML frontmatter at all
//	spec-present        — ## Spec section has ≥ 50 words
//	plan-present        — ## Plan has ≥ 1 non-empty checklist item
//	done-when-present   — ## Done when has ≥ 1 non-empty bullet,
//	                      OR `related:` frontmatter is populated
//	estimate-present    — estimate_hours: is a positive number
//
// Pure — no IO, deterministic on its input. Mirrors close.go's
// guard posture: small set of cheap checks, each clearly labelled
// so the operator can decide whether to fix or --force.
func CheckStructural(text string) []StructuralFailure {
	var out []StructuralFailure

	fm, body, err := Parse(text)
	if err != nil {
		// Without frontmatter the rest of the checks have nothing
		// to read — short-circuit with one decisive failure.
		return []StructuralFailure{{
			Name:    "frontmatter-present",
			Message: "issue file has no YAML frontmatter (expected `---\\n...\\n---\\n`)",
		}}
	}

	if f := checkSpec(body); f != nil {
		out = append(out, *f)
	}
	if f := checkPlan(body); f != nil {
		out = append(out, *f)
	}
	if f := checkDoneWhen(fm, body); f != nil {
		out = append(out, *f)
	}
	if f := checkEstimate(fm); f != nil {
		out = append(out, *f)
	}
	return out
}

// specSectionRE captures the body of the `## Spec` section, stopping
// at the next top-level heading or end-of-text.
var specSectionRE = regexp.MustCompile(`(?ms)^## Spec\s*\n(.*?)(?:^## |\z)`)

func checkSpec(body string) *StructuralFailure {
	m := specSectionRE.FindStringSubmatch(body)
	if m == nil {
		return &StructuralFailure{
			Name:    "spec-present",
			Message: "no `## Spec` section found",
		}
	}
	words := strings.Fields(stripCodeFences(m[1]))
	if len(words) < 50 {
		return &StructuralFailure{
			Name:    "spec-present",
			Message: "`## Spec` has " + strconv.Itoa(len(words)) + " words; need ≥ 50",
		}
	}
	return nil
}

// nonEmptyPlanItemRE matches `- [s] X` where X is non-empty after trim.
var nonEmptyPlanItemRE = regexp.MustCompile(`(?m)^- \[[ x.]\]\s+\S`)

func checkPlan(body string) *StructuralFailure {
	m := PlanSectionRE.FindStringSubmatchIndex(body)
	if m == nil {
		return &StructuralFailure{
			Name:    "plan-present",
			Message: "no `## Plan` section found",
		}
	}
	section := body[m[2]:m[3]]
	if !nonEmptyPlanItemRE.MatchString(section) {
		return &StructuralFailure{
			Name:    "plan-present",
			Message: "`## Plan` has no non-empty checklist items (placeholders like `- [ ]` don't count)",
		}
	}
	return nil
}

var doneWhenSectionRE = regexp.MustCompile(`(?ms)^## Done when\s*\n(.*?)(?:^## |\z)`)
var bulletRE = regexp.MustCompile(`(?m)^[-*]\s+\S`)

func checkDoneWhen(fm, body string) *StructuralFailure {
	// First try: ## Done when section with at least one non-empty bullet.
	if m := doneWhenSectionRE.FindStringSubmatch(body); m != nil {
		if bulletRE.MatchString(m[1]) {
			return nil
		}
	}
	// Fallback: related: frontmatter populated.
	if v, ok := GetField(fm, "related"); ok {
		v = strings.TrimSpace(v)
		// Reject empty `related: []` or just `related:`.
		if v != "" && v != "[]" {
			return nil
		}
	}
	return &StructuralFailure{
		Name:    "done-when-present",
		Message: "`## Done when` has no non-empty bullet AND `related:` frontmatter is empty",
	}
}

func checkEstimate(fm string) *StructuralFailure {
	v, ok := GetField(fm, "estimate_hours")
	if !ok || v == "" {
		return &StructuralFailure{
			Name:    "estimate-present",
			Message: "`estimate_hours:` is missing from frontmatter",
		}
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n <= 0 {
		return &StructuralFailure{
			Name:    "estimate-present",
			Message: "`estimate_hours: " + v + "` is not a positive number",
		}
	}
	return nil
}

// stripCodeFences removes fenced code blocks (```…```) from a markdown
// snippet so the Spec word count reflects prose, not embedded code.
// Naive — doesn't handle nested fences or indented code — but good
// enough for our gate purpose.
var fencedCodeRE = regexp.MustCompile("(?s)```.*?```")

func stripCodeFences(s string) string {
	return fencedCodeRE.ReplaceAllString(s, " ")
}
