package bench

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProjectBenchMetricsOmitsUnknownValues(t *testing.T) {
	got := ProjectBenchMetrics(Measurements{GrossNanos: int64(2 * time.Second)}, nil)
	want := map[string]float64{"wall_seconds": 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestProjectBenchMetricsKeepsMeasuredZerosAndRoleCounters(t *testing.T) {
	m := Measurements{GrossNanos: int64(10 * time.Second), Autonomous: true, Worker: RoleUsage{InputTokens: resultPtr(int64(20)), OutputTokens: resultPtr(int64(0))}, TL: RoleUsage{InputTokens: resultPtr(int64(2))}, Reviewer: RoleUsage{OutputTokens: resultPtr(int64(3))}, ReviewRounds: resultPtr(0), GateRefusals: resultPtr(1)}
	want := map[string]float64{"wall_seconds": 10, "automated_wall_seconds": 10, "human_wait_seconds": 0, "worker_input_tokens": 20, "worker_output_tokens": 0, "tl_input_tokens": 2, "reviewer_output_tokens": 3, "review_rounds": 0, "gate_refusals": 1}
	if got := ProjectBenchMetrics(m, nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestProjectBenchMetricsHiddenRateRequiresCaseEvidence(t *testing.T) {
	a := Acceptance{Status: "cases", CommandPassed: resultPtr(false), Passed: resultPtr(2), Failed: resultPtr(1), Total: resultPtr(3), ParentRed: true, LandingGreen: true, PatchSHA256: strings.Repeat("a", 64)}
	if got := ProjectBenchMetrics(Measurements{}, &a)["hidden_pass_rate"]; got != 2.0/3 {
		t.Fatal(got)
	}
	for _, status := range []string{"command", "unavailable", "cases"} {
		bad := Acceptance{Status: status, CommandPassed: resultPtr(true)}
		if _, ok := ProjectBenchMetrics(Measurements{}, &bad)["hidden_pass_rate"]; ok {
			t.Fatalf("invented rate for %s", status)
		}
	}
}
