package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestFleetMergeIntentsCapabilityOnPublicCommand(t *testing.T) {
	deps := fleetCommandDeps{
		git: fleetTestGit{},
		normalizeVantage: func(fleet.GitReader, string) (fleet.Vantage, error) {
			return fleet.Vantage{FleetRoot: t.TempDir()}, nil
		},
		collectInventory: func(context.Context, string, fleet.InventoryOptions) (fleet.Inventory, error) {
			return fleet.Inventory{}, nil
		},
	}
	stdout, _, err := executeFleetCommand(t, deps, "inventory", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got fleet.Inventory
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}
	if got.Capabilities.MergeIntents != 1 {
		t.Fatalf("missing producer capability: %s", stdout)
	}
}

type fakeFleetIntentRecord struct {
	intent       landingline.Intent
	live, exists bool
	failure      error
}
type fakeFleetIntentSource struct {
	stores     map[string]map[string]fakeFleetIntentRecord
	denied     map[string]error
	reads      map[string]int
	pause      string
	pauseReads int
}

func (f *fakeFleetIntentSource) observe(_ context.Context, root, note string) ([]mergeIntentReport, error) {
	common := root
	f.reads[common]++
	var reports []mergeIntentReport
	for path, record := range f.stores[common] {
		r := reportMergeIntent(landingline.ObservedIntent{Path: path, Intent: record.intent, Locked: record.live, Err: record.failure}, record.exists, nil)
		if r.State == "orphaned" {
			r.Note = note
		}
		reports = append(reports, r)
	}
	return reports, f.denied[common]
}

func fleetIntentRow(root, common string) fleet.TreeRow {
	return fleet.TreeRow{RepoIdentity: common, RepoRoot: root, TreePath: root, Branch: "main", Facts: fleet.MeasuredFacts{Error: "not collected in fixture"}, Issues: []fleet.IssueAssociation{}, Policy: fleet.PolicyCapability{Diagnostic: &fleet.PolicyDiagnostic{Code: fleet.DiagnosticMissingPolicy, Message: "private fixture"}}}
}

func TestFleetMergeIntentsStatefulPublicCommand(t *testing.T) {
	base := canonicalRoot(t.TempDir())
	primary, slot, dependency := filepath.Join(base, "sample"), filepath.Join(base, "worktree/sample-slot1/sample"), filepath.Join(base, "worktree/sample-slot1/base")
	common, depCommon := filepath.Join(primary, ".git"), filepath.Join(dependency, ".git")
	path, depPath := filepath.Join(common, "sdlc/landing/intents/a"), filepath.Join(depCommon, "sdlc/landing/intents/a")
	in := landingline.Intent{Version: 1, Branch: "000624-example", Head: strings.Repeat("a", 40), Root: slot, Workspace: "wrong:9", PID: 42, StartedAt: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)}
	dep := in
	dep.Root = dependency
	dep.Workspace = ""
	fake := &fakeFleetIntentSource{stores: map[string]map[string]fakeFleetIntentRecord{common: {path: {intent: in, live: true, exists: true}}, depCommon: {depPath: {intent: dep, exists: true}}}, denied: map[string]error{}, reads: map[string]int{}}
	rows := []fleet.TreeRow{fleetIntentRow(primary, common), fleetIntentRow(slot, common), fleetIntentRow(dependency, depCommon)}
	deps := fleetCommandDeps{git: fleetTestGit{}, normalizeVantage: func(fleet.GitReader, string) (fleet.Vantage, error) { return fleet.Vantage{FleetRoot: base}, nil }, collectInventory: func(context.Context, string, fleet.InventoryOptions) (fleet.Inventory, error) {
		return fleet.Inventory{Rows: rows}, nil
	}, observeMergeIntents: fake.observe, pauseNote: func() string { fake.pauseReads++; return fake.pause }}
	read := func() fleet.Inventory {
		t.Helper()
		before := fake.pauseReads
		out, _, err := executeFleetCommand(t, deps, "inventory", "--json")
		if err != nil {
			t.Fatal(err)
		}
		var got fleet.Inventory
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatal(err)
		}
		if fake.pauseReads != before+1 {
			t.Fatal("pause read repeated per store/row")
		}
		for _, row := range rows {
			if row.MergeIntentsState != "" {
				t.Fatal("mutated caller-owned rows")
			}
		}
		return got
	}
	first := read()
	if len(first.Rows[0].MergeIntents) != 0 || len(first.Rows[1].MergeIntents) != 1 || len(first.Rows[2].MergeIntents) != 1 {
		t.Fatalf("wrong root placement: %+v", first.Rows)
	}
	if first.Rows[1].MergeIntents[0].State != "live" || first.Rows[2].MergeIntents[0].State != "orphaned" {
		t.Fatal("wrong state")
	}
	if fake.reads[common] != 1 || fake.reads[depCommon] != 1 {
		t.Fatalf("repeated store reads: %v", fake.reads)
	}
	record := fake.stores[common][path]
	record.live = false
	fake.stores[common][path] = record
	fake.pause = "the machine queue is paused by a literal owner"
	second := read()
	orphan := second.Rows[1].MergeIntents[0]
	if orphan.State != "orphaned" || orphan.Note != fake.pause || orphan.Reentry != "sdlc merge --branch "+in.Branch+" --yes" || orphan.Root != slot {
		t.Fatalf("bad orphan: %+v", orphan)
	}
	fake.pause = ""
	delete(fake.stores[common], path)
	third := read()
	if third.Rows[1].MergeIntentsState != "complete" || len(third.Rows[1].MergeIntents) != 0 {
		t.Fatal("terminal removal did not yield complete empty")
	}
	record.exists = false
	fake.stores[common][path] = record
	if got := read().Rows[1].MergeIntents[0]; got.State != "stale" || got.Reentry != "" {
		t.Fatalf("deleted branch: %+v", got)
	}
	bad := filepath.Join(common, "sdlc/landing/intents/bad")
	fake.stores[common][bad] = fakeFleetIntentRecord{failure: errors.New("corrupt v1")}
	mixed := read()
	if mixed.Rows[1].MergeIntentsState != "partial" || len(mixed.Rows[1].MergeIntents) != 1 || len(mixed.Diagnostics) != 1 || mixed.Diagnostics[0].Stage != "merge_intents" {
		t.Fatalf("lost mixed observations: %+v", mixed)
	}
	delete(fake.stores[common], bad)
	fake.stores[common] = nil
	fake.denied[common] = errors.New("store unreadable")
	last := read()
	if last.Rows[1].MergeIntentsState != "unreadable" || last.Rows[2].MergeIntentsState != "complete" {
		t.Fatalf("store failure spread or vanished: %+v", last.Rows)
	}
}

func TestFleetMergeIntentsPlacementConflictsAndGrowth(t *testing.T) {
	for _, n := range []int{8, 16} {
		base := canonicalRoot(t.TempDir())
		common := filepath.Join(base, ".git")
		var rows []fleet.TreeRow
		var reports []mergeIntentReport
		for i := 0; i < n; i++ {
			root := filepath.Join(base, fmt.Sprint(i))
			rows = append(rows, fleetIntentRow(root, common))
			path := filepath.Join(common, "sdlc/landing/intents", fmt.Sprint(i))
			reports = append(reports, mergeIntentReport{State: "unreadable", Root: root, Identity: path, Path: path, Note: "inspection only"})
		}
		got, c := placeFleetMergeIntents(fleet.Inventory{Rows: rows}, []fleetIntentRead{{Common: common, Root: base, Reports: reports}})
		if c.Rows != n || c.Records != n || c.RootLookups != n {
			t.Fatalf("N=%d growth: %+v", n, c)
		}
		for _, row := range got.Rows {
			if len(row.MergeIntents) != 1 {
				t.Fatal("lost root")
			}
		}
		duplicate := reports[0]
		dedup, _ := placeFleetMergeIntents(fleet.Inventory{Rows: rows}, []fleetIntentRead{{Common: common, Root: base, Reports: []mergeIntentReport{duplicate, duplicate}}})
		if len(dedup.Rows[0].MergeIntents) != 1 {
			t.Fatal("same physical record duplicated")
		}
		conflict := duplicate
		conflict.Note = "different authority"
		conflicted, _ := placeFleetMergeIntents(fleet.Inventory{Rows: rows}, []fleetIntentRead{{Common: common, Root: base, Reports: []mergeIntentReport{duplicate, conflict}}})
		if len(conflicted.Rows[0].MergeIntents) != 0 || len(conflicted.Diagnostics) != 1 {
			t.Fatal("conflict became usable evidence")
		}
		unmatched := duplicate
		unmatched.Root = filepath.Join(base, "outside")
		missing, _ := placeFleetMergeIntents(fleet.Inventory{Rows: rows}, []fleetIntentRead{{Common: common, Root: base, Reports: []mergeIntentReport{unmatched, unmatched}}})
		if len(missing.Diagnostics) != 1 || missing.Rows[0].MergeIntentsState != "partial" {
			t.Fatal("unmatched root guessed or diagnosed twice")
		}
	}
}

func TestFleetMergeIntentsBoundedJSON(t *testing.T) {
	inventory := fleet.Inventory{Diagnostics: []fleet.RepoDiagnostic{{RepoPath: "/fixture", Stage: "merge_intents", Message: "x"}}}
	raw, err := json.Marshal(inventory)
	if err != nil {
		t.Fatal(err)
	}
	inventory.Diagnostics[0].Message = strings.Repeat("x", maxFleetInventoryJSON-len(raw))
	var out bytes.Buffer
	if err := writeFleetInventoryJSON(context.Background(), &out, inventory); err != nil || out.Len() != maxFleetInventoryJSON {
		t.Fatalf("exact bound: bytes=%d err=%v", out.Len(), err)
	}
	inventory.Diagnostics[0].Message += "x"
	out.Reset()
	if err := writeFleetInventoryJSON(context.Background(), &out, inventory); err == nil || out.Len() != 0 {
		t.Fatalf("overflow leaked JSON: bytes=%d err=%v", out.Len(), err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out.Reset()
	if err := writeFleetInventoryJSON(ctx, &out, fleet.Inventory{}); !errors.Is(err, context.Canceled) || out.Len() != 0 {
		t.Fatal("canceled publication")
	}
	if err := writeFleetInventoryJSON(context.Background(), shortFleetWriter{}, fleet.Inventory{}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write accepted: %v", err)
	}
}

type shortFleetWriter struct{}

func (shortFleetWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }

func TestFleetMergeIntentsStoreReadsGrowByStoresNotRows(t *testing.T) {
	for _, n := range []int{8, 16} {
		base := canonicalRoot(t.TempDir())
		common := filepath.Join(base, ".git")
		var rows []fleet.TreeRow
		for i := 0; i < n; i++ {
			root := filepath.Join(base, fmt.Sprint(i))
			rows = append(rows, fleetIntentRow(root, common))
		}
		fake := &fakeFleetIntentSource{stores: map[string]map[string]fakeFleetIntentRecord{}, denied: map[string]error{}, reads: map[string]int{}}
		got, err := collectFleetMergeIntents(context.Background(), fleet.Inventory{Rows: rows}, fake.observe, func() string { fake.pauseReads++; return "" })
		if err != nil || len(got.Rows) != n || fake.reads[common] != 1 || fake.pauseReads != 1 {
			t.Fatalf("N=%d repeated observations: rows=%d reads=%v pauses=%d err=%v", n, len(got.Rows), fake.reads, fake.pauseReads, err)
		}
	}
}

func TestFleetMergeIntentsMissingReadIsUnknown(t *testing.T) {
	base := canonicalRoot(t.TempDir())
	common := filepath.Join(base, ".git")
	got, _ := placeFleetMergeIntents(fleet.Inventory{Rows: []fleet.TreeRow{fleetIntentRow(base, common)}}, nil)
	if got.Rows[0].MergeIntentsState != "unreadable" || got.Rows[0].MergeIntentsError == "" {
		t.Fatalf("missing observation became absence: %+v", got.Rows[0])
	}
}

func TestFleetMergeIntentsUndiscoveredStoreHasDiagnostic(t *testing.T) {
	base := canonicalRoot(t.TempDir())
	inventory := fleet.Inventory{Diagnostics: []fleet.RepoDiagnostic{{RepoPath: base, Stage: "git", Message: "common directory unavailable"}}}
	got, _ := placeFleetMergeIntents(inventory, nil)
	found := 0
	for _, d := range got.Diagnostics {
		if d.Stage == "merge_intents" {
			found++
			if !strings.Contains(d.Message, "unavailable") {
				t.Fatal("lost discovery reason")
			}
		}
	}
	if found != 1 {
		t.Fatalf("no row and no merge-read diagnostic: %+v", got)
	}
}

func TestFleetMergeIntentsPartialPayloadGrowth(t *testing.T) {
	previous := 0
	for _, n := range []int{8, 16} {
		const base = "/fixture"
		common := filepath.Join(base, ".git")
		var rows []fleet.TreeRow
		var reports []mergeIntentReport
		for i := 0; i < n; i++ {
			root := filepath.Join(base, fmt.Sprintf("%02d", i))
			path := filepath.Join(common, "sdlc/landing/intents", fmt.Sprintf("%02d", i))
			rows = append(rows, fleetIntentRow(root, common))
			reports = append(reports, mergeIntentReport{State: "unreadable", Root: root, Identity: path, Path: path, Note: "branch authority unavailable"})
		}
		got, _ := placeFleetMergeIntents(fleet.Inventory{Rows: rows}, []fleetIntentRead{{Common: common, Root: base, Reports: reports}})
		raw, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		if previous > 0 && len(raw) > 2*previous {
			t.Fatalf("8→16 partial rows/records grew beyond linear encoded bytes: %d→%d", previous, len(raw))
		}
		previous = len(raw)
	}
}

func TestFleetMergeIntentsPreservesDiagnosticOrder(t *testing.T) {
	input := fleet.Inventory{Diagnostics: []fleet.RepoDiagnostic{
		{RepoIdentity: "same", RepoPath: "/z", Stage: "facts", Message: "a"},
		{RepoIdentity: "same", RepoPath: "/a", Stage: "facts", Message: "z"},
	}}
	got, _ := placeFleetMergeIntents(input, nil)
	if got.Diagnostics[0].RepoPath != "/a" {
		t.Fatalf("changed the existing repository/path-first diagnostic order: %+v", got.Diagnostics)
	}
}

func TestFleetMergeIntentsCancellationDoesNotPublishPartialEvidence(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	base := canonicalRoot(t.TempDir())
	row := fleetIntentRow(base, filepath.Join(base, ".git"))
	got, err := collectFleetMergeIntents(ctx, fleet.Inventory{Rows: []fleet.TreeRow{row}}, func(context.Context, string, string) ([]mergeIntentReport, error) { cancel(); return nil, nil }, func() string { return "" })
	if !errors.Is(err, context.Canceled) || got.Capabilities.MergeIntents != 0 {
		t.Fatalf("canceled partial observation became a report: %+v %v", got, err)
	}
}

func TestFleetMergeIntentsSeveralBranchesAndAmbiguousRoot(t *testing.T) {
	const root = "/fixture/worktree/sample-slot1/sample"
	const common = "/fixture/sample/.git"
	row := fleetIntentRow(root, common)
	var reports []mergeIntentReport
	for _, suffix := range []string{"a", "b"} {
		branch := "000624-" + suffix
		path := filepath.Join(common, "sdlc/landing/intents", suffix)
		reports = append(reports, mergeIntentReport{State: "orphaned", Branch: branch, Head: strings.Repeat("a", 40), Root: root, PID: 42, StartedAt: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC), Reentry: "sdlc merge --branch " + branch + " --yes", Path: path, Identity: path})
	}
	got, _ := placeFleetMergeIntents(fleet.Inventory{Rows: []fleet.TreeRow{row}}, []fleetIntentRead{{Common: common, Root: root, Reports: reports}})
	if len(got.Rows[0].MergeIntents) != 2 || got.Rows[0].MergeIntentsState != "complete" {
		t.Fatalf("several branches on one root lost: %+v", got)
	}
	if _, err := json.Marshal(got); err != nil {
		t.Fatal(err)
	}
	ambiguous, _ := placeFleetMergeIntents(fleet.Inventory{Rows: []fleet.TreeRow{row, row}}, []fleetIntentRead{{Common: common, Root: root, Reports: reports[:1]}})
	if len(ambiguous.Diagnostics) != 1 || len(ambiguous.Rows[0].MergeIntents) != 0 || len(ambiguous.Rows[1].MergeIntents) != 0 {
		t.Fatalf("ambiguous root guessed an owner: %+v", ambiguous)
	}
}

func TestFleetMergeIntentsHeldCorruptionIsNeverOrphaned(t *testing.T) {
	report := reportMergeIntent(landingline.ObservedIntent{Path: "/fixture/.git/sdlc/landing/intents/a", Locked: true, Err: errors.New("partial owner write")}, true, nil)
	if report.State != "live" || report.Reentry != "" || report.Root != "" {
		t.Fatalf("corrupt held record invented orphan authority: %+v", report)
	}
}
