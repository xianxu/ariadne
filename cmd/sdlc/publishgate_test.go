package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// publishRepo inits a temp git repo, chdir's in (so gitx.RunGit/Capture bind to it),
// creates workshop/issues, and returns a git helper + the base SHA (post-init) to
// use as the merge/push window base. Restores cwd on cleanup.
func publishRepo(t *testing.T) (git func(args ...string), base string) {
	t.Helper()
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	git = func(args ...string) { t.Helper(); testfix.Git(t, dir, args...) }
	if err := os.MkdirAll("workshop/issues", 0o755); err != nil {
		t.Fatal(err)
	}
	base = strings.TrimSpace(gitx.Capture("rev-parse", "HEAD"))
	return git, base
}

func issuePathFor(n int) string {
	return filepath.Join("workshop/issues", fmt.Sprintf("%06d-x.md", n))
}

// writeIssueStatus writes an issue file at status, commits it touching the file.
func writeIssueStatus(t *testing.T, git func(...string), n int, status, subject string) {
	t.Helper()
	p := issuePathFor(n)
	// Embed the subject in the body so each write differs (a re-close, like the real
	// one, adds a Log line → a real commit, not an empty "nothing to commit").
	body := fmt.Sprintf("---\nid: %06d\nstatus: %s\nactual_hours: 1\n---\n# T\n\n%s\n", n, status, subject)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", subject)
}

// commitCode makes a code-only commit (no issue-file touch) → drift after close.
func commitCode(t *testing.T, git func(...string), name string) {
	t.Helper()
	os.WriteFile(name, []byte(name+"\n"), 0o644)
	git("add", name)
	git("commit", "-q", "-m", "code: "+name)
}

func TestCodecompleteAnchorCommit(t *testing.T) {
	git, _ := publishRepo(t)
	writeIssueStatus(t, git, 69, "working", "#69: wip")
	writeIssueStatus(t, git, 69, "codecomplete", "#69: close → codecomplete")
	closeSHA := strings.TrimSpace(gitx.Capture("rev-parse", "HEAD"))

	if got := codecompleteAnchorCommit(issuePathFor(69)); got != closeSHA {
		t.Fatalf("anchor = %q, want the codecomplete commit %q", got, closeSHA)
	}

	// A later code commit that does NOT touch the issue file must NOT move the anchor.
	commitCode(t, git, "fix.go")
	if got := codecompleteAnchorCommit(issuePathFor(69)); got != closeSHA {
		t.Errorf("code drift must not move the anchor: got %q, want %q", got, closeSHA)
	}

	// A re-close (writes codecomplete again, touching the issue file) MUST advance it.
	writeIssueStatus(t, git, 69, "codecomplete", "#69: re-close after drift")
	reSHA := strings.TrimSpace(gitx.Capture("rev-parse", "HEAD"))
	if got := codecompleteAnchorCommit(issuePathFor(69)); got != reSHA {
		t.Errorf("re-close must advance the anchor: got %q, want %q", got, reSHA)
	}
}

func TestMergedCodecompleteIssues(t *testing.T) {
	git, base := publishRepo(t)
	writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
	writeIssueStatus(t, git, 70, "working", "#70 wip")

	got, err := mergedCodecompleteIssues(base, "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != issuePathFor(69) {
		t.Fatalf("want only the codecomplete issue #69, got %v", got)
	}
}

func TestMergedCodecompleteIssuesPreservesGitError(t *testing.T) {
	t.Setenv("PATH", "")
	_, err := mergedCodecompleteIssues("base", "workshop/issues")
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), `git diff base..HEAD: exec: "git": executable file not found in $PATH`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if !errors.Is(err, exec.ErrNotFound) {
		t.Fatalf("errors.Is(%v, exec.ErrNotFound) = false", err)
	}
}

func TestRunPublishGate(t *testing.T) {
	t.Run("clean: HEAD == anchor passes", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
		if err := runPublishGate(base, "workshop/issues", io.Discard); err != nil {
			t.Errorf("HEAD==anchor should pass, got: %v", err)
		}
	})

	t.Run("drift: commit after close refuses", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
		commitCode(t, git, "late.go")
		err := runPublishGate(base, "workshop/issues", io.Discard)
		if err == nil || !strings.Contains(err.Error(), "landed after `sdlc close`") {
			t.Errorf("post-close drift should refuse with a re-run-close message, got: %v", err)
		}
	})

	t.Run("multi-issue: latest anchor, no false drift", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "codecomplete", "#69 close") // anchor X
		writeIssueStatus(t, git, 70, "codecomplete", "#70 close") // anchor Y = HEAD
		if err := runPublishGate(base, "workshop/issues", io.Discard); err != nil {
			t.Errorf("two sequential closes (latest anchor==HEAD) should pass, got: %v", err)
		}
	})

	t.Run("re-close after drift passes", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
		commitCode(t, git, "drift.go")
		writeIssueStatus(t, git, 69, "codecomplete", "#69 re-close") // advances anchor to HEAD
		if err := runPublishGate(base, "workshop/issues", io.Discard); err != nil {
			t.Errorf("re-close (anchor advanced to HEAD) should pass, got: %v", err)
		}
	})

	t.Run("no codecomplete issue is a no-op", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "working", "#69 wip")
		if err := runPublishGate(base, "workshop/issues", io.Discard); err != nil {
			t.Errorf("no codecomplete issue should pass (no-op), got: %v", err)
		}
	})
}

func TestPublishCodecompleteIssues(t *testing.T) {
	git, _ := publishRepo(t)
	writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
	writeIssueStatus(t, git, 70, "working", "#70 wip")
	before, err := os.ReadFile(issuePathFor(69))
	if err != nil {
		t.Fatal(err)
	}
	_, bodyBefore, err := issue.Parse(string(before))
	if err != nil {
		t.Fatal(err)
	}

	flipped, err := publishCodecompleteIssues("workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	if len(flipped) != 1 || flipped[0] != issuePathFor(69) {
		t.Fatalf("want only #69 flipped, got %v", flipped)
	}
	got69, _ := os.ReadFile(issuePathFor(69))
	if !strings.Contains(string(got69), "status: done") {
		t.Errorf("#69 should be flipped to done:\n%s", got69)
	}
	fmAfter, bodyAfter, err := issue.Parse(string(got69))
	if err != nil {
		t.Fatal(err)
	}
	if bodyAfter != bodyBefore {
		t.Errorf("body changed during status flip:\nbefore %q\nafter  %q", bodyBefore, bodyAfter)
	}
	if updated, _ := issue.GetField(fmAfter, "updated"); updated != time.Now().Format("2006-01-02") {
		t.Errorf("updated = %q, want today", updated)
	}
	got70, _ := os.ReadFile(issuePathFor(70))
	if !strings.Contains(string(got70), "status: working") {
		t.Errorf("#70 (working) must be untouched:\n%s", got70)
	}
}

// commitDocs makes a docs-only commit (root-level *.md — the measured 6/6
// friction shape: lessons.md / plan ticks / atlas after the close commit).
func commitDocs(t *testing.T, git func(...string), name string) {
	t.Helper()
	os.WriteFile(name, []byte("# notes\n"), 0o644)
	git("add", name)
	git("commit", "-q", "-m", "docs: "+name)
}

// #174 leg C: post-close deltas with no code surface pass the publish gate —
// the boundary review's claims are about code behavior, and docs are not
// reviewable code surface (#177's hasCodePath definition, shared here).
func TestRunPublishGate_DocsOnly(t *testing.T) {
	t.Run("docs-only drift after close passes (#174)", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
		commitDocs(t, git, "lessons.md")
		var stderr strings.Builder
		if err := runPublishGate(base, "workshop/issues", &stderr); err != nil {
			t.Errorf("docs-only delta should pass: %v", err)
		}
		for _, want := range []string{"doc-only", "#174"} {
			if !strings.Contains(stderr.String(), want) {
				t.Errorf("docs-only pass line missing %q:\n%s", want, stderr.String())
			}
		}
	})

	t.Run("mixed docs+code drift refuses", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
		commitDocs(t, git, "lessons.md")
		commitCode(t, git, "late.go")
		err := runPublishGate(base, "workshop/issues", io.Discard)
		if err == nil || !strings.Contains(err.Error(), "landed after `sdlc close`") {
			t.Errorf("mixed delta should refuse with the pinned message, got: %v", err)
		}
	})

	t.Run("multi-issue: two anchors + trailing docs commit passes", func(t *testing.T) {
		git, base := publishRepo(t)
		writeIssueStatus(t, git, 69, "codecomplete", "#69 close") // older anchor
		writeIssueStatus(t, git, 70, "codecomplete", "#70 close") // newest anchor
		commitDocs(t, git, "lessons.md")
		if err := runPublishGate(base, "workshop/issues", io.Discard); err != nil {
			t.Errorf("docs-only delta past the newest anchor should pass: %v", err)
		}
	})
}

// TestFormatPublishGateDocsOnly_ContractElements pins the pass line's content
// and, critically, that it collides with no gatesig classifier pattern — the
// refusal vocabulary ("landed after") lives one branch away (#172).
func TestFormatPublishGateDocsOnly_ContractElements(t *testing.T) {
	msg := formatPublishGateDocsOnly(3, "abc1234")
	for _, w := range []string{"3", "abc1234", "doc-only", "#174"} {
		if !strings.Contains(msg, w) {
			t.Errorf("formatPublishGateDocsOnly missing %q in:\n%s", w, msg)
		}
	}
	assertNoGatesigCollision(t, msg)
}

// #174 close review I1: helptext under cmd/ is *.md but ships inside the
// binary (//go:embed) — a post-close helptext edit is code surface for the
// PUBLISH decision and must not ride the doc-only pass.
func TestRunPublishGate_EmbeddedHelptextIsCodeSurface(t *testing.T) {
	git, base := publishRepo(t)
	writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
	if err := os.MkdirAll("cmd/sdlc/helptext", 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile("cmd/sdlc/helptext/close.md", []byte("edited\n"), 0o644)
	git("add", "cmd/sdlc/helptext/close.md")
	git("commit", "-q", "-m", "docs: helptext tweak")
	err := runPublishGate(base, "workshop/issues", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "landed after `sdlc close`") {
		t.Errorf("embedded-helptext delta should refuse, got: %v", err)
	}
}

// TestPublishGateHasCodeSurface pins the tightened predicate directly:
// hasCodePath's docs stay docs, EXCEPT under cmd/.
func TestPublishGateHasCodeSurface(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  bool
	}{
		{"lessons.md is docs", []string{"lessons.md"}, false},
		{"workshop is docs", []string{"workshop/issues/000174-x.md"}, false},
		{"atlas is docs", []string{"atlas/workflow/x.md"}, false},
		{"go file is code", []string{"cmd/sdlc/close.go"}, true},
		{"embedded helptext md is code (ships in the binary)", []string{"cmd/sdlc/helptext/close.md"}, true},
		{"mixed docs + helptext is code", []string{"lessons.md", "cmd/sdlc/helptext/push.md"}, true},
		{"empty is docs", nil, false},
	}
	for _, tc := range cases {
		if got := publishGateHasCodeSurface(tc.paths); got != tc.want {
			t.Errorf("%s: publishGateHasCodeSurface(%v) = %v, want %v", tc.name, tc.paths, got, tc.want)
		}
	}
}

// publishFlowIssue commits a #69 history the way the quick flow leaves it: the
// issue at working with its flow record, code under the issue (files → added
// lines), then the close commit that writes codecomplete and is the publish
// anchor. record is the frontmatter flow value.
func publishFlowIssue(t *testing.T, git func(...string), record string, code map[string]int) {
	t.Helper()
	write := func(status string) {
		body := fmt.Sprintf("---\nid: 000069\nstatus: %s\nactual_hours: 1\nflow: %s\n---\n# T\n\n%s\n", status, record, status)
		if err := os.WriteFile(issuePathFor(69), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("working")
	git("add", ".")
	git("commit", "-q", "-m", "#69: issue-sync: spec/plan at change-code")
	for path, n := range code {
		os.MkdirAll(filepath.Dir(path), 0o755)
		if err := os.WriteFile(path, []byte(strings.Repeat("var _ = 1\n", n)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("add", ".")
	git("commit", "-q", "-m", "#69: implement, fixes after the verdict included")
	write("codecomplete")
	git("add", ".")
	git("commit", "-q", "-m", "#69: close")
}

const (
	quickRecord = `{kind: quick, provenance: inferred, spec: "1a2b3c4d", done: "5e6f7a8b"}`
	fullRecord  = `{kind: full, provenance: inferred}`
)

// TestRunPublishGate_QuickGrewPastReview (#231): close measures the head its
// small-diff review saw, and fixes made after the verdict ride into the close
// commit unmeasured. The publish check re-measures a quick issue's final diff
// over close's own window: up to flow.MaxAddedLinesAfterReview it publishes, one
// line past it sends the issue back to close. Only quick issues and only code
// lines count, and the docs-only pass path checks it too.
func TestRunPublishGate_QuickGrewPastReview(t *testing.T) {
	limit := flow.MaxAddedLinesAfterReview
	cases := []struct {
		name    string
		record  string
		code    map[string]int
		docs    bool // a docs-only commit after the close (the #174 pass path)
		refuses bool
	}{
		{"quick at the limit publishes", quickRecord, map[string]int{"cmd/a.go": limit - 20, "cmd/b.go": 20}, false, false},
		{"quick one line past refuses", quickRecord, map[string]int{"cmd/a.go": limit - 20, "cmd/b.go": 21}, false, true},
		{"quick past it via the docs-only path refuses", quickRecord, map[string]int{"cmd/a.go": limit + 1}, true, true},
		{"test lines never count", quickRecord, map[string]int{"cmd/a.go": 10, "cmd/a_test.go": 5 * limit}, false, false},
		{"a full issue is not the quick flow's to re-measure", fullRecord, map[string]int{"cmd/a.go": 5 * limit}, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			git, base := publishRepo(t)
			publishFlowIssue(t, git, c.record, c.code)
			if c.docs {
				commitDocs(t, git, "lessons.md")
			}
			err := runPublishGate(base, "workshop/issues", io.Discard)
			if !c.refuses {
				if err != nil {
					t.Errorf("want a publish, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("want a refusal sending the issue back to close, got a publish")
			}
			for _, want := range []string{"quick flow", "sdlc close --issue 69", strconv.Itoa(limit+1) + " added lines"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("refusal missing %q:\n%v", want, err)
				}
			}
		})
	}
}
