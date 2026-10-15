# Boundary Review — ariadne#371 (whole-issue close)

| field | value |
|-------|-------|
| issue | 371 — sdlc issue new --details FILE: file an issue with its body in one call (today: new, edit, publish) |
| repo | ariadne |
| issue file | workshop/issues/000371-sdlc-issue-new-details-file-file-an-issue-with-its-body-in-one-call-today-new-edit-publish.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0820d860ceb54656f8a7c6cab6f209e2b6a65729..4ad181aafb45810d42611e4881e3245a0801d1af |
| command | sdlc close --issue 371 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T01:32:27-07:00 |
| verdict | SHIP |

## Review

Tests passed except one legacy-mode test, which also fails at the base commit in this checkout, so the diff didn't cause it. Writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

The diff delivers what the Spec and Plan promise. `issue.ParseDetails` is a pure parser that reads code fences and refuses any structure it doesn't expect. `Render` takes a `Sections` map, and sections the file leaves out keep the template's text. `--details FILE|-` is parsed in `runIssueNew` before `openTracker` or the legacy path, so a bad file reserves no ID and publishes nothing. The parsed sections go through the existing `render(id)` closure, so the card and the full body publish in the one creation commit. `--no-claim` skips `claimAtCreation`. The help text, `sdlc issue --help`, the atlas and AGENTS.base §2 are all updated. Results of the test runs:
- The new unit tests pass.
- The new fake-origin tests pass.
- `go vet` is clean.
- The `IssueNew|Help` run has one failure, `TestLegacyRepoLockConcurrentIssueNewSerializesAllocation`. It also fails at base `0820d860` in this checkout, so it comes from the review environment, not this diff.

Nothing blocks SHIP. There are only Minor notes.

1. **Strengths**
   - `cmd/sdlc/internal/issue/details.go:23-26`: a fence that never closes is caught by running `FenceSpans` both ways (`UnterminatedIsFenced` vs `UnterminatedIsProse`) and comparing. This reuses the shared fence scanner instead of adding a new one (ARCH-DRY).
   - The parse happens before anything is reserved (`issue.go:370`), and `TestIssueNewDetailsRefusedBeforeReserving` checks that neither main nor `issue-tracker` moves. That is the ordering guarantee the issue asks for (ARCH-ORDER).
   - The file is untrusted input and becomes a typed `map[section]body` at the boundary, checked against `vocab.Issue().Sections()`. Unknown or repeated headings, frontmatter, an H1 or a preamble are refused, never guessed at (ARCH-SECURE).
   - The section names come from the cue model, not a hardcoded list, so the parser and `Render` work from one source of section names.
   - `TestIssueNewDetailsPublishesBodyInOneCall` checks three things: exactly one main commit, the card carries the Problem, and the checkout is untouched (`assertCheckoutUntouched`).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/sdlc/helptext/issue.md:123-125`: the inserted sentences moved the "(`start-plan` branches from fresh main)" note. It used to follow "next main merge"; now it hangs off the `--no-claim` sentence, where it means nothing.
   - `issue.go:497`: `--no-claim` from `:0` prints the "(--no-claim)" notice instead of the "filed outside a numbered slot" one. It's harmless but slightly misleading.
   - `ParseDetails` only treats a line as a heading when `## ` starts at column 0. CommonMark also allows up to 3 leading spaces, so an indented `  ## Spec` becomes body text, or is refused as preamble. Since the parser refuses rather than guessing, this is acceptable.
   - In the legacy path, `--details` works because `Sections` is passed through (`legacymode.go:157`), but no test covers it.

5. **Test coverage**
   - Unit level: full body, a `## ` inside a fence, an explicitly empty section, all 8 refusals, and `Render` keeping Log's dated heading.
   - Integration, against the stateful fake origin: one commit, refusal before reserving, the `--from-github` conflict, stdin with `--dry-run`, and `--no-claim` from `pair:1`.
   - Gaps: no legacy-path `--details` test, and no end-to-end check that a fenced `# ` inside a supplied section survives `SplitCardWithFormat`. That case is only checked at the unit level.

6. **Architecture**
   - **ARCH-DRY: pass.** It reuses `FenceSpans`, `Render` and the `render(id)` closure.
   - **ARCH-PURE: pass.** `ParseDetails` and `Render` are pure; the file and stdin are read only in the shell, and stdin is injected through the `issueNewStdin` variable.
   - **ARCH-PURPOSE: pass.** Both Done-when items are delivered, plus `--no-claim`, and the legacy path gets the same sections.
   - **ARCH-MOCK: pass.** The tests use the existing `newTrackerRepo` fake origin.
   - **ARCH-CONSTRAINTS: pass.** The input is a small file read once; there's no size cap, which is acceptable for a CLI.
   - **ARCH-SECURE: pass**, as above.
   - **ARCH-ORDER: pass.** It holds no state between events: validation finishes before the reservation's single effect.
   - **ARCH-FUNERAL: pass.** It creates nothing durable beyond the existing issue artifacts, and it only reads the details file.

7. **Plan revisions:** a small accuracy fix. The plan says `ParseDetails` is scanned "with `markdoc.ScanMarkdownLines`", but the code uses `FenceSpans` (the same markdoc fence logic) plus the unclosed-fence check. Suggested entry: "## Revisions — ParseDetails uses `FenceSpans` (both unterminated modes) rather than `ScanMarkdownLines`, and refuses an unclosed fence."

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: doc-edit-displaces-context
    title: |
      helptext/issue.md: the start-plan parenthetical now attaches to the --no-claim sentence
    detail: |
      The inserted --details/--no-claim sentences separated "(`start-plan` branches from fresh main)" from "next main merge", which it qualifies. Move the parenthetical back after "merge".
  - id: new
    severity: Minor
    kind: defect
    family: test-coverage-parallel-path
    title: |
      No test covers --details on the legacy (untracked) issue new path
    detail: |
      legacymode.go passes Sections to Render, but no test checks the body there; a legacy-mode fixture asserting the supplied Spec appears would pin it.
```
