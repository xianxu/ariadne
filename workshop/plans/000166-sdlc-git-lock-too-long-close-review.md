# Boundary Review - ariadne#166 (whole-issue close)

| field | value |
|-------|-------|
| issue | 166 - sdlc git lock is too long |
| repo | ariadne |
| issue file | workshop/issues/000166-sdlc-git-lock-too-long.md |
| boundary | whole-issue close |
| window | b290512127f61337811d858315b2a02eb2f076b2..HEAD |
| command | sdlc close --issue 166 |
| reviewer | codex |

## 2026-07-07T17:03:00-07:00 - REWORK

The first boundary review found two blocking issues:

- Stale finalization validation covered HEAD and the issue file but not a precomputed project-file edit. Because `computeClose` can prepare `projectEditText` before the unlocked review, finalization could overwrite a concurrent project update after reacquiring only the ariadne repo lock.
- The plan listed a `RepoLockMode` entity, but the implementation used untyped string constants.

Resolution:

- `closeResult` now carries the original project text when a project edit is prepared.
- `CloseReviewSnapshot` now validates the project file still matches that original text before applying the cached edit.
- A regression test covers a real temp brain project file changed during the unlocked review.
- `RepoLockMode` is implemented as a typed internal value.

## 2026-07-07 - REWORK

The second boundary review confirmed the behavioral fixes and found remaining artifact/docs issues:

- `CloseReviewSnapshot` was still listed under Pure Entities even though its implementation reads git and files.
- Help/atlas stale-check prose mentioned only HEAD and the issue file, omitting prepared project-file edits.
- The generated sidecar contained captured prompt/diff text with whitespace that made `git diff --check` fail.

Resolution:

- The plan now classifies `CloseReviewSnapshot` as integration behavior.
- Helptext and atlas docs now mention stale checks for prepared project-file edits.
- This sidecar was normalized to preserve the review findings without carrying the huge captured prompt/diff payload that caused later boundary-review dispatches to exceed OS argument limits.
