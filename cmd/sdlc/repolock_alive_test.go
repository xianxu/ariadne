package main

import (
	"os"
	"testing"
)

// A process we may not signal still exists (side-quest of #310): pid 1 answers
// kill(0) with EPERM for a non-root user, as a peer slot outside the sandbox does.
func TestProcessAliveTreatsEPERMAsAlive(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root may signal pid 1")
	}
	if !processAlive(1) {
		t.Error("processAlive(1) = false; EPERM must read as alive")
	}
	if !processAlive(os.Getpid()) {
		t.Error("processAlive(self) = false")
	}
	if processAlive(0) || processAlive(-1) {
		t.Error("non-positive pids must read as dead")
	}
}
