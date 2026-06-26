package estimate

import (
	"fmt"
	"path/filepath"
	"strings"
)

// source.go is the estimator-SOURCE pointer (#134): the discovery surface that
// names BOTH halves of estimate derivation so an agent never satisfies the block
// grammar while estimating per-primitive hours from memory.
//
//   - the SHARED METHOD is single-sourced in sdlc — the `## Estimate` grammar +
//     closed vocabulary (helptext/estimate.md, `Models()`/`Primitives()` here);
//   - the REPO-LOCAL CALIBRATION (the actual per-primitive hour ranges, drifting
//     as closes accrue) lives in a brain artifact, resolved by SourcePath.
//
// All pure (path strings + rendering). The stat/mtime that decides Exists/Stale
// is the thin IO seam in package main (estimatesource.go), which fills a
// SourceStatus and hands it to SourceGuidance.

// VelocityPath is the single source for a file under the brain velocity dir:
// <brainDir>/data/life/42shots/velocity/<filename>. Both the calibration-ledger
// (close.go) and the estimate-logic doc (SourcePath) resolve through it, so the
// brain-relative tail lives in exactly one place (ARCH-DRY).
func VelocityPath(brainDir, filename string) string {
	return filepath.Join(brainDir, "data", "life", "42shots", "velocity", filename)
}

// SourcePath resolves the repo-local calibration doc for model: the
// $WF_ESTIMATOR_SRC override (passed as `override`) wins; else the brain velocity
// doc named after the model (e.g. estimate-logic-v2 → estimate-logic-v2.md).
func SourcePath(brainDir, model, override string) string {
	if override != "" {
		return override
	}
	return VelocityPath(brainDir, model+".md")
}

// SourceStatus is what the IO seam resolves about the calibration source: where
// it is, which model it backs, whether it exists, and whether it looks stale (the
// calibration ledger is newer than the doc — new closes since the last recalibration).
type SourceStatus struct {
	Path   string
	Model  string
	Exists bool
	Stale  bool
}

// SourceGuidance renders the one discoverable block naming both sources +
// the calibration source's status. Pure. The Missing case is a LOUD next-action
// (never an empty string), satisfying #134's "fail loudly, don't let agents rely
// on memory."
func SourceGuidance(s SourceStatus) string {
	var b strings.Builder
	b.WriteString("ESTIMATE DERIVATION — two sources, both required:\n\n")

	b.WriteString("  Shared method (single-sourced in sdlc):\n")
	b.WriteString("    • grammar + closed vocabulary:  `sdlc change-code --help`  (helptext/estimate.md)\n")
	fmt.Fprintf(&b, "    • recognized models:            %s\n\n", strings.Join(Models(), ", "))

	b.WriteString("  Repo-local calibration (the per-primitive hours you derive against):\n")
	fmt.Fprintf(&b, "    • %s  [%s]\n\n", s.Path, statusTag(s))

	b.WriteString("  " + nextAction(s) + "\n")
	return b.String()
}

func statusTag(s SourceStatus) string {
	switch {
	case !s.Exists:
		return "MISSING"
	case s.Stale:
		return "stale"
	default:
		return "ok"
	}
}

// nextAction is the status-specific imperative — always present so the surface
// never silently lets an agent fall back to memory.
func nextAction(s SourceStatus) string {
	switch {
	case !s.Exists:
		return fmt.Sprintf("MISSING: %s not found. Set $WF_ESTIMATOR_SRC to the calibration doc "+
			"(or pass --brain-dir); do NOT estimate per-primitive hours from memory.", s.Path)
	case s.Stale:
		return "stale: the calibration ledger is newer than this doc — the per-primitive " +
			"hours may have drifted (recalibration tracked in #127). Derive against it, " +
			"but treat the numbers as provisional."
	default:
		return "ok: read it to pick per-primitive design/impl hours, then itemize them in the " +
			"`## Estimate` block."
	}
}
