package fleet

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeVantage_EquivalentFakeVantages(t *testing.T) {
	fleetRoot := t.TempDir()
	primary := filepath.Join(fleetRoot, "ariadne")
	linked := filepath.Join(fleetRoot, "linked", "ariadne-feature")
	common := filepath.Join(primary, ".git")
	for _, dir := range []string{
		primary,
		filepath.Join(primary, "cmd", "sdlc"),
		linked,
		filepath.Join(linked, "cmd", "sdlc"),
		common,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(fleetRoot, "ariadne-link")
	linkErr := os.Symlink(primary, link)

	reader := &vantageStub{
		roots: map[string]string{
			primary:                               primary,
			filepath.Join(primary, "cmd", "sdlc"): primary,
			linked:                                linked,
			filepath.Join(linked, "cmd", "sdlc"):  linked,
			link:                                  primary,
		},
		worktrees: worktreePorcelain(primary, linked),
		common:    common,
	}

	want := Vantage{
		RepoIdentity: mustCanonicalPath(t, common),
		PrimaryRoot:  mustCanonicalPath(t, primary),
		FleetRoot:    mustCanonicalPath(t, fleetRoot),
	}
	for _, vantage := range []string{
		primary,
		filepath.Join(primary, "cmd", "sdlc"),
		linked,
		filepath.Join(linked, "cmd", "sdlc"),
	} {
		got, err := NormalizeVantage(reader, vantage)
		if err != nil {
			t.Fatalf("NormalizeVantage(%q): %v", vantage, err)
		}
		if got.RepoIdentity != want.RepoIdentity || got.PrimaryRoot != want.PrimaryRoot || got.FleetRoot != want.FleetRoot {
			t.Errorf("NormalizeVantage(%q) stable tuple = %+v, want repo=%q primary=%q fleet=%q", vantage, got, want.RepoIdentity, want.PrimaryRoot, want.FleetRoot)
		}
		if got.WorktreeRoot != mustCanonicalPath(t, reader.roots[vantage]) {
			t.Errorf("NormalizeVantage(%q) worktree root = %q, want %q", vantage, got.WorktreeRoot, mustCanonicalPath(t, reader.roots[vantage]))
		}
	}
	t.Run("symlink", func(t *testing.T) {
		if linkErr != nil {
			t.Skipf("symlinks unavailable: %v", linkErr)
		}
		got, err := NormalizeVantage(reader, link)
		if err != nil {
			t.Fatal(err)
		}
		if got.RepoIdentity != want.RepoIdentity || got.PrimaryRoot != want.PrimaryRoot || got.FleetRoot != want.FleetRoot || got.WorktreeRoot != want.PrimaryRoot {
			t.Errorf("NormalizeVantage(%q) = %+v, want primary symlink vantage", link, got)
		}
	})

	// The nested primary vantage proves that a relative --git-common-dir value
	// is interpreted relative to the command directory rather than its worktree.
	if !reader.relativeCommonSeen {
		t.Error("NormalizeVantage never resolved a relative git-common-dir response")
	}
}

func TestCanonicalProspectivePathResolvesExistingSymlinkAncestorAndMissingSuffix(t *testing.T) {
	real := filepath.Join(t.TempDir(), "real")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	makeFleetTestSymlink(t, real, link)
	requested, containingDir, err := CanonicalProspectivePath(filepath.Join(link, "future", "child"))
	if err != nil {
		t.Fatal(err)
	}
	canonicalReal, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	if requested != filepath.Join(canonicalReal, "future", "child") || containingDir != canonicalReal {
		t.Fatalf("CanonicalProspectivePath = (%q, %q), want (%q, %q)", requested, containingDir, filepath.Join(canonicalReal, "future", "child"), canonicalReal)
	}
}

func TestCanonicalProspectivePathUsesParentForExistingFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "target")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	requested, containingDir, err := CanonicalProspectivePath(file)
	if err != nil {
		t.Fatal(err)
	}
	canonicalFile, err := filepath.EvalSymlinks(file)
	if err != nil {
		t.Fatal(err)
	}
	if requested != canonicalFile || containingDir != filepath.Dir(canonicalFile) {
		t.Fatalf("CanonicalProspectivePath = (%q, %q), want (%q, %q)", requested, containingDir, canonicalFile, filepath.Dir(canonicalFile))
	}
}

func TestCanonicalProspectivePathResolvesSymlinkBeforeParentTraversal(t *testing.T) {
	tempRoot := t.TempDir()
	repo := filepath.Join(tempRoot, "repo")
	outside := filepath.Join(tempRoot, "outside")
	linkTarget := filepath.Join(outside, "nested")
	for _, dir := range []string{repo, linkTarget, filepath.Join(outside, "existing")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(repo, "link")
	makeFleetTestSymlink(t, linkTarget, link)

	separator := string(filepath.Separator)
	for _, tt := range []struct {
		name string
		path string
		want string
		dir  string
	}{
		{"existing", link + separator + ".." + separator + "existing", filepath.Join(outside, "existing"), filepath.Join(outside, "existing")},
		{"prospective", link + separator + ".." + separator + "future" + separator + "child", filepath.Join(outside, "future", "child"), outside},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requested, containingDir, err := CanonicalProspectivePath(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := filepath.EvalSymlinks(tt.want)
			if os.IsNotExist(err) {
				canonicalOutside, resolveErr := filepath.EvalSymlinks(outside)
				if resolveErr != nil {
					t.Fatal(resolveErr)
				}
				want = filepath.Join(canonicalOutside, "future", "child")
			} else if err != nil {
				t.Fatal(err)
			}
			wantDir, err := filepath.EvalSymlinks(tt.dir)
			if err != nil {
				t.Fatal(err)
			}
			if requested != want || containingDir != wantDir {
				t.Fatalf("CanonicalProspectivePath(%q) = (%q, %q), want (%q, %q)", tt.path, requested, containingDir, want, wantDir)
			}
		})
	}
}

func TestCanonicalProspectivePathRejectsDanglingSymlinkAndAmbiguousMissingParent(t *testing.T) {
	repo := t.TempDir()
	dangling := filepath.Join(repo, "dangling")
	makeFleetTestSymlink(t, filepath.Join(repo, "absent-target"), dangling)
	file := filepath.Join(repo, "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	separator := string(filepath.Separator)
	for _, path := range []string{
		dangling,
		dangling + separator + "child",
		filepath.Join(repo, "missing") + separator + ".." + separator + "target",
		file + separator + ".." + separator + "target",
	} {
		if _, _, err := CanonicalProspectivePath(path); err == nil {
			t.Fatalf("CanonicalProspectivePath(%q) succeeded, want fail-closed error", path)
		}
	}
}

func TestCanonicalProspectivePathRejectsDirectorySyntaxAfterFile(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "file-link")
	makeFleetTestSymlink(t, file, link)
	separator := string(filepath.Separator)
	for _, tt := range []struct {
		name string
		path string
	}{
		{"file trailing separator", file + separator},
		{"file dot component", file + separator + "."},
		{"symlink to file trailing separator", link + separator},
		{"symlink to file dot component", link + separator + "."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := CanonicalProspectivePath(tt.path); err == nil {
				t.Fatalf("CanonicalProspectivePath(%q) succeeded, want non-directory traversal error", tt.path)
			}
		})
	}
}

func makeFleetTestSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		if os.IsPermission(err) || strings.Contains(strings.ToLower(err.Error()), "not supported") || strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
			t.Skipf("symlinks unavailable: %v", err)
		}
		t.Fatal(err)
	}
}

func TestNormalizeVantage_EquivalentRealGitVantages(t *testing.T) {
	fleetRoot := t.TempDir()
	primary := filepath.Join(fleetRoot, "ariadne")
	gitAt(t, fleetRoot, "init", primary)
	gitAt(t, primary, "config", "user.email", "fleet-test@example.test")
	gitAt(t, primary, "config", "user.name", "Fleet Test")
	if err := os.WriteFile(filepath.Join(primary, "README.md"), []byte("fleet\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAt(t, primary, "add", "README.md")
	gitAt(t, primary, "commit", "-m", "initial")

	primaryNested := filepath.Join(primary, "cmd", "sdlc")
	if err := os.MkdirAll(primaryNested, 0o755); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(fleetRoot, "linked", "ariadne-feature")
	if err := os.MkdirAll(filepath.Dir(linked), 0o755); err != nil {
		t.Fatal(err)
	}
	gitAt(t, primary, "worktree", "add", "-b", "fleet-test-linked", linked)
	linkedNested := filepath.Join(linked, "cmd", "sdlc")
	if err := os.MkdirAll(linkedNested, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(fleetRoot, "ariadne-link")
	linkErr := os.Symlink(primary, link)

	wantPrimary := mustCanonicalPath(t, primary)
	wantCommon := mustCanonicalPath(t, filepath.Join(primary, ".git"))
	wantFleet := mustCanonicalPath(t, fleetRoot)
	for _, vantage := range []string{primary, primaryNested, linked, linkedNested} {
		got, err := NormalizeVantage(execGitReader{}, vantage)
		if err != nil {
			t.Fatalf("NormalizeVantage(%q): %v", vantage, err)
		}
		if got.RepoIdentity != wantCommon || got.PrimaryRoot != wantPrimary || got.FleetRoot != wantFleet {
			t.Errorf("NormalizeVantage(%q) stable tuple = %+v, want repo=%q primary=%q fleet=%q", vantage, got, wantCommon, wantPrimary, wantFleet)
		}
		if got.WorktreeRoot != wantPrimary && got.WorktreeRoot != mustCanonicalPath(t, linked) {
			t.Errorf("NormalizeVantage(%q) worktree root = %q, want primary or linked root", vantage, got.WorktreeRoot)
		}
	}
	t.Run("symlink", func(t *testing.T) {
		if linkErr != nil {
			t.Skipf("symlinks unavailable: %v", linkErr)
		}
		got, err := NormalizeVantage(execGitReader{}, link)
		if err != nil {
			t.Fatal(err)
		}
		if got.RepoIdentity != wantCommon || got.PrimaryRoot != wantPrimary || got.FleetRoot != wantFleet || got.WorktreeRoot != wantPrimary {
			t.Errorf("NormalizeVantage(%q) = %+v, want primary symlink vantage", link, got)
		}
	})
}

func TestNormalizeVantage_RealGitSkipsUnrelatedPrunableWorktree(t *testing.T) {
	fleetRoot := t.TempDir()
	primary := newGitRepo(t, fleetRoot, "ariadne")
	prunable := filepath.Join(fleetRoot, "linked", "gone")
	healthy := filepath.Join(fleetRoot, "linked", "healthy")
	if err := os.MkdirAll(filepath.Dir(prunable), 0o755); err != nil {
		t.Fatal(err)
	}
	gitAt(t, primary, "worktree", "add", "-b", "fleet-test-gone", prunable)
	prunableCanonical := mustCanonicalPath(t, prunable)
	if err := os.RemoveAll(prunable); err != nil {
		t.Fatal(err)
	}
	gitAt(t, primary, "worktree", "add", "-b", "fleet-test-healthy", healthy)

	porcelain := gitOutputAt(t, healthy, "worktree", "list", "--porcelain", "-z")
	if !strings.Contains(porcelain, "worktree "+prunableCanonical+"\x00") || !strings.Contains(porcelain, "prunable") {
		t.Fatalf("fixture did not produce an unrelated prunable worktree:\n%q", porcelain)
	}
	got, err := NormalizeVantage(execGitReader{}, healthy)
	if err != nil {
		t.Fatalf("NormalizeVantage(healthy worktree): %v", err)
	}
	if got.WorktreeRoot != mustCanonicalPath(t, healthy) || got.PrimaryRoot != mustCanonicalPath(t, primary) {
		t.Errorf("NormalizeVantage(healthy worktree) = %+v, want healthy/current and primary roots", got)
	}
}

func TestNormalizeVantage_RealGitSkipsUnrelatedLockedMissingWorktree(t *testing.T) {
	fleetRoot := t.TempDir()
	primary := newGitRepo(t, fleetRoot, "ariadne")
	locked := filepath.Join(fleetRoot, "linked", "locked-gone")
	healthy := filepath.Join(fleetRoot, "linked", "healthy")
	if err := os.MkdirAll(filepath.Dir(locked), 0o755); err != nil {
		t.Fatal(err)
	}
	gitAt(t, primary, "worktree", "add", "-b", "fleet-test-locked", locked)
	gitAt(t, primary, "worktree", "lock", locked)
	lockedCanonical := mustCanonicalPath(t, locked)
	if err := os.RemoveAll(locked); err != nil {
		t.Fatal(err)
	}
	gitAt(t, primary, "worktree", "add", "-b", "fleet-test-healthy", healthy)

	porcelain := gitOutputAt(t, healthy, "worktree", "list", "--porcelain", "-z")
	if !strings.Contains(porcelain, "worktree "+lockedCanonical+"\x00") || !strings.Contains(porcelain, "locked") {
		t.Fatalf("fixture did not produce an unrelated locked missing worktree:\n%q", porcelain)
	}
	got, err := NormalizeVantage(execGitReader{}, healthy)
	if err != nil {
		t.Fatalf("NormalizeVantage(healthy worktree): %v", err)
	}
	if got.WorktreeRoot != mustCanonicalPath(t, healthy) || got.PrimaryRoot != mustCanonicalPath(t, primary) {
		t.Errorf("NormalizeVantage(healthy worktree) = %+v, want healthy/current and primary roots", got)
	}
}

func TestNormalizeVantage_SkipsUnresolvableUnrelatedListedWorktree(t *testing.T) {
	fleetRoot := t.TempDir()
	primary := filepath.Join(fleetRoot, "ariadne")
	healthy := filepath.Join(fleetRoot, "linked", "healthy")
	missing := filepath.Join(fleetRoot, "linked", "locked-gone")
	common := filepath.Join(primary, ".git")
	for _, dir := range []string{primary, healthy, common} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	porcelain := worktreePorcelain(primary) + "worktree " + missing + "\x00HEAD deadbeef\x00branch refs/heads/locked\x00locked unavailable\x00\x00" + worktreePorcelain(healthy)
	got, err := NormalizeVantage(&vantageStub{
		roots:     map[string]string{healthy: healthy},
		worktrees: porcelain,
		common:    common,
	}, healthy)
	if err != nil {
		t.Fatalf("NormalizeVantage(healthy worktree): %v", err)
	}
	if got.WorktreeRoot != mustCanonicalPath(t, healthy) || got.PrimaryRoot != mustCanonicalPath(t, primary) {
		t.Errorf("NormalizeVantage(healthy worktree) = %+v, want healthy/current and primary roots", got)
	}
}

func TestNormalizeVantage_RealGitPreservesTerminalControlBytesInWorktreePath(t *testing.T) {
	for _, suffix := range []string{"\n", "\r"} {
		t.Run(fmt.Sprintf("%q", suffix), func(t *testing.T) {
			fleetRoot := t.TempDir()
			primary := newGitRepo(t, fleetRoot, "ariadne"+suffix)
			got, err := NormalizeVantage(execGitReader{}, primary)
			if err != nil {
				t.Fatal(err)
			}
			if got.PrimaryRoot != mustCanonicalPath(t, primary) || got.WorktreeRoot != mustCanonicalPath(t, primary) {
				t.Errorf("NormalizeVantage() = %+v, want terminal-control-byte worktree root %q", got, mustCanonicalPath(t, primary))
			}
		})
	}
}

func TestNormalizeVantage_RefusesMissingContainingWorktree(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing")
	other := filepath.Join(root, "other")
	if err := os.Mkdir(missing, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(other, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := NormalizeVantage(&vantageStub{
		roots:     map[string]string{missing: missing},
		worktrees: worktreePorcelain(other),
		common:    filepath.Join(root, ".git"),
	}, missing)
	if err == nil || !strings.Contains(err.Error(), "containing worktree") {
		t.Fatalf("NormalizeVantage() error = %v, want actionable containing-worktree error", err)
	}
}

func TestNormalizeVantage_RefusesMalformedWorktreeOutput(t *testing.T) {
	root := t.TempDir()
	_, err := NormalizeVantage(&vantageStub{
		roots:     map[string]string{root: root},
		worktrees: "not porcelain\x00",
		common:    filepath.Join(root, ".git"),
	}, root)
	if err == nil || !strings.Contains(err.Error(), "parse git worktree list") {
		t.Fatalf("NormalizeVantage() error = %v, want actionable parse error", err)
	}
}

func TestNormalizeVantage_PreservesWhitespaceAndTerminalControlBytesInWorktreePaths(t *testing.T) {
	for _, suffix := range []string{"\n", "\r"} {
		t.Run(fmt.Sprintf("%q", suffix), func(t *testing.T) {
			fleetRoot := t.TempDir()
			primary := filepath.Join(fleetRoot, "ariadne "+suffix)
			common := filepath.Join(primary, ".git")
			if err := os.MkdirAll(common, 0o755); err != nil {
				t.Fatal(err)
			}
			got, err := NormalizeVantage(&vantageStub{
				roots:     map[string]string{primary: primary},
				worktrees: worktreePorcelain(primary),
				common:    common,
			}, primary)
			if err != nil {
				t.Fatal(err)
			}
			if got.PrimaryRoot != mustCanonicalPath(t, primary) || got.RepoIdentity != mustCanonicalPath(t, common) {
				t.Errorf("NormalizeVantage() = %+v, want whitespace/control-byte-preserving primary and common paths", got)
			}
		})
	}
}

type vantageStub struct {
	roots              map[string]string
	worktrees          string
	common             string
	relativeCommonSeen bool
}

func (s *vantageStub) GitInDir(dir string, args ...string) ([]byte, error) {
	switch strings.Join(args, " ") {
	case "rev-parse --show-toplevel":
		root, ok := s.roots[dir]
		if !ok {
			return nil, fmt.Errorf("unexpected vantage %q", dir)
		}
		return []byte(root + "\n"), nil
	case "worktree list --porcelain -z":
		return []byte(s.worktrees), nil
	case "rev-parse --git-common-dir":
		commandDir, err := canonicalPath(dir)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(commandDir, s.common)
		if err != nil {
			return nil, err
		}
		if !filepath.IsAbs(rel) {
			s.relativeCommonSeen = true
		}
		return []byte(rel + "\n"), nil
	default:
		return nil, fmt.Errorf("unexpected git command in %q: %s", dir, strings.Join(args, " "))
	}
}

type execGitReader struct{}

func (execGitReader) GitInDir(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitAt(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git -C %s %s: %v\n%s", dir, strings.Join(args, " "), err, out)
	}
}

func gitOutputAt(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git -C %s %s: %v", dir, strings.Join(args, " "), err)
	}
	return string(out)
}

func newGitRepo(t *testing.T, parent, name string) string {
	t.Helper()
	primary := filepath.Join(parent, name)
	gitAt(t, parent, "init", primary)
	gitAt(t, primary, "config", "user.email", "fleet-test@example.test")
	gitAt(t, primary, "config", "user.name", "Fleet Test")
	if err := os.WriteFile(filepath.Join(primary, "README.md"), []byte("fleet\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAt(t, primary, "add", "README.md")
	gitAt(t, primary, "commit", "-m", "initial")
	return primary
}

func mustCanonicalPath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
}

func worktreePorcelain(primary string, linked ...string) string {
	var records []string
	for i, path := range append([]string{primary}, linked...) {
		branch := "main"
		if i > 0 {
			branch = fmt.Sprintf("linked-%d", i)
		}
		records = append(records, "worktree "+path+"\x00HEAD deadbeef\x00branch refs/heads/"+branch)
	}
	return strings.Join(records, "\x00\x00") + "\x00\x00"
}
