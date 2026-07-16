package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

func boardDoc(t *testing.T, tasks string) *projectdoc.Doc {
	t.Helper()
	d, err := projectdoc.ParseDoc("---\ntype: project\nname: demo\nstatus: executing\ndeadline: 2026-09-01\nplanned_finish: 2026-08-20\n---\n## Breakdown\n" + tasks + "\n## Log\n### 2026-07-10 — retro\n")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestLookupIssueMetaCrossRepoAndArchive(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "ariadne")
	peer := filepath.Join(parent, "nous")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(peer, "workshop", "history", "issues")
	if err := os.MkdirAll(archive, 0o755); err != nil {
		t.Fatal(err)
	}
	issue := "---\nid: 000007\nstatus: done\nestimate_hours: 3.5\nactual_hours: 4.25\ndeps: [ariadne#1, nous#2]\n---\n# x\n"
	if err := os.WriteFile(filepath.Join(archive, "000007-x.md"), []byte(issue), 0o644); err != nil {
		t.Fatal(err)
	}
	meta, err := lookupIssueMeta("nous#7", root)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Status != "done" || meta.EstimateHours != 3.5 || meta.ActualHours != 4.25 || strings.Join(meta.Deps, ",") != "ariadne#1,nous#2" {
		t.Fatalf("meta = %+v", meta)
	}
}

func TestProjectStatusCommandRegistered(t *testing.T) {
	project, _, err := buildRoot().Find([]string{"project"})
	if err != nil {
		t.Fatal(err)
	}
	found, _, err := project.Find([]string{"status"})
	if err != nil || found == project {
		t.Fatalf("project status not registered: %v", err)
	}
}

func TestComputeBoardDerivesProgressFrontierAndThreads(t *testing.T) {
	d := boardDoc(t, "- [x] shipped [ariadne#1]\n- [ ] alpha [ariadne#2]\n- [ ] beta [ariadne#3]\n- [ ] plain task\n- [ ] missing [nous#9]")
	metas := map[string]issueMeta{
		"ariadne#1": {Status: "working", EstimateHours: 5},
		"ariadne#2": {Status: "open", EstimateHours: 3, Deps: []string{"ariadne#3"}},
		"ariadne#3": {Status: "done", ActualHours: 2},
	}
	b, err := computeBoard(d, func(ref string) (issueMeta, error) {
		m, ok := metas[ref]
		if !ok {
			return issueMeta{}, errors.New("peer unavailable")
		}
		return m, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if b.Done != 1 || b.Total != 5 || b.RemainingHours != 3 {
		t.Fatalf("board totals = %+v", b)
	}
	if len(b.Frontier) != 1 || b.Frontier[0] != "ariadne#2" {
		t.Fatalf("frontier = %v", b.Frontier)
	}
	if b.LastRetro != "2026-07-10" {
		t.Fatalf("last retro = %q", b.LastRetro)
	}
	if len(b.Threads) != 2 || strings.Join(b.Threads[0], ",") != "ariadne#2,ariadne#3" || strings.Join(b.Threads[1], ",") != "nous#9" {
		t.Fatalf("threads = %v", b.Threads)
	}
	if b.Rows[0].Warning == "" || b.Rows[4].Warning == "" {
		t.Fatalf("expected mismatch/unresolved warnings: %+v", b.Rows)
	}
}

func TestComputeBoardIndependentRefsFormIndependentThreads(t *testing.T) {
	d := boardDoc(t, "- [ ] one [ariadne#1]\n- [ ] two [ariadne#2]\n- [ ] three [ariadne#3]")
	b, err := computeBoard(d, func(string) (issueMeta, error) { return issueMeta{Status: "open", EstimateHours: 1}, nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Threads) != 3 {
		t.Fatalf("threads = %v, want 3 components", b.Threads)
	}
}

func TestComputeBoardFrontierIncludesUnblockedActiveIssue(t *testing.T) {
	d := boardDoc(t, "- [ ] underway [#1]")
	b, err := computeBoard(d, func(string) (issueMeta, error) {
		return issueMeta{Status: "working", EstimateHours: 2}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Frontier) != 1 || b.Frontier[0] != "#1" {
		t.Fatalf("frontier = %v, want unblocked active issue #1", b.Frontier)
	}
}

func TestRenderBoardIncludesOperationalLines(t *testing.T) {
	b := board{Name: "demo", Status: "executing", Deadline: "2026-09-01", PlannedFinish: "2026-08-20", Done: 1, Total: 3, RemainingHours: 22, Frontier: []string{"ariadne#2"}, Blocked: []string{"nous#9"}, Threads: [][]string{{"ariadne#2", "ariadne#3"}, {"nous#9"}}, LastRetro: "2026-07-10"}
	out := renderBoard(b, "2026-07-20")
	for _, want := range []string{"demo — executing", "1/3 done", "Σ remaining ≈ 22h", "frontier: ariadne#2", "blocked: nous#9", "threads: 2 — [ariadne#2, ariadne#3] / [nous#9]", "last retro: 2026-07-10"} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q:\n%s", want, out)
		}
	}
}
