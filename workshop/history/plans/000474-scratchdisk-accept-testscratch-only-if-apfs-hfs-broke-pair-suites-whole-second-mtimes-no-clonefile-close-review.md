# Boundary Review — ariadne#474 (whole-issue close)

| field | value |
|-------|-------|
| issue | 474 — scratchdisk: accept TestScratch only if APFS (HFS+ broke pair suites: whole-second mtimes, no clonefile) |
| repo | ariadne |
| issue file | workshop/issues/000474-scratchdisk-accept-testscratch-only-if-apfs-hfs-broke-pair-suites-whole-second-mtimes-no-clonefile.md |
| boundary | whole-issue close |
| milestone | — |
| window | 29a710c8c5e0611a35b05a737d1683177d91a0cd..9e2fa3897c0f2ba0c360a41d966097931738e916 |
| command | sdlc close --issue 474 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T15:30:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned diff implements Darwin APFS admission through the shared scratchdisk rule, including overrides. HFS and unknown filesystems fall back with a reason, while rejected roots remain available for checkout cleanup. No blocking defects found.

Done-when: current — the documented platform revision preserves the required APFS acceptance, HFS rejection, fallback note, and APFS guidance.

```findings
{}
```

1. **Strengths**
   - Admission remains centralized in `scratchdisk.Choose` (`scratchdisk.go:32`).
   - `Read` obtains filesystem type through the existing statfs call without capability writes (`scratchdisk.go:69`).
   - Rejected roots remain sweepable (`scratchdisk.go:99`).
   - AGENTS guidance and atlas describe the APFS requirement and override behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Scratchdisk and reviewbox package tests passed. Focused suite caller tests passed using both HFS and APFS temporary roots, exercising fallback and acceptance. The HFS fixture verified one warning and no scratch directory creation. Linux cross-compilation passed; Linux execution was not performed. Two unrelated Seatbelt tests skipped because sandbox application was unavailable.

6. **Architecture**
   - **ARCH-DRY: pass** — suite and review placement share one admission rule.
   - **ARCH-PURE: pass** — admission and root selection remain pure; statfs stays in `Read`.
   - **ARCH-PURPOSE: pass** — all scratchdisk consumers inherit the guard.
   - **ARCH-EVOLVE: pass** — no persisted format changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a6f12f14-e2bb-419f-b34a-ab14fcf2c40c","repository":"github.com/xianxu/ariadne","issue_id":"000474","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/18f955a362c1f746","reviewed_head":"9e2fa3897c0f2ba0c360a41d966097931738e916","slot":"ariadne:3","requested_at":"2026-10-19T22:29:56.951077Z"},"phase":"finished","pid":92208,"started_at":"2026-10-19T22:29:57.086173Z","finished_at":"2026-10-19T22:30:57.142235Z","session_id":"01a15649-284a-7992-9109-86a3ba67e958","outcome":"returned"}] -->
