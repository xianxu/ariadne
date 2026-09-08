package gitx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// trunkFixture builds a repo with a local bare origin, seeds `queue.md` on main,
// and pushes. Returns (repo, origin).
//
// ARCH-MOCK: git is the external binary and a real throwaway repo with a real
// bare origin is its portable stateful fake. A function-call mock cannot produce
// a genuine non-fast-forward rejection, which is the behavior that matters most
// in this file.
func trunkFixture(t *testing.T, seed string) (string, string) {
	t.Helper()
	repo := testfix.Repo(t, testfix.InitialCommit())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, repo, "remote", "add", "origin", origin)
	if err := os.WriteFile(filepath.Join(repo, "queue.md"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "queue.md")
	testfix.Git(t, repo, "commit", "-q", "-m", "seed")
	testfix.Git(t, repo, "push", "-q", "-u", "origin", "main")
	return repo, origin
}

// showTrunk reads a path out of the bare origin's main.
func showTrunk(t *testing.T, origin, path string) string {
	t.Helper()
	return testfix.Capture(t, origin, "show", "main:"+path)
}

// TestNewTrunkFile_RefusesEmptyDir is written BEFORE the feature it guards.
//
// The dangerous failure mode for this type is silent operation on the process
// cwd: gitx's other git shim (`run`, window.go) carries no Dir, so a TrunkFile
// that forgot to pass one would fetch from — and push `<commit>:main` to — the
// real ariadne origin during `go test`. A guard makes that impossible rather
// than merely avoided by convention.
func TestNewTrunkFile_RefusesEmptyDir(t *testing.T) {
	if _, err := NewTrunkFile("", "origin", "main"); err == nil {
		t.Fatal("empty dir must be refused — otherwise git runs against the process cwd")
	}
}

// TestTrunkFile_ReadsFromTrunkNotWorktree pins the whole point of the type: the
// trunk is the base, never the checkout.
//
// Deliberately does NOT testfix.Chdir(): `dir` carries the scoping, so a
// cwd-dependent implementation fails here instead of passing in CI while
// destroying a developer's checkout.
func TestTrunkFile_ReadsFromTrunkNotWorktree(t *testing.T) {
	repo, _ := trunkFixture(t, "from main\n")
	testfix.Git(t, repo, "checkout", "-q", "-b", "feature")
	if err := os.WriteFile(filepath.Join(repo, "queue.md"), []byte("from branch\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	got, err := tf.Read("queue.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "from main\n" {
		t.Errorf("Read = %q, want the TRUNK copy %q", got, "from main\n")
	}
}

// A path absent from the trunk is not an error — it is an empty file. Without
// this the first-ever write to a new queue would have to special-case a git
// error string, which is the kind of thing that breaks on a git upgrade.
func TestTrunkFile_ReadMissingPathIsEmptyNotError(t *testing.T) {
	repo, _ := trunkFixture(t, "x\n")
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	got, err := tf.Read("nope.md")
	if err != nil {
		t.Fatalf("missing path must not error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Read of a missing path = %q, want empty", got)
	}
}

// runGitIn is the seam the capability audit added. These assert the three
// parameters gitx's existing `run` shim does NOT carry, each of which the
// plumbing needs: Dir, Env, and stderr on failure.
func TestRunGitIn_CarriesDirEnvAndStderr(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())

	// Dir: rev-parse resolves against the named repo, not the process cwd.
	out, _, err := runGitIn(repo, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatalf("Dir: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); !strings.HasSuffix(got, filepath.Base(repo)) {
		t.Errorf("Dir not honored: toplevel = %q, want it under %q", got, repo)
	}

	// Env: GIT_INDEX_FILE reaches the child.
	idx := filepath.Join(t.TempDir(), "theindex")
	if _, errOut, err := runGitIn(repo, []string{"GIT_INDEX_FILE=" + idx}, "read-tree", "HEAD"); err != nil {
		t.Fatalf("Env: %v\n%s", err, errOut)
	}
	if _, err := os.Stat(idx); err != nil {
		t.Errorf("GIT_INDEX_FILE did not reach the child: %v", err)
	}

	// stderr: git's own message survives a failure. `run` (.Output()) drops it,
	// and three call sites render it — the retry exhaustion message, the offline
	// warning, and the refused-intent report.
	out, errOut, err := runGitIn(repo, nil, "cat-file", "blob", "main:definitely-absent")
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(string(errOut), "definitely-absent") {
		t.Errorf("stderr dropped — got %q, want git's own message", errOut)
	}
	if len(out) != 0 {
		t.Errorf("stdout must stay separable from stderr, got %q", out)
	}
}

// Update writes to the trunk from a checkout that is NOT on main, with no
// worktree on main anywhere, and leaves the working tree untouched.
func TestTrunkFile_UpdateWithNoMainWorktree(t *testing.T) {
	repo, origin := trunkFixture(t, "- a\n")
	testfix.Git(t, repo, "checkout", "-q", "-b", "feature")

	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	err = tf.Update("queue.md", "queue: add b", func(old []byte) ([]byte, error) {
		return append(append([]byte{}, old...), []byte("- b\n")...), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := showTrunk(t, origin, "queue.md"); got != "- a\n- b\n" {
		t.Errorf("trunk = %q, want %q", got, "- a\n- b\n")
	}
	if dirty := testfix.Capture(t, repo, "status", "--porcelain"); dirty != "" {
		t.Errorf("working tree touched: %q", dirty)
	}
	if br := strings.TrimSpace(testfix.Capture(t, repo, "branch", "--show-current")); br != "feature" {
		t.Errorf("branch changed to %q", br)
	}
}

// THE load-bearing test for M1.
//
// A peer lands a commit between our read and our push. The push is rejected
// non-fast-forward, and recovery must RE-RUN THE TRANSFORM against the peer's
// content — not re-push our own bytes, which would silently drop their line.
// That distinction is the whole design (ARCH-ORDER), and only a real bare origin
// can produce the rejection that exercises it (ARCH-MOCK).
func TestTrunkFile_RetryReRunsTransformOnMovedBase(t *testing.T) {
	repo, origin := trunkFixture(t, "- a\n")
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}

	calls := 0
	err = tf.Update("queue.md", "queue: add mine", func(old []byte) ([]byte, error) {
		calls++
		if calls == 1 {
			pushPeerLine(t, origin, "- a\n- peer\n") // peer wins the race
		}
		return append(append([]byte{}, old...), []byte("- mine\n")...), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	got := showTrunk(t, origin, "queue.md")
	for _, want := range []string{"- peer\n", "- mine\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("trunk missing %q:\n%s", want, got)
		}
	}
	if calls != 2 {
		t.Errorf("transform called %d times, want 2 (re-run against the moved base)", calls)
	}
}

// Exhaustion surfaces GIT'S OWN rejection text, reachable only because runGitIn
// returns stderr separately. Asserting on a generic wrapper message would pass
// against a shim that dropped stderr and prove nothing.
func TestTrunkFile_RetryExhaustionSurfacesGitRejection(t *testing.T) {
	repo, origin := trunkFixture(t, "- a\n")
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}

	calls := 0
	err = tf.Update("queue.md", "queue: doomed", func(old []byte) ([]byte, error) {
		calls++
		pushPeerLine(t, origin, fmt.Sprintf("- a\n- peer%d\n", calls)) // peer always wins
		return append(append([]byte{}, old...), []byte("- mine\n")...), nil
	})
	if err == nil {
		t.Fatal("expected refusal after the attempt budget")
	}
	if calls != 3 {
		t.Errorf("transform called %d times, want 3 (the bound)", calls)
	}
	if !strings.Contains(err.Error(), "non-fast-forward") && !strings.Contains(err.Error(), "fetch first") {
		t.Errorf("error must carry git's own rejection text, got: %v", err)
	}
}

// A transform error aborts without touching the trunk.
func TestTrunkFile_TransformErrorLeavesTrunkUntouched(t *testing.T) {
	repo, origin := trunkFixture(t, "- a\n")
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	boom := errors.New("boom")
	if err := tf.Update("queue.md", "m", func([]byte) ([]byte, error) { return nil, boom }); !errors.Is(err, boom) {
		t.Fatalf("want the transform's error, got %v", err)
	}
	if got := showTrunk(t, origin, "queue.md"); got != "- a\n" {
		t.Errorf("trunk changed on a failed transform: %q", got)
	}
}

// The temp index is absolute and is removed even when a git step INSIDE
// commitAndPush fails.
//
// The first version of this test aborted in the transform, which runs BEFORE any
// index is created — so GIT_INDEX_FILE was never set, the captured path stayed
// empty, and every assertion sat inside a branch that could not be taken. A
// cleanup test has to fail somewhere cleanup is actually owed.
func TestTrunkFile_TempIndexIsAbsoluteAndRemovedOnGitFailure(t *testing.T) {
	repo, _ := trunkFixture(t, "- a\n")
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}

	var seen string
	orig := runGitIn
	runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
		for _, e := range env {
			if strings.HasPrefix(e, "GIT_INDEX_FILE=") {
				seen = strings.TrimPrefix(e, "GIT_INDEX_FILE=")
			}
		}
		if len(args) > 0 && args[0] == "write-tree" {
			return nil, []byte("injected failure"), errors.New("write-tree failed")
		}
		return orig(dir, env, args...)
	}
	defer func() { runGitIn = orig }()

	if err := tf.Update("queue.md", "m", func(o []byte) ([]byte, error) {
		return append(o, []byte("- b\n")...), nil
	}); err == nil {
		t.Fatal("expected the injected git failure to surface")
	}

	if seen == "" {
		t.Fatal("GIT_INDEX_FILE was never set — the test would assert nothing")
	}
	if !filepath.IsAbs(seen) {
		t.Errorf("GIT_INDEX_FILE must be absolute (it resolves against cmd.Dir otherwise), got %q", seen)
	}
	if _, err := os.Stat(seen); !os.IsNotExist(err) {
		t.Errorf("temp index %s survived a mid-sequence git failure", seen)
	}
}

// BR-5: the path M2's seed actually takes — Update on a file absent from the
// trunk. Nothing else in this file exercises a first-ever write.
func TestTrunkFile_UpdateCreatesAbsentPath(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	err = tf.Update("workshop/queue.md", "queue: create", func(old []byte) ([]byte, error) {
		if len(old) != 0 {
			t.Errorf("absent path must read as empty, got %q", old)
		}
		return []byte("- first\n"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := showTrunk(t, origin, "workshop/queue.md"); got != "- first\n" {
		t.Errorf("created file = %q, want %q", got, "- first\n")
	}
	// The pre-existing file must survive: read-tree seeds the index from the
	// base tree, so a create must not blow away siblings.
	if got := showTrunk(t, origin, "queue.md"); got != "seed\n" {
		t.Errorf("sibling file lost on create: %q", got)
	}
}

// BR-2: a push refusal that retrying cannot fix must surface immediately rather
// than burning the attempt budget and reporting contention that never happened.
func TestTrunkFile_NonRetryablePushFailsFast(t *testing.T) {
	repo, _ := trunkFixture(t, "- a\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")

	calls := 0
	orig := runGitIn
	runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
		if len(args) > 0 && args[0] == "push" {
			// git's wording for refusals retrying cannot fix; the trunk does NOT move.
			return nil, []byte("! [remote rejected] main -> main (pre-receive hook declined)"),
				errors.New("exit status 1")
		}
		return orig(dir, env, args...)
	}
	defer func() { runGitIn = orig }()

	err := tf.Update("queue.md", "m", func(o []byte) ([]byte, error) {
		calls++
		return append(o, 'z', '\n'), nil
	})
	if err == nil {
		t.Fatal("expected the refusal to surface")
	}
	if calls != 1 {
		t.Errorf("transform called %d times, want 1 — the trunk never moved, so this is not a CAS failure", calls)
	}
	if !strings.Contains(err.Error(), "hook declined") {
		t.Errorf("must surface git's actual reason, got: %v", err)
	}
}

// pushPeerLine simulates another checkout landing a commit on the trunk.
func pushPeerLine(t *testing.T, origin, content string) {
	t.Helper()
	peer := testfix.Repo(t)
	testfix.Git(t, peer, "remote", "add", "origin", origin)
	testfix.Git(t, peer, "fetch", "-q", "origin", "main")
	testfix.Git(t, peer, "checkout", "-q", "-B", "main", "origin/main")
	if err := os.WriteFile(filepath.Join(peer, "queue.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, peer, "add", "queue.md")
	testfix.Git(t, peer, "commit", "-q", "-m", "peer")
	testfix.Git(t, peer, "push", "-q", "origin", "main")
}

// The pushed blob must check out byte-identical to what the caller intended,
// with .gitattributes applied. This is what `hash-object --path` buys: without
// it, a repo normalizing EOL for *.md stores bytes whose checkout differs from
// the file we thought we published.
func TestTrunkFile_RoundTripsUnderGitattributes(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	if err := os.WriteFile(filepath.Join(repo, ".gitattributes"), []byte("*.md text eol=lf\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", ".gitattributes")
	testfix.Git(t, repo, "commit", "-q", "-m", "attrs")
	testfix.Git(t, repo, "push", "-q", "origin", "main")

	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	want := "- a\n- b\n"
	if err := tf.Update("queue.md", "m", func([]byte) ([]byte, error) {
		return []byte("- a\r\n- b\r\n"), nil // CRLF in, LF expected out
	}); err != nil {
		t.Fatal(err)
	}
	if got := showTrunk(t, origin, "queue.md"); got != want {
		t.Errorf("round-trip = %q, want %q (gitattributes not applied)", got, want)
	}
}

// Offline: a READ degrades to the stale tracking ref and says so; a WRITE
// refuses, because a CAS push has no base to compare against. Same asymmetry
// issueids.go settled for id allocation — one offline policy in this binary,
// not two (ARCH-DRY).
func TestTrunkFile_OfflineDegradesReadRefusesWrite(t *testing.T) {
	repo, _ := trunkFixture(t, "- stale\n")
	testfix.Git(t, repo, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))

	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}

	got, warn, err := tf.ReadDegraded("queue.md")
	if err != nil {
		t.Fatalf("offline read must degrade, not fail: %v", err)
	}
	if string(got) != "- stale\n" {
		t.Errorf("degraded read = %q, want the stale ref's content", got)
	}
	if warn == "" || !strings.Contains(warn, "queue.md") {
		t.Errorf("degraded read must warn loudly and name the risk, got %q", warn)
	}

	err = tf.Update("queue.md", "m", func(o []byte) ([]byte, error) { return o, nil })
	if err == nil {
		t.Fatal("offline write must refuse — a CAS push has nothing to compare against")
	}
	if !strings.Contains(err.Error(), "offline") && !strings.Contains(err.Error(), "unreachable") {
		t.Errorf("refusal must name the cause, got: %v", err)
	}
}

// A repo with no origin at all reads its local ref and says so, rather than
// failing with a git error the operator has to decode.
func TestTrunkFile_NoRemoteIsNamed(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	_, warn, err := tf.ReadDegraded("README")
	if err != nil {
		t.Fatalf("no-remote read must not error: %v", err)
	}
	if !strings.Contains(warn, "origin") {
		t.Errorf("warning must name the missing remote, got %q", warn)
	}
}

// FirstLine is pure, so it gets a unit test rather than being exercised only
// through integration paths — the boundary the package claims should be visible
// from outside (ARCH-PURE).
func TestFirstLine(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", "no output"},
		{"   \n\n", "no output"},
		{"only line", "only line"},
		{"\n\n  real cause  \nnoise\nmore noise", "real cause"},
		{"fatal: x\nhint: y", "fatal: x"},
	} {
		if got := FirstLine(tc.in); got != tc.want {
			t.Errorf("FirstLine(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// The temp index lives in a private directory, not a deleted-then-recreated
// path in shared /tmp — there must be no window in which the name is unclaimed
// and a symlink could be planted to redirect the write (ARCH-SECURE).
func TestTempIndexPath_NoUnclaimedNameWindow(t *testing.T) {
	p, cleanup, err := tempIndexPath()
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(p) {
		t.Errorf("index path must be absolute, got %q", p)
	}
	dir := filepath.Dir(p)
	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("containing dir must exist so the name inside it is ours: %v", err)
	}
	if perm := fi.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("containing dir mode = %o, want no group/other access", perm)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("the index file itself must not pre-exist; git creates it")
	}
	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("cleanup must remove the whole directory")
	}
}

// commit-tree does not consult commit.gpgsign on its own, so this path must read
// it — and git stores the value VERBATIM, so `yes`/`1`/`on` are all true and a
// raw == "true" compare silently produces unsigned commits. Table includes the
// negative cases, which nothing covered when this was two near-identical tests.
func TestTrunkFile_SigningFollowsGitBoolSemantics(t *testing.T) {
	for _, tc := range []struct {
		cfg  string
		want bool
	}{
		{"true", true}, {"yes", true}, {"1", true}, {"on", true},
		{"false", false}, {"0", false}, {"off", false},
	} {
		t.Run(tc.cfg, func(t *testing.T) {
			repo, _ := trunkFixture(t, "x\n")
			testfix.Git(t, repo, "config", "commit.gpgsign", tc.cfg)

			var sawDashS bool
			orig := runGitIn
			runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
				if len(args) > 0 && args[0] == "commit-tree" {
					var kept []string
					for _, a := range args {
						if a == "-S" {
							sawDashS = true
							continue
						}
						kept = append(kept, a)
					}
					args = kept
				}
				return orig(dir, env, args...)
			}
			defer func() { runGitIn = orig }()

			tf, _ := NewTrunkFile(repo, "origin", "main")
			if err := tf.Update("queue.md", "m", func(o []byte) ([]byte, error) {
				return append(o, 'y', '\n'), nil
			}); err != nil {
				t.Fatal(err)
			}
			if sawDashS != tc.want {
				t.Errorf("commit.gpgsign=%q: signed=%v, want %v", tc.cfg, sawDashS, tc.want)
			}
		})
	}
}

// THE regression test for the wipe.
//
// When the presence check cannot tell whether the path exists — a broken or
// unavailable object, not an absent one — the read must FAIL. The two-state
// versions returned "absent", so the transform received empty content, appended
// to nothing, and Update published a file containing only the new line: the trunk
// silently truncated, reported as success. Data loss with a green exit code.
func TestTrunkFile_UnreadablePathRefusesRatherThanTruncating(t *testing.T) {
	repo, origin := trunkFixture(t, "- a\n- b\n- c\n")

	orig := runGitIn
	runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
		if len(args) > 0 && args[0] == "ls-tree" {
			return nil, []byte("fatal: unable to read object"), errors.New("exit status 128")
		}
		return orig(dir, env, args...)
	}
	defer func() { runGitIn = orig }()

	tf, _ := NewTrunkFile(repo, "origin", "main")
	called := false
	err := tf.Update("queue.md", "queue: add d", func(old []byte) ([]byte, error) {
		called = true
		return append(append([]byte{}, old...), []byte("- d\n")...), nil
	})

	if err == nil {
		t.Fatal("an unreadable path must refuse, not publish")
	}
	if called {
		t.Error("the transform must not run on content we could not read — that is how the truncation got published")
	}
	if got := showTrunk(t, origin, "queue.md"); got != "- a\n- b\n- c\n" {
		t.Errorf("trunk was modified despite an unreadable base: %q", got)
	}
}

// A contended update must not fetch twice per attempt: the fetch that decides
// retryability is the same one the next attempt needs for its base.
func TestTrunkFile_OneFetchPerAttempt(t *testing.T) {
	repo, origin := trunkFixture(t, "- a\n")

	fetches := 0
	orig := runGitIn
	runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
		if len(args) > 0 && args[0] == "fetch" {
			fetches++
		}
		return orig(dir, env, args...)
	}
	defer func() { runGitIn = orig }()

	tf, _ := NewTrunkFile(repo, "origin", "main")
	calls := 0
	if err := tf.Update("queue.md", "m", func(old []byte) ([]byte, error) {
		calls++
		if calls == 1 {
			pushPeerLine(t, origin, "- a\n- peer\n")
		}
		return append(append([]byte{}, old...), []byte("- mine\n")...), nil
	}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("transform called %d times, want 2", calls)
	}
	// 2 attempts: one fetch to open, one after the rejection which the retry reuses.
	if fetches > 2 {
		t.Errorf("%d fetches for 2 attempts — the post-push fetch should serve as the retry's base", fetches)
	}
}
