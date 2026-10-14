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
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
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
	ciMissing                       // a check main requires has not reported on the head
	ciLocalNeeded                   // a declared local check has no recorded pass on the head (#345)
)

// localGatesWorkflow is the seeded local-gates guard (#309). It runs main's copy on
// pull_request_target and posts localGatesVerdict on the PR head; while main carries
// it, merge requires that status, because a run that never reported must not
// read as green (BR-11).
const (
	localGatesWorkflow = ".github/workflows/local-gates-guard.yml"
	localGatesVerdict  = "local-gates-guard/verdict"
)

func (r ciReason) passes() bool { return r == ciGreen || r == ciNoWorkflows }

// ciVerdict is decideCIGate's answer: the reason, and the checks behind it.
type ciVerdict struct {
	Reason      ciReason
	Offender    []prCheck // the checks that are not green, in report order
	LocalNeeded []string  // declared local checks with no recorded pass on the head (#345)
	Substituted []string  // "CI label ← local status" pairs the gate accepted (#345)
}

// ciInputs is everything decideCIGate reads: the checks on the head, whether the
// head carries workflows, the base's local-check declaration, and which cancelled
// checks GitHub reports as never having had a runner (by checkLabel).
type ciInputs struct {
	Checks           []prCheck
	HeadHasWorkflows bool
	Local            localCheckDecl
	Runnerless       map[string]bool
	// RecordedAncestors names local checks with a passing record on an
	// ancestor of the head whose difference from it is docs-only (#332):
	// close's own commits do not invalidate the suite run before close.
	RecordedAncestors map[string]bool
}

// isCommitStatus reports whether a check is a commit status rather than a check
// run: no workflow and no Actions job link. A recorded local run is a commit
// status, so a CI job that merely carries the sdlc-local/ name is not one.
func isCommitStatus(c prCheck) bool {
	_, job := jobIDFromLink(c.Link)
	return c.Workflow == "" && !job
}

// decideCIGate is the gate's decision, pure. Red outranks unknown outranks
// pending outranks a needed local check: the refusal names the most actionable
// class, and lists every non-green check. An unknown bucket fails closed. A
// cancelled job that never got a runner, with a declared substitute, is
// satisfied by a passing recorded local run on the head (#345); a job that ran
// and was cancelled or failed stays red whatever was recorded locally.
func decideCIGate(in ciInputs) ciVerdict {
	recorded := map[string]bool{}
	for _, c := range in.Checks {
		if c.Bucket == "pass" && isCommitStatus(c) {
			recorded[c.Name] = true
		}
	}
	for name, ok := range in.RecordedAncestors {
		if ok {
			recorded[localStatusContext(name)] = true
		}
	}
	var v ciVerdict
	need := func(name string) {
		for _, n := range v.LocalNeeded {
			if n == name {
				return
			}
		}
		v.LocalNeeded = append(v.LocalNeeded, name)
	}
	for _, c := range in.Local.Checks {
		if c.Required && !recorded[localStatusContext(c.Name)] {
			need(c.Name)
		}
	}
	if len(in.Checks) == 0 {
		switch {
		case in.HeadHasWorkflows:
			v.Reason, v.LocalNeeded = ciNotRegistered, nil
		case len(v.LocalNeeded) > 0:
			v.Reason = ciLocalNeeded
		default:
			v.Reason = ciNoWorkflows
		}
		return v
	}
	rank := map[ciReason]int{ciGreen: 0, ciLocalNeeded: 1, ciPending: 2, ciUnknown: 3, ciRed: 4}
	raise := func(r ciReason) {
		if rank[r] > rank[v.Reason] {
			v.Reason = r
		}
	}
	for _, c := range in.Checks {
		var r ciReason
		switch c.Bucket {
		case "pass", "skipping":
			continue
		case "cancel":
			label := checkLabel(c)
			if sub, ok := in.Local.substituteFor(label); ok && in.Runnerless[label] {
				if recorded[localStatusContext(sub.Name)] {
					v.Substituted = append(v.Substituted, label+" ← "+localStatusContext(sub.Name))
				} else {
					need(sub.Name)
				}
				continue
			}
			r = ciRed
		case "fail":
			r = ciRed
		case "pending":
			r = ciPending
		default:
			r = ciUnknown
		}
		v.Offender = append(v.Offender, c)
		raise(r)
	}
	if len(v.LocalNeeded) > 0 {
		raise(ciLocalNeeded)
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
	case ciMissing:
		fmt.Fprintf(&b, "CI gate: PR #%d has not reported %s, which main's local-gates guard posts on every PR head", pr, localGatesVerdict)
	case ciHeadLag:
		fmt.Fprintf(&b, "CI gate: PR #%d's head on GitHub is not the pushed head merge pins, so its checks belong to another commit", pr)
	case ciLocalNeeded:
		fmt.Fprintf(&b, "CI gate: PR #%d needs recorded local checks on its head: %s", pr, strings.Join(v.LocalNeeded, ", "))
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
	case ciLocalNeeded:
		for _, n := range v.LocalNeeded {
			fmt.Fprintf(&b, "\n  → sdlc local-check run --name %s", n)
		}
		b.WriteString("\n  then re-run `sdlc merge` (" + localChecksPath + " on main declares them).")
	case ciRed:
		b.WriteString("\n  → fix the failing check and push (a guard failure on local-gates paths needs the operator's `local-gates-verified` label), then re-run `sdlc merge`.")
	default:
		b.WriteString("\n  → wait for the checks (`gh pr checks " + strconv.Itoa(pr) + " --watch`), then re-run `sdlc merge`.")
		b.WriteString("\n  A repository whose workflows never run on PRs stays pending here; only then is `--no-ci` the answer.")
	}
	return &ciGateError{Reason: v.Reason, Message: b.String()}
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

func (realGH) JobRun(ctx context.Context, repo string, id int64) (jobRun, error) {
	if !landingRepoValid(repo) || id <= 0 {
		return jobRun{}, fmt.Errorf("ci gate: invalid repository or job id")
	}
	out, err := runLandingGH(ctx, "api", "repos/"+repo+"/actions/jobs/"+strconv.FormatInt(id, 10))
	if err != nil {
		return jobRun{}, err
	}
	var j jobRun
	if err := json.Unmarshal(out, &j); err != nil {
		return jobRun{}, fmt.Errorf("decode job %d: %w", id, err)
	}
	return j, nil
}

func (realGH) PostStatus(ctx context.Context, repo, sha, statusContext, state, description string) error {
	if !landingRepoValid(repo) || !landingOIDValid(sha) || !localStatusContextValid(statusContext) || (state != "success" && state != "failure") {
		return fmt.Errorf("local check: invalid status target %q %q %q %q", repo, sha, statusContext, state)
	}
	_, err := runLandingGH(ctx, "api", "-X", "POST", "repos/"+repo+"/statuses/"+sha,
		"-f", "state="+state, "-f", "context="+statusContext, "-f", "description="+description)
	return err
}

func (realGH) CommitStatuses(ctx context.Context, repo, sha string) (map[string]string, error) {
	if !landingRepoValid(repo) || !landingOIDValid(sha) {
		return nil, fmt.Errorf("ci gate: invalid repository or commit")
	}
	out, err := runLandingGH(ctx, "api", "repos/"+repo+"/commits/"+sha+"/statuses?per_page=100")
	if err != nil {
		return nil, err
	}
	return parseCommitStatuses(out)
}

// parseCommitStatuses reads GitHub's status list, newest first: the first
// state seen for a context is its current one.
func parseCommitStatuses(data []byte) (map[string]string, error) {
	var list []struct {
		Context string `json:"context"`
		State   string `json:"state"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decode commit statuses: %w", err)
	}
	out := map[string]string{}
	for _, s := range list {
		if _, seen := out[s.Context]; !seen && s.Context != "" {
			out[s.Context] = s.State
		}
	}
	return out, nil
}

func (realGH) PRComment(ctx context.Context, repo string, number int, body string) error {
	if !landingRepoValid(repo) || number <= 0 {
		return fmt.Errorf("local check: invalid repository or PR number")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "pr", "comment", strconv.Itoa(number), "--repo", repo, "--body-file", "-")
	cmd.WaitDelay = time.Second
	cmd.Stdin = strings.NewReader(body)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh pr comment: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// requiredChecks names the checks main requires on every PR head: the guard's
// verdict while main (base) carries the guard workflow.
func requiredChecks(git func(args ...string) (string, error), base string) ([]string, error) {
	out, err := git("ls-tree", "--name-only", base, localGatesWorkflow)
	if err != nil {
		return nil, fmt.Errorf("read %s's tree for %s: %w", base, localGatesWorkflow, err)
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	return []string{localGatesVerdict}, nil
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

// localCheckRunner runs one declared local check on sha and records it
// (runLocalCheck bound to a checkout). nil: the gate names the run instead.
type localCheckRunner func(ctx context.Context, c localCheck, sha string) (localRunRecord, error)

// newLocalCheckRunner binds runLocalCheck to a checkout for the CI gate.
func newLocalCheckRunner(d localRunDeps) localCheckRunner {
	return func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
		return runLocalCheck(ctx, d, c, sha)
	}
}

// readLocalChecks reads the declaration from rev's tree; none is an empty one.
func readLocalChecks(git func(args ...string) (string, error), rev string) (localCheckDecl, error) {
	out, err := git("ls-tree", "--name-only", rev, localChecksPath)
	if err != nil {
		return localCheckDecl{}, fmt.Errorf("read %s's tree for %s: %w", rev, localChecksPath, err)
	}
	if strings.TrimSpace(out) == "" {
		return localCheckDecl{}, nil
	}
	data, err := git("show", rev+":"+localChecksPath)
	if err != nil {
		return localCheckDecl{}, fmt.Errorf("read %s on %s: %w", localChecksPath, rev, err)
	}
	return parseLocalChecks([]byte(data))
}

// runnerlessChecks asks GitHub, for each cancelled check that has a declared
// substitute and an Actions job link, whether the job ever got a runner. Only
// those checks are asked about, so the calls are bounded by the declaration.
func runnerlessChecks(ctx context.Context, gh ghCaller, repo string, checks []prCheck, decl localCheckDecl) (map[string]bool, error) {
	out := map[string]bool{}
	for _, c := range checks {
		if c.Bucket != "cancel" {
			continue
		}
		label := checkLabel(c)
		if _, ok := decl.substituteFor(label); !ok {
			continue
		}
		id, ok := jobIDFromLink(c.Link)
		if !ok {
			continue
		}
		j, err := gh.JobRun(ctx, repo, id)
		if err != nil {
			return nil, err
		}
		out[label] = j.runnerless()
	}
	return out, nil
}

// unrecordedRequired names the required local checks with no passing record
// on the head itself.
func unrecordedRequired(decl localCheckDecl, checks []prCheck) []string {
	var out []string
	for _, c := range decl.Checks {
		if !c.Required {
			continue
		}
		recorded := false
		for _, ch := range checks {
			if ch.Name == localStatusContext(c.Name) && ch.Bucket == "pass" && isCommitStatus(ch) {
				recorded = true
			}
		}
		if !recorded {
			out = append(out, c.Name)
		}
	}
	return out
}

// splitNUL splits git's -z output.
func splitNUL(out string) []string {
	var paths []string
	for _, p := range strings.Split(out, "\x00") {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// localAncestorLimit bounds the walk for a record on an ancestor.
const localAncestorLimit = 30

// localRecordsOnAncestors finds, for each named check, a passing record on a
// first-parent ancestor of head whose difference from head is docs-only (the
// publish gate's line: embedded docs count as code). The walk stops at the
// first ancestor with a code difference: an older one only differs by more.
func localRecordsOnAncestors(ctx context.Context, gh ghCaller, repo string, git func(args ...string) (string, error), head, base string, names []string) (map[string]bool, error) {
	found := map[string]bool{}
	if len(names) == 0 {
		return found, nil
	}
	mb, err := git("merge-base", head, base)
	if err != nil {
		return nil, fmt.Errorf("merge base of %s and %s: %w", abbrevSHA(head), base, err)
	}
	parents, err := git("rev-list", "--parents", "--max-count=1", head)
	if err != nil {
		return nil, fmt.Errorf("parents of %s: %w", abbrevSHA(head), err)
	}
	if len(strings.Fields(parents)) == 0 {
		return nil, fmt.Errorf("parents of %s: git returned no commit", abbrevSHA(head))
	}
	if len(strings.Fields(parents)) == 1 {
		return found, nil // a root commit has no ancestor to inherit from
	}
	out, err := git("rev-list", "--first-parent", fmt.Sprintf("--max-count=%d", localAncestorLimit), head+"~1", "--not", strings.TrimSpace(mb))
	if err != nil {
		return nil, fmt.Errorf("ancestors of %s: %w", abbrevSHA(head), err)
	}
	for _, a := range strings.Fields(out) {
		diff, err := git(gitx.ChangedPathArgs(a, head)...)
		if err != nil {
			return nil, err
		}
		if publishGateHasCodeSurface(splitNUL(diff)) {
			break
		}
		statuses, err := gh.CommitStatuses(ctx, repo, a)
		if err != nil {
			return nil, err
		}
		for _, n := range names {
			if statuses[localStatusContext(n)] == "success" {
				found[n] = true
			}
		}
		if len(found) == len(names) {
			break
		}
	}
	return found, nil
}

// runCIGate reads PR number's checks and refuses unless they are green. A query
// failure refuses too: unable to read is not green. noCI waives only this gate.
// A declared local check the verdict needs (#345) runs through run, on head, and
// the checks are read again before the gate passes.
func runCIGate(ctx context.Context, gh ghCaller, repo string, number int, base, head string, git func(args ...string) (string, error), run localCheckRunner, noCI bool, stderr io.Writer) error {
	return runPreparedCIGate(ctx, gh, repo, number, base, head, git, run, noCI, stderr, nil)
}

func runPreparedCIGate(ctx context.Context, gh ghCaller, repo string, number int, base, head string, git func(args ...string) (string, error), run localCheckRunner, noCI bool, stderr io.Writer, facts *ciGateFacts) error {
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
	if facts == nil {
		var err error
		facts, err = prepareCIGateFactsContext(ctx, git, base, head)
		if err != nil {
			return err
		}
	}
	decl, hasWF, required := facts.decl, facts.hasWorkflows, facts.required
	read := func() ([]prCheck, ciVerdict, error) {
		before, err := gh.PRHeadOID(ctx, repo, number)
		if err != nil {
			return nil, ciVerdict{}, unread("head", err)
		}
		checks, err := gh.PRChecks(ctx, repo, number)
		if err != nil {
			return nil, ciVerdict{}, unread("checks", err)
		}
		after, err := gh.PRHeadOID(ctx, repo, number)
		if err != nil {
			return nil, ciVerdict{}, unread("head", err)
		}
		if before != head || after != head {
			return nil, ciVerdict{}, ciVerdict{Reason: ciHeadLag}.refusal(number)
		}
		runnerless, err := runnerlessChecks(ctx, gh, repo, checks, decl)
		if err != nil {
			return nil, ciVerdict{}, unread("Actions job", err)
		}
		ancestors, err := localRecordsOnAncestors(ctx, gh, repo, git, head, base, unrecordedRequired(decl, checks))
		if err != nil {
			return nil, ciVerdict{}, unread("local check records", err)
		}
		return checks, decideCIGate(ciInputs{Checks: checks, HeadHasWorkflows: hasWF, Local: decl, Runnerless: runnerless, RecordedAncestors: ancestors}), nil
	}
	checks, v, err := read()
	if err != nil {
		return err
	}
	missing := func(checks []prCheck) error {
		for _, name := range required {
			if !hasCheck(checks, name) {
				return ciVerdict{Reason: ciMissing}.refusal(number)
			}
		}
		return nil
	}
	// A refusal already known is given before any local run spends minutes.
	if v.Reason == ciLocalNeeded {
		if err := missing(checks); err != nil {
			return err
		}
	}
	if v.Reason == ciLocalNeeded && run != nil {
		for _, name := range v.LocalNeeded {
			c, _ := decl.check(name)
			rec, err := run(ctx, c, head)
			if err != nil {
				return fmt.Errorf("CI gate: %w", err)
			}
			if !rec.passed() {
				return fmt.Errorf("CI gate: local check %s failed on %s (%s); log %s\n  → fix it and push, then re-run `sdlc merge`", name, abbrevSHA(head), rec.statusDescription(), rec.LogPath)
			}
		}
		// The run took minutes: read again, so a push or a new failure since
		// counts, and the recorded statuses are what passes the gate.
		if checks, v, err = read(); err != nil {
			return err
		}
	}
	if !v.Reason.passes() {
		return v.refusal(number)
	}
	if err := missing(checks); err != nil {
		return err
	}
	for _, s := range v.Substituted {
		cwarn(stderr, "CI job never got a runner; accepted the recorded local run instead: "+s)
	}
	if v.Reason == ciNoWorkflows {
		cok(stderr, "CI checks: none, and no workflows on the head — nothing to wait for")
	} else {
		cok(stderr, fmt.Sprintf("CI checks: %d green on PR #%d", len(checks), number))
	}
	return nil
}

func hasCheck(checks []prCheck, name string) bool {
	for _, c := range checks {
		if c.Name == name {
			return true
		}
	}
	return false
}
