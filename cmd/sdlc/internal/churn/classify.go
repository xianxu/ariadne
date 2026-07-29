// Package churn answers two questions about a work window that a line count alone
// cannot: WHERE the lines landed, and how many TIMES they were written.
//
// It exists because ariadne#187 needed the plan gate's cost to be measurable. The
// motivating case is pair#127: a 126-line change that took six `change-code`
// invocations and five rejections. Its FINAL diff scored as merely process-heavy —
// modest code, lots of prose. What that missed is the waste: one file rewritten five
// times. Insertions-across-commits over insertions-in-the-final-diff surfaces it as a
// rework multiple, so "the gate made me rewrite this five times" becomes a number.
//
// Everything here is PURE (ARCH-PURE): the git invocations live at the caller's seam
// (cmd/sdlc/churnreport.go) and hand this package raw numstat text. That is what lets
// the whole classification and ratio be tested on in-memory strings with no repo.
//
// Deliberately NOT measured: a comment-vs-code split (#187 D1). This house style is
// comment-dense by design, so that ratio is descriptive at best and a Goodhart target
// at worst — it would reward stripping the reasoning this repo exists to preserve.
package churn

import "strings"

// Bucket is where a changed path's lines count toward. The four are chosen so the
// close-time line reads as an answer to "what KIND of work was this window": product
// code, its tests, the map, and the paperwork.
type Bucket string

const (
	// CodeProd is the DEFAULT — see ClassifyPath for why that is a decision.
	CodeProd Bucket = "code-prod"
	CodeTest Bucket = "code-test"
	Atlas    Bucket = "atlas"
	Workshop Bucket = "workshop"
)

// ClassifyPath buckets one repo-relative path. The rule, in order:
//
//  1. a leading `atlas/` segment    → Atlas
//  2. a leading `workshop/` segment → Workshop
//  3. `*_test.go`, or any `testdata/` segment → CodeTest
//  4. everything else → CodeProd
//
// Order is observable and pinned by test: a test file under `workshop/` is workshop
// churn. The prefixes match a path SEGMENT rather than a substring, so an `atlasctl/`
// or `docs/atlas/` directory is not silently absorbed into the map's numbers.
//
// CodeProd being the DEFAULT is a decision, not a fallthrough nobody chose. Two classes
// land there on purpose:
//
//   - Embedded markdown — `internal/judge/prompts/*.md`, `helptext/*.md`, `*.cue`,
//     `AGENTS.base.md`. It ships inside the binary via //go:embed and is exactly the
//     surface #187 changes; counting it as prose would understate the code this repo
//     actually writes.
//   - Build and config files — `go.mod`, Makefiles, CI workflows, the base manifest.
//     They are versioned, reviewed, and break the build when wrong.
//
// The cost of that choice is that a lockfile-sized diff lands in code-prod. That must
// be a visible choice rather than an accident, which is why it is stated here and
// table-tested rather than left implicit.
func ClassifyPath(path string) Bucket {
	switch {
	case hasSegmentPrefix(path, "atlas"):
		return Atlas
	case hasSegmentPrefix(path, "workshop"):
		return Workshop
	case isTestPath(path):
		return CodeTest
	default:
		return CodeProd
	}
}

// hasSegmentPrefix reports whether path's FIRST path segment is exactly seg. Substring
// matching would fold `atlasctl/main.go` into the atlas bucket.
func hasSegmentPrefix(path, seg string) bool {
	return path == seg || strings.HasPrefix(path, seg+"/")
}

// isTestPath reports whether a path is test material: a Go test file, or anything under
// a `testdata/` directory at any depth (fixtures and fuzz corpora included).
func isTestPath(path string) bool {
	if strings.HasSuffix(path, "_test.go") {
		return true
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == "testdata" {
			return true
		}
	}
	return false
}
