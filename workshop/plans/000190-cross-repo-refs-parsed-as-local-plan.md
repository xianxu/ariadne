# Cross-Repo Issue Refs Parsed As Local — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop `pair#127` from resolving to local issue 127, so cross-repo references never
absorb another issue's measured hours.

**Architecture:** One pure `issueref` package owns the answer to "is this `#N` a local issue
ref?", and the three sites that currently each carry their own copy of `#(\d+)\b` derive from
it. The rule turns out to be expressible directly in RE2 — `\B#(\d+)\b`, the mirror of the
trailing `\b` this codebase already relies on — so the fix is a boundary assertion plus a
qualifier capture, not a hand-rolled scanner.

**Tech Stack:** Go, RE2 (`regexp`), existing `cmd/sdlc/internal/{gitx,activetime}` packages.

---

## Problem, restated from measurement

`#(\d+)\b` has no left boundary. So `pair#127` matches as `127`, and **ariadne#127 exists** —
`000127-recalibrate-estimate-logic-v2-high.md`, archived. At ariadne#187's close, 46.1 minutes
of #187's work were charged to that unrelated issue, corrupting the calibration data it had
itself produced.

Measured over ariadne's last 400 commit subjects:

| form | count | today | wanted |
|---|---|---|---|
| bare `#187` (the §12 convention) | 312 | local ✓ | local |
| foreign `pair#127`, `pair#129`, `pair#105`, `pair#104` | 5 | **local ✗** | foreign |
| self-qualified `ariadne#180` | 1 | local ✓ | local (must not regress) |

That last row is why the parser captures the qualifier instead of just asserting a boundary:
`\B#` alone would make `ariadne#180` foreign inside ariadne, trading one bug for a smaller one.

### The three sites, and why fixing one is not enough

| site | feeds | consequence today |
|---|---|---|
| `gitx/window.go:384` `issueRefRE` | `DiscoverWindowIssues` → `Options.Issues`, the tracked mention set | admits a foreign issue as a mention target |
| `activetime/commit.go:67` `allIssuePattern()` | `Commit.Issues` → `selectClaimant`/`attributeRun` | **commit-weighted** share splits equally with the foreign issue (`attributeRun`: `perCommit := weight * active / len(commitIssues)`) |
| `activetime/util.go:34` `issuePattern()` | `parseEventMentions` → mention counts | every `pair#127` in transcript prose counts as local `#127` |

The middle row is the one that invalidates the issue's originally-filed Spec: attribution is
corrupted on the **commit** path too, so "make commit boundaries outrank mentions" would have
left the bug in place.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Ref` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `Find` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `LocalNums` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `CountLocal` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `issueRefRE` | `cmd/sdlc/internal/gitx/window.go` | deleted |
| `allIssuePattern` | `cmd/sdlc/internal/activetime/util.go` | deleted |
| `issuePattern` | `cmd/sdlc/internal/activetime/util.go` | deleted |
| `parseEventMentions` | `cmd/sdlc/internal/activetime/util.go` | modified |
| `uniqueRefs` | `cmd/sdlc/internal/activetime/util.go` | deleted |

- **Ref** — one parsed `#N` occurrence: `{Qualifier, Num string}`. `Qualifier == ""` is a bare
  ref; otherwise it is the repo name that preceded the `#`.
  - **Relationships:** N refs per text. No ownership — a value type.
  - **DRY rationale:** collapses three independent copies of `#(\d+)\b` into one source. The
    three copies are precisely why this bug shipped: fixing the regex a reader happens to find
    would leave the other two paths wrong, and nothing connects them today.
  - **Future extensions:** the retained `Qualifier` is what keeps cross-repo attribution open —
    a future `pair#127` row in the active-time table needs exactly this field. Deliberately
    parsed and kept, not discarded, even though nothing reads it yet beyond the local check.

- **Find** — `Find(text string) []Ref`, every occurrence in order.
  - **DRY rationale:** the single scanner. `LocalNums` and `CountLocal` are thin filters over it.

- **LocalNums** — `LocalNums(text, selfRepo string) []string`, deduped bare numbers preserving
  first-seen order, keeping refs whose qualifier is empty **or** equals `selfRepo`.
  - **Relationships:** replaces `uniqueRefs(allIssuePattern(), …)` at the commit sites, which
    today do the same dedupe-preserving-order job in two places.

- **CountLocal** — `CountLocal(text, selfRepo string, tracked map[string]bool) map[string]int`,
  mention counts restricted to `tracked`.
  - **Relationships:** replaces `parseEventMentions(text, issuePattern(issues))`. The
    pattern-as-parameter indirection goes away: the tracked set is data, not a compiled regex.
  - **Contract to preserve:** an EMPTY `tracked` set yields no mentions, matching today's
    `issuePattern(nil) == nil` → "match nothing" guard that `Compute` depends on.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `DiscoverWindowIssues` | `cmd/sdlc/internal/gitx/window.go` | modified | `git log` |
| `loadWindowCommits` | `cmd/sdlc/internal/activetime/commit.go` | modified | `git log` |
| `Options.RepoName` | `cmd/sdlc/internal/activetime/compute.go` | new | caller-supplied identity |

- **DiscoverWindowIssues** — already shells to `git log`; gains the self-qualifier by taking the
  repo directory's basename via the existing `RepoTopLevel()`.
  - **Injected into:** nothing new; it hands `[]string` to `Options.Issues` as before.

- **loadWindowCommits** — already shells to `git log` through the package's `gitRun` shim.
  - **Injected into:** `Commit.Issues`, consumed by the pure `selectClaimant`/`attributeRun`.

- **Options.RepoName** — the caller's repo identity, so `activetime` can recognize a
  self-qualified ref without importing `gitx` (it deliberately does not — it carries its own
  `gitRun` shim). Empty string keeps today's bare-only behavior, so no test fixture is forced
  to know a repo name.
  - **Injected into:** `CountLocal` and `LocalNums` calls inside the package.
  - **Future extensions:** if cross-repo attribution lands, this becomes "the local repo" in a
    set of known repos.

**Test surface.** `issueref` is pure with a colocated `ref_test.go` — no IO, no mocks. The two
git-touching integration points keep their existing test styles (`gitx` has its `run` shim;
`activetime` has `gitRun`), so no new fake is introduced and `ARCH-MOCK` is unaffected: this
change adds no external dependency.

---

## Chunk 1 — the fix

### Task 1: The `issueref` package

**Files:**
- Create: `cmd/sdlc/internal/issueref/ref.go`
- Create: `cmd/sdlc/internal/issueref/ref_test.go`

- [ ] **Step 1: Write the failing table test.** Every row below is a real form taken from
      ariadne's own history or the #190 investigation — not invented shapes.

```go
package issueref

import (
	"reflect"
	"testing"
)

func TestFindSeparatesLocalFromForeign(t *testing.T) {
	cases := []struct {
		text string
		want []Ref
	}{
		// The convention (312 of 400 recent subjects).
		{"#187 M2: churn — four-bucket classification", []Ref{{Num: "187"}}},
		{"fixes #127, #128", []Ref{{Num: "127"}, {Num: "128"}}},
		{"(#127)", []Ref{{Num: "127"}}},
		{"PR #106", []Ref{{Num: "106"}}},
		// A RANGE must stay two local refs. `-` is a non-word char, so the boundary rule
		// does not mistake the second for a qualified ref — this is the false positive a
		// hand-written "preceded by [A-Za-z0-9_.-]" class would have introduced.
		{"#174-#176", []Ref{{Num: "174"}, {Num: "176"}}},
		// The bug: one subject carrying both a local and a foreign ref.
		{"#187 M2: pair#127 replay harness + round 1 evidence",
			[]Ref{{Num: "187"}, {Qualifier: "pair", Num: "127"}}},
		// Every real repo-name shape in the workspace.
		{"pair#127", []Ref{{Qualifier: "pair", Num: "127"}}},
		{"brain-family#12", []Ref{{Qualifier: "brain-family", Num: "12"}}},
		{"parley.nvim#12", []Ref{{Qualifier: "parley.nvim", Num: "12"}}},
		{"42shots#12", []Ref{{Qualifier: "42shots", Num: "12"}}},
		{"xianxu.dev#3", []Ref{{Qualifier: "xianxu.dev", Num: "3"}}},
		// Self-qualified: parsed WITH its qualifier; localness is decided by the caller.
		{"ariadne#180", []Ref{{Qualifier: "ariadne", Num: "180"}}},
		{"no refs here", nil},
	}
	for _, c := range cases {
		if got := Find(c.text); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Find(%q) = %+v, want %+v", c.text, got, c.want)
		}
	}
}

// LocalNums keeps bare refs and self-qualified ones, drops foreign, dedupes preserving
// first-seen order (the contract uniqueRefs held at both commit sites).
func TestLocalNums(t *testing.T) {
	const subject = "#187 M2: pair#127 replay harness; also #187 and ariadne#180"
	if got, want := LocalNums(subject, "ariadne"), []string{"187", "180"}; !reflect.DeepEqual(got, want) {
		t.Errorf("LocalNums(selfRepo=ariadne) = %v, want %v", got, want)
	}
	// With no self-repo known, only bare refs are local — and nothing panics.
	if got, want := LocalNums(subject, ""), []string{"187"}; !reflect.DeepEqual(got, want) {
		t.Errorf("LocalNums(selfRepo=\"\") = %v, want %v", got, want)
	}
	// A foreign ref must never appear, even when its number matches a real local issue.
	for _, n := range LocalNums("pair#127", "ariadne") {
		if n == "127" {
			t.Error("pair#127 resolved to local 127 — the #190 defect")
		}
	}
}

// CountLocal is the mention counter. The tracked set gates it, and an EMPTY set must yield
// nothing — Compute relies on that (today via issuePattern(nil) == nil).
func TestCountLocal(t *testing.T) {
	text := "working #187, replaying pair#127, more #187, and #190"
	tracked := map[string]bool{"187": true, "127": true, "190": true}
	got := CountLocal(text, "ariadne", tracked)
	want := map[string]int{"187": 2, "190": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CountLocal = %v, want %v — 127 is pair's, not ours", got, want)
	}
	if got := CountLocal(text, "ariadne", nil); len(got) != 0 {
		t.Errorf("an empty tracked set must yield no mentions, got %v", got)
	}
	if got := CountLocal("", "ariadne", tracked); len(got) != 0 {
		t.Errorf("empty text must yield no mentions, got %v", got)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/sdlc/internal/issueref/`
Expected: build failure — `undefined: Ref`, `undefined: Find`.

- [ ] **Step 3: Implement**

```go
// Package issueref answers one question: does this `#N` refer to an issue in THIS repo?
//
// It exists because `#(\d+)\b` — copied into three separate files — has no LEFT boundary, so
// `pair#127` matched as local issue 127. ariadne#127 exists (an archived issue about
// recalibrating estimates), so 46 minutes of ariadne#187's work were charged to it, corrupting
// the very calibration data that issue had produced (ariadne#190).
//
// The rule is expressible directly in RE2: `\B#` asserts that the character before `#` is NOT
// a word character, which is exactly the mirror of the trailing `\b` these patterns already
// used. No lookbehind and no hand-rolled scanner needed. `\b`'s word class is [0-9A-Za-z_],
// which matches how repo names end — and deliberately EXCLUDES `-` and `.`, so a range like
// `#174-#176` stays two local refs rather than reading the second as `174-`-qualified.
//
// Pure: no IO, no clock, no git. The repo's own name arrives as a parameter.
package issueref

import "regexp"

// Ref is one parsed `#N`. Qualifier is the repo name that preceded it ("" for a bare ref).
//
// The qualifier is KEPT rather than discarded once localness is decided: it is the field a
// future cross-repo attribution row (`pair#127` as its own line in the active-time table)
// would need, and dropping it here would foreclose that (ariadne#190 Spec).
type Ref struct {
	Qualifier string
	Num       string
}

// refRE captures an optional repo qualifier and the number. The qualifier must START
// alphanumeric (so a leading `-` or `.` is not swallowed) and may then contain the `-`/`.`
// that real repo names carry: brain-family, parley.nvim, xianxu.dev.
//
// `\B` before the unqualified `#` is the boundary assertion; the qualified alternative
// consumes the word characters itself, so both forms are recognized by ONE pass and the
// caller can tell them apart.
var refRE = regexp.MustCompile(`([A-Za-z0-9][A-Za-z0-9_.-]*)?#(\d+)\b`)

// Find returns every `#N` in text, in order, each tagged with its qualifier.
func Find(text string) []Ref {
	var out []Ref
	for _, m := range refRE.FindAllStringSubmatch(text, -1) {
		out = append(out, Ref{Qualifier: m[1], Num: m[2]})
	}
	return out
}

// IsLocal reports whether r names an issue in the repo called selfRepo. A bare ref is always
// local; a qualified one is local only when the qualifier IS this repo (`ariadne#180` inside
// ariadne). selfRepo "" means "unknown", so only bare refs count.
func (r Ref) IsLocal(selfRepo string) bool {
	return r.Qualifier == "" || (selfRepo != "" && r.Qualifier == selfRepo)
}

// LocalNums returns the local issue numbers in text, deduped preserving first-seen order.
func LocalNums(text, selfRepo string) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range Find(text) {
		if !r.IsLocal(selfRepo) || seen[r.Num] {
			continue
		}
		seen[r.Num] = true
		out = append(out, r.Num)
	}
	return out
}

// CountLocal counts local mentions of tracked issues. An empty tracked set yields an empty
// map — the contract Compute depends on (previously expressed as a nil regexp).
func CountLocal(text, selfRepo string, tracked map[string]bool) map[string]int {
	counts := map[string]int{}
	if len(tracked) == 0 || text == "" {
		return counts
	}
	for _, r := range Find(text) {
		if r.IsLocal(selfRepo) && tracked[r.Num] {
			counts[r.Num]++
		}
	}
	return counts
}
```

> **Note on `refRE` vs the probed `\B#(\d+)\b`.** Both express the same boundary. This form
> additionally *captures* the qualifier, which `\B#` cannot — and the qualifier is required for
> the self-qualified row (`ariadne#180`, 1 of 400 subjects) not to regress. Step 1's
> `#174-#176` case is the guard that the alternation did not reintroduce the greedy-class bug.

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./cmd/sdlc/internal/issueref/ -v`
Expected: PASS, all three tests.

- [ ] **Step 5: Mutation-verify the boundary.** A guard that cannot fail is worse than none.

```bash
# Drop the qualifier alternation → foreign refs become local again.
# Expect TestFindSeparatesLocalFromForeign and TestLocalNums to FAIL.
```
Temporarily change `refRE` to `regexp.MustCompile(`#(\d+)\b`)` with `m[1]` as Num and no
qualifier; confirm the foreign cases fail; restore.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/internal/issueref/
git commit -m "#190: issueref — one boundary-aware answer to 'is this #N ours?'

pair#127 matched #(\d+)\b as local 127, and ariadne#127 exists, so a cross-repo
ref absorbed another issue's hours. The rule is a left boundary — the mirror of
the trailing \b already in use — plus a captured qualifier so a self-qualified
ariadne#180 does not regress to foreign.

The qualifier is kept rather than discarded: it is what a future cross-repo
attribution row would need.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: `gitx.DiscoverWindowIssues` derives from `issueref`

**Files:**
- Modify: `cmd/sdlc/internal/gitx/window.go` (delete `issueRefRE` at `:384`, rewrite the scan at `:404`)
- Modify: `cmd/sdlc/internal/gitx/window_test.go`

- [ ] **Step 1: Write the failing test.** `DiscoverWindowIssues` shells to `git log`, and this
      package's `run` shim exists for exactly this — override it rather than building a repo.

```go
// A subject carrying a foreign ref must not admit that number to the tracked set. This is
// the entry point of the #190 chain: whatever lands here becomes Options.Issues, which
// becomes the mention pattern.
func TestDiscoverWindowIssuesExcludesForeignRefs(t *testing.T) {
	restore := run
	t.Cleanup(func() { run = restore })
	run = func(name string, args ...string) ([]byte, error) {
		return []byte("#187 M2: pair#127 replay harness + round 1 evidence\n" +
			"#187 M2: churn — four-bucket classification\n"), nil
	}
	got, err := DiscoverWindowIssues("2026-07-29T00:00:00Z", "2026-07-30T00:00:00Z", "187")
	if err != nil {
		t.Fatal(err)
	}
	for _, iss := range got {
		if iss == "127" {
			t.Errorf("pair#127 admitted 127 to the tracked set: %v", got)
		}
	}
	if len(got) != 1 || got[0] != "187" {
		t.Errorf("DiscoverWindowIssues = %v, want [187]", got)
	}
}
```

> **`DiscoverWindowIssues` currently uses `exec.Command` directly (`window.go:394`), not the
> `run` shim.** Route it through `run` as part of this task — the shim is documented as the
> path all new callers use, and the test above needs it. Note this in the commit body: it is a
> seam correction, not scope creep, and it is what makes the entry point testable at all.

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/sdlc/internal/gitx/ -run TestDiscoverWindowIssuesExcludesForeignRefs -v`
Expected: FAIL — `127` present.

- [ ] **Step 3: Implement.** Delete `issueRefRE`; scan with `issueref.LocalNums`, passing the
      repo basename as the self-qualifier.

```go
// selfRepoName returns this repo's directory basename — the qualifier that counts as LOCAL
// (`ariadne#180` inside ariadne). "" when the toplevel is unavailable, which degrades to
// bare-refs-only rather than failing: an unknown repo name must not cost us the whole scan.
func selfRepoName() string {
	top, err := RepoTopLevel()
	if err != nil {
		return ""
	}
	return filepath.Base(top)
}
```

Then in `DiscoverWindowIssues`, replace the `issueRefRE.FindAllStringSubmatch` loop with:

```go
	self := selfRepoName()
	for _, line := range strings.Split(text, "\n") {
		for _, num := range issueref.LocalNums(line, self) {
			seen[num] = struct{}{}
		}
	}
```

- [ ] **Step 4: Run to verify it passes.** Run the whole package — `CommitWindow` and the
      `subjectAnchorRE` guards must be unaffected.

Run: `go test ./cmd/sdlc/internal/gitx/ -v`
Expected: PASS, no other test changed behavior.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/internal/gitx/
git commit -m "#190: DiscoverWindowIssues drops foreign refs

The entry point of the chain: whatever lands here becomes Options.Issues and
then the mention pattern, so admitting pair#127 as 127 here is what let a
cross-repo ref collect transcript minutes.

Also routes the git call through the package run shim, which is what makes the
entry point testable — a seam correction the shim's own doc already asks for.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: `activetime` derives from `issueref` on both paths

**Files:**
- Modify: `cmd/sdlc/internal/activetime/util.go` (delete `issuePattern`, `allIssuePattern`, `uniqueRefs`; rewrite `parseEventMentions`)
- Modify: `cmd/sdlc/internal/activetime/commit.go:67`
- Modify: `cmd/sdlc/internal/activetime/compute.go` (add `Options.RepoName`, drop the `pat` plumbing)
- Modify: `cmd/sdlc/internal/activetime/event.go` (the `pat` parameter threading)
- Modify: `cmd/sdlc/internal/activetime/{util_test.go,commit_test.go,compute_test.go}`
- Modify: `cmd/sdlc/actual.go:107`, `cmd/sdlc/activetime.go:206` (supply `RepoName`)

- [ ] **Step 1: Write the failing tests — BOTH paths, because both are broken.**

```go
// The commit path. Commit.Issues drives selectClaimant and attributeRun, and attributeRun
// splits weight*active EQUALLY across the slice — so a foreign entry here silently hands
// half a commit's weighted share to an unrelated local issue. This is the path the issue's
// originally-filed Spec missed entirely.
func TestCommitIssuesExcludeForeignRefs(t *testing.T) {
	// withGitRun (commit_test.go:9) is the package's existing shim-swap helper — reuse it
	// rather than hand-rolling the save/restore (ARCH-DRY).
	withGitRun(t, func(dir string, args ...string) ([]byte, error) {
		return []byte("abc1234\t2026-07-29T10:30:00-07:00\t#187 M2: pair#127 replay harness\n"), nil
	})
	commits, err := loadWindowCommits(".", "2026-07-29T00:00:00Z", "2026-07-30T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 {
		t.Fatalf("want 1 commit, got %d", len(commits))
	}
	if got, want := commits[0].Issues, []string{"187"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Commit.Issues = %v, want %v — a foreign ref would take half the weighted share", got, want)
	}
}

// The mention path.
func TestParseEventMentionsExcludesForeignRefs(t *testing.T) {
	tracked := map[string]bool{"187": true, "127": true}
	got := parseEventMentions("working #187; replaying pair#127; more #187", "ariadne", tracked)
	if got["127"] != 0 {
		t.Errorf("pair#127 counted as local 127: %v", got)
	}
	if got["187"] != 2 {
		t.Errorf("mentions[187] = %d, want 2", got["187"])
	}
}

// End to end through Compute: the #190 shape. One run, no claimant, prose mentioning both a
// local and a foreign ref — the foreign issue must receive ZERO minutes and the local one
// must receive the whole segment, not a proportional slice.
func TestComputeDoesNotAttributeToForeignIssue(t *testing.T) {
	// … events mentioning "#187" and "pair#127" across a gap, no commits …
	res, err := Compute(Options{ /* Files: fixture, Issues: []string{"187","127"}, RepoName: "ariadne", … */ })
	if err != nil {
		t.Fatal(err)
	}
	if res.PerIssue["127"] != 0 {
		t.Errorf("foreign ref drew %.1f minutes", res.PerIssue["127"])
	}
	if res.PerIssue["187"] <= 0 {
		t.Error("the local issue should hold the whole segment")
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./cmd/sdlc/internal/activetime/ -run 'Foreign' -v`
Expected: FAIL on all three — `127` present with non-zero minutes.

- [ ] **Step 3: Implement.**
  - `parseEventMentions(text, selfRepo string, tracked map[string]bool)` → delegates to
    `issueref.CountLocal`. The compiled-pattern parameter disappears; the tracked set is data.
  - `loadWindowCommits` → `Issues: issueref.LocalNums(parts[2], selfRepo)`.
  - `Options.RepoName` threads to both. `Compute` builds `tracked` from `opts.Issues` once and
    passes it down in place of `pat`.
  - Delete `issuePattern`, `allIssuePattern`, `uniqueRefs`.

  **Preserve the empty-set contract.** `issuePattern(nil) == nil` today means "match nothing";
  `CountLocal` with an empty `tracked` returns an empty map. Same behavior, and
  `compute_test.go`'s existing no-issues cases are the guard — they must pass unchanged.

- [ ] **Step 4: Supply `RepoName` at the two CLI call sites.** `cmd/sdlc/actual.go:107` and
      `cmd/sdlc/activetime.go:206`, both `RepoName: repoIdentity()`. **Verified:**
      `repoIdentity()` → `repoNameAndRoot()` → `filepath.Base(root)`
      (`reviewsidecar.go:113-126`), so it already yields the bare basename (`"ariadne"`) — no
      `filepath.Base` wrapper needed, and `""` on failure degrades to bare-refs-only exactly as
      `Ref.IsLocal` specifies.

- [ ] **Step 5: Run the full suite.** `parity_test.go` is the one to watch: it pins v3 parity
      against the Python original, whose regex had the same missing boundary. If a parity case
      contains a qualified ref its expected value changes — and that change is CORRECT. Update
      it with a comment recording why, rather than preserving bug-for-bug parity.

Run: `go test ./cmd/sdlc/... -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/internal/activetime/ cmd/sdlc/actual.go cmd/sdlc/activetime.go
git commit -m "#190: activetime attributes only local refs, on BOTH paths

Commit.Issues was the path the filed Spec missed: attributeRun splits
weight*active equally across it, so a foreign ref took half a commit's weighted
share. Fixing only the mention path would have left that in place.

Options.RepoName carries the self-qualifier so activetime need not import gitx
(it deliberately does not — it has its own gitRun shim).

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: The regression check with a known answer

**Files:**
- Create: `workshop/plans/000190-evidence.md`

This is the one check whose correct answer is already known from #187's close, which makes it
worth more than any fixture: **46.1 minutes currently charged to ariadne#127 must return to
#187**, and #187's measured actual must rise from 2.32h.

- [ ] **Step 1: Record the BEFORE state** from #187's close output (already captured in its
      issue Log): `#127 46.1m/77% mention fallback`, `#187 84.5m/62%`, actual 2.32h.

- [ ] **Step 2: Re-measure the same window after the fix.** #187 is archived, so prefer the
      standalone verb over `sdlc actual --issue 187` (which needs the live issue file — check
      whether `issueFilePath` searches `workshop/history/`; if it does, run both):

```bash
sdlc active-time --since 2026-07-29T10:00:00-07:00 --until 2026-07-29T13:00:00-07:00 \
  --issues 187,127 --threshold 15
```

- [ ] **Step 3: Assert the three outcomes** in the evidence file:
  - `127` receives **0** minutes (it is `pair#127` throughout);
  - `187` gains what `127` held;
  - no `mention fallback` warning names `127`.

- [ ] **Step 4: State the ledger consequence honestly.** ariadne#187's calibration row records
      actual 2.32h and ratio 3.6×; the true actual is higher, so the recorded ratio is too
      generous. **Do not silently rewrite the row** — #117's integrity rule and #178's
      measured-not-typed gate both exist to keep that history honest. Record in the evidence
      file what the corrected measurement is, and note that the row predates the fix. Whether
      to re-measure historical rows is a separate decision with its own issue.

- [ ] **Step 5: Commit the evidence.**

---

### Task 5: Atlas + close

**Files:**
- Modify: `atlas/workflow/ledger-landscape.md` (the "How many hours did this issue actually take?" worked example)
- Modify: `atlas/index.md` only if a new file was added (none planned)

- [ ] **Step 1: Document the rule** where attribution is already described — the worked example
      at `ledger-landscape.md:43` names the engine and its unit but not what counts as a ref.
      Add: bare `#N` and `<thisrepo>#N` are local; `<other>#N` is foreign and attributable to no
      local issue; `-`/`.` are excluded from the boundary class so `#174-#176` stays two refs.

- [ ] **Step 2: `go test ./... && go vet ./...`**

- [ ] **Step 3: Tick every Plan row; write the `## Log` entry.**

- [ ] **Step 4: `sdlc actual --issue 190` to preview; then
      `sdlc close --issue 190 --verified '<evidence>'`** with `--actual` omitted so close
      measures and adopts it. The binary auto-dispatches the mandatory close review.

- [ ] **Step 5:** `workshop/lessons.md` entry **only if** the close review surfaces something
      not already prevented by code or tooling. Candidate, if the review agrees: three copies of
      one regex is how a single-line bug reached three subsystems — the `issueref` extraction is
      the code-enforced half, so a lesson is only warranted for the part tooling cannot catch.

---

## Risks and open questions

1. **`parity_test.go` encoding the bug — CHECKED, it does not.**
   `grep -n '[A-Za-z]#[0-9]' cmd/sdlc/internal/activetime/parity_test.go` returns nothing, so no
   parity fixture carries a qualified ref and parity is unaffected by this change. Recorded as a
   resolved check rather than a live risk, because the reasoning still matters if a fixture is
   ever added: bug-for-bug parity with the superseded Python is not a property worth defending,
   and a legitimately-changed expectation must be edited with a comment, never silently.
2. **Historical ledger rows keep their pre-fix numbers.** Every calibration row measured before
   this fix may be slightly wrong in the same direction. Deliberately out of scope: rewriting
   measured history is exactly what #117/#178 forbid. Task 4 Step 4 records the discrepancy so a
   future decision has data.
3. **`\b`'s word class is ASCII.** A non-ASCII repo name would not be recognized as a
   qualifier. No such repo exists in the workspace; noted so the constraint is a known one.
4. **Foreign refs vanish rather than being reported.** `Ref.Qualifier` is retained so the table
   *could* show `pair#127`, but nothing renders it in this issue. If the operator wants to see
   cross-repo time, that is the widening point — stated so the omission reads as scoped, not
   forgotten.
