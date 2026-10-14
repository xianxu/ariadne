package transcripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEligibleSourcesBoundBrainDescendant(t *testing.T) {
	at := func(min int) time.Time { return time.Date(2026, 10, 13, 0, min, 0, 0, time.UTC) }
	sessions := []SessionEvidence{
		{Harness: "codex", ID: "root", CWD: "/owner", File: "root", Started: at(0)},
		{Harness: "codex", ID: "child", CWD: "/brain", File: "child", Started: at(12), Spawned: true, ParentID: "root"},
		{Harness: "codex", ID: "unbound", CWD: "/brain", File: "unbound", Started: at(12)},
	}
	got, _, err := EligibleSources(sessions, []SourceWindow{{Start: at(10), End: at(20), Began: at(10), Roots: []string{"/owner"}}})
	if err != nil || len(got["child"]) != 1 || len(got["unbound"]) != 0 {
		t.Fatalf("bound descendant lost or unrelated brain admitted: %+v %v", got, err)
	}
}

func TestEligibleSourcesUsesLexicalPaths(t *testing.T) {
	root := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	sessions := []SessionEvidence{{Harness: "codex", ID: "root", File: "root", CWD: alias, Started: evidenceTime(0)}}
	windows := []SourceWindow{{Start: evidenceTime(10), End: evidenceTime(20), Began: evidenceTime(10), Roots: []string{root}}}
	got, _, err := EligibleSources(sessions, windows)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("eligibility resolved filesystem alias: %v", got)
	}
}

func evidenceTime(n int) time.Time { return time.Date(2026, 10, 14, 0, n, 0, 0, time.UTC) }

func TestEligibleSourcesRetainsHistoricalRootAfterRelocate(t *testing.T) {
	sessions := []SessionEvidence{{Harness: "codex", ID: "root", File: "root.jsonl", CWD: "/slot3/repo", Started: evidenceTime(0)}}
	windows := []SourceWindow{{Start: evidenceTime(10), End: evidenceTime(20), Began: evidenceTime(10), Roots: []string{"/slot3/repo"}}, {Start: evidenceTime(20), End: evidenceTime(30), Began: evidenceTime(10), Roots: []string{"/slot3/repo", "/slot4/repo"}}}
	got, _, err := EligibleSources(sessions, windows)
	if err != nil {
		t.Fatal(err)
	}
	if len(got["root.jsonl"]) != 2 || !got["root.jsonl"][1].End.Equal(evidenceTime(30)) {
		t.Fatalf("historical root lost: %v", got)
	}
}

func TestEligibleSourcesRejectsOldSpawnAfterReclaim(t *testing.T) {
	sessions := []SessionEvidence{{Harness: "codex", ID: "root", File: "root", CWD: "/repo", Started: evidenceTime(0)}, {Harness: "codex", ID: "old", ParentID: "root", Spawned: true, File: "old", CWD: "/repo", Started: evidenceTime(5)}, {Harness: "codex", ID: "new", ParentID: "root", Spawned: true, File: "new", CWD: "/repo", Started: evidenceTime(22)}}
	got, _, err := EligibleSources(sessions, []SourceWindow{{Start: evidenceTime(20), End: evidenceTime(30), Began: evidenceTime(20), Roots: []string{"/repo"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got["old"]) != 0 || len(got["new"]) != 1 || !got["new"][0].Start.Equal(evidenceTime(22)) {
		t.Fatalf("wrong spawn eligibility: %v", got)
	}
}

func TestEligibleSourcesVerifiesAncestorAtDispatch(t *testing.T) {
	base := SessionEvidence{Harness: "codex", ID: "root", File: "root", CWD: "/repo", Started: evidenceTime(0)}
	windows := []SourceWindow{{Start: evidenceTime(10), End: evidenceTime(30), Began: evidenceTime(10), Roots: []string{"/repo"}}}
	for _, tc := range []struct {
		name      string
		sessions  []SessionEvidence
		wantErr   bool
		wantChild bool
	}{
		{"missing", []SessionEvidence{{Harness: "codex", ID: "child", ParentID: "missing", Spawned: true, File: "child", CWD: "/repo", Started: evidenceTime(15)}}, true, false},
		{"cycle", []SessionEvidence{{Harness: "codex", ID: "child", ParentID: "child", Spawned: true, File: "child", CWD: "/repo", Started: evidenceTime(15)}}, true, false},
		{"cycle with later ancestor", []SessionEvidence{{Harness: "codex", ID: "child", ParentID: "parent", Spawned: true, File: "child", CWD: "/repo", Started: evidenceTime(15)}, {Harness: "codex", ID: "parent", ParentID: "child", Spawned: true, File: "parent", CWD: "/repo", Started: evidenceTime(25)}}, true, false},
		{"foreign parent", []SessionEvidence{{Harness: "codex", ID: "root", File: "root", CWD: "/other", Started: evidenceTime(0)}, {Harness: "codex", ID: "child", ParentID: "root", Spawned: true, File: "child", CWD: "/repo", Started: evidenceTime(15)}}, false, false},
		{"chain", []SessionEvidence{base, {Harness: "codex", ID: "middle", ParentID: "root", Spawned: true, File: "middle", CWD: "/repo", Started: evidenceTime(12)}, {Harness: "codex", ID: "child", ParentID: "middle", Spawned: true, File: "child", CWD: "/repo", Started: evidenceTime(15)}}, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _, err := EligibleSources(tc.sessions, windows)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v", err)
			}
			if (len(got["child"]) > 0) != tc.wantChild {
				t.Fatalf("eligibility=%v", got)
			}
		})
	}
}

func TestEligibleSourcesPreservesContinuousEngagementOrigin(t *testing.T) {
	sessions := []SessionEvidence{{Harness: "claude", ID: "root", File: "root", CWD: "/repo", Started: evidenceTime(0)}}
	got, _, err := EligibleSources(sessions, []SourceWindow{{Start: evidenceTime(10), End: evidenceTime(20), Began: evidenceTime(5), Roots: []string{"/repo"}}})
	if err != nil || len(got["root"]) != 1 || got["root"][0].Began != evidenceTime(5) {
		t.Fatalf("engagement origin lost: %v %v", got, err)
	}
}
