package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/pkg/workspace"
)

const assetPolicyPath = issueassets.PolicyPath

// resolveAssetPolicy never writes policy. Unknown host identity or visibility is
// a conservative decision, while an invalid operator file is a repairable error.
func resolveAssetPolicy(ctx context.Context, root string) (issueassets.Decision, error) {
	p := filepath.Join(root, filepath.FromSlash(assetPolicyPath))
	raw, err := readAssetRegular(root, assetPolicyPath, 4096)
	if err == nil {
		policy, err := issueassets.ParsePolicy(raw)
		if err != nil {
			return issueassets.Decision{}, fmt.Errorf("%s: %v; ask the operator to repair %s", assetPolicyPath, err, assetPolicyPath)
		}
		d := issueassets.ResolvePolicy(&policy, "")
		d.Source = assetPolicyPath
		return d, nil
	}
	if !os.IsNotExist(err) {
		return issueassets.Decision{}, fmt.Errorf("%s: %v; ask the operator to repair %s", p, err, assetPolicyPath)
	}
	unknown := issueassets.ResolvePolicy(nil, "")
	unknown.Source = "host visibility unknown"
	id, err := workspace.Resolve(execGitRunner{}, root, "")
	if err != nil {
		return unknown, nil
	}
	rest := "main"
	if id.RestingBranch != nil {
		rest = *id.RestingBranch
	}
	target, err := gitx.ResolvePublicationTarget(ctx, root, rest)
	if err != nil || !strings.HasPrefix(target.Repository, "github.com/") {
		return unknown, nil
	}
	repo := strings.TrimPrefix(target.Repository, "github.com/")
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	visibility, err := ghClient.RepoVisibility(queryCtx, repo)
	if err != nil {
		unknown.Source = "host visibility unavailable for " + repo
		return unknown, nil
	}
	d := issueassets.ResolvePolicy(nil, strings.ToLower(visibility))
	d.Source = "host visibility " + strings.ToLower(visibility) + " for " + repo
	return d, nil
}

func requireAssetPolicy(d issueassets.Decision) error {
	if d.Allowed {
		return nil
	}
	return fmt.Errorf("repository asset policy is text-only (%s); ask the operator to change %s", d.Source, assetPolicyPath)
}

// readAssetRegular checks each repository-relative component before opening.
// Root may be the system's canonical temp-directory alias; child symlinks cannot
// redirect policy or evidence outside this root.
func readAssetRegular(root, rel string, limit int64) ([]byte, error) {
	if filepath.IsAbs(rel) || filepath.Clean(rel) != rel || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("unsafe asset path %q; use a repository-relative file", rel)
	}
	current := root
	parts := strings.Split(rel, string(filepath.Separator))
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symlink at %s; replace it with a regular file or directory", rel)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return nil, fmt.Errorf("non-directory parent of %s; repair that directory", rel)
		}
		if i == len(parts)-1 && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("nonregular file %s; use a regular file", rel)
		}
	}
	// os.Root prevents a concurrently substituted symlink from escaping the root.
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	f, err := r.Open(rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("nonregular file %s; use a regular file", rel)
	}
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("%s exceeds %d bytes; reduce the file before retrying", rel, limit)
	}
	return raw, nil
}
