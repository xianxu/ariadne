package gitx

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// patchFixture is the #304 Task 3 fixture: main carries a.txt and shared.txt;
// branch feat changes a.txt (commit H1); main then advances with foreign.txt and
// an edit at the END of shared.txt (non-overlapping with the branch patch).
// The working tree is left on feat at H1, cwd inside the repo.
type patchFixture struct {
	dir, base, h1, mainTip string
}

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAll(t *testing.T, dir string, msg ...string) string {
	t.Helper()
	testfix.Git(t, dir, "add", "-A")
	args := []string{"commit", "-q"}
	for _, m := range msg {
		args = append(args, "-m", m)
	}
	testfix.Git(t, dir, args...)
	return revParse(t, dir, "HEAD")
}

func revParse(t *testing.T, dir, rev string) string {
	t.Helper()
	return strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", rev))
}

const sharedBody = "s1\ns2\ns3\ns4\ns5\ns6\ns7\ns8\n"

func patchRepo(t *testing.T) patchFixture {
	t.Helper()
	dir := testfix.Repo(t, testfix.Chdir())
	writeFile(t, dir, "a.txt", "a1\na2\na3\n")
	writeFile(t, dir, "shared.txt", sharedBody)
	base := commitAll(t, dir, "base")

	testfix.Git(t, dir, "checkout", "-q", "-b", "feat")
	writeFile(t, dir, "a.txt", "a1\nBRANCH\na3\n")
	h1 := commitAll(t, dir, "#304 M1: change a")

	testfix.Git(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, "foreign.txt", "foreign\n")
	writeFile(t, dir, "shared.txt", sharedBody+"MAIN-END\n")
	mainTip := commitAll(t, dir, "main advances")

	testfix.Git(t, dir, "checkout", "-q", "feat")
	return patchFixture{dir: dir, base: base, h1: h1, mainTip: mainTip}
}

func diffNames(t *testing.T, dir, a, b string, paths ...string) []string {
	t.Helper()
	args := append([]string{"diff", "--name-only", a, b, "--"}, paths...)
	var names []string
	for _, n := range strings.Split(strings.TrimSpace(testfix.Capture(t, dir, args...)), "\n") {
		if n != "" {
			names = append(names, n)
		}
	}
	return names
}

func mustRebased(t *testing.T, reviewed string) (string, []string) {
	t.Helper()
	s, conflicted, err := RebasedReviewedBase("main", reviewed)
	if err != nil {
		t.Fatalf("RebasedReviewedBase: %v", err)
	}
	return s, conflicted
}

// (a) B_now == B_r: S reduces to the reviewed tree on the reviewed base.
func TestRebasedReviewedBase_NoIntegration(t *testing.T) {
	f := patchRepo(t)
	s, conflicted := mustRebased(t, f.h1)
	if len(conflicted) != 0 {
		t.Fatalf("conflicted = %v, want none", conflicted)
	}
	if got, want := revParse(t, f.dir, s+"^{tree}"), revParse(t, f.dir, f.h1+"^{tree}"); got != want {
		t.Errorf("tree(S) = %s, want tree(H1) %s", got, want)
	}
	if got := revParse(t, f.dir, s+"^"); got != f.base {
		t.Errorf("parent(S) = %s, want B_r %s", got, f.base)
	}
}

// (b) after merging main, nothing new since the review.
func TestRebasedReviewedBase_AfterMergeEmpty(t *testing.T) {
	f := patchRepo(t)
	testfix.Git(t, f.dir, "merge", "-q", "--no-edit", "main")
	s, _ := mustRebased(t, f.h1)
	if names := diffNames(t, f.dir, s, "HEAD"); len(names) != 0 {
		t.Errorf("diff S HEAD = %v, want empty", names)
	}
}

// (c) after rebasing onto main (H1 itself is no longer on the branch), still empty.
func TestRebasedReviewedBase_AfterRebaseEmpty(t *testing.T) {
	f := patchRepo(t)
	testfix.Git(t, f.dir, "rebase", "-q", "main")
	if revParse(t, f.dir, "HEAD") == f.h1 {
		t.Fatal("fixture: rebase did not rewrite H1")
	}
	s, _ := mustRebased(t, f.h1)
	if names := diffNames(t, f.dir, s, "HEAD"); len(names) != 0 {
		t.Errorf("diff S HEAD = %v, want empty", names)
	}
}

// (d) new work after the merge is exactly the interdiff.
func TestRebasedReviewedBase_NewWorkAfterMerge(t *testing.T) {
	f := patchRepo(t)
	testfix.Git(t, f.dir, "merge", "-q", "--no-edit", "main")
	writeFile(t, f.dir, "b.txt", "b\n")
	commitAll(t, f.dir, "#304 M2: add b")
	s, _ := mustRebased(t, f.h1)
	if names := diffNames(t, f.dir, s, "HEAD"); !reflect.DeepEqual(names, []string{"b.txt"}) {
		t.Errorf("diff S HEAD = %v, want [b.txt]", names)
	}
}

// (e) main rewrote the reviewed line; the resolution shows in the interdiff.
func TestRebasedReviewedBase_ConflictResolutionInterdiff(t *testing.T) {
	f := patchRepo(t)
	testfix.Git(t, f.dir, "checkout", "-q", "main")
	writeFile(t, f.dir, "a.txt", "a1\nMAIN\na3\n")
	commitAll(t, f.dir, "main rewrites a2")
	testfix.Git(t, f.dir, "checkout", "-q", "feat")
	if _, err := runEnv(nil, "git", "merge", "-q", "--no-edit", "main"); err == nil {
		t.Fatal("fixture: merge should conflict")
	}
	writeFile(t, f.dir, "a.txt", "a1\nRESOLVED\na3\n")
	commitAll(t, f.dir, "merge main")
	writeFile(t, f.dir, "b.txt", "b\n")
	commitAll(t, f.dir, "#304 M2: add b")

	s, conflicted, err := RebasedReviewedBase("main", f.h1)
	if err != nil {
		t.Fatalf("RebasedReviewedBase: %v", err)
	}
	if !reflect.DeepEqual(conflicted, []string{"a.txt"}) {
		t.Errorf("conflicted = %v, want [a.txt]", conflicted)
	}
	patch := testfix.Capture(t, f.dir, "diff", s, "HEAD", "--", "a.txt")
	if !strings.Contains(patch, "+RESOLVED") || !strings.Contains(patch, "-<<<<<<<") {
		t.Errorf("diff S HEAD -- a.txt should replace markers with the resolution:\n%s", patch)
	}
	names := diffNames(t, f.dir, s, "HEAD")
	if !reflect.DeepEqual(names, []string{"a.txt", "b.txt"}) {
		t.Errorf("diff S HEAD = %v, want [a.txt b.txt]", names)
	}
}

// (f) an unresolvable reviewed head is an error naming it as such.
func TestRebasedReviewedBase_UnresolvableReviewed(t *testing.T) {
	patchRepo(t)
	_, _, err := RebasedReviewedBase("main", strings.Repeat("ab", 20))
	if err == nil || !strings.Contains(err.Error(), "reviewed head") {
		t.Fatalf("err = %v, want one mentioning \"reviewed head\"", err)
	}
}

// (g) S is a pure function of its inputs.
func TestRebasedReviewedBase_Deterministic(t *testing.T) {
	f := patchRepo(t)
	testfix.Git(t, f.dir, "merge", "-q", "--no-edit", "main")
	s1, _ := mustRebased(t, f.h1)
	s2, _ := mustRebased(t, f.h1)
	if s1 != s2 {
		t.Errorf("S differs across calls: %s vs %s", s1, s2)
	}
}

// (i) criss-cross history: two merge bases between main and HEAD.
func TestRebasedReviewedBase_CrissCrossRefuses(t *testing.T) {
	dir := testfix.Repo(t, testfix.Chdir())
	writeFile(t, dir, "x.txt", "x\n")
	commitAll(t, dir, "base")
	testfix.Git(t, dir, "checkout", "-q", "-b", "feat")
	writeFile(t, dir, "f.txt", "f\n")
	f1 := commitAll(t, dir, "F1")
	testfix.Git(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, "m.txt", "m\n")
	m1 := commitAll(t, dir, "M1")
	testfix.Git(t, dir, "merge", "-q", "--no-edit", f1) // main: merge(M1, F1)
	testfix.Git(t, dir, "checkout", "-q", "feat")
	testfix.Git(t, dir, "merge", "-q", "--no-edit", m1) // feat: merge(F1, M1)

	_, _, err := RebasedReviewedBase("main", f1)
	if err == nil {
		t.Fatal("want a criss-cross error")
	}
	for _, want := range []string{"criss-cross", f1, m1} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err %q should mention %s", err, want)
		}
	}
}

// (j) the explicit --merge-base: the review saw a newer main than HEAD now sits
// on (B_now precedes B_r). S must be B_now + the branch patch only; letting git
// pick the base (merge-base(B_now, H_r) == B_now) would fold main's foreign.txt
// into S and show it as a deletion in the interdiff.
func TestRebasedReviewedBase_ExplicitBaseWhenHeadOnOlderMain(t *testing.T) {
	f := patchRepo(t)
	// Reviewed head carries the patch on top of the newer main.
	testfix.Git(t, f.dir, "checkout", "-q", "-b", "reviewed", f.mainTip)
	writeFile(t, f.dir, "a.txt", "a1\nBRANCH\na3\n")
	hr := commitAll(t, f.dir, "#304 M1: change a (on newer main)")
	// HEAD (feat at H1) still sits on the older base.
	testfix.Git(t, f.dir, "checkout", "-q", "feat")

	s, _ := mustRebased(t, hr)
	if got := revParse(t, f.dir, s+"^"); got != f.base {
		t.Fatalf("parent(S) = %s, want B_now %s", got, f.base)
	}
	if names := diffNames(t, f.dir, s, "HEAD"); len(names) != 0 {
		t.Errorf("diff S HEAD = %v, want empty (main's changes are not the branch's)", names)
	}
}

// (h) Close-Token lookup survives rebase and merge, matches exactly, and is lost
// to a squash that rewrites the message.
func TestCommitsWithCloseToken(t *testing.T) {
	f := patchRepo(t)
	writeFile(t, f.dir, "c.txt", "c\n")
	commitAll(t, f.dir, "#304: close evidence", "Close-Token: tokA")

	check := func(label string, rangeArgs []string, token string, want int) {
		t.Helper()
		got, err := CommitsWithCloseToken(rangeArgs, token)
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		if len(got) != want {
			t.Errorf("%s: token %q found %v, want %d commit(s)", label, token, got, want)
		}
		for _, sha := range got {
			if subj := strings.TrimSpace(testfix.Capture(t, f.dir, "log", "-1", "--format=%s", sha)); subj != "#304: close evidence" {
				t.Errorf("%s: matched %s (%q), want the trailer commit", label, sha, subj)
			}
		}
	}

	testfix.Git(t, f.dir, "branch", "premerge")
	testfix.Git(t, f.dir, "merge", "-q", "--no-edit", "main")
	check("after merge", []string{"main..HEAD"}, "tokA", 1)
	check("different token", []string{"main..HEAD"}, "tokB", 0)
	check("prefix", []string{"main..HEAD"}, "tok", 0)
	check("substring", []string{"main..HEAD"}, "okA", 0)

	testfix.Git(t, f.dir, "checkout", "-q", "-B", "rebased", "premerge")
	testfix.Git(t, f.dir, "rebase", "-q", "main")
	check("after rebase", []string{"main..HEAD"}, "tokA", 1)

	testfix.Git(t, f.dir, "checkout", "-q", "-b", "squashed", "main")
	testfix.Git(t, f.dir, "merge", "-q", "--squash", "premerge")
	testfix.Git(t, f.dir, "commit", "-q", "-m", "#304: squashed landing")
	check("after squash", []string{"main..HEAD"}, "tokA", 0)

	if _, err := CommitsWithCloseToken([]string{"HEAD"}, ""); err == nil {
		t.Error("empty token should be an error")
	}
}
