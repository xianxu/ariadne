---
name: xx-issues
description: "Use when creating or editing issues in workshop/issues/."
---

# Issues

Issue *mechanics* — the canonical template, frontmatter fields, status values,
required sections — live in the `sdlc` binary. **Run `sdlc issue --help`** for
the contract, and use the verbs instead of hand-editing:

- `sdlc issue new "<title>"` — create (allocates the next ID; `--from-github N`
  seeds from a GitHub issue). Never hand-scan for the next ID.
- `sdlc issue set-status <status> --issue N` — flip status (carries the
  transition guards); `sdlc claim --issue N` to start work.
- `sdlc issue list` / `sdlc issue show <N>` — inspect.
- `sdlc close --issue N --verified '<evidence>'` — close. `--actual` is **measured, not typed**: omit it (close measures + adopts it; milestones still suggest) or run `sdlc actual --issue N` to preview; never hand-type the hours.

This skill keeps only the judgment the binary can't encode.

## Closing: what the binary leaves to you

`sdlc close` enforces the mechanics (status: done, `actual_hours`, atlas touch,
log entry) and measures the actual itself (the in-binary active-time-v3 engine
over the issue's commit window), printing the suggested `--actual`. For the full
per-segment breakdown, run `sdlc active-time`. The judgment is yours:
- inspect the per-segment table (`sdlc active-time`) for misclassified work;
- decide whether a discovered peer issue is real work or a stray mention;
- choose the rounded ACTUAL;
- write VERIFIED as behavior evidence ("tests pass"), not "code written".

Don't move the file to `workshop/history/` — that happens at periodic cleanup
(`sdlc push` / `sdlc merge`), not at close. Deep-dive on the v3 attribution
method: `brain/data/life/42shots/velocity/baseline-v3.md`.

## Log + side quests

- `## Log` is append-only — don't edit past entries. Use `### YYYY-MM-DD —
  session summary` for major-sitting wrap-ups (one paragraph: what landed, what
  deferred, in-flight decisions worth remembering); plain `### YYYY-MM-DD` for
  individual notes.
- `## Side quests` (optional, multi-day issues): one line per unbudgeted-but-
  shipped piece — name + ~time + commit ref. Pairs with the `side-quest:` commit
  verb (AGENTS.md §12); lets velocity calibration count effort that otherwise
  dissolves into the diff.

## Cross-repo references

When referencing an issue **outside the current repo** (e.g., a brain note
pointing at a parley issue, a parley issue depending on work in ariadne), use
the qualified form `<repo>#<NNN>`:

- `parley.nvim#123` — issue 123 in the `parley.nvim` repo
- `brain#42` — issue 42 in the `brain` repo

Within the current repo, keep using bare `#NNN`; the qualifier exists to
disambiguate across repos, so using it everywhere is just noise. This mirrors
GitHub's `owner/repo#NNN` convention; the repo slug is the sibling directory
name.

**Why the bare form needs disambiguation across repos:** issue numbers restart
at 1 in every repo, so `#42` is ambiguous when context spans repos, and it
collides with forked-upstream history (a fork's git log can mention the
upstream's old `#42` — a different issue). Qualified references are the
principled fix.

**Where qualified refs apply:** `deps:` frontmatter for cross-repo dependencies
(`deps: [parley.nvim#119]`); issue body / plan / log pointing at peer-repo work;
brain notes, project files, and roadmaps that span repos. Existing bare `#NNN`
references stay valid; tools that grep for issue refs should match both forms.

## Rules

- Keep slugs short + descriptive (3-5 words).
- Use `deps` for blocking relationships; qualified `<repo>#NNN` when the
  dependency lives in a peer repo.
- `## Log` is append-only.
