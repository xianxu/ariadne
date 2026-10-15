package transcripts

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func qoderFixture(t *testing.T, root, data string) string {
	t.Helper()
	path := filepath.Join(root, "misleading-project-slug", "copied-name.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func qoderRows(cwd string) string {
	return fmt.Sprintf(`{"type":"workspace-directories","sessionId":"S","directories":[%q]}
{"type":"runtime-config","sessionId":"S","timestamp":1791936000000,"model":"qoder-test"}
{"type":"user","sessionId":"S","timestamp":"2026-10-14T00:00:00Z","parentUuid":"message-parent"}
{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T00:01:00Z","parentUuid":"message-parent","message":{"id":"response","stop_reason":"end_turn","usage":{"credits":1.25}}}
`, cwd)
}
func TestQoderDiscoveryRecordedWorkspaceAndPinnedPrefix(t *testing.T) {
	root := t.TempDir()
	data := qoderRows("/repo")
	path := qoderFixture(t, root, data)
	h := QoderHarness(root)
	if got := Select(context.Background(), []string{"/repo"}, []Harness{h}); len(got.Files) != 1 || got.Files[0] != path {
		t.Fatalf("sources=%+v", got)
	}
	if got := Select(context.Background(), []string{"/misleading/project/slug"}, []Harness{h}); len(got.Files) != 0 {
		t.Fatalf("slug inferred cwd: %+v", got)
	}
	got, warnings, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{h})
	if err != nil || len(warnings) != 0 || len(got) != 1 {
		t.Fatalf("evidence=%+v warnings=%v err=%v", got, warnings, err)
	}
	s := got[0]
	if s.Harness != "qoder" || s.ID != "S" || s.CWD != "/repo" || s.ParentID != "" || s.Spawned || s.Started.Format(time.RFC3339) != "2026-10-14T00:00:00Z" || s.Length != int64(len(data)) || s.Digest != fmt.Sprintf("%x", sha256.Sum256([]byte(data))) {
		t.Fatalf("evidence=%+v", s)
	}
	usage, err := DiscoverUsageEvidence(context.Background(), []Harness{h}, time.Time{})
	if err != nil || len(usage.Sources) != 1 || len(usage.Sessions) != 1 || usage.Sources[0].ID != "S" || !usage.Sources[0].MetadataComplete {
		t.Fatalf("usage=%+v err=%v", usage, err)
	}
	// Repeated immutable metadata and stale runtime observations must not move start.
	data += `{"type":"runtime-config","sessionId":"S","timestamp":1000,"model":"old"}` + "\n" + strings.Split(data, "\n")[0] + "\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	later, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{h})
	if err != nil || len(later) != 1 || !later[0].Started.Equal(s.Started) || later[0].Length <= s.Length {
		t.Fatalf("later=%+v err=%v", later, err)
	}
}
func TestQoderDiscoveryRejectsAmbiguousWorkspaceRetainsUsage(t *testing.T) {
	for _, extra := range []string{
		`{"type":"workspace-directories","sessionId":"S","directories":["/other"]}`,
		`{"type":"workspace-directories","sessionId":"S","directories":["/repo","/other"]}`,
	} {
		t.Run(extra, func(t *testing.T) {
			root := t.TempDir()
			qoderFixture(t, root, qoderRows("/repo")+extra+"\n")
			hs := []Harness{QoderHarness(root)}
			got, warnings, err := DiscoverEvidence(context.Background(), []string{"/repo"}, hs)
			if len(got) != 0 || (err == nil && len(warnings) == 0) {
				t.Fatalf("ambiguous admitted: %+v %v %v", got, warnings, err)
			}
			usage, err := DiscoverUsageEvidence(context.Background(), hs, time.Time{})
			if err != nil || len(usage.Sources) != 1 || len(usage.Sessions) != 0 || len(usage.UnknownFiles) != 1 {
				t.Fatalf("unknown usage=%+v err=%v", usage, err)
			}
		})
	}
}
func TestQoderDiscoveryScansBeyondLegacyPrefixAndRejectsLatePoison(t *testing.T) {
	root := t.TempDir()
	data := qoderRows("/repo") + fmt.Sprintf(`{"type":"progress","content":%q}`, strings.Repeat("x", metadataLimit+1024)) + "\n"
	path := qoderFixture(t, root, data)
	hs := []Harness{QoderHarness(root)}
	if got, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, hs); err != nil || len(got) != 1 {
		t.Fatalf("large valid prefix: %+v %v", got, err)
	}
	data += `{"type":"user","sessionId":"copied-session","timestamp":"2026-10-14T00:02:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	if got, warnings, err := DiscoverEvidence(context.Background(), []string{"/repo"}, hs); len(got) != 0 || (err == nil && len(warnings) == 0) {
		t.Fatalf("late poison accepted: %+v %v %v", got, warnings, err)
	}
	usage, err := DiscoverUsageEvidence(context.Background(), hs, time.Time{})
	if err != nil || len(usage.Sources) != 1 || usage.Sources[0].MetadataComplete || len(usage.UnknownFiles) != 1 {
		t.Fatalf("poison usage=%+v err=%v", usage, err)
	}
}
func TestQoderDiscoveryCancellationAndUntimedMetadata(t *testing.T) {
	root := t.TempDir()
	qoderFixture(t, root, strings.Split(qoderRows("/repo"), "\n")[0]+"\n")
	hs := []Harness{QoderHarness(root)}
	if got, _, _ := DiscoverEvidence(context.Background(), []string{"/repo"}, hs); len(got) != 0 {
		t.Fatalf("untimed metadata admitted: %+v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := DiscoverEvidence(ctx, []string{"/repo"}, hs); err == nil {
		t.Fatal("actual ignored cancellation")
	}
	if _, err := DiscoverUsageEvidence(ctx, hs, time.Time{}); err == nil {
		t.Fatal("usage ignored cancellation")
	}
}

func TestQoderMetadataPinnedAppendTruncateAndReplacement(t *testing.T) {
	root := t.TempDir()
	data := qoderRows("/repo")
	path := qoderFixture(t, root, data)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// The usage discovery sample bounds proof; later contradictory appends belong
	// to a future sample rather than changing the already sampled prefix.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(`{"type":"workspace-directories","sessionId":"other","directories":["/other"]}` + "\n")
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	s, err := readQoderEvidence(context.Background(), path, info.Size(), info)
	if err != nil || s.ID != "S" || s.Digest != fmt.Sprintf("%x", sha256.Sum256([]byte(data))) {
		t.Fatalf("pinned append=%+v %v", s, err)
	}
	if err := os.Truncate(path, info.Size()-10); err != nil {
		t.Fatal(err)
	}
	if _, err := readQoderEvidence(context.Background(), path, info.Size(), info); err == nil {
		t.Fatal("truncation accepted")
	}
	replacement := filepath.Join(filepath.Dir(path), "replacement")
	if err := os.WriteFile(replacement, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readQoderEvidence(context.Background(), path, info.Size(), info); err == nil {
		t.Fatal("replacement accepted")
	}
}
func TestQoderDiscoveryMalformedAndMissingWorkspaceStayVisible(t *testing.T) {
	for _, data := range []string{
		strings.Join(strings.Split(qoderRows("/repo"), "\n")[1:], "\n"),
		qoderRows("/repo") + `{"type":"user","sessionId":"S","timestamp":"PRIVATE-CONTENT"}` + "\n",
		`{"type":"runtime-config","timestamp":1000,"model":"copied"}` + "\n" + qoderRows("/repo"),
	} {
		t.Run(fmt.Sprint(len(data)), func(t *testing.T) {
			root := t.TempDir()
			qoderFixture(t, root, data)
			hs := []Harness{QoderHarness(root)}
			actual, warnings, err := DiscoverEvidence(context.Background(), []string{"/repo"}, hs)
			if len(actual) != 0 || (len(warnings) == 0 && err == nil) {
				t.Fatalf("incomplete proof invisible: %+v %v %v", actual, warnings, err)
			}
			for _, warning := range warnings {
				if strings.Contains(warning, "PRIVATE-CONTENT") {
					t.Fatal("raw content leaked")
				}
			}
			usage, err := DiscoverUsageEvidence(context.Background(), hs, time.Time{})
			if err != nil || len(usage.Sources) != 1 || len(usage.UnknownFiles) != 1 || len(usage.Sessions) != 0 {
				t.Fatalf("unknown=%+v err=%v", usage, err)
			}
		})
	}
}

func TestQoderUsageMetadataIncompletePreservesPinnedLength(t *testing.T) {
	root := t.TempDir()
	path := qoderFixture(t, root, qoderRows("/repo"))
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	c := &usageCandidate{UsageSource: UsageSource{SessionEvidence: SessionEvidence{Harness: "qoder", File: path, Length: info.Size()}, Info: info}}
	if err := os.Truncate(path, 0); err != nil {
		t.Fatal(err)
	}
	if err := readUsageMetadata(context.Background(), c); err == nil {
		t.Fatal("truncation accepted")
	}
	if c.Length != info.Size() {
		t.Fatalf("metadata failure lost prefix bound: got %d want %d", c.Length, info.Size())
	}
}

func TestQoderActualDiscoveryCandidateBoundIsVisible(t *testing.T) {
	root := t.TempDir()
	path := qoderFixture(t, root, qoderRows("/repo"))
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "second.jsonl"), []byte(qoderRows("/repo")), 0600); err != nil {
		t.Fatal(err)
	}
	got, _, err := discoverEvidence(context.Background(), []string{"/repo"}, []Harness{QoderHarness(root)}, 1)
	if err == nil || !strings.Contains(err.Error(), "qoder candidate limit") || len(got) != 0 {
		t.Fatalf("candidate bound accepted partial proof: %+v %v", got, err)
	}
}

func TestQoderActualDiscoveryRefusesPartialOwnedEvidence(t *testing.T) {
	for _, bad := range []string{
		qoderRows("/repo") + "not-json\n",
		qoderRows("/repo") + `{"type":"workspace-directories","sessionId":"S","directories":["/other"]}` + "\n",
		qoderRows("/other") + `{"type":"workspace-directories","sessionId":"S","directories":["/repo"]}` + "\n",
	} {
		t.Run(fmt.Sprint(len(bad)), func(t *testing.T) {
			root := t.TempDir()
			good := strings.ReplaceAll(qoderRows("/repo"), `"S"`, `"good"`)
			path := qoderFixture(t, root, good)
			if err := os.WriteFile(filepath.Join(filepath.Dir(path), "bad-owned.jsonl"), []byte(bad), 0600); err != nil {
				t.Fatal(err)
			}
			got, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{QoderHarness(root)})
			if err == nil || len(got) != 0 {
				t.Fatalf("partial known-owner actual admitted: %+v err=%v", got, err)
			}
		})
	}
}
func TestQoderActualDiscoveryUnrelatedMalformedSourceDoesNotBlock(t *testing.T) {
	root := t.TempDir()
	path := qoderFixture(t, root, qoderRows("/repo"))
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "bad-unrelated.jsonl"), []byte(strings.ReplaceAll(qoderRows("/other"), `"S"`, `"other"`)+"not-json\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, warnings, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{QoderHarness(root)})
	if err != nil || len(got) != 1 || len(warnings) != 1 {
		t.Fatalf("unrelated source blocks actual: %+v %v %v", got, warnings, err)
	}
}

func TestQoderSourcesReceivesCallerCancellation(t *testing.T) {
	root := t.TempDir()
	qoderFixture(t, root, qoderRows("/repo"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := QoderHarness(root).Sources(ctx, []string{"/repo"}); len(got.Files) != 0 {
		t.Fatalf("canceled Qoder selection returned sources: %+v", got)
	}
	if got := Select(ctx, []string{"/repo"}, []Harness{QoderHarness(root)}); len(got.Files) != 0 {
		t.Fatalf("Select lost caller cancellation: %+v", got)
	}
}
