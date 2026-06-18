# Estimate-shell (reconcile · judge · close-the-loop) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give `estimate_hours` a deterministic shell — a parseable `## Estimate` block that `sdlc change-code` reconciles, an estimate-quality LLM judge, and an auto-calibration ledger written at `sdlc close` that scores every estimate against its measured actual.

**Architecture:** A new **pure** package `cmd/sdlc/internal/estimate` holds all logic (grammar parse, reconciliation math, vocabulary, ledger-row formatting) — unit-tested with zero IO. Three thin IO seams inject it: a reconciliation guard in `change-code`, an estimate-quality judge mirroring the existing plan-quality judge, and a ledger append in `close`. The shell is **model-agnostic**: it enforces that whatever model the `model:` provenance line names was actually applied (today `estimate-logic-v2`). Each ledger row carries a `window-trusted` flag so it can ship before #116 without logging garbage.

**Tech Stack:** Go (cobra CLI, table-driven tests). Reuses `internal/issue` (frontmatter parse/section regex), `internal/judge` (fresh-context LLM dispatch), `internal/activetime` (the actual measurement), `helptext` (embedded Long descriptions).

---

## Core concepts

### The `## Estimate` block grammar (the new contract)

A fenced ` ```estimate ` block inside the issue's `## Estimate` section:

```
## Estimate

​```estimate
model: estimate-logic-v2
familiarity: 1.0
item: greenfield-go-module   design=0.3 impl=0.6
item: smaller-go-module      design=0.2 impl=0.6
item: smaller-go-module      design=0.2 impl=0.5
item: atlas-docs             design=0.0 impl=0.2
item: milestone-review       design=0.0 impl=0.6
design-buffer: 0.30
total: 3.4
​```
```

**Reconciliation rule (deterministic, pure):**
`recomputed = Σ(item.design) × (1 + design-buffer) + Σ(item.impl) × familiarity`
(familiarity multiplies impl, buffer lifts design — matches estimate-logic-v2 Steps 5–6.)
Pass iff `|recomputed − total| ≤ tol` **and** `|total − frontmatter.estimate_hours| ≤ tol`, where `tol = max(0.05, 0.05 × total)` (absorbs 1-dp rounding).

**Field rules:** `model:` ∈ recognized set; `familiarity:` float (default 1.0); `design-buffer:` float (default 0.30); ≥1 `item:`; each item slug ∈ closed v2 vocabulary; `total:` and `estimate_hours` present and reconcile. Each violation yields a precise next-action message (the error *is* the spec).

**Closed v2 vocabulary** (slugs, canonical in `vocab.go`, documented in `helptext/estimate.md`, mirrors the estimate-logic-v2.md primitive table): `pensive`, `issue-spec`, `typed-data-prototype`, `skill-or-dispatcher`, `smaller-go-module`, `greenfield-go-module`, `api-integration`, `greenfield-service`, `tui-screen`, `cross-cutting-refactor`, `cross-repo-refactor-small`, `cross-repo-refactor-large`, `atlas-docs`, `lua-neovim`, `milestone-review`, `real-api-discovery`, `scope-pivot`, `ux-rename-iteration`, `method-b-decisions`.

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Block` (parsed `## Estimate`: model, familiarity, items, buffer, total) | `cmd/sdlc/internal/estimate/block.go` | new |
| `Item` (slug, design, impl) | `cmd/sdlc/internal/estimate/block.go` | new |
| `Vocabulary` (recognized primitive slugs + model versions) | `cmd/sdlc/internal/estimate/vocab.go` | new |
| `ParseBlock(section string) (Block, error)` | `cmd/sdlc/internal/estimate/parse.go` | new |
| `Check(b Block, estimateHours float64) []Failure` (reconcile + vocab + provenance) | `cmd/sdlc/internal/estimate/check.go` | new |
| `Failure` (one reconciliation violation + next-action message) | `cmd/sdlc/internal/estimate/check.go` | new |
| `LedgerRow` (issue, est, est-design, est-impl, actual, ratio, model, mode, window-trusted, date) | `cmd/sdlc/internal/estimate/ledger.go` | new |
| `FormatRow(LedgerRow) string` (one TSV/markdown line) | `cmd/sdlc/internal/estimate/ledger.go` | new |
| `DriftVerdict(rows []LedgerRow, n int) (warn bool, msg string)` (>2× same-direction over last N trusted) | `cmd/sdlc/internal/estimate/drift.go` | new |

- **Block / Item** — the parsed estimate. PURE: `ParseBlock` is a string→struct function; no IO.
  - **DRY rationale:** one grammar definition consumed by both the change-code guard and (later) the `sdlc estimate` engine; first occurrence of a pattern the engine will reuse.
  - **Future extensions:** add `cross-repo-buffer:` (+20%, v2 Step 6.2); a `range:` low/high pair; #112's attention touchpoints become new vocab slugs + a new `model:` value — the same parser carries them.
- **Check / Failure** — pure reconciliation. Mirrors `issue.CheckStructural` / `issue.CheckEstimate` shape (`[]Failure`, each with a message), so change-code composes it identically.
  - **DRY rationale:** reuses the existing structural-failure refusal pattern rather than inventing a new error channel.
- **LedgerRow / FormatRow / DriftVerdict** — pure data + formatting + the drift rule. Tested without touching the filesystem; the append IO is the thin seam.
  - **Future extensions:** `DriftVerdict` thresholds become tunable; `mode` (supervised|delegated) feeds a per-mode calibration once #112's supervision variable matters.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| reconciliation guard | `cmd/sdlc/changecode.go` (~after :144) | modified | issue-file read |
| `judge.EstimateQuality` category + prompt | `cmd/sdlc/internal/judge/prompts.go` | modified | — |
| estimate-quality judge dispatch | `cmd/sdlc/changecode.go` (in `!NoJudge` block, ~:147) | modified | LLM subprocess |
| calibration-ledger append | `cmd/sdlc/close.go` (~:574, after actual known) | modified | ledger file write |
| ledger file | `brain/data/life/42shots/velocity/calibration-ledger.tsv` | new | filesystem |

- **reconciliation guard** — reads the issue body, extracts the `## Estimate` section (reuse a `plan.go`-style `EstimateSectionRE`), calls `estimate.ParseBlock` + `estimate.Check`; on failure refuses like the structural gate. Gated by a new `--no-estimate-recon` flag (and bypassed by `--force`).
  - **Injected into:** nothing — it injects the pure `estimate` package into change-code. Pure logic stays in `estimate`; this is the thin caller.
- **estimate-quality judge** — second judge in the `!NoJudge` block, mirroring `runPlanQualityJudge` (`changecode.go:267`). New `judge.EstimateQuality` category + `estimateQualityPrompt`; reuses `judge.BuildPrompt`/`Dispatch`/`Classify`. Reads spec + `## Estimate` body + the model doc path; verdict trailer.
  - **Future extensions:** when #112 lands, the prompt's model-doc reference swaps; the harness is unchanged.
- **calibration-ledger append** — after the actual is known at close, compose a `LedgerRow` (estimate from frontmatter, actual from `--actual`/`computeActual`, `window-trusted: no` until #116, `mode` from a close flag or inferred) and append a line to the ledger; run `DriftVerdict` and print a warning on drift.
  - **Injected into:** the pure `FormatRow`/`DriftVerdict`; this seam only does file append + stderr print.
  - **Test surface:** a temp-dir fake ledger path (env override `WF_CALIB_LEDGER`) so the close test writes to a tmpfile, not the real brain path.

---

## Milestones (each an `Mx` review boundary, closed separately)

- M1 — pure `internal/estimate` package (grammar, parse, check, vocab, ledger-row, drift). No wiring.
- M2 — change-code enforcement: reconciliation guard + estimate-quality judge + helptext (change-code/issue/estimate).
- M3 — close-the-loop: ledger append + drift warning at `close`, backfill past points, helptext/close + atlas.

---

## Chunk 1: M1 — pure estimate package

### Task 1: `Block`/`Item` types + `Vocabulary`

**Files:**
- Create: `cmd/sdlc/internal/estimate/block.go`, `cmd/sdlc/internal/estimate/vocab.go`
- Test: `cmd/sdlc/internal/estimate/vocab_test.go`

- [ ] **Step 1: Write the failing test** — `vocab_test.go`: assert `KnownPrimitive("smaller-go-module")` true, `KnownPrimitive("made-up")` false, `KnownModel("estimate-logic-v2")` true, `KnownModel("vibes")` false.
- [ ] **Step 2: Run** `go test ./cmd/sdlc/internal/estimate/ -run Vocab` → FAIL (package/func missing).
- [ ] **Step 3: Implement** `block.go` (`type Item struct{ Slug string; Design, Impl float64 }`, `type Block struct{ Model string; Familiarity, DesignBuffer, Total float64; Items []Item }`) and `vocab.go` (`var primitives = map[string]bool{…}` from the closed vocab list; `var models = map[string]bool{"estimate-logic-v2":true, "estimate-logic-v2.1":true, "estimate-logic-v3":true}`; `KnownPrimitive`, `KnownModel`).
- [ ] **Step 4: Run** → PASS.
- [ ] **Step 5: Commit** `#117 M1: estimate Block/Item types + vocabulary`.

### Task 2: `ParseBlock`

**Files:**
- Create: `cmd/sdlc/internal/estimate/parse.go`
- Test: `cmd/sdlc/internal/estimate/parse_test.go`

- [ ] **Step 1: Write failing tests** — table-driven: (a) the **canonical green fixture** from Core concepts (5 items, familiarity 1.0, buffer 0.30, total 3.4) parses to the expected `Block` — reuse it verbatim so example/dogfood/test agree; (b) missing fenced block → error; (c) item line with non-numeric design → error; (d) defaults: absent `familiarity:` → 1.0, absent `design-buffer:` → 0.30.
- [ ] **Step 2: Run** → FAIL.
- [ ] **Step 3: Implement** `ParseBlock(section string) (Block, error)`: locate the ` ```estimate … ``` ` fence (regexp `(?ms)^\x60\x60\x60estimate\s*\n(.*?)^\x60\x60\x60`), parse `key: value` and `item: <slug> design=<f> impl=<f>` lines (split fields, `strconv.ParseFloat`). Return parse errors with line context. Pure.
- [ ] **Step 4: Run** → PASS.
- [ ] **Step 5: Commit** `#117 M1: ParseBlock estimate grammar`.

### Task 3: `Check` (reconciliation + vocab + provenance)

**Files:**
- Create: `cmd/sdlc/internal/estimate/check.go`
- Test: `cmd/sdlc/internal/estimate/check_test.go`

- [ ] **Step 1: Write failing tests** — `Check(b, estimateHours)` returns no failures for the reconciling example; returns a failure when (a) `total` ≠ recomputed; (b) `estimate_hours` ≠ `total`; (c) an item slug is unknown; (d) `model` unknown; (e) zero items. Assert each `Failure.Message` names the fix (e.g. `"## Estimate total: 3.3 ≠ recomputed 2.8 (Σdesign×1.30 + Σimpl×fam); fix the items or the total"`).
- [ ] **Step 2: Run** → FAIL.
- [ ] **Step 3: Implement** `Check`: `recomputed := sumDesign*(1+b.DesignBuffer) + sumImpl*b.Familiarity`; `tol := math.Max(0.05, 0.05*b.Total)`; append a `Failure` per violated rule. Pure.
- [ ] **Step 4: Run** → PASS.
- [ ] **Step 5: Commit** `#117 M1: Check reconciliation + vocab + provenance`.

### Task 4: `LedgerRow` / `FormatRow` / `DriftVerdict`

**Files:**
- Create: `cmd/sdlc/internal/estimate/ledger.go`, `cmd/sdlc/internal/estimate/drift.go`
- Test: `cmd/sdlc/internal/estimate/ledger_test.go`, `drift_test.go`

- [ ] **Step 1: Write failing tests** — `FormatRow` produces a stable TSV line `issue\test\actual\…\window-trusted` with fixed column order; `ParseRows(text)` round-trips. `DriftVerdict` over crafted rows: 3 trusted rows all >2× over-estimate → `warn=true`; mixed/within-2× → `warn=false`; untrusted rows excluded from the count.
- [ ] **Step 2: Run** → FAIL.
- [ ] **Step 3: Implement** `LedgerRow` struct, `FormatRow`, `ParseRows`, and `DriftVerdict(rows, n)`. All pure.
- [ ] **Step 4: Run** `go test ./cmd/sdlc/internal/estimate/...` → PASS; `go vet ./cmd/sdlc/internal/estimate/...` clean.
- [ ] **Step 5: Commit** `#117 M1: LedgerRow format + DriftVerdict`.

- [ ] **M1 close:** `sdlc milestone-close --issue 117 --milestone M1` (auto-dispatches the fresh-context review M1..base). Fix Critical/Important; log the verdict.

---

## Chunk 2: M2 — change-code enforcement

### Task 5: `## Estimate` section extractor (reuse the plan.go pattern)

**Files:**
- Modify: `cmd/sdlc/internal/issue/plan.go` (or a sibling `estimate.go` in `internal/issue`)
- Test: `cmd/sdlc/internal/issue/…_test.go`

- [ ] Add `EstimateSectionRE = regexp.MustCompile(`(?ms)^## Estimate\s*\n(.*?)(?:^## |\z)`)` and `EstimateSection(body string) (string, bool)`, mirroring `PlanSectionRE`. TDD: present → body; absent → `false`. Commit `#117 M2: ## Estimate section extractor`.

### Task 6: reconciliation guard in change-code

**Files:**
- Modify: `cmd/sdlc/changecode.go` (insert after the estimate gate ~:144; add `--no-estimate-recon` flag ~:51)
- Test: `cmd/sdlc/changecode_test.go`

- [ ] **Step 1: Failing test** — `runChangeCode` on an issue whose `estimate_hours` doesn't reconcile with its `## Estimate` block → refusal naming the mismatch; a reconciling issue → passes the guard; `--no-estimate-recon` → skips; `--force <reason>` → bypass logged.
- [ ] **Step 2:** Run → FAIL.
- [ ] **Step 3: Implement** `estimateReconRefusal(issueContent string, noRecon bool) *issue.StructuralFailure`: `EstimateSection` → `estimate.ParseBlock` → `estimate.Check(block, estimateHoursFromFM)`; first failure → `*StructuralFailure`. Wire it gated by `!f.NoEstimateRecon` between the estimate gate and the judge, following the `estimateRefusal` pattern at `:135/:201`.
- [ ] **Step 4:** Run → PASS.
- [ ] **Step 5:** Commit `#117 M2: change-code reconciliation guard`.

### Task 7: estimate-quality judge

**Files:**
- Modify: `cmd/sdlc/internal/judge/prompts.go` (add `EstimateQuality` category + `estimateQualityPrompt`), `cmd/sdlc/changecode.go` (second judge in `!NoJudge` block)
- Test: `cmd/sdlc/internal/judge/prompts_test.go`, `changecode_test.go`

- [ ] **Step 1: Failing test** — `judge.BuildPrompt(judge.EstimateQuality, input)` includes the `## Estimate` body + spec + a pointer to `estimate-logic-v2.md`, and asks "was the model applied / are per-primitive hours plausible / does it ignore delegation+fragmentation"; `Classify` maps the verdict. A `changecode_test` with a fake judge (existing test seam) asserts the estimate judge runs in the `!NoJudge` path and is skipped under `--no-judge`.
- [ ] **Step 2:** Run → FAIL.
- [ ] **Step 3: Implement** the category, prompt builder, and a `runEstimateQualityJudge` mirroring `runPlanQualityJudge` (`:267`), dispatched right after it inside `!f.NoJudge`. **Registration touchpoints** (all derive from one switch): add `EstimateQuality` to the `BuildPrompt` switch (`prompts.go:118`) and `Label()` (`:59`), but **do NOT add it to `AllCategories()` (`:44`)** — that list enrolls categories in push/merge **bulk dispatch**, and this is a change-code-time-only gate. Mirror `TestPlanQuality_RegisteredInCategories` with an inverse test asserting `EstimateQuality ∉ AllCategories()` so it can never silently become merge-time.
- [ ] **Step 4:** Run → PASS.
- [ ] **Step 5:** Commit `#117 M2: estimate-quality judge`.

### Task 8: helptext (change-code / issue / new estimate.md)

**Files:**
- Modify: `cmd/sdlc/helptext/change-code.md`, `cmd/sdlc/helptext/issue.md`
- Create: `cmd/sdlc/helptext/estimate.md` (the `## Estimate` grammar + vocab + reconciliation contract)

- [ ] Document the `## Estimate` block + reconciliation in `change-code.md` (new gate), the contract in `issue.md` (body sections), and the full grammar/vocab in `estimate.md`. TDD: `helptext/embed_test.go` already asserts every `.md` is loadable — add `estimate` to any coverage list. Commit `#117 M2: helptext for the estimate contract`.
- [ ] **Vocab drift guard (ARCH-DRY).** Add a test in `internal/estimate` (or `helptext`) asserting `helptext/estimate.md` mentions **exactly** the slug set in `vocab.go` (every canonical slug appears; no stray slug documented) — so the doc mirror can't silently drift from the `vocab.go` source of truth. Commit `#117 M2: vocab/helptext drift-guard test`.

- [ ] **M2 close:** `sdlc milestone-close --issue 117 --milestone M2`. Fix Critical/Important; log verdict.

---

## Chunk 3: M3 — close-the-loop ledger

### Task 9: ledger append + drift warning at close

**Files:**
- Modify: `cmd/sdlc/close.go` (after actual known, ~:519/:574); add `WF_CALIB_LEDGER` env override + optional `--mode supervised|delegated` flag
- Test: `cmd/sdlc/close_test.go`

- [ ] **Step 1: Failing test** — closing an issue with `WF_CALIB_LEDGER` pointed at a tmpfile appends one `FormatRow` line with `window-trusted: no` (no `started:` yet → see #116), the estimate from frontmatter, the actual from `--actual`; a crafted ledger with 3 trusted >2× rows makes close print a drift warning. **Plus a brain-absent test** (finding #1): with no `WF_CALIB_LEDGER` set and the resolved `brain/...` dir absent, `sdlc close` **succeeds** (skips the append) and prints a one-line `[!] calibration ledger skipped (no brain/ ledger dir)` warning — it must NOT fail the close. Use the existing close-test harness (it already fakes the actual).
- [ ] **Step 2:** Run → FAIL.
- [ ] **Step 3: Implement** the append: resolve ledger path (`WF_CALIB_LEDGER` else `brain/data/life/42shots/velocity/calibration-ledger.tsv`). **Graceful degradation (Important):** `sdlc` is base-layer and propagates to downstream repos that may have no sibling `brain/` — if the override is unset **and** the resolved ledger dir doesn't exist, skip-with-warning and return nil (close must never break on a missing ledger). Otherwise build `LedgerRow` (window-trusted = `false` until #116 stamps `started:`; detect via frontmatter `started:` presence so it auto-flips once #116 lands), `estimate.FormatRow`, append; read back rows + `estimate.DriftVerdict` → stderr warning. Pure logic stays in `estimate`; close only does IO.
- [ ] **Step 4:** Run → PASS.
- [ ] **Step 5:** Commit `#117 M3: calibration-ledger append + drift at close`.

### Task 10: backfill past points + ledger doc

**Files:**
- Create: `brain/data/life/42shots/velocity/calibration-ledger.tsv` (header + backfilled rows)
- Modify: `brain/data/life/42shots/velocity/estimate-logic-v2.md` (point the "Validation log" at the ledger), `brain/data/life/42shots/velocity/SKILL.md` (pointer)

- [ ] Backfill `window-trusted: no` historical rows from issues we have both sides for: #110 (est 5 / act 0.89), #111 (est 7 / act 0.35), charon #13 (est range / act 4.5). Mark them clearly as gut-estimate (no `## Estimate` block) vs by-the-book where known. Update the v2 doc's hand-kept validation table to defer to the ledger. Commit `#117 M3: backfill calibration ledger + repoint v2 validation log`.

### Task 11: helptext/close + atlas

**Files:**
- Modify: `cmd/sdlc/helptext/close.md`; `atlas/` (estimate-shell surface + the form/essence/feedback split; `atlas/index.md` link)

- [ ] Document the close-time ledger behavior in `close.md`; add an atlas note describing the estimate shell (reconcile/judge/ledger) and its relation to active-time-v3. Commit `#117 M3: helptext/close + atlas for the estimate shell`.

- [ ] **M3 / issue close:** `sdlc close --issue 117 --milestone M3 --verified '<evidence>'` (final milestone → close auto-dispatches the end-of-issue review over M1+M2+M3). `--actual` measured, not typed.

---

## Verification

- `go build ./... && go test ./... && go vet ./...` green across `cmd/sdlc`.
- Live: write a deliberately non-reconciling `## Estimate` block on a scratch issue → `sdlc change-code` refuses with the precise message; fix it → passes.
- Live: close a scratch issue with `WF_CALIB_LEDGER=$(mktemp)` → one row appended, `window-trusted: no`; inspect the file.
- The estimate-quality judge runs on a fabricated block and flags it (fresh-context).

## Notes / dependencies

- **#116** stamps `started:`; Task 9 reads its presence to set `window-trusted`, so rows auto-upgrade once #116 lands — no rework here.
- **#112** parked: if its attention model is ever adopted, it adds vocab slugs + a `model:` value; the grammar/guard/judge/ledger are unchanged (model-agnostic by design).
- Vocabulary lives canonically in `vocab.go` (DRY); `helptext/estimate.md` documents it; the brain `estimate-logic-v2.md` table is the human narrative it mirrors — keep them reconciled if the primitive set changes.
