# Boundary Review — ariadne#172 (milestone M3)

| field | value |
|-------|-------|
| issue | 172 — sdlc painpoint audit |
| repo | ariadne |
| issue file | workshop/issues/000172-sdlc-painpoint-audit.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 9c129f95499b26f444581f5937b65cae94f7ea92..HEAD |
| command | sdlc milestone-close --issue 172 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-07-14T14:13:21-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All verification is done. Here is the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3 delivers what it claims, and I could verify most claims independently: the Go codex reader derives faithfully from the atlas spec (`atlas/workflow/introspect.md` §"Codex transcript format" — first-`session_meta` keying, fork-replay skip vs sub-agent keep, JSON-encoded-string `arguments.cmd`, plain-string output), `go test ./cmd/sdlc/...` and `go vet` are green, and a live end-to-end run of `--friction-report --json` over the real corpus reproduces every headline number in the Log: codex 43 vs claude 37 bypasses, **exactly 40** fork-replays skipped (the spec's skip-40-not-119 contract), no-reclose-guard 25/3, no-actual 38 refusals, 18 firing-order anomalies, 52 unattributed publishes. The M2-review deferred items (Failed flag, single `allGateEvents` classification, `forEachRec` extraction) all landed with correct semantics — including the subtle REWORK-before-failure-skip ordering. Nothing blocks the boundary; the findings below are cheap hardening.

## 1. Strengths

- **Spec-as-DRY-point honored, not just cited.** `codex.go`'s format knowledge (fork trap, first-meta keying, double-encoded `arguments`) matches the atlas spec line-for-line, and the deliberate divergence from the spec's hint-gated `is_error` is explicitly documented at `codex.go:83-88` rather than silently drifting — exactly the right way to disagree with a spec section that answers a different question.
- **The classifier reuse is real and tested end-to-end**: `codex_test.go:97-99` asserts a codex-wrapped output flows through the same `classifyOutputLine` into a `no-atlas` bypass, and the golden's `root.jsonl` third call plants cat-n contamination inside codex output and asserts it classifies to none — the anti-contamination property carried across agents, not re-argued.
- **Fork-skip is counted, not hidden** (`CodexForksSkipped` in JSON + a rendered footnote), and the live corpus yielding exactly the spec's 40 is strong validation the meta-keying is right.
- **`detectFiringOrder` index-alignment is correct**: `seqs[k]` carries original invocation indices, so `events[i]` at `friction.go:514` lines up; the REWORK-rollback-before-failure-skip reorder (`friction.go:504-521`) handles the "REWORK close also reads as failed" trap and says so.
- **Plan Revisions entry is honest and complete** — all five deviations recorded at the boundary with rationale, matching what the code actually does.

## 2. Critical findings

None.

## 3. Important findings

1. **Plan Task 9 Step 1's walk-level test was not delivered as specified — the two-agent walk path is only exercised by live smoke** (`friction_test.go:596` is the sole `RunFrictionReport` test, and it's the zero-transcripts error case). The plan promised "walk over 1 Claude + 1 codex transcript each with a `--no-judge` bypass → both counted, agent-tagged"; what shipped is `TestAggregatePerAgent` (aggregate-level, pre-built invocations) plus parser-level goldens. The enumerate → `codexMeta` → `repoLabelFromPath` → tag → merge chain in `RunFrictionReport` (`friction.go:1004-1058`), including the documented "missing corpus on ONE side is fine" contract, has no test. Fix sketch: one `t.TempDir()` test building `claude/<slug>/x.jsonl` + `codex/2026/07/01/rollout-x.jsonl`, assert both bypasses agent-tagged in the JSON; a second call with only one root present asserts no error. ~25 lines, and it's exactly the class of wiring bug this diff could ship.
2. **The cross-language golden has no Python consumer — the Python side still hand-maintains its own fork fixture** (ARCH-DRY). Nothing under `construct/local/introspect/scripts/` reads `testdata/codex-golden/`; `test_normalize.py:125` tests fork-skip against its own inline fixture. Both sides derive from the spec, but the golden's stated purpose ("cross-language drift here is the exact failure the spec-level DRY exists to prevent", `expected.json`) is only enforced on the Go side. The plan's "no live python3" scoped the *Go test* correctly, but the shared-fixture half is a deferred consumer recorded only in passive voice (Revisions #4: "a Python-side test *can* consume"). Cheap fix: point a `test_normalize.py` case at the golden dir for the keep/skip decision (the only genuinely shared decision — bypass/refusal classification is Go-only), or track it as explicit M4/follow-up work rather than an untracked possibility.

## 4. Minor findings

- gofmt: the diff adds drift to already-dirty files — the `Failed` comment block splits `SdlcInvocation`'s alignment group (`friction.go`, 16→35 gofmt-diff lines) and `IsError` re-aligns the `rec` struct comments (`session.go`, 14→16). All four flagged files were dirty at base, so this is tolerated, but the pile is growing; one `gofmt -w` on the package would zero it.
- `repoLabelFromPath` (`friction.go:663`) excludes `/private/var/folders` but not the unresolved `/var/folders/…` form of macOS TMPDIR; a codex cwd recorded without the `/private` prefix would leak in as repo "folders"-something.
- `parseCodexInvocations` doesn't check `payload.name == "exec_command"` — any `function_call` whose arguments carry a matching `cmd` string counts. No current codex tool collides, but the check is one line.
- Scanner setup (64KB/16MB buffer) is now written three times (`forEachRec`, `codexMeta`, `parseCodexInvocations`); a shared line-iterator or at least a shared buffer-size constant would stop the third copy (ARCH-DRY, small).
- `scanTranscript` discards `forEachRec`'s error return — a >16MB line silently truncates the rest of that transcript. Pre-existing semantics, but the new signature makes the discard visible; worth a deliberate `_ =` or comment.
- `RefusalRetry` carries no `Agent` field — the Log's "codex's 3 refusals resolved 3/3 via bypass" claim had to be inferred rather than read from the JSON; M4's triage will want the split.
- `transcripts_scanned` asymmetry: Claude counts enumerated files (unreadable ones included via `len(refs)`), codex counts only processed-and-included files.

## 5. Test coverage notes

Parser-level coverage is strong: meta-kind table test (including fork-detection off the *first* meta with the parent's meta later in the file), golden keep/skip + classification + failed counts with contamination planted, the failed-invocation ladder regression, refusal-classified-through-wrapper-exit-71, and `repoLabelFromPath`/`enumerateCodexTranscripts` table tests. The gaps are at the seam, not the core: no walk-level integration test (Important #1), no one-sided-corpus test, and `codexOutputFailed`'s first-match assumption (wrapper exit line precedes `Output:`, so a body mentioning "Process exited with code 3" can't false-positive) is relied on but not pinned by a test.

## 6. Architectural notes (ARCH-* pass/flag)

- **ARCH-DRY — pass with one flag.** The cross-language DRY point (the atlas spec) is genuinely honored: no format re-discovery, and the fixture comments cite the spec rather than restating it. Flag: the golden's one-sidedness (Important #2) leaves the Python fork fixture as a parallel hand-maintained artifact; also the small scanner-setup triplication (Minor).
- **ARCH-PURE — pass.** `codexMeta`/`parseCodexInvocations`/`repoLabelFromPath` are pure over bytes/strings and tested without IO or mocks (the golden reads testdata, which is conventional); `RunFrictionReport` remains the single thin IO seam and the codex loop stays enumeration + ReadFile + delegate. The `forEachRec` extraction actively improved this.
- **ARCH-PURPOSE — pass.** The issue's Spec demands *both agents* precisely because Claude-only gave a biased picture, and the diff delivers the full purpose, not a subset: codex ingested, agent-tagged, per-agent split rendered, and the finding it was built for materialized (codex's re-close signature — invisible in M2 — is now the headline). Shadow-sweep of the spec's consumers: Python introspect and Go `codex.go` both derive from the spec; the only residual hand-maintained restatement is the Python fork fixture (Important #2), which is a test artifact, not a purpose gap.

## 7. Plan revision recommendations

The plan's Revisions section already covers the code faithfully (including the `parseCodexEvents` → `parseCodexInvocations` naming mapping for the Core-Concepts row). One addition warranted: extend Revisions #4 (or the M4 scope) to state explicitly whether the Python-side golden consumer is deferred work with an owner (M4 / a follow-up issue in #173's orbit) or intentionally dropped — right now it exists only as "a Python-side test can consume," which is neither tracked nor declined. If Important #1's walk-level test is added, no plan change is needed for it; if it's declined, Task 9 Step 1's checkbox should get a Revisions note that the walk test was downgraded to aggregate-level + live smoke.

---

**Ops note for the operator:** during verification, a review command of mine misfired — `mktemp -d` failed under the sandbox and a `git archive | tar -x` intended for a temp dir extracted three base-commit files (`friction.go`, `friction_test.go`, `session.go`) over the working tree. I restored them immediately with `git restore`; `git status` is clean and HEAD (`6ff7024`) is unchanged. No lasting effect, but flagging it for transparency.
