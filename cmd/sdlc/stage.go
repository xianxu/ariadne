// stage.go — `sdlc stage [:N]` puts this slot's HEAD on slot :N (default :0)
// as a detached commit for a smoke test (#344); the slot keeps its branch. The
// rules live in checkStage (stageplan.go); this file observes and switches.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func NewStageCmd() *cobra.Command {
	var build, clear, dryRun bool
	var abandoned string
	cmd := markMutatingCommand(&cobra.Command{
		Use: "stage [address]", Args: cobra.MaximumNArgs(1),
		SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			address := ":0"
			if len(args) > 0 {
				address = args[0]
			}
			if clear {
				if build {
					return errors.New("--clear and --build do not combine")
				}
				return runUnstage(commandContext(cmd.Context()), ".", address, abandoned, dryRun, cmd.OutOrStdout(), cmd.ErrOrStderr())
			}
			if abandoned != "" {
				return errors.New("--abandoned goes with --clear")
			}
			return runStage(commandContext(cmd.Context()), ".", address, build, dryRun, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	})
	cmd.Flags().BoolVar(&build, "build", false, "build and verify the target's Go binaries against the staged commit")
	cmd.Flags().BoolVar(&clear, "clear", false, "return the target to its resting branch, then rebuild and verify its Go binaries")
	cmd.Flags().StringVar(&abandoned, "abandoned", "", "with --clear: free the stage lease HOLDER left behind (its issue, as the refusal names it)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check both slots and print the stage; change nothing")
	return cmd
}

func runStage(ctx context.Context, dir, address string, build, dryRun bool, stdout, stderr io.Writer) error {
	facts, err := observeStage(dir, address)
	if err != nil {
		return err
	}
	if err := checkStage(facts); err != nil {
		return err
	}
	from, to, commit := facts.From, facts.To, facts.From.Head
	r := execGitRunner{}
	held, unlock, err := takeStageLease(r, to, stageHolder(from))
	if err != nil {
		return err
	}
	defer func() { unlock() }()
	assets, err := slotAssetEnvironment(ctx, to.Root)
	if err != nil {
		return err
	}
	issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
	if err := verifyAssetTransition(assets, commit, issuesDir); err != nil {
		return err
	}
	if n := len(from.Changes) + len(from.Untracked); n > 0 {
		cwarn(stderr, fmt.Sprintf("%s has %d uncommitted or untracked path(s); only its HEAD %s is staged", from.Address, n, shortOID(commit)))
	}
	if dryRun {
		fmt.Fprintf(stdout, "Would stage %s (%s) on %s\n", shortOID(commit), from.Branch, to.Address)
		cinfo(stderr, "dry-run — nothing was switched")
		return nil
	}
	again, err := observeStage(dir, address)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(facts, again) {
		return fmt.Errorf("%s or %s changed since the preflight; nothing was switched, rerun sdlc stage", from.Address, to.Address)
	}
	// The holder is recorded before the switch (#421), keeping the staged
	// commit until the switch is verified, so a crash after the switch still
	// leaves the target held.
	lease := fleet.StageLease{Commit: held.Commit, Holder: stageHolder(from), Branch: from.Branch, Slot: from.Address, At: time.Now()}
	if err := fleet.WriteStageLease(r, to.Root, lease); err != nil {
		return fmt.Errorf("write the stage lease in %s: %v; nothing was switched", to.Address, err)
	}
	if to.Head != commit || to.Branch != "" {
		if out, err := r.GitInDir(to.Root, "-c", "submodule.recurse=false", "switch", "--detach", "--no-overwrite-ignore", commit); err != nil {
			// A failed switch leaves the target's HEAD where it was (git may
			// have updated its index and tree first), so it gives back the
			// lease it held before; a refusal never leaves it held.
			if rerr := restoreStageLease(r, to.Root, held); rerr != nil {
				return fmt.Errorf("switch %s to %s: %v\n%s\nHEAD did not move (check its tree with git status), and restoring its stage lease failed: %v; run sdlc stage --clear %s", to.Address, shortOID(commit), err, out, rerr, to.Address)
			}
			return fmt.Errorf("switch %s to %s: %v\n%s\nnothing was staged: HEAD did not move (check its tree with git status)", to.Address, shortOID(commit), err, out)
		}
	}
	if err := verifySlot(r, to.Root, "", commit); err != nil {
		return err
	}
	if err := verifyMaterializedAssets(to.Root, assets.main.ViewOf(commit), issuesDir); err != nil {
		return err
	}
	// The commit is recorded only after the verified switch: a crash before it
	// leaves :N at a commit the slot's branch contains, which is leavable anyway.
	lease.Commit = commit
	if err := fleet.WriteStageLease(r, to.Root, lease); err != nil {
		cwarn(stderr, fmt.Sprintf("staged, but the record was not written (a later stage or --clear may refuse once %s is in no branch): %v", shortOID(commit), err))
	}
	switch {
	case to.Branch == "" && to.Head == commit:
		cok(stderr, fmt.Sprintf("%s already staged on %s", shortOID(commit), to.Address))
	case to.Branch == "":
		cok(stderr, fmt.Sprintf("replaced the staged %s on %s", shortOID(to.Head), to.Address))
	}
	fmt.Fprintf(stdout, "Staged %s (%s from %s) on %s in %s\n", commit, from.Branch, from.Address, to.Address, to.Root)
	cinfo(stderr, fmt.Sprintf("%s keeps %s; `sdlc stage --clear %s` returns %s to %s", from.Address, orDetached(from.Branch), to.Address, to.Address, to.Resting))
	if build {
		unlock()
		unlock = func() {}
		return stageBuild(ctx, to.Root, commit, stdout, stderr)
	}
	return nil
}

func runUnstage(ctx context.Context, dir, address, abandoned string, dryRun bool, stdout, stderr io.Writer) error {
	r := execGitRunner{}
	me, err := workspace.Resolve(r, dir, "")
	if err != nil {
		return err
	}
	id, err := workspace.Resolve(r, dir, address)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", address, err)
	}
	// The target is observed under the stage lock, so what --clear decides on
	// is what it changes.
	unlock, err := fleet.LockStage(r, id.WorktreeRoot)
	if err != nil {
		return err
	}
	defer unlock()
	if id, err = workspace.Resolve(r, dir, address); err != nil {
		return fmt.Errorf("resolve %s: %w", address, err)
	}
	to, err := observeMoveSide(r, id)
	if err != nil {
		return err
	}
	if to.Resting == "" {
		return fmt.Errorf("sdlc stage runs between durable slots; %s is not one", to.Root)
	}
	staged, err := checkUnstage(to)
	if err != nil {
		return err
	}
	held, err := readStageLease(r, to)
	if err != nil {
		return err
	}
	if err := checkClearLease(held, stageHolder(moveSide{Branch: workspaceText(me.Branch, ""), Address: workspaceText(me.Address, "")}), abandoned, to.Address); err != nil {
		return err
	}
	assets, err := slotAssetEnvironment(ctx, to.Root)
	if err != nil {
		return err
	}
	issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
	if !staged {
		if err := checkTargetClean(to); err != nil {
			return err
		}
		if err := verifyMaterializedAssets(to.Root, assets.main.ViewOf(to.Head), issuesDir); err != nil {
			return err
		}
		if dryRun && held.Holder != "" {
			fmt.Fprintf(stdout, "Would release %s's stage lease on %s (already on its resting branch %s)\n", held.Holder, to.Address, to.Resting)
			return nil
		}
		if !dryRun {
			if held.Holder != "" {
				removeStaged(r, to.Root)
			}
			if err := rebuildRestingStage(ctx, to.Root, to.Head, stdout, stderr); err != nil {
				return fmt.Errorf("%s is on %s at %s, but binary rebuild failed: %w; rerun sdlc stage --clear %s", to.Address, to.Resting, shortOID(to.Head), err, to.Address)
			}
			if held.Holder != "" {
				cok(stderr, fmt.Sprintf("%s is on its resting branch %s; released %s's stage lease", to.Address, to.Resting, held.Holder))
				return nil
			}
		}
		cok(stderr, fmt.Sprintf("%s is on its resting branch %s; nothing staged", to.Address, to.Resting))
		return nil
	}
	if err := verifyAssetTransition(assets, to.RestHead, issuesDir); err != nil {
		return err
	}
	if dryRun {
		fmt.Fprintf(stdout, "Would return %s from %s to %s\n", to.Address, shortOID(to.Head), to.Resting)
		return nil
	}
	if out, err := r.GitInDir(to.Root, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", to.Resting); err != nil {
		return fmt.Errorf("switch %s to %s: %v\n%s", to.Address, to.Resting, err, out)
	}
	if err := verifySlot(r, to.Root, to.Resting, to.RestHead); err != nil {
		return err
	}
	if err := verifyMaterializedAssets(to.Root, assets.main.ViewOf(to.RestHead), issuesDir); err != nil {
		return err
	}
	removeStaged(r, to.Root)
	if err := rebuildRestingStage(ctx, to.Root, to.RestHead, stdout, stderr); err != nil {
		return fmt.Errorf("%s is back on %s at %s, but binary rebuild failed: %w; rerun sdlc stage --clear %s", to.Address, to.Resting, shortOID(to.RestHead), err, to.Address)
	}
	fmt.Fprintf(stdout, "%s is back on %s (was staged at %s)\n", to.Address, to.Resting, shortOID(to.Head))
	return nil
}

func rebuildRestingStage(ctx context.Context, root, commit string, stdout, stderr io.Writer) error {
	if _, err := os.Stat(filepath.Join(root, "Makefile")); os.IsNotExist(err) {
		if _, binErr := os.Stat(filepath.Join(root, "bin")); os.IsNotExist(binErr) {
			return nil // non-buildable fixture/repository with no binaries
		}
		return fmt.Errorf("%s has bin/ but no Makefile to rebuild it after clear", root)
	} else if err != nil {
		return err
	}
	return stageBuild(ctx, root, commit, stdout, stderr)
}

func slotAssetEnvironment(ctx context.Context, root string) (*trackerEnv, error) {
	reader, err := gitx.NewTrunkFileContext(ctx, root, "origin", "main")
	if err != nil {
		return nil, err
	}
	return &trackerEnv{ctx: ctx, root: root, main: reader}, nil
}

// observeStage reads everything checkStage needs from both slots.
func observeStage(dir, address string) (stageFacts, error) {
	r := execGitRunner{}
	fromID, err := workspace.Resolve(r, dir, "")
	if err != nil {
		return stageFacts{}, err
	}
	toID, err := workspace.Resolve(r, dir, address)
	if err != nil {
		return stageFacts{}, fmt.Errorf("resolve %s: %w", address, err)
	}
	f := stageFacts{SameRepo: fromID.RepoIdentity == toID.RepoIdentity}
	if f.From, err = observeMoveSide(r, fromID); err != nil {
		return stageFacts{}, err
	}
	if f.To, err = observeMoveSide(r, toID); err != nil {
		return stageFacts{}, err
	}
	if f.To.Resting == "" {
		return stageFacts{}, fmt.Errorf("sdlc stage runs between durable slots; %s is not one", f.To.Root)
	}
	if f.From.Head == "" || !f.SameRepo {
		return f, nil // checkStage refuses
	}
	tree, err := r.GitInDir(f.To.Root, "ls-tree", "-r", "-z", "--name-only", f.From.Head)
	if err != nil {
		return stageFacts{}, fmt.Errorf("git ls-tree %s: %v\n%s", f.From.Head, err, tree)
	}
	f.Incoming = splitNonEmpty(string(tree), "\x00")
	return f, nil
}

// verifySlot re-resolves root and requires it on branch ("" = detached) at head.
func verifySlot(r execGitRunner, root, branch, head string) error {
	id, err := workspace.Resolve(r, root, "")
	if err != nil {
		return fmt.Errorf("the switch ran, but verifying %s failed: %w", root, err)
	}
	if got, at := workspaceText(id.Branch, ""), workspaceText(id.Head, ""); got != branch || at != head {
		return fmt.Errorf("the switch ran, but %s is on %s at %s, want %s at %s; inspect it before continuing", root, orDetached(got), at, orDetached(branch), head)
	}
	return nil
}

// takeStageLease locks target's lease and checks it for holder (#421); the
// caller holds the lock until the lease it writes is final.
func takeStageLease(r execGitRunner, to moveSide, holder string) (fleet.StageLease, func(), error) {
	unlock, err := fleet.LockStage(r, to.Root)
	if err != nil {
		return fleet.StageLease{}, nil, err
	}
	held, err := readStageLease(r, to)
	if err == nil {
		err = checkLease(held, holder, to.Address)
	}
	if err != nil {
		unlock()
		return fleet.StageLease{}, nil, err
	}
	return held, unlock, nil
}

// restoreStageLease puts back the lease root held before this run: the same
// record, or none.
func restoreStageLease(r execGitRunner, root string, held fleet.StageLease) error {
	if held == (fleet.StageLease{}) {
		p, err := fleet.StagedRecordPath(r, root)
		if err != nil {
			return err
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return fleet.WriteStageLease(r, root, held)
}

// readStageLease reads to's lease; a damaged one is never read as free.
func readStageLease(r execGitRunner, to moveSide) (fleet.StageLease, error) {
	held, err := fleet.ReadStageLease(r, to.Root)
	if err != nil {
		return fleet.StageLease{}, fmt.Errorf("%s's stage lease is unreadable: %w; check what %s is running, then remove the file by hand", to.Address, err, to.Address)
	}
	return held, nil
}

// removeStaged drops root's staged record; a missing one is already removed.
func removeStaged(r execGitRunner, root string) {
	if p, err := fleet.StagedRecordPath(r, root); err == nil {
		_ = os.Remove(filepath.Clean(p))
	}
}
