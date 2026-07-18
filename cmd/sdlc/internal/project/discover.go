package project

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// DiscoverScope selects how far DiscoverByIssueRef looks. The close gate uses
// ActiveOnly (it must not re-tick a `done` project already archived); the
// navigation consumers (`sdlc project find` / `resolve` / parley) use
// ActiveAndArchive so a record resolves regardless of its lifecycle location.
type DiscoverScope int

const (
	ActiveOnly DiscoverScope = iota
	ActiveAndArchive
)

// ProjectMatch is one discovered project record referencing an issue.
type ProjectMatch struct {
	Path    string // absolute path to the project .md
	RepoDir string // absolute repo root that owns the match
	Repo    string // repo basename
	Legacy  bool   // found under the deprecated brain/data/project home
}

// SiblingRepoDirs returns the absolute paths of every directory sibling of
// curRoot's parent — the fleet walk `resolveRepoDir` performs, factored out so
// the cross-repo project discovery reuses it (ARCH-DRY). It applies NO
// filtering: callers that need to exclude non-fleet dirs do so themselves, so
// `resolveRepoDir`'s matching stays behavior-identical.
func SiblingRepoDirs(parentDir string) ([]string, error) {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(parentDir, e.Name()))
		}
	}
	return dirs, nil
}

// isFleetSibling filters out obvious non-fleet directories a fleet-wide glob
// would otherwise scan: backup copies (`*.bak`), the worktree container, and
// dot-dirs. Without this, a stale `metis.bak/workshop/projects/` copy yields a
// duplicate match → a spurious tick (and, in the close path, a spurious peer
// commit). `resolveRepoDir` doesn't need this because it exact-matches one
// basename; the fleet-wide glob does.
func isFleetSibling(base string) bool {
	if strings.HasPrefix(base, ".") {
		return false
	}
	if base == "worktree" {
		return false
	}
	if strings.HasSuffix(base, ".bak") {
		return false
	}
	return true
}

// DiscoverByIssueRef returns every project record across the fleet whose body
// contains the marker "[<repoName>#<issueID>" (open-bracket form matches both
// "[metis#18]" and "[metis#18 M1]"). It always scans each sibling repo's
// canonical home (vocab.Project().Discovery().Home) and, when scope is
// ActiveAndArchive, the derived archive home; it also scans the deprecated
// brain/data/project/*.md legacy home. Multiple matches are legitimate
// membership, not an error. Deterministic given parentDir.
func DiscoverByIssueRef(parentDir, repoName, issueID string, scope DiscoverScope) ([]ProjectMatch, error) {
	marker := "[" + repoName + "#" + issueID
	disc := vocab.Project().Discovery()
	home := disc.Home                                                     // "workshop/projects"
	archive := vocab.ArchiveSubdir(disc.Archive, vocab.ArchiveProjects)   // "workshop/history/projects"

	var out []ProjectMatch
	seen := map[string]bool{}
	scan := func(repoDir, relDir string, legacy bool) {
		files, _ := filepath.Glob(filepath.Join(repoDir, relDir, "*.md"))
		for _, f := range files {
			real, evErr := filepath.EvalSymlinks(f)
			if evErr != nil || real == "" {
				real = f
			}
			if seen[real] {
				continue
			}
			data, rerr := os.ReadFile(f)
			if rerr != nil {
				continue // best-effort, matches FindByIssueRef
			}
			if strings.Contains(string(data), marker) {
				seen[real] = true
				out = append(out, ProjectMatch{
					Path: f, RepoDir: repoDir, Repo: filepath.Base(repoDir), Legacy: legacy,
				})
			}
		}
	}

	siblings, err := SiblingRepoDirs(parentDir)
	if err != nil {
		return nil, err
	}
	for _, repoDir := range siblings {
		if !isFleetSibling(filepath.Base(repoDir)) {
			continue
		}
		// Brain is identified by the canonical .brain/config.md predicate, not a
		// basename — a brain under any name still holds its projects in the
		// legacy data/project home (and must never be treated as a normal fleet
		// repo the close gate would auto-commit into, #176). Legacy home
		// (deprecated); under ActiveOnly terminal records are dropped below.
		if gitx.IsBrainRepo(repoDir) {
			scan(repoDir, filepath.Join("data", "project"), true)
			continue
		}
		scan(repoDir, home, false)
		if scope == ActiveAndArchive {
			scan(repoDir, archive, false)
		}
	}

	if scope == ActiveOnly {
		out = dropTerminalLegacy(out)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// dropTerminalLegacy removes Legacy (brain) matches whose status is terminal
// (done/dropped). The four migratable brain records are `done` and must not be
// re-ticked by the close gate during the M2→M6 window; a still-active brain
// record (metis-v2) keeps ticking. Non-legacy matches pass through untouched.
func dropTerminalLegacy(matches []ProjectMatch) []ProjectMatch {
	model := vocab.Project()
	kept := matches[:0]
	for _, m := range matches {
		if m.Legacy && legacyIsTerminal(model, m.Path) {
			continue
		}
		kept = append(kept, m)
	}
	return kept
}

func legacyIsTerminal(model *vocab.ProjectModel, path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false // unreadable → keep (best-effort, don't silently drop)
	}
	doc, err := ParseDoc(string(data))
	if err != nil {
		return false
	}
	meta, err := doc.Metadata()
	if err != nil {
		return false
	}
	return model.IsTerminal(meta.Status)
}
