// peerproject.go — close-time project updates in a PEER repository (#308, C1).
// A peer's project is never edited in its checkout: that checkout belongs to
// someone else (a TL's live ops tree, kink 38), and a local commit there
// diverges it from origin. The edit is re-derived on the peer's fetched main and
// published to its origin by one compare-and-swap commit (gitx.TrunkFile),
// leaving the peer's files, index and HEAD untouched. A moved origin is re-read
// and the edit recomputed on its bytes — never a replay of stale ones — within
// TrunkFile's bounded attempts.
package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// errPeerTLOwned refuses a peer project whose origin copy gained a TL: only the
// TL writes it (#326), even if the local copy close previewed had none.
var errPeerTLOwned = errors.New("TL-owned on origin")

// peerProjectPath resolves the project file to its repository-relative path in
// the peer, following symlinks (ops/workshop/projects/x.md is a link into
// teams/…): the commit must change the real file, not the link.
func peerProjectPath(repoDir, projectPath string) (string, error) {
	real, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		return "", err
	}
	root, err := filepath.EvalSymlinks(repoDir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, real)
	if err != nil || gitx.Escapes(rel) {
		return "", fmt.Errorf("%s resolves outside %s", projectPath, repoDir)
	}
	return filepath.ToSlash(rel), nil
}

// peerTrunk is the peer's main on its publication remote (the upstream of its
// resting branch), or an error naming why there is none.
func peerTrunk(ctx context.Context, repoDir string) (*gitx.TrunkFile, error) {
	resting := "main"
	if id, err := workspace.Resolve(execGitRunner{}, repoDir, ""); err == nil && id.RestingBranch != nil {
		resting = *id.RestingBranch
	}
	target, err := gitx.ResolvePublicationTarget(ctx, repoDir, resting)
	if err != nil {
		return nil, err
	}
	return gitx.NewTrunkFileContext(ctx, repoDir, target.Remote, "main")
}

// peerProjectTransform is the edit published on the peer's fetched bytes: a
// copy now TL-owned (or whose owner cannot be read) refuses; otherwise the same
// projectCloseEdit close previewed. Unchanged bytes publish nothing.
func peerProjectTransform(in projectEditInput) func([]byte) ([]byte, error) {
	return func(old []byte) ([]byte, error) {
		owners, err := project.TLOwners(string(old))
		if err != nil {
			return nil, fmt.Errorf("cannot read tl: on origin (%v)", err)
		}
		if len(owners) > 0 {
			return nil, fmt.Errorf("%w (tl: %v)", errPeerTLOwned, owners)
		}
		next, _, _, err := projectCloseEdit(string(old), in)
		if err != nil {
			return nil, err
		}
		return []byte(next), nil
	}
}

// publishPeerProject publishes one peer project edit to the peer's origin.
func publishPeerProject(ctx context.Context, e projectEdit, closingRef string) error {
	if gitx.IsBrainRepo(e.repoDir) {
		return errors.New("a brain capture repository is never written (#176)")
	}
	rel, err := peerProjectPath(e.repoDir, e.path)
	if err != nil {
		return err
	}
	tf, err := peerTrunk(ctx, e.repoDir)
	if err != nil {
		return err
	}
	return tf.Update(rel, fmt.Sprintf("project: close-time update (%s)", closingRef), peerProjectTransform(e.in))
}
