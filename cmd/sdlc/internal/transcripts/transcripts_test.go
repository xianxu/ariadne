package transcripts

import (
	"context"
	"reflect"
	"testing"
)

// fakeHarness lets the pure Select aggregator be tested without touching disk.
type fakeHarness struct {
	name string
	src  Sources
}

func (f fakeHarness) Name() string                                                { return f.name }
func (f fakeHarness) Sources(ctx context.Context, cwds []string) (Sources, error) { return f.src, nil }

// Select merges every harness's contribution into one Sources, deduping while
// preserving first-seen order (so the engine's dir/file lists are stable).
func TestSelectMergesAndDedups(t *testing.T) {
	hs := []Harness{
		fakeHarness{"a", Sources{Dirs: []string{"/d1"}, Files: []string{"/f1"}}},
		fakeHarness{"b", Sources{Dirs: []string{"/d1", "/d2"}, Files: []string{"/f2"}}},
	}
	got := requireSelectedSources(t, []string{"/repo"}, hs)
	want := Sources{Dirs: []string{"/d1", "/d2"}, Files: []string{"/f1", "/f2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Select = %+v, want %+v", got, want)
	}
}

// An empty harness slice yields a zero-value Sources, not a panic.
func TestSelectNoHarnesses(t *testing.T) {
	got := requireSelectedSources(t, []string{"/repo"}, nil)
	if len(got.Dirs) != 0 || len(got.Files) != 0 {
		t.Fatalf("Select with no harnesses = %+v, want empty", got)
	}
}

func requireSelectedSources(t *testing.T, cwds []string, hs []Harness) Sources {
	t.Helper()
	sources, err := Select(context.Background(), cwds, hs)
	if err != nil {
		t.Fatal(err)
	}
	return sources
}
func requireHarnessSources(t *testing.T, h Harness, cwds []string) Sources {
	t.Helper()
	sources, err := h.Sources(context.Background(), cwds)
	if err != nil {
		t.Fatal(err)
	}
	return sources
}

func TestHarnessSelectionReturnsCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, h := range []Harness{ClaudeHarness(t.TempDir()), CodexHarness(t.TempDir()), QoderHarness(t.TempDir())} {
		if _, err := h.Sources(ctx, []string{"/repo"}); err != context.Canceled {
			t.Fatalf("%s Sources error=%v", h.Name(), err)
		}
		if _, err := Select(ctx, []string{"/repo"}, []Harness{h}); err != context.Canceled {
			t.Fatalf("%s Select error=%v", h.Name(), err)
		}
	}
}
