package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestMigrationInventoryRefusesUnknownRootsAndFileKinds(t *testing.T) {
	for _, root := range migrationTestRoots {
		for _, scope := range []string{"main", "worktree"} {
			for _, kind := range []string{"root-file", "root-symlink", "document-symlink", "document-directory", "document-gitlink"} {
				t.Run(root+"/"+scope+"/"+kind, func(t *testing.T) {
					r := emptyInitRepo(t)
					at := r.root
					var worktrees []string
					if scope == "worktree" {
						at = filepath.Join(t.TempDir(), "uncertain worktree")
						r.git("worktree", "add", "-q", "--detach", at, "main")
						worktrees = []string{at}
					}
					p := root + "/000041-earlier.md"
					absRoot, absPath := filepath.Join(at, filepath.FromSlash(root)), filepath.Join(at, filepath.FromSlash(p))
					switch kind {
					case "root-file", "root-symlink":
						// Only disposable fixture roots are replaced; the active scaffold's
						// tracked .gitkeep intentionally becomes a staged deletion on main.
						if err := os.RemoveAll(absRoot); err != nil {
							t.Fatal(err)
						}
						if err := os.MkdirAll(filepath.Dir(absRoot), 0755); err != nil {
							t.Fatal(err)
						}
						if kind == "root-file" {
							writeRepoFile(t, at, root, "uncertain root\n")
						} else {
							target := t.TempDir()
							writeRepoFile(t, target, "000041-earlier.md", migrationTestBody("000041", "done"))
							if err := os.Symlink(target, absRoot); err != nil {
								t.Fatal(err)
							}
						}
					case "document-symlink":
						if err := os.MkdirAll(absRoot, 0755); err != nil {
							t.Fatal(err)
						}
						target := filepath.Join(t.TempDir(), "record.md")
						writeRepoFile(t, filepath.Dir(target), filepath.Base(target), migrationTestBody("000041", "done"))
						if err := os.Symlink(target, absPath); err != nil {
							t.Fatal(err)
						}
					case "document-directory":
						writeRepoFile(t, at, p+"/payload.txt", "not a document\n")
					case "document-gitlink":
						oid := r.git("rev-parse", "HEAD")
						testfix.Git(t, at, "update-index", "--add", "--cacheinfo", "160000,"+oid+","+p)
					}
					if scope == "main" {
						if kind != "document-gitlink" {
							r.git("add", "-A")
						}
						r.git("commit", "-qm", "uncertain inventory kind")
						r.git("push", "-q")
					}
					want := p
					if strings.HasPrefix(kind, "root-") {
						want = root
					}
					migrationAssertRefusalPreserves(t, r, want, worktrees...)
				})
			}
		}
	}
}

func TestMigrationInventoryUnreadableWorktreeRoot(t *testing.T) {
	for _, root := range migrationTestRoots {
		t.Run(root, func(t *testing.T) {
			r := emptyInitRepo(t)
			wt := filepath.Join(t.TempDir(), "unreadable")
			r.git("worktree", "add", "-q", "--detach", wt, "main")
			abs := filepath.Join(wt, filepath.FromSlash(root))
			if err := os.MkdirAll(abs, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(abs, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(abs, 0755) })
			if entries, err := os.ReadDir(abs); err == nil {
				_ = entries
				t.Skip("environment can read mode-000 directories")
			}
			// The byte snapshot helper must not try to read the intentionally unreadable
			// directory. Git state and refs remain independently observable here.
			snapshot := func() []string {
				return []string{testfix.Capture(t, r.origin, "for-each-ref", "--format=%(refname) %(objectname)"), r.git("rev-parse", "HEAD"), r.git("ls-files", "--stage", "-v", "-z"), testfix.Capture(t, wt, "rev-parse", "HEAD"), testfix.Capture(t, wt, "ls-files", "--stage", "-v", "-z")}
			}
			before := snapshot()
			stdout, stderr, err := executeSDLCTestCommand("issue", "migrate")
			if err == nil || !strings.Contains(stdout+stderr+err.Error(), root) {
				t.Errorf("unreadable root authorized migration: %v\n%s\n%s", err, stdout, stderr)
			}
			if !reflect.DeepEqual(before, snapshot()) {
				t.Error("unreadable-root refusal changed refs or index")
			}
		})
	}
}

func TestMigrationInventoryCancelledFilesystemObservation(t *testing.T) {
	r := emptyInitRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	env, err := openMigrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := issueFilesystemFiles(env, r.root, migrationTestRoots); err == nil {
		t.Fatal("cancelled inventory became an empty observation")
	}
}

// Git paths are NUL-framed. Legal filename bytes must survive classification,
// observation and diagnostics instead of turning into false absence.
func TestMigrationInventoryPreservesUnusualFilenameBytes(t *testing.T) {
	for _, root := range migrationTestRoots {
		for _, slug := range []string{"back\\slash", "tab\tname", "line\nname"} {
			t.Run(root+"/"+slug, func(t *testing.T) {
				r := emptyInitRepo(t)
				p := root + "/000041-" + slug + ".md"
				status := "done"
				if root == "workshop/issues" {
					status = "open"
				}
				writeRepoFile(t, r.root, p, migrationTestBody("000041", status))
				r.git("add", ".")
				r.git("commit", "-qm", "record with unusual filename")
				r.git("branch", "inherited-unusual-record")
				r.git("rm", "--", p)
				r.git("commit", "-qm", "main drops unusual record")
				r.git("push", "-q")
				migrationAssertRefusalPreserves(t, r, p)
			})
		}
	}
}
