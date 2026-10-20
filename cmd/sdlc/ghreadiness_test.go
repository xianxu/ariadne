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

func readinessRecord() map[string]any {
	return map[string]any{
		"number": 130, "state": "OPEN", "headRefName": "000472-drift", "headRefOid": strings.Repeat("a", 40),
		"baseRefName": "main", "baseRefOid": strings.Repeat("b", 40),
		"baseRef":        map[string]any{"name": "main", "target": map[string]any{"oid": strings.Repeat("c", 40)}},
		"headRepository": map[string]any{"nameWithOwner": "example/repo"}, "isCrossRepository": false,
		"mergeStateStatus": "CLEAN",
	}
}
func readinessJSON(t *testing.T, record map[string]any) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{"data": map[string]any{"repository": map[string]any{"nameWithOwner": "example/repo", "pullRequest": record}}})
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func TestLandingReadinessParse(t *testing.T) {
	got, err := parseLandingReadiness(readinessJSON(t, readinessRecord()), "example/repo", 130)
	want := landingReadiness{Number: 130, Repo: "example/repo", State: "OPEN", HeadRef: "000472-drift", HeadOID: strings.Repeat("a", 40), BaseRef: "main", BaseTargetOID: strings.Repeat("c", 40), MergeStateStatus: "CLEAN"}
	if err != nil || got != want {
		t.Fatalf("got %+v, %v; want %+v (live target, not historical baseRefOid)", got, err, want)
	}
	for _, state := range []string{"BEHIND", "UNKNOWN", "FUTURE_STATE", ""} {
		r := readinessRecord()
		r["mergeStateStatus"] = state
		got, err := parseLandingReadiness(readinessJSON(t, r), "example/repo", 130)
		if err != nil || got.MergeStateStatus != state {
			t.Fatalf("state %q: %+v %v", state, got, err)
		}
	}
	r := readinessRecord()
	delete(r, "mergeStateStatus")
	if got, err := parseLandingReadiness(readinessJSON(t, r), "example/repo", 130); err != nil || got.MergeStateStatus != "" {
		t.Fatalf("missing state: %+v %v", got, err)
	}
}
func TestLandingReadinessRejectsMalformedEvidence(t *testing.T) {
	cases := map[string]func(map[string]any){
		"number":             func(r map[string]any) { r["number"] = 131 },
		"number-type":        func(r map[string]any) { r["number"] = "130" },
		"closed":             func(r map[string]any) { r["state"] = "CLOSED" },
		"merged":             func(r map[string]any) { r["state"] = "MERGED" },
		"head-repo":          func(r map[string]any) { r["headRepository"] = map[string]any{"nameWithOwner": "fork/repo"} },
		"null-head-repo":     func(r map[string]any) { r["headRepository"] = nil },
		"cross-repo":         func(r map[string]any) { r["isCrossRepository"] = true },
		"missing-cross-repo": func(r map[string]any) { delete(r, "isCrossRepository") },
		"short-head":         func(r map[string]any) { r["headRefOid"] = "abcd" },
		"null-head":          func(r map[string]any) { r["headRefOid"] = nil },
		"head-branch":        func(r map[string]any) { r["headRefName"] = "bad..branch" },
		"base-branch":        func(r map[string]any) { r["baseRefName"] = "other" },
		"null-base":          func(r map[string]any) { r["baseRef"] = nil },
		"base-name": func(r map[string]any) {
			r["baseRef"] = map[string]any{"name": "other", "target": map[string]any{"oid": strings.Repeat("c", 40)}}
		},
		"short-live-base": func(r map[string]any) {
			r["baseRef"] = map[string]any{"name": "main", "target": map[string]any{"oid": "abcd"}}
		},
		"missing-live-base": func(r map[string]any) { r["baseRef"] = map[string]any{"name": "main", "target": nil} },
		"state-type":        func(r map[string]any) { r["mergeStateStatus"] = true },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			r := readinessRecord()
			change(r)
			if _, err := parseLandingReadiness(readinessJSON(t, r), "example/repo", 130); err == nil {
				t.Fatal("accepted malformed evidence")
			}
		})
	}
	valid := string(readinessJSON(t, readinessRecord()))
	for _, data := range []string{"null", "{}", `{"data":null}`, `{"data":{"repository":null}}`, valid[:len(valid)-1], valid + "{}", strings.Replace(valid, `"example/repo"`, `"other/repo"`, 1), strings.Replace(valid, `"data":`, `"errors":[{"message":"partial failure"}],"data":`, 1), strings.Repeat(" ", 65537) + valid} {
		if _, err := parseLandingReadiness([]byte(data), "example/repo", 130); err == nil {
			t.Fatalf("accepted malformed envelope: %.100s", data)
		}
	}
}
func TestLandingReadinessDispatch(t *testing.T) {
	dir := fakeLandingGH(t, "printf '%s\\n' \"$@\" > \"$LANDING_ARGV\"\ncat \"$LANDING_RESPONSE\"\n")
	argsPath := filepath.Join(dir, "argv")
	response := filepath.Join(dir, "response")
	t.Setenv("LANDING_ARGV", argsPath)
	t.Setenv("LANDING_RESPONSE", response)
	t.Setenv("GH_HOST", "untrusted.example")
	if err := os.WriteFile(response, readinessJSON(t, readinessRecord()), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := (realGH{}).LandingReadiness(context.Background(), "example/repo", 130); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"api", "graphql", "--hostname", "github.com", "-f", "query=" + landingReadinessQuery, "-f", "owner=example", "-f", "name=repo", "-F", "number=130"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("query argv: %q", got)
	}
	if strings.Contains(landingReadinessQuery, "baseRefOid") || !strings.Contains(landingReadinessQuery, "baseRef { name target { oid } }") {
		t.Fatal("query must read live base reference")
	}
}
func TestLandingReadinessRequestFailures(t *testing.T) {
	t.Run("invalid-input", func(t *testing.T) {
		fakeLandingGH(t, "echo should-not-run >&2\nexit 99\n")
		for _, repo := range []string{"", "--repo", "example/repo/extra", strings.Repeat("a", 300) + "/repo"} {
			_, err := (realGH{}).LandingReadiness(context.Background(), repo, 130)
			if err == nil || strings.Contains(err.Error(), "99") {
				t.Fatalf("invalid request dispatched: %q %v", repo, err)
			}
		}
		if _, err := (realGH{}).LandingReadiness(context.Background(), "example/repo", 0); err == nil || strings.Contains(err.Error(), "99") {
			t.Fatalf("invalid PR dispatched: %v", err)
		}
	})
	t.Run("failure", func(t *testing.T) {
		fakeLandingGH(t, "echo unavailable >&2\nexit 1\n")
		if _, err := (realGH{}).LandingReadiness(context.Background(), "example/repo", 130); err == nil || !strings.Contains(err.Error(), "unavailable") {
			t.Fatalf("lost error: %v", err)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		fakeLandingGH(t, "exec sleep 30\n")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		if _, err := (realGH{}).LandingReadiness(ctx, "example/repo", 130); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("cancel: %v", err)
		}
	})
}

func TestLandingReadinessOutputBound(t *testing.T) {
	for _, stream := range []string{"stdout", "stderr"} {
		t.Run(stream, func(t *testing.T) {
			body := "awk 'BEGIN { for (i=0;i<70000;i++) printf \"x\" }'"
			if stream == "stderr" {
				body += " >&2"
			}
			fakeLandingGH(t, body+"\n")
			if _, err := (realGH{}).LandingReadiness(context.Background(), "example/repo", 130); err == nil || !strings.Contains(err.Error(), "byte limit") {
				t.Fatalf("unbounded %s: %v", stream, err)
			}
		})
	}
}

// Opt-in read-only conformance; never calls a model or mutates a real PR.
func TestLandingReadinessLive(t *testing.T) {
	repo, rawNumber := os.Getenv("WF_READINESS_LIVE_REPO"), os.Getenv("WF_READINESS_LIVE_PR")
	if repo == "" || rawNumber == "" {
		t.Skip("set WF_READINESS_LIVE_REPO and WF_READINESS_LIVE_PR for read-only conformance")
	}
	number, err := strconv.Atoi(rawNumber)
	if err != nil {
		t.Fatal(err)
	}
	got, err := (realGH{}).LandingReadiness(context.Background(), repo, number)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("live readiness: repo=%s PR=%d head=%s base=%s status=%s", got.Repo, got.Number, got.HeadOID, got.BaseTargetOID, got.MergeStateStatus)
}
