# Make Actuals and Estimates Agent-Robust — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn `f62d099`'s one-agent Codex patch into a durable, harness-agnostic contract: transcript discovery becomes an explicit *harness registry*, and estimate derivation gets a single discoverable surface that names both the shared method and the repo-local calibration source — failing loudly when that source is absent.

**Architecture:** Two independent surfaces, two review boundaries.
- **M1 — measurement robustness.** Extract the inline Claude+Codex transcript selection from `actual.go` into a pure `internal/transcripts` package: a `Harness` interface, a `Registry`, and a pure `Select` aggregator over injected harnesses. Add robustness fixtures (malformed/irrelevant Codex sessions) and the missing commit-window matcher tests.
- **M2 — estimator-source pointer.** Mirror the `arch-principles` pattern (single-source + push-at-gate + pull command) for estimate derivation: a pure resolver that names the shared method (sdlc grammar + `vocab.Models()`) and the repo-local calibration doc (`<brain>/…/velocity/<model>.md`, `$WF_ESTIMATOR_SRC` override), a `sdlc estimate-source` pull command, and pushes at `start-plan` + `change-code`'s estimate-block error. Path resolution is DRY'd with `close.go`'s existing ledger resolution.

**Tech Stack:** Go (`cmd/sdlc`), cobra, the existing `internal/{activetime,gitx,estimate,judge,issue}` packages, `//go:embed` helptext.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `transcripts.Sources` | `cmd/sdlc/internal/transcripts/transcripts.go` | new |
| `transcripts.Select` | `cmd/sdlc/internal/transcripts/transcripts.go` | new |
| `transcripts.cwdToClaudeDir` | `cmd/sdlc/internal/transcripts/claude.go` | new (moved from `actual.go`) |
| `transcripts.codexSessionCWD` | `cmd/sdlc/internal/transcripts/codex.go` | new (moved from `actual.go`) |
| `gitx.IsShippedWorkSubject` (tests) | `cmd/sdlc/internal/gitx/window_test.go` | modified |
| `estimate.VelocityPath` | `cmd/sdlc/internal/estimate/source.go` | new |
| `estimate.SourcePath` | `cmd/sdlc/internal/estimate/source.go` | new |
| `estimate.SourceGuidance` | `cmd/sdlc/internal/estimate/source.go` | new |
| `close.go` ledger path | `cmd/sdlc/close.go:628` | modified |
| `estimateSourceLine` (start-plan nudge) | `cmd/sdlc/startplan.go` | modified |

Detail:

- **`transcripts.Sources`** — the engine's input: `{Dirs []string; Files []string}`. `Dirs` are globbed for `*.jsonl` (Claude's one-dir-per-cwd); `Files` are pre-selected session files (Codex's cwd-matched files). This is exactly the shape `activetime.Compute` already consumes (`Options.Dirs`/`Options.Files`), so it is a rename+move of the existing `actualSources`, not a new contract.
  - **Relationships:** 1 `Sources` aggregated from N `Harness` contributions per `actual` invocation.
  - **DRY rationale:** Eliminates the two parallel selection paths inline in `actual.go` (`selectActualDirs` Claude path + `selectCodexSessionFiles` Codex path) by giving every harness one uniform `Sources(cwds)` method merged by one `Select`.
  - **Future extensions:** A third agent CLI (e.g. a future harness) is one new file implementing `Harness` + one registry entry — no change to `actual.go` or the engine. This *is* the issue's stated purpose ("treat transcript discovery as a harness abstraction, not a Claude-only path convention") — ARCH-PURPOSE.

- **`transcripts.Select(cwds []string, hs []Harness) Sources`** — pure aggregator: maps each harness's `Sources(cwds)` and merges+dedups into one `Sources`. Pure over the *injected* harness slice, so unit tests pass fakes/temp-rooted real harnesses; `actual.go` passes `transcripts.DefaultHarnesses()`. ARCH-PURE: the aggregation is pure; each harness's `Sources` method is the thin IO seam (Stat/WalkDir/ReadFile).

- **`transcripts.cwdToClaudeDir`** — the `/`+`.` → `-` cwd→folder encoder (moved verbatim from `actual.go:cwdToTranscriptDir`). Pure; its existing test moves with it.

- **`transcripts.codexCWDFromBytes([]byte) string`** (pure) + **`codexSessionCWD(path)`** (thin `os.ReadFile` seam over it) — extracts `session_meta.payload.cwd`, tolerating malformed/missing records (skip-line loop). The robustness matrix (no `session_meta`, unparseable first line, empty file → `""`, never panics) is a *parse-level* concern, so `codexCWDFromBytes` is unit-tested **directly on bytes** (ARCH-PURE — no temp files for the pure assertions); `codexSessionCWD` is the one-line IO seam. (Plan-quality finding #2.)

- **`gitx.IsShippedWorkSubject` tests** — the matcher already accepts `#N …` and `<area>: #N …` and rejects loose refs (`f62d099`); this issue adds the missing **table coverage** so the contract is pinned against regression (Done-when 3). No production change expected unless a case fails.

- **`estimate.VelocityPath(brainDir, filename string) string`** — pure, the **single source** for `<brainDir>/data/life/42shots/velocity/<filename>`. Lives in `internal/estimate` (the lower layer `close.go` already imports for `ParseBlock`/`FormatRow`), so BOTH `close.go` (ledger) and the new `estimatesource.go` (calibration doc) call it — one source, no test-enforced mirror. (Plan-quality finding #1: the lower-layer home, not package-main, is the real single source.)
- **`estimate.SourcePath(brainDir, model, override string) string`** — pure resolver: `override` (from `$WF_ESTIMATOR_SRC`) wins; else `VelocityPath(brainDir, model+".md")`.
  - **Future extensions:** when `brainDir` becomes config-driven rather than `../brain`, only `VelocityPath` changes — one edit, both callers.

- **`estimate.SourceGuidance(src SourceStatus) string`** — pure renderer producing the one-block guidance: (a) shared method = the `## Estimate` grammar (`sdlc change-code --help`) + recognized `vocab.Models()`, (b) repo-local calibration = the resolved path + a status verb (`ok` / `MISSING — derive from memory is not allowed, run …` / `stale — ledger newer than calibration`). `SourceStatus` carries `{Path, Model, Exists bool, Stale bool}` computed by the thin IO seam; the wording is table-testable.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `transcripts.DefaultHarnesses` | `cmd/sdlc/internal/transcripts/transcripts.go` | new | `$HOME`, filesystem |
| `claudeHarness.Sources` | `cmd/sdlc/internal/transcripts/claude.go` | new | `os.Stat` over `~/.claude/projects` |
| `codexHarness.Sources` | `cmd/sdlc/internal/transcripts/codex.go` | new | `filepath.WalkDir` over `~/.codex/sessions` |
| `estimateSourceStatus` | `cmd/sdlc/estimatesource.go` | new | `os.Stat`/mtime over brain |
| `NewEstimateSourceCmd` | `cmd/sdlc/estimatesource.go` | new | cobra command (pull) |
| `runStartPlan` estimator line | `cmd/sdlc/startplan.go:48` | modified | stdout |
| `estimateReconRefusal` message | `cmd/sdlc/changecode.go:249` | modified | gate error text |

Detail:

- **`transcripts.DefaultHarnesses()`** — returns `[]Harness{ClaudeHarness(defaultClaudeRoot()), CodexHarness(defaultCodexRoot())}` using `$HOME`. The one place the real roots are wired; tests construct harnesses with temp roots instead.
  - **Injected into:** `actual.go:computeActual` calls `transcripts.Select(cwds, transcripts.DefaultHarnesses())`.

- **`claudeHarness.Sources(cwds)`** — for each cwd, `filepath.Join(root, cwdToClaudeDir(cwd))`; keeps only existing dirs; returns them as `Sources.Dirs`. (Was `selectActualSources`'s Claude loop.)

- **`codexHarness.Sources(cwds)`** — `WalkDir(root)` over date-sharded `YYYY/MM/DD/*.jsonl`; keep files whose `codexSessionCWD` ∈ `cwds`; returns sorted `Sources.Files`. (Was `selectCodexSessionFiles`.) Robustness: a malformed session file is skipped, never aborts the walk.

- **`estimateSourceStatus(brainDir, model, override)`** — the IO seam: resolves `estimate.SourcePath`, `os.Stat`s it (+ the sibling `calibration-ledger.tsv` for the stale compare), returns `estimate.SourceStatus`. Graceful-degradation contract identical to the ledger (`close.go:616-619`): a downstream repo with no brain MUST NOT break `change-code`/`start-plan` — those PUSH surfaces warn-and-continue; only the `sdlc estimate-source` PULL command exits non-zero when the source is missing (that is where "fail loudly" lives without breaking base-layer dependents).

- **`NewEstimateSourceCmd`** — `sdlc estimate-source [--model v2|v2.1|v3] [--brain-dir ../brain]`; prints `estimate.SourceGuidance` and exits non-zero if the calibration source is missing/inaccessible. The pull counterpart to `arch-principles`.

- **`runStartPlan`** — after the estimate nudge (`startplan.go:75`), emit one `estimateSourceLine` naming the resolved calibration source (best-effort; warn-and-continue if absent).

- **`estimateReconRefusal`** — extend the `no ## Estimate block` message (`changecode.go:249`) to name the calibration source: "…derive estimate_hours via the calibrated method — run `sdlc estimate-source` to see the shared grammar + your repo's calibration doc."

---

## Milestone M1 — Measurement robustness (transcript-harness abstraction + fixtures)

**Review boundary.** Closes with `sdlc milestone-close --issue 134 --milestone M1`.

### Task M1.1: Create the `transcripts` package skeleton + `Sources`/`Select`

**Files:**
- Create: `cmd/sdlc/internal/transcripts/transcripts.go`
- Test: `cmd/sdlc/internal/transcripts/transcripts_test.go`

- [ ] **Step 1 — Write the failing test** for the pure aggregator with two fake harnesses:

```go
package transcripts

import ("reflect"; "testing")

type fakeHarness struct{ name string; src Sources }
func (f fakeHarness) Name() string { return f.name }
func (f fakeHarness) Sources(cwds []string) Sources { return f.src }

func TestSelectMergesAndDedups(t *testing.T) {
	hs := []Harness{
		fakeHarness{"a", Sources{Dirs: []string{"/d1"}, Files: []string{"/f1"}}},
		fakeHarness{"b", Sources{Dirs: []string{"/d1", "/d2"}, Files: []string{"/f2"}}},
	}
	got := Select([]string{"/repo"}, hs)
	want := Sources{Dirs: []string{"/d1", "/d2"}, Files: []string{"/f1", "/f2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Select = %+v, want %+v", got, want)
	}
}
```

- [ ] **Step 2 — Run, expect FAIL** (`go test ./cmd/sdlc/internal/transcripts/` → undefined `Sources`/`Select`/`Harness`).
- [ ] **Step 3 — Implement** `Sources`, the `Harness` interface, and `Select` (dedup-preserving-order merge).
- [ ] **Step 4 — Run, expect PASS.**
- [ ] **Step 5 — Commit:** `#134 M1: transcripts.Sources + Select aggregator`.

### Task M1.2: Move the Claude harness (encoder + dir selection)

**Files:**
- Create: `cmd/sdlc/internal/transcripts/claude.go`
- Test: `cmd/sdlc/internal/transcripts/claude_test.go` (move `TestCwdToTranscriptDir` + `TestSelectActualDirs` from `actual_test.go`)

- [ ] **Step 1 — Move the encoder test** `TestCwdToTranscriptDir` (rename target `cwdToClaudeDir`) and the dir-selection test (now `ClaudeHarness(root).Sources(cwds)`), keeping the "unrelated folder excluded / missing folder skipped" assertions.
- [ ] **Step 2 — Run, expect FAIL.**
- [ ] **Step 3 — Implement** `cwdToClaudeDir` (verbatim from `actual.go:58-60`) + `claudeHarness{root}` with `Name()`/`Sources()`; `ClaudeHarness(root) Harness` constructor.
- [ ] **Step 4 — Run, expect PASS.**
- [ ] **Step 5 — Commit:** `#134 M1: Claude transcript harness`.

### Task M1.3: Move the Codex harness + add robustness fixtures

**Files:**
- Create: `cmd/sdlc/internal/transcripts/codex.go`
- Test: `cmd/sdlc/internal/transcripts/codex_test.go`

- [ ] **Step 1 — Write tests** covering the Codex robustness matrix (Done-when 1): date-sharded `YYYY/MM/DD/*.jsonl`; `session_meta.payload.cwd` selection (matching vs unrelated cwd); **malformed first line** then a valid `session_meta`; a session with **no `session_meta`** (→ excluded, not panicking); an **empty file** (→ excluded). Assert returned `Files` are sorted and contain only cwd-matched paths.

```go
func TestCodexHarnessSelectsByCWDAndSurvivesMalformed(t *testing.T) {
	root := t.TempDir()
	day := filepath.Join(root, "2026", "06", "26")
	mustMkdir(t, day)
	write(t, filepath.Join(day, "match.jsonl"),
		`{not json`+"\n"+
		`{"timestamp":"2026-06-26T16:00:00Z","type":"session_meta","payload":{"cwd":"/w/repo"}}`)
	write(t, filepath.Join(day, "nometa.jsonl"),
		`{"timestamp":"2026-06-26T16:00:00Z","type":"response_item","payload":{}}`)
	write(t, filepath.Join(day, "empty.jsonl"), ``)
	write(t, filepath.Join(day, "other.jsonl"),
		`{"timestamp":"2026-06-26T16:00:00Z","type":"session_meta","payload":{"cwd":"/w/other"}}`)
	got := CodexHarness(root).Sources([]string{"/w/repo"})
	want := Sources{Files: []string{filepath.Join(day, "match.jsonl")}}
	if !reflect.DeepEqual(got, want) { t.Fatalf("got %+v want %+v", got, want) }
}
```

- [ ] **Step 2 — Run, expect FAIL.**
- [ ] **Step 1b — Add a direct byte-level test** for the pure `codexCWDFromBytes`: the malformed-first-line, no-`session_meta`, and empty-input cases asserted on `[]byte` literals (no temp files).
- [ ] **Step 3 — Implement** the pure `codexCWDFromBytes([]byte) string`, the thin `codexSessionCWD(path)` seam over it, and `codexHarness{root}.Sources` (move `actual.go:97-147`), constructor `CodexHarness(root) Harness`.
- [ ] **Step 4 — Run, expect PASS.**
- [ ] **Step 5 — Commit:** `#134 M1: Codex transcript harness + robustness fixtures`.

### Task M1.4: Rewire `actual.go` onto the registry; delete the inline selection

**Files:**
- Modify: `cmd/sdlc/actual.go` (drop `transcriptsRoot`/`codexSessionsRoot` vars, `cwdToTranscriptDir`, `actualSources`, `selectActualDirs`, `selectActualSources`, `selectCodexSessionFiles`, `codexSessionCWD`); `computeActual` calls `transcripts.Select(cwds, transcripts.DefaultHarnesses())`.
- Modify: `cmd/sdlc/actual_test.go` (remove the moved tests; keep `TestSelectActualSourcesIncludesMatchingCodexSessions` as an integration test over `transcripts.Select` with temp-rooted real harnesses, or move it to the package).

- [ ] **Step 1 — Update** `computeActual` to build `cwds := nonEmpty(brainAbs, repoTop)` and call `transcripts.Select`; map `Sources` → `activetime.Options{Dirs, Files}`. Telemetry-gap detail string updated to reference the registry roots.
- [ ] **Step 2 — Delete** the now-dead inline functions/vars.
- [ ] **Step 3 — Run** `go test ./cmd/sdlc/...` → expect PASS (the registry tests + the unchanged `statusFromResult`/`windowStart` tests).
- [ ] **Step 4 — Build** `go build ./cmd/sdlc` → expect clean.
- [ ] **Step 5 — Commit:** `#134 M1: actual.go consumes the transcript-harness registry`.

### Task M1.5: Commit-window matcher table tests

**Files:**
- Modify: `cmd/sdlc/internal/gitx/window_test.go`

- [ ] **Step 1 — Write** a table test for `IsShippedWorkSubject("134", subject)` (Done-when 3): accept `#134: subject`, `#134 M2: subject`, `sdlc: #134 measure …`, `<area>: #134 …`; reject `docs: mention #134`, `see #134`, `#1340: …` (number-boundary), and bookkeeping (`#134: claim …`, `#134: close …`). Add the canonical-vs-area pair the issue names explicitly.
- [ ] **Step 2 — Run, expect PASS** (matcher already correct) — or FAIL → fix the matcher (root-cause, not the test).
- [ ] **Step 3 — Commit:** `#134 M1: pin commit-window ownership matcher coverage`.

### Task M1.6: Atlas + helptext — the harness abstraction contract

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md` (the "Measured actuals" para, ~line 167-179): replace "Claude cwd dirs … and Codex session files …" with the **harness-registry** framing (one `Harness` per agent CLI, `transcripts.Select`, adding a harness = one entry).
- Modify: `atlas/index.md` if the new package warrants a pointer.
- Modify: `cmd/sdlc/helptext/actual.md` (the "Selects transcript sources" para): name the registry + the harness-agnostic contract.

- [ ] **Step 1 — Edit** atlas + helptext per above.
- [ ] **Step 2 — Run** `go test ./cmd/sdlc/...` (helptext drift tests) → expect PASS.
- [ ] **Step 3 — Commit:** `#134 M1: document the transcript-harness abstraction`.

### Task M1.7: Close M1

- [ ] Tick M1 plan rows; run `sdlc milestone-close --issue 134 --milestone M1` (auto-dispatches the boundary review). Fix Critical/Important; log the verdict in `## Log`.

---

## Milestone M2 — Estimator-source pointer (shared method + repo-local calibration)

**Review boundary.** Closes with the full-issue `sdlc close`.

### Task M2.1: `estimate.VelocityPath` (single-source the brain path) + DRY `close.go`

Per plan-quality finding #1, the canonical relative-tail builder lives in the
**lower** layer (`internal/estimate`, already imported by `close.go`), so both
callers share one source — no package-main helper, no "assert tails match" test.

**Files:**
- Create: `cmd/sdlc/internal/estimate/source.go` — `VelocityPath(brainDir, filename string) string`.
- Test: `cmd/sdlc/internal/estimate/source_test.go` — `TestVelocityPath`.
- Modify: `cmd/sdlc/close.go:628` to call `estimate.VelocityPath(f.BrainDir, "calibration-ledger.tsv")`.

- [ ] **Step 1 — Write** `TestVelocityPath` pinning the join + an `$WF_CALIB_LEDGER`-untouched note.
- [ ] **Step 2 — Run, expect FAIL.**
- [ ] **Step 3 — Implement** `VelocityPath` + rewire `close.go:628` (the override branch is unchanged).
- [ ] **Step 4 — Run** `go test ./cmd/sdlc/...` → PASS (existing close tests unchanged).
- [ ] **Step 5 — Commit:** `#134 M2: single-source the brain velocity path in internal/estimate`.

### Task M2.2: `estimate.SourcePath` + `SourceGuidance` (pure)

**Files:**
- Modify: `cmd/sdlc/internal/estimate/source.go` (same file as M2.1)
- Test: `cmd/sdlc/internal/estimate/source_test.go`

- [ ] **Step 1 — Write tests:** `SourcePath(brain, model, override)` honours `override` then defaults to `VelocityPath(brain, model+".md")`; `SourceGuidance` renders the shared-method line (grammar pointer + `Models()`) and a status verb for each of `{Exists, Missing, Stale}`. Assert a `Missing` status renders a loud next-action and never an empty string.
- [ ] **Step 2 — Run, expect FAIL.**
- [ ] **Step 3 — Implement** `SourcePath` (delegates to `VelocityPath` — no duplicate tail), `SourceStatus` struct, `SourceGuidance` (pure).
- [ ] **Step 4 — Run, expect PASS.**
- [ ] **Step 5 — Commit:** `#134 M2: estimate source resolver + guidance renderer`.

### Task M2.3: `sdlc estimate-source` pull command

**Files:**
- Create: `cmd/sdlc/estimatesource.go` (`NewEstimateSourceCmd`, `estimateSourceStatus` IO seam).
- Create: `cmd/sdlc/helptext/estimate-source.md`.
- Modify: `cmd/sdlc/main.go` (register the command + helptext).
- Test: `cmd/sdlc/estimatesource_test.go`.

- [ ] **Step 1 — Write tests:** the command flags (`--model`, `--brain-dir`) are registered; `estimateSourceStatus` over a temp brain returns `Exists` when the doc is present and `Missing` (→ non-zero exit) when absent; the stale compare fires when the ledger mtime is newer than the calibration doc.
- [ ] **Step 2 — Run, expect FAIL.**
- [ ] **Step 3 — Implement** the command (prints `SourceGuidance`, exits non-zero on `Missing`), the IO seam (stat doc + sibling ledger via `brainVelocityPath`), the helptext, and registration.
- [ ] **Step 4 — Run** `go test ./cmd/sdlc/...` + `go run ./cmd/sdlc estimate-source --brain-dir ../brain` → shows the real calibration path + `ok`.
- [ ] **Step 5 — Commit:** `#134 M2: sdlc estimate-source pull command`.

### Task M2.4: Push at the planning gates

**Files:**
- Modify: `cmd/sdlc/startplan.go:48-77` — after the estimate nudge, emit `estimateSourceLine(issue)` (best-effort, warn-and-continue).
- Modify: `cmd/sdlc/changecode.go:249` — extend the `no ## Estimate block` message to point at `sdlc estimate-source`.
- Test: extend `startplan_test.go` / `changecode_test.go` for the new wording (pure renderers).

- [ ] **Step 1 — Write/extend tests** for the new start-plan line (present source vs missing) and the change-code message text.
- [ ] **Step 2 — Run, expect FAIL.**
- [ ] **Step 3 — Implement** both pushes (reuse `estimate.SourceGuidance` / a one-line variant).
- [ ] **Step 4 — Run** `go test ./cmd/sdlc/...` → PASS; `go run ./cmd/sdlc start-plan --issue 134` shows the source line.
- [ ] **Step 5 — Commit:** `#134 M2: push estimator-source pointer at start-plan + change-code`.

### Task M2.5: Atlas + helptext — the estimator-source contract

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md` (the "estimate shell" para, ~line 138-153): document the shared-method-vs-repo-local-calibration split + the `estimate-source` surface + `$WF_ESTIMATOR_SRC`.
- Modify: `cmd/sdlc/helptext/estimate.md` — add a "WHERE THE CALIBRATION LIVES" note pointing at `sdlc estimate-source`.

- [ ] **Step 1 — Edit** atlas + helptext.
- [ ] **Step 2 — Run** `go test ./cmd/sdlc/...` → PASS.
- [ ] **Step 3 — Commit:** `#134 M2: document the estimator-source contract`.

### Task M2.6: Verify + close

- [ ] Tick all plan rows. Run `go test ./cmd/sdlc/...` (full) → PASS.
- [ ] Real-harness check (Done-when 2): `sdlc actual --issue 134` measures this issue's own Claude session without hand-passing paths.
- [ ] `sdlc actual --issue 134` (or `sdlc active-time`) first; then `sdlc close --issue 134 --verified '<evidence: tests + estimate-source live run + real actual>'` (let close compute `--actual`).

---

## Notes / scope boundaries

- **Out of scope (future):** single-sourcing the `model:` literal that `internal/judge/prompts.go` hardcodes into the estimate-quality judge prompt (a separate DRY pass); recalibrating the numbers themselves (that is **#127**, which this issue's `estimate-source` surface makes more discoverable but does not perform).
- **ARCH-DRY** drove M2.1 (kill the duplicated brain path) and the harness registry (one selection method, not two parallel paths).
- **ARCH-PURE** drove the `transcripts` split (pure `Select` + pure encoders/parsers behind thin per-harness `Sources` seams) and `estimate.SourcePath`/`SourceGuidance` (pure; IO in `estimatesource.go`).
- **ARCH-PURPOSE** is the spine of M1: the issue's stated purpose is the *abstraction*, not just "Codex happens to parse now" — so the deliverable is a registry a new harness plugs into, and the docs that name that contract.
- **Graceful degradation:** every brain-dependent surface (start-plan/change-code push) warns-and-continues when brain is absent (base-layer propagates downstream); only the explicit `sdlc estimate-source` pull "fails loudly" — reconciling Spec area 5 with the must-not-break-downstream constraint.

## Revisions

- **2026-06-26 (M1 boundary review):** Task M1.5 / the entity-table row named
  `gitx.IsShippedWorkSubject` as the matcher to test, but the test was
  intentionally implemented against **`issueSubjectDescriptor(issue, subject,
  allowClosePrefix=true)`** — the function `CommitWindow` actually calls for
  active-time-window ownership. `IsShippedWorkSubject` is the *#76 ship-probe*,
  which layers a bookkeeping denylist the window deliberately does not apply (the
  window counts claim/close commits, since they carry real minutes). Testing the
  ship-probe would have pinned the wrong contract. A correctness improvement over
  the plan's wording, not a gap (M1 review: SHIP).
- **2026-06-26 (post-M1):** added a `nonEmpty` glue table test (M1 review advisory
  #5) so `computeActual`'s brain+repo cwd assembly is unit-covered, not only
  exercised by the dogfood run.
- **2026-06-26 (M2 close review, FIX-THEN-SHIP):** (a) the Important finding — the
  start-plan `SourceLine` push was wired but unasserted — fixed by adding
  `"estimate-source"` to `TestRunStartPlan_RendersAtPlanLens`'s `want` set (the
  lessons.md #72 un-wiring guard). (b) `reportEstimateSource`'s stdout print is now
  asserted, not just its error. (c) File-home reconciliation vs the Core-concepts
  tables: `DefaultHarnesses` ships in `internal/transcripts/defaults.go` (table said
  `transcripts.go`); the start-plan push is the lower-layer pure
  `estimate.SourceLine` in `internal/estimate/source.go` (table named a
  package-main `estimateSourceLine` — `startplan.go` only *wires* it). Both are
  improvements (DefaultHarnesses isolation; the pure renderer is reused + tested);
  only the file homes drifted from the table.
