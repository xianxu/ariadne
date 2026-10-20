// Package scratchdisk finds the TestScratch RAM disk (#468): where suites'
// temp files and review checkouts go when it is mounted with room, sparing the
// shared SSD. Absent or full, callers keep their normal paths.
package scratchdisk

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

const (
	// DefaultRoot is where the operator's login item mounts the RAM disk.
	DefaultRoot = "/Volumes/TestScratch"
	// Floor is the free space below which callers fall back.
	Floor = 2 << 30
)

// Probe is what the RAM disk looked like when read.
type Probe struct {
	Root       string
	Mounted    bool // its own file system, so an unmounted path is never written to the SSD
	FreeBytes  uint64
	FileSystem string // statfs type; Darwin scratch requires APFS (#474)
	Err        error
}

// Choose decides whether to use the disk. note is empty when the disk is
// simply absent: that is the normal path, not news.
func Choose(p Probe, floor uint64) (use bool, note string) {
	switch {
	case p.Root == "" || !p.Mounted:
		return false, ""
	case p.Err != nil:
		return false, fmt.Sprintf("test scratch %s unreadable (%v); using the normal folder", p.Root, p.Err)
	case runtime.GOOS == "darwin" && p.FileSystem != "apfs":
		name := p.FileSystem
		if name == "" {
			name = "unknown"
		}
		return false, fmt.Sprintf("test scratch %s is %s; APFS required for nanosecond timestamps and clonefile; using the normal folder", p.Root, name)
	case p.FreeBytes < floor:
		return false, fmt.Sprintf("test scratch %s has %d MiB free (< %d MiB); using the normal folder", p.Root, p.FreeBytes>>20, floor>>20)
	}
	return true, ""
}

// Read probes the disk. WF_TEST_SCRATCH names another root (a fixture, or the
// operator's override) and skips only the mount check, never Darwin's APFS
// requirement; "off" turns it off.
func Read() Probe {
	root, explicit := os.LookupEnv("WF_TEST_SCRATCH")
	switch {
	case root == "off":
		return Probe{}
	case !explicit || root == "":
		root, explicit = DefaultRoot, false
	}
	p := Probe{Root: root, Mounted: explicit}
	if !explicit {
		var disk, parent syscall.Stat_t
		if syscall.Stat(root, &disk) != nil || syscall.Stat(filepath.Dir(root), &parent) != nil {
			return p
		}
		p.Mounted = disk.Dev != parent.Dev
	}
	var fs syscall.Statfs_t
	if err := syscall.Statfs(root, &fs); err != nil {
		p.Err = err
		return p
	}
	p.FreeBytes = uint64(fs.Bavail) * uint64(fs.Bsize)
	p.FileSystem = fileSystemName(&fs)
	return p
}

// Use is Read then Choose at Floor: the root to use, or "" with a note.
func Use() (root, note string) {
	p := Read()
	if ok, note := Choose(p, Floor); !ok {
		return "", note
	}
	return p.Root, ""
}

// Roots places a family of disposable folders: create is where a new one goes
// (<disk>/<sub> when the disk is mounted with room, else fallback), and all is
// every place one may live, for a sweep. A folder's path is fixed when it is
// made, so one started on either root stays there; all lists the disk's
// folder whenever the disk is mounted and readable, full or not, so a sweep
// still reclaims what an earlier run left on it. An os.TempDir fallback also
// enumerates all normal roots MkdirTemp may repair into. note says why a mounted disk
// was passed over.
func Roots(sub, fallback string) (create string, all []string, note string) {
	create, all, note = roots(Read(), sub, fallback)
	if fallback == os.TempDir() {
		// MkdirTemp can repair this fallback. Sweep every possible repaired
		// root even when the inherited root works again; old runs stay put.
		user, _ := UserTempDir() // cleanup is best-effort if the lookup fails
		all = tempRoots(append(all, user, normalTempFallback)...)
	}
	return create, all, note
}

func roots(p Probe, sub, fallback string) (create string, all []string, note string) {
	use, note := Choose(p, Floor)
	create, all = fallback, []string{fallback}
	if p.Mounted && p.Err == nil {
		ram := filepath.Join(p.Root, sub)
		all = append([]string{ram}, all...)
		if use {
			create = ram
		}
	}
	return create, all, note
}

// MkdirTemp makes a fresh folder for the sub family, on the disk when Roots
// picks it and the disk accepts the write, else under fallback: a mounted disk
// that refuses writes (an agent sandbox denies it) costs the RAM, never the run.
// An unusable os.TempDir fallback is repaired lazily; custom roots retain their errors.
func MkdirTemp(sub, fallback, pattern string) (string, error) {
	return mkdirTemp(sub, fallback, pattern, NormalTempDir)
}

func mkdirTemp(sub, fallback, pattern string, normal func() (string, error)) (string, error) {
	create, _, _ := Roots(sub, fallback)
	if create != fallback {
		if os.MkdirAll(create, 0o755) == nil {
			if dir, err := os.MkdirTemp(create, pattern); err == nil {
				return dir, nil
			}
		}
	}
	dir, err := makeTemp(fallback, pattern)
	if err == nil || fallback != os.TempDir() {
		return dir, err
	}
	root, selectErr := normal()
	if selectErr != nil {
		return "", fmt.Errorf("normal temp fallback after %v: %w", err, selectErr)
	}
	return makeTemp(root, pattern)
}

func makeTemp(root, pattern string) (string, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(root, pattern)
}
