package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func TestSessionContinuityPolicyFansOutToEveryHarness(t *testing.T) {
	repo := filepath.Join("..", "..")
	policyRaw, err := os.ReadFile(filepath.Join(repo, "AGENTS.base.md"))
	if err != nil {
		t.Fatal(err)
	}
	const policyHeading = "### 14. Session Continuity"
	policy := string(policyRaw)
	start := strings.Index(policy, policyHeading)
	if start < 0 {
		t.Fatalf("AGENTS.base.md missing %q", policyHeading)
	}
	policy = policy[start:]
	end := strings.Index(policy, "\n## Core Design Principles")
	if end < 0 {
		t.Fatal("Session Continuity policy missing Core Design Principles boundary")
	}
	policy = policy[:end]
	const triggerClause = "When the harness reports that the active context is more than 60% full, proactively checkpoint the session before starting another substantial unit of work."
	const checkpointClause = "Finish the current atomic action and update its durable issue/plan/log state first."
	const fallbackClause = "If an exact percentage is unavailable, treat a harness context-pressure or compaction warning as the trigger."
	const continuationRoute = "Apply the canonical **`continuation` datatype** for the checkpoint"
	for name, marker := range map[string]string{
		"threshold-to-action clause": triggerClause,
		"checkpoint ordering clause": checkpointClause,
		"fallback trigger clause":    fallbackClause,
		"canonical route":            continuationRoute,
		"writer boundary":            "writer owns the restart",
		"durable-write precondition": "after a successful durable write",
		"no double restart":          "don't separately restart",
		"no-writer behavior":         "no writer",
	} {
		if !strings.Contains(policy, marker) {
			t.Errorf("Session Continuity policy missing %s marker %q", name, marker)
		}
	}
	if checkpointAt, routeAt := strings.Index(policy, checkpointClause), strings.Index(policy, continuationRoute); checkpointAt >= routeAt {
		t.Errorf("Session Continuity policy must checkpoint durable state before applying continuation: checkpoint=%d route=%d", checkpointAt, routeAt)
	}

	proto, err := os.ReadFile(filepath.Join(repo, "construct", "datatype", "continuation.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(proto), "60%") {
		t.Error("60% trigger belongs only in AGENTS.base.md, not the continuation procedure")
	}

	manifest, err := os.ReadFile(filepath.Join(repo, "construct", "base.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	intents, err := intent.ParseManifest(string(manifest))
	if err != nil {
		t.Fatalf("parse construct/base.manifest: %v", err)
	}
	exportSource := ""
	for _, in := range intents {
		if in.Kind == intent.Prose && in.Visibility == intent.Export && in.Source == "AGENTS.base.md" {
			exportSource = in.Source
			break
		}
	}
	if exportSource == "" {
		t.Fatal("construct/base.manifest missing active exported prose AGENTS.base.md intent")
	}
	exportRow := "export prose " + exportSource

	parent := t.TempDir()
	foundation := filepath.Join(parent, "foundation")
	leaf := filepath.Join(parent, "leaf")
	mkfile(t, filepath.Join(foundation, "construct", "base.manifest"), exportRow+"\n")
	mkfile(t, filepath.Join(foundation, "AGENTS.base.md"), policy)
	mkfile(t, filepath.Join(leaf, "construct", "deps"), "substrate ../foundation\n")
	mkfile(t, filepath.Join(leaf, "construct", "base.manifest"), "internal prose AGENTS.local.md\n")
	mkfile(t, filepath.Join(leaf, "AGENTS.local.md"), "# Leaf policy\n")

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("compile policy fixture: %v", err)
	}
	for _, entry := range plan.TargetAll.EntryFiles() {
		raw, err := os.ReadFile(filepath.Join(leaf, entry))
		if err != nil {
			t.Fatalf("read composed %s: %v", entry, err)
		}
		if !strings.Contains(string(raw), policy) {
			t.Errorf("%s did not derive the complete Session Continuity policy", entry)
		}
	}
}
