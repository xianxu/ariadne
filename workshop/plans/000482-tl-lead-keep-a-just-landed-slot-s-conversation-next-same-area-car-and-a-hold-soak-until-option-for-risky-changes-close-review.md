# Boundary Review — ariadne#482 (whole-issue close)

| field | value |
|-------|-------|
| issue | 482 — tl-lead: keep a just-landed slot's conversation (next same-area car), and a 'hold: soak until' option for risky changes |
| repo | ariadne |
| issue file | workshop/issues/000482-tl-lead-keep-a-just-landed-slot-s-conversation-next-same-area-car-and-a-hold-soak-until-option-for-risky-changes.md |
| boundary | whole-issue close |
| milestone | — |
| window | 718afe6dbad3fb6db1b71981e6402a6a0bce8c92..885657462e486119f46f41619a0b894f5c8af136 |
| command | sdlc close --issue 482 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T23:29:06-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking whether the `## Slots` renderer keeps a hand-added hold note.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The diff is small and uses only prose, and most of it is correct. Two rules were added to `cmd/tl-lead/help.md`. The DISPATCH rule keeps a just-landed slot's conversation (help.md:362-370). The LANDING rule adds a "hold: soak until X" with three exit conditions (help.md:420-425). The help test pins both rules, and `go test ./cmd/tl-lead/ -run TestLeadHelp` passes. Spec point 1 asked whether `couch --prepare` can skip the reload. The Log answers it, and the answer holds: the help text already treats `--relaunch` and `--reload-context` as distinct restarts (help.md:400). So no pair issue was needed, and the second Done-when clause is met by its "else" branch. The one gap is in the hold rule. It tells the TL to write the hold into `## Slots`, but the binary renders that whole section on every pass. The renderer drops any hand-written note, and it lists the held slot as **free**. The spec says "The hold shows in the status file's Slots line", and that part is not actually delivered.

**1. Strengths**
- The DISPATCH rule says why prepare is skipped and names the two exceptions: a fresh start for unrelated or nearly full context, and a relaunch for a Pair binary change (help.md:365-370).
- The soak hold gives three concrete exit conditions and lets the TL open an extra slot to keep throughput (help.md:422-425), which matches the Spec.
- The Spec 1 open question (couch --prepare) was settled with evidence in the Log, not left as a vague follow-up.
- The new test rows follow the existing `TestLeadHelpCoversPlaybook` table pattern (main_test.go:56-60).

**2. Critical findings**
None.

**3. Important findings**
- **The hold marker in `## Slots` cannot survive the renderer, and the held slot shows as free** (help.md:424; family `hand-state-in-rendered-section`).
  - `tl-lead status slots FILE` rebuilds the whole section from `gatherSlots` and `renderSlots` (slots.go:61-117, status.go:320). The help says to re-render on every pass (help.md:266-274).
  - A slot that landed and went idle has `HostBranch == RestingBranch`, so `renderSlots` puts it on the `free:` line (slots.go:71-72). A hand-added `hold: soak until …` is overwritten on the next render.
  - The family in this window has two instances:
    - (a) help.md:424 tells the TL to edit a section that is rendered, not written by hand.
    - (b) help.md:397's "Free for new work means idle, on its resting branch, clean, and with no claim in progress" was not amended. A held slot meets that definition and can be dispatched.
  - Fix sketch, cheapest option: put the hold under a hand-written section (Threads or Needs you), and add "and not held for a soak" to the free definition at help.md:397.
  - Fuller fix: have `renderSlots` read a hold marker (for example, carry it through from the prior section, or keep a small holds record) so it renders `held: soak until X` instead of `free`. Then pin it with a slots test.

**4. Minor findings**
- `atlas/workflow/tl-skills.md` has no entry for the keep-conversation or soak-hold rules. Other help.md rules get a one-line atlas pointer (for example #433 lifecycle notices), so a one-liner would keep this consistent.
- The "soak exits" test row pins the text across hard line wraps (`the\n    next train lands…`), so any reflow breaks it. Existing rows do the same thing, so this is only a note.

**5. Test coverage notes**
The help tests check that the phrases exist, which is the right check for prose rules. Nothing tests the Slots-line display of a hold, because no code supports one. That is the gap behind the Important finding.

**6. Architectural notes**
- ARCH-DRY: pass. The new rules reuse the existing train rule and the restart guidance (help.md:400) instead of restating them.
- ARCH-PURE: pass. No code changed.
- ARCH-PURPOSE: flag. The Spec's "hold shows in the Slots line" is asked for as prose in a section a binary owns. This is the easy subset, not a delivered display. See the Important finding.
- ARCH-EVOLVE: pass. No persisted format changed. If a later fix adds a hold record or marker, it must be additive and must survive slots re-renders.

**7. Plan revision recommendations**
- Add a "## Revisions" entry. Either: "the soak hold is recorded in <hand-written section>, because `## Slots` is fully rendered by `tl-lead status slots`; help.md's free-slot definition excludes held slots." Or: add a Plan item for a renderer-supported hold marker, with a slots test.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: hand-state-in-rendered-section
    title: |
      Soak hold is to be written in the binary-rendered Slots section, which overwrites it and lists the held slot as free
    detail: |
      help.md:424 says to show "hold: soak until X" on the slot's line in the Slots section, but tl-lead status slots re-renders that whole section every pass (slots.go:61-117, status.go:320, help.md:266-274). An idle slot on its resting branch goes on the free: line (slots.go:71-72), so the hand-written note is lost and the Spec clause "the hold shows in the status file's Slots line" is not delivered. Second instance of the same family: help.md:397's free-for-new-work definition was not updated to exclude held slots, so a held slot can still be dispatched. Fix: record the hold in a hand-written section and amend help.md:397, or teach renderSlots a hold marker and pin it with a slots test.
  - id: new
    severity: Minor
    kind: defect
    family: atlas-pointer-for-help-rules
    title: |
      No atlas pointer for the new keep-conversation and soak-hold tl-lead rules
    detail: |
      atlas/workflow/tl-skills.md points to other help.md rules (for example the 433 lifecycle notices) but was not updated for these two new rules.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"01315642-de00-4f87-b471-23668c374b18","repository":"github.com/xianxu/ariadne","issue_id":"000482","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b3516f20cd8e94c2","reviewed_head":"885657462e486119f46f41619a0b894f5c8af136","slot":"ariadne:4","requested_at":"2026-10-20T06:28:14.624786Z"},"phase":"finished","pid":25939,"started_at":"2026-10-20T06:28:14.760979Z","finished_at":"2026-10-20T06:29:05.804689Z","session_id":"b6973a74-5ca9-49ce-ba8d-467e21a1a142","outcome":"returned"}] -->
