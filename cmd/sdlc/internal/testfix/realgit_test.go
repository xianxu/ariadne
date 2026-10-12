package testfix

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrependPath(t *testing.T) {
	sep := string(os.PathListSeparator)
	cases := []struct{ dir, path, want string }{
		{"/g", "", "/g"},
		{"/g", "/usr/bin", "/g" + sep + "/usr/bin"},
		{"/g", "/g" + sep + "/usr/bin", "/g" + sep + "/usr/bin"},
		{"/g", "/usr/bin" + sep + "/g", "/g" + sep + "/usr/bin" + sep + "/g"},
	}
	for _, c := range cases {
		if got := PrependPath(c.dir, c.path); got != c.want {
			t.Errorf("PrependPath(%q, %q) = %q, want %q", c.dir, c.path, got, c.want)
		}
	}
}

func TestPreferRealGitResolvesGitFromItsExecPath(t *testing.T) {
	out, err := exec.Command("git", "--exec-path").Output()
	if err != nil {
		t.Skip("git not installed")
	}
	execPath := strings.TrimSpace(string(out))
	t.Setenv("PATH", os.Getenv("PATH")) // restored after the test

	PreferRealGit()

	got, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(got) != execPath {
		t.Fatalf("git resolves to %s, want it under the exec-path %s", got, execPath)
	}
}

// #302: the GIT_CONFIG_* environment never reaches the receive side of a local
// push (git unsets it for git-receive-pack), so QuietBackgroundGit's template
// must: a push into a repository created under it starts no maintenance.
func TestQuietBackgroundGitReachesTheReceiveSide(t *testing.T) {
	quietVars := []string{"GIT_CONFIG_COUNT", "GIT_CONFIG_KEY_0", "GIT_CONFIG_VALUE_0", "GIT_CONFIG_KEY_1", "GIT_CONFIG_VALUE_1", "GIT_TEMPLATE_DIR"}
	for _, k := range quietVars {
		t.Setenv(k, "") // restored after the test
		_ = os.Unsetenv(k)
	}
	push := func() string { // GIT_TRACE of a push into a fresh bare origin
		origin := filepath.Join(t.TempDir(), "origin.git")
		Git(t, "", "init", "--bare", "-q", origin)
		cmd := exec.Command("git", "push", "-q", origin, "HEAD:refs/heads/main")
		cmd.Dir = Repo(t, InitialCommit())
		cmd.Env = append(os.Environ(), "GIT_TRACE=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("push: %v\n%s", err, out)
		}
		return string(out)
	}
	if !strings.Contains(push(), "maintenance run --auto") {
		t.Skip("this git starts no maintenance on the receive side")
	}
	cleanup := QuietBackgroundGit()
	t.Cleanup(cleanup)
	if out := push(); strings.Contains(out, "maintenance run") {
		t.Fatalf("receive-side maintenance still ran:\n%s", out)
	}
}
