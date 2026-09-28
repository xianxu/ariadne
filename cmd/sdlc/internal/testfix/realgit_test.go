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
