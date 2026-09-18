package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// runVerbSequence drives ONE issue through the real commands — claim,
// change-code, a code commit, close — exactly as an agent would, passing no
// flow flag anywhere. It returns the prompt the boundary review received and
// the issue as close left it.
func runVerbSequence(t *testing.T, id int, code map[string]string) (prompt, issueText string) {
	t.Helper()
	repo, _ := syncRepo(t)
	name := fmt.Sprintf("%06d-e2e.md", id)
	writeSyncIssue(t, repo, name, fmt.Sprintf("---\nid: %06d\nstatus: open\ndeps: []\n---\n\n# e2e\n\n"+
		"## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Log\n", id))
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "file the issue")
	git(t, repo, "push", "-q", "origin", "main")

	run := func(args ...string) {
		t.Helper()
		if _, stderr, err := executeSDLCTestCommand(args...); err != nil {
			t.Fatalf("sdlc %s: %v\n%s", strings.Join(args, " "), err, stderr)
		}
	}
	idArg := fmt.Sprint(id)
	run("claim", "--issue", idArg)
	run("change-code", "--issue", idArg, "--worktree=no")

	for p, text := range code {
		os.MkdirAll(filepath.Join(repo, filepath.Dir(p)), 0o755)
		os.WriteFile(filepath.Join(repo, p), []byte(text), 0o644)
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", fmt.Sprintf("#%d: implement", id))

	_, last := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	run("close", "--issue", idArg, "--verified", "e2e", "--no-actual", "--no-atlas")
	b, err := os.ReadFile(filepath.Join(repo, syncIssuesDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return *last, string(b)
}

func e2eFlow(t *testing.T, text string) flow.Flow {
	t.Helper()
	fm, _, err := issue.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	f, recorded, err := flow.FromFrontmatter(fm)
	if err != nil || !recorded {
		t.Fatalf("no flow record after the sequence (err %v):\n%s", err, text)
	}
	return f
}

// TestQuickAndFullFlowsSameVerbSequence is #231's acceptance test: the agent's
// sequence of verbs is identical on both flows, and the gates decide. A small
// issue (one code file, no plan) stays quick and is reviewed with the
// small-diff recipe; a large one (over the line limit) is inferred quick at
// change-code — nothing about it said otherwise — and upgraded to full at close,
// where the diff exists, and gets the full review.
func TestQuickAndFullFlowsSameVerbSequence(t *testing.T) {
	prompt, text := runVerbSequence(t, 301, map[string]string{"cmd/a.go": goLines(20)})
	if f := e2eFlow(t, text); f.Kind() != flow.Quick {
		t.Errorf("small issue ended %+v, want quick", f)
	}
	if !strings.Contains(prompt, smallDiffMarker) {
		t.Error("small issue was not reviewed with the small-diff recipe")
	}

	prompt, text = runVerbSequence(t, 302, overLineLimit())
	if f := e2eFlow(t, text); f.Kind() != flow.Full || f.Provenance() != flow.Inferred {
		t.Errorf("large issue ended %+v, want full/inferred", f)
	}
	if strings.Contains(prompt, smallDiffMarker) || !strings.Contains(prompt, "each of the 8 entries") {
		t.Error("large issue was not reviewed with the full recipe")
	}
	if !strings.Contains(text, "flow upgraded quick → full") {
		t.Errorf("large issue's Log does not record the upgrade:\n%s", text)
	}
}
