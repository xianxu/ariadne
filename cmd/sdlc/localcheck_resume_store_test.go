package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func resumeStoreParent(t *testing.T, root, id string) string {
	t.Helper()
	a := storedAttemptFixture(id)
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "attempts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, id+".json")
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestResumeStoreParentLock(t *testing.T) {
	root := t.TempDir()
	p := resumeStoreParent(t, root, "aa")
	first, err := lockResumeParent(context.Background(), root, "aa")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := lockResumeParent(ctx, root, "aa"); !errors.Is(err, context.Canceled) {
		t.Fatalf("lock cancellation: %v", err)
	}
	if err := os.Rename(p, p+".old"); err != nil {
		t.Fatal(err)
	}
	resumeStoreParent(t, root, "aa")
	if err := first.verify(); err == nil {
		t.Fatal("replacement retained authority")
	}
	if _, err := lockResumeParent(context.Background(), root, "../aa"); err == nil {
		t.Fatal("traversal accepted")
	}
	if err := os.Symlink(p, filepath.Join(root, "attempts", "bb.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := lockResumeParent(context.Background(), root, "bb"); err == nil {
		t.Fatal("symlink parent accepted")
	}
}

func TestResumeStoreLockWaitsForOwner(t *testing.T) {
	root := t.TempDir()
	resumeStoreParent(t, root, "aa")
	first, err := lockResumeParent(context.Background(), root, "aa")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := lockResumeParent(ctx, root, "aa"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second caller did not wait cancellably: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := lockResumeParent(context.Background(), root, "aa")
	if err != nil {
		t.Fatal(err)
	}
	second.Close()
}

func TestResumeStorePublication(t *testing.T) {
	root := t.TempDir()
	p := resumeStoreParent(t, root, "aa")
	lock, err := lockResumeParent(context.Background(), root, "aa")
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	in := resumeCoreFixture(t)
	in.Parent = lock.Attempt
	in.ParentRecordSHA256 = resumeRecordDigest(lock.Raw)
	in.ParentLogSHA256 = lock.Attempt.LogSHA256
	l := resumeCoreLink(in)
	raw, _ := json.Marshal(l)
	if err := publishResumeEvidence(root, lock, raw); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(p)
	if string(before) != string(lock.Raw) {
		t.Fatal("parent rewritten")
	}
	if err := publishResumeEvidence(root, lock, raw); err == nil {
		t.Fatal("replaced immutable link")
	}
	l.ID = "dd"
	other, _ := json.Marshal(l)
	if err := publishResumeEvidence(root, lock, other); err == nil {
		t.Fatal("second link to same parent")
	}
	got, _ := os.ReadFile(filepath.Join(root, "attempts", "bb.json"))
	if string(got) != string(raw) {
		t.Fatal("link changed")
	}
	if err := os.WriteFile(p, append(before, ' '), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishResumeEvidence(root, lock, other); err == nil {
		t.Fatal("parent edit accepted")
	}
}

func TestResumeSnapshotIndexesTypedProof(t *testing.T) {
	root := t.TempDir()
	resumeStoreParent(t, root, "aa")
	l := resumeCoreLink(resumeCoreFixture(t))
	raw, _ := json.Marshal(l)
	if err := os.WriteFile(filepath.Join(root, "attempts", "bb.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	s, err := readSuiteEvidence(root, 100)
	if err != nil || len(s.Warnings) != 0 || len(s.Unsupported) != 0 || len(s.Resumes) != 1 || len(s.Attempts) != 1 {
		t.Fatalf("snapshot %+v %v", s, err)
	}
	if s.recordDigests["bb"] != resumeRecordDigest(raw) || s.recordDigests["aa"] == "" {
		t.Fatal("missing exact raw digests")
	}
	if !s.knownSuiteHeads()[l.SHA] {
		t.Fatal("link allowed legacy substitution")
	}
}

func TestResumeSnapshotPendingGuardsScope(t *testing.T) {
	root := t.TempDir()
	in := resumeCoreFixture(t)
	raw, _ := json.Marshal(in.Parent)
	if err := os.MkdirAll(filepath.Join(root, "attempts"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "attempts", "aa.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	l := resumeCoreLink(in)
	bytes, _ := json.Marshal(l)
	if err := os.WriteFile(filepath.Join(root, "attempts", "bb.json"), bytes, 0600); err != nil {
		t.Fatal(err)
	}
	s, err := readSuiteEvidence(root, 100)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.proofError(proofSubject{Repo: l.Repo, SHA: l.SHA, Check: suiteCheckName}); err == nil {
		t.Fatal("pending recognized link ignored by pass guard")
	}
	if err := s.proofError(proofSubject{Repo: l.Repo, SHA: l.SHA, Check: suiteCheckName, ID: l.RerunOf}); err == nil {
		t.Fatal("pending parent accepted")
	}
	if err := s.proofError(proofSubject{Repo: l.Repo, SHA: l.SHA, Check: "lint"}); err != nil {
		t.Fatalf("unrelated check blocked: %v", err)
	}
	if err := s.proofError(proofSubject{Repo: "other/repo", SHA: l.SHA, Check: suiteCheckName}); err != nil {
		t.Fatalf("unrelated repository blocked: %v", err)
	}
}

func TestResumeStorePrunePinsPending(t *testing.T) {
	root := t.TempDir()
	resumeStoreParent(t, root, "aa")
	lock, err := lockResumeParent(context.Background(), root, "aa")
	if err != nil {
		t.Fatal(err)
	}
	in := resumeCoreFixture(t)
	in.Parent = lock.Attempt
	in.ParentRecordSHA256 = resumeRecordDigest(lock.Raw)
	in.ParentLogSHA256 = lock.Attempt.LogSHA256
	l := resumeCoreLink(in)
	raw, _ := json.Marshal(l)
	if err := publishResumeEvidence(root, lock, raw); err != nil {
		t.Fatal(err)
	}
	lock.Close()
	old := time.Now().Add(-40 * 24 * time.Hour)
	for _, id := range []string{"aa", "bb"} {
		p := filepath.Join(root, "attempts", id+".json")
		if err := os.Chtimes(p, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if err := pruneSuiteAttempts(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"aa", "bb"} {
		if _, err := os.Stat(filepath.Join(root, "attempts", id+".json")); err != nil {
			t.Fatalf("pending member %s pruned: %v", id, err)
		}
	}
}
