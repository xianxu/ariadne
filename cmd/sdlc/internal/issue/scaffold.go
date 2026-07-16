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

	"github.com/xianxu/ariadne/pkg/vocab"
)

// idPrefixRE matches a leading zero-padded 6-digit issue ID in a filename.
var idPrefixRE = regexp.MustCompile(`^(\d{6})-`)

// hyphenRunRE collapses runs of hyphens produced by slugification.
var hyphenRunRE = regexp.MustCompile(`-+`)

// NextID scans issuesDir + historyDir (flat legacy) + the archive's issues
// subdir (#181 layout) for filenames starting with a 6-digit ID and returns
// the next ID, zero-padded to 6 chars. Missing dirs are treated as empty (so
// a fresh repo yields "000001"). The plans subdir is skipped — plan/review
// files carry the same ids as their issues, so it can't hold a new max.
func NextID(issuesDir, historyDir string) (string, error) {
	archivedIssues := vocab.ArchiveSubdir(historyDir, vocab.ArchiveIssues)
	max := 0
	for _, dir := range []string{issuesDir, historyDir, archivedIssues} {
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
// included). The section list, their order, their seed placeholders, and the
// initial `status:` are DERIVED from the cue model (`construct/vocabulary/issue.cue`
// `scaffold.sections` + `categories.open`) via vocab.Issue() — not hardcoded here
// (#145). The cue model is the single source of the template shape; this function
// is its renderer. `sdlc issue --help` documents a superset of the same sections
// (a test enforces documented ⊇ modeled).
//
// A freshly rendered issue is intentionally a skeleton (empty Spec, empty
// Done-when bullet): it sits at the initial status (`open`) and only has to
// satisfy CheckStructural later, at `sdlc change-code`, once the author fills it in.
func Render(s ScaffoldSpec) string {
	m := vocab.Issue()
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", s.ID)
	fmt.Fprintf(&b, "status: %s\n", m.InitialStatus())
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

	// Body sections come from the cue model: their names, order, and static seed
	// placeholders live in issue.cue `scaffold.sections`. Two sections carry
	// DYNAMIC creation content that stays here, keyed by name — keep these names
	// in sync with the model (TestScaffold_SpecialSectionsPresent pins them).
	sections := m.Sections()
	for i, sec := range sections {
		fmt.Fprintf(&b, "## %s\n\n", sec.Name)
		content := sec.Seed
		switch sec.Name {
		case "Problem":
			content = strings.TrimSpace(s.ProblemBody) // --from-github body, else blank
		case "Log":
			content = fmt.Sprintf("### %s", s.Today) // dated session subheading
		}
		if content == "" {
			continue
		}
		b.WriteString(content)
		if i < len(sections)-1 {
			b.WriteString("\n\n")
		} else {
			b.WriteString("\n") // last section closes the file with a single newline
		}
	}
	return b.String()
}
