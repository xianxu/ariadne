package main

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchScoringUnknownOutcomeNeverReexecutes(t *testing.T) {
	task := preparedFixtureTask(t)
	out := t.TempDir()
	calls := 0
	e := benchSessionEvidence{MetisRun: t.TempDir(), Candidate: strings.Repeat("a", 40)}
	d := benchScoringDeps{Evaluate: func(context.Context, bench.Task, string, string, string) (bench.Acceptance, error) {
		calls++
		return bench.Acceptance{Status: "unavailable", Reason: "unknown"}, &runnerCleanupUnconfirmed{Detail: "child alive"}
	}}
	if _, err := scoreBenchEvidence(context.Background(), task, e, t.TempDir(), out, d); err == nil {
		t.Fatal("unknown outcome completed")
	}
	if _, err := scoreBenchEvidence(context.Background(), task, e, t.TempDir(), out, d); err == nil || calls != 1 {
		t.Fatal("hidden evaluation repeated", calls, err)
	}
	if _, err := os.Stat(filepath.Join(out, "acceptance.json")); !os.IsNotExist(err) {
		t.Fatal("unknown became acceptance", err)
	}
}
func TestBenchScoringInvalidProfileNeverExecutes(t *testing.T) {
	task := preparedFixtureTask(t)
	out := t.TempDir()
	e := benchSessionEvidence{MetisRun: t.TempDir(), Candidate: strings.Repeat("a", 40), Measurements: bench.Measurements{Validity: "invalid"}}
	a, err := scoreBenchEvidence(context.Background(), task, e, "", out, benchScoringDeps{Evaluate: func(context.Context, bench.Task, string, string, string) (bench.Acceptance, error) {
		t.Fatal("invalid profile executed")
		return bench.Acceptance{}, errors.New("unreachable")
	}})
	if err != nil || a.Status != "unavailable" {
		t.Fatal(a, err)
	}
}
