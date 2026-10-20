package scratchdisk

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestNormalTempDirPreservesUsableInheritedRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TMPDIR", root)
	t.Setenv("GOTMPDIR", filepath.Join(root, "untouched"))
	got, err := NormalTempDir()
	if err != nil || got != root {
		t.Fatalf("got %q, %v", got, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("probe not cleaned: %v %v", entries, err)
	}
	if os.Getenv("TMPDIR") != root || os.Getenv("GOTMPDIR") != filepath.Join(root, "untouched") {
		t.Fatal("changed global environment")
	}
}

func TestNormalTempSelectionRejectsFileAndUsesPrivateFolder(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "file")
	if err := os.WriteFile(bad, []byte("owned fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(root, "normal")
	if err := os.Mkdir(good, 0o700); err != nil {
		t.Fatal(err)
	}
	got, err := normalTempDir(bad, func() (string, error) { return good, nil }, "unused", CheckTempDir)
	if err != nil || got != good {
		t.Fatalf("got %q, %v", got, err)
	}
	entries, err := os.ReadDir(good)
	if err != nil || len(entries) != 0 {
		t.Fatalf("probe not cleaned: %v %v", entries, err)
	}
}

func TestNormalTempSelectionCandidates(t *testing.T) {
	denied := errors.New("fixture permission denied")
	for _, tc := range []struct {
		name, inherited, user, fallback, good, want string
		userErr                                     error
		probes                                      []string
		lookups                                     int
	}{
		{name: "current succeeds lazily", inherited: "current", user: "user", fallback: "default", good: "current", want: "current", probes: []string{"current"}},
		{name: "user wins", inherited: "current", user: "user", fallback: "default", good: "user", want: "user", probes: []string{"current", "user"}, lookups: 1},
		{name: "Unix default after lookup error", inherited: "current", userErr: denied, fallback: "default", good: "default", want: "default", probes: []string{"current", "default"}, lookups: 1},
		{name: "duplicate candidates tried once", inherited: "current", user: "current/.", fallback: "current", probes: []string{"current"}, lookups: 1},
		{name: "empty user ignored", inherited: "current", fallback: "default", probes: []string{"current", "default"}, lookups: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Candidate controls use absolute paths; relative behavior has its own fixtures.
			absolute := func(path string) string {
				if path == "" {
					return ""
				}
				out, err := filepath.Abs(path)
				if err != nil {
					t.Fatal(err)
				}
				return out
			}
			tc.inherited, tc.user, tc.fallback = absolute(tc.inherited), absolute(tc.user), absolute(tc.fallback)
			tc.good, tc.want = absolute(tc.good), absolute(tc.want)
			for i := range tc.probes {
				tc.probes[i] = absolute(tc.probes[i])
			}
			var probes []string
			lookups := 0
			got, err := normalTempDir(tc.inherited, func() (string, error) { lookups++; return tc.user, tc.userErr }, tc.fallback, func(path string) error {
				probes = append(probes, path)
				if path == tc.good {
					return nil
				}
				return denied
			})
			if got != tc.want || (err == nil) != (tc.want != "") {
				t.Fatalf("got %q, %v", got, err)
			}
			if err != nil && (!errors.Is(err, denied) || !strings.Contains(err.Error(), "normal temp")) {
				t.Fatalf("hidden failure: %v", err)
			}
			if !reflect.DeepEqual(probes, tc.probes) || lookups != tc.lookups {
				t.Fatalf("probes %q, lookups %d", probes, lookups)
			}
		})
	}
}

func TestMkdirTempRepairsOnlyDefaultFallback(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "file")
	if err := os.WriteFile(bad, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", bad)
	t.Setenv("WF_TEST_SCRATCH", "off")
	calls := 0
	selectNormal := func() (string, error) { calls++; return root, nil }
	dir, err := mkdirTemp("sub", os.TempDir(), "x-*", selectNormal)
	if err != nil || filepath.Dir(dir) != root || calls != 1 {
		t.Fatalf("default: %q %v calls=%d", dir, err, calls)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	calls = 0
	_, err = mkdirTemp("sub", filepath.Join(bad, "custom"), "x-*", selectNormal)
	if err == nil || calls != 0 {
		t.Fatalf("custom: %v calls=%d", err, calls)
	}
	calls = 0
	dir, err = mkdirTemp("sub", root, "x-*", selectNormal)
	if err != nil || calls != 0 {
		t.Fatalf("usable fallback: %v calls=%d", err, calls)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
}

func TestMkdirTempReportsNormalSelectionFailure(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "file")
	if err := os.WriteFile(bad, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", bad)
	t.Setenv("WF_TEST_SCRATCH", "off")
	denied := errors.New("no usable normal root")
	_, err := mkdirTemp("sub", os.TempDir(), "x-*", func() (string, error) { return "", denied })
	if !errors.Is(err, denied) {
		t.Fatalf("selection failure hidden: %v", err)
	}
}

func TestUserTempDirPlatformLookup(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\n[ \"$1\" = DARWIN_USER_TEMP_DIR ] || exit 1\nprintf '%s/\\n' \"$SDLC_USER_TEMP_FIXTURE\"\n"
	if err := os.WriteFile(filepath.Join(bin, "getconf"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("SDLC_USER_TEMP_FIXTURE", root)
	t.Setenv("TMPDIR", filepath.Join(root, "bad-inherited"))
	got, err := UserTempDir()
	want := ""
	if runtime.GOOS == "darwin" {
		want = root
	}
	if got != want || err != nil {
		t.Fatalf("got %q %v; want %q", got, err, want)
	}
	if runtime.GOOS != "darwin" {
		return
	}
	bad := filepath.Join(root, "file")
	if err := os.WriteFile(bad, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", bad)
	got, err = NormalTempDir()
	if got != root || err != nil {
		t.Fatalf("normal fallback: %q %v", got, err)
	}
	t.Setenv("WF_TEST_SCRATCH", "off")
	dir, err := MkdirTemp("sub", os.TempDir(), "owned-*")
	if err != nil || filepath.Dir(dir) != root {
		t.Fatalf("checkout fallback: %q %v", dir, err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if os.Getenv("TMPDIR") != bad {
		t.Fatal("changed inherited TMPDIR")
	}
}

func TestNormalTempDirRelativeInheritedRootSurvivesChildCWD(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.Mkdir("normal", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("checkout", 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", "normal")
	got, err := NormalTempDir()
	want := filepath.Join(root, "normal")
	if err != nil || got != want {
		t.Fatalf("got %q %v; want absolute %q", got, err, want)
	}
	t.Chdir(filepath.Join(root, "checkout"))
	dir, err := os.MkdirTemp(got, "child-*")
	if err != nil {
		t.Fatalf("temp root changed meaning in child cwd: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if filepath.Dir(dir) != want || os.Getenv("TMPDIR") != "normal" {
		t.Fatalf("dir %q, TMPDIR %q", dir, os.Getenv("TMPDIR"))
	}
}

func TestNormalTempRelativeCandidateProbeAndDedup(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	want := filepath.Join(root, "normal")
	for _, source := range []string{"inherited", "user", "default"} {
		t.Run(source, func(t *testing.T) {
			current, user, fallback := "", "", ""
			switch source {
			case "inherited":
				current = "normal"
			case "user":
				user = "normal"
			case "default":
				fallback = "normal"
			}
			var probed []string
			got, err := normalTempDir(current, func() (string, error) { return user, nil }, fallback, func(path string) error { probed = append(probed, path); return nil })
			if got != want || err != nil || !reflect.DeepEqual(probed, []string{want}) {
				t.Fatalf("got %q %v, probed %q; want %q", got, err, probed, want)
			}
		})
	}
	calls := 0
	denied := errors.New("fixture deny")
	_, err := normalTempDir("normal", func() (string, error) { return want, nil }, "normal/.", func(path string) error {
		calls++
		if path != want {
			t.Fatalf("relative probe %q", path)
		}
		return denied
	})
	if calls != 1 || !errors.Is(err, denied) {
		t.Fatalf("calls %d, error %v", calls, err)
	}
}
