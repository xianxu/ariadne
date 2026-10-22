# Boundary Review — ariadne#550 (whole-issue close)

| field | value |
|-------|-------|
| issue | 550 — merge runner waits forever when its head lacks the required local-check record; run it or exit with the next action |
| repo | ariadne |
| issue file | workshop/issues/000550-merge-runner-waits-forever-when-its-head-lacks-the-required-local-check-record-run-it-or-exit-with-the-next-action.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6c70244f02106bf8f7e612ae20587e12769f8b76..bdd23477ef613a7ebc0b0d38e503b61e6960f61b |
| command | sdlc close --issue 550 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T08:23:10-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation fulfills the explicit-refusal option: confirmed missing local checks end the retained wait with actionable commands, including when server CI masks their absence. Focused tests passed; no blocking correctness or architecture findings emerged.

Done-when: current — explicit demand was already an allowed outcome; the revisions refine diagnostics, cleanup, and race behavior within that contract.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: patch-whitespace-hygiene
    title: |
      Capture fixture leaves a whitespace-check warning
    detail: |
      cmd/sdlc/testdata/landing/pair779-missing-check.txt:25 adds a blank line at EOF, causing the pinned-range git diff --check to exit 2. This is the only instance reported across the window; remove the trailing blank line outside the captured output.
```

1. **Strengths**
   - `cigate.go:231` copies missing-check metadata into the typed error, preserving it independently of aggregate CI precedence.
   - `landingcar.go:76` uses typed evidence and the shared command formatter rather than parsing diagnostic text.
   - `landingqueue.go:213` joins readiness before ending wait evidence; settlement handling precedes terminal-cause substitution.
   - README, help, and atlas consistently explain refusal, retry, and admitted check execution.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** The fixture whitespace warning above.

5. **Test coverage notes:** The focused classifier, CI gate, car readiness, retained-wait, and race-ordering tests passed in **111.437s**. Coverage includes masked CI verdicts, satisfied head/ancestor/fallback records, read failures, ticket retirement, cancellation, admission, settlement, survivor handoff, and the public merge entry. The full suite was not rerun during this review.

6. **Architectural notes**
   - **ARCH-DRY: pass** — commands derive from one formatter; existing classification and refusal machinery are reused.
   - **ARCH-PURE: pass** — classification and formatting remain pure; Git, GitHub, and lock operations remain behind existing integration seams.
   - **ARCH-PURPOSE: pass** — the change addresses the missing-record family across aggregate CI verdicts and carries the result through retained waiting and survivor handoff.
   - **ARCH-EVOLVE: pass** — new metadata is in-process only; persisted formats and rewrite behavior are unchanged.

7. **Plan revision recommendations:** None. The listed entities exist at their documented locations, and unfinished landing steps remain explicitly unchecked.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b1bd0cd8-203c-4ac9-b777-939017dbe790","repository":"github.com/xianxu/ariadne","issue_id":"000550","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2303c849270d56da","reviewed_head":"bdd23477ef613a7ebc0b0d38e503b61e6960f61b","slot":"ariadne:4","requested_at":"2026-10-22T15:20:11.317433Z"},"phase":"finished","pid":56128,"started_at":"2026-10-22T15:20:11.494192Z","finished_at":"2026-10-22T15:23:01.335411Z","session_id":"01a16432-c7d0-7172-b9e7-edf5d46be820","outcome":"returned"}] -->
