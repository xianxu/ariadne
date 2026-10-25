package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCouchResumeReceiptConformance checks the deployed Couch CLI's polling,
// JSON receipt reporting and exit mapping, then gives its actual stdout and
// process facts to our classifier. Only the broker is fake: no supervisor,
// agent, visible terminal or real slot is started. Run at every tl-lead close
// that changes resume and whenever upgrading the supported Couch binary:
// TL_LEAD_RESUME_CONFORMANCE=1 go test ./cmd/tl-lead -run '^TestCouchResumeReceiptConformance$' -count=1 -v
func TestCouchResumeReceiptConformance(t *testing.T) {
	if os.Getenv("TL_LEAD_RESUME_CONFORMANCE") != "1" {
		t.Skip("set TL_LEAD_RESUME_CONFORMANCE=1 for isolated real-Couch receipt conformance")
	}
	bin, err := exec.LookPath("couch")
	if err != nil {
		t.Fatal("opted into conformance but couch is not on PATH:", err)
	}
	t.Logf("real Couch binary: %s", bin)
	for _, status := range []string{"succeeded", "refused", "failed"} {
		t.Run(status, func(t *testing.T) {
			env, store, root := resumeConformanceEnvironment(t)
			done := resumeConformanceBroker(t, store, status)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, bin, "--resume", "receipt-conformance:7", "--json")
			cmd.Env, cmd.Dir = env, root
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			runErr := cmd.Run()
			r := resumeResult{stdout: stdout.String(), stderr: stderr.String(), started: cmd.Process != nil, timedOut: ctx.Err() != nil, exitCode: -1}
			if runErr != nil {
				r.processError = runErr.Error()
			}
			if cmd.ProcessState != nil {
				r.normal = cmd.ProcessState.Exited()
				r.exitCode = cmd.ProcessState.ExitCode()
			}
			wantExit := 1
			if status == "succeeded" {
				wantExit = 0
			}
			if r.timedOut || !r.normal || r.exitCode != wantExit {
				t.Fatalf("real Couch exit=%d normal=%v timeout=%v; want%d; stdout=%q stderr=%q", r.exitCode, r.normal, r.timedOut, wantExit, r.stdout, r.stderr)
			}
			var receipt struct{ ID, Op, Target, Status, Detail, Warning, Code, Diagnostic string }
			if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil || receipt.ID == "" || receipt.Op != "resume" || receipt.Target != "receipt-conformance:7" || receipt.Status != status || receipt.Detail != "isolated broker terminal evidence" || receipt.Warning != "conformance advisory" {
				t.Fatalf("real CLI did not produce terminal standalone receipt: %+v err=%v stdout=%q stderr=%q", receipt, err, r.stdout, r.stderr)
			}
			if strings.Contains(r.stdout, "Operation") || strings.Contains(r.stdout, "queued") || strings.Contains(r.stdout, "running") {
				t.Fatalf("CLI leaked transport envelope or intermediate receipt: %s", r.stdout)
			}
			classified := classifyResumeResult("receipt-conformance:7", r)
			if status == "succeeded" {
				if classified != nil {
					t.Fatalf("actual success classified as %v", classified)
				}
			} else if classified == nil || errors.Is(classified, errUncertain) || !strings.Contains(classified.Error(), status) {
				t.Fatalf("actual terminal %s classified as %v", status, classified)
			}
			if status != "succeeded" {
				var coded codedResumeFailure
				if !errors.As(classified, &coded) {
					t.Fatalf("real terminal failure lost its stable identity: %v", classified)
				}
				key, stable := coded.failureKey()
				if !stable || key != (resumeFailureKey{Status: status, Code: "fixture-terminal", Diagnostic: "isolated stable cause"}) || receipt.Code != key.Code || receipt.Diagnostic != key.Diagnostic {
					t.Fatalf("real receipt failure identity: receipt=%+v key=%+v stable=%v", receipt, key, stable)
				}
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatal("isolated broker:", err)
				}
			case <-ctx.Done():
				t.Fatal("broker did not observe resume then running and terminal status polls")
			}
			t.Logf("real Couch: queued -> running -> %s; standalone receipt; exit %d; classification confirmed", status, r.exitCode)
		})
	}
}

// Supply the whole child environment: inherited live store, slot identity,
// XDG and singleton overrides cannot select production state. Scratch selection
// models Couch's documented isolated runtime, without launching its supervisor.
func resumeConformanceEnvironment(t *testing.T) (env []string, store, root string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store = filepath.Join(root, "store")
	data, identity := filepath.Join(root, "data", "pair"), filepath.Join(root, "identity")
	for _, dir := range []string{store, data, identity, filepath.Join(root, "home"), filepath.Join(root, "tmp"), filepath.Join(root, "singleton")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	selection := map[string]any{"version": 1, "roots": map[string]string{"store_dir": store, "pair_data_dir": data, "identity_dir": identity}}
	raw, err := json.Marshal(selection)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "singleton", "selection.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + filepath.Join(root, "home"), "TMPDIR=" + filepath.Join(root, "tmp"), "XDG_DATA_HOME=" + filepath.Join(root, "data"), "COUCH_ISOLATED_ROOT=" + root, "COUCH_STORE_DIR=" + store, "COUCH_PAIR_DATA_DIR=" + data, "COUCH_IDENTITY_DIR=" + identity, "COUCH_THREAD_SCOPE=receipt-scope", "COUCH_THREAD_TAG=receipt-tag", "PAIR_SESSION_NAME=receipt-session", "PAIR_LAUNCH_NONCE=receipt-nonce"}
	return env, store, root
}

// This is the external couchmessage.SocketPath/framing contract, deliberately
// independent of our classifier. Couch still validates the request, owns the
// request ID, polls and turns these broker envelopes into its public output.
func resumeConformanceBroker(t *testing.T, store, status string) <-chan error {
	t.Helper()
	digest := sha256.Sum256([]byte(store + "\x00broker"))
	//lint:allow tmp-literal — Couch's cross-process socket contract always uses this UID-private directory, independent of TMPDIR.
	dir := filepath.Join("/tmp", fmt.Sprintf("pair-message-%d", os.Getuid()))
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
		t.Fatalf("unsafe Couch socket directory: %v %v", info, err)
	}
	socket := filepath.Join(dir, fmt.Sprintf("%x.sock", digest[:20]))
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: socket})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	if err := os.Chmod(socket, 0600); err != nil {
		t.Fatal(err)
	}
	if err := listener.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	finished := make(chan struct{})
	go func() { defer close(finished); done <- serveResumeConformance(listener, status) }()
	t.Cleanup(func() { listener.Close(); <-finished })
	return done
}

func serveResumeConformance(listener *net.UnixListener, status string) error {
	var id string
	for step := 0; step < 3; step++ {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return err
		}
		err = func() error {
			defer conn.Close()
			if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				return err
			}
			var header [4]byte
			if _, err := io.ReadFull(conn, header[:]); err != nil {
				return err
			}
			size := binary.BigEndian.Uint32(header[:])
			if size == 0 || size > 64<<10 {
				return fmt.Errorf("unexpected request frame size %d", size)
			}
			raw := make([]byte, int(size))
			if _, err := io.ReadFull(conn, raw); err != nil {
				return err
			}
			var req struct{ Op, Scope, Tag, Session, Nonce, ID, Target string }
			if err := json.Unmarshal(raw, &req); err != nil {
				return err
			}
			if req.Scope != "receipt-scope" || req.Tag != "receipt-tag" || req.Session != "receipt-session" || req.Nonce != "receipt-nonce" || req.ID == "" {
				return fmt.Errorf("unexpected caller identity: %+v", req)
			}
			if step == 0 {
				if req.Op != "resume" || req.Target != "receipt-conformance:7" {
					return fmt.Errorf("expected isolated resume: %+v", req)
				}
				id = req.ID
			} else if req.Op != "operation-status" || req.ID != id || req.Target != "" {
				return fmt.Errorf("expected status poll for %s: %+v", id, req)
			}
			// The transport response is never the CLI's standalone receipt.
			// Each intermediate state forces the real CLI to poll again.
			code, state := "ok", status
			if step == 0 {
				code, state = "accepted", "queued"
			}
			if step == 1 {
				state = "running"
			}
			reply := map[string]any{"Code": code, "Operation": map[string]any{
				"ID": id, "Op": "resume", "Target": "receipt-conformance:7", "Status": state,
				"Detail": "isolated broker terminal evidence", "Warning": "conformance advisory",
				"Code": "fixture-terminal", "Diagnostic": "isolated stable cause",
			}}
			body, err := json.Marshal(reply)
			if err != nil {
				return err
			}
			binary.BigEndian.PutUint32(header[:], uint32(len(body)))
			_, err = io.Copy(conn, io.MultiReader(bytes.NewReader(header[:]), bytes.NewReader(body)))
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
