---
id: 000212
status: open
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
---

# Add dup to the issue status vocabulary

## Problem

There is no way to close an issue as a duplicate. `categories.terminal` is
`["done", "wontfix", "punt"]` (`construct/vocabulary/issue.cue`), and none of
them is true of a duplicate:

- `wontfix` — "rejected; will not be done". False: the work *is* being done,
  under another id.
- `punt` — "deferred". False: nothing is deferred.
- `done` — false, and it would pollute velocity calibration with a close that
  measured no work.

So closing a duplicate records a claim about the work that is wrong. Concrete
instance from 2026-09-02: `pair#178` was closed `wontfix` while its entire
content lives on as `pair#165` — the file says "rejected" about work that is
scheduled.

**`dup` alone is not enough.** A duplicate that does not say *which* issue
supersedes it is barely better than `wontfix` plus prose: the reader still has
to search. The status needs a pointer beside it.

## Spec

**1. The status.** `categories.terminal` gains `"dup"`, with
`when.dup: "duplicate; the work is tracked by another issue"`. `#Status`,
`#Terminal` and the three laws derive from `categories`, so nothing else in the
model needs touching.

**2. Lifecycle edges**, mirroring `wontfix`/`punt` so every state that can be
abandoned can be deduped, and the `escapable`/`reachable` laws hold:

```
{from: "open",         to: "dup", event: "dedupe"}
{from: "working",      to: "dup", event: "dedupe"}
{from: "blocked",      to: "dup", event: "dedupe"}
{from: "codecomplete", to: "dup", event: "dedupe"}
{from: "dup",          to: "working", event: "reopen"}
```

The reopen edge matters: a wrong dedupe is a judgment call made in a hurry, and
every other terminal status is already reversible.

**3. `duplicate_of`, required when `status: dup`.** The model already has this
exact shape for actuals —

```
if status == "done" || status == "codecomplete" {
    actual_hours!: (number & >0) | #ActualNotApplicable
}
```

— so use the same compiled guard: `if status == "dup" { duplicate_of!: string }`.
The value is an issue ref, qualified `<repo>#<id>` across repos and bare within
one. Frontmatter is already open (`...`), so the field costs no schema
loosening.

**4. No actuals guard.** `dup` follows `wontfix`/`punt`: measured hours are not
required. Where real work happened before the dedupe, that time belongs to the
surviving issue — a note for whoever closes it, not a mechanism here.

**5. Consumers derive, or this is not done.** The status set is supposed to come
from the model — `sdlc issue --help` states "the status set above is derived
from the model", and `AGENTS.md` says not to hardcode the enum. A shadow sweep
must confirm no consumer gained a hand-written `"dup"`: `pkg/vocab` readers,
`sdlc issue set-status`'s accepted values, the helptext, the archival rule for
terminal statuses, and any status rendering (`sdlc state`, `sdlc issue list`).
A hand-maintained restatement is a deferred consumer, not a finished one
(`ARCH-PURPOSE`).

## Done when

- `vocabulary vet` passes with `dup` in `categories.terminal`, its `when` entry,
  and the five edges; `documented-value`, `reachable` and `escapable` all hold.
- `make weave` regenerates `construct/generated/vocabulary/issue.json` with the
  new status and edges, and `vocabulary check` reports fresh.
- `sdlc issue set-status dup --issue N --duplicate-of <ref>` sets both fields,
  and **refuses** without the ref.
- `sdlc issue set-status dup` on an issue in each of open / working / blocked /
  codecomplete succeeds; `dup → working` reopens.
- A `dup` issue archives on merge/push like other terminal statuses.
- Shadow sweep: no consumer restates the status list; every one derives.
- Real-world check: `pair#178` is re-statused from `wontfix` to `dup` with
  `duplicate_of: pair#165`, and reads truthfully afterward.

## Plan

- [ ] `dup` in `categories.terminal` + `when` + the five lifecycle edges.
- [ ] `duplicate_of` compiled guard, mirroring the actuals guard.
- [ ] `--duplicate-of` on `sdlc issue set-status`, refusing when absent.
- [ ] `make weave`; confirm the exported JSON and freshness check.
- [ ] Shadow sweep for hardcoded status lists.
- [ ] Re-status `pair#178`.

## Log

### 2026-09-02

Raised while merging `pair#178` into `pair#165`. `#178` currently carries
`wontfix`, which is a false statement about work that is scheduled — the
closest available lie. That file names this issue as the fix, so re-statusing
it is the acceptance check rather than a follow-up.
