package main

import (
	"fmt"
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
