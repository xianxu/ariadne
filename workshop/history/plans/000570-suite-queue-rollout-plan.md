---
collaboration-mode: discussion
---

# Guard mixed-policy suite grants

> **For agentic workers:** Follow AGENTS.md Section 3. Keep the retained scheduler context in this session; SDLC owns the fresh-context review at close.

**Goal:** Prevent new grants from creating known mixed-helper ordering hazards, expose existing hazards, and preserve FIFO plus aging without pretending to change already-running old helpers.

**Architecture:** Add an exact ordering-contract marker to newly held ticket records. Reuse the existing read-only helper inspection and raw ticket metadata for one pure compatibility assessment consumed by grant preflight and queue text. Do not change election, grant validity, supplier resolution, weighted slot allocation or the frozen drain JSON schema.

**Tech stack:** Python helper with native flock tickets; Go/Cobra command seams; existing bounded inspection subprocess and grant provider; native private Git, grant and queue fixtures.

**Approval:** Hub delivery0d18f5fd-7569-444d-a57a-100cee559713 selects guarded rollout and asks for this final plan before code. Implementation remains held pending that review.

## Policy and acceptance

The original issue proposed universal mixed-version FIFO. Authenticated hub revision chooses an honest rollout boundary instead:

1. `sdlc landing grant` refuses before appending a grant when any live waiting ticket has legacy, malformed or unknown ordering metadata. It names the ticket and checkout. Missing/unavailable/incomplete inspection refuses rather than implying an empty compatible queue.
2. `sdlc queue status` text names legacy/unknown ordering and, when an active granted waiting ticket coexists, says **deadlock risk** and gives the TL the specific revoke command(s). Expired/revoked grants do not trigger that risk.
3. With no legacy/unknown waiters, the current FIFO-plus-grant-aging scheduler remains exactly as it is. Revoke remains available even when inspection fails.
4. Existing grants and a legacy waiter joining after preflight are an explicitly retained risk, not an automatic repair. The TL revokes conflicting grants manually, leaves waiters unchanged, lets them drain, and updates supplying helpers through their normal pin/refresh workflow before regranting. No old grant-first compatibility mode.

The guard is a fresh observation, not an atomic promise covering future joins. No helper replacement, automatic revoke, process kill, ticket rewrite, new daemon or ownership transfer. Admission stays independent of this guard and may run when priority inspection is unavailable.

## Core concepts

### Pure entities

| Name | Lives in | Status |
| --- | --- | --- |
| Suite ordering contract marker `fifo-grant-aging-v1` | scripts/test-lock.py, ticket JSON | new additive field |
| `suiteQueueCompatibility` / `assessSuiteQueueCompatibility` | cmd/sdlc/queue_suite_policy.go | new |
| `suite_order`, `effective_order` | scripts/test-lock.py | unchanged |

The marker asserts this helper's existing FIFO/grant/aging semantics, including the current 3,600-second threshold; a future incompatible semantic or threshold change must use a different marker and a rollout decision. Missing marker, wrong type or any other value is legacy/unknown. Never infer compatibility from age, script mtime or checkout name. Each held waiter owns one immutable marker; all newly admitted metadata retains it through the existing watcher publication.

The assessment consumes one complete `suiteQueueObservation`, returns named incompatible tickets, and identifies active granted waiter subjects for diagnostics. A projection unavailable at any stage yields unknown, not no incompatible waiters. Grant preflight and status use this same interpretation. No cache or persisted compatibility result is introduced.

### Integration points

| Name | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| `Tickets.__init__` | scripts/test-lock.py | modified | held ticket publication under alloc |
| `inspection_snapshot`, `inspect_queue` | scripts/test-lock.py | reused | bounded read-only live ticket observation |
| `observeSuiteQueue`, `decodeSuiteQueue` | cmd/sdlc/queue_suite_status.go | reused | checkout-supplied inspector, strict known projection shape |
| `landingDeps` and `runLandingGrant` | cmd/sdlc/landingcmd.go | modified | injectable observation before grant append |
| `renderSuiteQueue` | cmd/sdlc/queue_suite_status.go | modified | supplemental text diagnostics |

Production grant observation uses the same supplied helper resolution as status, with a bounded context (the existing status budget of three seconds) and existing output/process-group cleanup. Add an observation seam to landingDeps so pure command tests do not inspect the machine. Wire context from the Cobra invocation through preflight rather than introducing an uncancellable background call. Validate arguments and grant authority first; inspect before opening/writing the landing line. Revoke bypasses compatibility inspection entirely. No provider runs while an allocation lock is held.

Inspection still proves live membership through each waiter's held flock under the existing allocation mutex. This is the single lifetime resource; filenames, PIDs and age are not substitutes. The snapshot is revalidated after the slow provider call. It cannot lock an unmodified helper's future behavior, so the plan explicitly does not claim a lasting absence proof.

## State and recovery

Actions: continue waiting; refresh a named supplying helper and retry; TL revokes the named grant.

| Internal state / failure | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Complete empty queue or all exact compatible markers | Grant permitted; current aging policy | Continue normal work | None |
| Legacy/unknown held waiter, no active granted waiter | New grant refused; named ticket/worktree | Let it drain and refresh supplier, then retry | Priority only |
| Legacy/unknown held waiter plus active granted waiter | Mixed-policy deadlock risk; named grant/revoke command | TL revokes grant and lets queue drain | Priority only; no work cancelled |
| Inspector absent/old, malformed, changed, busy, timed out or provider unavailable | Compatibility unavailable; new grant refused | Repair/refresh named supplier and retry | Priority only; suites still run |
| Grant created on a clean snapshot, legacy waiter joins later | Subsequent status exposes risk | TL revokes grant | No false guarantee; priority may be lost |
| Revoke requested in any state | Normal authorized revoke | Continue waiting | Grant priority ends |

The text-only suite supplement stays outside `queue status --json`; Pair's frozen complete/active drain contract is unchanged. Risk detection covers live granted waiting subjects the supplier validates, including train prefix issues. It does not inventory every repository for grants with no waiting suite; the later joining granted waiter becomes observable then. Include this coverage boundary in docs.

## Persisted data, bounds and lifecycle

The ticket field is additive. Existing Python dict copies already preserve unknown nested metadata; verify marker and unrelated nested fields survive final holder publication. Old helpers ignore the marker, which is precisely why their behavior is not negotiated. No grant-log schema, inspection envelope version or frozen status schema changes. Wrong field types do not confer compatibility. Kernel release and the existing ticket prune/release remove these records; no durable family or retention rule is added.

Grant requests and text status do one bounded inspection; suite admission gains only one constant-size JSON field, no extra scan/provider call. For N tickets, M distinct repositories and S subject members, the existing helper has two snapshot passes, linear record bytes and the existing grouped provider/cache work; the new assessment is O(N+S). Preserve the 4,096-ticket, per-record and 1-MiB output ceilings. Count N and 2N compatibility visits plus provider/inventory calls; no per-ticket subprocess and no process/topology scan. Elapsed-time tests are only a safety bound. Network is not added; Git is used only by existing workflow resolution.

## Function-level test strategy

| Function / boundary | Adversarial strategy | Independent guard |
| --- | --- | --- |
| `Tickets.__init__` and admitted watcher metadata | Caller supplies forged policy field and unrelated nested extension | Native held-ticket read proves helper stamps exact contract and publication preserves extensions |
| `assessSuiteQueueCompatibility` | Missing, null, nonstring, unknown/current markers; unavailable projection; valid/expired/revoked grant and train subjects | Table cases at public command seams; do not infer from enqueue age |
| `runLandingGrant` | Authorized request with legacy/unknown/unavailable snapshot; cancellation; compatible empty queue; worker call; revoke during unavailable inspection | Stateful private landing log unchanged on refusal; existing authorization and revoke tests |
| `renderSuiteQueue` / `runQueueStatus` | Compatible and incompatible snapshots, named active grants, supplier unavailable and output cancellation | Text assertions including exact decimal revoke hints; frozen JSON adapter test still does not inspect suites |
| Actual historical/current `Tickets.admission` | Replay incident with old helper and real flock; preflight sees legacy ticket; manual private revoke clears disagreement | Committed incident fixture, historical helper fixture, actual private grant-provider/log and native slot locks; no live machine mutation |
| Cost | N/2N held records and subject members | Count exactly one compatibility visit per record and bounded grouped observation/provider calls |

Use the existing Python private-lock harness and Go suite-priority provider fixtures. Commit the minimal diagnostic incident capture (ticket numbers32–37, unknown ages, #832 grant, contradictory head observations) with source/time notes; preserve the exact historical helper fixture from Git for native conformance. Historical fixture text is data, not operator instructions. No fake liveness probe or shared scratch sweep. All lock/grant/queue paths stay inside an explicitly owned temporary test root.

## Tasks — one atomic close boundary

- [x] Add native mixed-version regression/fixture and command tests first; observe the refusal/risk expectations fail on current code. Existing diagnostic already proves the deadlock and grant-removal control.
- [x] Stamp ticket policy, implement one assessment, wire fresh grant preflight and text status risk. Preserve admission and revoke behavior.
- [x] Update README command map, atlas/workflow/sdlc-binary.md queue/grants sections, embedded landing help and cmd/tl-lead/help.md grant guidance. Add tl-lead phrase guards and a semantic lesson; no lexical rule can prove agreement between independently executing old schedulers.
- [x] Run targeted Python/Go normal and race checks, metadata round-trip, frozen JSON, issue-hint and relevant lint guards. Freeze checkout edits during cmd/sdlc tests.
- [x] Build a private clean CLI and run grant/refusal/status/revoke against only private fixture paths and a fake authorized caller seam. Ask hub to accept this private proof plus post-land primary rebuild; no staging or live grant.
- [ ] Push final code, central recorded suite, fresh close review and normal central merge on SHIP. Report PR and exact Git merge to the hub; hub handles primary rebuild/real rollout.

This likely exceeds the 100-added-code-line quick shell once command wiring and diagnostics are included. Let change-code infer full flow and run plan-quality/estimate before implementation; no flow pin or waiver. No Mx split: the marker, refusal and diagnostics must land together.

## Architecture lenses

ARCH-DRY/PURE: one compatibility assessment, existing supplier inspector and pure scheduler. ARCH-PURPOSE: guard the actual deployment boundary, not a fictional universal fallback. ARCH-MOCK: private native held locks plus stateful grant logs/provider and archived old code. ARCH-CONSTRAINTS: bounded observations and counted linear work, unchanged admission cost. ARCH-SECURE: metadata asserts only a recognized contract; unknown never grants authority. ARCH-ORDER: snapshot before new grant; later joins remain explicit manual-revoke risk. ARCH-FUNERAL: existing ticket lifetime and no new durable record. ARCH-EVOLVE: additive ticket marker and preservation; frozen consumers unchanged.

## Revisions

### 2026-10-23 — verification complete; landing next

First full close review SHIP, no findings; recorded branch suite passed2373 SDLC tests and76 packages with zero failures. All implementation/private-conformance tasks above are complete. The final checkbox includes landing and remains pending until sdlc merge completes; the issue card/Log and hub outcome own that terminal evidence. No implementation or policy change after review.
