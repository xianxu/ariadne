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

// walkFleetProjects is the shared fleet walk (#182): for every fleet sibling
// under parentDir, it visits each project-home *.md (deduped by resolved path)
// and calls visit(path, repoDir, legacy). Homes: each non-brain repo's
// canonical home, plus — when includeArchive — the derived archive home; a
// brain repo (canonical .brain/config.md predicate, #176) contributes only its
// deprecated data/project legacy home. Non-fleet siblings (.bak, worktree,
// dot-dirs) are skipped. Both DiscoverByIssueRef and ListActiveProjectFiles
// drive off this one walk so "where the fleet's projects live" has one source.
func walkFleetProjects(parentDir string, includeArchive bool, visit func(path, repoDir string, legacy bool)) error {
	disc := vocab.Project().Discovery()
	home := disc.Home                                                   // "workshop/projects"
	archive := vocab.ArchiveSubdir(disc.Archive, vocab.ArchiveProjects) // "workshop/history/projects"

	siblings, err := SiblingRepoDirs(parentDir)
	if err != nil {
		return err
	}
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
			seen[real] = true
			visit(f, repoDir, legacy)
		}
	}
	for _, repoDir := range siblings {
		if !isFleetSibling(filepath.Base(repoDir)) {
			continue
		}
		if gitx.IsBrainRepo(repoDir) {
			scan(repoDir, filepath.Join("data", "project"), true)
			continue
		}
		scan(repoDir, home, false)
		if includeArchive {
			scan(repoDir, archive, false)
		}
	}
	return nil
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
	var out []ProjectMatch
	err := walkFleetProjects(parentDir, scope == ActiveAndArchive, func(path, repoDir string, legacy bool) {
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return // best-effort, matches FindByIssueRef
		}
		if containsIssueMarker(string(data), marker) {
			out = append(out, ProjectMatch{
				Path: path, RepoDir: repoDir, Repo: filepath.Base(repoDir), Legacy: legacy,
			})
		}
	})
	if err != nil {
		return nil, err
	}
	if scope == ActiveOnly {
		out = dropTerminalLegacy(out)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// ProjectFile is one active project record's location, from the fleet walk.
type ProjectFile struct {
	Path    string
	RepoDir string
	Repo    string
	Legacy  bool // deprecated brain/data/project home
}

// ListActiveProjectFiles returns every ACTIVE project record across the fleet
// (canonical home + brain legacy home; NOT the archive — archived/terminal
// records hold no current load), excluding excludePath (resolved via
// EvalSymlinks so a symlink to the subject still matches). Reuses the shared
// walkFleetProjects so the fleet enumeration has one source (#182 M2; the
// calendar forecast's contention input). Deterministic (sorted by path).
func ListActiveProjectFiles(parentDir, excludePath string) ([]ProjectFile, error) {
	exclude := excludePath
	if r, err := filepath.EvalSymlinks(excludePath); err == nil && r != "" {
		exclude = r
	}
	var out []ProjectFile
	err := walkFleetProjects(parentDir, false, func(path, repoDir string, legacy bool) {
		real := path
		if r, evErr := filepath.EvalSymlinks(path); evErr == nil && r != "" {
			real = r
		}
		if real == exclude {
			return
		}
		out = append(out, ProjectFile{Path: path, RepoDir: repoDir, Repo: filepath.Base(repoDir), Legacy: legacy})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// containsIssueMarker reports whether text contains the ref marker followed by
// a NON-DIGIT boundary — "[metis#18]" or "[metis#18 M2]", never "[metis#180]".
// A bare Contains on the open-bracket prefix would false-positive on longer
// ids sharing the prefix (#171 M4 review: a close of #18 must not tick #180's
// project, and `project find` must not navigate to it).
func containsIssueMarker(text, marker string) bool {
	for i := 0; ; {
		j := strings.Index(text[i:], marker)
		if j < 0 {
			return false
		}
		k := i + j + len(marker)
		if k >= len(text) || text[k] < '0' || text[k] > '9' {
			return true
		}
		i = k
	}
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
