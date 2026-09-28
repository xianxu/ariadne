package main

import (
	"os/exec"
	"testing"
)

// runGitCmd keeps stderr out of a successful command's output — a parsed
// stream (a -z status, a SHA) must not carry warnings — and returns it after
// stdout on failure so error messages keep the diagnostics (#259).
func TestRunGitCmdSeparatesStderr(t *testing.T) {
	out, err := runGitCmd(exec.Command("sh", "-c", "printf 'data'; printf 'warning: noise' >&2"))
	if err != nil || string(out) != "data" {
		t.Fatalf("success = (%q, %v), want stdout alone", out, err)
	}
	out, err = runGitCmd(exec.Command("sh", "-c", "printf 'partial'; printf 'fatal: broke' >&2; exit 3"))
	if err == nil || string(out) != "partialfatal: broke" {
		t.Fatalf("failure = (%q, %v), want stdout then stderr with the error", out, err)
	}
}
