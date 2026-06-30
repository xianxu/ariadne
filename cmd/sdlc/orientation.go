package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// reviewOrientation carries the boundary-review prompt's repo + issue anchors
// (#137), derived from the live git context so a fresh reviewer is oriented to
// the ACTUAL repo under review — never a hardcoded "ariadne". Plain strings,
// passed into the pure judge layer via judge.PromptInput (ARCH-PURE).
type reviewOrientation struct {
	Repo      string // repo name (git-root basename), e.g. "pair"
	RepoRoot  string // absolute git-root path
	IssueRef  string // "<repo>#<num>[ <milestone>]"
	IssueFile string // path to the issue file under review
	Boundary  string // "whole-issue close" | "milestone Mx close"
	RepoNote  string // base-vs-downstream orientation note
}

// boundaryOrientation derives the orientation for a boundary review of issue
// issueNum (milestone "" ⇒ whole-issue close). The repo name is the git-root
// basename; the issue ref is "<repo>#<num>[ <milestone>]". A repo is the ariadne
// base layer iff construct/base.manifest exists at its root (downstream repos
// carry construct/ but not the manifest); the note tells the reviewer to apply
// THIS repo's conventions. Best-effort — falls back to "<unknown-repo>" rather
// than blocking the review.
func boundaryOrientation(issuesDir string, issueNum int, milestone string) reviewOrientation {
	name, root := repoNameAndRoot()
	if name == "" {
		name = "<unknown-repo>"
	}

	ref := fmt.Sprintf("%s#%d", name, issueNum)
	boundary := "whole-issue close"
	if milestone != "" {
		ref += " " + milestone
		boundary = "milestone " + milestone + " close"
	}

	issueFile := ""
	if p, err := issueFilePath(issuesDir, issueNum); err == nil {
		issueFile = p
	}

	note := "a downstream repo built on the ariadne base layer; apply THIS repo's conventions and tracker, not ariadne's"
	if root != "" {
		if _, err := os.Stat(filepath.Join(root, "construct", "base.manifest")); err == nil {
			note = "the ariadne base-layer repo itself (changes here propagate to dependent repos)"
		}
	}

	return reviewOrientation{
		Repo:      name,
		RepoRoot:  root,
		IssueRef:  ref,
		IssueFile: issueFile,
		Boundary:  boundary,
		RepoNote:  note,
	}
}
