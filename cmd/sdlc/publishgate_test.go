package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// publishRepo inits a temp git repo, chdir's in (so gitx.RunGit/Capture bind to it),
// creates workshop/issues, and returns a git helper + the base SHA (post-init) to
// use as the merge/push window base. Restores cwd on cleanup.
func publishRepo(t *testing.T) (git func(args ...string), base string) {
	t.Helper()
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	git = func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	git("config", "user.email", "t@e.com")
	git("config", "user.name", "T")
	git("config", "commit.gpgsign", "false")
	if err := os.MkdirAll("workshop/issues", 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile("README", []byte("x\n"), 0o644)
	git("add", "README")
	git("commit", "-q", "-m", "init")
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
