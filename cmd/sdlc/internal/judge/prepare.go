package judge

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewprepare"
)

// cappedOutput drains both streams while retaining a fixed diagnostic budget.
type cappedOutput struct {
	mu        sync.Mutex
	data      []byte
	truncated bool
}

func (b *cappedOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	left := reviewprepare.OutputLimit - len(b.data)
	if len(p) > left {
		p = p[:left]
		b.truncated = true
	}
	b.data = append(b.data, p...)
	return n, nil
}

func runPreparation(ctx context.Context, opts DispatchOptions, attempt int) (err error) {
	if opts.Preparation == nil {
		return nil
	}
	cfg := opts.Preparation
	start := time.Now()
	result := reviewprepare.Result{Head: opts.PreparationHead, Command: append([]string(nil), cfg.Command...), Attempt: attempt}
	capture := &cappedOutput{}
	defer func() {
		result.Duration = time.Since(start)
		result.Output = string(capture.data)
		result.Truncated = capture.truncated
		if err != nil {
			result.Error = err.Error()
		}
		if opts.Evidence != nil {
			opts.Evidence.Preparation = append(opts.Evidence.Preparation, result)
		}
	}()
	if opts.Checkout == "" || len(cfg.Command) == 0 {
		return fmt.Errorf("review preparation requires a disposable checkout and command")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutMinutes)*time.Minute)
	defer cancel()
	before, err := trackedPreparationState(ctx, opts.Checkout)
	if err != nil {
		return err
	}
	name, args := cfg.Command[0], cfg.Command[1:]
	if opts.Wrap != nil {
		name, args = opts.Wrap(name, args)
	}
	if opts.Stderr != nil {
		fmt.Fprintf(opts.Stderr, "    … preparing review checkout (attempt %d): %q\n", attempt, cfg.Command)
	}
	if err = runCommand(ctx, nil, opts.Checkout, capture, capture, name, args...); err != nil {
		return fmt.Errorf("review preparation attempt %d: %w", attempt, err)
	}
	after, err := trackedPreparationState(ctx, opts.Checkout)
	if err != nil {
		return err
	}
	if !bytes.Equal(before, after) {
		return fmt.Errorf("review preparation changed tracked checkout state; generate ignored or untracked assets only")
	}
	return nil
}

// Keep HEAD and staged objects, but fingerprint actual filesystem bytes and
// modes independently of Git's normalization, text conversion and index flags.
// Existing dirty issue overlays are the baseline, not a reason to refuse.
func trackedPreparationState(ctx context.Context, dir string) ([]byte, error) {
	state, err := preparationRepositoryState(ctx, dir, make(map[string]bool))
	if err != nil {
		return nil, fmt.Errorf("inspect prepared tracked tree: %w", err)
	}
	return state, nil
}

func preparationRepositoryState(ctx context.Context, dir string, seen map[string]bool) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	canonical, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return nil, err
	}
	if seen[canonical] {
		return nil, fmt.Errorf("recursive tracked repository %s", canonical)
	}
	seen[canonical] = true
	root, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	git := func(args ...string) ([]byte, error) {
		out, err := exec.CommandContext(ctx, "git", append([]string{"-C", canonical}, args...)...).Output()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return out, err
	}
	// In particular, a gitlink must never fall back to its parent's repository.
	top, err := git("rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	actual, err := filepath.EvalSymlinks(strings.TrimSpace(string(top)))
	if err != nil {
		return nil, err
	}
	if actual != canonical {
		return nil, fmt.Errorf("tracked repository %s resolves to %s", canonical, actual)
	}
	head, err := git("rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	index, err := git("ls-files", "--stage", "-z")
	if err != nil {
		return nil, err
	}
	state := sha256.New()
	preparationField(state, head)
	preparationField(state, index)
	// A parent symlink must be recorded without following it to a same-content
	// replacement. Cache each component; untracked siblings are never enumerated.
	components := map[string]os.FileMode{}
	for _, entry := range bytes.Split(index, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		header, path, ok := bytes.Cut(entry, []byte{'\t'})
		fields := bytes.Fields(header)
		if !ok || len(fields) != 3 || !filepath.IsLocal(string(path)) {
			return nil, fmt.Errorf("invalid tracked index entry")
		}
		parts := strings.Split(string(path), "/")
		for i := range parts {
			rel := filepath.Join(parts[:i+1]...)
			mode, known := components[rel]
			if !known {
				mode, err = preparationPathState(ctx, root, rel, state)
				if err != nil {
					return nil, err
				}
				components[rel] = mode
			}
			if i < len(parts)-1 {
				if mode&os.ModeSymlink != 0 {
					return nil, fmt.Errorf("tracked path %s has symlink ancestor %s", path, rel)
				}
				if !mode.IsDir() {
					break
				}
				continue
			}
			if string(fields[0]) == "160000" {
				if mode == os.ModeIrregular {
					continue
				} // absent, uninitialized gitlink
				if !mode.IsDir() {
					return nil, fmt.Errorf("gitlink %s is not a directory", rel)
				}
				module, err := root.OpenRoot(rel)
				if err != nil {
					return nil, err
				}
				_, metaErr := module.Lstat(".git")
				if metaErr == nil {
					module.Close()
					nested, err := preparationRepositoryState(ctx, filepath.Join(canonical, rel), seen)
					if err != nil {
						return nil, err
					}
					preparationField(state, nested)
				} else {
					// Empty directories are the other uninitialized representation. Refuse
					// populated non-repositories rather than ignore ambiguous source bytes.
					d, err := module.Open(".")
					module.Close()
					if err != nil {
						return nil, err
					}
					entries, readErr := d.ReadDir(1)
					d.Close()
					if !errors.Is(metaErr, os.ErrNotExist) || len(entries) != 0 || (readErr != nil && !errors.Is(readErr, io.EOF)) {
						return nil, fmt.Errorf("gitlink %s is populated without its own repository", rel)
					}
				}
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return state.Sum(nil), nil
}

// Every field is length-framed so arbitrary filenames and contents are
// unambiguous. The state is fixed-size even when tracked files are very large.
func preparationField(state hash.Hash, value []byte) {
	var size [8]byte
	binary.LittleEndian.PutUint64(size[:], uint64(len(value)))
	state.Write(size[:])
	state.Write(value)
}

// ModeIrregular denotes an absent component. Actual irregular/special objects
// are fingerprinted by mode only and are never opened (in particular FIFOs).
func preparationPathState(ctx context.Context, root *os.Root, path string, state hash.Hash) (os.FileMode, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	preparationField(state, []byte(path))
	info, err := root.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		preparationField(state, []byte("absent"))
		return os.ModeIrregular, nil
	}
	if err != nil {
		return 0, err
	}
	mode := info.Mode()
	var bits [4]byte
	binary.LittleEndian.PutUint32(bits[:], uint32(mode))
	preparationField(state, bits[:])
	switch {
	case mode.IsRegular():
		f, err := root.Open(path)
		if err != nil {
			return 0, err
		}
		opened, err := f.Stat()
		if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
			f.Close()
			return 0, fmt.Errorf("tracked file changed while opening %s", path)
		}
		content := sha256.New()
		err = copyPreparationContents(ctx, content, f)
		closeErr := f.Close()
		if err != nil {
			return 0, err
		}
		if closeErr != nil {
			return 0, closeErr
		}
		preparationField(state, content.Sum(nil))
	case mode&os.ModeSymlink != 0:
		target, err := root.Readlink(path)
		if err != nil {
			return 0, err
		}
		preparationField(state, []byte(target))
	}
	return mode, nil
}

func copyPreparationContents(ctx context.Context, dst io.Writer, src io.Reader) error {
	var buf [32 << 10]byte
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := src.Read(buf[:])
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if errors.Is(err, io.EOF) {
			return ctx.Err()
		}
		if err != nil {
			return err
		}
	}
}
