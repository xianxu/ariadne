package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// sideFixture lays out a done issue and its plan, each with a side folder;
// the issue's folder holds a side doc of a side doc (#424).
func sideFixture(t *testing.T, root string) {
	t.Helper()
	w := func(p, s string) { mkArtifact(t, filepath.Join(root, p), s) }
	w("workshop/issues/000143-x.md", "---\nid: 143\nstatus: done\n---\n\n# x\n\n→ [discussion](000143-x.side/01-q.md)\n")
	w("workshop/issues/000143-x.side/01-q.md", "---\ntype: side-doc\nparent: ../000143-x.md\n---\n\n_From [x](../000143-x.md), section \"x\"._\n→ [discussion](02-n.md)\n")
	w("workshop/issues/000143-x.side/02-n.md", "---\ntype: side-doc\nparent: 01-q.md\n---\n\n_From [q](01-q.md), section \"q\"._\n")
	w("workshop/plans/000143-x-plan.md", "the plan → [discussion](000143-x-plan.side/01-p.md)\n")
	w("workshop/plans/000143-x-plan.side/01-p.md", "---\ntype: side-doc\nparent: ../000143-x-plan.md\n---\n\n_From [plan](../000143-x-plan.md), section \"p\"._\n")
	w("workshop/issues/000144-y.md", "---\nid: 144\nstatus: working\n---\n\n# y\n")
	w("workshop/issues/000144-y.side/01-open.md", "---\ntype: side-doc\nparent: ../000144-y.md\n---\n")
}

var sideArchived = []string{
	"workshop/history/issues/000143-x.side/01-q.md",
	"workshop/history/issues/000143-x.side/02-n.md",
	"workshop/history/plans/000143-x-plan.side/01-p.md",
}

// Both legacy archive paths move the issue's and its plan's side folders with
// them, leave an open issue's folder alone, and pass the pre-commit verify.
func TestArchiveSideFolders(t *testing.T) {
	for _, merge := range []bool{false, true} {
		t.Run(map[bool]string{false: "push", true: "merge"}[merge], func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
			sideFixture(t, root)
			testfix.Git(t, root, "add", ".")
			testfix.Git(t, root, "commit", "-qm", "fixture")
			var out bytes.Buffer
			var moves []preparedArchiveMove
			var err error
			if merge {
				moves, err = archiveDoneIssuesInDir(context.Background(), &out, "", root, "workshop/issues", "workshop/history", "workshop/plans")
			} else {
				moves, err = archiveDoneIssues(context.Background(), &out, "", "workshop/issues", "workshop/history", "workshop/plans")
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range sideArchived {
				if _, err := os.Stat(filepath.Join(root, p)); err != nil {
					t.Errorf("%s not archived: %v", p, err)
				}
			}
			for _, p := range []string{"workshop/issues/000143-x.side", "workshop/plans/000143-x-plan.side"} {
				if _, err := os.Stat(filepath.Join(root, p)); !os.IsNotExist(err) {
					t.Errorf("%s remains: %v", p, err)
				}
			}
			if _, err := os.Stat(filepath.Join(root, "workshop/issues/000144-y.side/01-open.md")); err != nil {
				t.Errorf("an open issue's side folder moved: %v", err)
			}
			if len(moves) != 5 { // issue, plan, three side docs
				t.Errorf("moves = %d: %v", len(moves), moves)
			}
			if err := verifyArchivedSideMoves(root, moves); err != nil {
				t.Error(err)
			}
			testfix.Git(t, root, archiveAddArgs(moves)...) // the moves stage
		})
	}
}

// An occupied destination folder or a symlink refuses before anything moves.
func TestArchiveSideRefusesBeforeEffects(t *testing.T) {
	for _, kind := range []string{"occupied", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
			sideFixture(t, root)
			if kind == "occupied" {
				mkArtifact(t, filepath.Join(root, "workshop/history/issues/000143-x.side/09-old.md"), "an older doc")
			} else if err := os.Symlink("/etc/hosts", filepath.Join(root, "workshop/issues/000143-x.side/03-link.md")); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if _, err := archiveDoneIssues(context.Background(), &out, "", "workshop/issues", "workshop/history", "workshop/plans"); err == nil {
				t.Fatal("archived anyway")
			}
			for _, p := range []string{"workshop/issues/000143-x.md", "workshop/issues/000143-x.side/01-q.md", "workshop/plans/000143-x-plan.md"} {
				if _, err := os.Stat(filepath.Join(root, p)); err != nil {
					t.Errorf("the refusal moved %s: %v", p, err)
				}
			}
		})
	}
}

// An archive interrupted after some side docs moved is finished by push's
// recovery: the moved issue is the authority, and the shared finisher moves
// the rest. A conflicting copy refuses; an unowned side doc is unrelated.
func TestPreparedArchiveMovesRecoversSideFolders(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	sideFixture(t, root)
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "fixture")
	// The interrupted state: the issue and one side doc moved, one not.
	for _, mv := range [][2]string{
		{"workshop/issues/000143-x.md", "workshop/history/issues/000143-x.md"},
		{"workshop/issues/000143-x.side/01-q.md", "workshop/history/issues/000143-x.side/01-q.md"},
	} {
		mkArtifact(t, mv[1], "")
		if err := os.Rename(mv[0], mv[1]); err != nil {
			t.Fatal(err)
		}
	}
	status := testfix.Capture(t, root, "status", "--porcelain=v1", "--untracked-files=all")
	moves, other, err := preparedArchiveMoves(context.Background(), statusOf(t, status), "workshop/issues", "workshop/history", "workshop/plans")
	if err != nil || len(other) != 0 {
		t.Fatalf("moves=%v other=%v err=%v", moves, other, err)
	}
	for _, m := range moves {
		if strings.Contains(m.IssuePath, ".side") {
			if err := finishArchiveMemberMoves(".", []preparedArchiveMove{m}); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, p := range []string{"workshop/history/issues/000143-x.side/01-q.md", "workshop/history/issues/000143-x.side/02-n.md"} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s not recovered: %v", p, err)
		}
	}
	if err := verifyArchivedSideMoves(".", moves); err != nil {
		t.Error(err)
	}
	testfix.Git(t, root, archiveAddArgs(moves)...) // the recovered moves stage

	// A conflicting active copy of a moved side doc refuses.
	mkArtifact(t, "workshop/issues/000143-x.side/01-q.md", "edited meanwhile")
	status = testfix.Capture(t, root, "status", "--porcelain=v1", "--untracked-files=all")
	if _, _, err := preparedArchiveMoves(context.Background(), statusOf(t, status), "workshop/issues", "workshop/history", "workshop/plans"); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Errorf("a conflicting copy: %v", err)
	}
	// A side doc no moved document owns is an unrelated change.
	_, other, err = preparedArchiveMoves(context.Background(), statusOf(t, "?? workshop/issues/000150-z.side/01-a.md\n"), "workshop/issues", "workshop/history", "workshop/plans")
	if err != nil || len(other) == 0 {
		t.Errorf("an unowned side doc: other=%v err=%v", other, err)
	}
}

// The landing archive (#424) carries the issue's and its plan's side folders
// in its one commit, its proof accepts the result, and an occupied
// destination refuses with main unchanged.
func TestLandingArchiveSideFolders(t *testing.T) {
	for _, occupied := range []bool{false, true} {
		root, origin, pr := laFixture(t)
		laWrite(t, root, "workshop/issues/000246-landing.side/01-q.md", "---\ntype: side-doc\nparent: ../000246-landing.md\n---\nq\n")
		laWrite(t, root, "workshop/issues/000246-landing.side/02-n.md", "---\ntype: side-doc\nparent: 01-q.md\n---\nn\n")
		laWrite(t, root, "workshop/plans/000246-landing-plan.side/01-p.md", "---\ntype: side-doc\nparent: ../000246-landing-plan.md\n---\np\n")
		if occupied {
			laWrite(t, root, "workshop/history/issues/000246-landing.side/09-old.md", "older\n")
		}
		pr.HeadOID = laCommit(t, root, "side docs")
		pr.MergeOID = pr.HeadOID
		testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
		before := laGit(t, origin, "rev-parse", "main")
		err := laArchive(root, pr)
		if occupied {
			if err == nil || !strings.Contains(err.Error(), "occupied") || laGit(t, origin, "rev-parse", "main") != before {
				t.Fatalf("an occupied side folder: %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		files := testfix.Capture(t, origin, "ls-tree", "-r", "--name-only", "main")
		for _, want := range []string{"workshop/history/issues/000246-landing.side/01-q.md", "workshop/history/issues/000246-landing.side/02-n.md", "workshop/history/plans/000246-landing-plan.side/01-p.md"} {
			if !strings.Contains(files, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(files, "workshop/issues/000246-landing.side/") || strings.Contains(files, "workshop/plans/000246-landing-plan.side/") {
			t.Errorf("active side folders remain:\n%s", files)
		}
		tip := laGit(t, origin, "rev-parse", "main")
		if complete, err := laProof(root, tip, pr); err != nil || !complete {
			t.Fatalf("proof %v %v", complete, err)
		}
	}
}

// The trunk archive (abandon → archiveIssueOnMain) carries the side folder in
// its write, reopen brings it back beside the live details, and a retry that
// finds side-doc residue beside the moved details refuses (#424, paths d, e
// and the reverse move).
func TestAbandonAndReopenCarrySideFolder(t *testing.T) {
	r, _, _ := startedHere(t)
	side := "workshop/issues/000009-s09.side/01-q.md"
	archived := "workshop/history/issues/000009-s09.side/01-q.md"
	writeRepoFile(t, r.root, side, "---\ntype: side-doc\nparent: ../000009-s09.md\n---\nq\n")
	r.git("add", ".")
	r.git("commit", "-qm", "a side doc")
	if err := abandon9("punt", "later"); err != nil {
		t.Fatal(err)
	}
	r.git("fetch", "-q", "origin")
	if got := r.git("show", "origin/main:"+archived); !strings.Contains(got, "parent: ../000009-s09.md") {
		t.Fatalf("archived side doc=%q", got)
	}
	if gitSucceeds(r.root, "cat-file", "-e", "origin/main:"+side) {
		t.Fatal("active side doc retained on main")
	}
	if out, err := reopen9(); err != nil {
		t.Fatalf("reopen: %v %s", err, out)
	}
	if got := r.git("show", "HEAD:"+side); !strings.Contains(got, "parent: ../000009-s09.md") {
		t.Fatalf("restored side doc=%q", got)
	}
	if gitSucceeds(r.root, "cat-file", "-e", "HEAD:"+archived) {
		t.Fatal("archived side doc retained after reopen")
	}
}

func TestAbandonRetryRefusesSideResidue(t *testing.T) {
	r, _, _ := startedHere(t)
	if err := abandon9("punt", "later"); err != nil {
		t.Fatal(err)
	}
	r.git("pull", "-q", "--ff-only")
	writeRepoFile(t, r.root, "workshop/issues/000009-s09.side/01-left.md", "left behind\n")
	r.git("add", ".")
	r.git("commit", "-qm", "side residue")
	r.git("push", "-q", "origin", "main")
	before := r.originMain()
	if err := abandon9("punt", "later"); err == nil || !strings.Contains(err.Error(), "side-doc residue") {
		t.Fatalf("side residue accepted: %v", err)
	}
	if r.originMain() != before {
		t.Fatal("refusal changed main")
	}
}
