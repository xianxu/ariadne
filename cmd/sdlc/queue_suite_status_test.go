package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

const suiteQueueFixture = `{"version":1,"available":true,"policy_seconds":3600,"protected_through":2,"priority_overridden":true,"tickets":[{"number":1,"position":1,"age_seconds":null,"protected":true,"granted":false,"priority_overridden":false,"reason":"protected_prefix","record":{"repo":"first","future":{"nested":true}}},{"number":2,"position":2,"age_seconds":3600,"protected":true,"granted":false,"priority_overridden":false,"reason":"aged_fifo","record":{"repo":"old"}},{"number":3,"position":3,"age_seconds":20,"protected":false,"granted":true,"priority_overridden":true,"reason":"grant_priority","record":{"repo":"granted"}}]}`

func suiteQueueHelperFixture(t *testing.T, script string) string {
	t.Helper()
	supplier := t.TempDir()
	if err := os.Mkdir(filepath.Join(supplier, "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"Makefile.workflow": "# supplier\n", "scripts/test-lock.py": script} {
		if err := os.WriteFile(filepath.Join(supplier, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	consumer := t.TempDir()
	if err := os.Symlink(filepath.Join(supplier, "Makefile.workflow"), filepath.Join(consumer, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	return consumer
}

func TestSuiteQueueObservationUsesSupplier(t *testing.T) {
	root := suiteQueueHelperFixture(t, "import sys\nassert sys.argv[1:] == ['inspect', '--supervised']\nprint("+quoteHelperShell(suiteQueueFixture)+")\n")
	got := observeSuiteQueue(context.Background(), root)
	if !got.Available || len(got.Tickets) != 3 || got.PolicySeconds != 3600 {
		t.Fatalf("observation: %+v", got)
	}
	var out bytes.Buffer
	renderSuiteQueue(&out, got)
	for _, want := range []string{"3600", "60", "ticket 1", "20s", "aged_fifo", "aging overrode grant", "age unknown"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %s", want, out.String())
		}
	}
}

func TestSuiteQueueDecodeRefusesIncompleteProjection(t *testing.T) {
	for _, raw := range []string{
		`{}`, `{"version":2,"available":true,"policy_seconds":3600,"tickets":[]}`,
		strings.Replace(suiteQueueFixture, `"position":2`, `"position":1`, 1),
		strings.Replace(suiteQueueFixture, `"number":2`, `"number":1`, 1),
		strings.Replace(suiteQueueFixture, `"age_seconds":3600`, `"age_seconds":-1`, 1),
		strings.Replace(suiteQueueFixture, `"policy_seconds":3600`, `"policy_seconds":0`, 1),
		suiteQueueFixture + ` {}`, strings.Replace(suiteQueueFixture, `"record":{"repo":"old"}`, `"record":null`, 1),
	} {
		if _, err := decodeSuiteQueue([]byte(raw)); err == nil {
			t.Errorf("accepted malformed observation %s", raw)
		}
	}
	// Additive observations do not need a schema bump.
	raw := strings.Replace(suiteQueueFixture, `"version":1`, `"future":{"nested":[1]},"version":1`, 1)
	if _, err := decodeSuiteQueue([]byte(raw)); err != nil {
		t.Fatal(err)
	}
}

func TestSuiteQueueObservationUnavailable(t *testing.T) {
	for _, tc := range []struct{ name, script string }{
		{"old", "import sys\nsys.exit('unknown command inspect')\n"},
		{"malformed", "print('{}')\n"},
		{"stdout overflow", "print('x' * (2 << 20))\n"},
		{"stderr overflow", "import sys\nsys.stderr.write('x' * (2 << 20))\n"},
		{"changed", `print('{"version":1,"available":false,"policy_seconds":3600,"reason":"snapshot_changed","tickets":[]}')`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := suiteQueueHelperFixture(t, tc.script)
			got := observeSuiteQueue(context.Background(), root)
			if got.Available || got.Reason == "" {
				t.Fatalf("should be unavailable: %+v", got)
			}
			var out bytes.Buffer
			renderSuiteQueue(&out, got)
			if !strings.Contains(out.String(), "unavailable") || !strings.Contains(out.String(), "sdlc queue status") {
				t.Fatalf("missing recovery %s", out.String())
			}
		})
	}
	got := observeSuiteQueue(context.Background(), t.TempDir())
	if got.Available || got.Reason == "" {
		t.Fatalf("missing supplier: %+v", got)
	}
}

func TestSuiteQueueStatusTextAndFrozenJSON(t *testing.T) {
	d := queueFixture(t)
	projection, err := decodeSuiteQueue([]byte(suiteQueueFixture))
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	d.suite = func(context.Context) suiteQueueObservation { calls++; return projection }
	command := newQueueCmd(d)
	var out bytes.Buffer
	command.SetOut(&out)
	command.SetArgs([]string{"status"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !strings.Contains(out.String(), "aging overrode grant") {
		t.Fatalf("text projection absent: %s", out.String())
	}
	raw, err := executeQueue(t, d, "status", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("JSON queried supplemental scheduler")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["tickets"]; ok {
		t.Fatal("frozen JSON expanded")
	}
}

func TestSuiteQueueObservationCancellation(t *testing.T) {
	root := suiteQueueHelperFixture(t, "import time\ntime.sleep(60)\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	got := observeSuiteQueue(ctx, root)
	if got.Available || !strings.Contains(got.Reason, "canceled") {
		t.Fatalf("cancellation: %+v", got)
	}
	if time.Since(start) > time.Second {
		t.Fatal("cancelled observation started slow work")
	}
}

func TestSuiteQueueStatusObservationsOverlapAndJoin(t *testing.T) {
	d := queueFixture(t)
	projection, err := decodeSuiteQueue([]byte(suiteQueueFixture))
	if err != nil {
		t.Fatal(err)
	}
	suiteStarted, legacyStarted := make(chan struct{}), make(chan struct{})
	d.suite = func(ctx context.Context) suiteQueueObservation {
		close(suiteStarted)
		select {
		case <-legacyStarted:
			return projection
		case <-ctx.Done():
			return suiteQueueObservation{Reason: ctx.Err().Error()}
		}
	}
	d.legacy = func(ctx context.Context, _ []queuegate.Operation) ([]queuegate.Activity, []string) {
		close(legacyStarted)
		select {
		case <-suiteStarted:
			return nil, nil
		case <-ctx.Done():
			return nil, []string{"observations serialized"}
		}
	}
	cmd := newQueueCmd(d)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "aging overrode grant") || strings.Contains(out.String(), "observations serialized") {
		t.Fatalf("observation lifetime: %s", out.String())
	}
}

func TestSuiteQueueInspectionCompatibilityDiagnostic(t *testing.T) {
	raw, err := os.ReadFile("testdata/suite-inspect-compat/pre560.stderr")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, diagnostic string
		code             int
		unsupported      bool
	}{
		{"captured acquire-only helper", string(raw), 2, true},
		{"missing supervised option", "usage: test-lock.py\ntest-lock.py: error: unrecognized arguments: --supervised\n", 2, true},
		{"other usage error", "test-lock.py: error: invalid slots value\n", 2, false},
		{"runtime failure", "provider could not read state\n", 1, false},
		{"misleading runtime text", string(raw), 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := suiteQueueHelperFixture(t, "import sys\nsys.stderr.write("+fmt.Sprintf("%q", tc.diagnostic)+")\nsys.exit("+fmt.Sprint(tc.code)+")\n")
			helper := filepath.Join(filepath.Dir(mustSuiteWorkflow(t, root)), "scripts", "test-lock.py")
			got := observeSuiteQueue(context.Background(), root)
			if got.Available || !strings.Contains(got.Reason, helper) || !strings.Contains(got.Reason, strings.TrimSpace(tc.diagnostic)) {
				t.Fatalf("lost helper diagnostic: %+v", got)
			}
			unsupported := strings.Contains(got.Reason, "does not support queue inspection")
			if unsupported != tc.unsupported {
				t.Fatalf("unsupported=%v want %v: %s", unsupported, tc.unsupported, got.Reason)
			}
			if unsupported && !strings.Contains(got.Reason, "update the supplying ariadne checkout") {
				t.Fatal("missing supplier recovery")
			}
		})
	}
}

func mustSuiteWorkflow(t *testing.T, root string) string {
	t.Helper()
	mk, err := filepath.EvalSymlinks(filepath.Join(root, "Makefile.workflow"))
	if err != nil {
		t.Fatal(err)
	}
	return mk
}

func TestSuiteQueueInspectionDiagnosticBound(t *testing.T) {
	for _, n := range []int{1024, 2048} {
		got := suiteQueueInspectFailure("/supplier/scripts/test-lock.py", errors.New("run failure"), []byte(strings.Repeat("x", n))).Error()
		if strings.Count(got, "x") != 1024 || (strings.Contains(got, "[truncated]") != (n > 1024)) {
			t.Fatalf("diagnostic bound at %d: %s", n, got)
		}
	}
}

func TestSuiteQueueAcquireOnlySupplierStillAdmits(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "acquire-argv")
	script := `import argparse, json, sys
p=argparse.ArgumentParser()
p.add_argument('cmd',choices=['acquire'])
p.add_argument('--watch',type=int,required=True)
p.add_argument('--slots',type=int,default=1)
a=p.parse_args()
open(` + fmt.Sprintf("%q", marker) + `, 'w').write(json.dumps(sys.argv[1:]))
print('acquired')
`
	root := suiteQueueHelperFixture(t, script)
	helper := filepath.Join(filepath.Dir(mustSuiteWorkflow(t, root)), "scripts", "test-lock.py")
	observed := observeSuiteQueue(context.Background(), root)
	if observed.Available || !strings.Contains(observed.Reason, "does not support queue inspection") {
		t.Fatalf("old helper status: %+v", observed)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("status attempted admission: %v", err)
	}
	d := localRunDeps{Root: root, Stderr: io.Discard}
	held, release, err := acquireSuiteSlot(context.Background(), d, root, 2, helper)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if held == "" {
		t.Fatal("no lifetime holder")
	}
	raw, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	var argv []string
	if err := json.Unmarshal(raw, &argv); err != nil {
		t.Fatal(err)
	}
	if len(argv) != 5 || argv[0] != "acquire" || argv[1] != "--watch" || argv[3] != "--slots" || argv[4] != "2" {
		t.Fatalf("admission argv changed: %v", argv)
	}
}
