package acquire

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
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

func (c Client) Ensure(ctx context.Context, dir, source string, requireLayer bool) (retErr error) {
	state := uninspected
	advance := func(event acquisitionEvent) error {
		next, err := transition(state, event)
		if err == nil {
			state = next
		}
		return err
	}
	defer func() {
		if retErr != nil && state != unconfirmed {
			_, _ = transition(state, operationFailed)
		}
	}()
	lexical, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	dir = canonical(dir)
	scoped := c.Policy != nil && requireLayer
	if scoped {
		var src Source
		if source != "" {
			src, err = sourceAt(source, filepath.Dir(dir))
			if err != nil {
				return err
			}
		}
		if err := c.Policy.validate(lexical, dir, src); err != nil {
			return err
		}
		if _, err := os.Lstat(dir); err == nil {
			if err := c.privateExisting(ctx, dir); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if source == "" {
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("missing repository %s: record its source in construct/deps: %w", dir, err)
		}
		if requireLayer {
			if err := manifest(dir); err != nil {
				return err
			}
		}
		return advance(existingVerified)
	}
	s, err := sourceAt(source, filepath.Dir(dir))
	if err != nil {
		return err
	}
	if err := reclaimStages(dir); err != nil {
		return err
	}
	if _, err := os.Lstat(dir); err == nil {
		if err := c.checkExisting(ctx, dir, s, requireLayer); err != nil {
			return err
		}
		return advance(existingVerified)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := advance(absentRemote); err != nil {
		return err
	}
	parent := filepath.Dir(dir)
	tmp, err := newStage(dir)
	if err != nil {
		return err
	}
	preserveStage := false
	defer func() {
		if !preserveStage {
			retErr = errors.Join(retErr, staging.Remove(tmp))
		}
	}()
	if err := advance(stageCreated); err != nil {
		return err
	}
	checkout := filepath.Join(tmp, "checkout")
	args := []string{"clone"}
	if scoped {
		args = append(args, "--branch", "main")
	}
	args = append(args, "--", s.URL, checkout)
	if _, err = c.gitOwned(ctx, parent, tmp, args...); err != nil {
		return err
	}
	if err := advance(cloneSucceeded); err != nil {
		return err
	}
	if scoped {
		main, err := c.git(ctx, checkout, "rev-parse", "--verify", "refs/remotes/origin/main^{commit}")
		if err != nil {
			return fmt.Errorf("remote must provide branch main: %w", err)
		}
		head, err := c.git(ctx, checkout, "rev-parse", "--verify", "HEAD^{commit}")
		if err != nil {
			return err
		}
		if main != head {
			return fmt.Errorf("new dependency HEAD differs from origin/main")
		}
	}
	if requireLayer {
		if err = manifest(checkout); err != nil {
			return err
		}
	}
	if err := advance(checksSucceeded); err != nil {
		return err
	}
	// A racing checkout is a conflict; never deliberately replace it.
	if _, err = os.Lstat(dir); err == nil {
		preserveStage = scoped
		_ = advance(destinationAppeared)
		return fmt.Errorf("destination %s appeared during clone; inspect it and retry", dir)
	} else if !os.IsNotExist(err) {
		return err
	}
	if scoped {
		if err := c.Policy.validate(lexical, canonical(dir), s); err != nil {
			return err
		}
	}
	if err := advance(destinationAbsent); err != nil {
		return err
	}
	if err = os.Rename(checkout, dir); err != nil {
		preserveStage = scoped
		_ = advance(publicationUncertain)
		return fmt.Errorf("publish clone %s: %w", dir, err)
	}
	return advance(publishSucceeded)
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
		return fmt.Errorf("inspect origin for destination %s; check its Git configuration: %w", dir, err)
	}
	if origin == "" {
		return fmt.Errorf("destination %s has no origin; configure it to match construct/deps or choose another path", dir)
	}
	actual, err := sourceAt(origin, dir)
	if err != nil {
		return fmt.Errorf("destination %s has an invalid origin; fix its Git configuration: %w", dir, err)
	}
	if actual.Identity != s.Identity {
		return fmt.Errorf("destination %s origin conflicts with its declared source; fix construct/deps or choose another path", dir)
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
	destinationKinds := map[string]string{}
	mounts := map[string]string{}
	queue := []string{root}
	queued := map[string]bool{root: true}
	for len(queue) > 0 {
		owner := queue[0]
		queue = queue[1:]
		if seen[owner] {
			continue
		}
		seen[owner] = true
		content, err := c.readDeclarations(filepath.Join(owner, "construct/deps"))
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
				if c.Policy != nil {
					lexical, err := filepath.Abs(dest)
					if err != nil {
						return result, err
					}
					if err := c.Policy.validate(lexical, canonical(dest), src); err != nil {
						return result, err
					}
				}
				dest = canonical(dest)
				if !queued[dest] {
					if c.MaxLayers > 0 && len(queued) >= c.MaxLayers {
						return result, fmt.Errorf("dependency layer limit %d exceeded at %s", c.MaxLayers, dest)
					}
					queued[dest] = true
					queue = append(queue, dest)
				}
				edges[owner] = append(edges[owner], dest)
			}
			if c.Policy != nil {
				if row.Kind == "data" && dest == c.Policy.HostRoot {
					return result, fmt.Errorf("data destination %s collides with environment host", dest)
				}
				if kind, ok := destinationKinds[dest]; ok && kind != row.Kind {
					return result, fmt.Errorf("data and substrate destination %s collide", dest)
				}
				destinationKinds[dest] = row.Kind
				if row.Kind == "substrate" && dryRun {
					if _, err := os.Lstat(dest); err == nil {
						if err := c.privateExisting(ctx, dest); err != nil {
							return result, err
						}
					}
				}
			}
			if src.Identity != "" {
				if old, ok := destinations[dest]; ok && old != src.Identity {
					return result, fmt.Errorf("destination %s has conflicting declared sources; fix construct/deps or choose another path", dest)
				}
				destinations[dest] = src.Identity
			}
			_, statErr := os.Lstat(dest)
			if os.IsNotExist(statErr) {
				if row.Source == "" {
					return result, fmt.Errorf("missing substrate %s declared in %s: record its source in construct/deps", dest, owner)
				}
				if dryRun {
					result.Missing = append(result.Missing, dest)
					continue
				}
				if err := c.Ensure(ctx, dest, src.URL, row.Kind == "substrate"); err != nil {
					return result, err
				}
			} else if statErr != nil {
				return result, statErr
			} else if !dryRun {
				if err := c.Ensure(ctx, dest, src.URL, row.Kind == "substrate"); err != nil {
					return result, err
				}
			} else if row.Source != "" {
				if err := c.checkExisting(ctx, dest, src, row.Kind == "substrate"); err != nil {
					return result, err
				}
			} else if err := manifest(dest); err != nil {
				return result, err
			}
		}
	}
	if len(result.Missing) > 0 {
		return result, fmt.Errorf("incomplete dependency graph (dry-run): missing %s; restore declared dependencies and retry", strings.Join(result.Missing, ", "))
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

// ReadDeclarations reads a dependency document. A positive maxBytes rejects
// nonordinary files and bounds the read before allocation; zero preserves the
// ordinary acquisition behavior of os.ReadFile.
func ReadDeclarations(path string, maxBytes int64) ([]byte, error) {
	return (Client{MaxDeclarationBytes: maxBytes}).readDeclarations(path)
}

// readDeclarations keeps ordinary acquisition compatibility while allowing
// refresh to reject nonordinary documents and cap reads before allocation.
func (c Client) readDeclarations(path string) ([]byte, error) {
	if c.MaxDeclarationBytes <= 0 {
		return os.ReadFile(path)
	}
	st, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("dependency declaration %s must be an ordinary file", path)
	}
	if st.Size() > c.MaxDeclarationBytes {
		return nil, fmt.Errorf("dependency declaration %s exceeds byte limit %d", path, c.MaxDeclarationBytes)
	}
	// Avoid following a replacement symlink or blocking on a replacement FIFO.
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open dependency declaration %s: %w", path, err)
	}
	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	st, err = f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("dependency declaration %s must be an ordinary file", path)
	}
	if st.Size() > c.MaxDeclarationBytes {
		return nil, fmt.Errorf("dependency declaration %s exceeds byte limit %d", path, c.MaxDeclarationBytes)
	}
	content, err := io.ReadAll(io.LimitReader(f, c.MaxDeclarationBytes))
	if err != nil {
		return nil, err
	}
	var extra [1]byte
	n, err := f.Read(extra[:])
	if n > 0 {
		return nil, fmt.Errorf("dependency declaration %s exceeds byte limit %d", path, c.MaxDeclarationBytes)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return content, nil
}
