---
collaboration-mode: discussion
---

# Issue-only main drift implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution approach.
> Keep the landing state changes together in the main session; the close gate
> owns the fresh-context review. No separate ad-hoc boundary review.

**Goal:** Issue filing and archiving on main must not consume a ready PR's code
integration budget or force a new PR head when GitHub can merge the tested head.

**Architecture:** Extend the existing integration state machine, not a second
landing loop. A conservative Git observation classifies every newly reachable
main commit. Exact-head gates still run; fresh GitHub CLEAN evidence permits an
unchanged-head merge for issue-only drift. Otherwise use the existing integration,
leased push, exact-head CI and ancestor-suite reuse without charging an issue-only
integration to the three-attempt code budget.

**Tech stack:** Go; existing gitRunner and landingGH; real temporary bare Git
origins plus the existing stateful GitHub/check-status fake.

## Approved decisions and alternatives

Hub approved this boundary on 2026-10-19, before implementation. No workflow,
label, branch-protection or persisted policy changes. Do not add ops/ now.

- Recommended: unchanged head plus fresh GitHub readiness, conservative integration
  fallback. Preserves the exact status/head binding and server policy.
- Only exempting the counter is simpler but still needlessly pushes a ready head
  and waits for CI on permissive repositories; keep it as the fallback.
- Inferring reusable server checks from paths-ignore is unsafe: PR path filters
  consider the entire PR diff. Never copy old CI statuses to a new commit.

GitHub defines CLEAN as mergeable and passing status; BEHIND is out of date.
Missing, unknown and non-CLEAN states confer no permission for the optimization.
References: [merge state enum](https://docs.github.com/en/graphql/reference/pulls#mergestatestatus),
[workflow diff and filter rules](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow#git-diff-comparisons).

## Core concepts

| Name | Kind | Location | Status |
|---|---|---|---|
| Issue-record path policy | PURE | cmd/sdlc/mainmovement.go | new |
| integrationState / transitionIntegration | PURE | cmd/sdlc/integrationstate.go | modified |
| landingPR readiness fact | PURE data | cmd/sdlc/ghlanding.go | modified |

The path policy has one list: workshop/issues/, workshop/plans/,
workshop/history/. Initially admit only Markdown record paths under those roots;
non-record payloads, code, configuration and every other root are ordinary.
A later policy can extend this one list; ops/ is explicitly excluded today.
Do not reuse the broader docs-only classifier (which admits arbitrary Markdown,
atlas and docs) to authorize avoiding integration. Existing suite reuse keeps its
own established, broader docs-only contract.

The state machine remains the only writer of head, old lease head, selected main,
phase, retry count and the pending integration's charge. Add the issue-only fact
to observed-main events and a validated readiness fact to the final gate event.
An integration that actually includes code remains chargeable even when a later
issue-only observation arrives during its replay/push.

| Seam | Location | Wraps / fake |
|---|---|---|
| Main movement observation | cmd/sdlc/mainmovement.go | gitRunner; real Git fixture histories |
| Fresh PR readiness | cmd/sdlc/ghlanding.go | existing landingGH.LandingPRs, extended parsed field |
| Landing transaction | cmd/sdlc/landing.go, landingintegrate.go | existing repo lock, leases, expected-head merge |
| GitHub/check status model | cmd/sdlc/landing_test.go, cigate_fake_test.go | bare remote refs, per-head checks/statuses, policy and outcome |

## Observation and trust rules

1. Use full verified commit IDs and enumerate main commits not reachable from the
   relevant head, not a net tree diff. A code edit followed by its revert remains
   ordinary. Include merge-parent differences and both sides of renames, using
   NUL-delimited paths, no rename collapsing and no shell interpolation.
2. Use one history/path enumeration per observation (for example git log with
   --format=, --name-only, -z, --no-renames, --root and merge-parent diffs).
   A missing ancestry/history object, malformed path record or Git error prints
   that the optimization is unavailable and takes the ordinary integration path.
   Do not turn an unreadable range into an empty/issue-only range.
3. For an interrupted integration, classify against its validated remote lease
   head, not the local merge that already contains main, so recovery cannot hide
   a code integration's charge. Retain integrationResume's two-parent/tree proof.
4. Add optional mergeStateStatus to landingPRFields and parseLandingPRs. Accept
   only the exact string CLEAN for direct merge. Unknown future strings or missing
   readiness fields disable the optimization; malformed identity still refuses.
   Do not weaken existing repository, PR-number, head/base OID or branch validation.
5. After the existing exact-head gates pass, refresh PR evidence. Bind CLEAN to
   the selected repository, PR number, branch, head and the fetched main BaseOID.
   Reacquire the repo transaction lock, revalidate the checkout/head and fetch main
   again. A stale base or non-CLEAN policy falls back to normal integration; a
   changed head/identity preserves the branch and refuses, as today.
6. Retain duplicate-ID, instance-conformance and publish gates against the pinned
   destination, including issue data main published after the branch point.
   Those gates already use the merge base and do not require main to be an ancestor.
   A conflicting issue record is never forced through: CLEAN must be current,
   and a read-only merge-tree preview can conservatively send it through normal
   integration's existing conflict abort/refusal.
7. A failed server merge is still an uncertain outcome: preserve state and use
   the existing recovery command. Never reinterpret it as definite rejection and
   issue an automatic second merge or integration.

## State and effects

Actions: merge normally; retry the printed merge command; resolve/re-close a
reported conflict; reconcile a possibly completed landing. No new durable state.

| Internal state / event | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Gates; ordinary/code main movement | bounded code integration, same three-attempt rule | normal merge / retry after bound | none; branch preserved |
| Gates; issue-only main movement | keep tested head and run normal gates | normal merge | none |
| Gates passed; issue-only + exact fresh CLEAN | expected-head server merge | reconcile if outcome uncertain | no status copying or unchecked head |
| Gates passed; strict, unknown, stale-base or record conflict | integrate normally; no issue-only attempt charge; fresh exact-head CI | normal merge / resolve named conflict | none; no gate waived |
| Replay/push; newer issue-only main commits | retain the pending merge and its original charge; re-observe before landing | normal merge | none |
| Replay/push; code/mixed/unreadable movement | existing rollback/retry path and charge semantics | normal merge / retry | existing preserved integration/lease behavior |
| Push registration or CI interrupted | existing recoverable/refused state | printed merge recovery | no fabricated success |
| Server merge error or queued/unconfirmed result | existing uncertain/not-yet-merged refusal | reconcile / retry as printed | none; no second effect assumed safe |
| Confirmed MERGED + main reachability | existing archive and rest cleanup | existing cleanup recovery | none |

An issue-only integration's registered push or retry reset adds zero to attempts;
code integrations retain their existing charge. The machine records this before
IO and owns the count at both registration and reset. Allow metadata-only progress
at an exhausted code budget, but never allow one more code integration.

The repo's existing landing turn is the owner-lifetime resource that serializes
landings. Issue publication deliberately does not hold that resource; making it
wait would reintroduce the operator freeze this issue removes. Keep the existing
transaction lock and expected-head server pin. No new locks, leases, policy cache,
CI bypass, global publication freeze, timeout increase or background actor.
Server/base races beyond the existing final revalidation remain server-enforced;
a policy observation is evidence for one exact snapshot, not authority over later
unknown main changes.

## Operating envelope / architecture

- ARCH-DRY: one issue-record list and one movement helper used at the outer loop,
  post-registration, replay/push and final gate observations; one state machine.
- ARCH-PURE: classification and state transitions take values; Git/API reads stay
  in the IO shell. State tests do not mock IO.
- ARCH-PURPOSE: cover first entry, CI/suite wait, registration and recovery, not
  merely the final loop. Confirm actual fewer pushes/CI reads, not just log text.
- ARCH-MOCK: extend the existing stateful fake with effective allow-behind / strict
  policy and per-head status storage. Real Git creates and merges actual commits.
- ARCH-CONSTRAINTS: per observation O(C + P) in new main commits and changed paths;
  one bounded Git history call, no per-path subprocesses. Unit/counted fixtures at
  N and 2N reject a growing number of Git calls. Readiness costs one fresh listing
  at the final gate, outside the repo lock; retain existing GitHub time budgets.
- ARCH-SECURE: only validated OIDs/identities and typed readiness; unknown or failed
  policy observations never authorize direct merge. No secrets or new permissions.
- ARCH-ORDER: the table above extends transitionIntegration; test interrupted,
  concurrent-main and uncertain-server effects through that single owner.
- ARCH-ELABORATE: no generic workflow parser or protection/ruleset reconstruction;
  effective GitHub readiness is the narrow fact needed here.
- ARCH-EVOLVE: no persisted format change. The optional GitHub field is typed and
  additive; missing/future values cause conservative fallback.

## Tasks

### Classifier and state (one atomic close boundary)

Files: mainmovement.go/mainmovement_test.go; integrationstate.go/integrationstate_test.go.

- [ ] Write red tests for safe record roots and ops exclusion; arbitrary docs,
  embedded help, config, mixed commits, code-add/revert, cross-root rename, merge
  commit, empty/unreadable history, names with whitespace and no-prefix lookalikes.
- [ ] Write red transition sequences for three issue-only observations, metadata
  fallback registration/reset not charging, ordinary code still charging/refusing,
  metadata after budget exhaustion, and a pending code integration retaining its
  original charge through metadata replay/push and resume.
- [ ] Implement the minimal shared observation/classifier and state changes.
  Run go test ./cmd/sdlc -run 'Test(MainMovement|IssueRecord|IntegrationState)' -count=1.
- [ ] Commit the classifier/state slice with its passing focused evidence.

### GitHub readiness and full landing wiring

Files: ghlanding.go/ghlanding_test.go; landing.go; landingintegrate.go;
landing_test.go; landingintegrate_test.go; landingeconomy_test.go.

- [ ] Add red adapter tests for exact CLEAN, BEHIND/UNKNOWN/missing/future readiness,
  invalid readiness type and changed repo/PR/head/base identity.
- [ ] Extend the real-Git/stateful-GitHub fixture: create a ready tested PR; publish
  three distinct issue-only main commits while its gate waits; assert unchanged
  PR head, one code integration at most (the initial fixture setup), one suite,
  one expected-head merge and no fresh-head CI cycle.
- [ ] Add strict/unknown-policy fixtures that integrate issue-only changes without
  spending attempts, retain exact-head CI and reuse the suite; pending checks on
  the new head must block, not inherit the old head's green checks.
- [ ] Cover same-record conflict, metadata-to-code/mixed drift during the final
  revalidation, code-add/revert, pre-push movement, interrupted resume and uncertain
  server merge. Verify no duplicate merge effect or unleased publication.
- [ ] Wire observations into every existing integration seam, refresh readiness
  after gates and revalidate before deciding the direct merge or normal fallback.
  Keep the normal central binary for all SDLC gates; no private landing binary.
- [ ] Run targeted landing/state/CI/adapter/economy/recovery tests, including the
  existing docs-suite-reuse and code/config/embedded-help invalidation controls.
- [ ] Commit with issue, Agent and Co-Authored-By trailers.

### Verification and delivery

Files: cmd/sdlc/helptext/merge.md; atlas/workflow/sdlc-binary.md; README.md if its
existing merge contract needs the changed attempt semantics; workshop/lessons.md
for any review-derived rule (no ceremonial wording tests for prose).

- [ ] Document the single issue-record list, effective CLEAN evidence, unchanged
  head optimization, conservative fallback, code-only attempt budget and recovery.
- [ ] Run a disposable local real-Git/fake-GitHub harness exercising both permissive
  and strict policy (no shared staging, real PR merge or branch-protection change).
  Read-only live conformance: compare one existing GitHub PR's selected fields
  against the parser; do not mutate that PR or apply any label.
- [ ] Push/open PR with sdlc pr; run the recorded branch suite through normal sdlc.
  Close the tested branch through the binary's mandatory fresh review.
- [ ] On SHIP, enter sdlc merge --yes promptly. Let its landing turn integrate and
  test code changes as required; report landed PR and Git-derived merge hash,
  archive/card evidence and both attempts if a registered flake reruns.

No Mx markers: this is one atomic behavior change with one close review boundary.
