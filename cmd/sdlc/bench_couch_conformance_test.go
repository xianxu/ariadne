package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestBenchBridgeReplacementEnvironment(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "owned", "bridge")
	ready := benchBridgeReadiness{Version: 1, Namespace: filepath.Join(root, "store"), Socket: filepath.Join(root, "sock"), Environment: map[string]string{"COUCH_ISOLATED_ROOT": root, "COUCH_STORE_DIR": filepath.Join(root, "store"), "PATH": filepath.Join(root, "bin"), "HOME": filepath.Join(root, "home")}}
	raw, _ := json.Marshal(ready)
	env, _, err := benchBridgeEnvironment(append(raw, '\n'))
	if err != nil || len(env) != 4 {
		t.Fatalf("replacement environment: %v %v", env, err)
	}
	for _, change := range []func(*benchBridgeReadiness){
		func(r *benchBridgeReadiness) { r.Version = 2 },
		func(r *benchBridgeReadiness) {
			r.Environment["PATH"] += string(filepath.ListSeparator) + "/ambient/bin"
		},
		func(r *benchBridgeReadiness) { r.Environment["COUCH_STORE_DIR"] = "/ambient/store" },
		func(r *benchBridgeReadiness) { r.Environment["BAD=KEY"] = "bad" },
	} {
		var broken benchBridgeReadiness
		_ = json.Unmarshal(raw, &broken)
		change(&broken)
		input, _ := json.Marshal(broken)
		if _, _, err := benchBridgeEnvironment(append(input, '\n')); err == nil {
			t.Fatalf("accepted invalid environment %s", input)
		}
	}
	for _, input := range [][]byte{[]byte("{}\n{}\n"), make([]byte, (64<<10)+1), []byte(`{"Version":1,"Version":1}`)} {
		if _, _, err := benchBridgeEnvironment(input); err == nil {
			t.Fatalf("accepted invalid readiness")
		}
	}
}

// The bridge owns the supervisor, its native children, and their storage. This
// test owns only the client and the control pipe; EOF + successful Wait is the
// bridge's cleanup receipt. Never delete its scratch on an unconfirmed stop.
func TestBenchCouchNativeServiceConformance(t *testing.T) {
	bridge := os.Getenv("BENCH_COUCH_CONFORMANCE_BRIDGE")
	if bridge == "" {
		t.Skip("set BENCH_COUCH_CONFORMANCE_BRIDGE to Pair's native bridge executable")
	}
	if !filepath.IsAbs(bridge) || filepath.Clean(bridge) != bridge {
		t.Fatal("bridge executable must be an absolute clean path")
	}
	// Pair's native wrapper socket leaves at most 22 bytes for this parent.
	// The bridge itself creates and owns fresh candidate/fleet directories; an
	// additional MkdirTemp nesting cannot fit even on the RAM scratch volume.
	scratch := os.Getenv("BENCH_COUCH_CONFORMANCE_SCRATCH_PARENT")
	explicitScratch := scratch != ""
	if !explicitScratch {
		scratch = os.TempDir()
		if info, err := os.Stat("/Volumes/TestScratch"); err == nil && info.IsDir() {
			scratch = "/Volumes/TestScratch"
		}
	}
	physicalScratch, err := filepath.EvalSymlinks(scratch)
	if err != nil {
		t.Fatal(err)
	}
	if explicitScratch && (scratch != physicalScratch || !filepath.IsAbs(scratch) || filepath.Clean(scratch) != scratch) {
		t.Fatal("explicit scratch parent must be an existing physical absolute directory")
	}
	scratch = physicalScratch
	info, err := os.Stat(scratch)
	if err != nil || !info.IsDir() {
		t.Fatal("scratch parent must be an existing directory")
	}
	if len(scratch) > 22 {
		t.Fatal("native socket path requires an existing short BENCH_COUCH_CONFORMANCE_SCRATCH_PARENT")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 330*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bridge, "--scratch-parent", scratch)
	// The bridge builds its candidate. Its native owner subsequently replaces the
	// environment again. No live-fleet COUCH authority enters either client.
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "COUCH_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	processgroup.Configure(cmd)
	join := configureRunnerCleanup(cmd)
	control, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout, cmd.Stderr = stdoutWriter, stderrWriter
	defer stdout.Close()
	defer stderr.Close()
	defer stdoutWriter.Close()
	defer stderrWriter.Close()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	var diagnostics benchBridgeDiagnostic
	diagDone := make(chan struct{})
	go func() { _, _ = io.Copy(&diagnostics, stderr); close(diagDone) }()
	type observation struct {
		line []byte
		err  error
	}
	observed := make(chan observation, 1)
	stdoutDone := make(chan struct{})
	var tail benchBridgeDiagnostic
	var tailErr error
	reader := bufio.NewReaderSize(stdout, 64<<10)
	go func() {
		defer close(stdoutDone)
		line, err := reader.ReadSlice('\n')
		observed <- observation{append([]byte(nil), line...), err}
		_, tailErr = io.Copy(&tail, reader)
	}()
	joined := false
	isolated := "unknown"
	stop := func() error {
		joined = true
		_ = control.Close()
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		var waitErr error
		select {
		case waitErr = <-done:
		case <-time.After(25 * time.Second):
			cancel()
			waitErr = fmt.Errorf("bridge cleanup unconfirmed; retained scratch %s (isolated %s)", scratch, isolated)
			select {
			case err := <-done:
				waitErr = errors.Join(waitErr, err)
			case <-time.After(3 * time.Second):
				return waitErr
			}
		}
		joined = true
		waitErr = errors.Join(waitErr, join())
		select {
		case <-diagDone:
		case <-time.After(3 * time.Second):
			return fmt.Errorf("bridge diagnostic drain unconfirmed; retain %s", scratch)
		}
		select {
		case <-stdoutDone:
		case <-time.After(3 * time.Second):
			return fmt.Errorf("bridge stdout drain unconfirmed; retain %s", scratch)
		}
		t.Logf("native bridge diagnostics:\n%s", diagnostics.String())
		if waitErr != nil {
			return fmt.Errorf("native cleanup unconfirmed; retain owned bridge artifacts under %s (isolated %s): %w", scratch, isolated, waitErr)
		}
		if tailErr != nil || tail.Len() != 0 {
			return fmt.Errorf("extra bridge stdout %q: %v", tail.String(), tailErr)
		}
		if isolated != "unknown" {
			if _, err := os.Lstat(isolated); !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("native root remains %s: %v", isolated, err)
			}
		}
		return nil
	}
	defer func() {
		if !joined {
			if err := stop(); err != nil {
				t.Error(err)
			}
		}
	}()
	var line []byte
	select {
	case result := <-observed:
		if result.err != nil {
			t.Fatalf("bridge readiness: %v", result.err)
		}
		line = result.line
	case <-ctx.Done():
		t.Fatal("bridge readiness deadline")
	}
	env, ready, err := benchBridgeEnvironment(line)
	if err != nil {
		t.Fatal(err)
	}
	isolated = ready.Environment["COUCH_ISOLATED_ROOT"]
	if isolated == scratch || !benchPathWithin(scratch, isolated) {
		t.Fatal("bridge root outside owned scratch")
	}
	couch := filepath.Join(isolated, "bin", "couch")
	deps := benchCouchDeps{Command: func(call context.Context, args ...string) ([]byte, error) {
		bounded, stop := context.WithTimeout(call, 5*time.Second)
		defer stop()
		child := exec.CommandContext(bounded, couch, args...)
		child.Env = env
		child.Dir = isolated
		processgroup.Configure(child)
		cleanup := configureRunnerCleanup(child)
		var out, diag benchBoundedOutput
		child.Stdout = &out
		child.Stderr = &diag
		err := errors.Join(child.Run(), cleanup(), bounded.Err())
		if err != nil {
			return out.Bytes(), fmt.Errorf("native Couch request %v: %w: %s", args, err, diag.String())
		}
		return out.Bytes(), nil
	}}
	actors, err := benchActors(ctx, deps)
	if err != nil {
		t.Fatal(err)
	}
	for _, actor := range actors {
		if actor.Binding.Repository != "" && !benchPathWithin(isolated, actor.Binding.Repository) {
			t.Fatalf("ambient actor leaked: %+v", actor.Binding)
		}
	}
	// Read only: an unknown receipt must never authorize dispatch or cancellation.
	raw, receiptErr := deps.command(ctx, "--message-status", "bench-conformance-absent", "--json")
	var receipt struct {
		Code, Error string
		Receipt     *benchMessageReceipt
	}
	if err := readBenchCouchJSON(raw, &receipt); err != nil {
		t.Fatalf("native receipt JSON: %s (%v)", raw, err)
	}
	if receipt.Code == "ok" || receipt.Error == "" || receipt.Receipt != nil {
		t.Fatalf("unknown receipt was not refused: %s (%v)", raw, receiptErr)
	}
	if _, err := readBenchDelivery(raw, benchRuntimeActor{}, "bench-conformance-absent", ""); err == nil {
		t.Fatal("refused receipt authorized delivery")
	}
	t.Logf("native service qualified: actors=%d receipt_code=%s; no agent/model/root replay qualification", len(actors), receipt.Code)
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	t.Logf("native cleanup confirmed: EOF, Wait exit 0, isolated root absent; bridge confirms owned candidate cleanup")
}

type benchBridgeReadiness struct {
	Version           int
	Namespace, Socket string
	Environment       map[string]string
}

func benchBridgeEnvironment(raw []byte) ([]string, benchBridgeReadiness, error) {
	var r benchBridgeReadiness
	bad := func() ([]string, benchBridgeReadiness, error) {
		return nil, r, errors.New("invalid bridge replacement environment")
	}
	if len(raw) > 64<<10 || len(raw) == 0 || raw[len(raw)-1] != '\n' || bytes.Count(raw, []byte{'\n'}) != 1 {
		return bad()
	}
	if err := readBenchCouchJSON(raw, &r); err != nil {
		return nil, r, err
	}
	root := r.Environment["COUCH_ISOLATED_ROOT"]
	clean := func(path string) bool {
		return filepath.IsAbs(path) && filepath.Clean(path) == path && !strings.ContainsRune(path, 0)
	}
	if r.Version != 1 || !clean(root) || root == string(filepath.Separator) || !clean(r.Namespace) || !clean(r.Socket) || !benchPathWithin(root, r.Namespace) || !benchPathWithin(root, r.Socket) || r.Environment["COUCH_STORE_DIR"] != r.Namespace {
		return bad()
	}
	paths := filepath.SplitList(r.Environment["PATH"])
	if len(paths) == 0 || paths[0] != filepath.Join(root, "bin") {
		return bad()
	}
	for _, path := range paths {
		if !clean(path) || !benchPathWithin(root, path) {
			return bad()
		}
	}
	var env []string
	for key, value := range r.Environment {
		if key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(value, 0) {
			return bad()
		}
		for i, c := range key {
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' || (i > 0 && c >= '0' && c <= '9')) {
				return bad()
			}
		}
		env = append(env, key+"="+value)
	}
	sort.Strings(env)
	return env, r, nil
}

type benchBridgeDiagnostic struct{ bytes.Buffer }

func (b *benchBridgeDiagnostic) Write(p []byte) (int, error) {
	n := len(p)
	remaining := (64 << 10) - b.Len()
	if remaining > 0 {
		if remaining > len(p) {
			remaining = len(p)
		}
		_, _ = b.Buffer.Write(p[:remaining])
	}
	return n, nil
}
