package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
)

func TestFleetInventory_RealGitPortableFleet(t *testing.T) {
	isolateFleetIntegrationGit(t)
	fleetRoot := t.TempDir()
	alpha := filepath.Join(fleetRoot, "alpha")
	zeta := filepath.Join(fleetRoot, "zeta")
	bare := filepath.Join(fleetRoot, "bare.git")
	linked := filepath.Join(fleetRoot, "worktree", "alpha-linked")
	for _, repo := range []string{zeta, alpha} {
		runFleetIntegrationGit(t, fleetRoot, "init", "-b", "main", repo)
		runFleetIntegrationGit(t, repo, "config", "user.name", "Fleet Integration")
		runFleetIntegrationGit(t, repo, "config", "user.email", "fleet@example.test")
		runFleetIntegrationGit(t, repo, "config", "commit.gpgsign", "false")
		runFleetIntegrationGit(t, repo, "config", "core.hooksPath", os.DevNull)
		if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("initial\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		runFleetIntegrationGit(t, repo, "add", "tracked.txt")
		runFleetIntegrationGit(t, repo, "commit", "-m", "initial")
		runFleetIntegrationGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	}
	runFleetIntegrationGit(t, fleetRoot, "clone", "--bare", alpha, bare)
	if err := os.MkdirAll(filepath.Dir(linked), 0o755); err != nil {
		t.Fatal(err)
	}
	runFleetIntegrationGit(t, alpha, "worktree", "add", "-b", "000123-feature", linked)
	if err := os.Mkdir(filepath.Join(fleetRoot, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(fleetRoot, "broken", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFleetIntegrationPolicy(t, alpha, `{"version":1,"admission":{"key":{"kind":"worktree","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"provision-worktree"}}`)
	writeFleetIntegrationPolicy(t, bare, `{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"unbounded"}}}`)
	writeFleetIntegrationPolicy(t, zeta, `{"version":1,"admission":{"key":{"kind":"repo","roots":["forbidden/*"]},"capacity":{"kind":"unbounded"}}}`)
	writeFleetIntegrationIssue(t, alpha, "000123-feature.md", "working")

	options := fleet.InventoryOptions{Git: execGitRunner{}}
	first, err := fleet.CollectInventory(fleetRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	if first.Rows == nil || first.Diagnostics == nil || len(first.Rows) != 4 {
		t.Fatalf("real inventory = %+v, want four rows and non-nil collections", first)
	}
	brokenPath, err := filepath.EvalSymlinks(filepath.Join(fleetRoot, "broken"))
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Diagnostics) != 2 || first.Diagnostics[0].Stage != "git" || first.Diagnostics[0].RepoPath != brokenPath || first.Diagnostics[1].Stage != "facts" {
		t.Fatalf("real diagnostics = %+v, want isolated broken-repo Git and bare-facts diagnostics", first.Diagnostics)
	}
	for i := 1; i < len(first.Rows); i++ {
		previous, current := first.Rows[i-1], first.Rows[i]
		if previous.RepoIdentity > current.RepoIdentity || previous.RepoIdentity == current.RepoIdentity && previous.TreePath > current.TreePath {
			t.Fatalf("real rows are not stable sorted: %+v", first.Rows)
		}
	}
	linkedRow := findFleetIntegrationRow(t, first, linked)
	if len(linkedRow.Issues) != 1 || linkedRow.Issues[0].Ref != "alpha#000123" || linkedRow.Issues[0].Provenance != fleet.IssueProvenanceBranchPrefix {
		t.Fatalf("linked row issues = %+v, want alpha#000123 with same-repo provenance", linkedRow.Issues)
	}
	if !linkedRow.Policy.OK || linkedRow.Policy.Value == nil || linkedRow.Policy.Value.KeyKind != "worktree" {
		t.Fatalf("linked row policy = %+v, want declaration capability", linkedRow.Policy)
	}
	zetaRow := findFleetIntegrationRow(t, first, zeta)
	if zetaRow.Policy.OK || zetaRow.Policy.Diagnostic == nil || zetaRow.Policy.Diagnostic.Code != fleet.DiagnosticInvalidPolicy {
		t.Fatalf("zeta policy = %+v, want invalid policy without erased row", zetaRow.Policy)
	}
	bareRow := findFleetIntegrationRow(t, first, bare)
	if !bareRow.Bare || bareRow.Branch != "" || bareRow.Detached || bareRow.Facts.Available || bareRow.Facts.Error == "" {
		t.Fatalf("bare row = %+v, want preserved bare state and explicit unavailable facts", bareRow)
	}

	if err := os.WriteFile(filepath.Join(linked, "untracked.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := fleet.CollectInventory(fleetRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	dirty := findFleetIntegrationRow(t, second, linked).Facts.DirtyCount
	if dirty == nil || *dirty != 1 {
		t.Fatalf("recollected dirty count = %v, want 1", dirty)
	}
}

func isolateFleetIntegrationGit(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		name, _, ok := strings.Cut(entry, "=")
		if !ok || !strings.HasPrefix(name, "GIT_") {
			continue
		}
		value, existed := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(name, value)
			} else {
				_ = os.Unsetenv(name)
			}
		})
	}
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_COUNT", "0")
}

func runFleetIntegrationGit(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	out, err := (execGitRunner{}).GitInDir(dir, args...)
	if err != nil {
		t.Fatalf("git -C %q %s: %v\n%s", dir, strings.Join(args, " "), err, out)
	}
	return out
}

func writeFleetIntegrationPolicy(t *testing.T, repo, raw string) {
	t.Helper()
	dir := filepath.Join(repo, ".sdlc")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fleet.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFleetIntegrationIssue(t *testing.T, repo, name, status string) {
	t.Helper()
	dir := filepath.Join(repo, "workshop", "issues")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "---\nstatus: " + status + "\n---\n# test\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
}

func findFleetIntegrationRow(t *testing.T, inventory fleet.Inventory, path string) fleet.TreeRow {
	t.Helper()
	want, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range inventory.Rows {
		if row.TreePath == want {
			return row
		}
	}
	t.Fatalf("no row for %q: %+v", want, inventory.Rows)
	return fleet.TreeRow{}
}
