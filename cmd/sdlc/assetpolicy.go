package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
	return resolveAssetHostPolicy(ctx, root)
}

func resolveAssetHostPolicy(ctx context.Context, root string) (issueassets.Decision, error) {
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
	// Keep the publication host in the gh argument: bare owner/repo would
	// inherit GH_HOST and could authorize a different repository.
	repo := "https://" + target.Repository
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

// requireAssetPublicationPolicy intersects local restrictions with the policy at
// the exact destination base. A stale checkout can restrict publication, never
// authorize bytes that the destination forbids. Call once per candidate attempt.
func requireAssetPublicationPolicy(ctx context.Context, root string, destination *gitx.TrunkView) error {
	local, err := resolveAssetPolicy(ctx, root)
	if err != nil {
		return err
	}
	if err := requireAssetPolicy(local); err != nil {
		return err
	}
	raw, present, err := destinationAssetPolicy(ctx, root, destination)
	if err != nil {
		return err
	}
	var policy issueassets.Decision
	if !present {
		policy, err = resolveAssetHostPolicy(ctx, root)
	} else {
		parsed, parseErr := issueassets.ParsePolicy(raw)
		if parseErr != nil {
			return fmt.Errorf("destination %s: %v; ask the operator to repair it", assetPolicyPath, parseErr)
		}
		policy = issueassets.ResolvePolicy(&parsed, "")
		policy.Source = assetPolicyPath
	}
	if err != nil {
		return err
	}
	policy.Source = "destination " + destination.Ref() + ": " + policy.Source
	return requireAssetPolicy(policy)
}

// Read exact entries, including directories and symlink ancestors. The immutable
// blob's size is checked before reading, so unrelated evidence is never loaded.
func destinationAssetPolicy(ctx context.Context, root string, view *gitx.TrunkView) ([]byte, bool, error) {
	fail := func(err error) ([]byte, bool, error) {
		return nil, false, fmt.Errorf("cannot read destination %s at %s: %w; fetch that destination and ask the operator to restore a regular, valid policy file before retrying", assetPolicyPath, view.Ref(), err)
	}
	for _, p := range []string{"workshop", assetPolicyPath} {
		raw, _, err := gitx.RunGitContext(ctx, "-C", root, "ls-tree", "-z", "--full-tree", view.Ref(), "--", p)
		if err != nil {
			return fail(err)
		}
		if len(raw) == 0 {
			return nil, false, nil
		}
		row := strings.TrimSuffix(string(raw), "\x00")
		header, name, ok := strings.Cut(row, "\t")
		fields := strings.Fields(header)
		if !ok || name != p || len(fields) != 3 {
			return fail(fmt.Errorf("invalid policy tree entry"))
		}
		if p == "workshop" {
			if fields[0] != "040000" || fields[1] != "tree" {
				return fail(fmt.Errorf("policy parent %s is not a directory", p))
			}
			continue
		}
		if (fields[0] != "100644" && fields[0] != "100755") || fields[1] != "blob" {
			return fail(fmt.Errorf("policy is not a regular file"))
		}
		sizeRaw, _, err := gitx.RunGitContext(ctx, "-C", root, "cat-file", "-s", fields[2])
		if err != nil {
			return fail(err)
		}
		size, err := strconv.ParseInt(strings.TrimSpace(string(sizeRaw)), 10, 64)
		if err != nil || size < 0 || size > 4096 {
			return fail(fmt.Errorf("policy exceeds the 4096-byte limit or has invalid size"))
		}
		content, err := view.ReadBlob(fields[2])
		if err != nil {
			return fail(err)
		}
		return content, true, nil
	}
	return nil, false, nil
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
