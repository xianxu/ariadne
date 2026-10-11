---
id: 000202
status: codecomplete
deps: []
github_issue:
created: 2026-08-21
updated: 2026-10-11
estimate_hours:
card_mirror: '11a1f7f8e039df5f52a0f1d9db4d139a04397f38' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T23:14:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.20
---

# A finding has no implementer-side disposition: only the judge can retract

## Problem

`construct/vocabulary/finding.cue` models advisory-ness for **severity** and not
for **authority**.

Severity is handled well. `categories.advisory: ["Minor"]` never blocks;
`hardBlocking: ["Critical"]` means an `Important` finding is demoted past the
round cap; `when."Important"` even reads "fix before the gate **if cheap**;
blocks until disposed". Those are real, and they work.

But the closing dispositions are:

```cue
dispositions: {
  closing: ["addressed", "withdrawn"]
  open:    ["not-addressed"]
}
whenDisposed: {
  "withdrawn": "the judge retracts it (mistaken, or overtaken by a design change)"
}
```

**Only the judge can withdraw.** There is no disposition meaning *the implementer
considered this and declined, with reasons*. A grep for rebuttal / dispute /
agent-note across `cmd/`, `pkg/` and `construct/` returns nothing: there is no
channel for the implementer to answer a finding at all.

AGENTS.md §3 matches: *"Fix Critical/Important before crossing the boundary"* —
imperative, with no "or record a reasoned disagreement".

The doctrine we want already exists in the base layer, in the wrong place.
`superpowers-receiving-code-review` has a whole **"When To Push Back"** section
("Push back with technical reasoning if wrong"). §3 explicitly says NOT to run
those skills at an SDLC boundary — so the pushback doctrine lives precisely where
the gate is not.

### The consequence: findings time out instead of being adjudicated

A **wrong** finding and a **correct-but-not-worth-it** finding leave through the
same door — the round cap — and the ledger does not distinguish them. So the next
round's judge inherits no counter-evidence and can re-derive the same mistake.

Measured in `tools#4`, which ran ten boundary rounds and 41 findings:

- **BR-40** claimed "the byte path is asserted by nothing". False: that
  measurement ran `-run TestPTY -tags conformance`; against the full suite the
  same mutation reddens `TestEditorLoopCtrlCExitsZero` with `exit = 9, want 0`.
  The implementer had no disposition for "disputed, here is the
  counter-measurement", so the rebuttal went into the plan's prose — the channel
  **BR-28 in that same issue had already measured as 0% addressed**. It exited by
  round cap, not by adjudication, and nothing records that it was wrong.
- **BR-33** was accepted with a stated engineering reason (the obvious fix
  collides with a separately-pinned invariant). No disposition for that either;
  it remains listed under "Open findings" at close.

## Spec

Add an implementer-side closing disposition — `disputed` is the working name —
with these properties:

- **It is `closing`, not `open`.** It stops the finding blocking. `finding.cue`'s
  own comment already explains why dispositions are partitioned rather than flat:
  so a new one cannot be accepted by the round validator while the open-set
  computation has no case for it. This change is the shape that comment
  anticipates (it uses `deferred` as its example).
- **It requires evidence, not an opinion.** A rationale is mandatory, and where
  the finding shipped with a measurement, so must the disposition — the rule
  `tools#4` landed in lessons.md is *a finding is closed only when you have
  re-run the measurement that produced it*, and a dispute is the case where
  re-running it produced a different answer. A `disputed` with no counter-measure
  should be a protocol error, exactly as a disposition naming an unknown id
  already is.
- **The next judge must see it.** Carry the rationale into the following round's
  prompt so the judge either re-raises **against the evidence** or withdraws. The
  current failure mode is that a rebuttal is invisible to the next round.
- **It is loud.** Per the escape-hatch principle: every structure needs a bypass,
  and using it must be announced, never silent. It belongs in the churn ledger as
  its own column beside `addressed` / `withdrawn` (`churnreport.go` already treats
  a closing disposition with no column as a defect), and on the operator gate
  line.

## Design risks

- **Abuse.** An agent that disputes everything escapes every gate. This is the
  main reason to require a counter-measurement and to make the count visible per
  boundary: a round that disputes 6 of 8 findings should look alarming on the
  gate line. Consider whether `Critical` should be disputable at all, or whether
  a dispute on a `hardBlocking` finding needs operator acknowledgement.
- **It must not become the lazy exit.** `tools#4` is also the evidence *for* the
  gate: rounds 7-10 caught a 9.6 MB blob, a guard that read the index instead of
  history, and a guard that reported green on a partial scan — all self-inflicted
  by the implementer, all correctly refused. A dispute channel must not make that
  refusal cheaper to dodge. The asymmetry to encode: disputing costs evidence,
  fixing costs a diff.
- **Judge learning.** Worth deciding whether a sustained dispute should feed back
  into the judge prompt at all, or stay per-issue.

## Related

- **#195** — finding-family escalation. Same underlying problem (findings that
  recur without converging) from the other side: #195 is about the judge
  escalating, this is about the implementer answering.
- **#201** — review artifacts capture harness stderr. Also surfaced by `tools#4`.
- `construct/vocabulary/finding.cue` is the single source; `cmd/sdlc/boundaryledger.go`,
  `cmd/sdlc/churnreport.go` and the judge prompt are the consumers that must derive
  from it rather than restate it.

## Done when

- `finding.cue` models a closing, implementer-only `disputed` disposition (an `authority`
  partition) and `appealOnly` (= `hardBlocking`); the judge's fence offers only the judge's
  dispositions, and a judge fence emitting `disputed` is dropped as a protocol error.
- `sdlc finding dispute --issue N --id ID --rationale … --evidence …` records the dispute in
  the gate ledger named by the id's prefix (PQ plan gate, BR boundary review) and commits it
  narrowly on the issue branch; it refuses a blank rationale or evidence, an unknown or
  unissued id, a finding already disputed and not yet answered, a judge-settled finding, a
  foreign owner, and a checkout off the issue branch, writing nothing on any refusal.
- A dispute settles an Important/Minor finding; on a Critical it is recorded but the finding
  keeps blocking until the judge withdraws it. A later judge `not-addressed` re-opens it.
- The next review round's prompt lists each live dispute with its rationale and evidence and
  asks the judge to withdraw or re-raise it against the evidence; it is not listed as settled.
- Disputes are loud: the verb warns; both gates print "N of M finding(s) settled by
  implementer DISPUTE" at warning level on every round; the close's plan-gate line and the
  calibration ledger (appended `gate_disputed`, `boundary_disputed`) count them.
- Every ledger-gate refusal names the dispute verb; help, AGENTS.base.md §3, atlas and README
  describe it.

## Plan

Durable plan: `workshop/plans/000202-a-finding-has-no-implementer-side-disposition-only-the-judge-can-retract-plan.md`
(single boundary — closes in one `sdlc close`).

- [x] Model: `authority` (judge vs implementer), `appealOnly: hardBlocking`, closing `disputed`
- [x] Ledger: `Disputes` stream merged into one round-ordered timeline; judge may not emit `disputed`
- [x] Prompt + decision: next judge answers each live dispute; gate line counts disputes
- [x] Verb `sdlc finding dispute` (evidence mandatory, owner-only, narrow commit, recovery contract)
- [x] Loudness: dispute warning on both gates, `gate_disputed` + `boundary_disputed` TSV columns
- [x] Docs: help, gatefindings routing, AGENTS.base.md §3, atlas, README
- [x] Close: `make test` run — #202 tests green; three environmental reds triaged in `--verified`; `sdlc close`

## Log




- 2026-10-11: closed — BR-1 fixed (dispute text quoted line-by-line as labeled data; TestPromptCarriesLiveDisputes asserts an injection-shaped evidence line is quoted). BR-2 fixed (finding-dispute help: a dispute settles the ledger half; a REWORK verdict still holds the close). BR-3 disputed via sdlc finding dispute with counter-measurement: TestPlanningReviewStaleBeforePersistence/{plan,estimate}/ledger PASS. go test ./cmd/sdlc/internal/gatestate/ green. (fixed-then-ship); review verdict: SHIP
- 2026-10-11: closed — make test: every #202 test green (gatestate dispute/timeline/prompt/decide incl. mutation checks on appeal-only, timeline ordering, judge-only fence; vocab authority partition; finding dispute e2e records/commits/refuses/rerun/ownership; per-site refusal routing e2e; calibration columns e2e through runClose). Suite reds are environmental, each rerun: processgroup TestCancellationKillsDescendants = sandbox denies /bin/ps; TestPlanningReviewConcurrencySchedules = 2s budget at load avg 30, passes alone; hermeticity lock leak = shared common-dir .git/sdlc.lock held by another slot, not reproduced by the #202 tests alone.; review verdict: SHIP
- 2026-10-11: flow upgraded quick → full — 648 added lines in code files (limit 100)
- 2026-10-10: claimed (ops:0 dispatch); design in the durable plan. Critical is appeal-only
  (dispute recorded and shown to the judge, keeps blocking) rather than refused or
  operator-acknowledged — an ack flag is a bypass the agent can pass itself.
- 2026-10-10: change-code inferred the QUICK flow (design under 500 lines). The code is past
  the 100-line shell, but `--flow` is the operator's pin, so I left it: close upgrades to the
  full boundary review once the diff leaves the shell. No plan-quality round or estimate ran.
- 2026-10-10: built in six commits (model, ledger, prompt/decision, verb, loudness, docs).
  Design deltas from the plan: the dispute line lives on `Decision.DisputeLine()` and is
  printed only by `stampAndPersist` (in Reason too, it printed twice); `AdmitDispute`
  (not `DecideDispute`) is the verb's pure admission check. Every ledger-gate refusal
  (plan-quality block, boundary SHIP-with-open-blocking, boundary REWORK) routes to the
  dispute verb, each pinned end-to-end (`TestLedgerGateRefusalsRouteToDispute`).
  Mutation-checked: appeal-only settle rule, timeline `<` vs `<=`, judge-only fence list.
- 2026-10-11: full `make test`: all #202 tests pass. Three reds, each environmental and
  rerun: processgroup `TestCancellationKillsDescendants` (sandbox denies `/bin/ps`);
  `TestPlanningReviewConcurrencySchedules` (2s budget at load average ~30, passes alone);
  a hermeticity lock "leak" (`.git/sdlc.lock` lives in the common dir every slot shares;
  not reproduced by the #202 tests alone).
