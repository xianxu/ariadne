// finding.go — `sdlc finding dispute`: the implementer's channel to answer a gate finding
// (ariadne#202).
//
// The judge could settle a finding (`addressed`, `withdrawn`); the implementer could only
// fix it. A WRONG finding and a correct-but-not-worth-it one therefore left by the same
// door, the round cap, and the next judge inherited no counter-evidence. This verb records
// a `disputed` disposition — a rationale plus a counter-measurement — in the gate's
// ledger, where the next review round must answer it against that evidence.
//
// The policy (who may dispute, what a dispute settles) is the finding model's and
// gatestate's; this file is the IO shell: ownership, branch, ledger read/write, one
// narrow commit.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// gateLedgerKinds are the gates whose findings a dispute may name, keyed by id prefix.
// The prefix is the gate's own data (gateLedgerKind.IDPrefix), so no prefix literal
// lives here.
var gateLedgerKinds = []gateLedgerKind{planGateKind, boundaryGateKind}

// gateKindForID resolves a finding id to the gate that issued it. Pure.
func gateKindForID(id string) (gateLedgerKind, error) {
	var known []string
	for _, k := range gateLedgerKinds {
		if strings.HasPrefix(id, k.IDPrefix+"-") {
			return k, nil
		}
		known = append(known, k.IDPrefix+"-<n> ("+k.Gate+")")
	}
	sort.Strings(known)
	return gateLedgerKind{}, fmt.Errorf("%q is not a gate finding id; known: %s", id, strings.Join(known, ", "))
}

type findingDisputeFlags struct {
	Issue     int
	ID        string
	Rationale string
	Evidence  string
	IssuesDir string
	PlansDir  string
}

// NewFindingCmd is the `finding` noun: verbs an implementer runs on a gate finding.
func NewFindingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "finding",
		Short:         "Answer a gate finding as the implementer (dispute)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"finding\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newFindingDisputeCmd())
	return cmd
}

func newFindingDisputeCmd() *cobra.Command {
	f := findingDisputeFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "dispute --issue N --id ID --rationale TEXT --evidence TEXT",
		Short:         "Decline a gate finding with a rationale and a counter-measurement",
		Long:          renderLong("finding-dispute"),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			return runFindingDispute(cmd.Context(), cmd.ErrOrStderr(), &f, time.Now().UTC().Format(time.RFC3339))
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID whose gate raised the finding")
	cmd.Flags().StringVar(&f.ID, "id", "", "the finding's id, as the gate printed it (PQ-n, BR-n)")
	cmd.Flags().StringVar(&f.Rationale, "rationale", "", "why the finding is wrong or not worth fixing")
	cmd.Flags().StringVar(&f.Evidence, "evidence", "", "the counter-measurement: the command re-run and what it showed")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	cmd.Flags().StringVar(&f.PlansDir, "plans-dir", "", "directory holding the gate ledgers (default workshop/plans, or WF_PLANS_DIR)")
	return cmd
}

func runFindingDispute(ctx context.Context, stderr io.Writer, f *findingDisputeFlags, timestamp string) error {
	ctx = commandContext(ctx)
	if f.Issue <= 0 {
		return errors.New("--issue N is required")
	}
	id := strings.TrimSpace(f.ID)
	rationale, evidence := strings.TrimSpace(f.Rationale), strings.TrimSpace(f.Evidence)
	if rationale == "" || evidence == "" {
		return errors.New("--rationale and --evidence are both required: a dispute is a counter-measurement " +
			"(the command you re-ran and what it showed), not an opinion")
	}
	kind, err := gateKindForID(id)
	if err != nil {
		return err
	}
	if tracked, err := repositoryTracked(ctx, "."); err != nil {
		return err
	} else if !tracked {
		return errors.New("`sdlc finding dispute` runs in issue tracker repositories (it needs the card's owner)")
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	cardID := fmt.Sprintf("%06d", f.Issue)
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, err := snap.Require(cardID)
	if err != nil {
		return err
	}
	if err := requireCardOwnership(env, card); err != nil {
		return err
	}
	if branch := issue.BranchName(card.Path); env.branch != branch {
		return fmt.Errorf("#%s's gate ledgers live on its branch %s (this checkout is on %q)", issue.CLIRef(cardID), branch, env.branch)
	}
	_, owner, _, err := ownership(env, card)
	if err != nil {
		return err
	}

	issuePath, err := issueFilePath(f.IssuesDir, f.Issue)
	if err != nil {
		return err
	}
	plansDir, name := resolvePlansDir(f.PlansDir), filepath.Base(issuePath)
	l, err := readGateLedger(kind, plansDir, name, f.Issue)
	if err != nil {
		return err
	}
	fnd, err := gatestate.AdmitDispute(l, id)
	if err != nil {
		return fmt.Errorf("#%s: %w", issue.CLIRef(cardID), err)
	}
	l.Disputes = append(l.Disputes, gatestate.Dispute{
		ID: id, Rationale: rationale, Evidence: evidence,
		After: gatestate.HighestRound(l), Timestamp: timestamp, By: describeClaimant(owner),
	})
	if err := writeGateLedger(kind, plansDir, name, l, repoIdentity()); err != nil {
		return err
	}
	if err := commitOnly(env, disputeCommitMessage(f.Issue, id), kind.path(plansDir, name)); err != nil {
		return fmt.Errorf("the dispute is written but not committed: %w", err)
	}

	cwarn(stderr, fmt.Sprintf("%s [%s] %s — DISPUTED, not fixed (%s ledger)", id, fnd.Severity, fnd.Title, kind.Gate))
	if vocab.Finding().IsAppealOnly(fnd.Severity) {
		cwarn(stderr, fmt.Sprintf("%s is %s, which a dispute cannot settle: it keeps blocking until the next review round withdraws it", id, fnd.Severity))
	} else {
		cinfo(stderr, fmt.Sprintf("%s no longer blocks; the next review round sees your evidence and may re-raise it", id))
	}
	return nil
}

// disputeCommitMessage is the narrow commit's subject. Pure.
func disputeCommitMessage(issueNum int, id string) string {
	return fmt.Sprintf("#%d: dispute %s", issueNum, id)
}
