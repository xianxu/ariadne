package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// sidecarMeta is everything a fresh reader needs to orient on a persisted
// boundary review (#136): which issue/repo, which boundary + window, who
// reviewed, when, the verdict, and the semantic final review body. All fields are plain
// values — the clock (Timestamp) is captured at the IO boundary and passed in so
// the renderers below stay pure (ARCH-PURE).
type sidecarMeta struct {
	IssueNum  int
	Title     string
	Repo      string
	IssueFile string
	Milestone string // "" ⇒ whole-issue close
	Base      string // long base SHA
	Head      string
	Command   string
	Agent     string
	Timestamp string // RFC3339
	Verdict   string
	Body      string
}

// sidecarPathFor derives a sidecar path from an issue filename stem plus a suffix. The one
// stem derivation shared by the boundary-review sidecar (#136) and the plan-gate ledger
// (#187), so both track the issue's slug from a single source (ARCH-DRY).
func sidecarPathFor(plansDir, issueFileName, suffix string) string {
	stem := strings.TrimSuffix(filepath.Base(issueFileName), ".md")
	return filepath.Join(plansDir, stem+"-"+suffix+".md")
}

// sidecarPath derives the deterministic sidecar path from the issue filename
// stem: `NNNNNN-slug-close-review.md` for a whole-issue close, or
// `NNNNNN-slug-m<x>-review.md` for milestone `Mx` (lowercased). Reusing the
// issue filename's stem keeps one slug source of truth (ARCH-DRY).
func sidecarPath(plansDir, issueFileName, milestone string) string {
	suffix := "close-review"
	if milestone != "" {
		suffix = strings.ToLower(milestone) + "-review"
	}
	return sidecarPathFor(plansDir, issueFileName, suffix)
}

// boundaryKind is the human label for the review boundary.
func boundaryKind(milestone string) string {
	if milestone == "" {
		return "whole-issue close"
	}
	return "milestone " + milestone
}

// sidecarCommand reconstructs the invoking command line for the metadata.
func sidecarCommand(issueNum int, milestone string) string {
	if milestone == "" {
		return fmt.Sprintf("sdlc close --issue %d", issueNum)
	}
	return fmt.Sprintf("sdlc milestone-close --issue %d --milestone %s", issueNum, milestone)
}

// renderReviewEntry renders one review as self-contained markdown. The initial
// write (isRevision=false) emits an H1 + metadata table + `## Review` + body; a
// re-run (isRevision=true) emits a `## Re-review` section (no H1) for appending
// under a `---` separator, so re-runs preserve prior evidence (#136 D2). Pure.
func renderReviewEntry(m sidecarMeta, isRevision bool) string {
	var b strings.Builder
	if isRevision {
		fmt.Fprintf(&b, "## Re-review — %s (%s)\n\n", m.Timestamp, m.Verdict)
	} else {
		// Repo-derived, not hardcoded "ariadne" (#137) — a downstream review's
		// durable artifact must name its own repo, matching the | repo | cell.
		repo := m.Repo
		if repo == "" {
			repo = "<unknown-repo>"
		}
		fmt.Fprintf(&b, "# Boundary Review — %s#%d (%s)\n\n", repo, m.IssueNum, boundaryKind(m.Milestone))
	}
	milestoneCell := m.Milestone
	if milestoneCell == "" {
		milestoneCell = "—"
	}
	b.WriteString("| field | value |\n|-------|-------|\n")
	fmt.Fprintf(&b, "| issue | %d — %s |\n", m.IssueNum, m.Title)
	fmt.Fprintf(&b, "| repo | %s |\n", m.Repo)
	fmt.Fprintf(&b, "| issue file | %s |\n", m.IssueFile)
	fmt.Fprintf(&b, "| boundary | %s |\n", boundaryKind(m.Milestone))
	fmt.Fprintf(&b, "| milestone | %s |\n", milestoneCell)
	fmt.Fprintf(&b, "| window | %s..%s |\n", m.Base, m.Head)
	fmt.Fprintf(&b, "| command | %s |\n", m.Command)
	fmt.Fprintf(&b, "| reviewer | %s |\n", m.Agent)
	fmt.Fprintf(&b, "| timestamp | %s |\n", m.Timestamp)
	fmt.Fprintf(&b, "| verdict | %s |\n", m.Verdict)
	b.WriteString("\n## Review\n\n")
	b.WriteString(m.Body)
	if !strings.HasSuffix(m.Body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

// repoNameAndRoot resolves the current repo's name (git-root basename) and its
// root path from the live git context, or ("", "") if the root can't be resolved.
// The single derivation site for "which repo are we in" (consolidates the
// duplicated filepath.Base(gitx.RepoTopLevel()) the #136 review flagged).
func repoNameAndRoot() (name, root string) {
	root, err := gitx.RepoTopLevel()
	if err != nil || root == "" {
		return "", ""
	}
	return filepath.Base(root), root
}

// repoIdentity returns the repo's top-level basename (e.g. "ariadne"), or "" if
// it can't be resolved — non-fatal metadata.
func repoIdentity() string {
	name, _ := repoNameAndRoot()
	return name
}

// nowRFC3339 is the single clock touch for the sidecar, isolated here so the
// renderers stay pure and deterministic.
func nowRFC3339() string { return time.Now().Format(time.RFC3339) }

// atomicWriteFile writes data to path via a temp file + rename, creating the
// parent directory if needed. First atomic-write helper in cmd/sdlc.
func atomicWriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// writeReviewSidecar persists one boundary review under workshop/plans/. The
// first write for a boundary creates the document; a re-run for the same
// boundary appends a timestamped `## Re-review` section rather than overwriting
// prior evidence (#136 D2). Returns the sidecar path. The thin IO seam over the
// pure renderers (ARCH-PURE).
func writeReviewSidecar(p boundaryReviewParams, verdict, body, timestamp string) (string, error) {
	issuePath, err := issueFilePath(p.IssuesDir, p.IssueNum)
	if err != nil {
		return "", err
	}
	title := "(no title)"
	if raw, rerr := os.ReadFile(issuePath); rerr == nil {
		title = issueTitleFromContent(string(raw))
	}
	path := sidecarPath(p.PlansDir, filepath.Base(issuePath), p.Milestone)
	m := sidecarMeta{
		IssueNum:  p.IssueNum,
		Title:     title,
		Repo:      repoIdentity(),
		IssueFile: issuePath,
		Milestone: p.Milestone,
		Base:      p.BaseLong,
		Head:      p.Head,
		Command:   sidecarCommand(p.IssueNum, p.Milestone),
		Agent:     p.Agent,
		Timestamp: timestamp,
		Verdict:   verdict,
		Body:      body,
	}
	if prior, rerr := os.ReadFile(path); rerr == nil {
		return path, atomicWriteFile(path, []byte(string(prior)+"\n---\n\n"+renderReviewEntry(m, true)))
	}
	return path, atomicWriteFile(path, []byte(renderReviewEntry(m, false)))
}
