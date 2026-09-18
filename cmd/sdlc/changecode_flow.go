package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// changeCodeFlow is change-code's flow decision over one issue (#231).
type changeCodeFlow struct {
	flow    flow.Flow
	reason  string // why this kind, for the info line
	content string // the issue text with the record written
	warning string // a malformed prior record, resolved to full
}

// decideChangeCodeFlow infers (or takes the operator's pin for) the issue's
// flow and returns the issue text with the one-line record written. Pure over
// the issue and plan text (ARCH-PURE); applyChangeCodeFlow is the IO shell.
//
// The inference inputs are the two artifacts the agent already produces under
// the constitution: Mx milestone rows in the Plan (read through the ONE
// enumeration close's verdict gate uses), and a durable plan — planContent is
// the same lookup plan-quality judges, so "a plan exists" means exactly "a plan
// plan-quality would have reviewed". On quick, the record also carries the
// contract hashes the close-time freshness check compares against, refreshed on
// every run so the anchor moves when the contract is re-fixed.
func decideChangeCodeFlow(issueContent, planContent, pin string) (changeCodeFlow, error) {
	fm, body, err := issue.Parse(issueContent)
	if err != nil {
		// No frontmatter: nothing to record, and the structural gate refuses the
		// issue on the full flow anyway.
		return changeCodeFlow{flow: flow.Flow{Kind: flow.Full}, reason: "the issue has no frontmatter", content: issueContent}, nil
	}
	var out changeCodeFlow
	var recorded *flow.Flow
	rec, isRecorded, perr := flow.FromFrontmatter(fm)
	switch {
	case perr != nil:
		out.warning = fmt.Sprintf("flow record unreadable (%v) — treating it as full and rewriting it", perr)
		recorded = &flow.Flow{Kind: flow.Full, Provenance: flow.Inferred}
	case isRecorded:
		recorded = &rec
	}
	plan, _ := issue.PlanItemsBody(body)
	hasMx := len(issue.MilestonesInPlanOrder(plan)) > 0
	hasPlan := planContent != ""
	fl, err := flow.Decide(flow.DecideInput{Recorded: recorded, Pin: pin, HasMilestones: hasMx, HasPlan: hasPlan})
	if err != nil {
		return changeCodeFlow{}, err
	}
	if fl.Kind == flow.Quick {
		fl = flow.WithContract(fl, body)
	}
	out.flow = fl
	out.reason = flowReason(fl, recorded, pin, hasMx, hasPlan)
	out.content = issue.Compose(issue.SetField(fm, flow.Field, flow.Format(fl)), body)
	return out, nil
}

// flowReason names why Decide landed where it did, in the order Decide checks.
func flowReason(fl flow.Flow, recorded *flow.Flow, pin string, hasMx, hasPlan bool) string {
	switch {
	case pin != "":
		return "pinned with --flow"
	case hasMx && fl.Provenance == flow.Inferred:
		return "the Plan has Mx milestones"
	case recorded != nil && recorded.Provenance == flow.Operator:
		return "the operator's pin stands"
	case recorded != nil && recorded.Kind == flow.Full:
		return "already full — gates never downgrade"
	case hasPlan:
		return "a durable plan exists"
	default:
		return "no Mx milestones and no durable plan"
	}
}

// flowInfoLine is the one line change-code prints about the flow. The skipped
// gate names are derived from the gate declaration, not restated.
func flowInfoLine(fl flow.Flow, reason string) string {
	if fl.Kind == flow.Quick {
		return fmt.Sprintf("flow: quick (%s: %s) — change-code runs none of %s; close runs the small-diff "+
			"review, or the full one if the diff leaves the shell: %s",
			fl.Provenance, reason, strings.Join(changeCodeGateOrder(), ", "), flow.ShellSummary())
	}
	return fmt.Sprintf("flow: full (%s: %s)", fl.Provenance, reason)
}

// applyChangeCodeFlow is the IO shell: decide, report, and write the record
// unless --dry-run. It returns the flow and the issue text the gates should see.
func applyChangeCodeFlow(stderr io.Writer, f *changeCodeFlags, issuePath, issueContent, planContent string) (flow.Flow, string) {
	d, err := decideChangeCodeFlow(issueContent, planContent, f.Flow)
	if err != nil {
		die(stderr, err.Error())
	}
	if d.warning != "" {
		cwarn(stderr, d.warning)
	}
	cinfo(stderr, flowInfoLine(d.flow, d.reason))
	if f.DryRun || d.content == issueContent {
		return d.flow, issueContent
	}
	if err := os.WriteFile(issuePath, []byte(d.content), 0o644); err != nil {
		die(stderr, fmt.Sprintf("record flow in %s: %v", issuePath, err))
	}
	return d.flow, d.content
}

// activeChangeCodeGates is the gate list this run executes: every gate on the
// full flow, none on the quick flow (#231) — no durable plan to judge, no
// estimate, and the one review is close's. Filtering here, rather than a check
// in each closure, keeps changeCodeGates the complete declaration its ordering
// guards test.
func activeChangeCodeGates(c *changeCodeCtx) []gate {
	if c.flow.Kind == flow.Quick {
		return nil
	}
	return changeCodeGates(c)
}
