// unclaim.go — `sdlc unclaim`: release the lock (#284). A normal operation:
// the owner is cleared and the status never changes. An open claim publishes
// its unpublished edits first, so a claim-to-shape cycle never loses work.
// Every release records who let go, so a rerun recognises its own release.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type unclaimFlags struct {
	Issues    []int
	Note      string
	IssuesDir string
	DryRun    bool
}

// NewUnclaimCmd returns the cobra command for `sdlc unclaim`.
func NewUnclaimCmd() *cobra.Command {
	f := unclaimFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "unclaim",
		Short:         "Release the lock on issues this workspace owns (status unchanged)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"unclaim\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			return runUnclaim(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntSliceVar(&f.Issues, "issue", nil, "issue ID(s) to release: --issue 284 or --issue 284,285")
	cmd.Flags().StringVar(&f.Note, "note", "", "a line for each issue's ## Log (where it stands, what's next)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "check and describe; change nothing")
	return cmd
}

// errAlreadyReleased is unclaimDecision's answer to a card this workspace
// already released: the rerun of a release whose response was lost.
var errAlreadyReleased = errors.New("already released by this workspace")

// unclaimDecision releases one card (#284): this workspace must own it, on a
// status that holds the lock; the claimant is cleared and a release recorded
// (with branch and head for started work). An unowned card this workspace
// released is errAlreadyReleased. Pure.
func unclaimDecision(card []byte, id string, me issue.Claimant, branch, head string) ([]byte, error) {
	fm, _, err := issue.Parse(string(card))
	if err != nil {
		return nil, err
	}
	status, _ := issue.GetField(fm, "status")
	if !vocab.Issue().CanHoldOwner(status) {
		return nil, fmt.Errorf("#%s is %s; it holds no lock to release", issue.CLIRef(id), status)
	}
	recorded, has, err := issue.CardClaimant(card)
	if err != nil {
		return nil, fmt.Errorf("card #%s: %w", id, err)
	}
	if !has {
		if rel, ok, err := issue.CardRelease(card); err != nil {
			return nil, fmt.Errorf("card #%s: %w", id, err)
		} else if ok {
			by := rel.By.Claimant()
			if issue.MatchClaimant(&by, me) == issue.OwnershipMine {
				return nil, errAlreadyReleased
			}
		}
		return nil, fmt.Errorf("#%s has no owner; there is nothing to release", issue.CLIRef(id))
	}
	if issue.MatchClaimant(&recorded, me) != issue.OwnershipMine {
		return nil, fmt.Errorf("#%s is owned by %s; only its owner releases it (a transfer is the operator's `sdlc reclaim`)", issue.CLIRef(id), describeClaimant(recorded))
	}
	out, err := issue.ClearCardClaimant(card)
	if err != nil {
		return nil, err
	}
	return issue.SetCardRelease(out, &issue.Release{By: issue.ReleasedBy(me), Branch: branch, Head: head})
}

// unclaimSetDecision releases open claims together, on the bytes one attempt
// read: any refusal aborts the set; cards already released by this workspace
// are skipped; a set of nothing but those is errAlreadyReleased.
func unclaimSetDecision(current map[string]tracker.Record, ids []string, me issue.Claimant) (map[string][]byte, error) {
	out := map[string][]byte{}
	for _, id := range ids {
		rec, ok := current[id]
		if !ok {
			return nil, fmt.Errorf("no readable card #%s on the tracker", id)
		}
		next, err := unclaimDecision(rec.Raw, id, me, "", "")
		if errors.Is(err, errAlreadyReleased) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out[id] = next
	}
	if len(out) == 0 {
		return nil, errAlreadyReleased
	}
	return out, nil
}

func runUnclaim(ctx context.Context, stdout, stderr io.Writer, f *unclaimFlags) error {
	nums := claimIssues(&claimFlags{Issues: f.Issues})
	if len(nums) == 0 {
		return errors.New("unclaim requires --issue N (or a list: --issue 284,285)")
	}
	if strings.ContainsAny(f.Note, "\r\n") {
		return errors.New("--note is one line")
	}
	if tracked, err := repositoryTracked(ctx, f.IssuesDir); err != nil {
		return err
	} else if !tracked {
		return errors.New("a legacy repository has no owners to release; `sdlc unclaim` works in issue tracker repositories")
	}
	dirs, err := resolveIDDirs(f.IssuesDir, envOr("WF_HISTORY_DIR", "workshop/history"))
	if err != nil {
		return err
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	me, err := claimantIdentity(env)
	if err != nil {
		return err
	}
	ids := make([]string, len(nums))
	cards := map[string]tracker.Record{}
	for i, n := range nums {
		id := fmt.Sprintf("%06d", n)
		card, err := snap.Require(id)
		if err != nil {
			return err
		}
		ids[i], cards[id] = id, card
		switch status, _ := issue.GetField(card.Card.Frontmatter, "status"); {
		case !vocab.Issue().CanHoldOwner(status):
			return fmt.Errorf("#%s is %s; it holds no lock to release", issue.CLIRef(id), status)
		case !vocab.Issue().IsOpen(status):
			// Started work is a handoff: its branch travels, one issue at a time.
			if len(nums) > 1 {
				return fmt.Errorf("#%s is %s: releasing started work hands its branch off, one issue at a time (`sdlc unclaim --issue %s`)", issue.CLIRef(id), status, issue.CLIRef(id))
			}
			return runHandoff(env, stdout, stderr, card, dirs.Rel[0], f.Note, me, f.DryRun)
		}
	}
	// Decide on the cards as read: refusals stop here, before any effect, and
	// cards this workspace already released (a rerun) skip the publish.
	pending := []string{}
	for _, id := range ids {
		if _, err := unclaimDecision(cards[id].Raw, id, me, "", ""); errors.Is(err, errAlreadyReleased) {
			continue
		} else if err != nil {
			return err
		}
		pending = append(pending, id)
	}
	if len(pending) == 0 {
		cok(stderr, fmt.Sprintf("%s already released by this workspace; nothing to do", claimRefs(ids)))
		if !f.DryRun && env.onRest() {
			if view, err := env.main.Snapshot(); err != nil {
				cwarn(stderr, fmt.Sprintf("%s not fast-forwarded to main: reading main failed: %v", env.resting, err))
			} else if warn := fastForwardRest(env, view.Ref()); warn != "" {
				cwarn(stderr, warn)
			}
		}
		fmt.Fprintln(stdout, "released")
		return nil
	}
	if f.DryRun {
		cinfo(stderr, fmt.Sprintf("dry-run — would publish any unpublished edits of %s, then release them in one tracker commit", claimRefs(pending)))
		return nil
	}
	if f.Note != "" {
		for _, id := range pending {
			_, p := localDetail(env, dirs.Rel[0], cards[id].Path)
			if _, err := appendUnclaimNote(p, f.Note); err != nil { // once, whatever the reruns (#284 BR-7)
				return fmt.Errorf("#%s: the note needs its local details: %w", issue.CLIRef(id), err)
			}
		}
	}
	// Publish while the claim is still held: republishOwned needs the owner, and
	// a failure leaves the claim in place with nothing lost. A checkout without
	// an issue's details has nothing of it to publish.
	var local []string
	for _, id := range pending {
		if _, abs := localDetail(env, dirs.Rel[0], cards[id].Path); fileExists(abs) {
			local = append(local, id)
		}
	}
	if len(local) > 0 {
		if err := republishOwned(env, io.Discard, stderr, local, dirs.Rel[0], false); err != nil {
			return fmt.Errorf("%w\n      the claim is kept; fix the above and rerun `sdlc unclaim --issue %s`", err, claimArg(pending))
		}
	}
	decide := func(current map[string]tracker.Record) (map[string][]byte, error) {
		return unclaimSetDecision(current, pending, me)
	}
	err = cardsPublish(env, pending, operationToken("unclaim"), nil, decide, nil)
	invalidateIssueRecords(env.ctx)
	if errors.Is(err, errAlreadyReleased) {
		err = nil
	}
	if err = uncertainCardWrite(err, "sdlc unclaim --issue "+claimArg(ids)); err != nil {
		return err
	}
	for _, id := range pending {
		cok(stderr, fmt.Sprintf("#%s released: no owner; status unchanged. Anyone may claim it.", issue.CLIRef(id)))
	}
	fmt.Fprintln(stdout, "released")
	return nil
}

// fileExists reports whether p names anything, without following a link.
func fileExists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}
