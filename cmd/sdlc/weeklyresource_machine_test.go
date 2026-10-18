package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeResourceRunner struct {
	results map[string]struct {
		output string
		err    error
	}
	calls []string
}

func TestDeviceGaugeRecoversAfterUnknownComparison(t *testing.T) {
	// Each intermediate week crosses the JSON boundary used by a published
	// snapshot. Its comparison can be unknown while its probe remains valid.
	for _, tc := range []struct {
		name                 string
		first, second, third DeviceGauge
		compare              func(DeviceGauge, *DeviceGauge) DeviceGauge
		want                 int64
	}{
		{"write after reboot", DeviceGauge{Device: "disk0", Epoch: "boot-1", Value: 100, Observed: true}, DeviceGauge{Device: "disk0", Epoch: "boot-2", Value: 10, Observed: true}, DeviceGauge{Device: "disk0", Epoch: "boot-2", Value: 30, Observed: true}, compareDeviceCounter, 20},
		{"wear after missing probe", DeviceGauge{Device: "disk0", State: "unknown", Reason: "probe unavailable"}, DeviceGauge{Device: "disk0", Value: 8, Observed: true}, DeviceGauge{Device: "disk0", Value: 9, Observed: true}, compareWearGauge, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first := tc.compare(tc.first, nil)
			second := tc.compare(tc.second, &first)
			if second.State != "unknown" || !second.Observed {
				t.Fatalf("intermediate week lost valid observation: %+v", second)
			}
			data, err := json.Marshal(second)
			if err != nil {
				t.Fatal(err)
			}
			var saved DeviceGauge
			if err := json.Unmarshal(data, &saved); err != nil {
				t.Fatal(err)
			}
			if err := validateDeviceGauge(saved, tc.name); err != nil {
				t.Fatal(err)
			}
			third := tc.compare(tc.third, &saved)
			if third.State != "compared" || third.Delta == nil || *third.Delta != tc.want {
				t.Fatalf("third week did not recover: %+v", third)
			}
		})
	}
}

func TestDeviceGaugeRecoversAfterUnknownPreviousWeek(t *testing.T) {
	for _, reason := range []string{"previous week invalid", "observations 9 days apart"} {
		t.Run(reason, func(t *testing.T) {
			current := ResourceMachine{
				BytesWritten: DeviceGauge{Device: "disk0", Epoch: "boot-2", Value: 10, Observed: true},
				SSDWear:      DeviceGauge{Device: "disk0", Value: 8, Observed: true},
			}
			intermediate := compareResourceMachine(current, previousResourceResult{State: "unknown", Reason: reason}, .25)
			if !intermediate.BytesWritten.Observed || !intermediate.SSDWear.Observed || intermediate.BytesWritten.State != "unknown" || intermediate.SSDWear.State != "unknown" {
				t.Fatalf("unknown prior erased current probes: %+v", intermediate)
			}
			section := WeeklyResourceSection{
				ObservedAt:      time.Date(2026, 10, 19, 1, 0, 0, 0, time.UTC),
				Families:        []ResourceComparison{},
				LargestLedgers:  []ResourceFileComparison{},
				Suites:          []WeeklyResourceSuite{},
				CoverageUnknown: []string{"no family inventory"},
				Machine:         intermediate,
			}
			// Fill the independent machine gauges so the persisted-section
			// validator exercises the actual weekly snapshot boundary.
			for _, row := range []*ResourceComparison{&section.Machine.GoCache, &section.Machine.AgentScratchFree, &section.Machine.AgentScratchTotal} {
				row.ResourceObservation = ResourceObservation{Kind: "go-cache", Definition: "/source", Complete: true}
				row.State = "baseline"
			}
			section.Machine.AgentScratchFree.Kind = "agentscratch-free"
			section.Machine.AgentScratchTotal.Kind = "agentscratch-total"
			data, err := json.Marshal(struct {
				Resources *WeeklyResourceSection `json:"resources"`
			}{&section})
			if err != nil {
				t.Fatal(err)
			}
			var saved struct {
				Resources WeeklyResourceSection `json:"resources"`
			}
			if err := json.Unmarshal(data, &saved); err != nil {
				t.Fatal(err)
			}
			if err := validatePersistedWeeklyResources(data, &saved.Resources); err != nil {
				t.Fatal(err)
			}
			third := compareResourceMachine(ResourceMachine{
				BytesWritten: DeviceGauge{Device: "disk0", Epoch: "boot-2", Value: 30, Observed: true},
				SSDWear:      DeviceGauge{Device: "disk0", Value: 9, Observed: true},
			}, previousResourceResult{State: "complete", Section: &saved.Resources}, .25)
			if third.BytesWritten.State != "compared" || !sameInt64(third.BytesWritten.Delta, ptrInt64(20)) || third.SSDWear.State != "compared" || !sameInt64(third.SSDWear.Delta, ptrInt64(1)) {
				t.Fatalf("next week did not compare with valid observations: %+v", third)
			}
		})
	}
}

func TestUnknownPreviousWeekKeepsCurrentProbeFailure(t *testing.T) {
	got := compareResourceMachine(ResourceMachine{
		BytesWritten: DeviceGauge{Device: "disk0", State: "unknown", Reason: "ioreg unavailable"},
		SSDWear:      DeviceGauge{Device: "disk0", State: "unknown", Reason: "smartctl unavailable"},
	}, previousResourceResult{State: "unknown", Reason: "prior snapshot invalid"}, .25)
	if got.BytesWritten.Reason != "ioreg unavailable" || got.SSDWear.Reason != "smartctl unavailable" || got.BytesWritten.Observed || got.SSDWear.Observed {
		t.Fatalf("current probe failures were obscured: %+v", got)
	}
}

func (f *fakeResourceRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, key)
	result, ok := f.results[key]
	if !ok {
		return nil, errors.New("unmodeled command: " + key)
	}
	return []byte(result.output), result.err
}

func TestCollectResourceMachineUsesReadOnlyProbes(t *testing.T) {
	cache := t.TempDir()
	if err := os.WriteFile(filepath.Join(cache, "one"), []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	runner := &fakeResourceRunner{results: map[string]struct {
		output string
		err    error
	}{
		"go env GOCACHE": {output: cache + "\n"},
		"ioreg -r -c IOBlockStorageDriver -w 0 -l -d 2": {output: `+-o IOBlockStorageDriver <class IOBlockStorageDriver, id 0x1>
  "Statistics" = {"Bytes (Write)"=12}
  +-o APPLE SSD Media <class IOMedia>
    "BSD Name" = "disk0"
    "Removable" = No`},
		"sysctl -n kern.boottime": {output: "{ sec = 1234, usec = 0 }"},
		"smartctl -a /dev/disk0":  {output: "Percentage Used: 8%\nRead error log failed", err: errors.New("exit status 4")},
	}}
	got := collectResourceMachine(context.Background(), resourceMachineDeps{
		Runner: runner,
		Capacity: func(path string) (uint64, uint64, error) {
			if path != "/Volumes/AgentScratch" {
				t.Fatalf("unexpected capacity path %q", path)
			}
			return 200, 100, nil
		},
	}, "/Volumes/AgentScratch", "disk0")
	if !got.GoCache.Complete || got.GoCache.Bytes != 3 || !got.AgentScratchFree.Complete || got.AgentScratchFree.Bytes != 100 || got.AgentScratchTotal.Bytes != 200 {
		t.Fatalf("cache and scratch = %+v", got)
	}
	if got.BytesWritten.State != "observed" || got.BytesWritten.Value != 12 || got.BytesWritten.Epoch != "1234" || got.SSDWear.State != "observed" || got.SSDWear.Value != 8 {
		t.Fatalf("device probes = %+v", got)
	}
	if len(runner.calls) != 4 {
		t.Fatalf("expected four read-only probes, got %v", runner.calls)
	}
	runner.results["smartctl -a /dev/disk0"] = struct {
		output string
		err    error
	}{err: errors.New("permission denied")}
	got = collectResourceMachine(context.Background(), resourceMachineDeps{Runner: runner, Capacity: func(string) (uint64, uint64, error) { return 200, 100, nil }}, "/Volumes/AgentScratch", "disk0")
	if got.SSDWear.State != "unknown" || !strings.Contains(got.SSDWear.Reason, "permission denied") {
		t.Fatalf("unavailable wear reason = %+v", got.SSDWear)
	}
}

func TestCompareResourceMachinePreservesUnknowns(t *testing.T) {
	prior := previousResourceResult{State: "complete", Section: &WeeklyResourceSection{Machine: ResourceMachineTrend{
		GoCache:      ResourceComparison{ResourceObservation: ResourceObservation{Kind: "go-cache", Definition: "/cache", Bytes: 100, Complete: true}},
		BytesWritten: DeviceGauge{Device: "disk0", Epoch: "boot-1", Value: 100, Observed: true, State: "baseline"},
		SSDWear:      DeviceGauge{Device: "disk0", Value: 8, Observed: true, State: "baseline"},
	}}}
	current := ResourceMachine{
		GoCache:      ResourceObservation{Kind: "go-cache", Definition: "/cache", Bytes: 130, Complete: true},
		BytesWritten: DeviceGauge{Device: "disk0", Epoch: "boot-2", Value: 20, Observed: true, State: "observed"},
		SSDWear:      DeviceGauge{Device: "disk0", Value: 9, Observed: true, State: "observed"},
	}
	got := compareResourceMachine(current, prior, .25)
	if !got.GoCache.Alert || !sameInt64(got.GoCache.DeltaBytes, ptrInt64(30)) {
		t.Fatalf("cache growth = %+v", got.GoCache)
	}
	if got.BytesWritten.State != "unknown" || got.BytesWritten.Delta != nil {
		t.Fatalf("reboot must suppress write delta: %+v", got.BytesWritten)
	}
	if got.SSDWear.State != "compared" || !sameInt64(got.SSDWear.Delta, ptrInt64(1)) {
		t.Fatalf("wear compares across reboot on same disk: %+v", got.SSDWear)
	}
	current.SSDWear = DeviceGauge{Device: "disk0", State: "unknown", Reason: "permission denied"}
	got = compareResourceMachine(current, prior, .25)
	if got.SSDWear.State != "unknown" || !strings.Contains(got.SSDWear.Reason, "permission denied") {
		t.Fatalf("missing wear must retain reason: %+v", got.SSDWear)
	}
}
