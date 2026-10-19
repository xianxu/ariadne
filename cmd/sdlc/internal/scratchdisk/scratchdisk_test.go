package scratchdisk

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChoose(t *testing.T) {
	for _, tc := range []struct {
		name     string
		p        Probe
		use      bool
		noteWant string
	}{
		{"off", Probe{}, false, ""},
		{"unmounted is the normal path, silently", Probe{Root: DefaultRoot}, false, ""},
		{"probe error", Probe{Root: "/x", Mounted: true, Err: errors.New("boom")}, false, "unreadable"},
		{"below the floor", Probe{Root: "/x", Mounted: true, FreeBytes: 1 << 30}, false, "1024 MiB free"},
		{"ok", Probe{Root: "/x", Mounted: true, FreeBytes: 3 << 30}, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			use, note := Choose(tc.p, Floor)
			if use != tc.use || (tc.noteWant == "") != (note == "") || !strings.Contains(note, tc.noteWant) {
				t.Fatalf("use=%v note=%q", use, note)
			}
		})
	}
}

// An explicit root skips the mount check; "off" disables; a fixture root
// reports its free space.
func TestReadOverride(t *testing.T) {
	t.Setenv("WF_TEST_SCRATCH", "off")
	if root, _ := Use(); root != "" {
		t.Fatalf("off: root %q", root)
	}
	dir := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", dir)
	p := Read()
	if !p.Mounted || p.Err != nil || p.Root != dir || p.FreeBytes == 0 {
		t.Fatalf("fixture probe %+v", p)
	}
}

// A mounted disk with room takes new folders; off, the fallback alone does.
func TestRoots(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", dir)
	create, all, note := Roots("sub", "/fallback")
	if create != dir+"/sub" || len(all) != 2 || all[0] != dir+"/sub" || all[1] != "/fallback" || note != "" {
		t.Fatalf("mounted: %q %q %q", create, all, note)
	}
	t.Setenv("WF_TEST_SCRATCH", "off")
	create, all, _ = Roots("sub", "/fallback")
	if create != "/fallback" || len(all) != 1 {
		t.Fatalf("off: %q %q", create, all)
	}
}

// A disk that refuses the write falls back to the normal folder.
func TestMkdirTempFallsBack(t *testing.T) {
	ram, fallback := t.TempDir(), t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", ram)
	dir, err := MkdirTemp("sub", fallback, "x-*")
	if err != nil || filepath.Dir(dir) != filepath.Join(ram, "sub") {
		t.Fatalf("writable disk: %q %v", dir, err)
	}
	if err := os.RemoveAll(filepath.Join(ram, "sub")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(ram, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ram, 0o700) })
	dir, err = MkdirTemp("sub", fallback, "x-*")
	if err != nil || filepath.Dir(dir) != fallback {
		t.Fatalf("refusing disk: %q %v", dir, err)
	}
}
