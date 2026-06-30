package judge

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBinAugmentedEnv(t *testing.T) {
	sep := string(os.PathListSeparator)

	t.Run("PATH present → bin dir prepended, rest preserved", func(t *testing.T) {
		got := binAugmentedEnv("/w/ariadne/bin", []string{"HOME=/h", "PATH=/usr/bin" + sep + "/bin"})
		if want := "PATH=/w/ariadne/bin" + sep + "/usr/bin" + sep + "/bin"; !envContains(got, want) {
			t.Errorf("PATH not prepended\n got %v\nwant entry %q", got, want)
		}
		if !envContains(got, "HOME=/h") {
			t.Errorf("unrelated env var dropped: %v", got)
		}
	})

	t.Run("PATH absent → synthesized", func(t *testing.T) {
		if got := binAugmentedEnv("/w/ariadne/bin", []string{"HOME=/h"}); !envContains(got, "PATH=/w/ariadne/bin") {
			t.Errorf("PATH not synthesized: %v", got)
		}
	})

	t.Run("empty / bare dir → no-op", func(t *testing.T) {
		in := []string{"PATH=/usr/bin"}
		if !reflect.DeepEqual(binAugmentedEnv("", in), in) {
			t.Errorf("empty dir should be a no-op")
		}
		if !reflect.DeepEqual(binAugmentedEnv(".", in), in) {
			t.Errorf("'.' dir should be a no-op")
		}
	})
}

func envContains(env []string, want string) bool {
	for _, e := range env {
		if e == want {
			return true
		}
	}
	return false
}

// TestMinimalPathResolvesSdlc is the #138 process-level fixture: a subprocess
// started with a deliberately narrow PATH (no owner bin/) still resolves `sdlc`
// once binAugmentedEnv prepends the owner bin/ — proving a fresh review agent
// would find the binary without the user's shell startup files. Real `sh -c`
// spawn, no agent process.
func TestMinimalPathResolvesSdlc(t *testing.T) {
	bin := t.TempDir()
	sdlc := filepath.Join(bin, "sdlc")
	if err := os.WriteFile(sdlc, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// A narrow base PATH that does NOT contain the temp bin dir.
	base := []string{"PATH=/usr/bin" + string(os.PathListSeparator) + "/bin"}
	cmd := exec.Command("sh", "-c", "command -v sdlc")
	cmd.Env = binAugmentedEnv(bin, base)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("`command -v sdlc` failed with augmented env: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != sdlc {
		t.Errorf("resolved %q, want the injected owner-bin sdlc %q", got, sdlc)
	}
}
