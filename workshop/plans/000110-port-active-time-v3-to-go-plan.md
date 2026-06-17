# Port active-time-v3 to a native Go package Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the `python3 active-time-v3.py` subprocess behind `sdlc actual` with a native Go package, and expose the script's standalone CLI as `sdlc active-time`.

**Architecture:** A new `cmd/sdlc/internal/activetime` package holds the v3 attribution as a **pure core** (gap-truncated active-minutes, segment construction, the commit-weight/mention split) behind a thin **IO seam** (transcript `.jsonl` event loading + a `git log` window loader). `Compute(Options) Result` is the single entrypoint; `computeActual` (actual.go) calls it in-process instead of shelling out and regex-scraping stdout, and a new `sdlc active-time` subcommand renders the per-segment table for manual inspection. The Python script and its subprocess/stdout-parse/script-resolution glue are deleted. (`ARCH-DRY`: one implementation, two call sites. `ARCH-PURE`: math is pure + unit-tested without mocks; only event/commit loading touches IO.)

**Tech Stack:** Go 1.26, `github.com/spf13/cobra`, stdlib `encoding/json`, `time`, `os/exec`. Reuses `cmd/sdlc/internal/gitx` for the existing window/peer discovery.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `Event` | `cmd/sdlc/internal/activetime/event.go` | new |
| `Commit` | `cmd/sdlc/internal/activetime/commit.go` | new |
| `Segment` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `Result` / `Status` | `cmd/sdlc/internal/activetime/compute.go` | new |
| `activeMinutes` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `attributeSegment` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `buildSegments` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `parseEventMentions` | `cmd/sdlc/internal/activetime/event.go` | new |

- **Event** — `{Time time.Time; Mentions map[string]int}`. One transcript line we care about (a human user turn, or an assistant text turn when `IncludeAssistant`), with the per-event count of tracked-issue `#N` mentions in its text.
  - **Relationships:** N events per session file; flattened across all `Dirs` into one time-sorted slice.
  - **DRY rationale:** First occurrence; the mention-count shape is reused by both the segment path and the no-commits fallback.

- **Commit** — `{Time time.Time; SHA string; Subject string; Issues []string}`. A window commit with its tracked-issue refs (deduped, order-preserving) parsed from the subject.
  - **Relationships:** Commits define segment boundaries; each non-suffix segment is anchored by the commit at its end.
  - **DRY rationale:** Issue-ref parsing reuses the same `issuePattern` regex used for event mentions.

- **Segment** — `{Start, End time.Time; Active float64; Commit *Commit; Mentions map[string]int; Alloc map[string]float64; IsPrefix bool}`. The unit of attribution: `[Start, End)` event span, its gap-truncated active minutes, and the resulting per-issue allocation.
  - **Relationships:** 1 prefix (optional) + 1 per commit + 1 suffix (optional), in time order.
  - **Future extensions:** #092 (fat-segment fix) will add a max-segment-span split here — out of scope for this port.

- **Result / Status** — `Result{Status; PerIssue map[string]float64; TotalActive float64; Segments []Segment; NumEvents, NumCommits int}`. `Status` ∈ `{Measured, TelemetryGap, EmptyWindow}` — mirrors v3's exit-code contract (3 = telemetry gap, 0 = measured-or-empty; the misinvoke "2" is CLI-layer validation, not a Compute state).
  - **DRY rationale:** Single structured return consumed by both `computeActual` (reads `PerIssue[issue]`) and the subcommand renderer (prints `Segments` + `PerIssue`). Replaces today's stdout-regex (`parseV3PrimaryHours`) + exit-code (`classifyV3`) decoding.

- **activeMinutes(times, thresholdMin)** — sum of inter-event gaps, each capped at the threshold; returns minutes. Pure; 1:1 with the Python `active_minutes`.

- **attributeSegment(active, commitIssues, mentions, weight)** — allocates one segment's active minutes: `weight·active/len(commitIssues)` per commit-named issue, `(1-weight)·active·mentionShare` per mentioned issue; full segment by mention when there's no commit signal; an `"#unattributed"` bucket when transcript share has no mentions. Pure; 1:1 with Python `attribute_segment`.

- **buildSegments(events, commits, opts)** — constructs boundaries (`firstEvent ∪ commitTimes ∪ lastEvent+1s`, deduped by instant), walks events into `[start,end)` segments, anchors each by the commit whose time equals `End`, applies `PrefixWeight` to a real pre-first-commit prefix, returns `[]Segment`. Pure (operates on already-loaded slices).

**Test surface:** `segment_test.go` covers `activeMinutes` (gap capping, single/empty), `attributeSegment` (commit-only weight=1.0, mixed weight, no-commit mention-only, no-mention unattributed), and `buildSegments` (prefix detection, anchor-by-instant, suffix). `event_test.go` covers `parseEventMentions` and tool-result skipping. All run without IO.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `loadEvents` | `cmd/sdlc/internal/activetime/event.go` | new | filesystem (`.jsonl` reads) |
| `loadWindowCommits` | `cmd/sdlc/internal/activetime/commit.go` | new | `git log` (explicit `-C <repo>`) |
| `gitRun` (runner shim) | `cmd/sdlc/internal/activetime/commit.go` | new | `exec.Command` |
| `Compute` | `cmd/sdlc/internal/activetime/compute.go` | new | the two loaders above |
| `computeActual` | `cmd/sdlc/actual.go` | modified | now calls `Compute` (was: `exec` python3) |
| `NewActiveTimeCmd` | `cmd/sdlc/activetime.go` | new | cobra / stdout |

- **loadEvents(dirs, issuePattern, includeAssistant, since, until)** — globs `*.jsonl` under each dir, parses each line, yields time-filtered `Event`s sorted by time. Mirrors Python `walk_session_events` + `load_events` exactly (user-text vs pure-tool-result handling; assistant text only when requested).
  - **Injected into:** `Compute`. Pure `parseEventMentions` is split out so the JSON/string-shape logic is unit-testable without files.

- **loadWindowCommits(repo, sinceISO, untilISO, issuePattern)** — `git -C <repo> log --pretty=%H\t%aI\t%s --reverse [--since --until]`, parses to `[]Commit` with subject-parsed tracked-issue refs. Mirrors Python `load_commits`. **Takes an explicit repo path** (the `--git-repo` flag points at arbitrary repos) — this is why it does not reuse `gitx`'s cwd-based helpers; the `gitRun` shim mirrors `gitx`'s `run` var pattern (an injectable IO seam), not a duplicated query (`ARCH-DRY` considered: this is a distinct invocation — full window log with timestamps+subjects over an explicit path — that `gitx` does not provide).
  - **Injected into:** `Compute`. `gitRun` is a package `var` so `commit_test.go` drives fixtures without spawning git.

- **Compute(Options) Result** — the thin orchestrator: load events + commits, branch (no-events→TelemetryGap/EmptyWindow; no-commits→whole-window mention fallback; else→`buildSegments`), fold per-issue totals. The only place IO and pure logic meet.
  - **Injected into:** `computeActual` (in-process) and `NewActiveTimeCmd` (CLI).

- **computeActual** — keeps its signature `(repoTop, brainAbs, issueNum) actualResult` and the `selectActualDirs` brain+repo heuristic + `gitx.CommitWindow`/`DiscoverWindowIssues`; only its **engine** swaps from subprocess to `Compute`. `close.go` and `printActual` are unaffected (they consume `actualResult`).

---

## Chunk 1: M1 — the `activetime` package

### Task 1: Pure core types + `activeMinutes` + `attributeSegment`

**Files:**
- Create: `cmd/sdlc/internal/activetime/segment.go`
- Create: `cmd/sdlc/internal/activetime/event.go` (just the `Event` type for now)
- Create: `cmd/sdlc/internal/activetime/commit.go` (just the `Commit` type for now)
- Test: `cmd/sdlc/internal/activetime/segment_test.go`

- [ ] **Step 1: Write failing tests for `activeMinutes` and `attributeSegment`**

```go
package activetime

import (
	"math"
	"testing"
	"time"
)

func tm(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestActiveMinutes(t *testing.T) {
	// empty / single → 0
	if activeMinutes(nil, 15) != 0 {
		t.Fatal("empty should be 0")
	}
	if activeMinutes([]time.Time{tm("2026-01-01T00:00:00Z")}, 15) != 0 {
		t.Fatal("single should be 0")
	}
	// three events: 5-min gap (kept) + 30-min gap (capped to 15) = 20
	got := activeMinutes([]time.Time{
		tm("2026-01-01T00:00:00Z"),
		tm("2026-01-01T00:05:00Z"),
		tm("2026-01-01T00:35:00Z"),
	}, 15)
	if !approx(got, 20) {
		t.Fatalf("want 20, got %v", got)
	}
}

func TestAttributeSegmentCommitOnly(t *testing.T) {
	// weight 1.0, two commit issues, no mention share → 30 each, no unattributed.
	out := attributeSegment(60, []string{"8", "10"}, map[string]int{"8": 3}, 1.0)
	if !approx(out["8"], 30) || !approx(out["10"], 30) {
		t.Fatalf("want 30/30, got %v", out)
	}
	if _, ok := out["#unattributed"]; ok {
		t.Fatalf("weight 1.0 must not produce unattributed: %v", out)
	}
}

func TestAttributeSegmentMixedWeight(t *testing.T) {
	// weight 0.5, one commit issue (#8), mentions 8:1,10:3 → commit 30 to #8;
	// transcript 30 split 1:3 → #8 +7.5, #10 +22.5.
	out := attributeSegment(60, []string{"8"}, map[string]int{"8": 1, "10": 3}, 0.5)
	if !approx(out["8"], 37.5) || !approx(out["10"], 22.5) {
		t.Fatalf("want 37.5/22.5, got %v", out)
	}
}

func TestAttributeSegmentNoCommitMentionOnly(t *testing.T) {
	out := attributeSegment(40, nil, map[string]int{"8": 1, "10": 1}, 1.0)
	if !approx(out["8"], 20) || !approx(out["10"], 20) {
		t.Fatalf("want 20/20, got %v", out)
	}
}

func TestAttributeSegmentNoMentionUnattributed(t *testing.T) {
	out := attributeSegment(40, nil, nil, 1.0)
	if !approx(out["#unattributed"], 40) {
		t.Fatalf("want 40 unattributed, got %v", out)
	}
}
```

- [ ] **Step 2: Run to confirm fail** — `go test ./cmd/sdlc/internal/activetime/` → FAIL (undefined).

- [ ] **Step 3: Implement the types + the two pure functions** (1:1 with Python lines 192–233):

```go
package activetime

import "time"

// Event is one transcript line we attribute: a human user turn, or (when
// IncludeAssistant) an assistant text turn. Mentions counts tracked-issue #N
// refs in this single event's text.
type Event struct {
	Time     time.Time
	Mentions map[string]int
}

// Commit is a window commit with its tracked-issue subject refs (deduped,
// order-preserving). Time is the author date (%aI).
type Commit struct {
	Time    time.Time
	SHA     string // short (7)
	Subject string
	Issues  []string
}

// Segment is the unit of attribution: an [Start,End) event span, its
// gap-truncated active minutes, and the resulting per-issue allocation.
type Segment struct {
	Start, End time.Time
	Active     float64
	Commit     *Commit
	Mentions   map[string]int
	Alloc      map[string]float64
	IsPrefix   bool
}

// unattributedKey buckets transcript share that has no mention signal.
const unattributedKey = "#unattributed"

// activeMinutes sums inter-event gaps, each capped at thresholdMin. Mirrors
// active-time-v3.py active_minutes. Caller passes a time-sorted slice; we sort
// defensively to match the Python (it sorts internally).
func activeMinutes(times []time.Time, thresholdMin int) float64 {
	if len(times) < 2 {
		return 0
	}
	sorted := make([]time.Time, len(times))
	copy(sorted, times)
	sortTimes(sorted)
	cap := time.Duration(thresholdMin) * time.Minute
	var total time.Duration
	for i := 1; i < len(sorted); i++ {
		gap := sorted[i].Sub(sorted[i-1])
		if gap <= cap {
			total += gap
		} else {
			total += cap
		}
	}
	return total.Minutes()
}

// attributeSegment allocates active minutes per the v3 rule. Mirrors
// active-time-v3.py attribute_segment.
func attributeSegment(active float64, commitIssues []string, mentions map[string]int, weight float64) map[string]float64 {
	out := map[string]float64{}
	if active <= 0 {
		return out
	}
	var transcriptShare float64
	if len(commitIssues) > 0 {
		perCommit := weight * active / float64(len(commitIssues))
		for _, iss := range commitIssues {
			out[iss] += perCommit
		}
		transcriptShare = (1 - weight) * active
	} else {
		transcriptShare = active
	}
	if transcriptShare <= 0 {
		return out
	}
	total := 0
	for _, n := range mentions {
		total += n
	}
	if total > 0 {
		for iss, n := range mentions {
			out[iss] += transcriptShare * float64(n) / float64(total)
		}
	} else {
		out[unattributedKey] += transcriptShare
	}
	return out
}
```

Add `sortTimes` (a `sort.Slice` on `.Before`) in `segment.go`.

- [ ] **Step 4: Run tests → PASS.**
- [ ] **Step 5: Commit** — `#110 M1: activetime pure core — activeMinutes + attributeSegment`.

### Task 2: `buildSegments` (pure boundary/anchor logic)

**Files:**
- Modify: `cmd/sdlc/internal/activetime/segment.go`
- Test: `cmd/sdlc/internal/activetime/segment_test.go`

**The subtle parity points (must match Python lines 324–372):**
1. Boundaries = `{events[0].Time} ∪ {c.Time for c in commits} ∪ {events[-1].Time + 1s}`, **deduped by instant** and sorted. (Python `sorted(set(...))`; aware-datetime set dedupes by instant — replicate with a UTC-UnixNano key.)
2. `hasPrefix = events[0].Time.Before(commits[0].Time)`.
3. Walk events into `[segStart, segEnd)`; skip empty segments.
4. Anchor = the commit whose `Time.Equal(segEnd)` (works because every commit time is in the boundary set). Suffix segment has no anchor.
5. `isPrefix = hasPrefix && i == 0`; use `PrefixWeight` there, else `CommitWeight`.

- [ ] **Step 1: Write failing tests**

```go
func TestBuildSegmentsPrefixAndAnchor(t *testing.T) {
	events := []Event{
		{Time: tm("2026-01-01T00:00:00Z"), Mentions: map[string]int{"8": 1}}, // prefix
		{Time: tm("2026-01-01T00:10:00Z")},                                   // prefix
		{Time: tm("2026-01-01T00:40:00Z"), Mentions: map[string]int{"8": 1}}, // in commit seg
	}
	commits := []Commit{
		{Time: tm("2026-01-01T00:50:00Z"), SHA: "abc1234", Subject: "#8 work", Issues: []string{"8"}},
	}
	segs := buildSegments(events, commits, 1.0 /*commitWeight*/, 0.5 /*prefixWeight*/, 15 /*thresholdMin*/)
	if len(segs) == 0 {
		t.Fatal("expected segments")
	}
	if !segs[0].IsPrefix {
		t.Fatal("first segment should be prefix")
	}
	// The commit-anchored segment must carry the commit.
	last := segs[len(segs)-1]
	if last.Commit == nil || last.Commit.SHA != "abc1234" {
		t.Fatalf("expected anchored commit, got %+v", last.Commit)
	}
}
```

- [ ] **Step 2: Run → FAIL.**
- [ ] **Step 3: Implement** `buildSegments` taking **resolved scalar weights** (not `Options`) — the prefix-weight defaulting lives in `Compute`, so this stays a pure function of concrete values:

```go
import "sort"

func sortTimes(ts []time.Time) {
	sort.Slice(ts, func(i, j int) bool { return ts[i].Before(ts[j]) })
}

func buildSegments(events []Event, commits []Commit, commitWeight, prefixWeight float64, thresholdMin int) []Segment {
	// 1. boundaries, deduped by instant.
	bset := map[int64]time.Time{}
	add := func(t time.Time) { bset[t.UTC().UnixNano()] = t }
	add(events[0].Time)
	for _, c := range commits {
		add(c.Time)
	}
	add(events[len(events)-1].Time.Add(time.Second))
	boundaries := make([]time.Time, 0, len(bset))
	for _, t := range bset {
		boundaries = append(boundaries, t)
	}
	sortTimes(boundaries)

	hasPrefix := events[0].Time.Before(commits[0].Time)

	var segs []Segment
	eIdx := 0
	for i := 0; i < len(boundaries)-1; i++ {
		segStart, segEnd := boundaries[i], boundaries[i+1]
		var segEvents []Event
		for eIdx < len(events) && events[eIdx].Time.Before(segEnd) {
			if !events[eIdx].Time.Before(segStart) {
				segEvents = append(segEvents, events[eIdx])
			}
			eIdx++
		}
		if len(segEvents) == 0 {
			continue
		}
		times := make([]time.Time, len(segEvents))
		mentions := map[string]int{}
		for k, e := range segEvents {
			times[k] = e.Time
			for iss, n := range e.Mentions {
				mentions[iss] += n
			}
		}
		active := activeMinutes(times, thresholdMin)

		var anchor *Commit
		for ci := range commits {
			if commits[ci].Time.Equal(segEnd) {
				anchor = &commits[ci]
				break
			}
		}
		var commitIssues []string
		if anchor != nil {
			commitIssues = anchor.Issues
		}
		isPrefix := hasPrefix && i == 0
		weight := commitWeight
		if isPrefix {
			weight = prefixWeight
		}
		segs = append(segs, Segment{
			Start: segStart, End: segEnd, Active: active,
			Commit: anchor, Mentions: mentions,
			Alloc:    attributeSegment(active, commitIssues, mentions, weight),
			IsPrefix: isPrefix,
		})
	}
	return segs
}
```

- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Commit** — `#110 M1: activetime buildSegments — boundary/anchor/prefix logic`.

### Task 3: IO loaders — events + window commits

**Files:**
- Modify: `cmd/sdlc/internal/activetime/event.go` (add `parseEventMentions`, `walkSessionEvents`, `loadEvents`, `issuePattern`)
- Modify: `cmd/sdlc/internal/activetime/commit.go` (add `gitRun`, `loadWindowCommits`)
- Test: `cmd/sdlc/internal/activetime/event_test.go`, `cmd/sdlc/internal/activetime/commit_test.go`

- [ ] **Step 1: Pure-mention + JSON-shape failing tests** (no files needed — parse a single decoded line struct):

`parseEventMentions(text, issuePattern)` returns `map[string]int`; test that `"#8 and #8 then #10"` with pattern over `{8,10}` → `{8:2,10:1}`, and `#9` (untracked) is ignored.

For the JSON shape, write `loadEvents` against a `t.TempDir()` with a hand-written `.jsonl` fixture containing: a user string-content line (counts), a user line that is a **pure `tool_result`** (skipped), a user line with empty/whitespace text (skipped), an assistant text line with a mention (counted only when `includeAssistant`), **an assistant line with empty/no text (must STILL be emitted with empty mentions — see fidelity point below)**, and an out-of-window line (filtered). Assert the returned events' count + mentions + sort order. **Add an explicit assertion that the empty-text assistant event is present in the stream** (this is the parity trap).

- [ ] **Step 2: Run → FAIL.**

- [ ] **Step 3: Implement** mirroring Python `walk_session_events`/`load_events` (lines 74–151). Key fidelity points:
  - `issuePattern(issues []string) *regexp.Regexp` = `#(` + alternation of escaped issue numbers + `)\b` — RE2 supports `\b`. Capture group 1 is the number.
  - user content: string → use directly; list → join `text` blocks + any `{content: string}` blocks; **if a `tool_result` block is present and no text parts, skip the event** (pure tool result, not human typing); **empty/whitespace text → skip (this `if not text.strip(): continue` guard is in the `user` branch ONLY).**
  - **assistant (when `includeAssistant`): join `text` blocks if content is a list, else leave text empty — then ALWAYS emit the event (with whatever mentions, possibly none). Python's assistant branch (lines 113–132) has NO empty-text skip: an assistant turn always contributes its timestamp to the active-time stream.** Since `computeActual` always passes `IncludeAssistant: true`, a symmetric `text == "" → skip` here would silently drop timestamps and **lower every measured actual** — do NOT add one. This asymmetry between the user and assistant branches is load-bearing.
  - timestamp from `d.timestamp` (RFC3339; the `Z` suffix parses natively in Go); since/until filter inclusive of bounds exactly as Python (`ts < since` skip, `ts > until` skip).
  - JSON decode per line with `json.Unmarshal`; malformed line → skip (Python's `except: continue`).
  - `loadEvents` globs `*.jsonl` per existing dir, accumulates, sorts by time.

```go
// gitRun is the package-level git runner (mirrors gitx's run shim) so
// commit_test.go can inject fixtures without spawning git.
var gitRun = func(repo string, args ...string) ([]byte, error) {
	full := append([]string{"-C", repo}, args...)
	return exec.Command("git", full...).Output()
}

func loadWindowCommits(repo, sinceISO, untilISO string, pat *regexp.Regexp) ([]Commit, error) {
	args := []string{"log", "--pretty=format:%H%x09%aI%x09%s", "--reverse"}
	if sinceISO != "" {
		args = append(args, "--since="+sinceISO)
	}
	if untilISO != "" {
		args = append(args, "--until="+untilISO)
	}
	out, err := gitRun(repo, args...)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		ts, err := time.Parse(time.RFC3339, parts[1])
		if err != nil {
			continue
		}
		commits = append(commits, Commit{
			Time: ts, SHA: short7(parts[0]), Subject: parts[2],
			Issues: uniqueRefs(pat, parts[2]),
		})
	}
	return commits, nil
}
```

  > Note: the Python uses a `%x00%n` record terminator; the simpler `\t`-delimited single-line-per-commit form above is equivalent because `%s` (subject) never contains a newline. `uniqueRefs` = `pat.FindAllStringSubmatch` → dedupe group-1 preserving order (Python lines 178–186). `short7` truncates to 7 chars.

- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Commit** — `#110 M1: activetime IO loaders — events + window commits`.

### Task 4: `Compute` entrypoint + status classification

**Files:**
- Create: `cmd/sdlc/internal/activetime/compute.go`
- Test: `cmd/sdlc/internal/activetime/compute_test.go`

`Options` (full): `Dirs []string; GitRepo, SinceISO, UntilISO string; Issues []string; CommitWeight float64; PrefixWeight *float64; ThresholdMin int; IncludeAssistant bool`. **`PrefixWeight` is a `*float64` (nil = unset), NOT a float sentinel** — Python (active-time-v3.py:282) honors `--prefix-commit-weight 0` (falls back to `CommitWeight` only when the flag is `None`); a `!= 0` float sentinel would wrongly treat an explicit `0` as unset. `Compute` resolves `prefixWeight := opts.CommitWeight; if opts.PrefixWeight != nil { prefixWeight = *opts.PrefixWeight }`.

- [ ] **Step 1: Failing test** — drive `Compute` through a `gitRun` fixture (override the package var) + a tempdir `.jsonl`, asserting:
  - commits + 0 events → `Status == TelemetryGap`.
  - 0 commits + 0 events → `Status == EmptyWindow`.
  - events + commits → `Status == Measured`, `PerIssue["8"]` matches a hand-computed value, `TotalActive` set.
  - events + **no** commits → `Status == Measured` via whole-window mention fallback.

- [ ] **Step 2: Run → FAIL.**

- [ ] **Step 3: Implement** (mirrors Python `main` branching, lines 279–402, minus printing):

```go
type Status int

const (
	Measured Status = iota
	TelemetryGap // commits exist but 0 transcript events (v3 exit 3)
	EmptyWindow  // nothing to measure (v3 exit 0, "nothing to measure")
)

type Result struct {
	Status      Status
	PerIssue    map[string]float64 // minutes (renderer / caller divide by 60)
	TotalActive float64            // minutes
	Segments    []Segment
	NumEvents   int
	NumCommits  int
}

func Compute(opts Options) (Result, error) {
	pat := issuePattern(opts.Issues)
	events, err := loadEvents(opts.Dirs, pat, opts.IncludeAssistant, opts.SinceISO, opts.UntilISO)
	if err != nil {
		return Result{}, err
	}
	commits, err := loadWindowCommits(opts.GitRepo, opts.SinceISO, opts.UntilISO, pat)
	if err != nil {
		return Result{}, err
	}
	res := Result{PerIssue: map[string]float64{}, NumEvents: len(events), NumCommits: len(commits)}

	if len(events) == 0 {
		if len(commits) > 0 {
			res.Status = TelemetryGap
		} else {
			res.Status = EmptyWindow
		}
		return res, nil
	}
	prefixWeight := opts.CommitWeight
	if opts.PrefixWeight != nil {
		prefixWeight = *opts.PrefixWeight
	}

	if len(commits) == 0 {
		// whole-window mention fallback (Python lines 309–319)
		times := make([]time.Time, len(events))
		mentions := map[string]int{}
		for i, e := range events {
			times[i] = e.Time
			for iss, n := range e.Mentions {
				mentions[iss] += n
			}
		}
		active := activeMinutes(times, opts.ThresholdMin)
		res.TotalActive = active
		res.PerIssue = attributeSegment(active, nil, mentions, opts.CommitWeight)
		res.Status = Measured
		return res, nil
	}

	res.Segments = buildSegments(events, commits, opts.CommitWeight, prefixWeight, opts.ThresholdMin)
	for _, s := range res.Segments {
		res.TotalActive += s.Active
		for iss, m := range s.Alloc {
			res.PerIssue[iss] += m
		}
	}
	res.Status = Measured
	return res, nil
}
```

- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Commit** — `#110 M1: activetime Compute entrypoint + status`.

### Task 5: Port the #68 guard tests + parity verification

**Files:**
- Test: `cmd/sdlc/internal/activetime/compute_test.go` (extend)

- [ ] **Step 1: Port `test_active_time_v3.py`'s behavioral guards as Go tests** against real git in `t.TempDir()` (the package's git fixtures already use the live binary; gate with the existing repo's test conventions):
  - commits-but-0-events window → `TelemetryGap`. (was: exit 3)
  - empty window (commit dated outside `--since/--until`) → `EmptyWindow`. (was: exit 0)
  - (the "no --dir → exit 2" guard moves to the **subcommand** test in Task 7, since it's CLI-layer validation.)

- [ ] **Step 2: Run → PASS.**

- [ ] **Step 3: PARITY GATE (manual, recorded in `## Log`).** Pick a recently-closed multi-issue window (e.g. from `workshop/history`), run the Python and a tiny Go harness over identical `--dir/--git-repo/--since/--until/--issue` args, and diff per-issue hours. Use a scratch `go run` or a temporary `TestParity` that prints `PerIssue`. **Must match to the 2-decimal hours `sdlc actual` reports.** Record the window + both outputs in the issue Log. This is the M1 done-gate — do not proceed to M2 until parity holds.

- [ ] **Step 4: Commit** — `#110 M1: activetime #68 guard tests + parity check`.

- [ ] **Step 5: `sdlc milestone-close --milestone M1`** (fresh-context review auto-dispatched; fix Critical/Important; log the `Review-Verdict:`).

---

## Chunk 2: M2 — integrate into `sdlc` + retire the Python

### Task 6: Rewrite `computeActual` to call `activetime.Compute`

**Files:**
- Modify: `cmd/sdlc/actual.go`
- Modify: `cmd/sdlc/actual_test.go`

- [ ] **Step 1: Update `actual_test.go`** — the existing tests stub `v3Runner` and assert on `classifyV3`/`parseV3PrimaryHours`. Replace with the `classify`-style unit on a new small pure helper `statusFromResult(res activetime.Result, issueNum string) (actualStatus, float64)` and test it directly (TelemetryGap→gap, Measured+issue-present→measured, Measured+absent→empty). **Delete the three obsolete tests in full** (delete the whole `func`, not a partial line range): `TestParseV3PrimaryHours` (`actual_test.go:27–57`), `TestResolveActualScript` (`58–105`, the M3 #104 substrate-ancestor owner-resolution test — that whole mechanism is deleted), and `TestClassifyV3` (`143–165`). Keep `TestCwdToTranscriptDir`, `TestSelectActualDirs`, `TestActualCmd_Registered`. **Confirm `close_actualdev_test.go:44` still passes** — it drives the `actualNoWindow` path through `computeActual`, which the rewrite preserves (no commits reference the issue → `actualNoWindow`).

- [ ] **Step 2: Run → FAIL (compile).**

- [ ] **Step 3: Rewrite `computeActual`** and delete the dead glue:
  - **Delete:** `v3Runner`, `resolveActualScript`, `actualScriptRel`, `statFile`, `parseV3PrimaryHours`, `classifyV3`, the `python3` `exec.LookPath`, and the `actualNoScript` *meaning* "script/python missing". **Keep** `substrateChain` (used by `startplan.go`/`propagatebase.go` — do **not** delete).
  - **Repurpose** the `actualNoScript` enum constant → rename `actualError` (compute/IO failure carries `Detail`); update its one `printActual` case (message already generic: "can't auto-compute (%s) — fall back…").
  - New body:

```go
func computeActual(repoTop, brainAbs, issueNum string) actualResult {
	res := actualResult{Issue: issueNum}
	firstSHA, firstISO, lastISO, _ := gitx.CommitWindow(issueNum)
	if firstSHA == "" {
		res.Status = actualNoWindow
		return res
	}
	res.Window = firstSHA[:8] + " → HEAD"
	res.Peers, _ = gitx.DiscoverWindowIssues(firstISO, lastISO, issueNum)
	res.Dirs = selectActualDirs(repoTop, brainAbs)
	if len(res.Dirs) == 0 {
		res.Status, res.Detail = actualTelemetryGap, "no brain/repo transcript dirs found under "+transcriptsRoot
		return res
	}
	out, err := activetime.Compute(activetime.Options{
		Dirs: res.Dirs, GitRepo: repoTop, SinceISO: firstISO, UntilISO: lastISO,
		Issues: res.Peers, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		res.Status, res.Detail = actualError, err.Error()
		return res
	}
	res.Status, res.Hours = statusFromResult(out, issueNum)
	return res
}

// statusFromResult maps a Compute Result to the actual outcome for issueNum.
// Pure — the integration contract, unit-tested without git/files.
func statusFromResult(out activetime.Result, issueNum string) (actualStatus, float64) {
	switch out.Status {
	case activetime.TelemetryGap:
		return actualTelemetryGap, 0
	case activetime.EmptyWindow:
		return actualEmptyWindow, 0
	default: // Measured
		if mins, ok := out.PerIssue[issueNum]; ok {
			return actualMeasured, mins / 60
		}
		return actualEmptyWindow, 0
	}
}
```

  - Drop now-unused imports (`bytes`, `io`, `os/exec`, `strconv` if unused, `regexp`). Add the `activetime` import.

- [ ] **Step 4: Run `go test ./cmd/sdlc/...` → PASS; `go vet ./cmd/sdlc/...` clean.**
- [ ] **Step 5: Commit** — `#110 M2: computeActual calls activetime.Compute (drop python subprocess)`.

### Task 7: `sdlc active-time` subcommand + renderer + helptext

**Files:**
- Create: `cmd/sdlc/activetime.go` (`NewActiveTimeCmd`, the table renderer)
- Create: `cmd/sdlc/helptext/active-time.md`
- Modify: `cmd/sdlc/main.go` (wire `add(NewActiveTimeCmd(), "active-time", …)`)
- Test: `cmd/sdlc/activetime_test.go`

- [ ] **Step 1: Failing test.** **Exit-code mechanism (this is the crux — get it right):** `die` (term.go:52) is hardcoded to `os.Exit(1)` and `main.go:30` maps *every* `RunE` error to `os.Exit(1)` — there is **no** typed-error→exit-code path. The only precedent for a custom code is a **direct `os.Exit`** (changecode.go:333). So factor the command body into a **testable pure-ish core**:

  ```go
  // runActiveTime executes the active-time computation, writes the table to
  // out and diagnostics to errOut, and RETURNS the process exit code
  // (0 measured/empty · 2 misinvoke · 3 telemetry-gap). No os.Exit here — the
  // cobra RunE wrapper calls os.Exit(runActiveTime(...)), so this core is
  // unit-testable with bytes.Buffers (no subprocess, no real exit).
  func runActiveTime(opts activetime.Options, out, errOut io.Writer) int
  ```

  `RunE` parses flags into `Options`, then `os.Exit(runActiveTime(opts, cmd.OutOrStdout(), cmd.ErrOrStderr()))`. The **test calls `runActiveTime` directly** with `bytes.Buffer`s and asserts the returned `int` + buffer contents — no cobra/subprocess needed. Cases: empty `--dir`/`--issue` → returns **2** + "no --dir given" on errOut (the #68 misinvoke guard — note this is exit **2**, NOT `die`'s 1); commits-but-0-events fixture window → returns **3** + "TELEMETRY UNAVAILABLE"; a measured window → returns **0** + the table on out.

- [ ] **Step 2: Implement** `runActiveTime` (the renderer, mirroring Python's table lines 284–402) and the cobra command exposing every original flag: `--dir` (StringArray), `--git-repo` (required), `--since`, `--until`, `--issue` (StringArray, ≥1 required), `--commit-weight` (default 1.0), `--prefix-commit-weight` (bind to a `*float64` only when the flag was set — use `cmd.Flags().Changed("prefix-commit-weight")`, so an explicit `0` is honored), `--threshold-min` (default 15), `--include-assistant`. `runActiveTime` validates `--dir`/`--issue` non-empty → **return 2** with the #68 message on errOut (do NOT use `die` — it exits 1). Render to `out`:
  - header (`# v3 segment-anchored attribution`, weights, issues, counts);
  - `TelemetryGap` → errOut TELEMETRY-UNAVAILABLE message + **return 3**;
  - `EmptyWindow` → "nothing to measure" + **return 0**;
  - no-commits fallback / segment table + per-issue totals (`hr` + `min`), dividing minutes by 60 → **return 0**.
  - Render the `unattributedKey` bucket as `unattributed` (not the Python's cosmetic `##unattributed`) — purely a display fix; never parsed.
  - **Sort** all per-segment rows (by start time) and per-issue totals (by issue) explicitly — Go map iteration is unordered (see Risks).

- [ ] **Step 3: Write `helptext/active-time.md`** — adapt the Python module docstring (the v3 method explanation + flags + edge cases) into the helptext house style (see `actual.md`/`close.md`). Note it is the manual-inspection sibling of `sdlc actual`.

- [ ] **Step 4: Wire in `main.go`** next to `actual` — `add(NewActiveTimeCmd(), "active-time", "Per-issue active-time attribution table (the v3 engine, standalone)")` and set `.Long` via `helptext.MustGet("active-time")` (the `add` helper already does this by key).

- [ ] **Step 5: Add `TestActiveTimeEmbedded`** in `helptext/embed_test.go` mirroring the existing `TestCloseEmbedded`/`TestPushEmbedded` guards (assert `helptext.MustGet("active-time")` is non-empty), for symmetry with the other verbs. `embed.go` uses `//go:embed *.md`, so the new file is auto-picked-up; no embed-code change needed.
- [ ] **Step 6: Run `go build ./... && go test ./cmd/sdlc/...` → PASS.** Manually run `sdlc active-time --help` and a real window; eyeball the table vs the Python's.
- [ ] **Step 7: Commit** — `#110 M2: sdlc active-time subcommand + helptext`.

### Task 8: Migrate doc/explainer references off the Python script

**Files (live code + current-state docs — must migrate):**
- Modify: `cmd/sdlc/helptext/actual.md` (lines 1, 7, 20, 29 — drop "python command"/"active-time-v3.py"/"python3 absent"; the engine is in-binary now)
- Modify: `cmd/sdlc/helptext/close.md` (lines 107, 115 — "runs active-time-v3 itself" stays true; ensure no script/python wording)
- Modify: `cmd/sdlc/helptext/milestone-close.md` (line 52 — fine as "active-time-v3" the method; verify)
- Modify: `construct/local/issues/SKILL.md` (line 24 — the "prints the exact `active-time-v3.py` command" paragraph → point manual inspection at `sdlc active-time`)
- Modify: **`scripts/close-issue.py` (lines 96, 234)** — the legacy Python issue-close fallback (run by `Makefile.workflow` when `bin/sdlc` isn't built) *builds a command string* telling the user to `python3 construct/local/issues/active-time-v3.py …`. After the script is deleted (Task 9) this prints a pointer to a missing file. Redirect the printed guidance to `sdlc active-time …` (close-issue.py is an ariadne base-layer file — the right place to fix per the base-layer-source-of-truth rule). Line 196 ("active-time-v3 algorithm") names the *method* — leave it.
- Modify: **open issue `workshop/issues/000092-…md` (lines 15, 119)** — references `construct/local/issues/active-time-v3.py` as the *source to fix*. It's a live (open) issue; repoint its source references at `cmd/sdlc/internal/activetime/segment.go` (`buildSegments`) so #092's eventual implementer lands in the right place.
- Audit: `cmd/sdlc/close.go` (comments at 321/658/690/744/749 — "active-time-v3" as the *method* name is fine; remove any "python" wording)

**Explicitly OUT of scope (historical — atlas-current-state-only: history lives in git/plans/history, do NOT overwrite):** `docs/vision/2026-05-25-01-pensive-…:84`, `workshop/plans/000104-skill-system-v2-plan.md:361`, `workshop/history/*`, `workshop/pensive/2026-06-02-…`, and `workshop/lessons.md:128` (a description of the #68 *incident* + its still-valid rule — the script name is incidental to the lesson, not a broken live pointer). Leave all of these untouched.

- [ ] **Step 1:** Whole-tree grep for residual pointers: `git grep -n "active-time-v3\.py\|active_time_v3\|python3 .*active-time"`. Classify every hit as **live/current-state** (migrate) or **historical** (leave, per the scope-out list above). Migrate only those that name the *script/python* and are live; keep "active-time-v3" where it names the algorithm/version.
- [ ] **Step 2:** Edit each live file above.
- [ ] **Step 3:** `go test ./cmd/sdlc/...` (the embed test from Task 7 Step 5 confirms `active-time.md` resolves).
- [ ] **Step 4: Commit** — `#110 M2: migrate explainers off active-time-v3.py → sdlc active-time`.

### Task 9: Delete the Python script + its test

**Files:**
- Delete: `construct/local/issues/active-time-v3.py`
- Delete: `construct/local/issues/test_active_time_v3.py`

- [ ] **Step 1:** `git rm` both. Re-run the Task-8 `git grep` and confirm **zero residual references in live code + current-state docs** (cmd/, construct/, scripts/, atlas/, open issues). Remaining hits must be only the historical files on Task 8's scope-out list (history/, plans/, pensive/, docs/vision/, lessons.md) — those are expected and intentional.
- [ ] **Step 2:** `go build ./... && go test ./cmd/sdlc/... && go vet ./cmd/sdlc/...` all green.
- [ ] **Step 3: Commit** — `#110 M2: delete active-time-v3.py + test_active_time_v3.py (ported to Go)`.

### Task 10: Atlas update

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md` (line 142 — `actual` no longer shells to `construct/local/issues/active-time-v3.py`; add the `active-time` verb + the `internal/activetime` package to the surface)
- Modify: `atlas/workflow/ledger-landscape.md` (line 43 — `actual_hours` is "derived from `active-time-v3.py`" → derived from the in-binary `activetime` engine, per `feedback_atlas_current_state_only`)
- Modify: `atlas/workflow/weave.md` (line 107 — names `active-time-v3.py`; reconcile to the in-binary package — this current-state atlas file was missed by a cmd-scoped grep, caught by the whole-tree `git grep`)
- Check: `atlas/index.md` links any new/changed atlas file.

- [ ] **Step 1:** Read each, reconcile to the new truth (atlas is current-state, not a changelog — chase every stale `active-time-v3.py` reference to ground).
- [ ] **Step 2: Commit** — `#110 M2: atlas — activetime package + sdlc active-time verb`.

### Task 11: M2 verification + close

- [ ] **Step 1:** `go build ./... && go test ./cmd/sdlc/... && go vet ./cmd/sdlc/...` — all green (paste output into Log).
- [ ] **Step 2:** Re-run the PARITY check end-to-end through the real binary: `sdlc actual --issue <recently-closed-N>` and confirm the suggested `--actual` equals the pre-port value, **with no `python3` process** (the script is gone). Record in Log.
- [ ] **Step 3:** `sdlc close --issue 110 --milestone M2 --verified '<evidence>'` (computes actuals; satisfies atlas/verified/verdict gates). The auto-dispatched fresh-context review runs here; fix Critical/Important before the boundary.

---

## Risks & non-goals

- **Numeric drift is the headline risk** — velocity calibration depends on these hours. The parity gate (Task 5 Step 3 + Task 11 Step 2) is the defense; do not skip it. Float summation order differs between Python dict iteration and Go map iteration, but final per-issue sums are order-independent (addition is associative within f64 tolerance at this magnitude); the 2-decimal `hr` rounding absorbs any last-ULP difference.
- **Go map iteration is unordered** — anywhere the Python relied on `sorted(...)` for *output* (the per-segment table, per-issue totals), the renderer must sort explicitly. Attribution math itself is order-independent.
- **Non-goal:** #092 (a long inter-commit gap creating one fat over-attributing segment) is a *behavior change* and stays out — this port reproduces current behavior verbatim so parity is checkable. #092 builds on `buildSegments` afterward.
- **Non-goal:** changing the brain+repo dir heuristic, the commit window cap, or the `--commit-weight 1.0` choice — all preserved.
- **Non-goal:** the v1 `construct/local/issues/active-time.py` (the older per-session estimator predecessor) is a separate standalone tool, **intentionally untouched**. It is not wired into the binary; the only reference to it is inside `active-time-v3.py`'s own docstring (which gets deleted with the script). Note for Task 8/9: the whole-tree grep's `python3 .*active-time` pattern *will* match v1 invocations — those are expected, leave them.
- **Base-layer transition (this is the ariadne owner repo).** #104 made `active-time-v3.py` owner-resolvable so derivatives without a local copy found it in their substrate ancestor at runtime. After deletion, that resolution vanishes — but the steady state is the in-process engine. Transition behavior is **graceful degradation, not a break**: a derivative still running an *old* `sdlc` binary after the script is gone walks its substrate chain, fails to find the script, and falls back to the `actualNoScript`/judgment-estimate path until it rebuilds. On the next `sdlc` rebuild (build-in-owner) it picks up the in-process `activetime` engine and the fallback disappears. No derivative action required beyond the normal rebuild.
```
