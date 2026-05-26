package main

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestMilestoneCloseCmd_Registered(t *testing.T) {
	cmd := NewMilestoneCloseCmd()
	if cmd.Use != "milestone-close" {
		t.Errorf("Use = %q want milestone-close", cmd.Use)
	}
	for _, want := range []string{"issue", "milestone", "actual", "verified", "force", "dry-run", "no-judge", "agent"} {
		if cmd.Flags().Lookup(want) == nil {
			t.Errorf("flag --%s not registered", want)
		}
	}
}

func TestDispatchMilestoneReview_PromptBuildAndDispatch(t *testing.T) {
	// Replace judge.Run with a fake that captures the args, returns a
	// "clean" response. Lets us verify the dispatch path without
	// invoking a real agent.
	orig := judge.Run
	defer func() { judge.Run = orig }()

	var seenName string
	var seenArgs []string
	judge.Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		seenName = name
		seenArgs = args
		return []byte("No DRY violations found.\n"), nil
	}

	prompt := judge.BuildPrompt(judge.MilestoneReview, judge.PromptInput{
		Diff:     "+ new code\n",
		Base:     "abc123",
		Head:     "HEAD",
		IssueRef: "ariadne#31 M6",
	})
	if !strings.Contains(prompt, "ariadne#31 M6") {
		t.Errorf("prompt missing issue ref:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Base: abc123") {
		t.Errorf("prompt missing base line:\n%s", prompt)
	}

	out, err := judge.Dispatch(context.Background(), judge.DispatchOptions{
		Agent:        judge.AgentClaude,
		Prompt:       prompt,
		AllowedTools: judge.MilestoneReview.AllowedTools(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if judge.Classify(out) != judge.Clean {
		t.Errorf("expected Clean classification, got %s", judge.Classify(out))
	}
	if seenName != "claude" {
		t.Errorf("dispatch agent = %q want claude", seenName)
	}
	if seenArgs[len(seenArgs)-1] != prompt {
		t.Errorf("dispatch prompt not last arg")
	}
}

func TestMilestoneCloseFlagSurface(t *testing.T) {
	// Flag validation lives in runMilestoneClose (not via cobra
	// MarkFlagRequired) so error format matches die()'s red prefix —
	// same posture as close, judge, set-status, etc. Confirm we
	// didn't accidentally mark --milestone required at the cobra layer.
	cmd := NewMilestoneCloseCmd()
	f := cmd.Flags().Lookup("milestone")
	if f == nil {
		t.Fatal("--milestone not registered")
	}
	if _, ok := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; ok {
		t.Errorf("--milestone should NOT use MarkFlagRequired (use runtime die() instead)")
	}
}
