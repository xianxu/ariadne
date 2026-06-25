# Plan: Instance-conformance validator (#124)

Validate typed-markdown artifact **files** against their datatype's cue schema:
`file.md → [locator] → normalized instance → cue vet + required-section check → diagnostics`.
The schema defends the skeleton (well-formedness only); the LLM owns the soul (semantic
quality). No deterministic write-back — diagnostics out, LLM/human fixes, re-validate.

Depends on #122 (done): the vocabulary layer (`construct/vocabulary/*.cue`), the
`cmd/vocabulary` cue runner + DAG-merge, `pkg/layergraph.MergeByName`.

## Resolved design decisions

- **Locator = generic + imperative** (ARCH-DRY, Simplicity-First). Two generic locators:
  frontmatter (YAML split) and sections (`## ` heading split). NO declared-per-datatype
  locator DSL — that's speculative until a 3rd datatype needs something the generic ones
  can't express (this answers the Spec's open question). Bespoke locators (fenced blocks,
  `- [ ]` lines) stay out of scope until a schema actually needs to constrain them.
- **Engine = one binary** (operator decision; matches Spec "one binary, many triggers"):
  `vocabulary validate-instance --type <noun> <file.md>`, reusing `resolveVocab()` +
  `CueRunner` in place. sdlc's pre-merge gate + `sdlc issue validate` **shell out** to it
  (sdlc already shells `cue` + `claude`; the vocabulary binary is a built substrate dep).
  Gate tests use a process-level fake (per the "model external services" discipline).
- **Sections: reuse `structural.go`'s policy, enforced added-files-only** (corrected after the
  plan-quality judge MEASURED the corpus — a flat required-trio was provably wrong: `000052`
  (status `working`, in-flight) has no `## Done when`; ~23 history files lack it; and a flat list
  would diverge from `structural.go`'s `checkDoneWhen` Done-when-**or**-`related:` fallback).
  Corrected design:
  - The required-section policy is **single-sourced from `structural.go`**: lift its
    presence checks (Spec present, Plan has items, Done-when-bullet-or-`related:`) into a shared
    `pkg/` function both `structural.go` (which adds its ≥50-word semantic check on top) and the
    validator call. NO `requiredSections` in the `.cue` (the fallback is logic, not a flat list;
    and section requirements are lifecycle-dependent, not a clean cue invariant).
  - **Frontmatter** validation is hard + universal (every changed issue, old or new) — this is
    the clean invariant and it catches the motivating bad-`status` case on existing tickets too.
  - **Section** validation is enforced at the gate only on **newly-ADDED** issue files (git `A`
    status) — new issues must be well-formed; pre-existing/legacy tickets are grandfathered, never
    retroactively failed. (Operator: "validate forward, don't fail old tickets.")
- **Loud escape hatch** ([[feedback_escape_hatch_loud_claim]]): the gate ships `--no-validate` /
  `--force <reason>`, and using it must claim LOUDLY (a prominent WARN + the reason in the
  `## Log` / commit) — never a silent bypass. Structure + fast iteration.
- **Open `#Issue`** (decided after the plan-quality judge measured the corpus + named the
  closed-schema cost). cue definitions are closed by default; add `...` to OPEN it. Rationale:
  a *closed* schema is a field allowlist that must track organically-growing frontmatter
  (`target:`/`references:`/`related:` already exist across real files; `000122` — this issue's
  dep — has `target:`), and a false-positive rejection at a **fail-closed merge gate** trains
  operators to reach for `--no-validate`, defeating the gate ([[feedback_escape_hatch_loud_claim]]
  — minimize forced escapes). An OPEN schema still catches the **high-value** cases: a bad
  `status` *value* (the `#Status` enum) and a typo'd *required* field (`statuss:` → `status`
  absent → required-field failure). It only misses typos of *optional* fields (low-stakes). So
  the schema change is minimal: keep the existing constrained fields (`id`, `status`,
  `estimate_hours?`, `actual_hours?` + the done-guard) and add `...`. No need to enumerate every
  field. (The existing #122 schema tests + `vocabulary vet` must still pass.)

## The load-bearing mechanism — `cue vet` of an instance

Vet extracted frontmatter against the `#Issue` *definition* (not a top-level field):

    cue vet -d '#Issue' <frontmatter.yaml> construct/vocabulary/issue.cue

`-d` unifies the data file against the named definition. **Spike DONE (cue v0.16.1):**
`cue vet -d '#Issue' valid.yaml issue.cue` → exit 0; `status: in-progress` → exit 1 with a
conflict diagnostic; extra fields (`deps`/`created`) → "field not allowed" (confirms `#Issue`
is closed, so completion is required); and cue's YAML loader keeps unquoted `created:
2026-06-24` / `started: 2026-06-25T12:14:37-07:00` as JSON strings, so `: string` field types
vet clean (the date/timestamp trap does not materialize). The package's concrete
`categories`/`lifecycle`/`laws` do not interfere with `-d`. Extend `CueRunner` with
`VetInstance(dataPath, schemaPath, def string) error`. The data temp file MUST have a
`.yaml` extension or cue won't parse it as YAML.

## Milestones

**3 review boundaries (M1–M3), matching the issue `## Plan` Mx rows exactly** — `sdlc
milestone-close` keys off those rows. M1 = the generic frontmatter engine; M2 = the gate +
the issue-specific section overlay + on-demand surface; M3 = generalize.

### M1 — Validator engine: generic frontmatter conformance
- Spike DONE (see "load-bearing mechanism" below): `cue vet -d '#Issue'` works; closed `#Issue`
  rejects extra fields (so completion is required); cue's YAML loader keeps unquoted
  date/timestamp scalars as strings, so `: string` field types are safe.
- **Open `#Issue`** (`construct/vocabulary/issue.cue`): add `...` so organically-growing
  frontmatter (`target`/`references`/`related`/future fields) is allowed; keep the constrained
  fields (`id`, `status`, `estimate_hours?`, `actual_hours?` + the done-guard). NO
  `requiredSections` (sections are M2's overlay). Minimal change — no need to enumerate fields.
  `vocabulary vet` + the existing #122 schema tests still pass (cross-artifact guard).
- `CueRunner.VetInstance(dataPath, schemaPath, def)` + `osCue` impl (writes data to a
  **`.yaml`** temp so cue parses it as YAML); fake in tests.
- **Pure diagnostic transform** (ARCH-PURE — the load-bearing fragile piece): a pure function
  `cueStderr → []Diagnostic` that collapses cue's verbose disjunction errors (the spike showed
  6 lines for one bad enum) into one clear per-field message
  (`status "in-progress" is not valid (want: open|working|blocked|done|wontfix|punt)`).
  Unit-tested directly against captured cue-stderr fixtures — the Done-when ("actionable enough
  an LLM fixes the file") rests on this, so test it in isolation, not only end-to-end.
- `vocabulary validate-instance --type <noun> <file>`: resolveVocab → schema path; split
  frontmatter → `.yaml` temp → VetInstance with `#<TitleCase(noun)>`. Emit diagnostics; exit
  non-zero on failure. Generic + frontmatter-only — works for any noun with a `.cue`.
- **Tests (the contract — now TRUE against the corpus):** the full corpus of real
  `workshop/issues/*.md` + `history/*.md` PASSES **frontmatter** validation; `status:
  in-progress` REJECTED with a clear diagnostic; unknown field (`statuss:`) fails; `done` issue
  missing `actual_hours` fails (the compiled guard); a valid file passes.
- **Done:** the issue's bad-status scenario is rejected; the whole real corpus passes frontmatter.

### M2 — Wire the fail-closed gate + section overlay + on-demand surface
- **Section policy, single-sourced from `structural.go`** (ARCH-DRY): export
  `issue.CheckSectionsPresence(text)` from `cmd/sdlc/internal/issue` (Spec present; Plan has ≥1
  item; Done-when-bullet-**or**-`related:` fallback). NO new `pkg/` — both section callers
  (`sdlc issue validate` + `validateChangedIssues`) live in `cmd/sdlc` and can import the
  internal package (the judge's correction; only the frontmatter *split* needs the
  `pkg/frontmatter` lift, since `cmd/vocabulary` can't reach internal). `structural.go` calls it
  internally and composes its ≥50-word Spec check on top (change-code time); the validator uses
  presence-only (no word-count). One source, two callers.
- `sdlc issue validate [--issue N | <file> | --all]` (cmd/sdlc/issue.go): full validation —
  shells `vocabulary validate-instance --type issue` (frontmatter) + runs the shared section
  check; renders diagnostics; exit non-zero on failure. (On-demand = informative, full check.)
- Pre-merge gate: in `cmd/sdlc/preflight.go` (push + merge), add a deterministic
  `validateChangedIssues` step BEFORE the judge loop:
  - **frontmatter** check on **every** changed `workshop/issues/*.md` (old or new) — universal;
  - **section** check **only on newly-ADDED** issue files (git `--name-status` `A`) — grandfathers
    legacy/in-flight tickets, validates forward.
  - **Loud escape hatch:** `--no-validate` / `--force <reason>` ([[feedback_escape_hatch_loud_claim]])
    — a prominent WARN + the reason on use; never silent. It's a hard check, not an LLM judge.
- Process-level fake for the vocabulary binary in sdlc gate tests (injected runner seam — the
  vocabulary binary is external to sdlc, so model it process-level, not a function mock).
- **Tests:** the gate REJECTS a window where a changed issue has `status: in-progress`
  (frontmatter, even on a modified old file); REJECTS a newly-ADDED issue missing `## Plan`;
  PASSES a window that modifies a legacy ticket which lacks `## Done when` (grandfathered — added
  vs modified is the line); `--no-validate` bypasses with a loud warning; `sdlc issue validate
  --issue N` reports diagnostics for a hand-broken file.

### M3 — Generalize to a second datatype (pensive)
- `construct/vocabulary/pensive.cue`: `#Pensive` { type: "pensive", date: string,
  topic: string, mode: or(["ideas","eureka","thoughts"]), description: string,
  references?: [...string] } (closed). No `requiredSections` (pensive is narrative —
  confirmed against `construct/datatype/pensive.md`: required type/date/topic/mode/description,
  no required body sections).
- Validate a real `workshop/pensive/*.md` through `vocabulary validate-instance --type
  pensive` — proves the locator/validator path isn't issue-specific (the ONLY new thing per
  datatype is the `.cue`).
- **Tests:** a real pensive file passes; a `mode: musing` (bad enum) pensive fails.

## Reuse (ARCH-DRY — don't rebuild)
- **Frontmatter split** (M1): `cmd/vocabulary` cannot import `cmd/sdlc/internal/issue` across the
  internal boundary, so lift the canonical frontmatter split into `pkg/frontmatter` (already has
  `Description`) and have `cmd/sdlc/internal/issue.Parse` delegate — one source for the parse.
- **Section-presence policy** (M2): export `issue.CheckSectionsPresence` from
  `cmd/sdlc/internal/issue` (reusing `checkPlan` has-items + `checkDoneWhen` bullet-or-`related:`
  + Spec-present); `structural.go` composes its ≥50-word check on top; `preflight.go` +
  `sdlc issue validate` import it directly (both in cmd/sdlc — NO new pkg). Single source.
- `resolveVocab()` + `CueRunner` + `layergraph.MergeByName`: reuse in place (engine is in
  cmd/vocabulary).
- The `vocabulary` binary is already a built substrate dep (CI/make/weave invoke it).

## Risks / watch-items
- ~~`cue vet -d` semantics~~ — spiked, confirmed (above).
- ~~Date/timestamp scalars vs `string`~~ — spiked: cue's YAML loader keeps them as strings;
  `: string` is safe. The real-corpus test stays the guard for any other odd variant
  (`github_issue:` empty→null, `deps: []` — both spiked clean).
- The pure cue-stderr→diagnostic transform is the most fragile piece and the one the Done-when
  rests on ("actionable enough an LLM fixes it"). Make it a pure function, unit-test against
  captured stderr fixtures directly (not only end-to-end).
- Cross-binary shell at the merge gate: needs the vocabulary binary present; the process-level
  fake keeps sdlc gate tests hermetic.
- Scope discipline: well-formedness ONLY. Resist smuggling `Spec ≥ N words` into the
  validator (it's a soul-check — #122's explicit boundary).
