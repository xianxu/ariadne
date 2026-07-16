# Boundary Review Sidecar Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist the full boundary-review transcript from `sdlc close` and `sdlc milestone-close` to a durable sidecar file under `workshop/plans/`, so an agent can reopen the review after scrollback loss or context compaction.

**Architecture:** Add a pure sidecar module (`reviewsidecar.go`) that renders a self-contained review document from structured metadata (`renderReviewEntry`) and computes its deterministic path (`sidecarPath`) — both unit-tested with zero IO (ARCH-PURE). A thin IO writer (`writeReviewSidecar`) gathers the metadata (issue title, repo identity, clock) and atomically writes/append-revises the file. The write is wired into the **single shared** `dispatchBoundaryReview()` choke point that both close and milestone-close already funnel through, so neither path duplicates persistence logic (ARCH-DRY). The change is strictly additive: trailers, log annotation, verdict parsing, and gate behavior are untouched (ARCH-PURPOSE — deliver durability without regressing the existing review contract).

**Tech Stack:** Go; pure unit tests colocated in `cmd/sdlc/reviewsidecar_test.go`; integration via the existing `judge.Run` test seam; SDLC atlas docs.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `sidecarMeta` | `cmd/sdlc/reviewsidecar.go` | new |
| `sidecarPath` | `cmd/sdlc/reviewsidecar.go` | new |
| `renderReviewEntry` | `cmd/sdlc/reviewsidecar.go` | new |

- **sidecarMeta** — value struct holding everything a fresh reader needs: `IssueNum int`, `Title string`, `Repo string`, `IssueFile string`, `Milestone string` (`""` ⇒ whole-issue close), `Base string` (long SHA), `Head string`, `Command string`, `Agent string`, `Timestamp string` (RFC3339, passed in — the clock is IO), `Verdict string`, `Body string` (the full review output).
  - **Relationships:** 1:1 with a single review run; N:1 with an Issue (an issue accrues a close-review plus one per milestone, and re-runs append within the same file).
  - **DRY rationale:** One struct is the single source of the sidecar's field set; both render paths (initial doc, re-review section) read the same shape, so adding a metadata field never means editing two renderers.
  - **Future extensions:** Add `DiffStat` or `PromptSHA` fields later without touching the writer/dispatch wiring.

- **sidecarPath** — `sidecarPath(plansDir, issueFileName, milestone string) string`. Pure path derivation: strips `.md` from the issue filename stem (`000136-review-sidecar`) and appends `-close-review.md`, or `-m<lower>-review.md` for a milestone (`M2` → `-m2-review.md`).
  - **DRY rationale:** The `NNNNNN-slug` stem is reused from the *issue filename* (which `issueFilePath` already resolves), not re-slugified from the title — one slug source of truth.
  - **Future extensions:** A third boundary kind (e.g. a pre-merge sidecar) is one more suffix branch.

- **renderReviewEntry** — `renderReviewEntry(m sidecarMeta, isRevision bool) string`. Pure markdown render. `isRevision=false` ⇒ `# Boundary Review …` H1 + metadata table + `## Review` + body. `isRevision=true` ⇒ `## Re-review — <ts> (<verdict>)` heading + metadata table + body (no H1), for appending under a `---` separator.
  - **DRY rationale:** One renderer, parameterized by a bool, covers both the create and the re-run cases — no parallel near-identical functions.
  - **Future extensions:** A machine-readable frontmatter block could be prepended here without changing callers.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `writeReviewSidecar` | `cmd/sdlc/reviewsidecar.go` | new | filesystem + git + clock |
| `atomicWriteFile` | `cmd/sdlc/reviewsidecar.go` | new | `os` temp+rename |
| `dispatchBoundaryReview` | `cmd/sdlc/milestoneclose.go:474` | modified | review subprocess |
| `boundaryReviewParams` | `cmd/sdlc/milestoneclose.go:447` | modified | dispatch inputs |
| `reviewResult` | `cmd/sdlc/milestoneclose.go:60` | modified | dispatch outputs |

- **writeReviewSidecar** — `writeReviewSidecar(p boundaryReviewParams, verdict, body, timestamp string) (string, error)`. Resolves the issue file via the existing `issueFilePath(p.IssuesDir, p.IssueNum)`, parses its `# Title`, derives repo identity from `filepath.Base(gitx.RepoTopLevel())`, computes the path via `sidecarPath`, and either writes the initial doc (file absent) or appends a re-review section (file present — **never silently overwrites**). Returns the path written.
  - **Injected into:** Called once from `dispatchBoundaryReview` after `output` is captured. The pure renderers receive a fully-populated `sidecarMeta`, so they stay testable without touching disk.
  - **Future extensions:** Could gain a `--no-sidecar` opt-out flag; today it always persists an actually-run review.

- **atomicWriteFile** — `atomicWriteFile(path string, data []byte) error`: `MkdirAll(dir)`, write `path + ".tmp"`, `os.Rename` into place. Satisfies the spec's "persist … atomically."
  - **Injected into:** `writeReviewSidecar` (both the create and the append→full-rewrite paths).
  - **Future extensions:** First atomic-write helper in `cmd/sdlc`; if a second caller appears, promote to `internal/`.

- **dispatchBoundaryReview** *(modified)* — after `fmt.Fprint(stdout, output)` (milestoneclose.go:493), call `writeReviewSidecar(...)`, print a compact `cok(stderr, "review sidecar: <path>")` line, and store the path on `reviewResult`. A write error is **non-fatal** (the review already ran; mirror the existing not-fatal philosophy at lines 488–491) — log a `cwarn` and continue. The full body stays on stdout (see Design decision D3).

- **boundaryReviewParams** *(modified)* — add `IssueNum int`, `Milestone string`, `PlansDir string`. Both call sites already hold these (close: `f.Issue`, `Milestone: ""`; milestone-close: `f.Issue`, `f.Milestone`); `PlansDir` from `envOr("WF_PLANS_DIR", "workshop/plans")`.

- **reviewResult** *(modified)* — add `SidecarPath string` (empty when no review ran). Lets future callers/tests reference the path; trailer/log logic is unchanged.

**Test surface.** `sidecarPath` + `renderReviewEntry` get pure colocated tests (no IO). `writeReviewSidecar` gets an IO test against `t.TempDir()` with a hand-written issue file (create + re-run-append + metadata completeness). The dispatch wiring is covered by extending the existing `judge.Run`-stub tests (`closereview_test.go`, `milestoneclose_test.go`) to also assert the sidecar file appears — and that trailers/annotation still fire (existing-behavior-intact).

---

## Design decisions

- **D1 — Naming.** Whole-issue close → `workshop/plans/NNNNNN-slug-close-review.md`; milestone `Mx` → `workshop/plans/NNNNNN-slug-mx-review.md` (lowercased). Slug comes from the resolved issue filename stem, so it always matches the issue/plan sibling.
- **D2 — Re-run = append, never overwrite.** Re-running a boundary review appends a timestamped `## Re-review — <ts> (<verdict>)` section under a `---` separator, preserving all prior evidence. This matches AGENTS.md §1's documented artifact-revision convention ("append a `## Revisions` section, don't overwrite"). Chosen over a collision-safe numeric suffix so all reviews for one boundary stay in one openable file.
- **D3 — Keep the full body on stdout; ADD a compact path line.** The change is additive. The motivating problem (issue ## Problem) is the *absence of a durable file*, not stdout verbosity; the agent at the gate still needs the findings inline to act on Critical/Important before crossing. So `dispatchBoundaryReview` keeps printing the body to stdout **and** prints `review sidecar: <path>` (the Done-when "compact verdict + sidecar path" is satisfied by the trailer block + this line). Suppressing the stdout body would regress the in-session gate and violate "existing behavior remains intact" — explicitly out of scope (ARCH-PURPOSE: serve durability, don't degrade the gate).
- **D4 — Only persist actually-run reviews.** The `--no-judge`, `--dry-run`, and not-run/error paths construct `reviewResult` *without* calling `dispatchBoundaryReview`, so they write no sidecar (there is no body to persist; the trailer already records `not-run`). YAGNI — no stub files.
- **D5 — Single write site.** Persistence lives only inside the shared `dispatchBoundaryReview`; close and milestone-close inherit it for free (ARCH-DRY). No write code in `finishBoundaryReview` or the milestone Step-4/5 path.

---

## Chunk 1: Sidecar module + wiring + docs

### Task 1: Pure path derivation

**Files:**
- Create: `cmd/sdlc/reviewsidecar.go`
- Test: `cmd/sdlc/reviewsidecar_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestSidecarPath(t *testing.T) {
	cases := []struct{ name, issueFile, milestone, want string }{
		{"close", "000136-review-sidecar.md", "", "workshop/plans/000136-review-sidecar-close-review.md"},
		{"milestone M2", "000136-review-sidecar.md", "M2", "workshop/plans/000136-review-sidecar-m2-review.md"},
		{"milestone lowercased", "000069-x.md", "M4b", "workshop/plans/000069-x-m4b-review.md"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sidecarPath("workshop/plans", c.issueFile, c.milestone); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run it, confirm it fails to compile** — `go test ./cmd/sdlc/ -run TestSidecarPath` → undefined: `sidecarPath`.

- [ ] **Step 3: Implement `sidecarMeta` + `sidecarPath`**

```go
type sidecarMeta struct {
	IssueNum                          int
	Title, Repo, IssueFile            string
	Milestone                         string // "" ⇒ whole-issue close
	Base, Head, Command, Agent        string
	Timestamp, Verdict, Body          string
}

func sidecarPath(plansDir, issueFileName, milestone string) string {
	stem := strings.TrimSuffix(filepath.Base(issueFileName), ".md")
	suffix := "close-review"
	if milestone != "" {
		suffix = strings.ToLower(milestone) + "-review"
	}
	return filepath.Join(plansDir, stem+"-"+suffix+".md")
}
```

- [ ] **Step 4: Run test → PASS.**

- [ ] **Step 5: Commit** — `git add cmd/sdlc/reviewsidecar*.go && git commit -m "#136: sidecar path derivation (pure)"`

### Task 2: Pure render

**Files:** Modify `cmd/sdlc/reviewsidecar.go`; Test `cmd/sdlc/reviewsidecar_test.go`

- [ ] **Step 1: Write the failing test** — assert every required metadata field + body is present, and the revision heading toggles.

```go
func TestRenderReviewEntry(t *testing.T) {
	m := sidecarMeta{
		IssueNum: 136, Title: "sdlc boundary review sidecar", Repo: "ariadne",
		IssueFile: "workshop/issues/000136-review-sidecar.md", Milestone: "",
		Base: "abc1234def", Head: "HEAD", Command: "sdlc close --issue 136",
		Agent: "claude", Timestamp: "2026-06-29T15:40:00-07:00",
		Verdict: "SHIP", Body: "VERDICT: SHIP\n\nLooks good.",
	}
	doc := renderReviewEntry(m, false)
	for _, want := range []string{
		"# Boundary Review", "136", "sdlc boundary review sidecar", "ariadne",
		"workshop/issues/000136-review-sidecar.md", "abc1234def", "sdlc close --issue 136",
		"claude", "2026-06-29T15:40:00-07:00", "SHIP", "Looks good.", "## Review",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("initial doc missing %q\n---\n%s", want, doc)
		}
	}
	rev := renderReviewEntry(m, true)
	if !strings.Contains(rev, "## Re-review") || strings.Contains(rev, "# Boundary Review") {
		t.Errorf("revision section should use ## Re-review and no H1:\n%s", rev)
	}
}
```

- [ ] **Step 2: Run it → fails** (undefined: `renderReviewEntry`).

- [ ] **Step 3: Implement `renderReviewEntry`** — H1 (only when `!isRevision`) or `## Re-review — <ts> (<verdict>)`, a metadata table (issue, repo, issue file, boundary kind = close/milestone, milestone-or-`—`, window `Base..Head`, command, reviewer, timestamp, verdict), then `## Review` + body. Use a `strings.Builder`.

- [ ] **Step 4: Run test → PASS.**

- [ ] **Step 5: Commit** — `#136: render review sidecar entry (pure)`

### Task 3: Atomic write + IO writer (create + re-run append)

**Files:** Modify `cmd/sdlc/reviewsidecar.go`; Test `cmd/sdlc/reviewsidecar_test.go`

- [ ] **Step 1: Write the failing test** — create then re-run; assert both bodies present, no overwrite, file at expected path.

```go
func TestWriteReviewSidecar_CreateThenAppend(t *testing.T) {
	dir := t.TempDir()
	issues := filepath.Join(dir, "issues")
	plans := filepath.Join(dir, "plans")
	os.MkdirAll(issues, 0o755)
	os.WriteFile(filepath.Join(issues, "000136-review-sidecar.md"),
		[]byte("---\nid: 000136\nstatus: working\n---\n# sdlc boundary review sidecar\n## Plan\n"), 0o644)

	p := boundaryReviewParams{IssueNum: 136, Milestone: "", IssuesDir: issues, PlansDir: plans,
		BaseLong: "abc1234", Head: "HEAD", Agent: "claude"}

	path1, err := writeReviewSidecar(p, "SHIP", "first review body", "2026-06-29T15:40:00-07:00")
	if err != nil { t.Fatal(err) }
	if filepath.Base(path1) != "000136-review-sidecar-close-review.md" {
		t.Errorf("unexpected path: %s", path1)
	}
	path2, err := writeReviewSidecar(p, "FIX-THEN-SHIP", "second review body", "2026-06-29T16:00:00-07:00")
	if err != nil { t.Fatal(err) }
	if path2 != path1 { t.Errorf("re-run should target same file") }

	data, _ := os.ReadFile(path1)
	s := string(data)
	for _, want := range []string{"first review body", "second review body", "## Re-review", "sdlc boundary review sidecar"} {
		if !strings.Contains(s, want) { t.Errorf("missing %q after re-run", want) }
	}
	if entries, _ := filepath.Glob(filepath.Join(plans, "*.tmp")); len(entries) != 0 {
		t.Errorf("temp file leaked: %v", entries)
	}
}
```

- [ ] **Step 2: Run it → fails** (undefined: `writeReviewSidecar`, and `boundaryReviewParams` lacks `IssueNum`/`Milestone`/`PlansDir` — add those fields in this step too).

- [ ] **Step 3: Implement `atomicWriteFile` + `writeReviewSidecar`**

```go
func atomicWriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writeReviewSidecar(p boundaryReviewParams, verdict, body, timestamp string) (string, error) {
	issuePath, err := issueFilePath(p.IssuesDir, p.IssueNum)
	if err != nil {
		return "", err
	}
	path := sidecarPath(p.PlansDir, filepath.Base(issuePath), p.Milestone)
	m := sidecarMeta{
		IssueNum: p.IssueNum, Title: issueTitle(issuePath), Repo: repoIdentity(),
		IssueFile: issuePath, Milestone: p.Milestone, Base: p.BaseLong, Head: p.Head,
		Command: sidecarCommand(p.IssueNum, p.Milestone), Agent: p.Agent,
		Timestamp: timestamp, Verdict: verdict, Body: body,
	}
	if _, statErr := os.Stat(path); statErr == nil {
		prior, rerr := os.ReadFile(path)
		if rerr != nil {
			return "", rerr
		}
		return path, atomicWriteFile(path, []byte(string(prior)+"\n\n---\n\n"+renderReviewEntry(m, true)))
	}
	return path, atomicWriteFile(path, []byte(renderReviewEntry(m, false)))
}
```

  Helpers (same file): `issueTitle(path)` reads the file and returns the first `# ` line (fallback `""`); `repoIdentity()` returns `filepath.Base(gitx.RepoTopLevel())` (fallback `""` on error); `sidecarCommand(n, ms)` returns `sdlc close --issue N` or `sdlc milestone-close --issue N --milestone Mx`.

- [ ] **Step 4: Run test → PASS.**

- [ ] **Step 5: Commit** — `#136: writeReviewSidecar — create + re-run append (atomic)`

### Task 4: Wire into the shared dispatch

**Files:** Modify `cmd/sdlc/milestoneclose.go` (`dispatchBoundaryReview` ~474, `reviewResult` ~60, `boundaryReviewParams` ~447 already extended in Task 3); Modify `cmd/sdlc/close.go:754` and `cmd/sdlc/milestoneclose.go:154` call sites to pass `IssueNum`/`Milestone`/`PlansDir`.

- [ ] **Step 1: Add the write to `dispatchBoundaryReview`** — after the stdout body print (line ~496) and verdict parse, before `return`:

```go
verdict := judge.ParseVerdict(output)
// ... existing unknown-verdict warning ...
res := reviewResult{Verdict: verdict, Base: p.Base, Head: p.Head, BaseLong: p.BaseLong}
if path, werr := writeReviewSidecar(p, verdict.String(), output, nowRFC3339()); werr != nil {
	cwarn(stderr, fmt.Sprintf("review sidecar not written: %v", werr))
} else {
	res.SidecarPath = path
	cok(stderr, "review sidecar: "+path)
}
return res
```

  Add `SidecarPath string` to `reviewResult`. `verdict.String()` — confirm `judge.Verdict` stringifies to the SHIP/FIX-THEN-SHIP/REWORK label (per the digest the enum values match the trailer labels); if not, map explicitly. `nowRFC3339()` is a tiny local `func() string { return time.Now().Format(time.RFC3339) }` (the only clock touch — keeps render pure).

- [ ] **Step 2: Set the new params at both call sites.** `close.go:754` block: add `IssueNum: f.Issue, Milestone: "", PlansDir: envOr("WF_PLANS_DIR", "workshop/plans")`. `milestoneclose.go:154` block: add `IssueNum: f.Issue, Milestone: f.Milestone, PlansDir: envOr("WF_PLANS_DIR", "workshop/plans")`.

- [ ] **Step 3: Build** — `go build ./cmd/sdlc/` → OK.

- [ ] **Step 4: Run the full sdlc suite** — `go test ./cmd/sdlc/...`. Existing `closereview_test.go` / `milestoneclose_test.go` must stay green (existing-behavior-intact). They stub `judge.Run`; the new write needs an issue file + plans dir reachable from their temp repo — if a test now errors on `issueFilePath`/write, it confirms the path is exercised; adjust the test's temp layout (don't weaken the assertion).

- [ ] **Step 5: Commit** — `#136: persist boundary review to sidecar at the shared dispatch`

### Task 5: Integration assertion (sidecar written through dispatch)

**Files:** Modify `cmd/sdlc/closereview_test.go` (or `milestoneclose_test.go`), reusing its temp-repo + `judge.Run`-stub setup.

- [ ] **Step 1: Extend the dispatch test** — after the stubbed review returns `VERDICT: SHIP\n…`, assert the sidecar file exists at `workshop/plans/NNNNNN-slug-close-review.md` (and `-m<x>-review.md` for the milestone test), contains the stubbed body, and that the existing trailer + log-line assertions still pass.

- [ ] **Step 2: Run → PASS.**

- [ ] **Step 3: Commit** — `#136: test — dispatch writes the review sidecar`

### Task 6: Docs

**Files:** Modify `atlas/workflow/sdlc-binary.md` (and check for a review-specific atlas page); keep `atlas/index.md` linkage intact.

- [ ] **Step 1: Document the sidecar** — under the close/milestone-close / boundary-review section, add: location (`workshop/plans/`), the `-close-review.md` / `-m<x>-review.md` naming, the re-run = append-revision semantics, and "agents reopen this after scrollback loss / compaction." One compact subsection.

- [ ] **Step 2: Commit** — `#136: atlas — document the boundary-review sidecar`

---

## Done-when mapping

| Issue Done-when | Delivered by |
|---|---|
| Persisted under `workshop/plans/` for milestone + issue close | Task 4 (shared dispatch) |
| Naming covers both boundaries + re-run without silent overwrite | Tasks 1, 3 (D1, D2) |
| Terminal prints verdict + sidecar path compactly | Task 4 (`cok` path line + existing trailer) |
| Tests cover path naming, metadata, preserved body, existing verdict/trailer | Tasks 1–3, 5 |
| Help/atlas tells agents where to find the sidecar | Task 6 |

## Non-goals / observations

- Not persisting `--no-judge`/`--dry-run`/not-run boundaries (D4) — no body exists.
- Not suppressing the stdout review body (D3).
- Archival: sidecars live in `workshop/plans/` and match the `NNNNNN-` id prefix, so whatever sweeps plans/issues to `workshop/history/` at merge carries them along — no archive-logic change in scope; verify during close.

## Revisions

- **2026-06-29 — reviewer = resolved agent (close-review FIX-THEN-SHIP follow-up).**
  The boundary review flagged that `sidecarMeta.Agent` was populated from the raw
  `--agent` flag (`p.Agent`), which defaults to `""`, so the `| reviewer |` cell
  rendered empty in the common default invocation even though the spec requires
  "reviewer … if known." Fix: `dispatchBoundaryReview` now sets
  `p.Agent = string(opts.Agent)` (the **resolved** dispatch agent) before
  `writeReviewSidecar`. Pinned by a new assertion in
  `TestRunCloseWithReview_IssueClose_Dispatches` (reviewer cell non-empty) and a
  D4 no-write assertion in `TestRunCloseWithReview_NoJudge_Skips`. The Minor
  `repoIdentity` triplication finding was deliberately **not** taken: `close.go:814`
  uses `RepoTopLevel` for a brain path (not basename) and `branchcreate.go` needs
  both basename and `filepath.Dir(repoTop)`, so a single `gitx.RepoName()` wouldn't
  cleanly unify them (Simplicity-First).
