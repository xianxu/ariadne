package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func retentionWrite(t *testing.T, root, id string, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Join(root, "attempts"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "attempts", id+".json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	return raw
}

func retentionFixture(t *testing.T, terminal bool) (string, ResumeLink, time.Time) {
	t.Helper()
	return retentionFixtureOf(t, terminal, resumeCorePassed(t))
}

func retentionFixtureOf(t *testing.T, terminal bool, in ResumeProofInput) (string, ResumeLink, time.Time) {
	t.Helper()
	root := t.TempDir()
	parentLog := []byte("parent failure\n")
	in.Parent.LogPath = "failed/aa.log"
	in.Parent.LogSHA256 = resumeRecordDigest(parentLog)
	in.ParentRecordSHA256 = resumeRecordDigest(retentionWrite(t, root, in.Parent.ID, in.Parent))
	in.ParentLogSHA256 = in.Parent.LogSHA256
	link := *resumeCoreLink(in)
	if in.Link != nil { // keep the caller's check, version and selection
		link.Version, link.Check, link.Selection = in.Link.Version, in.Link.Check, in.Link.Selection
	}
	retentionWrite(t, root, link.ID, link)
	if err := os.MkdirAll(filepath.Join(root, "failed"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "failed", "aa.log"), parentLog, 0600); err != nil {
		t.Fatal(err)
	}
	if terminal {
		in.Child.LogPath = "failed/cc.log"
		retentionWrite(t, root, in.Child.ID, in.Child)
		if err := os.WriteFile(filepath.Join(root, "failed", "cc.log"), []byte(in.ChildLog), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root, link, link.CreatedAt.Add(40 * 24 * time.Hour)
}

func TestResumeRetentionPendingPinsWholeBundle(t *testing.T) {
	root, link, now := retentionFixture(t, false)
	protected, err := pruneResumeBundles(root, now)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{link.RerunOf, link.ID, link.ChildID} {
		if !protected[id] {
			t.Errorf("member %s unprotected", id)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", link.ID+".json")); err != nil {
		t.Fatal(err)
	}
}

func TestResumeRetentionTerminalAgeAndBinding(t *testing.T) {
	for _, mode := range []string{"expired", "recent", "bad-log", "conflict", "orphan", "pending-orphan"} {
		t.Run(mode, func(t *testing.T) {
			root, link, now := retentionFixture(t, mode != "pending-orphan")
			switch mode {
			case "recent":
				now = link.CreatedAt.Add(24 * time.Hour)
			case "bad-log":
				if err := os.WriteFile(filepath.Join(root, "failed", "cc.log"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			case "conflict":
				other := link
				other.ID = "dd"
				retentionWrite(t, root, other.ID, other)
			case "orphan", "pending-orphan":
				if err := os.Remove(filepath.Join(root, "attempts", "aa.json")); err != nil {
					t.Fatal(err)
				}
			}
			_, err := pruneResumeBundles(root, now)
			if err != nil && mode != "conflict" {
				t.Fatal(err)
			}
			_, err = os.Stat(filepath.Join(root, "attempts", link.ID+".json"))
			wantGone := mode == "expired" || mode == "orphan"
			if os.IsNotExist(err) != wantGone {
				t.Fatalf("link existence %v, wantGone %v", err, wantGone)
			}
		})
	}
}

func TestResumeRetentionFaultsKeepFenceWhileParentSelectable(t *testing.T) {
	// Count the successful production path, then fail before each mutation/sync.
	root, _, now := retentionFixture(t, true)
	ops := resumeRetentionIO{remove: os.Remove, removeAll: os.RemoveAll, sync: syncResumeDirectory}
	count := 0
	traced := func(base func(string) error) func(string) error {
		return func(path string) error { count++; return base(path) }
	}
	ops.remove = traced(ops.remove)
	ops.removeAll = traced(ops.removeAll)
	ops.sync = traced(ops.sync)
	if _, err := pruneResumeBundlesWithIO(root, now, ops); err != nil {
		t.Fatal(err)
	}
	if count < 6 {
		t.Fatalf("too few retirement effects: %d", count)
	}
	for failAt := 1; failAt <= count; failAt++ {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			root, link, now := retentionFixture(t, true)
			n := 0
			wrap := func(base func(string) error) func(string) error {
				return func(path string) error {
					n++
					if n == failAt {
						return errors.New("injected IO failure")
					}
					return base(path)
				}
			}
			_, err := pruneResumeBundlesWithIO(root, now, resumeRetentionIO{remove: wrap(os.Remove), removeAll: wrap(os.RemoveAll), sync: wrap(syncResumeDirectory)})
			if err == nil {
				t.Fatalf("fault %d missed (effects %d)", failAt, n)
			}
			if _, e := os.Stat(filepath.Join(root, "attempts", link.RerunOf+".json")); e == nil {
				if _, e := os.Stat(filepath.Join(root, "attempts", link.ID+".json")); e != nil {
					t.Fatalf("selectable parent lost fence: %v", e)
				}
			}
			// Every point is retryable without exposing the original execution again.
			if _, err := pruneResumeBundles(root, now); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(root, "attempts", link.ID+".json")); !os.IsNotExist(err) {
				t.Fatalf("retry left fence: %v", err)
			}
		})
	}
}

func TestResumeRetentionLockedParentAndUnreadableSnapshot(t *testing.T) {
	root, link, now := retentionFixture(t, true)
	lock, err := lockResumeParent(context.Background(), root, link.RerunOf)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err = pruneResumeBundles(root, now)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("pruner waited behind parent owner")
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", link.ID+".json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "attempts", "dd.json"), []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := pruneResumeBundles(root, now); err == nil {
		t.Fatal("incomplete snapshot allowed pruning")
	}
}

func TestResumeRetentionReceiptAndFutureEvidence(t *testing.T) {
	for _, mode := range []string{"receipt", "bad-receipt", "future"} {
		t.Run(mode, func(t *testing.T) {
			root, l, now := retentionFixture(t, false)
			raw, err := os.ReadFile(filepath.Join(root, "attempts", l.ID+".json"))
			if err != nil {
				t.Fatal(err)
			}
			r := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: l.Repo, SHA: l.SHA, Check: l.Check, RerunOf: l.RerunOf, LinkID: l.ID, LinkSHA256: resumeRecordDigest(raw), ChildID: l.ChildID, Actor: "operator", Note: "stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt.Add(time.Hour)}
			if mode == "bad-receipt" {
				r.LinkSHA256 = strings.Repeat("e", 64)
			}
			retentionWrite(t, root, r.ID, r)
			if mode == "future" {
				retentionWrite(t, root, "ee", map[string]any{"version": 99, "id": "ee", "repo": l.Repo, "sha": l.SHA, "check": "suite", "rerun_of": l.RerunOf})
			}
			protected, err := pruneResumeBundles(root, now)
			if mode == "future" {
				if err == nil || !protected["ee"] || !protected[l.RerunOf] {
					t.Fatalf("future scope: %v %v", protected, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = os.Stat(filepath.Join(root, "attempts", l.ID+".json"))
			if err != nil {
				t.Fatalf("receipt binding %s: %v", mode, err)
			}
		})
	}
}

func TestResumeRetentionIndexLinearWork(t *testing.T) {
	counts := []int{}
	for _, n := range []int{100, 200} {
		s := SuiteEvidenceSnapshot{}
		for i := 0; i < n; i++ {
			p := fmt.Sprintf("%04x", i*4+1)
			child := fmt.Sprintf("%04x", i*4+2)
			link := fmt.Sprintf("%04x", i*4+3)
			s.Attempts = append(s.Attempts, SuiteAttempt{ID: p}, SuiteAttempt{ID: child, RerunOf: p})
			s.Resumes = append(s.Resumes, ResumeLink{ID: link, RerunOf: p, ChildID: child})
		}
		x := indexResumeRetention(s)
		counts = append(counts, x.visits)
		if len(x.protected) != 3*n {
			t.Fatalf("protected %d, want %d", len(x.protected), 3*n)
		}
	}
	if counts[1] != 2*counts[0] {
		t.Fatalf("nonlinear visits %v", counts)
	}
}

func TestResumeRetentionSharedMembersAndLogs(t *testing.T) {
	for _, mode := range []string{"member", "log"} {
		t.Run(mode, func(t *testing.T) {
			root, l, now := retentionFixture(t, true)
			if mode == "member" {
				other := l
				other.ID = "ee"
				other.RerunOf = "ff"
				other.ChildID = l.ChildID
				retentionWrite(t, root, other.ID, other)
			} else {
				child := resumeCorePassed(t).Child
				child.ID = "dd"
				child.RerunOf = "ff"
				child.Check = "suite-rerun"
				child.LogPath = "local-checks/" + child.SHA + "/suite-rerun.log"
				// Use the same valid per-head log path for both actual completed records.
				actual := resumeCorePassed(t).Child
				actual.LogPath = child.LogPath
				retentionWrite(t, root, actual.ID, actual)
				retentionWrite(t, root, child.ID, child)
				path := suiteAttemptLogFile(root, *child)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(resumeCorePassed(t).ChildLog), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := pruneResumeBundles(root, now); err != nil && mode != "member" {
				t.Fatal(err)
			}
			if mode == "member" {
				if _, err := os.Stat(filepath.Join(root, "attempts", l.ID+".json")); err != nil {
					t.Fatal("shared member bundle was retired")
				}
			} else {
				child := resumeCorePassed(t).Child
				child.LogPath = "local-checks/" + child.SHA + "/suite-rerun.log"
				if _, err := os.Stat(suiteAttemptLogFile(root, *child)); err != nil {
					t.Fatal("shared log removed", err)
				}
			}
		})
	}
}

func TestResumeRetentionOrphanReceiptRemainsPinned(t *testing.T) {
	root, l, now := retentionFixture(t, false)
	raw, err := os.ReadFile(filepath.Join(root, "attempts", l.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	r := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: l.Repo, SHA: l.SHA, Check: l.Check, RerunOf: l.RerunOf, LinkID: l.ID, LinkSHA256: resumeRecordDigest(raw), ChildID: l.ChildID, Actor: "operator", Note: "stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt.Add(time.Hour)}
	retentionWrite(t, root, r.ID, r)
	for _, id := range []string{l.RerunOf, l.ID} {
		if err := os.Remove(filepath.Join(root, "attempts", id+".json")); err != nil {
			t.Fatal(err)
		}
	}
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", r.ID+".json")); err != nil {
		t.Fatalf("orphan receipt was lost: %v", err)
	}
}

func TestResumeRetentionDurableOrderAndOrphanLog(t *testing.T) {
	root, l, now := retentionFixture(t, true)
	trace := []string{}
	wrap := func(kind string, base func(string) error) func(string) error {
		return func(path string) error { trace = append(trace, kind+":"+filepath.Base(path)); return base(path) }
	}
	_, err := pruneResumeBundlesWithIO(root, now, resumeRetentionIO{remove: wrap("remove", os.Remove), removeAll: wrap("tree", os.RemoveAll), sync: wrap("sync", syncResumeDirectory)})
	if err != nil {
		t.Fatal(err)
	}
	at := func(event string) int {
		for i, e := range trace {
			if e == event {
				return i
			}
		}
		t.Fatalf("missing %s: %v", event, trace)
		return -1
	}
	parent, link, child := at("remove:aa.json"), at("remove:bb.json"), at("remove:cc.json")
	if parent != 0 || trace[parent+1] != "sync:attempts" || trace[link+1] != "sync:attempts" || child < link+2 || at("remove:cc.log") < link+2 {
		t.Fatalf("unsafe durable order: %v", trace)
	}
	// Simulate a crash after parent unlink, before its raw log cleanup.
	root, l, now = retentionFixture(t, true)
	if err := os.Remove(filepath.Join(root, "attempts", l.RerunOf+".json")); err != nil {
		t.Fatal(err)
	}
	if _, err := pruneResumeBundles(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "failed", l.RerunOf+".log")); !os.IsNotExist(err) {
		t.Fatalf("orphan raw parent log remains: %v", err)
	}
}

func TestResumeRetentionRequiresMeasuredCompletion(t *testing.T) {
	for _, which := range []string{"parent", "child", "receipt-parent"} {
		for _, measured := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/measured=%t", which, measured), func(t *testing.T) {
				receipt := which == "receipt-parent"
				root, l, now := retentionFixture(t, !receipt)
				id := l.RerunOf
				if which == "child" {
					id = l.ChildID
				}
				path := filepath.Join(root, "attempts", id+".json")
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var a SuiteAttempt
				if err := json.Unmarshal(raw, &a); err != nil {
					t.Fatal(err)
				}
				a.WallSeconds = nil
				if measured && !receipt {
					a.WallSeconds = new(float64)
				}
				raw = retentionWrite(t, root, id, a)
				if which != "child" {
					l.ParentRecordSHA256 = resumeRecordDigest(raw)
				}
				linkRaw := retentionWrite(t, root, l.ID, l)
				ids := []string{l.RerunOf, l.ID, l.ChildID}
				paths := []string{filepath.Join(root, "attempts", l.RerunOf+".json"), filepath.Join(root, "attempts", l.ID+".json"), filepath.Join(root, "failed", l.RerunOf+".log")}
				if receipt {
					r := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: l.Repo, SHA: l.SHA, Check: l.Check, RerunOf: l.RerunOf, LinkID: l.ID, LinkSHA256: resumeRecordDigest(linkRaw), ChildID: l.ChildID, Actor: "operator", Note: "stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt.Add(time.Hour)}
					retentionWrite(t, root, r.ID, r)
					ids = append(ids, r.ID)
					paths = append(paths, filepath.Join(root, "attempts", r.ID+".json"))
				} else {
					paths = append(paths, filepath.Join(root, "attempts", l.ChildID+".json"), filepath.Join(root, "failed", l.ChildID+".log"))
				}
				before := make(map[string]string, len(paths))
				for _, path := range paths {
					b, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					before[path] = string(b)
					old := now.Add(-40 * 24 * time.Hour)
					if err := os.Chtimes(path, old, old); err != nil {
						t.Fatal(err)
					}
				}
				if err := pruneSuiteAttempts(root, now); err != nil {
					t.Fatal(err)
				}
				for _, path := range paths {
					b, err := os.ReadFile(path)
					if measured && !receipt {
						if !os.IsNotExist(err) {
							t.Errorf("measured zero did not retire %s: %v", path, err)
						}
						continue
					}
					if err != nil {
						t.Errorf("incomplete completion lost %s: %v", path, err)
						continue
					}
					if string(b) != before[path] {
						t.Errorf("incomplete completion changed %s", path)
					}
				}
				if !measured || receipt {
					protected, err := pruneResumeBundles(root, now)
					if err != nil {
						t.Fatal(err)
					}
					for _, id := range ids {
						if !protected[id] {
							t.Errorf("incomplete completion did not protect %s", id)
						}
					}
				}
			})
		}
	}
}

func TestResumeRetentionPinsUnreleasedEnvironmentAuthority(t *testing.T) {
	for _, orphan := range []bool{false, true} {
		for _, mode := range []string{"locked", "malformed", "unreadable", "dangling-marker"} {
			t.Run(fmt.Sprintf("orphan=%t/%s", orphan, mode), func(t *testing.T) {
				f := resumeEnvironmentPrivate(t)
				if err := protectResumeEnvironment(f.d, f.wt, f.link.ChildID, f.inflight); err != nil {
					t.Fatal(err)
				}
				marker := filepath.Join(f.inflight, runnerArtifactOwnerFile)
				switch mode {
				case "malformed":
					if err := os.WriteFile(marker, []byte("unknown owner"), 0600); err != nil {
						t.Fatal(err)
					}
				case "dangling-marker":
					if err := os.Remove(marker); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(filepath.Join(t.TempDir(), "missing-owner"), marker); err != nil {
						t.Fatal(err)
					}
				case "unreadable":
					if err := os.Remove(marker); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(marker, 0700); err != nil {
						t.Fatal(err)
					}
				}
				if orphan {
					if err := os.Remove(filepath.Join(f.store, "attempts", f.link.ID+".json")); err != nil {
						t.Fatal(err)
					}
				}
				if err := pruneSuiteAttempts(f.store, f.receipt.FinishedAt.Add(40*24*time.Hour)); err != nil {
					t.Fatal(err)
				}
				for _, path := range []string{filepath.Join(f.store, "attempts", f.receipt.ID+".json"), filepath.Join(f.inflight, "run.log"), marker} {
					if _, err := os.Lstat(path); err != nil {
						t.Errorf("GC lost unreleased authority %s: %v", path, err)
					}
				}
			})
		}
	}
}

func TestResumeRetentionInterruptedGraphsRemainPermanent(t *testing.T) {
	for _, orphan := range []bool{false, true} {
		t.Run(fmt.Sprintf("orphan=%t", orphan), func(t *testing.T) {
			visits := []int{}
			for _, n := range []int{8, 16} {
				root := t.TempDir()
				base := resumeCoreFixture(t)
				now := base.Parent.FinishedAt.Add(100 * 24 * time.Hour)
				if err := os.MkdirAll(filepath.Join(root, "failed"), 0700); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < n; i++ {
					p := base.Parent
					p.ID = fmt.Sprintf("%08x", i*4+1)
					p.LogPath = filepath.Join("failed", p.ID+".log")
					p.LogSHA256 = resumeRecordDigest([]byte("original failure"))
					l := *resumeCoreLink(base)
					l.RerunOf = p.ID
					l.ID = fmt.Sprintf("%08x", i*4+2)
					l.ChildID = fmt.Sprintf("%08x", i*4+3)
					l.ParentLogSHA256 = p.LogSHA256
					l.ParentRecordSHA256 = resumeRecordDigest(retentionWrite(t, root, p.ID, p))
					raw := retentionWrite(t, root, l.ID, l)
					r := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: fmt.Sprintf("%08x", i*4+4), Repo: l.Repo, SHA: l.SHA, Check: l.Check, RerunOf: p.ID, LinkID: l.ID, LinkSHA256: resumeRecordDigest(raw), ChildID: l.ChildID, Actor: "operator", Note: "stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt.Add(time.Hour)}
					retentionWrite(t, root, r.ID, r)
					if err := os.WriteFile(filepath.Join(root, p.LogPath), []byte("original failure"), 0600); err != nil {
						t.Fatal(err)
					}
					inflight := filepath.Join(root, "inflight", l.ChildID)
					if err := os.MkdirAll(inflight, 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(inflight, "run.log"), []byte("partial rerun"), 0600); err != nil {
						t.Fatal(err)
					}
					marker := filepath.Join(inflight, runnerArtifactOwnerFile)
					switch i % 4 {
					case 1:
						gitdir := filepath.Join(root, "unused-git", p.ID)
						if err := os.MkdirAll(gitdir, 0700); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(marker, []byte("gitdir: "+gitdir+"\n"), 0600); err != nil {
							t.Fatal(err)
						}
					case 2:
						if err := os.WriteFile(marker, []byte("changed owner"), 0600); err != nil {
							t.Fatal(err)
						}
					case 3:
						if err := os.Mkdir(marker, 0700); err != nil {
							t.Fatal(err)
						}
					}
					if orphan {
						for _, id := range []string{p.ID, l.ID} {
							if err := os.Remove(filepath.Join(root, "attempts", id+".json")); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
				if err := filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					old := now.Add(-90 * 24 * time.Hour)
					return os.Chtimes(path, old, old)
				}); err != nil {
					t.Fatal(err)
				}
				before := resumeEnvironmentSnapshot(t, root)
				snapshot, err := readSuiteEvidence(root, testStatsRecordLimit)
				if err != nil {
					t.Fatal(err)
				}
				visits = append(visits, indexResumeRetention(snapshot).visits)
				removals := 0
				wrap := func(base func(string) error) func(string) error {
					return func(path string) error { removals++; return base(path) }
				}
				for repeat := 0; repeat < 3; repeat++ {
					if _, err := pruneResumeBundlesWithIO(root, now, resumeRetentionIO{remove: wrap(os.Remove), removeAll: wrap(os.RemoveAll), sync: syncResumeDirectory}); err != nil {
						t.Fatal(err)
					}
					if err := pruneSuiteAttempts(root, now); err != nil {
						t.Fatal(err)
					}
				}
				if removals != 0 {
					t.Errorf("%d retained receipts caused %d removals", n, removals)
				}
				if !reflect.DeepEqual(before, resumeEnvironmentSnapshot(t, root)) {
					t.Error("permanently retained graph changed bytes or paths")
				}
			}
			if visits[1] != 2*visits[0] {
				t.Fatalf("nonlinear retained-graph visits %v", visits)
			}
		})
	}
}

func TestResumeRetentionAgedPendingGraphZeroEffects(t *testing.T) {
	root, l, now := retentionFixture(t, false)
	if err := filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		old := now.Add(-100 * 24 * time.Hour)
		return os.Chtimes(path, old, old)
	}); err != nil {
		t.Fatal(err)
	}
	before := resumeEnvironmentSnapshot(t, root)
	removals := 0
	remove := func(string) error { removals++; return errors.New("pending graph removal forbidden") }
	for repeat := 0; repeat < 3; repeat++ {
		protected, err := pruneResumeBundlesWithIO(root, now, resumeRetentionIO{remove: remove, removeAll: remove, sync: syncResumeDirectory})
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{l.ID, l.RerunOf, l.ChildID} {
			if !protected[id] {
				t.Errorf("pending graph lost ID protection %s", id)
			}
		}
		if err := pruneSuiteAttempts(root, now); err != nil {
			t.Fatal(err)
		}
	}
	if removals != 0 || !reflect.DeepEqual(before, resumeEnvironmentSnapshot(t, root)) {
		t.Fatalf("aged pending graph mutated: removals=%d", removals)
	}
}
