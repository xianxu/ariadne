// validategate.go — the #124 instance-conformance gate, run by `sdlc push` +
// `sdlc merge` before the irreversible action and INDEPENDENTLY of the LLM judges
// (so --no-judge doesn't skip it, and --no-validate doesn't skip the judges).
//
// It is a DETERMINISTIC hard check, not a judge:
//   - FRONTMATTER conformance (cue, via the `vocabulary validate-instance` binary)
//     on EVERY changed modeled instance (added or modified) — the universal
//     invariant that catches a hand-edited bad status on an existing record.
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
	"github.com/xianxu/ariadne/pkg/vocab"
)

// Seams — swapped in tests so the gate runs hermetically (no git, no vocabulary
// binary). Production points them at the real implementations.
var (
	diffNameStatusFn           = gitx.DiffNameStatus
	validateFrontmatterFn      = shellValidateFrontmatter
	readIssueFileFn            = os.ReadFile
	validateChangedInstancesFn = validateChangedInstances
)

// nounGate binds one vocabulary noun to the repo directory containing its
// instances. Only issues carry the legacy section-presence policy.
type nounGate struct {
	noun          string
	dir           string
	checkSections bool
}

func nounGates(issuesDir string) []nounGate {
	if issuesDir == "" {
		issuesDir = envOr("WF_ISSUES_DIR", vocab.Issue().Discovery().Home)
	}
	return []nounGate{
		{noun: "issue", dir: issuesDir, checkSections: true},
		{noun: "project", dir: vocab.Project().Discovery().Home},
	}
}

// validateChangedInstances is the fail-closed gate. base/head are the caller's
// review window; gates declare which changed paths derive from which noun model.
func validateChangedInstances(base, head string, gates []nounGate, stdout, stderr io.Writer) error {
	changes, err := diffNameStatusFn(base, head)
	if err != nil {
		return fmt.Errorf("instance-conformance gate: %w", err)
	}

	var problems []string
	checked := 0
	for _, ch := range changes {
		gate, ok := gateForPath(ch.Path, gates)
		if ch.Status == "D" || !ok {
			continue
		}
		checked++

		// Frontmatter — every changed instance (added OR modified).
		out, conforms, runErr := validateFrontmatterFn(gate.noun, ch.Path)
		if runErr != nil {
			// Could not RUN the validator (binary missing) — a setup failure, not a
			// conformance verdict. The GATE fails closed (hard return); the on-demand
			// `sdlc issue validate` (validateIssueFull) deliberately differs — it treats
			// can't-run as a per-file problem and continues, since it's informative.
			return fmt.Errorf("instance-conformance gate could not run on %s: %w", ch.Path, runErr)
		}
		if !conforms {
			problems = append(problems, ch.Path+" (frontmatter):\n"+indentLines(strings.TrimSpace(out), "      "))
		}

		// Sections — newly-ADDED files only (grandfather legacy/in-flight; a rename "R"
		// is NOT "A", so a renamed/archived ticket is never section-validated).
		if ch.Status == "A" && gate.checkSections {
			if data, rerr := readIssueFileFn(ch.Path); rerr == nil {
				for _, f := range issue.CheckSectionsPresence(string(data)) {
					problems = append(problems, ch.Path+" (section): "+f.Message)
				}
			}
		}
	}

	if len(problems) > 0 {
		cwarn(stderr, fmt.Sprintf("instance-conformance gate: %d nonconforming changed instance file(s) — fix and re-run, or --no-validate to bypass (loud):", len(problems)))
		for _, p := range problems {
			fmt.Fprintln(stdout, "  - "+p)
		}
		return fmt.Errorf("instance-conformance gate: %d nonconforming instance file(s)", len(problems))
	}
	cok(stderr, fmt.Sprintf("instance-conformance gate: %d changed instance file(s) conform", checked))
	return nil
}

// shellValidateFrontmatter runs `vocabulary validate-instance --type <noun> <file>`.
// ok=false (+ diagnostics in output) = nonconforming; err != nil = the validator
// could not RUN (e.g. binary not on PATH) — a setup failure distinct from
// nonconformance, surfaced loudly so the operator builds the binary or --no-validate.
func shellValidateFrontmatter(noun, file string) (output string, ok bool, err error) {
	out, runErr := exec.Command("vocabulary", "validate-instance", "--type", noun, file).CombinedOutput()
	if runErr == nil {
		return string(out), true, nil
	}
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		return string(out), false, nil // ran, exited non-zero → nonconforming
	}
	return string(out), false, fmt.Errorf("`vocabulary validate-instance` did not run (build the vocabulary binary onto PATH, or pass --no-validate): %w", runErr)
}

func gateForPath(path string, gates []nounGate) (nounGate, bool) {
	for _, gate := range gates {
		if isInstanceFile(path, gate.dir) {
			return gate, true
		}
	}
	return nounGate{}, false
}

// isInstanceFile reports whether path is a markdown file below dir.
func isInstanceFile(path, dir string) bool {
	dir = strings.TrimSuffix(filepath.ToSlash(dir), "/") + "/"
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
