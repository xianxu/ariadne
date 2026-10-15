package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/actualcorrection"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// The seam returns production measured inputs; tests supply a bounded known
// snapshot while exercising real tracker/calibration/main publication.
var measureActualCorrection = func(ctx context.Context, root, brain, id, milestone, plans string) actualResult {
	return computeActualFor(ctx, root, brain, id, milestone, plans)
}

func newIssueCorrectActualCmd() *cobra.Command {
	var id int
	var reason, brain, targetRow string
	var dry, exclude bool
	cmd := markMutatingCommand(&cobra.Command{Use: "correct-actual", Short: "Correct a completed issue's actual from pinned measured evidence", Args: cobra.NoArgs, SilenceErrors: true})
	cmd.RunE = func(c *cobra.Command, _ []string) error {
		return withRequiredRepoTransactionLock(c, func() error {
			return runActualCorrection(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), envOr("WF_ISSUES_DIR", "workshop/issues"), id, reason, brain, dry, exclude, targetRow)
		})
	}
	cmd.Flags().IntVar(&id, "issue", 0, "completed issue ID (required)")
	cmd.Flags().StringVar(&reason, "reason", "", "reason for the correction (required)")
	cmd.Flags().StringVar(&brain, "brain-dir", "../brain", "brain repository and calibration destination")
	cmd.Flags().BoolVar(&dry, "dry-run", false, "preview measured correction without writing")
	cmd.Flags().BoolVar(&exclude, "exclude", false, "exclude a proven invalid sample without inventing hours")
	cmd.Flags().StringVar(&targetRow, "target-row", "", "full SHA256 of an eligible calibration row (required when ambiguous)")
	return cmd
}

func correctionDigest(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }

// The same mutex covers ordinary close's ledger rewrite and correction append.
// It uses the existing context-aware lock implementation, not a second journal.
func withActualCalibrationLock(ctx context.Context, stderr io.Writer, ledgerPath string, run func() error) error {
	lock, err := repolock.Acquire(ctx, repolock.Options{GitCommonDir: filepath.Dir(ledgerPath), Command: "calibration", Stderr: stderr, CallerHandlesSignals: true})
	if err != nil {
		return err
	}
	defer lock.Release()
	if err = ctx.Err(); err != nil {
		return err
	}
	return run()
}

func actualCorrectionTarget(text, ref, old, date string) (estimate.LedgerRow, error) {
	return actualCorrectionTargetForGeneration(text, ref, old, date, "")
}

type actualCorrectionAmbiguity struct{ Candidates []estimate.LedgerRow }

func (e *actualCorrectionAmbiguity) Error() string { return "ambiguous calibration close generation" }

func actualCorrectionTargetForGeneration(text, ref, old, date, generation string, selectors ...string) (estimate.LedgerRow, error) {
	selector := ""
	if len(selectors) > 0 {
		selector = selectors[0]
	}
	if selector != "" && !validActualTargetRow(selector) {
		return estimate.LedgerRow{}, errors.New("--target-row must be a full lowercase SHA256 digest; copy an eligible row ID from the ambiguity diagnostic")
	}

	rows, err := estimate.ParseRowsChecked(text)
	if err != nil {
		return estimate.LedgerRow{}, err
	}
	rows, err = estimate.ResolveCorrections(rows)
	if err != nil {
		return estimate.LedgerRow{}, err
	}
	value, err := strconv.ParseFloat(old, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return estimate.LedgerRow{}, errors.New("recorded actual is not numeric; there is no eligible target")
	}
	var candidates []estimate.LedgerRow
	seen := map[string]bool{}
	for _, row := range rows {
		if row.Issue != ref || math.Abs(row.Actual-value) > 0.000001 {
			continue
		}
		if row.Correction != nil {
			if generation != "" && row.Correction.Generation != generation {
				continue
			}
		} else if row.Date != date {
			continue
		}
		if selector != "" && row.RawID != selector {
			continue
		}
		if !seen[row.RawID] {
			candidates = append(candidates, row)
			seen[row.RawID] = true
		}
	}
	if len(candidates) == 0 {
		if selector != "" {
			return estimate.LedgerRow{}, fmt.Errorf("target row %s is not an eligible effective row for this issue, actual and completed close generation; retry without --target-row to inspect eligible targets", selector)
		}
		return estimate.LedgerRow{}, errors.New("no exact effective calibration row for recorded actual and completed close generation")
	}
	if len(candidates) > 1 {
		return estimate.LedgerRow{}, &actualCorrectionAmbiguity{Candidates: candidates}
	}
	return candidates[0], nil
}

func validActualTargetRow(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !(c >= 'a' && c <= 'f' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func actualCorrectionRetryCommand(id int, reason, brain, target string) string {
	return fmt.Sprintf("sdlc issue correct-actual --issue %d --reason %s --brain-dir %s --target-row %s", id, quoteHelperShell(reason), quoteHelperShell(brain), target)
}

func appendActualCorrectionRow(ctx context.Context, stderr io.Writer, file, line, token, digest string) error {
	if correctionDigest([]byte(line)) != digest || strings.ContainsAny(line, "\r\n") {
		return errors.New("correction row digest mismatch")
	}
	return withActualCalibrationLock(ctx, stderr, file, func() error {
		raw, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		rows, err := estimate.ParseRowsChecked(string(raw))
		if err != nil {
			return err
		}
		resolved, err := estimate.ResolveCorrections(rows)
		if err != nil {
			return err
		}
		effective := map[string]bool{}
		for _, row := range resolved {
			effective[row.RawID] = true
		}
		// Header upgrades replace the file atomically under the same ledger
		// lock. Sample lines remain byte-identical, preserving raw identities.
		replace := func(content []byte) error {
			info, err := os.Stat(file)
			if err != nil {
				return err
			}
			temp, err := os.CreateTemp(filepath.Dir(file), ".actual-correction-*")
			if err != nil {
				return err
			}
			defer os.Remove(temp.Name())
			defer temp.Close()
			if err := temp.Chmod(info.Mode()); err != nil {
				return err
			}
			if _, err := temp.Write(content); err != nil {
				return err
			}
			if err := temp.Sync(); err != nil {
				return err
			}
			if err := temp.Close(); err != nil {
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			return os.Rename(temp.Name(), file)
		}
		found := false
		var owned estimate.LedgerRow
		for _, row := range rows {
			if row.Correction != nil && row.Correction.Operation == token {
				if row.RawID != digest {
					return errors.New("calibration operation ID already has different row evidence")
				}
				found = true
				owned = row
			}
		}
		if found {
			if owned.Correction.Disposition == "replace" && !effective[owned.RawID] {
				return errors.New("calibration operation was superseded before confirmation")
			}
			if owned.Correction.Disposition == "exclude" {
				for _, row := range rows {
					if row.Correction != nil {
						for _, target := range row.Correction.Targets {
							if target == owned.RawID {
								return errors.New("calibration exclusion was superseded before confirmation")
							}
						}
					}
				}
			}
			if upgraded, changed := estimate.UpgradeHeader(string(raw)); changed {
				return replace([]byte(upgraded))
			}
			return nil
		}
		// Validate targets against the current ledger before any append. A competing
		// correction that suppressed our exact target cannot be silently adopted.
		added, err := estimate.ParseRowsChecked(line + "\n")
		if err != nil || len(added) != 1 || added[0].Correction == nil || added[0].Correction.Operation != token {
			return errors.New("invalid prepared calibration correction")
		}
		for _, target := range added[0].Correction.Targets {
			if !effective[target] {
				return errors.New("calibration target was superseded before append")
			}
		}
		if _, err = estimate.ResolveCorrections(append(rows, added[0])); err != nil {
			return err
		}
		if upgraded, changed := estimate.UpgradeHeader(string(raw)); changed {
			if !strings.HasSuffix(upgraded, "\n") {
				upgraded += "\n"
			}
			return replace([]byte(upgraded + line + "\n"))
		}
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			return err
		}
		defer f.Close()
		prefix := ""
		if len(raw) > 0 && raw[len(raw)-1] != '\n' {
			prefix = "\n"
		}
		if _, err = f.WriteString(prefix + line + "\n"); err != nil {
			return err
		}
		return f.Sync()
	})
}

type actualCorrectionDocument struct {
	Version            int                         `json:"version"`
	State              actualcorrection.State      `json:"state"`
	Inputs             gatestate.MeasurementInputs `json:"inputs"`
	CardPath           string                      `json:"card_path"`
	MirrorPath         string                      `json:"mirror_path"`
	OriginalMirrorBlob string                      `json:"original_mirror_blob"`
	MirrorBlob         string                      `json:"mirror_blob"`
	Row                string                      `json:"row"`
	TrackerCandidate   gitx.Candidate              `json:"tracker_candidate"`
	MainCandidate      gitx.Candidate              `json:"main_candidate"`
}

func (d actualCorrectionDocument) validate() error {
	if d.Version != 1 {
		return errors.New("unknown actual correction document version")
	}
	if err := actualcorrection.Validate(d.State); err != nil {
		return err
	}
	if err := d.Inputs.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(d.Inputs)
	if err != nil {
		return err
	}
	s := d.State.Spec
	if correctionDigest(raw) != s.MeasurementDigest || d.Inputs.Repository != s.Repository || d.Inputs.CodeHead != s.ReviewedHEAD {
		return errors.New("recovery measurement identity mismatch")
	}
	for _, p := range []string{d.CardPath, d.MirrorPath} {
		if p == "" || path.IsAbs(p) || path.Clean(p) != p || strings.HasPrefix(p, "../") || !strings.HasSuffix(p, ".md") {
			return errors.New("invalid correction document path")
		}
	}
	for _, oid := range []string{d.OriginalMirrorBlob, d.MirrorBlob} {
		if len(oid) != 40 && len(oid) != 64 {
			return errors.New("invalid mirror blob pin")
		}
	}
	if correctionDigest([]byte(d.Row)) != s.CalibrationRowDigest {
		return errors.New("recovery calibration row changed")
	}
	rows, err := estimate.ParseRowsChecked(d.Row + "\n")
	if err != nil || len(rows) != 1 {
		return errors.New("malformed recovery calibration row")
	}
	c := rows[0].Correction
	disposition := "replace"
	if s.Exclude {
		disposition = "exclude"
	}
	if c == nil || c.Operation != s.Token || c.Generation != s.CompletionToken || c.InputDigest != s.MeasurementDigest || c.Disposition != disposition || !reflect.DeepEqual(c.Targets, s.TargetHashes) {
		return errors.New("recovery calibration identity mismatch")
	}
	return nil
}

// Strict field spelling/presence and duplicate rejection are enforced before
// decoding a persisted operation. Reflection derives the schema from its types.
func strictCorrectionJSON(raw []byte, dst any) error {
	if len(raw) == 0 || len(raw) > 64<<10 {
		return errors.New("invalid correction document size")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	var walk func() error
	walk = func() error {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for dec.More() {
				key, err := dec.Token()
				if err != nil {
					return err
				}
				s, ok := key.(string)
				if !ok || seen[s] {
					return errors.New("duplicate correction JSON field")
				}
				seen[s] = true
				if err = walk(); err != nil {
					return err
				}
			}
		case '[':
			for dec.More() {
				if err = walk(); err != nil {
					return err
				}
			}
		default:
			return errors.New("invalid correction JSON delimiter")
		}
		_, err = dec.Token()
		return err
	}
	if err := walk(); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return errors.New("trailing correction JSON")
	}
	var shape func(json.RawMessage, reflect.Type) error
	shape = func(b json.RawMessage, t reflect.Type) error {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() == reflect.Struct {
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(b, &obj); err != nil || obj == nil {
				return errors.New("correction JSON requires object")
			}
			allowed := map[string]bool{}
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				tag := strings.Split(f.Tag.Get("json"), ",")
				name := tag[0]
				if name == "" {
					name = f.Name
				}
				if name == "-" {
					continue
				}
				allowed[name] = true
				child, ok := obj[name]
				optional := strings.Contains(f.Tag.Get("json"), "omitempty")
				if !ok {
					if optional {
						continue
					}
					return fmt.Errorf("missing correction JSON field %s", name)
				}
				if bytes.Equal(bytes.TrimSpace(child), []byte("null")) {
					return fmt.Errorf("null correction JSON field %s", name)
				}
				if err := shape(child, f.Type); err != nil {
					return err
				}
			}
			for key := range obj {
				if !allowed[key] {
					return fmt.Errorf("unknown correction JSON field %s", key)
				}
			}
		} else if t.Kind() == reflect.Slice {
			var list []json.RawMessage
			if err := json.Unmarshal(b, &list); err != nil {
				return err
			}
			for _, child := range list {
				if err := shape(child, t.Elem()); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := shape(raw, reflect.TypeOf(dst).Elem()); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func listActualCorrections(ctx context.Context, root string) ([]gitx.RecoveryEntry, error) {
	store, err := gitx.NewRecoveryStoreFor(ctx, root, "actual-correction")
	if err != nil {
		return nil, err
	}
	return store.List()
}
func readActualCorrection(entry gitx.RecoveryEntry) (actualCorrectionDocument, error) {
	var d actualCorrectionDocument
	if err := strictCorrectionJSON(entry.Data.Document, &d); err != nil {
		return d, err
	}
	if err := d.validate(); err != nil {
		return d, err
	}
	if entry.Token != d.State.Spec.Token {
		return d, errors.New("correction ref token mismatch")
	}
	return d, nil
}

// actualCorrectionMainMirror reads the one authoritative details location,
// sharing the active/legacy/archive directory model with other issue verbs.
func actualCorrectionMainMirror(main *gitx.TrunkView, issuesDir, cardPath string) (string, []byte, error) {
	var mirrorPath string
	var content []byte
	seen := map[string]bool{}
	for _, dir := range issue.IDDirs(issuesDir, envArchiveDirs().History) {
		p := path.Join(filepath.ToSlash(dir), path.Base(cardPath))
		if seen[p] {
			continue
		}
		seen[p] = true
		exists, err := main.Exists(p)
		if err != nil {
			return "", nil, err
		}
		if !exists {
			continue
		}
		if mirrorPath != "" {
			return "", nil, errors.New("ambiguous main details location")
		}
		mirrorPath = p
		content, err = main.Read(p)
		if err != nil {
			return "", nil, err
		}
	}
	if mirrorPath == "" {
		return "", nil, errors.New("completed issue has no main details mirror to correct")
	}
	return mirrorPath, content, nil
}

// A completed repeat has no private recovery ref. Its remote card, effective
// calibration operation and main projection must independently prove the same
// operation before it can return without measuring or producing another row.
func actualCorrectionCompleteRepeat(env *trackerEnv, card tracker.Record, completion issue.Completion, issuesDir, ledger, ref, reason string, exclude bool, selectors ...string) (issue.ActualCorrection, bool, error) {
	record, exists, err := issue.CardActualCorrection(card.Raw)
	if err != nil {
		return record, false, err
	}
	intent := record.IntentReason
	if intent == "" {
		intent = record.Reason
	}
	actual, _ := issue.GetField(card.Card.Frontmatter, "actual_hours")
	if !exists || intent != reason || record.Exclude != exclude || record.NewActual != actual || record.CompletionToken != completion.Token || record.Repository != completion.Repository || record.EvidenceCommit != completion.EvidenceCommit || record.ReviewedHEAD != completion.ReviewedHEAD {
		return record, false, nil
	}
	refuse := func(cause error) (issue.ActualCorrection, bool, error) {
		return record, false, fmt.Errorf("correction %s is not proved complete: %w; restore its calibration/mirror evidence or original recovery receipt, then run sdlc issue recovery reconcile --issue %s", record.Token, cause, issue.CLIRef(card.ID))
	}
	raw, err := os.ReadFile(ledger)
	if err != nil {
		return refuse(err)
	}
	rows, err := estimate.ParseRowsChecked(string(raw))
	if err != nil {
		return refuse(err)
	}
	effective, err := estimate.ResolveCorrections(rows)
	if err != nil {
		return refuse(err)
	}
	disposition := "replace"
	if exclude {
		disposition = "exclude"
	}
	var owned *estimate.LedgerRow
	for _, row := range rows {
		if row.Correction == nil || row.Correction.Operation != record.Token {
			continue
		}
		c := row.Correction
		if row.Issue != ref || c.Generation != completion.Token || c.InputDigest != record.MeasurementDigest || c.Disposition != disposition {
			return refuse(errors.New("calibration operation binding mismatch"))
		}
		if exclude {
			if row.Actual != 0 {
				return refuse(errors.New("exclusion correction contains numeric sample"))
			}
		} else {
			n, err := strconv.ParseFloat(record.NewActual, 64)
			if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(row.Actual-n) > 0.000001 {
				return refuse(errors.New("calibration actual does not match recorded correction"))
			}
		}
		copy := row
		owned = &copy
	}
	if owned == nil {
		return refuse(errors.New("calibration operation evidence missing"))
	}
	if len(selectors) > 0 && selectors[0] != "" && !slices.Contains(owned.Correction.Targets, selectors[0]) {
		return record, false, fmt.Errorf("target row %s conflicts with completed correction %s; retry with --target-row %s", selectors[0], record.Token, owned.Correction.Targets[0])
	}

	if exclude {
		// Exclusions deliberately produce no effective numeric row. Validate
		// resolution above, then prove no later correction superseded this row.
		for _, row := range rows {
			if row.Correction != nil {
				for _, target := range row.Correction.Targets {
					if target == owned.RawID {
						return refuse(errors.New("calibration operation was superseded"))
					}
				}
			}
		}
	} else {
		found := false
		for _, row := range estimate.NewestPerIssue(effective) {
			if row.RawID == owned.RawID {
				found = true
			}
		}
		if !found {
			return refuse(errors.New("calibration operation is no longer effective"))
		}
	}
	main, err := env.main.Snapshot()
	if err != nil {
		return refuse(err)
	}
	_, mirror, err := actualCorrectionMainMirror(main, issuesDir, card.Path)
	if err != nil {
		return refuse(err)
	}
	baseline, err := issue.MirrorBaselineOID(mirror)
	if err != nil {
		return refuse(err)
	}
	if baseline != card.BlobOID {
		return refuse(errors.New("main mirror does not name the corrected card generation"))
	}
	if _, err := issue.RefreshMirror(mirror, card.Raw, card.Raw); err != nil {
		return refuse(err)
	}
	latest, err := env.repo.Snapshot()
	if err != nil {
		return refuse(err)
	}
	current, err := latest.Require(card.ID)
	if err != nil {
		return refuse(err)
	}
	if current.BlobOID != card.BlobOID {
		return refuse(errors.New("card generation changed while confirming completion"))
	}
	return record, true, nil
}

func runActualCorrection(ctx context.Context, stdout, stderr io.Writer, issuesDir string, id int, reason, brainDir string, dryRun, exclude bool, selectors ...string) error {
	targetRow := ""
	if len(selectors) > 0 {
		targetRow = selectors[0]
	}
	if targetRow != "" && !validActualTargetRow(targetRow) {
		return errors.New("--target-row must be a full lowercase SHA256 digest; copy an eligible row ID from the ambiguity diagnostic")
	}

	intentReason := reason
	if id <= 0 || strings.TrimSpace(reason) == "" || len(reason) > 4<<10 {
		return errors.New("--issue and bounded --reason are required")
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	store, err := gitx.NewRecoveryStoreFor(ctx, env.root, "actual-correction")
	if err != nil {
		return err
	}
	entries, err := store.List()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		loaded, err := store.Load(entry.Token)
		if err != nil {
			return err
		}
		d, err := readActualCorrection(loaded)
		if err != nil {
			return err
		}
		if d.State.Spec.IssueID == fmt.Sprintf("%06d", id) {
			if targetRow != "" && !slices.Contains(d.State.Spec.TargetHashes, targetRow) {
				return fmt.Errorf("target row %s conflicts with pending correction %s; resume its pinned target with sdlc issue recovery reconcile --issue %d", targetRow, d.State.Spec.Token, id)
			}
			if dryRun {
				fmt.Fprintf(stdout, "Pending correction %s: %s -> %s (%s)\n", entry.Token, d.State.Spec.PreviousActual, d.State.Spec.NewActual, d.State.Phase)
				return nil
			}
			return resumeActualCorrection(ctx, stdout, stderr, env.root, entry.Token)
		}
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, err := snap.Require(fmt.Sprintf("%06d", id))
	if err != nil {
		return err
	}
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	if status != "codecomplete" && status != "done" {
		return errors.New("actual correction requires codecomplete or done issue")
	}
	completion, ok, err := issue.CardCompletion(card.Raw)
	if err != nil {
		return err
	}
	if !ok || completion.Repository != env.target.Repository {
		return errors.New("actual correction requires matching completed close generation")
	}
	old, _ := issue.GetField(card.Card.Frontmatter, "actual_hours")
	date, err := env.git("show", "-s", "--format=%cs", completion.EvidenceCommit)
	if err != nil {
		return err
	}
	trailers, err := env.git("show", "-s", "--format=%B", completion.EvidenceCommit)
	if err != nil {
		return err
	}
	if !strings.Contains("\n"+trailers+"\n", "\nClose-Actual: "+old+"\n") {
		prior, exists, e := issue.CardActualCorrection(card.Raw)
		if e != nil {
			return e
		}
		if !exists || prior.CompletionToken != completion.Token || prior.NewActual != old {
			return errors.New("completed evidence does not prove recorded actual; repair generation evidence first")
		}
	}
	identity, err := resolveWorkspace(env.root)
	if err != nil {
		return err
	}
	brainAbs, err := filepath.Abs(brainDir)
	if err != nil {
		return err
	}
	ledger := calibrationLedgerPath(brainAbs)
	ledger, err = filepath.Abs(ledger)
	if err != nil {
		return err
	}
	repeated, complete, err := actualCorrectionCompleteRepeat(env, card, completion, issuesDir, ledger, identity.Repo+"#"+strconv.Itoa(id), reason, exclude, targetRow)
	if err != nil {
		return err
	}
	if complete {
		fmt.Fprintf(stdout, "Correction %s already complete: #%s actual %s\n", repeated.Token, card.ID, repeated.NewActual)
		return nil
	}
	raw, err := os.ReadFile(ledger)
	if err != nil {
		return fmt.Errorf("calibration targets: %w", err)
	}
	target, err := actualCorrectionTargetForGeneration(string(raw), identity.Repo+"#"+strconv.Itoa(id), old, date, completion.Token, targetRow)
	if err != nil {
		var ambiguous *actualCorrectionAmbiguity
		if errors.As(err, &ambiguous) {
			var detail strings.Builder
			fmt.Fprintf(&detail, "%s in %s; completion %s. Select the row whose provenance belongs to this close (never select by ordering):", err, ledger, completion.Token)
			for _, row := range ambiguous.Candidates {
				retry := actualCorrectionRetryCommand(id, reason, brainDir, row.RawID)
				if exclude {
					retry += " --exclude"
				}
				fmt.Fprintf(&detail, "\n  %s (date %s, model %q, mode %q, flow %q, review %q, churn %d/%d/%d/%d, boundary rounds %d)\n  %s --dry-run", row.RawID, row.Date, row.Model, row.Mode, row.FlowKind, row.ReviewRecipe, row.ChurnProd, row.ChurnTest, row.ChurnAtlas, row.ChurnWorkshop, row.BoundaryRounds, retry)
			}
			return errors.New(detail.String())
		}
		return fmt.Errorf("calibration target in %s; completion %s: %w", ledger, completion.Token, err)
	}
	measurementRefusal := func(cause error) error {
		return fmt.Errorf("%w; if this sample is proven invalid, exclude it without inventing hours: %s --exclude", cause, actualCorrectionRetryCommand(id, intentReason, brainDir, target.RawID))
	}
	measured := measureActualCorrection(ctx, env.root, brainAbs, strconv.Itoa(id), "", "")
	var inputs gatestate.MeasurementInputs
	nextActual := "N/A"
	if exclude {
		stamp, err := env.git("show", "-s", "--format=%cI", completion.EvidenceCommit)
		if err != nil {
			return err
		}
		inputs = gatestate.MeasurementInputs{Version: gatestate.MeasurementInputsVersion, Method: gatestate.MeasurementMethodLegacyV1, Repository: env.target.Repository, CodeHead: completion.ReviewedHEAD, Since: stamp, Until: stamp, GapMinutes: 30}
		fmt.Fprintf(stderr, "Exclusion: pin-only generation evidence; no numeric adoption. Measurement status %d: %s\n", measured.Status, measured.Detail)
		diagnostic := measured.Detail
		if len(diagnostic) > 256 {
			diagnostic = diagnostic[:256]
		}
		if measured.Status != actualMeasured {
			reason += fmt.Sprintf("; excluded without numeric measurement (status %d: %s)", measured.Status, diagnostic)
		}
		if len(reason) > 4<<10 {
			return errors.New("reason plus exclusion evidence exceeds 4 KiB")
		}
	} else {
		if measured.Status != actualMeasured || measured.Inputs == nil {
			return measurementRefusal(fmt.Errorf("correction requires trustworthy measured input: status %d, %s, %s", measured.Status, measured.Detail, strings.Join(measured.Warnings, "; ")))
		}
		inputs = *measured.Inputs
		if math.IsNaN(measured.Hours) || math.IsInf(measured.Hours, 0) || measured.Hours < 0 {
			return measurementRefusal(errors.New("correction measured invalid actual"))
		}
		nextActual = strconv.FormatFloat(measured.Hours, 'f', 2, 64)
	}
	if err := inputs.Validate(); err != nil {
		return measurementRefusal(err)
	}
	if inputs.Repository != env.target.Repository || inputs.CodeHead != completion.ReviewedHEAD {
		return measurementRefusal(errors.New("measurement is not pinned to the completed close generation"))
	}
	inputJSON, err := json.Marshal(inputs)
	if err != nil {
		return err
	}
	token := operationToken("actual-correction")
	evidence := issue.ActualCorrection{Version: 1, Token: token, CompletionToken: completion.Token, Repository: completion.Repository, EvidenceCommit: completion.EvidenceCommit, ReviewedHEAD: completion.ReviewedHEAD, PreviousActual: old, NewActual: nextActual, Reason: reason, IntentReason: intentReason, MeasurementDigest: correctionDigest(inputJSON), Inputs: string(inputJSON), Exclude: exclude}
	candidate, err := issue.SetCardField(card.Raw, "actual_hours", nextActual)
	if err != nil {
		return err
	}
	candidate, err = issue.SetCardActualCorrection(candidate, evidence)
	if err != nil {
		return err
	}
	candidateOID, err := issue.CardBlobOID(candidate, env.format)
	if err != nil {
		return err
	}
	main, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	mirrorPath, original, err := actualCorrectionMainMirror(main, issuesDir, card.Path)
	if err != nil {
		return err
	}
	baseline, err := issue.MirrorBaselineOID(original)
	if err != nil {
		return err
	}
	baselineCard, err := env.repo.ReadCardBlob(baseline)
	if err != nil {
		return err
	}
	mirror, err := issue.RefreshMirror(original, baselineCard, candidate)
	if err != nil {
		return err
	}
	row := target
	row.RawID = ""
	row.Actual = measured.Hours
	if !exclude && inputs.Method == gatestate.MeasurementMethodEngagementV1 {
		row.WindowTrusted = true
	}
	row.Date = time.Now().Format("2006-01-02")
	disposition := "replace"
	if exclude {
		disposition = "exclude"
		row.Actual = 0
		row.BuildHours = 0
		row.ConvergeHours = 0
		row.SplitSource = ""
	} else if measured.Split != nil {
		row.BuildHours = measured.Split.Build
		row.ConvergeHours = measured.Split.Converge
		row.SplitSource = measured.Split.Source
	} else {
		row.BuildHours = 0
		row.ConvergeHours = 0
		row.SplitSource = ""
	}
	row.Correction = &estimate.CorrectionRecord{Version: 1, Operation: token, Generation: completion.Token, Disposition: disposition, Targets: []string{target.RawID}, InputDigest: evidence.MeasurementDigest}
	line := estimate.FormatRow(row)
	spec := actualcorrection.Spec{Version: actualcorrection.Version, Token: token, Repository: env.target.Repository, IssueID: card.ID, CompletionToken: completion.Token, EvidenceCommit: completion.EvidenceCommit, ReviewedHEAD: completion.ReviewedHEAD, ExpectedCardBlob: card.BlobOID, CandidateCardBlob: candidateOID, MeasurementDigest: evidence.MeasurementDigest, Reason: reason, CalibrationDestination: ledger, TargetHashes: row.Correction.Targets, PreviousActual: old, NewActual: nextActual, Exclude: exclude, CalibrationRowDigest: correctionDigest([]byte(line)), MirrorProjectionDigest: correctionDigest(mirror)}
	state := actualcorrection.State{Spec: spec, Phase: actualcorrection.Prepared}
	if err := actualcorrection.Validate(state); err != nil {
		return err
	}
	if !exclude {
		printActualMeasurement(stderr, measured)
	}
	fmt.Fprintf(stdout, "#%s actual: %s -> %s; close %s; reviewed %s; input %s; target %s\n", card.ID, old, nextActual, completion.Token, completion.ReviewedHEAD, spec.MeasurementDigest, target.RawID)
	if dryRun {
		return nil
	}
	if _, err = gitx.WriteBlob(ctx, env.root, candidate); err != nil {
		return err
	}
	originalOID, err := gitx.WriteBlob(ctx, env.root, original)
	if err != nil {
		return err
	}
	mirrorOID, err := gitx.WriteBlob(ctx, env.root, mirror)
	if err != nil {
		return err
	}
	d := actualCorrectionDocument{Version: 1, State: state, Inputs: inputs, CardPath: card.Path, MirrorPath: mirrorPath, OriginalMirrorBlob: originalOID, MirrorBlob: mirrorOID, Row: line}
	if err = d.validate(); err != nil {
		return err
	}
	document, err := json.Marshal(d)
	if err != nil {
		return err
	}
	if len(document) > 64<<10 {
		return errors.New("correction operation exceeds 64 KiB")
	}
	if _, err = store.Save(token, "", gitx.RecoveryData{Document: document, Objects: []string{card.BlobOID, candidateOID, originalOID, mirrorOID, completion.EvidenceCommit, completion.ReviewedHEAD}}); err != nil {
		return err
	}
	return resumeActualCorrection(ctx, stdout, stderr, env.root, token)
}

func resumeActualCorrection(ctx context.Context, stdout, stderr io.Writer, root, token string) error {
	env, err := openTrackerAt(ctx, root)
	if err != nil {
		return err
	}
	store, err := gitx.NewRecoveryStoreFor(ctx, env.root, "actual-correction")
	if err != nil {
		return err
	}
	entry, err := store.Load(token)
	if err != nil {
		return err
	}
	d, err := readActualCorrection(entry)
	if err != nil {
		return err
	}
	if d.State.Spec.Repository != env.target.Repository {
		return errors.New("correction recovery repository mismatch")
	}
	save := func() error {
		if err := d.validate(); err != nil {
			return err
		}
		raw, err := json.Marshal(d)
		if err != nil {
			return err
		}
		if len(raw) > 64<<10 {
			return errors.New("correction operation exceeds 64 KiB")
		}
		s := d.State.Spec
		objects := []string{s.ExpectedCardBlob, s.CandidateCardBlob, s.EvidenceCommit, s.ReviewedHEAD, d.OriginalMirrorBlob, d.MirrorBlob}
		for _, c := range []gitx.Candidate{d.TrackerCandidate, d.MainCandidate} {
			if c.OID != "" {
				objects = append(objects, c.OID, c.Base)
			}
		}
		next, err := store.Save(token, entry.OID, gitx.RecoveryData{Document: raw, Objects: objects})
		if err != nil {
			return err
		}
		entry.OID = next
		return nil
	}
	event := func(kind actualcorrection.EventKind) actualcorrection.Event {
		s := d.State
		return actualcorrection.Event{Kind: kind, Spec: s.Spec, TrackerCommit: s.TrackerCommit, CandidateCardBlob: s.Spec.CandidateCardBlob, CalibrationRowDigest: s.Spec.CalibrationRowDigest, MirrorProjectionDigest: s.Spec.MirrorProjectionDigest}
	}
	advance := func(e actualcorrection.Event) error {
		next, _, err := actualcorrection.Transition(d.State, e)
		if err != nil {
			return err
		}
		d.State = next
		return save()
	}
	current := func() (tracker.Record, error) {
		snap, err := env.repo.Snapshot()
		if err != nil {
			return tracker.Record{}, err
		}
		card, err := snap.Require(d.State.Spec.IssueID)
		if err != nil {
			return card, err
		}
		s := d.State.Spec
		c, ok, err := issue.CardCompletion(card.Raw)
		if err != nil {
			return card, err
		}
		if !ok || c.Token != s.CompletionToken || c.Repository != s.Repository || c.EvidenceCommit != s.EvidenceCommit || c.ReviewedHEAD != s.ReviewedHEAD || card.Path != d.CardPath {
			return card, errors.New("completed close generation changed; correction progress retained")
		}
		wanted := s.CandidateCardBlob
		if (d.State.Phase == actualcorrection.Prepared || d.State.Phase == actualcorrection.UncertainTracker) && card.BlobOID == s.ExpectedCardBlob {
			return card, nil
		}
		if card.BlobOID != wanted {
			return card, errors.New("card generation changed; correction progress retained")
		}
		record, exists, err := issue.CardActualCorrection(card.Raw)
		if err != nil {
			return card, err
		}
		if !exists || record.Token != s.Token || record.MeasurementDigest != s.MeasurementDigest {
			return card, errors.New("remote card belongs to another correction operation")
		}
		return card, nil
	}
	confirmTracker := func() error {
		s := d.State.Spec
		if d.TrackerCandidate.OID == "" {
			return errors.New("uncertain correction has no pinned tracker candidate")
		}
		blob, err := env.repo.CardBlobAt(d.TrackerCandidate.OID, d.CardPath)
		if err != nil {
			return err
		}
		if blob != s.CandidateCardBlob {
			return errors.New("tracker candidate card proof mismatch")
		}
		msg, err := env.git("show", "-s", "--format=%B", d.TrackerCandidate.OID)
		if err != nil {
			return err
		}
		if !strings.Contains("\n"+msg+"\n", "\nTracker-Operation: "+s.Token+"\n") {
			return errors.New("tracker candidate operation proof mismatch")
		}
		card, err := current()
		if err != nil {
			return err
		}
		if card.BlobOID != s.CandidateCardBlob {
			return errors.New("tracker candidate not current; correction remains pending")
		}
		e := event(actualcorrection.TrackerConfirmed)
		e.TrackerCommit = d.TrackerCandidate.OID
		return advance(e)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.State.Phase == actualcorrection.Complete {
			if err := store.Delete(token, entry.OID); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "Correction %s complete: #%s actual %s\n", token, d.State.Spec.IssueID, d.State.Spec.NewActual)
			return nil
		}
		card, err := current()
		if err != nil {
			return err
		}
		_, effect, err := actualcorrection.Transition(d.State, actualcorrection.Event{Kind: actualcorrection.Repeat})
		if err != nil {
			return err
		}
		switch effect {
		case actualcorrection.PublishTracker:
			next, err := gitx.ReadBlob(ctx, env.root, d.State.Spec.CandidateCardBlob)
			if err != nil {
				return err
			}
			err = env.repo.UpdateCardWithTrailers(card, next, token, []string{"Actual-Correction: " + token}, func(base, candidate string) error {
				d.TrackerCandidate = gitx.Candidate{Base: base, OID: candidate}
				// Persist uncertainty BEFORE the write. A crash cannot turn an unknown
				// publication into a fresh measured correction or another writer's value.
				return advance(event(actualcorrection.TrackerUncertain))
			})
			invalidateIssueRecords(ctx)
			if err != nil {
				return fmt.Errorf("correction tracker pending (%s): %w", token, err)
			}
			if err = confirmTracker(); err != nil {
				return err
			}
		case actualcorrection.ProbeTracker:
			if d.TrackerCandidate.OID == "" {
				return errors.New("uncertain correction missing publication candidate")
			}
			outcome, _, err := env.repo.Probe(d.TrackerCandidate)
			if err != nil {
				return err
			}
			switch outcome {
			case gitx.ProbeReachable:
				if err := confirmTracker(); err != nil {
					return err
				}
			case gitx.ProbeAbsentSame:
				pushed, err := env.repo.Push(d.TrackerCandidate)
				if err != nil {
					return err
				}
				if pushed != gitx.PushAccepted {
					return errors.New("correction tracker publication remains pending")
				}
				if err := confirmTracker(); err != nil {
					return err
				}
			default:
				return errors.New("correction tracker publication not confirmed; generation moved or publication remains uncertain")
			}
		case actualcorrection.AppendCalibration:
			s := d.State.Spec
			if err := appendActualCorrectionRow(ctx, stderr, s.CalibrationDestination, d.Row, s.Token, s.CalibrationRowDigest); err != nil {
				return err
			}
			// Re-observe card generation after the separate ledger effect, preserving
			// the pending phase if a concurrent writer superseded it meanwhile.
			if _, err := current(); err != nil {
				return err
			}
			if err := advance(event(actualcorrection.CalibrationConfirmed)); err != nil {
				return err
			}
		case actualcorrection.PublishMirror:
			mirror, err := gitx.ReadBlob(ctx, env.root, d.MirrorBlob)
			if err != nil {
				return err
			}
			if correctionDigest(mirror) != d.State.Spec.MirrorProjectionDigest {
				return errors.New("pinned mirror projection changed")
			}
			// A persisted candidate after interruption is reconciled before preparing
			// anything new. Reachability plus exact token proves our own publication.
			if d.MainCandidate.OID != "" {
				outcome, _, err := env.main.ProbeCandidate(d.MainCandidate)
				if err != nil {
					return err
				}
				switch outcome {
				case gitx.ProbeReachable:
				case gitx.ProbeAbsentSame:
					pushed, err := env.main.PushCandidate(d.MainCandidate)
					if err != nil {
						return err
					}
					if pushed != gitx.PushAccepted {
						return errors.New("correction mirror publication remains pending")
					}
				default:
					return errors.New("correction mirror candidate not confirmed; main generation moved")
				}
			} else {
				original, err := gitx.ReadBlob(ctx, env.root, d.OriginalMirrorBlob)
				if err != nil {
					return err
				}
				err = env.main.UpdateManyPrepared("#"+issue.CLIRef(d.State.Spec.IssueID)+": actual: refresh corrected card mirror\n\nActual-Correction: "+token, func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
					if _, err := current(); err != nil {
						return gitx.TrunkWrite{}, err
					}
					main, err := view.Read(d.MirrorPath)
					if err != nil {
						return gitx.TrunkWrite{}, err
					}
					if !bytes.Equal(main, original) {
						return gitx.TrunkWrite{}, errors.New("main details generation changed; correction progress retained")
					}
					return gitx.TrunkWrite{Write: map[string][]byte{d.MirrorPath: mirror}, ExactBytes: true}, nil
				}, func(base, candidate string) error {
					d.MainCandidate = gitx.Candidate{Base: base, OID: candidate}
					return save()
				})
				if err != nil {
					return fmt.Errorf("correction mirror pending (%s): %w", token, err)
				}
			}
			if d.MainCandidate.OID == "" {
				return errors.New("mirror equality does not prove this correction operation")
			}
			msg, err := env.git("show", "-s", "--format=%B", d.MainCandidate.OID)
			if err != nil {
				return err
			}
			if !strings.Contains("\n"+msg+"\n", "\nActual-Correction: "+token+"\n") {
				return errors.New("mirror candidate operation proof mismatch")
			}
			candidate, err := env.main.ViewOf(d.MainCandidate.OID).Read(d.MirrorPath)
			if err != nil {
				return err
			}
			if !bytes.Equal(candidate, mirror) {
				return errors.New("mirror candidate projection proof mismatch")
			}
			main, err := env.main.Snapshot()
			if err != nil {
				return err
			}
			observed, err := main.Read(d.MirrorPath)
			if err != nil {
				return err
			}
			if !bytes.Equal(observed, mirror) {
				return errors.New("main mirror generation changed after correction publication")
			}
			if _, err := current(); err != nil {
				return err
			}
			if err := advance(event(actualcorrection.MirrorConfirmed)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected correction effect %s", effect)
		}
	}
}
