# Boundary Review — ariadne#377 (whole-issue close)

| field | value |
|-------|-------|
| issue | 377 — tl-lead: port the ops TL rules added after #350's read (Operator instructions + QA files, Needs-you only asks, CI glance, render-from-binaries, quiet pane) and opinstr.py; fix status frontmatter key |
| repo | ariadne |
| issue file | workshop/issues/000377-tl-lead-port-the-ops-tl-rules-added-after-350-s-read-operator-instructions-qa-files-needs-you-only-asks-ci-glance-render-from-binaries-quiet-pane-and-opinstr-py-fix-status-frontmatter-key.md |
| boundary | whole-issue close |
| milestone | — |
| window | 18915f6c17d9b4aeb5e618d3f736705ec355d2b3..043a22dc7e22e03a0a36bc4577fe5a533c5ce843 |
| command | sdlc close --issue 377 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T21:18:26-07:00 |
| verdict | SHIP |

## Review

Tests pass and the Go port matches `opinstr.py`. Last checks: whether README lists the `tl-lead` subcommands, and how Couch section reads handle `###` subsections.

```verdict
verdict: SHIP
confidence: high
```

The change delivers what the issue asks for, and I found nothing that blocks shipping. `help.md` now has every rule in the plan's gap table (G1–G8, G10, G11), the five extra rules from the fresh comparison, the `couch-owner:` key with a note that `owner:` is still read, and the corrected section order. I read `opinstr.py` from `~/workspace/ops/skills-draft/xx-tl/scripts/` next to `cmd/tl-lead/instructions.go`, and the port matches it. Both use the same bullet regex, the same answered regex, the same position-based numbering, the same trailing-blank trimming, and the same rule of appending the link after the last non-blank line. Where the Python scanned for `## ` headings, the Go code reads the section through the existing `SessionSync.ReadSection` call. The never-retry hash guard on `link` is implemented and tested. `go test ./cmd/tl-lead ./cmd/tl-worker` passes, and `go vet` and `gofmt` report nothing.

1. **Strengths**
   - `cmd/tl-lead/instructions.go:220-259`: the parser and linker are pure functions over the section text. The golden fixture `opSection` covers all four cases the issue names (multi-line bullets with blank lines inside, repeated numbers, image lines, both answered forms) plus preamble and empty-section cases.
   - `linkInstruction` (`instructions.go:281`) makes one write attempt, and treats a changed hash before the write and a stale result from Couch as the same exit-6 refusal. `TestInstructionsLinkNeverRetries` uses the stateful fake's `editBefore` and checks `writeCalls == 1`, so a later change that sends `link` through the retrying `writeSection` would fail that test.
   - `TestLeadHelpNamesSubcommands` (`main_test.go:142`) now walks the cobra command tree instead of keeping a hand-written list. It has a floor (fail if fewer than 10 commands are walked), so an empty tree can't pass, and future verbs are covered automatically.
   - `TestLeadHelpCoversPlaybook` checks one phrase per ported rule, so a later edit can't silently drop a rule.
   - The atlas (`atlas/workflow/tl-skills.md`) was updated with the new verbs, the retired `opinstr.py`, and the new section order.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `instructions.go:257`: `TARGET` is written into the operator's section exactly as given. A target containing a newline would add new lines, possibly a fake `N. ` bullet. A target containing `)` would break the link. It's cheap to reject targets with newlines or whitespace (ARCH-SECURE).
   - `instructions.go:257`: `TrimRight(" \t")` and Python's `rstrip()` differ only on `\r` and other unusual whitespace. That matters only for a file with CRLF line endings.
   - `render.go`, `status_test.go`, `trains_test.go`: gofmt-only changes, outside the issue's scope but harmless.

5. **Test coverage notes:** the tests cover the parser, the linker, list/link through the fake, the hash-mismatch refusal, the race where the operator edits between list and link, the lock (exit 3), a missing `--hash`, a non-numeric POS, and an already-answered bullet. One case is untested: `list`/`link` when the `## Operator instructions` section is missing. It goes through `refusal`, the same path the other verbs use.

6. **Architecture notes**
   - **ARCH-DRY:** pass. The section read reuses Couch's section read; `writeBody` and `refusal` are shared; the single-shot write is a deliberate variant of `writeSection`, not a copy of it.
   - **ARCH-PURE:** pass. The parser, linker and `renderOpen` are pure, and the cobra commands are a thin layer over the seam.
   - **ARCH-PURPOSE:** pass. Every row in the gap table lands in the help or the binary, each with a test guard. Removing the ops draft is ops-side work the issue's Spec assigns after this lands.
   - **ARCH-MOCK:** pass. No new external surface; both verbs go through `SessionSync`, which already has a stateful fake and an opt-in live conformance test.
   - **ARCH-CONSTRAINTS:** pass. One read and at most one write per call.
   - **ARCH-SECURE:** pass, apart from the Minor `TARGET` validation above. Operator text is only scanned and re-emitted, never executed.
   - **ARCH-ORDER:** pass. The event most likely to be mishandled (the operator editing between list and link) is pinned by the hash and tested with an injected ordering.
   - **ARCH-FUNERAL:** pass. Nothing new is persisted; the temporary body file is removed by `writeBody`.

7. **Plan revision recommendations:** none. The plan's Revisions section already matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: unvalidated-write-argument
    title: |
      instructions link writes TARGET verbatim; a newline or ')' in it corrupts the operator's section
    detail: |
      instructions.go:257 appends TARGET exactly as given. A newline in it adds lines that can parse as a new bullet; reject newlines and whitespace in TARGET.
  - id: new
    severity: Minor
    kind: defect
    family: port-parity-whitespace
    title: |
      linkBullet trims only space and tab, where opinstr.py rstrip() also trims CR
    detail: |
      Matters only for a status file with CRLF line endings; strings.TrimRightFunc(unicode.IsSpace) would match the Python.
```

---

## Re-review — 2026-10-14T21:20:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 377 — tl-lead: port the ops TL rules added after #350's read (Operator instructions + QA files, Needs-you only asks, CI glance, render-from-binaries, quiet pane) and opinstr.py; fix status frontmatter key |
| repo | ariadne |
| issue file | workshop/issues/000377-tl-lead-port-the-ops-tl-rules-added-after-350-s-read-operator-instructions-qa-files-needs-you-only-asks-ci-glance-render-from-binaries-quiet-pane-and-opinstr-py-fix-status-frontmatter-key.md |
| boundary | whole-issue close |
| milestone | — |
| window | d61f045ccd0821cdf86f26e2a8ff627cf162eea0..132a4e367c77022d1bbfa9badb1f1e7e204b05b6 |
| command | sdlc close --issue 377 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T21:20:25-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both prior Minor findings are fixed in the interdiff, and each fix has a regression test that fails without it. In a disposable checkout I reverted only the two changes to `instructions.go` and ran `TestLinkBullet`. It failed in five places: the four bad TARGETs (`""`, `"qa/a b.md"`, a value containing a newline, `"qa/(x).md"`) were no longer refused, and the CRLF case returned `"1. a\r → [answer](…)"`. I then restored the file and confirmed the tree was clean. On the pinned head, `go test ./cmd/tl-lead` passes, and `gofmt` and `go vet` report nothing. Apart from the code, the interdiff adds only the generated gate ledger and the close-review record. `## Spec`, `## Revisions` and `## Done when` are unchanged, and no conflict resolutions are listed.

1. **Strengths**
   - `cmd/tl-lead/instructions.go:66` handles the whole class the finding names, not just newlines. Any `unicode.IsSpace` character (which covers newline, CR and tab), either parenthesis, and an empty target are all refused. That covers both ways the finding showed corruption: a fake bullet and a link that ends early.
   - The check sits in the pure `linkBullet`, so `linkInstruction` reaches it after the hash check and before any write (`instructions.go:112`). A bad TARGET never reaches `WriteSection`.
   - `instructions.go:79` now uses `TrimRightFunc(unicode.IsSpace)`, which matches Python's `rstrip()` on all whitespace, not only on CR.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - Nit, not raised as a finding: the refusal message says TARGET "must be a relative path", but an absolute path with no whitespace or parentheses is accepted.

5. **Test coverage notes:** `instructions_test.go:103-111` covers each rejection rule and the CRLF trim, and both go red when the fix is reverted.

6. **Architectural notes (interdiff only)**
   - **ARCH-DRY:** pass. The check lives in one place.
   - **ARCH-PURE:** pass. The validation is in the pure linker and is tested without IO.
   - **ARCH-PURPOSE:** pass. Both findings were fixed at the class level: all whitespace and both parentheses for BR-1, all Unicode whitespace for BR-2.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      instructions.go:66 refuses an empty TARGET, any unicode whitespace and either paren before linking; the refusal loop in instructions_test.go:103 goes red when the check is reverted.
  - id: BR-2
    disposition: addressed
    note: |
      instructions.go:79 trims with TrimRightFunc(unicode.IsSpace), matching rstrip(); the CRLF test at instructions_test.go:109 goes red when reverted to TrimRight(" \t").
```
