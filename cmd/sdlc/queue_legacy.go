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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

type queueLegacyProcess struct {
	PID                      int
	Command, Root, Workspace string
}

type queueLegacyOptions struct {
	FleetRoot string
	LockBase  string
	Slots     int
	// Private is an explicitly isolated portable cohort. Its default observer
	// examines only its private registry/resources, never the machine process
	// table. This proves nothing about legacy processes outside that cohort.
	Private      bool
	Processes    func(context.Context) ([]queueLegacyProcess, error)
	CommonDirs   func(context.Context, string) ([]string, error)
	ProcessStart func(int) (string, error)
	// Cost makes the single indexing pass and constant-time lookups observable
	// in operation-count tests; production leaves it nil.
	Cost func(string)
}

type queueLegacyProcessKey struct {
	pid                      int
	started, root, workspace string
}
type queueLegacyBindingKey struct{ id, kind string }
type queueLegacyIndex struct {
	bindings  map[queueLegacyBindingKey]queuegate.Operation
	pending   map[queueLegacyBindingKey]queuegate.Operation
	processes map[queueLegacyProcessKey]bool
}

func queueLegacyIndexKnown(known []queuegate.Operation, cost func(string)) queueLegacyIndex {
	idx := queueLegacyIndex{bindings: map[queueLegacyBindingKey]queuegate.Operation{}, pending: map[queueLegacyBindingKey]queuegate.Operation{}, processes: map[queueLegacyProcessKey]bool{}}
	for _, op := range known {
		if cost != nil {
			cost("index")
		}
		for _, kind := range op.Kinds {
			if op.Phase == "waiting" || op.Phase == "running" || op.Phase == "interrupting" {
				idx.pending[queueLegacyBindingKey{op.ID, kind}] = op
			}
			if op.Phase == "running" || op.Phase == "interrupting" {
				idx.bindings[queueLegacyBindingKey{op.ID, kind}] = op
			}
		}
		if queueLegacyHasKind(op, "merge") {
			for _, root := range []string{"", op.Root} {
				for _, workspace := range []string{"", op.Workspace} {
					idx.processes[queueLegacyProcessKey{op.PID, op.PIDStarted, root, workspace}] = true
				}
			}
		}
	}
	return idx
}

// observeLegacyQueue is status-only. Held flock resources establish legacy
// presence. Process observations only add uncertainty; they never grant signal,
// cleanup, or cancellation authority.
func observeLegacyQueue(ctx context.Context, o queueLegacyOptions, known []queuegate.Operation) ([]queuegate.Activity, []string) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	active := []queuegate.Activity{}
	details := []string{}
	seen := map[string]bool{}
	add := func(id, kind, workspace, detail string) {
		key := id + "/" + kind
		if !seen[key] {
			seen[key] = true
			active = append(active, queuegate.Activity{ID: id, Kind: kind, Workspace: workspace, State: "unknown"})
			details = append(details, detail)
		}
	}
	note := func(detail string) { details = append(details, detail) }
	if o.ProcessStart == nil {
		o.ProcessStart = processgroup.StartTime
	}
	if o.CommonDirs == nil {
		o.CommonDirs = queueLegacyCommonDirs
	}
	if len(known) > queuegate.MaxOperations {
		return active, []string{"known queue operation bound exceeded; legacy observation incomplete"}
	}
	index := queueLegacyIndexKnown(known, o.Cost)
	if o.Slots < 1 {
		o.Slots = 2
	}
	if o.Slots > queuegate.MaxOperations {
		note("legacy suite slot bound exceeded")
		return active, details
	}
	probe := func(path, resource, kind string) {
		if ctx.Err() != nil {
			return
		}
		raw, held, err := queueLegacyHeld(path)
		if err != nil {
			add(queueLegacyID(path, "unreadable"), kind, "", fmt.Sprintf("legacy %s inspection unavailable at %s: %v", kind, path, err))
			return
		}
		if !held {
			return
		}
		fields, err := queueLegacyFields(raw)
		if err != nil {
			add(queueLegacyID(path, "unreadable"), kind, "", fmt.Sprintf("legacy %s has unreadable held owner at %s: %v", kind, path, err))
			return
		}
		if o.Cost != nil {
			o.Cost("lookup")
		}
		if queueLegacyBound(fields, kind, index, o.ProcessStart) {
			return
		}
		canonical, _ := json.Marshal(fields)
		id := queueLegacyID(resource, string(canonical))
		add(id, kind, "", fmt.Sprintf("legacy %s held at %s (owner %s) lacks a verified queue lease", kind, path, string(canonical)))
	}
	if o.LockBase == "" {
		note("legacy machine suite lock path unavailable")
	} else {
		for i := 0; i < o.Slots; i++ {
			path := o.LockBase
			if i > 0 {
				path += "." + strconv.Itoa(i)
			}
			probe(path, o.LockBase, "suite")
		}
		queueLegacyPending(ctx, o, index, note)
	}
	dirs, err := o.CommonDirs(ctx, o.FleetRoot)
	if err != nil {
		note("legacy fleet common-directory inspection unavailable: " + err.Error())
	}
	if len(dirs) > queuegate.MaxOperations {
		note("legacy fleet common-directory bound exceeded")
	} else {
		checked := map[string]bool{}
		for _, dir := range dirs {
			dir = filepath.Clean(dir)
			if checked[dir] {
				continue
			}
			checked[dir] = true
			probe(filepath.Join(dir, "sdlc", "landing", "head.lock"), dir, "landing")
		}
	}
	processes := o.Processes
	if processes == nil && !o.Private {
		processes = queueLegacyProcesses
	}
	if processes != nil {
		rows, err := processes(ctx)
		if err != nil {
			note("legacy merge process inspection unavailable: " + err.Error())
		}
		if len(rows) > queuegate.MaxOperations {
			note("legacy merge process bound exceeded")
		} else {
			for _, p := range rows {
				if ctx.Err() != nil {
					break
				}
				if o.Private && !queueLegacyWithin(o.FleetRoot, p.Root) {
					continue
				}
				if !queueLegacyMergeCommand(p.Command) {
					continue
				}
				started, err := o.ProcessStart(p.PID)
				if errors.Is(err, processgroup.ErrNoProcess) {
					continue
				}
				if err != nil || started == "" {
					add(queueLegacyID("process", fmt.Sprintf("%d/unknown", p.PID)), "merge", p.Workspace, fmt.Sprintf("legacy merge pid %d kernel incarnation unavailable", p.PID))
					continue
				}
				if o.Cost != nil {
					o.Cost("lookup")
				}
				if index.processes[queueLegacyProcessKey{p.PID, started, p.Root, p.Workspace}] {
					continue
				}
				add(queueLegacyID("process", fmt.Sprintf("%d/%s", p.PID, started)), "merge", p.Workspace, fmt.Sprintf("legacy merge pid %d incarnation %s has no queue lifetime lease", p.PID, started))
			}
		}
	}
	if ctx.Err() != nil {
		note("legacy queue inspection deadline or context unavailable: " + ctx.Err().Error())
	}
	if len(active) > queuegate.MaxOperations {
		return []queuegate.Activity{}, []string{"legacy active entry bound exceeded; observation incomplete"}
	}
	return active, details
}

func queueLegacyID(resource, identity string) string {
	sum := sha256.Sum256([]byte(resource + "\x00" + identity))
	return "legacy-" + hex.EncodeToString(sum[:16])
}

// A held old waiter may admit later without consulting pause. It is not active,
// but prevents a complete snapshot until it drains or proves a new binding.
// This bounded, read-only probe is called exclusively by status observation.
func queueLegacyPending(ctx context.Context, o queueLegacyOptions, index queueLegacyIndex, note func(string)) {
	probe := func(path string) {
		if ctx.Err() != nil {
			return
		}
		raw, held, err := queueLegacyHeld(path)
		if err != nil {
			note(fmt.Sprintf("legacy pending admission inspection unavailable at %s: %v", path, err))
			return
		}
		if !held {
			return
		}
		fields, err := queueLegacyFields(raw)
		if err == nil {
			if o.Cost != nil {
				o.Cost("lookup")
			}
			if queueLegacyBindingMatches(fields, "suite", index.pending, o.ProcessStart) {
				return
			}
		}
		note(fmt.Sprintf("legacy pending admission at %s lacks a verified queue lease; it may admit without consulting pause", path))
	}
	probe(o.LockBase + ".queue")
	dir := o.LockBase + ".tickets"
	f, err := os.Open(dir)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		note(fmt.Sprintf("legacy pending admission directory unavailable at %s: %v", dir, err))
		return
	}
	defer f.Close()
	entries, err := f.ReadDir(queuegate.MaxOperations + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		note(fmt.Sprintf("legacy pending admission directory unreadable at %s: %v", dir, err))
		return
	}
	if len(entries) > queuegate.MaxOperations {
		note("legacy pending admission entry bound exceeded at " + dir)
		return
	}
	for _, entry := range entries {
		if ctx.Err() != nil {
			return
		}
		// alloc is the allocation mutex, not a waiting lifetime resource.
		if entry.Name() == "alloc" {
			continue
		}
		if o.Cost != nil {
			o.Cost("pending")
		}
		probe(filepath.Join(dir, entry.Name()))
	}
}
func queueLegacyHasKind(op queuegate.Operation, kind string) bool {
	for _, k := range op.Kinds {
		if k == kind {
			return true
		}
	}
	return false
}
func queueLegacyWithin(root, path string) bool {
	if root == "" || path == "" {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// queueLegacyHeld never creates or rewrites resources. An unlocked old file is
// historical residue; an unreadable held file remains present and uncertain.
func queueLegacyHeld(path string) ([]byte, bool, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	err = flock.Exclusive(f, true)
	if err == nil {
		return nil, false, nil
	}
	if !errors.Is(err, syscall.EWOULDBLOCK) {
		return nil, false, err
	}
	b, err := io.ReadAll(io.LimitReader(f, queuegate.MaxJSON+1))
	if err != nil {
		return nil, true, err
	}
	if len(b) > queuegate.MaxJSON {
		return nil, true, errors.New("held record exceeds 1 MiB")
	}
	return b, true, nil
}

func queueLegacyBound(fields map[string]json.RawMessage, kind string, index queueLegacyIndex, start func(int) (string, error)) bool {
	return queueLegacyBindingMatches(fields, kind, index.bindings, start)
}

func queueLegacyBindingMatches(fields map[string]json.RawMessage, kind string, bindings map[queueLegacyBindingKey]queuegate.Operation, start func(int) (string, error)) bool {
	var id, root, inc string
	var pid int
	if json.Unmarshal(fields["queue_operation_id"], &id) != nil || json.Unmarshal(fields["queue_root"], &root) != nil || json.Unmarshal(fields["queue_pid"], &pid) != nil || json.Unmarshal(fields["queue_pid_started"], &inc) != nil || id == "" || root == "" || pid <= 1 || inc == "" {
		return false
	}
	op, ok := bindings[queueLegacyBindingKey{id, kind}]
	if ok && op.Root == root && op.PID == pid && op.PIDStarted == inc {
		actual, err := start(pid)
		return err == nil && actual == inc
	}
	return false
}

// Legacy records share the gate codec's structural validation; their old
// fields remain opaque except for the explicit queue incarnation binding.
func queueLegacyFields(b []byte) (map[string]json.RawMessage, error) {
	return queuegate.StrictObject(b)
}

func queueLegacyCommonDirs(ctx context.Context, root string) ([]string, error) {
	if root == "" {
		return nil, errors.New("fleet root unavailable")
	}
	repos, err := project.FleetRepoDirs(root)
	if err != nil {
		return nil, err
	}
	if len(repos) > queuegate.MaxOperations {
		return nil, errors.New("fleet repository bound exceeded")
	}
	dirs := []string{}
	problems := []error{}
	for _, repo := range repos {
		if ctx.Err() != nil {
			return dirs, ctx.Err()
		}
		_, err := os.Stat(filepath.Join(repo, ".git"))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", repo, err))
			continue
		}
		cmd := exec.CommandContext(ctx, "git", "-C", repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
		out, err := cmd.Output()
		if err != nil {
			problems = append(problems, fmt.Errorf("Git common directory %s: %w", repo, err))
			continue
		}
		if len(out) > queuegate.MaxText {
			problems = append(problems, errors.New("Git common directory output exceeds bound"))
			continue
		}
		dir := strings.TrimSuffix(string(out), "\n")
		if !filepath.IsAbs(dir) {
			problems = append(problems, errors.New("nonabsolute Git common directory"))
			continue
		}
		dirs = append(dirs, filepath.Clean(dir))
	}
	return dirs, errors.Join(problems...)
}

type queueLegacyOutput struct {
	mu       sync.Mutex
	b        bytes.Buffer
	overflow bool
}

func (b *queueLegacyOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	remaining := queuegate.MaxJSON - b.b.Len()
	if len(p) > remaining {
		b.overflow = true
		b.b.Write(p[:remaining])
	} else {
		b.b.Write(p)
	}
	return len(p), nil
}

func queueLegacyProcesses(ctx context.Context) ([]queueLegacyProcess, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return nil, errors.New("legacy process inspection unsupported on this platform")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var out, stderr queueLegacyOutput
	cmd := exec.CommandContext(ctx, "ps", "-axww", "-o", "pid=,uid=,command=")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.WaitDelay = 100 * time.Millisecond
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("process inventory: %w", err)
	}
	if out.overflow || stderr.overflow || out.b.Len()+stderr.b.Len() > queuegate.MaxJSON {
		return nil, errors.New("process inventory exceeds 1 MiB")
	}
	return queueLegacyParseProcesses(out.b.Bytes(), os.Getuid())
}

func queueLegacyParseProcesses(data []byte, uid int) ([]queueLegacyProcess, error) {
	rows := []queueLegacyProcess{}
	if len(data) > queuegate.MaxJSON {
		return nil, errors.New("process inventory exceeds 1 MiB")
	}
	if !utf8.Valid(data) {
		return nil, errors.New("process inventory contains invalid UTF-8")
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		return nil, errors.New("truncated process inventory")
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, errors.New("malformed process inventory row")
		}
		pid, pidErr := strconv.Atoi(fields[0])
		owner, uidErr := strconv.Atoi(fields[1])
		if pidErr != nil || uidErr != nil || pid <= 0 {
			return nil, errors.New("malformed process inventory identity")
		}
		if owner != uid {
			continue
		}
		command := strings.Join(fields[2:], " ")
		if queueLegacyMergeCommand(command) {
			rows = append(rows, queueLegacyProcess{PID: pid, Command: command})
		}
	}
	return rows, nil
}

func queueLegacyMergeCommand(command string) bool {
	fields := strings.Fields(command)
	if len(fields) < 2 {
		return false
	}
	base := filepath.Base(fields[0])
	if base != "sdlc" && !strings.HasPrefix(base, "sdlc-") {
		return false
	}
	// Root flags can precede the subcommand. Conservative matching may hold a
	// drain for inspection, but never grants authority over the observed PID.
	for _, arg := range fields[1:] {
		if arg == "merge" {
			return true
		}
	}
	return false
}
