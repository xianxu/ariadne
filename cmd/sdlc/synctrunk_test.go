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
}

func (f *fakePublisher) UpdateMany(_ string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error)) error {
	for i := 0; i <= f.rerun; i++ {
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
