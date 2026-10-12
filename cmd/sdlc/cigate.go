// cigate.go — the CI gate inside `sdlc merge` (#309).
//
// Local sdlc gates do the heavy checks; server CI does cheap deterministic ones
// and guards the local gates' configuration. This gate is where the two meet:
// merge refuses to land a PR whose checks on its head are red or still pending,
// and names each one. Both merge paths (the durable landing and the legacy
// worktree path) call runCIGate, and both pin the merged head to the one whose
// checks were read (`--match-head-commit`).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// prCheck is one check GitHub reports on a PR's head, as `gh pr checks` buckets it.
type prCheck struct {
	Name     string `json:"name"`
	Bucket   string `json:"bucket"`
	Workflow string `json:"workflow"`
	Link     string `json:"link"`
}

// ciReason classifies the gate's answer, so a caller and a test read the reason
// rather than guess it from a bool.
type ciReason int

const (
	ciGreen         ciReason = iota // every check passed or was skipped
	ciNoWorkflows                   // no checks, and the head carries no workflows: nothing to wait for
	ciRed                           // a check failed or was cancelled
	ciPending                       // a check has not finished
	ciNotRegistered                 // no checks yet, but the head carries workflows
	ciUnknown                       // a bucket this gate does not know
	ciHeadLag                       // GitHub's PR head is not (or stopped being) the head merge pins
)

func (r ciReason) passes() bool { return r == ciGreen || r == ciNoWorkflows }

// ciVerdict is decideCIGate's answer: the reason, and the checks behind it.
type ciVerdict struct {
	Reason   ciReason
	Offender []prCheck // the checks that are not green, in report order
}

// decideCIGate is the gate's decision, pure. Red outranks unknown outranks
// pending: the refusal names the most actionable class, and lists every
// non-green check. An unknown bucket fails closed.
func decideCIGate(checks []prCheck, headHasWorkflows bool) ciVerdict {
	if len(checks) == 0 {
		if headHasWorkflows {
			return ciVerdict{Reason: ciNotRegistered}
		}
		return ciVerdict{Reason: ciNoWorkflows}
	}
	v := ciVerdict{Reason: ciGreen}
	rank := map[ciReason]int{ciGreen: 0, ciPending: 1, ciUnknown: 2, ciRed: 3}
	for _, c := range checks {
		var r ciReason
		switch c.Bucket {
		case "pass", "skipping":
			continue
		case "fail", "cancel":
			r = ciRed
		case "pending":
			r = ciPending
		default:
			r = ciUnknown
		}
		v.Offender = append(v.Offender, c)
		if rank[r] > rank[v.Reason] {
			v.Reason = r
		}
	}
	return v
}

// refusal renders a failing verdict as the next-action message merge prints.
func (v ciVerdict) refusal(pr int) error {
	var b strings.Builder
	switch v.Reason {
	case ciRed:
		fmt.Fprintf(&b, "CI gate: PR #%d has failing checks on its head", pr)
	case ciPending:
		fmt.Fprintf(&b, "CI gate: PR #%d has checks still running on its head", pr)
	case ciUnknown:
		fmt.Fprintf(&b, "CI gate: PR #%d has checks in a state this gate does not know (refusing, fail closed)", pr)
	case ciNotRegistered:
		fmt.Fprintf(&b, "CI gate: PR #%d reports no checks yet, but its head carries .github/workflows (GitHub has not registered them)", pr)
	case ciHeadLag:
		fmt.Fprintf(&b, "CI gate: PR #%d's head on GitHub is not the pushed head merge pins, so its checks belong to another commit", pr)
	default:
		return nil
	}
	for _, c := range v.Offender {
		fmt.Fprintf(&b, "\n    %s: %s", c.Bucket, checkLabel(c))
		if c.Link != "" {
			fmt.Fprintf(&b, "  %s", c.Link)
		}
	}
	switch v.Reason {
	case ciRed:
		b.WriteString("\n  → fix the failing check and push (a guard failure on local-gates paths needs the operator's `local-gates-verified` label), then re-run `sdlc merge`.")
	default:
		b.WriteString("\n  → wait for the checks (`gh pr checks " + strconv.Itoa(pr) + " --watch`), then re-run `sdlc merge`.")
		b.WriteString("\n  A repository whose workflows never run on PRs stays pending here; only then is `--no-ci` the answer.")
	}
	return errors.New(b.String())
}

func checkLabel(c prCheck) string {
	if c.Workflow != "" && c.Workflow != c.Name {
		return c.Workflow + " / " + c.Name
	}
	return c.Name
}

// noChecksMessage is what gh prints (stderr, exit 1) for a PR with no checks.
// Measured with gh 2.93 on 10-12: "no checks reported on the '<branch>' branch".
const noChecksMessage = "no checks reported"

// parsePRChecks interprets one `gh pr checks --json` run. gh exits non-zero for
// failing (1) and pending (8) checks while still printing the JSON, so stdout is
// read first: a JSON array is the answer whatever the exit status. The no-checks
// message is an empty list. Anything else is a query failure.
func parsePRChecks(stdout, stderr []byte, runErr error) ([]prCheck, error) {
	out := bytes.TrimSpace(stdout)
	if bytes.HasPrefix(out, []byte("[")) {
		var checks []prCheck
		if err := json.Unmarshal(out, &checks); err != nil {
			return nil, fmt.Errorf("decode gh pr checks: %w", err)
		}
		for _, c := range checks {
			if c.Name == "" || c.Bucket == "" {
				return nil, fmt.Errorf("gh pr checks: a check without a name or bucket: %+v", c)
			}
		}
		return checks, nil
	}
	if strings.Contains(string(stderr), noChecksMessage) {
		return nil, nil
	}
	if runErr != nil {
		return nil, fmt.Errorf("gh pr checks: %w: %s", runErr, strings.TrimSpace(string(stderr)))
	}
	return nil, fmt.Errorf("gh pr checks printed no check list: %q", string(out))
}

func (realGH) PRHeadOID(ctx context.Context, repo string, number int) (string, error) {
	if !landingRepoValid(repo) || number <= 0 {
		return "", fmt.Errorf("ci gate: invalid repository or PR number")
	}
	out, err := runLandingGH(ctx, "pr", "view", strconv.Itoa(number), "--repo", repo, "--json", "headRefOid", "--jq", ".headRefOid")
	if err != nil {
		return "", err
	}
	oid := strings.TrimSpace(string(out))
	if !landingOIDValid(oid) {
		return "", fmt.Errorf("gh pr view: malformed head OID %q", oid)
	}
	return oid, nil
}

func (realGH) PRChecks(ctx context.Context, repo string, number int) ([]prCheck, error) {
	if !landingRepoValid(repo) || number <= 0 {
		return nil, fmt.Errorf("ci gate: invalid repository or PR number")
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI not on PATH: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "pr", "checks", strconv.Itoa(number), "--repo", repo, "--json", "name,bucket,workflow,link")
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("gh pr checks interrupted: %w", ctx.Err())
	}
	return parsePRChecks(stdout.Bytes(), stderr.Bytes(), runErr)
}

// headHasWorkflows reports whether head's tree carries .github/workflows, via
// the given git call (each merge path has its own runner).
func headHasWorkflows(git func(args ...string) (string, error), head string) (bool, error) {
	out, err := git("ls-tree", "--name-only", head, ".github/workflows")
	if err != nil {
		return false, fmt.Errorf("read %s's tree for .github/workflows: %w", head, err)
	}
	return strings.TrimSpace(out) != "", nil
}

// runCIGate reads PR number's checks and refuses unless they are green. A query
// failure refuses too: unable to read is not green. noCI waives only this gate.
func runCIGate(ctx context.Context, gh ghCaller, repo string, number int, head string, git func(args ...string) (string, error), noCI bool, stderr io.Writer) error {
	if noCI {
		cwarn(stderr, "⚠️  --no-ci: SKIPPING the CI gate (#309) — landing without reading the PR's checks. Escape hatch for a broken CI system: say why in your log.")
		return nil
	}
	// gh pr checks reads the checks of whatever head GitHub records for the
	// PR, which can lag a fresh push. Reading that head before and after the
	// checks binds them to the pinned head; a mismatch is pending (BR-1).
	unread := func(what string, err error) error {
		return fmt.Errorf("CI gate: could not read PR #%d's %s (refusing; unread is not green): %w\n  → retry, or --no-ci if the CI system itself is broken", number, what, err)
	}
	before, err := gh.PRHeadOID(ctx, repo, number)
	if err != nil {
		return unread("head", err)
	}
	checks, err := gh.PRChecks(ctx, repo, number)
	if err != nil {
		return unread("checks", err)
	}
	after, err := gh.PRHeadOID(ctx, repo, number)
	if err != nil {
		return unread("head", err)
	}
	if before != head || after != head {
		return ciVerdict{Reason: ciHeadLag}.refusal(number)
	}
	hasWF, err := headHasWorkflows(git, head)
	if err != nil {
		return fmt.Errorf("CI gate: %w", err)
	}
	v := decideCIGate(checks, hasWF)
	if !v.Reason.passes() {
		return v.refusal(number)
	}
	if v.Reason == ciNoWorkflows {
		cok(stderr, "CI checks: none, and no workflows on the head — nothing to wait for")
	} else {
		cok(stderr, fmt.Sprintf("CI checks: %d green on PR #%d", len(checks), number))
	}
	return nil
}
