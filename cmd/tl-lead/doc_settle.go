package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

const maxWorkingdocBytes = 1 << 20

type settlementResult struct {
	Target, Commit     string
	Published, Stamped bool
	Warnings           []string
}

var workingdocPrefix = regexp.MustCompile(`^(?:\d{4}-\d{2}-\d{2}-)?\d+-`)

func workingdocTarget(m workingdocMetadata, source, requested string, now time.Time) (string, error) {
	if m.Pointer != nil {
		if requested != "" && requested != m.Pointer.Path {
			return "", fmt.Errorf("already settled into %s; relocation requires reconciliation", m.Pointer.Path)
		}
		return m.Pointer.Path, nil
	}
	if requested == "" {
		slug := strings.TrimSuffix(path.Base(source), ".md")
		slug = workingdocPrefix.ReplaceAllString(slug, "")
		if slug == "" {
			return "", fmt.Errorf("source has no topic slug")
		}
		requested = "workshop/workingdoc/" + now.Format("2006-01-02") + "-" + slug + ".md"
	}
	if !workingdocPath(requested) {
		return "", fmt.Errorf("invalid workingdoc destination %q", requested)
	}
	return requested, nil
}

func readWorkingdoc(root, rel string) ([]byte, bool, error) {
	full, err := safeMovePath(root, rel)
	if err != nil {
		return nil, false, err
	}
	info, err := os.Lstat(full)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() {
		return nil, false, fmt.Errorf("not a regular file: %s", rel)
	}
	if info.Size() > maxWorkingdocBytes {
		return nil, false, fmt.Errorf("%s exceeds the 1 MiB workingdoc limit", rel)
	}
	f, err := os.Open(full)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxWorkingdocBytes+1))
	if err != nil {
		return nil, false, err
	}
	if len(b) > maxWorkingdocBytes {
		return nil, false, fmt.Errorf("%s exceeds the 1 MiB workingdoc limit", rel)
	}
	return b, true, nil
}

func workingdocEligible(target string, mode docPublicationMode) bool {
	if !workingdocPath(target) {
		return false
	}
	for _, prefix := range []string{"workshop/workingdoc/", "workshop/projects/", "workshop/targets/", "workshop/parley/", "workshop/pensive/"} {
		if strings.HasPrefix(target, prefix) {
			return true
		}
	}
	if target == "workshop/lessons.md" {
		return true
	}
	if mode == docOperating {
		parts := strings.Split(target, "/")
		return len(parts) >= 5 && parts[0] == "teams" && parts[1] != "" && parts[2] == "projects"
	}
	return false
}

func lockWorkingdoc(common string) (func(), error) {
	f, err := os.OpenFile(filepath.Join(common, "tl-lead-doc-settle.flock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another settlement holds the repository lock: %w", err)
	}
	// The stable lock inode is retained; deleting it would let a waiter and a
	// new caller hold locks on different inodes. It has constant disk cost.
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}

func checkWorkingdocLock(ss SessionSync, file string, registered bool) error {
	if _, err := os.Lstat(file + ".lock"); err == nil {
		return fmt.Errorf("%s: %w", file, ErrLocked)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if registered {
		st, err := ss.Status(file)
		if err != nil {
			return err
		}
		if st.Holder == "operator" {
			return fmt.Errorf("%s: %w", file, ErrLocked)
		}
	}
	return nil
}

func applyWorkingdoc(ss SessionSync, file, before, after string) error {
	if before == after {
		return nil
	}
	if before == "" {
		return fmt.Errorf("registered empty destination %s needs an initial document before settlement", file)
	}
	recs := []editRecord{{Old: before, New: after, Occurrence: 1, Explain: "Record the confirmed settlement copy"}}
	if got, err := applyRecords(before, recs); err != nil || got != after {
		return fmt.Errorf("invalid settlement edit: %v", err)
	}
	f, err := os.CreateTemp(os.TempDir(), "tl-lead-doc-records-*.json")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = json.NewEncoder(f).Encode(recs); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return ss.Apply(file, name, fileHash(before))
}

func prepareWorkingdoc(ss SessionSync, root, target string, before []byte, present, registered bool, payload []byte) error {
	full, err := safeMovePath(root, target)
	if err != nil {
		return err
	}
	current, exists, err := readWorkingdoc(root, target)
	if err != nil {
		return err
	}
	if present != exists || !bytes.Equal(current, before) {
		return fmt.Errorf("destination changed before preparation: %s", target)
	}
	if bytes.Equal(current, payload) && exists {
		return nil
	}
	if registered {
		return applyWorkingdoc(ss, full, string(before), string(payload))
	}
	if err = os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	if _, err = safeMovePath(root, target); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(full), ".workingdoc-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(payload); err != nil {
		f.Close()
		return err
	}
	if err = f.Chmod(0644); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if !present {
		// Link creates an absent destination without replacing a racing create.
		return os.Link(name, full)
	}
	current, exists, err = readWorkingdoc(root, target)
	if err != nil {
		return err
	}
	if !exists || !bytes.Equal(current, before) {
		return fmt.Errorf("destination changed during preparation: %s", target)
	}
	return os.Rename(name, full)
}

func settleWorkingdoc(ctx context.Context, d deps, repo DocRepository, root, source, requested string) (settlementResult, error) {
	var result settlementResult
	if err := ctx.Err(); err != nil {
		return result, err
	}
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return result, err
	}
	if filepath.IsAbs(source) {
		source, err = filepath.Rel(root, source)
		if err != nil {
			return result, err
		}
	}
	source = filepath.ToSlash(source)
	if !strings.HasSuffix(source, ".md") || !sidedoc.InSideDir(source) {
		return result, fmt.Errorf("source must be a Markdown side document inside this repository")
	}
	common, err := repo.CommonDir(ctx, root)
	if err != nil {
		return result, err
	}
	unlock, err := lockWorkingdoc(common)
	if err != nil {
		return result, err
	}
	defer unlock()
	mode, err := repo.Mode(ctx, root)
	if err != nil {
		return result, err
	}
	if mode != docProduct && mode != docOperating {
		return result, fmt.Errorf("unknown document publication mode")
	}
	before, present, err := readWorkingdoc(root, source)
	if err != nil {
		return result, err
	}
	if !present {
		return result, fmt.Errorf("source missing: %s", source)
	}
	m, err := parseWorkingdoc(string(before))
	if err != nil {
		return result, fmt.Errorf("%s: %w", source, err)
	}
	target, err := workingdocTarget(m, source, requested, d.clock())
	if err != nil {
		return result, err
	}
	result.Target = target
	if !workingdocEligible(target, mode) {
		return result, fmt.Errorf("destination %s is not eligible prose for this repository", target)
	}
	if err = repo.ValidateTarget(ctx, root, target); err != nil {
		return result, err
	}
	ss := boundedMoveSync(d.ss)
	regs, err := ss.RegistrationSnapshot()
	if err != nil {
		return result, err
	}
	sourceFull := filepath.Join(root, filepath.FromSlash(source))
	targetFull, err := safeMovePath(root, target)
	if err != nil {
		return result, err
	}
	if regs[sourceFull] == nil {
		return result, fmt.Errorf("source unregistered; run couch --session-sync register %s --owner SLOT, then rerun", source)
	}
	for _, f := range []string{sourceFull, targetFull} {
		if err = checkWorkingdocLock(ss, f, regs[f] != nil); err != nil {
			return result, err
		}
	}
	liveReferences := map[string]bool{}
	exists := func(rel string) bool {
		if rel == "session.md" || sidedoc.InSideDir(rel) {
			liveReferences[rel] = true
		}
		p, err := safeMovePath(root, rel)
		if err != nil {
			return false
		}
		_, err = os.Lstat(p)
		return err == nil
	}
	projected, warnings, err := projectWorkingdoc(string(before), source, target, exists)
	if err != nil {
		return result, fmt.Errorf("%s: %w", source, err)
	}
	payload := []byte(projected)
	result.Warnings = warnings
	for rel := range liveReferences {
		result.Warnings = append(result.Warnings, "live reference is not a durable tracked dependency: "+rel)
	}
	sort.Strings(result.Warnings)
	if len(payload) > maxWorkingdocBytes {
		return result, fmt.Errorf("projected copy exceeds 1 MiB")
	}
	// Reserve the complete supported pointer before publication, so stamping
	// cannot turn an at-limit source into an unreadable document.
	reserved, err := stampWorkingdoc(string(before), settlementPointer{target, strings.Repeat("f", 64)})
	if err != nil {
		return result, err
	}
	if len(reserved) > maxWorkingdocBytes {
		return result, fmt.Errorf("source has no room for its settlement pointer within 1 MiB")
	}
	if err = repo.Prepare(ctx, root, target); err != nil {
		return result, err
	}
	prior, hadTarget, err := readWorkingdoc(root, target)
	if err != nil {
		return result, err
	}
	expected := ""
	if m.Pointer != nil {
		expected = m.Pointer.Commit
		blobs, err := repo.ReadBaselines(ctx, root, []settlementPointer{*m.Pointer})
		if err != nil {
			return result, err
		}
		baseline, ok := blobs[*m.Pointer]
		if !ok {
			return result, fmt.Errorf("unverifiable settlement baseline")
		}
		if !hadTarget || (!bytes.Equal(prior, baseline) && !bytes.Equal(prior, payload)) {
			return result, fmt.Errorf("destination conflict: %s; reconcile its independent edits", target)
		}
	} else if hadTarget && !bytes.Equal(prior, payload) {
		return result, fmt.Errorf("destination collision: %s", target)
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = prepareWorkingdoc(ss, root, target, prior, hadTarget, regs[targetFull] != nil, payload); err != nil {
		return result, err
	}
	state := settlementState{Phase: settlementPlanned}
	effect := nextSettlementEffect(state)
	stop := func(event settlementEvent, cause error) error {
		var transitionErr error
		state, effect, transitionErr = transitionSettlement(state, event)
		return errors.Join(cause, transitionErr)
	}
	var commit string
	err = runSettlementEffect(state, effect, settlementEffectPublish, func() error {
		var publishErr error
		commit, publishErr = repo.Publish(ctx, root, target, payload, expected)
		return publishErr
	})
	if err != nil {
		var warning interface {
			error
			PublishedCommit() string
		}
		if !errors.As(err, &warning) || warning.PublishedCommit() != commit {
			return result, stop(settlementUnconfirmed, fmt.Errorf("copy prepared; publication failed or unconfirmed, source not stamped; retry the same command: %w", err))
		}
		result.Warnings = append(result.Warnings, warning.Error())
	}
	result.Commit = commit
	pointer := settlementPointer{target, commit}
	if _, err = parseSettlementPointer(target + " @ " + commit); err != nil {
		return result, stop(settlementFailed, fmt.Errorf("publisher returned invalid receipt: %w", err))
	}
	blobs, err := repo.ReadBaselines(ctx, root, []settlementPointer{pointer})
	if err != nil {
		return result, stop(settlementUnconfirmed, fmt.Errorf("publication readback unconfirmed; source not stamped: %w", err))
	}
	blob, ok := blobs[pointer]
	if !ok || !bytes.Equal(blob, payload) {
		return result, stop(settlementFailed, fmt.Errorf("published copy differs from the validated snapshot; source not stamped"))
	}
	state, effect, err = transitionSettlement(state, settlementPublished)
	if err != nil {
		return result, err
	}
	result.Published = true
	stamped, err := stampWorkingdoc(string(before), pointer)
	if err != nil {
		return result, stop(settlementFailed, err)
	}
	if err = ctx.Err(); err != nil {
		return result, stop(settlementFailed, fmt.Errorf("copy published; source not stamped: %w", err))
	}
	if err = runSettlementEffect(state, effect, settlementEffectStamp, func() error { return applyWorkingdoc(ss, sourceFull, string(before), stamped) }); err != nil {
		current, _, readErr := readWorkingdoc(root, source)
		if readErr != nil || string(current) != stamped {
			return result, stop(settlementUnconfirmed, fmt.Errorf("copy published at %s; source stamp failed or unconfirmed, reconcile source and retry: %w", commit, err))
		}
	}
	state, effect, err = transitionSettlement(state, settlementStamped)
	if err != nil {
		return result, err
	}
	result.Stamped = state.Phase == settlementDone
	return result, nil
}

// Every externally visible settlement effect passes this authorization seam.
// The pure phase model supplies its next effect; a terminal or mismatched
// request cannot call the repository or Couch even if a caller is reordered.
func runSettlementEffect(state settlementState, effect, requested settlementEffect, apply func() error) error {
	if requested == settlementEffectNone || effect != requested || nextSettlementEffect(state) != effect {
		return fmt.Errorf("settlement phase %d does not authorize effect %d", state.Phase, effect)
	}
	return apply()
}
