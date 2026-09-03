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
	"sort"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// idPrefixRE matches a leading zero-padded 6-digit issue ID in a filename.
var idPrefixRE = regexp.MustCompile(`^(\d{6})-`)

// hyphenRunRE collapses runs of hyphens produced by slugification.
var hyphenRunRE = regexp.MustCompile(`-+`)

// IDDirs returns the three directories that hold id-prefixed issue filenames:
// the live dir, the flat legacy history dir, and the #181 archive subdir. The
// plans subdir is skipped — plan/review files carry their issue's id, so they
// can never hold a new max.
//
// One source for "where do ids live", shared by the local scan and the
// remote-ref scan, so the two cannot look in different places.
func IDDirs(issuesDir, historyDir string) []string {
	return []string{issuesDir, historyDir, vocab.ArchiveSubdir(historyDir, vocab.ArchiveIssues)}
}

// IDFromFilename returns the 6-digit id a filename carries, or 0 when it
// carries none. Pure; the single parser of the id prefix.
func IDFromFilename(name string) int {
	m := idPrefixRE.FindStringSubmatch(name)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// NextID returns the id after the highest in any of the given sets, zero-padded
// to 6 chars. No sets, or all empty, yields "000001".
//
// PURE (#213): it takes id sets rather than reading directories, because the
// defect it exists to fix is that the LOCAL directory is the wrong source. A
// branch cut before an issue landed on the trunk has a workshop/issues/ that
// never contained it, so a local-only scan reallocates a published id — silently,
// since the two files get different slugs and therefore never collide as paths.
//
// The caller unions the local scan with the trunk's ids. Union, not replacement:
// unpushed issues on this branch are real and must still be excluded.
func NextID(idSets ...[]int) string {
	max := 0
	for _, set := range idSets {
		for _, n := range set {
			if n > max {
				max = n
			}
		}
	}
	return fmt.Sprintf("%06d", max+1)
}

// ScanLocalIDs reads the on-disk issue directories and returns every id found.
// Missing dirs are treated as empty, so a fresh repo yields none. The thin IO
// half of allocation; NextID does the deciding.
func ScanLocalIDs(issuesDir, historyDir string) ([]int, error) {
	var ids []int
	for _, dir := range IDDirs(issuesDir, historyDir) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", dir, err)
		}
		for _, e := range entries {
			if n := IDFromFilename(e.Name()); n > 0 {
				ids = append(ids, n)
			}
		}
	}
	return ids, nil
}

// IDsInTreeListing parses `git ls-tree --name-only <ref> <dir>/` output into ids.
// Pure, so the remote-ref path is testable without a repo; the caller runs git.
func IDsInTreeListing(out string) []int {
	var ids []int
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i := strings.LastIndex(line, "/"); i >= 0 {
			line = line[i+1:]
		}
		if n := IDFromFilename(line); n > 0 {
			ids = append(ids, n)
		}
	}
	return ids
}

// IDCollision is one id claimed by more than one file within a single tree.
type IDCollision struct {
	ID    int
	Paths []string
}

// DuplicateIDsInRef finds ids claimed by two or more DIFFERENT paths in one
// ls-tree listing, sorted by id.
//
// A different question from "does this branch reuse a trunk id", and the only one
// that can see a collision already merged (#213): once both files are on the
// trunk, a branch-vs-trunk comparison finds two agreeing trees and nothing to
// report. This asks whether a single tree contradicts itself.
//
// Pure — takes the listing, so the git call stays in the caller's IO shell.
func DuplicateIDsInRef(listing string) []IDCollision {
	byID := map[int][]string{}
	var order []int
	for _, line := range strings.Split(listing, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		base := line
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		n := IDFromFilename(base)
		if n == 0 {
			continue
		}
		// The same path seen twice (overlapping dirs in the listing) is not a
		// collision — only two DIFFERENT paths claiming one id are.
		dup := false
		for _, seen := range byID[n] {
			if seen == line {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		if len(byID[n]) == 0 {
			order = append(order, n)
		}
		byID[n] = append(byID[n], line)
	}
	var out []IDCollision
	for _, id := range order {
		if len(byID[id]) > 1 {
			out = append(out, IDCollision{ID: id, Paths: byID[id]})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
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
