# Boundary Review — ariadne#250 (whole-issue close)

| field | value |
|-------|-------|
| issue | 250 — Seeded merge-check builds weave from source while the tap is unpublished |
| repo | ariadne |
| issue file | workshop/issues/000250-seeded-merge-check-builds-weave-from-source-while-the-tap-is-unpublished.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f16e3f98d208f8f62c0047f14a37ac0697c2622..3a6c65d91ef700f2cf157a9aa692e68c04d32461 |
| command | sdlc close --issue 250 |
| reviewer | claude |
| timestamp | 2026-09-24T20:18:22-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The change does what the Spec asks for. A consumer that isn't ariadne source now runs `brew tap xianxu/ariadne` first. If the tap fails, it emits a `::warning::` naming #241, shallow-clones ariadne, runs `brew bundle` on the clone's Brewfile, builds `weave` into the same `$RUNNER_TEMP/weave-candidate` directory, and adds that directory to `$GITHUB_PATH`. `bootstrap.sh` is unchanged. It still execs the `weave` on PATH when `weave dependencies --help` succeeds, so `brew install` is skipped only when the fallback provided the binary.

`portable-ci.test.sh` runs the workflow's real run blocks for three gateways: tap published, tap unpublished, and ariadne source. It also keeps the existing failure cases, and I re-ran it and it passes. The removal note is in `workshop/issues/000241-publish-weave-startup.md:101`. The last Done-when clause (a pair PR's merge-check passes after re-seeding) can only be checked after merge, so it is outside this window. Nothing blocks SHIP. There are two Minor findings.

**Strengths**
1. The test runs the workflow YAML's real run blocks, not a copy of them. So a mutation such as deleting the `elif` branch or its `$GITHUB_PATH` line turns the test red; the Log says both mutations were checked.
2. The unpublished-tap case checks the full event order (`brew:tap`, `clone`, `brew:bundle …/ariadne-source/Brewfile`, `candidate-build`, `compile`, `setup`, `checked`). Because the list is complete, it also proves `brew install` never ran.
3. The fake `git` only intercepts `clone` and passes every other call through to the real git. The later merge-check blocks still get real git.
4. The fallback reuses the source branch's candidate path, and `bootstrap.sh` is untouched. `weave` stays the only gateway and no second install path was added.
5. `elif ! brew tap` runs in a conditional context, so a failed tap under `set -e` goes to the fallback and does not abort the step.

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
- **ARCH-DRY:** `.github/workflows/merge-check.yml:25-27` and `:35-37` repeat the same three lines: `mkdir -p` of the candidate directory, `go build … ./cmd/weave`, and `echo … >> "$GITHUB_PATH"`. The first branch could set `src=.`, the fallback could set `src=<clone>`, and one shared build tail could follow. Since #241 deletes this branch soon, this is optional. These two branches are the only instances in the window.
- **Stale step name:** `merge-check.yml:17` is still called "Build candidate gateway for ariadne source CI", but the step now also covers consumers while the tap is unpublished. Something like "Build candidate gateway (ariadne source / unpublished tap)" would fit. The in-step comment was updated; the step name is the only stale label.

**Test coverage notes**
- All three Done-when gateways are covered, in both tap states.
- The fake `git` uses `${@: -1}`, which works only because the helper writes a `#!/bin/bash` shebang.
- The published case now expects an extra `brew:tap` row. That matches the new real behaviour.
- When a real `brew tap` fails for a transient network reason, the step falls back to a source build instead of failing. The `::warning::` makes this visible, so it's acceptable.

**Architectural notes**
- **ARCH-DRY:** passes, apart from the Minor finding above.
- **ARCH-PURE:** passes. This is CI glue, and it is tested through injected fake commands.
- **ARCH-PURPOSE:** passes. The fix goes in the seeded workflow that every consumer gets, not in a pair-only patch. Deleting the branch is recorded in #241, which is the right owner.

**Plan revision recommendations:** none.

```findings
findings:
  - id: new
    severity: Minor
    family: duplicated-candidate-build-tail
    title: |
      Candidate mkdir/go build/GITHUB_PATH block duplicated across both gateway branches
    detail: |
      merge-check.yml lines 25-27 and 35-37 are the same three steps apart from the build directory; set src (. or the clone) and share one tail. These two branches are the only instances in the window; optional because #241 deletes the fallback.
  - id: new
    severity: Minor
    family: stale-label-after-scope-change
    title: |
      Step name "Build candidate gateway for ariadne source CI" no longer describes the consumer fallback
    detail: |
      merge-check.yml line 17. The in-step comment was updated; the step name is the only stale label in the window.
```
