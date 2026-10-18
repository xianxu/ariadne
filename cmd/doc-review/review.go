package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/pkg/reviewpolicy"
)

// AgentCLI is a reviewer CLI, the one agent list pkg/reviewpolicy holds for
// every binary (#441; doc-review once kept its own copy).
type AgentCLI = reviewpolicy.Agent

const (
	AgentCodex  = reviewpolicy.Codex
	AgentGemini = reviewpolicy.Gemini
	AgentClaude = reviewpolicy.Claude
)

type reviewFlags struct {
	Agent    AgentCLI // named on the command line; it wins over the policy
	AgentSet bool
	File     string
	Out      string
	DryRun   bool
}

// runAgent is the subprocess shim. Tests replace it to assert the command line
// and inject a canned report without spawning a real agent. Production execs.
var runAgent = func(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// reportPath computes the sidecar report path: <dir>/<stem>-<agent>-check.md.
// The doc's extension (.md) is stripped from the stem; a doc without one just
// gets the suffix appended.
func reportPath(file string, agent AgentCLI) string {
	dir := filepath.Dir(file)
	base := filepath.Base(file)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.Join(dir, fmt.Sprintf("%s-%s-check.md", stem, agent))
}

// runReview is the binary's one job: choose the reviewer (#441, under #409's
// policy: never the document author's own family), build the baked review
// prompt for <file>, dispatch the reviewer READ-ONLY, capture its report, and
// write the sidecar, then tell the main agent to triage. The reviewer never
// edits the doc; this orchestrator does the only write (the report).
func runReview(stdout, stderr io.Writer, f *reviewFlags) error {
	abs, err := filepath.Abs(f.File)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", f.File, err)
	}
	if st, err := os.Stat(abs); err != nil {
		return fmt.Errorf("cannot read %s: %w", f.File, err)
	} else if st.IsDir() {
		return fmt.Errorf("%s is a directory, not a document", f.File)
	}
	policy, err := docPolicy(filepath.Dir(abs))
	if err != nil {
		return err
	}
	env := sessionEnv()
	authors, unread := docAuthors(abs, env)
	choice := chooseReviewers(policy, string(f.Agent), f.AgentSet, env, authors, onPath)
	if unread != "" {
		choice.Degraded = append(choice.Degraded, unread)
	}
	prompt := buildPrompt(abs)

	if f.DryRun {
		first, ok := choice.First()
		if !ok {
			return &reviewpolicy.NoReviewerError{Skips: choice.Skips}
		}
		name, args, _, tmp := buildArgs(first.Agent, prompt)
		if tmp != "" {
			defer os.Remove(tmp)
		}
		fmt.Fprintf(stdout, "agent:  %s\n", name)
		fmt.Fprintf(stdout, "argv:   %s %s\n", name, shellJoin(args))
		fmt.Fprintf(stdout, "review: %s\n", abs)
		fmt.Fprintf(stdout, "report: %s\n", outPath(f, abs, first.Agent))
		fmt.Fprintf(stdout, "record: %s\n", reviewpolicy.Record(first, choice, choice.Skips))
		return nil
	}

	skips := append([]reviewpolicy.Skip(nil), choice.Skips...)
	for _, s := range choice.Skips {
		announceSkip(stderr, s)
	}
	for _, c := range choice.Candidates {
		report, launchErr, err := reviewOnce(stderr, c.Agent, prompt, abs, f.File)
		if launchErr != nil && choice.Forced {
			return launchErr // the agent the user named could not start; there is no other to try
		}
		if launchErr != nil {
			// The CLI could not start, so no review began: the next family.
			s := reviewpolicy.Skip{Agent: c.Agent, Reason: "could not launch: " + launchErr.Error()}
			skips = append(skips, s)
			announceSkip(stderr, s)
			continue
		}
		if err != nil {
			return err
		}
		out := outPath(f, abs, c.Agent)
		record := reviewpolicy.Record(c, choice, skips)
		if err := os.WriteFile(out, []byte(wrapReport(filepath.Base(f.File), c.Agent, record, report)), 0o644); err != nil {
			return fmt.Errorf("write report %s: %w", out, err)
		}
		fmt.Fprintf(stdout, "\nReport written: %s\nReviewer: %s\n", out, record)
		fmt.Fprintf(stdout, "\nNEXT (main agent): read the report, triage each finding, and update %s where you AGREE.\n"+
			"The review is advisory and read-only — you own the document. Do not apply findings blindly;\n"+
			"verify the claimed corrections before editing, and leave a note for ones you reject.\n", filepath.Base(f.File))
		return nil
	}
	return &reviewpolicy.NoReviewerError{Skips: skips}
}

// outPath is the report path for the reviewer that ran: --out, or the sidecar
// named after that reviewer.
func outPath(f *reviewFlags, abs string, agent AgentCLI) string {
	if f.Out != "" {
		return f.Out
	}
	return reportPath(abs, agent)
}

// reviewOnce runs one reviewer. launchErr is a CLI that could not start (no
// review began, so the caller may try another family); err is a reviewer that
// ran but gave no usable report, which ends the review with today's message.
func reviewOnce(stderr io.Writer, agent AgentCLI, prompt, abs, file string) (report string, launchErr, err error) {
	name, args, fromFile, tmp := buildArgs(agent, prompt) // a fresh -o capture per attempt
	if tmp != "" {
		defer os.Remove(tmp)
	}
	fmt.Fprintf(stderr, "doc-review: dispatching %s (read-only) to review %s …\n", name, abs)
	raw, runErr := runAgent(context.Background(), name, args...)
	if isLaunchFailure(runErr) {
		return "", fmt.Errorf("%s failed to run: %w", name, runErr), nil
	}
	report = string(raw)
	if fromFile {
		b, rerr := os.ReadFile(tmp)
		if rerr != nil {
			return "", nil, fmt.Errorf("%s produced no report file (output was:\n%s)", name, truncate(string(raw), 800))
		}
		report = string(b)
	}
	report = strings.TrimSpace(report)
	if report == "" {
		hint := ""
		if agent == AgentCodex {
			// A codex that ran and crashed (gpt-image-2 bug) is not a failed
			// launch, so it does not move on by itself.
			hint = "\n  codex can crash on some setups (gpt-image-2 bug); retry with a different vendor:\n    doc-review gemini " + file
		}
		return "", nil, fmt.Errorf("%s returned an empty report (likely a launch/auth problem); raw output:\n%s%s", name, truncate(string(raw), 800), hint)
	}
	return report, nil, nil
}

func announceSkip(w io.Writer, s reviewpolicy.Skip) {
	fmt.Fprintf(w, "  [!] reviewer %s (%s) unavailable: %s; trying the next\n", s.Agent, familyName(s.Agent), s.Reason)
}

func familyName(a AgentCLI) string {
	if f := reviewpolicy.FamilyOf(string(a)); f != "" {
		return string(f)
	}
	return "unknown family"
}

// buildArgs returns the argv for the chosen reviewer plus how to capture its
// report. All three modes are READ-ONLY: the reviewer can read + web-search but
// cannot edit the document.
//
//   - codex: `exec --sandbox read-only` writes its final message to a temp file
//     via -o (fromFile=true). The sandbox blocks edits; web access comes from the
//     repo's codex config.
//   - gemini: `-p` is non-interactive, so it cannot approve any write; report is
//     stdout. Search is built in.
//   - claude: `-p` with a read-only --allowedTools allowlist (no Edit/Write/Bash);
//     report is stdout.
func buildArgs(agent AgentCLI, prompt string) (name string, args []string, fromFile bool, tmp string) {
	switch agent {
	case AgentCodex:
		tf, _ := os.CreateTemp("", "doc-review-*.md")
		path := ""
		if tf != nil {
			path = tf.Name()
			tf.Close()
		}
		args = []string{"exec", "--sandbox", "read-only", "--skip-git-repo-check"}
		if path != "" {
			args = append(args, "-o", path)
		}
		args = append(args, prompt)
		return "codex", args, path != "", path

	case AgentGemini:
		return "gemini", []string{"-p", prompt}, false, ""

	case AgentClaude:
		return "claude", []string{
			"-p",
			"--allowedTools", "Read Grep Glob WebSearch WebFetch",
			"--permission-mode", "dontAsk", // #310: reviewers never auto-approve
			prompt,
		}, false, ""

	default:
		return string(agent), []string{prompt}, false, ""
	}
}

// isLaunchFailure distinguishes a real launch failure (binary missing, auth,
// ctx cancelled) from the reviewer running and exiting non-zero (e.g. "found N
// issues, exit 1"). Mirrors dispatch.go's exit-code policy.
func isLaunchFailure(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false
	}
	return true
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func shellJoin(args []string) string {
	parts := make([]string, len(args))
	for i, a := range args {
		if strings.ContainsAny(a, " \t\n'\"$`\\|&;<>(){}*?[]#~=") {
			parts[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		} else {
			parts[i] = a
		}
	}
	return strings.Join(parts, " ")
}
