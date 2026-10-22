package golint

import "testing"

func TestUnitChildProcessGroup(t *testing.T) {
	const owner = "cmd/sdlc/flake_run.go"
	const group = "github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	for _, tc := range []struct {
		name, file, imports, body string
		bad                       bool
	}{
		{"configure", owner, `"` + group + `"`, `processgroup.Configure(cmd)`, true},
		{"terminate", owner, `"` + group + `"`, `processgroup.Terminate(cmd, true)`, true},
		{"import alias", owner, `pg "` + group + `"`, `pg.Configure(cmd)`, true},
		{"dot import", owner, `. "` + group + `"`, `Terminate(cmd, true)`, true},
		{"function alias", owner, `pg "` + group + `"`, `configure := pg.Configure; configure(cmd)`, true},
		{"sysprocattr assignment", owner, `"syscall"`, `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}`, true},
		{"sysprocattr literal", owner, `"os/exec"`, `_ = exec.Cmd{SysProcAttr: attrs}`, true},
		{"sysprocattr read", owner, `"os/exec"`, `_ = cmd.SysProcAttr`, true},
		{"setpgid", owner, `"syscall"`, `syscall.Setpgid(0, 0)`, true},
		{"setsid alias", owner, `sys "syscall"`, `sys.Setsid()`, true},
		{"kill group", owner, `sys "syscall"`, `sys.Kill(-pid, signal)`, true},
		{"kill current group", owner, `"syscall"`, `syscall.Kill(0, signal)`, true},
		{"raw positive pid", owner, `"syscall"`, `syscall.Kill(42, signal)`, true},
		{"raw dynamic pid", owner, `"syscall"`, `syscall.Kill(pid, signal)`, true},
		{"syscall function alias", owner, `sys "syscall"`, `kill := sys.Kill; kill(pid, signal)`, true},
		{"syscall dot import", owner, `. "syscall"`, `Setsid()`, true},
		{"managed batch", owner, `"os/exec"`, `runManagedCommandBatch(ctx, deps)`, true},
		{"slot batch", owner, `"os/exec"`, `runUnderSlotBatch(ctx, deps)`, true},
		{"single slot", owner, `"os/exec"`, `runUnderSlot(ctx, deps)`, true},
		{"single slot alias", owner, `"os/exec"`, `run := runUnderSlot; run(ctx, deps)`, true},
		{"resource batch", owner, `"os/exec"`, `runUnderResourceBatch(ctx, deps)`, true},
		{"single resource", owner, `"os/exec"`, `runUnderResource(ctx, deps)`, true},
		{"single resource alias", owner, `"os/exec"`, `run := runUnderResource; run(ctx, deps)`, true},
		{"cleanup", owner, `"os/exec"`, `configureRunnerCleanup(cmd)`, true},
		{"bounded cleanup", owner, `"os/exec"`, `configureRunnerCleanupWithin(cmd, timeout)`, true},
		{"bounded cleanup alias", owner, `"os/exec"`, `configure := configureRunnerCleanupWithin; configure(cmd, timeout)`, true},
		{"slot acquisition", owner, `"os/exec"`, `acquireSuiteSlot(ctx, deps)`, true},
		{"managed alias", owner, `"os/exec"`, `run := runManagedCommandBatch; run(ctx, deps)`, true},
		{"nested reference", owner, `"os/exec"`, `_ = runManagedCommandBatch(ctx, deps).Result`, true},
		{"sibling adapter", "cmd/sdlc/flake_run_child.go", `"` + group + `"`, `processgroup.Configure(cmd)`, true},
		{"absolute owner", "/repo/" + owner, `"` + group + `"`, `processgroup.Configure(cmd)`, true},
		{"inherited child", owner, `"os/exec"`, `cmd := exec.CommandContext(ctx, argv[0], argv[1:]...); cmd.Cancel = func() error { return cmd.Process.Kill() }; _ = cmd.Start(); _ = cmd.Wait()`, false},
		{"outer group owner", "cmd/sdlc/localcheck_run.go", `"` + group + `"`, `processgroup.Configure(cmd); runManagedCommandBatch(ctx, deps)`, false},
		{"test fixture", "cmd/sdlc/flake_run_test.go", `"` + group + `"`, `processgroup.Configure(cmd)`, false},
		{"sibling test fixture", "cmd/sdlc/flake_run_child_test.go", `"` + group + `"`, `processgroup.Configure(cmd)`, false},
		{"other directory", "cmd/other/flake_run.go", `"` + group + `"`, `processgroup.Configure(cmd)`, false},
		{"unrelated package", owner, `pg "example.org/processgroup"`, `pg.Configure(cmd)`, false},
		{"shadowed import", owner, `pg "` + group + `"`, `pg := unrelated; pg.Configure(cmd)`, false},
		{"comment string", owner, `"os/exec"`, "// processgroup.Configure(cmd)\n_ = `cmd.SysProcAttr`", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte("package main\nimport " + tc.imports + "\nfunc f() {" + tc.body + "}")
			got, err := Check(tc.file, src, Catalog)
			if err != nil {
				t.Fatal(err)
			}
			var hits int
			for _, finding := range got {
				if finding.Rule == "unit-child-process-group" {
					hits++
				}
			}
			if (hits > 0) != tc.bad {
				t.Fatalf("got %d group findings, bad=%v: %v", hits, tc.bad, got)
			}
		})
	}
}
