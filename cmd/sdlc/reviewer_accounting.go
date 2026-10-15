package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
	"github.com/xianxu/ariadne/pkg/frontmatter"
)

// attachReviewerAccounting prepares an accounting-only CAS recorder before the
// repo lock is released. It carries no admission authority and no prompts.
func attachReviewerAccounting(ctx context.Context, opts *judge.DispatchOptions, issueID int, recipe string, round int, head string) error {
	if recipe == "" {
		recipe = "milestone-review"
	}
	if issueID <= 0 {
		return nil
	}
	identity, e := resolveWorkspace(".")
	if e != nil {
		return e
	}
	root := identity.WorktreeRoot
	resting := "main"
	if identity.RestingBranch != nil {
		resting = *identity.RestingBranch
	}
	target, e := gitx.ResolvePublicationTarget(ctx, root, resting)
	if e != nil {
		target.Repository = identity.RepoIdentity
	}
	if head == "" {
		head, e = ownedActualGit(ctx, root, "rev-parse", "HEAD")
		if e != nil {
			return e
		}
	}
	slot := "unknown-slot"
	if home := identity.HomeSlot(); home != "" {
		slot = home
	}
	store, e := reviewusage.NewStore(root)
	if e != nil {
		return e
	}
	if err := cleanupReviewerAccounting(ctx, root, target.Remote, store); err != nil {
		return err
	}
	opts.UsageIntent = &reviewusage.Intent{Repository: target.Repository, IssueID: fmt.Sprintf("%06d", issueID), Recipe: recipe, Round: round, ReviewedHEAD: head, Slot: slot, Root: root}
	opts.UsageRecorder = store.Record
	if opts.Evidence == nil {
		opts.Evidence = &judge.DispatchEvidence{}
	}
	return nil
}

func persistEstimateUsage(f *changeCodeFlags, name string, records []reviewusage.Record) error {
	path, e := findIssueFileByName(f.IssuesDir, name)
	if e != nil {
		return e
	}
	file := filepath.Base(path)
	ledger, e := readPlanGateLedger(f.PlansDir, file, f.Issue)
	if e != nil {
		return e
	}
	ledger.ReviewerRuns = append(ledger.ReviewerRuns, records...)
	if e = writePlanGateLedger(f.PlansDir, file, ledger, repoIdentity()); e != nil {
		return e
	}
	return f.review.refreshOwnLedger()
}

// A private receipt is collected only after its entire exact typed record is
// reachable from a committed ledger/sidecar. Failed, stale and interrupted
// reviews have no such proof and remain retained. The namespace is bounded.
func cleanupReviewerAccounting(ctx context.Context, root, remote string, store *reviewusage.Store) error {
	entries, e := store.List(ctx)
	if e != nil {
		return e
	}
	if len(entries) == 0 {
		return nil
	}
	wanted := map[string]bool{}
	for _, entry := range entries {
		wanted[entry.Record.Intent.IssueID] = true
	}
	proven := map[reviewusage.Record]bool{}
	bytesRead := 0
	for _, ref := range []string{"HEAD", "refs/remotes/" + remote + "/main"} {
		pin, e := ownedActualGit(ctx, root, "rev-parse", "--verify", ref+"^{commit}")
		if e != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		paths, e := ownedActualGit(ctx, root, "ls-tree", "-r", "--name-only", pin, "--", "workshop/plans", "workshop/history/plans")
		if e != nil {
			continue
		}
		for _, path := range strings.Split(paths, "\n") {
			name := filepath.Base(path)
			if len(name) < 7 || !wanted[name[:6]] || !strings.HasSuffix(name, "-gate.md") && !strings.HasSuffix(name, "-review.md") {
				continue
			}
			raw, e := ownedActualGit(ctx, root, "show", pin+":"+path)
			if e != nil {
				continue
			}
			bytesRead += len(raw)
			if bytesRead > 16<<20 {
				return nil
			}
			for _, record := range reviewerRecordsInArtifact(raw) {
				proven[record] = true
			}
		}
	}
	for _, entry := range entries {
		if proven[entry.Record] {
			if e := store.Delete(ctx, entry.Record.Intent.ID, entry.OID); e != nil {
				return fmt.Errorf("collect committed reviewer accounting: %w", e)
			}
		}
	}
	return nil
}

// decodeReviewerAccountingArtifact is shared by reporting and reachability
// cleanup. Invalid evidence remains available for identity quarantine, but never
// becomes accepted accounting or authorizes deletion of a private receipt.
func decodeReviewerAccountingArtifact(raw string) ([]reviewusage.Record, usage.Coverage) {
	var records []reviewusage.Record
	var coverage usage.Coverage
	retain := func(rr []reviewusage.Record, malformed bool) {
		for _, record := range rr {
			if malformed {
				// Preserve recoverable identity while ensuring this partial parse
				// cannot acquire attribution authority.
				record.Version = 0
			}
			if err := reviewusage.Validate(record); err != nil {
				coverage.Add("unavailable_reviewer_bindings", err.Error())
			}
			records = append(records, record)
		}
	}
	// Boundary review prose has no frontmatter; its typed evidence is in the
	// comment below. A present but unreadable ledger must be reported.
	if strings.HasPrefix(strings.TrimSpace(raw), "---") {
		ledger, err := gatestate.ParseSidecar(raw)
		malformed := err != nil
		if malformed {
			coverage.Add("unavailable_reviewer_bindings", "malformed reviewer ledger")
			// A semantic ledger error (for example in actual_inputs) can still
			// leave independently typed reviewer identity recoverable.
			if fm, _, splitErr := frontmatter.Split(raw); splitErr == nil {
				_ = yaml.Unmarshal([]byte(fm), &ledger)
			}
		}
		retain(ledger.ReviewerRuns, malformed)
		for _, round := range ledger.Rounds {
			if len(round.ReviewerRuns) == 0 && !round.NoCap {
				coverage.Add("legacy_reviewer_binding_gaps", "")
			}
			retain(round.ReviewerRuns, malformed)
		}
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "<!-- sdlc-review-usage") {
			continue
		}
		malformed := !strings.HasPrefix(line, "<!-- sdlc-review-usage ") || !strings.HasSuffix(line, " -->")
		if malformed {
			coverage.Add("unavailable_reviewer_bindings", "malformed reviewer accounting comment")
		}
		var encoded []json.RawMessage
		payload := strings.TrimSuffix(strings.TrimPrefix(line, "<!-- sdlc-review-usage "), " -->")
		if err := json.Unmarshal([]byte(payload), &encoded); err != nil || encoded == nil {
			coverage.Add("unavailable_reviewer_bindings", "malformed reviewer accounting JSON")
			continue
		}
		for _, item := range encoded {
			record, err := reviewusage.ParseAccountingEvidence(item)
			retain([]reviewusage.Record{record}, malformed || err != nil)
		}
	}
	return records, coverage
}

func reviewerRecordsInArtifact(raw string) []reviewusage.Record {
	records, _ := decodeReviewerAccountingArtifact(raw)
	valid := records[:0]
	for _, record := range records {
		if reviewusage.Validate(record) == nil {
			valid = append(valid, record)
		}
	}
	return valid
}
