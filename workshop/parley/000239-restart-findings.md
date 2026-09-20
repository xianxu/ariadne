---
type: parley
topic: "#239 — what we actually learned, and whether the ticket should be restarted"
created: 2026-09-20
related: [ariadne#239, ariadne#240, ariadne#210, ariadne#222, ariadne#225]
---

# #239 restart findings

Written to survive a context clear. Operator's read: *"you are treating various
historical artifacts as gospel, they are just oversights, likely."* That is
correct, and it is the through-line of everything below.

## The one-paragraph state

M1–M3 are implemented and closed (M3's close is blocked on a gate artifact, not
on code). The deliverable is ~150 lines: the `seed-once` verb, `mergeManagedBlock`,
`IgnoreEntries`. M4 (the fleet untrack) is not started and is where the real risk
lives. Branch `000239-minimal-committed-base-layer-surface`, pushed.

Roughly 60% of the elapsed work was **meta-work** — tooling to enforce comment
hygiene, then tooling to test that tooling — not the deliverable. See
*What went wrong in the process*.

## Operator questions, answered with measurements

### 1. scaffold / touch targets — "that's just a very minor thing"

**Agreed; I oversold it.** The rule really is *"gitignore everything weave
creates except the bootstrap core."* The scaffold/touch wrinkle is an
implementation detail, not a design insight: git does not track empty
directories, so `scaffold workshop/issues` needs no ignore line at all — you
simply must not emit `/workshop/issues/`, which would ignore the *contents*. One
`case` in a switch. I headlined it as the load-bearing reframe; it is not.

### 2. `.colima/` — "what file needs to be tracked?"

Six files, all ariadne's own, none of which a derivative should change:

```
.colima/Makefile  .colima/colima.sh  .colima/vm-rc.sh
.colima/vm-setup.sh  .colima/vnc-setup.sh  .colima/test/colima.test.sh
```

They are already tracked. The blanket `/.colima/` line only bit if someone
**added a new file** there — e.g. a second test beside `colima.test.sh` — which
would be silently invisible to `git add`. So it is **latent, not live**. I
called it "a live bug"; that was overstatement. Removing it costs nothing and is
a fine side-effect, but it is not a reason this ticket exists.

### 3. kbench's 1160 files — "tell me the issue you face"

This one is real, and it is **not a #239 problem — it is a pre-existing
`sdlc propagate-base` bug** that M4 would be the first thing to trigger
fleet-wide.

- `commitConsumption` (`cmd/sdlc/propagatebase.go:248`) runs
  `git ls-files -i -c --exclude-standard` — *tracked-but-ignored* files — and
  `git rm --cached` on every one. Its intent is to untrack files that weave just
  made ignored.
- But that command reads the **entire ignore configuration**, including nested,
  repo-owned `.gitignore` files that have nothing to do with weave.
- kbench returns **1160** such paths, all under `competition/arc-agi-3/runs/`,
  matched by its own `competition/arc-agi-3/.gitignore:24` → `runs/20*/`.
- That pattern is deliberate for a competition host: *ignore future run
  directories, keep the ones already committed*. "Tracked AND ignored" is the
  intended steady state, not drift.
- So one `propagate-base` run would `git rm --cached` and commit away 1160
  deliberately-committed research artifacts.

**Fix belongs in `propagate-base`, not in #239:** untrack only paths whose
ignore-match comes from weave's own managed block, by pattern provenance
(`git check-ignore -v` reports the source file and line). Verified on ariadne:
the single tracked-but-ignored path there is `construct/staging/.gitignore`,
matched at `.gitignore:7` — *outside* the block — exactly the case the filter
must leave alone.

### 4. CI and weave — "all derivative CI needs to run weave in some form"

**Your intuition is right and the current behaviour is the oversight.** Measured:
`merge-check.yml:38` runs `BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh`, and
`bootstrap.sh:34-36` exits **before** the `make bootstrap` handoff. So CI clones
the peer chain and then never weaves.

Everything I built around that is scaffolding on an accident:

- the whole "pre-weave consumer" class I enumerated,
- keeping `scripts/merge-checks.d/*` in the tracked set,
- the `run-merge-checks.sh` owner-fallback,
- the astro/parli "vacuously green CI" finding.

**If CI ran weave, that entire class dissolves** and the committed surface can
shrink to the true bootstrap core. That is one line in `merge-check.yml` versus
four workarounds. I should have named CLONE_ONLY as the suspect instead of
designing around it.

*(Open question for the redesign: CLONE_ONLY presumably exists to keep CI fast —
`make bootstrap` also builds tools and installs. A `make weave`-only step is
probably what is wanted, not the full handoff.)*

### 5. What is `--target`?

`weave compile --target claude|codex|gemini` compiles **one harness's face**
instead of the Union. A "face" is that harness's entry file + skill directory
(`plan.Target`, `cmd/weave/internal/plan/target.go`):

| target | entry file | skill dir |
|---|---|---|
| claude | `CLAUDE.md` | `.claude/skills/` |
| codex  | `AGENTS.md` | `.agents/skills/` |
| gemini | `GEMINI.md` | `.agents/skills/` |

Default is the Union (all three). It came from #107 Option B.

**Measured: `--target` has no non-test caller anywhere in the fleet** — it
appears only in `main.go`'s own doc comment and in tests. So the hazard I
designed the `TargetAll` pin around (a lean compile shrinking the wholesale-
replaced ignore block) **is theoretical**. The pin is cheap and correct, but it
is not evidence of difficulty, and I cited it as such.

## What went wrong in the process

Worth keeping, because it is the reusable part.

1. **I treated oversights as constraints.** CLONE_ONLY, the seed semantics, the
   `--target` lean mode — I designed *around* each instead of asking whether it
   was intentional. The operator's read is correct.
2. **I over-applied "fix the CLASS, not the instance."** A comment went stale, so
   I wrote a merge check (`46-removed-symbol-references.sh`). It had a bug that
   made it silently pass, so I wrote a fixture harness. The harness counted any
   non-zero exit as success, so I added a self-test. Then a fix to the check
   inverted a false-positive into a **false negative masking 121 symbols**.
   Three levels of tooling to police stale comments, which produced two real
   defects of its own and caught nothing in the deliverable.
3. **The review loop fed on itself.** Rounds 9–12 are mostly findings about
   tooling added in rounds 6–8, not about the committed-surface invariant.
4. **I oversold the difficulty** in my own summaries (points 1, 2 and 5 above),
   which made the work look more necessary than it was.

## Recommendation

Restart the ticket from goals rather than continuing. Suggested framing to
settle **before** any more code:

- **Goal.** A derivative commits only what is needed to get to a first
  `make weave` (plus its own source). Everything else weave produces is ignored.
- **Decide first: does CI weave?** If yes (recommended), the committed core is
  genuinely tiny — `bootstrap.sh`, `merge-check.yml`, `construct/deps` — and
  several M1–M3 complications disappear.
- **Split out what is not this ticket.** The `propagate-base` provenance bug
  (kbench) is its own issue; so is registering `go test` in a gate; so is the
  `--target`-has-no-caller question (retire it?).
- **Keep from M1–M3:** `seed-once` (a real ownership bug — `seed` was destroying
  repo-owned Makefiles), `mergeManagedBlock` (retirement of an entry is only
  possible with a managed region), `IgnoreEntries` (the derivation).
- **Consider dropping:** `46-removed-symbol-references.sh` and its fixtures —
  net-negative so far. `45-verb-enumeration.sh` is simple and clean; keep.

## Pointers

- Issue: `workshop/issues/000239-minimal-committed-base-layer-surface.md`
  (`## Revisions` has the full round-by-round trail)
- Plan: `workshop/plans/000239-minimal-committed-base-layer-surface-plan.md`
- Reviews: `…-m1-review.md`, `…-m2-review.md`, `…-m3-review.md`
- Lessons added this session: `workshop/lessons.md` (falsification; staging by path)
- Related: **#240** (issue-update threading), **#210** (`go test ./...` red since
  a plan was archived), **#222** (concurrent trunk edits)
