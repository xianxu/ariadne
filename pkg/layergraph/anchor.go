package layergraph

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/workspace"
)

// Anchors finds a declaring layer's fallback anchor (#230). A
// *workspace.NoAnchorError is the reason there is none, and resolution stays
// local; any other error is a failure to find out.
type Anchors interface {
	PrimaryCheckout(layer string) (string, error)
}

// GitAnchors is the production Anchors: workspace.DependencyAnchor over git.
type GitAnchors struct{}

func (GitAnchors) PrimaryCheckout(layer string) (string, error) {
	return workspace.DependencyAnchor(workspace.ExecGit{}, layer)
}

// Resolver resolves substrate rows for one walk: FS probes, plus an anchor
// looked up only when a relative target is absent, memoized per owner. Every
// reader that maps a construct/deps row to a path resolves through it
// (ARCH-DRY); the callers, and the two readers that provably need not, are
// listed once in atlas/workflow/setup-and-replication.md (Linked worktrees).
type Resolver struct {
	FS      FS
	Anchors Anchors
	memo    map[string]anchorResult
}

type anchorResult struct {
	primary string
	err     error
}

// NewResolver uses fs's own Anchors when it implements one (test fakes),
// else GitAnchors.
func NewResolver(fs FS) *Resolver {
	a, ok := fs.(Anchors)
	if !ok {
		a = GitAnchors{}
	}
	return &Resolver{FS: fs, Anchors: a}
}

// lstater is an FS that can stat without following a final symlink.
type lstater interface {
	Lstat(path string) (os.FileInfo, error)
}

// probe classifies p. Presence is judged without following a final symlink
// and fails toward present: a dangling or unreadable local target is an
// existing invalid target, never a reason to fall back (#230 spec 5).
func (r *Resolver) probe(p string) Probe {
	if hasManifest(r.FS, p) {
		return Layer
	}
	var err error
	if l, ok := r.FS.(lstater); ok {
		_, err = l.Lstat(p)
	} else {
		_, err = r.FS.Stat(p)
	}
	if errors.Is(err, iofs.ErrNotExist) {
		return Absent
	}
	return NotLayer
}

// Substrate resolves one row declared by owner. An absent relative row that
// did not fall back carries the reason in NoFallback; the error return is a
// failure to find owner's anchor (Git could not answer), which the caller
// must not read as "not a worktree". With that error the Resolution is the
// absent local spelling, for a caller that reports rather than fails.
func (r *Resolver) Substrate(owner, target string) (Resolution, error) {
	if filepath.IsAbs(target) || r.probe(filepath.Join(owner, target)) != Absent {
		return ChooseSubstrate(owner, target, "", r.probe), nil
	}
	a := r.anchor(owner)
	var none *workspace.NoAnchorError
	if a.err != nil && !errors.As(a.err, &none) {
		local := filepath.Join(owner, target)
		return Resolution{Path: local, Local: local, Probe: Absent}, fmt.Errorf("resolve substrate %s declared in %s: find its primary checkout: %w", target, owner, a.err)
	}
	res := ChooseSubstrate(owner, target, a.primary, r.probe)
	switch {
	case res.Via != "":
	case a.err != nil:
		res.NoFallback = a.err
	default:
		res.NoFallback = &FallbackAbsentError{Primary: a.primary, Candidate: filepath.Join(a.primary, target)}
	}
	return res, nil
}

func (r *Resolver) anchor(owner string) anchorResult {
	if a, ok := r.memo[owner]; ok {
		return a
	}
	if r.memo == nil {
		r.memo = map[string]anchorResult{}
	}
	var a anchorResult
	a.primary, a.err = r.Anchors.PrimaryCheckout(owner)
	r.memo[owner] = a
	return a
}

// FallbackAbsentError is a relative row absent both beside its linked
// worktree and beside the worktree's primary checkout (#230). Acquisition
// refuses it rather than cloning a duplicate beside the worktree.
type FallbackAbsentError struct{ Primary, Candidate string }

func (e *FallbackAbsentError) Error() string {
	return fmt.Sprintf("absent beside primary checkout %s too (%s)", e.Primary, e.Candidate)
}
