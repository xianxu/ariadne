package reviewbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFenceDeniesSourceWrites is the live conformance check: under the probed
// fence, a reviewer command can write its checkout but neither the source's
// files nor its Git state. It skips where no fence applies — notably inside an
// agent's own sandbox, where Seatbelt cannot nest (run it unsandboxed or in CI).
func TestFenceDeniesSourceWrites(t *testing.T) {
	fence, reason := Probe()
	if fence == FenceNone {
		t.Skipf("no write fence here (%s); the review would run checkout-only", reason)
	}
	src, _, h2 := sourceRepo(t)
	co, err := Create(t.TempDir(), src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer co.Close()
	before := snapshot(t, src)
	protect, err := Canonical([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	iso := Isolation{Fence: fence, Dir: co.Dir, Protect: protect}

	run := func(script string) error {
		name, args := iso.Wrap("sh", []string{"-c", script})
		return exec.Command(name, args...).Run()
	}
	if err := run("echo ok > in-checkout"); err != nil {
		t.Fatalf("write inside the checkout failed under the fence: %v", err)
	}
	if _, err := os.Stat(filepath.Join(co.Dir, "in-checkout")); err != nil {
		t.Fatalf("the wrapped command did not run in the checkout: %v", err)
	}
	if err := run("echo x > " + filepath.Join(src, "a.txt")); err == nil {
		t.Error("write to a source file succeeded under the fence")
	}
	if err := run("git -C " + src + " branch evil"); err == nil {
		t.Error("git ref write in the source succeeded under the fence")
	}
	if after := snapshot(t, src); after != before {
		t.Fatal("the fenced command changed the source checkout")
	}
}

// The probe claims a fence only after seeing it deny a write: a "fence" that
// runs the command but enforces nothing is reported as none.
func TestProbeRejectsAFenceThatDoesNotDeny(t *testing.T) {
	fence, reason := verifyFence(FenceNone, "env") // Wrap with FenceNone is env -C: no enforcement
	if fence != FenceNone || !strings.Contains(reason, "did not deny") {
		t.Fatalf("verifyFence = %q, %q; want none, did not deny", fence, reason)
	}
}

func TestProbeReportsAMissingBinary(t *testing.T) {
	if fence, reason := verifyFence(FenceSeatbelt, "no-such-fence-binary"); fence != FenceNone || !strings.Contains(reason, "not found") {
		t.Fatalf("verifyFence = %q, %q", fence, reason)
	}
}
