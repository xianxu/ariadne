package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type fleetCLIOutput struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	canceled chan struct{}
	once     sync.Once
}

func (w *fleetCLIOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.buf.Write(p)
	if strings.Contains(string(p), "caller canceled; joining admitted calls") {
		w.once.Do(func() { close(w.canceled) })
	}
	return n, err
}
func (w *fleetCLIOutput) String() string { w.mu.Lock(); defer w.mu.Unlock(); return w.buf.String() }

type fleetCLIState struct {
	Agent, Mode                   string
	Argv                          [][]string
	Active, Started, Ended, Awake []string
	MaxActive                     int   `json:"max_active"`
	ResumePIDs                    []int `json:"resume_pids"`
}

// The whole binary runs against a stateful executable with a private portable
// backend. The separate real-Couch receipt test qualifies its wire/exit model;
// this test starts no broker, agent or live fleet and never enables scheduling.
func TestFleetResumeExecutableConformance(t *testing.T) {
	if os.Getenv("TL_LEAD_RESUME_CONFORMANCE") != "1" {
		t.Skip("set TL_LEAD_RESUME_CONFORMANCE=1 and TL_LEAD_RESUME_BIN to a clean private tl-lead build")
	}
	bin, err := filepath.EvalSymlinks(os.Getenv("TL_LEAD_RESUME_BIN"))
	if err != nil || !filepath.IsAbs(bin) {
		t.Fatal("TL_LEAD_RESUME_BIN must name an existing absolute private build")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Fatal("isolated executable conformance needs python3:", err)
	}
	script, err := filepath.Abs("testdata/fleet-couch.py")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ agent, mode string }{
		{"claude", "success"}, {"codex", "success"}, {"qoder", "success"},
		{"codex", "shared"}, {"claude", "lost"}, {"codex", "signal"},
	} {
		t.Run(tc.agent+"/"+tc.mode, func(t *testing.T) {
			env, _, root := resumeConformanceEnvironment(t)
			ops := filepath.Join(root, "ops")
			gitInit(t, ops)
			if err := os.WriteFile(filepath.Join(ops, ".hub"), []byte("fixture\n"), 0600); err != nil {
				t.Fatal(err)
			}
			backend := filepath.Join(root, "fake-couch")
			fakeBin := filepath.Join(root, "bin")
			for _, dir := range []string{backend, fakeBin} {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink(script, filepath.Join(fakeBin, "couch")); err != nil {
				t.Fatal(err)
			}
			// An accidental legacy inventory call must fail inside the fixture,
			// never fall through to the operator's real sdlc executable.
			if err := os.WriteFile(filepath.Join(fakeBin, "sdlc"), []byte("#!/bin/sh\necho unexpected-sdlc-call >&2\nexit 29\n"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(backend, "test-only"), []byte("isolated tl-lead conformance\n"), 0600); err != nil {
				t.Fatal(err)
			}
			seed := map[string]any{"agent": tc.agent, "mode": tc.mode, "argv": []any{}, "pids": []int{}, "resume_pids": []int{}, "active": []string{}, "started": []string{}, "ended": []string{}, "awake": []string{}, "max_active": 0}
			raw, err := json.Marshal(seed)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(backend, "state.json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			for i, entry := range env {
				if strings.HasPrefix(entry, "PATH=") {
					env[i] = "PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH")
				}
			}
			env = append(env, "TL_LEAD_FAKE_COUCH_ROOT="+backend)
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
			defer cancel()
			read := func() fleetCLIState {
				t.Helper()
				var s fleetCLIState
				raw, err := os.ReadFile(filepath.Join(backend, "state.json"))
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(raw, &s); err != nil {
					t.Fatal(err)
				}
				return s
			}
			release := func(addr string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(backend, "release-"+addr), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			stdout := fleetCLIOutput{canceled: make(chan struct{})}
			var stderr bytes.Buffer
			done := make(chan struct{})
			var runErr error
			await := func(check func(fleetCLIState) bool) fleetCLIState {
				t.Helper()
				tick := time.NewTicker(20 * time.Millisecond)
				defer tick.Stop()
				for {
					s := read()
					if check(s) {
						return s
					}
					select {
					case <-done:
						t.Fatalf("private CLI returned before fixture condition: %v\n%s\n%s", runErr, stdout.String(), stderr.String())
					case <-ctx.Done():
						t.Fatalf("owned fixture condition unavailable: %v state=%+v", ctx.Err(), s)
					case <-tick.C:
					}
				}
			}
			cmd := exec.CommandContext(ctx, bin, "fleet", "resume", "--apply", "--concurrency", "2")
			cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = ops, env, &stdout, &stderr
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			go func() { runErr = cmd.Wait(); close(done) }()
			t.Cleanup(func() {
				for n := 1; n <= 4; n++ {
					release(fmt.Sprintf("fixture:%d", n))
				}
				cleanup, stop := context.WithTimeout(context.WithoutCancel(t.Context()), 10*time.Second)
				defer stop()
				select {
				case <-done:
				case <-cleanup.Done():
					_ = cmd.Process.Kill()
					<-done
					t.Error("owned tl-lead did not join after releasing its fake calls")
				}
				tick := time.NewTicker(20 * time.Millisecond)
				defer tick.Stop()
				for {
					alive := false
					for _, pid := range read().ResumePIDs {
						p, err := os.FindProcess(pid)
						if err != nil {
							alive = true
							continue
						}
						probe := p.Signal(syscall.Signal(0))
						if !errors.Is(probe, syscall.ESRCH) && !errors.Is(probe, os.ErrProcessDone) {
							alive = true
						}
					}
					if !alive {
						return
					}
					select {
					case <-cleanup.Done():
						t.Error("owned fake-call absence unconfirmed; no process was stopped")
						return
					case <-tick.C:
					}
				}
			})
			initial := await(func(s fleetCLIState) bool { return len(s.Started) == 2 })
			first := append([]string(nil), initial.Started...)
			sort.Strings(first)
			if strings.Join(first, ",") != "fixture:1,fixture:2" {
				t.Fatal(first)
			}
			if tc.mode == "signal" {
				if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
					t.Fatal(err)
				}
				// The root must stay alive to collect both admitted outcomes.
				select {
				case <-done:
					t.Fatalf("caller signal abandoned admitted calls: %v", runErr)
				case <-stdout.canceled:
				case <-ctx.Done():
					t.Fatal("caller cancellation was not observed:", ctx.Err())
				}
				release("fixture:1")
				release("fixture:2")
			} else if tc.mode == "lost" {
				release("fixture:1")
				await(func(s fleetCLIState) bool {
					n := 0
					for _, a := range s.Argv {
						if len(a) > 0 && a[0] == "--actors" {
							n++
						}
					}
					return n == 2
				})
				release("fixture:2")
			} else {
				release("fixture:2")
				if tc.mode == "shared" { // first stable failure may admit the third pending worker
					await(func(s fleetCLIState) bool { return len(s.Started) == 3 })
					release("fixture:1")
					release("fixture:3")
				} else {
					await(func(s fleetCLIState) bool { return len(s.Started) == 3 })
					release("fixture:1")
					await(func(s fleetCLIState) bool { return len(s.Ended) >= 2 })
					if len(read().Started) != 3 {
						t.Fatal("remaining wave crossed pending barrier")
					}
					release("fixture:3")
					await(func(s fleetCLIState) bool { return len(s.Started) == 4 })
					release("fixture:4")
				}
			}
			select {
			case <-done:
			case <-ctx.Done():
				t.Fatal("private CLI did not finish:", ctx.Err())
			}
			final := read()
			if tc.mode == "success" {
				if runErr != nil || !strings.Contains(stdout.String(), "worker waves 1 and 2 complete") {
					t.Fatalf("%v\n%s\n%s", runErr, &stdout, &stderr)
				}
				// A separate, explicit replay observes cleared marks/live actors.
				replay := exec.CommandContext(ctx, bin, "fleet", "resume", "--apply", "--concurrency", "2")
				replay.Dir, replay.Env = ops, env
				if out, err := replay.CombinedOutput(); err != nil {
					t.Fatalf("explicit replay: %v %s", err, out)
				}
				final = read()
				if len(final.Started) != 4 {
					t.Fatal("explicit replay resumed a live worker twice")
				}
			} else if runErr == nil || !strings.Contains(stdout.String(), "worker waves incomplete") {
				t.Fatalf("partial operation claimed completion: %v\n%s\n%s", runErr, &stdout, &stderr)
			}
			if final.MaxActive != 2 || len(final.Active) != 0 {
				t.Fatalf("resource/retirement: %+v", final)
			}
			for _, argv := range final.Argv {
				if len(argv) == 0 {
					t.Fatal("empty fake call")
				}
				switch argv[0] {
				case "--recover-plan-from-sdlc", "--actors", "--resume-candidates", "--resume":
				default:
					t.Fatalf("unexpected authority: %v", argv)
				}
				if len(argv) > 1 && (argv[0] == "--resume" || argv[0] == "--resume-candidates") && (argv[1] == "fixture:5" || argv[1] == "fixture:6") {
					t.Fatalf("stale/unmarked worker admitted: %v", argv)
				}
			}
			t.Logf("private build=%s agent=%s mode=%s max_active=%d started=%v ended=%v; no live Couch/scheduler driven", bin, tc.agent, tc.mode, final.MaxActive, final.Started, final.Ended)
		})
	}
}
