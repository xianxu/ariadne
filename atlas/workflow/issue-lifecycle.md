# Issue Lifecycle

## Flow

```
Issue created (sdlc issue new "<title>", or sdlc issue new --from-github 42) → workshop/issues/NNNNNN-slug.md → sdlc claim → sdlc start-plan → design (complex → durable plan via superpowers-writing-plans → workshop/plans/NNNNNN-slug-plan.md) → sdlc change-code (in-place branch by default) → work → sdlc close (local acceptance review → codecomplete) → sdlc pr → sdlc merge (deterministic publish → done)   [direct sdlc push on main still available, but not the default]
```

## States

| Status | Meaning |
|--------|---------|
| open | Not started |
| working | An agent is working on it |
| blocked | Waiting on something |
| codecomplete | Code complete; passed the local acceptance review (`sdlc close`), awaiting merge (#160) |
| done | Merged/published, awaiting archive |
| wontfix | Declined |
| punt | Deferred |

**The two-gate publish model (#160).** `sdlc close` is the **local acceptance
gate** — it runs the fresh-context boundary review (all LLM review lives here,
incl. docs/atlas + README sync) and flips `working → codecomplete`, NOT straight
to `done`. `sdlc merge`/`push` are the **deterministic publish gate**: they verify
`HEAD` is unchanged since close (the *reviewed-HEAD-unchanged* invariant — nothing
drifted after the review; doc-only bookkeeping deltas are tolerated, #174), run no
LLM judge, flip `codecomplete → done`, and archive. `codecomplete` is written **only** by `sdlc close` (set-status refuses it),
which is what makes the commit carrying it a trustworthy anchor for that invariant.
So `done` now means "reviewed AND published," not "an agent thinks it's finished."

## Transitions

1. **Create**: `sdlc issue new "<title>"` allocates the next ID and writes the canonical template (the no-GitHub entry path); `sdlc issue new --from-github <num>` (or the older `sdlc fetch`) seeds it from a GitHub issue. See `sdlc issue --help` for the canonical issue-file contract.
2. **Claim**: `sdlc claim --issue N` flips an open issue to `working` and publishes the issue-state claim to main in one step (`--no-start` to skip the flip). A **cheap lock** — no estimate demanded (#113), so claim early (at brainstorm start). The flip stamps an explicit `started:` timestamp (#116) that anchors the active-time window at engagement start, so `sdlc actual` measures design attention instead of dropping it (superseding the older `WorkingTransitionISO` git-log heuristic; gap-truncation keeps a dormant claim→work gap from inflating the actual).
3. **Plan**: `sdlc start-plan` marks the design entry — it delivers the `at-plan` architecture lens, points at the durable-plan path, and tells you NOT to derive `estimate_hours` yet — `change-code` asks for it only after the plan clears plan-quality (#187). For complex work, author the plan via the **`superpowers-writing-plans`** skill into `workshop/plans/NNNNNN-slug-plan.md` (version-controlled — never the harness builtin's ephemeral `~/.claude/plans/`, #72).
4. **Work**: Agent works within the issue file — updates Plan, Log, Spec sections
5. **Default — branch + PR**: `sdlc change-code` creates an **in-place branch** (a branch in the current checkout) after the gates; `sdlc pr` opens the pull request; `sdlc merge` merges it server-side, archives done issues, and switches back to main. `--worktree=yes` gets an isolated worktree instead (parallel work).
6. **Shortcut — direct on main**: `sdlc push` (auto-commit, pre-merge checks, push, archive, close GH issues) still exists for quick one-liners, but is no longer the default (#51).

## Worktree layout

Worktrees are created at `../worktree/<repo-dir-name>/<branch-name>/`, keeping
worktrees from different repos separated. The `<repo-dir-name>` is the basename
of the current working directory (i.e., the repo folder name).

```
../worktree/
└── my-repo/
    ├── 000042-add-feature/    ← branch: 000042-add-feature
    └── 000051-fix-bug/        ← branch: 000051-fix-bug
```

**Branching decision** (#51): `sdlc change-code --issue N` runs structural checks,
the `estimate_hours` gate (#113 — relocated here from `claim`; `--no-estimate`
bypasses), the **estimate-reconciliation** gate + **estimate-quality** judge
(#117 — estimate_hours must reconcile with an itemized `## Estimate` block;
`--no-estimate-recon` / `--no-judge` bypass), and the plan-quality judge, then
branches. The default (no `--worktree` flag) is
**in-place** — a branch in the current checkout, no worktree dir; the common
case, chosen without prompting. `--worktree=yes` gets an isolated worktree (the
layout above); `--worktree=ask` restores the interactive prompt, or for a
non-interactive agent emits the `ASK_BRANCHING_STRATEGY` sentinel (exit 2) so the
agent can ask the operator and rerun with `--worktree=yes|no`.

**Terminal detection (`isTTY`, shared).** The "is this interactive?" decision
behind both `change-code --worktree=ask` and `sdlc merge`'s final-confirm
fail-fast is the one `isTTY` helper. It must be a **real `isatty`** (the
`TIOCGETA`/`TCGETS` ioctl, stdlib-only in `tty_*.go`), NOT an `os.ModeCharDevice`
check: `/dev/null` — an agent's usual redirected stdin — is a char device but not
a terminal, so the char-device proxy misclassified agents as interactive and
merge ran all its judges before aborting at the unanswerable prompt (#141). One
helper, one fix, both callers.

**Navigation**: worktree creation writes the path to `.goto`; the shell `g`
alias reads it to `cd` you there. `sdlc merge` writes the main worktree path
back into `.goto` for the return trip.

## Issue file structure

`sdlc issue new` writes this shape (see `sdlc issue --help` for the
authoritative field/section contract; the on-disk template is rendered by
`Render` in `cmd/sdlc/internal/issue/scaffold.go`):

```markdown
---
id: 000042
status: open
deps: []
github_issue: 42
target:            # optional; a workshop/targets/ slug
created: 2026-04-20
updated: 2026-04-20
estimate_hours:    # derived after the plan clears plan-quality; required by change-code, not claim (#113, #187)
                   # actual_hours: added at close; number or N/A when status=done
---

# Title

## Problem
What's wrong / what's needed (seeded from the GitHub body on --from-github).

## Spec
- brainstorming results (if needed)

## Done when
- acceptance criteria

## Plan
- [ ] checklist of work

## Log
### 2026-04-20 — session summary
One paragraph: what was attempted, what landed, what got deferred.

### 2026-04-20
- individual decisions, discoveries

## Side quests
- (optional; recommended for multi-day issues) name + ~time + commit ref
```

## Reading a section (`internal/issue/fence.go`, #211)

Issue bodies are parsed by ONE fence-aware scanner. `FenceSpans` classifies each
line as inside or outside a fenced code block; `ScanMarkdownLines` visits the
prose lines; `SectionBody` finds `## <heading>` and runs to the next heading that
is **not** inside a fence. `PlanSectionBody` is the Plan's named shortcut.

**Why it can't be a regex.** Go's `regexp` is RE2 — no backreferences — so
CommonMark's "a closing fence is the same character and at least as long as the
opener" is inexpressible, and a four-backtick block containing a three-backtick
line is parsed wrong. Measured: a fence-consuming regex handles 4 of 5 forms.

**Why it matters.** This repo's deliverables are frequently markdown documents —
registry entries, datatype templates, helptext, skills — so an issue specifying
one quotes it, and the quoted headings are `##` because the target file's are.
The prior `^## ` terminator ended the section at the quoted heading. For `## Plan`
that was a false PASS, not a false refusal: the close gates count things whose
*absence* means pass, so plan-unchecked saw 0 open items, the milestone scan
missed milestones, and `CountPlanItems` (behind `sdlc state`) under-reported.

**The unterminated-fence policy is a parameter, and the choice is load-bearing:**

| consumer | policy | why |
| --- | --- | --- |
| `SectionBody`, plan extraction | `UnterminatedIsProse` | a stray opener must never hide `## Plan` from the gates |
| `stripCodeFences` (word count) | `UnterminatedIsProse` | pre-existing, deliberate |
| `StripFenced` (plan counters) | `UnterminatedIsProse` | a quoted `- [ ]` is not open work |
| `project` section scan | `UnterminatedIsFenced` | pre-existing behavior, unchanged |
| `SplitFences` (#179 `migrate`) | `UnterminatedIsFenced` | **own scanner** — see below |

Inheriting the fenced policy for `SectionBody` would be worse than the bug:
instead of one truncated section, every heading after the stray fence vanishes.
The price of the prose policy is over-segmentation — a `##` after an unterminated
opener reads as a real heading — which is visible and recoverable, and pinned by
its own test so it doesn't get "fixed".

**One exception, deliberate.** `SplitFences` keeps its own scanner because it is
CHARACTER-oriented, not line-oriented: its contract covers inline pairs mid-line
(`` a```one``` mid ```two```z `` is two fenced segments with prose between) and
byte-exact boundaries that fall inside a line, which line classification cannot
express. It also answers a different question — "may a rewriter edit these bytes"
rather than "where does this section end" — so merging them would change what
`migrate` rewrites across repos to serve a tidiness this class doesn't need.

**Bounding a search to a fence-aware section does not make the SEARCH
fence-aware.** A section can quote its own format — the real `## Log` of an issue
about log formatting contains a `### <date>` example — so a matcher scanning that
section's raw text still lands inside the quote even though the section boundary
was right. Two helpers close that level:

- `StripFenced(s)` for readers (`logHasEntryToday`, every plan-item counter via
  `PlanItemsBody`);
- `FindLineOutsideFences(s, re)` for callers that need a byte OFFSET to splice
  (`insertLogLine`'s day-header lookup).

The write side needs it too: the milestone tick used to `ReplaceAll` over the
whole issue body, so a `- [ ] Mx` in any quoted example was ticked. It is now
scoped to the Plan section and fence-filtered.

**Everything that finds a heading is fence-aware**, which took three sweeps to
finish: `SectionBody`, `PlanSectionBody`/`PlanItemsBody`, `logHasEntryToday`,
`insertLogLine` (heading, section end, and the day-header search within it),
`planGateContent`'s Estimate strip, the plan-item counters, and the milestone tick. `insertLogLine`'s previous
*last-match* heuristic (#66) was a weaker workaround for this same defect — it
was added because first-match filed a close line into #66's own quoted example,
and it fails when a quoted `## Log` sits AFTER the real one. `logHasEntryToday`
was a live instance: it read the fenced `## Log` at line 22 of
`workshop/history/issues/000066-*.md` instead of the real one at line 68.

Guarded by a corpus property test over every `workshop/**/*.md` (2875 sections
across 406 files at the time of writing): no heading outside a fence may become
unreachable through `SectionBody`.

For frontmatter and section conventions see the **xx-issues skill**
(`construct/local/issues/SKILL.md`). For the cross-artifact closing
sweep (actual_hours, project-file update, atlas update, validation
log) see **AGENTS.md §5 closing checklist**.

## Closing

Each `sdlc push` / `sdlc merge` archives done issues into `history/`. Before that, run the **closing checklist** from AGENTS.md §5:

1. Verify behavior.
2. Tick the completed `## Plan` items; `sdlc close` flips `status` to
   `codecomplete` (#160 — the local acceptance review; `merge`/`push` later flip
   it to `done`). Atomic single-pass work uses plain `- [ ]` checkboxes and closes
   in one `sdlc close`; only tag `Mx` rows when the work has ≥2 separate review
   boundaries you'll `milestone-close` individually (AGENTS.md §3 — an `Mx` tag is
   a review boundary, not a task label).
3. **Record `actual_hours`** in the frontmatter: a measured positive number, or
   explicit `N/A` only when measurement is not applicable.
4. Update the parent project file (if any).
5. Update `atlas/` for any new architectural surface.
6. ~~Append validation-log entry~~ — now automatic: on a full-issue close `sdlc
   close` appends numeric estimate↔actual pairs to the calibration ledger
   (#117), superseding the manual validation-log step. `actual_hours: N/A`
   closes are schema-valid but excluded from calibration.
