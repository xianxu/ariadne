package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type resumeEnvironmentFixture struct {
	d                                                 localRunDeps
	store, wt, inflight, scratchRoot, scratch, gitdir string
	link                                              ResumeLink
	receipt                                           ResumeReconciliation
}

func resumeEnvironmentPrivate(t *testing.T) resumeEnvironmentFixture {
	t.Helper()
	d, sha, _, _ := localRunRepo(t, "unused")
	wt := filepath.Join(t.TempDir(), localCheckTmpPrefix+"private")
	if out, err := d.Git("worktree", "add", "--detach", wt, sha); err != nil {
		t.Fatalf("add: %v %s", err, out)
	}
	t.Cleanup(func() { _, _ = d.Git("worktree", "remove", "--force", "--force", wt) })
	in := resumeCoreFixture(t)
	in.Parent.Repo = d.Repo
	in.Parent.SHA = sha
	l := *resumeCoreLink(in)
	l.Owner.PID = 999999999
	store := t.TempDir()
	inflight := filepath.Join(store, "inflight", l.ChildID)
	if err := os.MkdirAll(inflight, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inflight, "run.log"), []byte("partial raw evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	raw := retentionWrite(t, store, l.ID, l)
	r := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: l.Repo, SHA: l.SHA, Check: l.Check, RerunOf: l.RerunOf, LinkID: l.ID, LinkSHA256: resumeRecordDigest(raw), ChildID: l.ChildID, Actor: "operator", Note: "verified descendants stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt.Add(time.Hour)}
	retentionWrite(t, store, r.ID, r)
	scratchRoot := t.TempDir()
	scratch := filepath.Join(scratchRoot, "sdlc-run-suite-rerun-999999999-123")
	if err := os.Mkdir(scratch, 0700); err != nil {
		t.Fatal(err)
	}
	if err := attachRunnerArtifactOwner(scratch, wt); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(wt, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	gitdir, err := parseRunnerOwner(raw, wt, true)
	if err != nil {
		t.Fatal(err)
	}
	return resumeEnvironmentFixture{d: d, store: store, wt: wt, inflight: inflight, scratchRoot: scratchRoot, scratch: scratch, gitdir: gitdir, link: l, receipt: r}
}

func TestResumeEnvironmentProtectBindsExistingArtifacts(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	reason, err := os.ReadFile(filepath.Join(f.gitdir, "locked"))
	if err != nil {
		t.Fatal("resume worktree is not locked", err)
	}
	if strings.TrimSpace(string(reason)) != resumeEnvironmentReason(f.link.ChildID) {
		t.Fatalf("lock reason %q", reason)
	}
	raw, err := os.ReadFile(filepath.Join(f.inflight, runnerArtifactOwnerFile))
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseRunnerOwner(raw, f.inflight, false)
	if err != nil || got != f.gitdir {
		t.Fatalf("owner %s %v", raw, err)
	}
}

func TestResumeEnvironmentRetiresOnlyBoundInputs(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{f.wt, f.scratch} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("bound input survived %s: %v", path, err)
		}
	}
	for _, path := range []string{filepath.Join(f.inflight, "run.log"), filepath.Join(f.store, "attempts", f.receipt.ID+".json")} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("retained audit missing %s: %v", path, err)
		}
	}
	// A marker whose worktree was already removed grants no further effects.
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
}

func TestResumeEnvironmentRefusesUntrustedBindings(t *testing.T) {
	for _, mode := range []string{"receipt-outcome", "receipt-unpublished", "wrong-reason", "wrong-sha", "foreign-owner", "wrong-back-pointer", "symlink-owner", "symlink-gitdir", "symlink-inflight", "other-pid", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			f := resumeEnvironmentPrivate(t)
			if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(f.inflight, runnerArtifactOwnerFile)
			write := func(path, value string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(value), 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx := context.Background()
			switch mode {
			case "receipt-outcome":
				f.receipt.Outcome = "pending"
			case "receipt-unpublished":
				if err := os.Remove(filepath.Join(f.store, "attempts", f.receipt.ID+".json")); err != nil {
					t.Fatal(err)
				}
			case "wrong-reason":
				write(filepath.Join(f.gitdir, "locked"), resumeEnvironmentReason("ee")+"\n")
			case "wrong-sha":
				write(filepath.Join(f.gitdir, "HEAD"), strings.Repeat("e", 40)+"\n")
			case "foreign-owner":
				write(marker, "gitdir: "+filepath.Join(t.TempDir(), "worktrees", "foreign")+"\n")
			case "wrong-back-pointer":
				write(filepath.Join(f.wt, ".git"), "gitdir: "+f.d.Root+"\n")
			case "symlink-owner":
				raw, err := os.ReadFile(marker)
				if err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(t.TempDir(), "owner")
				write(target, string(raw))
				if err := os.Remove(marker); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, marker); err != nil {
					t.Fatal(err)
				}
			case "symlink-gitdir":
				target := filepath.Join(t.TempDir(), "registration")
				if err := os.Rename(f.gitdir, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, f.gitdir); err != nil {
					t.Fatal(err)
				}
			case "symlink-inflight":
				target := filepath.Join(t.TempDir(), "child")
				if err := os.Rename(f.inflight, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, f.inflight); err != nil {
					t.Fatal(err)
				}
			case "other-pid":
				other := filepath.Join(f.scratchRoot, "sdlc-run-suite-rerun-999999998-123")
				if err := os.Rename(f.scratch, other); err != nil {
					t.Fatal(err)
				}
				f.scratch = other
			case "cancelled":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			if err := retireResumeEnvironmentWithRoots(ctx, f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err == nil {
				t.Fatal("untrusted binding authorized cleanup")
			}
			for _, path := range []string{f.wt, f.scratch, filepath.Join(f.inflight, "run.log")} {
				if _, err := os.Stat(path); err != nil {
					t.Errorf("refused cleanup deleted %s: %v", path, err)
				}
			}
		})
	}
}

func TestResumeEnvironmentLostRemovalReplyAndMissingMarker(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	f.d.CleanupGit = func(args ...string) (string, error) {
		if len(args) > 1 && args[0] == "worktree" && args[1] == "remove" {
			if _, err := os.Stat(f.scratch); !os.IsNotExist(err) {
				t.Errorf("worktree removal preceded scratch cleanup: %v", err)
			}
			want := []string{"worktree", "remove", "--force", "--force", f.wt}
			if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
				t.Errorf("unexpected cleanup command %v", args)
			}
		}
		out, err := f.d.Git(args...)
		if err == nil && len(args) > 1 && args[0] == "worktree" && args[1] == "remove" {
			return out, errors.New("reply lost after removal")
		}
		return out, err
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err == nil {
		t.Fatal("lost response claimed confirmed cleanup")
	}
	if _, err := os.Stat(filepath.Join(f.inflight, "run.log")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(f.store, "attempts", f.receipt.ID+".json")); err != nil {
		t.Fatal(err)
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(f.inflight, runnerArtifactOwnerFile)); err != nil {
		t.Fatal(err)
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
}

func TestResumeEnvironmentIgnoresUnrelatedScratch(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	other := resumeEnvironmentPrivate(t)
	unrelated := filepath.Join(f.scratchRoot, "sdlc-run-suite-123-456")
	if err := os.Mkdir(unrelated, 0700); err != nil {
		t.Fatal(err)
	}
	if err := attachRunnerArtifactOwner(unrelated, other.wt); err != nil {
		t.Fatal(err)
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{unrelated, other.wt, other.scratch} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("unrelated artifact lost: %s %v", path, err)
		}
	}
}

func TestResumeEnvironmentRevalidatesInflightOwnerBeforeDeletion(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	reads := 0
	f.d.CleanupGit = func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "rev-parse" {
			reads++
			if reads == 2 {
				if err := os.WriteFile(filepath.Join(f.inflight, runnerArtifactOwnerFile), []byte("changed owner"), 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		return f.d.Git(args...)
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err == nil {
		t.Fatal("changed inflight owner authorized deletion")
	}
	for _, path := range []string{f.wt, f.scratch, filepath.Join(f.inflight, "run.log")} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("changed owner lost %s: %v", path, err)
		}
	}
}
