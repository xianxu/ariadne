// Package weavefs is weave's filesystem seam: the injectable FS interface that
// isolates every disk read/mutation behind one boundary (ARCH-PURE — the pure
// compiler core never touches disk; only the walk readers and plan.Apply do,
// through this interface). OSFS is the production implementation backed by the
// os package; tests inject a t.TempDir()-rooted OSFS (real FS, hermetic dir) so
// the seam is exercised end-to-end without mocks.
package weavefs

import "os"

// FS is the filesystem the IO seam reads and mutates through. Reads back the
// manifest/deps/prose files for the walk; the mutations (Symlink/WriteFile/
// Mkdir) are what plan.Apply executes. Paths are absolute (the walk resolves
// them against the repo root before calling). Keeping this an interface lets
// Apply and the walk be exercised against a hermetic t.TempDir without mocking
// the os package.
type FS interface {
	// ReadFile returns the bytes of the file at path.
	ReadFile(path string) ([]byte, error)
	// Stat reports whether path exists and, if so, its FileInfo. Used for the
	// present-skip + idempotency checks (ported from setup.sh's [[ -e ]] guards).
	Stat(path string) (os.FileInfo, error)
	// Lstat is Stat without following a final symlink (so an existing symlink
	// is detected as a symlink, like setup.sh's [[ -L ]]).
	Lstat(path string) (os.FileInfo, error)
	// Readlink returns the target of the symlink at path.
	Readlink(path string) (string, error)
	// Remove deletes path (a single file or symlink).
	Remove(path string) error
	// RemoveAll deletes path and any children (setup.sh's `rm -rf` for a
	// regular file/dir occupying a symlink's slot).
	RemoveAll(path string) error
	// MkdirAll creates dir and any missing parents (setup.sh's `mkdir -p`).
	MkdirAll(path string) error
	// Symlink creates a symlink at name pointing at oldname (the link target,
	// which the seam computes relative to name's dir — see plan.Apply).
	Symlink(oldname, name string) error
	// WriteFile writes data to path, creating it if needed (setup.sh seeds /
	// the composed AGENTS.md).
	WriteFile(path string, data []byte) error
}

// OSFS is the production FS backed by the os package. Its zero value is ready
// to use. Tests use it too, rooted at a t.TempDir(), so the seam is verified
// against a real filesystem (ARCH: faithful over mocked).
type OSFS struct{}

func (OSFS) ReadFile(path string) ([]byte, error)      { return os.ReadFile(path) }
func (OSFS) Stat(path string) (os.FileInfo, error)      { return os.Stat(path) }
func (OSFS) Lstat(path string) (os.FileInfo, error)     { return os.Lstat(path) }
func (OSFS) Readlink(path string) (string, error)       { return os.Readlink(path) }
func (OSFS) Remove(path string) error                   { return os.Remove(path) }
func (OSFS) RemoveAll(path string) error                { return os.RemoveAll(path) }
func (OSFS) MkdirAll(path string) error                 { return os.MkdirAll(path, 0o755) }
func (OSFS) Symlink(oldname, name string) error         { return os.Symlink(oldname, name) }
func (OSFS) WriteFile(path string, data []byte) error   { return os.WriteFile(path, data, 0o644) }

// ensure OSFS satisfies FS at compile time.
var _ FS = OSFS{}
