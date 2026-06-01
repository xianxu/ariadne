// scaffold.go — issue-file creation primitives: next-ID allocation, slug
// derivation, and the canonical new-issue template renderer.
//
// Extracted from cmd/sdlc/fetch.go (ariadne#56) so the same deterministic
// core backs both `sdlc issue new` (blank) and `sdlc issue new --from-github`
// (the old `sdlc fetch`). The agent used to do the ID scan by hand via the
// xx-issues skill — racy under parallel workstreams; this makes it one call.
package issue

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// idPrefixRE matches a leading zero-padded 6-digit issue ID in a filename.
var idPrefixRE = regexp.MustCompile(`^(\d{6})-`)

// hyphenRunRE collapses runs of hyphens produced by slugification.
var hyphenRunRE = regexp.MustCompile(`-+`)

// NextID scans issuesDir + historyDir for filenames starting with a
// 6-digit ID and returns the next ID, zero-padded to 6 chars. Missing
// dirs are treated as empty (so a fresh repo yields "000001").
func NextID(issuesDir, historyDir string) (string, error) {
	max := 0
	for _, dir := range []string{issuesDir, historyDir} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("read %s: %w", dir, err)
		}
		for _, e := range entries {
			m := idPrefixRE.FindStringSubmatch(e.Name())
			if m == nil {
				continue
			}
			n, _ := strconv.Atoi(m[1])
			if n > max {
				max = n
			}
		}
	}
	return fmt.Sprintf("%06d", max+1), nil
}

// Slugify lowercases a title, replaces every non-alphanumeric rune with a
// hyphen, collapses hyphen runs, and trims leading/trailing hyphens.
func Slugify(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	return strings.Trim(hyphenRunRE.ReplaceAllString(b.String(), "-"), "-")
}

// ScaffoldSpec carries the inputs for rendering a new issue file from the
// canonical template. The optional fields default cleanly: empty
// GithubIssue/ProblemBody → a blank issue; populated → the --from-github
// path (the GH body is seeded under ## Problem).
type ScaffoldSpec struct {
	ID          string   // zero-padded 6-digit, e.g. "000057"
	Title       string   // the # heading
	Today       string   // ISO date for created/updated + the Log entry
	GithubIssue string   // "" for a blank issue
	ProblemBody string   // "" for blank; the GH issue body for --from-github
	Deps        []string // optional; rendered as deps: [a, b]
	Target      string   // optional; adds a `target:` line when non-empty
}

// Render returns the canonical new-issue file content (trailing newline
// included). This is the single source of truth for the on-disk template;
// `sdlc issue --help` documents the same shape in prose.
//
// A freshly rendered issue is intentionally a skeleton (empty Spec, empty
// Done-when bullet): it sits at status `open` and only has to satisfy
// CheckStructural later, at `sdlc change-code`, once the author fills it in.
func Render(s ScaffoldSpec) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", s.ID)
	b.WriteString("status: open\n")
	fmt.Fprintf(&b, "deps: [%s]\n", strings.Join(s.Deps, ", "))
	if s.GithubIssue != "" {
		fmt.Fprintf(&b, "github_issue: %s\n", s.GithubIssue)
	} else {
		b.WriteString("github_issue:\n") // no trailing space when empty
	}
	if s.Target != "" {
		fmt.Fprintf(&b, "target: %s\n", s.Target)
	}
	fmt.Fprintf(&b, "created: %s\n", s.Today)
	fmt.Fprintf(&b, "updated: %s\n", s.Today)
	b.WriteString("estimate_hours:\n")
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", s.Title)
	b.WriteString("## Problem\n\n")
	if body := strings.TrimSpace(s.ProblemBody); body != "" {
		b.WriteString(body)
		b.WriteString("\n\n")
	}
	b.WriteString("## Spec\n\n")
	b.WriteString("## Done when\n\n-\n\n")
	b.WriteString("## Plan\n\n- [ ]\n\n")
	b.WriteString("## Log\n\n")
	fmt.Fprintf(&b, "### %s\n", s.Today)
	return b.String()
}
