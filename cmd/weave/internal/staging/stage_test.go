package staging

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestStageOwnerProcess(t *testing.T) {
	destination := os.Getenv("WEAVE_STAGE_TEST_DESTINATION")
	if destination == "" {
		t.Skip("producer subprocess only")
	}
	stage, err := New(destination)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(stage, "output"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, "output", "partial.json"), []byte("partial"), 0644); err != nil {
		t.Fatal(err)
	}
	fmt.Println(stage)
	for {
		time.Sleep(time.Hour)
	}
}

func TestKilledProducerStageIsReclaimedBeforeRetry(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "generation")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStageOwnerProcess$")
	cmd.Env = append(os.Environ(), "WEAVE_STAGE_TEST_DESTINATION="+destination)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	scanner := bufio.NewScanner(pipe)
	if !scanner.Scan() {
		t.Fatal("producer did not report stage", scanner.Err())
	}
	stage := scanner.Text()
	if _, err := os.Stat(filepath.Join(stage, "output", "partial.json")); err != nil {
		t.Fatal(err)
	}
	if err := Reclaim(destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stage); err != nil {
		t.Fatal("live producer stage removed", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	if err := Reclaim(destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("killed producer residue: %v", err)
	}
	retry, err := New(destination)
	if err != nil {
		t.Fatal(err)
	}
	if err := Remove(retry); err != nil {
		t.Fatal(err)
	}
}
