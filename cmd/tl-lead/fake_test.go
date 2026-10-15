package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fake's refusals match couch's help (`couch --skill`, Session-sync files):
// a held lock, a stale hash, a missing hash, and a fold of an ID not owed.
func TestFakeSessionSyncRefusals(t *testing.T) {
	body := filepath.Join(t.TempDir(), "b.md")
	if err := os.WriteFile(body, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		setup func(f *fakeSessionSync) (hash string, fold []string)
		check func(error) bool
	}{
		{"lock", func(f *fakeSessionSync) (string, []string) { f.holder = "operator"; return fakeHash("old"), nil },
			func(err error) bool { return errors.Is(err, ErrLocked) }},
		{"stale hash", func(f *fakeSessionSync) (string, []string) { return fakeHash("other"), nil },
			func(err error) bool { return errors.Is(err, ErrStale) }},
		{"missing hash", func(f *fakeSessionSync) (string, []string) { return "", nil },
			func(err error) bool { return err != nil && strings.Contains(err.Error(), "--if-hash") }},
		{"fold not owed", func(f *fakeSessionSync) (string, []string) { return fakeHash("old"), []string{"nope"} },
			func(err error) bool { return err != nil && strings.Contains(err.Error(), "not-owed") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFake(map[string]string{"S": "old"})
			hash, fold := tc.setup(f)
			err := f.WriteSection("file", "S", body, hash, fold)
			if !tc.check(err) {
				t.Errorf("err = %v", err)
			}
			if f.writes != 0 || f.sections["S"] != "old" {
				t.Errorf("a refused write changed state: writes=%d S=%q", f.writes, f.sections["S"])
			}
		})
	}
	f := newFake(map[string]string{"S": "old"})
	f.owed = []string{"ev"}
	if err := f.WriteSection("file", "S", body, fakeHash("old"), []string{"ev"}); err != nil {
		t.Fatal(err)
	}
	if f.sections["S"] != "new" || len(f.owed) != 0 {
		t.Errorf("an accepted write: S=%q owed=%v", f.sections["S"], f.owed)
	}
}
