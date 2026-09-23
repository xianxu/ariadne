//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCLISignalProcess(t *testing.T) {
	switch os.Getenv("SDLC_SIGNAL_ROLE") {
	case "cli":
		var args []string
		if err := json.Unmarshal([]byte(os.Getenv("SDLC_SIGNAL_ARGS")), &args); err != nil {
			os.Exit(3)
		}
		os.Args = append([]string{"sdlc"}, args...)
		main()
		os.Exit(0)
	case "lock-owner", "legacy-lock":
		verb := "change-code"
		if os.Getenv("SDLC_SIGNAL_ROLE") == "legacy-lock" {
			verb = "push"
		}
		root := &cobra.Command{Use: "sdlc"}
		root.AddCommand(markMutatingCommand(&cobra.Command{Use: verb, RunE: func(cmd *cobra.Command, _ []string) error {
			if err := os.WriteFile(os.Getenv("SDLC_SIGNAL_READY"), []byte("ready"), 0600); err != nil {
				return err
			}
			<-cmd.Context().Done()
			common, err := repoLockGitCommonDir()
			if err != nil {
				return err
			}
			if _, err := os.Stat(filepath.Join(common, "sdlc.lock")); err != nil {
				return fmt.Errorf("lock released before cancellation unwind: %w", err)
			}
			if err := os.WriteFile(os.Getenv("SDLC_SIGNAL_PROOF"), []byte("lock held through cancellation"), 0600); err != nil {
				return err
			}
			return cmd.Context().Err()
		}}))
		wrapRepoLockCommands(root)
		os.Exit(executeCLI(root, []string{verb}))
	case "reviewer":
		signal.Ignore(syscall.SIGTERM, syscall.SIGINT)
		if err := os.WriteFile(os.Getenv("SDLC_SIGNAL_READY"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
			os.Exit(4)
		}
		fmt.Println("VERDICT: SHIP")
		for {
			time.Sleep(time.Hour)
		}
	}
}

func TestCLISignalCancelsOwnedReviewer(t *testing.T) {
	for _, verb := range []string{"judge", "close", "change-code"} {
		for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
			t.Run(verb+"/"+sig.String(), func(t *testing.T) {
				issues := closeRepo(t, 69)
				plans := filepath.Join("workshop", "plans")
				if err := os.MkdirAll(plans, 0755); err != nil {
					t.Fatal(err)
				}
				dir := t.TempDir()
				ready := filepath.Join(dir, "reviewer.pid")
				script := "#!/bin/sh\nexport SDLC_SIGNAL_ROLE=reviewer\nexec \"$SDLC_SIGNAL_BINARY\" -test.run='^TestCLISignalProcess$'\n"
				if err := os.WriteFile(filepath.Join(dir, "claude"), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
				args := []string{"judge", "plan-quality", "--agent", "claude", "--base", "HEAD^", "--head", "HEAD"}
				if verb == "close" {
					args = []string{"close", "--issue", "69", "--actual", "1", "--verified", "signal fixture", "--no-atlas", "--agent", "claude", "--brain-dir", "../nonexistent-brain"}
				}
				if verb == "change-code" {
					args = []string{"change-code", "--issue", "69", "--flow", "full", "--no-structural", "--no-estimate", "--no-estimate-recon", "--agent", "claude", "--worktree", "no", "--force", "signal cancellation is not waivable"}
				}
				raw, _ := json.Marshal(args)
				cmd := exec.Command(os.Args[0], "-test.run=^TestCLISignalProcess$")
				cmd.WaitDelay = 2 * time.Second
				cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "SDLC_SIGNAL_ROLE=cli", "SDLC_SIGNAL_ARGS="+string(raw), "SDLC_SIGNAL_BINARY="+os.Args[0], "SDLC_SIGNAL_READY="+ready, "GORACE=atexit_sleep_ms=0", "AGENT_CMD=claude")
				var output bytes.Buffer
				cmd.Stdout = &output
				cmd.Stderr = &output
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
				done := make(chan error, 1)
				go func() { done <- cmd.Wait() }()
				reviewer := 0
				finished := false
				t.Cleanup(func() {
					if reviewer > 0 {
						_ = syscall.Kill(-reviewer, syscall.SIGKILL)
					}
					if !finished {
						_ = cmd.Process.Kill()
						select {
						case <-done:
						case <-time.After(3 * time.Second):
							t.Error("CLI cleanup did not finish")
						}
					}
				})
				timer := time.NewTimer(20 * time.Second)
				defer timer.Stop()
				tick := time.NewTicker(10 * time.Millisecond)
				defer tick.Stop()
				for reviewer == 0 {
					select {
					case err := <-done:
						finished = true
						t.Fatalf("CLI exited before review: %v\n%s", err, output.String())
					case <-timer.C:
						t.Fatal("CLI reviewer did not start")
					case <-tick.C:
						if raw, err := os.ReadFile(ready); err == nil && len(raw) > 0 {
							reviewer, err = strconv.Atoi(string(raw))
							if err != nil {
								t.Fatal(err)
							}
						}
					}
				}
				issueBefore := readIssue(t, issues)
				plansBefore := reviewPlanFiles(t, plans)
				if err := cmd.Process.Signal(sig); err != nil {
					t.Fatal(err)
				}
				select {
				case err := <-done:
					finished = true
					if err == nil {
						t.Error("signal returned success")
					}
				case <-time.After(8 * time.Second):
					t.Fatal("CLI did not complete bounded cancellation")
				}
				if !strings.Contains(output.String(), "interrupted") && !strings.Contains(output.String(), "context canceled") {
					t.Errorf("missing cancellation diagnostic:\n%s", output.String())
				}
				if err := syscall.Kill(reviewer, 0); err != syscall.ESRCH {
					t.Errorf("reviewer was not reaped before CLI returned: %v", err)
				}
				if after := readIssue(t, issues); after != issueBefore {
					t.Error("signal changed issue")
				}
				after := reviewPlanFiles(t, plans)
				if len(after) != len(plansBefore) {
					t.Errorf("signal wrote review artifacts: before=%v after=%v", plansBefore, after)
				}
				for p, data := range plansBefore {
					if after[p] != data {
						t.Errorf("signal changed %s", p)
					}
				}
			})
		}
	}
}

func TestCLISignalOwnershipKeepsLockUntilUnwind(t *testing.T) {
	for _, role := range []string{"lock-owner", "legacy-lock"} {
		t.Run(role, func(t *testing.T) {
			closeRepo(t, 69)
			dir := t.TempDir()
			ready := filepath.Join(dir, "ready")
			proof := filepath.Join(dir, "proof")
			cmd := exec.Command(os.Args[0], "-test.run=^TestCLISignalProcess$")
			cmd.Env = append(os.Environ(), "SDLC_SIGNAL_ROLE="+role, "SDLC_SIGNAL_READY="+ready, "SDLC_SIGNAL_PROOF="+proof, "GORACE=atexit_sleep_ms=0")
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			finished := false
			t.Cleanup(func() {
				if !finished {
					_ = cmd.Process.Kill()
					<-done
				}
			})
			deadline := time.NewTimer(5 * time.Second)
			defer deadline.Stop()
			tick := time.NewTicker(10 * time.Millisecond)
			defer tick.Stop()
			waiting := true
			for waiting {
				select {
				case err := <-done:
					finished = true
					t.Fatalf("CLI exited early: %v %s", err, out.String())
				case <-deadline.C:
					t.Fatal("lock did not become ready")
				case <-tick.C:
					if _, err := os.Stat(ready); err == nil {
						waiting = false
					}
				}
			}
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				finished = true
				if err == nil {
					t.Error("signal returned success")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("signal cleanup did not finish")
			}
			_, err := os.Stat(proof)
			if role == "lock-owner" && err != nil {
				t.Fatalf("managed CLI exited before guarded unwind: %v\n%s", err, out.String())
			}
			if role == "legacy-lock" && !os.IsNotExist(err) {
				t.Fatalf("legacy command lost immediate cleanup behavior: %v", err)
			}
			common, err := repoLockGitCommonDir()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(common, "sdlc.lock")); !os.IsNotExist(err) {
				t.Errorf("signal leaked repository lock: %v", err)
			}
		})
	}
}

type cancelDuringCloseSnapshot struct {
	gitRunner
	cancel context.CancelFunc
}

func (r cancelDuringCloseSnapshot) GitInDir(dir string, args ...string) ([]byte, error) {
	r.cancel()
	return r.gitRunner.GitInDir(dir, args...)
}

func TestCLICancelledCloseBypassDoesNotWrite(t *testing.T) {
	for _, when := range []string{"before", "during-peer-read"} {
		t.Run(when, func(t *testing.T) {
			issues := closeRepo(t, 69)
			root, _ := os.Getwd()
			before := readIssue(t, issues)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f := closeFlagsFor(issues)
			f.Context = ctx
			f.Force = true
			f.NoActual = true
			res := closeResult{issuePath: filepath.Join(issues, "000069-x.md"), issueText: before, newIssueText: before + "\nclosed\n", repoTop: root}
			var runner gitRunner = execGitRunner{}
			if when == "before" {
				cancel()
			} else {
				peer := testfix.Repo(t, testfix.InitialCommit())
				path := filepath.Join(peer, "project.md")
				if err := os.WriteFile(path, []byte("before"), 0600); err != nil {
					t.Fatal(err)
				}
				res.projectEdits = []projectEdit{{path: path, repoDir: peer, oldText: "before", newText: "after"}}
				runner = cancelDuringCloseSnapshot{gitRunner: runner, cancel: cancel}
			}
			msg, died := expectDie(t, func() { applyClose(io.Discard, io.Discard, runner, f, res) })
			if !died || !strings.Contains(msg, "interrupted") {
				t.Fatalf("cancelled bypass continued: died=%v message=%s", died, msg)
			}
			if got := readIssue(t, issues); got != before {
				t.Fatal("cancelled bypass changed issue")
			}
			for _, p := range res.projectEdits {
				raw, err := os.ReadFile(p.path)
				if err != nil || string(raw) != p.oldText {
					t.Fatalf("cancelled bypass changed peer: %s %v", raw, err)
				}
			}
		})
	}
}
