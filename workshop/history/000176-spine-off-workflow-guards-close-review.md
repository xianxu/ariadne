# Boundary Review — ariadne#176 (whole-issue close)

| field | value |
|-------|-------|
| issue | 176 — spine guards for off-workflow invocations: change-code on done issues + non-SDLC repos |
| repo | ariadne |
| issue file | workshop/issues/000176-spine-off-workflow-guards.md |
| boundary | whole-issue close |
| milestone | — |
| window | 66ca9bfb74af6922253ac2f678f99576d0123e8a..HEAD |
| command | sdlc close --issue 176 |
| reviewer | claude |
| timestamp | 2026-07-14T18:04:47-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The guard family is correctly built and well-tested: both arms of `guardSpineRepo` behave as specced, `die` has halt semantics so there's no fall-through, the guard-first ordering is pinned by the message-asserting drift test, and the repo-top anchoring fix (Revisions delta 1) is right. `go build`/`vet`/`go test ./cmd/sdlc/...` all pass. I also verified the brain-side charter actually landed (`~/workspace/brain/AGENTS.local.md` leads with the capture-repo charter, and `.brain/config.md` exists). What keeps this from a clean SHIP is one purpose gap: the docs and code comments claim the `WF_SPINE_GUARD=off` bypass is "measurable by the #172 friction instrument," but the instrument's single source (`GateCatalog`) got no row for the new gate family — the classifier can never count these bypasses or refusals, and the new no-collision test actively guarantees it stays blind to them. Plus one test the plan claims but doesn't deliver.

## 1. Strengths

- **Drift-proof verb enumeration (ARCH-DRY pass):** `repoguard_test.go:17` derives the guarded set from `processmanual.WorkflowVerbs()` — the same catalog the friction instrument uses — so a new lifecycle verb added to the catalog automatically fails the test until it carries the guard. This is the right single-source shape.
- **Guard-first placement is pinned, not just asserted:** `TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs` requires the *charter message*, so a verb whose own validation fires first (e.g. "--issue is required") fails the test. That enforces the "refusal, not a confusing downstream error" contract mechanically.
- **The repo-top anchoring** (`repoguard.go:69-75`) with the cwd-relative `WF_ISSUES_DIR` override matching how the verbs read it is the correct resolution of the subdirectory false-positive the suite caught — and the Revisions entry documents it honestly.
- **Precision on reads-vs-lifecycle:** wiring the guard into exactly the 7 lifecycle RunEs (exemption by construction, not classification) matches the Spec's "sdlc legitimately reads brain" requirement with zero classification machinery.
- **Refusal messages carry the next action** (charter + positive path; new-issue/`set-status` reopen path), consistent with the repo's errors-as-next-action-specs convention.

## 2. Critical findings

None.

## 3. Important findings

1. **ARCH-PURPOSE — the "measurable by the friction instrument" claim is not wired.** `repoguard.go:20-21`, `atlas/workflow/sdlc-binary.md:401`, and the plan all state the env hatch "cwarn-ACKs so the #172 friction instrument can measure bypasses." But the classifier derives *exclusively* from `GateCatalog` rows (`gatesig.go:10-14` calls it "the single source"; matching happens at `friction.go:265`/`274` via each row's `ackRE`/`refusalRE`), and no row was added for the spine-repo guard. So the instrument counts neither `WF_SPINE_GUARD=off` bypasses nor brain/non-SDLC refusals — `TestSpineGuardLinesNoGatesigCollision` in fact *guarantees* the instrument stays blind to all four lines. The ACK is greppable in raw transcripts, but "the instrument can measure it" is a hand-maintained restatement, not a derived consumer. Fix: either add a `GateCatalog` entry for the guard family (an env-gate variant of `GateSig` — bypass ACK `spine repo guard bypassed`, refusals `is a brain \(capture repo\)` / `not an SDLC repo`), or correct the three doc sites to say the ACK is transcript-greppable and instrument wiring is follow-up work.

2. **Plan claims a test that doesn't exist.** Plan Task 1 Step 1(e) (checked `[x]`): "…non-done statuses pass". `repoguard_test.go` has no non-done case — `TestGuardIssueNotDone` only covers `status: done`. This is exactly the false-positive class worth pinning (guard firing on a live issue would block all work). Cheap fix: add a `status: working` issue to the same test and assert `start-plan`/`change-code` don't die with the done message. Alternatively revise the plan to stop claiming it — but the test is ~10 lines.

## 4. Minor findings

- `repoguard.go:77`: `notSDLCRepoMsg` is called with the literal `"workshop/issues"` even when a `WF_ISSUES_DIR` override was what failed the stat — the message would misname the missing dir. Pass `issuesDir` through.
- `startplan.go:42`: the done-guard resolves `workshop/issues` cwd-relative while `guardSpineRepo` anchors at repo top — from a subdirectory the done-guard silently skips (locate fails). Consistent with start-plan's other reads (`issueEstimate`), but a gap vs the family's "correct from any subdirectory" property.
- `repoguard_test.go:143` comment says "the three new lines" then checks four; atlas `sdlc-binary.md` likewise says "three new lines". Say four (or "the new lines").
- Plan Task 2 says the constitution one-liner goes in "AGENTS.base.md §0/§5"; it landed in §1's Peer-Repo brain bullet (a fine home, but the Revisions entry doesn't note the delta).
- The GUARDS family note landed only in `helptext/close.md`; the other six guarded verbs' help never mentions the repo guard or the env hatch. Defensible reading of the plan wording, but an agent refused at `claim` who runs `sdlc claim --help` won't find it.
- ARCH-DRY (tiny): `guardIssueNotDone` re-implements the read→`issue.Parse`→`GetField("status")` shape that `issueStatus` (`setstatus.go:136`) already owns; a shared path-based `issueStatusAt(path)` would serve both. (The literal `"done"` comparison is fine — it's the documented #122 value-specific carve-out used consistently across the package.)

## 5. Test coverage notes

- The delivered tests are hermetic e2e through the real command tree with no mocks — good.
- Gaps beyond the non-done case above: (a) the `WF_ISSUES_DIR`-override arm of `guardSpineRepo` (override set, dir exists → pass; override set, dir missing → die) is untested; (b) the Spec's key precision property — reads (`state`, `issue`, `estimate-source`) proceed in a brain repo — was only live-verified manually. A one-case test (`sdlc state` in a brain hermetic repo must not die with the charter) would pin the property against someone later wiring the guard into a read verb.

## 6. Architectural notes

- **ARCH-DRY: pass** (with the minor `issueStatus` note). The catalog-derived drift test is the load-bearing DRY move and it's done right.
- **ARCH-PURE: pass.** The guard is a thin IO preflight (two stats + env read); messages are pure functions tested without IO; the e2e tests need no mocks. At ~40 lines, folding decision and `die` together is proportionate.
- **ARCH-PURPOSE: flag** — Important finding 1. Otherwise the purpose is genuinely delivered on both layers (binary gate + brain constitution, both verified), and the reads-stay-unguarded precision requirement is met by construction.
- For upcoming work: if the friction instrument grows a `GateCatalog` row for this family, note the ACK grammar is new (cwarn line, env-based, no `--force` variant) — it'll need a fourth grammar or an env-gate flag on `GateSig`, and the no-collision test should then assert against *other* gates' patterns only.

## 7. Plan revision recommendations

Add a `## Revisions` entry covering: (1) the instrument-measurability outcome once finding 1 is resolved (catalog row added, or claim softened to "transcript-greppable; instrument wiring deferred to a named follow-up"); (2) either the added non-done-pass test or removal of that clause from Step 1(e); (3) a one-line note that the constitution edit landed in §1 (Peer Repo) rather than §0/§5, and that Done-when #3 ("brain drops out of the friction report") is verifiable only at the next re-measure, not at this close.
