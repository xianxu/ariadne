package tracker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func writeDetails(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Vary the card and the details independently: card-owned fields follow the
// card, detail-owned fields follow the details, and a missing half is unknown.
func TestRecordsTakeEachFieldFromItsOwner(t *testing.T) {
	r, root, _ := fixture(t)
	dir := filepath.Join(root, "workshop", "issues")
	// Stale mirror: details still say blocked with its own deps and title.
	writeDetails(t, dir, "000252-test.md", "---\nid: 000252\nstatus: blocked\ndeps: [x#1]\n---\n\n# Mirror title\n")
	writeDetails(t, dir, "000300-uncarded.md", "---\nid: 000300\nstatus: working\ndeps: []\n---\n\n# Uncarded\n")
	rs, err := LoadRecords(context.Background(), r, dir, Fresh)
	if err != nil || !rs.Tracker || rs.Stale {
		t.Fatalf("load: %+v %v", rs, err)
	}
	rec, _ := rs.get("000252")
	if rec.Status() != "open" || rec.Title() != "Tracker title" {
		t.Fatalf("card-owned fields from the mirror: %s %q", rec.Status(), rec.Title())
	}
	if deps, ok := rec.Field("deps"); !ok || deps != "[x#1]" {
		t.Fatalf("detail-owned deps: %q %v", deps, ok)
	}
	orphan, _ := rs.get("000300")
	if s, ok := orphan.Field("status"); ok || s != "" {
		t.Fatalf("uncarded details answered a card field: %q", s)
	}
	if err := os.Remove(filepath.Join(dir, "000252-test.md")); err != nil {
		t.Fatal(err)
	}
	rs, _ = LoadRecords(context.Background(), r, dir, Fresh)
	cardOnly, _ := rs.get("000252")
	if _, ok := cardOnly.Field("deps"); ok || cardOnly.Status() != "open" {
		t.Fatal("card-only record invented details or lost its card")
	}
}

func TestRecordsWithoutTrackerUseDetails(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	origin := filepath.Join(t.TempDir(), "o.git")
	testfix.Git(t, root, "init", "--bare", "-q", origin)
	testfix.Git(t, root, "remote", "add", "publication", origin)
	repo, err := NewRepository(context.Background(), root, "publication")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "workshop", "issues")
	writeDetails(t, dir, "000001-legacy.md", "---\nid: 000001\nstatus: working\n---\n\n# Legacy\n")
	rs, err := LoadRecords(context.Background(), repo, dir, Fresh)
	if err != nil || rs.Tracker {
		t.Fatalf("legacy: %+v %v", rs, err)
	}
	if rec, _ := rs.get("000001"); rec.Status() != "working" || rec.Title() != "Legacy" {
		t.Fatal("pre-tracker details are not the record")
	}
}

func TestRecordsOfflineReadLastFetchedCardsAsStale(t *testing.T) {
	r, root, _ := fixture(t)
	if _, err := r.Snapshot(); err != nil { // leaves a tracking ref behind
		t.Fatal(err)
	}
	testfix.Git(t, root, "remote", "set-url", "publication", filepath.Join(t.TempDir(), "gone.git"))
	dir := filepath.Join(root, "workshop", "issues")
	if _, err := LoadRecords(context.Background(), r, dir, Fresh); err == nil {
		t.Fatal("a fresh read succeeded offline")
	}
	rs, err := LoadRecords(context.Background(), r, dir, PreferFresh)
	if err != nil || !rs.Stale || !rs.Tracker {
		t.Fatalf("offline preferred read: %+v %v", rs, err)
	}
	if rec, ok := rs.get("000252"); !ok || rec.Status() != "open" || !strings.HasPrefix(rec.Title(), "Tracker") {
		t.Fatal("stale read lost the last-fetched card")
	}
}

// composeRecords is pure: vary cards and detail files independently.
func TestComposeRecordsJoinsByIDWithoutIO(t *testing.T) {
	card, err := issue.ParseCard([]byte(testCard))
	if err != nil {
		t.Fatal(err)
	}
	cards := []Record{{ID: "000252", Path: testPath, Card: card, Raw: []byte(testCard)}}
	files := []DetailFile{
		{Path: "/d/000252-test.md", Raw: []byte("---\nid: 000252\nstatus: blocked\ndeps: [a#1]\n---\n# T\n")},
		{Path: "/d/000252-copy.md", Raw: []byte("---\nid: 000252\n---\n# Copy\n")},
		{Path: "/d/000301-unread.md", ReadErr: errors.New("permission denied")},
		{Path: "/d/000302-.md"},
	}
	rs := ComposeRecords(Records{Tracker: true}, cards, nil, files)
	all := rs.All()
	if len(all) != 3 || !all[1].Duplicate || all[2].ID != "000301" || !all[2].DetailUnreadable {
		t.Fatalf("composition: %+v", all)
	}
	rec, _ := rs.get("000252")
	if rec.Status() != "open" || rec.DetailPath != "/d/000252-test.md" {
		t.Fatalf("card-owned status or first details: %+v", rec)
	}
	if deps, _ := rec.Field("deps"); deps != "[a#1]" {
		t.Fatalf("detail-owned deps: %q", deps)
	}
}

// The operating envelope's read path (#252 M3): 10,000 cards joined with 100
// active details, the composed view every reader consumes.
func BenchmarkComposeRecordsTenThousandCardsHundredDetails(b *testing.B) {
	cards := make([]Record, 0, 10000)
	for id := 1; id <= 10000; id++ {
		key := fmt.Sprintf("%06d", id)
		raw := []byte(strings.ReplaceAll(testCard, "000252", key))
		card, err := issue.ParseCard(raw)
		if err != nil {
			b.Fatal(err)
		}
		cards = append(cards, Record{ID: key, Path: "workshop/issue-cards/" + key + "-sample.md", Card: card, Raw: raw})
	}
	details := make([]DetailFile, 0, 100)
	for id := 9901; id <= 10000; id++ {
		key := fmt.Sprintf("%06d", id)
		details = append(details, DetailFile{Path: "workshop/issues/" + key + "-sample.md", Raw: []byte("---\nid: " + key + "\nstatus: working\ndeps: []\n---\n\n# T\n\n## Problem\nx\n")})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		rs := ComposeRecords(Records{Tracker: true}, cards, nil, details)
		if rec, ok := rs.get("010000"); !ok || rec.DetailPath == "" || len(rs.All()) != 10000 {
			b.Fatal("compose lost records")
		}
	}
}

// #288: an unreadable card composes as CardErr with its details; card-owned
// fields read as unknown — never the (possibly stale) mirror.
func TestComposeRecordsCarriesUnreadableCards(t *testing.T) {
	cause := errors.New("tracker workshop/issue-cards/000007-bad.md: invalid claimant")
	files := []DetailFile{{Path: "/d/000007-bad.md", Raw: []byte("---\nid: 000007\nstatus: working\n---\n# Bad\n")}}
	rs := ComposeRecords(Records{Tracker: true}, nil, []UnreadableCard{{ID: "000007", Path: "workshop/issue-cards/000007-bad.md", Err: cause}}, files)
	rec, ok := rs.get("000007")
	if !ok || rec.Card != nil || !errors.Is(rec.CardErr, cause) || rec.DetailPath == "" {
		t.Fatalf("composed: %+v", rec)
	}
	if s, ok := rec.Field("status"); ok || s != "" {
		t.Fatalf("status of an unreadable card read %q from the mirror", s)
	}
	if _, ok, err := rs.Require("000007"); !ok || !errors.Is(err, cause) {
		t.Fatalf("Require(unreadable): ok=%v err=%v", ok, err)
	}
	if _, ok, err := rs.Require("000008"); ok || err != nil {
		t.Fatalf("Require(missing): ok=%v err=%v", ok, err)
	}
}
