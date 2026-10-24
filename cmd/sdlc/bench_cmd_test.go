package main

import "testing"

func TestBenchCLIOnlyMetisStepsAndPublicProtocol(t *testing.T) {
	cmd := NewBenchCmd()
	for _, name := range []string{"prepare", "session", "score"} {
		got, _, err := cmd.Find([]string{"step", name})
		if err != nil || got.Name() != name || got.Parent().Name() != "step" {
			t.Errorf("missing step %s: %v", name, err)
		}
	}
	for _, name := range []string{"run", "record", "status", "clean", "cancel", "prepare", "score"} {
		got, _, err := cmd.Find([]string{name})
		if err == nil && got != cmd {
			t.Errorf("standalone engine survives: %s", name)
		}
	}
	for _, name := range []string{"decision", "finish"} {
		got, _, err := cmd.Find([]string{name})
		if err != nil || got.Name() != name {
			t.Errorf("missing public protocol %s", name)
		}
	}
}
