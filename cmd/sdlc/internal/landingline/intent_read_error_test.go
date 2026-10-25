package landingline

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntentOpenFailureIsNotAbsence(t *testing.T) {
	// Both supported hosts reject a component longer than NAME_MAX. Unlike
	// chmod this remains a deterministic open error for privileged test users.
	path := filepath.Join(t.TempDir(), strings.Repeat("x", 300))
	if _, err := os.Open(path); err == nil || os.IsNotExist(err) {
		t.Fatalf("fixture requires a non-ENOENT open error: %v", err)
	}
	got, present := observeIntent(path)
	if !present || got.Err == nil || got.Path != path {
		t.Fatalf("open failure became absence: %+v, present=%v", got, present)
	}
	missing := filepath.Join(t.TempDir(), "gone")
	if _, present := observeIntent(missing); present {
		t.Fatal("confirmed disappearance should be omitted")
	}
}

func TestIntentPermissionFailureRetainsAnObservation(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "intents")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "denied")
	if err := os.WriteFile(path, []byte("not readable"), 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0600) })
	if f, err := os.Open(path); err == nil {
		f.Close()
		t.Skip("privileged user bypasses file permissions")
	} else if !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	got, err := OpenIntents(root).List()
	if err != nil || len(got) != 1 || !errors.Is(got[0].Err, os.ErrPermission) {
		t.Fatalf("permission error became absence: %+v %v", got, err)
	}
}

func TestIntentNonregularEntriesAreInspectionEvidence(t *testing.T) {
	root := t.TempDir()
	store := OpenIntents(root)
	dir := filepath.Join(root, "intents")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte(`{"secret":"must not be read"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "symlink")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "directory"), 0700); err != nil {
		t.Fatal(err)
	}
	got, err := store.List()
	if err != nil || len(got) != 2 {
		t.Fatalf("nonregular records vanished: %+v %v", got, err)
	}
	for _, report := range got {
		if report.Err == nil || !strings.Contains(report.Err.Error(), "regular") {
			t.Fatalf("nonregular record followed: %+v", report)
		}
	}
}
