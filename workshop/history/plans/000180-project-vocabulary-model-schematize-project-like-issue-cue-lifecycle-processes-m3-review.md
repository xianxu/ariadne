# Boundary Review — ariadne#180 (milestone M3)

| field | value |
|-------|-------|
| issue | 180 — project vocabulary model: schematize project like issue (cue + lifecycle + processes) |
| repo | ariadne |
| issue file | workshop/issues/000180-project-vocabulary-model-schematize-project-like-issue-cue-lifecycle-processes.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | e3efc14aefe354c714a20fcff0f194ccaa213478..HEAD |
| command | sdlc milestone-close --issue 180 --milestone M3 |
| reviewer | codex |
| timestamp | 2026-07-16T13:42:40-07:00 |
| verdict | REWORK |

## Initial review — REWORK

The model-derived command family, lifecycle enforcement, guard ordering, and
pure-core separation were sound. Two correctness defects blocked the boundary:

- `RenderScaffold` interpolated free-text CLI fields directly into YAML. Colons
  could break parsing and `#` could silently alter the stored value.
- show, validate, and set-status joined unchecked slugs to `ProjectsDir`, so
  traversal could escape the modeled live-project directory.

README coverage for the new command family was also missing. Required
remediation: YAML-safe scalar serialization with hostile-value validation;
one shared slug/path resolver used by every slug consumer with traversal tests;
README, plan Core concepts/revision, and lessons updates.

## Remediation

Commit `ee146a4` addressed the review:

- Scaffold strings are deterministically double-quoted and typed reads decode
  them. Tests cover colons, comments, quotes, newlines, boolean/date-like text,
  and a real `vocabulary validate-instance --type project` subprocess.
- PURE `project.ResolvePath` accepts canonical lowercase kebab slugs only and
  is the sole locator used by new/show/validate/set-status. Traversal refuses
  before downstream IO.
- README.md documents the runnable M3 workflow; the plan Core concepts and
  revision record plus `workshop/lessons.md` capture the new boundaries.

Verification: `go test ./...`, `construct/vocabulary/vet_test.sh`, and
`git diff --check` pass; live list/show/validate succeeds against
`workshop/projects/project-management-primitive.md`.

---

## Re-review — 2026-07-16T13:47:32-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| boundary | milestone M3 |
| window | e3efc14aefe354c714a20fcff0f194ccaa213478..HEAD |
| reviewer | codex |
| verdict | FIX-THEN-SHIP |

The re-review confirmed both blockers and the documentation gap were resolved.
ARCH-DRY, ARCH-PURE, and ARCH-PURPOSE passed. No Critical correctness finding
remained.

The sole Important finding was artifact hygiene: the generated sidecar had
retained 460 KB of raw prompt, diff, and duplicated reviewer transcript. This
bounded record replaces that raw capture while preserving the verdicts,
findings, remediation, and verification evidence required for the boundary.
