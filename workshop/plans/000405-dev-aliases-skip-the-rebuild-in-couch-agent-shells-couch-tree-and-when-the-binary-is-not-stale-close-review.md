# Boundary Review — ariadne#405 (whole-issue close)

| field | value |
|-------|-------|
| issue | 405 — dev-aliases: skip the rebuild in Couch agent shells (COUCH_TREE) and when the binary is not stale |
| repo | ariadne |
| issue file | workshop/issues/000405-dev-aliases-skip-the-rebuild-in-couch-agent-shells-couch-tree-and-when-the-binary-is-not-stale.md |
| boundary | whole-issue close |
| milestone | — |
| window | f0b5eb265f46a143d783d6bd39e836e71d0f265a..d6c05eb0e036548762023a566c4361365ab54ca1 |
| command | sdlc close --issue 405 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T21:02:37-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass (55 of 55) and I've checked the local `replace` directives. Writing up the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. `construct/dev-aliases.sh` now emits a guard that skips the build: `[ -x bin/X ] && { COUCH_TREE set || nothing newer than bin/X }`. It sits in front of the #355 build-to-temp-and-rename block, which is unchanged, and in front of the #299 `dev-prepare` step, so a skipped build skips both. Every Done-when clause is tested in both states against a counting stub `go`/`make`: missing builds, fresh skips, stale builds, Couch with a stale binary skips, Couch with a missing binary builds, and `.git`/`bin` are excluded. The test file passes 55 of 55 in this checkout. Two problems keep this from SHIP, and both are cheap to fix. First, the atlas says a slot "never writes" to `:0`, but the code (and a test) builds into `:0` from a slot when the binary is missing. Second, the freshness scan only looks inside the owner. An owner that pulls in a sibling repo through a local `go.mod` `replace` (metis does: `replace github.com/xianxu/ariadne => ../ariadne`) now keeps running a stale binary after that sibling is edited. Before this change it was rebuilt on every call.

**1. Strengths**
- `construct/dev-aliases.sh:285` defines the guard once as `skip` and splices it into the single `printf`, so the plain form and the `dev-prepare` form share one guard. The #355 atomic-rename body is untouched.
- The tests at `construct/scripts/test/dev-aliases.test.sh:282-322` set file times with `touch -t`, so no case depends on timestamp resolution. The stub `go`/`make` trace counts exactly what ran, including that `dev-prepare` is skipped too.
- The existing #355 concurrency test now pushes a source file's time into the future (`touch -t 2030…`) so its rebuilds still happen. That neighbouring test was disturbed by the change and was updated to match.
- `unset COUCH_TREE` at the top of the test stops a slot's own environment from quietly turning every build test into a skip.

**2. Critical findings**
None.

**3. Important findings**
- **The atlas says slots "never write" to `:0`, but they do** (`atlas/workflow/base-layer.md:93-94`). The atlas says an agent shell in a slot "runs whatever `:0` last built and never writes there". In fact, with `COUCH_TREE` set and `bin/X` missing, the alias builds into `:0`. That is the Spec's "a missing bin/X always builds", and the test "in a Couch slot a missing binary still builds" pins it. Two lines later the same paragraph says "A missing `bin/X` always builds", which contradicts the earlier claim.
  - Instances in this window: `base-layer.md:93-94` ("never writes there") and `atlas/workflow/issue-tracker.md:129` ("the alias never rebuilds in a Couch slot"). The second is defensible but should say "when `bin/X` exists". The header comment in `dev-aliases.sh` is accurate.
  - Fix: reword both, for example "…runs whatever `:0` last built; it writes there only when `bin/X` is missing."
- **The freshness scan misses local `replace` dependencies** (`construct/dev-aliases.sh:285`). The scan is limited to the owner tree. `~/workspace/metis/go.mod:10` has `replace github.com/xianxu/ariadne => ../ariadne`, and metis owns `cmd/metis`. Editing ariadne no longer marks `bin/metis` stale, so the alias silently runs an old binary. The brain repos replace `../nous` too, though they own no `cmd/`. This regression is within the Spec's letter but neither the header nor the atlas mentions it. Only "a toolchain change" is listed as not counting.
  - Fix, either option: (a) while generating, parse `replace … => ../X` (or any relative path) from the owner's `go.mod` and add those directories as extra `find` roots, each pruning its own `.git`/`bin`; or (b) at minimum, document the exclusion next to the toolchain caveat, in both the header comment and the atlas.

**4. Minor findings**
- If a source file is edited during `go build` (after the compiler reads it, before the binary is written), the edit is older than the new `bin/X` and won't trigger a rebuild. The window is narrow; worth noting.
- `find` errors on unreadable subdirectories go to stderr and leave stdout empty, which counts as "fresh". The skip fails open in that case.
- The header comment says the scan takes "about 0.1s on a 7k-file tree"; the Spec says 0.13s on 7.6k files. Harmless.

**5. Test coverage notes**
Every Done-when state is exercised in both modes, and the exact-shape assertions for the plain form, the #355 rename and the #299 `dev-prepare` form were updated. No test covers a dependency outside the owner tree (the second Important finding). If option (a) is taken, add a sibling `../dep` fixture with a newer file and assert a rebuild.

**6. Architectural notes**
- **ARCH-DRY: pass.** One `skip` string feeds both emitted forms. The test copies the guard text once in `SKIP_SDLC` and once in the exact `nous` line. That is an intentional shape pin, not duplicated logic.
- **ARCH-PURE: pass.** This is shell-emitted glue. The decision (exists && (couch || fresh)) is a single expression in the emitted shell function, and it is tested end-to-end through injected stub binaries.
- **ARCH-PURPOSE: pass, with the dependency caveat.** Both purposes are delivered: no writes from slots when the binary exists, and no build when nothing changed. Staleness that comes from local-replace dependencies is part of "not stale" and should either be covered or explicitly scoped out.

**7. Plan revision recommendations**
- Add a `## Revisions` entry stating how local `go.mod` `replace` dependencies are handled: either scanned, or explicitly out of scope with the atlas and header updated to say so.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: doc-absolute-claim-contradicts-code
    title: |
      atlas says a slot "never writes" to :0, but a missing bin/X in a Couch slot builds into :0
    detail: |
      atlas/workflow/base-layer.md:93-94 says "an agent shell in a slot runs whatever :0 last built and never writes there", but the guard requires [ -x bin/X ], so with COUCH_TREE set and bin/X missing it builds into :0 (pinned by the test "in a Couch slot a missing binary still builds"). Family instances in window: base-layer.md:93-94; issue-tracker.md:129 "the alias never rebuilds in a Couch slot" (qualify with "when bin/X exists"). The dev-aliases.sh header is accurate.
  - id: new
    severity: Important
    kind: defect
    family: staleness-scan-scope
    title: |
      freshness scan ignores local go.mod replace deps, so edits to a sibling repo leave a stale binary running
    detail: |
      The find at construct/dev-aliases.sh:285 scans only the owner tree. metis/go.mod has "replace github.com/xianxu/ariadne => ../ariadne" and owns cmd/metis, so an edit to ariadne no longer rebuilds bin/metis; before #405 it was rebuilt every call. Fix: at generation time, add relative replace targets from the owner go.mod as extra find roots, or at least document the exclusion in the header comment and the atlas next to the toolchain caveat.
  - id: new
    severity: Minor
    kind: defect
    family: staleness-scan-scope
    title: |
      edits during a build and find errors both count as fresh
    detail: |
      A source edit after the compiler reads it but before the binary is written is older than bin/X. find errors on unreadable dirs leave stdout empty, which counts as fresh, so the skip fails open.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0774e655-2a7e-4eca-b233-f8c18f42da8c","repository":"github.com/xianxu/ariadne","issue_id":"000405","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/21e4dfe034c45ada","reviewed_head":"d6c05eb0e036548762023a566c4361365ab54ca1","slot":"ariadne:2","requested_at":"2026-10-16T04:00:42.643361Z"},"phase":"finished","pid":97842,"started_at":"2026-10-16T04:00:43.117483Z","finished_at":"2026-10-16T04:02:32.20003Z","session_id":"c50dc749-a50a-4a4d-83b4-b8926fc42dd0","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T21:10:17-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 405 — dev-aliases: skip the rebuild in Couch agent shells (COUCH_TREE) and when the binary is not stale |
| repo | ariadne |
| issue file | workshop/issues/000405-dev-aliases-skip-the-rebuild-in-couch-agent-shells-couch-tree-and-when-the-binary-is-not-stale.md |
| boundary | whole-issue close |
| milestone | — |
| window | f0b5eb265f46a143d783d6bd39e836e71d0f265a..e944fe19aadd870e3db09890ca816bd5e54122d0 |
| command | sdlc close --issue 405 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T21:10:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The three fixes from round 1 are in place, and each behavioural fix has a test that fails without it. I removed the `.start` stamp and the `|| echo stale` fallback in this checkout, and the behavioural cases went red ("a scan error counts as stale", "an edit made during the build makes the next call rebuild"). The replace-target roots have their own fixture and assertions. `dev-aliases.test.sh` passes 63/63 at head. One new correctness gap remains, in the same family as BR-2. The scan only looks at regular files (`-type f`), so deleting a source file never makes `bin/X` stale. I reproduced this: after `rm cmd/b.go`, the files-only scan finds nothing and the scan with directories finds `./cmd`. Without the skip, every call used to rebuild. Now a deletion leaves an outdated binary running, which is the wrong direction for a check meant to fail toward building. `## Done when` also no longer covers the replace-target scope added in the Revisions.

Done-when: stale — it still says "bin/X is newer than every owner file". The round-1 Revision widened the freshness set to local `go.mod` replace targets outside the owner, and made a scan error and a mid-build edit count as stale. None of those appear in Done when, though tests pin all of them.

**Strengths**
- `construct/dev-aliases.sh:272-286`: `fresh_scan` works out the roots when the aliases are generated. It skips replace targets inside the owner and does not follow dependencies' own `go.mod` files, which matches Go's rule that only the main module's replaces apply. Escaping the paths with `%q` handles spaces.
- The skip guard is ordered correctly: `[ -x bin/X ] && { COUCH_TREE || fresh; }`. A missing binary therefore always builds, and dev-prepare is skipped along with the build.
- The tests set file times with `touch -t`, so no case depends on mtime resolution, and a counting stub for go and make records what ran. Every case of the original Done when is exercised both ways: Couch slot with the binary present and missing, fresh and stale, `.git`/`bin` excluded.
- The BR-1 sweep is accurate. base-layer.md, issue-tracker.md and the script header each now state the missing-binary exception.

**Critical findings**
- None.

**Important findings**
1. **Deleting a file never makes the binary stale** (`construct/dev-aliases.sh:285`, the `-type f` in the `printf`). **This is the 3rd finding in family `staleness-scan-scope`.**
   - **Rule:** the freshness check must see every change to what the build reads. That means file contents and also which files exist (adds, deletes, renames), in every root.
   - **Fix the rule:** drop `-type f`. A directory's mtime changes whenever an entry is added, removed or renamed, so directories newer than `bin/X` then count. `.git` and `bin` stay pruned. Editor swap files will sometimes cause a rebuild that wasn't needed, which is the safe direction.
   - **Sweep the docs that say "regular file":** the header at `dev-aliases.sh:20`, the `fresh_scan` comment at `:267`, atlas `base-layer.md:91` ("no file … is newer"), and the Spec.
   - **Add a test:** age everything, delete a file under `cmd/`, and expect `make go`.
2. **`## Done when` doesn't cover what this diff builds** (issue file, `## Done when` bullet 1). Add a clause:
   - files in a local replace target outside the owner count, except under its `.git` and `bin`;
   - a scan error counts as stale;
   - an edit made during the build makes the next call rebuild.

**Minor findings**
- **"The one write a slot makes" leaves out dev-prepare** (`dev-aliases.sh:23-24`, `base-layer.md:97`). **This is the 2nd finding in family `doc-absolute-claim-contradicts-code`.**
  - When a slot builds a missing binary for an owner that declares dev-prepare, it also runs `make -s dev-prepare` in `:0`, which writes generated inputs such as pair's embed tree.
  - **Rule:** describe the exception as the event ("building a missing binary, dev-prepare included"), not as a count of writes.
  - **Instances:** the header line 23, base-layer.md:97, and the Revisions entry ("the one write a slot can make").
- The replace roots are read from `go.mod` when the shell sources the aliases, so a replace added later only takes effect after a re-source. This is the same caveat as dev-prepare detection, and the docs could say so.

**Test coverage notes**
- No case covers a deleted or renamed file, which is the gap behind Important 1.
- The mutation check confirmed that the BR-3 behaviour (scan error, mid-build edit) is pinned by behavioural tests, not only by assertions on the emitted shape.

**Architectural notes**
- **ARCH-DRY: pass.** The guard is built once by `fresh_scan` and spliced into both the plain and dev-prepare forms. The duplicated test stubs are small and local.
- **ARCH-PURE: pass.** `fresh_scan` is generation-time text assembly: it reads `go.mod`, resolves paths and emits a string. The runtime IO stays in the emitted one-liner.
- **ARCH-PURPOSE: flag** (Important 1). The purpose is to skip a build only when nothing it reads has changed. Membership changes are part of what it reads, so the scope is still a subset.

**Plan revision recommendations**
- Add to `## Revisions`: "close round 2: the freshness scan counts directories too, so an added, deleted or renamed file makes bin/X stale. The exception for a slot is stated as building a missing binary, dev-prepare included."
- Update `## Done when` as described in Important 2.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      base-layer.md:93-97 now says a slot never rebuilds an existing binary and that a missing bin/X builds from a slot too; issue-tracker.md:129-130 says the alias rebuilds only a missing binary; the test "in a Couch slot a missing binary still builds" pins it.
  - id: BR-2
    disposition: addressed
    note: |
      fresh_scan (dev-aliases.sh:272-286) adds outside local replace targets as find roots with .git and bin pruned; tests assert the roots and that "a newer file in a replace target makes it stale".
  - id: BR-3
    disposition: addressed
    note: |
      Mutation check: removing touch -r .start and the echo-stale fallback makes "a scan error counts as stale" and "an edit made during the build makes the next call rebuild" fail.
findings:
  - id: new
    severity: Important
    kind: defect
    family: staleness-scan-scope
    title: |
      freshness scan ignores deleted or renamed files because it checks only regular files
    detail: |
      3rd in family. Rule: freshness must observe every change to the build input set, contents and membership, in every root. The -type f at construct/dev-aliases.sh:285 hides deletions: reproduced, rm cmd/b.go gives an empty files-only scan, while dropping -type f reports ./cmd. Before #405 every call rebuilt. Fix: drop -type f so directories newer than bin/X count, keeping .git and bin pruned. Sweep the "regular file" wording at dev-aliases.sh:20 and :267, base-layer.md:91 and the Spec. Add a test that deletes a file and expects a rebuild.
  - id: new
    severity: Important
    kind: defect
    family: done-when-tracks-revisions
    title: |
      Done when does not cover the replace-target, scan-error and mid-build-stamp behaviour added in round 1
    detail: |
      Done when bullet 1 still says "newer than every owner file". The round-1 Revision widened freshness to local replace targets outside the owner (excluding their .git and bin) and made a scan error and a mid-build edit count as stale. Add these clauses to Done when; tests already pin all three.
  - id: new
    severity: Minor
    kind: defect
    family: doc-absolute-claim-contradicts-code
    title: |
      "the one write a slot makes" leaves out dev-prepare, which also runs in :0 when a slot builds a missing binary
    detail: |
      2nd in family. Rule: describe the exception as the event ("building a missing binary, dev-prepare included"), not as a count of writes. Instances: dev-aliases.sh:23-24, base-layer.md:97, and the issue's Revisions entry.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2c60d22d-3d7a-49d7-be10-5d8f088a410b","repository":"github.com/xianxu/ariadne","issue_id":"000405","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/7d47185813774b60","reviewed_head":"e944fe19aadd870e3db09890ca816bd5e54122d0","slot":"ariadne:2","requested_at":"2026-10-16T04:08:28.557529Z"},"phase":"finished","pid":59791,"started_at":"2026-10-16T04:08:28.777789Z","finished_at":"2026-10-16T04:10:17.095755Z","session_id":"3b131d79-d92b-4ec1-bc6e-5f966493f2ca","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T21:13:36-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 405 — dev-aliases: skip the rebuild in Couch agent shells (COUCH_TREE) and when the binary is not stale |
| repo | ariadne |
| issue file | workshop/issues/000405-dev-aliases-skip-the-rebuild-in-couch-agent-shells-couch-tree-and-when-the-binary-is-not-stale.md |
| boundary | whole-issue close |
| milestone | — |
| window | f0b5eb265f46a143d783d6bd39e836e71d0f265a..14ab4a40905e79d5b9346a1ab10131cc8e2d9389 |
| command | sdlc close --issue 405 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-15T21:13:36-07:00 |
| verdict | SHIP |

## Review

Revert-test done (the fix's tests go red without it). One last check: whether a binary that was just built reads as fresh on the next call, with no artificial aging of file times.

```verdict
verdict: SHIP
confidence: high
```

All three open findings are fixed, and the fix behind BR-4 is backed by tests that fail without it. The freshness scan in `construct/dev-aliases.sh:287` now counts directories as well as files (the `-type f` filter is gone), so adding, deleting or renaming a file makes `bin/X` stale. To confirm this, I put `-type f` back in the head checkout and re-ran the suite: 4 tests failed, including "an added file makes it stale" and "a deleted file makes it stale". Without the revert the suite passes 65 of 65. The `## Done when` section now covers round 1's replace targets, the scan-error rule and the build-start stamp. The rule that a slot writes to `:0` only to build a missing binary is now stated as an event in every place the prompt named. One Minor wording gap remains: the Spec bullet at line 33 still says "regular file under the owner". It does not block.

1. **Strengths**
   - The scan has one source, `fresh_scan` (`construct/dev-aliases.sh:267-288`). The emitted test string is pinned exactly in the `SKIP_SDLC` and `nous` tests, so the scan's form cannot drift unnoticed.
   - The behaviour tests use a stub `go` and `make` that count calls, and they set file times explicitly. No case depends on timestamp resolution.
   - Directory times work alongside the build-start stamp. `mkdir -p bin` runs before `.start` is created, and `bin` is excluded from the scan, so a build does not make its own output stale. I checked this without any artificial aging: in a fresh tree, three calls produced exactly one `go build`.
   - The `.git` and `bin` folders of replace targets stay excluded, and a test pins that (`dev-aliases.test.sh:350-351`).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The Spec bullet at `workshop/issues/000405-…md:33` still says "no regular file under the owner (outside `.git` and `bin`)". It omits directories and replace targets. The round-2 Revisions entry does record the change. Every other place that describes the scan scope has been updated: `dev-aliases.sh:19-22`, `:267-269`, `base-layer.md:91-93`, `## Done when` and the test header at `dev-aliases.test.sh:283-285`.

5. **Test coverage**
   - Every clause of `## Done when` is tested, in both states where a clause names two:
     - slot with an existing binary vs. slot with a missing binary
     - fresh vs. newer file
     - file added and file deleted
     - scan error
     - edit made mid-build
     - replace target unchanged vs. newer, plus its `.git`/`bin` exclusion
   - The #355 rename shape and the #299 dev-prepare shape are still pinned.

6. **Architecture**
   - **ARCH-DRY: pass.** A single generator emits every scan, and there is no duplicated logic.
   - **ARCH-PURE: pass.** `fresh_scan` is a pure generator of a string, the actual IO happens in the emitted shell, and the integration tests drive that shell through injected stubs.
   - **ARCH-PURPOSE: pass.** The issue's purpose (skip the build in a slot, skip it when the binary is not stale, and never miss a change) is fully delivered. Every place that describes the scan follows `fresh_scan`, except the Minor Spec line above.

7. **Plan revisions:** none beyond the Minor item. The round-2 Revisions entry already records the change.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      -type f was dropped at dev-aliases.sh:287. With it reverted, the tests "an added file makes it stale" and "a deleted file makes it stale" fail; with the fix, 65/65 pass. The wording was swept at dev-aliases.sh:19-22 and :267-269 and at base-layer.md:91-93. Spec line 33 was left as is (Minor, raised below).
  - id: BR-5
    disposition: addressed
    note: |
      Done when bullet 1 now names the replace targets outside the owner, files and directories, add, delete and rename, scan error = stale, and the build-start stamp. Tests pin each one (dev-aliases.test.sh:321-356).
  - id: BR-6
    disposition: addressed
    note: |
      Worded as an event at dev-aliases.sh:24-25 and base-layer.md:99-100 ("a slot writes to :0 only to build a missing binary, dev-prepare included"). The round-2 Revisions entry says it replaces round 1's count.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: staleness-scan-scope
    title: |
      Spec bullet still states the scan as "no regular file under the owner"
    detail: |
      This is the 4th finding in family staleness-scan-scope. Rule: every prose statement of what the freshness scan observes must match fresh_scan: files and directories, in the owner and in each local replace target outside it, excluding .git and bin. The only remaining instance in the window is the issue Spec at line 33. Every other one is already swept: dev-aliases.sh:19-22 and :267-269, base-layer.md:91-93, Done when bullet 1, and the test header at dev-aliases.test.sh:283-285. The Revisions entry records the change, so this is cosmetic.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"be801d49-c5d3-4528-9945-ce847523910c","repository":"github.com/xianxu/ariadne","issue_id":"000405","recipe":"small-diff-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/b3a7c7f1ab837e00","reviewed_head":"14ab4a40905e79d5b9346a1ab10131cc8e2d9389","slot":"ariadne:2","requested_at":"2026-10-16T04:12:00.247055Z"},"phase":"finished","pid":25017,"started_at":"2026-10-16T04:12:00.529914Z","finished_at":"2026-10-16T04:13:25.128517Z","session_id":"02a57ce7-ade0-4096-b8a8-4e14c079a131","outcome":"returned"}] -->
