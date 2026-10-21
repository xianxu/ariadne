package golint

import (
	"strings"
	"testing"
)

func TestResumeCleanupEffect(t *testing.T) {
	for _, tc := range []struct {
		name, function, body string
		bad                  bool
	}{
		{"raw removal", "retireResumeEnvironmentInputs", `return ops.removeAll(p)`, true},
		{"raw sync", "releaseResumeEnvironmentOwner", `return ops.sync(p)`, true},
		{"scratch", "confirmResumeScratchCleanup", `return scratch.Cleanup()`, true},
		{"future cleanup", "cleanupResumeScratchInputs", `return os.Remove(p)`, true},
		{"future unlink", "unlinkResumeEnvironmentMarker", `return disk.RemoveAll(p)`, true},
		{"injected directory sync", "confirmResumeScratchCleanup", `return sync(p)`, true},
		{"deferred raw effect", "cleanupResumeEnvironment", `return resumeCleanupEffect(ctx,authorize,func()error{defer ops.remove(p);return nil})`, true},
		{"directory sync", "finishResumeScratch", `return syncResumeDirectory(p)`, true},
		{"git removal", "retireResumeEnvironmentInputs", `_,err:=resumeEnvironmentGit(d,"worktree","remove","--force",p);return err`, true},
		{"git unlock", "releaseResumeEnvironmentOwner", `_,err:=d.CleanupGit("worktree","unlock",p);return err`, true},
		{"git prune", "cleanupResumeEnvironment", `_,err:=d.Git("worktree","prune");return err`, true},
		{"effect alias", "cleanupResumeEnvironment", `remove:=ops.remove;return remove(p)`, true},
		{"git alias", "cleanupResumeEnvironment", `git:=d.Git;_,err:=git("worktree","remove",p);return err`, true},
		{"wrapped", "retireResumeEnvironmentInputs", `return resumeCleanupEffect(ctx,authorize,func()error{return ops.removeAll(p)})`, false},
		{"wrapped git", "retireResumeEnvironmentInputs", `return resumeCleanupEffect(ctx,authorize,func()error{_,err:=resumeEnvironmentGit(d,"worktree","remove",p);return err})`, false},
		{"two effects", "cleanupResumeEnvironment", `return resumeCleanupEffect(ctx,authorize,func()error{ops.remove(p);return ops.sync(p)})`, true},
		{"nested callback", "cleanupResumeEnvironment", `return resumeCleanupEffect(ctx,authorize,func()error{go func(){ops.remove(p)}();return nil})`, true},
		{"named callback", "cleanupResumeEnvironment", `effect:=func()error{return ops.remove(p)};return resumeCleanupEffect(ctx,authorize,effect)`, true},
		{"helper alias", "cleanupResumeEnvironment", `run:=resumeCleanupEffect;return run(ctx,authorize,func()error{return ops.remove(p)})`, true},
		{"shadowed helper", "cleanupResumeEnvironment", `resumeCleanupEffect:=other;return resumeCleanupEffect(ctx,authorize,func()error{return ops.remove(p)})`, true},
		{"unrelated", "readResumeEnvironment", `return ops.removeAll(p)`, false},
		{"setup", "protectResumeEnvironment", `_,err:=d.Git("worktree","lock",p);return err`, false},
		{"git read", "confirmResumeEnvironment", `_,err:=d.Git("rev-parse","HEAD");return err`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			imports := `import "os"`
			if strings.Contains(tc.body, "disk.") {
				imports = `import disk "os"`
			}
			src := []byte("package p\n" + imports + "\nfunc " + tc.function + "()error{" + tc.body + "}")
			for _, file := range []string{"cmd/sdlc/localcheck_resume_environment.go", "cmd/sdlc/localcheck_resume.go"} {
				findings, err := Check(file, src, Catalog)
				if err != nil {
					t.Fatal(err)
				}
				var hits int
				for _, f := range findings {
					if f.Rule == "resume-cleanup-effect" {
						hits++
					}
				}
				if (hits > 0) != tc.bad {
					t.Errorf("%s: got %d cleanup findings, bad=%v: %v", file, hits, tc.bad, findings)
				}
			}
		})
	}
}

func TestResumeCleanupEffectOwnerScope(t *testing.T) {
	src := []byte(`package p;func cleanupResumeEnvironment()error{return ops.remove(p)}`)
	for _, file := range []string{"cmd/sdlc/localcheck_resume_retention.go", "cmd/sdlc/localcheck_resume_environment_test.go", "other/localcheck_resume.go"} {
		findings, err := Check(file, src, Catalog)
		if err != nil || len(findings) != 0 {
			t.Errorf("%s: %v, %v", file, findings, err)
		}
	}
}

func TestResumeCleanupEffectStrictRunnerDefer(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		bad        bool
	}{
		{"raw strict cleanup", `defer func(){if d.StrictGroupCleanup{scratch.Cleanup();return};scratch.Cleanup()}()`, true},
		{"raw strict sync", `defer func(){if d.StrictGroupCleanup{syncResumeDirectory(p);return};scratch.Cleanup()}()`, true},
		{"raw strict alias", `cleanup:=scratch.Cleanup;defer func(){if d.StrictGroupCleanup{cleanup();return};scratch.Cleanup()}()`, true},
		{"resume scratch helper", `defer func(){if d.StrictGroupCleanup{err=finishResumeScratch(ctx,d,run,scratch,err);return};scratch.Cleanup()}()`, false},
		{"wrapped strict effect", `defer func(){if d.StrictGroupCleanup{resumeCleanupEffect(ctx,authorize,func()error{return scratch.Cleanup()});return};scratch.Cleanup()}()`, false},
		{"ordinary cleanup", `defer func(){scratch.Cleanup()}()`, false},
		{"strict setup sync", `if d.StrictGroupCleanup{syncResumeDirectory(p)}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			findings, err := Check("cmd/sdlc/localcheck_run.go", []byte("package p;func runUnderResourceBatch(){"+tc.body+"}"), Catalog)
			if err != nil {
				t.Fatal(err)
			}
			var hits int
			for _, finding := range findings {
				if finding.Rule == "resume-cleanup-effect" {
					hits++
				}
			}
			if (hits > 0) != tc.bad {
				t.Errorf("got %d cleanup findings, bad=%v: %v", hits, tc.bad, findings)
			}
		})
	}
}
