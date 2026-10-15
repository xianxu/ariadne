package judge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func usageIntent() reviewusage.Intent {
	return reviewusage.Intent{Repository: "github.com/example/repo", IssueID: "000365", Recipe: "plan-quality", Round: 1, ReviewedHEAD: strings.Repeat("a", 40), Slot: "repo:3"}
}

func TestDispatchReviewUsageStatefulCLIAccountsEachRetry(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	store, err := reviewusage.NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(t.TempDir(), "attempt-count")
	script := filepath.Join(t.TempDir(), "fake-reviewer.sh")
	body := fmt.Sprintf(`#!/bin/sh
count=0
if [ -f %s ]; then count=$(cat %s); fi
count=$((count + 1))
printf '%%s' "$count" > %s
printf '{"type":"system","subtype":"init","session_id":"00000000-0000-4000-8000-%%012d"}\n' "$count"
if [ "$count" -eq 1 ]; then text='no verdict'; else text='VERDICT: CLEAN'; fi
printf '{"type":"assistant","message":{"content":[{"type":"text","text":"%%s"}]}}\n' "$text"
printf '{"type":"result","subtype":"success","session_id":"00000000-0000-4000-8000-%%012d"}\n' "$count"
`, shellQuote(state), shellQuote(state), shellQuote(state))
	if err := os.WriteFile(script, []byte(body), 0700); err != nil {
		t.Fatal(err)
	}
	saved := Run
	t.Cleanup(func() { Run = saved })
	Run = runProcess
	base := usageIntent()
	base.ReviewedHEAD = strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	evidence := DispatchEvidence{}
	text, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Checkout: root, Containment: ContainCheckout, UsageIntent: &base, UsageRecorder: store.Record, Evidence: &evidence, Wrap: func(string, []string) (string, []string) { return "/bin/sh", []string{script} }})
	if err != nil || !strings.Contains(text, "CLEAN") {
		t.Fatalf("stateful retry: %q %v", text, err)
	}
	entries, err := store.List(context.Background())
	if err != nil || len(entries) != 2 {
		t.Fatalf("attempt receipts %v: %v", entries, err)
	}
	seen := map[int]reviewusage.Record{}
	for _, entry := range entries {
		r := entry.Record
		if r.Phase != reviewusage.Finished || r.PID <= 0 || r.SessionID == "" {
			t.Fatalf("launched attempt lost binding: %+v", r)
		}
		seen[r.Intent.Attempt] = r
	}
	if len(seen) != 2 || seen[1].Intent.ID == seen[2].Intent.ID || seen[1].SessionID == seen[2].SessionID {
		t.Fatal("retry overwrote first attempt")
	}
	if len(evidence.Protocol) != 1 || len(evidence.ReviewerRuns) != 2 || len(evidence.UsageDiagnostics) != 0 {
		t.Fatalf("protocol/accounting evidence %+v", evidence)
	}
}

// This opt-in launches the installed CLI using the unchanged production argv.
// Run on a supported CLI upgrade or before a benchmark batch, never in usage.
func TestLiveReviewUsageIdentityConformance(t *testing.T) {
	if os.Getenv("WF_LIVE_USAGE_CONFORMANCE") != "1" {
		t.Skip("opt-in live CLI identity conformance")
	}
	for _, agent := range []AgentCLI{AgentClaude, AgentCodex} {
		t.Run(string(agent), func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
			// The CLIs record the canonical cwd; macOS temp roots commonly
			// enter through /var or /tmp symlinks to /private equivalents.
			root, err := filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			name, args, err := BuildArgs(DispatchOptions{Agent: agent, Prompt: "Respond only with VERDICT: CLEAN. Do not invoke tools.", Containment: ContainNone})
			if err != nil {
				t.Fatal(err)
			}
			pid := 0
			out, err := runProcess(ctx, func(p int) { pid = p }, name, args...)
			if err != nil || pid <= 0 {
				t.Fatalf("live CLI did not return: %v", err)
			}
			id, err := reviewusage.CaptureIdentity(string(agent), out.Stdout, out.Stderr)
			if err != nil {
				t.Fatalf("live startup identity unavailable: %v", err)
			}
			// Use the usage feature's raw-store metadata adapter. Exact
			// session identity and cwd, never message bodies, bind this run.
			discovery, err := transcripts.DiscoverUsageEvidence(ctx, transcripts.DefaultHarnesses(), time.Now().Add(-5*time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			matched := false
			for _, source := range discovery.Sessions {
				if source.Harness == string(agent) && source.ID == id && filepath.Clean(source.CWD) == filepath.Clean(root) {
					matched = true
				}
			}
			if !matched {
				t.Fatal("CLI identity absent from recorded transcript metadata")
			}
		})
	}
}

const usageSession = "11111111-1111-4111-8111-111111111111"

func usageStream() []byte {
	return []byte(`{"type":"system","subtype":"init","session_id":"` + usageSession + `"}` + "\n" + `{"type":"assistant","message":{"content":[{"type":"text","text":"VERDICT: CLEAN"}]}}` + "\n" + `{"type":"result","subtype":"success","session_id":"` + usageSession + `"}` + "\n")
}
func TestDispatchReviewUsageBothReturnPathsAndCancellation(t *testing.T) {
	for _, progress := range []bool{false, true} {
		for _, canceled := range []bool{false, true} {
			t.Run(string([]byte{'0' + byte(boolInt(progress)), '0' + byte(boolInt(canceled))}), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				saved := Run
				t.Cleanup(func() { Run = saved })
				Run = func(ctx context.Context, onStart func(int), _ string, _ ...string) (ProcessOutput, error) {
					onStart(42)
					if canceled {
						cancel()
						return ProcessOutput{Stdout: usageStream()}, ctx.Err()
					}
					return ProcessOutput{Stdout: usageStream()}, nil
				}
				base := usageIntent()
				var records []reviewusage.Record
				evidence := DispatchEvidence{}
				opts := DispatchOptions{Agent: AgentClaude, Checkout: t.TempDir(), UsageIntent: &base, Evidence: &evidence, UsageRecorder: func(ctx context.Context, r reviewusage.Record) error {
					if r.Phase == reviewusage.Finished {
						if ctx.Err() != nil {
							t.Error("terminal accounting reused canceled context")
						}
						deadline, ok := ctx.Deadline()
						if !ok || time.Until(deadline) > 2*time.Second {
							t.Error("terminal accounting context not bounded")
						}
					}
					records = append(records, r)
					return nil
				}}
				if progress {
					opts.Stderr = io.Discard
				}
				text, err := dispatchOnce(ctx, opts)
				if canceled {
					if !errors.Is(err, context.Canceled) || text != "" {
						t.Fatalf("cancellation changed semantic policy: %q %v", text, err)
					}
				} else if err != nil || !strings.Contains(text, "CLEAN") {
					t.Fatalf("response %q %v", text, err)
				}
				if len(records) != 3 || records[0].Phase != reviewusage.Requested || records[1].Phase != reviewusage.Started || records[2].Phase != reviewusage.Finished {
					t.Fatalf("missing attempt progress: %+v", records)
				}
				last := records[2]
				expectedRoot, rootErr := filepath.EvalSymlinks(opts.Checkout)
				if rootErr != nil {
					t.Fatal(rootErr)
				}
				if last.SessionID != usageSession || last.PID != 42 || last.Intent.Root != expectedRoot {
					t.Fatalf("wrong process binding: %+v", last)
				}
				if canceled && last.Outcome != "canceled" {
					t.Fatal("canceled return not accounted")
				}
				if len(evidence.ReviewerRuns) != 1 || evidence.ReviewerRuns[0] != last {
					t.Fatal("copy lost terminal accounting")
				}
			})
		}
	}
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func TestDispatchReviewUsageSaveFailurePreventsLaunch(t *testing.T) {
	saved := Run
	t.Cleanup(func() { Run = saved })
	calls := 0
	Run = func(context.Context, func(int), string, ...string) (ProcessOutput, error) {
		calls++
		return ProcessOutput{}, nil
	}
	base := usageIntent()
	_, err := dispatchOnce(context.Background(), DispatchOptions{Agent: AgentClaude, Checkout: t.TempDir(), UsageIntent: &base, UsageRecorder: func(context.Context, reviewusage.Record) error { return errors.New("requested save failed") }})
	if err == nil || calls != 0 {
		t.Fatal("failed requested persistence launched provider")
	}
}

func TestDispatchReviewUsageCodexPreambleCapture(t *testing.T) {
	saved := Run
	t.Cleanup(func() { Run = saved })
	const preamble = "OpenAI Codex (v0.1.0)\n--------\nworkdir: /repo\nmodel: gpt-6\nprovider: openai\napproval: never\nsandbox: read-only\nreasoning effort: high\nreasoning summaries: auto\nsession id: " + usageSession + "\n--------\nuser\nprompt\n"
	for _, valid := range []bool{true, false} {
		Run = func(_ context.Context, onStart func(int), _ string, _ ...string) (ProcessOutput, error) {
			onStart(42)
			stderr := preamble
			if !valid {
				stderr = "user\n" + preamble
			}
			return ProcessOutput{Stdout: []byte("VERDICT: CLEAN"), Stderr: []byte(stderr)}, nil
		}
		base := usageIntent()
		evidence := DispatchEvidence{}
		text, err := dispatchOnce(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: t.TempDir(), UsageIntent: &base, Evidence: &evidence, UsageRecorder: func(context.Context, reviewusage.Record) error { return nil }})
		if err != nil || text != "VERDICT: CLEAN" || len(evidence.ReviewerRuns) != 1 {
			t.Fatalf("capture changed review semantics: %q %v", text, err)
		}
		r := evidence.ReviewerRuns[0]
		if valid {
			if r.SessionID != usageSession || r.Diagnostic != "" {
				t.Fatalf("startup identity lost: %+v", r)
			}
		} else if r.SessionID != "" || r.Diagnostic == "" {
			t.Fatal("echoed identity was guessed")
		}
	}
}
func TestDispatchReviewUsageStartWriteFailureRetainsPartialEvidence(t *testing.T) {
	saved := Run
	t.Cleanup(func() { Run = saved })
	Run = func(ctx context.Context, onStart func(int), _ string, _ ...string) (ProcessOutput, error) {
		onStart(42)
		return ProcessOutput{Stdout: usageStream()}, nil
	}
	base := usageIntent()
	evidence := DispatchEvidence{}
	durable := map[string]reviewusage.Record{}
	_, err := dispatchOnce(context.Background(), DispatchOptions{Agent: AgentClaude, Checkout: t.TempDir(), UsageIntent: &base, Evidence: &evidence, UsageRecorder: func(_ context.Context, r reviewusage.Record) error {
		if r.Phase != reviewusage.Requested {
			return errors.New("accounting store unavailable")
		}
		durable[r.Intent.ID] = r
		return nil
	}})
	if err != nil {
		t.Fatalf("nonsemantic persistence failure changed verdict: %v", err)
	}
	if len(durable) != 1 || len(evidence.UsageDiagnostics) != 2 || len(evidence.ReviewerRuns) != 1 {
		t.Fatal("partial persistence not visible")
	}
	for _, r := range durable {
		if r.Phase != reviewusage.Requested {
			t.Fatal("unknown write claimed progress")
		}
	}
}

func TestDispatchReviewUsageCanonicalRoot(t *testing.T) {
	saved := Run
	t.Cleanup(func() { Run = saved })
	Run = func(_ context.Context, onStart func(int), _ string, _ ...string) (ProcessOutput, error) {
		onStart(42)
		return ProcessOutput{Stdout: usageStream()}, nil
	}
	physical := t.TempDir()
	canonical, err := filepath.EvalSymlinks(physical)
	if err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(t.TempDir(), "review-root-link")
	if err := os.Symlink(physical, linked); err != nil {
		t.Fatal(err)
	}
	base := usageIntent()
	evidence := DispatchEvidence{}
	_, err = dispatchOnce(context.Background(), DispatchOptions{Agent: AgentClaude, Checkout: linked, UsageIntent: &base, Evidence: &evidence, UsageRecorder: func(context.Context, reviewusage.Record) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if len(evidence.ReviewerRuns) != 1 || evidence.ReviewerRuns[0].Intent.Root != canonical {
		t.Fatal("accounting root does not bind physical checkout")
	}
}
