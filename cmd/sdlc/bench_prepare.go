package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/markdoc"
)

// Only gate executable selection is injectable. Gates themselves always execute
// as real child processes in the sanitized checkout, never in the source tree.
type benchPrepareDeps struct{ GateArgv, GateEnv []string }

func prepareBenchTemplate(ctx context.Context, t bench.Task, templatePath string) error {
	if err := bench.ValidateTask(t); err != nil {
		return err
	}
	for _, pin := range []string{t.ParentOID, t.LandingOID} {
		kind, err := runBenchPrepareCommand(ctx, t.SourceRepo, nil, "git", "--no-replace-objects", "-c", "protocol.allow=never", "-c", "protocol.file.allow=never", "cat-file", "-t", pin)
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(kind)) != "commit" {
			return errors.New("parent and landing pins must name local commits")
		}
	}
	original, err := gitx.ReadLocalBlob(ctx, t.SourceRepo, t.OriginalIssueOID)
	if err != nil {
		return err
	}
	if err := verifyBenchOriginalIssue(t, original); err != nil {
		return err
	}
	patch, err := readBenchPatch(t.Acceptance.Patch)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(patch)
	if hex.EncodeToString(sum[:]) != t.Acceptance.PatchSHA256 {
		return errors.New("acceptance patch digest differs from descriptor")
	}
	if err := verifyBenchPatchPaths(ctx, filepath.Dir(templatePath), patch, t.Acceptance.Paths); err != nil {
		return err
	}
	files, err := gitx.ReadLocalTree(ctx, t.SourceRepo, t.ParentOID, func(f gitx.TreeFile) bool { return !benchProcessArtifact(f.Path) })
	if err != nil {
		return err
	}
	return writeBenchTemplate(ctx, templatePath, files)
}

func instantiateBenchSession(ctx context.Context, s *benchSession, templatePath string, d benchPrepareDeps) (retErr error) {
	release, err := s.acquireRun()
	if err != nil {
		return err
	}
	defer release()
	r, state, err := s.read()
	if err != nil {
		return err
	}
	if state.Phase() != bench.PhaseNew {
		return fmt.Errorf("prepare requires new run; retained phase %s", state.Phase())
	}
	if err := bench.ValidateTask(r.Task); err != nil {
		return err
	}
	if len(d.GateArgv) == 0 {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		d.GateArgv = []string{exe}
	}
	op := "prepare-" + r.ID
	_, effects, err := s.apply(bench.Event{ID: op + "-begin", RunID: r.ID, Kind: bench.BeginPrepare, OperationID: op})
	if err != nil {
		return err
	}
	if len(effects) != 1 || effects[0].Kind != bench.Prepare {
		return errors.New("preparation effect not authorized")
	}
	defer func() {
		if retErr != nil {
			var unconfirmed *runnerCleanupUnconfirmed
			if errors.As(retErr, &unconfirmed) {
				return
			}
			_, _, stateErr := s.apply(bench.Event{ID: op + "-failed", RunID: r.ID, Kind: bench.PrepareFailed, OperationID: op, Reason: retErr.Error()})
			retErr = errors.Join(retErr, stateErr)
		}
	}()
	t := r.Task
	files, err := readBenchTemplate(ctx, templatePath)
	if err != nil {
		return err
	}
	public := filepath.Join(s.root, "public")
	if err := os.Mkdir(public, 0700); err != nil {
		return fmt.Errorf("refuse existing preparation directory: %w", err)
	}
	// Preparation owns only this newly created directory. On failure preserve
	// the raw observer record but remove its incomplete disposable material.
	defer func() {
		if retErr != nil {
			var unconfirmed *runnerCleanupUnconfirmed
			if errors.As(retErr, &unconfirmed) {
				return
			}
			if err := s.check(); err != nil {
				retErr = errors.Join(retErr, err)
				return
			}
			retErr = errors.Join(retErr, os.RemoveAll(public))
		}
	}()
	work, hub, origin := filepath.Join(public, "work"), filepath.Join(public, "hub"), filepath.Join(public, "origin.git")
	if err := os.Mkdir(work, 0700); err != nil {
		return err
	}
	if err := os.Mkdir(hub, 0700); err != nil {
		return err
	}
	if err := materializeBenchTree(work, files); err != nil {
		return err
	}
	contextPublic := benchPublicContext{Version: 1, RunID: r.ID, WorkRoot: work, HubRoot: hub, OriginPath: origin, OriginIdentity: "file:" + origin}
	for _, root := range []string{work, hub} {
		if err := writeBenchJSONExclusive(filepath.Join(root, benchPublicContextPath), contextPublic); err != nil {
			return err
		}
	}
	if err := writeBenchJSONExclusive(filepath.Join(public, "task.json"), bench.ProjectPublic(t)); err != nil {
		return err
	}
	issuePath := filepath.Join(work, "workshop", "issues", t.ID+"-bench-replay.md")
	created, err := time.Parse(time.RFC3339Nano, r.CreatedAt)
	if err != nil {
		return err
	}
	date := created.UTC().Format("2006-01-02")
	issueText := fmt.Sprintf("---\nid: %s\nstatus: open\ndeps: []\ncreated: %s\nupdated: %s\n---\n\n# Bench replay %s\n\n## Problem\n\n## Spec\n\n%s\n\n## Done when\n\n%s\n\n## Plan\n\n## Log\n", t.ID, date, date, t.ID, strings.TrimSpace(t.Spec), strings.TrimSpace(t.DoneWhen))
	if err := writeBenchExclusive(issuePath, []byte(issueText), 0644); err != nil {
		return err
	}
	if err := writeBenchExclusive(filepath.Join(hub, ".hub"), []byte("# Provision an exact worker slot claim before dispatch.\n"), 0644); err != nil {
		return err
	}
	if err := writeBenchExclusive(filepath.Join(hub, "team.md"), []byte("# Bench team\n\nOne disposable replay hub. Follow normal tl-lead and tl-lead bench rules.\n"), 0644); err != nil {
		return err
	}
	if _, err := runBenchPrepareCommand(ctx, public, nil, "git", "init", "--bare", "-q", "-b", "main", origin); err != nil {
		return err
	}
	for _, root := range []string{work, hub} {
		for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"config", "user.name", "Model Bench"}, {"config", "user.email", "model-bench@localhost.invalid"}, {"config", "core.hooksPath", os.DevNull}, {"add", "-f", "-A"}, {"commit", "-qm", agentid.StampCurrent("bench: sanitized replay seed", ctx)}} {
			if _, err := runBenchPrepareCommand(ctx, root, nil, "git", args...); err != nil {
				return err
			}
		}
	}
	seed, err := runBenchPrepareCommand(ctx, work, nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	committed, err := gitx.ReadLocalTree(ctx, work, strings.TrimSpace(string(seed)), nil)
	if err != nil {
		return err
	}
	byPath := make(map[string]gitx.TreeFile, len(committed))
	for _, f := range committed {
		byPath[f.Path] = f
	}
	for _, expected := range files {
		actual, present := byPath[expected.Path]
		if !present || actual.Mode != expected.Mode || !bytes.Equal(actual.Content, expected.Content) {
			return fmt.Errorf("Git attributes, ignore rules or filesystem changed parent entry %s; replay requires exact product bytes/modes", expected.Path)
		}
	}
	for _, args := range [][]string{{"remote", "add", "origin", origin}, {"push", "-q", "-u", "origin", "main"}} {
		if _, err := runBenchPrepareCommand(ctx, work, nil, "git", args...); err != nil {
			return err
		}
	}
	// Real migration retains the original ID, real tracker CAS, and the normal
	// card mirror. This seed contains no historic tracker/completion authority.
	dry, err := runBenchGate(ctx, work, d, "issue", "migrate")
	if err != nil {
		return err
	}
	digest := regexp.MustCompile(`(?m)digest: ([0-9a-f]{16})`).FindSubmatch(dry)
	if len(digest) != 2 {
		return errors.New("migration dry run did not produce one reviewed digest")
	}
	if _, err := runBenchGate(ctx, work, d, "issue", "migrate", "--apply", "--expect", string(digest[1])); err != nil {
		return err
	}
	if _, err := runBenchPrepareCommand(ctx, work, nil, "git", "pull", "--ff-only", "-q"); err != nil {
		return err
	}
	if _, present, err := tracker.ReadCutoverMarker(work); err != nil || !present {
		return fmt.Errorf("real tracker bootstrap not confirmed (marker present=%v): %v", present, err)
	}
	if _, _, err := s.apply(bench.Event{ID: op + "-prepared", RunID: r.ID, Kind: bench.Prepared, OperationID: op}); err != nil {
		return err
	}
	return nil
}

func verifyBenchOriginalIssue(t bench.Task, raw []byte) error {
	fm, body, err := issue.Parse(string(raw))
	if err != nil {
		return err
	}
	id, _ := issue.GetField(fm, "id")
	if id != t.ID {
		return errors.New("original issue blob has a different issue ID")
	}
	for heading, want := range map[string]string{"Spec": t.Spec, "Done when": t.DoneWhen} {
		got, present := markdoc.SectionBody(body, heading)
		if !present || strings.TrimSpace(got) != strings.TrimSpace(want) {
			return fmt.Errorf("original issue %s differs from descriptor", heading)
		}
	}
	return nil
}

func benchProcessArtifact(p string) bool {
	// A new tracker cannot import mirrors/completions whose original object
	// graph was intentionally removed. Start with one fresh replay inventory.
	for _, prefix := range []string{"workshop/issues", "workshop/plans", "workshop/history"} {
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			return true
		}
	}
	return p == tracker.CutoverMarkerPath
}

func validateBenchTree(files []gitx.TreeFile) error {
	seen := map[string]bool{}
	spellings := map[string]string{}
	links := map[string]string{}
	total := 0
	for _, f := range files {
		p := f.Path
		lower := strings.ToLower(p)
		if p == "" || path.IsAbs(p) || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../") || strings.ContainsAny(p, "\\\x00\r\n") || seen[lower] {
			return fmt.Errorf("unsafe or colliding parent path %q", p)
		}
		for _, part := range strings.Split(lower, "/") {
			if part == ".git" {
				return fmt.Errorf("Git control path in parent: %s", p)
			}
		}
		if lower == strings.ToLower(benchPublicContextPath) || strings.HasPrefix(lower, strings.ToLower(benchPublicContextPath)+"/") {
			return errors.New("parent collides with reserved bench context")
		}
		if lower == ".sdlc" || lower == "workshop" {
			return fmt.Errorf("parent file collides with preparation directory %s", p)
		}
		parts := strings.Split(p, "/")
		for index := range parts {
			prefix := strings.Join(parts[:index+1], "/")
			folded := strings.ToLower(prefix)
			if previous, exists := spellings[folded]; exists && previous != prefix {
				return fmt.Errorf("parent directory casing collision: %s and %s", previous, prefix)
			}
			spellings[folded] = prefix
		}
		seen[lower] = true
		total += len(f.Content)
		if total > 64<<20 {
			return errors.New("parent exceeds 64 MiB tree budget")
		}
		if f.Mode != "100644" && f.Mode != "100755" && f.Mode != "120000" {
			return fmt.Errorf("unsupported parent mode %s at %s", f.Mode, p)
		}
		if f.Mode == "120000" {
			target := string(f.Content)
			resolved := path.Clean(path.Join(path.Dir(p), target))
			if target == "" || path.IsAbs(target) || strings.ContainsAny(target, "\\\x00\r\n") || resolved == ".." || strings.HasPrefix(resolved, "../") {
				return fmt.Errorf("escaping parent symlink %s", p)
			}
			links[lower] = target
		}
	}
	// Symlink ancestry is forbidden even when its immediate target is inside:
	// writing through it would turn an export path into a different tree path.
	for _, f := range files {
		for dir := path.Dir(f.Path); dir != "."; dir = path.Dir(dir) {
			if seen[strings.ToLower(dir)] {
				return fmt.Errorf("parent file/directory ancestry collision at %s", f.Path)
			}
		}
	}
	// Interpret links against the exported tree, allowing a contained dangling
	// target while refusing cycles and chains whose '..' escapes after an
	// intermediate symlink changes its effective depth.
	for _, f := range files {
		if f.Mode != "120000" {
			continue
		}
		queue := strings.Split(f.Path, "/")
		var stack []string
		visits := 0
		for len(queue) > 0 {
			part := queue[0]
			queue = queue[1:]
			if part == "" || part == "." {
				continue
			}
			if part == ".." {
				if len(stack) == 0 {
					return fmt.Errorf("parent symlink chain escapes checkout: %s", f.Path)
				}
				stack = stack[:len(stack)-1]
				continue
			}
			candidate := strings.Join(append(append([]string(nil), stack...), part), "/")
			if target, isLink := links[strings.ToLower(candidate)]; isLink {
				visits++
				if visits > 40 {
					return fmt.Errorf("parent symlink cycle or excessive chain: %s", f.Path)
				}
				queue = append(strings.Split(target, "/"), queue...)
				continue
			}
			stack = append(stack, part)
		}
	}
	return nil
}

func materializeBenchTree(root string, files []gitx.TreeFile) error {
	for _, f := range files {
		if f.Mode == "120000" {
			continue
		}
		mode := os.FileMode(0644)
		if f.Mode == "100755" {
			mode = 0755
		}
		if err := writeBenchExclusive(filepath.Join(root, filepath.FromSlash(f.Path)), f.Content, mode); err != nil {
			return err
		}
	}
	// Install symlinks last so no file write can follow one.
	for _, f := range files {
		if f.Mode != "120000" {
			continue
		}
		p := filepath.Join(root, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		if err := os.Symlink(string(f.Content), p); err != nil {
			return err
		}
	}
	return nil
}

func readBenchPatch(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("acceptance patch must be a regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, bench.TaskRecordLimit+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > bench.TaskRecordLimit || len(raw) == 0 {
		return nil, errors.New("acceptance patch is empty or exceeds 1 MiB")
	}
	return raw, nil
}

func verifyBenchPatchPaths(ctx context.Context, dir string, patch []byte, allow []string) error {
	out, err := runBenchPrepareCommand(ctx, dir, patch, "git", "apply", "--numstat", "-z")
	if err != nil {
		return fmt.Errorf("parse acceptance patch: %w", err)
	}
	want := map[string]bool{}
	for _, p := range allow {
		want[p] = true
	}
	seen := map[string]bool{}
	for _, row := range bytes.Split(out, []byte{0}) {
		if len(row) == 0 {
			continue
		}
		fields := bytes.SplitN(row, []byte{'\t'}, 3)
		if len(fields) != 3 || string(fields[0]) == "-" || string(fields[1]) == "-" || !want[string(fields[2])] || seen[string(fields[2])] {
			return errors.New("acceptance patch touches unsupported, duplicate or undeclared test paths")
		}
		seen[string(fields[2])] = true
	}
	if len(seen) != len(want) {
		return errors.New("acceptance patch paths differ from descriptor")
	}
	return nil
}

func writeBenchExclusive(p string, raw []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(raw)
	return errors.Join(writeErr, f.Close())
}
func writeBenchJSONExclusive(p string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(raw)+1 > benchRecordLimit {
		return errors.New("bench JSON evidence exceeds 1 MiB before publication")
	}
	return writeBenchExclusive(p, append(raw, '\n'), 0644)
}

func benchPrepareEnv() []string {
	var env []string
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		if strings.HasPrefix(key, "GIT_") || key == "COUCH_TREE" || key == "COUCH_SLOT" || key == "COUCH_HUB" {
			continue
		}
		env = append(env, value)
	}
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_NO_LAZY_FETCH=1", "GIT_NO_REPLACE_OBJECTS=1", "GIT_TERMINAL_PROMPT=0")
}

type benchBoundedOutput struct{ bytes.Buffer }

func (b *benchBoundedOutput) Write(p []byte) (int, error) {
	if len(p) > bench.TaskRecordLimit-b.Len() {
		return 0, errors.New("bench command output exceeds 1 MiB")
	}
	return b.Buffer.Write(p)
}
func runBenchPrepareCommand(ctx context.Context, dir string, input []byte, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = benchPrepareEnv()
	cmd.Stdin = bytes.NewReader(input)
	processgroup.Configure(cmd)
	join := configureRunnerCleanup(cmd)
	var output benchBoundedOutput
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := errors.Join(cmd.Run(), join()); err != nil {
		return nil, fmt.Errorf("bench preparation %s %v: %w: %s", name, args, err, output.String())
	}
	return output.Bytes(), nil
}
func runBenchGate(ctx context.Context, root string, d benchPrepareDeps, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	argv := append(append([]string(nil), d.GateArgv[1:]...), args...)
	cmd := exec.CommandContext(ctx, d.GateArgv[0], argv...)
	cmd.Dir = root
	cmd.Env = append(benchPrepareEnv(), d.GateEnv...)
	processgroup.Configure(cmd)
	join := configureRunnerCleanup(cmd)
	var output benchBoundedOutput
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := errors.Join(cmd.Run(), join()); err != nil {
		return nil, fmt.Errorf("real bench gate %v: %w: %s", args, err, output.String())
	}
	return output.Bytes(), nil
}
