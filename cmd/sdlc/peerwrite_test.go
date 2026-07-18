package main

import (
	"strings"
	"testing"
)

// TestPlanPeerWrites is the decision table for the safe peer-write commit
// mechanics (#171 M3): a peer repo's project-file edit auto-commits only when
// git state makes the commit unambiguous — on main with a clean index and not
// a brain capture repo. Everything else is report-only: the file is written
// but the commit is handed to the operator with the reason + exact next step.
func TestPlanPeerWrites(t *testing.T) {
	cur := "/fleet/ariadne"
	edits := map[string][]string{
		"/fleet/ariadne": {"workshop/projects/a.md"}, // current repo → rides the normal close commit
		"/fleet/nous":    {"workshop/projects/b.md"}, // peer on main, clean → commit
		"/fleet/metis":   {"workshop/projects/c.md"}, // peer off-main → report-only
		"/fleet/kbench":  {"workshop/projects/d.md"}, // peer on main, staged changes → report-only
		"/fleet/brain":   {"data/project/e.md"},      // brain capture repo → report-only (#176)
		"/fleet/hermes":  {"workshop/projects/f.md"}, // no state entry → unknown → report-only
		"/fleet/tyche":   {"workshop/projects/g.md"}, // target file already dirty → report-only
		"/fleet/erebus":  {"workshop/projects/h.md"}, // branch undeterminable → report-only
	}
	states := map[string]RepoGitState{
		"/fleet/nous":   {Branch: "main"},
		"/fleet/metis":  {Branch: "feature-x"},
		"/fleet/kbench": {Branch: "main", HasStagedChanges: true},
		"/fleet/brain":  {Branch: "main", IsBrain: true},
		"/fleet/tyche":  {Branch: "main", TargetFilesDirty: true},
		"/fleet/erebus": {Branch: ""},
		// /fleet/hermes deliberately absent
	}

	got := planPeerWrites(edits, states, cur, "ariadne#171")

	byRepo := map[string]PeerWriteDecision{}
	for _, d := range got {
		byRepo[d.RepoDir] = d
	}
	if _, ok := byRepo[cur]; ok {
		t.Errorf("current repo %s must be omitted from peer decisions (it rides the close commit)", cur)
	}
	if len(got) != 7 {
		t.Fatalf("want 7 peer decisions, got %d: %+v", len(got), got)
	}

	nous := byRepo["/fleet/nous"]
	if !nous.Commit {
		t.Errorf("nous (main, clean) should commit; reason=%q", nous.Reason)
	}
	if len(nous.Files) != 1 || nous.Files[0] != "workshop/projects/b.md" {
		t.Errorf("nous files = %v, want [workshop/projects/b.md]", nous.Files)
	}
	if !strings.Contains(nous.Message, "ariadne#171") {
		t.Errorf("commit message should cite the closing ref, got %q", nous.Message)
	}

	metis := byRepo["/fleet/metis"]
	if metis.Commit {
		t.Error("metis (off-main) must be report-only")
	}
	if !strings.Contains(metis.Reason, "feature-x") {
		t.Errorf("metis reason should name the branch, got %q", metis.Reason)
	}
	if !strings.Contains(metis.NextAction, "git add") || !strings.Contains(metis.NextAction, "workshop/projects/c.md") {
		t.Errorf("metis next action should be an exact add+commit command, got %q", metis.NextAction)
	}

	kbench := byRepo["/fleet/kbench"]
	if kbench.Commit {
		t.Error("kbench (staged changes) must be report-only")
	}
	if !strings.Contains(kbench.Reason, "staged") {
		t.Errorf("kbench reason should mention staged changes, got %q", kbench.Reason)
	}

	brain := byRepo["/fleet/brain"]
	if brain.Commit {
		t.Error("brain must never be auto-committed into (#176)")
	}
	if !strings.Contains(brain.Reason, "brain") {
		t.Errorf("brain reason should say it is a brain repo, got %q", brain.Reason)
	}

	hermes := byRepo["/fleet/hermes"]
	if hermes.Commit {
		t.Error("hermes (unknown git state) must be report-only, never commit blind")
	}
	if !strings.Contains(hermes.Reason, "unknown") {
		t.Errorf("hermes reason should say the git state is unknown, got %q", hermes.Reason)
	}

	tyche := byRepo["/fleet/tyche"]
	if tyche.Commit {
		t.Error("tyche (dirty target file) must be report-only — never absorb another session's uncommitted edits")
	}
	if !strings.Contains(tyche.Reason, "uncommitted edits") {
		t.Errorf("tyche reason should mention the pre-existing uncommitted edits, got %q", tyche.Reason)
	}

	erebus := byRepo["/fleet/erebus"]
	if erebus.Commit {
		t.Error("erebus (undeterminable branch) must be report-only")
	}
	if !strings.Contains(erebus.Reason, "could not be determined") {
		t.Errorf("erebus reason should say the branch could not be determined (never a garbled error text), got %q", erebus.Reason)
	}

	// Deterministic order: sorted by RepoDir.
	for i := 1; i < len(got); i++ {
		if got[i-1].RepoDir > got[i].RepoDir {
			t.Errorf("decisions not sorted by RepoDir: %s before %s", got[i-1].RepoDir, got[i].RepoDir)
		}
	}
}

func TestPlanPeerWrites_EmptyEdits(t *testing.T) {
	if got := planPeerWrites(nil, nil, "/fleet/ariadne", "ariadne#171"); len(got) != 0 {
		t.Errorf("empty edits should yield no decisions, got %+v", got)
	}
	// Only the current repo edited → still no peer decisions.
	only := map[string][]string{"/fleet/ariadne": {"workshop/projects/a.md"}}
	if got := planPeerWrites(only, nil, "/fleet/ariadne", "ariadne#171"); len(got) != 0 {
		t.Errorf("current-repo-only edits should yield no peer decisions, got %+v", got)
	}
}
