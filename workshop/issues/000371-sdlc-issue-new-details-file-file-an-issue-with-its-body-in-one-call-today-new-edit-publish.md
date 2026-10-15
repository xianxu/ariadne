---
id: 000371
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-15
estimate_hours:
card_mirror: '70d425a7172013f510a4020023f7460fabbe7069' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T00:11:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "5b408f42", done: "55abce04"}
---

# sdlc issue new --details FILE: file an issue with its body in one call (today: new, edit, publish)

## Problem

Filing an issue with content takes three steps: `sdlc issue new "title"` (publishes an empty template), edit the details file, `sdlc issue publish --issue N`. On 2026-10-14 the TL did this ~20 times; it twice collided with a pair:0 stage (the edit needs the resting checkout) and repeatedly waited on a repo lock between steps. Retro item (a) from couch-broadcast-web.

## Spec

`sdlc issue new "title" --details FILE` (and `--details -` for stdin): FILE holds the body sections (`## Problem`, `## Spec`, `## Done when`, optional Plan/Log), validated against the issue-file contract; the card and the full details publish in the one creation commit, behind the existing creation receipt. Works from any checkout without writing to it (so a staged or busy checkout is never touched). Partial bodies are fine; missing sections keep the template's.

## Done when

- `issue new --details FILE` creates and publishes card plus body in one call; the file never appears dirty in the invoking checkout (test).
- Invalid section structure is refused before reserving an ID (test).

## Plan

**Design.**
- **Pure core** (`internal/issue`, ARCH-PURE):
  - `ParseDetails(text) (map[string]string, error)` reads a details file. It holds `## <Section>` blocks only, scanned fence-aware with `markdoc.ScanMarkdownLines`, so a `## ` inside a code fence is body text.
  - Refused: frontmatter; a `# ` H1 (the title is the positional argument; a second H1 would also break `SplitCardWithFormat`); text before the first `##`; a heading that isn't one of the model's sections (`vocab.Issue().Sections()`); a section given twice; an empty file.
  - `ScaffoldSpec` gains `Sections map[string]string`. `Render` uses a supplied section's body in place of its seed. Sections not supplied keep the template's, including Log's dated heading.
  - `ProblemBody` (`--from-github`) stays as it is. A `--details` that also carries `## Problem` together with `--from-github` is refused: two sources for one section.
- **Shell** (`issue.go`): `--details FILE` (`-` for stdin) is read and parsed in `runIssueNew` before `openTracker` reserves anything. A bad file refuses with no ID taken and nothing on origin. The parsed sections go through the existing `render(id)` closure, so the card and the full details go in the one creation commit behind the existing receipt, and `checkDraft`'s bundle validation runs on the result. `--dry-run` prints the merged body. Nothing is written to the checkout; the details file is only read. The legacy path (`runLegacyIssueNew`) passes the same `Sections` to `Render`, so `--details` works there too.
- **`--no-claim`** (the TL's ask): filed from a numbered slot, the issue stays unclaimed. It skips `claimAtCreation` and prints the same unclaimed line as filing from `:0`. This replaces new-then-unclaim for filing on another slot's behalf.
- **ARCH-SECURE:** the details file is untrusted input, turned into a typed `map[section]body` at the boundary. Unknown structure refuses rather than guesses. **ARCH-ORDER:** validate before reserving, so a refusal has no effects. **ARCH-FUNERAL:** creates nothing new. **ARCH-MOCK:** the tests use the existing `newTrackerRepo` fake origin.

- [x] `issue.ParseDetails` and `Render` sections, with unit tests: a full body; a partial one (missing sections keep the template's); a fenced `## ` stays body; each refusal (frontmatter, H1, preamble, unknown heading, duplicate, empty)
- [x] `issue new --details FILE|-` and `--no-claim`. Tests against the fake origin: card and body published in one creation commit, the checkout untouched (`assertCheckoutUntouched`); an invalid file refused before any ID is reserved (the tracker's max ID unchanged, main unchanged); stdin; `--from-github` with a Problem in the file refused; `--no-claim` from a slot leaves it unclaimed; `--dry-run` shows the merged body
- [x] Help text for `issue new`, the `sdlc issue --help` contract if it lists the flags, the atlas, and the AGENTS §2 creation line if it names the flow

## Log

### 2026-10-14
- Plan: parse the details into sections before reserving; `Render` takes the sections; `--no-claim` per the TL. Quick flow.
- Built. Unit tests: `ParseDetails` (full body, fenced `## ` kept as body, an empty section, 8 refusals) and `Render` with sections. Fake-origin tests: one creation commit carrying card and body, checkout untouched; unknown section and `# ` title refused before reserving (main and issue-tracker unchanged); `--from-github` with a Problem refused; stdin with `--dry-run`; `--no-claim` from `pair:1` files unclaimed. `go test -run IssueNew` green. A live `--dry-run` against this repo renders the merged body. Docs: the `issue new` help, `sdlc issue --help`, AGENTS.base §2, the atlas.
