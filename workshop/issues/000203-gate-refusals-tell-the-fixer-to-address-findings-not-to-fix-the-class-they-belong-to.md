---
id: 000203
status: working
deps: []
github_issue:
created: 2026-08-22
updated: 2026-08-22
estimate_hours: 1.38
started: 2026-08-22T10:04:13-07:00
---

# gate refusals tell the fixer to address findings, not to fix the class they belong to

## Problem

The `family:` mechanism (#194) is fully built on the *reviewer's* side and
absent on the *fixer's*. It has three specified consumers:

1. **the reviewer** — `judge/prompts/milestone-review.md` tells it to slug the
   underlying RULE not the symptom, reuse prior slugs verbatim, and (per
   `helptext/close.md`) escalate on a repeat from "fix this instance" to "state
   the rule that covers all of them";
2. **the ledger** — persists families and computes repeat counts;
3. **the operator** — the convergence line, `Not converging: fix rules, not
   instances.`

Nobody tells the agent *receiving* the findings to do anything differently. What
it actually reads at the moment it starts fixing is:

    changecode.go:554  address the findings above and re-run …
    close.go:1194      address them, or have the review dispose them explicitly …
    close.go:1237      address the findings, then re-run `sdlc close`
    close.go:1809      1. Fix the findings NOW, before committing this close.

"Address them" reads as "address each of them", which is precisely the per-site
patching the family machinery exists to detect. The convergence line is a
*diagnosis* emitted after the round, not a *procedure* offered before it — by
the time you read "Not converging" you have already spent the round.

**Evidence (parley.nvim#202).** Four boundary rounds against a cap of three,
with two families each surviving three separate rounds:

| family | findings | rounds |
|---|---|---|
| `invariant-without-regression-guard` | BR-2, BR-8, BR-13, BR-17 | 1, 1, 3, 4 |
| `stale-restatement-of-moved-source` | BR-4, BR-5, BR-9, BR-16, BR-18 | 1, 1, 1, 3, 4 |

Each round the agent fixed the site the finding named and stopped. Round 1's
BR-2 ("this invariant has no executable guard") would have retired BR-13 and
BR-17 too, had the agent enumerated *every invariant it had asserted in prose*
instead of guarding the one named. The rule was available — it was sitting in
the `family:` field — and was read as a label rather than as a worklist.

The tendency also shows one stage earlier: #202's first plan was the narrow
per-spec patch, inherited uncritically from the issue's own Done-when. The
general fix came from the plan gate (PQ-5), not from the planning agent.

## Spec

- The discipline — **a finding names one instance; the deliverable is the class:
  name it, enumerate it, sweep it in the same round** — has exactly ONE
  statement, in `judge/architecture.md`'s `ARCH-PURPOSE` entry, whose `at-review`
  lens already carries the neighbouring "shadow-sweep: enumerate the consumers"
  discipline.
- Every surface that hands findings to the fixer **routes to that statement and
  cites the marker**. None paraphrases the procedure. This is the #128 pattern,
  already binding here: the constitution stopped restating ARCH-* definitions and
  now routes to `sdlc arch-principles`, guarded by
  `TestArchitecture_NarrativeRoutesToArchPrinciples`.
- The routing line itself has one owner too — eight hand-maintained citations
  would be the same defect one level down.
- A marker citation is safe at these sites specifically: `ArchitectureBlock`'s
  contract warns that "a marker alone would be a dangling pointer in a
  fresh-context subagent", but gate refusals are read by the **main thread**,
  which already received the block from `sdlc start-plan` and can pull it any
  time via `sdlc arch-principles` — the documented non-gate PULL.
- The reviewer-facing side is untouched; it is already correct.

## The class (enumeration)

Enumerated **mechanically**, not by hand — every non-test `cwarn`/`die` in
`cmd/sdlc` whose message reports judge findings and directs the reader:

    grep -rn "cwarn(\|die(" cmd/sdlc/*.go | grep -v _test | grep -iE "finding|fix the|address"

**In class — eight sites:**

| site | enclosing func | context |
|---|---|---|
| `changecode.go:554` | `runPlanQualityJudge` (416) | plan gate blocked, ledger path |
| `changecode.go:572` | `classifyFallback` (563) | plan-quality, verdict-token path |
| `changecode.go:722` | `runEstimateQualityJudge` (664) | estimate-quality findings |
| `close.go:1194` | `finalizeBoundaryReview` (1144) | boundary gate, open blocking |
| `close.go:1237` | `finalizeBoundaryReview` (1144) | boundary review REWORK |
| `close.go:1809` | `formatFixThenShipProtocol` (1806) | FIX-THEN-SHIP step 1 |
| `milestoneclose.go:626` | `dispatchBoundaryReview` (590) | boundary review findings |
| `judge.go:172` | `runJudge` (69) | ad-hoc `sdlc judge` |

**Out of class, with reasons** — the scan surfaces them; none reports judge
findings to fix: `changecode.go:251` (deterministic structural failures),
`changecode.go:279` (estimate arithmetic), `changecode.go:495` +
`boundaryledger.go:161` (*no* findings block parsed — a prompt/parse fault),
`boundaryledger.go:101` (ledger unreadable), `close.go:1180` (bypass notice),
`close.go:1190` (states non-finalization; `:1194` carries the instruction),
`migrate.go:308` (ref resolution).

Doc surfaces: `helptext/close.md` FINDING FAMILIES and `helptext/change-code.md`
THE PLAN GATE.

The first read of this issue found two sites; a hand enumeration found seven and
was still wrong; the mechanical scan found eight. Recorded because it is the
issue's own thesis: an enumeration written from memory is another instance.

## Done when

- [x] `ARCH-PURPOSE` carries the class-vs-instance discipline in its
      `principle`, `at-plan`, and `at-review` lenses — the single statement.
- [x] One pure formatter owns the routing line, and **all eight** code sites
      emit it. No site paraphrases the procedure.
- [x] A **drift guard** mirroring `TestArchitecture_NarrativeRoutesToArchPrinciples`:
      the emitted line routes to `sdlc arch-principles` and cites `ARCH-PURPOSE`,
      and the registry entry still carries the discipline the citation promises.
      Restating the procedure in the formatter would make this test vacuous, so
      it asserts routing, not wording.
- [x] An **enumeration guard** — a Go test scanning non-test `cmd/sdlc` sources
      for the class signature, asserting every match routes through the formatter
      or sits on an explicit, reasoned exclusion list. A table over the eight
      known funcs cannot deliver this: it is structurally blind to a ninth,
      exactly as the first draft of this plan was blind to four of the eight.
- [x] `helptext/close.md` FINDING FAMILIES states the fixer's half by routing to
      the marker; `helptext/change-code.md`'s plan-gate section does the same.
- [x] Helptext render/golden tests pass and the full Go suite is green.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module     design=0.04 impl=0.16
item: greenfield-go-module  design=0.10 impl=0.24
item: atlas-docs            design=0.20 impl=0.06
item: milestone-review      design=0.00 impl=0.14
item: scope-pivot           design=0.25 impl=0.10
design-buffer: 0.15
total: 1.38
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* (`design-buffer: 0.15` is a rate, not hours.)

Derivation:

- **smaller-go-module** — `gatefindings.go`'s formatter plus wiring eight call
  sites: extending an established pattern (`gatepersist.go`), not inventing one.
  v2 design pick 0.2 of 0–0.3, ×0.2 Step-3 (the plan fixes the file, the header
  rationale, and the eight targets) → 0.04. v2 impl pick 0.4 of 0.2–0.5, ×0.40
  per v3.1 → 0.16.
- **greenfield-go-module** — the two guards, of which the enumeration guard is a
  genuinely new single-concern thing: a source scan over non-test `cmd/sdlc` with
  a reasoned exclusion list. v2 design pick 0.5 (low end of 0.5–2 — the plan
  states the mechanism), ×0.2 → 0.10. v2 impl pick **0.6 of 0.3–0.8, above the
  midpoint**: I have read this repo's Go but not written in it, and the false-
  positive boundary between a findings refusal and a parse-fault message is
  fiddly. Taken here rather than as a blanket ×1.5 familiarity, which would
  double-count against the four items that *are* familiar. ×0.40 → 0.24.
- **atlas-docs** — `ARCH-PURPOSE`'s three lenses, `helptext/close.md`,
  `helptext/change-code.md`, atlas. Design 0.20 at the **top** of 0.05–0.2 and
  **undiscounted**: the registry entry is the single source every consumer routes
  to, so its wording is the deliverable, not a restatement of one. Impl 0.15
  ×0.40 → 0.06.
- **milestone-review** — the one mandatory close review. Design 0.0; impl pick
  0.35 of 0.2–0.5, ×0.40 → 0.14.
- **scope-pivot** — round 1's plan gate (PQ-1 found my hand enumeration missed
  four of eight sites) plus the operator's ARCH-DRY redirect, which moved the
  design from "state the rule in two places" to "state it once in ARCH-PURPOSE
  and route". Both landed inside the `sdlc actual` window (`started: 10:04`), so
  they are priced. v2 design pick 0.25 of 0.2–0.5 undiscounted — the revision
  *was* the design decision; impl 0.25 ×0.40 → 0.10.
- **Step 2.5 (library availability)** — n/a: no external dependency; the change
  is entirely internal wiring and prose.
- **Step 5 familiarity ×1.0** — the codebase area was enumerated mechanically
  before estimating, and four of five items extend patterns already in the tree.
- **Step 6 buffer +15%** — Step-3's ×0.2 discount was applied to the two code
  primitives, so the v2.1 rule of thumb halves the buffer.

Recompute: (0.04+0.10+0.20+0.00+0.25) × 1.15 + (0.16+0.24+0.06+0.14+0.10) × 1.0
= 0.59 × 1.15 + 0.70 = 0.6785 + 0.70 = **1.38**

## Non-goals

- **A new `ARCH-*` marker.** `ARCH-PURPOSE` already owns "deliver the purpose,
  not the easy subset" and an enumeration discipline (shadow-sweep). A finding's
  named site *is* the easy subset of its class, so this is the same principle,
  not a new one — and a fifth marker would split it.
- **Restating the procedure at the eight sites or in helptext.** That is the
  defect this issue is about, one level down, and #128 already ruled against it
  for the constitution.
- **Rendering the full ARCH-PURPOSE entry at each refusal.** ~10 lines on every
  gate refusal drowns the findings it is meant to frame; the main thread already
  holds the block and can pull it.
- **Changing the reviewer prompt or the family slugging rules.** They work;
  parley.nvim#202 proves it — the ledger diagnosed non-convergence correctly and
  on time. The defect is downstream of the diagnosis.
- **Automated enforcement of the sweep itself.** Whether an enumeration was
  actually swept is not mechanically checkable. This installs the routing at the
  moment of use; the ledger's repeat-family counter already measures compliance.
- **`AGENTS.base.md`.** It already routes to `sdlc arch-principles` and retains
  marker awareness (guarded), so extending `ARCH-PURPOSE` reaches it for free —
  no fleet-wide prose change needed. That is the payoff of routing over
  restating.

## Plan

- [x] Extend `ARCH-PURPOSE` in `cmd/sdlc/internal/judge/architecture.md`: the
      class-vs-instance discipline into `principle`, `at-plan` (flag a plan that
      fixes a named instance where the class is enumerable), and `at-review`.
- [x] Write both guards FIRST, red: the drift guard (routing + marker + registry
      entry) and the enumeration guard (source scan over non-test `cmd/sdlc`).
      The enumeration guard must fail on the current tree naming all eight sites.
- [x] Add `gatefindings.go` — one pure formatter owning the routing line, header
      in the house style naming why it is a file and not eight strings
      (cf. `gatepersist.go`, whose tail diverged five times before extraction).
- [x] Wire the eight sites: `runPlanQualityJudge`, `classifyFallback`,
      `runEstimateQualityJudge`, `finalizeBoundaryReview` (both arms),
      `formatFixThenShipProtocol`, `dispatchBoundaryReview`, `runJudge`.
      Existing seams: `changecode_test.go` drives `runPlanQualityJudge`
      directly; `finalizeBoundaryReview` is reached via `runCloseWithReview`
      (`close_finalize_test.go`, `close_ledger_test.go`).
- [x] Route `helptext/close.md` FINDING FAMILIES and `helptext/change-code.md`
      THE PLAN GATE to the marker.
- [x] Full Go suite green; helptext render/golden tests pass.

## Log

### 2026-08-22

Filed from a parley.nvim#202 retrospective: the user observed a tendency to
"keep patching" rather than fix the class, and the #202 gate ledger measured it
(table above). Scope and process chosen by the user: point-of-use plus
`ARCH-PURPOSE`, landed through ariadne's full SDLC.

Judgment call on artifact tier: seven surfaces but well under 100 lines of real
change, and the substance is one sentence with one owner — so the plan lives in
this issue rather than a separate `workshop/plans/` doc. Flagging it explicitly
for the plan gate to rule on.

### 2026-08-22 (implementation)

Landed as designed. `ARCH-PURPOSE` gained the class-vs-instance discipline in all
three lenses; `gatefindings.go` owns the one routing line; all eight sites emit
it; `helptext/close.md` and `helptext/change-code.md` route to the marker rather
than restating the procedure.

**Both guards proven red on their own regression, not just green on HEAD:**

- *Enumeration guard* — `TestEveryFixerFacingSiteRoutes` walks the AST of every
  non-test `cmd/sdlc/*.go`, matches the class signature (says findings exist OR
  need disposing, AND directs the reader), and requires the routing line in the
  same `cwarn`/`die` call — call granularity, not function, because
  `finalizeBoundaryReview` has two arms and a function-level check would pass
  with one routed. Pass 2 covers string builders like
  `formatFixThenShipProtocol`, which never calls `cwarn` itself. Before wiring it
  failed naming exactly the eight enumerated sites; after wiring, appending a
  simulated ninth site to `state.go` — a file none of the eight live in — failed
  with `state.go:458`, so it is not pinned to an instance list.
- *Drift guard* — `TestFixTheClassLine_RoutesToArchPrinciples` asserts ROUTING,
  never wording (asserting wording is what would let the line become a second
  copy of the principle). Stripping the discipline from `ARCH-PURPOSE` fails it
  with `ARCH-PURPOSE no longer carries "CLASS"`.

One reasoned exclusion in the enumeration guard: `fixTheClassLine` itself, whose
own literal matches the signature it describes and cannot route through itself.

**Judgment call flagged for the close review.** Editing `architecture.md` broke
`TestBuildPrompt_Golden`, whose header says "⛔ After the initial capture, NEVER
re-run -update-golden to 'fix' a failure: a failure means a .md drifted — fix the
.md, not the golden." That guard is aimed at the #153 prompts→markdown
*refactor*, where byte-identity is the contract; a deliberate registry edit is
the case it does not cover (and could not, or #71's ARCH-SHIM could never land).
I re-captured, then verified the diff is exactly the twelve added ARCH-PURPOSE
lines across the four injected prompts and nothing else. The ⛔'s wording not
covering deliberate registry change is a real gap, but it belongs to a different
class than this issue's, so it is noted here rather than swept into scope.

Full Go suite green (`go build ./... && go test ./...`).
