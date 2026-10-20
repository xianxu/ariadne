package scratchdisk

import (
	"syscall"
	"testing"
)

func TestDarwinFileSystemName(t *testing.T) {
	for _, name := range []string{"apfs", "hfs", ""} {
		var fs syscall.Statfs_t
		for i, c := range []byte(name) {
			fs.Fstypename[i] = int8(c)
		}
		fs.Fstypename[len(name)+1] = 'x' // ignore bytes after the terminating NUL
		if got := fileSystemName(&fs); got != name {
			t.Fatalf("got %q, want %q", got, name)
		}
	}
}

func TestReadDarwinOverrideChecksActualFileSystem(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", root)
	var fs syscall.Statfs_t
	if err := syscall.Statfs(root, &fs); err != nil {
		t.Fatal(err)
	}
	p := Read()
	want := fileSystemName(&fs)
	if p.Err != nil || p.FileSystem != want || want == "" {
		t.Fatalf("probe %+v, actual %q", p, want)
	}
	use, note := Choose(p, 0)
	if use != (want == "apfs") || (use && note != "") || (!use && note == "") {
		t.Fatalf("use=%v note=%q for %q", use, note, want)
	}
}

// Read-only conformance against the operator's mount, when it exists.
func TestDarwinMountedScratchAdmission(t *testing.T) {
	t.Setenv("WF_TEST_SCRATCH", "")
	p := Read()
	if !p.Mounted {
		t.Skip("TestScratch is not mounted")
	}
	if p.Err != nil {
		t.Fatal(p.Err)
	}
	use, note := Choose(p, 0)
	t.Logf("root=%s filesystem=%s use=%v note=%q", p.Root, p.FileSystem, use, note)
	if use != (p.FileSystem == "apfs") || (!use && note == "") {
		t.Fatalf("probe=%+v use=%v note=%q", p, use, note)
	}
}
