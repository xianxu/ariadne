package reviewbox

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckoutOwnDependencyResetCloseSweep(t *testing.T) {
	src, h1, h2 := sourceRepo(t)
	root := t.TempDir()
	primary, err := Create(root, src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer primary.Close()
	child, err := Create(root, src, h1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer child.Close()
	owner, ok := any(primary).(interface {
		OwnDependency(*Checkout) error
		ResetContext(context.Context) error
	})
	if !ok {
		t.Fatal("checkout lacks dependency ownership and contextual reset")
	}
	if err := owner.OwnDependency(child); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(child.Dir, "a.txt"), "changed")
	write(t, filepath.Join(primary.Dir, "a.txt"), "changed")
	if removed, problems := Sweep(root); removed != 0 || len(problems) != 0 {
		t.Fatalf("live sweep: %d %v", removed, problems)
	}
	if err := owner.ResetContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		co   *Checkout
		want string
	}{{primary, "two\n"}, {child, "one\n"}} {
		got, err := os.ReadFile(filepath.Join(item.co.Dir, "a.txt"))
		if err != nil || string(got) != item.want {
			t.Fatalf("reset %s: %q %v", item.co.Dir, got, err)
		}
	}
	if err := primary.Close(); err != nil {
		t.Fatal(err)
	}
	for _, co := range []*Checkout{primary, child} {
		for _, p := range []string{co.Dir, co.lockPath()} {
			if _, err := os.Stat(p); !os.IsNotExist(err) {
				t.Fatalf("retained %s: %v", p, err)
			}
		}
	}
	if err := primary.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateContextCancellationReapsDescendantsAndCleansLock(t *testing.T) {
	root := t.TempDir()
	tools := t.TempDir()
	ready := filepath.Join(tools, "child.pid")
	write(t, filepath.Join(tools, "git"), "#!/bin/sh\nsleep 60 &\nprintf '%s' \"$!\" > \"$REVIEW_TEST_READY\"\nwait\n")
	if err := os.Chmod(filepath.Join(tools, "git"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("REVIEW_TEST_READY", ready)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := CreateContext(ctx, root, "unused", "unused", nil); done <- err }()
	pid := testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil)
	if pid == 0 {
		t.Fatal("Git child never started")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	if err := testfix.AwaitExit(pid, testfix.HangCeiling); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("checkout debris: %v %v", entries, err)
	}
}

func TestCreateDependencyContextIgnoresConfigTemplatesAndSmudge(t *testing.T) {
	src, _, h2 := sourceRepo(t)
	config := filepath.Join(t.TempDir(), "gitconfig")
	template := t.TempDir()
	if err := os.MkdirAll(filepath.Join(template, "hooks"), 0755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "executed")
	hook := filepath.Join(template, "hooks", "post-checkout")
	write(t, hook, "#!/bin/sh\nprintf ran > \"$REVIEW_TEST_MARKER\"\n")
	if err := os.Chmod(hook, 0755); err != nil {
		t.Fatal(err)
	}
	write(t, config, "[init]\n templateDir = "+template+"\n[filter \"bad\"]\n smudge = touch "+marker+"\n required = true\n")
	// The source's committed attributes select a globally configured driver;
	// dependency checkout must materialize the bytes without running that driver.
	write(t, filepath.Join(src, ".gitattributes"), "a.txt filter=bad\n")
	testfix.Git(t, src, "add", ".gitattributes")
	testfix.Git(t, src, "commit", "-qm", "attributes")
	h2 = strings.TrimSpace(testfix.Capture(t, src, "rev-parse", "HEAD"))
	t.Setenv("GIT_CONFIG_GLOBAL", config)
	t.Setenv("REVIEW_TEST_MARKER", marker)
	co, err := CreateDependencyContext(t.Context(), t.TempDir(), src, h2)
	if err != nil {
		t.Fatal(err)
	}
	defer co.Close()
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("dependency executed inherited template or filter: %v", err)
	}
	if err := co.ResetContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("reset executed inherited template or filter: %v", err)
	}
}

func TestCheckoutOwnDependencyRejectsInvalidOwnership(t *testing.T) {
	src, _, head := sourceRepo(t)
	root := t.TempDir()
	makeCheckout := func() *Checkout {
		co, err := Create(root, src, head, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = co.Close() })
		return co
	}
	primary, child, other := makeCheckout(), makeCheckout(), makeCheckout()
	if err := primary.OwnDependency(nil); err == nil {
		t.Fatal("accepted nil")
	}
	if err := primary.OwnDependency(primary); err == nil {
		t.Fatal("accepted self")
	}
	if err := primary.OwnDependency(child); err != nil {
		t.Fatal(err)
	}
	if err := primary.OwnDependency(child); err == nil {
		t.Fatal("accepted duplicate")
	}
	if err := other.OwnDependency(child); err == nil {
		t.Fatal("accepted second owner")
	}
	if err := child.OwnDependency(primary); err == nil {
		t.Fatal("accepted cycle")
	}
	if err := other.Close(); err != nil {
		t.Fatal(err)
	}
	if err := primary.OwnDependency(other); err == nil {
		t.Fatal("accepted closed child")
	}
}

func TestCheckoutCloseAttemptsEveryOwnedCheckoutAfterFailure(t *testing.T) {
	src, _, head := sourceRepo(t)
	root := t.TempDir()
	primary, err := Create(root, src, head, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer primary.Close()
	var children []*Checkout
	for range 2 {
		co, err := Create(root, src, head, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer co.Close()
		if err := primary.OwnDependency(co); err != nil {
			t.Fatal(err)
		}
		children = append(children, co)
	}
	// Replace one published lock name with a nonempty directory so Close's lock
	// removal fails without depending on UID-specific chmod behavior.
	bad := children[0]
	if err := os.Remove(bad.lockPath()); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(bad.lockPath(), 0700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(bad.lockPath(), "block"), "x")
	if err := primary.Close(); err == nil {
		t.Fatal("lost child cleanup failure")
	}
	for _, co := range append(children, primary) {
		if _, err := os.Stat(co.Dir); !os.IsNotExist(err) {
			t.Fatalf("checkout remained after sibling failure: %s %v", co.Dir, err)
		}
		if co.lock != nil {
			t.Fatalf("lock still held: %s", co.Dir)
		}
	}
	if _, err := os.Stat(children[1].lockPath()); !os.IsNotExist(err) {
		t.Fatalf("later child lock remains: %v", err)
	}
	if _, err := os.Stat(primary.lockPath()); !os.IsNotExist(err) {
		t.Fatalf("primary lock remains: %v", err)
	}
	if err := primary.Close(); err != nil {
		t.Fatalf("not idempotent: %v", err)
	}
}

func TestCheckoutOwnedDependenciesSweptAfterOwnerDeath(t *testing.T) {
	src, _, head := sourceRepo(t)
	root := t.TempDir()
	primary, err := Create(root, src, head, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer primary.Close()
	child, err := Create(root, src, head, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer child.Close()
	if err := primary.OwnDependency(child); err != nil {
		t.Fatal(err)
	}
	// Closing the descriptors without cleanup models kernel lock release after
	// process exit. Neither a manifest nor an owner process is needed by Sweep.
	if err := primary.lock.Close(); err != nil {
		t.Fatal(err)
	}
	if err := child.lock.Close(); err != nil {
		t.Fatal(err)
	}
	if removed, problems := Sweep(root); removed != 2 || len(problems) != 0 {
		t.Fatalf("sweep owned family: %d %v", removed, problems)
	}
}

func TestCheckoutResetAttemptsDependenciesAfterPrimaryFailure(t *testing.T) {
	src, _, head := sourceRepo(t)
	other, _, otherHead := sourceRepo(t)
	root := t.TempDir()
	primary, err := Create(root, src, head, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer primary.Close()
	child, err := Create(root, other, otherHead, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer child.Close()
	if err := primary.OwnDependency(child); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(child.Dir, "a.txt"), "dirty")
	moved := src + "-unavailable"
	if err := os.Rename(src, moved); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(moved, src)
	if err := primary.ResetContext(t.Context()); err == nil {
		t.Fatal("missing source should fail reset")
	}
	got, err := os.ReadFile(filepath.Join(child.Dir, "a.txt"))
	if err != nil || string(got) != "two\n" {
		t.Fatalf("later dependency reset skipped: %q %v", got, err)
	}
}
