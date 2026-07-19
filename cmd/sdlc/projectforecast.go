// projectforecast.go — the IO seams that feed the pure calendar forecast core
// (#182 M2). ListFleetProjects assembles each active fleet project's contention
// load (reusing project.ListActiveProjectFiles for the walk + computeBoard for
// remaining hours); loadThroughputBaseline reads the blessed baseline;
// forecastForProject wires this project + the fleet + the baseline into
// project.ComputeForecast. The pure core (forecast.go) never touches the
// filesystem — all of that lives here.
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

// projectFileName is the slug fallback name when metadata is unavailable.
func projectFileName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".md")
}

// projectRepoDir derives a project file's owning repo root from its canonical
// <repo>/workshop/projects/<name>.md location (the issue-lookup vantage).
func projectRepoDir(path string) string {
	return filepath.Dir(filepath.Dir(filepath.Dir(path)))
}

// errNoBaseline is the single fallback signal: no blessed throughput baseline
// exists, or it can't be read/parsed. Every consumer maps it to its own
// fallback (commit → legacy --reality; show/status → a quiet hint line).
var errNoBaseline = errors.New("no throughput baseline")

// loadThroughputBaseline returns the current (last) blessed baseline. Absence
// AND read/parse failure both map to errNoBaseline (spec Errors: "ledger
// unreadable → same fallback as no-baseline") — consumers never distinguish.
// WF_THROUGHPUT_BASELINE overrides the brain path (test seam).
func loadThroughputBaseline(brainDir string) (estimate.ThroughputBaseline, error) {
	text, err := os.ReadFile(throughputBaselinePath(brainDir))
	if err != nil {
		return estimate.ThroughputBaseline{}, errNoBaseline
	}
	rows, err := estimate.ParseBaselineTSV(string(text))
	if err != nil || len(rows) == 0 {
		return estimate.ThroughputBaseline{}, errNoBaseline
	}
	return rows[len(rows)-1], nil
}

// ListFleetProjects builds the contention loads for every ACTIVE project in the
// fleet except excludePath (the subject). Each load's remaining hours come from
// the same computeBoard the status board uses, resolved from that project's own
// repo vantage; a project whose breakdown doesn't resolve falls back to its
// Phase-A estimate, and one with neither is `unknown` (weight 0 + warning) —
// never silently dropped, since a silent drop reads as "no contention".
func ListFleetProjects(parentDir, excludePath string) []projectdoc.ProjectLoad {
	files, err := projectdoc.ListActiveProjectFiles(parentDir, excludePath)
	if err != nil {
		// A fleet-walk failure (e.g. unreadable parent) degrades to a solo
		// forecast; surface it rather than silently reading as "no contention".
		cwarn(os.Stderr, "fleet project walk failed: "+err.Error()+" — forecasting without cross-project contention")
		return nil
	}
	var loads []projectdoc.ProjectLoad
	for _, f := range files {
		load := projectdoc.ProjectLoad{Repo: f.Repo}
		d, derr := readProject(f.Path)
		if derr != nil {
			load.Name = projectFileName(f.Path)
			load.RemainingSource = "unknown"
			load.Warning = "unreadable/unparsable project file: " + derr.Error()
			loads = append(loads, load)
			continue
		}
		loads = append(loads, projectLoadFromDoc(d, f))
	}
	return loads
}

// projectLoadFromDoc turns one project doc into a contention load: board
// remaining if the breakdown resolves, else Phase-A, else unknown.
func projectLoadFromDoc(d *projectdoc.Doc, f projectdoc.ProjectFile) projectdoc.ProjectLoad {
	meta, err := d.Metadata()
	name := projectFileName(f.Path)
	if err == nil && meta.Name != "" {
		name = meta.Name
	}
	load := projectdoc.ProjectLoad{Name: name, Repo: f.Repo}
	if err != nil {
		load.Status, load.RemainingSource, load.Warning = "unknown", "unknown", "metadata error: "+err.Error()
		return load
	}
	load.Status = meta.Status

	// The board is authoritative once the breakdown resolves into issue rows —
	// even at 0 remaining (a fully-done project is maximally mature and reads
	// ~0, NOT the coarse Phase-A number). Phase-A is a fallback only for a
	// project whose breakdown hasn't resolved into any issues yet.
	b, berr := computeBoard(d, func(ref string) (issueMeta, error) { return projectIssueLookupFn(ref, f.RepoDir) })
	if berr == nil && boardRowsResolved(b) {
		load.RemainingHours, load.RemainingSource = b.RemainingHours, "board"
		return load
	}
	if phaseA, present, perr := projectdoc.ParsePhaseA(d.SectionBody("Estimate")); present && perr == nil && phaseA > 0 {
		load.RemainingHours, load.RemainingSource = phaseA, "phase-a"
		return load
	}
	load.RemainingSource = "unknown"
	load.Warning = "no resolvable breakdown rows and no **phase-a:** estimate"
	return load
}

// boardRowsResolved reports whether the breakdown produced at least one row
// that resolved to a real issue (status != "unresolved"). An empty or
// all-unresolvable breakdown has NOT resolved, so the caller falls back to
// Phase-A rather than reporting a spurious 0.
func boardRowsResolved(b board) bool {
	for _, row := range b.Rows {
		if row.RefText != "" && row.IssueStatus != "unresolved" {
			return true
		}
	}
	return false
}

// forecastForProject is the shared assembly every consumer calls: build this
// project's load + the fleet's other loads + the baseline, then ComputeForecast.
// Returns errNoBaseline (bubbled) so each consumer picks its own fallback.
func forecastForProject(d *projectdoc.Doc, projectPath, parentDir, brainDir, today string) (projectdoc.Forecast, string, error) {
	baseline, err := loadThroughputBaseline(brainDir)
	if err != nil {
		return projectdoc.Forecast{}, "", err
	}
	meta, merr := d.Metadata()
	if merr != nil {
		return projectdoc.Forecast{}, "", merr
	}
	// Repo is the repo basename fleet-wide (consistent with sibling loads); the
	// project's own name lives in ProjectLoad.Name via projectLoadFromDoc.
	repoDir := projectRepoDir(projectPath)
	this := projectLoadFromDoc(d, projectdoc.ProjectFile{Path: projectPath, RepoDir: repoDir, Repo: filepath.Base(repoDir)})
	others := ListFleetProjects(parentDir, projectPath)
	f, cerr := projectdoc.ComputeForecast(baseline, this, others, today)
	if cerr != nil {
		return projectdoc.Forecast{}, meta.Deadline, cerr
	}
	return f, meta.Deadline, nil
}
