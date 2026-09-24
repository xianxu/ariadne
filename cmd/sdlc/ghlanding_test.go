package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func landingGHRecord() map[string]any {
	return map[string]any{"number": 130, "state": "MERGED", "headRefName": "000246-landing", "headRefOid": strings.Repeat("a", 40), "baseRefName": "main", "baseRefOid": strings.Repeat("b", 40), "mergeCommit": map[string]any{"oid": strings.Repeat("c", 40)}, "url": "https://github.com/example/repo/pull/130", "isCrossRepository": false, "headRepository": map[string]any{"name": "repo", "nameWithOwner": "example/repo"}, "headRepositoryOwner": map[string]any{"login": "example"}}
}
func landingGHJSON(t *testing.T, records ...map[string]any) []byte {
	t.Helper()
	if records == nil {
		records = []map[string]any{}
	}
	b, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func TestLandingGHParse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
		fail   bool
	}{
		{"merged", func(map[string]any) {}, false},
		{"open-queued", func(r map[string]any) { r["state"] = "OPEN"; r["mergeCommit"] = nil }, false},
		{"closed", func(r map[string]any) { r["state"] = "CLOSED"; r["mergeCommit"] = nil }, false},
		{"fork", func(r map[string]any) { r["isCrossRepository"] = true }, true},
		{"wrong-head-owner", func(r map[string]any) { r["headRepositoryOwner"] = map[string]any{"login": "other"} }, true},
		{"wrong-head-repo", func(r map[string]any) {
			r["headRepository"] = map[string]any{"name": "other", "nameWithOwner": "example/other"}
		}, true},
		{"wrong-base-repo", func(r map[string]any) { r["url"] = "https://github.com/other/repo/pull/130" }, true},
		{"wrong-url-number", func(r map[string]any) { r["url"] = "https://github.com/example/repo/pull/131" }, true},
		{"wrong-base", func(r map[string]any) { r["baseRefName"] = "develop" }, true},
		{"wrong-branch", func(r map[string]any) { r["headRefName"] = "other" }, true},
		{"short-head", func(r map[string]any) { r["headRefOid"] = "abc123" }, true},
		{"missing-base", func(r map[string]any) { delete(r, "baseRefOid") }, true},
		{"missing-merge", func(r map[string]any) { r["mergeCommit"] = nil }, true},
		{"null-head-repository", func(r map[string]any) { r["headRepository"] = nil }, true},
		{"missing-cross-repo", func(r map[string]any) { delete(r, "isCrossRepository") }, true},
		{"unknown-state", func(r map[string]any) { r["state"] = "QUEUED" }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := landingGHRecord()
			tc.mutate(record)
			got, err := parseLandingPRs(landingGHJSON(t, record), "example/repo", "000246-landing")
			if (err != nil) != tc.fail {
				t.Fatalf("error=%v want refusal=%v", err, tc.fail)
			}
			if !tc.fail && (len(got) != 1 || got[0].Number != 130 || got[0].HeadOID != record["headRefOid"] || got[0].BaseOID != record["baseRefOid"] || got[0].Repo != "example/repo") {
				t.Fatalf("lost structured evidence: %+v", got)
			}
		})
	}
	for _, data := range []string{"null", "{}", "[", "[] trailing"} {
		if _, err := parseLandingPRs([]byte(data), "example/repo", "000246-landing"); err == nil {
			t.Errorf("accepted %q", data)
		}
	}
	if got, err := parseLandingPRs([]byte("[]"), "example/repo", "000246-landing"); err != nil || len(got) != 0 {
		t.Fatalf("absence is not an error: %v %v", got, err)
	}
}

func TestLandingGHBoundedRead(t *testing.T) {
	records := make([]map[string]any, 101)
	for i := range records {
		records[i] = landingGHRecord()
		records[i]["number"] = i + 1
		records[i]["url"] = "https://github.com/example/repo/pull/" + strconv.Itoa(i+1)
	}
	if got, err := parseLandingPRs(landingGHJSON(t, records[:100]...), "example/repo", "000246-landing"); err != nil || len(got) != 100 {
		t.Fatalf("100 records: %d %v", len(got), err)
	}
	if _, err := parseLandingPRs(landingGHJSON(t, records...), "example/repo", "000246-landing"); err == nil {
		t.Fatal("silently truncated evidence")
	}
	if _, err := parseLandingPRs(landingGHJSON(t, records[0], records[0]), "example/repo", "000246-landing"); err == nil {
		t.Fatal("duplicate PR record accepted")
	}
}

func fakeLandingGH(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "gh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}
func TestLandingGHDispatch(t *testing.T) {
	dir := fakeLandingGH(t, "printf '%s\\n' \"$@\" > \"$LANDING_ARGV\"\ncat \"$LANDING_RESPONSE\"\n")
	argsPath := filepath.Join(dir, "argv")
	response := filepath.Join(dir, "response")
	t.Setenv("LANDING_ARGV", argsPath)
	t.Setenv("LANDING_RESPONSE", response)
	if err := os.WriteFile(response, landingGHJSON(t, landingGHRecord()), 0644); err != nil {
		t.Fatal(err)
	}
	gh := realGH{}
	if got, err := gh.LandingPRs(context.Background(), "example/repo", "000246-landing"); err != nil || len(got) != 1 {
		t.Fatalf("query: %v %v", got, err)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"pr", "list", "--repo", "example/repo", "--head", "000246-landing", "--state", "all", "--limit", "101", "--json", landingPRFields}
	if got := strings.Split(strings.TrimSpace(string(args)), "\n"); !reflect.DeepEqual(got, want) {
		t.Fatalf("query argv: %q", got)
	}
	if err := gh.LandingMerge(context.Background(), "example/repo", 130, strings.Repeat("a", 40)); err != nil {
		t.Fatal(err)
	}
	args, err = os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"pr", "merge", "130", "--repo", "example/repo", "--merge", "--match-head-commit", strings.Repeat("a", 40)}
	if got := strings.Split(strings.TrimSpace(string(args)), "\n"); !reflect.DeepEqual(got, want) {
		t.Fatalf("merge argv (must never delete branches): %q", got)
	}
}
func TestLandingGHFailuresAndCancellation(t *testing.T) {
	t.Run("failure-is-not-absence", func(t *testing.T) {
		fakeLandingGH(t, "echo unavailable >&2\nexit 1\n")
		if _, err := (realGH{}).LandingPRs(context.Background(), "example/repo", "issue"); err == nil || !strings.Contains(err.Error(), "unavailable") {
			t.Fatalf("lost query failure: %v", err)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		fakeLandingGH(t, "exec sleep 30\n")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		start := time.Now()
		_, err := (realGH{}).LandingPRs(ctx, "example/repo", "issue")
		if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 3*time.Second {
			t.Fatalf("cancellation: %v elapsed %s", err, time.Since(start))
		}
	})
	t.Run("invalid-request", func(t *testing.T) {
		fakeLandingGH(t, "echo MUST-NOT-RUN\nexit 0\n")
		for _, repo := range []string{"", "other/owner/repo", "--repo"} {
			if _, err := (realGH{}).LandingPRs(context.Background(), repo, "issue"); err == nil {
				t.Fatalf("accepted repo %q", repo)
			}
		}
		if err := (realGH{}).LandingMerge(context.Background(), "example/repo", 0, "short"); err == nil {
			t.Fatal("accepted malformed merge request")
		}
	})
}

// Opt-in, read-only: select a branch with an existing merged PR (e.g. PR130).
// It exercises the production parser/adapter; it never invokes LandingMerge.
func TestLandingGHLiveConformance(t *testing.T) {
	repo, branch := os.Getenv("SDLC_LANDING_LIVE_REPO"), os.Getenv("SDLC_LANDING_LIVE_BRANCH")
	if repo == "" || branch == "" {
		t.Skip("set SDLC_LANDING_LIVE_REPO and SDLC_LANDING_LIVE_BRANCH for read-only GitHub conformance")
	}
	records, err := (realGH{}).LandingPRs(context.Background(), repo, branch)
	if err != nil {
		t.Fatal(err)
	}
	expected := os.Getenv("SDLC_LANDING_LIVE_PR")
	for _, pr := range records {
		if pr.State == "MERGED" && (expected == "" || strconv.Itoa(pr.Number) == expected) {
			if pr.HeadOID == "" || pr.BaseOID == "" || pr.MergeOID == "" {
				t.Fatal("missing immutable integration evidence")
			}
			t.Logf("merged PR %s#%d head=%s base=%s merge=%s", pr.Repo, pr.Number, pr.HeadOID, pr.BaseOID, pr.MergeOID)
			return
		}
	}
	t.Fatal("no expected merged PR evidence returned")
}

func TestLandingGHOID(t *testing.T) {
	for _, tc := range []struct {
		oid   string
		valid bool
	}{
		{strings.Repeat("a", 40), true}, {strings.Repeat("b", 64), true},
		{"", false}, {strings.Repeat("0", 40), false}, {strings.Repeat("g", 40), false},
		{strings.Repeat("A", 40), false}, {strings.Repeat("a", 39), false},
	} {
		if got := landingOIDValid(tc.oid); got != tc.valid {
			t.Errorf("OID %q valid=%v want %v", tc.oid, got, tc.valid)
		}
	}
}

func FuzzLandingGHParser(f *testing.F) {
	valid, err := json.Marshal([]map[string]any{landingGHRecord()})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Add([]byte("[]"))
	f.Add([]byte("null"))
	f.Add([]byte("[{\"number\":130}]"))
	f.Fuzz(func(t *testing.T, data []byte) {
		records, err := parseLandingPRs(data, "example/repo", "000246-landing")
		if err != nil {
			return
		}
		if len(records) > landingPRLimit {
			t.Fatal("accepted excessive evidence")
		}
		seen := map[int]bool{}
		for _, pr := range records {
			if pr.Number <= 0 || seen[pr.Number] || pr.Repo != "example/repo" || pr.HeadRef != "000246-landing" || pr.BaseRef != "main" || !landingOIDValid(pr.HeadOID) || !landingOIDValid(pr.BaseOID) {
				t.Fatalf("accepted unbound evidence: %+v", pr)
			}
			seen[pr.Number] = true
			if pr.State != "OPEN" && pr.State != "CLOSED" && pr.State != "MERGED" {
				t.Fatalf("accepted unknown state: %+v", pr)
			}
			if pr.State == "MERGED" && !landingOIDValid(pr.MergeOID) {
				t.Fatalf("accepted incomplete integration: %+v", pr)
			}
		}
	})
}
