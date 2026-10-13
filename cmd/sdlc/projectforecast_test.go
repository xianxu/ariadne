package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// writeFleetProject writes a minimal executing project whose Breakdown cites
// taskRef, under <parent>/<repo>/workshop/projects/<name>.md.
func writeFleetProject(t *testing.T, parent, repo, name, status, taskRef, _ string) string {
	t.Helper()
	return writeProjectFile(t, filepath.Join(parent, repo), name, "---\ntype: project\nname: "+name+"\ngoal: g\ndone_when: d\nstatus: "+status+"\n---\n## Charter\nWhy.\n## Trains\n## Breakdown\n- [ ] task ["+taskRef+"]\n## Log\n## Decisions\n")
}

// writeTrainProject writes an executing project with the given Trains body.
func writeTrainProject(t *testing.T, repoDir, name, trains string) string {
	t.Helper()
	return writeProjectFile(t, repoDir, name, "---\ntype: project\nname: "+name+"\ngoal: g\ndone_when: d\nstatus: executing\ntl: x@ops:0\n---\n## Charter\nWhy.\n## Trains\n"+trains+"\n## Breakdown\n- [ ]\n## Log\n## Decisions\n")
}

func writeProjectFile(t *testing.T, repoDir, name, body string) string {
	t.Helper()
	dir := filepath.Join(repoDir, "workshop", "projects")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name+".md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeForecastLedger points WF_CALIB_LEDGER at a ledger whose eligible rows
// all calibrate to multiplier 1 (build + converge == estimate), k = 0.25.
func writeForecastLedger(t *testing.T) {
	t.Helper()
	var b strings.Builder
	b.WriteString(estimate.Header() + "\n")
	for _, issue := range []string{"000001", "000002", "000003"} {
		b.WriteString(estimate.FormatRow(estimate.LedgerRow{Issue: issue, Estimate: 5, Actual: 5, BuildHours: 4, ConvergeHours: 1, SplitSource: "boundary", FlowKind: "full", Date: "2026-10-01"}) + "\n")
	}
	path := filepath.Join(t.TempDir(), "calibration-ledger.tsv")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_CALIB_LEDGER", path)
}

// stubForecastLookup replaces the shared card resolver with a map-backed fake
// (the projectclose_test pattern); roots records each lookup's vantage.
func stubForecastLookup(t *testing.T, metas map[string]issueMeta, roots *[]string) {
	t.Helper()
	orig := projectIssueLookupFn
	projectIssueLookupFn = func(_ context.Context, ref, root string) (issueMeta, error) {
		if roots != nil {
			*roots = append(*roots, root)
		}
		m, ok := metas[ref]
		if !ok {
			return issueMeta{}, os.ErrNotExist
		}
		return m, nil
	}
	t.Cleanup(func() { projectIssueLookupFn = orig })
}

func TestProjectForecastVerbPrintsRange(t *testing.T) {
	_, _, slot := projectWorkspaceFixture(t)
	path := writeTrainProject(t, slot, "fixture", `| Train | Cars, in order | Slot |
|---|---|---|
| A | ariadne#1 → #2 | :1 |
| B | ariadne#3 | :2 |

- ariadne#2 ⇒ ariadne#3`)
	writeForecastLedger(t)
	stubForecastLookup(t, map[string]issueMeta{
		"ariadne#1": {Status: "done", EstimateHours: 9},
		"ariadne#2": {Status: "working", EstimateHours: 2},
		"ariadne#3": {Status: "open"}, // unestimated: the 5h prior
	}, nil)
	var out bytes.Buffer
	if err := runProjectForecast(context.Background(), &out, &bytes.Buffer{}, &projectForecastFlags{}, path); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"critical path 7.0–7.0h (point 7.0h, ends in train \"B\")", "7.0h remaining across 2 of 2 train(s)", "calibration n=3, ×1.00", "k=0.25", "unestimated, at the 5.0h prior: ariadne#3"} {
		if !strings.Contains(got, want) {
			t.Errorf("forecast missing %q:\n%s", want, got)
		}
	}
	// show prints the same statement best-effort.
	var show bytes.Buffer
	if err := runProjectShow(context.Background(), &show, nil, &projectShowFlags{Slug: "fixture", ProjectsDir: filepath.Dir(path)}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(show.String(), "critical path 7.0–7.0h") || !strings.Contains(show.String(), "tl: x@ops:0") {
		t.Errorf("show missing the forecast or tl:\n%s", show.String())
	}
}

func TestProjectForecastRefusals(t *testing.T) {
	_, _, slot := projectWorkspaceFixture(t)
	noTrains := writeTrainProject(t, slot, "empty", "")
	writeForecastLedger(t)
	if err := runProjectForecast(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, &projectForecastFlags{}, noTrains); err == nil || !strings.Contains(err.Error(), "no trains") {
		t.Fatalf("no trains = %v", err)
	}
	var show bytes.Buffer
	if err := runProjectShow(context.Background(), &show, nil, &projectShowFlags{Slug: "empty", ProjectsDir: filepath.Dir(noTrains)}); err != nil || strings.Contains(show.String(), "forecast") {
		t.Fatalf("show on a trainless project must stay quiet: %v\n%s", err, show.String())
	}
	trains := writeTrainProject(t, slot, "trains", "| Train | Cars |\n|-|-|\n| A | ariadne#1 |")
	t.Setenv("WF_CALIB_LEDGER", filepath.Join(t.TempDir(), "absent.tsv"))
	if err := runProjectForecast(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, &projectForecastFlags{}, trains); err == nil || !strings.Contains(err.Error(), "calibration ledger") {
		t.Fatalf("no ledger = %v", err)
	}
	if err := runProjectForecast(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, &projectForecastFlags{Slug: "x"}, trains); err == nil {
		t.Fatal("file and --slug together accepted")
	}
}

// #329 D1: an archived record keeps `status: done` and its old `deadline:`
// forever. Neither validation surface may reach it — `project validate --all`
// walks only the live home, and the changed-instance gate binds the project
// noun only to the live home — or the frozen archive would start failing.
func TestProjectValidationScopeExcludesArchive(t *testing.T) {
	root := t.TempDir()
	live := writeProjectFile(t, root, "live", "---\ntype: project\nname: live\ngoal: g\ndone_when: d\nstatus: executing\n---\n## Charter\n## Trains\n## Breakdown\n## Log\n## Decisions\n")
	archDir := filepath.Join(root, "workshop", "history", "projects")
	if err := os.MkdirAll(archDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(archDir, "old.md"), []byte("---\ntype: project\nname: old\nstatus: done\ndeadline: 2026-01-01\n---\n## PRD\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var seen []string
	orig := validateFrontmatterFn
	validateFrontmatterFn = func(_, file string) (string, bool, error) {
		seen = append(seen, file)
		b, _ := os.ReadFile(file)
		return "deadline: retired", !strings.Contains(string(b), "deadline:"), nil
	}
	t.Cleanup(func() { validateFrontmatterFn = orig })
	if err := runProjectValidate(&bytes.Buffer{}, &bytes.Buffer{}, &projectValidateFlags{All: true, ProjectsDir: filepath.Dir(live)}, nil); err != nil {
		t.Fatalf("validate --all reached the archive: %v (validated %v)", err, seen)
	}
	if len(seen) != 1 || seen[0] != live {
		t.Fatalf("validated %v, want only the live record", seen)
	}
	gates := nounGates("workshop/issues")
	if _, ok := gateForPath("workshop/history/projects/old.md", gates); ok {
		t.Fatal("the changed-instance gate binds a noun to the project archive")
	}
	if g, ok := gateForPath("workshop/projects/live.md", gates); !ok || g.noun != "project" {
		t.Fatalf("live project path gate = %+v %v", g, ok)
	}
}

// #329: on-demand `project validate` checks the model's sections too, so a
// record still shaped as a shipping plan (PRD/Estimate) is named per section.
func TestProjectValidateChecksSections(t *testing.T) {
	root := t.TempDir()
	old := writeProjectFile(t, root, "old", "---\ntype: project\nname: old\ngoal: g\ndone_when: d\nstatus: defined\n---\n## PRD\n## Estimate\n## Breakdown\n## Log\n")
	orig := validateFrontmatterFn
	validateFrontmatterFn = func(string, string) (string, bool, error) { return "", true, nil }
	t.Cleanup(func() { validateFrontmatterFn = orig })
	var out bytes.Buffer
	err := runProjectValidate(&out, &bytes.Buffer{}, &projectValidateFlags{}, []string{old})
	if err == nil {
		t.Fatal("a PRD/Estimate record conformed")
	}
	for _, want := range []string{"missing `## Charter`", "missing `## Trains`", "missing `## Decisions`"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("validate output missing %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "`## Breakdown`") || strings.Contains(out.String(), "`## Log`") {
		t.Errorf("present sections reported missing:\n%s", out.String())
	}
}

// An optional scaffold section is written by `project new` but not required.
func TestMissingSectionsSkipsOptional(t *testing.T) {
	d, err := projectdoc.ParseDoc("---\ntype: project\n---\n## A\n")
	if err != nil {
		t.Fatal(err)
	}
	got := missingSections(d, []vocab.Section{{Name: "A"}, {Name: "B"}, {Name: "C", Optional: true}})
	if len(got) != 1 || !strings.Contains(got[0], "`## B`") {
		t.Fatalf("missing = %v, want only B", got)
	}
}
