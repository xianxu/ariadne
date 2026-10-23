package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSuiteQueueLaneRendering(t *testing.T) {
	lane := `,"lane":{"version":1,"available":true,"busy":true,"holder":{"repo":"echo","worktree":"/w/echo","pid":7,"started":"2026-10-23T09:00:00"},"waiters":[` +
		`{"number":4,"position":1,"age_seconds":12,"protected":false,"granted":false,"priority_overridden":false,"reason":"light_fifo","record":{"repo":"parley"}},` +
		`{"number":6,"position":2,"age_seconds":3,"protected":false,"granted":false,"priority_overridden":false,"reason":"light_fifo","record":{"repo":"echo2"}}]}}`
	withLane := strings.TrimSuffix(suiteQueueFixture, "}") + lane
	for name, tc := range map[string]struct{ raw, want string }{
		"held with waiters": {withLane, `held by repo "echo", worktree "/w/echo", since 2026-10-23T09:00:00; 2 waiting`},
		"waiters in order":  {withLane, "light ticket 4: position 1; age 12s; repo \"parley\", worktree \"\"\n  light ticket 6: position 2"},
		"free":              {strings.Replace(strings.Replace(withLane, `"busy":true`, `"busy":false`, 1), `"holder":{"repo":"echo","worktree":"/w/echo","pid":7,"started":"2026-10-23T09:00:00"}`, `"holder":null`, 1), "stopped at 240s): free; 2 waiting"},
		"older helper":      {suiteQueueFixture, "light lane: not supported by this suite helper"},
		"lane unavailable":  {strings.TrimSuffix(suiteQueueFixture, "}") + `,"lane":{"version":1,"available":false,"reason":"record_invalid"}}`, `light lane: observation unavailable: "record_invalid"`},
		"queue unavailable": {`{"version":1,"available":false,"policy_seconds":3600,"reason":"provider_unavailable","tickets":[]` + lane, "2 waiting"},
	} {
		t.Run(name, func(t *testing.T) {
			observed, err := decodeSuiteQueue([]byte(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			renderSuiteQueue(&out, observed)
			if !strings.Contains(out.String(), tc.want) {
				t.Fatalf("missing %q in:\n%s", tc.want, out.String())
			}
		})
	}
	for name, bad := range map[string]string{
		"lane version":      strings.Replace(withLane, `"lane":{"version":1`, `"lane":{"version":2`, 1),
		"lane position":     strings.Replace(withLane, `"number":6,"position":2`, `"number":6,"position":3`, 1),
		"lane duplicate":    strings.Replace(withLane, `"number":6`, `"number":4`, 1),
		"unavailable blank": strings.TrimSuffix(suiteQueueFixture, "}") + `,"lane":{"version":1,"available":false}}`,
	} {
		if _, err := decodeSuiteQueue([]byte(bad)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// The real helper's inspection, as sdlc queue status reads it, names the
// light holder and its waiter (#559 Done-when 3).
func TestSuiteQueueRealHelperShowsLightLane(t *testing.T) {
	source, err := os.ReadFile("../../scripts/test-lock.py")
	if err != nil {
		t.Fatal(err)
	}
	root := suiteQueueHelperFixture(t, string(source))
	t.Setenv("WF_TEST_LOCK_FILE", filepath.Join(t.TempDir(), "lock"))
	t.Setenv("WF_TEST_PRIORITY_PROVIDER", "/nonexistent")
	sleeper := func() int {
		cmd := exec.Command("sleep", "60")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
		return cmd.Process.Pid
	}
	light := func(pid int) *exec.Cmd {
		cmd := exec.Command("python3", "../../scripts/test-lock.py", "acquire", "--watch", strconv.Itoa(pid), "--lane", "light")
		cmd.Env = append(os.Environ(), "WF_TEST_LOCK_OWNER=/w/"+strconv.Itoa(pid))
		return cmd
	}
	holder, waiter := sleeper(), sleeper()
	if out, err := light(holder).Output(); err != nil || strings.TrimSpace(string(out)) != "acquired-light" {
		t.Fatalf("light acquire: %q %v", out, err)
	}
	queued := light(waiter)
	if err := queued.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = queued.Process.Kill(); _ = queued.Wait() })
	var out bytes.Buffer
	for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
		out.Reset()
		renderSuiteQueue(&out, observeSuiteQueue(context.Background(), root))
		if strings.Contains(out.String(), "1 waiting") {
			break
		}
	}
	for _, want := range []string{`held by repo "` + strconv.Itoa(holder) + `", worktree "/w/` + strconv.Itoa(holder) + `"`, "1 waiting", "light ticket 1: position 1", `worktree "/w/` + strconv.Itoa(waiter) + `"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in:\n%s", want, out.String())
		}
	}
}
