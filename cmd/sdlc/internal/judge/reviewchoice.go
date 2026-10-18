// reviewchoice.go — the one place every review site picks and runs its
// reviewer (#409): resolve the candidates once, try them in order, move on
// only from a reviewer that could not review at all, and say so.
package judge

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/xianxu/ariadne/pkg/reviewpolicy"
)

// ErrReviewerUnavailable is a candidate that could not review at all (see
// Unavailable): trying the next one duplicates no review.
var ErrReviewerUnavailable = errors.New("reviewer unavailable")

// UnavailableError names the candidate and the evidence. Reason is short and
// is what a skip, the trailer and the ledger record; Detail is the full
// diagnosis (a launch failure's owner bin and PATH) for the operator only.
type UnavailableError struct {
	Agent  AgentCLI
	Reason string
	Detail string
}

func (e *UnavailableError) Error() string {
	s := fmt.Sprintf("%s: %s: %s", ErrReviewerUnavailable, e.Agent, e.Reason)
	if e.Detail != "" {
		s += " (" + e.Detail + ")"
	}
	return s
}
func (e *UnavailableError) Unwrap() error { return ErrReviewerUnavailable }

// The choice, its refusal and its record are pkg/reviewpolicy's (#441).
type (
	ReviewChoice    = reviewpolicy.Choice
	NoReviewerError = reviewpolicy.NoReviewerError
)

// ResolveReview resolves every review site's reviewers (it replaced #310's
// ResolveReviewerCLI and ResolveAgentCLI): reviewpolicy.Resolve.
func ResolveReview(explicit string, explicitSet bool, env AgentDefaultEnv, policy ReviewerPolicy, authors []Family, kind ChangeKind, available Availability) ReviewChoice {
	return reviewpolicy.Resolve(explicit, explicitSet, env, policy, authors, kind, available)
}

// ReviewerRecord is reviewpolicy.Record: the Review-Reviewer trailer's text.
func ReviewerRecord(used Candidate, choice ReviewChoice, skips []Skip) string {
	return reviewpolicy.Record(used, choice, skips)
}

// noReviewer is DispatchChoice's refusal: a NoReviewerError that is also a
// review that did not run, for callers matching either.
type noReviewer struct{ *NoReviewerError }

func (e noReviewer) Unwrap() []error { return []error{e.NoReviewerError, ErrReviewDidNotRun} }

// DispatchChoice runs the review with the first candidate that can review,
// announcing each one passed over. It returns the output, the reviewer that
// produced it and every candidate skipped on the way (with choice.Skips
// first). A failure other than unavailability ends the review as before.
func DispatchChoice(ctx context.Context, opts DispatchOptions, choice ReviewChoice) (string, Candidate, []Skip, error) {
	skips := append([]Skip(nil), choice.Skips...)
	for _, s := range choice.Skips {
		announceSkip(opts.Stderr, s)
	}
	for _, c := range choice.Candidates {
		opts.Agent, opts.Model = c.Agent, c.Model
		out, err := Dispatch(ctx, opts)
		var unavailable *UnavailableError
		if errors.As(err, &unavailable) {
			s := Skip{Agent: c.Agent, Reason: unavailable.Reason}
			skips = append(skips, s)
			announceSkip(opts.Stderr, s)
			continue
		}
		return out, c, skips, err
	}
	return "", Candidate{}, skips, noReviewer{&NoReviewerError{Skips: skips}}
}

func announceSkip(w io.Writer, s Skip) {
	if w != nil {
		fmt.Fprintf(w, "  [!] reviewer %s (%s) unavailable: %s; trying the next\n", s.Agent, valueOrDefault(string(FamilyOf(string(s.Agent))), "unknown family"), s.Reason)
	}
}
