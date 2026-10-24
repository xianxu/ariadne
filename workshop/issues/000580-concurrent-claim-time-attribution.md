---
id: 000580
status: codecomplete
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e2ad286b9440d33a9b74e8e2b51c1b6b8a6f3659' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T19:27:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.45
---

# Active-time attribution cross-counts concurrent same-slot claims instead of active branches

## Problem

On2026-10-23, parley.nvim:3 held concurrent claims for parley.nvim#418 and parley.nvim#322. Actual work stayed on #418's issue branch until its landing at14:29 and hub dispatch to #322 at14:30. A fresh main-based #322 branch began then. At14:42, sdlc close --issue322 proposed4.59h measured actual, including a dominant203.5m/74% segment11:18–14:42. That segment includes #418's targeted flake classification, queued verification and landing, not #322 implementation.

Observed measurement text: window2026-10-23T00:22:39-07:00 through2026-10-23T14:42:03-07:00; provenance active-time-v3-engagement-v1;1 ownership interval;5 admitted sessions; dominant #322203.5m/74%11:18–14:42. Code candidate19dda8d9339507f2cfd8b368a8505a25d133e023; tracker5dcff7790ce1157ff18229a5c7d019573a3befcd. The reporter stopped the close review before finalization; tracker remained working with no completion and no actual adopted. TL35643b0f directed --no-actual, no typed hours, and this unclaimed filing.

The printed exclusions already removed some segments adopted by #418's earlier closes. Thus this particular capture directly proves wrong-issue attribution of #418's post-close tail, not that the same exact interval was persisted into both final cards. Concurrent claims must not cross-count one slot's work or assign an unadopted tail to the other claim merely because its ownership interval remains open.

## Spec

Attribute each interval to the issue whose branch the slot was actually working on, while retaining verified ownership/session provenance and disjoint adoption. A second held claim is a reservation, not evidence of active work. Preserve unknown branch evidence explicitly rather than choosing a concurrent claim. Branch transitions, post-close work, queued tests and the final landing tail must remain associated with the actual owning issue until that work ends. This is an attribution repair in Ariadne, separate from Parley's product fix.

## Done when

- A deterministic fixture with two simultaneous same-slot claims, activity on A through close/test/landing, then a main-based branch for B, attributes the intervals exactly once to their witnessed active issue. It reproduces the observed A-tail-to-B failure without the correction. Close/actual/active-time share the corrected classifier; previously adopted intervals remain excluded and ambiguous branch evidence is reported rather than guessed. No manually typed time totals are used to hide the defect.

## Plan

Investigate branch/engagement evidence in active-time-v3 and build the same-slot two-claim regression before changing classification.

Approved implementation and verification: [concurrent claim attribution plan](../plans/000580-concurrent-claim-time-attribution-plan.md). Policy and version approvals are recorded in its Revisions; recorded pre-close verification and mandatory review remain before landing.

## Log

Filed unclaimed from parley.nvim:3 at TL request. Product candidate is parley.nvim PR344; #418 landed as PR338. Relevant hub dispatch1341251d at14:30:52 authorizes starting #322 after #418. Clean #322 implementation began14:30; #418 tail of concern is11:18–14:30. No measurement-source repair was attempted.

2026-10-23: Claimed and started in ariadne:3 after #574 landed. Confirmed the structural cause: EligibleSources admits a persistent root session over every matching owned worktree window, with no branch discriminator; resolveOwnedActualInputsFor retains those windows and the engine's issue-scoped commits can assign A's unadopted tail to B. Existing adopted-credit exclusions cannot prove active ownership of that tail. Own native HEAD reflog confirms per-worktree checkout transitions exist, including issue-to-resting and fresh issue-branch starts; full operation grammar and continuity qualification are planned, not claimed proven. The issue contains a reported incident summary, not raw captured terminal bytes. Proposed a pure shared branch/claim interval selector with stable private HEAD-history capture, saved-window replay and an explicit sole-claim planning compatibility policy. No code, fixtures or production measurement changes yet; plan goes to the hub.

2026-10-23 19:32 PDT: Authenticated TL approved the branch/sole-claim/unknown-concurrency policy and the shared captured-history selector with additive provenance and immutable replay. Approval recorded in plan Revisions; entering change-code before test/implementation work.

2026-10-23: change-code entered (quick inferred from design/no milestones so far; full close review will apply if the implementation exceeds the shell). Added TestOwnedActualConcurrentClaimDoesNotCreditForeignBranchTail through the actual production resolver with real tracker claims and a private worktree. RED in 2.830s: B receives its whole 00:02–00:31 claim window although A remains the active branch through code, close and unadopted verification/landing tail; B's main-based branch starts at 00:20. This is a labelled reconstruction of the issue report, not a raw terminal replay. No production code changed. Next: native HEAD-history grammar/continuity qualification, then pure selector and bounded IO integration.

Pure selector controls now pass (0.150s): exact concurrent A/B partition, no credit without ownership, sole-claim planning labels, ambiguity beginning at the second claim, preserved engagement origin, invalid branch identifiers, and counted 8/16 transition sweeps with disjoint/exhaustive selection. A native private Git probe confirmed worktree-local HEAD logs and rename/checkout records; initial worktree creation does not itself name the branch. Vocabulary inspection resolved the terminal-owner distinction: terminal cards keep attribution identity but are not reservations. A new terminal-owner control first failed, then passed after separating reservation cardinality from historical owner evidence. Plan Revisions records these findings. Capture/parser, tracker IO, snapshot provenance and production wiring remain unimplemented; the production incident fixture is intentionally still RED. No suite, review, private runner or other background work is active.

Pure HEAD-history decoder plus native worktree controls are green (0.721s). Native RED identified missing-tab creation and paired rename records, now normalized only in those exact forms. Decoder also refuses partial records, object-chain gaps, reversed timestamps and contradictory branch/tip witnesses, and preserves detached/unsupported intervals as unknown. Found and corrected a plan assumption: active-time currently has no ownership route. Authenticated TL approved the default tracked-repo semantic change (explicit dirs/time bounds only narrow eligibility; multi-issue intervals partition before allocation; no-tracker behavior unchanged), now recorded in Revisions before wiring. Production incident remains RED; no production resolver integration or full suite has run.

### 2026-10-23 20:03 PDT — production resolver branch selection
- 2026-10-23: closed — Two-claim production resolver and real-command fixture partition A 17min/B 11min; disabling selection reproduces foreign tail attribution. BR-1 saved-replay canonical-kind contradictions fail before the fix and reject after it. Corrected private executable conformance 8.194s, focused race 8.122s, Go catalog 1.392s; prior broad owned/replay controls pass. Fresh final-code pre-close suite queued; its pass remains required before landing.; review verdict: SHIP
- 2026-10-23: flow upgraded quick → full — 1558 added lines in code files (limit 100); an earlier round of this close already ran the full review

The incident resolver regression is GREEN (3.738 s), retaining A’s unadopted tail exclusion and B’s switch boundary. Added expired-log refusal and private native capture controls for separate worktree logs, truncated/malformed/oversized history and cancellation; combined targeted controls pass (5.441 s). Immutable tracker event reads now retain lifecycle status, with the existing engagement derivation unchanged (tracker targeted tests 2.371 s). Source selection is wired before measurement persistence; additive audit provenance, standalone command wiring and the rest of the verification matrix remain unfinished. No suite, close or PR yet.

### 2026-10-23 20:12 PDT — command, dispatch and compatibility checkpoint

Tracked active-time now routes through the same resolver before allocation; the shared-source incident command control gives A 17 minutes and B 11 minutes, and wide explicit raw options cannot bypass missing branch evidence. Counted 8/16 request controls ensure repeated source/event use does not re-read each card generation or re-capture a worktree. Existing no-tracker command tests pass. Combined targeted controls pass (7.322 s). The broader owned-measurement run found contributor pruning dropped new dispatch windows and incomplete unrelated legacy histories were treated as global absence/refusal; both fixed without changing historical engagement derivation. The three affected production consumer/queue controls now pass (23.303 s).

TL approved a new method/version for newly branch-classified snapshots (receipt 7398f142-811e-4267-bc4d-30d5c623c953), because an older engine ignoring original dispatch windows silently undercounts crossing spans. Recorded in plan Revisions. Version wiring, narrowed-option controls, remaining native/adversarial tests, docs, suite and close are still unfinished. No new snapshot has been adopted; no suite, PR or live mutation.

### 2026-10-23 20:35 PDT — version-3 implementation and targeted verification

Recorded the approved refusing method/version boundary in help, README and atlas. New actual/close/manual candidates share branch selection; original dispatch and selected windows are retained and checked against branch witnesses. Native unsupported symbolic-ref is fenced, and neither a later raw current ref nor expired live log rewrites an adopted result. Explicit manual limits remove only out-of-window uncertainty; wide raw inputs still cannot bypass proof. Three-harness parent/child lineage uses one captured owner root, not a child brain cwd.

Verification: broader owned/queue/hybrid/command/replay controls PASS 128.536 s; all activetime, gatestate and tracker tests PASS 14.131 s / 0.348 s / 47.206 s; static Go catalog PASS 1.391 s; focused pure/codec race checks PASS 2.919 s / 1.209 s; targeted command/native/hybrid race checks PASS 16.523 s. Tamper controls went red when a foreign branch witness or widened selected window was accepted, then green with shared saved-proof validation (production incident, Qoder close/replay and shared consumer checks PASS 26.168 s). The symbolic-ref probe’s initial failure was a refuted fixture premise, not a production defect; its corrected negative control passes. No recorded suite, PR or adoption yet. Next: private executable/overlay proof, pushed-head pre-close suite, mandatory close review and landing.

### 2026-10-23 20:40 PDT — private conformance and independent red control

Private Go 1.27.3 executable at Git revision 1b97d456581241c25e2bdbab03b16b1765596ae2 has modified=false and passes the real active-time command on the isolated two-claim fixture (7.451 s). A Go overlay that disables only branch narrowing reproduces B crediting A's unadopted tail; expected-red result and raw log are retained under the common Git directory sdlc/test-stats/evidence/000580/selector-disabled-result.json and selector-disabled.log. No tracked source was mutated by this proof.

A final pure control found sole-claim fallback absorbing a witnessed foreign canonical issue branch when that issue had no owner witness. Corrected the selector to exclude a witnessed foreign issue branch regardless of reservation state; ownership/lineage remain prerequisites to credit it to its own issue. That red/green control and the selector suite pass (0.244 s). The compatibility boundary and original-dispatch behavior are unchanged. Rebuilding the private binary for this final source, then the recorded pushed-head pre-close suite and mandatory close review. No live data was adopted or changed.

### 2026-10-23 20:45 PDT — pushed implementation and verification checkpoint

PR 425 is open. The final private executable at Git revision df0f0a5166db83f7b297cbf9fb46ff38b5f58589 (Go 1.27.3, modified=false) passes the isolated real-command incident conformance in 18.725 s. The independent selector-disabled overlay remains an expected RED, with original tail-misattribution evidence retained under sdlc/test-stats/evidence/000580 in the common Git directory. Native capture controls additionally prove symlink aliases select the same worktree history and a removed worktree cannot fall back to the primary log (PASS 1.248 s). These last edits affect tests and this checkpoint only.

Implementation and targeted verification are complete; no historical card was changed or adopted, no staging occurred, and no full suite or review is currently running. Next: push this checkpoint, run central sdlc local-check run --pre-close on the pushed head, then mandatory close review and central merge on SHIP. The approved version-3 method refuses older readers rather than allowing a silent undercount; old adopted snapshots remain immutable.

Close preflight required a bullet in Done when; formatted the existing acceptance paragraph as one bullet without changing its criteria. Measured actual preview succeeded; no review was dispatched by that preflight refusal. The existing recorded suite remains queued on unchanged code.

### 2026-10-23 20:54 PDT — first boundary review correction

Full review returned REWORK with one Important finding, BR-1 (retained-attribution-consistency). Reproduced its class through saved replay: canonical own or foreign issue branches relabelled ownership-only were accepted. The shared saved-proof validator now enforces every attribution kind's branch identity, so neither replay nor direct execution can accept that contradiction. Saved-replay mutation controls first failed, then incident/capture controls passed in 10.660 s. No old adopted snapshot semantics changed. Added the rule to lessons and the plan's Revisions. Cancelled only this checkout's old-head queued suite through SDLC; it exited before admission. Next: targeted race/static checks, commit and push, fresh pre-close suite and round-2 close.

2026-10-23 20:54 PDT: BR-1 focused race regression PASS 8.122 s; Go catalog PASS 1.392 s. Rebuilt private bin/sdlc from Git revision dbef315ea42c6ecb402358b047cabeb411411f5e (Go 1.27.3, modified=false); isolated real-command conformance PASS 8.194 s. PR 425 updated to that revision. Next: push this text-only evidence checkpoint, queue final-code pre-close suite and run close round 2; no background run remains at this checkpoint.

### 2026-10-23 22:17 PDT — complete suite classification and correction

Recorded attempt 9ac60fc393d946bc6b53e2b94c11a27b on dd39e615a6561921075b0066ba0cd09a8e0ca22d completed after 3932 s queueing and 15m6s execution: 2444 cmd/sdlc tests and 76 other packages. Exactly TestParkedClaimExcludesOtherAdoptedIssueInSameSlot and TestVerbContextsReachTrackerReads failed; all other shards/packages passed. These are deterministic branch integration defects, not flakes; no failure was registered or resumed. Raw result is retained under the common Git directory sdlc/test-stats/failed/9ac60fc393d946bc6b53e2b94c11a27b.log.

The context fallback now reuses commandContext, preserving a supplied cancellation context and the established nil-context boundary. The older protected-credit fixture previously generated a fresh #649 adoption while its worktree remained on #636; the new classifier correctly excludes that. Its intended historical-adoption protection is now explicit: construct a valid immutable version-1 ownership-only snapshot and replay it, then verify the new candidate still excludes that historical credited interval. No production historical sample is rewritten. Both named controls pass together in 24.900 s. Next: push these corrections, queue a fresh pre-close suite, and run fixed-then-ship close review of the post-SHIP delta. No background runner remains at this checkpoint.
