package judge

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite testdata/golden/*.prompt from current BuildPrompt")

// goldenInput exercises EVERY placeholder (non-empty diff, issue, plan, refs,
// changed issues, and the milestone-review repo-orientation fields) so the
// golden captures the fully-rendered output of every category.
var goldenInput = PromptInput{
	Diff:          "DIFF-BODY-LINE-1\nDIFF-BODY-LINE-2",
	ChangedIssues: []string{"workshop/issues/000001-a.md", "workshop/issues/000002-b.md"},
	Base:          "BASE_SHA", Head: "HEAD",
	IssueRef:     "pair#31 M2",
	IssueContent: "ISSUE-CONTENT-BODY",
	PlanContent:  "PLAN-CONTENT-BODY",
	Repo:         "pair", RepoRoot: "/abs/pair",
	IssueFile: "workshop/issues/000031-x.md",
	Boundary:  "milestone M2 close",
	RepoNote:  "REPO-ORIENTATION-NOTE",
}

func goldenCategories() []Category {
	return append(append([]Category{}, AllCategories()...), EstimateQuality)
}

// TestBuildPrompt_Golden is the byte-fidelity net for the prompts→markdown
// refactor (#153 M2). Capture once from the CURRENT BuildPrompt
// (-update-golden), then every extraction must keep it byte-identical.
//
// ⛔ After the initial capture, NEVER re-run -update-golden to "fix" a failure:
// a failure means a .md drifted — fix the .md, not the golden.
func TestBuildPrompt_Golden(t *testing.T) {
	for _, c := range goldenCategories() {
		if c == Lessons {
			continue // no prompt (BuildPrompt returns "")
		}
		got := BuildPrompt(c, goldenInput)
		path := filepath.Join("testdata", "golden", string(c)+".prompt")
		if *updateGolden {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("golden missing for %s (capture: go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden -update-golden): %v", c, err)
		}
		if got != string(want) {
			t.Errorf("BYTE DRIFT for %s: BuildPrompt output changed vs golden — fix the .md, not the golden.\n--- got ---\n%s", c, got)
		}
	}
}

// TestBuildPrompt_EmptyIssueRefFallback covers the ref=="" → "<unknown>" branch,
// which goldenInput (non-empty IssueRef) does not exercise.
func TestBuildPrompt_EmptyIssueRefFallback(t *testing.T) {
	in := goldenInput
	in.IssueRef = ""
	for _, c := range []Category{PlanQuality, EstimateQuality} {
		if !strings.Contains(BuildPrompt(c, in), "<unknown>") {
			t.Errorf("%s with empty IssueRef should render <unknown>", c)
		}
	}
}
