package churn

import "testing"

func TestClassifyPath(t *testing.T) {
	cases := map[string]Bucket{
		"cmd/sdlc/changecode.go":                CodeProd,
		"cmd/sdlc/changecode_test.go":           CodeTest,
		"cmd/sdlc/internal/judge/testdata/x.md": CodeTest,
		"pkg/vocab/finding.go":                  CodeProd,
		"atlas/index.md":                        Atlas,
		"atlas/workflow/gate-state.md":          Atlas,
		"workshop/issues/000187-x.md":           Workshop,
		"workshop/plans/000187-x-plan.md":       Workshop,

		// Embedded prompt/helptext markdown is PRODUCTION here — it ships inside the
		// binary via //go:embed and is exactly the surface #187 changes. Counting it as
		// prose would understate the code this repo actually writes.
		"cmd/sdlc/internal/judge/prompts/plan-quality.md": CodeProd,
		"cmd/sdlc/helptext/change-code.md":                CodeProd,
		"construct/vocabulary/finding.cue":                CodeProd,
		"AGENTS.base.md":                                  CodeProd,

		// The DEFAULT bucket, named explicitly rather than left to whichever switch arm
		// happens to be last. Build/config/meta files are production artifacts of the
		// repo: they are versioned, reviewed, and break the build when wrong. Routing
		// them to code-prod is a decision, and a lockfile-sized diff landing there must
		// be a visible choice rather than an accident.
		"go.mod":                   CodeProd,
		"go.sum":                   CodeProd,
		"Makefile.workflow":        CodeProd,
		".github/workflows/ci.yml": CodeProd,
		"construct/base.manifest":  CodeProd,
		"docs/vision/roadmap.md":   CodeProd,
	}
	for path, want := range cases {
		if got := ClassifyPath(path); got != want {
			t.Errorf("ClassifyPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// The rule is ORDERED, and the order is observable: a test file under workshop/ is
// workshop churn, not test churn. Pinning it here means a later reader who reorders
// the switch for tidiness gets a failure rather than a silently reshuffled metric.
func TestClassifyPathRuleOrderIsPrefixFirst(t *testing.T) {
	for path, want := range map[string]Bucket{
		"workshop/issues/x_test.go":    Workshop,
		"atlas/testdata/sample.md":     Atlas,
		"cmd/sdlc/testdata/fuzz/seed":  CodeTest,
		"cmd/sdlc/internal/churn/x.go": CodeProd,
	} {
		if got := ClassifyPath(path); got != want {
			t.Errorf("ClassifyPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// The prefix rules match a leading path SEGMENT, not a substring: a repo that grows a
// `docs/atlas/` or an `atlasctl/` directory must not have those counted as atlas churn.
func TestClassifyPathPrefixIsSegmentNotSubstring(t *testing.T) {
	for path, want := range map[string]Bucket{
		"atlasctl/main.go":      CodeProd,
		"docs/atlas/vision.md":  CodeProd,
		"workshopping/notes.md": CodeProd,
	} {
		if got := ClassifyPath(path); got != want {
			t.Errorf("ClassifyPath(%q) = %v, want %v", path, got, want)
		}
	}
}
