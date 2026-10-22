package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func typedConfirmationLog(t *testing.T, unit, label string, failed bool) []byte {
	t.Helper()
	exit := 0
	if failed {
		exit = 1
	}
	events := []flakes.UnitEvent{{Version: 1, Invocation: strings.Repeat("a", 32), Unit: unit, Event: "start"}}
	if label != "" {
		result := "pass"
		if failed {
			result = "fail"
		}
		events = append(events, flakes.UnitEvent{Version: 1, Invocation: strings.Repeat("a", 32), Unit: unit, Event: "case", Case: label, Result: result})
	}
	events = append(events, flakes.UnitEvent{Version: 1, Invocation: strings.Repeat("a", 32), Unit: unit, Event: "finish", Exit: &exit, Complete: true})
	var raw []byte
	for _, event := range events {
		b, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		raw = append(raw, append(b, '\n')...)
	}
	return raw
}

func TestTypedConfirmationStoreProof(t *testing.T) {
	unit, label := "lua:tests/native_spec.lua", "native retries admission"
	name := unit + "#" + label
	for _, variant := range []string{"exact", "unrelated-case", "missing-log", "tampered-log", "rejected-zero-exit", "different-contract"} {
		t.Run(variant, func(t *testing.T) {
			store := t.TempDir()
			if err := os.Mkdir(filepath.Join(store, "failed"), 0700); err != nil {
				t.Fatal(err)
			}
			now := time.Now().UTC()
			rows := []SuiteAttempt{
				{Version: 4, ID: "fail", Repo: "fixture", SHA: strings.Repeat("a", 40), Check: suiteCheckName, Exit: 1, Failures: []string{name}, FinishedAt: now, RunnerContractSHA256: strings.Repeat("b", 64)},
				{Version: 4, ID: "pass", Repo: "fixture", SHA: strings.Repeat("a", 40), Check: suiteCheckName, Outcome: string(flakes.Passed), FinishedAt: now.Add(time.Second), RunnerContractSHA256: strings.Repeat("b", 64)},
			}
			for i := range rows {
				selected := label
				if i == 1 && variant == "unrelated-case" {
					selected = "sibling case"
				}
				raw := typedConfirmationLog(t, unit, selected, i == 0)
				rows[i].LogPath = filepath.Join("failed", rows[i].ID+".log")
				rows[i].LogSHA256 = resumeRecordDigest(raw)
				if i == 1 && variant == "missing-log" {
					continue
				}
				if i == 1 && variant == "tampered-log" {
					raw = append(raw, ' ')
				}
				if err := os.WriteFile(filepath.Join(store, rows[i].LogPath), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if variant == "rejected-zero-exit" {
				rows[1].Outcome = string(flakes.Unattributable)
			}
			if variant == "different-contract" {
				rows[1].RunnerContractSHA256 = strings.Repeat("c", 64)
			}
			projected := withTypedLogFacts(flakeAttempts(rows), rows, storeTypedFacts(store))
			_, _, confirmed := flakes.SameHeadFailPass(projected, name)
			if confirmed != (variant == "exact") {
				t.Fatalf("confirmed=%v, want %v", confirmed, variant == "exact")
			}
			proof, missing := flakeEvidence(rows, "fixture", []string{name}, nil, storeTypedFacts(store))
			if (len(proof) == 1) != (variant == "exact") || (len(missing) == 0) != (variant == "exact") {
				t.Fatalf("proof=%v missing=%v", proof, missing)
			}
			hydrated := withSuiteTypedFacts(rows, storeTypedFacts(store))
			stats := aggregateSuiteAttempts(hydrated, now.Add(-time.Hour), "fixture")
			want := 0
			if variant == "exact" {
				want = 1
			}
			if len(stats) != 1 || stats[0].ConfirmedFlakeAttempts != want {
				t.Fatalf("stats=%+v want confirmed=%d", stats, want)
			}
		})
	}
}

func TestTypedFactsHydrationDoesNotReadLegacyLogs(t *testing.T) {
	rows := []SuiteAttempt{{Version: 2, ID: "legacy", Check: suiteCheckName, Exit: 0}}
	calls := 0
	got := withTypedLogFacts(flakeAttempts(rows), rows, func(SuiteAttempt) (flakes.UnitFacts, error) { calls++; return flakes.UnitFacts{}, nil })
	if calls != 0 || !got[0].Passed {
		t.Fatalf("calls=%d attempts=%+v", calls, got)
	}
}

// The production check-name projection must retain the distinction between a
// full-suite pass, an affected selection and a targeted flake child. Exact
// typed coverage is necessary but only a full-suite pass registers a typed
// flake; legacy Go evidence still accepts its targeted passing child.
func TestTypedConfirmationRequiresCompleteSuiteScope(t *testing.T) {
	for _, name := range []string{"script:tests/check.sh", "lua:tests/check_spec.lua#native selected case", "example.com/fixture.TestGo"} {
		id, err := flakes.ParseIdentity(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, check := range []string{suiteCheckName, affectedCheckName, suiteRerunCheck} {
			t.Run(name+"/"+check, func(t *testing.T) {
				store := t.TempDir()
				if err := os.Mkdir(filepath.Join(store, "failed"), 0700); err != nil {
					t.Fatal(err)
				}
				now := time.Now().UTC()
				rows := []SuiteAttempt{
					{Version: 4, ID: "failed", Repo: "fixture", SHA: strings.Repeat("a", 40), Check: suiteCheckName, Exit: 1, Failures: []string{name}, FinishedAt: now, RunnerContractSHA256: strings.Repeat("b", 64)},
					{Version: 4, ID: "passed", Repo: "fixture", SHA: strings.Repeat("a", 40), Check: check, Exit: 0, Outcome: string(flakes.Passed), FinishedAt: now.Add(time.Second), RunnerContractSHA256: strings.Repeat("b", 64)},
				}
				if id.Kind() == flakes.KindGo {
					rows[0].Version = 2
					rows[1].Version = 2
				} else {
					for i := range rows {
						raw := typedConfirmationLog(t, id.Unit(), id.Case(), i == 0)
						rows[i].LogPath = filepath.Join("failed", rows[i].ID+".log")
						rows[i].LogSHA256 = resumeRecordDigest(raw)
						if err := os.WriteFile(filepath.Join(store, rows[i].LogPath), raw, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				want := check == suiteCheckName || id.Kind() == flakes.KindGo && check == suiteRerunCheck
				projected := withTypedLogFacts(flakeAttempts(rows), rows, storeTypedFacts(store))
				if _, ok := flakes.Confirmed(projected, 0); ok != want {
					t.Errorf("whole confirmation=%v want %v", ok, want)
				}
				proof, missing := flakeEvidence(rows, "fixture", []string{name}, nil, storeTypedFacts(store))
				if (len(proof) == 1) != want || (len(missing) == 0) != want {
					t.Errorf("registration proof=%v missing=%v want confirmation %v", proof, missing, want)
				}
				stats := aggregateSuiteAttempts(withSuiteTypedFacts(rows, storeTypedFacts(store)), now.Add(-time.Hour), "fixture")
				count := 0
				if want {
					count = 1
				}
				if len(stats) != 1 || stats[0].ConfirmedFlakeAttempts != count {
					t.Errorf("stats=%+v want confirmed=%d", stats, count)
				}
			})
		}
	}
}
