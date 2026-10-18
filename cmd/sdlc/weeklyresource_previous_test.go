package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadPreviousResourceSectionExactWeek(t *testing.T) {
	root := t.TempDir()
	end := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	window, err := newWeeklyWindow(end, end)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "memory/work/devinfra/test-economy/weekly")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	priorPath := filepath.Join(dir, "2026-10-12.json")
	if got := readPreviousResourceSection(root, window); got.State != "baseline" || !strings.Contains(got.Reason, "missing") {
		t.Fatalf("missing exact prior = %+v", got)
	}
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(priorPath, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"schema_version":1,"window":{"start":"2026-10-05T00:00:00Z","end":"2026-10-12T00:00:00Z"},"repos":[]}`)
	if got := readPreviousResourceSection(root, window); got.State != "baseline" || !strings.Contains(got.Reason, "v1") {
		t.Fatalf("v1 has no resource baseline: %+v", got)
	}
	write(`{"schema_version":2,"window":{"start":"2026-10-04T00:00:00Z","end":"2026-10-11T00:00:00Z"},"resources":{"families":[]}}`)
	if got := readPreviousResourceSection(root, window); got.State != "unknown" || !strings.Contains(got.Reason, "window") {
		t.Fatalf("mismatched window = %+v", got)
	}
	write(`{"schema_version":2,"window":{"start":"2026-10-05T00:00:00Z","end":"2026-10-12T00:00:00Z"},"resources":{"observed_at":"2026-10-12T01:00:00Z","families":[{"kind":"couch-ledgers","definition":"v1","bytes":100,"complete":true}]}}`)
	got := readPreviousResourceSection(root, window)
	if got.State != "unknown" || !strings.Contains(got.Reason, "invalid") {
		t.Fatalf("v2 predecessor without machine coverage = %+v", got)
	}
	write(`{"schema_version":2,"window":{"start":"2026-10-05T00:00:00Z","end":"2026-10-12T00:00:00Z"},"resources":{"families":[{"kind":"couch-ledgers","definition":"v1","bytes":100,"complete":true}],"suites":[],"machine":{"go_cache":{"kind":"go-cache","complete":true},"agentscratch_free":{"kind":"agentscratch-free","complete":true},"agentscratch_total":{"kind":"agentscratch-total","complete":true},"bytes_written":{"device":"disk0","state":"observed"},"ssd_wear":{"device":"disk0","state":"observed"}}}}`)
	if got := readPreviousResourceSection(root, window); got.State != "unknown" || !strings.Contains(got.Reason, "observed_at") {
		t.Fatalf("missing point-observation time = %+v", got)
	}
	validBody := `{"schema_version":2,"window":{"start":"2026-10-05T00:00:00Z","end":"2026-10-12T00:00:00Z"},"resources":{"observed_at":"2026-10-12T01:00:00Z","families":[{"kind":"couch-ledgers","definition":"v1","bytes":100,"files":1,"complete":true,"state":"baseline"}],"largest_ledgers":[],"suites":[],"machine":{"go_cache":{"kind":"go-cache","definition":"/cache","bytes":0,"files":0,"complete":true,"state":"baseline"},"agentscratch_free":{"kind":"agentscratch-free","definition":"/scratch","bytes":0,"files":0,"complete":true,"state":"baseline"},"agentscratch_total":{"kind":"agentscratch-total","definition":"/scratch","bytes":0,"files":0,"complete":true,"state":"baseline"},"bytes_written":{"device":"disk0","value":0,"observed":true,"state":"baseline"},"ssd_wear":{"device":"disk0","value":0,"observed":true,"state":"baseline"}}}}`
	write(validBody)
	got = readPreviousResourceSection(root, window)
	if got.State != "complete" || got.Section == nil || len(got.Section.Families) != 1 || got.Section.Families[0].Bytes != 100 {
		t.Fatalf("valid predecessor = %+v", got)
	}
	for _, tc := range []struct{ body, want string }{
		{strings.Replace(validBody, `"bytes":100,`, "", 1), "bytes"},
		{strings.Replace(validBody, `"files":1,`, "", 1), "files"},
		{strings.Replace(validBody, `"value":0,`, "", 1), "value"},
		{strings.Replace(validBody, `"observed":true,`, "", 1), "observed"},
		{strings.Replace(validBody, `"suites":[]`, `"suites":[{"repo":"ariadne","wall":{"p50_seconds":1,"p90_seconds":2},"state":"baseline"}]`, 1), "samples"},
	} {
		write(tc.body)
		if got := readPreviousResourceSection(root, window); got.State != "unknown" || !strings.Contains(got.Reason, tc.want) {
			t.Fatalf("missing persisted %s accepted: %+v", tc.want, got)
		}
	}
	write(`{`)
	if got := readPreviousResourceSection(root, window); got.State != "unknown" || !strings.Contains(got.Reason, "decode") {
		t.Fatalf("malformed predecessor = %+v", got)
	}
}

func TestReadPreviousResourceSectionRejectsSemanticCorruption(t *testing.T) {
	root := t.TempDir()
	end := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	window, err := newWeeklyWindow(end, end)
	if err != nil {
		t.Fatal(err)
	}
	priorWindow, err := newWeeklyWindow(end.Add(-7*24*time.Hour), end)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "memory/work/devinfra/test-economy/weekly")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "2026-10-12.json")
	valid := func() WeeklySnapshot {
		obs := func(kind string) ResourceComparison {
			return ResourceComparison{ResourceObservation: ResourceObservation{Kind: kind, Definition: "/source", Bytes: 100, Files: 1, Complete: true}, State: "baseline"}
		}
		return WeeklySnapshot{SchemaVersion: 2, Window: priorWindow, Resources: &WeeklyResourceSection{
			ObservedAt:     end.Add(-7 * 24 * time.Hour).Add(time.Hour),
			Families:       []ResourceComparison{obs("family")},
			LargestLedgers: []ResourceFileComparison{{ResourceFile: ResourceFile{Path: "/ledger", Bytes: 100}, State: "baseline"}},
			Suites:         []WeeklyResourceSuite{{Repo: "ariadne", Wall: &SuiteQuantiles{P50: 10, P90: 20, Samples: 2}, State: "baseline"}},
			Machine:        ResourceMachineTrend{GoCache: obs("go-cache"), AgentScratchFree: obs("agentscratch-free"), AgentScratchTotal: obs("agentscratch-total"), BytesWritten: DeviceGauge{Device: "disk0", Epoch: "123", Value: 100, Observed: true, State: "baseline"}, SSDWear: DeviceGauge{Device: "disk0", Value: 8, Observed: true, State: "baseline"}},
		}}
	}
	for _, tc := range []struct {
		name, want string
		change     func(*WeeklySnapshot)
	}{
		{"negative family bytes", "family", func(s *WeeklySnapshot) { s.Resources.Families[0].Bytes = -100 }},
		{"negative file count", "family", func(s *WeeklySnapshot) { s.Resources.Families[0].Files = -1 }},
		{"missing family definition", "family", func(s *WeeklySnapshot) { s.Resources.Families[0].Definition = "" }},
		{"duplicate family", "family", func(s *WeeklySnapshot) { s.Resources.Families = append(s.Resources.Families, s.Resources.Families[0]) }},
		{"negative ledger", "ledger", func(s *WeeklySnapshot) { s.Resources.LargestLedgers[0].Bytes = -1 }},
		{"missing ledger path", "ledger", func(s *WeeklySnapshot) { s.Resources.LargestLedgers[0].Path = "" }},
		{"duplicate ledger", "ledger", func(s *WeeklySnapshot) {
			s.Resources.LargestLedgers = append(s.Resources.LargestLedgers, s.Resources.LargestLedgers[0])
		}},
		{"missing suite repo", "suite", func(s *WeeklySnapshot) { s.Resources.Suites[0].Repo = "" }},
		{"duplicate suite", "suite", func(s *WeeklySnapshot) { s.Resources.Suites = append(s.Resources.Suites, s.Resources.Suites[0]) }},
		{"invalid wall p50", "suite", func(s *WeeklySnapshot) { s.Resources.Suites[0].Wall.P50 = -1 }},
		{"invalid wall order", "suite", func(s *WeeklySnapshot) { s.Resources.Suites[0].Wall.P90 = 1 }},
		{"invalid wall samples", "suite", func(s *WeeklySnapshot) { s.Resources.Suites[0].Wall.Samples = 0 }},
		{"negative machine bytes", "go-cache", func(s *WeeklySnapshot) { s.Resources.Machine.GoCache.Bytes = -1 }},
		{"unpaired family delta", "family", func(s *WeeklySnapshot) { s.Resources.Families[0].DeltaBytes = ptrInt64(1) }},
		{"missing machine kind", "machine", func(s *WeeklySnapshot) { s.Resources.Machine.GoCache.Kind = "" }},
		{"missing device", "device", func(s *WeeklySnapshot) { s.Resources.Machine.BytesWritten.Device = "" }},
		{"invalid device state", "device", func(s *WeeklySnapshot) { s.Resources.Machine.SSDWear.State = "invalid" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snap := valid()
			tc.change(&snap)
			data, err := json.Marshal(snap)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatal(err)
			}
			got := readPreviousResourceSection(root, window)
			if got.State != "unknown" || !strings.Contains(got.Reason, tc.want) {
				t.Fatalf("%s: %+v", tc.name, got)
			}
		})
	}
}
