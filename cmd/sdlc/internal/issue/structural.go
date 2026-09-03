package issue

import (
	"regexp"
	"strconv"
	"strings"
)

// Section names the presence gates validate. Constants so the names are
// single-sourced within this file (ARCH-DRY) and pinned to the cue model by
// TestGatedSectionsSubsetOfModel. The gates' logic (word counts, checklist/bullet
// regex, related: fallback) is bespoke per section and intentionally NOT modeled
// in cue — only the section identity is.
const (
	secSpec     = "Spec"
	secDoneWhen = "Done when"
	secPlan     = "Plan"
)

// gatedSections is the set of sections CheckSectionsPresence enforces.
// INVARIANT (TestGatedSectionsSubsetOfModel): a subset of issue.cue
// scaffold.sections — a gate must not require a section the creation template
// never writes. (Note: checkPlan encodes "Plan" in PlanSectionBody, so a rename
// there needs a matching edit — the test fires to remind you.)
var gatedSections = []string{secSpec, secPlan, secDoneWhen}

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
//
// The estimate gate (`estimate-present`) is NOT part of this bundle: #113
// split it into the standalone CheckEstimate so `sdlc change-code` can run
// and bypass it independently (--no-estimate), honoring the per-gate
// --no-<gate> convention.
//
// Pure — no IO, deterministic on its input. Mirrors close.go's
// guard posture: small set of cheap checks, each clearly labelled
// so the operator can decide whether to fix or --force.
func CheckStructural(text string) []StructuralFailure {
	// Section PRESENCE is the shared policy (CheckSectionsPresence, also the
	// #124 instance-conformance validator's section overlay — single source, so
	// the change-code gate and the merge gate can't diverge). CheckStructural
	// adds the change-code-only ≥50-word Spec check on top.
	out := CheckSectionsPresence(text)
	if _, body, err := Parse(text); err == nil {
		if f := checkSpecWordCount(body); f != nil {
			out = append(out, *f)
		}
	}
	return out
}

// CheckSectionsPresence runs the PRESENCE-only section gates (well-formedness,
// no semantic quality). Shared by CheckStructural (change-code) and the #124
// validator's section overlay (`sdlc issue validate` + the pre-merge gate) so the
// required-section policy lives in ONE place. Empty return = all present.
//
//	frontmatter-present — issue has YAML frontmatter at all
//	spec-present        — ## Spec section exists
//	plan-present        — ## Plan has ≥ 1 non-empty checklist item
//	done-when-present   — ## Done when has ≥ 1 non-empty bullet OR `related:` populated
//
// Pure — no IO. Note `## Done when` is optional-with-`related:`-fallback (NOT a hard
// required section): #124's flat-required design was disproved against the corpus.
func CheckSectionsPresence(text string) []StructuralFailure {
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

	if f := checkSpecPresent(body); f != nil {
		out = append(out, *f)
	}
	if f := checkPlan(body); f != nil {
		out = append(out, *f)
	}
	if f := checkDoneWhen(fm, body); f != nil {
		out = append(out, *f)
	}
	return out
}

// CheckEstimate is the standalone estimate gate, split out of CheckStructural
// (#113) so `sdlc change-code` enforces + bypasses it as its own gate
// (--no-estimate). Returns the `estimate-present` failure when `estimate_hours:`
// is missing, empty, or not a positive number; nil when it's a positive number.
// Pure — parses the frontmatter itself so callers pass the full issue text,
// mirroring CheckStructural's signature.
func CheckEstimate(text string) *StructuralFailure {
	fm, _, err := Parse(text)
	if err != nil {
		return &StructuralFailure{
			Name:    "estimate-present",
			Message: "issue file has no YAML frontmatter to read `estimate_hours:` from",
		}
	}
	return checkEstimate(fm)
}

// checkSpecPresent — presence only (shared, well-formedness).
func checkSpecPresent(body string) *StructuralFailure {
	if _, ok := SectionBody(body, secSpec); !ok {
		return &StructuralFailure{
			Name:    "spec-present",
			Message: "no `## Spec` section found",
		}
	}
	return nil
}

// checkSpecWordCount — the ≥50-word SEMANTIC check, change-code-only (NOT part of
// CheckSectionsPresence). No-ops when ## Spec is absent (presence is reported by
// checkSpecPresent) so a missing Spec yields exactly one failure, not two.
func checkSpecWordCount(body string) *StructuralFailure {
	sec, ok := SectionBody(body, secSpec)
	if !ok {
		return nil
	}
	words := strings.Fields(stripCodeFences(sec))
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
	section, ok := PlanItemsBody(body)
	if !ok {
		return &StructuralFailure{
			Name:    "plan-present",
			Message: "no `## Plan` section found",
		}
	}
	if !nonEmptyPlanItemRE.MatchString(section) {
		return &StructuralFailure{
			Name:    "plan-present",
			Message: "`## Plan` has no non-empty checklist items (placeholders like `- [ ]` don't count)",
		}
	}
	return nil
}

var bulletRE = regexp.MustCompile(`(?m)^[-*]\s+\S`)

func checkDoneWhen(fm, body string) *StructuralFailure {
	// First try: ## Done when section with at least one non-empty bullet.
	if sec, ok := SectionBody(body, secDoneWhen); ok {
		if bulletRE.MatchString(sec) {
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

// stripCodeFences removes fenced code blocks from a markdown snippet so the Spec
// word count reflects prose, not embedded code.
//
// Since #211 M2 this is the shared scanner (StripFenced), not the naive
// `(?s)```.*?``` ` regex it used to be — so it now handles tilde fences, the
// closer-width rule, and indented fences, none of which the old one did. The
// stale "naive — doesn't handle nested fences or indented code" caveat and the
// blank line that detached this block from the function are gone with it.
//
// NOT built on SplitFences, deliberately: the two have different
// unterminated-fence policies. Here an unterminated tail stays in the output
// (counted as prose by the word-count gates — changing that would silently shift
// gate behavior); SplitFences classifies it Fenced (a rewriter must never touch
// the inside of a broken fence). TestUnterminatedPolicies_DisagreeOnPurpose pins
// that fork.
func stripCodeFences(s string) string {
	return StripFenced(s)
}

// FenceSegment is one run of markdown text, classified by whether it lies
// inside a ``` fence. Concatenating the Text of every segment reproduces
// the input byte-for-byte.
type FenceSegment struct {
	Text   string
	Fenced bool
}

// NOT rebased onto FenceSpans (#211 M2), deliberately — the one place this issue
// leaves two fence implementations standing, with the reason:
//
// SplitFences is CHARACTER-oriented, not line-oriented. Its contract includes
// inline pairs mid-line (`a```one``` mid ```two```z` is two fenced segments with
// prose between them, pinned by TestSplitFences) and byte-exact segment
// boundaries that fall inside a line. FenceSpans classifies whole LINES, which
// cannot express that without embedding a second scanner inside the first.
//
// It is also a different problem. This issue's class is "a heading-shaped line
// inside a fence read as structure"; SplitFences never looks for headings — it
// answers "may a rewriter edit these bytes". Merging them would change what
// `migrate` rewrites across repos to serve a tidiness the class doesn't need.
//
// SplitFences segments a markdown snippet into prose and fenced runs for
// rewriters that must skip code fences (#179 `sdlc migrate`). An
// unterminated trailing fence is classified Fenced — the conservative
// direction for a rewriter (see stripCodeFences's comment for why the
// word-count gate makes the opposite call).
func SplitFences(s string) []FenceSegment {
	var segs []FenceSegment
	rest := s
	for rest != "" {
		start := strings.Index(rest, "```")
		if start < 0 {
			segs = append(segs, FenceSegment{Text: rest})
			break
		}
		if start > 0 {
			segs = append(segs, FenceSegment{Text: rest[:start]})
		}
		end := strings.Index(rest[start+3:], "```")
		if end < 0 { // unterminated: the whole tail is fenced
			segs = append(segs, FenceSegment{Text: rest[start:], Fenced: true})
			break
		}
		stop := start + 3 + end + 3
		segs = append(segs, FenceSegment{Text: rest[start:stop], Fenced: true})
		rest = rest[stop:]
	}
	return segs
}
