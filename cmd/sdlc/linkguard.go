package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/linkguard"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/sidedoc"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// The side-link guard (#579): before an issue's text reaches main, every
// relative link in it, its plans and the projects that reference it must
// name a committed file. A side doc is committed with its artifact; a
// private, git-ignored one is folded in instead.

// sideLinkGit runs git in a directory, returning stdout.
type sideLinkGit func(dir string, args ...string) (string, error)

func execSideLinkGit(dir string, args ...string) (string, error) {
	out, err := execGitRunner{}.GitInDir(dir, args...)
	return string(out), err
}

// gitLinkState judges a root-relative path in the repository at root: in the
// published tree (inTree), absent from disk, ignored, or merely uncommitted.
// Paths reach git as literal pathspecs after "--".
func gitLinkState(git sideLinkGit, root string, inTree func(string) (bool, error)) func(string) (linkguard.LinkState, error) {
	return func(p string) (linkguard.LinkState, error) {
		ok, err := inTree(p)
		if err != nil {
			return 0, err
		}
		if ok {
			return linkguard.Committed, nil
		}
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(p))); errors.Is(err, os.ErrNotExist) {
			return linkguard.Absent, nil
		} else if err != nil {
			return 0, err
		}
		out, err := git(root, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "--", ":(literal)"+p)
		if err != nil {
			return 0, err
		}
		if strings.TrimSpace(out) != "" {
			return linkguard.Ignored, nil
		}
		return linkguard.Uncommitted, nil
	}
}

// headHas reports whether HEAD's tree holds a path, file or directory.
func headHas(git sideLinkGit, root string) func(string) (bool, error) {
	return func(p string) (bool, error) {
		out, err := git(root, "ls-tree", "--name-only", "HEAD", "--", ":(literal)"+p)
		return strings.TrimSpace(out) != "", err
	}
}

func readUnder(root string) func(string) (string, error) {
	return func(p string) (string, error) {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		return string(b), err
	}
}

// censusRepo runs the census over docs (root-relative; a missing one is
// skipped) against HEAD in the repository at root.
func censusRepo(git sideLinkGit, root string, docs []string) ([]linkguard.Refusal, error) {
	var in []linkguard.Doc
	for _, d := range docs {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(d)))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		in = append(in, linkguard.Doc{Path: d, Text: string(b)})
	}
	return linkguard.Census(in, gitLinkState(git, root, headHas(git, root)), readUnder(root))
}

// issueSideLinkRefusals judges the issue at issueRel, its plan documents (the
// archive's list) and the active projects across the fleet that reference it,
// each in its own repository. A project that cannot be found or judged is a
// warning, not a refusal, as close's own project update treats it.
func issueSideLinkRefusals(git sideLinkGit, root, issueRel, plansDir string, identity workspace.Identity) (refusals []linkguard.Refusal, warnings []string, err error) {
	var docs []string
	for _, d := range issueArchiveDocs(issueRel, issueRel, filepath.Join(root, plansDir), plansDir, plansDir) {
		docs = append(docs, filepath.ToSlash(d.Src))
	}
	if refusals, err = censusRepo(git, root, docs); err != nil {
		return nil, nil, err
	}
	roots, err := projectWorkspaceRoots(identity)
	if err != nil {
		return refusals, []string{fmt.Sprintf("side-link guard: projects not judged: %v", err)}, nil
	}
	id := issueIDPrefix(filepath.Base(issueRel))
	matches, err := project.DiscoverInRoots(roots, identity.Repo, strings.TrimLeft(id, "0"), project.ActiveOnly)
	if err != nil {
		return refusals, []string{fmt.Sprintf("side-link guard: projects not judged: %v", err)}, nil
	}
	for _, m := range matches {
		rel, err := filepath.Rel(m.RepoDir, m.Path)
		if err == nil {
			var rs []linkguard.Refusal
			if rs, err = censusRepo(git, m.RepoDir, []string{filepath.ToSlash(rel)}); err == nil {
				for _, r := range rs {
					r.Doc = path.Join(m.Repo, r.Doc)
					refusals = append(refusals, r)
				}
				continue
			}
		}
		warnings = append(warnings, fmt.Sprintf("side-link guard: project %s not judged: %v", m.Path, err))
	}
	return refusals, warnings, nil
}

// sideLinkError is the refusal every verb prints, with its bypass when it has one.
func sideLinkError(rs []linkguard.Refusal, bypass string) error {
	msg := "links to files that will not reach the repository:\n" + linkguard.FormatRefusals(rs)
	if bypass != "" {
		msg += "  (or pass " + bypass + ", with the reason in --verified)"
	}
	return errors.New(strings.TrimRight(msg, "\n"))
}

// publishSideDocs judges the details publish is about to put on main against
// main's tree, and returns the side docs it carries with them (D1): files the
// details reach by links, regular and inside the issue's own side folder,
// that main lacks or holds older (unchanged since base) copies of. A side doc
// whose main copy changed since base refuses, as the details do.
//
// With carry false (a first publication, which moves only the details), no
// side doc is carried, so a link main lacks refuses.
func publishSideDocs(git sideLinkGit, root, detailsRel string, details []byte, main, base *gitx.TrunkView, carry bool) (map[string][]byte, error) {
	own := sidedoc.SideDir(detailsRel) + "/"
	id := issueIDPrefix(path.Base(detailsRel))
	candidates := map[string]bool{}
	inner := gitLinkState(git, root, main.Exists)
	state := func(p string) (linkguard.LinkState, error) {
		if id != "" && assetIssueID(p) == id {
			return linkguard.Committed, nil // the issue's asset bundle: publication carries and checks it
		}
		st, err := inner(p)
		if err != nil || !carry || !strings.HasPrefix(p, own) || (st != linkguard.Committed && st != linkguard.Uncommitted) {
			return st, err
		}
		info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil || !info.Mode().IsRegular() {
			return st, nil // gone locally, a directory or a symlink: judged as it is
		}
		candidates[p] = true
		return linkguard.Committed, nil
	}
	rs, err := linkguard.Census([]linkguard.Doc{{Path: detailsRel, Text: string(details)}}, state, readUnder(root))
	if err != nil {
		return nil, err
	}
	if len(rs) > 0 {
		return nil, sideLinkError(rs, "")
	}
	carried := map[string][]byte{}
	for p := range candidates {
		local, err := readRegular(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			return nil, fmt.Errorf("side doc %s: %w", p, err)
		}
		onMain, err := main.Exists(p)
		if err != nil {
			return nil, err
		}
		if !onMain {
			carried[p] = local
			continue
		}
		mainCopy, err := main.Read(p)
		if err != nil {
			return nil, err
		}
		if string(mainCopy) == string(local) {
			continue
		}
		baseCopy, baseErr := base.Read(p)
		if baseErr != nil || string(baseCopy) != string(mainCopy) {
			return nil, fmt.Errorf("side doc %s changed on main since this checkout's base: bring main in, then publish again — nothing was published", p)
		}
		carried[p] = local
	}
	return carried, nil
}

// issueRelTo is p (absolute, or relative to the working directory) as a slash
// path relative to root.
func issueRelTo(root, p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		abs = real
	}
	if real, err := filepath.EvalSymlinks(root); err == nil {
		root = real
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(rel)
}

// changeCodeSideLinks is change-code's side-link check over the issue at
// issuePath, its plans and its projects; warnings go to stderr.
func changeCodeSideLinks(stderr io.Writer, issuePath, plansDir string) error {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	identity, err := resolveWorkspace(root)
	if err != nil {
		return err
	}
	rs, warnings, err := issueSideLinkRefusals(execSideLinkGit, root, issueRelTo(root, issuePath), resolvePlansDir(plansDir), identity)
	for _, w := range warnings {
		cwarn(stderr, w)
	}
	if err != nil || len(rs) == 0 {
		return err
	}
	return sideLinkError(rs, "--no-side-links (or --force <reason>)")
}

// sideChanged names a carried side doc whose local bytes no longer match what
// was published.
func sideChanged(root string, side map[string][]byte) (string, bool) {
	for p, b := range side {
		cur, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil || string(cur) != string(b) {
			return p, true
		}
	}
	return "", false
}

// readRegular reads a file that is, before and after the read, the same
// regular file (never a symlink swapped in between).
func readRegular(p string) ([]byte, error) {
	before, err := os.Lstat(p)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	after, err := os.Lstat(p)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errors.New("not the regular file it was when judged; publish again")
	}
	return b, nil
}
