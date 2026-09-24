package refresh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

const declarationLimit int64 = 1 << 20

func canonical(p string) (string, error) {
	p, e := filepath.Abs(p)
	if e != nil {
		return "", e
	}
	return filepath.EvalSymlinks(p)
}
func record(ctx context.Context, g acquire.GitRunner, p string, args ...string) (string, error) {
	s, e := g.Run(ctx, p, args...)
	if e != nil {
		return "", e
	}
	return parseRecord(s)
}
func oid(ctx context.Context, g acquire.GitRunner, p, ref string) (string, error) {
	s, e := g.Run(ctx, p, "rev-parse", "--verify", ref+"^{commit}")
	if e != nil {
		return "", e
	}
	return parseOID(s)
}
func readDeclarations(p string) ([]layergraph.Dependency, error) {
	b, e := acquire.ReadDeclarations(filepath.Join(p, "construct/deps"), declarationLimit)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	return layergraph.ParseRows(string(b))
}
func observe(ctx context.Context, g acquire.GitRunner, p string) (Snapshot, error) {
	s := Snapshot{path: p}
	actual, e := canonical(p)
	if e != nil {
		return s, e
	}
	if actual != p {
		return s, fmt.Errorf("checkout path binding changed")
	}
	top, e := record(ctx, g, p, "rev-parse", "--show-toplevel")
	if e != nil {
		return s, e
	}
	top, e = canonical(top)
	if e != nil || top != p {
		return s, fmt.Errorf("not the exact repository checkout")
	}
	common, e := record(ctx, g, p, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if e != nil {
		return s, e
	}
	s.common, e = canonical(common)
	if e != nil {
		return s, e
	}
	private, e := record(ctx, g, p, "rev-parse", "--absolute-git-dir")
	if e != nil {
		return s, e
	}
	s.gitdir, e = canonical(private)
	if e != nil {
		return s, e
	}
	b, e := g.Run(ctx, p, "symbolic-ref", "--quiet", "HEAD")
	if e != nil {
		return s, fmt.Errorf("detached HEAD; choose a working branch")
	}
	s.branch, e = parseBranch(b)
	if e != nil {
		return s, e
	}
	s.head, e = oid(ctx, g, p, "HEAD")
	if e != nil {
		return s, fmt.Errorf("HEAD must name an existing commit: %w", e)
	}
	origin, e := record(ctx, g, p, "remote", "get-url", "origin")
	if e != nil || origin == "" {
		return s, fmt.Errorf("a usable origin is required")
	}
	source, e := acquire.ResolveSource(origin, p)
	if e != nil {
		return s, fmt.Errorf("invalid origin")
	}
	s.origin = source.Identity
	for _, name := range []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "rebase-apply", "rebase-merge", "sequencer", "BISECT_START"} {
		v, e := record(ctx, g, p, "rev-parse", "--path-format=absolute", "--git-path", name)
		if e != nil {
			return s, e
		}
		_, e = os.Lstat(v)
		if e == nil {
			return s, fmt.Errorf("active Git operation (%s); resolve or abort it explicitly", name)
		}
		if !os.IsNotExist(e) {
			return s, e
		}
	}
	status, e := g.Run(ctx, p, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none")
	if e != nil {
		return s, e
	}
	if status != "" {
		return s, fmt.Errorf("checkout is dirty (including untracked files or submodules)")
	}
	s.rows, e = readDeclarations(p)
	return s, e
}
func sameStart(a, b Snapshot) bool {
	return a.path == b.path && a.common == b.common && a.gitdir == b.gitdir && a.branch == b.branch && a.head == b.head && a.origin == b.origin && reflect.DeepEqual(a.rows, b.rows)
}
func revalidate(ctx context.Context, g acquire.GitRunner, s Snapshot) error {
	now, e := observe(ctx, g, s.path)
	if e != nil {
		return e
	}
	if !sameStart(s, now) {
		return fmt.Errorf("checkout identity, branch, HEAD, origin or declarations changed; retry")
	}
	return nil
}
func ancestor(ctx context.Context, g acquire.GitRunner, p, a, b string) (bool, error) {
	_, e := g.Run(ctx, p, "merge-base", "--is-ancestor", a, b)
	if e == nil {
		return true, nil
	}
	var code interface{ ExitCode() int }
	if errors.As(e, &code) && code.ExitCode() == 1 {
		return false, nil
	}
	return false, e
}
func targetDeclarations(ctx context.Context, g acquire.GitRunner, s Snapshot) ([]layergraph.Dependency, error) {
	tree, e := g.Run(ctx, s.path, "ls-tree", "-z", s.target, "--", "construct/deps")
	if e != nil {
		return nil, e
	}
	if tree == "" {
		return nil, nil
	}
	if !strings.HasSuffix(tree, "\x00") || strings.Count(tree, "\x00") != 1 {
		return nil, fmt.Errorf("invalid declaration tree framing")
	}
	meta, path, ok := strings.Cut(strings.TrimSuffix(tree, "\x00"), "\t")
	f := strings.Fields(meta)
	if !ok || path != "construct/deps" || len(f) != 3 || (f[0] != "100644" && f[0] != "100755") || f[1] != "blob" || !oidPattern.MatchString(f[2]) {
		return nil, fmt.Errorf("target construct/deps must be an ordinary blob")
	}
	size, e := record(ctx, g, s.path, "cat-file", "-s", f[2])
	if e != nil {
		return nil, e
	}
	var n int64
	if _, e = fmt.Sscan(size, &n); e != nil || n < 0 || n > declarationLimit {
		return nil, fmt.Errorf("target construct/deps exceeds 1 MiB or invalid size")
	}
	content, e := g.Run(ctx, s.path, "cat-file", "blob", f[2])
	if e != nil {
		return nil, e
	}
	if int64(len(content)) != n {
		return nil, fmt.Errorf("target declaration read was truncated")
	}
	return layergraph.ParseRows(content)
}

// Rebase lacks merge's --no-overwrite-ignore. Check ignored paths against every
// target tree entry, including file/directory collisions, before either effect.
func ignoredCollisions(ctx context.Context, g acquire.GitRunner, s Snapshot, rebase bool) error {
	ignored, e := g.Run(ctx, s.path, "ls-files", "--others", "--ignored", "--exclude-standard", "-z")
	if e != nil {
		return e
	}
	if ignored == "" {
		return nil
	}
	ignoredPaths, e := parseNULRecords(ignored)
	if e != nil {
		return e
	}
	tree, e := g.Run(ctx, s.path, "ls-tree", "-r", "--name-only", "-z", s.target)
	if e != nil {
		return e
	}
	paths, e := parseNULRecords(tree)
	if e != nil {
		return e
	}
	if rebase {
		touched, e := g.Run(ctx, s.path, "log", "--format=", "--name-only", "-z", "--no-renames", "--diff-merges=first-parent", s.target+".."+s.head)
		if e != nil {
			return e
		}
		replay, e := parseNULRecords(touched)
		if e != nil {
			return e
		}
		paths = append(paths, replay...)
	}
	sort.Strings(paths)
	tracked := make(map[string]bool, len(paths))
	for _, p := range paths {
		tracked[p] = true
	}
	for _, i := range ignoredPaths {
		collision := tracked[i]
		for parent := filepath.Dir(i); parent != "."; parent = filepath.Dir(parent) {
			if tracked[filepath.ToSlash(parent)] {
				collision = true
				break
			}
		}
		child := sort.SearchStrings(paths, i+"/")
		if child < len(paths) && strings.HasPrefix(paths[child], i+"/") {
			collision = true
		}
		if collision {
			return fmt.Errorf("ignored path %q collides with target or replay; preserve it elsewhere before refresh", i)
		}
	}
	return nil
}
