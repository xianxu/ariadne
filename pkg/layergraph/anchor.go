package layergraph

import (
	"fmt"
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/workspace"
)

// Anchors finds a declaring layer's fallback anchor (#230). An error is the
// reason there is none; resolution then stays local.
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
// layer-graph consumer resolves through it (ARCH-DRY): this package's walk,
// weave's acquisition and sdlc's substrate walk.
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

// Substrate resolves one row declared by owner. The error is informational:
// the reason no fallback applied when the local target was absent.
func (r *Resolver) Substrate(owner, target string) (Resolution, error) {
	probe := func(p string) Probe {
		switch {
		case hasManifest(r.FS, p):
			return Layer
		case pathExists(r.FS, p):
			return NotLayer
		}
		return Absent
	}
	if filepath.IsAbs(target) || probe(filepath.Join(owner, target)) != Absent {
		return ChooseSubstrate(owner, target, "", probe), nil
	}
	a := r.anchor(owner)
	res := ChooseSubstrate(owner, target, a.primary, probe)
	switch {
	case res.Via != "":
		return res, nil
	case a.err != nil:
		return res, a.err
	}
	return res, fmt.Errorf("absent beside primary checkout %s too (%s)", a.primary, filepath.Join(a.primary, target))
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
