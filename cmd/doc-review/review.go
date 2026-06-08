package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AgentCLI names the second-vendor reviewer CLI. Mirrors the set in
// cmd/sdlc/internal/judge/dispatch.go (claude/codex/gemini); we keep our own
// copy rather than import across the sdlc internal/ boundary because the review
// contract differs from the judge contract (read-only + report capture, not
// --full-auto) — a thin, intentional duplication (ARCH-DRY: noted, not shared).
type AgentCLI string

const (
	AgentCodex  AgentCLI = "codex"
	AgentGemini AgentCLI = "gemini"
	AgentClaude AgentCLI = "claude"

	// DefaultAgent is a different vendor from the usual co-authoring agent
	// (Claude), satisfying the cross-model requirement out of the box.
	DefaultAgent = AgentCodex
)

func (a AgentCLI) known() bool {
	switch a {
	case AgentCodex, AgentGemini, AgentClaude:
		return true
	}
	return false
}

type reviewFlags struct {
	Agent  AgentCLI
	File   string
	Out    string
	DryRun bool
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

// runReview is the binary's one job: build the baked review prompt for <file>,
// dispatch the chosen agent READ-ONLY, capture its report, and write the sidecar
// — then tell the main agent to triage. The reviewer never edits the doc; this
// orchestrator does the only write (the report).
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

	out := f.Out
	if out == "" {
		out = reportPath(abs, f.Agent)
	}
	prompt := buildPrompt(abs)

	name, args, fromFile, tmp := buildArgs(f.Agent, prompt)
	if tmp != "" {
		defer os.Remove(tmp)
	}

	if f.DryRun {
		fmt.Fprintf(stdout, "agent:  %s\n", name)
		fmt.Fprintf(stdout, "argv:   %s %s\n", name, shellJoin(args))
		fmt.Fprintf(stdout, "review: %s\n", abs)
		fmt.Fprintf(stdout, "report: %s\n", out)
		return nil
	}

	fmt.Fprintf(stderr, "doc-review: dispatching %s (read-only) to review %s …\n", name, abs)
	raw, runErr := runAgent(context.Background(), name, args...)
	if isLaunchFailure(runErr) {
		hint := ""
		if f.Agent == AgentCodex {
			// codex has crashed on some setups (gpt-image-2 bug) — the
			// documented fallback is a different vendor.
			hint = "\n  codex can crash on some setups (gpt-image-2 bug); retry with a different vendor:\n    doc-review gemini " + f.File
		}
		return fmt.Errorf("%s failed to run: %w%s", name, runErr, hint)
	}

	report := string(raw)
	if fromFile {
		b, rerr := os.ReadFile(tmp)
		if rerr != nil {
			return fmt.Errorf("%s produced no report file (output was:\n%s)", name, truncate(string(raw), 800))
		}
		report = string(b)
	}
	report = strings.TrimSpace(report)
	if report == "" {
		return fmt.Errorf("%s returned an empty report (likely a launch/auth problem); raw output:\n%s", name, truncate(string(raw), 800))
	}

	if err := os.WriteFile(out, []byte(wrapReport(filepath.Base(f.File), f.Agent, report)), 0o644); err != nil {
		return fmt.Errorf("write report %s: %w", out, err)
	}

	fmt.Fprintf(stdout, "\nReport written: %s\n", out)
	fmt.Fprintf(stdout, "\nNEXT (main agent): read the report, triage each finding, and update %s where you AGREE.\n"+
		"The review is advisory and read-only — you own the document. Do not apply findings blindly;\n"+
		"verify the claimed corrections before editing, and leave a note for ones you reject.\n", filepath.Base(f.File))
	return nil
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
			"--permission-mode", "bypassPermissions",
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
