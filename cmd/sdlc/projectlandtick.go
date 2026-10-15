// projectlandtick.go — the landing tick (#373). When an issue lands, every live
// project that cites it as a row gets that row ticked and "(PR #N)" appended.
// The edit is project.TickLandedIssue, pure; this file is its IO shell. Each
// project is edited on its repository's ORIGIN main through one compare-and-swap
// commit (gitx.TrunkFile, as #308's peer publish), never in a checkout: the
// project's own checkout may be a TL's live ops tree (kink 38), and a resting
// branch only fast-forwards. `sdlc merge` runs it after archiving; `sdlc project
// tick` is the same call by hand.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// errLandingTickOff stops an automatic tick on a project whose landing_tick
// policy is off (a TL-owned project by default): merge prints the command.
var errLandingTickOff = errors.New("landing tick is off for this project")

// landingTickOpts says who is ticking. Auto is the merge hook, which honours
// each project's landing_tick policy; a hand-run tick is the explicit choice
// and writes regardless.
type landingTickOpts struct {
	Auto, DryRun bool
}

// landingTickTransform is the edit applied to origin's bytes of one project.
// The policy is re-read there, since the local copy merge previewed may be
// stale; a copy that will not parse refuses rather than being rewritten.
func landingTickTransform(repoName, issueID, pr string, auto bool) func([]byte) ([]byte, error) {
	return func(old []byte) ([]byte, error) {
		d, err := projectdoc.ParseDoc(string(old))
		if err != nil {
			return nil, fmt.Errorf("parse origin's copy: %w", err)
		}
		if auto {
			on, err := d.LandingTick()
			if err != nil {
				return nil, err
			}
			if !on {
				return nil, errLandingTickOff
			}
		}
		next, _ := projectdoc.TickLandedIssue(string(old), repoName, issueID, pr)
		return []byte(next), nil
	}
}

// landingTickCommand is the exact hand-run equivalent, printed when the
// automatic tick is off or failed.
func landingTickCommand(repoName, issueID, pr string) string {
	cmd := fmt.Sprintf("sdlc project tick --issue %s#%s", repoName, issueID)
	if pr != "" {
		cmd += " --pr " + pr
	}
	return cmd
}

// landedProjectMatches is the merge hook's discovery: live records citing
// repoName#issueID anywhere in the fleet, through the roots close uses.
func landedProjectMatches(identity workspace.Identity, repoName, issueID string) ([]projectdoc.ProjectMatch, error) {
	roots, err := projectWorkspaceRoots(identity)
	if err != nil {
		return nil, err
	}
	return projectdoc.DiscoverInRoots(roots, repoName, issueID, projectdoc.ActiveOnly)
}

// tickLandedProjects ticks repoName#issueID in each matched project. It never
// fails the caller: the issue has landed whatever happens here, so each
// problem is a warning naming the hand-run retry. Returns how many projects
// it could not tick.
func tickLandedProjects(ctx context.Context, stdout, stderr io.Writer, matches []projectdoc.ProjectMatch, repoName, issueID, pr string, opts landingTickOpts) int {
	retry := landingTickCommand(repoName, issueID, pr)
	missed := 0
	for _, m := range matches {
		label := m.Repo + "/" + filepath.Base(m.Path)
		switch {
		case !m.RowRef:
			cinfo(stderr, fmt.Sprintf("project %s carries %s#%s only as a train car (%s) — no row to tick", label, repoName, issueID, m.Train))
			continue
		case m.Legacy:
			cwarn(stderr, fmt.Sprintf("project %s is in the legacy brain home, which is never written (#176) — not ticked", label))
			missed++
			continue
		}
		rel, err := peerProjectPath(m.RepoDir, m.Path)
		if err != nil {
			cwarn(stderr, fmt.Sprintf("project %s NOT ticked: %v", label, err))
			missed++
			continue
		}
		transform := landingTickTransform(repoName, issueID, pr, opts.Auto)
		if opts.DryRun {
			fmt.Fprintln(stdout, dryRunTick(m.Path, rel, filepath.Base(m.RepoDir), transform))
			continue
		}
		changed := false
		err = publishProjectEdit(ctx, m.RepoDir, rel, fmt.Sprintf("project: tick %s#%s landed%s", repoName, issueID, prSuffix(pr)), func(old []byte) ([]byte, error) {
			next, err := transform(old)
			changed = err == nil && string(next) != string(old)
			return next, err
		})
		switch {
		case errors.Is(err, errLandingTickOff):
			cinfo(stderr, fmt.Sprintf("project %s is TL-owned and its landing tick is off (landing_tick) — not ticked; the TL runs:\n      %s", label, retry))
		case err != nil:
			cwarn(stderr, fmt.Sprintf("project %s NOT ticked on %s's origin: %v\n      retry: %s", label, filepath.Base(m.RepoDir), err, retry))
			missed++
		case changed:
			cok(stderr, fmt.Sprintf("ticked %s#%s in %s on %s's origin main (pull it into that checkout)", repoName, issueID, label, filepath.Base(m.RepoDir)))
		default:
			cinfo(stderr, fmt.Sprintf("project %s already records %s#%s%s — nothing to publish", label, repoName, issueID, prSuffix(pr)))
		}
	}
	return missed
}

func prSuffix(pr string) string {
	if pr == "" {
		return ""
	}
	return " (PR #" + pr + ")"
}

// dryRunTick previews one project's edit on the local copy (origin's may
// differ; the real run re-derives on it).
func dryRunTick(path, rel, repo string, transform func([]byte) ([]byte, error)) string {
	old, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("Would skip %s: %v", path, err)
	}
	next, err := transform(old)
	switch {
	case errors.Is(err, errLandingTickOff):
		return fmt.Sprintf("Would not tick %s: landing tick is off (TL-owned)", path)
	case err != nil:
		return fmt.Sprintf("Would skip %s: %v", path, err)
	case string(next) == string(old):
		return fmt.Sprintf("No change: %s", path)
	}
	return fmt.Sprintf("Would publish to %s's origin main: %s", repo, rel)
}

// publishProjectEdit commits one transform of a project file to its
// repository's origin main. Shared with close's peer publish (ARCH-DRY).
func publishProjectEdit(ctx context.Context, repoDir, rel, msg string, transform func([]byte) ([]byte, error)) error {
	if gitx.IsBrainRepo(repoDir) {
		return errors.New("a brain capture repository is never written (#176)")
	}
	tf, err := peerTrunk(ctx, repoDir)
	if err != nil {
		return err
	}
	return tf.Update(rel, msg, transform)
}

type projectTickFlags struct {
	Issue, PR string
	DryRun    bool
	root      string // test injection; "" ⇒ the current checkout
}

func newProjectTickCmd() *cobra.Command {
	f := projectTickFlags{}
	cmd := markMutatingCommand(&cobra.Command{Use: "tick", Short: "Tick a landed issue in every project citing it, on each project's origin", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectTick(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}})
	cmd.Flags().StringVar(&f.Issue, "issue", "", "landed issue ref (e.g. ariadne#373 or #373)")
	cmd.Flags().StringVar(&f.PR, "pr", "", "the landing PR number, appended to each row as (PR #N)")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would be published, write nothing")
	_ = cmd.MarkFlagRequired("issue")
	return cmd
}

func runProjectTick(ctx context.Context, stdout, stderr io.Writer, f *projectTickFlags) error {
	if f.PR != "" {
		if n, err := strconv.Atoi(f.PR); err != nil || n <= 0 {
			return fmt.Errorf("--pr %q: want a PR number", f.PR)
		}
	}
	root, err := currentRoot(f.root)
	if err != nil {
		return err
	}
	matches, repoName, ref, err := discoverProjectsInScope(f.Issue, root, projectdoc.ActiveOnly)
	if err != nil {
		return err
	}
	if missed := tickLandedProjects(commandContext(ctx), stdout, stderr, matches, repoName, strconv.Itoa(ref.ID), f.PR, landingTickOpts{DryRun: f.DryRun}); missed > 0 {
		return fmt.Errorf("%d project(s) not ticked", missed)
	}
	return nil
}

// tickOnLanding is merge's hook (#373): the landed branch's issue (one issue,
// one branch: #272) is ticked in every live project citing it, on each
// project's origin. Warn-only; a branch without an id prefix ticks nothing.
func tickOnLanding(ctx context.Context, stdout, stderr io.Writer, root, branch, pr string) {
	id := issue.IDFromFilename(branch)
	if id == 0 {
		return
	}
	identity, err := resolveWorkspace(root)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("landed; projects not ticked: %v", err))
		return
	}
	issueID := strconv.Itoa(id)
	matches, err := landedProjectMatches(identity, identity.Repo, issueID)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("landed; projects not ticked: %v\n      retry: %s", err, landingTickCommand(identity.Repo, issueID, pr)))
		return
	}
	tickLandedProjects(ctx, stdout, stderr, matches, identity.Repo, issueID, pr, landingTickOpts{Auto: true})
}
