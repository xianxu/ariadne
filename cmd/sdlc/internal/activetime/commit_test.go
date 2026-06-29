package activetime

import (
	"fmt"
	"strings"
	"testing"
)

// withGitRun swaps the package gitRun shim for the duration of a test.
func withGitRun(t *testing.T, fn func(repo string, args ...string) ([]byte, error)) {
	t.Helper()
	orig := gitRun
	gitRun = fn
	t.Cleanup(func() { gitRun = orig })
}

func TestLoadWindowCommits(t *testing.T) {
	// Fake `git log` output: tab-delimited %H \t %aI \t %s, oldest first.
	out := strings.Join([]string{
		"aaaaaaaaaaaaaaaa\t2026-01-01T00:50:00-07:00\t#8 first bit",
		"bbbbbbbbbbbbbbbb\t2026-01-01T01:20:00-07:00\t#8 M2: more (also #10) and #8 again",
		"cccccccccccccccc\t2026-01-01T02:00:00-07:00\tchore: no refs",
	}, "\n")
	withGitRun(t, func(repo string, args ...string) ([]byte, error) {
		if repo == "" {
			return nil, fmt.Errorf("expected a repo path")
		}
		return []byte(out), nil
	})

	commits, err := loadWindowCommits("/repo", "2026-01-01T00:00:00Z", "2026-01-02T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 3 {
		t.Fatalf("want 3 commits, got %d", len(commits))
	}
	if commits[0].SHA != "aaaaaaa" {
		t.Fatalf("SHA should be short7, got %q", commits[0].SHA)
	}
	// Order-preserving dedupe: #8 appears twice in commit 2 but once in Issues,
	// and #8 precedes #10 in the subject.
	if got := commits[1].Issues; len(got) != 2 || got[0] != "8" || got[1] != "10" {
		t.Fatalf("want issues [8 10] order-preserving deduped, got %v", got)
	}
	// A commit with no tracked refs has empty Issues.
	if len(commits[2].Issues) != 0 {
		t.Fatalf("want no issues for chore commit, got %v", commits[2].Issues)
	}
	if !commits[0].Time.Equal(tm("2026-01-01T00:50:00-07:00")) {
		t.Fatalf("commit time parse mismatch: %v", commits[0].Time)
	}
}

func TestLoadWindowCommitsParsesAllIssueRefsForClaimants(t *testing.T) {
	out := strings.Join([]string{
		"aaaaaaaaaaaaaaaa\t2026-01-01T00:00:00Z\t#1 c11",
		"bbbbbbbbbbbbbbbb\t2026-01-01T00:20:00Z\t#2 c21",
		"cccccccccccccccc\t2026-01-01T00:40:00Z\t#2 c22",
		"dddddddddddddddd\t2026-01-01T01:00:00Z\t#3 c31",
		"eeeeeeeeeeeeeeee\t2026-01-01T02:00:00Z\tchore: no refs",
	}, "\n")
	withGitRun(t, func(repo string, args ...string) ([]byte, error) {
		return []byte(out), nil
	})

	commits, err := loadWindowCommits("/repo", wideSince, wideUntil)
	if err != nil {
		t.Fatal(err)
	}
	if got := commits[1].Issues; len(got) != 1 || got[0] != "2" {
		t.Fatalf("intervening #2 commit should be parsed as a claimant even when only #1 was seeded, got %v", got)
	}
	if got := commits[3].Issues; len(got) != 1 || got[0] != "3" {
		t.Fatalf("intervening #3 commit should be parsed as a claimant even when only #1 was seeded, got %v", got)
	}
	if got := commits[4].Issues; len(got) != 0 {
		t.Fatalf("no-ref commit should remain a neutral boundary, got %v", got)
	}
}

func TestLoadWindowCommitsEmpty(t *testing.T) {
	withGitRun(t, func(repo string, args ...string) ([]byte, error) { return []byte(""), nil })
	commits, err := loadWindowCommits("/repo", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 0 {
		t.Fatalf("want 0 commits, got %d", len(commits))
	}
}
