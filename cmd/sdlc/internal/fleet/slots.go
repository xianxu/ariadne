package fleet

import (
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// Verdict is a checkout's or slot's readiness to take new work (#289). It is
// an observation: an action that reuses a slot re-checks it at action time.
type Verdict string

const (
	VerdictReady         Verdict = "ready"          // clean, settled, nothing unlanded or claimed
	VerdictHoldsWork     Verdict = "holds-work"     // work to resume: unlanded commits, an open issue, a claim
	VerdictUnknown       Verdict = "unknown"        // a probe it depends on failed; never ready
	VerdictMissing       Verdict = "missing"        // a declared checkout is absent
	VerdictNeedsRecovery Verdict = "needs-recovery" // dirty, mid-operation or detached
)

var verdictOrder = []Verdict{VerdictReady, VerdictHoldsWork, VerdictUnknown, VerdictMissing, VerdictNeedsRecovery}

func (v Verdict) rank() int {
	for i, o := range verdictOrder {
		if o == v {
			return i
		}
	}
	return -1
}

// Worst is the slot fold: the highest-ranked verdict.
func Worst(vs ...Verdict) Verdict {
	worst := VerdictReady
	for _, v := range vs {
		if v.rank() > worst.rank() {
			worst = v
		}
	}
	return worst
}

// MemberVerdict is one checkout's verdict with its reason codes and, for
// failed probes, their errors.
type MemberVerdict struct {
	Verdict Verdict  `json:"verdict"`
	Reasons []string `json:"reasons"`
	Errors  []string `json:"errors,omitempty"`
}

// JudgeCheckout judges one checkout row against its resting branch (pure).
// Facts that show needs-recovery decide it even when another probe failed;
// otherwise a failed probe the verdict depends on makes it unknown;
// otherwise unlanded commits, an open issue or this machine's claim hold
// work; otherwise it is ready.
func JudgeCheckout(row TreeRow, resting string) MemberVerdict {
	var recovery, holds, probes, errs []string
	probe := func(what, err string) { probes, errs = append(probes, "probe:"+what), append(errs, err) }
	f := row.Facts
	if !f.Available {
		probe("facts", f.Error)
	} else {
		if f.DirtyCount != nil && *f.DirtyCount > 0 {
			recovery = append(recovery, "dirty")
		}
		if f.Operation != "" {
			recovery = append(recovery, "operation:"+f.Operation)
		}
		if f.OperationError != "" {
			probe("operation", f.OperationError)
		}
		if !f.BaseAvailable || f.Ahead == nil {
			probe("base", f.BaseError)
		} else if *f.Ahead > 0 {
			holds = append(holds, "unlanded-commits")
		}
	}
	if row.Detached {
		recovery = append(recovery, "detached")
	}
	if row.Branch != "" && row.Branch != resting {
		if len(row.Issues) == 0 {
			if _, _, isIssue := issue.ParseFilename(issueBranchStem(row.Branch) + ".md"); isIssue {
				probe("issue", "the issue named by branch "+row.Branch+" could not be read")
			}
		}
		for _, a := range row.Issues {
			if !vocab.Issue().IsTerminal(a.DeclaredStatus) {
				holds = append(holds, "open-issue:"+a.Ref)
			}
		}
	}
	if claimsCarryValue(row.ClaimsState) {
		for _, c := range row.Claims {
			holds = append(holds, "claimed:"+c.Ref)
		}
	} else if row.ClaimsState != ClaimsAbsent {
		probe("claims", row.ClaimsError)
	}
	switch {
	case len(recovery) > 0:
		return MemberVerdict{Verdict: VerdictNeedsRecovery, Reasons: append(recovery, probes...), Errors: errs}
	case len(holds) > 0 && onlyClaimsUnread(probes):
		// A claim could only add holds-work; the work already held decides.
		return MemberVerdict{Verdict: VerdictHoldsWork, Reasons: holds}
	case len(probes) > 0:
		return MemberVerdict{Verdict: VerdictUnknown, Reasons: probes, Errors: errs}
	case len(holds) > 0:
		return MemberVerdict{Verdict: VerdictHoldsWork, Reasons: holds}
	}
	return MemberVerdict{Verdict: VerdictReady, Reasons: []string{}}
}

func onlyClaimsUnread(probes []string) bool {
	return len(probes) == 1 && probes[0] == "probe:claims"
}

// issueBranchStem is the leading component of an issue-prefixed branch name,
// as AssociateBranchIssue reads it.
func issueBranchStem(branch string) string {
	if i := strings.IndexByte(branch, '/'); i >= 0 {
		return branch[:i]
	}
	return branch
}
