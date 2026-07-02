# process-manual — dynamic session reconstruction (#157) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) for the execution approach (superpowers-subagent-driven-development or superpowers-executing-plans). Steps use checkbox (`- [x]`) syntax.

**Goal:** `sdlc process-manual --session <jsonl|current>` parses a Claude session transcript and emits, in timestamp order (segmented on the 60-min-gap / `away_summary` boundary), each **fired** injection event matched to its M1 catalog `Kind` — as a linked markdown report.

**Architecture:** New `session.go` in `cmd/sdlc/internal/processmanual` (co-located with M1's catalog so it matches against `InjectionSource` in-process — ARCH-DRY, no serialization boundary). Pure core: a tolerant JSONL parser → ordered `FiredEvent`s → segmentation → catalog match → `renderSessionReport`. Thin IO shell: locate the JSONL, read it. The verb gains a `--session` flag that branches `runProcessManual` to the reconstruction path; the static catalog path is unchanged.

**Tech Stack:** Go, `encoding/json`, existing `internal/processmanual` (M1 catalog + `renderManual` helpers) + `internal/judge` (`ParseVerdict`).

**Scope:** The **feasible core** only. Anomaly / "injected-but-ignored" detection is explicitly **deferred** (grounded as fuzzy — see the issue). Two hard limits are documented, not fought: (1) `agents-chain`/`memory` are session-start system-prompt injections that never appear in the transcript; (2) forked review *prompts* aren't in the main JSONL (only their *output*, via Bash stdout).

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `FiredEvent` | `cmd/sdlc/internal/processmanual/session.go` | new |
| `parseEvents` | `session.go` | new |
| `classifyToolUse` | `session.go` | new |
| `segmentEvents` | `session.go` | new |
| `renderSessionReport` | `session.go` | new |

- **`FiredEvent`** — one injection that fired: `{Time time.Time, Kind Kind, Tool string ("Bash"/"Skill"/"Read"), Detail string (verb / skill name / file basename), Verdict string (optional — for close/milestone-close), Link string (resolved from the M1 catalog)}`.
  - **DRY rationale:** `Kind` + `Link` reuse M1's vocabulary, so the same catalog resolves both the static manual and "what fired." (The issue notes `InjectionSource` was shaped for this; `FiredEvent` is the *dynamic* record that references the catalog rather than mutating the static struct.)

- **`parseEvents(data []byte) (events []FiredEvent, allTimes []time.Time, awaySummaryTimes []time.Time, err error)`** — pure over bytes. Scans JSONL lines; `json.Unmarshal` each into a **tolerant** struct — the full field set (verified against live data by the plan review):
  ```go
  type rec struct {
    Type      string    // "assistant" | "user" | "system" | … (skip unknown)
    Subtype   string    // for system records: "away_summary" is a segment boundary
    Timestamp time.Time
    Message   struct{ Content []struct {
      Type       string          // "tool_use" | "tool_result"
      Name, ID   string          // tool_use: name + its own id
      Input      json.RawMessage // tool_use input
      ToolUseID  string `json:"tool_use_id"` // tool_result: links back to the tool_use id
    } }
    ToolUseResult json.RawMessage // polymorphic: dict {stdout,…} | string | null
  }
  ```
  - **skips unknown `type`s** (newer sessions add bookkeeping types — never enumerate).
  - For `assistant` records, walk `tool_use` blocks → `classifyToolUse`; keep the fired ones.
  - Build the `tool_use_id → stdout` map from the **`user` records' `.message.content[].tool_use_id`** (NOT the tool_use's own `id`; the top-level `.toolUseResult` carries no id) so `close`/`milestone-close` events recover their `Verdict` via **`judge.ParseVerdict`** (ARCH-DRY — the exact fn `close` uses at `close.go:513`). Note `judge.ParseVerdict` parses the **reviewer's output** (a `VERDICT: SHIP` line / bare `SHIP` / fenced ```` ```verdict ```` block), NOT the `Review-Verdict:` git-trailer — and a real close stdout streams the full reviewer body *then* the trailer, so it resolves correctly.
  - `ToolUseResult` is polymorphic (dict/string/null); the stdout extraction must **swallow unmarshal errors** on the string/null cases (stay truly tolerant).
  - `allTimes` = every parsed record's timestamp (gap-based segmentation, so non-injection work between fired events doesn't cause false splits); `awaySummaryTimes` = timestamps of `type=="system" && subtype=="away_summary"` records (fed to `segmentEvents`).

- **`classifyToolUse(name string, input json.RawMessage) (Kind, detail string, ok bool)`** — pure match table (ports the *idea* of `construct/local/introspect/scripts/segment_text.py:88` `summarize_tool_input`, not code): `Skill` → `KindSkill` (`.skill`); `Bash` with `sdlc <verb>` → `KindHelpText` if `--help`, else `KindSDLCPrompt` (verb captured); `Read` of `lessons.md` → `KindLessons`; else `ok=false`. Regex anchored so `sdlc` inside a word doesn't match.

- **`segmentEvents(events []FiredEvent, allTimes []time.Time, awaySummaryTimes []time.Time) [][]FiredEvent`** — pure. Boundaries = an `away_summary` timestamp OR a `> 60min` gap between consecutive `allTimes` (constant ported from `construct/local/introspect/scripts/normalize.py`, `GAP_BOUNDARY_SECONDS = 60*60`). Buckets fired events into segments. First occurrence of a pattern likely to recur → a named `const gapBoundary = 60 * time.Minute`.

- **`renderSessionReport(segments [][]FiredEvent, catalog []InjectionSource, linkPrefix string) string`** — pure markdown: a header stating the two hard limits, then one `## Segment N` per segment with a chronological list — `HH:MM:SS · Kind · detail` linked to the matched catalog source, verdict inline for close/milestone-close. Reuses M1's link-prefix + fencing helpers where shared (ARCH-DRY).
  - **Link resolution (plan-review):** matching `Kind`+`Detail` against `catalog` resolves `Skill` (catalog Title = skill dir name) and `lessons` (single entry). But `KindSDLCPrompt` catalog entries are titled by *judge category* (`milestone-review`, `plan-quality`, …), **never by verb** — so a fired `sdlc close` (`Detail="close"`) resolves no exact entry. Fallback: link the verb event to `cmd/sdlc/helptext/<verb>.md` if that help-text entry exists, else render unlinked. Done-when only requires mapping to a `Kind` (satisfied); the fallback is a quality touch — **test the unresolved-verb case** so it's not silently unlinked.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `locateSessionJSONL` | `session.go` | new | `~/.claude/projects/<slug>` FS |
| `runProcessManual` (`--session` branch) | `processmanual.go` | modified | cobra / stdout |

- **`locateSessionJSONL(homeDir, absRepoRoot, sessionArg string) (string, error)`** — if `sessionArg` is a path, use it; if `"current"`, resolve to `<slug>/<$CLAUDE_CODE_SESSION_ID>.jsonl` when the env var is set (the **authoritative** signal — confirmed present in the runtime; the session file is `<sessionId>.jsonl`), else fall back to **newest `*.jsonl` by mtime** in `homeDir/.claude/projects/claudeProjectSlug(absRepoRoot)/` (running the Bash call appends to the current session → newest; guessy only under concurrent same-repo sessions). Reuse M1's `claudeProjectSlug` (ARCH-DRY). IO seam → tested with a temp dir.
- **`runProcessManual` `--session` branch** — when `--session` is set, `locateSessionJSONL` → read → `parseEvents` → `segmentEvents` → `renderSessionReport(…, Collect(opts)…)`; else the existing static-catalog path. `--out` reuses the existing write logic.

**Architecture principle citations:**
- **ARCH-PURE** — parser/classifier/segmenter/renderer are pure over bytes/slices, unit-tested against fixture JSONL with no IO; only `locateSessionJSONL` + the file read touch IO.
- **ARCH-DRY** — reuse M1's `InjectionSource` catalog + `claudeProjectSlug` + link/fence helpers in-process (the reason to stay Go-native, not shell to Python); reuse `judge.ParseVerdict` for verdicts; port introspect's segmentation *constant/algorithm*, not its process.
- **ARCH-PURPOSE** — deliver the ordered fired-event stream matched to the catalog (the whole point), and **state the two hard limits in the output** rather than silently omitting them; anomaly detection is a separate issue, not a half-built subset here.

---

## Chunk 1: core session reconstruction

### Task 1: `FiredEvent` + tolerant `parseEvents` (+ verdict recovery)
**Files:** Create `session.go`, `session_test.go`.
- [x] **Step 1: failing test** — a fixture JSONL string with: an `assistant` record whose `tool_use` is a Bash `sdlc close --issue 9` (id `toolu_X`); a following `user` record whose `.message.content[]` has a `tool_result` with `tool_use_id: toolu_X` and a `.toolUseResult` (dict) whose `stdout` contains a **reviewer-output** verdict line — `VERDICT: SHIP` (NOT just the `Review-Verdict: SHIP` trailer, which `ParseVerdict` returns `"unknown"` for — verified); a `Skill` tool_use; a `Read` of `…/workshop/lessons.md`; a `type:"system", subtype:"away_summary"` record; and an unknown `type` line. Assert `parseEvents` returns 3 `FiredEvent`s (close/skill/lessons), the close one has `Verdict=="SHIP"` (linked via `tool_use_id`), `awaySummaryTimes` has 1 entry, and the unknown line is skipped without error.
- [x] **Step 2:** run → FAIL (undefined).
- [x] **Step 3:** implement `FiredEvent`, the tolerant unmarshal structs, the tool_use walk, the `tool_use_id→stdout` map, and verdict recovery via `judge.ParseVerdict`.
- [x] **Step 4:** run → PASS.
- [x] **Step 5:** commit — `#157: FiredEvent + tolerant parseEvents (verdict via judge.ParseVerdict)`

### Task 2: `classifyToolUse` match table
**Files:** modify `session.go`, `session_test.go`.
- [x] Table test: `Skill{skill:x}`→KindSkill; `Bash{command:"sdlc close …"}`→KindSDLCPrompt(detail "close"); `Bash{command:"sdlc state --help"}`→KindHelpText; `Read{file_path:".../lessons.md"}`→KindLessons; `Bash{command:"ls"}`→ok=false; `Bash{command:"echo sdlcx"}`→ok=false (anchored). Red → implement → green → commit.

### Task 3: `segmentEvents` (60-min gap + away_summary)
**Files:** modify `session.go`, `session_test.go`.
- [x] Test: events at t, t+10min, t+90min (gap) → 2 segments; an `away_summary` timestamp between two events → split there too. Red → implement (`const gapBoundary = 60*time.Minute`) → green → commit.

### Task 4: `renderSessionReport`
**Files:** modify `session.go`, `session_test.go`.
- [x] Test: 2 segments of fired events + the M1 catalog → markdown with `## Segment 1/2`, chronological `HH:MM:SS · Kind · detail` lines, a matched link for a known Kind, the close event's `SHIP` verdict inline, and the header naming the two hard limits (agents-chain/memory invisible; forked prompts). Red → implement (resolve links against catalog; reuse link-prefix) → green → commit.

### Task 5: `locateSessionJSONL` + `--session` wiring
**Files:** modify `session.go`, `processmanual.go`, `processmanual_test.go`, `session_test.go`.
- [x] Test A (`locateSessionJSONL`): temp `home/.claude/projects/<slug>/` with two `*.jsonl` (different mtimes) → `"current"` returns the newest; an explicit path returns it as-is.
- [x] Test B (cobra): write a fixture JSONL to a temp file; `buildRoot()` with `["process-manual","--session",<path>]` → output contains `## Segment` and a fired event. Red → implement the `--session` flag + `runProcessManual` branch → green → commit.

### Task 6: atlas + real smoke run
**Files:** `atlas/workflow/process-manual.md` (note the dynamic pass now exists, #157), regenerate nothing (session report is not tracked).
- [x] Update the atlas doc: the dynamic pass is delivered (`--session`); keep the two hard limits documented.
- [x] Real smoke: `go run ./cmd/sdlc process-manual --session current | head -40` on a real recent session; sanity-check the ordered stream. Record in the issue `## Log`.
- [x] Commit.

---

## Estimate

The canonical reconciled derivation is the issue's `## Estimate` block (v3.1,
Method A) — **2.6h**: one `greenfield-go-module` (`session.go`: parser +
classifier + segmenter + renderer, all pure + fixture-tested) + a
`smaller-go-module` (`--session` wiring + `locateSessionJSONL`) + `atlas-docs` +
the single close-review (`milestone-review`). Deterministic, fixture-driven,
reuses M1. The estimate-quality gate reads the issue block, so this plan does not
carry a second (divergent) number.

## Explicitly deferred (a later issue, not #157)

Anomaly / "injected-but-ignored" detection: `agents-chain`/`memory` firing is
undetectable (no transcript event); "offered-but-never-fired" (skill_listing − Skill
tool_use) is a weak signal; "did the agent follow the guidance?" is an LLM-judge problem.
Out of scope by design.

## Revisions

- **2026-07-01 — `FiredEvent.Link` removed (implementation).** The Core-concepts
  sketch listed `FiredEvent{… Link string (resolved from the M1 catalog)}`, but links
  resolve at *render* time in `renderSessionReport` (which takes the catalog), so a
  `Link` field on the parse output would be dead. Delta: `FiredEvent` carries only the
  fired data (`Time`/`Kind`/`Detail`/`Verdict`); presentation resolves links. (The
  `Tool` field was likewise dropped at the boundary review — redundant with `Kind`,
  never rendered.) Cleaner pure-core / render split; done-when unaffected.
- **2026-07-01 — verdict recovery: trailer fallback (boundary-review I1).** Task 1's
  `parseEvents` note assumed a close stdout "streams the full reviewer body *then* the
  trailer, so it resolves correctly." Empirically incomplete: a *re-close* streams
  trailer-only stdout (no fresh body), and `judge.ParseVerdict` returns `unknown` for a
  bare `Review-Verdict:` trailer (~20% of verdict-bearing closes). Delta: `parseEvents`
  now falls back to `judge.ParseVerdictTrailer` when `ParseVerdict` is unknown; a
  trailer-only fixture + a direct unit test cover it.
