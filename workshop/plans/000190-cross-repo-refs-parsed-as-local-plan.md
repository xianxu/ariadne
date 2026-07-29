# Cross-Repo Issue Refs Parsed As Local — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop `pair#127` from resolving to local issue 127, so a cross-repo reference never
absorbs another issue's measured hours.

**Architecture:** The repo-qualified ref grammar already has a canonical validator
(`parseRef`) and one derived scanner (`migrate.go`'s `refScanRE`). This work moves the SCAN
half into `internal/issueref` so it can be reached from the internal packages, then routes
**five** call sites through it — the three broken ones plus `migrate.go`'s two. Net encodings
of the grammar: **5 → 1**, achieved by exporting the qualifier+id *fragment* so an anchored
variant composes instead of restating.

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

That last row is why the parser captures the qualifier rather than only asserting a boundary:
`\B#` alone would make `ariadne#180` foreign inside ariadne, trading one bug for a smaller one.

### The three broken sites, and why fixing one is not enough

| site | feeds | consequence today |
|---|---|---|
| `gitx/window.go:384` `issueRefRE` | `DiscoverWindowIssues` → `Options.Issues`, the tracked mention set | admits a foreign issue as a mention target |
| `activetime/commit.go:67` `allIssuePattern()` | `Commit.Issues` → `selectClaimant`/`attributeRun` | **commit-weighted** share splits equally with the foreign issue (`attributeRun`: `perCommit := weight * active / len(commitIssues)`) |
| `activetime/util.go:34` `issuePattern()` | `parseEventMentions` → mention counts | every `pair#127` in transcript prose counts as local `#127` |

The middle row invalidates the issue's originally-filed Spec: attribution is corrupted on the
**commit** path too, so "make commit boundaries outrank mentions" would have left the bug in
place.

### What already exists, and must not be duplicated

Verified before designing:

- **`parseRef` (`resolve.go:50-57`) is the canonical ref grammar** — `[repo]#id [Mx]`,
  `gh#id`, id 1–6 digits — documented in `helptext/resolve.md` and explicitly single-sourced.
- **`refScanRE` (`migrate.go:39-45`) is `([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})\b`** — the
  prose *candidate* scanner, whose own doc says it is "Derived from parseRef's grammar but not
  a second authority."
- **`spanRefRE` (`migrate.go:50-55`) is the same grammar ANCHORED**, plus an optional milestone
  tag: `^([A-Za-z0-9][A-Za-z0-9_.-]*)?#[0-9]{1,6}( M[0-9]+[a-z]?)?$` — the whole-span
  discriminator that keeps `` `#171` `` rewriting while `` `git log --grep "^#15"` `` does not.

So the qualified-ref grammar is already solved and already single-sourced. This work must
**join** that lineage, not start a parallel one. `parseRef` lives in package `main`, which
internal packages cannot import — that is the only reason a new package is needed at all, and
it is why the new package takes the SCAN half and leaves validation where it is.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Ref` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `ScanRE` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `Find` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `LocalNums` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `CountLocal` | `cmd/sdlc/internal/issueref/ref.go` | new |
| `issueRefRE` | `cmd/sdlc/internal/gitx/window.go` | deleted |
| `allIssuePattern` | `cmd/sdlc/internal/activetime/util.go` | deleted |
| `issuePattern` | `cmd/sdlc/internal/activetime/util.go` | deleted |
| `uniqueRefs` | `cmd/sdlc/internal/activetime/util.go` | deleted |
| `parseEventMentions` | `cmd/sdlc/internal/activetime/util.go` | modified |
| `refScanRE` | `cmd/sdlc/migrate.go` | deleted |
| `spanRefRE` | `cmd/sdlc/migrate.go` | deleted |

- **Ref** — one parsed `#N`: `{Qualifier, Num string}`. `Qualifier == ""` is a bare ref.
  - **Relationships:** N per text; a value type with no ownership.
  - **DRY rationale:** five encodings of one grammar become one. The duplication is not
    hypothetical — it is *why this bug shipped*: three of the five copies were unbounded, and
    nothing connected them, so fixing whichever a reader happened to find would leave two
    wrong. (The two in `migrate.go` were already correct — they are the ones that show what the
    grammar should have been all along.)
  - **Future extensions:** the retained `Qualifier` is what keeps cross-repo attribution open;
    a `pair#127` row in the active-time table needs exactly this field.

- **ScanRE / QualifiedIDPattern** — `QualifiedIDPattern` is the un-anchored qualifier+id
  fragment as a string **const**; `ScanRE` is it plus `\b`, compiled. The fragment is exported
  because `migrate.go` needs the grammar in two shapes — un-anchored candidate scanning
  (`refScanRE`) and a whole-span anchored discriminator with a trailing milestone group
  (`spanRefRE:55`). A regex cannot be re-anchored, so exporting only the compiled form would
  have forced `spanRefRE` to stay a separate encoding; exporting the fragment lets both
  compose. This is what makes the count 5 → 1 rather than 5 → 2.
  - **Grammar decision:** adopt `refScanRE`'s **`[0-9]{1,6}`**, not `\d+`. It is the bound
    `parseRef` documents and `TestRewriteRefs` pins (a 7+-digit run must match nothing), and
    workshop ids are 6-digit. Widening it here would fork the grammar in the same breath as
    consolidating it.

- **Find** — `Find(text string) []Ref`, every occurrence in order. The single scanner.

- **LocalNums** — `LocalNums(text, selfRepo string) []string`: local numbers, deduped,
  first-seen order (the contract `uniqueRefs` held at both commit sites).

- **CountLocal** — `CountLocal(text, selfRepo string, tracked map[string]bool) map[string]int`.
  - **Contract to preserve:** an EMPTY `tracked` set yields no mentions, matching today's
    `issuePattern(nil) == nil` → "match nothing" guard that `Compute` relies on.

- **`Ref.IsLocal` matches the qualifier EXACTLY** — deliberately unlike `resolveRepoDir`
  (`resolve.go:193-199`), which does exact-basename-wins then unique-prefix. Prefix matching is
  a navigation convenience; here it would be a correctness bug, silently re-introducing the
  cross-repo bleed this issue exists to remove (`brain` would match `brain-family`). Stated
  because the divergence from a sibling convention is intentional.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `DiscoverWindowIssues` | `cmd/sdlc/internal/gitx/window.go` | modified | `git log` |
| `loadWindowCommits` | `cmd/sdlc/internal/activetime/commit.go` | modified | `git log` |
| `selfQualifier` | `cmd/sdlc/internal/activetime/commit.go` | new | `opts.GitRepo` path |

- **DiscoverWindowIssues** — takes the self-qualifier as a **parameter**, and moves onto the
  package `run` shim (Task 2) so it is testable at all.
  - **Why a parameter, not a `RepoTopLevel()` call inside:** `RepoTopLevel` shells out via
    `exec.Command` directly (`window.go:524`), bypassing the `run` shim — so a self-qualifier
    resolved internally would be untestable, and the plan's own must-not-regress row
    (`ariadne#180` stays local) would ship with **no guard**. A `""` or wrong basename would
    silently drop it. There is exactly one production caller (`actual.go:99`), which already
    holds `repoTop`, so passing `filepath.Base(repoTop)` costs nothing and makes the dependency
    explicit.

- **loadWindowCommits** — already shells through the package's `gitRun` shim.

- **selfQualifier** — derives the local repo name from **`opts.GitRepo`**, not the process cwd.
  - **Why not an `Options.RepoName` from `repoIdentity()`:** the commits being parsed come from
    `opts.GitRepo` (`commit.go:42`), which the standalone verb takes as `--git-repo`
    (`activetime.go:203`). A cwd-derived qualifier would, with `--git-repo` pointed at a peer,
    drop that peer's own self-qualified refs as foreign *and* admit ariadne-qualified refs as
    local — reproducing this very bug class inside the diagnostic verb. `sdlc actual` passes
    `repoTop` (`actual.go:110`), so it would have looked correct in every test that goes
    through `actual`. The qualifier must name the repo the commits come from.
  - **Path handling:** `filepath.Base(filepath.Clean(abs))` after the existing `expandUser`;
    yields `""` for `"."`/`"/"`/unresolvable, which degrades to bare-refs-only exactly as
    `IsLocal("")` specifies.

**Test surface.** `issueref` is pure with a colocated `ref_test.go` — no IO, no mocks. Both git
seams keep their existing shims (`run`, `gitRun`), so no new fake is introduced and no external
dependency is added: `ARCH-MOCK` is unaffected.

---

## Chunk 1 — the fix

### Task 1: The `issueref` package

**Files:**
- Create: `cmd/sdlc/internal/issueref/ref.go`
- Create: `cmd/sdlc/internal/issueref/ref_test.go`

- [ ] **Step 1: Write the failing table test.** Every row is a real form from ariadne's own
      history or the #190 investigation — not invented shapes.

```go
func TestFindSeparatesLocalFromForeign(t *testing.T) {
	cases := []struct{ text string; want []Ref }{
		// The convention (312 of 400 recent subjects).
		{"#187 M2: churn — four-bucket classification", []Ref{{Num: "187"}}},
		{"fixes #127, #128", []Ref{{Num: "127"}, {Num: "128"}}},
		{"(#127)", []Ref{{Num: "127"}}},
		{"PR #106", []Ref{{Num: "106"}}},
		// A RANGE stays two LOCAL refs: `-` is a non-word char, so the second is not read as
		// `174-`-qualified. This is the false positive a hand-written preceded-by class
		// would have introduced, and it appears in real history ("#174-#176").
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
		// Self-qualified: parsed WITH its qualifier; localness is the caller's call.
		{"ariadne#180", []Ref{{Qualifier: "ariadne", Num: "180"}}},
		{"no refs here", nil},
	}
	// … assert reflect.DeepEqual(Find(c.text), c.want) …
}

// The {1,6} bound is inherited from refScanRE/parseRef, and RE2's trailing \b makes a
// 7+-digit run match NOTHING (not a truncated 6-digit prefix). TestRewriteRefs pins this
// for migrate; pin it here too, since this is now the source.
func TestFindRejectsOverlongIDs(t *testing.T) {
	if got := Find("#1234567"); got != nil {
		t.Errorf("Find(#1234567) = %+v, want nil (7 digits is not a ref)", got)
	}
	if got := Find("#123456"); len(got) != 1 || got[0].Num != "123456" {
		t.Errorf("Find(#123456) = %+v, want one ref", got)
	}
}

func TestLocalNums(t *testing.T) { /* bare + self-qualified kept, foreign dropped, deduped
	first-seen order; selfRepo "" keeps only bare; "pair#127" never yields "127" */ }

func TestCountLocal(t *testing.T) { /* tracked gates it; empty tracked → empty map (the
	Compute contract); empty text → empty map; pair#127 contributes nothing */ }

// IsLocal is EXACT, unlike resolveRepoDir's prefix matching — a prefix rule would let
// `brain` match `brain-family` and re-introduce the bleed this package removes.
func TestIsLocalIsExactNotPrefix(t *testing.T) {
	if (Ref{Qualifier: "brain", Num: "1"}).IsLocal("brain-family") {
		t.Error("prefix matching would re-introduce cross-repo bleed")
	}
}
```

- [ ] **Step 2: Run to verify it fails.** `go test ./cmd/sdlc/internal/issueref/` → build failure.

- [ ] **Step 3: Implement.** Signatures plus the strategy for each; the package doc carries the
      *why* (the 46-minute misattribution, and that `parseRef` remains the canonical validator
      while this owns the scan half).

```go
type Ref struct{ Qualifier, Num string }

// ScanRE is the grammar, exported so migrate.go derives rather than restates it.
// Same expression as the refScanRE it replaces, including the {1,6} bound.
var ScanRE = regexp.MustCompile(`([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})\b`)

func Find(text string) []Ref                                    // ScanRE.FindAllStringSubmatch → Refs, in order
func (r Ref) IsLocal(selfRepo string) bool                      // bare, or qualifier EXACTLY == selfRepo
func LocalNums(text, selfRepo string) []string                  // filter Find, dedupe first-seen
func CountLocal(text, selfRepo string, tracked map[string]bool) map[string]int
```

- [ ] **Step 4: Run to verify PASS.** `go test ./cmd/sdlc/internal/issueref/ -v`

- [ ] **Step 5: Mutation-verify the boundary.** Replace `ScanRE` with the old unbounded
      `#(\d+)\b` (Num from group 1, no qualifier). Expect
      `TestFindSeparatesLocalFromForeign`, `TestLocalNums`, `TestCountLocal` and
      `TestFindRejectsOverlongIDs` to FAIL. Restore. A guard that cannot fail is worse than none.

- [ ] **Step 6: Commit.**

---

### Task 2: `gitx.DiscoverWindowIssues` derives from `issueref`

**Files:**
- Modify: `cmd/sdlc/internal/gitx/window.go` (delete `issueRefRE` `:384`; rewrite the scan `:404`; route the git call through `run`)
- Modify: `cmd/sdlc/internal/gitx/window_test.go`

- [ ] **Step 1: Write the failing test.** Override the package `run` shim — the established
      pattern in this file (`window_test.go:262,307,334`).

```go
// The entry point of the chain: whatever lands here becomes Options.Issues, which becomes
// the mention pattern. A foreign ref must not be admitted.
func TestDiscoverWindowIssuesExcludesForeignRefs(t *testing.T) {
	// run = fixture returning "#187 M2: pair#127 replay harness…\n#187 M2: churn…\n"
	// assert result == ["187"], and specifically that "127" is absent.
}
```

> **`DiscoverWindowIssues` uses `exec.Command` directly (`window.go:394`), not the `run` shim** —
> verified. Route it through `run` as part of this task: the shim's own doc says all new callers
> in the package should use it, and the test above is impossible without it. Call it out in the
> commit body as a seam correction, not scope creep.

- [ ] **Step 2: Run to verify it fails** — `127` present.

- [ ] **Step 3: Implement.** Delete `issueRefRE`; add a `selfRepo string` parameter and scan
      each subject with `issueref.LocalNums(line, selfRepo)`. Update the one production caller
      (`actual.go:99`) to pass `filepath.Base(repoTop)`.

      **The self-qualified case gets its own test** — `ariadne#180` with `selfRepo: "ariadne"`
      must yield `180`, and with `selfRepo: ""` must not. That is the must-not-regress row from
      the measurement table, and making `selfRepo` a parameter is what allows it to be asserted
      at all.

- [ ] **Step 4: Run the whole package.** `CommitWindow`, `subjectAnchorRE` and
      `IsShippedWorkSubject` must be untouched. `go test ./cmd/sdlc/internal/gitx/ -v`

- [ ] **Step 5: Commit.**

---

### Task 3: `activetime` derives from `issueref` on both paths

**Files:**
- Modify: `cmd/sdlc/internal/activetime/util.go` (delete `issuePattern`, `allIssuePattern`, `uniqueRefs`; rewrite `parseEventMentions`)
- Modify: `cmd/sdlc/internal/activetime/commit.go` (`:67`, plus `selfQualifier`)
- Modify: `cmd/sdlc/internal/activetime/compute.go` (build `tracked` once; thread the qualifier)
- Modify: `cmd/sdlc/internal/activetime/event.go` (the `pat` parameter threading)
- Modify: `cmd/sdlc/internal/activetime/{util_test.go,commit_test.go,compute_test.go}`

**No CLI change and no `Options.RepoName`** — the qualifier comes from `opts.GitRepo`, which both
callers already set correctly. See the `selfQualifier` rationale in Core concepts.

- [ ] **Step 1: Write the failing tests — BOTH paths, because both are broken.**

```go
// The COMMIT path — the one the filed Spec missed. Commit.Issues drives selectClaimant and
// attributeRun, which splits weight*active EQUALLY across the slice, so a foreign entry
// silently takes half a commit's weighted share.
func TestCommitIssuesExcludeForeignRefs(t *testing.T) {
	// withGitRun (commit_test.go:9) is the package's existing shim-swap helper — reuse it
	// rather than hand-rolling save/restore (ARCH-DRY).
	withGitRun(t, func(dir string, args ...string) ([]byte, error) {
		return []byte("abc1234\t2026-07-29T10:30:00-07:00\t#187 M2: pair#127 replay harness\n"), nil
	})
	// loadWindowCommits(repo="/tmp/…/ariadne", …) → Issues == ["187"] only.
}

// The qualifier names the repo the COMMITS come from, not the cwd. With a peer repo, that
// peer's self-qualified refs are local and ariadne's are foreign — the inverse of the cwd
// bug this avoids.
func TestSelfQualifierComesFromGitRepo(t *testing.T) {
	// gitRun fixture subject: "pair#129 M1: …  (also ariadne#180)"
	// loadWindowCommits(repo=".../pair", …) → Issues contains "129", NOT "180".
}

// The MENTION path.
func TestParseEventMentionsExcludesForeignRefs(t *testing.T) {
	// "working #187; replaying pair#127; more #187", tracked{187,127}, self "ariadne"
	// → {187: 2}; 127 absent.
}

// End to end through Compute: the #190 shape — one run, no claimant, prose naming both.
func TestComputeDoesNotAttributeToForeignIssue(t *testing.T) {
	// PerIssue["127"] == 0; PerIssue["187"] holds the whole segment.
}
```

- [ ] **Step 2: Run to verify they fail.** `go test ./cmd/sdlc/internal/activetime/ -run Foreign -v`

- [ ] **Step 3: Implement.**
  - `parseEventMentions(text, selfRepo string, tracked map[string]bool)` → delegates to
    `issueref.CountLocal`. The compiled-pattern parameter disappears; the tracked set is data.
  - `loadWindowCommits` → `Issues: issueref.LocalNums(parts[2], selfQualifier(repo))`.
  - `Compute` builds `tracked` from `opts.Issues` once and threads it plus the qualifier in
    place of `pat`.
  - Delete `issuePattern`, `allIssuePattern`, `uniqueRefs`.
  - **Preserve the empty-set contract:** `compute_test.go`'s existing no-issues cases are the
    guard and must pass unchanged.

- [ ] **Step 4: Make the drop OBSERVABLE, not silent.** A foreign ref now contributes nothing;
      if that is invisible, a future reader cannot tell "correctly excluded" from "never seen".
      Add a warning when a window's commits or mentions contained foreign refs, naming the
      qualifier — e.g. `foreign refs ignored: pair#127 (×3)`. Reuses the existing
      `AttributionWarning`/`formatAttributionWarning` channel (`activetime.go:100`); no new
      output surface. **This also disposes the Done-when bullet on per-segment rule reporting**
      — see the Revisions note.

- [ ] **Step 5: Run the full suite.** `go test ./cmd/sdlc/... -v`
      **`parity_test.go` — CHECKED, unaffected:** `grep '[A-Za-z]#[0-9]'` over it returns
      nothing, so no fixture carries a qualified ref. If one is ever added, a legitimately
      changed expectation must be edited with a comment; bug-for-bug parity with the superseded
      Python is not a property worth defending.

- [ ] **Step 6: Commit.**

---

### Task 4: `migrate.go`'s BOTH encodings derive from `issueref`

**Files:**
- Modify: `cmd/sdlc/migrate.go` (delete `refScanRE` `:45` and `spanRefRE` `:55`)

The task that makes this a consolidation rather than a new encoding (`ARCH-DRY`).

- [ ] **Step 1: Replace `refScanRE` with `issueref.ScanRE`.** Group indices are identical, so
      its call sites are unchanged. Keep the comment recording that every candidate is still
      filtered through `parseRef` — `issueref` owns the SCAN, `parseRef` owns VALIDATION, and
      each doc should name the other.

- [ ] **Step 2: Recompose `spanRefRE` from the shared fragment.**

```go
var spanRefRE = regexp.MustCompile(`^` + issueref.QualifiedIDPattern + `( M[0-9]+[a-z]?)?$`)
```

      **Check the group indices before assuming this is transparent.** Today `spanRefRE` does
      NOT capture the id (`#[0-9]{1,6}`), while `QualifiedIDPattern` does — so the milestone
      group shifts from 2 to 3. Confirm whether any call site reads a submatch or only
      `MatchString`; if it reads groups, update the index in the same commit. This is exactly
      the kind of silent shift the composition is worth doing carefully.

- [ ] **Step 3: Run migrate's existing tests unchanged.** `TestRewriteRefs` is the guard that
      the grammar did not move — including the 7+-digit case — and the `#179` span cases
      (a styled `` `#171` `` rewrites; a quoted `` `git log --grep "^#15"` `` must not) are the
      guard for `spanRefRE`. `go test ./cmd/sdlc/ -run 'Migrate|RewriteRefs|Span' -v`
      Expected: PASS with **no test edits**. If a test needs changing, the grammar moved and
      that is a finding, not a fixup.

- [ ] **Step 4: Commit.**

---

### Task 5: The regression check with a known answer

**Files:**
- Create: `workshop/plans/000190-evidence.md`

The one check whose correct answer is already known: **46.1 minutes currently charged to
ariadne#127 must return to #187.**

- [ ] **Step 1: Record the BEFORE state.** Captured live before implementation:

```
#187 84.5m/50% mention fallback without issue commit boundary (2026-07-29 10:25 → 12:36)
#127 46.1m/77% mention fallback without issue commit boundary (2026-07-29 10:25 → 12:36)
#127 46.1m/77% dominant long attribution segment              (2026-07-29 10:25 → 12:36)
```

- [ ] **Step 2: Re-measure the same window.** The corrected invocation — **verified against
      `activetime.go`**: `--dir`, `--git-repo` and `--issue` are all REQUIRED (`:29-41`, exit 2
      otherwise); `--issue` is repeatable `StringArrayVar` (`:224`), *not* `--issues` with a
      comma list; the flag is `--threshold-min` (`:227`). `--include-assistant` matches what
      `sdlc actual` uses (`actual.go:114`), without which the streams are not comparable.

```bash
sdlc active-time \
  --dir ~/.claude/projects/-Users-xianxu-workspace-ariadne \
  --dir ../brain \
  --git-repo /Users/xianxu/workspace/ariadne \
  --issue 187 --issue 127 \
  --since 2026-07-29T10:00:00-07:00 --until 2026-07-29T13:00:00-07:00 \
  --threshold-min 15 --include-assistant
```

> Run the BEFORE measurement with this same command *before* Task 1 lands, so the comparison is
> like-for-like. Resolve the exact `--dir` values from what `transcripts.Select` picks for
> `sdlc actual` (`actual.go:101-103`) rather than guessing the slug.

- [ ] **Step 3: Assert three outcomes** in the evidence file: `127` receives **0** minutes; `187`
      gains what `127` held; no warning names `127` except the new foreign-refs-ignored line.

- [ ] **Step 4: Corroborate with `sdlc actual --issue 187`.** **Verified it works on the
      archived issue** (`computeActual` resolved it and reported 2.83h). But its window is
      `<first-commit> → HEAD`, so the number drifts as unrelated work lands — it is a
      corroborating signal, not the stable comparison. Step 2's fixed window is the measurement.

- [ ] **Step 5: State the ledger consequence honestly.** #187's calibration row records actual
      2.32h / ratio 3.6×; the true actual is higher, so the recorded ratio is too generous.
      **Do not rewrite the row** — #117's integrity rule and #178's measured-not-typed gate exist
      to keep that history honest. Record the corrected figure and note the row predates the fix.
      Whether to re-measure historical rows is a separate decision with its own issue.

- [ ] **Step 6: Commit the evidence.**

---

### Task 6: Atlas + close

- [ ] **Step 1: Document the rule** in `atlas/workflow/ledger-landscape.md`'s "How many hours
      did this issue actually take?" worked example (`:43`), which names the engine and its unit
      but not what counts as a ref: bare `#N` and `<thisrepo>#N` are local; `<other>#N` is
      foreign and attributable to no local issue; `-`/`.` are outside the boundary class so
      `#174-#176` stays two refs; the qualifier names the repo the **commits** come from.
      Cross-reference `helptext/resolve.md`, which owns the grammar.

- [ ] **Step 2:** `go test ./... && go vet ./... && sh construct/vocabulary/vet_test.sh`

- [ ] **Step 3:** Tick every Plan row; write the `## Log` entry.

- [ ] **Step 4:** `sdlc actual --issue 190` to preview, then
      `sdlc close --issue 190 --verified '<evidence>'` with `--actual` omitted so close measures
      and adopts it. The binary auto-dispatches the mandatory close review.

- [ ] **Step 5:** `workshop/lessons.md` **only if** the close review surfaces something not
      already prevented by code or tooling. The four-copies-of-one-grammar lesson is now
      code-enforced by the `issueref` consolidation, so it likely does not qualify.

---

## Risks and open questions

1. **`parity_test.go` encoding the bug — CHECKED, it does not.** No fixture carries a qualified
   ref, so v3 parity is unaffected. The reasoning is retained for the day one is added.
2. **Historical ledger rows keep their pre-fix numbers.** Every calibration row measured before
   this fix may be wrong in the same direction. Deliberately out of scope: rewriting measured
   history is what #117/#178 forbid. Task 5 Step 5 records the discrepancy so a future decision
   has data.
3. **`\b`'s word class is ASCII.** A non-ASCII repo name would not be recognized as a qualifier.
   No such repo exists in the workspace; noted so the constraint is a known one.
4. **`gh#id` refs are out of scope.** `parseRef` handles a `gh` token before `#` as a GitHub
   ref; `ScanRE` will parse `gh#42` as qualifier `gh`, num `42` — correctly NOT local, since a
   GitHub issue is not a workshop issue. Stated because it looks like an oversight and is not.
5. **The `--dir` slug in Task 5 must be resolved, not guessed.** Named as a step rather than
   hard-coded, because a wrong transcript dir yields a silent 0-event window (which `--dir`'s
   own exit-2 guard exists to prevent, but only when the flag is missing entirely).

---

## Revisions

### 2026-07-29 — plan-quality round 1: FAILURE, 4 Important + 1 Minor, all confirmed

**Reason:** `sdlc change-code --issue 190` round 1. Every finding was verified against the code
before acting; all four held up, and one changed the plan's central claim.

- **[Important] The grammar already existed twice — the plan added a third encoding while
  claiming to consolidate.** `migrate.go:45` `refScanRE` is
  `([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})\b` — **byte-identical** to what I proposed — and
  its own doc says it is derived from `parseRef` (`resolve.go:50-57`), the canonical grammar
  documented as not-to-be-re-encoded. Verified both. The plan now makes `issueref` the single
  scan source and **Task 4 retires `refScanRE`**, so the count goes 4 → 1 instead of 3 → 1+1.
  Two divergences settled explicitly: adopt the pinned `[0-9]{1,6}` bound (not `\d+`), and keep
  `IsLocal` EXACT rather than adopting `resolveRepoDir`'s prefix matching, because a prefix rule
  would let `brain` match `brain-family` and re-introduce the bleed this issue removes.
- **[Important] The qualifier was derived from the wrong repo.** I had `Options.RepoName` from
  `repoIdentity()` (cwd), while the commits being parsed come from `opts.GitRepo`
  (`commit.go:42`), a `--git-repo` flag (`activetime.go:203`). Pointed at a peer, that drops the
  peer's own refs as foreign and admits ariadne's as local — **this bug class, reproduced inside
  the diagnostic verb**, and invisible to any test going through `sdlc actual` (which passes
  `repoTop`). Now derived from `opts.GitRepo`, which also removes the CLI plumbing and the new
  `Options` field entirely.
- **[Important] Task 5's regression command would have exited 2 before computing.** Verified
  against `activetime.go`: `--dir`, `--git-repo`, `--issue` are all required (`:29-41`); `--issue`
  is repeatable, not `--issues` with commas (`:224`); it is `--threshold-min`, not `--threshold`
  (`:227`). The most valuable verification in the document could not run. Corrected, plus
  `--include-assistant` for comparability with `sdlc actual`, plus a BEFORE run on the same
  command so the comparison is like-for-like.
- **[Important] A Done-when bullet was neither planned nor dropped** — "`sdlc actual` output
  states which rule attributed each segment". **Disposed:** already satisfied, because
  `AttributionWarning.Reason` carries the rule name and `formatAttributionWarning`
  (`activetime.go:100-105`) renders it verbatim — that is exactly how #187's close surfaced
  `mention fallback without issue commit boundary`. But the finding surfaced a real gap next to
  it: after this fix a foreign ref contributes nothing *silently*. Task 3 Step 4 now adds a
  `foreign refs ignored: pair#127 (×3)` warning through that same channel, so the exclusion is
  observable rather than invisible.
- **[Minor] The plan restated the diff.** Full `ref.go` body and five pre-written commit
  messages, stale on arrival. Compressed to signatures plus a strategy line each; commit
  messages dropped. The corpus-derived test table and the mutation-verify step are kept — the
  judge agreed those earn their place, and the mutation step is the mechanical guard this gate
  asks every plan for.

**Also verified before this round, so the gate was not asked to take them on faith:**
`repoIdentity()` returns a bare basename (`reviewsidecar.go:113-126`); `parity_test.go` carries
no qualified refs; `commit_test.go:9` already has a `withGitRun` helper to reuse;
`sdlc actual --issue 187` works on the archived issue file.
