# Boundary Review — ariadne#289 (milestone M2)

| field | value |
|-------|-------|
| issue | 289 — Slot readiness in sdlc fleet inventory |
| repo | ariadne |
| issue file | workshop/issues/000289-slot-readiness-in-sdlc-fleet-inventory.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 407bd61baa85197b041573f3b4560aa8c31fc9d1..fa54dee5dcbcfd5763e7359e08371df6e2d00d57 |
| command | sdlc milestone-close --issue 289 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T17:09:57-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 does what the Spec and Done-when ask for. Slots are found from row paths alone. Membership comes from the declared `construct/deps` substrate rows, and an undeclared sibling is never a member. Dependency clones become ordinary rows and reuse the fleet primary's tracker read. A pure `AssembleSlots` takes the worst member verdict as the slot's verdict, and the contract is versioned and strictly validated. A real-git fleet fixture tests every verdict cause, in both the host and the clone. I checked out the head commit (`fa54dee5`) and ran `go test` for `internal/fleet`, `internal/recovery` and `pkg/workspace`: all green. Nothing blocks shipping. There is one Important finding: `membership.go` writes its own copy of weave's `construct/deps` reader, and the copy is weaker than weave's. This is the second time this issue has duplicated a weave/sdlc helper, so the fix should be the rule, not just this one instance.

1. **Strengths**
   - `discoverSlots` (`cmd/sdlc/internal/fleet/slots.go:135`) runs no git process. It checks each row against `workspace.SlotPath` and rejects look-alikes such as `pair-slot01` or another repository's slot path. `TestDiscoverSlots` covers those cases.
   - Tracker aliasing compares publication identities (`gitx.PublicationRepository`), not raw URLs, so this fleet's SSH primary and weave's HTTPS clones are recognised as one repository. The e2e test checks that no root under `worktree/` reads its own tracker, and that the clone's claim is placed rather than left dangling.
   - `TestFleetInventorySlotReadiness` is a real stateful fixture. It has 11 numbered slots, each with one cause, placed in the host or in the clone: claim, unlanded commits, a zero-commit open-issue branch, dirty files, a merge in progress, a detached HEAD, a missing clone, an undeclared sibling and a malformed declaration. It also round-trips the result through strict JSON.
   - Failure stays visible. `withProbe` never lets a failed probe leave a member ready, and a member that is not a row becomes `unknown` with the reason `probe:checkout`. `Slot.validate` enforces "reasons exactly when not ready" and "slot verdict = worst member" on both marshal and unmarshal.
   - `workspace.RestingBranch` is extracted, and `Classify` now uses it (ARCH-DRY).

2. **Critical:** none.

3. **Important**
   - `cmd/sdlc/internal/fleet/membership.go:85-105` — `readDeclaration` re-implements weave's `acquire.ReadDeclarations` (`cmd/weave/internal/acquire/acquire.go:404-430`) and restates its 1 MiB limit as a literal. The copy is weaker. Weave refuses non-regular files and opens with `O_NOFOLLOW|O_NONBLOCK`. `os.Open` follows a symlinked `construct/deps`, and blocks forever if `construct/deps` is a FIFO, which would hang the whole read-only `fleet inventory` (ARCH-SECURE).
     - **This is the 2nd finding in family `single-source-marker-list`.** M1 fixed the first instance, the operation-marker list, by moving it into `pkg/workspace`.
     - **The rule:** any read of slot state that both weave and sdlc perform lives once in `pkg/`; `cmd/sdlc` never re-implements `cmd/weave/internal` behavior.
     - **Fix the rule:** move `ReadDeclarations` and its limit into `pkg/layergraph` (for example `ReadDeclarationFile`), and have weave's refresh/acquire and sdlc's `readDeclaration` both call it. Then search `cmd/sdlc` for other re-derivations of weave behavior. The transitive substrate walk itself is one: weave's planner canonicalizes with EvalSymlinks, while `DeclaredMembers` cleans paths lexically. Either share the walk or record in the plan why it stays separate.

4. **Minor**
   - `cmd/sdlc/internal/fleet/slots_test.go` — the "null slots" mutation also adds an unknown key `"z"`. Strict decoding rejects that key on its own, so the `slots must be non-null` check is never what fails the test. Replace only the value instead.
   - `DeclaredMembers` builds member paths lexically (`filepath.Clean`), but rows use canonical paths. A symlinked environment or member would not match its row and would report `unknown` with the reason "not a Git checkout the inventory could read". That is conservative (never wrongly ready), but the message misleads, and the alias lookup would then miss, costing an extra tracker read.
   - `collectDependencyRows` calls `readDeclaration` and `statPath` directly while receiving `collect` and `git` as injected parameters. Inject all of them, or say why not (ARCH-PURE).
   - `sameOrigin` re-reads the primary's origin for every clone. This is cheap, but a per-run memo would be trivial to add.

5. **Test coverage**
   - Every Done-when cause is covered end to end.
   - The e2e test uses a merge for the active-operation case. Rebase coverage comes from M1's `pkg/workspace` test and the `JudgeCheckout` table, which is acceptable.
   - There is no test for a declared member that exists but is not a Git repository on the e2e path; only the pure `TestAssembleSlots` covers it.
   - There is no test for a clone whose origin differs from the primary's (the no-alias path).

6. **Architectural notes**
   - ARCH-DRY: flagged (the Important finding).
   - ARCH-PURE: pass, with the injection Minor.
   - ARCH-PURPOSE: pass. Every Done-when item is delivered, and there is no deferred "follow-up" that is the real point of the issue.
   - ARCH-MOCK: pass. The tests use real git end to end through the existing `GitReader` seam.
   - ARCH-CONSTRAINTS: pass. The Log measures 6.3s against the 5.9s baseline, inside the +1s target, and clones add no network read.
   - ARCH-SECURE: flagged, within the Important finding (symlink/FIFO on the declaration read). Members outside the environment and absolute paths are handled correctly.
   - ARCH-ORDER: pass. This is a single-shot observation, and the help text says an action must re-check the verdict.
   - ARCH-FUNERAL: pass. Nothing durable is created.
   - For Couch consumers: `schema_version` is now required on decode, so any older JSON emitted before this change is rejected. That is intended, but downstream should be told.

7. **Plan revisions:** if the declaration reader moves to `pkg/layergraph`, add a `## Revisions` entry that updates the Integration row for `readDeclaration` / `statPath` and notes the shared weave/sdlc reader.

```findings
findings:
  - id: new
    severity: Important
    family: single-source-marker-list
    title: |
      membership.go re-implements weave's construct/deps reader, weaker (follows symlinks, blocks on a FIFO)
    detail: |
      readDeclaration (cmd/sdlc/internal/fleet/membership.go:85) duplicates acquire.ReadDeclarations (cmd/weave/internal/acquire/acquire.go:404) and its 1 MiB limit, without the regular-file check or O_NOFOLLOW/O_NONBLOCK, so a FIFO construct/deps hangs the read-only inventory. This is the 2nd finding in family single-source-marker-list (M1 unified the operation markers into pkg/workspace). Rule: any slot-state read both weave and sdlc perform lives once in pkg/; cmd/sdlc never re-implements cmd/weave/internal behavior. Fix the rule: move ReadDeclarations (+limit) to pkg/layergraph, use it from weave refresh/acquire and fleet membership, and search for other duplicates (the substrate walk: weave canonicalizes, DeclaredMembers is lexical).
  - id: new
    severity: Minor
    family: contract-rejection-test-isolates-invariant
    title: |
      "null slots" contract mutation also adds an unknown key, so the non-null check is never what fails
    detail: |
      slots_test.go TestSlotsContract replaces `"slots":[{` with `"slots":null,"z":[{`; strict decoding rejects "z" regardless, so validateSlots' nil check has no test that isolates it.
  - id: new
    severity: Minor
    family: lexical-vs-canonical-path-join
    title: |
      DeclaredMembers builds member paths lexically while rows are canonical
    detail: |
      A symlinked member or environment does not match its row: the member reports unknown with a misleading "not a Git checkout" error, and the alias lookup misses (an extra tracker read). Conservative, never wrongly ready.
  - id: new
    severity: Minor
    family: io-not-injected
    title: |
      collectDependencyRows calls readDeclaration/statPath directly but has collect and git injected
```
