package processmanual

import (
	"reflect"
	"sort"
	"testing"
)

// The 12 spine bypass gates (#172), by distinct flag name — the closed vocabulary
// the friction audit measures. `no-start` (claim) is a workflow toggle, not a gate.
func TestGateFlagNames(t *testing.T) {
	got := GateFlagNames()
	want := []string{
		"no-actual", "no-atlas", "no-estimate", "no-estimate-recon",
		"no-judge", "no-plan-check", "no-project", "no-reclose-guard",
		"no-structural", "no-validate", "no-verdict", "no-verified",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GateFlagNames() = %v, want %v", got, want)
	}
}

// Per-command census — the catalog must agree with which spine command carries
// which gate (drift guard in package main asserts this vs the registered flags).
func TestGateFlagsForCommand(t *testing.T) {
	cases := map[string][]string{
		"close": {"no-actual", "no-atlas", "no-judge", "no-plan-check",
			"no-project", "no-reclose-guard", "no-verdict", "no-verified"},
		"milestone-close": {"no-actual", "no-atlas", "no-judge", "no-plan-check",
			"no-project", "no-reclose-guard", "no-verdict", "no-verified"},
		"change-code": {"no-estimate", "no-estimate-recon", "no-judge", "no-structural"},
		"merge":       {"no-judge", "no-validate"},
		"push":        {"no-judge", "no-validate"},
	}
	for cmd, want := range cases {
		got := GateFlagsFor(cmd)
		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GateFlagsFor(%q) = %v, want %v", cmd, got, want)
		}
	}
}

// no-judge on close/mclose auto-dispatches (never refuses); it must be marked so
// the refusal→retry detector doesn't hunt a refusal that can't exist.
func TestNoJudgeCloseHasNoRefusal(t *testing.T) {
	for _, g := range GateCatalog {
		if g.Flag == "no-judge" && contains(g.Commands, "close") {
			if g.HasRefusal {
				t.Errorf("close no-judge should have HasRefusal=false (auto-dispatch skip)")
			}
		}
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
