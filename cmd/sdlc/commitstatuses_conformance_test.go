package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestLiveCommitStatusesConformance exercises the production adapter against
// existing GitHub history. It only reads: it never posts statuses or changes a PR.
// Run after changing CommitStatuses/parseCommitStatuses or upgrading gh:
//
//	SDLC_LIVE_GH=1 SDLC_LIVE_COMMIT_STATUSES_GH="$(command -v gh)" \
//	go test ./cmd/sdlc -run '^TestLiveCommitStatusesConformance$' -count=1 -v
//
// Fixture: xianxu/ariadne PR #229, independently inspected on 2026-10-14.
// The same context went pending → failure → pending → success. Expectations
// below are fixed evidence, never derived from the adapter's answer. If someone
// appends statuses, inspect that history and deliberately refresh the fixture.
func TestLiveCommitStatusesConformance(t *testing.T) {
	if os.Getenv("SDLC_LIVE_GH") != "1" {
		t.Skip("set SDLC_LIVE_GH=1 and SDLC_LIVE_COMMIT_STATUSES_GH to read live commit statuses")
	}
	ghPath := os.Getenv("SDLC_LIVE_COMMIT_STATUSES_GH")
	if !filepath.IsAbs(ghPath) {
		t.Fatal("set SDLC_LIVE_COMMIT_STATUSES_GH to the absolute path of the installed gh executable")
	}
	resolved, err := filepath.EvalSymlinks(ghPath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		t.Fatal("SDLC_LIVE_COMMIT_STATUSES_GH must name an executable regular file")
	}
	// The harness may have fake gh directories in PATH. Bind both reads to the
	// explicitly selected installed CLI while retaining the caller's read auth.
	bin := t.TempDir()
	if err := os.Symlink(resolved, filepath.Join(bin, "gh")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_HOST", "github.com")
	t.Setenv("GH_PROMPT_DISABLED", "1")
	const repo = "xianxu/ariadne"
	const sha = "38ca18d3d9e95151f2ad3659e278ca45fe954b07"
	const statusContext = "local-gates-guard/verdict"
	type historyRow struct {
		ID      int64  `json:"id"`
		Context string `json:"context"`
		State   string `json:"state"`
	}
	want := []historyRow{
		{56288871627, statusContext, "success"}, // 2026-10-14T02:43:58Z
		{56288870476, statusContext, "pending"}, // 2026-10-14T02:43:56Z
		{56288093629, statusContext, "failure"}, // 2026-10-14T02:26:55Z
		{56288092590, statusContext, "pending"}, // 2026-10-14T02:26:54Z
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, resolved, "api", "repos/"+repo+"/commits/"+sha+"/statuses?per_page=100")
	cmd.WaitDelay = time.Second
	raw, err := cmd.Output()
	if err != nil {
		t.Fatalf("read independent live status history: %v", err)
	}
	var history []historyRow
	if err := json.Unmarshal(raw, &history); err != nil {
		t.Fatal(err)
	}
	var observed []historyRow
	for _, row := range history {
		if row.Context == statusContext {
			observed = append(observed, row)
		}
	}
	if len(observed) != len(want) {
		t.Fatalf("live fixture history changed: got %+v; inspect PR #229 history before updating fixed expectations", observed)
	}
	for i := range want {
		if observed[i] != want[i] {
			t.Fatalf("live history row %d = %+v, want %+v (newest first)", i, observed[i], want[i])
		}
	}
	got, err := (realGH{}).CommitStatuses(ctx, repo, sha)
	if err != nil {
		t.Fatal(err)
	}
	if got[statusContext] != "success" {
		t.Fatalf("production CommitStatuses returned %q, want newest success over older failure and pending statuses", got[statusContext])
	}
	t.Logf("%s@%s: four independently pinned status events; production adapter returns newest success", repo, sha)
}
