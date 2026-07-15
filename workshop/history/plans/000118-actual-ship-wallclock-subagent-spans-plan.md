# Fill subagent spans in active-time (ship wall-clock) — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `sdlc actual` / `sdlc active-time` measure **active ship wall-clock** by counting synchronous subagent-execution spans in full instead of truncating them as idle, and reconcile the "operator-attention" framing to "ship wall-clock".

**Architecture:** The active-time engine (`cmd/sdlc/internal/activetime`) keeps its pure-core / thin-IO split (ARCH-PURE). The IO seam (`walkSessionEvents`) gains structural detection of subagent dispatch→return pairs, emitting a new pure `TaskSpan` interval alongside `Event`s. The pure gap-accounting (`activeMinutes`) is generalized to a single union-of-intervals core (`activeMinutesUnion`) — capped inter-event gaps **unioned with** full-length task spans — and `activeMinutes` becomes a thin wrapper over it (ARCH-DRY: one gap-math implementation). Non-task idle still truncates at 15 min; only task-bounded gaps fill.

**Tech Stack:** Go (stdlib only: `encoding/json`, `time`, `sort`). Docs: Markdown (ariadne helptext + atlas; brain calibration files).

**Cross-repo note:** Engine + helptext + atlas changes land in **ariadne**. The calibration framing files (`calibration-findings.md`, `calibration-ledger.tsv`) live in **brain** (`data/life/42shots/velocity/`) and are edited as part of this issue.

**Key discovery driving the design (see issue Log + Spec):** the subagent-dispatch tool is named **`Agent`** in real transcripts (not `Task`; the `Task*` tools are the todo list). Detection keys off `tool_use.name == "Agent"`. Empirically all 33 historical Agent spans are < 15 min, so this fix is **unit-correctness + forward-looking** (it changes no current ledger row); the calibration banner is corrected accordingly rather than crediting the fix for the supervised ~3.5× overshoot.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `TaskSpan` | `cmd/sdlc/internal/activetime/event.go` | new |
| `activeMinutesUnion` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `unionMinutes` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `clampSpans` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `activeMinutes` | `cmd/sdlc/internal/activetime/segment.go` | modified |
| `buildSegments` | `cmd/sdlc/internal/activetime/segment.go` | modified |

- **TaskSpan** — one synchronous subagent execution, `{Start, End time.Time}`: dispatch (assistant `tool_use` name=="Agent") → matching return (`user` `tool_result` with the same id). The gap it spans is active project work, not idle.
  - **Relationships:** N:1 with a transcript session (many spans per file); produced by the parser, consumed by the active-minutes math. No ownership of Events — it is a parallel output stream keyed only by time.
  - **DRY rationale:** First occurrence of "a time interval that should count in full regardless of the gap cap." Modeled as an explicit interval so the union math (below) treats it uniformly with capped gaps — no special-case branch in the cap loop.
  - **Future extensions:** If background/async subagents ever need crediting, or non-Agent long-running tools, the producer widens to recognize more tool names; the interval shape is unchanged.

- **activeMinutesUnion(times, spans, thresholdMin)** — the single gap-math core. Builds one interval per inter-event gap (`[t[i-1], min(t[i], t[i-1]+cap)]`), adds one interval per task span (full length), and returns the **union** length in minutes.
  - **DRY rationale:** Replaces the bespoke "sum of capped gaps" loop. `activeMinutes(times, thr)` becomes `activeMinutesUnion(times, nil, thr)`. With no spans the gap intervals are adjacent and non-overlapping, so union == sum-of-capped-gaps — **bitwise-identical to the old behavior** (parity preserved; `TestActiveMinutes` and `TestAttributionGolden` still pass unchanged).
  - **Future extensions:** Any future "this interval counts in full" source (e.g. a manually-asserted focus block) is just another interval fed to the union.

- **unionMinutes(intervals)** — pure helper: sort intervals by start, merge overlaps, sum lengths in minutes. Makes parallel/overlapping spans collapse to wall-clock (overlap union), the right quantity per Spec.

- **clampSpans(spans, segStart, segEnd)** — pure helper: intersect each span with `[segStart, segEnd)`, dropping empties. Ensures a span straddling a commit boundary is split across segments (each segment counts only its portion) so the per-segment sum equals the whole-window union with no double-counting. **This invariant only holds if span-bearing segments are never skipped** — see `buildSegments` change (3) below; the return is a dropped `tool_result` (no Event), so the post-commit tail of a span often lands in a *zero-event* segment.

- **activeMinutes** (modified) — thin wrapper: `return activeMinutesUnion(times, nil, thresholdMin)`. Signature unchanged.

- **buildSegments** (modified) — gains a `spans []TaskSpan` parameter. Three changes: (1) the final boundary becomes `max(lastEvent+1s, max(span.End))` so a span whose return lands after the last emitted event is not cut; (2) each segment computes `active := activeMinutesUnion(segTimes, clampSpans(spans, segStart, segEnd), thresholdMin)`; (3) **the zero-event skip becomes `if len(segEvents) == 0 && len(clampedSpans) == 0 { continue }`** — a segment that carries a clamped span but no events is still emitted (its active comes from the span, attributed via the existing anchor-at-segEnd rule; with `CommitWeight=1.0` that's 100% to the anchor commit's issues, else unattributed). **Span boundaries are deliberately NOT added to the boundary set** — segments keep ending at commits so commit-anchored attribution is preserved; splitting at a span boundary would orphan the span's bulk into an anchor-less (mention-only ⇒ usually unattributed) segment.
  - **Why (3) is load-bearing — the commit-inside-span bug:** during a *synchronous* Agent span the main thread is blocked, but the **subagent itself commits** (a delegated build that ships code), so a commit referencing the issue lands *strictly inside* `[dispatch, return]`. The pre-commit piece (with the dispatch Event) counts and attributes fine; the post-commit tail `[commit, return)` has no Event (the return is a dropped `tool_result`) → its segment would be skipped and the tail silently lost. Change (3) keeps it, attributed to the commit anchoring that segment.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `walkSessionEvents` | `cmd/sdlc/internal/activetime/event.go` | modified | transcript .jsonl parse |
| `loadEvents` | `cmd/sdlc/internal/activetime/event.go` | modified | filesystem glob + window filter |
| `Compute` | `cmd/sdlc/internal/activetime/compute.go` | modified | engine orchestration |
| `contentBlock` | `cmd/sdlc/internal/activetime/event.go` | modified | JSON shape |

- **walkSessionEvents** (modified) — single pass over the file; in addition to emitting `Event`s, tracks Agent dispatch ids → timestamps in a local `pending map[string]time.Time`, and on a matching `tool_result` emits a `TaskSpan`. Span detection is **structural and independent of `includeAssistant`** (it reads tool_use/tool_result blocks, not assistant text). Returns `([]Event, []TaskSpan, error)`.
  - **Injected into:** nothing — it IS the IO seam. The pure `activeMinutesUnion` receives its `spans` output.
  - **Parity guard:** the existing event-emission switch is unchanged; ts-parsing is hoisted above it (a line with unparseable ts is still dropped, same as today), and span-tracking runs before the text-skip `continue`s so pure-tool_result lines (currently dropped as events) still contribute their span return.
- **loadEvents** (modified) — aggregates spans across all session files, then **clamps each span to the window** `[since, until]` (Start←max(Start,since), End←min(End,until)) and keeps it iff `End > Start`, sorts by `Start`. Returns `([]Event, []TaskSpan, error)`. Clamping (not just filtering by Start) gives a clean invariant: **all measured active time lies within the claim→close window** — a subagent still returning at the close instant is clamped to `until` rather than counting work past the window. (For the primary consumer `sdlc actual`, `IncludeAssistant=true`, so the dispatch is itself an Event and `events[0] ≤ span.Start` — every span is fully covered by the tiled segments.)
- **Compute** (modified) — destructures the new spans return; passes spans to `buildSegments` and to the no-commits whole-window fallback (`activeMinutesUnion(times, spans, thr)`). `Result`/`Segment` struct shapes are **unchanged**, so `runActiveTime` (renderer) and `computeActual` need no edits — span-containing segments simply carry a larger `.Active`.
  - **Injected into:** consumed by `computeActual` (`actual.go`) and `runActiveTime` (`activetime.go`), both unchanged.

**Test surface:** all four pure entities get colocated unit tests in `segment_test.go`; the parser changes get `event_test.go` cases on crafted JSONL (no IO mocks — real temp files via the existing `writeJSONL` helper); the end-to-end fill is covered in `compute_test.go` / `parity_test.go` via a crafted fixture with a >15-min Agent span. No process-level fake is needed — the "external service" here is the on-disk transcript, already exercised with real temp files.

---

## Chunk 1: Engine

### Task 1: `TaskSpan` + union gap-math core (pure)

**Files:**
- Modify: `cmd/sdlc/internal/activetime/event.go` (add `TaskSpan` type)
- Modify: `cmd/sdlc/internal/activetime/segment.go` (add `unionMinutes`, `activeMinutesUnion`, `clampSpans`; rewrite `activeMinutes` as wrapper)
- Test: `cmd/sdlc/internal/activetime/segment_test.go`

- [ ] **Step 1: Write failing tests** in `segment_test.go`:

```go
func TestActiveMinutesUnionFillsSpan(t *testing.T) {
	// Two events 30 min apart: bare gap caps at 15. With a Task span covering
	// the full 30 min, the span fills it → 30.
	times := []time.Time{tm("2026-01-01T00:00:00Z"), tm("2026-01-01T00:30:00Z")}
	if got := activeMinutesUnion(times, nil, 15); !approx(got, 15) {
		t.Fatalf("bare 30-min gap should cap at 15, got %v", got)
	}
	spans := []TaskSpan{{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T00:30:00Z")}}
	if got := activeMinutesUnion(times, spans, 15); !approx(got, 30) {
		t.Fatalf("span-covered 30-min gap should fill to 30, got %v", got)
	}
}

func TestActiveMinutesUnionParityNoSpans(t *testing.T) {
	// Identical to TestActiveMinutes: 5-min kept + 30-min capped = 20.
	times := []time.Time{
		tm("2026-01-01T00:00:00Z"), tm("2026-01-01T00:05:00Z"), tm("2026-01-01T00:35:00Z"),
	}
	if got := activeMinutesUnion(times, nil, 15); !approx(got, 20) {
		t.Fatalf("want 20, got %v", got)
	}
}

func TestActiveMinutesUnionOverlapIsWallClock(t *testing.T) {
	// Two overlapping (parallel) spans union to wall-clock, not summed effort.
	spans := []TaskSpan{
		{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T00:20:00Z")},
		{Start: tm("2026-01-01T00:10:00Z"), End: tm("2026-01-01T00:30:00Z")},
	}
	if got := activeMinutesUnion(nil, spans, 15); !approx(got, 30) {
		t.Fatalf("overlapping spans should union to 30, got %v (not 40)", got)
	}
}

func TestClampSpans(t *testing.T) {
	spans := []TaskSpan{{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T01:00:00Z")}}
	got := clampSpans(spans, tm("2026-01-01T00:20:00Z"), tm("2026-01-01T00:40:00Z"))
	if len(got) != 1 || !got[0].Start.Equal(tm("2026-01-01T00:20:00Z")) || !got[0].End.Equal(tm("2026-01-01T00:40:00Z")) {
		t.Fatalf("clamp to [00:20,00:40), got %+v", got)
	}
	// Non-overlapping span → dropped.
	if out := clampSpans(spans, tm("2026-01-01T02:00:00Z"), tm("2026-01-01T03:00:00Z")); len(out) != 0 {
		t.Fatalf("non-overlapping span should drop, got %+v", out)
	}
}
```

- [ ] **Step 2: Run, verify they fail**

Run: `cd /Users/xianxu/workspace/ariadne && go test ./cmd/sdlc/internal/activetime/ -run 'Union|ClampSpans' -v`
Expected: FAIL — `undefined: activeMinutesUnion` / `clampSpans` / `TaskSpan`.

- [ ] **Step 3: Add `TaskSpan` to `event.go`** (near `Event`):

```go
// TaskSpan is one synchronous subagent execution: the interval between an
// assistant `tool_use` dispatch (name "Agent") and its matching `tool_result`
// return, both timestamped in the operator's main transcript. The subagent runs
// in its own transcript (outside the dirs we read), so this gap shows as one big
// inter-event gap; it is active project work, not idle, and must count in full
// rather than truncate at the 15-min cap (#118).
type TaskSpan struct {
	Start, End time.Time
}
```

- [ ] **Step 4: Add the union core to `segment.go`** and rewrite `activeMinutes`:

```go
// interval is a half-open [s,e) time span used to compute active minutes as a
// union (so overlapping/parallel task spans collapse to wall-clock, not summed
// effort).
type interval struct{ s, e time.Time }

// unionMinutes merges overlapping intervals and returns the total covered
// duration in minutes. Pure.
func unionMinutes(ivals []interval) float64 {
	if len(ivals) == 0 {
		return 0
	}
	sort.Slice(ivals, func(i, j int) bool { return ivals[i].s.Before(ivals[j].s) })
	var total time.Duration
	cur := ivals[0]
	for _, iv := range ivals[1:] {
		if iv.s.After(cur.e) {
			total += cur.e.Sub(cur.s)
			cur = iv
		} else if iv.e.After(cur.e) {
			cur.e = iv.e
		}
	}
	total += cur.e.Sub(cur.s)
	return total.Minutes()
}

// activeMinutesUnion is the single gap-math core: each inter-event gap counts up
// to thresholdMin (idle truncation), each task span counts in full, and the
// result is the UNION of those intervals (#118). With spans == nil the gap
// intervals are adjacent and non-overlapping, so the union equals the old
// sum-of-capped-gaps — parity is exact.
func activeMinutesUnion(times []time.Time, spans []TaskSpan, thresholdMin int) float64 {
	sorted := make([]time.Time, len(times))
	copy(sorted, times)
	sortTimes(sorted)
	capGap := time.Duration(thresholdMin) * time.Minute
	var ivals []interval
	for i := 1; i < len(sorted); i++ {
		end := sorted[i]
		if sorted[i].Sub(sorted[i-1]) > capGap {
			end = sorted[i-1].Add(capGap)
		}
		ivals = append(ivals, interval{sorted[i-1], end})
	}
	for _, sp := range spans {
		if sp.End.After(sp.Start) {
			ivals = append(ivals, interval{sp.Start, sp.End})
		}
	}
	return unionMinutes(ivals)
}

// clampSpans intersects each span with [start,end), dropping empties, so a span
// that straddles a commit boundary is split across segments (each counts only
// its portion — no double-count when the per-segment actives are summed).
func clampSpans(spans []TaskSpan, start, end time.Time) []TaskSpan {
	var out []TaskSpan
	for _, sp := range spans {
		s, e := sp.Start, sp.End
		if s.Before(start) {
			s = start
		}
		if e.After(end) {
			e = end
		}
		if e.After(s) {
			out = append(out, TaskSpan{s, e})
		}
	}
	return out
}
```

Then replace the body of `activeMinutes` with:

```go
func activeMinutes(times []time.Time, thresholdMin int) float64 {
	return activeMinutesUnion(times, nil, thresholdMin)
}
```
(Keep its doc comment; note it now delegates to the union core.)

- [ ] **Step 5: Run all activetime tests, verify pass** (new + parity)

Run: `cd /Users/xianxu/workspace/ariadne && go test ./cmd/sdlc/internal/activetime/ -run 'Union|ClampSpans|ActiveMinutes|AttributionGolden' -v`
Expected: PASS — including the unchanged `TestActiveMinutes` and `TestAttributionGolden` (parity).

- [ ] **Step 6: Commit**

```bash
cd /Users/xianxu/workspace/ariadne
git add cmd/sdlc/internal/activetime/event.go cmd/sdlc/internal/activetime/segment.go cmd/sdlc/internal/activetime/segment_test.go
git commit -m "#118: TaskSpan + union gap-math core (parity-preserving)"
```

### Task 2: Parse Agent spans in the transcript (IO seam)

**Files:**
- Modify: `cmd/sdlc/internal/activetime/event.go` (`contentBlock`, `walkSessionEvents`, `loadEvents`)
- Test: `cmd/sdlc/internal/activetime/event_test.go`

- [ ] **Step 1: Write failing tests** in `event_test.go`:

```go
func TestLoadEventsTaskSpans(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "s.jsonl"), []string{
		// Agent dispatch (assistant tool_use, id A1) at 00:00.
		`{"timestamp":"2026-01-01T00:00:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"A1","name":"Agent","input":{}}]}}`,
		// return (user tool_result for A1) at 00:30 — a 30-min subagent span.
		`{"timestamp":"2026-01-01T00:30:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"A1","content":"done"}]}}`,
		// a non-Agent tool_use (Bash) — must NOT produce a span.
		`{"timestamp":"2026-01-01T00:31:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"B1","name":"Bash","input":{}}]}}`,
		`{"timestamp":"2026-01-01T00:32:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"B1","content":"ok"}]}}`,
	})
	_, spans, err := loadEvents([]string{dir}, nil, true, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want exactly 1 Agent span, got %d: %+v", len(spans), spans)
	}
	if !spans[0].Start.Equal(tm("2026-01-01T00:00:00Z")) || !spans[0].End.Equal(tm("2026-01-01T00:30:00Z")) {
		t.Fatalf("span should be [00:00,00:30], got %+v", spans[0])
	}
}

func TestLoadEventsDanglingDispatchNoSpan(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "s.jsonl"), []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"A1","name":"Agent","input":{}}]}}`,
		// no matching tool_result → no span.
	})
	_, spans, err := loadEvents([]string{dir}, nil, true, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 0 {
		t.Fatalf("dangling dispatch should yield no span, got %+v", spans)
	}
}
```
Also: update the **existing** `loadEvents` call sites in `event_test.go` (`TestLoadEventsShapes`, `TestLoadEventsMissingDirSkipped`) to the new 3-value return (`evs, _, err := ...`).

- [ ] **Step 2: Run, verify fail**

Run: `cd /Users/xianxu/workspace/ariadne && go test ./cmd/sdlc/internal/activetime/ -run 'TaskSpans|Dangling' -v`
Expected: FAIL — `loadEvents` returns 2 values, not 3 (compile error) until Step 3.

- [ ] **Step 3: Extend `contentBlock`** in `event.go` with the tool fields:

```go
type contentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Content   json.RawMessage `json:"content"`
	Name      string          `json:"name"`        // tool_use: tool name (e.g. "Agent")
	ID        string          `json:"id"`          // tool_use: id matched by tool_result
	ToolUseID string          `json:"tool_use_id"` // tool_result: id of the dispatch it answers
}
```

- [ ] **Step 4: Add span-tracking to `walkSessionEvents`** — change its signature and hoist ts-parsing:

```go
func walkSessionEvents(path string, pat *regexp.Regexp, includeAssistant bool) ([]Event, []TaskSpan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var events []Event
	var spans []TaskSpan
	pending := map[string]time.Time{} // Agent dispatch id → dispatch time
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var d rawLine
		if json.Unmarshal([]byte(line), &d) != nil {
			continue
		}
		if d.Timestamp == "" {
			continue
		}
		ts, err := parseISO(d.Timestamp)
		if err != nil {
			continue
		}
		// Structural span tracking — independent of includeAssistant and of the
		// text-skip below (a pure tool_result is dropped as an Event but still
		// closes a span). Decodes the content array once; non-array content has
		// no tool blocks.
		if blocks, ok := decodeBlocks(d.Message.Content); ok {
			for _, blk := range blocks {
				switch {
				case d.Type == "assistant" && blk.Type == "tool_use" && blk.Name == "Agent" && blk.ID != "":
					pending[blk.ID] = ts
				case d.Type == "user" && blk.Type == "tool_result" && blk.ToolUseID != "":
					if start, ok := pending[blk.ToolUseID]; ok {
						spans = append(spans, TaskSpan{Start: start, End: ts})
						delete(pending, blk.ToolUseID)
					}
				}
			}
		}
		// Event emission — unchanged logic.
		var text string
		switch {
		case d.Type == "user":
			t, skip := userText(d.Message.Content)
			if skip {
				continue
			}
			text = t
		case d.Type == "assistant" && includeAssistant:
			text = assistantText(d.Message.Content)
		default:
			continue
		}
		events = append(events, Event{Time: ts, Mentions: parseEventMentions(text, pat)})
	}
	return events, spans, nil
}
```

- [ ] **Step 5: Update `loadEvents`** to collect + window-filter + return spans:

```go
func loadEvents(dirs []string, pat *regexp.Regexp, includeAssistant bool, sinceISO, untilISO string) ([]Event, []TaskSpan, error) {
	// ... existing since/until parse, returning (nil, nil, err) on parse error ...
	var events []Event
	var spans []TaskSpan
	for _, d := range dirs {
		// ... existing dir stat + glob ...
		for _, f := range files {
			evs, sps, err := walkSessionEvents(f, pat, includeAssistant)
			if err != nil {
				continue
			}
			for _, e := range evs { /* existing window filter */ }
			for _, s := range sps {
				// Clamp to the window so all measured active time ∈ [since,until].
				if haveSince && s.Start.Before(since) {
					s.Start = since
				}
				if haveUntil && s.End.After(until) {
					s.End = until
				}
				if s.End.After(s.Start) {
					spans = append(spans, s)
				}
			}
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Time.Before(events[j].Time) })
	sort.Slice(spans, func(i, j int) bool { return spans[i].Start.Before(spans[j].Start) })
	return events, spans, nil
}
```
(Update every `return ..., err` in `loadEvents` to the 3-value form.)

- [ ] **Step 6: Run, verify pass**

Run: `cd /Users/xianxu/workspace/ariadne && go test ./cmd/sdlc/internal/activetime/ -run 'TaskSpans|Dangling|LoadEvents' -v`
Expected: PASS (new span tests + the updated existing shape tests).

- [ ] **Step 7: Commit**

```bash
cd /Users/xianxu/workspace/ariadne
git add cmd/sdlc/internal/activetime/event.go cmd/sdlc/internal/activetime/event_test.go
git commit -m "#118: detect Agent dispatch->return spans in the transcript parser"
```

### Task 3: Thread spans through `Compute` + `buildSegments`

**Files:**
- Modify: `cmd/sdlc/internal/activetime/compute.go` (`Compute`)
- Modify: `cmd/sdlc/internal/activetime/segment.go` (`buildSegments`)
- Test: `cmd/sdlc/internal/activetime/segment_test.go`, `cmd/sdlc/internal/activetime/compute_test.go` (or `parity_test.go`)

- [ ] **Step 1: Write failing tests.** Segment-level (in `segment_test.go`):

```go
func TestBuildSegmentsFillsSpan(t *testing.T) {
	// One commit; a 40-min Agent span sits in the suffix between two events that
	// are 40 min apart (bare gap would cap at 15). The span fills it.
	events := []Event{
		{Time: tm("2026-01-01T00:50:00Z"), Mentions: map[string]int{"8": 1}},
		{Time: tm("2026-01-01T01:00:00Z")},                                   // dispatch event
		{Time: tm("2026-01-01T01:40:00Z"), Mentions: map[string]int{"8": 1}}, // next turn after return
	}
	commits := []Commit{{Time: tm("2026-01-01T00:50:00Z"), SHA: "aaa", Subject: "#8", Issues: []string{"8"}}}
	spans := []TaskSpan{{Start: tm("2026-01-01T01:00:00Z"), End: tm("2026-01-01T01:40:00Z")}}
	withSpan := buildSegments(events, commits, spans, 1.0, 0.5, 15)
	noSpan := buildSegments(events, commits, nil, 1.0, 0.5, 15)
	var aw, an float64
	for _, s := range withSpan { aw += s.Active }
	for _, s := range noSpan { an += s.Active }
	// noSpan: 10 (kept) + 15 (capped) = 25. withSpan: 10 + 40 (filled) = 50.
	if !approx(an, 25) || !approx(aw, 50) {
		t.Fatalf("want noSpan=25 withSpan=50, got %v / %v", an, aw)
	}
}

func TestBuildSegmentsCommitInsideSpan(t *testing.T) {
	// The blocker case: a commit lands STRICTLY INSIDE a span (a subagent that
	// commits mid-run). The post-commit tail segment has NO events (the return is
	// a dropped tool_result), so it must not be skipped — the full span counts and
	// attributes to the commits anchoring each piece.
	events := []Event{
		{Time: tm("2026-01-01T00:00:00Z"), Mentions: map[string]int{"8": 1}}, // dispatch event
		{Time: tm("2026-01-01T00:50:00Z"), Mentions: map[string]int{"8": 1}}, // next turn (well after return)
	}
	commits := []Commit{
		{Time: tm("2026-01-01T00:10:00Z"), SHA: "aaa", Subject: "#8 mid", Issues: []string{"8"}}, // INSIDE span
		{Time: tm("2026-01-01T00:50:00Z"), SHA: "bbb", Subject: "#8 end", Issues: []string{"8"}},
	}
	spans := []TaskSpan{{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T00:40:00Z")}}
	segs := buildSegments(events, commits, spans, 1.0 /*commitWeight*/, 1.0 /*prefixWeight*/, 15)
	var tot, toIssue8 float64
	for _, s := range segs {
		tot += s.Active
		toIssue8 += s.Alloc["8"]
	}
	// Span is 40 min; the [00:00,00:10) piece (10) + [00:10,00:40) tail (30) must
	// both count → 40 total, all attributed to #8 (weight 1.0, both anchors #8).
	if !approx(tot, 40) {
		t.Fatalf("commit-inside-span: want 40 total active, got %v (tail dropped?)", tot)
	}
	if !approx(toIssue8, 40) {
		t.Fatalf("commit-inside-span: want 40 min → #8, got %v", toIssue8)
	}
}

func TestBuildSegmentsSpanTailPastLastEvent(t *testing.T) {
	// Dispatch is the LAST event; the return lands 30 min later (no further
	// event). The final boundary must extend so the tail is not cut.
	events := []Event{
		{Time: tm("2026-01-01T00:00:00Z"), Mentions: map[string]int{"8": 1}},
		{Time: tm("2026-01-01T00:05:00Z")}, // dispatch, last event
	}
	commits := []Commit{{Time: tm("2026-01-01T00:00:00Z"), SHA: "aaa", Subject: "#8", Issues: []string{"8"}}}
	spans := []TaskSpan{{Start: tm("2026-01-01T00:05:00Z"), End: tm("2026-01-01T00:35:00Z")}}
	segs := buildSegments(events, commits, spans, 1.0, 0.5, 15)
	var tot float64
	for _, s := range segs { tot += s.Active }
	// 5-min gap kept + 30-min span filled = 35.
	if !approx(tot, 35) {
		t.Fatalf("want 35 (tail not cut), got %v", tot)
	}
}
```
End-to-end (in `parity_test.go`, real git + transcript): a fixture with a single commit and a >15-min Agent dispatch→return, asserting `res.TotalActive` reflects the full span (a `TestComputeFillsAgentSpanEndToEnd`).

Also update existing `buildSegments` call sites in `segment_test.go` (`TestBuildSegmentsPrefixAndAnchor`, `TestBuildSegmentsNoPrefix`) to pass `nil` for the new `spans` arg.

- [ ] **Step 2: Run, verify fail**

Run: `cd /Users/xianxu/workspace/ariadne && go test ./cmd/sdlc/internal/activetime/ -run 'FillsSpan|TailPast|EndToEnd' -v`
Expected: FAIL — `buildSegments` arity mismatch (compile) until Step 3.

- [ ] **Step 3: Modify `buildSegments`** — add `spans []TaskSpan` after `commits`; extend the final boundary and union spans per segment:

```go
func buildSegments(events []Event, commits []Commit, spans []TaskSpan, commitWeight, prefixWeight float64, thresholdMin int) []Segment {
	bset := map[int64]time.Time{}
	add := func(t time.Time) { bset[t.UTC().UnixNano()] = t }
	add(events[0].Time)
	for _, c := range commits {
		add(c.Time)
	}
	// Final boundary extends past the last event to cover any span whose return
	// lands after it (a trailing subagent run), so its tail is not cut.
	last := events[len(events)-1].Time.Add(time.Second)
	for _, sp := range spans {
		if sp.End.After(last) {
			last = sp.End
		}
	}
	add(last)
	// ... boundaries build + sort + hasPrefix (unchanged) ...
	// inside the per-segment loop:
	clampedSpans := clampSpans(spans, segStart, segEnd)
	// CHANGED skip: keep a span-bearing segment even with no events (else the
	// post-commit tail of a span is silently dropped — the commit-inside-span bug).
	if len(segEvents) == 0 && len(clampedSpans) == 0 {
		continue
	}
	times, mentions := eventTimesAndMentions(segEvents)
	active := activeMinutesUnion(times, clampedSpans, thresholdMin)
	// ... anchor lookup + attributeSegment + append (unchanged) ...
}
```
Note: do **not** add `sp.Start`/`sp.End` to the boundary set — segments must keep ending at commits so commit-anchored attribution is preserved (a span boundary would orphan the span's bulk into an anchor-less, usually-unattributed segment).

- [ ] **Step 4: Modify `Compute`** in `compute.go`:

```go
events, spans, err := loadEvents(opts.Dirs, pat, opts.IncludeAssistant, opts.SinceISO, opts.UntilISO)
// ...
// no-commits fallback:
active := activeMinutesUnion(times, spans, opts.ThresholdMin)
// commits path:
res.Segments = buildSegments(events, commits, spans, opts.CommitWeight, prefixWeight, opts.ThresholdMin)
```

- [ ] **Step 5: Run full package tests + vet**

Run: `cd /Users/xianxu/workspace/ariadne && go test ./cmd/sdlc/... && go vet ./cmd/sdlc/...`
Expected: PASS, no vet findings.

- [ ] **Step 6: Commit**

```bash
cd /Users/xianxu/workspace/ariadne
git add cmd/sdlc/internal/activetime/segment.go cmd/sdlc/internal/activetime/compute.go cmd/sdlc/internal/activetime/segment_test.go cmd/sdlc/internal/activetime/parity_test.go
git commit -m "#118: thread Agent spans through Compute + buildSegments (fill, don't cap)"
```

---

## Chunk 2: Framing reconciliation + atlas

### Task 4: Reconcile "operator-attention" → "ship wall-clock"

**Files:**
- Modify (ariadne): `cmd/sdlc/helptext/estimate.md` (UNIT NOTE, lines 63–69)
- Verify (ariadne): `cmd/sdlc/helptext/{change-code,close,issue,actual}.md` — confirm no stale operator-attention framing (grep showed none; the issue's file list was aspirational — fix only where the wrong framing actually appears, per "chase stale refs to ground")
- Modify (brain): `data/life/42shots/velocity/calibration-findings.md` (reframe banner)
- Modify (brain): `data/life/42shots/velocity/calibration-ledger.tsv` (header comment)
- Modify (ariadne): atlas — `atlas/workflow/ledger-landscape.md` and/or `atlas/workflow/sdlc-binary.md` (active-time semantics note)

- [ ] **Step 1: estimate.md UNIT NOTE** — replace the "measures OPERATOR-ATTENTION / the two diverge as work is delegated" wording with: estimate-logic-v2 estimates BUILD-EFFORT and `sdlc actual` measures **ship wall-clock** (idle removed, subagent-execution spans kept, #118); both are the same unit (one engineer + AI ship-time), so the close-time ledger compares like with like. Keep the `vocab.go` canonical-source pointer.

- [ ] **Step 2: Verify the drift guard still passes.** The real guard is `TestEstimateHelptextMatchesVocab` (`cmd/sdlc/estimate_helptext_test.go`); it isolates only the `CLOSED PRIMITIVE VOCABULARY` block and **truncates at the literal string `UNIT NOTE`**, so it never inspects the UNIT NOTE prose and `vocab.go` needs no change. **Constraint:** the edit MUST keep the literal `UNIT NOTE` header line intact, or the guard's block-end delimiter breaks. (No estimate.md prose is embed-guarded — `TestActiveTimeEmbedded` covers active-time.md only.)

Run: `cd /Users/xianxu/workspace/ariadne && go test ./cmd/sdlc/ -run Estimate -v` (and the whole package in Task 5)
Expected: PASS.

- [ ] **Step 3: calibration-findings.md banner** — per the operator decision (build it, correct the rationale): rewrite the reframe banner so it no longer claims the ~3.5× supervised overshoot is "largely a wrong-ruler artifact." State the measured fact: all historical Agent spans are < 15 min, so filling them changes no current (supervised) ledger row; #118 is **unit-correctness + forward-looking** (matters once delegation produces >15-min spans). Keep the "stale v2 numbers" hypothesis **live** as the leading explanation for the current supervised overshoot. Add this as a `## Revisions` entry too (don't silently overwrite — append the reframe-correction with date + reason).

- [ ] **Step 4: calibration-ledger.tsv header** — adjust the comment so it describes `actual` as ship wall-clock (idle removed, subagent spans kept), and document the **4 pre-fix ledger rows decision**: re-measuring #116/#117 under the new engine yields identical numbers (their spans are sub-cap), so the rows are **kept as-is** — no `window_trusted` flip is caused by #118, no re-measure needed. Record this verdict in the issue `## Log` as well.

- [ ] **Step 5: Atlas note** — add a one-paragraph note (in `atlas/workflow/ledger-landscape.md`, near the active-time description) that active-time now measures ship wall-clock: non-task gaps truncate at 15 min, but `Agent` dispatch→return spans count in full (union of overlapping spans = wall-clock). Ensure `atlas/index.md` still links the file.

- [ ] **Step 6: Commit (two repos)**

```bash
cd /Users/xianxu/workspace/ariadne
git add cmd/sdlc/helptext/estimate.md atlas/
git commit -m "#118: reconcile actual framing to ship wall-clock (helptext + atlas)"
cd /Users/xianxu/workspace/brain
git add data/life/42shots/velocity/calibration-findings.md data/life/42shots/velocity/calibration-ledger.tsv
git commit -m "ariadne#118: correct calibration framing (ship wall-clock; spans sub-cap, no row change)"
```

---

## Chunk 3: Verification

### Task 5: End-to-end verification

- [ ] **Step 1: Full suite + vet**

Run: `cd /Users/xianxu/workspace/ariadne && go test ./... && go vet ./cmd/sdlc/...`
Expected: all PASS.

- [ ] **Step 2: Real-data smoke** — run `sdlc active-time` over a real window known to contain a long Agent span (the 833s span on 2026-06-03 in the ariadne project dir) and confirm that segment's `min` reflects the full ~13.9 min (it already did, being sub-cap — but verify the engine path is exercised and unchanged). Then craft / locate any >15-min span if available and confirm fill. Document the command + observed output in the issue `## Log`.

- [ ] **Step 3: `sdlc actual --issue 118`** (or `sdlc state`) sanity — confirm the binary runs the new engine without error on this issue's own window. Capture output for the close `--verified` evidence.

- [ ] **Step 3b: Demonstrate the "4 pre-fix rows unchanged" claim (don't just assert it).** Before Task 4 Step 4 records the keep-as-is decision in the durable ledger header, actually re-run `sdlc actual --issue 116` and `sdlc actual --issue 117` under the new engine and confirm the measured hours match the ledger's existing `actual` (0.41 / 0.93). If they differ, the spans-are-sub-cap census was incomplete — revisit the decision instead of recording it. (Per plan-quality INFO: verify-don't-assert.)

- [ ] **Step 4: Build the binary** (downstream consumers run it):

Run: `cd /Users/xianxu/workspace/ariadne && go build ./cmd/sdlc`
Expected: clean build.

---

## Done-when traceability (issue → plan)

- active-time fills Task-bounded gaps, truncates non-Task idle → Task 1 + Task 3 (`TestActiveMinutesUnionFillsSpan`, `TestBuildSegmentsFillsSpan`, `TestBuildSegmentsCommitInsideSpan`).
- `sdlc actual` on a delegated issue measures wall-clock not ~15 min → Task 3 end-to-end test + Task 5 smoke.
- Docs/framing reconciled (helptext + calibration-findings + ledger header) → Task 4.
- Decide on the 4 pre-fix ledger rows → Task 4 Step 4 (kept as-is: sub-cap spans ⇒ identical on re-measure).
- Atlas note → Task 4 Step 5.

## Notes / risks

- **Parity is the dominant risk.** `activeMinutes` → `activeMinutesUnion(…, nil, …)` must stay bit-identical; `TestActiveMinutes` + `TestAttributionGolden` are the guards (Task 1 Step 5). Do not change the event-emission switch in `walkSessionEvents`.
- **Commit-inside-span (the review-caught blocker).** A subagent that commits mid-run puts a commit strictly inside `[dispatch, return]`; the post-commit tail segment has no Event (return is a dropped `tool_result`) and must not be skipped. Fix = `buildSegments` change (3) (keep span-bearing zero-event segments); guard = `TestBuildSegmentsCommitInsideSpan`. Do NOT "fix" this by adding span boundaries — that orphans attribution.
- **Window semantics.** Spans are clamped to `[since, until]` in `loadEvents` (all measured time ∈ window). Span-time attribution rides on the surrounding/subagent commits referencing the issue (true in practice under `CommitWeight=1.0`); a span ending in an anchor-less suffix with a dispatch turn that doesn't mention the issue falls to unattributed — a pre-existing mention-attribution limitation, not new.
- **ARCH-DRY:** one gap-math implementation (`activeMinutesUnion`); `activeMinutes` is a wrapper. No parallel cap-loop.
- **ARCH-PURE:** span detection lives in the IO seam (`walkSessionEvents`); union/clamp math is pure and unit-tested with no mocks (real temp files for the parser).
- **Tool name:** key off `"Agent"`. If a future harness reintroduces `"Task"`, widen the `blk.Name == "Agent"` check to a small set — but do NOT match `TaskCreate`/`TaskUpdate` (todo tools, not subagents).
- **Single-pass close (no Mx):** this is one coherent deliverable; the mandatory fresh-eyes review runs at `sdlc close` over the branch-point→HEAD window (AGENTS.md §3).

## Revisions

### 2026-06-18 — close-review (FIX-THEN-SHIP) fixes
The mandatory boundary review (window 59366da..HEAD) returned FIX-THEN-SHIP — no
Critical, two Important, both now resolved:

1. **Framing scope was incomplete.** The reconciliation (Task 4) covered `helptext/`
   + brain but missed the contract-bearing **estimate-quality judge prompt**
   (`cmd/sdlc/internal/judge/prompts.go`), which still asserted "operator-attention
   (what the actual measures) will diverge." Reframed to: actual now measures ship
   wall-clock (same unit as build-effort → should converge); the residual heavy-
   fan-out gap is the parallelism/overlap discount (#118 non-goal), not operator-
   attention. (`propagatebase.go` + `vocab.go` mentions are genuinely out of scope —
   they don't describe the actual's unit.) Caution: that prompt is a backtick raw
   string — no backticks in inserted text.
2. **The committed end-to-end test was missing.** `TestComputeFillsAgentSpanEndToEnd`
   (Task 3 Step 1 + Done-when) had not been delivered, leaving the Compute↔span
   wiring uncovered. Now added in `parity_test.go`: real git commit + a 30-min Agent
   span → asserts `TotalActive`≈35 and `PerIssue["8"]`≈35 (un-filled would be ~5).

Minor (not blocking, noted): span matching is **per-transcript-file** — a dispatch/
return straddling a session-compaction boundary isn't paired. Forward-looking
(all historical spans within-file + sub-cap); caveat added to atlas/ledger-landscape.
