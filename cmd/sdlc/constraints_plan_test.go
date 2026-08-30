package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var constraintsPlanInventoryRows = []string{
	"| `ARCH-CONSTRAINTS` registry entry | PURE | `cmd/sdlc/internal/judge/architecture.md` | new |",
	"| `constraintsClauseContracts` | PURE, test-only | `cmd/sdlc/internal/judge/judge_test.go` | new |",
	"| `architectureEntry` | PURE, test-only | `cmd/sdlc/internal/judge/judge_test.go` | new |",
	"| `architectureClause` | PURE, test-only | `cmd/sdlc/internal/judge/judge_test.go` | new |",
	"| `constraintsContractViolations` | PURE, test-only | `cmd/sdlc/internal/judge/judge_test.go` | new |",
	"| `predicateIsNegated` | PURE, test-only | `cmd/sdlc/internal/judge/judge_test.go` | new |",
	"| `validConstraintsRegistryForTest` | PURE, test-only | `cmd/sdlc/internal/judge/judge_test.go` | new |",
	"| `constraintsPlanInventoryRows` | PURE, test-only | `cmd/sdlc/constraints_plan_test.go` | new |",
	"| `constraintsPlanInventoryViolations` | PURE, test-only | `cmd/sdlc/constraints_plan_test.go` | new |",
	"| `dry.prompt` | generated integration evidence | `cmd/sdlc/internal/judge/testdata/golden/dry.prompt` | modified |",
	"| `pure.prompt` | generated integration evidence | `cmd/sdlc/internal/judge/testdata/golden/pure.prompt` | modified |",
	"| `plan-quality.prompt` | generated integration evidence | `cmd/sdlc/internal/judge/testdata/golden/plan-quality.prompt` | modified |",
	"| `milestone-review.prompt` | generated integration evidence | `cmd/sdlc/internal/judge/testdata/golden/milestone-review.prompt` | modified |",
	"| `ArchitectureMarkers` | PURE | `cmd/sdlc/internal/judge/architecture.go` | unchanged; output derives fifth marker |",
	"| `ArchitectureBlock` | PURE | `cmd/sdlc/internal/judge/architecture.go` | unchanged; output embeds expanded registry |",
	"| `BuildPrompt` | PURE | `cmd/sdlc/internal/judge/prompts.go` | unchanged; output derives four architecture-aware prompts |",
	"| `CodeReviewBody` | PURE | `cmd/sdlc/internal/judge/review.go` | unchanged; output derives complete marker enumeration |",
	"| `runArchPrinciples` | INTEGRATION | `cmd/sdlc/archprinciples.go` | unchanged; output derives CLI pull |",
	"| `runStartPlan` | INTEGRATION | `cmd/sdlc/startplan.go` | unchanged; output derives planning push |",
}

func TestConstraintsPlanHasAuthoritativeCoreConceptInventory(t *testing.T) {
	body := readConstraintsPlan(t)
	if violations := constraintsPlanInventoryViolations(body); len(violations) > 0 {
		t.Fatalf("#205 authoritative Core concepts inventory violations:\n  %s", strings.Join(violations, "\n  "))
	}
}

func TestConstraintsPlanCoreConceptInventoryMutants(t *testing.T) {
	body := readConstraintsPlan(t)
	for _, row := range constraintsPlanInventoryRows {
		t.Run(row, func(t *testing.T) {
			start := strings.LastIndex(body, row)
			if start < 0 {
				t.Fatalf("fixture missing inventory row %q", row)
			}
			mutant := body[:start] + body[start+len(row):]
			if violations := constraintsPlanInventoryViolations(mutant); len(violations) == 0 {
				t.Fatal("removed inventory row unexpectedly satisfies contract")
			}
		})
	}
}

func readConstraintsPlan(t *testing.T) string {
	t.Helper()
	root := repoRootForTest(t)
	for _, relative := range []string{
		"workshop/plans/000205-make-operating-constraints-explicit-plan.md",
		"workshop/history/plans/000205-make-operating-constraints-explicit-plan.md",
	} {
		body, err := os.ReadFile(filepath.Join(root, relative))
		if err == nil {
			return string(body)
		}
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	t.Fatal("#205 plan not found in active or historical plans")
	return ""
}

func constraintsPlanInventoryViolations(body string) []string {
	const heading = "### 2026-08-29T18:38:00-07:00 — authoritative Core concepts inventory contract"
	start := strings.Index(body, heading)
	if start < 0 {
		return []string{"missing authoritative inventory heading"}
	}
	section := body[start:]
	if next := strings.Index(section[len(heading):], "\n### "); next >= 0 {
		section = section[:len(heading)+next]
	}
	normalized := strings.Join(strings.Fields(section), " ")
	var violations []string
	for _, row := range constraintsPlanInventoryRows {
		if !strings.Contains(normalized, row) {
			violations = append(violations, "missing "+row)
		}
	}
	return violations
}
