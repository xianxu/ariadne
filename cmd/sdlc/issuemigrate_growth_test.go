package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Counts are taken only around the complete adapter, excluding fixture setup,
// main pinning, and the legacy-close history algorithm (no codecomplete cards).
func TestMigrationInventoryProcessGrowth(t *testing.T) {
	type counts struct{ total, batches int }
	count := func(axis string, n int) counts {
		t.Helper()
		r := emptyInitRepo(t)
		documents := 1
		if axis == "documents" {
			documents = n
		}
		for i := 0; i < documents; i++ {
			id := fmt.Sprintf("%06d", i+1)
			writeRepoFile(t, r.root, "workshop/issues/"+id+"-active.md", migrationTestBody(id, "open"))
			// All three roots have a fixed shape at both document population sizes.
			writeRepoFile(t, r.root, fmt.Sprintf("workshop/history/%06d-flat.md", i+100), migrationTestBody(fmt.Sprintf("%06d", i+100), "done"))
			writeRepoFile(t, r.root, fmt.Sprintf("workshop/history/issues/%06d-archive.md", i+200), migrationTestBody(fmt.Sprintf("%06d", i+200), "done"))
		}
		if axis == "assets" {
			for i := 0; i < n; i++ {
				writeRepoFile(t, r.root, fmt.Sprintf("workshop/issues/assets/%06d-payload.md", i+900), strings.Repeat("asset bytes\n", 32*1024))
			}
		}
		r.git("add", ".")
		r.git("commit", "-qm", "published population")
		r.git("push", "-q")
		switch axis {
		case "documents":
			r.git("switch", "-qc", "all-active-details-changed")
			for i := 0; i < documents; i++ {
				id := fmt.Sprintf("%06d", i+1)
				writeRepoFile(t, r.root, "workshop/issues/"+id+"-active.md", migrationTestBody(id, "open")+"\nRelated branch prose.\n")
			}
			r.git("add", ".")
			r.git("commit", "-qm", "changed details")
			r.git("switch", "-q", "main")
		case "tips", "duplicates":
			tree, parent := r.git("rev-parse", "HEAD^{tree}"), r.git("rev-parse", "HEAD")
			duplicate := r.git("commit-tree", tree, "-p", parent, "-m", "shared branch tip")
			for i := 0; i < n; i++ {
				tip := duplicate
				if axis == "tips" {
					tip = r.git("commit-tree", tree, "-p", parent, "-m", fmt.Sprintf("distinct branch tip %d", i))
				}
				r.git("branch", fmt.Sprintf("growth-%d", i), tip)
			}
		case "worktrees":
			base := t.TempDir()
			for i := 0; i < n; i++ {
				r.git("worktree", "add", "-q", "--detach", filepath.Join(base, fmt.Sprintf("wt-%d", i)), "main")
			}
		}
		env, err := openMigrate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		main, err := env.main.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		trace := filepath.Join(t.TempDir(), "git.trace")
		t.Setenv("GIT_TRACE", trace)
		in, observationErr := migrationInventory(env, main)
		t.Setenv("GIT_TRACE", "")
		if observationErr != nil {
			t.Fatalf("%s/%d: %v", axis, n, observationErr)
		}
		manifest := tracker.PlanTrackerMigration(in)
		if len(manifest.Refusals) != 0 {
			t.Fatalf("%s/%d fixture refused: %+v", axis, n, manifest.Refusals)
		}
		if axis == "documents" && (len(in.Active) != documents || len(in.Archived) != documents*2 || len(in.Branches) != documents) {
			t.Fatalf("document population not observed: active=%d archives=%d changes=%d", len(in.Active), len(in.Archived), len(in.Branches))
		}
		if axis == "tips" && len(in.Tips) != n {
			t.Fatalf("distinct tip population: got %d want %d", len(in.Tips), n)
		}
		if axis == "duplicates" && (len(in.Tips) != 1 || len(in.Tips[0].Carriers) != n) {
			t.Fatalf("duplicate carriers lost or rescanned: %+v", in.Tips)
		}
		raw, err := os.ReadFile(trace)
		if err != nil {
			t.Fatal(err)
		}
		got := counts{}
		for _, line := range strings.Split(string(raw), "\n") {
			if !strings.Contains(line, "built-in: git ") {
				continue
			}
			got.total++
			if strings.Contains(line, "cat-file --batch") {
				got.batches++
			}
		}
		if got.total == 0 {
			t.Fatal("Git tracing recorded no commands")
		}
		t.Logf("%s/%d: Git processes=%d, cat-file batches=%d", axis, n, got.total, got.batches)
		return got
	}
	docs8, docs16 := count("documents", 8), count("documents", 16)
	if docs8 != docs16 || docs8.batches == 0 {
		t.Errorf("document count caused per-file processes: %+v -> %+v", docs8, docs16)
	}
	tips8, tips16 := count("tips", 8), count("tips", 16)
	if tips16.total <= tips8.total || tips16.total > 2*tips8.total {
		t.Errorf("distinct tip scan must grow at most linearly: %+v -> %+v", tips8, tips16)
	}
	duplicates8, duplicates16 := count("duplicates", 8), count("duplicates", 16)
	if duplicates8 != duplicates16 {
		t.Errorf("same-OID carriers caused repeated reads: %+v -> %+v", duplicates8, duplicates16)
	}
	worktrees0, worktrees8, worktrees16 := count("worktrees", 0), count("worktrees", 8), count("worktrees", 16)
	if worktrees8.total-worktrees0.total != 16 || worktrees16.total-worktrees8.total != 16 || worktrees0.batches != worktrees16.batches {
		t.Errorf("worktrees require two fixed index/status processes each and no repeated blobs: %+v -> %+v -> %+v", worktrees0, worktrees8, worktrees16)
	}
	assets8, assets16 := count("assets", 8), count("assets", 16)
	if assets8 != assets16 {
		t.Errorf("unrelated asset volume changed process/blob-read counts: %+v -> %+v", assets8, assets16)
	}
}
