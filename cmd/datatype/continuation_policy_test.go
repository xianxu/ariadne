package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionContinuityPolicyRoutesToDatatypeAndWriter(t *testing.T) {
	root := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "AGENTS.base.md"))
	if err != nil {
		t.Fatal(err)
	}
	agents := string(raw)
	start := strings.Index(agents, "### 14. Session Continuity")
	if start < 0 {
		t.Fatal("AGENTS.base.md missing Session Continuity policy")
	}
	policy := agents[start:]
	if end := strings.Index(policy, "\n## Core Design Principles"); end >= 0 {
		policy = policy[:end]
	}

	for name, marker := range map[string]string{
		"threshold":          "60%",
		"pressure fallback":  "context-pressure",
		"warning fallback":   "compaction warning",
		"canonical route":    "`continuation` datatype",
		"writer boundary":    "writer owns the restart",
		"no-writer behavior": "no writer",
	} {
		if !strings.Contains(policy, marker) {
			t.Errorf("Session Continuity policy missing %s marker %q", name, marker)
		}
	}

	proto, err := os.ReadFile(filepath.Join(root, "construct", "datatype", "continuation.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(proto), "60%") {
		t.Error("60% trigger belongs only in AGENTS.base.md, not the continuation procedure")
	}
}
