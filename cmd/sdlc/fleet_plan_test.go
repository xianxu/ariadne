package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory(t *testing.T) {
	path := filepath.Join(repoRootForTest(t), "workshop", "plans", "000200-sdlc-fleet-thread-inventory-plan.md")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	const heading = "### 2026-08-27 — authoritative corrected Core concepts inventory"
	start := strings.Index(text, heading)
	if start < 0 {
		t.Fatal("fleet plan missing authoritative corrected Core concepts inventory")
	}
	section := text[start:]
	if next := strings.Index(section[len(heading):], "\n### "); next >= 0 {
		section = section[:len(heading)+next]
	}
	normalized := strings.Join(strings.Fields(section), " ")
	for _, want := range []string{
		"| `PolicyDiagnostic` / `PolicyCapability` / `PolicyResult` | PURE | `cmd/sdlc/internal/fleet/types.go` | new |",
		"| `ResolvePolicy` | PURE | `cmd/sdlc/internal/fleet/policy.go` | new |",
		"| `IssueAssociation` | PURE | `cmd/sdlc/internal/fleet/issues.go` | new |",
		"| `AssociateBranchIssue` | INTEGRATION | `cmd/sdlc/internal/fleet/issues.go` | new |",
		"| `RenderInventory` / `RenderPolicy` | INTEGRATION | `cmd/sdlc/internal/fleet/render.go` | new |",
	} {
		if !strings.Contains(normalized, want) {
			t.Errorf("corrected Core concepts inventory missing %q:\n%s", want, section)
		}
	}
}
