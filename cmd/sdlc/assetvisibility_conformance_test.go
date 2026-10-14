package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// TestLiveAssetVisibilityConformance checks the installed gh CLI and live host
// against the production visibility seam, without reading or changing policy.
// Normal suites skip it. Run after changing RepoVisibility or upgrading gh:
//
//	SDLC_LIVE_ASSET_VISIBILITY_CONFORMANCE=1 \
//	SDLC_LIVE_ASSET_VISIBILITY_GH=/absolute/path/to/gh \
//	SDLC_LIVE_ASSET_VISIBILITY_REPO=owner/repository \
//	SDLC_LIVE_ASSET_VISIBILITY_EXPECT=PUBLIC \
//	go test ./cmd/sdlc -run '^TestLiveAssetVisibilityConformance$' -count=1 -v
//
// EXPECT is the repository's independently known host visibility. A private
// fixture can use PRIVATE; the caller supplies existing read-only credentials.
// Opting in makes missing prerequisites and host failures test failures, never
// skips, so a requested live verification cannot silently report success.
func TestLiveAssetVisibilityConformance(t *testing.T) {
	if os.Getenv("SDLC_LIVE_ASSET_VISIBILITY_CONFORMANCE") != "1" {
		t.Skip("set SDLC_LIVE_ASSET_VISIBILITY_CONFORMANCE=1, SDLC_LIVE_ASSET_VISIBILITY_REPO and SDLC_LIVE_ASSET_VISIBILITY_EXPECT to query the live host")
	}
	repo := os.Getenv("SDLC_LIVE_ASSET_VISIBILITY_REPO")
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9][A-Za-z0-9_.-]*$`).MatchString(repo) {
		t.Fatal("invalid live repository: set SDLC_LIVE_ASSET_VISIBILITY_REPO to owner/repository")
	}
	want := os.Getenv("SDLC_LIVE_ASSET_VISIBILITY_EXPECT")
	switch want {
	case "PUBLIC", "PRIVATE", "INTERNAL":
	default:
		t.Fatal("missing expected host visibility: set SDLC_LIVE_ASSET_VISIBILITY_EXPECT to the independently known PUBLIC, PRIVATE or INTERNAL value")
	}
	ghPath := os.Getenv("SDLC_LIVE_ASSET_VISIBILITY_GH")
	if !filepath.IsAbs(ghPath) {
		t.Fatal("installed gh path required: set SDLC_LIVE_ASSET_VISIBILITY_GH to the absolute path of the real gh executable")
	}
	resolved, err := filepath.EvalSymlinks(ghPath)
	if err != nil {
		t.Fatal("installed gh unavailable: repair SDLC_LIVE_ASSET_VISIBILITY_GH before rerunning")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		t.Fatal("installed gh is not executable: point SDLC_LIVE_ASSET_VISIBILITY_GH at the real gh executable")
	}
	// The production seam invokes gh through PATH. Put only the explicitly
	// selected binary first, so fake-gh directories from a surrounding harness
	// cannot satisfy this live check. Auth remains the caller's existing auth.
	bin := t.TempDir()
	if err := os.Symlink(resolved, filepath.Join(bin, "gh")); err != nil {
		t.Fatalf("prepare live gh path: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	// Deliberately conflict with the destination; the explicit repository URL
	// must select github.com even when the session defaults to another host.
	t.Setenv("GH_HOST", "visibility-conformance.invalid")
	t.Setenv("GH_PROMPT_DISABLED", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got, err := (realGH{}).RepoVisibility(ctx, "https://github.com/"+repo)
	if err != nil {
		t.Fatalf("live RepoVisibility failed (%v): check gh authentication and repository metadata access, then rerun", err)
	}
	if got != want {
		t.Fatalf("live visibility = %q, expected %q: verify the repository's actual host visibility and reconcile the expectation or production adapter", got, want)
	}
	t.Logf("production RepoVisibility ignores conflicting GH_HOST: %s is %s", repo, got)
}
