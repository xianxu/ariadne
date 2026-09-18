package churn

import (
	"strings"
	"testing"
)

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

// TestIsDoc is #177's docs rule, per path: *.md anywhere, or anything under
// workshop/, atlas/, docs/. Everything else is code surface.
func TestIsDoc(t *testing.T) {
	for path, want := range map[string]bool{
		"README.md":                   true,
		"cmd/sdlc/helptext/close.md":  true, // the per-path DOCS rule; IsEmbedded says it ships
		"workshop/issues/000231-x.md": true,
		"atlas/index.md":              true,
		"docs/vision/x.txt":           true,
		"cmd/sdlc/close.go":           false,
		"Makefile":                    false,
		".gitignore":                  false,
		"workshopper/x.go":            false, // a segment, not a substring
	} {
		if got := IsDoc(path); got != want {
			t.Errorf("IsDoc(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestIsEmbedded: markdown under cmd/ ships inside the binary via go:embed, so
// it is code even though it is *.md (#174).
func TestIsEmbedded(t *testing.T) {
	for path, want := range map[string]bool{
		"cmd/sdlc/internal/judge/prompts/plan.md": true,
		"cmd/sdlc/close.go":                       true,
		"command/x.md":                            false,
		"README.md":                               false,
	} {
		if got := IsEmbedded(path); got != want {
			t.Errorf("IsEmbedded(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestIsCodeFile is the quick-flow shell's notion of a code file (#231): code
// production surface — embedded prompts included — excluding docs, the process
// trees, and tests, so writing a test never pushes a change out of the shell.
func TestIsCodeFile(t *testing.T) {
	for path, want := range map[string]bool{
		"cmd/sdlc/close.go":                          true,
		"cmd/sdlc/internal/judge/prompts/plan.md":    true,
		"lua/parley/init.lua":                        true,
		"go.mod":                                     true,
		"Makefile":                                   true,
		"README.md":                                  false,
		"docs/guide.md":                              false,
		"workshop/issues/000231-x.md":                false,
		"atlas/index.md":                             false,
		"cmd/sdlc/close_test.go":                     false,
		"tests/parley/chat_spec.lua":                 false,
		"cmd/sdlc/internal/judge/testdata/golden.md": false,
	} {
		if got := IsCodeFile(path); got != want {
			t.Errorf("IsCodeFile(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestClassifyPathNonGoTestLayouts: the fleet is not all Go. parley.nvim keeps
// its tests under tests/**/*_spec.lua, and the quick flow's shell must not count
// them as code (#231). This widens what churn_test means for non-Go repos from
// #231 onward.
func TestClassifyPathNonGoTestLayouts(t *testing.T) {
	for _, path := range []string{
		"tests/parley/chat_spec.lua",
		"lua/parley/chat_spec.lua",
		"spec/parley/chat_test.lua",
		"test/test_parse.py",
		"pkg/parse_test.py",
		"src/app.test.ts",
		"src/app.spec.tsx",
		"web/__tests__/app.js",
		"construct/scripts/test/portable.test.sh",
	} {
		if got := ClassifyPath(path); got != CodeTest {
			t.Errorf("ClassifyPath(%q) = %v, want %v", path, got, CodeTest)
		}
	}
	for _, path := range []string{"lua/parley/test_helpers_config.lua", "cmd/sdlc/internal/testfix/testfix.go", "latest/x.go"} {
		if got := ClassifyPath(path); got != CodeProd {
			t.Errorf("ClassifyPath(%q) = %v, want %v — not a test layout", path, got, CodeProd)
		}
	}
}

// TestCodeFileRule: every clause of the printed rule holds for IsCodeFile, and
// every class the classifier separates is named by the rule.
func TestCodeFileRule(t *testing.T) {
	for _, c := range []struct {
		clause   string
		examples []string
		code     bool
	}{
		{"tests", []string{"cmd/x_test.go", "tests/a_spec.lua"}, false},
		{"docs", []string{"README.md", "docs/guide.md"}, false},
		{"workshop/", []string{"workshop/issues/000231-x.md"}, false},
		{"atlas/", []string{"atlas/index.md"}, false},
		{"cmd/", []string{"cmd/sdlc/internal/judge/prompts/plan.md", "cmd/sdlc/helptext/close.md"}, true},
	} {
		if !strings.Contains(CodeFileRule, c.clause) {
			t.Errorf("CodeFileRule does not name %q", c.clause)
		}
		for _, ex := range c.examples {
			if IsCodeFile(ex) != c.code {
				t.Errorf("IsCodeFile(%q) = %v, but CodeFileRule's %q clause says %v", ex, !c.code, c.clause, c.code)
			}
		}
	}
}
