---
id: 000201
status: working
deps: []
github_issue:
created: 2026-08-21
updated: 2026-08-25
estimate_hours: 1.09
started: 2026-08-25T11:51:47-07:00
---

# Boundary-review artifacts persist the harness transcript, not just the review

## Problem

Every `sdlc close` / `milestone-close` review document in `tools#4` opens with a
line that is not review content:

```
Ignoring 6 permissions.allow entries from .claude/settings.json: this workspace
has not been trusted. Run Claude Code interactively here once and accept the
trust dialog, or set projects["<repo>"].hasTrustDialogAccepted: true in
~/.claude.json.
```

The reviewer's stderr is being concatenated into the persisted artifact, so the
preamble lands once per round. In `tools#4` it now appears at
`close-review.md:18, 251, 485, 674, 878, 1043` — six occurrences, one per round,
and it was flagged by the reviewer itself at rounds 3, 4, 6, 7 and 8 as a
finding it could not converge because the cause is outside that repo.

Two distinct defects sit behind it:

1. **The artifact captures stderr it should not.** A review document should hold
   the review. Harness diagnostics belong on the operator's terminal, not
   committed into `workshop/plans/`. Anything the reviewer writes to stderr
   before the verdict block should be dropped or routed.

2. **A recurring finding with no home converges nowhere.** The reviewer
   correctly re-reports it every round because nothing in `tools` can fix it, so
   it costs a finding slot per round forever. Worth considering whether the
   review protocol should let a finding be dispositioned "external, tracked at
   `<ref>`" once, rather than re-litigated each boundary.

## Spec

- `*-review.md` contains boundary metadata plus the reviewer's semantic final
  response: verdict, summary, findings, test and architecture notes, plan
  revisions, and the structured findings fence. It contains no harness
  diagnostics, progress stream, tool transcript, prompt echo, or input diff.
- `*-gate.md` remains the machine-readable finding/disposition ledger and the
  source for the next round's `PriorFindings`; it is not a raw process log.
- The judge dispatch boundary captures stdout and stderr separately for Claude,
  Codex, and Gemini. Verdict parsing, structured-finding parsing, terminal
  review display, and sidecar persistence consume semantic stdout only.
- Captured stderr remains visible on the command's stderr, including for a
  non-zero reviewer exit or failed launch. Heartbeat output uses that same
  terminal diagnostic sink.
- Existing sidecars are not rewritten fleet-wide. Prompt transport, review
  window sizing, and reviewer checkout isolation remain separate work.
- Reproduce the original trust-dialog contamination by running `sdlc close` in
  a repo whose path is absent from `~/.claude.json`'s trusted projects.

## Notes

Surfaced by `tools#4`, where it recurred across eight close rounds. Related to
ariadne#195 (finding-family escalation), which addresses the general problem of
findings that recur without converging, but by a different mechanism.

## Revisions

### 2026-08-25 — define the artifact as review output, not process output

**Reason:** pair#146 exposed the same contract defect at a much larger scale.
Codex's combined process output contains the full input prompt and diff, tool
commands/results, harness diagnostics, and a repeated final answer. Five M3
rounds accumulated into a 5.03 MB, 74,289-line `*-review.md`; once committed,
that generated artifact entered the next review diff and recursively amplified.

**Delta:** the durable surfaces have semantic roles rather than mirroring a
subprocess stream:

- `*-review.md` stores boundary metadata plus the reviewer's final response:
  verdict, summary, findings, test notes, architecture notes, plan revisions,
  and the structured findings fence.
- `*-gate.md` remains the machine-readable finding/disposition ledger. It is
  the source rendered into the next round's `PriorFindings` block.
- Harness diagnostics and agent progress belong on the operator's terminal.
  They are not review content and are never persisted in either artifact.
- The dispatch boundary separates the agent's semantic final output from its
  diagnostic stream before verdict parsing or sidecar persistence. The rule is
  agent-independent even where individual CLIs route their streams differently.
- Existing raw sidecars are not silently rewritten fleet-wide. pair#146's
  oversized M3 artifact is condensed in that repo before its blocked close is
  retried.

Review-checkout isolation is deliberately separate (#204). Review-window and
large legitimate payload transport remain #162. The external-finding lifecycle
question remains the concern already tracked by #202/#195.

## Done when

- A fake reviewer that writes a valid verdict to stdout and diagnostics to
  stderr produces a sidecar containing the verdict/findings but none of the
  diagnostics.
- Diagnostics remain visible on the command's stderr, including on a failed
  reviewer launch or non-zero reviewer exit.
- Verdict parsing, structured-finding parsing, heartbeat reporting, and both
  synchronous and heartbeat dispatch paths consume the same semantic output.
- Claude, Codex, and Gemini command adapters obey the same artifact contract;
  an adapter regression cannot rejoin stderr to the persisted review.
- Atlas/help describe `*-review.md` as final review output and `*-gate.md` as
  structured gate state, not as raw/full transcripts.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.04 impl=0.12
item: cross-cutting-refactor design=0.10 impl=0.20
item: atlas-docs design=0.03 impl=0.08
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 1.09
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only. The calibration source was stale on
2026-08-25, so the primitive values are provisional; the thorough approved plan
uses the calibrated 15% design buffer.

## Plan

- [ ] Split the subprocess seam into semantic stdout and diagnostic stderr,
      preserving launch/non-zero-exit behavior and terminal diagnostics.
- [ ] Persist and parse only semantic stdout through the shared boundary-review
      path.
- [ ] Add fake-process regression coverage for all agent adapters plus the
      synchronous and heartbeat dispatch paths.
- [ ] Update the sidecar contract in atlas/help and verify the complete sdlc
      suite.

## Log

### 2026-08-25

- Scope narrowed with the operator: fix artifact semantics first; file reviewer
  checkout isolation separately as #204; then return to pair#146/couch.
