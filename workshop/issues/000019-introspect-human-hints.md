---
id: 000019
status: working
deps: [000018]
created: 2026-05-01
updated: 2026-05-01
---

# Human Hints as Strong Signal in Introspection

## Problem

The introspection pipeline today only consumes `~/.claude/projects/*.jsonl`
transcripts. Patterns reach the deployed `introspect-<activity>` skills only if
they show up *inside* a transcript as a redirect/endorsement, and only after
clearing the ≥2-distinct-segments threshold during the cluster pass.

This means:
1. A user who *knows* a rule should apply has no clean way to inject it.
   They must wait for the rule to manifest organically across two sessions and
   survive the cluster pass — slow, lossy, and easy to miss.
2. The current heuristic blind spots (rules the LLM extractor under-weights,
   subtle preferences that don't surface as redirects) have no escape hatch.

We want a side-channel: **explicit human hints** that the next pipeline run
treats as strong signal, each hint worth its own cluster, bypassing the
≥2-segment threshold. Hints are still subject to user review and retirement —
they're hints, not commandments — but they're authoritative by default.

While we're here, consolidate the scattered `~/.claude/introspect-cache/` and
(planned) `~/.claude/introspect-versions/` paths under a single
`~/.claude/introspect/` root for tidiness.

## Spec

### Directory consolidation

Migrate to a single root. The deployed skill files at
`~/.claude/skills/introspect-*/SKILL.md` stay where they are — Claude Code's
discovery mechanism owns that location.

```
~/.claude/introspect/
├── cache/<run-id>/         # was ~/.claude/introspect-cache/<run-id>/
├── hints/<activity>/*.md   # NEW — human hints
└── versions/v<N>/          # was ~/.claude/introspect-versions/ (still M7-unbuilt)
```

`<activity>` is one of the existing taxonomy buckets:
`debugging`, `exploration`, `planning`, `implementation`, `brainstorming`.

### Hint file format

Small markdown file under `hints/<activity>/<slug>.md`:

```markdown
---
activity: debugging
created: 2026-05-01
---

## Rule: <short imperative title>

<rule body — one or two paragraphs, same shape as a rendered cluster rule>

**Why:** <optional rationale — past incident, strong preference, etc.>
```

User authors hints either by hand (drop a file in) or via the umbrella skill's
`/xx-introspect hint` subcommand (see below). The slash command is a thin
writer over the same directory the pipeline reads — both paths produce
identical files, no duplication.

### Hint authoring via slash command

Extend the umbrella skill (`construct/local/introspect/SKILL.md`) with a
`hint` subcommand alongside the existing `extract` and `load`:

- `/xx-introspect hint <activity> <rule>` — single-shot: skill picks a slug
  from the rule, drafts the markdown body, writes the file, shows it back
  for confirmation.
- `/xx-introspect hint` — no args: skill infers activity from recent session
  context, prompts for the rule, then same as above.
- `/xx-introspect hint --list [<activity>]` — list existing hints, by activity
  if specified.
- `/xx-introspect hint --retire <slug>` — delete a hint file (same effect as
  retiring it through the review UI).

Activity arg validates against the five-bucket taxonomy; unknown values
prompt for clarification rather than silently mis-filing.

### Pipeline changes

The pipeline gains one new input source and one new transition.

**Cluster pass (new behavior):**
- Read hint files in addition to the existing `patterns.json`.
- Emit each hint as its own cluster, `source: hint`, no ≥2-segment requirement.
- For each hint, scan the run's evidence for *contradicting* transcript signals
  (user redirects against the hint's rule). If found, attach a
  `retirement_candidate: true` flag with pointers to the contradicting moments.
- Hints with no contradicting evidence pass through unchanged.

**User review (existing surface, extended):**
- Hint clusters render alongside extracted clusters, pre-marked `source: hint`.
- Retirement-candidate hints are surfaced first, with the contradicting
  evidence quoted so the user can decide: *keep* (ignore the contradiction),
  *modify* (edit the file), or *retire* (delete the file).
- Decisions write back to `hints/<activity>/` directly. Deletion is the retirement.

**Write-back to SKILL.md:**
- Hint-sourced rules render in the same format as extracted rules, but with a
  trailing `**Source:** human hint` line in place of the evidence-segment line.
- This keeps the SKILL.md format readable and lets future humans (and the
  pipeline itself) tell extracted from authored without scraping JSON.

### Retirement semantics

Eligible-for-retirement, not frozen. Concretely:

- A hint is **persistent** across runs — it lives in `hints/`, gets re-emitted
  every cluster pass.
- A hint becomes a **retirement candidate** when the cluster pass finds
  transcript evidence contradicting it. The user decides at review time.
- A hint is **retired** by deleting the file (manual or via the review UI).
  No tombstone — if you re-add it, it's a new hint.

This matches the existing "user-in-the-loop on every model judgment" principle
and the precision-over-recall rule. The pipeline never auto-deletes.

## Plan

- [x] **Path migration** (M1, 2026-05-01)
  - [x] Move `~/.claude/introspect-cache/` → `~/.claude/introspect/cache/`
        (one-time mv; backward-compat symlink left in place).
  - [x] Reserve `~/.claude/introspect/versions/` (M7 still unbuilt, but the
        path convention lands now).
  - [x] Create `~/.claude/introspect/hints/{debugging,exploration,planning,implementation,brainstorming}/`.
  - [x] Update `introspect-extract.sh`, `scripts/README.md`,
        `construct/local/introspect/SKILL.md`, and `atlas/introspect.md`
        to the new paths.

- [x] **Hint authoring (slash command)** (M2, 2026-05-01)
  - [x] Extend `construct/local/introspect/SKILL.md` umbrella with a `hint`
        subcommand: parse args, infer activity if missing, draft body,
        write to `~/.claude/introspect/hints/<activity>/<slug>.md`, confirm.
  - [x] Implement `--list` and `--retire <slug>` flags (read/delete only,
        no LLM call needed).
  - [x] Slug derivation: lowercase-hyphenated truncation of the rule's
        imperative title; collision-resolve by appending `-2`, `-3`, …
  - [x] Activity validation against the five-bucket taxonomy
        (debugging, exploration, planning, implementation, brainstorming).

- [ ] **Hint ingestion**
  - [ ] Add `read_hints.py` (or equivalent) that walks `hints/<activity>/`
        and emits a JSON array matching the `clusters.json` shape.
  - [ ] Cluster pass: union hints with extracted clusters; tag each
        with `source: "hint" | "extracted"`.

- [ ] **Retirement detection**
  - [ ] During cluster pass, run a contradiction probe per hint against the
        run's `patterns.json` / `moments.jsonl`. Cheap LLM call:
        "does any of this evidence contradict the hint?"
  - [ ] Attach `retirement_candidate` + `contradicting_moments` to flagged hints.

- [ ] **Review UX**
  - [ ] Cluster review surfaces hint clusters first (retirement candidates at top).
  - [ ] Provide three actions: keep / edit / retire (delete file).
  - [ ] Existing precision-over-recall behavior (ambiguous → skip) applies
        only to extracted clusters; hints don't go to ambiguous.

- [ ] **Write-back**
  - [ ] Update write-back template so `**Source:** human hint` renders for
        hint-sourced rules.
  - [ ] Verify the rendered SKILL.md round-trips cleanly through the next
        pipeline run (i.e. the source marker doesn't get re-extracted as a
        moment).

- [ ] **Docs**
  - [ ] Update `atlas/introspect.md` (directory layout, hint format,
        retirement semantics).
  - [ ] Update `construct/local/introspect/SKILL.md` (umbrella skill body) to
        mention hints as a supported input.
  - [ ] Update `construct/local/introspect/scripts/README.md` composition
        recipes if env vars changed.

- [ ] **Verification**
  - [ ] Author one real hint in `hints/debugging/` by hand, run a full
        pipeline pass, confirm it appears in `introspect-debugging/SKILL.md`
        with the source marker.
  - [ ] Author another via `/xx-introspect hint debugging "<rule>"`,
        confirm the file lands in the same place and round-trips identically.
  - [ ] Verify `--list` shows both, and `--retire <slug>` removes one of them
        cleanly.
  - [ ] Stage a synthetic contradicting moment, rerun, confirm the hint shows
        up as a retirement candidate in review.
  - [ ] Retire it via the review UI, confirm the file is deleted and the rule
        is gone from the next regenerated SKILL.md.

## Log

### 2026-05-01 — M1 path migration

- Created `~/.claude/introspect/{cache,hints,versions}/` and the five
  per-activity `hints/` subdirs.
- Moved the three existing run dirs from `~/.claude/introspect-cache/` into
  `~/.claude/introspect/cache/`. Left `~/.claude/introspect-cache` as a
  symlink pointing at the new location so the deployed
  `~/.claude/skills/introspect-*/SKILL.md` evidence pointers still resolve;
  the symlink can be removed at any time (next pipeline run will produce
  skills that reference the new path directly).
- Updated `construct/local/introspect/SKILL.md`,
  `construct/local/introspect/scripts/README.md`,
  `construct/local/introspect/scripts/introspect-extract.sh`, and
  `atlas/introspect.md` to point at the new layout.
- Manifest (`construct/base.manifest`) tracks no introspect paths, so no
  base-layer change needed.

### 2026-05-01 — M2 hint authoring slash command

- Added `/xx-introspect hint` subcommand spec to the umbrella
  `construct/local/introspect/SKILL.md`: authoring (with or without args),
  `--list`, `--retire <slug>`. Updated the skill's top description and
  three-subcommand list.
- The orchestrating Claude can now author/list/retire hints from any
  session that loads the `xx-introspect` skill. No pipeline-side work
  is required — hints just become files; M3 is what teaches the cluster
  pass to ingest them.
- Decision: `code-review` is part of the classification taxonomy but has
  no deployed `introspect-code-review` skill yet, so hints in that bucket
  are rejected for now (will become accepted once the deployed skill exists).
- `created:` field uses local date (YYYY-MM-DD) rather than ISO timestamp;
  hints don't need sub-day resolution.
