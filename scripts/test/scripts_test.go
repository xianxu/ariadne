// Package scripttest runs the repository's shell tests from `make test` (#362):
// the Go suite is what `make test` runs, so a shell test needs a Go entry point
// to count.
package scripttest

import (
	"os"
	"os/exec"
	"testing"

	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestMain(m *testing.M) {
	testgate.Enter() // one machine-wide test slot (#324)
	os.Exit(m.Run())
}

// TestConstructScripts runs every construct/scripts/test/*.test.sh through the
// one runner the CI merge check also runs, so both apply the same exclusions.
func TestConstructScripts(t *testing.T) {
	out, err := exec.Command("bash", "construct-scripts.test.sh").CombinedOutput()
	t.Logf("%s", out)
	if err != nil {
		t.Fatalf("construct/scripts/test: %v (failing tests are named above)", err)
	}
}
