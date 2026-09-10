package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// fakePublisher records what prepare produced, and can re-run it to model the
// CAS retry — because the property under test is that the collision check sits
// INSIDE the loop, and a double that calls prepare once cannot show that.
type fakePublisher struct {
	sets  []gitx.TrunkWrite
	view  *gitx.TrunkView
	runs  int
	rerun int // re-invoke prepare this many extra times, as a rejection would
	err   error
	// beforePrepare runs ahead of each prepare call, so a test can move the trunk
	// between attempts — the only way to model a peer landing mid-retry.
	beforePrepare func()
}

func (f *fakePublisher) UpdateMany(_ string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error)) error {
	for i := 0; i <= f.rerun; i++ {
		if f.beforePrepare != nil {
			f.beforePrepare()
		}
		f.runs++
		set, err := prepare(f.view)
		if err != nil {
			return err
		}
		f.sets = append(f.sets, set)
	}
	return f.err
}

// A file changed locally becomes a Write; a file that is gone becomes a Delete.
//
// The deletion half is the one that matters: the arm this replaces fails loudly
// on a missing source (os.ReadFile, claim.go:459), so silently omitting the path
// would report success while the file stayed published.
func TestSyncViaTrunk_DeletedFileBecomesDelete(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	issues := filepath.Join(repo, "workshop", "issues")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	live := "workshop/issues/000300-live.md"
	if err := os.WriteFile(filepath.Join(repo, live), []byte("body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "-A")
	testfix.Git(t, repo, "commit", "-q", "-m", "seed")

	gone := "workshop/issues/000301-gone.md"
	if err := os.WriteFile(filepath.Join(repo, gone), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "-A")
	testfix.Git(t, repo, "commit", "-q", "-m", "add")
	if err := os.Remove(filepath.Join(repo, gone)); err != nil {
		t.Fatal(err)
	}
	// Touch the live one so both appear as changed.
	if err := os.WriteFile(filepath.Join(repo, live), []byte("body v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pub := &fakePublisher{view: viewFor(t, repo)}
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues"}
	if err := syncViaTrunk(&out, &errOut, f, execGitRunner{}, "msg", pub); err != nil {
		t.Fatalf("%v\n%s", err, errOut.String())
	}
	if len(pub.sets) != 1 {
		t.Fatalf("prepare produced %d sets, want 1", len(pub.sets))
	}
	set := pub.sets[0]
	if _, ok := set.Write[live]; !ok {
		t.Errorf("changed file missing from Write: %+v", set.Write)
	}
	if len(set.Delete) != 1 || set.Delete[0] != gone {
		t.Errorf("Delete = %v, want [%s] — a locally deleted issue must leave the trunk", set.Delete, gone)
	}
}

// A republication that finds a DIFFERENT slug at its id refuses, and the refusal
// names both paths. It must never renumber: the id is published, so it is
// already in the branch name, commit subjects, deps: and sidecars (#188).
func TestSyncViaTrunk_RepublishRefusesOnForeignSlug(t *testing.T) {
	if got, foreign := decideCollision(207, "workshop/issues/000207-mine.md",
		map[int][]string{207: {"workshop/issues/000207-theirs.md"}}, false); got != verdictRefuse {
		t.Fatalf("verdict = %v, want refuse", got)
	} else {
		err := collisionRefusal(207, "workshop/issues/000207-mine.md", foreign)
		for _, want := range []string{"000207", "000207-mine.md", "000207-theirs.md", "Not renumbering", "ariadne#188"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal missing %q:\n%s", want, err)
			}
		}
	}
}

// The collision check must run on EVERY attempt, not once. A publisher that
// re-invokes prepare (as a CAS rejection does) must see the check re-evaluated.
func TestSyncViaTrunk_CheckRunsOnEveryAttempt(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if err := os.MkdirAll(filepath.Join(repo, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "workshop/issues/000400-x.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pub := &fakePublisher{rerun: 1, view: viewFor(t, repo)} // one rejection, one retry
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues"}
	if err := syncViaTrunk(&out, &errOut, f, execGitRunner{}, "msg", pub); err != nil {
		t.Fatalf("%v\n%s", err, errOut.String())
	}
	if pub.runs != 2 {
		t.Errorf("prepare ran %d times, want 2 — the check must re-evaluate on the retry", pub.runs)
	}
}

// Nothing to copy is not nothing to publish: a body already committed here still
// needs routing, which is the pre-#206 no-op this arm must not reintroduce.
func TestSyncViaTrunk_PublishExistingWithNoDirtyFiles(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if err := os.MkdirAll(filepath.Join(repo, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := "workshop/issues/000500-committed.md"
	if err := os.WriteFile(filepath.Join(repo, p), []byte("body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "-A")
	testfix.Git(t, repo, "commit", "-q", "-m", "committed")

	pub := &fakePublisher{view: viewFor(t, repo)}
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues", Issue: 500, PublishExisting: true}
	if err := syncViaTrunk(&out, &errOut, f, execGitRunner{}, "msg", pub); err != nil {
		t.Fatalf("%v\n%s", err, errOut.String())
	}
	if len(pub.sets) != 1 || len(pub.sets[0].Write) != 1 {
		t.Errorf("a committed-but-unpublished body must still be routed, got %+v", pub.sets)
	}
}

var _ = errors.Is

// viewFor gives the fake a usable TrunkView over the repo's own HEAD, so prepare
// can run refIDSpace against a real ref without a bare origin.
func viewFor(t *testing.T, repo string) *gitx.TrunkView {
	t.Helper()
	tf, err := gitx.NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	return tf.ViewOf("HEAD")
}

// The collision check must be WIRED INTO the arm, not merely implemented beside
// it. Found by mutation: disabling the check in syncViaTrunk left the whole
// suite green, because only the pure decideCollision was covered. A guard whose
// integration is untested is a guard that can be deleted silently.
func TestSyncViaTrunk_RefusesWhenTrunkHasForeignSlugAtOurID(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	issues := filepath.Join(repo, "workshop", "issues")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	// The TRUNK carries someone else's file at id 000600.
	theirs := "workshop/issues/000600-theirs.md"
	if err := os.WriteFile(filepath.Join(repo, theirs), []byte("theirs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "-A")
	testfix.Git(t, repo, "commit", "-q", "-m", "trunk carries theirs")

	// We are publishing a DIFFERENT slug at the same id.
	mine := "workshop/issues/000600-mine.md"
	if err := os.WriteFile(filepath.Join(repo, mine), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pub := &fakePublisher{view: viewFor(t, repo)}
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues"}
	err := syncViaTrunk(&out, &errOut, f, execGitRunner{}, "msg", pub)
	if err == nil {
		t.Fatal("publishing a foreign slug at a published id must refuse")
	}
	for _, want := range []string{"000600", "000600-mine.md", "000600-theirs.md", "Not renumbering"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal missing %q:\n%v", want, err)
		}
	}
	// And nothing was staged for publication.
	for _, set := range pub.sets {
		if _, ok := set.Write[mine]; ok {
			t.Error("a refused publication must not stage the colliding file")
		}
	}
}

// THE test for the re-allocate arm: the collision must be handled when it
// appears INSIDE the retry window, not only when it is visible up front.
//
// The fake re-invokes prepare as a CAS rejection does, and the id space grows
// between attempts — modelling a peer landing at our id after we read and before
// we pushed. If the re-allocation were computed once outside the loop, attempt 2
// would re-push the colliding id and land the duplicate as a clean fast-forward,
// which is ariadne#188's hole.
func TestSyncViaTrunk_ReallocatesOnMidRetryCollision(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if err := os.MkdirAll(filepath.Join(repo, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	mine := "workshop/issues/000700-mine.md"
	if err := os.WriteFile(filepath.Join(repo, mine),
		[]byte("---\nid: 000700\nstatus: open\n---\n\n# Mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A peer lands 000700-theirs.md on the trunk between attempt 1 and attempt 2.
	attempt := 0
	pub := &fakePublisher{rerun: 1, view: viewFor(t, repo), beforePrepare: func() {
		attempt++
		if attempt == 2 {
			p := filepath.Join(repo, "workshop/issues/000700-theirs.md")
			if err := os.WriteFile(p, []byte("---\nid: 000700\n---\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			// Stage ONLY the peer's file: an `add -A` would sweep up ours too,
			// putting our path on the trunk and making this correctly a refusal
			// rather than the re-allocation the test is about.
			testfix.Git(t, repo, "add", "workshop/issues/000700-theirs.md")
			testfix.Git(t, repo, "commit", "-q", "-m", "peer lands at our id")
		}
	}}

	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues", FirstPublication: true}
	rc, err := syncViaTrunkWithRealloc(&out, &errOut, f, execGitRunner{}, "msg", pub)
	if err != nil {
		t.Fatalf("%v\n%s", err, errOut.String())
	}
	if rc == nil {
		t.Fatal("a mid-retry collision must produce a re-allocation; nil means the decision ran outside the loop")
	}
	if rc.OldID != 700 || rc.NewID != 701 {
		t.Errorf("re-allocated %d -> %d, want 700 -> 701", rc.OldID, rc.NewID)
	}
	// The final set must carry the NEW path, never the colliding one.
	last := pub.sets[len(pub.sets)-1]
	if _, ok := last.Write[mine]; ok {
		t.Error("the colliding path was still published — the retry re-pushed the same id")
	}
	if body, ok := last.Write["workshop/issues/000701-mine.md"]; !ok {
		t.Errorf("re-derived path missing from the published set: %+v", last.Write)
	} else if !strings.Contains(string(body), "id: 000701") {
		t.Errorf("frontmatter not moved with the filename:\n%s", body)
	}
	if !strings.Contains(errOut.String(), "now 000701") {
		t.Errorf("the id change must be announced loudly:\n%s", errOut.String())
	}
}

// A republication never re-allocates, whatever the trunk holds. This is the
// regression the first draft of #207's Spec would have shipped.
func TestSyncViaTrunk_RepublicationNeverReallocates(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if err := os.MkdirAll(filepath.Join(repo, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The trunk already carries OUR file at this id — the ordinary sync case.
	mine := "workshop/issues/000800-mine.md"
	if err := os.WriteFile(filepath.Join(repo, mine),
		[]byte("---\nid: 000800\n---\n\n# v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "-A")
	testfix.Git(t, repo, "commit", "-q", "-m", "published")
	if err := os.WriteFile(filepath.Join(repo, mine),
		[]byte("---\nid: 000800\n---\n\n# v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pub := &fakePublisher{view: viewFor(t, repo)}
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues"} // FirstPublication false
	rc, err := syncViaTrunkWithRealloc(&out, &errOut, f, execGitRunner{}, "msg", pub)
	if err != nil {
		t.Fatalf("republishing our own body must succeed: %v\n%s", err, errOut.String())
	}
	if rc != nil {
		t.Fatalf("a republication renumbered the issue %d -> %d — the defect this design exists to prevent",
			rc.OldID, rc.NewID)
	}
	if _, ok := pub.sets[0].Write[mine]; !ok {
		t.Errorf("the body was not published under its own id: %+v", pub.sets[0].Write)
	}
}

// Only the changed ISSUE files are published — an unrelated dirty file in the
// working tree is not swept in.
//
// Inherited from TestSyncViaMainWorktree_CommitsOnlyTheCopiedIssueFiles, which
// is deleted with the arm it covered. Its other half — "an untracked peer file
// in the MAIN worktree is left alone" — does not migrate, because it is now
// structurally impossible rather than merely asserted: this arm never opens a
// worktree at all.
func TestSyncViaTrunk_PublishesOnlyChangedIssueFiles(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if err := os.MkdirAll(filepath.Join(repo, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	issue := "workshop/issues/000900-x.md"
	if err := os.WriteFile(filepath.Join(repo, issue), []byte("---\nid: 000900\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Unrelated dirty work sitting beside it.
	if err := os.WriteFile(filepath.Join(repo, "peer-work.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pub := &fakePublisher{view: viewFor(t, repo)}
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues"}
	if err := syncViaTrunk(&out, &errOut, f, execGitRunner{}, "msg", pub); err != nil {
		t.Fatalf("%v\n%s", err, errOut.String())
	}
	set := pub.sets[0]
	if _, ok := set.Write[issue]; !ok {
		t.Errorf("the changed issue file was not published: %+v", set.Write)
	}
	for p := range set.Write {
		if !strings.HasPrefix(p, "workshop/issues/") {
			t.Errorf("published a non-issue path: %s", p)
		}
	}
}

// A FAILED publish must not clean up. finish() removes the ORIGINAL local file,
// so running it on the error path deletes the operator's issue while nothing
// reached the trunk — the Spec's ordering inverted, and data loss.
func TestSyncViaTrunk_FailedPublishKeepsTheOriginalFile(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if err := os.MkdirAll(filepath.Join(repo, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	mine := filepath.Join(repo, "workshop/issues/001000-mine.md")
	if err := os.WriteFile(mine, []byte("---\nid: 001000\n---\n\n# body\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pub := &fakePublisher{view: viewFor(t, repo), err: errors.New("push rejected")}
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues", FirstPublication: true}
	if err := syncViaTrunk(&out, &errOut, f, execGitRunner{}, "msg", pub); err == nil {
		t.Fatal("expected the publish failure to surface")
	}
	if _, err := os.Stat(mine); err != nil {
		t.Errorf("the original issue file was deleted after a FAILED publish: %v", err)
	}
}

// Re-allocation must avoid ids held by UNPUBLISHED LOCAL files, not just the
// trunk. The trunk alone is not the id space — ariadne#213 settled that for
// allocation, and stepping onto a local reservation mints exactly the duplicate
// this arm prevents.
func TestSyncViaTrunk_ReallocationSkipsLocalUnpublishedIDs(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	issues := filepath.Join(repo, "workshop", "issues")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	// The TRUNK holds a foreign slug at 001100, forcing a re-allocation.
	if err := os.WriteFile(filepath.Join(issues, "001100-theirs.md"), []byte("---\nid: 001100\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "-A")
	testfix.Git(t, repo, "commit", "-q", "-m", "trunk")

	// 001101 exists LOCALLY and is unpublished — the next free id must skip it.
	if err := os.WriteFile(filepath.Join(issues, "001101-local-unpublished.md"), []byte("---\nid: 001101\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mine := "workshop/issues/001100-mine.md"
	if err := os.WriteFile(filepath.Join(repo, mine), []byte("---\nid: 001100\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pub := &fakePublisher{view: viewFor(t, repo)}
	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues", FirstPublication: true}
	rc, err := syncViaTrunkWithRealloc(&out, &errOut, f, execGitRunner{}, "msg", pub)
	if err != nil {
		t.Fatalf("%v\n%s", err, errOut.String())
	}
	if rc == nil {
		t.Fatal("expected a re-allocation")
	}
	if rc.NewID == 1101 {
		t.Fatal("re-allocated onto 001101, which a local unpublished file already holds — the duplicate this prevents")
	}
	if rc.NewID != 1102 {
		t.Errorf("re-allocated to %d, want 1102 (next free past both the trunk and local)", rc.NewID)
	}
}

// rc must be reset per attempt. If attempt 1 re-allocates and attempt 2 does
// NOT — the collision cleared on the new base — a carried-over rc reports an id
// change the final push never made, and finish() then deletes the original file
// that WAS just published under its own name. Same data-loss family as
// FailedPublishKeepsTheOriginalFile, reached by a different route.
func TestSyncViaTrunk_StaleReallocationIsNotCarriedAcrossAttempts(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	issues := filepath.Join(repo, "workshop", "issues")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	theirs := filepath.Join(issues, "001200-theirs.md")
	if err := os.WriteFile(theirs, []byte("---\nid: 001200\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "-A")
	testfix.Git(t, repo, "commit", "-q", "-m", "peer holds our id")

	mine := "workshop/issues/001200-mine.md"
	if err := os.WriteFile(filepath.Join(repo, mine), []byte("---\nid: 001200\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	attempt := 0
	pub := &fakePublisher{rerun: 1, view: viewFor(t, repo), beforePrepare: func() {
		attempt++
		if attempt == 2 {
			// The peer withdrew: on THIS base our id is free again.
			testfix.Git(t, repo, "rm", "-q", "workshop/issues/001200-theirs.md")
			testfix.Git(t, repo, "commit", "-q", "-m", "peer withdrew")
		}
	}}

	var out, errOut bytes.Buffer
	f := &claimFlags{IssuesDir: "workshop/issues", FirstPublication: true}
	rc, err := syncViaTrunkWithRealloc(&out, &errOut, f, execGitRunner{}, "msg", pub)
	if err != nil {
		t.Fatalf("%v\n%s", err, errOut.String())
	}
	if rc != nil {
		t.Fatalf("attempt 2 found no collision, so no re-allocation happened — "+
			"a carried-over rc (%06d -> %06d) would report an id change the push never made, "+
			"and finish() would delete the file just published under its own name",
			rc.OldID, rc.NewID)
	}
	last := pub.sets[len(pub.sets)-1]
	if _, ok := last.Write[mine]; !ok {
		t.Errorf("the final attempt must publish the ORIGINAL path: %+v", last.Write)
	}
}
