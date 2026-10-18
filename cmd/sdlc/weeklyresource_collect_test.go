package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectResourceFamiliesFromDeclaredPatterns(t *testing.T) {
	root := t.TempDir()
	write := func(name, data string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("repos/a/ledger-one.jsonl", "abc")
	write("repos/a/other.log", "ignore")
	write("archive/x/one", "xy")
	write("archive/x/two", "z")
	outside := filepath.Join(t.TempDir(), "large")
	if err := os.WriteFile(outside, []byte(strings.Repeat("x", 100)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "repos/a/ledger-link.jsonl")); err != nil {
		t.Fatal(err)
	}
	inventory := []byte(`{"version":1,"sources":[{"kind":"couch-ledgers","root_token":"pair-data","patterns":["repos/*/ledger-*.jsonl"]},{"kind":"captures","root_token":"pair-data","patterns":["archive/*"]},{"kind":"missing","root_token":"absent","patterns":["*.jsonl"]}]}`)
	rows, visited, err := collectResourceFamilies(inventory, []ResourceRoot{{Token: "pair-data", Path: root}})
	if err != nil {
		t.Fatal(err)
	}
	if visited < 7 || visited > 20 {
		t.Fatalf("single traversal count = %d", visited)
	}
	byKind := map[string]ResourceObservation{}
	for _, row := range rows {
		byKind[row.Kind] = row
	}
	if row := byKind["couch-ledgers"]; !row.Complete || row.Bytes != 3 || row.Files != 1 || row.Definition == "" {
		t.Fatalf("ledger row = %+v", row)
	}
	if row := byKind["captures"]; !row.Complete || row.Bytes != 3 || row.Files != 2 {
		t.Fatalf("directory selector row = %+v", row)
	}
	if row := byKind["missing"]; row.Complete || !strings.Contains(row.Reason, "absent") {
		t.Fatalf("missing root must be unknown: %+v", row)
	}
}

func TestCollectResourceFamiliesDoesNotDoubleAssign(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "one.jsonl"), []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	inventory := []byte(`{"version":1,"sources":[{"kind":"first","root_token":"data","patterns":["*.jsonl"]},{"kind":"second","root_token":"data","patterns":["one.*"]}]}`)
	rows, _, err := collectResourceFamilies(inventory, []ResourceRoot{{Token: "data", Path: root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Complete || rows[1].Complete || rows[0].Bytes != 0 || rows[1].Bytes != 0 {
		t.Fatalf("ambiguous assignment must count nowhere: %+v", rows)
	}
}

func TestCollectResourceFamiliesOverlappingRootsRemoveEarlierAssignment(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "one.jsonl"), []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	inventory := []byte(`{"version":1,"sources":[{"kind":"first","root_token":"first-root","patterns":["*.jsonl"]},{"kind":"second","root_token":"second-root","patterns":["one.*"]},{"kind":"third","root_token":"second-root","patterns":["*.jsonl"]}]}`)
	rows, _, err := collectResourceFamilies(inventory, []ResourceRoot{{Token: "first-root", Path: root}, {Token: "second-root", Path: root}})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Complete || row.Bytes != 0 || row.Files != 0 {
			t.Fatalf("shared physical file must count nowhere: %+v", rows)
		}
	}
}

func TestCollectResourceDetailsRanksLedgersInSameWalk(t *testing.T) {
	root := t.TempDir()
	for i, size := range []int{2, 5, 5, 1} {
		name := filepath.Join(root, "ledger-"+string(rune('a'+i))+".jsonl")
		if err := os.WriteFile(name, []byte(strings.Repeat("x", size)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	inventory := []byte(`{"version":1,"sources":[{"kind":"couch-ledgers","root_token":"data","patterns":["ledger-*.jsonl"]}]}`)
	got, err := collectResourceDetails(inventory, []ResourceRoot{{Token: "data", Path: root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Families) != 1 || got.Families[0].Bytes != 13 || got.Visited != 5 || len(got.LargestLedgers) != 4 {
		t.Fatalf("one walk and four ledgers: %+v", got)
	}
	if got.LargestLedgers[0].Bytes != 5 || !strings.HasSuffix(got.LargestLedgers[0].Path, "ledger-b.jsonl") || !strings.HasSuffix(got.LargestLedgers[1].Path, "ledger-c.jsonl") {
		t.Fatalf("largest first, stable tie: %+v", got.LargestLedgers)
	}
	for i := 4; i < 14; i++ {
		name := filepath.Join(root, "ledger-"+string(rune('a'+i))+".jsonl")
		if err := os.WriteFile(name, []byte(strings.Repeat("x", i+1)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err = collectResourceDetails(inventory, []ResourceRoot{{Token: "data", Path: root}})
	if err != nil || len(got.LargestLedgers) != 10 || got.LargestLedgers[0].Bytes != 14 || got.LargestLedgers[9].Bytes != 5 {
		t.Fatalf("top ten after more ledgers = %+v, %v", got.LargestLedgers, err)
	}
}

func TestCollectResourceDetailsVisitsLinearly(t *testing.T) {
	inventory := []byte(`{"version":1,"sources":[{"kind":"files","root_token":"data","patterns":["*.bin"]}]}`)
	visit := func(n int) int {
		t.Helper()
		root := t.TempDir()
		for i := 0; i < n; i++ {
			name := filepath.Join(root, string(rune('a'+i))+".bin")
			if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		got, err := collectResourceDetails(inventory, []ResourceRoot{{Token: "data", Path: root}})
		if err != nil || got.Families[0].Files != n {
			t.Fatalf("n=%d result=%+v err=%v", n, got, err)
		}
		return got.Visited
	}
	n, twice := visit(8), visit(16)
	if twice > 2*n || twice < n {
		t.Fatalf("walk work must scale with entries: n=%d 2n=%d", n, twice)
	}
}

func TestCollectResourceDetailsBoundsLargeWalk(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.bin", "b.bin", "c.bin", "d.bin"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	inventory := []byte(`{"version":1,"sources":[{"kind":"files","root_token":"data","patterns":["*.bin"]}]}`)
	got, err := collectResourceDetailsWithLimit(inventory, []ResourceRoot{{Token: "data", Path: root}}, 3)
	if err != nil || len(got.Families) != 1 || got.Families[0].Complete || !strings.Contains(got.Families[0].Reason, "limit") {
		t.Fatalf("bounded walk must keep named partial coverage: %+v %v", got, err)
	}
}
