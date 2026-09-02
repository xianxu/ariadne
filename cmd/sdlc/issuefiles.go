package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const issueFilenamePattern = issue.FilenamePattern

type issueFileRef struct {
	Path        string
	Status      string
	Frontmatter string
	Body        string
}

type issueFileScanError struct {
	Output []byte
	Err    error
}

func (e *issueFileScanError) Error() string { return e.Err.Error() }

func (e *issueFileScanError) Unwrap() error { return e.Err }

func scanIssueFiles(baseRef, issuesDir string, runGit func(...string) ([]byte, error)) ([]issueFileRef, error) {
	var paths []string
	if baseRef != "" {
		out, err := runGit("diff", "--name-only", baseRef+"..HEAD", "--", issuesDir+"/*.md")
		if err != nil {
			return nil, &issueFileScanError{Output: out, Err: err}
		}
		paths = splitNonEmptyLines(string(out))
	} else {
		paths, _ = filepath.Glob(filepath.Join(issuesDir, issueFilenamePattern))
		sort.Strings(paths)
	}

	refs := make([]issueFileRef, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fm, body, err := issue.Parse(string(data))
		if err != nil {
			continue
		}
		status, _ := issue.GetField(fm, "status")
		refs = append(refs, issueFileRef{
			Path:        path,
			Status:      status,
			Frontmatter: fm,
			Body:        body,
		})
	}
	return refs, nil
}

func issueFilenameParts(name string) (id, slug string, ok bool) {
	return issue.ParseFilename(name)
}

func issueFilename(name string) bool {
	_, _, ok := issueFilenameParts(name)
	return ok
}

func codecompleteIssueFiles(refs []issueFileRef) []issueFileRef {
	return filterIssueFiles(refs, func(ref issueFileRef) bool {
		return ref.Status == "codecomplete"
	})
}

func notDoneIssueFiles(refs []issueFileRef) []issueFileRef {
	return filterIssueFiles(refs, func(ref issueFileRef) bool {
		return ref.Status != "codecomplete" && !vocab.Issue().IsTerminal(ref.Status)
	})
}

func terminalIssueFiles(refs []issueFileRef) []issueFileRef {
	return filterIssueFiles(refs, func(ref issueFileRef) bool {
		return vocab.Issue().IsTerminal(ref.Status)
	})
}

func filterIssueFiles(refs []issueFileRef, keep func(issueFileRef) bool) []issueFileRef {
	var filtered []issueFileRef
	for _, ref := range refs {
		if keep(ref) {
			filtered = append(filtered, ref)
		}
	}
	return filtered
}

// issueFilesForID globs the NNNNNN-*.md files belonging to one issue id. The
// single source for "which files are issue N's" as a PATH question — shared by
// resolveBranchName (which file names the branch) and syncPathspec (which files
// a sync stages and commits), so the two cannot disagree about what an --issue
// filter means. changedIssueFiles answers a different question (which of the
// already-changed paths belong to N) by prefix-matching that same convention.
func issueFilesForID(issuesDir string, id int) []string {
	matches, _ := filepath.Glob(filepath.Join(issuesDir, fmt.Sprintf("%06d", id)+"-*.md"))
	return matches
}

// issueIDFromPath returns the issue id encoded in an issue file's name, or 0
// when the name doesn't follow the NNNNNN- convention (a --name branch pointed
// at a non-issue file). Reuses issueIDPrefix so the convention is parsed once.
func issueIDFromPath(path string) int {
	id, err := strconv.Atoi(issueIDPrefix(filepath.Base(path)))
	if err != nil {
		return 0
	}
	return id
}
