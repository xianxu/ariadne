package main

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestClassifyResumeResult(t *testing.T) {
	receipt := func(status, target string) string {
		return fmt.Sprintf(`{"ID":"request-1","Op":"resume","Target":%q,"Status":%q,"Detail":"slot busy","Future":true}`, target, status)
	}
	for _, tc := range []struct {
		name, raw              string
		started, normal        bool
		code                   int
		timeout, uncertain, ok bool
	}{
		{"success", receipt("succeeded", "pair:3"), true, true, 0, false, false, true},
		{"refused", receipt("refused", "pair:3"), true, true, 1, false, false, false},
		{"failed", receipt("failed", "pair:3"), true, true, 1, false, false, false},
		{"missing", "", true, true, 1, false, true, false},
		{"wrong target", receipt("succeeded", "pair:4"), true, true, 0, false, true, false},
		{"wrong exit", receipt("succeeded", "pair:3"), true, true, 1, false, true, false},
		{"refusal exit zero", receipt("refused", "pair:3"), true, true, 0, false, true, false},
		{"running", receipt("running", "pair:3"), true, true, 0, false, true, false},
		{"truncated", `{"ID":"request-1"`, true, true, 1, false, true, false},
		{"signal", receipt("refused", "pair:3"), true, false, -1, false, true, false},
		{"timeout", receipt("succeeded", "pair:3"), true, true, 0, true, true, false},
		{"not attempted", "", false, false, -1, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := classifyResumeResult("pair:3", resumeResult{stdout: tc.raw, stderr: "diagnostic", started: tc.started, normal: tc.normal, exitCode: tc.code, timedOut: tc.timeout})
			if (err == nil) != tc.ok || errors.Is(err, errUncertain) != tc.uncertain {
				t.Fatalf("err=%v; ok=%v uncertain=%v", err, tc.ok, tc.uncertain)
			}
			if err != nil && !strings.Contains(err.Error(), "diagnostic") {
				t.Fatalf("lost stderr: %v", err)
			}
			if tc.started && err != nil && tc.raw != "" && !strings.Contains(err.Error(), tc.raw) {
				t.Fatalf("lost stdout: %v", err)
			}
		})
	}
	if resumeTimeout < 6*time.Minute {
		t.Fatal("outer timeout interrupts Couch's five-minute polling budget")
	}
}

func TestFleetResumeStopsAtFirstFailure(t *testing.T) {
	for _, uncertain := range []bool{false, true} {
		t.Run(fmt.Sprint(uncertain), func(t *testing.T) {
			d, c, cwd := fleetFixture(t)
			if uncertain {
				c.resumeErr["pair:6"] = fmt.Errorf("%w: request req-6", errUncertain)
			}
			var out strings.Builder
			err := fleetResume(t.Context(), &out, d, cwd, true)
			if err == nil {
				t.Fatalf("failure must stop with nonzero: %s", out.String())
			}
			if strings.Join(c.attempts(), ",") != "pair:2,pair:6" {
				t.Fatalf("attempted %v", c.attempts())
			}
			if !strings.Contains(out.String(), "not attempted: pair:8") || strings.Contains(out.String(), "uncertain pair:8") {
				t.Fatalf("missing exact remainder: %s", out.String())
			}
			if uncertain && (!strings.Contains(out.String(), "req-6") || c.actorsReads != 2) {
				t.Fatalf("missing diagnostic/recheck (%d): %s", c.actorsReads, out.String())
			}
		})
	}
}

func TestFleetResumeSerialCompletion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d, c, cwd := fleetFixture(t)
		c.resumeErr = map[string]error{}
		c.delays = map[string]time.Duration{"pair:2": 10 * time.Second, "pair:6": 20 * time.Second, "pair:8": 30 * time.Second}
		done := make(chan error, 1)
		go func() { done <- fleetResume(t.Context(), &strings.Builder{}, d, cwd, true) }()
		synctest.Wait()
		if strings.Join(c.attempts(), ",") != "pair:2" {
			t.Fatalf("didn't wait for first completion: %v", c.attempts())
		}
		//lint:allow test-fixed-sleep — advance synctest virtual time to the modeled operation boundary; no wall-clock wait
		time.Sleep(10 * time.Second)
		synctest.Wait()
		if strings.Join(c.attempts(), ",") != "pair:2,pair:6" {
			t.Fatalf("second completion ordering: %v", c.attempts())
		}
		//lint:allow test-fixed-sleep — advance synctest virtual time to the modeled operation boundary; no wall-clock wait
		time.Sleep(20 * time.Second)
		synctest.Wait()
		if strings.Join(c.attempts(), ",") != "pair:2,pair:6,pair:8" {
			t.Fatalf("third completion ordering: %v", c.attempts())
		}
		select {
		case err := <-done:
			t.Fatalf("completed before third terminal outcome: %v", err)
		default:
		}
		//lint:allow test-fixed-sleep — advance synctest virtual time to the modeled operation boundary; no wall-clock wait
		time.Sleep(30 * time.Second)
		synctest.Wait()
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})
}

func TestExecResumeTerminalReceipt(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "couch")
	raw := `{"ID":"receipt-3","Op":"resume","Target":"pair:3","Status":"refused","Detail":"busy"}`
	script := "#!/bin/sh\n[ \"$1 $2 $3\" = '--resume pair:3 --json' ] || exit 9\nprintf '%s\\n' '" + raw + "'\necho 'receipt detail' >&2\nexit 1\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	err := (execResume{bin: bin}).Resume("pair:3")
	if err == nil || errors.Is(err, errUncertain) || !strings.Contains(err.Error(), "receipt-3") || !strings.Contains(err.Error(), "receipt detail") {
		t.Fatalf("terminal evidence: %v", err)
	}
}

func TestFleetResumeHelpIsOperatorStarted(t *testing.T) {
	root := buildRoot(deps{})
	var out strings.Builder
	root.SetOut(&out)
	root.SetArgs([]string{"fleet", "resume", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "only when the operator asks; never at startup") {
		t.Fatal(out.String())
	}
}

func TestFleetResumeAllIsVisibilityOnly(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	c.resumeErr = map[string]error{}
	var out strings.Builder
	if err := fleetResume(t.Context(), &out, d, cwd, true, true); err != nil {
		t.Fatal(err)
	}
	for _, addr := range []string{"brain:1", "echo:1", "pair:7", "ops-echo:0"} {
		if !strings.Contains(out.String(), addr) {
			t.Fatalf("--all hid %s: %s", addr, out.String())
		}
	}
	if strings.Join(c.attempts(), ",") != "pair:2,pair:6,pair:8" || c.candReads != 5 {
		t.Fatalf("expanded mutation authority: attempts=%v reads=%d", c.attempts(), c.candReads)
	}
}

// FIFO barriers prove the real subprocess remains awaited until its receipt.
func TestExecResumeWaitsForTerminalReceipt(t *testing.T) {
	dir := t.TempDir()
	readyPath, releasePath := filepath.Join(dir, "ready"), filepath.Join(dir, "release")
	for _, p := range []string{readyPath, releasePath} {
		if err := unix.Mkfifo(p, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ready, err := os.OpenFile(readyPath, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer ready.Close()
	release, err := os.OpenFile(releasePath, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer release.Close()
	bin := filepath.Join(dir, "couch")
	script := fmt.Sprintf("#!/bin/sh\nprintf r > '%s'\nread permit < '%s'\nprintf '%%s\\n' '{\"ID\":\"receipt-3\",\"Op\":\"resume\",\"Target\":\"pair:3\",\"Status\":\"succeeded\"}'\n", readyPath, releasePath)
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- (execResume{bin: bin, resume: 3 * time.Second}).Resume("pair:3") }()
	started := make(chan error, 1)
	go func() { _, err := io.ReadFull(ready, make([]byte, 1)); started <- err }()
	select {
	case err := <-started:
		if err != nil {
			t.Fatal(err)
		}
	case err := <-done:
		t.Fatalf("returned before ready barrier: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("subprocess never reached barrier")
	}
	select {
	case err := <-done:
		t.Fatalf("returned without terminal receipt: %v", err)
	default:
	}
	if _, err := release.WriteString("continue\n"); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestFleetResumeUncertainObservationIsDiagnostic(t *testing.T) {
	for _, mode := range []string{"running", "unreadable"} {
		t.Run(mode, func(t *testing.T) {
			d, c, cwd := fleetFixture(t)
			c.resumeErr["pair:6"] = fmt.Errorf("%w: request req-6", errUncertain)
			if mode == "running" {
				c.failureAddsActor = true
			} else {
				c.failureActorsErr = errors.New("recheck denied")
			}
			var out strings.Builder
			if err := fleetResume(t.Context(), &out, d, cwd, true); !errors.Is(err, errUncertain) {
				t.Fatalf("promoted uncertain result: %v", err)
			}
			want := "observed pair:6 running; operation outcome remains uncertain"
			if mode == "unreadable" {
				want = "observation unavailable: recheck denied"
			}
			if !strings.Contains(out.String(), want) || strings.Join(c.attempts(), ",") != "pair:2,pair:6" {
				t.Fatalf("observation changed mutation decision: %s", out.String())
			}
		})
	}
}
