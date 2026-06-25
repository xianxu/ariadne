---
id: 000124
status: open
deps: [ariadne#122]
github_issue:
created: 2026-06-24
updated: 2026-06-24
estimate_hours:
---

# Instance-conformance: validate typed-markdown artifacts against their datatype schema (extract then cue vet)

## Problem

#122 makes `construct/vocabulary/issue.cue` the formal type of an issue (frontmatter
enum/fields + lifecycle), and wires `sdlc`'s *verbs* to it. But nothing validates that
a real `workshop/issues/NNNNNN-*.md` **file** conforms to that type. A direct edit —
an LLM "tidying" frontmatter to `status: in-progress` (the enum says `working`), a
bulk migration, a mangled merge, a `nous-resolve` AI-merge — bypasses the verb guards
and lands an ill-formed instance that no tool flags. Symptoms are silent: `claim`
assumes already-claimed, `close` clobbers, `state` mis-categorizes. Because issue files
publish to `main` (the bulletin board), the bad value propagates to peer agents.

The shell around a typed artifact's *shape* is today only **prose** (`AGENTS.md`: "do
NOT hand-edit `status:`") — a soft shell an LLM violates by default, since the Edit
tool knows nothing about the lifecycle. The vocabulary layer makes a **hard** shell
possible for the first time (there's now a formal `#Issue` to vet against); this issue
builds it. Generalizes beyond `issue`: it's the **instance-conformance** closed-loop
for *all* typed markdown (datatype = prose skill + schema). Design captured in
`workshop/pensive/2026-06-24-01-pensive-cue-schema-layer-nouns-verbs.md` ("Typing
markdown — the general frame").

## Spec

A validator: **`artifact → (locator) → list of typed instances → cue vet each`.**

- **The validator is well-formedness only.** Enum membership, required fields, field
  types, required/allowed sections. It does **not** check semantic quality — `Spec ≥ 50
  words` and friends are soul-checks in disguise and belong to the LLM, not here (see
  #122 Log + pensive). The schema defends the skeleton; the LLM owns the soul.
- **The extractor/locator is the genuinely-new component** (`cue vet` is the easy
  half). Per type: where do instances live and how are they normalized into the
  schema's shape? Generic locators for frontmatter (YAML) and sections (`## ` split);
  bespoke locators for fenced blocks / `- [ ]` lines. Open design question: can a
  locator be *declared* per datatype, or do bespoke ones stay imperative glue?
- **No deterministic write-back.** The validator emits clear diagnostics; the fix is the
  LLM (or the human in parley) editing the markdown, then re-validate — same agentic
  loop. Deterministic = the check; LLM = the actuator.
- **One binary, many triggers** (single mechanism): the agent runs it after an Edit (or
  a PostToolUse hook fires it), parley shells out on save (LSP-style inline
  diagnostics), the pre-merge gate runs it before `main`, CI runs it on PR. The type
  lives in *one* `construct/vocabulary/*.cue` all of them read.
- **Reuse, don't rebuild:** the frontmatter parse (`pkg/frontmatter`), the
  `cmd/vocabulary` cue runner + DAG-merge (#122 M2), and the section-extraction the
  `sdlc` structural gates already do imperatively.

## Done when

- A validator (`sdlc issue validate` and/or a `vocabulary check-instance` surface) takes
  a typed markdown file, extracts its instances, and `cue vet`s them against the type;
  a `status: in-progress` issue file is **rejected with a clear diagnostic**, a valid
  one passes.
- It is wired at ≥1 fail-closed boundary — the **pre-merge / `sdlc push` gate** (the
  high-value one: catches it before `main`) — and is invocable on demand by an agent.
- Diagnostics are actionable enough that an LLM, given the output, fixes the file and
  re-validates clean (the agentic-loop write-back).
- Generalizes past `issue`: at least one other datatype (e.g. `pensive`) gets a
  `construct/vocabulary/<type>.cue` and validates through the same path — proving the
  locator/validator isn't issue-specific.
- No semantic-quality checks smuggled into the validator (well-formedness only).

## Plan

- [ ] Design at `start-plan` (the locator model — declared vs imperative per type; where the validator surface lives; which boundaries wire it)
- [ ] Frontmatter-instance validation: extract frontmatter → `cue vet` against `#Issue`; reject the bad-status scenario with a clear diagnostic
- [ ] Section-instance validation: extract `## ` sections → vet required/allowed/present (NOT word-count quality)
- [ ] Wire the pre-merge / `sdlc push` gate; make it agent-invocable
- [ ] Generalize to a second datatype to prove the path isn't issue-specific

## Log

### 2026-06-24

- Filed as a follow-up to #122 (deps: ariadne#122 — needs the vocabulary layer + the
  `cmd/vocabulary` cue infra). Motivated by the hand-edited-bad-status walkthrough: the
  formal enum from #122 guards *verbs*, not *files*; this closes the instance-conformance
  loop so a typed artifact actually defends its own shape.
