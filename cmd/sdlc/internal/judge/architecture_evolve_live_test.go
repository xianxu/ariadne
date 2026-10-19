package judge

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// TestLiveArchitectureEvolve checks observed plan-quality judgments, not canned
// responses. It spends credentials only when explicitly requested:
//
//	SDLC_LIVE_ARCH_EVOLVE=1 go test ./cmd/sdlc/internal/judge -run '^TestLiveArchitectureEvolve$' -count=1 -v -timeout 20m
//
// SDLC_LIVE_ARCH_EVOLVE_AGENT and SDLC_LIVE_ARCH_EVOLVE_MODEL optionally select
// the installed reviewer; the default is codex with its CLI's default model.
func TestLiveArchitectureEvolve(t *testing.T) {
	if os.Getenv("SDLC_LIVE_ARCH_EVOLVE") != "1" {
		t.Skip("set SDLC_LIVE_ARCH_EVOLVE=1 to invoke the real plan-quality judge")
	}
	agent := AgentCLI(os.Getenv("SDLC_LIVE_ARCH_EVOLVE_AGENT"))
	if agent == "" {
		agent = AgentCodex
	}
	if _, err := exec.LookPath(string(agent)); err != nil {
		t.Fatalf("requested live reviewer %q is unavailable: %v", agent, err)
	}
	for _, tc := range []struct {
		name       string
		wantEvolve bool
	}{
		{"optional-field-refusal", true},
		{"additive-round-trip", false},
		{"unsafe-action-guard", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := os.ReadFile(filepath.Join("testdata", "architecture-evolve", tc.name+".md"))
			if err != nil {
				t.Fatal(err)
			}
			checkout := testfix.Repo(t, testfix.InitialCommit())
			if err := os.WriteFile(filepath.Join(checkout, "plan.md"), body, 0o644); err != nil {
				t.Fatal(err)
			}
			testfix.Git(t, checkout, "add", "plan.md")
			testfix.Git(t, checkout, "commit", "-qm", "plan fixture")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			output, err := Dispatch(ctx, DispatchOptions{
				Agent: agent, Model: os.Getenv("SDLC_LIVE_ARCH_EVOLVE_MODEL"),
				Prompt: BuildPrompt(PlanQuality, PromptInput{
					IssueRef: "fixture/" + tc.name, IssueContent: string(body),
				}),
				Checkout: checkout, Containment: ContainCheckout,
				AllowedTools: PlanQuality.AllowedTools(), Timeout: 5 * time.Minute,
			})
			if err != nil {
				t.Fatalf("live dispatch: %v\n%s", err, boundedEvolveText(output))
			}
			report, ok := gatestate.ParseFindingsBlock(output)
			if !ok {
				t.Fatalf("real judge omitted a valid findings block:\n%s", boundedEvolveText(output))
			}
			found, relevant := false, false
			for _, finding := range report.New {
				text := finding.Title + " " + finding.Detail + " " + finding.Family
				t.Logf("%s: %s", finding.Severity, boundedEvolveText(text))
				if strings.Contains(text, "ARCH-EVOLVE") {
					found = true
					lower := strings.ToLower(text)
					relevant = relevant || strings.Contains(lower, "field") || strings.Contains(lower, "version") || strings.Contains(lower, "refus")
				}
			}
			t.Logf("agent=%s findings=%d ARCH-EVOLVE=%t", agent, len(report.New), found)
			if found != tc.wantEvolve || (tc.wantEvolve && !relevant) {
				t.Fatalf("ARCH-EVOLVE finding=%t relevant=%t, want finding=%t:\n%s", found, relevant, tc.wantEvolve, boundedEvolveText(output))
			}
		})
	}
}

func boundedEvolveText(s string) string {
	const limit = 4000
	if len(s) > limit {
		return s[:limit] + " [truncated]"
	}
	return s
}
