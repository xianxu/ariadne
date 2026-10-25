package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/sidedoc"
)

func TestLandingArchiveOrphanPlanSideFolders(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprint(present), func(t *testing.T) {
			root, origin, pr := laFixture(t)
			owner := "workshop/plans/000246-landing-plan.md"
			if !present {
				if err := os.Remove(filepath.Join(root, owner)); err != nil {
					t.Fatal(err)
				}
			}
			side := "workshop/plans/000246-landing-plan.side/evidence/000999-capture.json"
			archived := strings.Replace(side, "workshop/plans/", "workshop/history/plans/", 1)
			content := "{\"future\":{\"nested\":[1,true]},\"capture\":\"kept\\nbytes\"}\n"
			other := "workshop/plans/000247-other-plan.side/evidence.log"
			laWrite(t, root, side, content)
			laWrite(t, root, other, "unrelated\n")
			pr.HeadOID = laCommit(t, root, "#246: orphan side evidence")
			pr.MergeOID = pr.HeadOID
			pr.LandedOID = pr.HeadOID
			testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
			beforeBlob := laGit(t, root, "rev-parse", pr.HeadOID+":"+side)
			if err := laArchive(root, pr); err != nil {
				t.Fatal(err)
			}
			tip := laGit(t, origin, "rev-parse", "main")
			if got := laGit(t, origin, "rev-parse", tip+":"+archived); got != beforeBlob {
				t.Fatalf("side bytes changed: %s vs %s", got, beforeBlob)
			}
			if gitSucceeds(origin, "cat-file", "-e", tip+":"+side) {
				t.Fatal("active side evidence remains")
			}
			if !gitSucceeds(origin, "cat-file", "-e", tip+":"+other) {
				t.Fatal("another issue's side tree moved")
			}
			if present != gitSucceeds(origin, "cat-file", "-e", tip+":workshop/history/plans/000246-landing-plan.md") {
				t.Fatal("owner document invented or lost")
			}
			if complete, err := laProof(root, tip, pr); err != nil || !complete {
				t.Fatalf("archive proof: complete%v error%v", complete, err)
			}
			if err := laArchive(root, pr); err != nil {
				t.Fatalf("confirmed retry failed: %v", err)
			}
			if got := laGit(t, origin, "rev-parse", "main"); got != tip {
				t.Fatal("retry published another archive")
			}
		})
	}
}

func TestFilesystemArchiveOrphanPlanSideFolders(t *testing.T) {
	for _, merge := range []bool{false, true} {
		t.Run(fmt.Sprint(merge), func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
			sideFixture(t, root)
			if err := os.Remove(filepath.Join(root, "workshop/plans/000143-x-plan.md")); err != nil {
				t.Fatal(err)
			}
			testfix.Git(t, root, "add", ".")
			testfix.Git(t, root, "commit", "-qm", "orphan fixture")
			original, err := os.ReadFile("workshop/plans/000143-x-plan.side/01-p.md")
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			var moves []preparedArchiveMove
			if merge {
				moves, err = archiveDoneIssuesInDir(context.Background(), &out, "", root, "workshop/issues", "workshop/history", "workshop/plans")
			} else {
				moves, err = archiveDoneIssues(context.Background(), &out, "", "workshop/issues", "workshop/history", "workshop/plans")
			}
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile("workshop/history/plans/000143-x-plan.side/01-p.md")
			if err != nil || !bytes.Equal(original, got) {
				t.Fatalf("orphan tree not preserved: %v", err)
			}
			if _, err := os.Stat("workshop/plans/000143-x-plan.side"); !os.IsNotExist(err) {
				t.Fatalf("active orphan remains: %v", err)
			}
			if _, err := os.Stat("workshop/history/plans/000143-x-plan.md"); !os.IsNotExist(err) {
				t.Fatal("invented owner document")
			}
			if err := verifyArchivedSideMoves(root, moves); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPreparedArchiveMovesRecoversOrphanPlanSideFolder(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	sideFixture(t, root)
	if err := os.Remove("workshop/plans/000143-x-plan.md"); err != nil {
		t.Fatal(err)
	}
	second := "workshop/plans/000143-x-plan.side/02.json"
	mkArtifact(t, second, "{\"kept\":true}\n")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "orphan fixture")
	for _, mv := range [][2]string{{"workshop/issues/000143-x.md", "workshop/history/issues/000143-x.md"}, {"workshop/plans/000143-x-plan.side/01-p.md", "workshop/history/plans/000143-x-plan.side/01-p.md"}} {
		mkArtifact(t, mv[1], "")
		if err := os.Rename(mv[0], mv[1]); err != nil {
			t.Fatal(err)
		}
	}
	status := testfix.Capture(t, root, "status", "--porcelain=v1", "--untracked-files=all")
	moves, other, err := preparedArchiveMoves(context.Background(), statusOf(t, status), "workshop/issues", "workshop/history", "workshop/plans")
	if err != nil || len(other) != 0 {
		t.Fatalf("orphan recovery: other%v error%v", other, err)
	}
	for _, m := range moves {
		if strings.Contains(m.IssuePath, ".side/") {
			if err := finishArchiveMemberMoves(".", []preparedArchiveMove{m}); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got, err := os.ReadFile("workshop/history/plans/000143-x-plan.side/02.json"); err != nil || string(got) != "{\"kept\":true}\n" {
		t.Fatalf("orphan member not recovered: %q %v", got, err)
	}
	if err := verifyArchivedSideMoves(".", moves); err != nil {
		t.Fatal(err)
	}
}

func TestAbandonOrphanPlanSideFolder(t *testing.T) {
	r, _, _ := startedHere(t)
	side := "workshop/plans/000009-s09-orphan.side/01.md"
	archived := strings.Replace(side, "workshop/plans/", "workshop/history/plans/", 1)
	content := "---\ntype: side-doc\nparent: ../../issues/000009-s09.md\n---\nkept\n"
	writeRepoFile(t, r.root, side, content)
	r.git("add", ".")
	r.git("commit", "-qm", "orphan plan side")
	if err := abandon9("punt", "later"); err != nil {
		t.Fatal(err)
	}
	r.git("fetch", "-q", "origin")
	if got := r.git("show", "origin/main:"+archived); got != strings.TrimSpace(content) {
		t.Fatalf("orphan archive bytes: %q", got)
	}
	if gitSucceeds(r.root, "cat-file", "-e", "origin/main:"+side) {
		t.Fatal("active orphan remains after abandon")
	}
	if out, err := reopen9(); err != nil {
		t.Fatalf("reopen: %v %s", err, out)
	}
	if got := r.git("show", "HEAD:"+side); got != strings.TrimSpace(content) {
		t.Fatalf("reopen lost orphan: %q", got)
	}
	writeRepoFile(t, r.root, side, strings.Replace(content, "kept", "second", 1))
	r.git("add", ".")
	r.git("commit", "-qm", "second orphan episode")
	if err := abandon9("wontfix", "finished"); err != nil {
		t.Fatal(err)
	}
	r.git("fetch", "-q", "origin")
	if got := r.git("show", "origin/main:"+archived); !strings.Contains(got, "second") {
		t.Fatalf("successor orphan was not replaced: %q", got)
	}
}

func TestSideArchiveInventoryLinearVisits(t *testing.T) {
	for _, n := range []int{8, 16} {
		var paths []string
		for i := 0; i < n; i++ {
			paths = append(paths, fmt.Sprintf("workshop/plans/000143-plan-%d.side/nested/evidence.json", i))
		}
		paths = append(paths, "workshop/plans/000144-other.side/unrelated.md")
		visits := 0
		index := indexSideMembers(paths, func(p string) (string, bool) {
			visits++
			return sidedoc.Owner(p)
		})
		if visits != len(paths) || len(index) != len(paths) {
			t.Fatalf("N%d: visits%d indexed%d paths%d", n, visits, len(index), len(paths))
		}
		docs := sideArchiveDocsForIssue("workshop/issues/000143-owner.md", "workshop/history/issues/000143-owner.md", paths, "workshop/plans", "workshop/history")
		if len(docs) != n+1 {
			t.Fatalf("N%d: docs%d, want issue plus every orphan owner", n, len(docs))
		}
		current := landingArchiveSnapshot{}
		for _, p := range paths {
			current[p] = landingArchiveBlob{mode: "100644", content: []byte(p)}
		}
		moves := 0
		err := planLandingSideMoves(current, docs, func(src, dst string, body []byte) error {
			moves++
			if strings.Contains(src, "000144-") || dst != strings.Replace(src, "workshop/plans/", "workshop/history/plans/", 1) || string(body) != src {
				t.Fatalf("incorrect member move %s -> %s", src, dst)
			}
			return nil
		})
		if err != nil || moves != n {
			t.Fatalf("N%d: moves%d error%v", n, moves, err)
		}
	}
}
