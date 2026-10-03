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

---

## Re-review — 2026-10-02T17:17:39-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 289 — Slot readiness in sdlc fleet inventory |
| repo | ariadne |
| issue file | workshop/issues/000289-slot-readiness-in-sdlc-fleet-inventory.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 407bd61baa85197b041573f3b4560aa8c31fc9d1..a3e68bfb07dcaf8943ee28e61bb05c347f3fa33e |
| command | sdlc milestone-close --issue 289 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T17:17:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I checked out nothing; the review ran read-only against the pinned range and re-ran the affected packages.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The fix commit `a3e68bfb` correctly handles the reader half of BR-8. The guarded `construct/deps` reader moved to `pkg/layergraph.ReadDeclaration`, and both weave's acquire/refresh and fleet's `readDeclaration` now use it. The contract-mutation tests now change exactly one field, and the declaration read/stat are now passed into `collectDependencyRows`. `go test` is green for `internal/fleet`, `internal/recovery`, `pkg/layergraph`, `pkg/workspace`, weave `acquire` and weave `refresh` at HEAD.

Two things are still open. First, BR-8 explicitly asked to either share the substrate walk with weave or record in the plan why it stays separate. Neither was done. `DeclaredMembers` still walks lexically, and canonicalization was added afterwards, only for present members. Second, that canonicalization (BR-10) has no regression test.

Neither is a crash. One can, in a rare symlinked layout, wrongly report a slot as ready. Nothing in the window blocks the boundary.

1. **Strengths**
   - `pkg/layergraph/read.go:17` is a byte-faithful move of weave's guarded reader (lstat → `O_NOFOLLOW|O_NONBLOCK` → fstat → bounded read → one-byte overflow probe). Weave's FIFO/symlink/oversize table (`cmd/weave/internal/acquire/limits_test.go:30`) still exercises it through `Restore`.
   - `editJSON` (`claims_test.go:249`) fixes the whole class, not just one test: it was applied to the claims contract (`null claims`, `missing machine`, `null dangling claims`) as well as the `null slots` case BR-9 named.
   - `DeclarationLimit` is now defined once. `refresh/git.go:18` refers to it instead of restating `1 << 20`.
   - The injection seam (`inventory.go:208`) is now consistent: collect, git, read and stat are all parameters.

2. **Critical:** none.

3. **Important**
   - **BR-8, still open in part** (`single-source-marker-list`). The rule BR-8 stated was "any slot-state read both weave and sdlc perform lives once in `pkg/`". It covers the substrate walk as well as the file reader.
     - Weave's walk (`acquire.Client.Restore`, `acquire.go:249`) canonicalizes each destination before dedup, queueing and the policy check. The policy validates `canonical(dest)` against the environment.
     - `DeclaredMembers` (`membership.go:59-75`) does the dedup (`visited`) and the inside-the-environment check (`filepath.Dir(p) != envRoot`) on the lexical path. `collectDependencyRows` canonicalizes only afterwards.
     - **Failure case:** `<env>/ariadne` is a symlink to the shared `ariadne:0` checkout. Fleet treats it as an in-environment member, canonicalizes it to the `:0` row, and judges the slot by `:0`'s state, possibly ready. Weave's policy would refuse that layout.
     - **Fix:** do the canonicalization inside the walk, before `visited` and the environment check. Better, extract the substrate-walk core shared with `Restore` into `pkg/layergraph`. Otherwise, add a `## Revisions` line explaining why the walks stay separate, plus a test with a symlinked member.

4. **Minor**
   - **BR-10, not addressed: no regression test.** No fleet test creates a symlinked member or environment. The only symlink-related call is `EvalSymlinks(t.TempDir())` in `fleetslots_test.go:23`, so removing the canonicalization block in `inventory.go:220-235` would not turn any test red. The lexical `visited` set can also produce duplicate members once two spellings canonicalize to the same path.
   - **New** (3rd finding in `single-source-marker-list`): `pkg/layergraph.Walk` still reads `construct/deps` through `OSFS.ReadFile` → `os.ReadFile` (`walk.go:96`, `fs.go:43`). That read follows symlinks, can block on a FIFO and has no size bound, and it sits in the same package as the new guarded reader. Weave's compile and `cmd/datatype` use it.
     - **Rule:** every `construct/deps` read goes through `layergraph.ReadDeclaration`. Route `OSFS` declaration reads through it, and add a guard test that no non-test code opens `construct/deps` any other way.
   - `pkg/layergraph/read.go` has no colocated test. Its only coverage is indirect, through weave's acquire.
   - Carried over: `sameOrigin` re-reads the primary's origin for every clone.

5. **Test coverage**
   - BR-9: the `null slots` mutation now reaches `validateSlots`' nil check (`types.go:425`) without also adding an unknown key.
   - BR-8 (reader): the shared reader is pinned by weave's limits table. No fleet test puts a FIFO at `construct/deps`. That is acceptable because the fleet side is a one-line delegation.
   - BR-10: none (see above).
   - BR-11: a structural change with no behavior change, so no test is needed.

6. **Architecture**
   - **ARCH-DRY:** flagged. The reader is unified, but the walk is not, and `Walk` still uses an unguarded reader.
   - **ARCH-PURE:** pass. BR-11 is fixed and `DeclaredMembers` is pure over read and stat.
   - **ARCH-PURPOSE:** flagged, mildly. BR-8's sweep ("search for other duplicates") was done for the reader but not for the walk or `layergraph.Walk`.
   - **ARCH-MOCK:** pass. The tests use real git through the `GitReader` seam.
   - **ARCH-CONSTRAINTS:** pass. Each present member adds one `EvalSymlinks` and no git process.
   - **ARCH-SECURE:** flagged, Minor. The environment-membership check runs on the lexical path while trust is decided on the canonical one. `Walk`'s reader is unguarded.
   - **ARCH-ORDER:** pass. This is a single-shot observation, and the help text tells callers to re-check the verdict at action time.
   - **ARCH-FUNERAL:** pass. Nothing durable is created.

7. **Plan revisions:** the M2 BR-8 Revision entry says "present members are canonicalized before matching rows". It should add either "the walk stays separate from `acquire.Restore` because …" or the move of the walk core into `pkg/layergraph`. It should also note that the environment check now runs on canonical paths, if that gets fixed.

```findings
dispose:
  - id: BR-8
    disposition: not-addressed
    note: |
      The reader half is fixed (layergraph.ReadDeclaration shared, weave limits_test pins it). Still open: the substrate walk is neither shared with acquire.Restore nor recorded as separate, and DeclaredMembers dedups and checks environment membership lexically (membership.go:59-75), so a member symlinked to a :0 checkout counts as in-environment and takes that checkout's verdict, a layout weave's canonical policy refuses.
  - id: BR-9
    disposition: addressed
    note: |
      editJSON sets slots to null alone (slots_test.go:195), so validateSlots' nil check (types.go:425) is what fails; the sweep also covered the claims contract mutations.
  - id: BR-10
    disposition: not-addressed
    note: |
      The canonicalization was added (inventory.go:220-235) without a regression test; no fleet test creates a symlinked member or environment, so removing it stays green. The lexical visited set can also yield duplicate members after canonicalization.
  - id: BR-11
    disposition: addressed
    note: |
      read and stat are now parameters of collectDependencyRows (inventory.go:208), passed from CollectInventory.
findings:
  - id: new
    severity: Minor
    family: single-source-marker-list
    title: |
      pkg/layergraph.Walk still reads construct/deps via OSFS os.ReadFile, beside the new guarded reader
    detail: |
      3rd finding in family. Rule: every construct/deps read goes through layergraph.ReadDeclaration. walk.go:96 and fs.go:43 follow symlinks, block on a FIFO and are unbounded (used by weave compile and cmd/datatype). Fix: route OSFS declaration reads through ReadDeclaration and add a guard test that no non-test code opens construct/deps another way.
  - id: new
    severity: Minor
    family: test-placement
    title: |
      pkg/layergraph.ReadDeclaration has no colocated test; covered only through weave acquire's Restore
```
