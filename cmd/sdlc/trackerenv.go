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
	"os/exec"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
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
func openTracker(ctx context.Context) (*trackerEnv, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	identity, err := workspace.Resolve(execGitRunner{}, ".", "")
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
func (e *trackerEnv) git(args ...string) (string, error) {
	cmd := exec.CommandContext(e.ctx, "git", args...)
	cmd.Dir = e.root
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}

func (e *trackerEnv) checkout(mainBase string) tracker.Checkout {
	branch := ""
	if e.branch != "" {
		branch = "refs/heads/" + e.branch
	}
	return tracker.Checkout{Root: e.root, Repository: e.target.Repository, Branch: branch, HEAD: e.head, MainBase: mainBase, ObjectFormat: e.format}
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
