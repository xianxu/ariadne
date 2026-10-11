package reviewbox

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Checkout is one review round's disposable clone. Its lifecycle (ARCH-FUNERAL):
// Create makes it, the round's dispatcher removes it with Close once the
// reviewer has exited, and Sweep reclaims it if its owner died first. Each id
// is absent → locked (lock file named and held) → checked out → removed.
type Checkout struct {
	Dir  string
	lock *os.File
}

const lockSuffix = ".lock"

// Root is where review checkouts live: WF_REVIEW_CHECKOUT_ROOT, else
// <user cache dir>/sdlc/review-checkouts.
func Root() (string, error) {
	if r := os.Getenv("WF_REVIEW_CHECKOUT_ROOT"); r != "" {
		return r, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locate review checkout root: %w", err)
	}
	return filepath.Join(cache, "sdlc", "review-checkouts"), nil
}

// Create sweeps root, then makes a fresh shared clone of src at head under it.
// The clone reads objects through alternates (no copy) and has its own refs,
// config, index and hooks; its origin remote is removed so nothing in it points
// back at src. overlays are repository-relative paths copied from src's
// working tree (the live issue and plan text, which may be uncommitted).
// On error nothing it made is left behind.
func Create(root, src, head string, overlays []string) (co *Checkout, err error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create review checkout root: %w", err)
	}
	if _, err := Sweep(root); err != nil {
		return nil, err
	}
	id, err := newID()
	if err != nil {
		return nil, err
	}
	// Hold the lock before its name is visible, so a concurrent sweep can never
	// see an unheld lock for a live checkout.
	tmp := filepath.Join(root, id+lockSuffix+".tmp")
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create review checkout lock: %w", err)
	}
	if err := lockExclusive(f, false); err != nil {
		f.Close()
		os.Remove(tmp)
		return nil, fmt.Errorf("lock review checkout: %w", err)
	}
	co = &Checkout{Dir: filepath.Join(root, id), lock: f}
	if err := os.Rename(tmp, co.lockPath()); err != nil {
		f.Close()
		os.Remove(tmp)
		return nil, fmt.Errorf("publish review checkout lock: %w", err)
	}
	defer func() {
		if err != nil {
			co.Close()
			co = nil
		}
	}()

	if err := git("", "clone", "--quiet", "--shared", "--no-checkout", src, co.Dir); err != nil {
		return co, err
	}
	if err := git(co.Dir, "remote", "remove", "origin"); err != nil {
		return co, err
	}
	if err := git(co.Dir, "checkout", "--quiet", "--detach", head); err != nil {
		return co, err
	}
	for _, rel := range overlays {
		if err := overlay(src, co.Dir, rel); err != nil {
			return co, err
		}
	}
	return co, nil
}

// Close removes the checkout and releases its lock. Idempotent.
func (c *Checkout) Close() error {
	if c == nil || c.lock == nil {
		return nil
	}
	err := os.RemoveAll(c.Dir)
	if rerr := os.Remove(c.lockPath()); rerr != nil && !errors.Is(rerr, fs.ErrNotExist) && err == nil {
		err = rerr
	}
	c.lock.Close() // closing the descriptor drops the flock
	c.lock = nil
	if err != nil {
		return fmt.Errorf("remove review checkout %s: %w", c.Dir, err)
	}
	return nil
}

func (c *Checkout) lockPath() string { return c.Dir + lockSuffix }

// Sweep removes every checkout under root whose owner is gone: a lock that can
// be taken is held by nobody, because the kernel drops a flock when its holder
// dies (no pid-reuse ambiguity). A directory without a lock is debris from a
// sweep or a create that was killed midway. Returns how many it removed.
func Sweep(root string) (int, error) {
	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("sweep review checkouts: %w", err)
	}
	present := map[string]bool{}
	for _, e := range entries {
		present[e.Name()] = true
	}
	removed := 0
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasSuffix(name, lockSuffix):
			id := strings.TrimSuffix(name, lockSuffix)
			if reclaimed, err := reclaim(root, id); err != nil {
				return removed, err
			} else if reclaimed {
				removed++
			}
		case e.IsDir() && !present[name+lockSuffix]:
			if err := os.RemoveAll(filepath.Join(root, name)); err != nil {
				return removed, fmt.Errorf("sweep review checkout %s: %w", name, err)
			}
			removed++
		}
		// <id>.lock.tmp is a create in progress (held) or a create killed
		// before rename; the latter is tiny and its directory never existed.
	}
	return removed, nil
}

// reclaim removes id's checkout and lock when no live owner holds the lock.
func reclaim(root, id string) (bool, error) {
	path := filepath.Join(root, id+lockSuffix)
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil // its owner closed it meanwhile
	}
	if err != nil {
		return false, fmt.Errorf("sweep review checkout %s: %w", id, err)
	}
	defer f.Close()
	if err := lockExclusive(f, true); err != nil {
		return false, nil // held: a live review owns it
	}
	if err := os.RemoveAll(filepath.Join(root, id)); err != nil {
		return false, fmt.Errorf("sweep review checkout %s: %w", id, err)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, fmt.Errorf("sweep review checkout lock %s: %w", id, err)
	}
	return true, nil
}

func overlay(src, dst, rel string) error {
	clean := filepath.Clean(rel)
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("review checkout overlay must be repository-relative: %q", rel)
	}
	data, err := os.ReadFile(filepath.Join(src, clean))
	if errors.Is(err, fs.ErrNotExist) {
		return nil // nothing live to overlay; the checkout keeps Head's copy, if any
	}
	if err != nil {
		return fmt.Errorf("read review overlay %s: %w", rel, err)
	}
	target := filepath.Join(dst, clean)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("write review overlay %s: %w", rel, err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("write review overlay %s: %w", rel, err)
	}
	return nil
}

func git(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("review checkout: git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("name review checkout: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
