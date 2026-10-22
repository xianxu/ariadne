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
	cmd.Flags().BoolVar(&clear, "clear", false, "fetch main, fast-forward the resting branch, then rebuild and verify its Go binaries")
	cmd.Flags().StringVar(&abandoned, "abandoned", "", "with --clear: free the stage lease HOLDER left behind (its issue, as the refusal names it)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview without checkout or build changes (clear fetches main to report lag)")
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
	landedHolder := false
	if err := checkClearLease(held, stageHolder(moveSide{Branch: workspaceText(me.Branch, ""), Address: workspaceText(me.Address, "")}), abandoned, to.Address); err != nil {
		caller := moveSide{Branch: workspaceText(me.Branch, ""), Address: workspaceText(me.Address, ""), Resting: workspaceText(me.RestingBranch, "")}
		if abandoned != "" || me.RepoIdentity != id.RepoIdentity || !landedStageCaller(held, caller) {
			return err
		}
		if staged && to.Head != held.Commit {
			return landedStageRefusal(held, to.Address, errors.New("target is no longer at the recorded staged commit"))
		}
		landedHolder = true
	}
	assets, err := slotAssetEnvironment(ctx, to.Root)
	if err != nil {
		return err
	}
	if err := checkTargetClean(to); err != nil {
		return err
	}
	pin, err := assets.main.Snapshot()
	if err != nil {
		if dryRun {
			fmt.Fprintf(stdout, "%s freshness unverified: fetch origin/main failed: %v; no checkout, lease or build changes\n", to.Address, err)
			return nil
		}
		return fmt.Errorf("fetch origin/main before clearing %s: %w; no rebuild; repair the remote and rerun sdlc stage --clear %s", to.Address, err, to.Address)
	}
	var proof *landedStageProof
	if landedHolder {
		proofCtx, cancel := context.WithTimeout(ctx, landingRecoveryTimeout)
		defer cancel()
		proof, err = proveLandedStageHolder(proofCtx, to.Root, pin.Ref(), held)
		if err != nil {
			return landedStageRefusal(held, to.Address, err)
		}
	}
	git := checkoutGit(ctx, to.Root)
	counts, err := git("rev-list", "--left-right", "--count", to.RestHead+"..."+pin.Ref())
	if err != nil {
		return fmt.Errorf("compare %s with fetched main: %w; no rebuild", to.Resting, err)
	}
	var ahead, behind int
	if n, err := fmt.Sscanf(string(counts), "%d %d", &ahead, &behind); err != nil || n != 2 {
		return fmt.Errorf("invalid resting/main comparison %q; no rebuild", counts)
	}
	fmt.Fprintf(stdout, "%s: %s is %d behind, %d ahead of origin/main at %s\n", to.Address, to.Resting, behind, ahead, pin.Ref())
	if ahead > 0 {
		return fmt.Errorf("%s resting branch %s is ahead or diverged from origin/main (%d local commits); no rebuild; reconcile %s without discarding work, then rerun sdlc stage --clear %s", to.Address, to.Resting, ahead, to.Resting, to.Address)
	}
	issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
	for _, target := range []string{to.RestHead, pin.Ref()} {
		if err := verifyAssetTransition(assets, target, issuesDir); err != nil {
			return err
		}
	}
	// Clear switches through rest before merging; its attributes also govern
	// assets that exist only in the freshly fetched tree.
	if err := verifyAssetMaterialization(ctx, to.Root, pin, issuesDir, to.RestHead); err != nil {
		return err
	}
	if proof != nil {
		if err := proof.revalidate(); err != nil {
			return landedStageRefusal(held, to.Address, err)
		}
		now, err := workspace.Resolve(boundedLandingGit(proof.env.ctx, r), dir, "")
		if err != nil || !reflect.DeepEqual(me, now) {
			return landedStageRefusal(held, to.Address, fmt.Errorf("caller changed during proof: %v", err))
		}
		proofGit := boundedLandingGit(proof.env.ctx, r)
		targetID, err := workspace.Resolve(proofGit, to.Root, "")
		if err != nil {
			return landedStageRefusal(held, to.Address, err)
		}
		currentTarget, err := observeMoveSide(proofGit, targetID)
		if err != nil || !reflect.DeepEqual(to, currentTarget) {
			return landedStageRefusal(held, to.Address, fmt.Errorf("target changed during proof: %v", err))
		}
		current, err := readStageLease(proofGit, to)
		if err != nil || !reflect.DeepEqual(held, current) {
			return landedStageRefusal(held, to.Address, fmt.Errorf("stage lease changed during proof: %v", err))
		}
		if err := proof.env.ctx.Err(); err != nil {
			return landedStageRefusal(held, to.Address, err)
		}
		fmt.Fprintf(stdout, "Confirmed landed holder %s from %s; archive verified at %s\n", held.Holder, held.Slot, pin.Ref())
	}
	if dryRun {
		if held.Holder != "" && !staged {
			fmt.Fprintf(stdout, "Would release %s's stage lease on %s (already on its resting branch %s)\n", held.Holder, to.Address, to.Resting)
		}
		fmt.Fprintf(stdout, "Would return %s to %s, fast-forward to %s and rebuild\n", to.Address, to.Resting, pin.Ref())
		return nil
	}
	if staged {
		if out, err := git("-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", to.Resting); err != nil {
			return fmt.Errorf("switch %s to %s: %v\n%s", to.Address, to.Resting, err, out)
		}
	}
	if err := verifySlot(r, to.Root, to.Resting, to.RestHead); err != nil {
		return err
	}
	if err := verifyMaterializedAssets(to.Root, assets.main.ViewOf(to.RestHead), issuesDir); err != nil {
		return err
	}
	if out, err := git("-c", "submodule.recurse=false", "merge", "--ff-only", "--no-autostash", "--no-overwrite-ignore", pin.Ref()); err != nil {
		return fmt.Errorf("%s is on %s but fast-forward to %s failed: %w\n%s; no rebuild; inspect the checkout then rerun sdlc stage --clear %s", to.Address, to.Resting, pin.Ref(), err, out, to.Address)
	}
	if err := verifySlot(r, to.Root, to.Resting, pin.Ref()); err != nil {
		return err
	}
	if err := verifyMaterializedAssets(to.Root, pin, issuesDir); err != nil {
		return err
	}
	removeStaged(r, to.Root)
	if err := rebuildRestingStage(ctx, to.Root, pin.Ref(), stdout, stderr); err != nil {
		return fmt.Errorf("%s is back on %s at %s, but binary rebuild failed: %w; rerun sdlc stage --clear %s", to.Address, to.Resting, shortOID(pin.Ref()), err, to.Address)
	}
	if staged {
		fmt.Fprintf(stdout, "%s is back on %s (was staged at %s)\n", to.Address, to.Resting, shortOID(to.Head))
	} else if held.Holder != "" {
		cok(stderr, fmt.Sprintf("%s is on its resting branch %s; released %s's stage lease", to.Address, to.Resting, held.Holder))
	} else {
		cok(stderr, fmt.Sprintf("%s is on its resting branch %s; nothing staged", to.Address, to.Resting))
	}
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
func readStageLease(r gitRunner, to moveSide) (fleet.StageLease, error) {
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
