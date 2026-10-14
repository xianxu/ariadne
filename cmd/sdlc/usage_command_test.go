package main

import (
	"bytes"
	"context"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUsageAuthorityIncludesReviewTailAndPinsCompletion(t *testing.T) {
	t.Run("active local", func(t *testing.T) {
		r, _, _, owner, _ := ownedResolverFixture(t)
		now := time.Date(2026, 10, 14, 0, 20, 0, 0, time.UTC)
		physical, e := filepath.EvalSymlinks(r.origin)
		if e != nil {
			t.Fatal(e)
		}
		r.git("remote", "set-url", "origin", physical)
		if e := os.Rename(r.origin, r.origin+".offline"); e != nil {
			t.Fatal(e)
		}
		t.Cleanup(func() { _ = os.Rename(r.origin+".offline", r.origin) })
		a, e := resolveUsageAuthority(context.Background(), r.root, 335, now)
		if e != nil {
			t.Fatal(e)
		}
		if !a.Until.Equal(now) || len(a.Windows) != 1 || a.Windows[0].Roots[0] != owner.Worktree || !a.Windows[0].End.Equal(now) {
			t.Fatalf("%+v", a)
		}
	})
	t.Run("completed evidence", func(t *testing.T) {
		r, _, _ := ownedCompletedFixture(t, false)
		ownedFixtureAt(t, 30)
		testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "later unrelated")
		a, e := resolveUsageAuthority(context.Background(), r.root, 335, time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC))
		if e != nil {
			t.Fatal(e)
		}
		want := time.Date(2026, 10, 14, 0, 11, 0, 0, time.UTC)
		if !a.Until.Equal(want) {
			t.Fatalf("completion token right edge=%s want%s", a.Until, want)
		}
	})
}
func TestUsageAuthorityPlanningNeedsNoCodeCommit(t *testing.T) {
	r, _, _, _, _ := ownedResolverFixture(t)
	testfix.Git(t, r.root, "reset", "--hard", "HEAD^")
	a, e := resolveUsageAuthority(context.Background(), r.root, 335, time.Date(2026, 10, 14, 0, 25, 0, 0, time.UTC))
	if e != nil || len(a.Windows) == 0 {
		t.Fatal(a, e)
	}
}
func TestUsageFirstOutputFleetGlanceAndJSONNull(t *testing.T) {
	now := time.Now().UTC()
	out := usageOutput{SchemaVersion: 1, AgentGlance: []agentGlance{{Agent: "codex", LastHourTokens: 42, BurnTokensPerHour: 9, BurnWindow: usage.Window{Since: now.Add(-3 * time.Hour), Until: now}, Projection: usage.Projection{Status: "insufficient_calibration"}}}, Window: usage.Window{Since: now.Add(-time.Hour), Until: now}, IssueParts: []usageIssuePart{{Role: "worker", Tokens: 10}, {Role: "reviewer", Tokens: 32, Rounds: 2, Attempts: 3}, {Role: "total", Tokens: 42}}}
	var b bytes.Buffer
	if e := renderUsage(&b, out); e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(b.String(), "codex: last hour 42 tokens; burn 9 tokens/h (3h); insufficient_calibration\n") || !strings.Contains(b.String(), "reviewer: 32 tokens (2 rounds, 3 attempts)") {
		t.Fatal(b.String())
	}
}
func TestUsageCalibrateFlagsAndValidation(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	for _, args := range [][]string{{"calibrate", "--agent", "codex", "--unit", "credits"}, {"calibrate", "--agent", "codex", "--unit", "credits", "--remaining", "NaN"}, {"calibrate", "--agent", "claude", "--unit", "percent", "--remaining", "101"}} {
		cmd := NewUsageCmd()
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		if e := cmd.Execute(); e == nil {
			t.Fatalf("invalid accepted %v", args)
		}
	}
	if e := usage.ValidateObservation(usage.Observation{Agent: "codex", Unit: "credits", Remaining: math.Inf(1), At: time.Now()}); e == nil {
		t.Fatal("inf accepted")
	}
}
func TestUsageSinceValidation(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	for _, s := range []string{"0h", "-1h", "tomorrow", "2027-01-01T00:00:00Z"} {
		if _, e := usageSince(s, now); e == nil {
			t.Fatal(s)
		}
	}
	got, e := usageSince("2h", now)
	if e != nil || !got.Equal(now.Add(-2*time.Hour)) {
		t.Fatal(got, e)
	}
}

func TestUsageSourcePinAllowsAppendButRefusesRewrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source")
	if e := os.WriteFile(path, []byte("one\n"), 0600); e != nil {
		t.Fatal(e)
	}
	pin, _ := os.Stat(path)
	if !usageSourceStable(pin, pin, pin) {
		t.Fatal("stable source refused")
	}
	if e := os.WriteFile(path, []byte("two\n"), 0600); e != nil {
		t.Fatal(e)
	}
	stamp := pin.ModTime().Add(time.Second)
	os.Chtimes(path, stamp, stamp)
	rewritten, _ := os.Stat(path)
	if usageSourceStable(pin, rewritten, rewritten) {
		t.Fatal("same-size mutation accepted")
	}
	if e := os.WriteFile(path, []byte("one\nmore\n"), 0600); e != nil {
		t.Fatal(e)
	}
	appendInfo, _ := os.Stat(path)
	if !usageSourceStable(pin, pin, appendInfo) {
		t.Fatal("append refused")
	}
	other := filepath.Join(t.TempDir(), "replacement")
	os.WriteFile(other, []byte("one\n"), 0600)
	replacement, _ := os.Stat(other)
	if usageSourceStable(pin, replacement, replacement) {
		t.Fatal("replacement accepted")
	}
}
