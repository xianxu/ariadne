package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

var migrationTestRoots = []string{"workshop/issues", "workshop/history", "workshop/history/issues"}

func migrationTestBody(id, status string) string {
	return "---\nid: " + id + "\nstatus: " + status + "\ncreated: 2026-09-01\n---\n\n# Earlier work\n\n## Problem\n\nPreserve this record.\n"
}

// Keep semantic index contents and flags, filesystem bytes, and all publication
// refs: a refusal may refresh remote-tracking refs, but has no other write grant.
func migrationPreservationSnapshot(t *testing.T, r *trackerRepo, roots ...string) map[string]string {
	t.Helper()
	state := map[string]string{
		"origin":     testfix.Capture(t, r.origin, "for-each-ref", "--format=%(refname) %(objectname)"),
		"local refs": r.git("for-each-ref", "--format=%(refname) %(objectname)", "refs/heads", "refs/tags"),
	}
	for _, root := range append([]string{r.root}, roots...) {
		for _, args := range [][]string{{"rev-parse", "HEAD"}, {"rev-parse", "--symbolic-full-name", "HEAD"}, {"status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignored"}, {"ls-files", "--stage", "-v", "-z"}, {"diff", "--binary"}, {"diff", "--cached", "--binary"}} {
			state[root+":"+strings.Join(args, " ")] = testfix.Capture(t, root, args...)
		}
		for _, dir := range []string{"workshop/issues", "workshop/history"} {
			err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(dir)), func(p string, d os.DirEntry, err error) error {
				if os.IsNotExist(err) {
					return nil
				}
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				info, err := d.Info()
				if err != nil {
					return err
				}
				var raw []byte
				if info.Mode()&os.ModeSymlink != 0 {
					target, err := os.Readlink(p)
					if err != nil {
						return err
					}
					raw = []byte(target)
				} else {
					raw, err = os.ReadFile(p)
					if err != nil {
						return err
					}
				}
				state[p] = fmt.Sprintf("%s:%x", info.Mode(), sha256.Sum256(raw))
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return state
}

func migrationAssertRefusalPreserves(t *testing.T, r *trackerRepo, want string, worktrees ...string) {
	t.Helper()
	before := migrationPreservationSnapshot(t, r, worktrees...)
	stdout, stderr, err := executeSDLCTestCommand("issue", "migrate")
	if err == nil || !strings.Contains(stdout+stderr+fmt.Sprint(err), want) {
		t.Errorf("evidence %q was not named in a refusal: %v\n%s\n%s", want, err, stdout, stderr)
	}
	if after := migrationPreservationSnapshot(t, r, worktrees...); !reflect.DeepEqual(before, after) {
		for key, value := range before {
			if after[key] != value {
				t.Errorf("refusal changed %s:\nbefore %q\nafter %q", key, value, after[key])
			}
		}
		for key := range after {
			if _, ok := before[key]; !ok {
				t.Errorf("refusal created %s", key)
			}
		}
	}
}

func TestMigrationInventoryCompleteCommittedScopes(t *testing.T) {
	for _, root := range migrationTestRoots {
		for _, scope := range []string{"other-clone-remote", "inherited", "unrelated", "detached-worktree"} {
			t.Run(root+"/"+scope, func(t *testing.T) {
				r := emptyInitRepo(t)
				p := root + "/000041-earlier.md"
				status := "done"
				if root == "workshop/issues" {
					status = "open"
				}
				body := migrationTestBody("000041", status)
				var otherWT []string
				switch scope {
				case "other-clone-remote":
					clone := filepath.Join(t.TempDir(), "other-clone")
					testfix.Git(t, "", "clone", "-q", r.origin, clone)
					testfix.Git(t, clone, "switch", "-qc", "archive-only")
					writeRepoFile(t, clone, p, body)
					testfix.Git(t, clone, "add", ".")
					testfix.Git(t, clone, "commit", "-qm", "unpublished remote issue")
					testfix.Git(t, clone, "push", "-q", "origin", "archive-only")
					// No local fetch: this is the original two-clone reviewer reproduction.
				case "inherited":
					writeRepoFile(t, r.root, p, body)
					r.git("add", ".")
					r.git("commit", "-qm", "record before fork")
					r.git("branch", "inherited-record")
					r.git("rm", p)
					r.git("commit", "-qm", "main drops record")
					r.git("push", "-q")
				case "unrelated":
					r.git("switch", "--orphan", "unrelated-record")
					writeRepoFile(t, r.root, p, body)
					r.git("add", ".")
					r.git("commit", "-qm", "independent record")
					r.git("switch", "-q", "main")
				case "detached-worktree":
					wt := filepath.Join(t.TempDir(), "detached worktree")
					r.git("worktree", "add", "-q", "--detach", wt, "main")
					writeRepoFile(t, wt, p, body)
					testfix.Git(t, wt, "add", ".")
					testfix.Git(t, wt, "commit", "-qm", "detached committed record")
					if got := testfix.Capture(t, wt, "status", "--porcelain"); got != "" {
						t.Fatalf("dirty fixture: %s", got)
					}
					otherWT = []string{wt}
				}
				migrationAssertRefusalPreserves(t, r, p, otherWT...)
			})
		}
	}
}

func TestMigrationInventoryNoBranchNameExemption(t *testing.T) {
	for _, ref := range []string{"refs/heads/main", "refs/heads/main-slot3", "refs/heads/issue-tracker", "refs/heads/HEAD", "refs/remotes/origin/HEAD"} {
		t.Run(ref, func(t *testing.T) {
			r := emptyInitRepo(t)
			p := "workshop/history/issues/000041-earlier.md"
			r.git("switch", "-qc", "evidence-builder")
			writeRepoFile(t, r.root, p, migrationTestBody("000041", "done"))
			r.git("add", ".")
			r.git("commit", "-qm", "reserved-name evidence")
			oid := r.git("rev-parse", "HEAD")
			r.git("switch", "-q", "main")
			if ref == "refs/remotes/origin/HEAD" {
				// A real heads/HEAD on origin survives the explicit all-heads refresh.
				r.git("push", "-q", "origin", oid+":refs/heads/HEAD")
			} else {
				r.git("update-ref", ref, oid)
			}
			r.git("branch", "-D", "evidence-builder")
			before := migrationPreservationSnapshot(t, r)
			env, err := openMigrate(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			main, err := env.main.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			in, err := migrationInventory(env, main)
			if err != nil {
				t.Fatalf("inventory failed before named-evidence policy: %v", err)
			}
			refusals := fmt.Sprint(tracker.PlanTrackerMigration(in).Refusals)
			if !strings.Contains(refusals, p) {
				t.Errorf("reserved ref %s hid its archive: %s", ref, refusals)
			}
			if after := migrationPreservationSnapshot(t, r); !reflect.DeepEqual(before, after) {
				t.Error("inventory changed publication or caller state")
			}
		})
	}
}

func TestIssueMigrateImportsFlatHistoryAndReservesIDs(t *testing.T) {
	r := emptyInitRepo(t)
	flat := "workshop/history/000041-flat.md"
	canonical := "workshop/history/issues/000042-canonical.md"
	for _, p := range []string{flat, canonical} {
		id := "000041"
		if p == canonical {
			id = "000042"
		}
		writeRepoFile(t, r.root, p, migrationTestBody(id, "done"))
	}
	r.git("add", ".")
	r.git("commit", "-qm", "published archives")
	r.git("push", "-q")
	digest, out, err := migrateDryRun(t)
	if err != nil || !strings.Contains(out, "cards: 2") {
		t.Fatalf("both archives must import: %v\n%s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v %s", err, stderr)
	}
	for _, p := range []string{flat, canonical} {
		if r.git("show", "origin/main:"+p) != r.git("show", "HEAD:"+p) {
			t.Errorf("archive rewritten: %s", p)
		}
	}
	env, err := openMigrate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	snap, err := env.tracker.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snap.MaxID() != 42 || len(snap.Records()) != 2 {
		t.Fatalf("archive IDs not reserved: max=%d cards=%d", snap.MaxID(), len(snap.Records()))
	}
}

func TestMigrationInventoryArchiveMatchingIgnoresPathButNotBytes(t *testing.T) {
	for _, kind := range []string{"moved-path", "stale-differing", "older-tip-absence"} {
		t.Run(kind, func(t *testing.T) {
			r := emptyInitRepo(t)
			p := "workshop/history/issues/000041-earlier.md"
			body := migrationTestBody("000041", "done")
			if kind == "older-tip-absence" {
				r.git("branch", "older-tip")
			}
			writeRepoFile(t, r.root, p, body)
			r.git("add", ".")
			r.git("commit", "-qm", "published archive")
			r.git("push", "-q")
			if kind != "older-tip-absence" {
				r.git("switch", "-qc", "archive-copy")
				if kind == "moved-path" {
					r.git("mv", p, "workshop/history/000041-renamed.md")
				} else {
					writeRepoFile(t, r.root, p, body+"\nOlder private annotation.\n")
					r.git("add", p)
				}
				r.git("commit", "-qm", "branch archive copy")
				r.git("switch", "-q", "main")
			}
			if kind == "stale-differing" {
				migrationAssertRefusalPreserves(t, r, p)
				return
			}
			if _, out, err := migrateDryRun(t); err != nil {
				t.Fatalf("published archive evidence refused (%s): %v\n%s", kind, err, out)
			}
		})
	}
}

func TestMigrationInventoryRelatedProseAndRenameRemainAllowed(t *testing.T) {
	r := legacyRepo(t)
	r.git("switch", "-qc", "related-details")
	p := "workshop/issues/000001-one.md"
	writeRepoFile(t, r.root, p, r.git("show", "HEAD:"+p)+"\nPrivate design prose.\n")
	r.git("mv", "workshop/issues/000002-two.md", "workshop/issues/000002-renamed.md")
	r.git("add", ".")
	r.git("commit", "-qm", "related prose and rename")
	r.git("switch", "-q", "main")
	if _, out, err := migrateDryRun(t); err != nil {
		t.Fatalf("related active detail edits refused: %v\n%s", err, out)
	}
}

func TestMigrationInventoryDirtyAllRoots(t *testing.T) {
	for _, root := range migrationTestRoots {
		for _, kind := range []string{"untracked", "ignored", "staged-add-missing", "unstaged", "staged-delete", "assume-unchanged", "skip-worktree", "conflict", "symlink-document"} {
			t.Run(root+"/"+kind, func(t *testing.T) {
				r := emptyInitRepo(t)
				p := root + "/000041-earlier.md"
				status := "done"
				if root == "workshop/issues" {
					status = "open"
				}
				body := migrationTestBody("000041", status)
				switch kind {
				case "unstaged", "staged-delete", "assume-unchanged", "skip-worktree", "conflict":
					writeRepoFile(t, r.root, p, body)
					r.git("add", ".")
					r.git("commit", "-qm", "published issue")
					r.git("push", "-q")
				}
				wt := filepath.Join(t.TempDir(), "other worktree")
				r.git("worktree", "add", "-q", "--detach", wt, "main")
				switch kind {
				case "untracked":
					writeRepoFile(t, wt, p, body)
				case "ignored":
					// Shared excludes exercise files invisible even to --untracked-files=all.
					writeRepoFile(t, r.root, ".git/info/exclude", "/"+root+"/000041-earlier.md\n")
					writeRepoFile(t, wt, p, body)
				case "staged-add-missing":
					writeRepoFile(t, wt, p, body)
					testfix.Git(t, wt, "add", p)
					if err := os.Remove(filepath.Join(wt, filepath.FromSlash(p))); err != nil {
						t.Fatal(err)
					}
				case "unstaged":
					writeRepoFile(t, wt, p, body+"\nLocal edit.\n")
				case "staged-delete":
					testfix.Git(t, wt, "rm", "-q", p)
				case "assume-unchanged", "skip-worktree":
					testfix.Git(t, wt, "update-index", "--"+kind, p)
					writeRepoFile(t, wt, p, body+"\nHidden edit.\n")
				case "conflict":
					// Real unmerged stages without relying on a textual merge conflict.
					oid := r.git("rev-parse", "HEAD:"+p)
					testfix.Git(t, wt, "update-index", "--force-remove", p)
					indexInput := "100644 " + oid + " 1\t" + p + "\n100644 " + oid + " 2\t" + p + "\n100644 " + oid + " 3\t" + p + "\n"
					migrationGitStdin(t, wt, indexInput, "update-index", "--index-info")
				case "symlink-document":
					if err := os.MkdirAll(filepath.Dir(filepath.Join(wt, filepath.FromSlash(p))), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(filepath.Join(t.TempDir(), "missing-target"), filepath.Join(wt, filepath.FromSlash(p))); err != nil {
						t.Fatal(err)
					}
				}
				migrationAssertRefusalPreserves(t, r, p, wt)
			})
		}
	}
}

func TestMigrationInventoryNormalizationAndNonIssueAssetsAreClean(t *testing.T) {
	r := emptyInitRepo(t)
	writeRepoFile(t, r.root, ".gitattributes", "*.md text eol=crlf\n")
	for _, root := range migrationTestRoots {
		status := "done"
		if root == "workshop/issues" {
			status = "open"
		}
		id := "000041"
		if root == "workshop/history" {
			id = "000042"
		}
		if root == "workshop/history/issues" {
			id = "000043"
		}
		writeRepoFile(t, r.root, root+"/"+id+"-earlier.md", migrationTestBody(id, status))
	}
	r.git("add", ".")
	r.git("commit", "-qm", "normalized published issues")
	r.git("push", "-q")
	wt := filepath.Join(t.TempDir(), "normalized")
	r.git("worktree", "add", "-q", "--detach", wt, "main")
	for _, root := range migrationTestRoots {
		// Nested issue-shaped attachments and direct non-doc assets are irrelevant.
		writeRepoFile(t, wt, root+"/assets/999999-nested.md", "asset\n")
		writeRepoFile(t, wt, root+"/000099-side.md.side/999999-nested.md", "side bundle\n")
		writeRepoFile(t, wt, root+"/999999-asset.bin", "asset\n")
	}
	raw, err := os.ReadFile(filepath.Join(wt, "workshop/issues/000041-earlier.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "\r\n") {
		t.Fatal("normalization fixture did not materialize CRLF")
	}
	if _, out, err := migrateDryRun(t); err != nil {
		t.Fatalf("normalized/asset-only checkout refused: %v\n%s", err, out)
	}
}

func TestIssueMigrateReobservesEvidenceBeforeApply(t *testing.T) {
	r := emptyInitRepo(t)
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("initial dry run: %v\n%s", err, out)
	}
	wt := filepath.Join(t.TempDir(), "appeared-after-review")
	r.git("worktree", "add", "-q", "--detach", wt, "main")
	p := "workshop/history/issues/000041-earlier.md"
	writeRepoFile(t, wt, p, migrationTestBody("000041", "done"))
	before := migrationPreservationSnapshot(t, r, wt)
	stdout, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest)
	if err == nil || !strings.Contains(stdout+stderr+fmt.Sprint(err), p) {
		t.Errorf("changed evidence did not block apply: %v\n%s\n%s", err, stdout, stderr)
	}
	if after := migrationPreservationSnapshot(t, r, wt); !reflect.DeepEqual(before, after) {
		t.Error("refused apply changed remote or checkout state")
	}
}

func migrationGitStdin(t *testing.T, root, input string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Stdin = strings.NewReader(input)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
