// move.go — `sdlc move [:N]` moves this slot's issue branch into slot :N
// (default :0) and returns this slot to its resting branch (#260). The rules
// live in checkMove (moveplan.go); this file observes and switches.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"io"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func NewMoveCmd() *cobra.Command {
	var dryRun bool
	cmd := markMutatingCommand(&cobra.Command{
		Use: "move [address]", Args: cobra.MaximumNArgs(1),
		SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			address := ""
			if len(args) > 0 {
				address = args[0]
			}
			return runMove(commandContext(cmd.Context()), ".", address, dryRun, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	})
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check both slots and print the move; change nothing")
	return cmd
}

func runMove(ctx context.Context, dir, address string, dryRun bool, stdout, stderr io.Writer) error {
	if address == "" {
		address = ":0"
	}
	facts, err := observeMove(dir, address)
	if err != nil {
		return err
	}
	if err := checkMove(facts); err != nil {
		return err
	}
	from, to, branch := facts.From, facts.To, facts.From.Branch
	fmt.Fprintf(stdout, "Move %s from %s to %s\n", branch, from.Address, to.Address)
	moveReport(stderr, fmt.Sprintf("%s commits not in %s (they stay on %s, but are not in the test)", to.Resting, branch, to.Resting), facts.Parked)
	moveReport(stderr, fmt.Sprintf("%s commits not on its upstream", to.Resting), facts.Unpublished)
	if dryRun {
		if _, _, ok := issue.ParseFilename(branch + ".md"); ok {
			cinfo(stderr, fmt.Sprintf("would then record %s as the issue's owner, if this workspace owns it (#277)", to.Address))
		}
		cinfo(stderr, "dry-run — nothing was switched")
		return nil
	}
	again, err := observeMove(dir, address)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(facts, again) {
		return fmt.Errorf("%s or %s changed since the preflight; nothing was switched, rerun sdlc move", from.Address, to.Address)
	}
	r := execGitRunner{}
	if out, err := r.GitInDir(from.Root, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", from.Resting); err != nil {
		return fmt.Errorf("switch %s to %s: %v\n%s\nnothing was moved", from.Address, from.Resting, err, out)
	}
	if out, err := r.GitInDir(to.Root, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", branch); err != nil {
		return fmt.Errorf("switch %s to %s: %v\n%s\n%s is intact at %s and %s is on %s; reconcile %s, then retry: git -C %s switch %s",
			to.Address, branch, err, out, branch, from.Head, from.Address, from.Resting, to.Address, to.Root, branch)
	}
	for _, check := range []struct{ root, branch, head string }{{to.Root, branch, from.Head}, {from.Root, from.Resting, from.RestHead}} {
		id, err := workspace.Resolve(r, check.root, "")
		if err != nil {
			return fmt.Errorf("both switches ran, but verifying %s failed: %w", check.root, err)
		}
		if got, head := workspaceText(id.Branch, "(detached)"), workspaceText(id.Head, ""); got != check.branch || head != check.head {
			return fmt.Errorf("both switches ran, but %s is on %s at %s, want %s at %s; inspect it before continuing", check.root, got, head, check.branch, check.head)
		}
	}
	cok(stderr, fmt.Sprintf("%s is on %s in %s; %s is back on %s", branch, to.Address, to.Root, from.Address, from.Resting))
	relocateAfterMove(ctx, to.Root, branch, stderr)
	return nil
}

// relocateAfterMove records the destination as the owner of the issue the moved
// branch carries (#277). It runs only after both switches are verified, so its
// network step can never strand the branch: a failure leaves the move complete,
// the claimant on the source, and names the convergent repair.
func relocateAfterMove(ctx context.Context, dest, branch string, stderr io.Writer) {
	id, _, ok := issue.ParseFilename(branch + ".md")
	if !ok {
		return // not an issue branch
	}
	if err := moveRelocation(ctx, dest, id); err != nil {
		cwarn(stderr, fmt.Sprintf("%s moved, but #%s's owner was not updated: %v\n      finish it with `sdlc claim --issue %s` in %s", branch, id, err, issue.CLIRef(id), dest))
	}
}

// moveRelocation is relocateAfterMove's effect, a variable so a test can fail
// it after the switches.
var moveRelocation = func(ctx context.Context, dest, id string) error {
	if _, err := os.Stat(filepath.Join(dest, filepath.FromSlash(tracker.CutoverMarkerPath))); errors.Is(err, os.ErrNotExist) {
		return nil // a pre-tracker repository has no owners to move
	}
	env, err := openTrackerAt(ctx, dest)
	if err != nil {
		return err
	}
	if tracked, _, err := env.repo.Presence(); err != nil || !tracked {
		return err // a pre-tracker repository has no owners to move
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, ok := snap.Card(id)
	if !ok {
		return nil
	}
	own, recorded, me, err := ownership(env, card)
	if err != nil || own == issue.OwnershipMine {
		return err
	}
	if own == issue.OwnershipUnknown {
		return fmt.Errorf("it has no recorded owner; `sdlc claim --issue %s --adopt` here records one", issue.CLIRef(id))
	}
	allowed, err := relocatable(env, card, recorded, me)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("it is owned by %s, not the slot it moved from — moving never takes ownership (reclaim is #278)", describeClaimant(recorded))
	}
	return relocateClaimant(env, card, me)
}

func moveReport(w io.Writer, title string, commits []string) {
	if len(commits) > 0 {
		cinfo(w, fmt.Sprintf("%s:\n  %s", title, strings.Join(commits, "\n  ")))
	}
}

// observeMove reads everything checkMove needs from both slots.
func observeMove(dir, address string) (moveFacts, error) {
	r := execGitRunner{}
	fromID, err := workspace.Resolve(r, dir, "")
	if err != nil {
		return moveFacts{}, err
	}
	toID, err := workspace.Resolve(r, dir, address)
	if err != nil {
		return moveFacts{}, fmt.Errorf("resolve %s: %w", address, err)
	}
	f := moveFacts{SameRepo: fromID.RepoIdentity == toID.RepoIdentity}
	if f.From, err = observeMoveSide(r, fromID); err != nil {
		return moveFacts{}, err
	}
	if f.To, err = observeMoveSide(r, toID); err != nil {
		return moveFacts{}, err
	}
	if f.From.Address == "" || f.To.Resting == "" {
		return moveFacts{}, fmt.Errorf("sdlc move runs between durable slots; %s or %s is not one", f.From.Root, f.To.Root)
	}
	git := func(args ...string) (string, error) {
		out, err := r.GitInDir(f.To.Root, args...)
		if err != nil {
			return "", fmt.Errorf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimRight(string(out), "\n"), nil
	}
	branch, rest := f.From.Branch, f.To.Resting
	if branch == "" || branch == f.From.Resting {
		return f, nil // checkMove refuses; the ranges below need an issue branch
	}
	tree, err := git("ls-tree", "-r", "-z", "--name-only", branch)
	if err != nil {
		return moveFacts{}, err
	}
	f.Incoming = splitNonEmpty(tree, "\x00")
	upstream, err := git("rev-parse", "--abbrev-ref", "--symbolic-full-name", rest+"@{upstream}")
	if err != nil {
		return moveFacts{}, fmt.Errorf("%s has no configured upstream: %w", rest, err)
	}
	for _, rng := range []struct {
		dst  *[]string
		base string
	}{{&f.Parked, branch}, {&f.Unpublished, upstream}} {
		out, err := git("log", "--format=%h %s", rng.base+".."+rest)
		if err != nil {
			return moveFacts{}, err
		}
		*rng.dst = splitNonEmpty(out, "\n")
	}
	return f, nil
}

func observeMoveSide(r execGitRunner, id workspace.Identity) (moveSide, error) {
	s := moveSide{Root: id.WorktreeRoot, Address: workspaceText(id.Address, ""), Branch: workspaceText(id.Branch, ""),
		Resting: workspaceText(id.RestingBranch, ""), Head: workspaceText(id.Head, "")}
	if s.Resting != "" {
		rest, err := r.GitInDir(s.Root, "rev-parse", "--verify", "refs/heads/"+s.Resting)
		if err != nil {
			return moveSide{}, fmt.Errorf("resolve %s in %s: %v\n%s", s.Resting, s.Root, err, rest)
		}
		s.RestHead = strings.TrimSpace(string(rest))
	}
	out, err := r.GitInDir(s.Root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return moveSide{}, fmt.Errorf("git status in %s: %v\n%s", s.Root, err, out)
	}
	entries, err := gitx.ParseStatusZ(out)
	if err != nil {
		return moveSide{}, err
	}
	for _, e := range entries {
		if e.XY == "??" {
			s.Untracked = append(s.Untracked, e.Path)
		} else {
			s.Changes = append(s.Changes, e)
		}
	}
	s.Operation, err = gitOperationInProgress(func(args ...string) (string, error) {
		out, err := r.GitInDir(s.Root, args...)
		return strings.TrimSpace(string(out)), err
	}, s.Root)
	return s, err
}

func splitNonEmpty(s, sep string) []string {
	var out []string
	for _, p := range strings.Split(s, sep) {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
