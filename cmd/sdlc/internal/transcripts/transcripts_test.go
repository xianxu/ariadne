package transcripts

import (
	"reflect"
	"testing"
)

// fakeHarness lets the pure Select aggregator be tested without touching disk.
type fakeHarness struct {
	name string
	src  Sources
}

func (f fakeHarness) Name() string                  { return f.name }
func (f fakeHarness) Sources(cwds []string) Sources { return f.src }

// Select merges every harness's contribution into one Sources, deduping while
// preserving first-seen order (so the engine's dir/file lists are stable).
func TestSelectMergesAndDedups(t *testing.T) {
	hs := []Harness{
		fakeHarness{"a", Sources{Dirs: []string{"/d1"}, Files: []string{"/f1"}}},
		fakeHarness{"b", Sources{Dirs: []string{"/d1", "/d2"}, Files: []string{"/f2"}}},
	}
	got := Select([]string{"/repo"}, hs)
	want := Sources{Dirs: []string{"/d1", "/d2"}, Files: []string{"/f1", "/f2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Select = %+v, want %+v", got, want)
	}
}

// An empty harness slice yields a zero-value Sources, not a panic.
func TestSelectNoHarnesses(t *testing.T) {
	got := Select([]string{"/repo"}, nil)
	if len(got.Dirs) != 0 || len(got.Files) != 0 {
		t.Fatalf("Select with no harnesses = %+v, want empty", got)
	}
}
