package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeResourceRunner struct {
	results map[string]struct {
		output string
		err    error
	}
	calls []string
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
		BytesWritten: DeviceGauge{Device: "disk0", Epoch: "boot-1", Value: 100, State: "observed"},
		SSDWear:      DeviceGauge{Device: "disk0", Value: 8, State: "observed"},
	}}}
	current := ResourceMachine{
		GoCache:      ResourceObservation{Kind: "go-cache", Definition: "/cache", Bytes: 130, Complete: true},
		BytesWritten: DeviceGauge{Device: "disk0", Epoch: "boot-2", Value: 20, State: "observed"},
		SSDWear:      DeviceGauge{Device: "disk0", Value: 9, State: "observed"},
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
