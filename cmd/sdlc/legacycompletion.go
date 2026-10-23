package main

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const legacyCompletionHistoryLimit = 10000

type legacyCompletionProof struct {
	Completion    issue.Completion
	HasCompletion bool
	Adoption      issue.LegacyAdoption
}

type legacySourceState struct{ commit, path, fm, status string }

// resolveLegacyCompletion reads only immutable objects. Missing old inputs may
// establish a windowless adoption; malformed saved evidence never does.
func resolveLegacyCompletion(ctx context.Context, root, sourceMain, repository string, source tracker.MigrationFile, plans string) (legacyCompletionProof, error) {
	return resolveLegacyGeneration(ctx, root, sourceMain, repository, source, plans, false)
}

// archiveBoundOnly proves the terminal generation, never review or measured
// credit. Only the old unmarked-import path may request this weaker fact.
func resolveLegacyGeneration(ctx context.Context, root, sourceMain, repository string, source tracker.MigrationFile, plans string, archiveBoundOnly bool) (legacyCompletionProof, error) {
	var out legacyCompletionProof
	fail := func(message string) (legacyCompletionProof, error) {
		return out, fmt.Errorf("legacy completion %s: %s", source.Path, message)
	}
	if !fullLegacyOID(sourceMain) || source.Path == "" || path.Clean(source.Path) != source.Path || path.IsAbs(source.Path) || strings.HasPrefix(source.Path, "../") || strings.ContainsAny(source.Path, "\r\n\x00:\\") {
		return fail("invalid immutable source pin or path")
	}
	id, _, ok := issue.ParseFilename(path.Base(source.Path))
	if !ok {
		return fail("source filename has no issue identity")
	}
	fm, _, err := issue.Parse(string(source.Raw))
	if err != nil {
		return out, err
	}
	sourceFields, err := legacyScalarFields(fm)
	if err != nil {
		return out, err
	}
	fields := map[string]map[string]string{fm: sourceFields}
	field := func(fm, key string) string { return fields[fm][key] }
	if field(fm, "id") != id || field(fm, "status") != "done" {
		return fail("source is not the matching terminal done issue")
	}
	if v := field(fm, "actual_hours"); v != "" {
		n, e := strconv.ParseFloat(v, 64)
		if e != nil || n < 0 || n != n || n > 1e9 {
			return fail("invalid adopted actual_hours")
		}
	}
	if v := field(fm, "started"); v != "" {
		if _, e := time.Parse(time.RFC3339Nano, v); e != nil {
			return fail("invalid started timestamp")
		}
	}
	raw, err := archiveReadContext(ctx, root, "cat-file", "blob", sourceMain+":"+source.Path)
	if err != nil {
		return out, err
	}
	if !bytes.Equal(raw, source.Raw) {
		return fail("source bytes differ from pinned main")
	}
	blob, err := ownedActualGit(ctx, root, "rev-parse", "--verify", sourceMain+":"+source.Path)
	if err != nil {
		return out, err
	}
	out.Adoption = issue.LegacyAdoption{Repository: repository, SourceMain: sourceMain, SourcePath: source.Path, SourceBlob: blob, WindowClass: issue.LegacyWindowlessTerminal}
	// --follow performs one bounded history walk, including a real archive
	// rename. The low threshold preserves identity checks across heavily edited
	// archives; a copied non-issue template is bounded by a proven issue birth.
	history, err := ownedActualGit(ctx, root, "--literal-pathspecs", "log", "--topo-order", "--full-history", "--follow", "-M1%", "--format=@@%H", "--name-only", "--max-count="+strconv.Itoa(legacyCompletionHistoryLimit+1), sourceMain, "--", source.Path)
	if err != nil {
		return out, err
	}
	blocks := strings.Split(strings.TrimSpace(history), "@@")
	if len(blocks)-1 > legacyCompletionHistoryLimit {
		return fail("source history exceeds complete-history bound")
	}
	var states []legacySourceState
	for blockIndex, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) == 0 || lines[0] == "" {
			continue
		}
		if !fullLegacyOID(lines[0]) {
			return fail("malformed or incomplete source history")
		}
		var p string
		for _, candidate := range lines[1:] {
			if candidate == "" {
				continue
			}
			candidateID, _, valid := issue.ParseFilename(path.Base(candidate))
			if valid && candidateID == id {
				if p != "" {
					return fail("ambiguous renamed source paths")
				}
				p = candidate
			}
		}
		if p == "" {
			// Git --follow may treat an issue created from a project template
			// as a copy and continue into that project's history (Echo #1).
			// Stop only after proving creation at the last issue observation;
			// a missing path within an existing generation still refuses.
			if len(states) > 0 {
				birth, e := legacyHistoryIssueBirth(ctx, root, states[len(states)-1])
				if e != nil {
					return out, e
				}
				if birth {
					for _, remaining := range blocks[blockIndex:] {
						for _, candidate := range strings.Split(strings.TrimSpace(remaining), "\n")[1:] {
							if candidateID, _, valid := issue.ParseFilename(path.Base(candidate)); valid && candidateID == id {
								return fail("ambiguous issue ancestry beyond its creation")
							}
						}
					}
					break
				}
			}
			return fail("source history lost its path")
		}
		text, e := ownedActualGit(ctx, root, "show", lines[0]+":"+p)
		if e != nil {
			return out, e
		}
		oldFM, _, e := issue.Parse(text)
		if e != nil {
			return out, e
		}
		oldFields, e := legacyScalarFields(oldFM)
		if e != nil {
			return out, e
		}
		fields[oldFM] = oldFields
		if field(oldFM, "id") != id {
			return fail("source history changes issue identity")
		}
		for _, key := range []string{"created", "started"} {
			a, b := field(oldFM, key), field(fm, key)
			if a != "" && b != "" && a != b {
				return fail("source history changes " + key + " generation")
			}
		}
		states = append(states, legacySourceState{lines[0], p, oldFM, field(oldFM, "status")})
	}
	if len(states) == 0 || len(states) > legacyCompletionHistoryLimit {
		return fail("source history is missing or exceeds complete-history bound")
	}
	var observations []tracker.LegacyCloseEvidence
	messages := make(map[string]string, len(states))
	for i, state := range states {
		message, e := ownedActualGit(ctx, root, "show", "-s", "--format=%B", state.commit)
		if e != nil {
			return out, e
		}
		messages[state.commit] = message
		bearing := strings.Contains(message, "Review-Verdict:") || strings.Contains(message, "Review-Window:")
		if !archiveBoundOnly && !bearing && (state.status == "done" || state.status == "codecomplete") && strings.Contains(strings.Split(message, "\n")[0], ": close") {
			return fail("close is missing review evidence")
		}
		observations = append(observations, tracker.LegacyCloseEvidence{Position: i, Commit: state.commit, Path: state.path, Status: state.status, Actual: field(state.fm, "actual_hours"), Started: field(state.fm, "started"), HasReview: bearing})
	}
	selected, err := tracker.SelectLegacyClose(observations, field(fm, "actual_hours"), field(fm, "started"))
	if err != nil {
		return out, err
	}
	out.Adoption.TerminalCommit = selected.TerminalCommit
	if selected.CloseCommit == "" {
		return out, nil
	}
	for _, candidate := range observations {
		if candidate.Commit == selected.CloseCommit || !candidate.HasReview || (candidate.Status != "done" && candidate.Status != "codecomplete") {
			continue
		}
		if _, err := ownedActualGit(ctx, root, "merge-base", "--is-ancestor", candidate.Commit, selected.CloseCommit); err != nil {
			return fail("incomparable legacy close generations")
		}
	}
	if archiveBoundOnly {
		return out, nil // HasCompletion stays false; no review/actual is certified.
	}

	for _, s := range states {
		if s.commit != selected.CloseCommit {
			continue
		}
		message := messages[s.commit]
		verdicts := 0
		for _, line := range strings.Split(message, "\n") {
			if strings.HasPrefix(line, "Review-Verdict:") {
				verdicts++
				if strings.TrimSpace(strings.TrimPrefix(line, "Review-Verdict:")) != "SHIP" {
					return fail("latest close did not ship")
				}
			}
		}
		if verdicts != 1 {
			return fail("close requires one SHIP review verdict")
		}
		base, head, e := ownedReviewWindow(message)
		if e != nil {
			return out, e
		}
		resolve := func(pin string) (string, error) {
			return ownedActualGit(ctx, root, "rev-parse", "--verify", pin+"^{commit}")
		}
		base, e = resolve(base)
		if e != nil {
			return out, e
		}
		head, e = resolve(head)
		if e != nil {
			return out, e
		}
		for _, pair := range [][2]string{{base, head}, {head, s.commit}, {s.commit, sourceMain}} {
			if _, e = ownedActualGit(ctx, root, "merge-base", "--is-ancestor", pair[0], pair[1]); e != nil {
				return fail("review/close ancestry mismatch")
			}
		}
		completion := issue.Completion{Token: tracker.MigrationToken(id), Repository: repository, ReviewedHEAD: head, EvidenceCommit: s.commit}
		num, _ := strconv.Atoi(id)
		ledger, e := ownedLedgerAt(ctx, root, s.commit, num, plans)
		if e != nil {
			return out, e
		}
		reviewed, found := gatestate.LatestFinalizedAt(ledger, "")
		if found && reviewed != head {
			return fail("close boundary ledger reviewed a different head")
		}
		if len(ledger.Rounds) > 0 && !found {
			return fail("saved boundary ledger has no finalized whole-issue close")
		}
		// The shared snapshot reader validates retained input structure/repository/head.
		rec := tracker.IssueRecord{ID: id, DetailFM: s.fm}
		saved, _, _, e := completedMeasurementSnapshot(ctx, root, rec, completion, repository, plans, false)
		if e != nil {
			return out, e
		}
		out.Completion = completion
		out.HasCompletion = true
		out.Adoption.CompletionToken = completion.Token
		out.Adoption.WindowClass = issue.LegacyWindowlessClose
		if saved != nil {
			out.Adoption.WindowClass = issue.LegacyRetained
		}
		return out, nil
	}
	return out, nil
}

// legacyHistoryIssueBirth proves the path was absent in every parent of its
// oldest issue observation. This bounds --follow's heuristic copy ancestry;
// neither similarity nor a missing name in log output proves a new generation.
func legacyHistoryIssueBirth(ctx context.Context, root string, state legacySourceState) (bool, error) {
	parents, err := ownedActualGit(ctx, root, "rev-list", "--parents", "-n", "1", state.commit)
	if err != nil {
		return false, err
	}
	parts := strings.Fields(parents)
	if len(parts) == 0 || parts[0] != state.commit {
		return false, fmt.Errorf("invalid issue creation ancestry")
	}
	for _, parent := range parts[1:] {
		present, err := ownedActualGit(ctx, root, "--literal-pathspecs", "ls-tree", "--name-only", parent, "--", state.path)
		if err != nil {
			return false, err
		}
		if present != "" {
			return false, nil
		}
	}
	return true, nil
}

func fullLegacyOID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

// verifyLegacyAdoption grants a bound to explicit windowless provenance or
// an old unbound terminal-only import proven against the actual cutover root.
func verifyLegacyAdoption(ctx context.Context, root, trackerPin string, rec tracker.IssueRecord, repository, plans string) (time.Time, error) {
	var zero time.Time
	if rec.Card == nil {
		return zero, nil
	}
	a, ok, err := issue.CardLegacyAdoption(rec.Card.Raw)
	if err != nil {
		return zero, err
	}
	if !ok {
		return recoverLegacyTerminalAdoption(ctx, root, trackerPin, rec, repository, plans)
	}
	if a.WindowClass == issue.LegacyRetained {
		return zero, nil
	}
	if a.Repository != repository || rec.Status() != "done" {
		return zero, fmt.Errorf("legacy adoption repository/status differs from current generation")
	}
	if _, corrected, e := issue.CardActualCorrection(rec.Card.Raw); e != nil || corrected {
		if e != nil {
			return zero, e
		}
		return zero, fmt.Errorf("legacy adoption has an actual correction")
	}
	sourceMain, trackerRoot, err := legacyMigrationSource(ctx, root, "HEAD")
	if err != nil {
		return zero, err
	}
	if sourceMain != a.SourceMain {
		return zero, fmt.Errorf("legacy adoption source main differs from actual cutover parent")
	}
	importedRaw, err := archiveReadContext(ctx, root, "cat-file", "blob", trackerRoot+":"+rec.Card.Path)
	if err != nil {
		return zero, fmt.Errorf("legacy adoption lacks imported tracker-root card: %w", err)
	}
	imported, err := issue.ParseCard(importedRaw)
	if err != nil {
		return zero, err
	}
	if imported.ID != rec.ID {
		return zero, fmt.Errorf("legacy adoption imported identity differs")
	}
	for _, key := range []string{"status", "actual_hours", "started"} {
		value, err := issue.FieldString(imported.Frontmatter, key)
		if err != nil {
			return zero, err
		}
		if value != rec.Text(key) {
			return zero, fmt.Errorf("legacy adoption imported %s differs from current generation", key)
		}
	}
	raw, err := archiveReadContext(ctx, root, "cat-file", "blob", a.SourceMain+":"+a.SourcePath)
	if err != nil {
		return zero, err
	}
	proof, err := resolveLegacyCompletion(ctx, root, a.SourceMain, repository, tracker.MigrationFile{Path: a.SourcePath, Raw: raw}, plans)
	if err != nil {
		return zero, err
	}
	if proof.Adoption != a {
		return zero, fmt.Errorf("legacy adoption differs from immutable source proof")
	}
	importedCompletion, importedHas, err := issue.CardCompletion(importedRaw)
	if err != nil {
		return zero, err
	}
	if importedHas && (!proof.HasCompletion || importedCompletion != proof.Completion) {
		return zero, fmt.Errorf("legacy adoption imported completion differs from source close generation")
	}
	importedAdoption, importedMarked, err := issue.CardLegacyAdoption(importedRaw)
	if err != nil {
		return zero, err
	}
	if importedMarked && importedAdoption != proof.Adoption {
		return zero, fmt.Errorf("legacy adoption imported provenance differs from source generation")
	}
	if _, corrected, err := issue.CardActualCorrection(importedRaw); err != nil || corrected {
		if err != nil {
			return zero, err
		}
		return zero, fmt.Errorf("legacy adoption imported generation has an actual correction")
	}

	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return zero, err
	}
	for _, key := range []string{"id", "actual_hours", "started"} {
		v, err := issue.FieldString(fm, key)
		if err != nil {
			return zero, err
		}
		if rec.Text(key) != v {
			return zero, fmt.Errorf("legacy adoption current %s differs from source generation", key)
		}
	}
	completion, has, err := issue.CardCompletion(rec.Card.Raw)
	if err != nil {
		return zero, err
	}
	if has != proof.HasCompletion || (has && completion != proof.Completion) {
		return zero, fmt.Errorf("legacy adoption completion differs from source close generation")
	}
	if err := legacyGenerationHistory(ctx, root, trackerPin, trackerRoot, rec, proof); err != nil {
		return zero, err
	}
	pin := a.TerminalCommit
	if proof.HasCompletion {
		pin = proof.Completion.EvidenceCommit
	}
	timestamp, err := ownedActualGit(ctx, root, "show", "-s", "--format=%cI", pin)
	if err != nil {
		return zero, err
	}
	end, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return zero, fmt.Errorf("legacy adoption object timestamp: %w", err)
	}
	return end, nil
}

// Decode scalars before identity/generation comparisons: YAML quoting has no
// authority, while a structured value in a known scalar field is malformed.
func legacyScalarFields(fm string) (map[string]string, error) {
	values := make(map[string]string, 5)
	for _, key := range []string{"id", "status", "created", "started", "actual_hours"} {
		value, err := issue.FieldString(fm, key)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, nil
}

// recoverUnboundLegacyAdoption is read-only compatibility for original
// migrations that retained numeric done cards without the newer provenance.
// Root membership and exact archived generation are positive authority; a
// missing completion on any other card is never evidence of a legacy close.
func recoverUnboundLegacyAdoption(ctx context.Context, root, trackerPin string, rec tracker.IssueRecord, repository, plans string) (issue.LegacyAdoption, bool, error) {
	return recoverUnboundLegacyGeneration(ctx, root, trackerPin, rec, repository, plans, false)
}

func recoverUnboundLegacyGeneration(ctx context.Context, root, trackerPin string, rec tracker.IssueRecord, repository, plans string, archiveBoundOnly bool) (issue.LegacyAdoption, bool, error) {
	var zero issue.LegacyAdoption
	if rec.Card == nil || rec.Status() != "done" {
		return zero, false, nil
	}
	if _, completed, err := issue.CardCompletion(rec.Card.Raw); err != nil || completed {
		return zero, false, err
	}
	if _, corrected, err := issue.CardActualCorrection(rec.Card.Raw); err != nil || corrected {
		return zero, false, err
	}
	if _, marked, err := issue.CardLegacyAdoption(rec.Card.Raw); err != nil || marked {
		return zero, false, err
	}
	actual, err := strconv.ParseFloat(rec.Text("actual_hours"), 64)
	if err != nil || actual < 0 || actual != actual || actual > 1e9 {
		return zero, false, nil
	}
	sourceMain, trackerRoot, err := legacyMigrationSource(ctx, root, "HEAD")
	if err != nil {
		return zero, false, err
	}
	listed, err := ownedActualGit(ctx, root, "--literal-pathspecs", "ls-tree", "--name-only", trackerRoot, "--", rec.Card.Path)
	if err != nil {
		return zero, false, err
	}
	if listed == "" {
		return zero, false, nil
	}
	raw, err := archiveReadContext(ctx, root, "cat-file", "blob", trackerRoot+":"+rec.Card.Path)
	if err != nil {
		return zero, false, err
	}
	card, err := issue.ParseCard(raw)
	if err != nil {
		return zero, false, err
	}
	if card.ID != rec.ID {
		return zero, false, fmt.Errorf("legacy terminal imported identity differs")
	}
	for _, key := range []string{"status", "actual_hours", "started"} {
		value, err := issue.FieldString(card.Frontmatter, key)
		if err != nil {
			return zero, false, err
		}
		if value != rec.Text(key) {
			return zero, false, nil
		}
	}
	if _, completed, err := issue.CardCompletion(raw); err != nil || completed {
		return zero, false, err
	}
	if _, corrected, err := issue.CardActualCorrection(raw); err != nil || corrected {
		return zero, false, err
	}
	// Explicit newer root provenance goes through its ordinary verifier;
	// stripping it from a current card cannot activate the old-import path.
	if _, marked, err := issue.CardLegacyAdoption(raw); err != nil || marked {
		return zero, false, err
	}
	imported := tracker.Record{ID: card.ID, Path: rec.Card.Path, Card: card, Raw: raw}
	source, err := legacyImportedSource(ctx, root, sourceMain, repository, imported)
	if err != nil {
		return zero, false, err
	}
	if path.Dir(source.Path) != vocab.ArchiveSubdir(vocab.Issue().Discovery().Archive, vocab.ArchiveIssues) {
		return zero, false, nil
	}
	proof, err := resolveLegacyGeneration(ctx, root, sourceMain, repository, source, plans, archiveBoundOnly)
	if err != nil {
		return zero, false, err
	}
	if proof.HasCompletion || proof.Adoption.WindowClass != issue.LegacyWindowlessTerminal {
		return zero, false, nil
	}
	if err := legacyGenerationHistory(ctx, root, trackerPin, trackerRoot, rec, proof); err != nil {
		return zero, false, err
	}
	return proof.Adoption, true, nil
}

func recoverLegacyTerminalAdoption(ctx context.Context, root, trackerPin string, rec tracker.IssueRecord, repository, plans string) (time.Time, error) {
	adoption, proven, err := recoverUnboundLegacyAdoption(ctx, root, trackerPin, rec, repository, plans)
	if err != nil || !proven {
		return time.Time{}, err
	}
	return legacyTerminalEnd(ctx, root, adoption)
}

func legacyTerminalEnd(ctx context.Context, root string, adoption issue.LegacyAdoption) (time.Time, error) {
	timestamp, err := ownedActualGit(ctx, root, "show", "-s", "--format=%cI", adoption.TerminalCommit)
	if err != nil {
		return time.Time{}, err
	}
	end, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("legacy terminal object timestamp: %w", err)
	}
	return end, nil
}

// An archive bound is relevance evidence only. Strict close, repair and replay
// callers continue through recoverLegacyTerminalAdoption instead.
func recoverLegacyArchiveBound(ctx context.Context, root, trackerPin string, rec tracker.IssueRecord, repository, plans string) (time.Time, error) {
	adoption, proven, err := recoverUnboundLegacyGeneration(ctx, root, trackerPin, rec, repository, plans, true)
	if err != nil || !proven {
		return time.Time{}, err
	}
	return legacyTerminalEnd(ctx, root, adoption)
}

// legacyGenerationHistory is the shared repair and exclusion authority fence.
// It proves continuity on the caller's immutable tracker snapshot; matching
// endpoints alone cannot distinguish a reopened generation.
// Only missing -> exact matching repair metadata may accumulate along the way.
func legacyGenerationHistory(ctx context.Context, root, trackerPin, trackerRoot string, rec tracker.IssueRecord, proof legacyCompletionProof) error {
	if !fullLegacyOID(trackerPin) || !fullLegacyOID(trackerRoot) || rec.Card == nil {
		return fmt.Errorf("legacy adoption requires an immutable tracker snapshot")
	}
	if _, err := ownedActualGit(ctx, root, "merge-base", "--is-ancestor", trackerRoot, trackerPin); err != nil {
		return fmt.Errorf("legacy adoption tracker snapshot does not descend from imported root: %w", err)
	}
	current, err := archiveReadContext(ctx, root, "cat-file", "blob", trackerPin+":"+rec.Card.Path)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, rec.Card.Raw) {
		return fmt.Errorf("legacy adoption current card differs from pinned tracker snapshot")
	}
	blob, err := ownedActualGit(ctx, root, "rev-parse", "--verify", trackerPin+":"+rec.Card.Path)
	if err != nil {
		return err
	}
	if rec.Card.BlobOID != "" && rec.Card.BlobOID != blob {
		return fmt.Errorf("legacy adoption card blob differs from pinned tracker snapshot")
	}
	history, err := ownedActualGit(ctx, root, "--literal-pathspecs", "log", "--full-history", "--topo-order", "--format=%H", "--max-count="+strconv.Itoa(legacyCompletionHistoryLimit+1), trackerPin, "--", rec.Card.Path)
	if err != nil {
		return fmt.Errorf("legacy adoption tracker history: %w", err)
	}
	commits := strings.Fields(history)
	if len(commits) > legacyCompletionHistoryLimit {
		return fmt.Errorf("legacy adoption tracker history exceeds complete-history bound")
	}
	rootIndex := -1
	for i, commit := range commits {
		if commit == trackerRoot {
			rootIndex = i
			break
		}
	}
	if rootIndex < 0 {
		return fmt.Errorf("legacy adoption tracker history does not reach imported root")
	}
	hadCompletion, hadAdoption := false, false
	for i := rootIndex; i >= 0; i-- {
		raw, err := archiveReadContext(ctx, root, "cat-file", "blob", commits[i]+":"+rec.Card.Path)
		if err != nil {
			return fmt.Errorf("legacy adoption tracker generation read: %w", err)
		}
		card, err := issue.ParseCard(raw)
		if err != nil {
			return err
		}
		if card.ID != rec.ID {
			return fmt.Errorf("legacy adoption tracker history changes issue identity")
		}
		for _, key := range []string{"status", "actual_hours", "started"} {
			value, err := issue.FieldString(card.Frontmatter, key)
			if err != nil {
				return err
			}
			if value != rec.Text(key) {
				return fmt.Errorf("legacy adoption tracker history changes %s generation", key)
			}
		}
		if _, corrected, err := issue.CardActualCorrection(raw); err != nil || corrected {
			if err != nil {
				return err
			}
			return fmt.Errorf("legacy adoption tracker history contains an actual correction")
		}
		completion, completed, err := issue.CardCompletion(raw)
		if err != nil {
			return err
		}
		if completed && (!proof.HasCompletion || completion != proof.Completion) {
			return fmt.Errorf("legacy adoption tracker history changes completion generation")
		}
		if hadCompletion && !completed {
			return fmt.Errorf("legacy adoption tracker history removes completion evidence")
		}
		adoption, marked, err := issue.CardLegacyAdoption(raw)
		if err != nil {
			return err
		}
		if marked && adoption != proof.Adoption {
			return fmt.Errorf("legacy adoption tracker history changes provenance generation")
		}
		if hadAdoption && !marked {
			return fmt.Errorf("legacy adoption tracker history removes provenance evidence")
		}
		hadCompletion, hadAdoption = completed, marked
	}
	return nil
}
