package main

import (
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/pkg/sidedoc"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// These rules are shared by filesystem archives and immutable remote archives.
func archiveDestination(historyDir string, kind vocab.ArchiveKind, basename string) string {
	return filepath.Join(vocab.ArchiveSubdir(historyDir, kind), filepath.Base(basename))
}
func planArtifactBelongsToIssue(issueBase, artifactBase string) bool {
	id := issueIDPrefix(issueBase)
	return id != "" && filepath.Base(artifactBase) == artifactBase && strings.HasPrefix(artifactBase, id+"-")
}
func publishedIssueContent(frontmatter, body, date string) ([]byte, error) {
	status, _ := issue.GetField(frontmatter, "status")
	if status != "codecomplete" {
		return nil, fmt.Errorf("cannot publish issue status %q; expected codecomplete", status)
	}
	frontmatter = issue.SetField(frontmatter, "status", "done")
	frontmatter = issue.SetField(frontmatter, "updated", date)
	return []byte(issue.Compose(frontmatter, body)), nil
}

// validateArchiveSideNames applies one policy to working-tree and pinned Git
// inventories (#532). Each path is inspected once; membership is an ID lookup.
func validateArchiveSideNames(paths, issues []string, issuesDir, plansDir string) error {
	ids := map[string]bool{}
	for _, issue := range issues {
		ids[issueIDPrefix(filepath.Base(issue))] = true
	}
	for _, name := range paths {
		for _, dir := range []string{issuesDir, plansDir} {
			rel, ok := strings.CutPrefix(filepath.ToSlash(name), filepath.ToSlash(filepath.Clean(dir))+"/")
			if !ok {
				continue
			}
			folder, _, _ := strings.Cut(rel, "/")
			if !strings.HasSuffix(folder, ".md.side") || !ids[issueIDPrefix(strings.TrimSuffix(folder, ".side"))] {
				continue
			}
			bad := path.Join(filepath.ToSlash(dir), folder)
			canonical := sidedoc.SideDir(strings.TrimSuffix(bad, ".side"))
			return fmt.Errorf("noncanonical side folder %s: rename it to %s, preserve its contents and update links, then retry", bad, canonical)
		}
	}
	return nil
}

// checkCloseSideNames reads only the two active roots, including untracked
// entries. It never follows a folder symlink or reads another issue's history.
func checkCloseSideNames(issuePath, plansDir string, readDir func(string) ([]os.DirEntry, error)) error {
	issuesDir := filepath.Dir(issuePath)
	var paths []string
	for _, dir := range []string{issuesDir, plansDir} {
		entries, err := readDir(dir)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read side-folder names in %s: %w", dir, err)
		}
		for _, entry := range entries {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	return validateArchiveSideNames(paths, []string{issuePath}, issuesDir, plansDir)
}
