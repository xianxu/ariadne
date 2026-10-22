package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// legacyMigrationSource binds a legacy proof to the actual cutover, rather
// than a caller-supplied historical commit which happens to hold similar data.
func legacyMigrationSource(ctx context.Context, root, main string) (sourceMain, trackerRoot string, err error) {
	read := func(args ...string) (string, error) {
		raw, err := archiveReadContext(ctx, root, args...)
		return strings.TrimSpace(string(raw)), err
	}
	raw, err := read("cat-file", "blob", main+":"+tracker.CutoverMarkerPath)
	if err != nil {
		return "", "", fmt.Errorf("read migration authority: %w", err)
	}
	trackerRoot, err = tracker.ParseCutoverMarker([]byte(raw))
	if err != nil {
		return "", "", err
	}
	history, err := read("--literal-pathspecs", "log", "--format=%H", "--diff-filter=A", "--max-count=10001", main, "--", tracker.CutoverMarkerPath)
	if err != nil {
		return "", "", fmt.Errorf("read cutover history: %w", err)
	}
	commits := strings.Fields(history)
	if len(commits) != 1 {
		return "", "", fmt.Errorf("migration authority requires exactly one marker-add commit, found %d", len(commits))
	}
	parents, err := read("rev-list", "--parents", "-n", "1", commits[0])
	if err != nil {
		return "", "", err
	}
	parts := strings.Fields(parents)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("cutover must have one unambiguous source parent")
	}
	marker, err := read("cat-file", "blob", commits[0]+":"+tracker.CutoverMarkerPath)
	if err != nil {
		return "", "", err
	}
	atCutover, err := tracker.ParseCutoverMarker([]byte(marker))
	if err != nil || atCutover != trackerRoot {
		return "", "", fmt.Errorf("migration marker changed its tracker root")
	}
	before, err := read("--literal-pathspecs", "ls-tree", "-z", parts[1], "--", tracker.CutoverMarkerPath)
	if err != nil {
		return "", "", err
	}
	if before != "" {
		return "", "", fmt.Errorf("cutover source already has a marker")
	}
	return parts[1], trackerRoot, nil
}

type legacyCompletionRepairPlan struct {
	Root, Source, Digest string
	Current              tracker.Record
	Replacement          []byte
	Complete             bool
	Proof                legacyCompletionProof
}

func runLegacyCompletionRepair(env *migrateEnv, stdout, stderr io.Writer, f issueMigrateFlags) error {
	plan, err := prepareLegacyCompletionRepair(env, fmt.Sprintf("%06d", f.Issue))
	if err != nil {
		return err
	}
	if plan.Complete {
		fmt.Fprintf(stdout, "issue #%d: legacy completion already repaired; no change\n", f.Issue)
		return nil
	}
	fmt.Fprintf(stdout, "issue #%d: repair legacy close %s (%s); actual unchanged\ndigest: %s\n", f.Issue, plan.Proof.Completion.EvidenceCommit, plan.Proof.Adoption.WindowClass, plan.Digest)
	command := fmt.Sprintf("sdlc issue migrate --repair-completions --issue %d --apply --expect %s", f.Issue, plan.Digest)
	if !f.Apply {
		fmt.Fprintf(stdout, "dry run; apply: %s\n", command)
		return nil
	}
	if f.Expect != plan.Digest {
		return fmt.Errorf("repair plan is %s, not reviewed %q; rerun the dry run", plan.Digest, f.Expect)
	}
	fmt.Fprintf(stderr, "Recovery: %s\n", command)
	err = applyLegacyCompletionRepair(env.ctx, env.tracker, plan, func(base, candidate string) error {
		if err := env.ctx.Err(); err != nil {
			return err
		}
		fmt.Fprintf(stderr, "repair candidate %s (tracker base %s)\n", candidate, base)
		return nil
	})
	if err != nil {
		return fmt.Errorf("legacy completion repair outcome not confirmed: %w; inspect with sdlc issue migrate --repair-completions --issue %d before retrying", err, f.Issue)
	}
	fmt.Fprintf(stdout, "issue #%d: legacy completion repaired (%s); checkout and actual unchanged\n", f.Issue, plan.Proof.Adoption.WindowClass)
	return nil
}

func prepareLegacyCompletionRepair(env *migrateEnv, id string) (legacyCompletionRepairPlan, error) {
	var plan legacyCompletionRepairPlan
	main, err := env.main.Snapshot()
	if err != nil {
		return plan, err
	}
	root, migrated, err := migratedAlready(env, main)
	if err != nil {
		return plan, err
	}
	if !migrated {
		return plan, fmt.Errorf("completion repair requires an already migrated repository")
	}
	source, sourceRoot, err := legacyMigrationSource(env.ctx, env.root, main.Ref())
	if err != nil {
		return plan, err
	}
	if sourceRoot != root {
		return plan, fmt.Errorf("cutover tracker authority changed")
	}
	snapshot, err := env.tracker.Snapshot()
	if err != nil {
		return plan, err
	}
	current, err := snapshot.Require(id)
	if err != nil {
		return plan, err
	}
	importedRaw, err := env.trunk.ReadAt(root, current.Path)
	if err != nil {
		return plan, fmt.Errorf("issue #%s was not imported at the tracker root: %w", issue.CLIRef(id), err)
	}
	importedCard, err := issue.ParseCard(importedRaw)
	if err != nil {
		return plan, err
	}
	imported := tracker.Record{ID: id, Path: current.Path, Raw: importedRaw, Card: importedCard}
	file, err := legacyRepairSource(env, source, imported)
	if err != nil {
		return plan, err
	}
	proof, err := resolveLegacyCompletion(env.ctx, env.root, source, env.target.Repository, file, plansDir())
	if err != nil {
		return plan, err
	}
	if !proof.HasCompletion {
		return plan, fmt.Errorf("issue #%s has no provable legacy close; terminal-only imports cannot fabricate one", issue.CLIRef(id))
	}
	record := tracker.IssueRecord{ID: current.ID, Card: &current}
	if err := legacyGenerationHistory(env.ctx, env.root, snapshot.Ref(), root, record, proof); err != nil {
		return plan, err
	}
	raw, complete, err := tracker.PlanLegacyCompletionRepair(imported, current, proof.Completion, proof.Adoption, true)
	if err != nil {
		return plan, err
	}
	plan = legacyCompletionRepairPlan{Root: root, Source: source, Current: current, Replacement: raw, Complete: complete, Proof: proof}
	plan.Digest = tracker.LegacyCompletionRepairDigest(env.target.Repository, root, source, current.Path, current.BlobOID, raw)
	return plan, nil
}

// legacyRepairSource reuses the pure migration selector against the actual
// immutable pre-cutover tree, including its duplicate-ID policy.
func legacyRepairSource(env *migrateEnv, source string, imported tracker.Record) (tracker.MigrationFile, error) {
	return legacyImportedSource(env.ctx, env.root, source, env.target.Repository, imported)
}

// legacyImportedSource is shared by explicit repair and read-only recovery of
// old imports. It has no remote publication or checkout capability.
func legacyImportedSource(ctx context.Context, root, source, repository string, imported tracker.Record) (tracker.MigrationFile, error) {
	d := vocab.Issue().Discovery()
	format, err := ownedActualGit(ctx, root, "rev-parse", "--show-object-format")
	if err != nil {
		return tracker.MigrationFile{}, err
	}
	in := tracker.MigrationInput{Repository: repository, ObjectFormat: format, Main: source}
	files := map[string]tracker.MigrationFile{}
	archiveDir := vocab.ArchiveSubdir(d.Archive, vocab.ArchiveIssues)
	listing, err := archiveReadContext(ctx, root, "--literal-pathspecs", "ls-tree", "-r", "-z", "--name-only", source, "--", d.Home, archiveDir)
	if err != nil {
		return tracker.MigrationFile{}, err
	}
	for _, name := range strings.Split(string(listing), "\x00") {
		id, _, ok := issue.ParseFilename(path.Base(name))
		if !ok || id != imported.ID || (path.Dir(name) != d.Home && path.Dir(name) != archiveDir) {
			continue
		}
		raw, err := archiveReadContext(ctx, root, "cat-file", "blob", source+":"+name)
		if err != nil {
			return tracker.MigrationFile{}, err
		}
		file := tracker.MigrationFile{Path: name, Raw: raw}
		files[name] = file
		if path.Dir(name) == d.Home {
			in.Active = append(in.Active, file)
		} else {
			in.Archived = append(in.Archived, file)
		}
	}
	m := tracker.PlanTrackerMigration(in)
	for _, c := range m.Cards {
		if c.ID == imported.ID {
			if _, err := issue.MirrorDetails(c.Raw, imported.Raw, format); err != nil {
				return tracker.MigrationFile{}, fmt.Errorf("source generation differs from imported card: %w", err)
			}
			return files[c.Source], nil
		}
	}
	return tracker.MigrationFile{}, fmt.Errorf("imported issue #%s has no unambiguous source at cutover", issue.CLIRef(imported.ID))
}

type legacyRepairPublisher interface {
	UpdateCardWithTrailers(tracker.Record, []byte, string, []string, func(string, string) error) error
}

func applyLegacyCompletionRepair(ctx context.Context, publisher legacyRepairPublisher, plan legacyCompletionRepairPlan, beforePush func(string, string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if plan.Complete {
		return nil
	}
	if plan.Digest == "" || len(plan.Replacement) == 0 {
		return errors.New("legacy repair has no reviewed replacement")
	}
	if beforePush == nil {
		return errors.New("legacy repair requires a publication observer")
	}
	return publisher.UpdateCardWithTrailers(plan.Current, plan.Replacement, "legacy-repair-"+plan.Current.ID+"-"+plan.Digest, []string{"Legacy-Completion-Source: " + plan.Source}, func(base, candidate string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := beforePush(base, candidate); err != nil {
			return err
		}
		return ctx.Err()
	})
}
