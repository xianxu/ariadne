package acquire

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0644); err != nil {
		t.Fatal(err)
	}
}
func gitFixture(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
	b, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("git %v: %s %v", args, b, e)
	}
	return strings.TrimSpace(string(b))
}
func origin(t *testing.T, base, name, deps string, layer bool) string {
	t.Helper()
	src := filepath.Join(base, name+"-work")
	put(t, filepath.Join(src, "README"), name)
	if layer {
		put(t, filepath.Join(src, "construct/base.manifest"), "")
	}
	if deps != "" {
		put(t, filepath.Join(src, "construct/deps"), deps)
	}
	gitFixture(t, src, "init")
	gitFixture(t, src, "add", ".")
	gitFixture(t, src, "commit", "-m", "init")
	bare := filepath.Join(base, name+".git")
	gitFixture(t, base, "clone", "--bare", src, bare)
	return bare
}
func TestEnsureCloneReuseAndConflict(t *testing.T) {
	base := t.TempDir()
	src := origin(t, base, "base", "", true)
	dest := filepath.Join(base, "clone")
	ctx := context.Background()
	if err := Ensure(ctx, dest, src, true); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(dest, "README"), "dirty")
	before := gitFixture(t, dest, "rev-parse", "HEAD")
	if err := Ensure(ctx, dest, src, true); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dest, "README"))
	if string(b) != "dirty" || gitFixture(t, dest, "rev-parse", "HEAD") != before {
		t.Fatal("mutated checkout")
	}
	other := origin(t, base, "other", "", true)
	if err := Ensure(ctx, dest, other, true); err == nil || !strings.Contains(err.Error(), "origin") {
		t.Fatalf("conflict: %v", err)
	}
}
func TestEnsureFailureDoesNotPublish(t *testing.T) {
	base := t.TempDir()
	src := origin(t, base, "plain", "", false)
	for _, source := range []string{src, filepath.Join(base, "absent.git")} {
		dest := filepath.Join(base, "clone")
		if err := Ensure(context.Background(), dest, source, true); err == nil {
			t.Fatal("expected error")
		}
		if _, err := os.Lstat(dest); !os.IsNotExist(err) {
			t.Fatalf("published failed clone: %v", err)
		}
		matches, _ := filepath.Glob(filepath.Join(base, ".clone-weave-*"))
		if len(matches) != 0 {
			t.Fatal(matches)
		}
	}
}
func TestRestoreTransitiveDiamondAndDataMounts(t *testing.T) {
	base := t.TempDir()
	orig := filepath.Join(base, "origins")
	os.MkdirAll(orig, 0755)
	data := origin(t, orig, "dataset", "", false)
	foundation := origin(t, orig, "foundation", "data "+data+" data/a\ndata "+data+" data/b\n", true)
	left := origin(t, orig, "left", "substrate ../foundation "+foundation+"\n", true)
	right := origin(t, orig, "right", "substrate ../foundation "+foundation+"\n", true)
	root := filepath.Join(base, "leaf")
	put(t, filepath.Join(root, "construct/deps"), "substrate ../left "+left+"\nsubstrate ../right "+right+"\n")
	got, err := Restore(context.Background(), root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Layers) != 4 || filepath.Base(got.Layers[0]) != "foundation" || len(got.Mounts) != 2 || got.Mounts[0].Source != got.Mounts[1].Source {
		t.Fatalf("%+v", got)
	}
	for _, m := range got.Mounts {
		if _, err := os.Lstat(m.Target); !os.IsNotExist(err) {
			t.Fatalf("created mount %s", m.Target)
		}
	}
}
func TestRestoreDryRunMissingAndLocalCycle(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "leaf")
	put(t, filepath.Join(root, "construct/deps"), "substrate ../missing https://github.com/org/missing.git\n")
	got, err := Restore(context.Background(), root, true)
	if err == nil || len(got.Missing) != 1 {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := os.Lstat(filepath.Join(base, "missing")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "construct/deps"), "substrate ../missing\n")
	if _, err := Restore(context.Background(), root, false); err == nil || !strings.Contains(err.Error(), "source") {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "construct/base.manifest"), "")
	put(t, filepath.Join(base, "missing/construct/base.manifest"), "")
	put(t, filepath.Join(base, "missing/construct/deps"), "substrate ../leaf\n")
	if _, err := Restore(context.Background(), root, false); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatal(err)
	}
}
func TestRestoreRejectsEscapingMount(t *testing.T) {
	for _, mount := range []string{"../outside", "/absolute", "."} {
		t.Run(mount, func(t *testing.T) {
			root := t.TempDir()
			put(t, filepath.Join(root, "construct/deps"), "data https://github.com/org/data.git "+mount+"\n")
			if _, err := Restore(context.Background(), root, true); err == nil || !strings.Contains(err.Error(), "mount") {
				t.Fatal(err)
			}
		})
	}
}

func TestRestoreSourceDestinationConflict(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "construct/deps"), "data https://github.com/one/shared.git data/a\ndata https://github.com/two/shared.git data/b\n")
	if _, err := Restore(context.Background(), root, true); err == nil || !strings.Contains(err.Error(), "conflicting declared sources") {
		t.Fatal(err)
	}
}
func TestRestoreMountParentSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "data")); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "construct/deps"), "data https://github.com/org/data.git data/x\n")
	if _, err := Restore(context.Background(), root, true); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatal(err)
	}
}
func TestEnsureWarmCheckoutDoesNotContactOrigin(t *testing.T) {
	base := t.TempDir()
	src := origin(t, base, "base", "", true)
	dest := filepath.Join(base, "clone")
	if err := Ensure(context.Background(), dest, src, true); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(src, src+"-offline"); err != nil {
		t.Fatal(err)
	}
	if err := Ensure(context.Background(), dest, src, true); err != nil {
		t.Fatal(err)
	}
}
func TestEnsureCanceledCloneLeavesNoDestination(t *testing.T) {
	base := t.TempDir()
	src := origin(t, base, "base", "", true)
	dest := filepath.Join(base, "clone")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Ensure(ctx, dest, src, true); err == nil {
		t.Fatal("expected cancellation")
	}
	if _, err := os.Lstat(dest); !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
func TestEnsureAuthoredDirectoryPreserved(t *testing.T) {
	base := t.TempDir()
	src := origin(t, base, "base", "", true)
	dest := filepath.Join(base, "authored")
	put(t, filepath.Join(dest, "mine"), "authored")
	if err := Ensure(context.Background(), dest, src, true); err == nil {
		t.Fatal("expected destination conflict")
	}
	b, err := os.ReadFile(filepath.Join(dest, "mine"))
	if err != nil || string(b) != "authored" {
		t.Fatalf("%s %v", b, err)
	}
}

func TestEnsureLocalOnlyLayer(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "construct/base.manifest"), "")
	if err := Ensure(context.Background(), root, "", true); err != nil {
		t.Fatal(err)
	}
	if err := Ensure(context.Background(), filepath.Join(root, "absent"), "", true); err == nil || !strings.Contains(err.Error(), "source") {
		t.Fatal(err)
	}
}

func TestEnsureEquivalentGitHubOriginWithoutNetwork(t *testing.T) {
	base := t.TempDir()
	src := origin(t, base, "base", "", true)
	dest := filepath.Join(base, "clone")
	if err := Ensure(context.Background(), dest, src, true); err != nil {
		t.Fatal(err)
	}
	gitFixture(t, dest, "remote", "set-url", "origin", "git@github.com:Org/base.git")
	if err := Ensure(context.Background(), dest, "https://github.com/org/base.git", true); err != nil {
		t.Fatal(err)
	}
}
