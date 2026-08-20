---
issue: 000194
created: 2026-08-20
---

# Boundary-Review Anchor Commit — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Anchor a boundary review to the concrete commit it read, so a commit landing
mid-review is classified as a *delta* (doc-only → finalize; code → refuse, by commit)
instead of blanket-invalidating a 20-minute run.

**Architecture:** Two changes, in order. (1) **Resolve the head once.** Today
`resolveReviewWindow` returns the literal string `"HEAD"`, and the diff is collected
*after* the repo lock is released — so the reviewed commit is never recorded and can
even drift between snapshot and dispatch. Resolve HEAD to a SHA under the lock and
thread that one value through the diff, the prompt, the trailer, the sidecar, and the
finalize check. (2) **Classify instead of compare.** `closeReviewSnapshot.validate()`
currently asks "is HEAD identical?"; replace that with a pure classifier over
`reviewedSHA..currentHEAD` that reuses `publishGateHasCodeSurface` — the exact predicate
`sdlc merge` already uses (ARCH-DRY, and the issue's Done-when names it).

**Tech Stack:** Go 1.x, `cmd/sdlc` (cobra), `cmd/sdlc/internal/gitx`, `go test` with the
existing hermetic-repo harness (`closeRepo`, `executeSDLCTestCommand`, `judge.Run`
override).

**Milestone shape:** single review boundary → **plain checkboxes, no `Mx` tags**
(AGENTS.md §3). One `sdlc close` at the end; the mandatory fresh-eyes review runs there.

---

## The one real design decision

The issue's Spec offers a choice for a **code** delta: *"finalize the reviewed portion
and report the unreviewed delta explicitly … **or** refuse, but say which commits are
unreviewed."*

**We refuse — and name the commits.** Reason: finalizing would violate Done-when #5
("the publish-time invariant is unchanged: no code ships unreviewed"). `runPublishGate`
anchors on `codecompleteAnchorCommit` — the *close commit*. If close finalized on top of
an unreviewed code delta, the close commit would sit **above** that delta, the publish
gate would compute `closeCommit..HEAD` = 0 commits, report
`reviewed-HEAD-unchanged ✓`, and ship code no reviewer ever read. Making that safe would
require re-anchoring the publish gate on a recorded reviewed-SHA — a second, larger
change to the publish contract that this issue explicitly does *not* ask for.

The doc-only case has no such hazard: the delta carries no code surface, so the
invariant "no **code** ships unreviewed" holds by construction — which is precisely the
#174 judgement `publishGateHasCodeSurface` already encodes one stage later.

Net effect on the reported pain: the ~20-minute freeze stops applying to the bookkeeping
that a close *itself* invites (lessons.md, atlas, plan ticks, issue edits on other
issues, parley/pensive notes) — the commits that made 4 of 8 runs dead time. A genuine
code commit still invalidates, and should: that is the gate doing its job.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `deltaCommit` | `cmd/sdlc/reviewanchor.go` | new |
| `reviewAnchorDelta` | `cmd/sdlc/reviewanchor.go` | new |
| `anchorOutcome` | `cmd/sdlc/reviewanchor.go` | new |
| `classifyReviewAnchor` | `cmd/sdlc/reviewanchor.go` | new |
| `formatAnchorDocsOnly` | `cmd/sdlc/reviewanchor.go` | new |
| `formatAnchorRefusal` | `cmd/sdlc/reviewanchor.go` | new |
| `closeReviewSnapshot` | `cmd/sdlc/close.go:1182` | modified |
| `resolveReviewWindow` | `cmd/sdlc/milestoneclose.go:243` | modified |
| `publishGateHasCodeSurface` | `cmd/sdlc/publishgate.go:175` | unchanged (reused) |

- **`reviewAnchorDelta`** — the facts about `reviewedSHA..currentHEAD`, gathered by IO,
  decided on by pure code. Fields: `Reviewed`, `Current` (long SHAs), `Descendant`
  (is `Current` a descendant of `Reviewed`), `Commits []deltaCommit`, `Paths []string`.
  - **Relationships:** 1:1 with a dispatched boundary review; held by
    `closeReviewSnapshot` for the lifetime of one close.
  - **DRY rationale:** Gives the close gate the *same* question the publish gate asks
    ("does this delta carry code surface?") without a second implementation of the
    answer — `classifyReviewAnchor` calls `publishGateHasCodeSurface`, it does not
    restate it. This is the ARCH-DRY reuse the issue's Spec §"Why the current guard is
    stricter than its own purpose" asks for.
  - **Future extensions:** Add `Sidecar string` if a re-attachable review (the issue's
    "second-order effect") ever needs to point back at the transcript for the delta.

- **`anchorOutcome`** — four-valued: `anchorUnchanged`, `anchorDocsOnly`,
  `anchorCodeDelta`, `anchorDiverged`. The fourth exists because `Reviewed` may not be
  an ancestor of `Current` at all — a rebase, `reset --hard`, or a branch switch during
  the review. `git diff A B` between unrelated commits happily returns paths, so
  without the ancestry check a rebase-away could masquerade as a doc-only delta. It
  refuses with the diverged message.
  - **DRY rationale:** First occurrence — the publish gate uses `revCount(a+"..HEAD")`,
    which conflates "diverged" with "0 ahead". Fixing that there is out of scope.

- **`classifyReviewAnchor(d reviewAnchorDelta) anchorOutcome`** — pure, total, no IO.
  Unit-tested in `cmd/sdlc/reviewanchor_test.go` **without git**: it takes the gathered
  facts as a struct (ARCH-PURE — the `exec` lives in the caller).

- **`formatAnchorDocsOnly` / `formatAnchorRefusal`** — pure string builders. Per
  `formatPublishGateDocsOnly`'s comment (#172), the **pass** line must share no
  vocabulary with the **refusal** line, because `gatesig` classifies transcripts by
  substring and a pass line echoing refusal words corrupts friction attribution. A test
  pins that non-collision.

- **`closeReviewSnapshot`** (modified) — `head string` becomes `reviewed string`
  (the concrete SHA, supplied by the caller rather than re-`rev-parse`d), and
  `validate()` delegates its HEAD question to the classifier. The issue-file and
  project-file checks are **unchanged**: the review read that text, so a mid-review edit
  to it is still a genuine invalidation (the issue's Spec says exactly this).

- **`resolveReviewWindow`** (modified) — returns a resolved `head` SHA instead of the
  literal `"HEAD"`, falling back to `"HEAD"` when `rev-parse` fails (keeps the
  documented `("?", "", "HEAD")` no-anchor return shape intact).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `gatherReviewAnchorDelta` | `cmd/sdlc/reviewanchor.go` | new | `git` (via `gitx`) |

- **`gatherReviewAnchorDelta(reviewed string) (reviewAnchorDelta, error)`** — the thin IO
  shell: `rev-parse HEAD`, `merge-base --is-ancestor`, `log --format=%H %s`,
  `gitx.DiffNames`. Returns an error only when git itself fails, so the caller keeps the
  **fail-closed** posture the publish gate documents ("if we can't verify, refuse").
  - **Injected into:** `closeReviewSnapshot.validate()`. Pure decision-making stays in
    `classifyReviewAnchor`, so the interesting logic is unit-testable with no repo.
  - **Test surface (ARCH-MOCK):** `git` is the external binary here, and this repo's
    established seam for it is the **hermetic real-git repo** (`hermeticrepo_test.go`,
    `closeRepo`), not a mocked runner — every existing close/merge gate test drives real
    git in a temp repo. We reuse that seam rather than introduce a second git double;
    the fake-vs-real conformance question is already settled by using the real binary.
  - **Future extensions:** Return `MergeBase` if a diverged-history delta ever needs to
    be described rather than merely refused.

---

## Chunk 1: anchor the review, classify the delta

### Task 1: Resolve the reviewed SHA once, and thread it

Today `Head` is the string `"HEAD"` from `resolveReviewWindow` down through
`collectDiff`, `judge.PromptInput`, `emitTrailerBlock` and the sidecar — so nothing
records what was actually reviewed. Worse: `reviewThenFinalizeLocked` **releases the
lock before** `dispatchBoundaryReview` runs, and `boundaryReviewDispatchOptions` then
resolves `"HEAD"` itself — so today the snapshot's `rev-parse` and the diff can already
disagree. Resolving once, under the lock, closes that gap and is the precondition for
everything else.

**Files:**
- Modify: `cmd/sdlc/milestoneclose.go:243-256` (`resolveReviewWindow`)
- Modify: `cmd/sdlc/milestoneclose.go:576-600` (`boundaryReviewDispatchOptions`)
- Modify: `cmd/sdlc/close.go:1189-1196` (`captureCloseReviewSnapshot` signature)
- Modify: `cmd/sdlc/close.go:1036-1043`, `cmd/sdlc/milestoneclose.go:200-212` (call sites)
- Test: `cmd/sdlc/milestonewindow_test.go`, `cmd/sdlc/closereview_test.go`

- [ ] **Step 1: Write the failing test** — `resolveReviewWindow` returns a concrete SHA.

```go
// cmd/sdlc/milestonewindow_test.go
func TestResolveReviewWindow_HeadIsConcreteSHA(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	_, _, head := resolveReviewWindow("69", "", filepath.Join(issuesDir, "000069-boundary-review.md"))
	if head == "HEAD" {
		t.Fatal("head must be resolved to a SHA, not the literal \"HEAD\"")
	}
	want := captureGit(t, "rev-parse", "HEAD")
	if head != strings.TrimSpace(want) {
		t.Fatalf("head = %q, want %q", head, strings.TrimSpace(want))
	}
}
```

- [ ] **Step 2: Run it, confirm it fails**

Run: `go test ./cmd/sdlc/ -run TestResolveReviewWindow_HeadIsConcreteSHA -v`
Expected: FAIL — `head must be resolved to a SHA`.

- [ ] **Step 3: Resolve the head**

```go
// cmd/sdlc/milestoneclose.go — in resolveReviewWindow, replace `head = "HEAD"`
// head is the CONCRETE SHA the review will read (#194), not the floating "HEAD":
// it is the anchor the finalize check classifies against, so it must be pinned
// under the caller's lock. Falls back to "HEAD" only when rev-parse fails (an
// empty repo / dry-run in a non-repo), preserving the documented no-anchor shape.
head = "HEAD"
if sha := strings.TrimSpace(gitx.Capture("rev-parse", "HEAD")); sha != "" {
	head = sha
}
```

- [ ] **Step 4: Use the pinned head for the diff and the prompt**

In `boundaryReviewDispatchOptions`, replace both literal `"HEAD"` uses with `p.Head`:

```go
diff, _, err := collectDiff(judge.MilestoneReview, p.BaseLong, p.Head, p.IssuesDir, "workshop/history")
...
in := judge.PromptInput{
	Diff: diff, Base: p.BaseLong, Head: p.Head,
	...
}
```

Also update `dispatchBoundaryReview`'s log line — `cinfo(..., "dispatching boundary
review (%s..%s) via %s …", p.BaseLong, shortSHA(p.Head), agent)`.

- [ ] **Step 5: Take the reviewed SHA as a parameter instead of re-reading it**

```go
// cmd/sdlc/close.go
// captureCloseReviewSnapshot pins the state the boundary review is about to read.
// reviewedSHA is the window head the caller already resolved (#194) — passing it in
// rather than re-`rev-parse`ing guarantees the snapshot, the reviewed diff, and the
// finalize check all name the SAME commit.
func captureCloseReviewSnapshot(r closeResult, reviewedSHA string) closeReviewSnapshot {
	return closeReviewSnapshot{
		reviewed:  reviewedSHA,
		issuePath: r.issuePath,
		issueText: r.issueText,
		projects:  r.projectEdits,
	}
}
```

Update both call sites to `captureCloseReviewSnapshot(r, head)` — they already compute
`head` on the line above, inside the lock.

- [ ] **Step 6: Fix the trailer/sidecar expectations**

`emitTrailerBlock` now renders `Review-Window: <base>..<reviewedShort>`. Shorten the head
for the trailer via the existing `shortSHA`. `Review-Window` has **no production parser**
(verified: only `close.go:1672` help text, `helptext/milestone-close.md`, and
`atlas/workflow/ledger-landscape.md` mention it), so this is a pure provenance
improvement. Update:
- `cmd/sdlc/closereview_test.go:226` — assert a `..`-joined pair of SHAs, not `"..HEAD"`.
- `cmd/sdlc/milestoneclose_test.go:120,134` — same.
- `cmd/sdlc/helptext/milestone-close.md:35` and `milestoneclose.go:395` doc comment —
  show `abc1234..def5678`.

(Fixtures that *construct* commit messages containing `Review-Window: abc1234..HEAD` —
`close_test.go:564,611,653`, `closereview_test.go:485`, `milestonewindow_test.go:85,164`
— are inputs to the `Review-Verdict:` grep and need **no** change.)

- [ ] **Step 7: Run the full package**

Run: `go test ./cmd/sdlc/ 2>&1 | tail -20`
Expected: PASS (`TestCloseCommand_HEADChangedDuringBoundaryReview_DoesNotFinalize` still
passes — the guard is still identity-based at this point).

- [ ] **Step 8: Commit**

```bash
git add cmd/sdlc/
git commit -m "#194: pin the boundary review to a concrete reviewed SHA

The window head was the literal string \"HEAD\", resolved independently by
the diff collector AFTER the repo lock was released — so nothing recorded
what was reviewed, and the snapshot could name a different commit than the
diff. Resolve once under the lock and thread that SHA through."
```

---

### Task 2: The pure delta classifier

**Files:**
- Create: `cmd/sdlc/reviewanchor.go`
- Test: `cmd/sdlc/reviewanchor_test.go`

- [ ] **Step 1: Write the failing tests** (pure — no git, no temp repo)

```go
// cmd/sdlc/reviewanchor_test.go
package main

import "strings"
import "testing"

func TestClassifyReviewAnchor(t *testing.T) {
	for _, tc := range []struct {
		name string
		d    reviewAnchorDelta
		want anchorOutcome
	}{
		{"identical head", reviewAnchorDelta{Reviewed: "aaa", Current: "aaa", Descendant: true}, anchorUnchanged},
		{"docs delta", reviewAnchorDelta{
			Reviewed: "aaa", Current: "bbb", Descendant: true,
			Commits: []deltaCommit{{SHA: "bbb", Subject: "lessons"}},
			Paths:   []string{"workshop/lessons.md", "atlas/index.md"},
		}, anchorDocsOnly},
		{"code delta", reviewAnchorDelta{
			Reviewed: "aaa", Current: "bbb", Descendant: true,
			Commits: []deltaCommit{{SHA: "bbb", Subject: "fix"}},
			Paths:   []string{"cmd/sdlc/close.go"},
		}, anchorCodeDelta},
		{"embedded helptext counts as code", reviewAnchorDelta{
			Reviewed: "aaa", Current: "bbb", Descendant: true,
			Commits: []deltaCommit{{SHA: "bbb", Subject: "helptext"}},
			Paths:   []string{"cmd/sdlc/helptext/close.md"},
		}, anchorCodeDelta},
		{"rebased away", reviewAnchorDelta{
			Reviewed: "aaa", Current: "ccc", Descendant: false,
			Paths: []string{"workshop/lessons.md"},
		}, anchorDiverged},
		{"unresolvable reviewed sha degrades to unchanged", reviewAnchorDelta{
			Reviewed: "", Current: "bbb",
		}, anchorUnchanged},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyReviewAnchor(tc.d); got != tc.want {
				t.Fatalf("classifyReviewAnchor = %v, want %v", got, tc.want)
			}
		})
	}
}

// The pass line must not collide with the refusal vocabulary — gatesig
// classifies transcripts by substring (#172, mirrors the publish gate's guard).
func TestAnchorPassLineNoRefusalVocabulary(t *testing.T) {
	d := reviewAnchorDelta{Reviewed: "aaaaaaaaaa", Current: "bbbbbbbbbb", Descendant: true,
		Commits: []deltaCommit{{SHA: "bbbbbbbbbb", Subject: "docs"}}}
	pass := formatAnchorDocsOnly(d)
	for _, forbidden := range []string{"NOT finalized", "unreviewed", "re-run", "stale"} {
		if strings.Contains(pass, forbidden) {
			t.Fatalf("pass line %q must not contain refusal word %q", pass, forbidden)
		}
	}
}

func TestAnchorRefusalNamesEveryCommit(t *testing.T) {
	d := reviewAnchorDelta{
		Reviewed: "aaaaaaaaaaaa", Current: "cccccccccccc", Descendant: true,
		Commits: []deltaCommit{
			{SHA: "cccccccccccc", Subject: "second fix"},
			{SHA: "bbbbbbbbbbbb", Subject: "first fix"},
		},
		Paths: []string{"cmd/sdlc/close.go"},
	}
	msg := formatAnchorRefusal(d, anchorCodeDelta, "sdlc close")
	for _, want := range []string{"cccccccc", "second fix", "bbbbbbbb", "first fix", "cmd/sdlc/close.go", "sdlc close"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("refusal must name %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "HEAD changed from") {
		t.Fatal("refusal must report commits, not a bare HEAD-identity message")
	}
}
```

- [ ] **Step 2: Run them, confirm they fail**

Run: `go test ./cmd/sdlc/ -run 'TestClassifyReviewAnchor|TestAnchorPassLine|TestAnchorRefusal' -v`
Expected: FAIL to compile — `undefined: reviewAnchorDelta`.

- [ ] **Step 3: Write `cmd/sdlc/reviewanchor.go`**

```go
// reviewanchor.go — #194. A boundary review is anchored to the COMMIT IT READ, not
// to "HEAD has not moved since". A commit landing while the ~20-minute review runs is
// a DELTA to be classified, not an invalidation:
//
//	doc-only delta → finalize (the reviewed code surface is unchanged)
//	code delta     → refuse, naming the commits (see the plan's design note: the
//	                 publish gate anchors on the close commit, so finalizing above
//	                 an unreviewed code delta would ship it unreviewed)
//	diverged       → refuse (history was rewritten; the delta is not describable)
//
// The code-surface question is answered by publishGateHasCodeSurface — the SAME
// predicate `sdlc merge` uses (#174), reused one stage earlier rather than restated
// (ARCH-DRY). Decision logic here is PURE; the git reads live in
// gatherReviewAnchorDelta (ARCH-PURE).
package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// deltaCommit is one commit in reviewedSHA..HEAD.
type deltaCommit struct {
	SHA     string
	Subject string
}

// reviewAnchorDelta is the gathered, git-free description of reviewedSHA..HEAD.
type reviewAnchorDelta struct {
	Reviewed   string // long SHA the review read ("" ⇒ nothing to check)
	Current    string // long SHA at HEAD now
	Descendant bool   // Current descends from Reviewed
	Commits    []deltaCommit
	Paths      []string
}

type anchorOutcome int

const (
	anchorUnchanged anchorOutcome = iota // no delta — finalize
	anchorDocsOnly                       // delta carries no code surface — finalize
	anchorCodeDelta                      // delta carries code — refuse, name the commits
	anchorDiverged                       // history rewritten — refuse, cannot classify
)

// classifyReviewAnchor decides what a delta means. Pure and total.
func classifyReviewAnchor(d reviewAnchorDelta) anchorOutcome {
	if d.Reviewed == "" || d.Reviewed == d.Current {
		return anchorUnchanged
	}
	if !d.Descendant {
		return anchorDiverged
	}
	if publishGateHasCodeSurface(d.Paths) {
		return anchorCodeDelta
	}
	return anchorDocsOnly
}

// gatherReviewAnchorDelta is the thin IO shell (ARCH-PURE). Errors ONLY on git
// failure, so the caller can fail closed the way the publish gate does.
func gatherReviewAnchorDelta(reviewed string) (reviewAnchorDelta, error) {
	d := reviewAnchorDelta{Reviewed: reviewed}
	if reviewed == "" {
		return d, nil
	}
	d.Current = strings.TrimSpace(gitx.Capture("rev-parse", "HEAD"))
	if d.Current == "" {
		return d, fmt.Errorf("cannot resolve HEAD")
	}
	if d.Current == reviewed {
		d.Descendant = true
		return d, nil
	}
	if _, err := gitx.RunGit("merge-base", "--is-ancestor", reviewed, d.Current); err != nil {
		return d, nil // not an ancestor ⇒ diverged; classifier refuses
	}
	d.Descendant = true
	out, err := gitx.RunGit("log", "--format=%H %s", reviewed+".."+d.Current)
	if err != nil {
		return d, fmt.Errorf("log %s..%s: %w", shortSHA(reviewed), shortSHA(d.Current), err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		sha, subject, _ := strings.Cut(line, " ")
		d.Commits = append(d.Commits, deltaCommit{SHA: sha, Subject: subject})
	}
	paths, err := gitx.DiffNames(reviewed, d.Current)
	if err != nil {
		return d, fmt.Errorf("diff %s..%s: %w", shortSHA(reviewed), shortSHA(d.Current), err)
	}
	d.Paths = paths
	return d, nil
}

// formatAnchorDocsOnly renders the PASS line. Deliberately shares no vocabulary with
// formatAnchorRefusal — gatesig classifies transcripts by substring, so a pass line
// echoing refusal words would corrupt friction attribution (#172). Pure.
func formatAnchorDocsOnly(d reviewAnchorDelta) string {
	return fmt.Sprintf("boundary review: anchored to %s; %d doc-only commit(s) arrived since — "+
		"no code surface, the reviewed code is intact (#194)", shortSHA(d.Reviewed), len(d.Commits))
}

// formatAnchorRefusal renders the refusal, naming EVERY commit the review did not
// cover (Done-when: "reported by commit, not as a bare HEAD changed"). Pure.
func formatAnchorRefusal(d reviewAnchorDelta, outcome anchorOutcome, verb string) string {
	if outcome == anchorDiverged {
		return fmt.Sprintf("history moved off the reviewed commit %s (HEAD %s is not a descendant) — "+
			"the review cannot be attributed to this history; re-run `%s`",
			shortSHA(d.Reviewed), shortSHA(d.Current), verb)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d code commit(s) landed after the reviewed commit %s — unreviewed:",
		len(d.Commits), shortSHA(d.Reviewed))
	for _, c := range d.Commits {
		fmt.Fprintf(&b, "\n    %s  %s", shortSHA(c.SHA), c.Subject)
	}
	var code []string
	for _, p := range d.Paths {
		if publishGateHasCodeSurface([]string{p}) {
			code = append(code, p)
		}
	}
	if len(code) > 0 {
		fmt.Fprintf(&b, "\n  code surface: %s", strings.Join(code, ", "))
	}
	fmt.Fprintf(&b, "\n  Re-run `%s` so the review covers them. (Doc-only commits during a review "+
		"are fine — they finalize on their own, #194.)", verb)
	return b.String()
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/sdlc/ -run 'TestClassifyReviewAnchor|TestAnchorPassLine|TestAnchorRefusal' -v`
Expected: PASS (6 subtests + 2).

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/reviewanchor.go cmd/sdlc/reviewanchor_test.go
git commit -m "#194: pure classifier for the reviewed-SHA..HEAD delta

Reuses publishGateHasCodeSurface — the predicate sdlc merge already applies
at publish time (#174) — one stage earlier, rather than restating it
(ARCH-DRY). Decision logic is pure; the git reads sit in a thin shell."
```

---

### Task 3: Wire the classifier into the finalize check

**Files:**
- Modify: `cmd/sdlc/close.go:1182-1227` (`closeReviewSnapshot`, `validate`)
- Modify: `cmd/sdlc/close.go:1128-1152` (`finalizeBoundaryReview` — surface the pass line)
- Test: `cmd/sdlc/close_finalize_test.go`

- [ ] **Step 1: Write the failing integration test** — the interleaving that motivated
      the issue (Done-when #6). Model it on the existing
      `TestCloseCommand_HEADChangedDuringBoundaryReview_DoesNotFinalize` at
      `close_finalize_test.go:139`, which already blocks the fake reviewer on a channel
      and commits concurrently.

```go
// cmd/sdlc/close_finalize_test.go
// #194: a DOC-ONLY commit landing mid-review is a delta, not an invalidation.
func TestCloseCommand_DocOnlyCommitDuringBoundaryReview_Finalizes(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	started := make(chan struct{})
	releaseReview := make(chan struct{})
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		close(started)
		<-releaseReview
		return []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n"), nil
	}

	done := make(chan struct {
		stdout, stderr string
		err            error
	}, 1)
	go func() {
		stdout, stderr, err := executeSDLCTestCommand("close", "--issue", "69", "--actual", "1",
			"--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir,
			"--brain-dir", "../nonexistent-brain")
		done <- struct {
			stdout, stderr string
			err            error
		}{stdout, stderr, err}
	}()

	waitForSignal(t, started, "boundary review to start")
	if err := os.MkdirAll("workshop", 0o755); err != nil {
		t.Fatalf("mkdir workshop: %v", err)
	}
	if err := os.WriteFile("workshop/lessons.md", []byte("- learned a thing\n"), 0o644); err != nil {
		t.Fatalf("write lessons: %v", err)
	}
	git(t, "", "add", "workshop/lessons.md")
	git(t, "", "commit", "-q", "-m", "#69: lessons from the review")
	close(releaseReview)

	var got struct {
		stdout, stderr string
		err            error
	}
	select {
	case got = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for close")
	}
	if got.err != nil {
		t.Fatalf("doc-only delta must finalize, got error: %v\n%s", got.err, got.stderr)
	}
	text := readIssue(t, issuesDir)
	for _, want := range []string{"status: codecomplete", "closed — tests pass"} {
		if !strings.Contains(text, want) {
			t.Fatalf("close did not finalize; missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(got.stderr, "doc-only commit(s) arrived since") {
		t.Fatalf("expected the anchored pass line on stderr:\n%s", got.stderr)
	}
}
```

Then **retarget** the existing code-delta test at `close_finalize_test.go:139` — it
commits `other.txt`, which is code surface, so it must still refuse, but now with the
commit named. Add to its assertions:

```go
	if !strings.Contains(got.err.Error(), "concurrent #69 side change") {
		t.Fatalf("refusal must name the unreviewed commit, got %v", got.err)
	}
	if strings.Contains(got.err.Error(), "HEAD changed from") {
		t.Fatal("refusal should report commits, not bare HEAD identity")
	}
```

- [ ] **Step 2: Run both, confirm the new one fails**

Run: `go test ./cmd/sdlc/ -run 'DuringBoundaryReview' -v`
Expected: `DocOnlyCommit…` FAILS with `boundary review stale: HEAD changed from …`;
`HEADChanged…` fails on the two new assertions.

- [ ] **Step 3: Rewrite `validate()`**

```go
// cmd/sdlc/close.go
type closeReviewSnapshot struct {
	// reviewed is the CONCRETE SHA the boundary review read (#194) — the anchor
	// validate() classifies against. The issue/project text is snapshotted too:
	// the review READ that prose, so a mid-review edit to it is still a genuine
	// invalidation (only the HEAD question was ever stricter than its purpose).
	reviewed  string
	issuePath string
	issueText string
	projects  []projectEdit
}

// validate reports whether the finalize may proceed. On a doc-only delta it returns
// (pass line, nil) so the caller can say what it decided; on a code delta or a
// rewritten history it returns an error naming the commits (#194).
func (s closeReviewSnapshot) validate() (string, error) {
	note := ""
	if s.reviewed != "" {
		d, err := gatherReviewAnchorDelta(s.reviewed)
		if err != nil {
			return "", err // fail closed on a git error, as the publish gate does
		}
		switch outcome := classifyReviewAnchor(d); outcome {
		case anchorDocsOnly:
			note = formatAnchorDocsOnly(d)
		case anchorCodeDelta, anchorDiverged:
			return "", fmt.Errorf("%s", formatAnchorRefusal(d, outcome, closeVerb(s.milestone)))
		}
	}
	if s.issuePath != "" {
		data, err := os.ReadFile(s.issuePath)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", s.issuePath, err)
		}
		if string(data) != s.issueText {
			return "", fmt.Errorf("%s changed", s.issuePath)
		}
	}
	for _, e := range s.projects {
		data, err := os.ReadFile(e.path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", e.path, err)
		}
		if string(data) != e.oldText {
			return "", fmt.Errorf("%s changed", e.path)
		}
	}
	return note, nil
}
```

Add a `milestone string` field to the snapshot, set from `f.Milestone` in
`captureCloseReviewSnapshot`, so the refusal names the right re-run verb (`sdlc close`
vs `sdlc milestone-close`) via the existing `closeVerb` (ARCH-DRY).

- [ ] **Step 4: Update the caller**

```go
// cmd/sdlc/close.go — finalizeBoundaryReview, the validate branch
if validate != nil {
	note, err := validate()
	if err != nil {
		emitTrailerBlock(stdout, review, kind)
		cwarn(stderr, fmt.Sprintf("boundary review: reviewed state changed while the lock was released — close NOT finalized: %v", err))
		return fmt.Errorf("boundary review stale: %w", err)
	}
	if note != "" {
		cinfo(stderr, note)
	}
}
```

Note the second `cwarn` ("re-run `%s` so the review covers the current repo state") is
**removed** — `formatAnchorRefusal` already carries a precise re-run instruction, and the
issue/project-changed branches keep their own message. Verify no test asserts on the
removed line: `grep -rn "so the review covers the current repo state" cmd/sdlc/`.

Change the `validate func() error` parameter type to `func() (string, error)`; the
non-locked callers pass `nil` and are unaffected.

- [ ] **Step 5: Run the package**

Run: `go test ./cmd/sdlc/ 2>&1 | tail -20`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/
git commit -m "#194: classify the mid-review delta instead of refusing on HEAD identity

A doc-only commit arriving during the ~20-minute review no longer discards
it; a code commit still refuses, but now names the commits it did not cover.
The publish-time invariant is untouched: only a delta with no code surface
finalizes, so no code ships unreviewed."
```

---

### Task 4: Docs + issue record

**Files:**
- Modify: `atlas/workflow/ledger-landscape.md:102` (`Review-Window:` now a SHA pair)
- Modify: `cmd/sdlc/helptext/close.md`, `cmd/sdlc/helptext/milestone-close.md`
- Modify: `workshop/issues/000194-review-anchor-commit.md` (`## Log`, tick `## Plan`)

- [ ] **Step 1: Locate every doc that states the old rule**

Run: `grep -rn "HEAD changed\|reviewed-HEAD-unchanged\|Review-Window" atlas/ cmd/sdlc/helptext/`
Read each hit; update the ones that describe the *close-time* rule. **Leave the
publish-time `reviewed-HEAD-unchanged` wording alone** — that invariant is unchanged.

- [ ] **Step 2: Add the close-time rule to the helptext**

In `cmd/sdlc/helptext/close.md` (and the milestone twin), under the boundary-review
section:

```markdown
  The review is anchored to the commit it read (#194). A commit landing while
  the review runs is classified, not fatal:
    doc-only delta → the close finalizes; the reviewed code is intact
    code delta     → refused, naming the commits the review did not cover
  Committing lessons/atlas/plan bookkeeping during a review is therefore safe.
```

- [ ] **Step 3: Tick the plan + write the Log entry** in the issue file, citing ARCH-DRY
      (reused `publishGateHasCodeSurface`) and ARCH-PURE (classifier pure, git in a
      shell), and recording the design decision that a code delta refuses rather than
      finalizes, with the publish-gate reason.

- [ ] **Step 4: Full verification**

```bash
go build ./... && go test ./... 2>&1 | tail -20
```
Expected: all packages PASS.

- [ ] **Step 5: Commit**

```bash
git add atlas/ cmd/sdlc/helptext/ workshop/issues/
git commit -m "#194: document the anchored close-time review rule"
```

---

## Done-when → task map

| Issue Done-when | Satisfied by |
|---|---|
| Review records the SHA it reviewed, finalizes against it | Task 1 (resolve + thread + trailer/sidecar), Task 3 (`validate`) |
| Mid-review commit classified, gate says what it decided | Task 2 (`classifyReviewAnchor`), Task 3 (pass line / refusal) |
| Doc-only delta finalizes via the *same* classifier as merge | Task 2 — calls `publishGateHasCodeSurface`, no second impl |
| Code delta reported by commit, not bare "HEAD changed" | Task 2 (`formatAnchorRefusal`), asserted in Task 2 Step 1 + Task 3 Step 1 |
| Publish-time invariant unchanged | `publishgate.go` untouched; design note above explains why a code delta must refuse |
| Test covers the interleaving | Task 3 Step 1 (both directions: doc-only finalizes, code refuses) |

## Out of scope

- Re-anchoring `runPublishGate` on a recorded reviewed-SHA (would be needed to finalize
  *through* a code delta; a change to the publish contract this issue does not ask for).
- Re-attachable / concurrent reviews — the issue lists these as a second-order effect
  this work merely unblocks, explicitly "not required here".
- `revCount`'s diverged/zero-ahead conflation in the publish gate.
