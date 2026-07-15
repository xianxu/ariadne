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
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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
