package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
)

func TestFleetEndToEnd(t *testing.T) {
	isolateFleetIntegrationGit(t)
	binary := buildFleetE2EBinary(t)

	fleetRoot := t.TempDir()
	alpha := newFleetE2ERepo(t, fleetRoot, "alpha", `{
  "version": 1,
  "admission": {
    "key": {"kind": "repo", "roots": []},
    "capacity": {"kind": "bounded", "limit": 1},
    "onCapacity": "reject"
  }
}`)
	peer := newFleetE2ERepo(t, fleetRoot, "peer", "")

	linked := filepath.Join(fleetRoot, "worktree", "alpha-linked")
	if err := os.MkdirAll(filepath.Dir(linked), 0o755); err != nil {
		t.Fatal(err)
	}
	runFleetE2EGit(t, alpha, "worktree", "add", "-b", "feature", linked)

	alphaNested := filepath.Join(alpha, "nested")
	linkedNested := filepath.Join(linked, "nested")
	for _, dir := range []string{alphaNested, linkedNested} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	broken := filepath.Join(fleetRoot, "broken")
	if err := os.MkdirAll(filepath.Join(broken, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	canonicalAlpha := canonicalFleetE2EPath(t, alpha)
	canonicalLinked := canonicalFleetE2EPath(t, linked)
	canonicalPeer := canonicalFleetE2EPath(t, peer)
	canonicalBroken := canonicalFleetE2EPath(t, broken)
	alphaIdentity := canonicalFleetE2EPath(t, filepath.Join(alpha, ".git"))

	vantages := []string{alpha, alphaNested, linked, linkedNested, peer}
	alias := filepath.Join(fleetRoot, "alpha-alias")
	if err := os.Symlink(alpha, alias); err != nil {
		if !fleetE2ESymlinkUnavailable(err) {
			t.Fatalf("create symlink vantage %q -> %q: %v", alias, alpha, err)
		}
		t.Logf("symlink vantage unavailable; continuing all other end-to-end cases: %v", err)
	} else {
		vantages = append(vantages, alias, filepath.Join(alias, "nested"))
	}
	var baseline fleet.Inventory
	var baselineJSON string
	for i, vantage := range vantages {
		got, raw := runFleetE2EInventory(t, binary, vantage)
		if i == 0 {
			baseline = got
			baselineJSON = raw
			continue
		}
		if !reflect.DeepEqual(got, baseline) {
			t.Fatalf("inventory from %q differs from primary-root inventory\nprimary: %#v\ncurrent: %#v", vantage, baseline, got)
		}
		if raw != baselineJSON {
			t.Fatalf("inventory JSON from %q is not byte-stable\nprimary: %s\ncurrent: %s", vantage, baselineJSON, raw)
		}
	}

	if len(baseline.Rows) != 3 {
		t.Fatalf("inventory rows = %d, want primary + linked + peer rows: %#v", len(baseline.Rows), baseline.Rows)
	}
	for i := 1; i < len(baseline.Rows); i++ {
		previous := baseline.Rows[i-1].RepoIdentity + "\x00" + baseline.Rows[i-1].TreePath
		current := baseline.Rows[i].RepoIdentity + "\x00" + baseline.Rows[i].TreePath
		if previous >= current {
			t.Fatalf("inventory rows are not strictly ordered by repo identity and tree path: %#v", baseline.Rows)
		}
	}

	healthyTrees := map[string]bool{
		canonicalAlpha:  false,
		canonicalLinked: false,
	}
	peerRows := 0
	for _, row := range baseline.Rows {
		if row.RepoIdentity == alphaIdentity {
			if _, ok := healthyTrees[row.TreePath]; !ok {
				t.Fatalf("unexpected healthy worktree path %q", row.TreePath)
			}
			healthyTrees[row.TreePath] = true
			assertFleetE2EHealthyRow(t, row, canonicalAlpha)
			continue
		}
		if row.RepoRoot != canonicalPeer || row.TreePath != canonicalPeer {
			t.Fatalf("unexpected unaffected row: %#v", row)
		}
		peerRows++
		if !row.Facts.Available || row.Facts.Head == "" {
			t.Fatalf("peer facts were lost because its declaration is missing: %#v", row.Facts)
		}
		if row.Policy.OK || row.Policy.Diagnostic == nil || row.Policy.Diagnostic.Code != fleet.DiagnosticMissingPolicy {
			t.Fatalf("peer policy capability = %#v, want localized missing-policy diagnostic", row.Policy)
		}
	}
	for tree, seen := range healthyTrees {
		if !seen {
			t.Errorf("healthy worktree %q missing from inventory", tree)
		}
	}
	if peerRows != 1 {
		t.Fatalf("peer rows = %d, want 1", peerRows)
	}
	if len(baseline.Diagnostics) != 1 {
		t.Fatalf("top-level diagnostics = %#v, want one malformed-repository diagnostic", baseline.Diagnostics)
	}
	if diagnostic := baseline.Diagnostics[0]; diagnostic.RepoPath != canonicalBroken || diagnostic.Stage != "git" {
		t.Fatalf("malformed-repository diagnostic = %#v, want repo_path %q and git stage", diagnostic, canonicalBroken)
	}

	primaryResult := runFleetE2EPolicy(t, binary, filepath.Join(alphaNested, "prospective", "task"))
	linkedResult := runFleetE2EPolicy(t, binary, filepath.Join(linkedNested, "prospective", "task"))
	if !reflect.DeepEqual(primaryResult, linkedResult) {
		t.Fatalf("repo-key policy differs across equivalent primary and linked prospective paths\nprimary: %#v\nlinked: %#v", primaryResult, linkedResult)
	}
	if !primaryResult.OK || primaryResult.Value == nil || primaryResult.Diagnostic != nil {
		t.Fatalf("healthy policy result = %#v, want success", primaryResult)
	}
	value := primaryResult.Value
	if value.RepoIdentity != alphaIdentity || value.AdmissionKey != alphaIdentity {
		t.Fatalf("repo-key identities = repo %q, key %q; want %q", value.RepoIdentity, value.AdmissionKey, alphaIdentity)
	}
	if value.PolicyVersion != 1 || len(value.PolicyDigest) != 64 || value.Capacity.Kind != "bounded" || value.Capacity.Limit == nil || *value.Capacity.Limit != 1 || value.OnCapacity != "reject" {
		t.Fatalf("resolved policy metadata = %#v, want committed bounded repo policy", value)
	}

	assertFleetE2EPolicyRefusal(t, binary, filepath.Join(peer, "prospective", "task"), fleet.DiagnosticMissingPolicy)
	writeFleetIntegrationPolicy(t, peer, `{not-json`)
	assertFleetE2EPolicyRefusal(t, binary, filepath.Join(peer, "prospective", "task"), fleet.DiagnosticInvalidPolicy)

	writeFleetIntegrationPolicy(t, alpha, `{
  "version": 1,
  "admission": {
    "key": {"kind": "declared-root", "roots": ["apps/*"]},
    "capacity": {"kind": "bounded", "limit": 1},
    "onCapacity": "reject"
  }
}`)
	assertFleetE2EPolicyRefusal(t, binary, filepath.Join(alpha, "outside", "prospective", "task"), fleet.DiagnosticOutsideDeclaredScope)
}

func TestFleetE2ERefusalValidationRejectsCrashLikeProcess(t *testing.T) {
	diagnosticCode := fleet.DiagnosticMissingPolicy
	raw, err := json.Marshal(fleet.PolicyResult{Diagnostic: &fleet.PolicyDiagnostic{
		Code: diagnosticCode, Message: "fleet policy declaration is missing", Path: "/repo/.sdlc/fleet.json",
	}})
	if err != nil {
		t.Fatal(err)
	}
	stdout := string(raw) + "\n"
	expectedStderr := "Error: fleet policy refused: " + diagnosticCode + "\n"

	tests := []struct {
		name     string
		exitCode int
		stderr   string
		wantErr  bool
	}{
		{name: "exact refusal", exitCode: 1, stderr: expectedStderr},
		{name: "crash exit", exitCode: 2, stderr: expectedStderr, wantErr: true},
		{name: "panic output", exitCode: 1, stderr: "panic: runtime failure\n", wantErr: true},
		{name: "unrelated suffix", exitCode: 1, stderr: expectedStderr + "unrelated\n", wantErr: true},
		{name: "usage output", exitCode: 1, stderr: expectedStderr + "Usage: sdlc fleet policy\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFleetE2ERefusal(stdout, tt.stderr, tt.exitCode, diagnosticCode)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate refusal error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newFleetE2ERepo(t *testing.T, fleetRoot, name, policy string) string {
	t.Helper()

	repo := filepath.Join(fleetRoot, name)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	runFleetE2EGit(t, repo, "init", "-b", "main")
	runFleetE2EGit(t, repo, "config", "user.name", "Fleet End-to-End")
	runFleetE2EGit(t, repo, "config", "user.email", "fleet-e2e@example.invalid")
	runFleetE2EGit(t, repo, "config", "commit.gpgsign", "false")
	runFleetE2EGit(t, repo, "config", "core.hooksPath", os.DevNull)
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte(name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if policy != "" {
		writeFleetIntegrationPolicy(t, repo, policy)
	}
	runFleetE2EGit(t, repo, "add", ".")
	runFleetE2EGit(t, repo, "commit", "-m", "initial")
	runFleetE2EGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	return repo
}

func runFleetE2EInventory(t *testing.T, binary, vantage string) (fleet.Inventory, string) {
	t.Helper()

	stdout, stderr, err := runFleetE2ECommand(t, binary, "fleet", "inventory", "--json", "--path", vantage)
	if err != nil {
		t.Fatalf("fleet inventory --path %q: %v\nstderr: %s\nstdout: %s", vantage, err, stderr, stdout)
	}
	if stderr != "" {
		t.Fatalf("fleet inventory --path %q stderr = %q, want empty", vantage, stderr)
	}
	var inventory fleet.Inventory
	if err := json.Unmarshal([]byte(stdout), &inventory); err != nil {
		t.Fatalf("decode fleet inventory --path %q: %v\nstdout: %s", vantage, err, stdout)
	}
	return inventory, stdout
}

func runFleetE2EPolicy(t *testing.T, binary, requested string) fleet.PolicyResult {
	t.Helper()

	stdout, stderr, err := runFleetE2ECommand(t, binary, "fleet", "policy", "--json", "--path", requested)
	if err != nil {
		t.Fatalf("fleet policy --path %q: %v\nstderr: %s\nstdout: %s", requested, err, stderr, stdout)
	}
	if stderr != "" {
		t.Fatalf("fleet policy --path %q stderr = %q, want empty", requested, stderr)
	}
	var result fleet.PolicyResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode fleet policy --path %q: %v\nstdout: %s", requested, err, stdout)
	}
	return result
}

func assertFleetE2EPolicyRefusal(t *testing.T, binary, requested, diagnosticCode string) {
	t.Helper()

	stdout, stderr, err := runFleetE2ECommand(t, binary, "fleet", "policy", "--json", "--path", requested)
	if err == nil {
		t.Fatalf("fleet policy --path %q succeeded; want %s refusal\nstdout: %s", requested, diagnosticCode, stdout)
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("fleet policy --path %q error = %T %v, want nonzero process exit", requested, err, err)
	}
	if validationErr := validateFleetE2ERefusal(stdout, stderr, exitError.ExitCode(), diagnosticCode); validationErr != nil {
		t.Fatalf("fleet policy --path %q invalid %s refusal: %v", requested, diagnosticCode, validationErr)
	}
}

func validateFleetE2ERefusal(stdout, stderr string, exitCode int, diagnosticCode string) error {
	if exitCode != 1 {
		return fmt.Errorf("exit code = %d, want 1", exitCode)
	}
	expectedStderr := "Error: fleet policy refused: " + diagnosticCode + "\n"
	if stderr != expectedStderr {
		return fmt.Errorf("stderr = %q, want exactly %q", stderr, expectedStderr)
	}
	if stdout == "" {
		return errors.New("typed stdout is empty")
	}
	var result fleet.PolicyResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return fmt.Errorf("decode typed stdout: %w", err)
	}
	if result.OK || result.Value != nil || result.Diagnostic == nil || result.Diagnostic.Code != diagnosticCode {
		return fmt.Errorf("typed result = %#v, want %s refusal", result, diagnosticCode)
	}
	return nil
}

func buildFleetE2EBinary(t *testing.T) string {
	t.Helper()

	name := "sdlc"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	result := runFleetE2ESubprocess(t, time.Minute, "", true, "go", "build", "-o", binary, ".")
	if result.err != nil {
		t.Fatalf("build fleet end-to-end binary: %v\n%s", result.err, result.combined)
	}
	return binary
}

func runFleetE2ECommand(t *testing.T, binary string, args ...string) (string, string, error) {
	t.Helper()

	result := runFleetE2ESubprocess(t, 30*time.Second, "", false, binary, args...)
	return result.stdout, result.stderr, result.err
}

func runFleetE2EGit(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	result := runFleetE2ESubprocess(t, 15*time.Second, dir, true, "git", args...)
	if result.err != nil {
		t.Fatalf("git -C %q %v: %v\n%s", dir, args, result.err, result.combined)
	}
	return []byte(result.combined)
}

type fleetE2ESubprocessResult struct {
	stdout   string
	stderr   string
	combined string
	err      error
}

func runFleetE2ESubprocess(t *testing.T, timeout time.Duration, dir string, combined bool, executable string, args ...string) fleetE2ESubprocessResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = dir
	command.Env = os.Environ()
	var result fleetE2ESubprocessResult
	if combined {
		output, err := command.CombinedOutput()
		result.combined = string(output)
		result.err = err
	} else {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		result.err = command.Run()
		result.stdout = stdout.String()
		result.stderr = stderr.String()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.err = fmt.Errorf("%s exceeded %s deadline: %w", executable, timeout, ctx.Err())
	}
	return result
}

func fleetE2ESymlinkUnavailable(err error) bool {
	if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.ENOSYS) {
		return true
	}
	var errno syscall.Errno
	return runtime.GOOS == "windows" && errors.As(err, &errno) && errno == 1314
}

func assertFleetE2EHealthyRow(t *testing.T, row fleet.TreeRow, repoRoot string) {
	t.Helper()

	if row.RepoRoot != repoRoot {
		t.Errorf("healthy row repo root = %q, want %q", row.RepoRoot, repoRoot)
	}
	if row.Bare || row.Detached || row.Branch == "" {
		t.Errorf("healthy row branch state is incomplete: %#v", row)
	}
	if !row.Facts.Available || row.Facts.Head == "" || row.Facts.CommitTimestamp == "" || !row.Facts.BaseAvailable || row.Facts.BaseRef != "origin/main" || row.Facts.Ahead == nil || *row.Facts.Ahead != 0 || row.Facts.Behind == nil || *row.Facts.Behind != 0 || row.Facts.DirtyCount == nil || *row.Facts.DirtyCount != 0 {
		t.Errorf("healthy row facts are incomplete: %#v", row.Facts)
	}
	if row.Issues == nil {
		t.Errorf("healthy row issues must be non-nil")
	}
	if !row.Policy.OK || row.Policy.Value == nil || row.Policy.Diagnostic != nil || row.Policy.Value.KeyKind != "repo" {
		t.Errorf("healthy row policy capability is incomplete: %#v", row.Policy)
	}
}

func canonicalFleetE2EPath(t *testing.T, path string) string {
	t.Helper()

	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("canonicalize %q: %v", path, err)
	}
	absolute, err := filepath.Abs(canonical)
	if err != nil {
		t.Fatalf("make %q absolute: %v", canonical, err)
	}
	return filepath.Clean(absolute)
}
