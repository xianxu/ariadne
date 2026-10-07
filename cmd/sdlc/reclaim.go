// reclaim.go — `sdlc reclaim` (#278): operator-directed transfer of an issue's
// recorded responsibility (#277's claimant) to the workspace running it, after
// the operator has coordinated out of band. Two steps: inspect (read-only)
// shows the current and proposed owner and the card revision; confirm, pinned
// to that revision with --expect and justified with --reason, publishes the
// transfer by compare-and-swap. The tracker commit's trailers are the record.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const (
	reclaimFromTrailer   = "Reclaim-From"
	reclaimToTrailer     = "Reclaim-To"
	reclaimReasonTrailer = "Reclaim-Reason"
)

// reclaimDecision is reclaim's pure core: the card with me as its claimant, and
// the claimant it replaces. The card must hold a live lock (a holdable status
// per the model, #283 — an open shaping claim included) with a recorded owner;
// rev is the card revision read now and expect the one the operator inspected.
// The owner's repeat is errAlreadyMine — so an identical retry, or one after a
// lost publication response, is decided by the card itself.
func reclaimDecision(card []byte, rev, expect, reason string, me issue.Claimant) ([]byte, issue.Claimant, error) {
	fm, _, err := issue.Parse(string(card))
	if err != nil {
		return nil, issue.Claimant{}, err
	}
	id, _ := issue.GetField(fm, "id")
	status, _ := issue.GetField(fm, "status")
	if !vocab.Issue().CanHoldOwner(status) {
		return nil, issue.Claimant{}, fmt.Errorf("#%s is %s; there is no live responsibility to reclaim", id, status)
	}
	recorded, has, err := issue.CardClaimant(card)
	if err != nil {
		return nil, issue.Claimant{}, err
	}
	switch {
	case !has && vocab.Issue().IsOpen(status):
		return nil, issue.Claimant{}, fmt.Errorf("#%s is open; nobody holds it — `sdlc claim --issue %s` takes it", id, issue.CLIRef(id))
	case !has:
		return nil, issue.Claimant{}, fmt.Errorf("#%s has no owner; nothing to reclaim — `sdlc claim --issue %s` takes it over", id, issue.CLIRef(id))
	}
	if issue.MatchClaimant(&recorded, me) == issue.OwnershipMine {
		return nil, recorded, errAlreadyMine
	}
	if expect == "" {
		return nil, recorded, errors.New("--expect is required: the card revision you inspected (run `sdlc reclaim --issue N` to see it)")
	}
	if rev != expect {
		return nil, recorded, fmt.Errorf("#%s changed since you inspected it (now %s, you saw %s); inspect it again before reclaiming", id, shortOID(rev), shortOID(expect))
	}
	reason = strings.TrimSpace(reason)
	if reason == "" || strings.ContainsAny(reason, "\r\n") {
		return nil, recorded, errors.New("--reason is required: one line saying why responsibility moves (the operator's out-of-band decision)")
	}
	next, err := issue.SetCardClaimant(card, me)
	return next, recorded, err
}

// reclaimTrailers records a transfer in its tracker commit.
func reclaimTrailers(from, to issue.Claimant, reason string) []string {
	return []string{
		reclaimFromTrailer + ": " + describeClaimant(from),
		reclaimToTrailer + ": " + describeClaimant(to),
		reclaimReasonTrailer + ": " + strings.TrimSpace(reason),
	}
}

// reclaimEvent is one past transfer read back from tracker history.
type reclaimEvent struct{ Commit, Date, From, To, Reason string }

// parseReclaimTrailers reads the trailers reclaimTrailers wrote, from one
// commit message; ok is false for a commit that is not a reclaim.
func parseReclaimTrailers(message string) (from, to, reason string, ok bool) {
	for _, line := range strings.Split(message, "\n") {
		key, value, found := strings.Cut(line, ": ")
		if !found {
			continue
		}
		switch key {
		case reclaimFromTrailer:
			from = value
		case reclaimToTrailer:
			to = value
		case reclaimReasonTrailer:
			reason = value
		}
	}
	return from, to, reason, from != "" && to != "" && reason != ""
}

type reclaimFlags struct {
	Issue          int
	Expect, Reason string
	IssuesDir      string
}

func NewReclaimCmd() *cobra.Command {
	var f reclaimFlags
	cmd := markMutatingCommand(&cobra.Command{
		Use: "reclaim", Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReclaim(commandContext(cmd.Context()), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID (required)")
	cmd.Flags().StringVar(&f.Expect, "expect", "", "the card revision you inspected; without it, reclaim only shows the transfer")
	cmd.Flags().StringVar(&f.Reason, "reason", "", "one line: why responsibility moves (recorded in tracker history)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	return cmd
}

// reclaimEffect publishes one transfer. A variable only so a test can lose the
// publication response after the effect landed; nothing but runReclaim calls it
// (TestReclaimIsOnlyOperatorInvoked).
var reclaimEffect = func(env *trackerEnv, card tracker.Record, next []byte, trailers []string) error {
	return cardPublish(env, card, next, operationToken("reclaim"), trailers, nil)
}

// runReclaim inspects (no --expect) or performs (--expect + --reason) a
// transfer of #N's responsibility to this workspace. Both read one fresh card;
// the decision and the compare-and-swap key on it.
func runReclaim(ctx context.Context, stdout, stderr io.Writer, f *reclaimFlags) error {
	if f.Issue <= 0 {
		return errors.New("--issue is required and must be positive")
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%06d", f.Issue)
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, err := snap.Require(id)
	if err != nil {
		return err
	}
	me, err := claimantIdentity(env)
	if err != nil {
		return err
	}
	if f.Expect == "" {
		if f.Reason != "" {
			return errors.New("--reason needs --expect: inspect first (`sdlc reclaim --issue N`), then confirm the revision it shows")
		}
		return inspectReclaim(stdout, env, snap.Ref(), card, me)
	}
	detailPath := path.Join(f.IssuesDir, path.Base(card.Path))
	next, from, err := reclaimDecision(card.Raw, card.BlobOID, f.Expect, f.Reason, me)
	if errors.Is(err, errAlreadyMine) {
		cok(stderr, fmt.Sprintf("#%s is already this workspace's (%s); nothing to reclaim — a retry of a reclaim that landed ends here", id, me.Worktree))
		if warn := refreshLocalMirror(env, detailPath); warn != "" { // the landed reclaim's mirror, after a lost response
			cwarn(stderr, warn)
		}
		return nil
	}
	if err != nil {
		return err
	}
	err = reclaimEffect(env, card, next, reclaimTrailers(from, me, f.Reason))
	invalidateIssueRecords(env.ctx)
	switch {
	case errors.Is(err, tracker.ErrCardChanged):
		return fmt.Errorf("#%s changed while reclaiming; nothing was transferred — inspect it again (`sdlc reclaim --issue %d`)", id, f.Issue)
	case errors.Is(err, gitx.ErrPublicationUncertain):
		return uncertainCardWrite(err, "sdlc reclaim with the same --expect and --reason")
	case err != nil:
		return err
	}
	cok(stderr, fmt.Sprintf("#%s reclaimed: %s → %s", id, describeClaimant(from), describeClaimant(me)))
	if warn := refreshLocalMirror(env, detailPath); warn != "" {
		cwarn(stderr, warn)
	}
	fmt.Fprintln(stdout, "reclaimed")
	return nil
}

// inspectReclaim shows what a reclaim would do and writes nothing.
func inspectReclaim(w io.Writer, env *trackerEnv, ref string, card tracker.Record, me issue.Claimant) error {
	id := card.ID
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	fmt.Fprintf(w, "#%s %s — %s\n  revision:       %s\n", issue.CLIRef(id), card.Card.Title, status, card.BlobOID)
	if recorded, has, err := issue.CardClaimant(card.Raw); err != nil {
		return err
	} else if has {
		fmt.Fprintf(w, "  current owner:  %s\n", describeClaimant(recorded))
	} else {
		fmt.Fprintln(w, "  current owner:  (none recorded)")
	}
	fmt.Fprintf(w, "  proposed owner: %s (this workspace)\n", describeClaimant(me))
	history, err := reclaimHistory(env, ref, card.Path)
	if err != nil {
		return err
	}
	if len(history) > 0 {
		fmt.Fprintln(w, "  past reclaims:")
		for _, e := range history {
			fmt.Fprintf(w, "    %s %s  %s → %s: %s\n", e.Date, e.Commit, e.From, e.To, e.Reason)
		}
	}
	_, _, err = reclaimDecision(card.Raw, card.BlobOID, card.BlobOID, "inspect", me)
	if errors.Is(err, errAlreadyMine) {
		fmt.Fprintln(w, "This workspace already owns it; nothing to reclaim.")
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "To move responsibility here — only after agreeing it with the current owner out of band:\n  sdlc reclaim --issue %s --expect %s --reason '<why>'\n", issue.CLIRef(id), card.BlobOID)
	return nil
}

// reclaimHistoryLimit bounds inspect's history: the most recent reclaims.
const reclaimHistoryLimit = 20

// reclaimHistory lists the card's most recent transfers, newest first, from
// the trailers in the tracker history.
func reclaimHistory(env *trackerEnv, ref, cardPath string) ([]reclaimEvent, error) {
	out, err := env.gitRaw(nil, "log", fmt.Sprintf("--max-count=%d", reclaimHistoryLimit), "--grep=^"+reclaimFromTrailer+": ",
		"--format=%h%x00%cs%x00%B%x01", ref, "--", cardPath)
	if err != nil {
		return nil, fmt.Errorf("read #%s's tracker history: %w", path.Base(cardPath), err)
	}
	var events []reclaimEvent
	for _, rec := range strings.Split(string(out), "\x01") {
		parts := strings.SplitN(strings.TrimLeft(rec, "\n"), "\x00", 3)
		if len(parts) != 3 {
			continue
		}
		if from, to, reason, ok := parseReclaimTrailers(parts[2]); ok {
			events = append(events, reclaimEvent{Commit: parts[0], Date: parts[1], From: from, To: to, Reason: reason})
		}
	}
	return events, nil
}
