package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func TestSessionContinuityPolicyFansOutToEveryHarness(t *testing.T) {
	repo := filepath.Join("..", "..")
	policyRaw, err := os.ReadFile(filepath.Join(repo, "AGENTS.base.md"))
	if err != nil {
		t.Fatal(err)
	}
	policy := string(policyRaw)
	for name, marker := range map[string]string{
		"threshold direction": "more than 60% full",
		"checkpoint boundary": "before starting another substantial unit of work",
		"pressure fallback":   "context-pressure",
		"warning fallback":    "compaction warning",
		"canonical route":     "`continuation` datatype",
		"writer boundary":     "writer owns the restart",
		"no double restart":   "don't separately restart",
		"no-writer behavior":  "no writer",
	} {
		if !strings.Contains(policy, marker) {
			t.Errorf("Session Continuity policy missing %s marker %q", name, marker)
		}
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
	const exportRow = "export    prose AGENTS.base.md"
	if !strings.Contains(string(manifest), exportRow) {
		t.Fatalf("construct/base.manifest missing canonical policy export %q", exportRow)
	}

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
	for _, entry := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
		raw, err := os.ReadFile(filepath.Join(leaf, entry))
		if err != nil {
			t.Fatalf("read composed %s: %v", entry, err)
		}
		if !strings.Contains(string(raw), "### 14. Session Continuity") ||
			!strings.Contains(string(raw), "more than 60% full") {
			t.Errorf("%s did not derive the Session Continuity policy", entry)
		}
	}
}
