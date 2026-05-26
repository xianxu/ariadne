package judge

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// execCommand is exposed as a var so the test can swap to a fake if
// needed; for now, all tests use the real exec.
var execCommand = exec.Command

func TestIsValid(t *testing.T) {
	for _, name := range []string{"dry", "pure", "plan", "specs", "lessons", "milestone-review"} {
		if !IsValid(name) {
			t.Errorf("IsValid(%q) = false; want true", name)
		}
	}
	for _, name := range []string{"", "dryrun", "Lessons", "milestone"} {
		if IsValid(name) {
			t.Errorf("IsValid(%q) = true; want false", name)
		}
	}
}

func TestCategoryAllowedTools(t *testing.T) {
	if Specs.AllowedTools() != "Edit,Read,Write,Grep,Glob,Bash" {
		t.Errorf("Specs.AllowedTools() = %q", Specs.AllowedTools())
	}
	if DRY.AllowedTools() != "Read,Grep,Glob,Bash" {
		t.Errorf("DRY.AllowedTools() = %q", DRY.AllowedTools())
	}
}

func TestCategoryNeedsAgent(t *testing.T) {
	if Lessons.NeedsAgent() {
		t.Error("Lessons should not need an agent")
	}
	for _, c := range []Category{DRY, PURE, Plan, Specs, MilestoneReview} {
		if !c.NeedsAgent() {
			t.Errorf("%s should need an agent", c)
		}
	}
}

func TestBuildPrompt_DRY(t *testing.T) {
	p := BuildPrompt(DRY, PromptInput{Diff: "DIFF_CONTENT"})
	for _, want := range []string{
		"DRY (Don't Repeat Yourself) violations",
		`No DRY violations found.`,
		"Do NOT modify any files",
		"DIFF_CONTENT",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q\n%s", want, p)
		}
	}
}

func TestBuildPrompt_Plan_ListsIssues(t *testing.T) {
	p := BuildPrompt(Plan, PromptInput{
		Diff:          "DIFF",
		ChangedIssues: []string{"workshop/issues/000031.md", "workshop/issues/000042.md"},
	})
	if !strings.Contains(p, "workshop/issues/000031.md\nworkshop/issues/000042.md") {
		t.Errorf("Plan prompt should list changed issues:\n%s", p)
	}
}

func TestBuildPrompt_Lessons_Empty(t *testing.T) {
	if got := BuildPrompt(Lessons, PromptInput{}); got != "" {
		t.Errorf("Lessons should produce empty prompt, got %q", got)
	}
}

func TestBuildPrompt_MilestoneReview_HasContract(t *testing.T) {
	p := BuildPrompt(MilestoneReview, PromptInput{
		IssueRef: "ariadne#31 M3",
		Base:     "9e8625e",
		Head:     "d7789e0",
		Diff:     "DIFF",
	})
	for _, want := range []string{
		"ariadne#31 M3",
		"Base: 9e8625e",
		"Head: d7789e0",
		"Critical (must fix before next milestone)",
		"anti-collusion property",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("milestone-review prompt missing %q", want)
		}
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   Outcome
	}{
		{"clean dry", "No DRY violations found.", Clean},
		{"clean pure", "No PURE violations found.", Clean},
		{"clean specs", "Everything is in sync.", Clean},
		{"clean plan", "No issue files changed", Clean},
		{"info lessons", LessonsReminder, Info},
		{"failure with content", "Found 2 violations:\n- foo.go:42 duplicates bar.go:99", Failure},
		{"failure empty", "", Failure},
		{"failure whitespace", "   \n\t\n", Failure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.output); got != tt.want {
				t.Errorf("Classify(%q) = %s, want %s", tt.output, got, tt.want)
			}
		})
	}
}

func TestBuildArgs_Claude(t *testing.T) {
	name, args, err := BuildArgs(DispatchOptions{
		Agent:        AgentClaude,
		Prompt:       "review this",
		AllowedTools: "Read,Grep",
	})
	if err != nil {
		t.Fatal(err)
	}
	if name != "claude" {
		t.Errorf("name = %q want claude", name)
	}
	want := []string{"-p", "--allowedTools", "Read,Grep", "--permission-mode", "bypassPermissions", "review this"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Errorf("args = %v\nwant: %v", args, want)
	}
}

func TestBuildArgs_Codex_SandboxAddsFullAuto(t *testing.T) {
	_, args, _ := BuildArgs(DispatchOptions{Agent: AgentCodex, Prompt: "p", IsSandbox: true})
	if args[1] != "--full-auto" {
		t.Errorf("expected --full-auto in sandbox mode, got %v", args)
	}
}

func TestBuildArgs_Gemini_NoSandboxFlag(t *testing.T) {
	_, args, _ := BuildArgs(DispatchOptions{Agent: AgentGemini, Prompt: "p"})
	for _, a := range args {
		if a == "--yolo" {
			t.Errorf("--yolo should not be present without sandbox flag: %v", args)
		}
	}
}

func TestBuildArgs_UnknownAgent(t *testing.T) {
	_, _, err := BuildArgs(DispatchOptions{Agent: "bogus", Prompt: "p"})
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}

func TestBuildArgs_DefaultIsClaude(t *testing.T) {
	name, _, err := BuildArgs(DispatchOptions{Prompt: "p", AllowedTools: "Read"})
	if err != nil {
		t.Fatal(err)
	}
	if name != "claude" {
		t.Errorf("empty Agent should default to claude, got %q", name)
	}
}

func TestFormatCommandLine_Quoting(t *testing.T) {
	cmd, err := FormatCommandLine(DispatchOptions{
		Agent:        AgentClaude,
		Prompt:       "this has spaces and 'quotes'",
		AllowedTools: "Read",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cmd, `'this has spaces and '\''quotes'\'''`) {
		t.Errorf("prompt not properly shell-quoted:\n%s", cmd)
	}
}

func TestDispatch_FakeRun_CapturesOutput(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	var gotName string
	var gotArgs []string
	Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		gotName = name
		gotArgs = args
		return []byte("No DRY violations found.\n"), nil
	}
	out, err := Dispatch(context.Background(), DispatchOptions{
		Agent:        AgentClaude,
		Prompt:       "review",
		AllowedTools: "Read",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotName != "claude" {
		t.Errorf("name = %q", gotName)
	}
	if gotArgs[len(gotArgs)-1] != "review" {
		t.Errorf("last arg should be the prompt, got %v", gotArgs)
	}
	if Classify(out) != Clean {
		t.Errorf("output should classify as Clean, got %s", Classify(out))
	}
}

func TestDispatch_LaunchError_Surfaces(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, errors.New("exec: command not found")
	}
	_, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "x", AllowedTools: "Read"})
	if err == nil {
		t.Error("expected error when Run fails to launch")
	}
}

// Regression for M3 review I3: non-zero exit from the subprocess (whether
// with or without output) should NOT be a Dispatch error — it's a finding
// for Classify to interpret. Real launch failures still error.
func TestDispatch_ExitErrorWithEmptyOutput_NotAnError(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	// Real *exec.ExitError requires a started process. Easiest path:
	// spawn `false` (always exits 1) via the actual exec package, no
	// args needed.
	Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return realExec("false")
	}
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "x", AllowedTools: "Read"})
	if err != nil {
		t.Errorf("non-zero exit should not surface as Dispatch error, got %v", err)
	}
	if Classify(out) != Failure {
		t.Errorf("empty output should classify as Failure, got %s", Classify(out))
	}
}

// realExec runs `false` (or any always-exit-non-zero binary) so we get a
// genuine *exec.ExitError. Wrapped here to keep the test's intent clear.
func realExec(name string) ([]byte, error) {
	return execCommand(name).CombinedOutput()
}
