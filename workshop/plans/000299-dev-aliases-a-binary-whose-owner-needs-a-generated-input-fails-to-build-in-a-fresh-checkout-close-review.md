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
