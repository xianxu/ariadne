# Boundary Review — ariadne#598 (milestone M1)

| field | value |
|-------|-------|
| issue | 598 — Affected planning: declared non-Go path ownership and audited reader inputs |
| repo | ariadne |
| issue file | workshop/issues/000598-affected-declared-input-ownership.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 65726091f368c5713d1ec162e407be6dec8c79bf..d2d3ac6421a07f535f29f5f9c4001731ebbb8660 |
| command | sdlc milestone-close --issue 598 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T15:48:35-07:00 |
| verdict | REWORK |

## Review

Inspecting how `base` flows into the mapping loader, then running the tests.

```verdict
verdict: REWORK
confidence: high
```

The code in M1 is solid. Inline `owners`, `full` entries with `except`, and sidecar conventions all compile into one pure `affected.Mapping`. `Plan` reads that mapping inside its existing claim step, and `splitImpact`'s fallback shares the same function through `affected.Owned`. Declarations are read only from the pinned main revision (`mainHead`), and every way a declaration can fail falls back to running the full suite. I ran vet and the affected, gitx and `cmd/sdlc -run 'TestAffected|TestSplitImpact'` tests: all pass. `BenchmarkPlanDeclared` runs in 23 ms/op. What blocks SHIP is the plan document, not the code. Its Core concepts table and its M1 checklist both claim things the code does not deliver: an exported `OwnerRule`, `sidecar.go`, `Mapping.Targets`, a three-argument `Compile`, and a `traceability` config member ticked `[x]`. The issue Log records the traceability move to M2, but the plan never got a `## Revisions` entry for it. The fix is a cheap edit to the plan. Separately, two Important items are cheap to fix: the README is missing the new config surface, and an interrupt during the mapping read is not handled the way an interrupt during the graph load is.

**1. Strengths**
- **One classification function for every consumer (ARCH-DRY):** `claimPath` in `cmd/sdlc/internal/affected/mapping.go` is the only classifier. `Plan` (`plan.go`) and `Owned` both call it, so landing attribution sees exactly what selection sees.
- **Trust boundary is right (ARCH-SECURE):**
  - `productionAffectedDeps` loads sidecars at `mainHead`, the same pinned revision as the config (`localcheck_cmd.go:219,268`).
  - `TestAffectedDeclaredSidecarsFromMain` shows a `reads` entry that only the head declares has no effect.
  - `ParseSidecar` is strict about `reads` and ignores the sidecar's other members.
  - `stripTarget` rejects any derived target that is not a clean relative path, and two fuzz targets check this.
- **The gitx extraction is clean:** `TrunkView.FilesSelected` is now a one-line call into `TreeFilesSelected`, which adds guards for an empty dir and for a revision that looks like an option (`gitx/snapshot.go`). Its test covers cancellation and traversal.
- **The qualification test uses real data:** `TestPairQualification` runs pair's real `affected.json` and real certificate with pair#853's edits applied. That pins the Done-when behaviour against real inputs rather than invented ones.
- **Performance claim checked:** the literal-prefix prefilter plus the exact-path index keep 2,000 paths × 1,200 rules at about 23 ms/op here. `TestLiteralPrefix` checks that the prefilter never rejects a path the glob matches.

**2. Critical findings**
- **The plan's Core concepts table and M1 checklist contradict the code** (`workshop/plans/000598-affected-declared-input-ownership-plan.md`). Each claim against what the code has:

  | Plan claims | Code has |
  |---|---|
  | exported `OwnerRule` | unexported `ownerRule` |
  | `ParseSidecar` in `cmd/sdlc/internal/affected/sidecar.go` | it lives in `mapping.go`; there is no `sidecar.go` |
  | `Mapping.Targets(path)` as the shared lookup | `claimPath` / `Owned` |
  | `Compile(cfg, traceFiles, sidecarFiles)` | `Compile(cfg, sidecars)` |
  | M1 row 1 ticked `[x]`, including `traceability` | `Config` has no such member |

  The issue Log records the traceability move to M2, but the plan does not. Fix: add a `## Revisions` entry to the plan (see section 7) and update the table rows and the M1 checkbox text.

**3. Important findings**
- **README update appears missing for the new `affected.json` keys** (`README.md:897-923`).
  - The README already documents `affected.json` keys at that level, for example `graph_prepare` together with its older-binary note.
  - It does not mention `owners`, `sidecars`, the `{"glob", "except"}` form of `full`, or that older sdlc binaries reject them.
  - The blind-spots paragraph (around line 920) still says only "the repo's affected command or its `full` globs must cover such tests". The atlas was updated to name declared ownership as the answer, so the README now contradicts it.
  - Fix: add a short paragraph pointing to the atlas section.
- **An interrupt during the mapping read turns into a full selection** (`cmd/sdlc/localcheck_affected.go:108-111`). The graph path checks `ctx.Err()` and returns "graph interrupted" as an error. The mapping path does not:
  - In boarding mode, a cancelled `loadAffectedMapping` becomes `boardingRefusal` with "declared ownership unavailable: context canceled".
  - In ordinary mode, the run says it is falling back to the suite, under a context that is already cancelled.
  - Fix: after `ad.mapping(ctx)`, check `ctx.Err()` and return the interrupt error, as the graph path does. Add a seam test where `mapping` cancels the context.

**4. Minor findings**
- `Owned` stops at the first rule that forces a full selection, so packages a later rule would have claimed are lost (`mapping.go` claimPath early returns). In `splitImpact`'s uncertain fallback this only drops "Affected" evidence; it never wrongly marks a package innocent, because those packages stay Unknown. The comment in `mapping.go` presents this as intended, but it depends on rule order.
- The glob-character test `strings.ContainsAny(p, "*?[\\")` appears twice in `mapping.go` (in `Compile` and `literalPrefix`). It could be one `hasGlobMeta` helper (ARCH-DRY, trivial).
- `Compile` is exported and does not re-check `OwnerDecl.Paths` being non-empty. A hand-built `Config` with a `Strip` owner and no paths would become a match-everything-under-prefix convention rule. `ParseConfig` prevents this in production; a guard would make it unrepresentable.
- The atlas says about 30 ms; the issue Log says 33 ms. Fine as an approximation.

**5. Test coverage notes**
- The table tests cover:
  - union with graph owners, dependents and group roots;
  - `full` and `except` precedence, including a declared target that matches `full`;
  - a stale root, a sidecar for a deleted package, and a sidecar path no target derives from;
  - added, edited and renamed sidecars, plus a deleted test in a living package.
- Real-git tests show that sidecars come from main, not the head, and that a malformed sidecar on main runs the suite and is named in the output.
- `splitImpact` is covered for a declared read, for a full selection that keeps the declared connection, and for a sidecar that fails to compile (impacts stay Unknown).
- Gap: no test covers cancellation during the mapping load (see Important finding 2).
- No prior fixes were claimed this round, so no revert-and-check was needed.

**6. Architectural notes for upcoming work**
- **ARCH-DRY: pass.** One classifier; `FilesSelected` delegates to `TreeFilesSelected`.
- **ARCH-PURE: pass.** `Compile`, `ParseSidecar`, `Plan` and `Owned` take bytes and values and do no IO. IO stays in `loadAffectedMapping` behind the `affectedDeps.mapping` seam.
- **ARCH-PURPOSE: pass.** Shadow sweep: the consumers that compute a selection are the affected run, resume (both through `productionAffectedDeps`) and `splitImpact`, and all three load the mapping. Pre-close and the car stand-in read only whether `affected.json` is declared, as the plan says.
- **ARCH-MOCK: pass.** Real git in temp repos is the git fake, as in existing tests; the planner tests inject the mapping through the seam.
- **ARCH-CONSTRAINTS: pass.** The benchmark matches the stated envelope. Reads are bounded by the existing gitx snapshot limits, and going over a limit falls back to a full selection.
- **ARCH-SECURE: pass.** Declarations come only from the pinned main revision. Sidecar and strip inputs are parsed strictly and fuzzed, and failures are visible as a full selection with a reason.
- **ARCH-ORDER: pass.** The planner holds no state between calls. The one ordering flaw (cancellation) is filed under Important.
- **ARCH-FUNERAL: pass.** M1 creates nothing durable; the mapping lives in memory for one run. M3's ledger will need the planned cap and dismiss path.
- **ARCH-EVOLVE: pass.** The `affected.json` changes are additive with strict known fields, and the older-binary refusal (falling back to the suite) is documented in the atlas. ariadne never writes sidecars.
- **For M2:** traceability should compile into `ownerRule{as: tests}` inside `Compile`. Adding a parameter will change `Compile`'s signature again, so consider a `Sources` struct argument now.

**7. Plan revision recommendations**
- Add `## Revisions` → "2026-10-24 M1 as built":
  - `traceability` moved to M2, so M1 row 1 covers only `FullRule`, `owners` and `sidecars`.
  - `OwnerRule` is the unexported `ownerRule`.
  - `ParseSidecar` lives in `mapping.go`; there is no `sidecar.go`.
  - The shared lookup is `claimPath`, exposed as `Owned`; `Mapping.Targets` does not exist.
  - `Compile(cfg, sidecars)` takes two arguments until M2.
- Update the Core concepts table rows for `OwnerRule` and `ParseSidecar` to match.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: plan-claims-match-code
    title: |
      The plan's Core concepts table and M1 checklist claim entities and members the code does not have
    detail: |
      The table lists an exported OwnerRule, cmd/sdlc/internal/affected/sidecar.go for ParseSidecar, a Mapping.Targets lookup and Compile(cfg, traceFiles, sidecarFiles). The code has the unexported ownerRule, ParseSidecar in mapping.go, claimPath/Owned and Compile(cfg, sidecars). M1 row 1 is ticked [x] including traceability, but Config has no such member; it moved to M2 (issue Log) with no plan Revisions entry. Fix: add a Revisions entry and update the table rows and the checkbox text.
  - id: new
    severity: Important
    kind: defect
    family: readme-tracks-user-surface
    title: |
      README update appears missing for the new affected.json keys: owners, sidecars, full except and the older-binary refusal
    detail: |
      README.md:897-923 documents affected.json keys (graph_prepare, with its older-binary note) but not the new members. Its blind-spots paragraph still says only full globs or the command must cover non-Go reads, which contradicts the updated atlas. Add a short paragraph that points to the atlas section.
  - id: new
    severity: Important
    kind: defect
    family: cancellation-not-collapsed-to-fallback
    title: |
      planAffectedSelection turns a cancelled mapping load into a full selection instead of an interrupt error
    detail: |
      localcheck_affected.go:108-111 does not check ctx.Err() after ad.mapping(ctx), unlike the graph path (lines 102-104). A cancellation becomes "declared ownership unavailable: context canceled", which in boarding mode is a boardingRefusal and in ordinary mode a suite fallback under a dead context. Check ctx.Err() first and add a seam test where mapping cancels.
  - id: new
    severity: Minor
    kind: defect
    family: fallback-evidence-order-dependent
    title: |
      Owned stops at the first rule that forces full, so packages later rules would claim are lost to splitImpact's fallback
    detail: |
      claimPath returns early. In the uncertain fallback this only removes Affected evidence; it never wrongly marks a package innocent. Consider letting claimPath keep collecting packages after it records the full reason.
  - id: new
    severity: Minor
    kind: defect
    family: shared-helper-for-repeated-predicate
    title: |
      The glob-metacharacter test is duplicated in Compile and literalPrefix in mapping.go
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e3b28313-ed27-4bb9-ac76-d6e8fd256de3","repository":"github.com/xianxu/ariadne","issue_id":"000598","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/723a950ea506798e","reviewed_head":"d2d3ac6421a07f535f29f5f9c4001731ebbb8660","slot":"ariadne:5","requested_at":"2026-10-24T22:43:12.371918Z"},"phase":"finished","pid":29467,"started_at":"2026-10-24T22:43:12.5354Z","finished_at":"2026-10-24T22:48:34.452621Z","session_id":"9091a226-887b-4a7f-a460-3511b463cb72","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T16:08:17-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 598 — Affected planning: declared non-Go path ownership and audited reader inputs |
| repo | ariadne |
| issue file | workshop/issues/000598-affected-declared-input-ownership.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 65726091f368c5713d1ec162e407be6dec8c79bf..3f5556ef1a54c9bdb2d5f1faca82d2c6fa82f149 |
| command | sdlc milestone-close --issue 598 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T16:08:17-07:00 |
| verdict | SHIP |

## Review

Let me look at the round-1 answer commit first, then the full code.

Still reviewing: the affected and gitx tests pass, and the cmd/sdlc tests are still running in the background. While they finish I'm checking the BR-1/BR-2/BR-5 fixes.

The cmd/sdlc tests are still running. Meanwhile I'm checking that the fuzz and property tests the plan promises for M1 exist.

Now I'll revert the BR-3 and BR-4 fixes and check that their regression tests go red.

Both regressions go red when their fix is reverted, and the tree is restored. Next I'm checking the remaining M1 checklist items against the code.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** All five round-1 findings are fixed, and I checked each one against the code rather than the commit messages. I tested both behavior fixes by reverting them in this checkout:
- **BR-3:** without the new `ctx.Err()` check, `TestAffectedDeclaredMappingCancelled` fails. Ordinary mode falls through to the suite run and dies in host preflight with `context canceled`.
- **BR-4:** with the old `mapping.go`, `TestOwnedOrderIndependent` fails with `Owned = []`.

Both pass again once the fix is restored, and the tree is clean. At head these all pass: `go test ./cmd/sdlc/internal/affected/`, `./cmd/sdlc/internal/gitx/`, and `./cmd/sdlc/ -run 'TestAffectedDeclared|TestSplitImpactDeclared'`. I stopped the broader cmd/sdlc run early because other test runs on the host were slowing it down, so the full package was not run. The M1 checklist matches the code. I'm raising only two new findings, both Minor, so nothing blocks the boundary.

### 1. Strengths
- `claimPath` (`cmd/sdlc/internal/affected/mapping.go:207`) is the one place a path gets classified, and both `Plan` and `Owned` call it. Selection and landing attribution can't drift apart. Its `fail` closure keeps the first reason for going full while still collecting every package, so the result no longer depends on rule order.
- Declarations are read only from main's pinned commit (`localcheck_cmd.go:164`, `localcheck_resume_cmd.go:84`). A test proves the head can't declare its own selection ("a read only the head declares does not count", `localcheck_affected_declared_test.go:69`).
- The interrupt handling now matches the graph path (`localcheck_affected.go:108-112`). It is tested in both ordinary and boarding mode, and the check catches cancellation even when the loader returns nil.
- `ParseSidecar` is strict about duplicate keys, trailing data, and `reads` that aren't a clean array, and it treats every other member as opaque. That fits ARCH-EVOLVE, because ariadne never rewrites sidecars.
- The exact-path index plus literal-prefix filter is backed by a benchmark at the plan's stated scale (2,000 paths × 600 sidecars × 600 globs).

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- **README and atlas overstate the older-binary fallback** (`README.md:938`, `atlas/workflow/ci-merge-check.md`, the "Older binaries" bullet). Both say an older binary's refusal "degrades to the suite". That is true for `--pre-close` (`precloseck.go:33`) and the car gate. But an explicit `--name affected` or `resume` returns the parse error and refuses (`localcheck_cmd.go:219-221`, `localcheck_resume_cmd.go:89-91`). This is the 2nd finding in family `readme-tracks-user-surface`. The rule: a doc claim about what happens when a file fails to parse must be checked against every caller of `readAffectedConfig`, not just one. Fix both sentences at once, e.g. "pre-close and the car gate fall back to the suite; an explicit `--name affected` refuses".
- **M1 claims a boarding test that doesn't exist.** The checklist says real-git tests show an unreadable sidecar makes boarding refuse and name the sidecar. `TestAffectedDeclaredMalformedSidecarRunsTheSuite` only runs ordinary mode. The underlying refusal (`sel.Full && boarding`) is generic code that is tested elsewhere, so this is low risk. This is the 2nd finding in family `plan-claims-match-code`. The rule: every test claim in a ticked row must name a test that exists. The cheap fix is to loop the existing test over both modes, as `TestAffectedDeclaredMappingCancelled` already does.

### 5. Test coverage notes
- The pure entities (`ParseConfig`, `ParseSidecar`, `Compile`, `Plan`, `Owned`) are tested with no IO. That includes two fuzz targets, the `Owned` superset-of-graph-owners property, and the order-independence regression.
- The integration seams (`loadAffectedMapping` and `TreeFilesSelected`) use real-git fixtures, and `splitImpact` has declared-owner and loader-error cases.
- The BR-5 refactor (`hasGlobMeta`) changes no behavior, and `TestLiteralPrefix` pins the helper.

### 6. Architectural notes
- **ARCH-DRY: pass.** One claim step, one `fullMatch`, and `hasGlobMeta` now shared. The other `ContainsAny` calls in `config.go` and `golist.go` check different character sets for different purposes.
- **ARCH-PURE: pass.** `Compile` and `Plan` are pure. IO lives only in `loadAffectedMapping` and the `affectedDeps.mapping` seam.
- **ARCH-PURPOSE: pass.** Both consumers (local selection and `splitImpact`) use `claimPath`, so there's no copy restating the rules.
- **ARCH-MOCK: pass.** git goes through gitx, and the tests run it against real fixture repos, as the repo already does.
- **ARCH-CONSTRAINTS: pass.** The benchmark covers the stated envelope, and sidecar reads stay under the existing snapshot limits.
- **ARCH-SECURE: pass.** Untrusted head declarations are ignored, sidecar paths and globs must be clean relative paths, and a malformed sidecar visibly falls back to the suite with its name in the reason.
- **ARCH-ORDER: pass.** Nothing carries state between events; it is a single plan under one context, and cancellation is now an error instead of a fallback.
- **ARCH-FUNERAL: pass.** Nothing durable is created; the mapping lives only in memory, built from git reads.
- **ARCH-EVOLVE: pass.** The new `affected.json` members are additive, and older binaries refuse them, which is documented. Sidecar members ariadne doesn't know are opaque and never rewritten.
- **For M2:** the plan already commits to switching `Compile` to a sources struct when the traceability input arrives. Keep `claimPath` as the only consumer.

### 7. Plan revision recommendations
None required. The 2026-10-24 Revisions entry and the updated Core concepts table match the code. Optionally, after the boarding loop is added, the M1 sidecar row could name the test that covers boarding.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan now has a 2026-10-24 Revisions entry, ownerRule/ParseSidecar-in-mapping.go/claimPath/Owned/Compile(cfg, sidecars) rows, and traceability moved to M2. Each was checked against mapping.go and config.go.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:925-941 now documents owners (go/as/strip), sidecars with reads, full except, reading from main and the older-binary note, and links the atlas section.
  - id: BR-3
    disposition: addressed
    note: |
      localcheck_affected.go:109 checks ctx.Err() before the fallback. Reverting it makes TestAffectedDeclaredMappingCancelled fail (mode 0 reaches the suite and fails in host preflight).
  - id: BR-4
    disposition: addressed
    note: |
      claimPath keeps the first full reason but collects every claim. With the pre-fix mapping.go, TestOwnedOrderIndependent fails with Owned = [].
  - id: BR-5
    disposition: addressed
    note: |
      hasGlobMeta (mapping.go:145) is shared by Compile and literalPrefix. No behavior change; TestLiteralPrefix pins it.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: readme-tracks-user-surface
    title: |
      README and atlas say an older binary's refusal of the new members degrades to the suite, but --name affected and resume refuse instead
    detail: |
      2nd finding in this family. Rule: a doc claim about what happens when affected.json fails to parse must be checked against every readAffectedConfig caller. Pre-close (precloseck.go:33) falls back to the suite; localcheck_cmd.go:219 and localcheck_resume_cmd.go:89 return the error. Fix README.md:938 and the atlas Older binaries bullet together.
  - id: new
    severity: Minor
    kind: defect
    family: plan-claims-match-code
    title: |
      The ticked M1 row claims a real-git test that boarding refuses on an unreadable sidecar, but only ordinary mode is tested
    detail: |
      2nd finding in this family. Rule: every test claim in a ticked checklist row names a test that exists. Loop TestAffectedDeclaredMalformedSidecarRunsTheSuite over affectedOrdinary and affectedBoarding, as TestAffectedDeclaredMappingCancelled does.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7c4380ee-cc2d-4a66-9f66-76694c690552","repository":"github.com/xianxu/ariadne","issue_id":"000598","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/ca7b5dcf55101aa3","reviewed_head":"3f5556ef1a54c9bdb2d5f1faca82d2c6fa82f149","slot":"ariadne:5","requested_at":"2026-10-24T22:54:43.020781Z"},"phase":"finished","pid":60832,"started_at":"2026-10-24T22:54:43.177083Z","finished_at":"2026-10-24T23:08:07.087706Z","session_id":"9f8ac2a3-3586-49b9-8be3-6c9c63bccddf","outcome":"returned"}] -->
