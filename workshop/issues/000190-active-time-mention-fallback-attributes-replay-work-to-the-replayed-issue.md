---
id: 000190
status: working
deps: []
github_issue:
created: 2026-07-29
updated: 2026-07-29
estimate_hours: 3.75
started: 2026-07-29T16:23:09-07:00
---

# active-time mention-fallback attributes replay work to the replayed issue

## Problem

The active-time engine's mention-fallback attributes work to whatever issue the session
TALKS about, so replaying or postmortem-ing issue X charges the time to X rather than to the
issue doing the replaying.

Measured live at ariadne#187's close: **46.1m/77% attributed to pair#127** by
`mention fallback without issue commit boundary`, because #187's Task 14 replayed pair#127
and its commits, prose and evidence file cite `#127` constantly. #187 measured 2.29h; the
true figure is higher, and pair#127 — long closed — gained 46 minutes it did not spend.

Both directions are wrong in a way that matters, because these numbers feed velocity
calibration: the replaying issue looks cheaper than it was, and a CLOSED issue's actual
silently grows after the fact.

This is a general hazard, not a #187 quirk. Any issue whose work is *about* another issue —
replays, postmortems, migrations, "fix the thing #N introduced" — hits it. The engine
already knows the difference in principle: it warns `without issue commit boundary`, meaning
it had no commit anchoring the segment and fell back to text mentions.

## Spec

- An issue-referencing **commit boundary** should outrank a text mention within the same
  window. The warning already identifies the weak case; the fix is to let the strong signal
  win rather than blending them.
- Attribution to an issue whose status is terminal (`done`/`wontfix`) is the loudest signal
  that a mention fallback is wrong — a closed issue is not accruing work. Treat it as at
  minimum a refusal to attribute, or a hard warning.
- **Do not** solve this by asking the operator to hand-correct hours: a typed actual is
  exactly what the close gate exists to prevent (#178). The engine must attribute correctly,
  or say it cannot.
- Whatever lands must be checkable against #187's window, where the right answer is known:
  the 46.1m belongs to #187.

## Done when

- A window whose commits all anchor `#A` while the prose mentions `#B` attributes to `#A`.
  Commit boundaries outrank mention fallback; the engine already distinguishes them in its
  warning text.
- Mention-fallback attribution to a **terminal-status** issue (`done`/`wontfix`) refuses or
  warns loudly — a closed issue is not accruing work, and silently growing its actual after
  the fact corrupts calibration history.
- **Regression check against a known answer:** re-measuring ariadne#187's real window
  returns the 46.1m currently charged to pair#127 back to #187.
- `sdlc actual` output states which rule attributed each segment, so a wrong number is
  diagnosable without reading the engine.
- No operator-facing workaround is introduced: hand-typing a corrected actual stays exactly
  as forbidden as it is today (#178).
- `atlas/workflow/` documents the precedence rule, since attribution now has one.

## Plan

Plan: `workshop/plans/000190-cross-repo-refs-parsed-as-local-plan.md`. Single-pass work —
plain checkboxes, one `sdlc close` (§3: tag `Mx` only for work with ≥2 boundaries closed
separately).

**Superseded by the 2026-07-29 Revisions entry** (the original rows planned a
commit-boundary-precedence rule and a terminal-status guard; both were dropped once the root
cause turned out to be a ref-parsing bug affecting the commit path too). These rows mirror the
plan's Tasks 1–6:

- [x] `issueref` package: the ONE qualifier+id grammar — `Ref`, `QualifiedIDPattern`, `ScanRE`,
      `Find`, `IsLocal` (exact, not prefix), `LocalNums`, `CountLocal`; corpus-derived table
      test + mutation-verified boundary (Task 1)
- [x] `gitx.DiscoverWindowIssues` takes `selfRepo` as a parameter and derives from `issueref`;
      moves onto the package `run` shim so the self-qualified case is testable at all (Task 2)
- [x] `activetime` derives on BOTH paths — `Commit.Issues` and transcript mentions — with the
      self-qualifier from `opts.GitRepo`, not the process cwd; plus a `foreign refs ignored`
      warning so the exclusion is observable rather than silent (Task 3)
- [x] `migrate.go`'s two encodings compose from `QualifiedIDPattern`, retiring `refScanRE` and
      `spanRefRE` — the step that makes this 5 → 1 rather than a sixth encoding (Task 4)
- [x] Regression check with a known answer: the 46.1m charged to ariadne#127 returns to #187,
      measured over a fixed window (Task 5)
- [ ] `atlas/workflow/ledger-landscape.md` documents the rule; `sdlc close --issue 190` (Task 6)

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Revised once on the estimate-quality review — see the note
below.*

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=0.60 impl=0.10
item: greenfield-go-module   design=0.45 impl=0.30
item: smaller-go-module      design=0.10 impl=0.15
item: smaller-go-module      design=0.15 impl=0.20
item: smaller-go-module      design=0.05 impl=0.12
item: cross-cutting-refactor design=0.10 impl=0.15
item: smaller-go-module      design=0.05 impl=0.15
item: atlas-docs             design=0.05 impl=0.06
item: atlas-docs             design=0.05 impl=0.08
item: milestone-review       design=0.00 impl=0.20
item: milestone-review       design=0.00 impl=0.20
item: milestone-review       design=0.00 impl=0.20
design-buffer: 0.15
total: 3.75
```

**Verified by recomputation:** Σdesign 1.60 × 1.15 = 1.8400, + Σimpl 1.91 = **3.7500** vs
stated 3.75 (δ 0.0000, tolerance 0.188). Exact, not tolerance-riding — #187's estimate review
flagged a 0.12 drift that reconciled only because 5% absorbed it, and that is the wrong habit
for a number the helptext calls DERIVED.

**Per-item derivation** (v3.1: design from the v2/v2.1 table unchanged, `impl=` at **40%** of
the v2 table's implementation hours, +15% design buffer since a thorough plan doc exists):

| item | what | why this value |
|---|---|---|
| `issue-spec` | root-cause investigation + Spec supersede | design 0.60 of 0.5–1.5: read `segment.go`/`compute.go`/`commit.go`/`util.go`, probed `\B` in a scratch program, measured a 400-subject corpus, then rewrote the filed Spec through `## Revisions` because both its mechanism and its fix were wrong |
| `greenfield-go-module` | the `issueref` package | **v2.1 Step 2.5 (library-availability) applied explicitly:** the grammar already exists (`parseRef` + `refScanRE`), so design HALVES from 0.90 to **0.45** — under the 0.5 table floor by that named rule, not by an implicit discount. impl 0.30 under the 0.32 ceiling |
| `smaller-go-module` | `gitx` derive + `selfRepo` param + `run`-shim move | impl 0.15 of a 0.20 ceiling |
| `smaller-go-module` | `activetime` both derive paths | impl 0.20 — **at** ceiling; the largest chunk (util/commit/compute/event + 4 tests + 5 `event_test.go` call sites) |
| `smaller-go-module` | the `foreign refs ignored` warning | split OUT of the row above rather than writing that one over its ceiling |
| `cross-cutting-refactor` | `migrate.go`'s two encodings compose from the fragment | design 0.10 under the 0.2 floor **for a named reason: no independent design work.** The decision (export the fragment un-anchored) is priced in Task 1's row; duplicating it here would double-count one decision. impl 0.15 covers the group-index check |
| `smaller-go-module` | the regression MEASUREMENT (Task 5) | **the estimate-quality review's best catch:** this was priced `atlas-docs` (4.8-min ceiling) while actually requiring a BEFORE run sequenced ahead of Task 1, resolving the transcript slug rather than guessing, a two-`--dir` fixed-window invocation, a corroborating `sdlc actual`, and three asserted outcomes. Re-slugged. The slug is imperfect (this is not a Go module) but its ceiling honestly fits the work, and pricing the issue's most valuable verification as prose was the real error |
| `atlas-docs` | `ledger-landscape.md` rule paragraph | impl 0.06 **under** the 0.08 ceiling — genuinely one paragraph |
| `atlas-docs` | the evidence DOC prose | at ceiling; the measurement runs moved to their own row above |
| `milestone-review` | the close review dispatch | 0.20, ceiling |
| `milestone-review` | fixing the close review's findings | **0.20, raised from 0.15** — the review caught that 0.15 sat *below* the 0.20 row it was split off from, contradicting the very evidence cited to justify the split (#187's close review returned 4 Important whose fixes were demonstrably not 9 minutes) |
| `milestone-review` | the `change-code` gate rounds | **added.** `started:` anchors the window at the claim commit, so 3 plan-quality rounds + 2 estimate-quality dispatches land INSIDE the measured actual. The block previously argued exactly this to decline the ×0.2 discount, then failed to carry it into the itemization — #187's "don't re-derive across rounds" is an argument against re-deriving, not against budgeting them once |

**Calibration context, stated not padded.** Design share **42.7%** pre-buffer (1.60/3.75),
inside the 41–61% peer band (#172 61%, #180 41%, #186 53%, #187 45.6%). Recent rows drift
OVER-estimate — #187 closed **3.6×** over (8.45 est / 2.32 measured). This estimate was derived
ONCE after the plan cleared (the timing #187 B1 introduced to stop cross-round accretion) and
then revised ONCE on the estimate-quality review, in the direction that review pointed: the four
INFO findings all said *floor, not ceiling*, and 3.29 → 3.75 is that correction, not padding.

**The ×0.2 spec-quality discount is deliberately NOT applied**, and the itemization now honors
that reasoning rather than only asserting it: with `started:` anchoring at the claim commit, the
investigation, plan authoring, and every gate round are inside the measured window, so all three
are itemized above.

**Standing advisory, recorded not acted on:** Tasks 1–4 are independently committable, so
within-session overlap could compress the measured actual below the sequential sum of the impl
rows. That is #117's data to collect (a #118 non-goal), and worth naming the recursion — the
calibration data this would feed is the data this very bug corrupted.

## Revisions` because both its mechanism and its fix were wrong |
| `greenfield-go-module` | the `issueref` package | design 0.45 of 0.5–2 — LOW because the grammar already existed (`parseRef`/`refScanRE`); the design work was deciding to join that lineage, not inventing one. impl 0.30 under the 0.32 ceiling |
| `smaller-go-module` | `gitx` derive + `selfRepo` param + `run`-shim move | impl 0.15 of 0.20 ceiling |
| `smaller-go-module` | `activetime` both derive paths | impl 0.20 — **at** ceiling; the largest single chunk (util/commit/compute/event + 4 tests) |
| `smaller-go-module` | the `foreign refs ignored` warning | split OUT of the row above rather than writing that one above its ceiling (#187's estimate review established splitting over exceeding) |
| `cross-cutting-refactor` | `migrate.go`'s two encodings compose from the fragment | includes the group-index shift check, which is the part that could silently break `spanRefRE` |
| `atlas-docs` | `ledger-landscape.md` rule paragraph | impl 0.06 **under** the 0.08 ceiling — it is genuinely one paragraph, and pinning it at ceiling just because the ceiling is low would be the uniform-ceiling habit #187 flagged |
| `atlas-docs` | regression evidence doc | at ceiling: includes the before/after measurement runs, not just prose |
| `milestone-review` | the close review dispatch | 0.20, ceiling |
| `milestone-review` | fixing the close review's findings | split, because one item cannot express review-plus-fix: #187's close review returned 4 Important, and fixing them (a new `UpgradeHeader` + tests, a `vet_test.sh` block, a registry entry, the I4 two-half fix) was demonstrably not 12 minutes |

**Calibration context, stated not padded.** Design share is **47.1%** pre-buffer
(1.55/3.29), inside the 41–61% band of the nearest peers (#172 61%, #180 41%, #186 53%,
#187 45.6%). The repo's recent rows drift OVER-estimate — #187 closed at **3.6×** over
(8.45 est / 2.32 measured), and its own postmortem note identifies the mechanism: an estimate
re-derived across gate rounds accretes items for work the judge surfaces. This estimate was
derived ONCE, after the plan cleared, which is exactly the timing #187 B1 introduced to stop
that accretion. If it still lands materially over, that is evidence the primitive table itself
needs the recalibration ariadne#127 tracks — and worth noting that #127 is the very issue whose
data this bug corrupted.

**The ×0.2 spec-quality discount is deliberately NOT applied.** `started:` anchors the window
at the claim commit, so the investigation, plan authoring and both gate rounds land INSIDE the
measured actual. Discounting design for a good spec while the measurement counts the work of
producing it would guarantee an under-estimate.

## Revisions

### 2026-07-29 — root cause found, and the Spec above describes the SYMPTOM

**Reason:** investigation before design. The filed Spec assumed the defect was mention
fallback out-competing commit boundaries, and proposed a precedence rule plus a
terminal-status guard. Both premises are wrong, and the real cause is narrower, sharper, and
affects **more paths** than mention fallback.

**The actual defect: a cross-repo issue ref is parsed as a LOCAL issue number.**

`#(\d+)\b` has no left boundary, so `pair#127` matches as `127`. My commit
`28428da #187 M2: pair#127 replay harness + round 1 evidence` therefore reads as referencing
local issue 127 — and **ariadne#127 exists**: `000127-recalibrate-estimate-logic-v2-high.md`,
long closed. So 46 minutes of #187's work were charged to an unrelated archived issue about
*recalibrating estimates*, corrupting precisely the calibration data that issue produced.

**Three sites share the missing boundary, not one:**

| site | what it feeds | consequence |
|---|---|---|
| `gitx/window.go:384` `issueRefRE` | `DiscoverWindowIssues` → the tracked peer set | admits a foreign issue as a mention target |
| `activetime/commit.go:67` `allIssuePattern()` | `Commit.Issues` → the claimant | **commit-weighted** share splits equally with the foreign issue |
| `activetime/util.go:34` `issuePattern()` | transcript mention counts | every `pair#127` in prose counts as local `#127` |

The second row is why the filed Spec's precedence rule would not have worked: attribution is
corrupted on the **commit** path too, so making commits outrank mentions would still
misattribute — `attributeRun` splits `weight * active` equally across `Commit.Issues`, and
`127` is in that slice. A precedence rule fixes nothing when both sides are poisoned by the
same parse.

**Superseding Spec:**
- A `#N` preceded by a repo-name character (letter, digit, `_`, `-`, `.`, `/`) is a FOREIGN
  ref and must not resolve to a local issue. One shared boundary rule, derived once and used
  by all three sites (`ARCH-DRY`) — the current three copies of the same regex are why one
  fix would otherwise miss two paths.
- Go's RE2 has no lookbehind, so the boundary is expressed the way `subjectAnchorRE`
  (`window.go:205`) already handles its lookahead problem: match the preceding character and
  reject it in code.
- **Cross-repo refs are not merely ignored — they are a separate question.** `pair#127` is
  real work on a real issue in another repo; the honest reading is "not attributable to any
  LOCAL issue". Whether the engine should ever attribute across repos is out of scope here,
  and the fix must not foreclose it.

**Retained from the original Spec:** the regression check against #187's window (the 46.1m
returns to #187), no operator-facing hand-correction, and documenting the rule in `atlas/`.
**Dropped:** the commit-boundary-outranks-mentions precedence rule (wrong premise, see
above) and the terminal-status guard (it would have masked this bug rather than fixed it —
ariadne#127 being closed is a symptom of the misparse, not the reason it was wrong; a
same-numbered *open* foreign issue would have slipped through).

## Log

### 2026-07-29
