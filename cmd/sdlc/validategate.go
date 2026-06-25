// validategate.go — the #124 instance-conformance gate, run by `sdlc push` +
// `sdlc merge` before the irreversible action and INDEPENDENTLY of the LLM judges
// (so --no-judge doesn't skip it, and --no-validate doesn't skip the judges).
//
// It is a DETERMINISTIC hard check, not a judge:
//   - FRONTMATTER conformance (cue, via the `vocabulary validate-instance` binary)
//     on EVERY changed issue file (added or modified) — the universal invariant that
//     catches the motivating hand-edited bad `status:` even on an existing ticket.
//   - SECTION presence (issue.CheckSectionsPresence, the SAME policy the change-code
//     structural gate uses — single source) on NEWLY-ADDED files only. New issues
//     must be well-formed; pre-existing/legacy/in-flight tickets are grandfathered
//     (#124: "validate forward, don't fail old tickets"). A rename is not "added".
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Seams — swapped in tests so the gate runs hermetically (no git, no vocabulary
// binary). Production points them at the real implementations.
var (
	diffNameStatusFn        = gitx.DiffNameStatus
	validateFrontmatterFn   = shellValidateFrontmatter
	readIssueFileFn         = os.ReadFile
	validateChangedIssuesFn = validateChangedIssues
)

// validateChangedIssues is the fail-closed gate. base/head are the caller's window
// (the SAME one the judges use — don't recompute, per the M1 review). Returns an
// error naming every nonconforming changed issue file; nil when all conform.
func validateChangedIssues(base, head, issuesDir string, stdout, stderr io.Writer) error {
	if issuesDir == "" {
		issuesDir = envOr("WF_ISSUES_DIR", "workshop/issues")
	}
	changes, err := diffNameStatusFn(base, head)
	if err != nil {
		return fmt.Errorf("instance-conformance gate: %w", err)
	}

	var problems []string
	checked := 0
	for _, ch := range changes {
		if ch.Status == "D" || !isIssueFile(ch.Path, issuesDir) {
			continue
		}
		checked++

		// Frontmatter — every changed issue (added OR modified): the universal invariant.
		out, ok, runErr := validateFrontmatterFn(ch.Path)
		if runErr != nil {
			// Could not RUN the validator (binary missing) — a setup failure, not a
			// conformance verdict. The GATE fails closed (hard return); the on-demand
			// `sdlc issue validate` (validateIssueFull) deliberately differs — it treats
			// can't-run as a per-file problem and continues, since it's informative.
			return fmt.Errorf("instance-conformance gate could not run on %s: %w", ch.Path, runErr)
		}
		if !ok {
			problems = append(problems, ch.Path+" (frontmatter):\n"+indentLines(strings.TrimSpace(out), "      "))
		}

		// Sections — newly-ADDED files only (grandfather legacy/in-flight; a rename "R"
		// is NOT "A", so a renamed/archived ticket is never section-validated).
		if ch.Status == "A" {
			if data, rerr := readIssueFileFn(ch.Path); rerr == nil {
				for _, f := range issue.CheckSectionsPresence(string(data)) {
					problems = append(problems, ch.Path+" (section): "+f.Message)
				}
			}
		}
	}

	if len(problems) > 0 {
		cwarn(stderr, fmt.Sprintf("instance-conformance gate: %d nonconforming changed issue file(s) — fix and re-run, or --no-validate to bypass (loud):", len(problems)))
		for _, p := range problems {
			fmt.Fprintln(stdout, "  - "+p)
		}
		return fmt.Errorf("instance-conformance gate: %d nonconforming issue file(s)", len(problems))
	}
	cok(stderr, fmt.Sprintf("instance-conformance gate: %d changed issue file(s) conform", checked))
	return nil
}

// shellValidateFrontmatter runs `vocabulary validate-instance --type issue <file>`.
// ok=false (+ diagnostics in output) = nonconforming; err != nil = the validator
// could not RUN (e.g. binary not on PATH) — a setup failure distinct from
// nonconformance, surfaced loudly so the operator builds the binary or --no-validate.
func shellValidateFrontmatter(file string) (output string, ok bool, err error) {
	out, runErr := exec.Command("vocabulary", "validate-instance", "--type", "issue", file).CombinedOutput()
	if runErr == nil {
		return string(out), true, nil
	}
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		return string(out), false, nil // ran, exited non-zero → nonconforming
	}
	return string(out), false, fmt.Errorf("`vocabulary validate-instance` did not run (build the vocabulary binary onto PATH, or pass --no-validate): %w", runErr)
}

// isIssueFile reports whether path is a `.md` under issuesDir (prefix match at any
// depth — issue files are flat today, but a nested one would still be validated).
func isIssueFile(path, issuesDir string) bool {
	dir := strings.TrimSuffix(filepath.ToSlash(issuesDir), "/") + "/"
	p := filepath.ToSlash(path)
	return strings.HasPrefix(p, dir) && strings.HasSuffix(p, ".md")
}

func indentLines(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
