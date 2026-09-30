# Boundary Review — ariadne#272 (whole-issue close)

| field | value |
|-------|-------|
| issue | 272 — One issue per branch: no work on an unlanded base |
| repo | ariadne |
| issue file | workshop/issues/000272-stacked-issue-development.md |
| boundary | whole-issue close |
| milestone | — |
| window | 01d0ae03f5d0872d0ac251dbd40dea4ac3d6b0e0..28205b9429bb95aef52da4862adb4887e084885f |
| command | sdlc close --issue 272 |
| reviewer | claude |
| timestamp | 2026-09-29T21:12:59-07:00 |
| verdict | unknown |

## Review

API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL) — the proxy refused the tunnel: check its credentials and that it allows this host

---

## Re-review — 2026-09-29T21:25:17-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 272 — One issue per branch: no work on an unlanded base |
| repo | ariadne |
| issue file | workshop/issues/000272-stacked-issue-development.md |
| boundary | whole-issue close |
| milestone | — |
| window | 01d0ae03f5d0872d0ac251dbd40dea4ac3d6b0e0..28205b9429bb95aef52da4862adb4887e084885f |
| command | sdlc close --issue 272 |
| reviewer | claude |
| timestamp | 2026-09-29T21:25:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change is small and mostly sound. `refuseUnlandedBase` catches the parley topology whether the checkout is already on the issue branch or is switching to it. Merged-in main and landed parents do not count against a branch. The soft instruction appears in AGENTS.base.md §2, the atlas table and `start-plan --help`. The handoff-guard defect was filed as #274 (`259024b6`), and no `change-code`/`close`/publish guard was added, as the operator decided. One clause of `## Done when` is only half-delivered. It says a branch that an unlanded descendant was built on "still passes". That holds only while the parent has not moved since the descendant forked. Once the parent gets another commit, which is exactly the parley #301-fixes-during-#302 pattern, `start-plan` on the parent wrongly refuses. The test does not cover that case.

1. **Strengths**
   - `unlandedSharedBase` (planningbranch.go:131) is a clean three-step predicate. It skips a branch that has landed, skips a branch built on tip, and checks whether the merge base is on main. Because it reads ownership from refs, it does not depend on a commit-subject convention.
   - It checks both entry paths (already on the branch, and switching to an existing one). The refusal runs before `git switch`, and the test confirms HEAD does not move.
   - It checks remote-only parent refs too. A remote copy of this issue's own branch is correctly skipped by name (`other == name`).
   - The table test covers all three refusal topologies plus merged-in main, a landed parent and a descendant.
   - The docs sweep is complete for the new behaviour: constitution, atlas lifecycle row and help text.

2. **Critical**: none.

3. **Important**
   - **The descendant exemption breaks once the parent advances** (planningbranch.go:132–136).
     - The exemption is `tip ⊑ ref` (tip is an ancestor of the other branch). Take parent `000009` with commit A, child `000010` forked at A, then parent commit B.
     - Now tip B is no longer an ancestor of the child, the merge base is A (beyond main), and the parent is refused with "carries unlanded work of 000010-child". But A is 000009's own work.
     - Test gap: the "a descendant built on this branch" case (planningbranch_test.go:189) switches back without committing on the parent, so it only covers the case where the parent has not moved.
     - Fix sketch: decide who owns the shared range `main..base`. Either refuse only when that range holds commits that are not this issue's own (for example, subjects not tagged `#<id>`, per §12), or record the branch's fork point from main at creation and require it to be on main.
     - At minimum, add a test case where the parent commits again after the child forks, and choose a behaviour on purpose.
     - This is the only instance of this rule in the window. The other descendant-shaped input, a child ref sitting exactly on the parent's tip, is already covered.

4. **Minor**
   - **A git error is silently treated as "unrelated histories"** (planningbranch.go:137–140). `err != nil || base == ""` returns `(false, nil)`, so any `merge-base` failure (bad ref, corrupt object) passes the guard silently. Only exit status 1 with empty output should mean "no common ancestor"; other errors should propagate. This is the only swallow site in the diff.
   - **The switch path fetches main twice.** It calls `env.main.Snapshot()` at line 52, then again inside `refuseUnlandedBase`. Passing the already-pinned ref in would give one fetch and one pinned view (ARCH-DRY). The already-on-branch path now also does a network fetch it did not do before, so a re-run of `start-plan` there can now fail offline.
   - **Cost grows with the number of refs.** It spawns up to four git processes per `NNNNNN-*` ref across `refs/heads` and `refs/remotes/<remote>`. That is fine today, but it will be slow on remotes with many stale issue branches.
   - **`path.Base(ref)` drops a namespace prefix.** `refs/heads/wip/000003-x` is treated as issue branch `000003-x`. This is harmless, but the name check is looser than the comment on `issueBranchRE` suggests.

5. **Test coverage**
   - All three refusal cases and three pass cases are pinned against a real git fixture, with no mocks.
   - Missing: a parent that advanced after a descendant forked (above), and a `merge-base` error path.
   - Done when's "Creating a new issue branch keeps branching at freshly pinned main" relies on existing tests; this diff does not change that path.

6. **Architecture**
   - **ARCH-DRY: pass**, with the double-Snapshot nit above. Main is pinned by reusing `trackerEnv`.
   - **ARCH-PURE: pass.** This is thin git glue in an IO-shell file, and the predicate is small. It could be a pure function over ancestry facts, but that isn't worth doing at this size.
   - **ARCH-PURPOSE: pass**, apart from the descendant false positive. The issue's purpose (one issue per branch, instruction plus a `start-plan` gate) is delivered, and the deferrals are genuinely separable (#273, #274, pair#352, pair#353).

7. **Plan revisions**: if the fix takes the narrower route, add a `## Revisions` entry. It should say that the descendant exemption covers only a parent that has not moved since the descendant forked, and record the ownership rule chosen for a parent that has advanced.

```findings
findings:
  - id: new
    severity: Important
    family: stack-ownership-by-ancestry
    title: |
      Descendant exemption fails once the parent advances: start-plan refuses the parent of an unlanded child
    detail: |
      unlandedSharedBase exempts ref only when tip is an ancestor of ref. If parent 000009 commits again after child 000010 forked, the merge base is the fork point beyond main, so the parent is refused as carrying the child's work, which contradicts the Done-when clause that a branch an unlanded descendant was built on still passes. The test case at planningbranch_test.go:189 never commits on the parent after the fork. This is the only instance of the family in the window. Fix by deciding who owns main..merge-base (commit issue tags, or the recorded fork point) and add the parent-advanced regression case.
  - id: new
    severity: Minor
    family: silent-error-swallow
    title: |
      merge-base failure is treated as unrelated histories and passes the guard
    detail: |
      planningbranch.go:137-140 returns (false, nil) on any merge-base error. Only exit 1 with empty output means no common ancestor; other errors should propagate. This is the only swallow site in the diff.
  - id: new
    severity: Minor
    family: redundant-main-snapshot
    title: |
      Switch path fetches main twice; the already-on-branch path adds a new fetch
    detail: |
      preparePlanningBranch pins main at line 52, then refuseUnlandedBase calls env.main.Snapshot() again. Pass the pinned ref in. Note the already-on-branch path now needs the network.
```

---

## Re-review — 2026-09-29T21:30:07-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 272 — One issue per branch: no work on an unlanded base |
| repo | ariadne |
| issue file | workshop/issues/000272-stacked-issue-development.md |
| boundary | whole-issue close |
| milestone | — |
| window | 01d0ae03f5d0872d0ac251dbd40dea4ac3d6b0e0..adfa4250aa0bf95b3835cc2ec2da8ae6493c8399 |
| command | sdlc close --issue 272 |
| reviewer | claude |
| timestamp | 2026-09-29T21:30:07-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All three prior findings are resolved, and the change delivers what `## Done when` asks for. Round 1 decided whose work a commit is by looking at which branch it came from. That rule refused a parent branch once the parent moved past the point where its child forked. The code now reads ownership from the commit's `#N` subject tag, which fixes this. Regression coverage: the case "a child built on this branch, which then advanced" (`planningbranch_test.go`). The merge-base call that swallowed errors is gone; the `rev-list`, `for-each-ref` and `log` calls in `refuseUnlandedBase` all return their errors. The switch path now reuses `pinned`. The branch you are already on still fetches main fresh, which the "beyond fresh main" wording in Done-when requires. `go test ./cmd/sdlc -run TestPreparePlanningBranch` passes. The docs changes (constitution line, atlas row, `start-plan --help`) match the code. What remains is minor: one false positive in how the tag is matched, a small duplicated regex, and one pass mode with no test.

1. **Strengths**
   - `planningbranch.go:103-138`: ownership comes from the `#N` tag, and the doc comment says plainly that an untagged shared commit is left to the soft instruction. The limit is stated, not hidden.
   - `issueTagRE` handles leading zeros (`#000003`) and rejects longer numbers (`#31` does not match `#3`), and the tests pin this. It also rejects cross-repo refs like `parley#3`, because an alphanumeric character before the `#` blocks the match.
   - Refusal is tested in both modes (already on the branch, and switching to it). Each refusal test checks that HEAD did not move.
   - A remote-only parent branch is covered too (`refs/remotes/<remote>`).
   - The lesson entry in `workshop/lessons.md` names the whole class (inferring ownership from git topology), not just this one site.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `issueTagRE` matches a `#N` anywhere in the subject. Suppose this branch's own commit is "#9: prep hook for #10" and an unlanded 000010 branch was built on this branch. That commit is then counted as #10's work, and #9 is refused, even though Done-when says a branch that an unlanded descendant was built on still passes. **Fix:** treat a commit as owned by the planning issue if its subject carries that issue's own tag (skip `onTip` commits that match `issueTagRE(id)`). This is the only instance in the window.
   - ARCH-DRY: `issueBranchRE` (`planningbranch.go:95`) repeats the pattern of `issueFamilyRE` (`migrate.go:178`) with an added capture group. Use one shared regex with the capture group.
   - `issueTagRE` recompiles a regexp for every issue-branch ref. This is not a hot path; hoisting it is optional.

5. **Test coverage**
   - Every passing case (main merged in, a child built on this branch, a landed parent, a `#31` stray commit) runs only on the already-on-branch path.
   - The switch path (`refs/heads/<name>` with `pinned`) is tested only for refusal. Add one passing switch case, for example an existing branch that merged main, checked out from rest.

6. **Architecture**
   - **ARCH-DRY:** flag (Minor). The regex duplication above; nothing else is duplicated.
   - **ARCH-PURE:** pass. `refuseUnlandedBase` does git IO interleaved with a small decision. It could be split into gathering the data plus a pure check, but at this size the integration tests using a real repo fake are the right tool.
   - **ARCH-PURPOSE:** pass. Done-when explicitly limits the gate to `start-plan` (operator decision), and the soft-instruction docs are delivered. Filing the handoff-guard defect is recorded in the Plan and cited as #274 in `lessons.md`.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Ownership is now read from #N tags; the case "a child built on this branch, which then advanced" pins the parent-advanced scenario and passes.
  - id: BR-2
    disposition: addressed
    note: |
      The merge-base call was removed; rev-list, for-each-ref and log errors all propagate from refuseUnlandedBase.
  - id: BR-3
    disposition: addressed
    note: |
      The switch path passes pinned in. The already-on-branch Snapshot is needed because Done-when checks against fresh main.
findings:
  - id: new
    severity: Minor
    family: commit-owner-leading-tag
    title: |
      A #N mention anywhere in this branch's own commit subject counts it as another issue's work
    detail: |
      issueTagRE matches anywhere in the subject, so a commit "#9: prep hook for #10", shared with an unlanded 000010 branch built on 000009, refuses #9. This breaks the Done-when clause for a descendant built on this branch. Fix: skip commits whose subject also carries the planning issue's own tag. This is the only instance in the window.
  - id: new
    severity: Minor
    family: shared-issue-id-regex
    title: |
      issueBranchRE repeats issueFamilyRE (migrate.go:178) with an added capture group
    detail: |
      ARCH-DRY: use one NNNNNN- prefix regex with a capture group for both. These are the only two instances.
  - id: new
    severity: Minor
    family: two-mode-clause-coverage
    title: |
      The switch-to-existing-branch path has no passing test case
    detail: |
      Every passing case runs on the already-on-branch path; the switch path is tested only for refusal. Add one passing switch case, e.g. an existing branch that merged main, checked out from rest.
```

---

## Re-review — 2026-09-29T21:35:33-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 272 — One issue per branch: no work on an unlanded base |
| repo | ariadne |
| issue file | workshop/issues/000272-stacked-issue-development.md |
| boundary | whole-issue close |
| milestone | — |
| window | 01d0ae03f5d0872d0ac251dbd40dea4ac3d6b0e0..75a79140e8e648d64f797b8e22c21e87611acd32 |
| command | sdlc close --issue 272 |
| reviewer | claude |
| timestamp | 2026-09-29T21:35:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The `start-plan` guard does what `## Done when` asks, and I checked this by running it. At `75a79140`, `TestPreparePlanningBranchRefusesAnUnlandedBase` passes all 10 cases. It covers the parley topology, which is refused on both the already-on-branch path and the switch path, plus a parent that exists only on the remote. The pass cases are: main merged in (on both paths), a child built on this branch, a child whose parent then advanced, `#31` against `#3`, the `#9: prep hook for #10` mention, and a landed parent. The constitution line, the atlas row, the help text and the lesson are all present. No `change-code`, `close` or publish guard was added, as the operator decided. All three open Minors are fixed.

The fix for BR-4, though, created a second definition of "which issue owns a commit subject" (`commitIssue`/`commitTagRE`). `gitx` already declares itself the single source for that rule (`issueSubjectDescriptor`/`subjectAnchorRE`), and the two rules disagree. That is non-blocking, but it is cheap to fix and it can still refuse a parent by mistake.

**1. Strengths**
- The rewritten lesson and design: ownership now comes from the `#N` subject tag, not from ancestry. That removes the whole BR-1 class, not just one site of it (`planningbranch.go:94-100`, lessons.md).
- On a refusal, HEAD does not move. The test checks this directly, and the check runs before `git switch` (`planningbranch.go:74`).
- Commits that main already has are excluded (`^mainTip`), so merging main in stays legal on both paths.
- The branch-prefix regex is now shared with `migrate.go:179`.

**2. Critical:** none.

**3. Important**
- `cmd/sdlc/planningbranch.go:137-147` — `commitIssue` re-implements commit-subject ownership. `cmd/sdlc/internal/gitx/window.go:212` (`issueSubjectDescriptor`, documented at `:248` as "the single source for the 'commit subject opens with #N' anchor") already defines it, with different rules:
  - `gitx` treats a loose `docs: mention #10 …` as owned by nobody.
  - `commitIssue` gives that commit to #10.
  - As a result, a parent #9 whose own commit is `docs: mention #10 hook`, and whose unlanded child `000010-*` is built on it, is refused. This is the BR-1 and BR-4 false-parent-refusal class again.

  **This is the 2nd finding in family `commit-owner-leading-tag`.** The rule that covers all of them: a commit's owning issue is decided in exactly one place, the `gitx` subject anchor (`#N …` or `<area>: #N …`, optionally with `close`). Every consumer calls it.
  - Fix: export a `gitx` owner check (e.g. `gitx.SubjectOwnedBy(issueNum, subject)`, wrapping `issueSubjectDescriptor(…, true)`) and delete `commitTagRE`/`commitIssue`.
  - Keep leading-zero tolerance by normalizing the id before the call. `#000003` is used in a test, while the `gitx` anchor matches the literal digits only.
  - Add a pass case for `docs: mention #10`.
  - Instances in this window: only `planningbranch.go:137-147`.

**4. Minor**
- An issue with id `000000` would give `TrimLeft("000000","0") == ""`, which equals `commitIssue` of an untagged subject, so untagged shared commits would count as #0's. The fix above removes this.
- `path.Base(ref)` treats `refs/heads/foo/000003-x` as an issue branch. That is probably intended; noted only.

**5. Test coverage notes**
- Every `## Done when` clause is exercised, both refuse and pass, on both paths. BR-6's switch-path pass case is at test line ~215.
- Missing: a pass case where a loose mention of another issue sits under an area prefix (see Important).

**6. Architectural notes**
- **ARCH-DRY:** flagged, as above.
- **ARCH-PURE:** pass. `refuseUnlandedBase` mixes `git` calls with the decision, but the ownership decision is a pure helper. Reusing the `gitx` pure function keeps it that way.
- **ARCH-PURPOSE:** pass. Only `start-plan` enforces the rule, which is what the operator chose; the soft instruction covers the other gates.

**7. Plan revision recommendations:** none. The Log already records the tag-ownership pivot.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      commitIssue takes the first #N; case "a child whose parent's own commit mentions it" would refuse under the old anywhere-match; test passes at 75a79140.
  - id: BR-5
    disposition: addressed
    note: |
      issueFamilyRE (migrate.go:179) now captures the id and is reused at planningbranch.go:116; issueBranchRE removed.
  - id: BR-6
    disposition: addressed
    note: |
      Case "switching to a branch that merged main in" starts from rest main with an existing issue branch and passes through refuseUnlandedBase then git switch.
findings:
  - id: new
    severity: Important
    family: commit-owner-leading-tag
    title: |
      commitIssue duplicates gitx's single-source subject-ownership anchor with divergent rules
    detail: |
      2nd in family commit-owner-leading-tag. Rule: a commit's owning issue is decided only by gitx issueSubjectDescriptor (window.go:212, "#N ..." or "<area>: #N ..."); export it (e.g. gitx.SubjectOwnedBy, normalizing leading zeros) and delete commitTagRE/commitIssue (planningbranch.go:137-147). Divergence: "docs: mention #10 hook" on parent #9 is owned by #10 here, so #9 is falsely refused when an unlanded 000010 child is built on it; add that as a pass case. Only instance in this window.
```
