package scratchdisk

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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
		{"below the floor", Probe{Root: "/x", Mounted: true, FileSystem: "apfs", FreeBytes: 1 << 30}, false, "1024 MiB free"},
		{"hfs falls back on Darwin", Probe{Root: "/x", Mounted: true, FileSystem: "hfs", FreeBytes: 3 << 30}, runtime.GOOS != "darwin", fileSystemNote("hfs")},
		{"unknown fails closed on Darwin", Probe{Root: "/x", Mounted: true, FreeBytes: 3 << 30}, runtime.GOOS != "darwin", fileSystemNote("unknown")},
		{"ok", Probe{Root: "/x", Mounted: true, FileSystem: "apfs", FreeBytes: 3 << 30}, true, ""},
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
	use, wantNote := Choose(Read(), Floor)
	wantCreate := "/fallback"
	if use {
		wantCreate = dir + "/sub"
	}
	create, all, note := Roots("sub", "/fallback")
	if create != wantCreate || len(all) != 2 || all[0] != dir+"/sub" || all[1] != "/fallback" || note != wantNote {
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
	want := fallback
	if use, _ := Choose(Read(), Floor); use {
		want = filepath.Join(ram, "sub")
	}
	dir, err := MkdirTemp("sub", fallback, "x-*")
	if err != nil || filepath.Dir(dir) != want {
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

// Non-Darwin override fixtures preserve their existing filesystem behavior.
func fileSystemNote(fs string) string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	return "is " + fs + "; APFS required"
}

func TestUnsupportedRootFallsBackButRemainsSweepable(t *testing.T) {
	p := Probe{Root: "/fixture", Mounted: true, FileSystem: "hfs", FreeBytes: 3 << 30}
	create, all, note := roots(p, "local-checks", "/fallback")
	wantCreate := "/fallback"
	if runtime.GOOS != "darwin" {
		wantCreate = "/fixture/local-checks"
	}
	if create != wantCreate || len(all) != 2 || all[0] != "/fixture/local-checks" || all[1] != "/fallback" || !strings.Contains(note, fileSystemNote("hfs")) {
		t.Fatalf("create=%q all=%q note=%q", create, all, note)
	}
}
