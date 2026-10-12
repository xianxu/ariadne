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
