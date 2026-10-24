package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A clean private executable, fake IPC table and fake GitHub, never live IPC or
// production queue state. Run explicitly with WF_SYSV_CLI_BIN=/absolute/sdlc.
func TestSysVPrivateCLIConformance(t *testing.T) {
	binary := os.Getenv("WF_SYSV_CLI_BIN")
	if binary == "" {
		t.Skip("requires a private executable via WF_SYSV_CLI_BIN")
	}
	if !filepath.IsAbs(binary) {
		t.Fatal("private executable must be absolute")
	}
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	t.Setenv("WF_TEST_LOCK_FILE", filepath.Join(t.TempDir(), "suite.lock"))
	t.Setenv("WF_TEST_LOCK_HELD_BY", "")
	t.Setenv("WF_TEST_SCRATCH", t.TempDir())
	d, _, _, _ := localRunRepo(t, "acquired")
	wireLocalCheckSignalCLI(t, d.Root)
	table := filepath.Join(t.TempDir(), "ipc-table")
	t.Setenv("WF_PRIVATE_SYSV_TABLE", table)
	header := "T ID KEY MODE OWNER GROUP CREATOR CGROUP NATTCH SEGSZ CPID LPID ATIME DTIME CTIME\nShared Memory:\n"
	if err := os.WriteFile(table, []byte(header), 0600); err != nil {
		t.Fatal(err)
	}
	bin := os.Getenv("SDLC_QUEUE_BIN")
	if err := os.WriteFile(filepath.Join(bin, "ipcs"), []byte("#!/bin/sh\ncat \"$WF_PRIVATE_SYSV_TABLE\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	statusFile := filepath.Join(t.TempDir(), "statuses.jsonl")
	t.Setenv("WF_PRIVATE_SYSV_STATUSES", statusFile)
	gh := `#!/usr/bin/env python3
import json,os,sys
a=sys.argv[1:]
if a[:2]==['pr','list']: sys.exit(0)
if a[:3]==['api','-X','POST'] and a[3].startswith('repos/fixture/queue/statuses/'):
    fields=dict(v.split('=',1) for v in a[4:] if v!='-f')
    with open(os.environ['WF_PRIVATE_SYSV_STATUSES'],'a') as f: f.write(json.dumps(fields)+'\n')
    print('{}'); sys.exit(0)
sys.exit('unexpected private gh argv: '+repr(a))
`
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(gh), 0700); err != nil {
		t.Fatal(err)
	}
	row := fmt.Sprintf("m 777 0x123 --rw------- %d staff %d staff 0 56 424242 0 no-entry no-entry 01:02:03\n", os.Geteuid(), os.Geteuid())
	decl := localCheckDecl{Version: 1, Checks: []localCheck{{Name: "sysv-smoke", Command: []string{"sh", "-c", `printf '%s' "$1" >> "$WF_PRIVATE_SYSV_TABLE"`, "fake-allocate", row}, TimeoutMinutes: 1, Weight: 1}}}
	raw, err := json.Marshal(decl)
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, d.Root, localChecksPath, string(raw))
	git(t, d.Root, "add", "-A")
	git(t, d.Root, "commit", "-qm", "private SysV check")
	git(t, d.Root, "push", "-q", "origin", "main")
	cli := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.Dir = d.Root
		cmd.WaitDelay = time.Second
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("private CLI %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	out := cli("local-check", "run", "--name", "sysv-smoke")
	if !strings.Contains(out, "id=777 creator=424242") || !strings.Contains(out, "not attributed") {
		t.Fatalf("%s", out)
	}
	// Repost must preserve original evidence even when current IPC is unreadable.
	if err := os.Remove(table); err != nil {
		t.Fatal(err)
	}
	cli("local-check", "republish", "--name", "sysv-smoke")
	statuses, err := os.ReadFile(statusFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(statuses), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("statuses %s", statuses)
	}
	for _, line := range lines {
		var status map[string]string
		if err := json.Unmarshal(line, &status); err != nil {
			t.Fatal(err)
		}
		if status["state"] != "success" || !strings.HasPrefix(status["description"], "SysV warning") {
			t.Fatalf("%s", line)
		}
	}
	t.Log("private CLI: stateful fake allocation, advisory success and digest-verified republish without IPC access")
}
