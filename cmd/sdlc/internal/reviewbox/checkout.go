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
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/scratchdisk"
)

// Checkout is one review round's disposable clone. Its lifecycle (ARCH-FUNERAL):
// Create makes it, the round's dispatcher removes it with Close once the
// reviewer has exited, and Sweep reclaims it if its owner died first. Each id
// is absent → locked (lock file named and held) → checked out → removed.
type Checkout struct {
	Dir      string
	lock     *os.File
	src      string
	head     string
	overlays []string
}

const (
	lockSuffix    = ".lock"
	tmpLockSuffix = ".lock.tmp"
	// tmpLockGrace is how old an unheld .lock.tmp must be before Sweep calls its
	// creator dead. Create renames it within microseconds; a minute is margin.
	tmpLockGrace = time.Minute
)

// Roots says where review checkouts live. create is where a new one goes;
// all is every root that may hold one, for Sweep. WF_REVIEW_CHECKOUT_ROOT,
// when set, is the only root. Otherwise a new checkout goes to the TestScratch
// RAM disk's review-checkouts when it is mounted with room (#468), else to
// <user cache dir>/sdlc/review-checkouts. A checkout's Dir is fixed at Create
// and Reset reuses it, so a checkout lives wholly on the root it started on and
// never moves; all lists both roots so a sweep still reclaims either. note
// says why the RAM disk was passed over, when it is mounted but full or
// unreadable.
func Roots() (create string, all []string, note string, err error) {
	if r := os.Getenv("WF_REVIEW_CHECKOUT_ROOT"); r != "" {
		return r, []string{r}, "", nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", nil, "", fmt.Errorf("locate review checkout root: %w", err)
	}
	create, all, note = scratchdisk.Roots("review-checkouts", filepath.Join(cache, "sdlc", "review-checkouts"))
	return create, all, note, nil
}

// Create makes a fresh shared clone of src at head under root. The clone
// reads objects through alternates (no copy) and has its own refs, config,
// index and hooks; its origin remote is removed so nothing in it points back at
// src. It keeps src's main authority (#447), read-only: copyMainAuthority.
// overlays are repository-relative paths copied from src's working tree
// (the live issue and plan text, which may be uncommitted). On error nothing it
// made is left behind. Callers Sweep root first; Create itself never fails on
// another checkout's debris.
func Create(root, src, head string, overlays []string) (co *Checkout, err error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create review checkout root: %w", err)
	}
	// Canonical, so the directory a fence profile names is the one the kernel
	// sees (macOS temp and cache paths can sit behind /var → /private/var).
	if root, err = filepath.EvalSymlinks(root); err != nil {
		return nil, fmt.Errorf("resolve review checkout root: %w", err)
	}
	id, err := newID()
	if err != nil {
		return nil, err
	}
	// The published <id>.lock is held from the moment it is visible. The
	// <id>.lock.tmp it is renamed from is unheld for the instant between create
	// and flock; Sweep leaves any .lock.tmp younger than tmpLockGrace alone, so
	// it never reclaims one a live Create is still locking.
	tmp := filepath.Join(root, id+tmpLockSuffix)
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create review checkout lock: %w", err)
	}
	if err := flock.Exclusive(f, false); err != nil {
		f.Close()
		os.Remove(tmp)
		return nil, fmt.Errorf("lock review checkout: %w", err)
	}
	co = &Checkout{Dir: filepath.Join(root, id), lock: f, src: src, head: head, overlays: overlays}
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
	return co, co.populate()
}

// Reset returns the checkout to its pristine pinned state in place: same
// directory, same lock, a new clone. A retried attempt of one review uses it,
// so attempt 2 never reads attempt 1's mutations (#204 close review).
func (c *Checkout) Reset() error {
	if err := removeTree(c.Dir); err != nil {
		return fmt.Errorf("reset review checkout %s: %w", c.Dir, err)
	}
	return c.populate()
}

func (c *Checkout) populate() error {
	if err := git("", "clone", "--quiet", "--shared", "--no-checkout", c.src, c.Dir); err != nil {
		return err
	}
	if err := git(c.Dir, "remote", "remove", "origin"); err != nil {
		return err
	}
	if err := git(c.Dir, "checkout", "--quiet", "--detach", c.head); err != nil {
		return err
	}
	if err := c.copyMainAuthority(); err != nil {
		return err
	}
	for _, rel := range c.overlays {
		if err := overlay(c.src, c.Dir, rel); err != nil {
			return err
		}
	}
	return nil
}

// copyMainAuthority gives the checkout the source's main authority (#447):
// main's upstream config, each named remote's fetch refspec, and the main
// refs, so a repository test that resolves main and the merge base runs as it
// does in the slot. Each remote's URL is inert, so the checkout can neither
// push nor fetch. It runs after the detached checkout, so updating
// refs/heads/main never moves a checked-out branch.
func (c *Checkout) copyMainAuthority() error {
	config, err := gitOutput(c.src, "config", "--local", "--get-regexp", authorityKeys) // the slot's own, not the user's
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		config, err = "", nil // no such keys
	}
	if err != nil {
		return fmt.Errorf("review checkout: read main authority: %w", err)
	}
	refs, err := gitOutput(c.src, "for-each-ref", "--format=%(objectname) %(refname)", "refs/remotes/", "refs/heads/main")
	if err != nil {
		return fmt.Errorf("review checkout: read main authority: %w", err)
	}
	a, err := mainAuthority(config, refs)
	if err != nil {
		return err
	}
	for _, r := range a.remotes {
		for _, k := range []string{"url", "pushurl"} {
			if err := git(c.Dir, "config", "remote."+r+"."+k, inertURL); err != nil {
				return err
			}
		}
	}
	for _, kv := range a.config {
		if err := git(c.Dir, "config", "--add", kv[0], kv[1]); err != nil {
			return err
		}
	}
	for _, r := range a.refs {
		if err := git(c.Dir, "update-ref", r[0], r[1]); err != nil {
			return err
		}
	}
	return nil
}

// Close removes the checkout and releases its lock. Idempotent. When the
// directory cannot be removed, the lock FILE stays (only the flock is
// released), so the next Sweep sees an unheld lock and retries the removal.
func (c *Checkout) Close() error {
	if c == nil || c.lock == nil {
		return nil
	}
	err := removeTree(c.Dir)
	if err == nil {
		if rerr := os.Remove(c.lockPath()); rerr != nil && !errors.Is(rerr, fs.ErrNotExist) {
			err = rerr
		}
	}
	c.lock.Close() // closing the descriptor drops the flock
	c.lock = nil
	if err != nil {
		return fmt.Errorf("remove review checkout %s (the next review's sweep retries): %w", c.Dir, err)
	}
	return nil
}

func (c *Checkout) lockPath() string { return c.Dir + lockSuffix }

// Sweep removes every checkout under root whose owner is gone: a lock that can
// be taken is held by nobody, because the kernel drops a flock when its holder
// dies (no pid-reuse ambiguity). A directory without a lock is debris from a
// sweep or a create that was killed midway; an unheld <id>.lock.tmp is a create
// killed before it published its lock, once it is older than tmpLockGrace.
// Best effort: one entry that cannot be
// removed is reported and skipped, never a reason to refuse the next review.
func Sweep(root string) (removed int, problems []error) {
	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, []error{fmt.Errorf("sweep review checkouts: %w", err)}
	}
	present := map[string]bool{}
	for _, e := range entries {
		present[e.Name()] = true
	}
	for _, e := range entries {
		name := e.Name()
		var gone bool
		var err error
		switch {
		case strings.HasSuffix(name, tmpLockSuffix):
			if info, ierr := e.Info(); ierr != nil || time.Since(info.ModTime()) < tmpLockGrace {
				continue // possibly a Create between its create and its flock
			}
			gone, err = reclaim(root, strings.TrimSuffix(name, tmpLockSuffix), name)
		case strings.HasSuffix(name, lockSuffix):
			gone, err = reclaim(root, strings.TrimSuffix(name, lockSuffix), name)
		case e.IsDir() && !present[name+lockSuffix] && !present[name+tmpLockSuffix]:
			if err = removeTree(filepath.Join(root, name)); err == nil {
				gone = true
			}
		}
		if err != nil {
			problems = append(problems, fmt.Errorf("sweep review checkout %s: %w", name, err))
		}
		if gone {
			removed++
		}
	}
	return removed, problems
}

// reclaim removes id's checkout and its lock file when no live owner holds it.
func reclaim(root, id, lockName string) (bool, error) {
	path := filepath.Join(root, lockName)
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil // its owner closed it meanwhile
	}
	if err != nil {
		return false, err
	}
	defer f.Close()
	if err := flock.Exclusive(f, true); err != nil {
		return false, nil // held: a live review owns it
	}
	if err := removeTree(filepath.Join(root, id)); err != nil {
		return false, err // the lock stays, so a later sweep retries
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	return true, nil
}

// removeTree removes dir even where a reviewer left read-only directories
// (Go's module cache and some test fixtures do): it restores owner write
// permission on every directory first.
func removeTree(dir string) error {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			if info, ierr := d.Info(); ierr == nil && info.Mode().Perm()&0o700 != 0o700 {
				_ = os.Chmod(path, info.Mode().Perm()|0o700)
			}
		}
		return nil
	})
	return os.RemoveAll(dir)
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

// gitOutput runs git in dir and returns its trimmed stdout.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("name review checkout: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
