---
id: 000203
status: codecomplete
deps: []
github_issue:
created: 2026-08-22
updated: 2026-08-22
estimate_hours: 1.38
started: 2026-08-22T10:04:13-07:00
actual_hours: 1.57
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

**The enumeration is computed, not listed here.** An earlier revision of this
section carried the eight in-class sites, the seven reasoned exclusions, and the
doc surfaces as tables — and they went stale within one round (BR-8: the doc line
named two surfaces after the fix had routed five across three files). A
hand-maintained restatement of a class the guards now compute is a deferred
consumer, which is the same defect this issue exists to close. So this section
routes, per the #128 pattern the issue applies everywhere else:

- **The surface set and the ruling on each member** — `fixerFacingSurfaces` in
  `cmd/sdlc/gatefindings_test.go`. It names what is guarded, what is ruled out,
  and why.
- **The in-class sites** — whatever `TestEveryFixerFacingSiteRoutes` and
  `TestEveryFixerFacingHelptextRoutes` match today. Run them; a violation prints
  the list.
- **The exclusions** — structural and stated at each guard (the routing line's
  own definition; >=6-space quoted output; reference sections).

What is worth recording here is the *evidence*, which does not drift: the first
read of this issue found **two** sites; a careful hand enumeration found **seven**
and was still wrong; the mechanical scan found **eight**. That gap is the issue's
thesis, and it recurred twice more under review — the doc class got a hand pass
(BR-1) and the guard's own signature was narrower than the class it claimed
across three rounds (BR-2, BR-7).

## Done when

- [x] `ARCH-PURPOSE` carries the class-vs-instance discipline in its
      `principle`, `at-plan`, and `at-review` lenses — the single statement.
- [x] One pure formatter owns the routing line, and every code site emits it.
      No site paraphrases the procedure. (The site list is computed by the
      guard, not restated here — see `## The class`.)
- [x] A **drift guard** mirroring `TestArchitecture_NarrativeRoutesToArchPrinciples`:
      the emitted line routes to `sdlc arch-principles` and cites `ARCH-PURPOSE`,
      and the registry entry still carries the discipline the citation promises.
      It asserts routing, not wording — asserting wording would let the line
      become a second copy of the principle.
- [x] An **enumeration guard** scanning non-test `cmd/sdlc` sources for the class
      signature, asserting every match routes or sits on an explicit, reasoned
      exclusion. It **raises the cost** of shipping an unrouted refusal; it is a
      syntactic over-approximation and does not make it impossible — what it
      approximates is stated at the guard and in `atlas/workflow/gate-state.md`
      (#203 BR-10: a syntactic approximation cannot back an absolute claim).
- [x] Every helptext passage that hands findings to the fixer routes to the
      marker — enforced by `TestEveryFixerFacingHelptextRoutes` rather than by a
      list here, so it cannot go stale.
- [x] The **surface set** is declared with a ruling on each member
      (`fixerFacingSurfaces`), and the canonical findings-reception skill
      generalizes before implementing, with an intent record so
      `/construct upgrade` cannot silently revert it.
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
- **`cmd/doc-review`'s "triage each finding" (`review.go:125`)** — ruled OUT
  (#203 BR-6), not overlooked. In principle it is the same framing, but it is a
  different binary with no family ledger, its review is explicitly advisory and
  read-only over a document the agent owns ("do not apply findings blindly"),
  and the ARCH-* registry is not delivered to it — routing there would mean
  teaching that tool to push or pull the registry. That is a separable
  extension in ARCH-PURPOSE's sense, so deferring is legitimate; stating it is
  the part that is not optional. Worth a follow-up issue if doc-review ever
  grows a ledger. The guard's scan boundary (`cmd/sdlc/*.go`, subpackages
  excluded) is documented at the test.
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
- 2026-08-22: closed — go build ./... && go test ./... green. BR-10 fixed both halves: attribution now COUNTS routing refs per statement (verified red on the exact two-messages-one-ref scenario in formatFixThenShipProtocol), and - the more important half - the CLAIM is fixed rather than the mechanism widened a fifth time: atlas/workflow/gate-state.md and Done-when now state what the guard approximates (statically visible literals only, + chains only, per-statement counting that cannot tell which message a ref is joined to) and say it raises the cost of an unrouted refusal without making it impossible. BR-11 fixed at the source, not the render output: construct/intents/superpowers.md Conversation 8 records the receiving-code-review change with Verify clauses per the a547179 precedent, so /construct upgrade cannot silently revert it. Also restored the ## Done when section that the prior commit span-edit swallowed.; review verdict: FIX-THEN-SHIP

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

### 2026-08-22 (close review round 1 — all six findings addressed)

Both blockers were this issue's own thesis turned back on it, so both got the
class treatment rather than a patch each.

- **BR-1 (Important) — I enumerated the code class mechanically and the doc class
  by hand.** Exactly the defect #203 installs a guard against, committed inside
  the issue about it. Fixed at the class: `TestEveryFixerFacingHelptextRoutes`
  now scans `helptext/*.md` the way the Go guard scans sources. Granularity is
  the paragraph, routed if it or the next paragraph cites the marker — matching
  how these files are written (statement, then elaboration). Two structural
  exclusions, both reasoned rather than line-keyed so they cannot rot: blocks
  indented >=6 spaces (quoted tool output — this is what correctly keeps
  `close.md`'s convergence-line examples out, since they *quote* the rule rather
  than instruct anyone), and `referenceSections` (FLAGS/USAGE/DEFERRED/… — prose
  describing the tool, which is why `judge.md`'s `--tools` note and its roadmap
  mention findings without handing any to anyone). The scan then found exactly
  the three genuine sites, which are now routed: `close.md:13`, `close.md:59`,
  `milestone-close.md:38`. BR-3 (unguarded helptext citations) is disposed by the
  same guard.
- **BR-2 (Important) — the guard's signature was narrower than the class it
  claimed.** All four blind spots closed: (a) a `cwarn`/`die` call's literals are
  now **concatenated before matching**, since the tree's prevailing style splits
  one message across adjacent literals — `close.go:1194` is three and
  `changecode.go:554` became two *in this diff*; (b) pass 2 walks **every**
  literal including package-level const/var, not just `FuncDecl` bodies; (c) the
  directive vocabulary is widened (a narrow signature being the same defect one
  level down); (d) pass 2 is **count-based** — routing refs >= matching literals
  per func — because function-granularity is what pass 1's own comment already
  declared insufficient.
- **BR-4 (Minor) — the scans assert the source routes; neither proved the line
  reaches an operator.** A formatter whose output were dropped or swallowed by a
  wrapper would pass both scans and reach nobody. `TestGatePathStderrCarriesRoutingLine`
  drives `runPlanQualityJudge` end-to-end with a findings verdict and reads
  stderr. Proven meaningful: removing the routing from `classifyFallback` fails it
  with "gate stderr does not carry the routing line".
- **BR-5 (Minor)** — done earlier: `fixTheClassNote()` owns the `"\n  "` join the
  six one-line sites had hand-spelled.
- **BR-6 (Minor)** — scan boundary documented at the test (`cmd/sdlc/*.go`,
  subpackages excluded, with why); `cmd/doc-review`'s sibling framing ruled OUT
  in Non-goals with its reasoning, not dropped silently.

Correction to the parked continuation: it recorded "BR-4 already done" — that was
**BR-5**. BR-4 was open and untouched, and is fixed here.

Full Go suite green (`go build ./... && go test ./...`).

### 2026-08-22 (close review round 2 — the rule, not two more shapes)

Round 2 disposed all six of round 1's findings but raised a **third round of
`guard-narrower-than-claimed-class`**, and the gate said so: *"Not converging:
fix rules, not instances."* It was right — rounds 1 and 2 had each widened the
guard to whatever shape the last finding named. BR-7 named two more (F: a split
message in a non-`cwarn` call, of which the tree has ~200 candidates; H: a second
unrouted line in an already-routing func, which `finalizeBoundaryReview` already
carried twice) and, correctly, refused to let me patch them.

**The rule, now stated and implemented:** match every fixer-facing message as a
whole string-valued **expression** — folding `+`-joined chains wherever they
occur, not inside two hardcoded emitter names — and attribute each match to a
routing reference in its own **statement**. Both halves collapse a family of
shapes the earlier drafts handled one at a time: per-literal matching misses any
message the tree splits across pieces (its prevailing style), and per-function
attribution credits one routing reference to every line in the func.

The redesign deleted `stringLiterals`, `enclosingFunc`, `countMatchingLiterals`,
`countRoutingRefs`, and the whole two-pass/`claimed` structure — a shorter guard
that is strictly stronger. It immediately caught shape H in the tree
(`formatFixThenShipProtocol` built its message in one statement and appended the
routing line in the next); fixed by merging into one `append`, which is the
honest shape anyway — the routing line should share a statement with the message
it routes. Both shapes re-verified red afterwards on synthetic sites.

- **BR-8 (Minor) — the issue's own enumeration was hand-maintained** and went
  stale within one round (named two doc surfaces after the fix routed five across
  three files). Fixed at the rule rather than by correcting the list: `## The
  class (enumeration)` now ROUTES at the guards, keeping only the evidence that
  does not drift. Done-when 5 likewise.
- **BR-9 (Minor) — the SET of surfaces was never declared**, so siblings were
  ruled in or out from memory. `fixerFacingSurfaces` now declares it with a
  ruling on each member. The finding it surfaced is the sharpest in the issue:
  `construct/adapted/superpowers-receiving-code-review/SKILL.md` — the canonical
  findings-*reception* surface, invoked at exactly the moment a gate hands
  findings over — ended with *"6. IMPLEMENT: One item at a time, test each"*,
  the per-site patching this issue exists to stop. It escaped every scan because
  it says "feedback"/"items", never "findings". Its response pattern now has a
  **GENERALIZE** step before IMPLEMENT and routes to ARCH-PURPOSE, pinned by
  `TestReceivingCodeReviewSkillGeneralizes`.

Full Go suite green.

### 2026-08-22 (close review round 3 — stop widening, fix the claim)

- **BR-10 (Important) — 4th in the family, and not a new shape.** It was the
  half of BR-7's own stated rule that round 3 dropped: *"count only unclaimed
  refs."* I implemented statement attribution as **membership** rather than
  **counting**, so one statement carrying two fixer-facing messages and one
  routing reference passed. Reachable in `formatFixThenShipProtocol` precisely
  because round 3's own fix put the reference into that append. Counting is a
  ~5-line change and is done; verified red on BR-10's exact scenario.

  The more important half is the reviewer's: **a syntactic approximation cannot
  back an absolute claim.** Counting closes this residue; the next (two
  references both joined to one message) is one level finer, without bound. So
  the claim is fixed rather than the mechanism widened a fifth time.
  `atlas/workflow/gate-state.md` now states what the guard *approximates* —
  statically visible literals only, `+` chains only, per-statement counting that
  cannot tell which message a reference is joined to — and says plainly that it
  raises the cost of shipping an unrouted refusal without making it impossible.
  Done-when 4 matches.
- **BR-11 (Important) — my BR-9 fix landed at a generated consumer.**
  `construct/adapted/` is render output: `/construct promote` is delete-then-copy
  and `/construct upgrade` re-renders from `construct/sources/` +
  `construct/intents/superpowers.md`, where a skill not named in the intent is
  copied from source as-is. `receiving-code-review` appears nowhere in that
  transcript, so the GENERALIZE step, the ARCH-PURPOSE routing, and the
  `family:`-as-worklist paragraph were all scheduled for deletion by the
  substrate's own pipeline. This is the issue's own principle one level out — the
  deliverable landed at the compiled consumer instead of the source it derives
  from. Added **Conversation 8** to `construct/intents/superpowers.md` with Verify
  clauses, following the exact precedent the reviewer cited (a547179, #71, the
  last ARCH-registry change to touch an adapted skill).
- **Mistake of mine, caught and fixed here:** the BR-8 edit replaced the span
  from `## The class` to `## Estimate` — which silently swallowed the entire
  `## Done when` section. Recovered from `99bb7a5` and rewritten to match the
  current state. A structural section vanished for one commit; worth remembering
  that a span-based edit is only as safe as the section boundary you assume.

Full Go suite green.

### 2026-08-22 (close round 4 — gate CLEARED; advisories fixed under FIX-THEN-SHIP)

Boundary gate passed after 4 rounds: no open blocking findings, 4 advisories
recorded. Fixed here before the close commit, per the #174 protocol.

- **BR-13 (Minor) — 3rd in `undocumented-scan-boundary`, and it said outright
  "do NOT rule this instance and stop."** `superpowers-requesting-code-review`
  says *"Fix Critical issues immediately"* — per-item, in a skill live at
  `.claude/skills/` and kept in play by AGENTS.md §3, escaping the doc scan by
  the same mechanism as its sibling ("feedback"/"issues", never "findings").
  Fixed the instance (class-first, routed to ARCH-PURPOSE, guarded by
  `TestRequestingCodeReviewSkillGeneralizes`, with **Conversation 9** in the
  intent so `/construct upgrade` cannot revert it) — but the *rule* is BR-10's
  remedy one level up: **sites are computed, the surface set is not.** Both the
  `fixerFacingSurfaces` header and the atlas now say so, and neither claims "the
  whole surface set" any more. Measured prevalence the reviewer supplied: 14
  adapted skills, exactly 2 carry fixer-facing directives, the declared set named
  1.
- **BR-14 (Minor).** `referencesRoutingLine` — the membership predicate BR-10
  named as the defect — survived directly below its counting replacement with
  zero call sites. Round 3's Log claimed the redesign deleted the whole two-pass
  structure; this was the piece that didn't. Deleted.
- **BR-12 (Minor).** The routing line was 120 chars (129 at the FIX-THEN-SHIP
  indent) against ~83 for its neighbours. Now 88 on one line. A post-colon split
  was rejected: the continuation indent differs per site, which is why
  `fixTheClassNote` exists.
- **BR-3 (Minor).** Covered rather than separately guarded: the doc scan requires
  the citation in every fixer-facing helptext passage, and
  `TestFixTheClassLine_RoutesToArchPrinciples` requires the destination to still
  carry the discipline. Between them a citation cannot go stale in either
  direction.

Estimate 1.38h / actual 1.57h (ratio 0.9x, trusted window). Four lessons added to
`workshop/lessons.md`, including the two the reviewer effectively taught: bound
the claim rather than widening the mechanism, and state which half a guard
computes.
