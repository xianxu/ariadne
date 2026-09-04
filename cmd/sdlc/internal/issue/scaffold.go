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
	"path/filepath"
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

// LocalPathsByID reads id-bearing filenames from directories on disk into the
// SAME id→paths shape a ref read produces, so "what this checkout claims" and
// "what the trunk claims" are the same kind of answer and can be compared
// without either side reparsing.
//
// It takes RESOLVED directories rather than deriving them (#213 BR-25): the
// caller resolves once, against the repo top level, so the local scan and the
// trunk read cannot look in different places.
//
// `found` counts the directories that actually existed. A scan that found NONE
// is blind, not empty — the caller must say so rather than reporting an empty
// id space as fact.
func LocalPathsByID(dirs []string) (byID map[int][]string, found int, err error) {
	byID = map[int][]string{}
	for _, dir := range dirs {
		entries, rerr := os.ReadDir(dir)
		if rerr != nil {
			if os.IsNotExist(rerr) {
				continue
			}
			return nil, found, fmt.Errorf("read %s: %w", dir, rerr)
		}
		found++
		for _, e := range entries {
			if n := IDFromFilename(e.Name()); n > 0 {
				addPath(byID, n, filepath.Join(dir, e.Name()))
			}
		}
	}
	return byID, found, nil
}

// PathsByID parses `git ls-tree --name-only <ref> <dir>/` output into id →
// every DISTINCT path claiming it.
//
// THE parser of an id listing (#213). There were three — one yielding ids for
// allocation, one yielding paths for the merge gate, one yielding duplicates
// for the lint verb — each with its own copy of "strip the directory, read the
// prefix, skip a path already seen". Allocation, collision detection and the
// merge predicate now all derive from this one map, so they cannot disagree
// about what a listing says.
//
// Pure, so every ref-reading path is testable without a repo; the caller runs
// git.
func PathsByID(listing string) map[int][]string {
	byID := map[int][]string{}
	for _, line := range strings.Split(listing, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name := line
		if i := strings.LastIndex(name, "/"); i >= 0 {
			name = name[i+1:]
		}
		if n := IDFromFilename(name); n > 0 {
			addPath(byID, n, line)
		}
	}
	return byID
}

// addPath records a path under an id, ignoring one already seen. The same path
// can arrive twice when the id directories overlap (workshop/history and its
// issues/ subdir), and that is not two claimants.
func addPath(byID map[int][]string, id int, path string) {
	for _, seen := range byID[id] {
		if seen == path {
			return
		}
	}
	byID[id] = append(byID[id], path)
}

// IDsIn returns every id in an id space, sorted.
func IDsIn(byID map[int][]string) []int {
	ids := make([]int, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// IDCollision is one id claimed by more than one file within a single tree.
type IDCollision struct {
	ID    int
	Paths []string
}

// DuplicatesIn finds ids claimed by two or more DIFFERENT paths within one id
// space, sorted by id.
//
// This is the only way an ALREADY-MERGED collision is visible at all: a
// branch-vs-trunk comparison sees two trees that agree and finds nothing, so
// the eight live collisions this issue was opened over were invisible to every
// range-based check.
func DuplicatesIn(byID map[int][]string) []IDCollision {
	var out []IDCollision
	for _, id := range IDsIn(byID) {
		if len(byID[id]) > 1 {
			paths := append([]string{}, byID[id]...)
			sort.Strings(paths)
			out = append(out, IDCollision{ID: id, Paths: paths})
		}
	}
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
