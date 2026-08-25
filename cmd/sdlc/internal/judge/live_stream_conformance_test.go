package judge

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestLiveAgentStreamConformance is an opt-in check for the external CLI
// channel contract #201 relies on. Run it after upgrading an agent CLI:
//
//	SDLC_LIVE_AGENT_STREAM_CONFORMANCE=1 go test ./cmd/sdlc/internal/judge -run TestLiveAgentStreamConformance -count=1
//
// Normal test runs never spend credentials or make network calls.
func TestLiveAgentStreamConformance(t *testing.T) {
	if os.Getenv("SDLC_LIVE_AGENT_STREAM_CONFORMANCE") != "1" {
		t.Skip("set SDLC_LIVE_AGENT_STREAM_CONFORMANCE=1 to invoke installed agent CLIs")
	}

	for _, agent := range []AgentCLI{AgentClaude, AgentCodex, AgentGemini} {
		t.Run(string(agent), func(t *testing.T) {
			opts := DispatchOptions{
				Agent:        agent,
				Prompt:       "Reply with exactly STREAM_OK and no other text.",
				AllowedTools: "",
				IsSandbox:    true,
			}
			name, args, err := BuildArgs(opts)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := exec.LookPath(name); err != nil {
				t.Skipf("%s is not installed: %v", name, err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			out, err := Run(ctx, nil, name, args...)
			if err != nil {
				t.Fatalf("%s invocation failed: %v\nstderr:\n%s", name, err, out.Stderr)
			}
			if got := strings.TrimSpace(string(out.Stdout)); got != "STREAM_OK" {
				t.Fatalf("%s semantic stdout = %q, want exactly STREAM_OK; stderr:\n%s", name, got, out.Stderr)
			}
			t.Logf("stdout bytes=%d, stderr bytes=%d", len(out.Stdout), len(out.Stderr))
		})
	}
}
