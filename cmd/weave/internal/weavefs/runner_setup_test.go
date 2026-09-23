package weavefs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
)

func TestSetupRunnerHelper(t *testing.T) {
	env := os.Getenv("WEAVE_SETUP_RUNNER_HELPER")
	if env == "" {
		return
	}
	lease, err := staging.AcquireSetup(env)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	runner := ExecRunner{ExtraFiles: []*os.File{lease}, Stdin: os.Stdin}
	stage := os.Getenv("WEAVE_SETUP_STAGE")
	fd := 3
	if stage != "" {
		fd = 4
	}
	// Release is an explicit barrier after the last setup write. No sleep is
	// used as evidence of either child readiness or descriptor release.
	script := fmt.Sprintf("echo ready; read release; exec %d>&-; exec 3>&-; echo released", fd)
	if stage == "" {
		err = runner.Run(env, []string{"sh", "-c", script})
	} else {
		err = runner.RunOwned(env, []string{"sh", "-c", script}, stage)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetupLeaseSurvivesRunnerParentDeath(t *testing.T) {
	for _, owned := range []bool{false, true} {
		t.Run(fmt.Sprintf("owned=%v", owned), func(t *testing.T) {
			env := t.TempDir()
			stage := ""
			if owned {
				var err error
				stage, err = staging.New(filepath.Join(env, "generation"))
				if err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			helper := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSetupRunnerHelper$")
			helper.Env = append(os.Environ(), "WEAVE_SETUP_RUNNER_HELPER="+env, "WEAVE_SETUP_STAGE="+stage)
			// Own the pipes so killing/waiting the parent does not close the child's
			// communication channel. EOF on cleanup also bounds the child lifetime.
			input, feed, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			defer feed.Close()
			output, sink, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			defer sink.Close()
			helper.Stdin, helper.Stdout, helper.Stderr = input, sink, sink
			if err := helper.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = helper.Process.Kill(); _ = helper.Wait() }()
			input.Close()
			sink.Close()
			lines := make(chan string, 4)
			go func() {
				scan := bufio.NewScanner(output)
				for scan.Scan() {
					lines <- scan.Text()
				}
				close(lines)
			}()
			expect := func(want string) {
				t.Helper()
				select {
				case got := <-lines:
					if got != want {
						t.Fatalf("got %q, want %q", got, want)
					}
				case <-ctx.Done():
					t.Fatal("process barrier timeout", want)
				}
			}
			expect("ready")
			if err := helper.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			_ = helper.Wait()
			if lease, err := staging.AcquireSetup(env); !errors.Is(err, staging.ErrSetupInUse) {
				if lease != nil {
					lease.Close()
				}
				t.Fatalf("parent death released live child's setup lease: %v", err)
			}
			if owned {
				if lease, err := staging.Exclusive(stage); !errors.Is(err, staging.ErrInUse) {
					if lease != nil {
						lease.Close()
					}
					t.Fatalf("stage descriptor lost: %v", err)
				}
			}
			if _, err := feed.Write([]byte("release\n")); err != nil {
				t.Fatal(err)
			}
			expect("released")
			lease, err := staging.AcquireSetup(env)
			if err != nil {
				t.Fatal("child release retained lease", err)
			}
			lease.Close()
			if owned {
				lease, err := staging.Exclusive(stage)
				if err != nil {
					t.Fatal(err)
				}
				lease.Close()
			}
		})
	}
}
