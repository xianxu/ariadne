package main

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
)

func TestBenchCLIExecutionBoundaryRefusesPublicationAndCapturesGateFailure(t *testing.T) {
	c := benchContextFixture(t)
	t.Chdir(c.WorkRoot)
	called := 0
	root := &cobra.Command{Use: "sdlc", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(&cobra.Command{Use: "merge", Aliases: []string{"ship"}, RunE: func(*cobra.Command, []string) error { called++; return nil }})
	if code := executeCLI(root, []string{"ship"}); code != 1 || called != 0 {
		t.Fatalf("bench alias escaped guard: code %d calls %d", code, called)
	}
	events, err := readBenchAudit(context.Background(), &c)
	if err != nil || len(events) != 2 || events[1].Success == nil || *events[1].Success {
		t.Fatalf("boundary failure lost: %+v %v", events, err)
	}
	if len(events[0].Route) != 1 || events[0].Route[0] != "merge" || events[0].Command[0] != "ship" {
		t.Fatalf("canonical gate route lost: %+v", events[0])
	}
}
