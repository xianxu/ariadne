// migrate.go — `sdlc migrate` (#179): move a markdown artifact to a peer
// repo, rewriting repo-relative refs so they resolve identically from the
// destination. Deterministic, no LLM — the ref grammar authority is
// parseRef (resolve.go); the scanner here only FINDS candidates, parseRef
// decides refhood.
//
// Two layers (ARCH-PURE, same shape as resolve.go):
//   - Pure core: rewriteRefs — fence/span-aware segmentation + the three
//     rewrite rules (bare → source-qualified, dest-qualified → bare,
//     everything else passes through). string→string+report, no IO.
//   - Thin IO shell: runMigrate — guards, dest-vantage verification via
//     resolveArtifacts, the file move, scoped two-repo commits, and the
//     inbound-ref sweep.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// refRewrite is one applied rewrite — the unit of the printed summary and
// of the dest-vantage verification.
type refRewrite struct {
	Line     int
	Old, New string
}

// refScanRE finds ref CANDIDATES in prose: an optional directly-attached
// repo token + '#' + 1–6 digits. Derived from parseRef's grammar but not a
// second authority — every candidate is filtered through parseRef before
// any rewrite (a non-parsing candidate like `#0` is a non-ref). Note the
// trailing \b: a 7+-digit run matches NOTHING (RE2 still demands the
// boundary after backtracking), which TestRewriteRefs pins.
var refScanRE = regexp.MustCompile(`([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})\b`)

// inlineSpanRE finds single-backtick code spans within one line.
var inlineSpanRE = regexp.MustCompile("`[^`\n]+`")

// spanRefRE decides whether an inline span's ENTIRE content is one ref
// (optionally with a milestone tag) — the grammar-anchored discriminator
// for mixed-content spans: `` `#171` `` is a styled ref and rewrites;
// `` `git log --grep "^#15"` `` is a quoted command and must not (#179
// plan: real corruption cases in this repo's markdown).
var spanRefRE = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9_.-]*)?#[0-9]{1,6}( M[0-9]+[a-z]?)?$`)

// rewriteToken applies the three rewrite rules to one parseRef-validated
// candidate. Textual (prefix add/strip), so digit padding and any trailing
// milestone text are preserved byte-for-byte. Returns the new token, whether
// it changed, and a skip reason ("" when the token was handled).
func rewriteToken(tok string, sourceRepo, destRepo string) (string, bool, string) {
	ref, err := parseRef(tok)
	if err != nil {
		return tok, false, fmt.Sprintf("not a valid ref (%v)", err)
	}
	switch {
	case ref.GitHub:
		return tok, false, "github ref — repo-relative, verify manually after the move"
	case ref.Repo == "":
		return sourceRepo + tok, true, ""
	case ref.Repo == destRepo:
		// Exact-basename match only. resolveRepoDir prefix-matches
		// (`parley` → `parley.nvim`), so a prefix-form ref migrating into
		// its full-named repo won't normalize — harmless: it still resolves.
		return strings.TrimPrefix(tok, destRepo), true, ""
	default:
		return tok, false, ""
	}
}

// rewriteRefs rewrites body's repo-relative refs for a move from sourceRepo
// to destRepo. Fenced blocks (incl. an unterminated trailing fence) pass
// through verbatim; an inline code span rewrites only when its whole content
// is a single ref; every scanned candidate is parseRef-filtered. Returns the
// rewritten body, the applied rewrites (with 1-indexed line numbers), and a
// report of skipped ref-like matter (gh refs, invalid ids, mixed-content
// spans) for the operator.
func rewriteRefs(body, sourceRepo, destRepo string) (string, []refRewrite, []string) {
	var out strings.Builder
	var rewrites []refRewrite
	var skipped []string
	offset := 0 // absolute byte offset of the current chunk within body
	lineAt := func(pos int) int { return 1 + strings.Count(body[:pos], "\n") }

	rewritePlain := func(text string, base int) {
		last := 0
		for _, loc := range refScanRE.FindAllStringIndex(text, -1) {
			tok := text[loc[0]:loc[1]]
			out.WriteString(text[last:loc[0]])
			newTok, changed, reason := rewriteToken(tok, sourceRepo, destRepo)
			out.WriteString(newTok)
			line := lineAt(base + loc[0])
			if changed {
				rewrites = append(rewrites, refRewrite{Line: line, Old: tok, New: newTok})
			} else if reason != "" {
				skipped = append(skipped, fmt.Sprintf("line %d: %s — %s", line, tok, reason))
			}
			last = loc[1]
		}
		out.WriteString(text[last:])
	}

	for _, seg := range issue.SplitFences(body) {
		if seg.Fenced {
			out.WriteString(seg.Text)
			offset += len(seg.Text)
			continue
		}
		text := seg.Text
		last := 0
		for _, loc := range inlineSpanRE.FindAllStringIndex(text, -1) {
			rewritePlain(text[last:loc[0]], offset+last)
			span := text[loc[0]:loc[1]]
			content := span[1 : len(span)-1]
			line := lineAt(offset + loc[0])
			switch {
			case spanRefRE.MatchString(content):
				newContent, changed, reason := rewriteToken(content, sourceRepo, destRepo)
				out.WriteString("`" + newContent + "`")
				if changed {
					rewrites = append(rewrites, refRewrite{Line: line, Old: content, New: newContent})
				} else if reason != "" {
					skipped = append(skipped, fmt.Sprintf("line %d: %s — %s", line, span, reason))
				}
			case refScanRE.MatchString(content):
				// Mixed-content span (command, grep pattern, multi-ref):
				// never rewritten, always surfaced.
				skipped = append(skipped, fmt.Sprintf("line %d: %s — code span with ref-like content, not rewritten", line, span))
				out.WriteString(span)
			default:
				out.WriteString(span)
			}
			last = loc[1]
		}
		rewritePlain(text[last:], offset+last)
		offset += len(text)
	}
	return out.String(), rewrites, skipped
}

// ── IO shell ──────────────────────────────────────────────────────────────

type migrateOpts struct {
	file         string // source file path (cwd-relative or absolute)
	destDir      string // destination repo dir
	destPath     string // --dest-path: repo-root-relative path at dest ("" = same as source)
	noCommit     bool
	noCleanCheck bool
	stderr       io.Writer // all migrate output is stderr (status/report); stdout stays clean
}

// gitInDir runs one git command in dir, returning combined output.
func gitInDir(dir string, args ...string) (string, error) {
	c := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := c.CombinedOutput()
	return string(out), err
}

// issueFamilyRE matches the id-keyed artifact filename prefix (NNNNNN-).
var issueFamilyRE = regexp.MustCompile(`^[0-9]{6}-`)

// isIssueFamilyPath reports whether root-relative relPath lies in one of the
// vocab Discovery dirs with an id-keyed name — the artifact class whose
// migration needs dest-side renumbering (v2), so v1 refuses it.
func isIssueFamilyPath(relPath string, d vocab.Discovery) bool {
	if !issueFamilyRE.MatchString(filepath.Base(relPath)) {
		return false
	}
	dir := filepath.ToSlash(filepath.Dir(relPath))
	for _, sub := range []string{d.Home, d.Plans, d.Archive} {
		if dir == filepath.ToSlash(sub) {
			return true
		}
	}
	return false
}

// verifyRefTarget strips a milestone tag off a rewritten ref: verification
// asserts the ISSUE exists, not that a review sidecar for the milestone does.
func verifyRefTarget(newForm string) string {
	if i := strings.IndexByte(newForm, ' '); i > 0 {
		return newForm[:i]
	}
	return newForm
}

// runMigrate moves o.file into o.destDir with refs rewritten (#179). Guard
// order and semantics per the plan: refusals via die (fail-closed, nothing
// half-moved); every rewrite verified from the DESTINATION's vantage before
// any write.
func runMigrate(o *migrateOpts) error {
	srcRoot, err := gitx.RepoTopLevel()
	if err != nil {
		die(o.stderr, "not inside a git repo: "+err.Error())
	}
	srcRepo := filepath.Base(srcRoot)

	// (0) source path normalization: must lie inside the cwd repo — the
	// transaction lock and the source-side commit are anchored there.
	// EvalSymlinks BOTH sides before comparing: os.Getwd (behind
	// filepath.Abs) prefers the logical $PWD env path, while git returns
	// the resolved one — under a symlinked cwd (macOS /tmp) the two
	// disagree on a prefix and Rel would misfire (live-dogfood regression).
	absFile, err := filepath.Abs(o.file)
	if err != nil {
		die(o.stderr, fmt.Sprintf("resolve %s: %v", o.file, err))
	}
	if _, err := os.Stat(absFile); err != nil {
		die(o.stderr, fmt.Sprintf("source file %s: %v", o.file, err))
	}
	if resolved, rerr := filepath.EvalSymlinks(absFile); rerr == nil {
		absFile = resolved
	}
	if resolved, rerr := filepath.EvalSymlinks(srcRoot); rerr == nil {
		srcRoot = resolved
	}
	relPath, err := filepath.Rel(srcRoot, absFile)
	if err != nil || strings.HasPrefix(relPath, "..") {
		die(o.stderr, fmt.Sprintf("%s is not inside the current repo (%s) — run migrate from the repo that owns the file (the lock and the source commit anchor there)", o.file, srcRepo))
	}
	relPath = filepath.ToSlash(relPath)

	// (1) tracked + unmodified — we migrate reviewed, committed state only.
	if out, err := gitInDir(srcRoot, "status", "--porcelain", "--untracked-files=all", "--", relPath); err != nil {
		die(o.stderr, fmt.Sprintf("git status %s: %v — %s", relPath, err, out))
	} else if strings.TrimSpace(out) != "" {
		die(o.stderr, fmt.Sprintf("%s has uncommitted changes (or is untracked) — commit it first; migrate moves reviewed state only", relPath))
	}

	// (2) issue-family artifacts need dest-side renumbering — v2 (#179).
	if isIssueFamilyPath(relPath, vocab.Issue().Discovery()) {
		die(o.stderr, fmt.Sprintf("%s is an id-keyed issue-family artifact — issue IDs are per-repo sequences, so migrating it must renumber it at the destination (v2, #179). v1 moves slug-named artifacts (project files, docs).", relPath))
	}

	// (3) destination shape: a git repo, not this repo, not a brain.
	destAbs, err := filepath.Abs(o.destDir)
	if err != nil {
		die(o.stderr, fmt.Sprintf("resolve %s: %v", o.destDir, err))
	}
	destTopOut, err := gitInDir(destAbs, "rev-parse", "--show-toplevel")
	if err != nil {
		die(o.stderr, fmt.Sprintf("%s is not a git repo: %s", o.destDir, strings.TrimSpace(destTopOut)))
	}
	destTop := strings.TrimSpace(destTopOut)
	destRepo := filepath.Base(destTop)
	if destTop == srcRoot {
		die(o.stderr, fmt.Sprintf("destination resolves to the same repo (%s) — a same-repo migrate would flip bare↔qualified forms for nothing; use git mv", srcRepo))
	}
	if _, err := os.Stat(filepath.Join(destTop, ".brain", "config.md")); err == nil {
		die(o.stderr, fmt.Sprintf("%s is a brain (capture repo) — SDLC process artifacts don't live in brain (#171); pick the work's center-of-gravity repo instead", destRepo))
	}

	// (4) destination cleanliness + free path.
	destRel := o.destPath
	if destRel == "" {
		destRel = relPath
	}
	destRel = filepath.ToSlash(destRel)
	destFile := filepath.Join(destTop, filepath.FromSlash(destRel))
	// Containment guard (close review I3): a traversal --dest-path
	// (../evil.md) would otherwise write a stray file OUTSIDE the dest repo
	// before git add fails — breaking fail-closed. Mirror the source-side
	// inside-repo check.
	if rel, rerr := filepath.Rel(destTop, destFile); rerr != nil || strings.HasPrefix(rel, "..") {
		die(o.stderr, fmt.Sprintf("--dest-path %s escapes the destination repo (%s) — pass a repo-root-relative path", o.destPath, destRepo))
	}
	if _, err := os.Stat(destFile); err == nil {
		die(o.stderr, fmt.Sprintf("destination path %s already exists in %s — pass --dest-path to place it elsewhere", destRel, destRepo))
	}
	if st, err := gitStatusPorcelain(destTop); err != nil {
		die(o.stderr, fmt.Sprintf("git status in %s: %v", destRepo, err))
	} else if st != "" {
		if !o.noCleanCheck {
			die(o.stderr, fmt.Sprintf("destination repo %s is dirty — commit/stash there first (or pass --no-clean-check to proceed; staging is explicit-path either way)", destRepo))
		}
		cwarn(o.stderr, "--no-clean-check: proceeding into a dirty destination (staging is explicit-path)")
	}

	// (5) rewrite + verify from the destination's vantage, BEFORE any write.
	raw, err := os.ReadFile(absFile)
	if err != nil {
		die(o.stderr, fmt.Sprintf("read %s: %v", relPath, err))
	}
	rewritten, rewrites, skipped := rewriteRefs(string(raw), srcRepo, destRepo)
	for _, s := range skipped {
		cwarn(o.stderr, "not rewritten — "+s)
	}
	for _, r := range rewrites {
		target := verifyRefTarget(r.New)
		if _, _, verr := resolveArtifacts(target, destTop); verr != nil {
			die(o.stderr, fmt.Sprintf("ref %s (line %d, rewritten to %s) does not resolve from %s: %v\n  Nothing was moved. Fix the ref (or fence it), then re-run.", r.Old, r.Line, r.New, destRepo, verr))
		}
		cinfo(o.stderr, fmt.Sprintf("line %d: %s → %s", r.Line, r.Old, r.New))
	}
	if len(rewrites) == 0 {
		// NOTE: with zero rewrites there is nothing to verify, so the
		// non-sibling-dest backstop (a side effect of dest-vantage
		// verification) is vacuous for ref-free files. Harmless in v1 —
		// the file carries no refs that could dangle.
		cinfo(o.stderr, "no refs needed rewriting")
	}

	// (6) move: write dest, remove source, stage both sides explicitly.
	if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
		die(o.stderr, fmt.Sprintf("mkdir for %s: %v", destRel, err))
	}
	if err := os.WriteFile(destFile, []byte(rewritten), 0o644); err != nil {
		die(o.stderr, fmt.Sprintf("write %s: %v", destFile, err))
	}
	if out, err := gitInDir(destTop, "add", "--", destRel); err != nil {
		die(o.stderr, fmt.Sprintf("git add in %s: %v — %s", destRepo, err, out))
	}
	if err := os.Remove(absFile); err != nil {
		die(o.stderr, fmt.Sprintf("remove %s: %v", relPath, err))
	}
	if out, err := gitInDir(srcRoot, "add", "--", relPath); err != nil {
		die(o.stderr, fmt.Sprintf("git add in %s: %v — %s", srcRepo, err, out))
	}

	// (7) scoped commits (or hand the staged state to the operator).
	if o.noCommit {
		cinfo(o.stderr, "--no-commit: both sides staged; commit with:")
		fmt.Fprintf(o.stderr, "    git -C %s commit -m %q\n", destTop, "migrate: receive "+destRel+" from "+srcRepo)
		fmt.Fprintf(o.stderr, "    git -C %s commit -m %q\n", srcRoot, "migrate: move "+relPath+" to "+destRepo)
	} else {
		if out, err := gitInDir(destTop, "commit", "-q", "-m", "migrate: receive "+destRel+" from "+srcRepo); err != nil {
			die(o.stderr, fmt.Sprintf("commit in %s: %v — %s", destRepo, err, out))
		}
		if out, err := gitInDir(srcRoot, "commit", "-q", "-m", "migrate: move "+relPath+" to "+destRepo); err != nil {
			die(o.stderr, fmt.Sprintf("commit in %s: %v — %s", srcRepo, err, out))
		}
		cok(o.stderr, fmt.Sprintf("moved %s → %s/%s (both sides committed, scoped)", relPath, destRepo, destRel))
	}

	// (8) inbound-ref sweep (report-only, v1): every parent-dir sibling repo,
	// including the two participants, minus the migrated file itself.
	reportInboundRefs(o.stderr, srcRoot, destTop, relPath, destRel)
	return nil
}

// reportInboundRefs prints tracked-file references to the migrated artifact
// (by old repo-relative path or basename) across the source's parent-dir
// sibling repos. Report-only (#179 v1): issue refs are location-independent,
// path references are not — the operator judges.
func reportInboundRefs(stderr io.Writer, srcRoot, destTop, relPath, destRel string) {
	parent := filepath.Dir(srcRoot)
	entries, err := os.ReadDir(parent)
	if err != nil {
		cwarn(stderr, "inbound-ref sweep skipped: "+err.Error())
		return
	}
	base := filepath.Base(relPath)
	var hits []string
	seen := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		repoDir := filepath.Join(parent, e.Name())
		if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
			continue
		}
		for _, pat := range []string{relPath, base} {
			out, gerr := gitInDir(repoDir, "grep", "-n", "-F", pat)
			if gerr != nil {
				continue // exit 1 = no match; other errors are non-fatal for a report
			}
			for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
				if line == "" {
					continue
				}
				file := line
				if i := strings.IndexByte(line, ':'); i > 0 {
					file = line[:i]
				}
				// The migrated file's own new home isn't an inbound ref.
				if repoDir == destTop && filepath.ToSlash(file) == destRel {
					continue
				}
				key := e.Name() + "/" + line
				if !seen[key] {
					seen[key] = true
					hits = append(hits, key)
				}
			}
		}
	}
	if len(hits) == 0 {
		cinfo(stderr, "inbound-ref sweep: no references to the old path across sibling repos")
		return
	}
	cwarn(stderr, fmt.Sprintf("inbound references to the moved artifact (%d) — path refs need hand-fixing (issue refs survive moves):", len(hits)))
	for _, h := range hits {
		fmt.Fprintf(stderr, "    %s\n", h)
	}
}

// NewMigrateCmd returns the cobra command for `sdlc migrate` (#179).
func NewMigrateCmd() *cobra.Command {
	var o migrateOpts
	cmd := markMutatingCommand(&cobra.Command{
		Use:   "migrate <file> <dest-repo-dir>",
		Short: "Move a markdown artifact to a peer repo, rewriting refs (#179)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Deliberately NOT guardSpineRepo'd: migrating an artifact OUT of
			// a brain is the #171 use case; the brain guard applies to the
			// DESTINATION instead (inside runMigrate).
			o.file, o.destDir = args[0], args[1]
			o.stderr = cmd.ErrOrStderr()
			return runMigrate(&o)
		},
	})
	cmd.Flags().StringVar(&o.destPath, "dest-path", "", "repo-root-relative path at the destination (default: same as the source's)")
	cmd.Flags().BoolVar(&o.noCommit, "no-commit", false, "stage both sides but leave the two commits to the operator")
	cmd.Flags().BoolVar(&o.noCleanCheck, "no-clean-check", false, "proceed even if the destination repo has uncommitted changes")
	cmd.SilenceErrors = true
	return cmd
}
