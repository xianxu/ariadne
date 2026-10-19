package scratchdisk

import (
	"errors"
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
