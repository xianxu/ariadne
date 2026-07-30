---
type: prose
issue: ariadne#190
task: 5
created: 2026-07-29
---

# Regression check: the 46.1 minutes return to #187

The one verification whose correct answer was known before the fix was written. ariadne#187's
close recorded, in its own `## Log`, that **46.1 minutes were attributed to `#127`** by mention
fallback. `#127` in this repo is `000127-recalibrate-estimate-logic-v2-high.md` — archived,
unrelated, and about *recalibrating estimates*. The minutes belonged to #187, whose Task 14
replayed **pair**#127.

So the fix has a falsifiable target: those minutes must come back, and `#127` must vanish from
the attribution set entirely.

## Before

Captured live from `sdlc actual --issue 187` **before** Task 1 landed (and recorded in #187's
own close Log, independently, at close time):

```
[ok] measured actual for #187: 2.83h   (window f59f49cb → HEAD)
  #185 12.5m/74% mention fallback without issue commit boundary (2026-07-28 21:48 → 22:04)
  #105  2.7m/100% mention fallback without issue commit boundary (2026-07-28 21:48 → 22:04)
  #187 84.5m/50% mention fallback without issue commit boundary (2026-07-29 10:25 → 12:36)
  #127 46.1m/77% dominant long attribution segment              (2026-07-29 10:25 → 12:36)
  #127 46.1m/77% mention fallback without issue commit boundary (2026-07-29 10:25 → 12:36)
(attributed across window issues: #105, #127, #129, #185, #187, #188)
```

Note `#129` in the window set too — that is **pair**#129, from
`c1f7d68`-era commit prose. A second foreign issue silently holding a claim on this repo's time.

## After

Same command, same window, after Tasks 1–4:

```
[ok] measured actual for #187: 3.83h   (window f59f49cb → HEAD)
  #185 12.5m/74% mention fallback without issue commit boundary (2026-07-28 21:48 → 22:04)
  #105  2.7m/100% mention fallback without issue commit boundary (2026-07-28 21:48 → 22:04)
  #187 130.6m/57% dominant long attribution segment             (2026-07-29 10:25 → 12:36)
  #187 130.6m/57% mention fallback without issue commit boundary (2026-07-29 10:25 → 12:36)
  pair#127 foreign ref ignored — another repo's issue, not attributable here (×2)
  pair#129 foreign ref ignored — another repo's issue, not attributable here (×1)
(attributed across window issues: #105, #185, #187, #188)
```

## The three assertions

| assertion | result |
|---|---|
| `#127` receives **0** minutes | ✅ absent from the attribution set entirely — not zeroed, *gone* |
| `#187` gains what `#127` held | ✅ **84.5m → 130.6m = +46.1m**, exactly the disputed quantity |
| no warning names `#127` | ✅ both `#127` rows gone; replaced by named `pair#127` / `pair#129` foreign-ref lines |

The +46.1m is exact, not approximate. That is the strongest form this check could take: the
number that moved is the same number the bug had misplaced, to the tenth of a minute.

**`#129` was a bonus catch.** The plan predicted one misattribution; the measurement found two.
`pair#129` had also been sitting in the tracked-issue set, eligible to claim mention share on
any segment whose prose named it.

## The fixed-window run (the plan's stated command)

The comparison above uses `sdlc actual`, whose window ends at `HEAD`. The plan also specified a
FIXED-window `sdlc active-time` invocation, so the primary evidence rests on boundaries that
cannot drift. Run after Tasks 1–4:

```
$ sdlc active-time \
    --dir ~/.claude/projects/-Users-xianxu-workspace-ariadne \
    --git-repo /Users/xianxu/workspace/ariadne \
    --issue 187 --issue 127 \
    --since 2026-07-29T10:00:00-07:00 --until 2026-07-29T13:00:00-07:00 \
    --threshold-min 15 --include-assistant

# events in window: 462  •  commits in window: 14

  #  start             end                 min  commit       issues  mentions    alloc
  1  2026-07-29 10:25  2026-07-29 12:36  130.6  (no anchor)          #187=11     #187=130.6m

# per-issue totals
  #187: 2.18 hr  (130.6 min)

  attribution warning: #187 130.6m/100% dominant long attribution segment
  attribution warning: #187 130.6m/100% mention fallback without issue commit boundary
  attribution warning: pair#129 foreign ref ignored — another repo's issue, not attributable here (×1)
```

Three things this shows that the `sdlc actual` comparison cannot:

- **`mentions #187=11` and nothing else.** Before the fix this same prose yielded mentions for
  BOTH `187` and `127`, and the segment's 130.6m was split 84.5/46.1 between them. The mention
  column now shows the foreign ref was never counted.
- **`#127` is absent from `per-issue totals` entirely** — not a zero row. It is no longer a
  participant.
- **`#187: 130.6 min` is the whole segment**, over boundaries that cannot drift.

**A practical limit of the warning, visible right here.** `pair#129` is named but `pair#127` is
not — because `foreignRefWarnings` scans COMMIT SUBJECTS, and the `pair#127` commits landed after
13:00, outside this window. The `pair#127` *mentions* inside the window were correctly excluded
but silently. So the observability improvement is real yet partial: it names foreign refs that
appear in commit subjects, not those appearing only in transcript prose. Stated in
`foreignRefWarnings`' doc, and repeated here because this run is the concrete demonstration.

## Corroboration, and why it is not the primary measurement

`sdlc actual --issue 187` works on the archived issue file (verified — `computeActual` resolved
it), but its window is `<first #187 commit> → HEAD`, so the total drifts upward as unrelated
work lands on the branch. Across this session it read 2.29h → 2.32h → 2.83h → 3.83h.

Only the **last** step of that sequence is this fix; the earlier drift is the moving `HEAD`.
The trustworthy comparison is therefore the per-segment figure — `#187`'s
`2026-07-29 10:25 → 12:36` segment going 84.5m → 130.6m while its window boundaries stayed
identical — not the headline total. Stated because quoting "2.83 → 3.83h" as the fix's effect
would overclaim by including drift the fix had nothing to do with.

## Ledger consequence — recorded, deliberately NOT corrected

ariadne#187's calibration row reads `actual 2.32 / ratio 3.6×`. The measurement above shows
that actual was too low: ~46 minutes of its work had been charged elsewhere.

**The row is not being rewritten.** #117's calibration-integrity rule and #178's
measured-not-typed gate both exist to keep that history honest, and hand-editing a historical
actual is exactly the move they forbid. What is recorded instead:

- #187's true actual is **higher** than 2.32h, so its 3.6× over-estimate ratio is **too
  generous** — the real over-estimate was smaller.
- Every calibration row measured before this fix may be low in the same direction, by however
  much cross-repo prose appeared in its window. Rows for issues that never mentioned another
  repo are unaffected.
- Whether to re-measure historical rows is a separate decision with its own trade-off (a
  re-measured row is no longer the number that was true at close). Not taken here.

Worth naming the recursion: the corrupted data feeds velocity calibration, and the issue that
absorbed the minutes — ariadne#127 — is the issue that tracks *recalibrating the estimate
model*. The bug was quietly poisoning the well it was drawing from.

## What the fix does NOT claim

- **Cross-repo time is not attributed, only named.** `pair#127` gets a warning line, not a row
  with minutes. Attributing across repos remains open (`Ref.Qualifier` is retained for exactly
  that), and is deliberately out of scope.
- **Transcript-side foreign refs are not counted in the warning.** `foreignRefWarnings` scans
  commit subjects, which `Commit` retains; event text is not kept past mention extraction. The
  exclusion is still *correct* on the mention path (that is Task 3's `mentionScope`) — it is the
  *count* in the warning that comes only from subjects.
