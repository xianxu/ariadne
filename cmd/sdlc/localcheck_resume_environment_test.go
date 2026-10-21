package main

import (
	"context"
	"errors"
	"fmt"
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
			if _, err := os.Lstat(f.wt); !os.IsNotExist(err) {
				t.Errorf("registration removal preceded checkout absence: %v", err)
			}
			if _, err := os.Stat(filepath.Join(f.gitdir, "locked")); err != nil {
				t.Errorf("checkout removal lost locked registration: %v", err)
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
	if err := os.Remove(filepath.Join(f.inflight, runnerArtifactOwnerFile)); err != nil && !os.IsNotExist(err) {
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

func TestResumeEnvironmentProtectionDurabilityPrecedesLock(t *testing.T) {
	for _, failAt := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			f := resumeEnvironmentPrivate(t)
			synced := []string{}
			originalGit := f.d.Git
			f.d.Git = func(args ...string) (string, error) {
				if len(args) > 1 && args[0] == "worktree" && args[1] == "lock" {
					if _, err := os.Stat(filepath.Join(f.inflight, runnerArtifactOwnerFile)); err != nil {
						t.Errorf("lock preceded owner marker: %v", err)
					}
					if len(synced) != 2 {
						t.Errorf("lock preceded ownership directory durability: %v", synced)
					}
				}
				return originalGit(args...)
			}
			sync := func(path string) error {
				synced = append(synced, path)
				if len(synced) == failAt {
					return errors.New("owner directory sync fault")
				}
				return syncResumeDirectory(path)
			}
			err := protectResumeEnvironmentWithSync(f.d, f.wt, f.link.ChildID, f.inflight, sync)
			if failAt == 0 {
				if err != nil {
					t.Fatal(err)
				}
				if len(synced) != 2 || synced[0] != f.inflight || synced[1] != filepath.Dir(f.inflight) {
					t.Fatalf("owner durability order %v", synced)
				}
			} else {
				if err == nil {
					t.Fatal("sync uncertainty authorized lock")
				}
				if _, err := os.Stat(filepath.Join(f.gitdir, "locked")); !os.IsNotExist(err) {
					t.Fatalf("uncertain marker has locked worktree: %v", err)
				}
			}
		})
	}
}

func resumeCompletedEnvironmentPrivate(t *testing.T) (resumeEnvironmentFixture, *resumeParentLock) {
	t.Helper()
	f := resumeEnvironmentPrivate(t)
	p := resumeCoreFixture(t).Parent
	p.Repo = f.d.Repo
	p.SHA = f.link.SHA
	f.link.ParentRecordSHA256 = resumeRecordDigest(retentionWrite(t, f.store, p.ID, p))
	f.link.ParentLogSHA256 = p.LogSHA256
	f.link.Owner = ResumeOwner{PID: os.Getpid(), Host: hostName()}
	var err error
	f.link.Owner.PIDStarted, err = processStartTime(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	f.link.Owner.BootID, err = queueBootID()
	if err != nil {
		t.Fatal(err)
	}
	retentionWrite(t, f.store, f.link.ID, f.link)
	scratch := filepath.Join(f.scratchRoot, fmt.Sprintf("sdlc-run-suite-rerun-%d-123", os.Getpid()))
	if err := os.Rename(f.scratch, scratch); err != nil {
		t.Fatal(err)
	}
	f.scratch = scratch
	f.d.runnerLifetime = &runnerArtifactLifetime{retirable: true}
	lock, err := lockResumeParent(context.Background(), f.store, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lock.Close() })
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	return f, lock
}

func TestResumeEnvironmentCompletedAuthorizationAndAbsence(t *testing.T) {
	for _, mode := range []string{"valid", "foreign-pid", "foreign-start", "foreign-host", "foreign-boot", "unconfirmed-group", "no-group", "changed-parent", "changed-link", "missing-marker", "removal-unconfirmed"} {
		t.Run(mode, func(t *testing.T) {
			f, lock := resumeCompletedEnvironmentPrivate(t)
			switch mode {
			case "foreign-pid":
				f.link.Owner.PID++
			case "foreign-start":
				f.link.Owner.PIDStarted += "changed"
			case "foreign-host":
				f.link.Owner.Host += "changed"
			case "foreign-boot":
				f.link.Owner.BootID += "changed"
			case "unconfirmed-group":
				f.d.runnerLifetime.retirable = false
			case "no-group":
				f.d.runnerLifetime = nil
			case "changed-parent":
				if err := os.WriteFile(lock.path, []byte("changed parent"), 0600); err != nil {
					t.Fatal(err)
				}
			case "changed-link":
				changed := f.link
				changed.Owner.Host += "changed"
				retentionWrite(t, f.store, changed.ID, changed)
			case "missing-marker":
				if err := os.Remove(filepath.Join(f.inflight, runnerArtifactOwnerFile)); err != nil {
					t.Fatal(err)
				}
			case "removal-unconfirmed":
				f.d.CleanupGit = func(args ...string) (string, error) {
					if len(args) > 1 && args[0] == "worktree" && args[1] == "remove" {
						return "unconfirmed", nil
					}
					return f.d.Git(args...)
				}
			}
			err := retireCompletedResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, lock, []string{f.scratchRoot})
			if mode != "valid" {
				if err == nil {
					t.Fatal("unconfirmed completion authorized retirement")
				}
				if mode != "removal-unconfirmed" {
					for _, path := range []string{f.wt, f.scratch} {
						if _, err := os.Stat(path); err != nil {
							t.Errorf("refused completion lost %s: %v", path, err)
						}
					}
				}
				if mode != "missing-marker" {
					if _, err := os.Stat(filepath.Join(f.inflight, runnerArtifactOwnerFile)); err != nil {
						t.Fatal("uncertainty dropped ownership marker", err)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{f.wt, f.gitdir, f.scratch, filepath.Join(f.inflight, runnerArtifactOwnerFile)} {
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Errorf("completed environment retained %s: %v", path, err)
				}
			}
			if raw, err := os.ReadFile(filepath.Join(f.inflight, "run.log")); err != nil || string(raw) != "partial raw evidence" {
				t.Fatalf("completed cleanup lost raw log: %q %v", raw, err)
			}
		})
	}
}

func TestResumeEnvironmentInterruptedRegistrationCleanupRetries(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	failed := false
	f.d.CleanupGit = func(args ...string) (string, error) {
		if !failed && len(args) > 1 && args[0] == "worktree" && args[1] == "remove" {
			failed = true
			return "", errors.New("registration removal interrupted")
		}
		return f.d.Git(args...)
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err == nil {
		t.Fatal("uncertain registration cleanup succeeded")
	}
	if _, err := os.Lstat(f.wt); !os.IsNotExist(err) {
		t.Fatalf("checkout not removed first: %v", err)
	}
	for _, path := range []string{filepath.Join(f.gitdir, "locked"), filepath.Join(f.inflight, runnerArtifactOwnerFile)} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("uncertain cleanup lost authority: %s %v", path, err)
		}
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{f.gitdir, filepath.Join(f.inflight, runnerArtifactOwnerFile)} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("retry retained %s: %v", path, err)
		}
	}
}

func TestResumeEnvironmentPreparedBeforeLockRetainsUnlockedInputs(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	// Crash after durable marker preparation but before the Git lock existed.
	if err := attachRunnerArtifactOwner(f.inflight, f.wt); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{f.inflight, filepath.Dir(f.inflight)} {
		if err := syncResumeDirectory(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{f.wt, f.gitdir, f.scratch, filepath.Join(f.inflight, runnerArtifactOwnerFile)} {
		if _, err := os.Lstat(path); err != nil {
			t.Errorf("unused preparation deleted without child lock %s: %v", path, err)
		}
	}
	for _, path := range []string{f.wt, f.scratch, f.inflight} {
		if runnerArtifactProtected(path) {
			t.Errorf("unlocked preparation pinned ordinary collection: %s", path)
		}
	}
}

func TestResumeEnvironmentSkipsUnrelatedDanglingScratchOwner(t *testing.T) {
	f := resumeEnvironmentPrivate(t)
	if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
		t.Fatal(err)
	}
	unrelated := filepath.Join(f.scratchRoot, "sdlc-run-suite-123-456")
	if err := os.Mkdir(unrelated, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(unrelated, runnerArtifactOwnerFile)
	if err := os.WriteFile(marker, []byte("gitdir: "+filepath.Join(t.TempDir(), "missing-registration")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := retireResumeEnvironmentWithRoots(context.Background(), f.d, f.store, f.link, f.receipt, []string{f.scratchRoot}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatal("unrelated dangling scratch was deleted", err)
	}
}
