package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestShardedReportRetainsSubtestDiagnostics(t *testing.T) {
	cmd := exec.Command("python3", "-B", filepath.Join("..", "..", "scripts", "test-shard.test.py"))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shard-report regression: %v\n%s", err, output)
	}
}
