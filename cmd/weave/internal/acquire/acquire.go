package acquire

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// ResolveSource resolves a local source relative to the checkout that owns its declaration.
func ResolveSource(raw, ownerDir string) (Source, error) { return sourceAt(raw, ownerDir) }

func sourceAt(raw, dir string) (Source, error) {
	s, err := NormalizeSource(raw)
	if err != nil {
		return s, err
	}
	if strings.HasPrefix(s.Identity, "file:") {
		p := strings.TrimPrefix(s.Identity, "file:")
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		s.Identity = "file:" + canonical(p)
		if !strings.HasPrefix(s.URL, "file://") {
			s.URL = p
		}
	}
	return s, nil
}

func canonical(p string) string {
	p, _ = filepath.Abs(p)
	if v, err := filepath.EvalSymlinks(p); err == nil {
		return v
	}
	parent := filepath.Dir(p)
	if parent != p {
		return filepath.Join(canonical(parent), filepath.Base(p))
	}
	return p
}

func manifest(dir string) error {
	p := filepath.Join(dir, "construct/base.manifest")
	st, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("layer %s requires construct/base.manifest: %w", dir, err)
	}
	if !st.Mode().IsRegular() {
		return fmt.Errorf("layer manifest %s must be a regular file", p)
	}
	return nil
}

// Ensure clones into a temporary sibling, validates, and publishes only complete
// checkouts. Existing matching checkouts are reused without fetch/pull/reset.
func Ensure(ctx context.Context, dir, source string, requireLayer bool) error {
	return (Client{}).Ensure(ctx, dir, source, requireLayer)
}

func (c Client) Ensure(ctx context.Context, dir, source string, requireLayer bool) error {
	dir = canonical(dir)
	if source == "" {
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("missing repository %s: record its source in construct/deps: %w", dir, err)
		}
		if requireLayer {
			return manifest(dir)
		}
		return nil
	}
	s, err := sourceAt(source, filepath.Dir(dir))
	if err != nil {
		return err
	}
	if err := reclaimStages(dir); err != nil {
		return err
	}
	if _, err := os.Lstat(dir); err == nil {
		return c.checkExisting(ctx, dir, s, requireLayer)
	} else if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(dir)
	tmp, err := newStage(dir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	checkout := filepath.Join(tmp, "checkout")
	if _, err = c.git(ctx, parent, "clone", "--", s.URL, checkout); err != nil {
		return err
	}
	if requireLayer {
		if err = manifest(checkout); err != nil {
			return err
		}
	}
	// A racing checkout is a conflict; never deliberately replace it.
	if _, err = os.Lstat(dir); err == nil {
		return fmt.Errorf("destination %s appeared during clone; inspect it and retry", dir)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.Rename(checkout, dir); err != nil {
		return fmt.Errorf("publish clone %s: %w", dir, err)
	}
	return nil
}

func (c Client) checkExisting(ctx context.Context, dir string, s Source, requireLayer bool) error {
	top, err := c.git(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("inspect destination %s: %w", dir, err)
	}
	if canonical(top) != dir {
		return fmt.Errorf("destination %s is not a repository checkout; move it or choose another path", dir)
	}
	origin, err := c.Origin(ctx, dir)
	if err != nil {
		return fmt.Errorf("destination %s has no origin matching %s: %w", dir, s.URL, err)
	}
	if origin == "" {
		return fmt.Errorf("destination %s has no origin matching %s", dir, s.URL)
	}
	actual, err := sourceAt(origin, dir)
	if err != nil {
		return err
	}
	if actual.Identity != s.Identity {
		return fmt.Errorf("destination %s origin %s conflicts with declared source %s; fix construct/deps or choose another path", dir, origin, s.URL)
	}
	if requireLayer {
		return manifest(dir)
	}
	return nil
}

type Mount struct{ Owner, Source, Target string }
type Result struct {
	Layers  []string
	Mounts  []Mount
	Missing []string
}

// Restore acquires every declared source, then delegates topology and cycle
// detection to layergraph.Resolve. Data mounts are descriptions only: the
// composition owner creates them. Dry runs never clone and report an incomplete
// graph instead of pretending absent repositories have no dependencies.
func Restore(ctx context.Context, root string, dryRun bool) (Result, error) {
	return (Client{}).Restore(ctx, root, dryRun)
}

func (c Client) Restore(ctx context.Context, root string, dryRun bool) (Result, error) {
	root = canonical(root)
	var result Result
	edges := map[string][]string{}
	seen := map[string]bool{}
	destinations := map[string]string{}
	dataSources := map[string]string{}
	mounts := map[string]string{}
	queue := []string{root}
	for len(queue) > 0 {
		owner := queue[0]
		queue = queue[1:]
		if seen[owner] {
			continue
		}
		seen[owner] = true
		content, err := os.ReadFile(filepath.Join(owner, "construct/deps"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, err
		}
		rows, err := layergraph.ParseRows(string(content))
		if err != nil {
			return result, err
		}
		for _, row := range rows {
			dest := row.Path
			var src Source
			if row.Source != "" {
				src, err = sourceAt(row.Source, owner)
				if err != nil {
					return result, fmt.Errorf("%s: %w", owner, err)
				}
			}
			if row.Kind == "data" {
				target, err := mountTarget(owner, row.Mount)
				if err != nil {
					return result, err
				}
				dest = filepath.Join(filepath.Dir(owner), src.Name)
				if previous := dataSources[src.Identity]; previous != "" {
					dest = previous
				} else {
					dataSources[src.Identity] = canonical(dest)
				}
				dest = canonical(dest)
				if old, ok := mounts[target]; ok && old != dest {
					return result, fmt.Errorf("data mount %s has conflicting sources %s and %s", target, old, dest)
				}
				mounts[target] = dest
				result.Mounts = append(result.Mounts, Mount{Owner: owner, Source: dest, Target: target})
			} else {
				if !filepath.IsAbs(dest) {
					dest = filepath.Join(owner, dest)
				}
				dest = canonical(dest)
				edges[owner] = append(edges[owner], dest)
			}
			if src.Identity != "" {
				if old, ok := destinations[dest]; ok && old != src.Identity {
					return result, fmt.Errorf("destination %s has conflicting declared sources %s and %s", dest, old, src.Identity)
				}
				destinations[dest] = src.Identity
			}
			_, statErr := os.Lstat(dest)
			if os.IsNotExist(statErr) {
				if row.Source == "" {
					return result, fmt.Errorf("missing substrate %s declared in %s: record its source in construct/deps", dest, owner)
				}
				if dryRun {
					result.Missing = append(result.Missing, fmt.Sprintf("%s from %s", dest, src.URL))
					continue
				}
				if err := c.Ensure(ctx, dest, src.URL, row.Kind == "substrate"); err != nil {
					return result, err
				}
			} else if statErr != nil {
				return result, statErr
			} else if row.Source != "" {
				if err := c.checkExisting(ctx, dest, src, row.Kind == "substrate"); err != nil {
					return result, err
				}
			} else if err := manifest(dest); err != nil {
				return result, err
			}
			if row.Kind == "substrate" {
				queue = append(queue, dest)
			}
		}
	}
	if len(result.Missing) > 0 {
		return result, fmt.Errorf("incomplete dependency graph (dry-run): missing %s", strings.Join(result.Missing, ", "))
	}
	var err error
	result.Layers, err = layergraph.Resolve(root, edges)
	return result, err
}

func mountTarget(owner, mount string) (string, error) {
	clean := filepath.Clean(mount)
	if mount == "" || filepath.IsAbs(mount) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("data mount %q must stay inside %s", mount, owner)
	}
	target := filepath.Join(owner, clean)
	// Canonicalize the parent only: an existing final symlink belongs to the
	// composition reconciler, but a parent symlink must not redirect writes out.
	parent := canonical(filepath.Dir(target))
	rel, err := filepath.Rel(owner, parent)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("data mount %q escapes %s through a parent symlink", mount, owner)
	}
	return filepath.Join(parent, filepath.Base(target)), nil
}
