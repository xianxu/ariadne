// Package layergraph is the module-level, importable home of the transitive
// construct/deps layer-graph walk — the SINGLE source of truth for "what is
// repo R's layer graph" (ARCH-DRY). It is shared by both cmd/weave (which loads
// the rich per-layer manifests/prose on top of the topology) and any future
// DAG-aware subsystem (e.g. cmd/datatype) that cannot reach into
// cmd/weave/internal/*. The pure graph reasoning (ParseDeps, Resolve) stays
// IO-free; the only disk access — reading construct/deps and stat-ing
// construct/base.manifest — sits behind the minimal FS seam below (ARCH-PURE).
package layergraph

import "os"

// FS is the minimal filesystem seam the layer-graph walk needs: it reads
// construct/deps (ReadFile), and stats construct/base.manifest (Stat). ReadDir
// is included for the per-layer prototype/skill reads a DAG-aware consumer
// (cmd/datatype's mergeTypes) performs over the same topology. Paths are
// absolute (the walk resolves them against the repo root before calling).
//
// Defined here — not in cmd/weave/internal/weavefs — so cmd/datatype (which
// cannot import weave's internal packages) has access to the same seam, and
// weave's existing weavefs.FS satisfies it structurally (a superset interface,
// no behavior change). The physical-path canonicalization the walk performs is
// a real-disk path concern (filepath.EvalSymlinks ≈ pwd -P) handled outside
// this seam, preserving the macOS /tmp→/private/tmp semantics the walk depends
// on.
type FS interface {
	// ReadFile returns the bytes of the file at path.
	ReadFile(path string) ([]byte, error)
	// ReadDir lists the directory entries at path (sorted by name, like
	// os.ReadDir). A missing dir surfaces as an error the caller treats as
	// "nothing here".
	ReadDir(path string) ([]os.DirEntry, error)
	// Stat reports whether path exists and, if so, its FileInfo (the
	// base.manifest-existence layer filter).
	Stat(path string) (os.FileInfo, error)
}

// OSFS is the production FS backed by the os package. Its zero value is ready
// to use. cmd/datatype (which has no weavefs) uses it directly; weave passes
// its own weavefs.FS (which satisfies this interface structurally).
type OSFS struct{}

func (OSFS) ReadFile(path string) ([]byte, error)       { return os.ReadFile(path) }
func (OSFS) ReadDir(path string) ([]os.DirEntry, error) { return os.ReadDir(path) }
func (OSFS) Stat(path string) (os.FileInfo, error)      { return os.Stat(path) }

// ensure OSFS satisfies FS at compile time.
var _ FS = OSFS{}
