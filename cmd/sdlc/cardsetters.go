// cardsetters.go — the verbs that change card-owned metadata (#252). The card
// on issue-tracker is authoritative for these fields; details carry a mirror
// that sdlc refreshes, so a hand edit there is refused rather than trusted.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// cardUpdate is one setter's pure decision: from the current card (and the
// local details body, "" when this checkout has none) to the next card bytes.
type cardUpdate func(env *trackerEnv, card tracker.Record, detailsBody string) ([]byte, error)

// runCardUpdate applies one decision by CAS on the tracker, then refreshes the
// local mirror (never on the resting branch). A dry run decides and stops.
func runCardUpdate(ctx context.Context, stdout, stderr io.Writer, issuesDir string, issueID int, what string, dryRun bool, decide cardUpdate) error {
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
		return errLegacyOnly("setting "+what, "edit the field in the details file's frontmatter, then `sdlc issue sync --issue N`")
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
	card, ok := snap.Card(id)
	if !ok {
		return fmt.Errorf("no card #%s on the tracker", id)
	}
	detailPath := path.Join(dirs.Rel[0], path.Base(card.Path))
	body := ""
	if raw, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(detailPath))); err == nil {
		// A malformed file is not "no Log": the reopen guard would misread it.
		_, b, perr := issue.Parse(string(raw))
		if perr != nil {
			return fmt.Errorf("%s is malformed (%v); repair it before changing the card", detailPath, perr)
		}
		body = b
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read %s: %w", detailPath, err)
	}
	next, err := decide(env, card, body)
	if err != nil {
		return err
	}
	if dryRun {
		cinfo(stderr, "dry-run — the card is unchanged")
		fmt.Fprintf(stdout, "Would update #%s: %s\n", id, what)
		return nil
	}
	err = uncertainCardWrite(cardPublish(env, card, next, operationToken("set"), nil, nil), fmt.Sprintf("the same `sdlc issue` setter for #%s", id))
	invalidateIssueRecords(env.ctx)
	switch {
	case errors.Is(err, tracker.ErrNoChange):
		cok(stderr, fmt.Sprintf("#%s already has that %s", id, what))
		return nil
	case errors.Is(err, tracker.ErrCardChanged):
		return fmt.Errorf("card #%s changed concurrently; rerun to apply %s to the current card", id, what)
	case err != nil:
		return err
	}
	cok(stderr, fmt.Sprintf("#%s: %s", id, what))
	if warn := refreshLocalMirror(env, detailPath); warn != "" {
		cwarn(stderr, warn)
	}
	fmt.Fprintln(stdout, card.Path)
	return nil
}

func newCardSetterCmd(use, short string, args cobra.PositionalArgs, bind func(*cobra.Command) func([]string) (string, cardUpdate, error)) *cobra.Command {
	var issueID int
	var dryRun bool
	cmd := markMutatingCommand(&cobra.Command{Use: use, Short: short, Args: args, SilenceErrors: true})
	build := bind(cmd)
	cmd.RunE = func(c *cobra.Command, a []string) error {
		what, decide, err := build(a)
		if err != nil {
			return err
		}
		return runCardUpdate(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), envOr("WF_ISSUES_DIR", "workshop/issues"), issueID, what, dryRun, decide)
	}
	cmd.Flags().IntVar(&issueID, "issue", 0, "issue ID (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "decide and print; do not publish")
	return cmd
}

func newIssueSetTitleCmd() *cobra.Command {
	return newCardSetterCmd("set-title <title>", "Retitle an issue's card (its path and branch keep their slug)", cobra.ExactArgs(1),
		func(*cobra.Command) func([]string) (string, cardUpdate, error) {
			return func(a []string) (string, cardUpdate, error) {
				return "title → " + a[0], func(_ *trackerEnv, card tracker.Record, _ string) ([]byte, error) {
					return issue.SetCardTitle(card.Raw, a[0])
				}, nil
			}
		})
}

func newIssueSetEstimateCmd() *cobra.Command {
	var hours float64
	return newCardSetterCmd("set-estimate", "Record estimate_hours on an issue's card (derive it after plan-quality)", cobra.NoArgs,
		func(cmd *cobra.Command) func([]string) (string, cardUpdate, error) {
			cmd.Flags().Float64Var(&hours, "hours", 0, "estimated hours, positive (required)")
			return func([]string) (string, cardUpdate, error) {
				if hours <= 0 {
					return "", nil, errors.New("--hours must be a positive number")
				}
				value := strconv.FormatFloat(hours, 'f', -1, 64)
				return "estimate_hours → " + value, func(_ *trackerEnv, card tracker.Record, _ string) ([]byte, error) {
					return issue.SetCardField(card.Raw, "estimate_hours", value)
				}, nil
			}
		})
}

func newIssueSetGitHubCmd() *cobra.Command {
	var number int
	return newCardSetterCmd("set-github", "Link an issue's card to its GitHub issue", cobra.NoArgs,
		func(cmd *cobra.Command) func([]string) (string, cardUpdate, error) {
			cmd.Flags().IntVar(&number, "number", 0, "GitHub issue number, positive (required)")
			return func([]string) (string, cardUpdate, error) {
				if number <= 0 {
					return "", nil, errors.New("--number must be a positive GitHub issue number")
				}
				value := strconv.Itoa(number)
				return "github_issue → " + value, func(_ *trackerEnv, card tracker.Record, _ string) ([]byte, error) {
					return issue.SetCardField(card.Raw, "github_issue", value)
				}, nil
			}
		})
}
