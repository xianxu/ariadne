# Boundary Review — ariadne#299 (whole-issue close)

| field | value |
|-------|-------|
| issue | 299 — dev-aliases: a binary whose owner needs a generated input fails to build in a fresh checkout |
| repo | ariadne |
| issue file | workshop/issues/000299-dev-aliases-a-binary-whose-owner-needs-a-generated-input-fails-to-build-in-a-fresh-checkout.md |
| boundary | whole-issue close |
| milestone | — |
| window | 007829a6511132af2a116e045a5f595471b758e6..4511e8d7de0822369904ef96c03ff258023485e4 |
| command | sdlc close --issue 299 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T16:57:53-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The ariadne half of #299 is delivered and does what the Spec asks. If an owner has a `dev-prepare:` rule in its `Makefile` or `Makefile.local`, its emitted functions now run `make -s dev-prepare >&2 &&` before `go build`. Owners without the rule get output identical to before. If preparation fails, the call fails loudly, and the binary's stdout stays clean. I checked the window with the stat, name-status and full-diff recipes. The test file passes 33/33 at head. I also swapped in the base `dev-aliases.sh` in a scratch copy, and 6 of 33 tests went red, which matches the plan's claim. The live `couch --help` check has moved to pair's half under a dated `## Revisions` entry. That is a real split: pair has to add its `dev-prepare` target in its own tree. So deferring it does not under-deliver this issue's purpose. Only Minor findings remain.

1. **Strengths**
   - The emitter checks for the rule once, when the aliases are generated (`construct/dev-aliases.sh:224-226`). Owners without a rule pay nothing per call, and dev-aliases never has to know how any owner builds (ARCH-DRY pass).
   - The tests cover both states in the Done-when: the declaring owner's form, and the plain form compared byte for byte (`dev-aliases.test.sh:67-72`). They also cover a `.PHONY` mention that should not count, and the rule in `Makefile` versus `Makefile.local`. That last pair also exercises the macOS `grep -qs` case where the other file is missing.
   - The stub-driven run (`dev-aliases.test.sh:82-104`) checks the order (prep, then build), that make's output doesn't reach the binary's stdout, and that a failed prep returns non-zero, prints to stderr and never builds. These test behaviour, not wording.
   - `--help` now prints the whole comment header with `awk` instead of a fixed `sed -n '2,30p'` range. The old range would have cut off the newly lengthened header. I ran it: the output ends at the Bash 3.2 note.
   - The atlas dev-binaries section is updated in the same range. No README update is needed, because no flag or subcommand changed.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **Rule detection is narrow** (`construct/dev-aliases.sh:225`). `^dev-prepare:` misses a rule that names several targets (`foo dev-prepare:`) and one with a space before the colon (`dev-prepare :`). It also falsely matches the variable assignment `dev-prepare:=x`. A rule in an included `Makefile.workflow` isn't detected either. The comment says declaring is "one rule", so I'd note the exact accepted shape in the header comment and the atlas section rather than change the code.
   - **Hidden assumption about `make`.** `make -s dev-prepare` relies on the owner's `Makefile` including `Makefile.local`. That holds for ariadne-styled owners (base `Makefile:11`), but the header doesn't say so.
   - **Atlas line too long.** In `atlas/workflow/base-layer.md:97`, the new text was joined onto the existing "The function only **builds + runs** — it" line, making it overlong. Re-wrap it.

5. **Test coverage notes:** Both the declared and undeclared states are covered, as are both Makefile locations, success and failure, and stdout cleanliness. The failure test gets its exit code from `fail_rc=$?` straight after the command substitution. That is a correct pattern.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The prefix is added to the one existing `printf`; there is no second template.
   - **ARCH-PURE: pass.** `owner_prepares` is a thin filesystem check in a shell emitter, and its output is pinned by the tests.
   - **ARCH-PURPOSE: pass.** I looked for other places that build an owner's binary. `make sdlc-build` (`atlas/workflow/sdlc-binary.md:1692`) builds in the owner separately and doesn't run `dev-prepare`. That's fine today because ariadne, sdlc's owner, has no generated inputs. If a downstream ever builds a binary owned by a repo with generated inputs, that path would need the same step.

7. **Plan revision recommendations:** none. The plan's checklist and its Revisions entry match the code.

```findings
findings:
  - id: new
    severity: Minor
    family: owner-declaration-detection-shape
    title: |
      dev-prepare detection regex accepts only `dev-prepare:` at column 0
    detail: |
      construct/dev-aliases.sh:225 misses rules that name several targets (`foo dev-prepare:`) and a space before the colon, falsely matches `dev-prepare:=x`, and ignores Makefile.workflow. Document the exact accepted declaration shape in the header comment and in atlas base-layer.md.
  - id: new
    severity: Minor
    family: doc-line-wrap
    title: |
      atlas base-layer.md:97 joins new prose onto an existing line, making it overlong
    detail: |
      "...keep the plain form. The function only **builds + runs** — it" should be re-wrapped to match the surrounding line width.
```

---

## Re-review — 2026-10-12T17:00:14-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 299 — dev-aliases: a binary whose owner needs a generated input fails to build in a fresh checkout |
| repo | ariadne |
| issue file | workshop/issues/000299-dev-aliases-a-binary-whose-owner-needs-a-generated-input-fails-to-build-in-a-fresh-checkout.md |
| boundary | whole-issue close |
| milestone | — |
| window | 007829a6511132af2a116e045a5f595471b758e6..1b95ae894abcc0a48d98c1d677ab4c010d01bdd6 |
| command | sdlc close --issue 299 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T17:00:14-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both prior findings are fixed, and I checked each against the code. The `dev-prepare` detection regex is wider now. It accepts a rule that names several targets, a space before the colon, and a double-colon rule. It rejects `:=` assignments and `.PHONY` mentions. The accepted shape is documented in the comment above `owner_prepares` (`construct/dev-aliases.sh:222-226`). The atlas paragraph has been re-wrapped. I ran the stat, name-status and full-diff recipes on the pinned window. The test file passes 35/35 at head. To check the regression test, I removed the `([^=]|$)` guard in the disposable checkout and ran the tests again. "A dev-prepare variable assignment is not a declaration" failed. I then restored the file, and the tree is clean. One new Minor finding is left, in the same family as BR-1, and it does not block the ship.

1. **Strengths**
   - The regex `^([^#=:]*[[:space:]])?dev-prepare[[:space:]]*:([^=]|$)` cannot match `.PHONY:` or comments, because the prefix excludes `:`, `#` and `=`. It also handles `dev-prepare:=x` and `dev-prepare := x` correctly. I probed both directly.
   - The new multi-target fixture is `assets dev-prepare : gen.stamp` (`dev-aliases.test.sh:83-84`). It covers two of the BR-1 shapes in one line: several targets, and a space before the colon. The old `^dev-prepare:` would fail it.
   - The `:=` test really pins the guard: removing `([^=]|$)` turns it red.
   - The header documents the accepted shape and the non-declarations right next to the function, and `--help` prints the whole header through `awk`.
   - The atlas paragraph (`atlas/workflow/base-layer.md:92-101`) is re-wrapped. Its lines are 75–79 characters, which matches the rest of the paragraph.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **This is the 2nd finding in family `owner-declaration-detection-shape`.** The rule behind both findings is that the code infers a make declaration by pattern-matching raw Makefile text. Each round of regex patching only narrows the gap. The regex still produces false positives. I tested each of these:
     - a tab-indented recipe line that contains `dev-prepare:`, e.g. `\t@echo run dev-prepare: now`;
     - the POSIX simple assignment `dev-prepare ::= x`;
     - a target-specific variable `dev-prepare: V = 1`, which creates no rule.

     In each case the owner's functions would emit `make -s dev-prepare`. That fails with "No rule to make target", so every call to that owner's binaries would fail.

     The fix that covers the whole class is to ask make itself rather than pattern-match text. For example, `make -C "$1" -pRrq : 2>/dev/null` lists the targets make has parsed (`-p` prints its database; `-q` with the dummy goal `:` builds nothing). Check that list for a `dev-prepare:` target that has a rule. This also covers `Makefile.workflow` and any other include.

     If that costs too much at alias-generation time, a smaller fix is to exclude recipe lines (add `\t` to the prefix class, or skip lines that start with a tab) and record `::=` and target-specific variables as known non-detected cases in the header.

5. **Test coverage notes:** Every shape in the BR-1 list now has a fixture: `.PHONY` only, a rule in `Makefile`, a rule in `Makefile.local`, `:=`, and several targets with a space before the colon. The bare `dev-prepare:=x` case has no test, but the `:=` test already exercises the same guard. Recipe lines, `::=` and target-specific variables have no tests.

6. **Architectural notes**
   - **ARCH-DRY: pass.** There is still one emitter template; the prefix goes into the same `printf`.
   - **ARCH-PURE: pass.** `owner_prepares` is a thin filesystem check, and the tests pin its output. Switching to `make -p` would make it IO-heavier, but it would still sit in the same thin shell.
   - **ARCH-PURPOSE: pass.** This issue delivers ariadne's half: detecting the step and emitting it. Pair's target is separable work, and a Revisions entry already records it.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Regex now accepts multi-target and space-before-colon rules and rejects := (fixtures at dev-aliases.test.sh:79-84). Removing the ([^=]|$) guard turns the := test red. Accepted shape documented in the comment above owner_prepares (construct/dev-aliases.sh:222-226); atlas says "a dev-prepare rule", which is accurate.
  - id: BR-2
    disposition: addressed
    note: |
      atlas/workflow/base-layer.md:92-101 re-wrapped, lines 75-79 chars, consistent with the paragraph.
findings:
  - id: new
    severity: Minor
    family: owner-declaration-detection-shape
    title: |
      Text-regex detection of dev-prepare still false-matches recipe lines, `::=`, target-specific vars
    detail: |
      2nd finding in this family. The rule: inferring make declarations by regex over Makefile text leaves gaps however it is patched. Verified matches at construct/dev-aliases.sh:228 for a tab-indented recipe line "@echo run dev-prepare: now", for "dev-prepare ::= x", and for "dev-prepare: V = 1"; each emits `make -s dev-prepare`, which then fails every call. Class fix: query make's own database (make -C owner -pRrq :) for a dev-prepare target with a rule, which also covers Makefile.workflow and other includes. Cheaper fallback: exclude tab-led lines and document ::= and target-specific vars as unsupported.
```

---

## Re-review — 2026-10-12T17:05:08-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 299 — dev-aliases: a binary whose owner needs a generated input fails to build in a fresh checkout |
| repo | ariadne |
| issue file | workshop/issues/000299-dev-aliases-a-binary-whose-owner-needs-a-generated-input-fails-to-build-in-a-fresh-checkout.md |
| boundary | whole-issue close |
| milestone | — |
| window | 007829a6511132af2a116e045a5f595471b758e6..31c214bd28e4ac07f4167701fda48c6c3e465aef |
| command | sdlc close --issue 299 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T17:05:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The one open finding, BR-3, is fixed at the class level. `owner_prepares` (`construct/dev-aliases.sh:233-240`) no longer reads Makefile text with a regex. It asks make itself (`make -q dev-prepare`) whether the owner's tree defines a `dev-prepare` rule. A cheap grep runs first, so owners whose Makefiles never mention the name skip the make call. I tested the fix by putting the previous commit's `dev-aliases.sh` (1b95ae89) back in the checkout. Three of the new cases failed: the target-specific variable, the recipe line, and the rule in Makefile.workflow. With the head version restored, all 38 tests pass under GNU Make 3.81. The emitted-function contract also holds:
- make's output goes to stderr, so the binary's stdout stays clean;
- a failed `dev-prepare` fails the call and skips the build;
- owners without the rule get exactly the old form.

The Done-when's live check (a fresh pair slot running `couch --help`) is formally moved to pair's half in `## Revisions`. Since that is pair's own tree, this is a legitimate split, not under-delivery. The remaining findings are Minor: stale prose about how detection works, and side effects of `make -q`.

1. **Strengths**
   - `construct/dev-aliases.sh:233-240`: make decides whether the rule exists, so includes, multi-target rules, assignments and target-specific variables all mean what make says they mean. The grep keeps the common no-rule case free.
   - A rule that exists but has a missing prerequisite is still emitted. The call then fails loudly instead of quietly skipping the step, which keeps the "failed prep fails the call" contract.
   - `construct/scripts/test/dev-aliases.test.sh:95-121`: the emitted function is actually run against stub `make` and `go`. The tests check stdout cleanliness, that prep runs before the build, and that a failed prep stops the build.
   - The plain form is pinned byte-for-byte (`nous_line`), so the "today's exact form" clause holds.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **Stale prose about the detection rule (family `doc-restates-detection-rule`).** Text written for the earlier regex no longer matches the code. All instances in this window:
     - (a) The `dev-aliases.sh:15` header says "in its Makefile or Makefile.local" but leaves out Makefile.workflow. The `owner_prepares` comment at line 222 and the atlas both name all three files.
     - (b) The issue's Plan bullet 1 still says `owner_prepares` "greps `^dev-prepare:`".
     - (c) Plan bullet 2 promises a test that "a `.PHONY` mention alone is not a declaration". No such test exists. Under make, `.PHONY: dev-prepare` on its own *does* count as a rule: `make -q` exits 0. That is harmless, since the emitted `make -s dev-prepare` is a no-op, but the claim is false.
     - Fix: point the header at the `owner_prepares` comment as the one description of the rule, and add a Plan revision for (b) and (c).
   - **Side effects of `make -q` while aliases are generated.** `make -q` parses the owner's Makefiles every time a shell sources the aliases. That runs `$(shell …)` calls (I verified a `touch` fires) and can rebuild included makefiles. It is limited to owners whose three files mention `dev-prepare`. Worth one sentence in the atlas note that the step should be cheap.
   - **One make call per binary.** `owner_prepares` runs once per binary, not once per owner, so an owner with N binaries pays for N make calls at shell startup. A per-repo cache inside the loop would fix this.
   - **`::=` is untested.** The BR-3 text names it, but GNU Make 3.81 doesn't support it, so there is no fixture. Acceptable.

5. **Test coverage:** both sides of each Done-when clause are covered: an owner with the rule versus one without, and prep that succeeds versus prep that fails. Detection is tested for:
   - a rule declared in Makefile, Makefile.local, or Makefile.workflow (reached through an include);
   - a multi-target rule;
   - a rule with a broken prerequisite;
   - things that must *not* count as a rule: a `:=` assignment, a target-specific variable, and a recipe line that mentions the name.

   The revert experiment confirms these tests fail without the fix.

6. **Architecture**
   - **ARCH-DRY: pass.** dev-aliases learns nothing about how an owner builds; make's own database is the only authority.
   - **ARCH-PURE: pass.** The IO is one small detection function plus a string choice in the emit loop. Tests run it through real make against fixtures, which suits a shell emitter.
   - **ARCH-PURPOSE: pass.** The ariadne mechanism is complete, and pair's adoption is pair-owned and tracked through Revisions. The finding-class sweep also happened: BR-3 was fixed for the whole class by delegating to make, not by patching one regex case.

7. **Plan revisions:** add a `## Revisions` entry stating that detection now asks make (`make -q dev-prepare`, after the grep), not the `^dev-prepare:` grep. It should drop the `.PHONY` test claim, or record that a bare `.PHONY: dev-prepare` counts as a no-op rule.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      owner_prepares now asks make (make -q dev-prepare) behind a grep prefilter; restoring 1b95ae89's version makes 3 new cases fail (target-specific var, recipe line, Makefile.workflow include); head passes 38/38.
findings:
  - id: new
    severity: Minor
    family: doc-restates-detection-rule
    title: |
      Prose about how dev-prepare is detected still describes the old regex
    detail: |
      Instances: (a) the construct/dev-aliases.sh:15 header says "Makefile or Makefile.local" and omits Makefile.workflow, which line 222 and the atlas name; (b) the issue Plan bullet 1 says owner_prepares "greps ^dev-prepare:"; (c) Plan bullet 2 claims a test that a .PHONY mention alone is not a declaration, but none exists and make -q treats a bare .PHONY: dev-prepare as a rule (rc 0, harmless no-op). Rule: describe detection once, in the owner_prepares comment, and point other prose at it; add a Plan revision.
  - id: new
    severity: Minor
    family: alias-emission-cost
    title: |
      make -q runs owner Makefile parse-time side effects at shell startup, once per binary
    detail: |
      Parsing the owner's Makefiles runs $(shell ...) calls (verified: a touch fired) and can rebuild included makefiles, every time a shell sources the aliases. owner_prepares runs once per binary, not once per owner. Limited to owners whose three files mention dev-prepare. Cache the result per repo in the loop, and note in the atlas that these Makefiles should keep parsing cheap.
```

---

## Re-review — 2026-10-12T17:08:47-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 299 — dev-aliases: a binary whose owner needs a generated input fails to build in a fresh checkout |
| repo | ariadne |
| issue file | workshop/issues/000299-dev-aliases-a-binary-whose-owner-needs-a-generated-input-fails-to-build-in-a-fresh-checkout.md |
| boundary | whole-issue close |
| milestone | — |
| window | 007829a6511132af2a116e045a5f595471b758e6..53766ba99864d585ef99ebd8230aadcab5177c20 |
| command | sdlc close --issue 299 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T17:08:47-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** Both open findings are addressed, and only one Minor finding remains. `owner_prepares` now asks make whether `dev-prepare` is a rule (`make -q`). It runs only when one of the three Makefiles the base layer reads mentions the name, and it runs once per owner. The suite reports 40 passed, 0 failed at the pinned head. I removed the per-owner cache and confirmed the regression test fails without it (`make detections: 2`), then restored the file. Done-when clause 1 is tested both ways: an owner that declares the step gets it in the emitted function, and an owner that declares nothing keeps the exact plain form. The new finding is a doc-comment line that wasn't re-wrapped.

Done-when: current — clause 1 matches what this diff builds and is tested both ways. Clause 2 (the live pair check) was moved to pair's issue by the "live check deferred" Revision. The "detection asks make" Revision changes how the step is detected, not what Done-when requires.

### 1. Strengths
- **Make decides what counts as a declaration.** `construct/dev-aliases.sh:234-241` treats exit 0 or 1 from `make -q` as declared. Exit 2 with "No rule to make target …dev-prepare'" is undeclared. Any other exit 2 is a declared but broken step, so it is still emitted and the call fails loudly instead of silently skipping it.
- **The error-message match is narrow.** The single-character `?` matches both make 3.81's backtick and make 4's quote. It doesn't match a missing prerequisite (`` `missing', needed by `dev-prepare' ``) or a target like `x-dev-prepare`.
- **One detection per owner.** The `prepared` cache (`:243-264`) lives in the same pipeline subshell as the loop, so it works. A test that counts calls to a stub `make` pins it.
- **The emitted function is actually run in the test.** It runs against stub `make` and `go` (`dev-aliases.test.sh:106-132`). That checks the order (prepare, then build), that the binary's stdout stays clean, and that a failed prepare fails the call without building.
- **`make` runs with `MAKEFLAGS`/`MAKELEVEL` unset,** so detection isn't affected when the aliases are emitted from inside a make run.

### 2. Critical
None.

### 3. Important
None.

### 4. Minor
- **`construct/dev-aliases.sh:16`: a header comment line is 115 characters long.** The rest of the header wraps at about 80.
  - **This is the 2nd finding in family `doc-line-wrap`.** The rule: when you edit a prose comment or doc paragraph, re-wrap the whole paragraph to the file's existing width (about 80 columns).
  - Instances of this family in the window:
    - (a) `dev-aliases.sh:16`, 115 characters.
    - (b) `dev-aliases.test.sh:64`, a comment at 83 characters.
  - The longer code and assertion lines in the test file follow that file's existing style and are not part of this family.

### 5. Test coverage notes
- Declared and undeclared owners, the exact plain form, multi-target rules, rules reached through an include of `Makefile.workflow`, a broken prerequisite, `:=`, target-specific variables and a recipe line that mentions the name are all covered. So are caching, ordering, clean stdout and a failed prepare.
- Not covered: a bare `.PHONY: dev-prepare`. The Revision explains this: make treats it as a rule that does nothing, so emitting it is a harmless no-op.

### 6. Architectural notes
- **ARCH-DRY: pass.** Make itself is the authority on what declares the step; there is no restated parser. The cache key and the emitted `prep` string appear twice inside one `case`, which is acceptable.
- **ARCH-PURE: pass.** `owner_prepares` is a thin layer over make and is tested through the generator's real output. That fits a shell emitter.
- **ARCH-PURPOSE: pass.** The ariadne half fully delivers the mechanism. Pair's half is a separable owner adoption that the TL has filed in pair, not the point of this issue being deferred.

### 7. Plan revisions
None needed. The "detection asks make" Revision already corrects Plan bullets 1-2.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      Header now points at owner_prepares (dev-aliases.sh:15-16); the Revision "detection asks make" corrects Plan bullets 1-2 and drops the .PHONY test with its reason.
  - id: BR-5
    disposition: addressed
    note: |
      Per-repo cache in the emit loop (dev-aliases.sh:243-264) with a counting-stub test; removing the cache turns it red (make detections 2). Parse cost is noted in the header.
findings:
  - id: new
    severity: Minor
    family: doc-line-wrap
    title: |
      dev-aliases.sh:16 header comment line is 115 characters long; the header wraps at about 80
    detail: |
      2nd finding in family doc-line-wrap. Rule: after editing a prose comment or doc paragraph, re-wrap the whole paragraph to the file's existing width (about 80 columns). Instances in this window: dev-aliases.sh:16 (115 characters); dev-aliases.test.sh:64 (a comment at 83 characters). Long code and assertion lines in the test follow that file's existing style and are excluded.
```
