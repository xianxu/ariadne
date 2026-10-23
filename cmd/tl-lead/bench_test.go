package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type benchForwardRecorder struct {
	calls    [][]string
	err      error
	deadline bool
}

func (r *benchForwardRecorder) run(ctx context.Context, args []string) ([]byte, error) {
	_, r.deadline = ctx.Deadline()
	r.calls = append(r.calls, append([]string(nil), args...))
	if r.err != nil {
		return nil, r.err
	}
	return []byte(fmt.Sprintf("proposal-%d\n", len(r.calls))), nil
}
func executeBench(t *testing.T, d deps, args ...string) (string, error) {
	t.Helper()
	root := buildRoot(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(append([]string{"bench"}, args...))
	err := root.Execute()
	return out.String(), err
}

func TestBenchLiteralStatefulForwarding(t *testing.T) {
	rec := &benchForwardRecorder{}
	d := deps{benchCommand: rec.run}
	reason := "literal $(touch /tmp/unwanted); `whoami`\nquoted reason"
	root := buildRoot(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"bench", "decision", "--run-id", "run-1", "--kind", "approve", "--reason", reason})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	candidate := strings.Repeat("a", 40)
	root = buildRoot(d)
	root.SetOut(&out)
	root.SetArgs([]string{"bench", "finish", "--run-id", "run-1", "--candidate", candidate})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"bench", "decision", "--run-id", "run-1", "--kind", "approve", "--reason", reason}, {"bench", "finish", "--run-id", "run-1", "--candidate", candidate}}
	if !reflect.DeepEqual(rec.calls, want) || !rec.deadline {
		t.Fatalf("forwarding/timeout: %+v deadline=%v", rec.calls, rec.deadline)
	}
	if out.String() != "proposal-1\nproposal-2\n" {
		t.Fatalf("receipt output changed: %q", out.String())
	}
}

func TestBenchMissingOrInvalidFlagsRefuseBeforeExecution(t *testing.T) {
	cases := [][]string{{"decision"}, {"decision", "--run-id", "r", "--kind", "approve"}, {"decision", "--kind", "approve", "--reason", "ok"}, {"decision", "--run-id", "r", "--kind", "", "--reason", "ok"}, {"decision", "--run-id", "r", "--kind", "approve", "--reason", " "}, {"decision", "--run-id", "r", "--kind", "bad\nkind", "--reason", "ok"}, {"finish"}, {"finish", "--candidate", strings.Repeat("a", 40)}, {"finish", "--run-id", "r", "--candidate", "short"}, {"finish", "--run-id", "../r", "--candidate", strings.Repeat("a", 40)}, {"finish", "--run-id", "r", "--candidate", strings.Repeat("A", 40)}, {"decision", "extra", "--run-id", "r", "--kind", "approve", "--reason", "ok"}}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			rec := &benchForwardRecorder{}
			_, err := executeBench(t, deps{benchCommand: rec.run}, args...)
			if err == nil || len(rec.calls) != 0 {
				t.Fatalf("invalid flags reached owner: calls=%v err=%v", rec.calls, err)
			}
		})
	}
}

func TestBenchUnknownDecisionCategoryRefusesBeforeExecution(t *testing.T) {
	rec := &benchForwardRecorder{}
	_, err := executeBench(t, deps{benchCommand: rec.run}, "decision", "--run-id", "r", "--kind", "approve-plan", "--reason", "ok")
	if err == nil || len(rec.calls) != 0 {
		t.Fatal("category outside durable protocol forwarded")
	}
}

type benchBrokenOutput struct{}

func (benchBrokenOutput) Write([]byte) (int, error) { return 0, errors.New("receipt writer failed") }
func TestBenchFailureExitSurvivesOutputFailure(t *testing.T) {
	root := buildRoot(deps{benchCommand: func(context.Context, []string) ([]byte, error) {
		return []byte("refusal output"), exitError{code: 23, err: errors.New("gate refusal")}
	}})
	root.SetOut(benchBrokenOutput{})
	root.SetArgs([]string{"bench", "finish", "--run-id", "r", "--candidate", strings.Repeat("a", 40)})
	err := root.Execute()
	var exit exitError
	if !errors.As(err, &exit) || exit.code != 23 || !strings.Contains(err.Error(), "receipt writer failed") {
		t.Fatalf("output failure replaced gate refusal: %v", err)
	}
}

func TestBenchForwardedRefusalAndHelp(t *testing.T) {
	refused := exitError{code: 23, err: errors.New("real gate refused")}
	rec := &benchForwardRecorder{err: refused}
	_, err := executeBench(t, deps{benchCommand: rec.run}, "finish", "--run-id", "r", "--candidate", strings.Repeat("a", 64))
	var exit exitError
	if !errors.As(err, &exit) || exit.code != 23 || !strings.Contains(err.Error(), "real gate refused") {
		t.Fatalf("refusal changed: %v", err)
	}
	root := buildRoot(deps{})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"bench", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{"normal TL rules", "tl-lead --help", "fixed", "never codes", "never lands", "trusted agents", "proposal", "--run-id"} {
		if !strings.Contains(out.String(), phrase) {
			t.Errorf("help missing %q", phrase)
		}
	}
}

func TestBenchRunnerProcessHelper(t *testing.T) {
	for i, arg := range os.Args {
		if arg != "--bench-test-child" || i+1 >= len(os.Args) {
			continue
		}
		switch os.Args[i+1] {
		case "refuse":
			fmt.Fprintln(os.Stderr, "gate rejected proposal")
			os.Exit(23)
		case "large":
			fmt.Print(strings.Repeat("x", docRepositoryOutputLimit+1))
			os.Exit(0)
		}
	}
}

func TestBenchBoundedNativeRunnerPreservesExitAndCancellation(t *testing.T) {
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-test.run=^TestBenchRunnerProcessHelper$", "--", "--bench-test-child", "refuse"}
	_, err = runBenchCommand(context.Background(), bin, args)
	var exit exitError
	if !errors.As(err, &exit) || exit.code != 23 || !strings.Contains(err.Error(), "gate rejected proposal") {
		t.Fatalf("native exit lost: %v", err)
	}
	args[len(args)-1] = "large"
	if _, err = runBenchCommand(context.Background(), bin, args); err == nil || !strings.Contains(err.Error(), "exceeds limit") {
		t.Fatalf("unbounded output: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	if _, err = runBenchCommand(ctx, bin, args); err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	if time.Since(started) > 5*time.Second {
		t.Fatal("cancelled call blocked")
	}
}
