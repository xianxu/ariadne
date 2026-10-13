package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedProjectFleet lays out a temp fleet for fleet-wide project navigation:
//
//	<parent>/ariadne                    — current repo (returned root)
//	<parent>/metis/workshop/projects/p.md            — active, refs metis#18
//	<parent>/metis/workshop/history/projects/q.md    — archived, refs metis#18
//	<parent>/brain (.brain/config.md)/data/project/r.md — legacy, refs metis#18
func seedProjectFleet(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	root := filepath.Join(parent, "ariadne")
	write := func(rel, content string) {
		t.Helper()
		abs := filepath.Join(parent, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "workshop"), 0o755); err != nil {
		t.Fatal(err)
	}
	proj := "---\nstatus: executing\n---\n\n# x\n\n## tasks\n\n- [ ] [metis#18] thing\n"
	write("metis/workshop/projects/p.md", proj)
	write("metis/workshop/history/projects/q.md", strings.Replace(proj, "executing", "done", 1))
	write("brain/.brain/config.md", "brain\n")
	write("brain/data/project/r.md", proj)
	projectWorkspaceGit(t, root, "init", "-b", "main")
	projectWorkspaceGit(t, filepath.Join(parent, "metis"), "init", "-b", "main")
	return root
}

func TestProjectFind_FleetWideArchiveInclusive(t *testing.T) {
	root := seedProjectFleet(t)
	var buf bytes.Buffer
	if err := runProjectFind(&buf, &projectFindFlags{Issue: "metis#18", root: root}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 matches (active, archived, legacy), got %d:\n%s", len(lines), buf.String())
	}
	joined := buf.String()
	for _, want := range []string{"metis/workshop/projects/p.md", "metis/workshop/history/projects/q.md", "brain/data/project/r.md"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s in:\n%s", want, joined)
		}
	}
	for _, l := range lines {
		if strings.Contains(l, "r.md") && !strings.Contains(l, "(legacy)") {
			t.Errorf("legacy match not flagged: %q", l)
		}
		if strings.Contains(l, "p.md") && strings.Contains(l, "(legacy)") {
			t.Errorf("non-legacy match flagged legacy: %q", l)
		}
	}
}

func TestProjectFind_RepoPrefixAndNoMatch(t *testing.T) {
	root := seedProjectFleet(t)
	// Prefix repo token resolves like `sdlc resolve` ("met" → metis).
	var buf bytes.Buffer
	if err := runProjectFind(&buf, &projectFindFlags{Issue: "met#18", root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "p.md") {
		t.Fatalf("prefix repo token did not resolve: %q", buf.String())
	}
	// No match → distinct error, not empty success.
	err := runProjectFind(&bytes.Buffer{}, &projectFindFlags{Issue: "metis#99", root: root})
	if err == nil || !strings.Contains(err.Error(), "no project") {
		t.Fatalf("want no-project error, got %v", err)
	}
}

func TestProjectFind_GitHubRefRejected(t *testing.T) {
	root := seedProjectFleet(t)
	err := runProjectFind(&bytes.Buffer{}, &projectFindFlags{Issue: "gh#42", root: root})
	if err == nil || !strings.Contains(err.Error(), "github") {
		t.Fatalf("want github-ref rejection, got %v", err)
	}
}

// A milestone token in the ref is accepted and ignored — project records are
// found per issue, not per milestone (documented in helptext/resolve.md).
func TestProjectFind_MilestoneTokenIgnored(t *testing.T) {
	root := seedProjectFleet(t)
	var buf bytes.Buffer
	if err := runProjectFind(&buf, &projectFindFlags{Issue: "metis#18 M2", root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "p.md") {
		t.Fatalf("milestone-tagged ref should still find the project: %q", buf.String())
	}
}

func TestResolveRun_KindProject(t *testing.T) {
	root := seedProjectFleet(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "metis#18", kind: "project", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 project paths, got %d:\n%s", len(lines), buf.String())
	}
	if !strings.Contains(buf.String(), "history/projects/q.md") {
		t.Fatalf("archived record must resolve under --kind project: %q", buf.String())
	}
}

func TestResolveRun_KindProjectJSON(t *testing.T) {
	root := seedProjectFleet(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "metis#18", kind: "project", root: root, asJSON: true, out: &buf}); err != nil {
		t.Fatal(err)
	}
	var res resolveResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v (%q)", err, buf.String())
	}
	if res.ID != 18 || len(res.Files) != 3 {
		t.Fatalf("bad result: %+v", res)
	}
	legacies := 0
	for _, f := range res.Files {
		if f.Kind != "project" {
			t.Fatalf("kind = %q, want project: %+v", f.Kind, res.Files)
		}
		if f.Legacy {
			legacies++
			if !strings.Contains(f.Path, "brain") {
				t.Fatalf("legacy flag on a non-brain record: %+v", f)
			}
		}
	}
	if legacies != 1 {
		t.Fatalf("want exactly 1 legacy row in JSON, got %d: %+v", legacies, res.Files)
	}
}

func TestResolveRun_KindUnknown(t *testing.T) {
	root := seedProjectFleet(t)
	err := runResolve(resolveOpts{ref: "metis#18", kind: "bogus", root: root, out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("want unknown-kind error, got %v", err)
	}
}

// Default (no --kind) resolution must stay behavior-identical: issue family
// only, no project records mixed in.
func TestResolveRun_DefaultKindUnchangedByProjects(t *testing.T) {
	root := seedTempRepo(t)
	// Drop a project referencing #144 next to the family; default resolve must not list it.
	projDir := filepath.Join(root, "workshop", "projects")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projDir, "x.md"), []byte("- [ ] [ariadne#144] t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "#144", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "workshop/projects") {
		t.Fatalf("default resolve leaked project records: %q", buf.String())
	}
}

// #329: find and resolve name the train and car carrying the issue. A project
// that carries the issue only as a train car (no `[ref]` row, the ops shape)
// is a match too; the close tick sees it as RowRef=false and leaves it alone.
func TestProjectFind_NamesTrainAndCar(t *testing.T) {
	root := seedProjectFleet(t)
	parent := filepath.Dir(root)
	trains := "---\nstatus: executing\ntl: x@ops:0\n---\n## Trains\n\n| Train | Cars, in order | Slot |\n|---|---|---|\n| Other | pair#1 | :2 |\n| Metis core | metis#17 → #18 (the fix) → #19 | metis:1 |\n"
	if err := os.WriteFile(filepath.Join(parent, "metis/workshop/projects/t.md"), []byte(trains), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runProjectFind(&buf, &projectFindFlags{Issue: "metis#18", root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "metis/workshop/projects/t.md  (train: Metis core, car 2/3)") {
		t.Fatalf("find missing the train line:\n%s", buf.String())
	}
	if strings.Contains(strings.Split(buf.String(), "p.md")[1][:1], "(") {
		t.Fatalf("a row-only match must print the path alone:\n%s", buf.String())
	}
	matches, _, err := discoverProjectsForRef("metis#18", root)
	if err != nil {
		t.Fatal(err)
	}
	var trainOnly, rowOnly bool
	for _, m := range matches {
		if strings.HasSuffix(m.Path, "/t.md") {
			trainOnly = !m.RowRef && m.Train == "Metis core" && m.Car == 2 && m.Cars == 3
		}
		if strings.HasSuffix(m.Path, "/p.md") {
			rowOnly = m.RowRef && m.Train == ""
		}
	}
	if !trainOnly || !rowOnly {
		t.Fatalf("matches = %+v", matches)
	}
}

func TestResolveRun_KindProjectJSONCarriesTrain(t *testing.T) {
	root := seedProjectFleet(t)
	parent := filepath.Dir(root)
	trains := "---\nstatus: executing\n---\n## Trains\n\n| Train | Cars | Slot |\n|---|---|---|\n| Core | metis#18 → #19 | :1 |\n"
	if err := os.WriteFile(filepath.Join(parent, "metis/workshop/projects/t.md"), []byte(trains), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "metis#18", kind: "project", root: root, asJSON: true, out: &buf}); err != nil {
		t.Fatal(err)
	}
	var res resolveResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	sawTrain := false
	for _, f := range res.Files {
		if strings.HasSuffix(f.Path, "/t.md") {
			sawTrain = true
			if f.Train != "Core" || f.Car != 1 {
				t.Fatalf("train file = %+v", f)
			}
		}
		if strings.HasSuffix(f.Path, "/p.md") && (f.Train != "" || f.Car != 0) {
			t.Fatalf("row-only file grew train fields: %+v", f)
		}
	}
	if !sawTrain {
		t.Fatalf("train-only record missing from JSON: %+v", res.Files)
	}
}
