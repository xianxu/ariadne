package main

import (
	"bytes"
	"testing"
)

func TestIssueIDGateIgnoresSideDirectories(t *testing.T) {
	repo, _ := idRepo(t)
	git(t, repo, "switch", "-q", "-c", "feature")
	writeRepoFile(t, repo, "workshop/issues/000001-first.side/000999-evidence.md", "# untyped evidence\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "side evidence")
	var diagnostic bytes.Buffer
	if err := refuseDuplicateIssueIDs(&diagnostic, idsDir, histDir, execGitRunner{}); err != nil {
		t.Fatalf("side directory falsely claims issue1: %v", err)
	}
	writeIssueAt(t, repo, idsDir, "000001-real-duplicate.md")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "real duplicate")
	if err := refuseDuplicateIssueIDs(&diagnostic, idsDir, histDir, execGitRunner{}); err == nil {
		t.Fatal("side exclusion hid a real duplicate")
	}
}
