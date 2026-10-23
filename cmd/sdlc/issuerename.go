// issuerename.go — `sdlc issue rename` (#566): give an unstarted issue's card
// and details a new slug. A filer-made name can be too long for a checkout to
// hold; before this verb the only repair was hand-written plumbing commits.
package main

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const issueRenameLong = `Rename an unstarted issue's card and details to a new slug (#566).

  sdlc issue rename --issue N [--slug S]

Without --slug the slug is derived from the card's title, cut at a word
boundary within 80 bytes: the fix for a filer-made name too long to check out.
--slug must be lowercase words joined by single hyphens, at most 80 bytes;
any other slug is refused, never corrected.

The card moves on issue-tracker, then the details move on main, each in one
compare-and-swap commit with its bytes unchanged. If the run stops between the
two, run it again: it reads what both refs hold and finishes, and a finished
rename is a no-op. Only an open issue with no issue branch renames: the issue
branch, plan and review files are named after it. The issue's owner may rename
it, or anyone while it is unclaimed. This checkout is not changed; a resting
branch picks the new names up with its next fast-forward (sdlc claim makes one).`

func newIssueRenameCmd() *cobra.Command {
	var issueID int
	var slug string
	cmd := markMutatingCommand(&cobra.Command{Use: "rename", Short: "Rename an unstarted issue's card and details to a new slug",
		Long: issueRenameLong, Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runIssueRename(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), envOr("WF_ISSUES_DIR", "workshop/issues"), issueID, slug)
		}})
	cmd.Flags().IntVar(&issueID, "issue", 0, "issue ID (required)")
	cmd.Flags().StringVar(&slug, "slug", "", "the new slug (default: derived from the card's title)")
	return cmd
}

func runIssueRename(ctx context.Context, stdout, stderr io.Writer, issuesDir string, issueID int, explicit string) error {
	if issueID <= 0 {
		return fmt.Errorf("--issue is required and must be positive (got %d)", issueID)
	}
	dirs, err := resolveIDDirs(issuesDir, envOr("WF_HISTORY_DIR", "workshop/history"))
	if err != nil {
		return err
	}
	if tracked, err := repositoryTracked(ctx, dirs.Abs[0]); err != nil {
		return err
	} else if !tracked {
		return errLegacyOnly("renaming an issue", "git mv the details file on a branch and land it")
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%06d", issueID)
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, err := snap.Require(id)
	if err != nil {
		return err
	}
	slug, err := issue.FileSlug(card.Card.Title, explicit)
	if err != nil {
		return err
	}
	newStem := id + "-" + slug
	newCard, newDetail := tracker.CardPath(id, slug), path.Join(dirs.Rel[0], newStem+".md")
	token := operationToken("rename")
	checked := false
	// Card, details, then the observation that both are done.
	for range 3 {
		snap, err := env.repo.Snapshot()
		if err != nil {
			return err
		}
		cur, err := snap.Require(id)
		if err != nil {
			return err
		}
		// The details are found by ID: after the card step the card no
		// longer names the old file.
		main, err := env.main.Snapshot()
		if err != nil {
			return err
		}
		names, err := env.git("ls-tree", "--name-only", main.Ref(), dirs.Rel[0]+"/")
		if err != nil {
			return err
		}
		olds, mainNew := tracker.DetailNames(strings.Fields(names), id, newDetail)
		step, err := tracker.NextRenameStep(tracker.RenameObservation{TrackerOld: cur.Path != newCard, TrackerNew: cur.Path == newCard, MainOld: len(olds) > 0, MainNew: mainNew})
		if err == nil && len(olds) > 1 {
			err = tracker.ErrRenameInconsistent
		}
		if err != nil {
			return fmt.Errorf("#%s: the card is %s and main holds %v (new name present: %v); %w", issue.CLIRef(id), cur.Path, olds, mainNew, err)
		}
		// Any step that moves a file needs the permission, including the
		// details step of a half-done rename; observing a finished one does not.
		if step != tracker.RenameDone && !checked {
			if err := renamable(env, cur); err != nil {
				return err
			}
			checked = true
		}
		switch step {
		case tracker.RenameCard:
			err = env.repo.MoveCard(id, cur.Path, newCard, token)
		case tracker.RenameDetails:
			err = env.main.UpdateMany(fmt.Sprintf("#%s: issue: rename details", issue.CLIRef(id)), tracker.MovePath(olds[0], newDetail))
		case tracker.RenameDone:
			invalidateIssueRecords(env.ctx)
			cok(stderr, fmt.Sprintf("#%s is named %s", issue.CLIRef(id), newStem))
			fmt.Fprintln(stdout, newDetail)
			return nil
		}
		if err != nil {
			return fmt.Errorf("#%s rename stopped (rerun `sdlc issue rename --issue %s --slug %s` to finish): %w", issue.CLIRef(id), issue.CLIRef(id), slug, err)
		}
	}
	return fmt.Errorf("#%s: rename did not settle; rerun it", issue.CLIRef(id))
}

// renamable refuses a rename that would orphan names derived from the stem:
// only an open issue with no issue branch, by its owner or while unclaimed.
func renamable(env *trackerEnv, card tracker.Record) error {
	if err := requireOwnedToEdit(env, card, "renames it"); err != nil {
		return err
	}
	if status, _ := issue.GetField(card.Card.Frontmatter, "status"); status != "open" {
		return fmt.Errorf("#%s is %s; rename it before start-plan (its branch, plan and review files are named after it)", issue.CLIRef(card.ID), status)
	}
	branch := issue.BranchName(card.Path)
	if _, err := env.git("rev-parse", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
		return fmt.Errorf("#%s has an issue branch %s; only an unstarted issue renames", issue.CLIRef(card.ID), branch)
	}
	heads, err := env.git("ls-remote", "--heads", env.target.Remote, "refs/heads/"+branch)
	if err != nil {
		return fmt.Errorf("check %s for an issue branch: %w", env.target.Remote, err)
	}
	if heads != "" {
		return fmt.Errorf("#%s has an issue branch %s on %s; only an unstarted issue renames", issue.CLIRef(card.ID), branch, env.target.Remote)
	}
	return nil
}
