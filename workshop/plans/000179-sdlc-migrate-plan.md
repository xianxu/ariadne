# sdlc migrate — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc migrate <file> <dest-repo-dir>` moves a markdown artifact to a peer repo with all repo-relative refs rewritten (bare `#N` → `<source>#N`, `<dest>#M` → `#M`), every touched ref verified to resolve, scoped commits on both sides, and inbound references across the fleet reported — deterministic, binary-owned, no LLM (#179, supporting #171's peer-repo addressing).

**Architecture:** ARCH-PURE two-layer, mirroring `resolve.go`'s documented shape: a pure core (`rewriteRefs` — fence-aware segmentation + the ref-scan regex + the three rewrite rules, string→string+report, no IO) and a thin IO shell (`runMigrate` — guards, existence verification via the existing `resolveArtifacts`, file move, scoped `git -C` commits, inbound `git grep` sweep). ARCH-DRY: the ref *grammar* authority stays `parseRef` (the scanner regex is derived from it and drift-tested against it); fence handling is lifted into an exported `issue.SplitFences` that `stripCodeFences` also routes through; dirty checks reuse the pinned-porcelain posture (`--untracked-files=all`, lessons Rule 4); repo-token→dir mapping reuses `resolveRepoDir` conventions (dest is a PATH argument, simpler).

**Tech Stack:** Go (cmd/sdlc); seams: `resolveArtifacts`/`parseRef` (resolve.go), `gitStatusPorcelain` posture (propagatebase.go:213), `markMutatingCommand` (repolock.go — auto lock wrap), `expectDie` (die_test.go), glob-embedded helptext (a new `helptext/migrate.md` needs no registration).

**Key semantic decisions (settling the issue's design questions):**
- **Issue-family artifacts REFUSE in v1** (renumbering is v2): a file under the vocab Discovery dirs (`workshop/issues|plans|history`) with an `NNNNNN-` prefix needs a fresh dest-side ID; the #171 driver (slug-named project files) doesn't. Refusal names the v2 path.
- **Brain guard is INVERTED from the spine verbs:** migrate must RUN in brain (moving an artifact *out* is the #171 use case), so it does NOT join `WorkflowVerbs()`/`guardSpineRepo`. Instead it refuses when the **destination** is a brain (`.brain/config.md`): "SDLC process artifacts don't live in brain (#171 amendment)."
- **`gh#N` refs are reported, not rewritten** (v1): the space-separated `repo gh#N` form makes prose-scanning ambiguous (rewriting the directly-attached `gh#N` inside an already-qualified `ariadne gh#N` would double-qualify). Rare in artifact bodies; print them as needs-manual-attention.
- **Fenced blocks are skipped; inline code spans rewrite ONLY when the whole span parses as a single ref.** Fences quote grammar/examples verbatim (the #66 meta-document lesson). Single-backtick spans are mixed-content: they style real refs (`` `#171` ``, `` `ariadne#144 M2` `` — must rewrite) but also quote commands and grep patterns (`` `git log --grep "^#15"` ``, `` `nous#41 #11` `` — rewriting corrupts them silently, real counterexamples in this repo's markdown). The grammar-anchored discriminator: a span whose ENTIRE content `parseRef`s as one ref is a styled ref → rewrite; any other span containing `#`-digit matter is skipped + listed in the report (like gh-refs). ARCH-DRY: parseRef stays the sole authority.
- **Scanned candidates are filtered through `parseRef`:** a match that doesn't parse (e.g. `#0` — grammar rejects id 0) is a non-ref → skipped + reported, never rewritten, never verified. This makes "everything rewritten parses" true by construction.
- **Known precision limits, mitigated by transparency + verification:** a hex color `#123456` in prose scans as a ref (6 digits parse) → verification fails to resolve it → the migration refuses naming it (fail-closed; the fix is fencing or rewording the color). "PR #96" scans as issue 96 and, if issue 96 exists, rewrites wrongly — accepted v1 residual; every rewrite is printed (`old → new`, line numbers) and `--no-commit` + `git diff` is the review path. The space-separated form `ariadne #15` (which parseRef would accept as one ref) is scanned as bare `#15` and re-qualified — producing `ariadne src#15`; rare, documented in helptext caveats.
- **No frontmatter mutation in v1** (no `updated:` bump): content changes are refs-only. The body scan INCLUDES frontmatter text (a qualified dep ref would be handled correctly; bare numerics don't match the grammar).
- **The round-trip invariant is canonicalization + idempotence, NOT first-trip byte-equality.** A pre-existing *self-qualified* ref (`src#12` inside src — real, e.g. `ariadne#95` in `workshop/targets/base-layer-mechanics.md`) passes through outbound (correct: it must stay qualified at dest) but normalizes to bare `#12` on the return trip (src is then the dest) — a semantic no-op, but not byte-identical. So: the first return may canonicalize self-qualified → bare; the SECOND round-trip is byte-stable. That is the spec's "ref-stable, no rewrite churn" — every rewrite in steady state is a no-op or a canonicalization, never a flip-flop.
- **Source-repo cleanliness is scoped to the FILE** (must be tracked + unmodified — we migrate reviewed state, and other WIP in the source repo is none of our business, staging is explicit-path per lessons Rule 1/2). **Dest-repo cleanliness is repo-wide** (spec: dirty destination refuses; `--no-clean-check` is the explicit bypass) and the dest path must not already exist.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `refScanRE` + `rewriteRefs` | `cmd/sdlc/migrate.go` | new |
| `refRewrite` | `cmd/sdlc/migrate.go` | new |
| `SplitFences` | `cmd/sdlc/internal/issue/structural.go` | new |

- **`rewriteRefs(body, sourceRepo, destRepo string) (out string, rewrites []refRewrite, skipped []string)`** — the whole rewrite semantics in one pure function. Splits body into fence/prose segments (`issue.SplitFences`); within prose, applies the inline-span rule (a backticked span rewrites only if its entire content `parseRef`s as one ref; other `#`-bearing spans → skipped report); scans with `refScanRE`, filters every candidate through `parseRef` (non-parsing → skipped report), then applies exactly three rules: bare `#N` → `sourceRepo#N`; `destRepo#M` → `#M`; anything else (source-qualified, third-repo, `gh`-marked) passes through — `gh` refs also land in the skipped report. Returns rewrites with line numbers for verification + the printed summary.
  - **Relationships:** consumed only by `runMigrate`; 1:1 with the migration.
  - **DRY rationale:** `parseRef` stays the sole grammar authority — the scanner only *finds candidates*; parseRef decides refhood, and a drift test pins that every REWRITTEN form parses (true by construction once the filter is in).
  - **Future extensions:** v2 issue renumbering adds a fourth rule (self-ref rewrite); `gh` handling graduates from report to rewrite.
- **`refRewrite`** — `{Line int; Old, New string}`; the unit of the printed summary and of verification.
- **`SplitFences(text string) []FenceSegment`** — exported fence segmenter (`{Text string; Fenced bool}`, concatenation-identity: joining segments reproduces the input byte-for-byte; an UNTERMINATED trailing fence is classified Fenced — conservative for rewriters: never rewrite inside a broken fence). NEW, colocated with `stripCodeFences` but **not** a refactor of it: stripCodeFences keeps its exact current semantics (space-replacement, unterminated tail counted as prose) because the structural/sizing gates consume it — a cross-referencing comment records that the two unterminated-fence policies differ deliberately (gate word-count vs rewriter safety).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `runMigrate` + `NewMigrateCmd` | `cmd/sdlc/migrate.go` | new | git (two repos), filesystem, stderr |
| verb registration | `cmd/sdlc/main.go` (`buildRoot`) | modified | cobra |
| `helptext/migrate.md` | `cmd/sdlc/helptext/` | new | embedded help |

- **`runMigrate(o migrateOpts)`** — guard order: (0) the source file lies INSIDE the cwd repo (`markMutatingCommand`'s lock is acquired from `os.Getwd()` — a path reaching into another repo would lock and commit the wrong repo → refuse), and its path is normalized to REPO-ROOT-relative (the `--dest-path` default, both commit subjects, and the inbound sweep all key on the root-relative form, not cwd-relative); (1) source file tracked + unmodified (`git status --porcelain -- <file>` empty; untracked = "commit it first"); (2) issue-family refusal; (3) dest dir is a git repo, is NOT the source repo (same top-level — including a dest path inside the source repo — refuses: with sourceRepo == destRepo the rules degenerate into bare↔qualified flip-flopping), is NOT a brain, dest path free; (4) dest repo clean (pinned porcelain) unless `--no-clean-check`; (5) `rewriteRefs` + verification FROM THE DESTINATION'S VANTAGE: every rewrite's NEW form must resolve against the DEST root (`resolveArtifacts("src#N", destRoot)` / `resolveArtifacts("#M", destRoot)`) — this exercises the exact post-move resolution path, so a dest that is not a sibling of the source (where `src#N` cannot resolve) refuses up front rather than shipping dangling refs; the skipped report prints, any non-resolving rewrite ABORTS before any write. Then: mkdir-p dest parent, write rewritten content, `git -C dest add -- <path>` + commit `migrate: receive <relpath> from <source>`; remove source file, `git -C source add -- <path>` + commit `migrate: move <relpath> to <dest>`; `--no-commit` stages both sides and prints the two commit commands instead. Finally the inbound sweep: for every sibling dir of the source's parent that has a `.git` (INCLUDING source and dest themselves), `git -C <sib> grep -n -F <relpath>` and `-F <basename>` over tracked files, de-duped, excluding the migrated file's own old/new paths, printed as `repo/file:line` (report-only, v1). Refusals go through `die` (expectDie-testable, #63 seam). Note: dest-token matching in the rewriter is exact-basename while `resolveRepoDir` prefix-matches — a `parley#160` ref migrating into `parley.nvim` won't normalize; harmless (still resolves), one-line comment.
  - **Injected into:** nothing — it injects paths/roots into the pure core and `resolveArtifacts`.
  - **Lock:** `markMutatingCommand` (auto `.git/sdlc.lock` on the SOURCE/cwd repo). The dest-repo write is outside the local lock — same accepted posture as close's brain project-file write today; noted in helptext.
- **Registration:** `add(NewMigrateCmd(), "migrate", "Move a markdown artifact to a peer repo, rewriting refs (#179)")` after `open` in the workflow-ordered list (it's a resolve-family tool). NOT added to `processmanual.WorkflowVerbs()` — must run in brain (outbound); a comment at the WorkflowVerbs site records why migrate is absent (the drift test enumerates membership, so absence is deliberate).

**Test surface:** pure — `TestRewriteRefs` table + the grammar round-trip drift test + `TestSplitFences` (concatenation identity, unterminated fence) + `stripCodeFences` regression. IO — fixture-based e2e in `migrate_test.go`: a temp PARENT dir holding two `git init` repos (source with issue `000012-*`, dest with issue `000005-*`), driving `runMigrate` directly (same posture as `TestFindMilestonesMissingVerdict_Integration`); refusal tests via `expectDie`; round-trip e2e asserting byte equality. Mutation-check (#63): invert the bare-ref rule → round-trip + rewrite tests go red.

---

## Chunk 1: pure core

### Task 1: `issue.SplitFences` (new segmenter, TDD)

**Files:** modify `cmd/sdlc/internal/issue/structural.go`, test `cmd/sdlc/internal/issue/structural_test.go`

- [ ] **Step 1:** Failing tests: `TestSplitFences` — segments alternate prose/fenced; concatenation reproduces input byte-for-byte (property over fixtures: no-fence, fence-at-start, back-to-back fences, unterminated trailing fence → tail segment `Fenced: true`).
- [ ] **Step 2:** red → implement: `type FenceSegment struct{ Text string; Fenced bool }`; `func SplitFences(s string) []FenceSegment`. `stripCodeFences` is NOT touched — add the cross-referencing comment on both (deliberately different unterminated-fence policies: gate word-count treats the tail as prose; the rewriter treats it as fenced). Its existing tests stay green trivially.
- [ ] **Step 3:** green; `go test ./cmd/sdlc/internal/issue/`.
- [ ] **Step 4:** Commit `#179: issue.SplitFences — fence segmenter for rewriters`.

### Task 2: `rewriteRefs` (pure, TDD)

**Files:** create `cmd/sdlc/migrate.go`, `cmd/sdlc/migrate_test.go`

- [ ] **Step 1:** Failing table test `TestRewriteRefs` — cases: bare `#12` → `src#12`; `dst#5` → `#5`; `src#12` passthrough (self-qualified — the round-trip canonicalization class); `third#9` passthrough; `gh#4` → untouched + reported; `ariadne gh#4` → untouched + reported (the directly-attached token is `gh`); fenced block containing `#99` untouched; unterminated trailing fence containing `#99` untouched; inline span `` `#12` `` rewritten (whole span = one ref); inline span `` `git log --grep "^#15"` `` untouched + reported (multi-token span); inline span `` `nous#41 #11` `` untouched + reported; milestone form `#15 M4` → `src#15 M4`; span `` `ariadne#144 M2` `` rewritten iff dest/source rules apply (whole span parses with milestone); heading `## Log` no match; 6-digit `#000175`; 7-digit `#1234567` matches NOTHING (pin the `\b` behavior); `#0` skipped + reported (parseRef rejects id 0); `#123456` (hex-alike) IS scanned and rewritten (verification owns rejecting it); punctuation forms `(#12)`, `#12,` match; multiple refs one line; rewrite report carries correct line numbers.
- [ ] **Step 2:** Failing drift test `TestRefScan_GrammarRoundTrip`: every `New` value produced by the table's rewrites must `parseRef` cleanly (true by construction via the parseRef filter — the test pins the construction).
- [ ] **Step 3:** red → implement. Scanner: `` regexp.MustCompile(`([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})\b`) `` over non-fenced segments; candidates through `parseRef` (non-parsing → skipped report); inline-span rule before scanning prose; group 1 empty = bare, `"gh"` = github (report), `== destRepo` = normalize, else passthrough.
- [ ] **Step 4:** green.
- [ ] **Step 5:** Commit `#179: rewriteRefs — pure fence/span-aware ref rewriter + grammar drift test`.

## Chunk 2: IO shell + e2e

### Task 3: `runMigrate` + command

**Files:** extend `cmd/sdlc/migrate.go`; modify `cmd/sdlc/main.go`; create `cmd/sdlc/helptext/migrate.md`

- [ ] **Step 1:** Failing e2e tests (fixture: `migrateRepos(t)` builds a parent-dir with `src/` + `dst/` git repos — each with `git config user.email/user.name` + `commit.gpgsign false`, all seeded files COMMITTED (the tracked-clean guard demands it) — issues `000012-x.md` in src and `000005-y.md` in dst under `workshop/issues/`, and a `data/project/p.md` in src containing `#12`, `dst#5`, `src#12` (self-qualified), `third#9`, a fenced `#99`, and prose). runMigrate is driven with cwd = src (chdir, like closeRepo):
  - happy path: dest file exists with rewritten content; source file gone; `git -C dst log -1` subject `migrate: receive …`; `git -C src log -1` subject `migrate: move …`; both commits touch exactly one path (`git show --stat`); rewrite summary on stderr.
  - dangling ref (`#77` nowhere): `expectDie`, message names `#77`, NOTHING moved (both repos' trees unchanged).
  - dest dirty → refuses (and `--no-clean-check` proceeds); dest path exists → refuses.
  - dest is brain (`.brain/config.md`) → refuses citing #171.
  - dest == source repo (same top-level; also a dest path inside src) → refuses (form-flipping degenerate).
  - source file outside the cwd repo (path reaching into a sibling) → refuses (lock/commit correctness).
  - issue-family source (`workshop/issues/000012-x.md`) → refuses naming v2/renumbering.
  - non-sibling dest (nested one level deeper): `#12`'s new form `src#12` can't resolve from dest → refuses naming it (the vantage-point check).
  - `--no-commit`: both sides staged (`git -C … diff --cached --name-only`), no new commits.
  - source file with uncommitted modification → refuses.
- [ ] **Step 2:** red → implement `runMigrate` per the Integration-points contract; `NewMigrateCmd` (args validation, flags `--dest-path`, `--no-commit`, `--no-clean-check`), `markMutatingCommand`; register in `buildRoot`; write `helptext/migrate.md` (grammar rules, the gh/PR-# caveats, the brain-dest refusal, lock scope note).
- [ ] **Step 3:** green; `go vet ./cmd/sdlc/...`.
- [ ] **Step 4:** Commit `#179: sdlc migrate — IO shell, guards, scoped two-repo commits`.

### Task 4: inbound-ref report + round-trip

- [ ] **Step 1:** Failing tests: (a) a third sibling repo with a tracked md referencing `data/project/p.md` → report contains `sib/notes.md:` line; (b) round-trip (canonicalization + idempotence): migrate src→dst, then dst→src (same runMigrate, roles swapped) → the returned file differs from the original ONLY by the self-qualified canonicalization (`src#12` → `#12`, a semantic no-op — assert the diff is exactly that); then migrate src→dst→src a SECOND time → byte-identical to the once-returned version (idempotence).
- [ ] **Step 2:** red → implement the sibling sweep: `git -C <sib> grep -n -F` over every parent-dir sibling that has a `.git` (including src and dst themselves), excluding the migrated file's own old and new paths.
- [ ] **Step 3:** green. **Mutation-check (#63):** invert the bare-ref rule (`#N` → left untouched) → round-trip AND happy-path content tests must go red; restore.
- [ ] **Step 4:** Commit `#179: migrate — inbound-ref report + round-trip canonicalization/idempotence`.

### Task 5: docs + bookkeeping

- [ ] **Step 1:** Atlas: `atlas/workflow/sdlc-binary.md` — short migrate paragraph near the resolve prose (grammar reuse, brain-dest inversion, v1 scope); check `atlas/index.md` linkage. Note in the WorkflowVerbs comment (repoguard/processmanual) why migrate is deliberately absent.
- [ ] **Step 2:** Verification: `go build -o /dev/null ./cmd/sdlc/ && go vet ./cmd/sdlc/... && go test ./cmd/sdlc/...`; `git diff --check`; live dogfood — build `bin/sdlc`, `sdlc migrate --help`, and a `--no-commit` dry pass on a scratch fixture (NOT a real artifact; the real 5-file migration belongs to #171's execution, not here).
- [ ] **Step 3:** Tick issue Plan, Log (cite ARCH-PURE for the core/shell split, ARCH-DRY for parseRef/SplitFences/porcelain reuse), close per the #174 protocol (bundle everything into the close commit).

### Notes for the implementer

- `resolveArtifacts(refStr, root string) ([]Artifact, ArtifactRef, error)` — not-found returns an ERROR (resolve.go:295); GitHub refs return `(nil, ref, nil)` (but gh refs never reach verification — they're report-only). ALL verification runs against the DEST root (the post-move vantage): `resolveArtifacts("src#N", destRoot)` for re-qualified bare refs, `resolveArtifacts("#M", destRoot)` for dest-normalized ones.
- helptext caveats section must list: PR-#N false-positive, hex-color fail-closed, `gh#` report-only, the space-separated `ariadne #15` → `ariadne src#15` wart, and multi-token inline spans being report-only.
- The scanner regex and `parseRef` MUST stay coupled through the drift test — if the grammar ever gains a form, the round-trip test is what breaks loudly.
- New stderr lines: run `assertNoGatesigCollision` over the refusal + summary lines (close_atlasskip_test.go pattern) — cheap insurance even though migrate isn't in the gate catalog.
- Don't add migrate to `processmanual.WorkflowVerbs()`; its drift test enumerates the spine — absence is the point (runs in brain, outbound).

## Revisions

### 2026-07-15 — close-review (FIX-THEN-SHIP) deltas

1. **Containment guard added** (review I3): `--dest-path` is now checked to
   stay inside the destination repo (`filepath.Rel` mirror of the source-side
   guard) — a traversal value previously wrote a stray file outside the repo
   before `git add` failed, breaking fail-closed. Happy-path + traversal
   tests added (the flag previously had zero coverage).
2. Live dogfood (Task 5) found the symlinked-cwd guard misfire (os.Getwd's
   logical-$PWD vs git's resolved paths) — fixed with EvalSymlinks on both
   sides + a $PWD-setting regression test, already logged in the issue.
3. Minor review items applied: unused `migrateOpts.stdout` dropped (all
   output is stderr); zero-rewrites verification vacuity documented at the
   branch; atlas/index.md hook line mentions migrate.
