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
	rule    flow.Rule // which of Decide's rules decided it, for the info line
	content string    // the issue text with the record written
	warning string    // a malformed prior record, resolved to full
}

// decideChangeCodeFlow infers (or takes the operator's pin for) the issue's
// flow and returns the issue text with the one-line record written. Pure over
// the issue and plan text (ARCH-PURE); reportChangeCodeFlow and
// recordChangeCodeFlow are the IO shell.
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
		// Nowhere to record a flow. Refuse rather than decide one here, outside
		// Decide: an issue without frontmatter is malformed on either flow (#231 BR-14).
		return changeCodeFlow{}, fmt.Errorf("the issue has no YAML frontmatter, so its flow cannot be recorded — fix the issue file (see `sdlc issue --help`)")
	}
	var out changeCodeFlow
	recorded, perr := flow.Recorded(fm)
	if perr != nil {
		out.warning = fmt.Sprintf("flow record unreadable (%v) — treating it as full and rewriting it", perr)
	}
	plan, _ := issue.PlanItemsBody(body)
	hasMx := len(issue.MilestonesInPlanOrder(plan)) > 0
	hasPlan := planContent != ""
	fl, rule, err := flow.Decide(flow.DecideInput{Recorded: recorded, Pin: pin, HasMilestones: hasMx, HasPlan: hasPlan})
	if err != nil {
		return changeCodeFlow{}, err
	}
	if fl.Kind() == flow.Quick {
		fl = flow.WithContract(fl, body)
	}
	out.flow = fl
	out.rule = rule
	out.content = issue.Compose(issue.SetField(fm, flow.Field, flow.Format(fl)), body)
	return out, nil
}

// flowInfoLine is the one line change-code prints about the flow. The skipped
// gate names are derived from the gate declaration, not restated.
func flowInfoLine(fl flow.Flow, rule flow.Rule) string {
	if fl.Kind() == flow.Quick {
		return fmt.Sprintf("flow: quick (%s: %s) — change-code runs none of %s; close runs the small-diff "+
			"review, or the full one if the diff leaves the shell: %s",
			fl.Provenance(), rule, strings.Join(changeCodeGateOrder(), ", "), flow.ShellSummary())
	}
	return fmt.Sprintf("flow: full (%s: %s)", fl.Provenance(), rule)
}

// reportChangeCodeFlow decides the flow and prints it — BEFORE the gates, which
// it decides between. It writes nothing: a gate refusal must leave the issue
// exactly as it was, because Decide treats a recorded full as permanent (no
// downgrade), so a record left behind by a refused run would stick (#231 BR-5).
func reportChangeCodeFlow(stderr io.Writer, f *changeCodeFlags, issueContent, planContent string) changeCodeFlow {
	d, err := decideChangeCodeFlow(issueContent, planContent, f.Flow)
	if err != nil {
		die(stderr, err.Error())
	}
	if d.warning != "" {
		cwarn(stderr, d.warning)
	}
	cinfo(stderr, flowInfoLine(d.flow, d.rule))
	return d
}

// recordChangeCodeFlow writes the decided record to the issue. runChangeCode
// calls it only after every gate passed and past the dry-run return, right
// before the sync commit that lands it (pinned by
// TestRunChangeCodeRecordsFlowAfterGates); the DryRun check here is the same
// guarantee held locally.
//
// The gates can run for minutes, and the issue may be edited meanwhile, so the
// record is RE-DERIVED from the file as it is now rather than written from the
// text read before the gates (#231 BR-11): the edit survives, and a quick
// record's contract hashes describe the edited text. If the edit changed the
// flow itself, the gates ran for the wrong one — it refuses, and the file is
// left as the editor left it.
func recordChangeCodeFlow(stderr io.Writer, f *changeCodeFlags, issuePath, name string, ran changeCodeFlow) {
	if f.DryRun {
		return
	}
	fresh, err := os.ReadFile(issuePath)
	if err != nil {
		die(stderr, fmt.Sprintf("re-read %s to record its flow: %v", issuePath, err))
	}
	now, err := decideChangeCodeFlow(string(fresh), readOptionalPlanFile(f.PlansDir, name), f.Flow)
	if err != nil {
		die(stderr, err.Error())
	}
	if err := flowDrift(ran, now); err != nil {
		die(stderr, err.Error())
	}
	if now.content == string(fresh) {
		return
	}
	if err := os.WriteFile(issuePath, []byte(now.content), 0o644); err != nil {
		die(stderr, fmt.Sprintf("record flow in %s: %v", issuePath, err))
	}
}

// flowDrift refuses when the issue as it is now infers a different flow than the
// one the gates ran under. Pure.
func flowDrift(ran, now changeCodeFlow) error {
	if ran.flow.Kind() == now.flow.Kind() && ran.flow.Provenance() == now.flow.Provenance() {
		return nil
	}
	return fmt.Errorf("the issue changed while change-code ran: it now infers flow %s (%s), but the gates ran "+
		"for %s (%s) — re-run `sdlc change-code`", now.flow.Kind(), now.rule, ran.flow.Kind(), ran.rule)
}

// activeChangeCodeGates is the gate list this run executes: every gate on the
// full flow, none on the quick flow (#231) — no durable plan to judge, no
// estimate, and the one review is close's. Filtering here, rather than a check
// in each closure, keeps changeCodeGates the complete declaration its ordering
// guards test.
func activeChangeCodeGates(c *changeCodeCtx) []gate {
	if c.flow.Kind() == flow.Quick {
		return nil
	}
	return changeCodeGates(c)
}
