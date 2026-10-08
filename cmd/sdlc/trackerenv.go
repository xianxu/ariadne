// trackerenv.go — the pinned checkout/publication context every tracker verb
// shares (#252): card authority lives on the issue-tracker branch of the
// configured publication remote, details live in this checkout.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type trackerEnv struct {
	ctx     context.Context
	root    string
	target  gitx.PublicationTarget
	repo    *tracker.Repository
	main    *gitx.TrunkFile
	format  string
	resting string // this checkout's resting branch
	branch  string // current branch; "" when detached
	head    string
}

// onRest reports whether the checkout sits on its resting branch.
func (e *trackerEnv) onRest() bool { return e.branch != "" && e.branch == e.resting }

// openTracker pins the checkout and its publication target. It performs no
// network IO; the first snapshot or candidate fetches.
func openTracker(ctx context.Context) (*trackerEnv, error) { return openTrackerAt(ctx, ".") }

// staleTrackerNote labels a read-only view built from the last-fetched
// tracker because the fresh fetch failed (#252): stale, never silent.
const staleTrackerNote = "issue tracker unreachable: card statuses are stale (as last fetched)"

// openTrackerAt pins the checkout containing dir (another repository's, for
// readers that inspect a dependency chain).
func openTrackerAt(ctx context.Context, dir string) (*trackerEnv, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	identity, err := workspace.Resolve(execGitRunner{}, dir, "")
	if err != nil {
		return nil, fmt.Errorf("resolve checkout: %w", err)
	}
	e := &trackerEnv{ctx: ctx, root: identity.WorktreeRoot, resting: "main"}
	if identity.RestingBranch != nil {
		e.resting = *identity.RestingBranch
	}
	if identity.Branch != nil {
		e.branch = *identity.Branch
	}
	if e.target, err = gitx.ResolvePublicationTarget(ctx, e.root, e.resting); err != nil {
		return nil, fmt.Errorf("%w\n      tracker verbs publish through %s's upstream; configure it as <remote>/main", err, e.resting)
	}
	if e.repo, err = tracker.NewRepository(ctx, e.root, e.target.Remote); err != nil {
		return nil, err
	}
	e.repo.GuardCutover(e.root)
	if e.main, err = gitx.NewTrunkFileContext(ctx, e.root, e.target.Remote, "main"); err != nil {
		return nil, err
	}
	if e.format, err = gitx.ObjectFormat(ctx, e.root); err != nil {
		return nil, err
	}
	if e.head, err = e.git("rev-parse", "--verify", "HEAD^{commit}"); err != nil {
		return nil, errors.New("tracker verbs need a checkout with a commit")
	}
	return e, nil
}

// git runs one command in the checkout and returns trimmed stdout — also on a
// non-zero exit, where some commands (merge-tree) still report a result. Stderr
// is kept out of the value so a warning can never be parsed as an object ID.
func (e *trackerEnv) git(args ...string) (string, error) { return e.gitEnv(nil, args...) }

// statusEntries runs `git -C dir status --porcelain=v1 -z <args>` and parses it
// byte-exact. Status output is column-structured: git() trims, which eats the
// first entry's leading status space and shifts its path (#259).
func (e *trackerEnv) statusEntries(dir string, args ...string) ([]gitx.StatusEntry, error) {
	out, err := e.gitRaw(nil, append([]string{"-C", dir, "status", "--porcelain=v1", "-z"}, args...)...)
	if err != nil {
		return nil, err
	}
	return gitx.ParseStatusZ(out)
}

// gitEnv is git with extra environment (a temporary index, for instance).
// Its output is trimmed, which suits single values (an OID, a ref, a count);
// column-structured output goes through gitRaw.
func (e *trackerEnv) gitEnv(extra []string, args ...string) (string, error) {
	out, err := e.gitRaw(extra, args...)
	return strings.TrimSpace(string(out)), err
}

// gitRaw runs git in the checkout and returns its stdout untouched; a failure
// carries git's stderr.
func (e *trackerEnv) gitRaw(extra []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(e.ctx, "git", args...)
	// Cancellation ends git and anything it spawned (a hook, a transport
	// helper), as gitx's trunk writes do: a bounded push must not wait on a
	// descendant still holding the output pipe (#286).
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	cmd.Dir = e.root
	if len(extra) > 0 {
		cmd.Env = append(os.Environ(), extra...)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// gitTest runs a Git predicate: exit 0 is true, exit 1 is a completed false
// observation, and anything else (a missing repository, a killed process) is
// an error — never evidence of absence or difference. What exit 1 means is the
// command's: `rev-parse -q --verify` also exits 1 for an unknown commit, so
// callers pass commits they have already resolved.
func (e *trackerEnv) gitTest(args ...string) (bool, error) {
	_, err := e.git(args...)
	if err == nil {
		return true, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// has reports whether a path exists at a commit (rev-parse exits 1 when not).
func (e *trackerEnv) has(commit, p string) (bool, error) {
	return e.gitTest("rev-parse", "-q", "--verify", commit+":"+p)
}

// ancestorOf reports whether commit a precedes (or equals) commit b.
func (e *trackerEnv) ancestorOf(a, b string) (bool, error) {
	return e.gitTest("merge-base", "--is-ancestor", a, b)
}

// closeAncestorOf is ancestorOf for judging close generations (#283): a
// reviewed commit this clone does not have — rewritten away by a rebase
// elsewhere, never fetched — precedes nothing here, rather than failing the
// close. Other callers keep ancestorOf's error for an unknown commit.
func (e *trackerEnv) closeAncestorOf(a, b string) (bool, error) {
	if known, err := e.gitTest("rev-parse", "-q", "--verify", a+"^{commit}"); err != nil || !known {
		return false, err
	}
	return e.ancestorOf(a, b)
}

// branchRef is the checkout's current branch as a full ref ("" when detached).
func (e *trackerEnv) branchRef() string {
	if e.branch == "" {
		return ""
	}
	return "refs/heads/" + e.branch
}

func (e *trackerEnv) checkout(mainBase string) tracker.Checkout {
	return tracker.Checkout{Root: e.root, Repository: e.target.Repository, Branch: e.branchRef(), HEAD: e.head, MainBase: mainBase, ObjectFormat: e.format}
}

func (e *trackerEnv) receipts() (*tracker.RecoveryReceipts, error) {
	store, err := gitx.NewRecoveryStore(e.ctx, e.root)
	if err != nil {
		return nil, err
	}
	return tracker.NewRecoveryReceipts(store, e.target.Repository), nil
}

// operationToken names one operation run: verb, issue and a random suffix, so
// its candidate commits prove ownership even against identical peer content.
func operationToken(verb string) string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // crypto/rand failure is unrecoverable
	}
	return verb + "-" + hex.EncodeToString(b[:])
}

// commitOnly records exactly these paths on the current branch, leaving any
// other staged or unstaged work untouched (an ordinary local checkpoint; it
// never publishes).
func commitOnly(e *trackerEnv, message string, paths ...string) error {
	if _, err := e.git(append([]string{"add", "--"}, paths...)...); err != nil {
		return err
	}
	_, err := e.git(append([]string{"commit", "-q", "--no-verify", "-m", message, "--only", "--"}, paths...)...)
	return err
}

// commandContext is the verb's Cobra context, or Background for direct callers.
func commandContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
