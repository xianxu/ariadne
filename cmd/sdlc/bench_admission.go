package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type benchCloseProof struct {
	context                      *benchPublicContext
	record                       benchSessionFacts
	run                          bench.Run
	env                          *trackerEnv
	card                         tracker.Record
	binding                      issue.Completion
	detail, reviewPath, gatePath string
	ledger                       gatestate.Ledger
	review                       observe.Review
}

// readBenchCloseProof pins the existing tracker, gate and observation owners at
// the immutable close evidence. It never guesses completion from pane prose or
// substitutes current metadata HEAD for the candidate the reviewer read.
func readBenchCloseProof(ctx context.Context, s *benchSession) (benchCloseProof, error) {
	var p benchCloseProof
	if ctx == nil || s == nil {
		return p, errors.New("bench admission needs context and store")
	}
	if err := ctx.Err(); err != nil {
		return p, err
	}
	record, run, err := s.read()
	if err != nil {
		return p, err
	}
	p.record, p.run = record, run
	root := filepath.Join(s.root, "public", "work")
	c, err := loadBenchContext(ctx, root)
	if err != nil {
		return p, err
	}
	if c == nil || c.RunID != record.ID || c.WorkRoot != root || c.HubRoot != filepath.Join(s.root, "public", "hub") || c.OriginPath != filepath.Join(s.root, "public", "origin.git") {
		return p, errors.New("bench public roots do not match observer run")
	}
	p.context = c
	env, err := openTrackerAt(ctx, root)
	if err != nil {
		return p, err
	}
	p.env = env
	if env.target.Repository != c.OriginIdentity || env.target.Remote != "origin" {
		return p, errors.New("bench completion is not on exact private backend")
	}
	replacements, err := env.git("for-each-ref", "--format=%(refname)", "refs/replace/")
	if err != nil {
		return p, err
	}
	if replacements != "" {
		return p, errors.New("bench admission refuses replacement object refs")
	}
	snapshot, err := env.repo.Snapshot()
	if err != nil {
		return p, err
	}
	p.card, err = snapshot.Require(record.Task.ID)
	if err != nil {
		return p, err
	}
	fm, _, err := issue.Parse(string(p.card.Raw))
	if err != nil {
		return p, err
	}
	if status, _ := issue.GetField(fm, "status"); status != "codecomplete" {
		return p, errors.New("bench requires current codecomplete card, before any landing")
	}
	p.binding, _, err = issue.CardCompletion(p.card.Raw)
	if err != nil {
		return p, err
	}
	if p.binding.Token == "" || p.binding.Repository != c.OriginIdentity || !benchCandidatePattern.MatchString(p.binding.ReviewedHEAD) || !benchCandidatePattern.MatchString(p.binding.EvidenceCommit) || p.binding.LandedCommit != "" {
		return p, errors.New("bench card lacks exact private unlanded completion binding")
	}
	d := vocab.Issue().Discovery()
	p.detail = path.Join(d.Home, path.Base(p.card.Path))
	p.reviewPath = filepath.ToSlash(sidecarPath(d.Plans, path.Base(p.detail), ""))
	p.gatePath = filepath.ToSlash(boundaryGatePath(d.Plans, path.Base(p.detail)))
	if err := verifyCloseEvidenceChain(env, p.card, p.binding, p.binding.ReviewedHEAD, env.head, p.detail, d.Plans); err != nil {
		return p, err
	}
	gate, err := benchAdmissionBlob(ctx, root, p.binding.EvidenceCommit, p.gatePath)
	if err != nil {
		return p, err
	}
	p.ledger, err = gatestate.ParseSidecar(string(gate))
	if err != nil {
		return p, err
	}
	id, err := strconv.Atoi(record.Task.ID)
	if err != nil {
		return p, err
	}
	if p.ledger.Gate != "boundary-review" || p.ledger.IDPrefix != "BR" || p.ledger.IssueNum != id {
		return p, errors.New("bench review ledger belongs to another gate/issue")
	}
	reviewed, ok := gatestate.LatestFinalizedAt(p.ledger, "")
	if !ok || reviewed != p.binding.ReviewedHEAD {
		return p, errors.New("bench completion lacks same-candidate finalized close review")
	}
	if err := verifyBenchCloseGate(p.ledger); err != nil {
		return p, err
	}
	details, err := benchAdmissionBlob(ctx, root, p.binding.EvidenceCommit, p.detail)
	if err != nil {
		return p, err
	}
	reviewText, err := benchAdmissionBlob(ctx, root, p.binding.EvidenceCommit, p.reviewPath)
	if err != nil {
		return p, err
	}
	var rounds []observe.ReviewRound
	for _, r := range p.ledger.Rounds {
		if r.Boundary == "" && !r.NoCap {
			rounds = append(rounds, observe.ReviewRound{Number: r.N, Head: r.Reviewed})
		}
	}
	observed := observe.Assemble(observe.Inputs{Issue: record.Task.ID, Tracked: true, TrackerRef: snapshot.Ref(), Card: p.card.Raw, CardPath: p.card.Path, CardBlob: p.card.BlobOID, Evidence: observe.Evidence{Source: p.binding.EvidenceCommit + ":" + d.Plans, Details: details, Artifacts: map[string]observe.ArtifactFacts{"close": {Found: true, Sidecar: string(reviewText), Rounds: rounds}}}})
	for _, review := range observed.Checkpoints.Reviews {
		if review.Boundary == "close" {
			p.review = review
		}
	}
	if p.review.State != observe.Present || p.review.Verdict != "SHIP" || p.review.RoundError != "" || !strings.HasSuffix(p.review.Window, ".."+p.binding.ReviewedHEAD) {
		return p, errors.New("bench completion is not bound to an exact SHIP review")
	}
	trailer, err := env.git("show", "-s", "--format=%(trailers:key=Review-Verdict,valueonly,separator=%x00)", p.binding.EvidenceCommit)
	if err != nil {
		return p, err
	}
	if trailer != "SHIP" {
		return p, errors.New("bench evidence does not carry exactly one SHIP verdict")
	}
	return p, ctx.Err()
}

func verifyBenchCloseGate(ledger gatestate.Ledger) error {
	decision := gatestate.DecideScoped(gatestate.FilterBoundary(ledger, ""), openScopeFor(ledger, ""), gatestate.DefaultRoundCap)
	if decision.Block {
		return errors.New("bench close review retains unresolved blocking findings")
	}
	return nil
}

func benchAdmissionBlob(ctx context.Context, root, commit, rel string) ([]byte, error) {
	out, diag, err := gitx.RunGitBoundedContext(ctx, root, []string{"GIT_NO_REPLACE_OBJECTS=1", "GIT_NO_LAZY_FETCH=1"}, benchRecordLimit, "--no-replace-objects", "cat-file", "blob", commit+":"+rel)
	if err != nil {
		return nil, fmt.Errorf("bench close evidence %s: %w: %s", rel, err, diag)
	}
	return out, nil
}

type benchAdmissionSuiteReader func(context.Context, string, string, string) (benchLocalCheckReceipt, error)

func verifyBenchWouldLand(ctx context.Context, s *benchSession, candidate string) (bench.WouldLandFacts, error) {
	return verifyBenchWouldLandWithSuite(ctx, s, candidate, readBenchLocalCheckReceipt)
}

// The only replaceable IO is the costly native suite reader. Production always
// uses its full saved-attempt/flake proof; graph/card/review checks remain real.
func verifyBenchWouldLandWithSuite(ctx context.Context, s *benchSession, candidate string, readSuite benchAdmissionSuiteReader) (bench.WouldLandFacts, error) {
	var facts bench.WouldLandFacts
	if !benchCandidatePattern.MatchString(candidate) {
		return facts, errors.New("bench candidate must be a full lowercase object ID")
	}
	p, err := readBenchCloseProof(ctx, s)
	if err != nil {
		return facts, err
	}
	if candidate != p.binding.ReviewedHEAD {
		return facts, errors.New("TL candidate differs from exact reviewed completion head")
	}
	hub, worker := p.run.Bindings()
	if hub.Root != p.context.HubRoot || worker.Root != p.context.WorkRoot || hub.Validate() != nil || worker.Validate() != nil {
		return facts, errors.New("bench admission lacks exact provisioned runtime roots")
	}
	if p.env.branch == "" || p.env.onRest() {
		return facts, errors.New("bench candidate must remain on its issue branch")
	}
	dirty, err := p.env.statusEntries(p.env.root, "--untracked-files=all")
	if err != nil {
		return facts, err
	}
	if len(dirty) != 0 {
		return facts, errors.New("bench worker tree must be clean before admission")
	}
	remoteHead := func() (string, error) {
		out, diag, err := gitx.RunGitInContext(ctx, p.context.OriginPath, nil, "rev-parse", "--verify", p.env.branchRef()+"^{commit}")
		if err != nil {
			return "", fmt.Errorf("bench private pushed branch: %w: %s", err, diag)
		}
		return strings.TrimSpace(string(out)), nil
	}
	pushed, err := remoteHead()
	if err != nil {
		return facts, err
	}
	if pushed != p.env.head {
		return facts, errors.New("bench close metadata is not pushed to its private branch")
	}
	if readSuite == nil {
		return facts, errors.New("bench suite verifier missing")
	}
	suite, err := readSuite(ctx, p.context.WorkRoot, candidate, "suite")
	if err != nil {
		return facts, err
	}
	if suite.RunID != p.record.ID || suite.WorkRoot != p.context.WorkRoot || suite.OriginIdentity != p.context.OriginIdentity || suite.Head != candidate || suite.Check != "suite" || suite.Record.SHA != candidate || suite.Record.Name != "suite" || !suite.Record.passed() || len(suite.Record.LogSHA256) != 64 || !suiteHex.MatchString(suite.Record.LogSHA256) {
		return facts, errors.New("bench needs exact candidate full-suite pass evidence")
	}
	current, err := p.env.git("rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return facts, err
	}
	latestPushed, err := remoteHead()
	if err != nil {
		return facts, err
	}
	if current != p.env.head || latestPushed != pushed {
		return facts, errors.New("bench code/metadata moved during admission")
	}
	snapshot, err := p.env.repo.Snapshot()
	if err != nil {
		return facts, err
	}
	card, err := snapshot.Require(p.record.Task.ID)
	if err != nil {
		return facts, err
	}
	if card.BlobOID != p.card.BlobOID {
		return facts, errors.New("bench completion card changed during admission")
	}
	finalContext, err := loadBenchContext(ctx, p.context.WorkRoot)
	if err != nil {
		return facts, err
	}
	if finalContext == nil || *finalContext != *p.context {
		return facts, errors.New("bench public context changed during admission")
	}
	receiptID, err := benchLocalCheckReceiptPath(p.context, candidate, "suite")
	if err != nil {
		return facts, err
	}
	// Both actual tips equal the proven evidence/mirror chain above. Normalize
	// their code identity only after that exact owner proof, retaining immutable
	// evidence and review IDs separately instead of calling metadata "reviewed".
	facts = bench.WouldLandFacts{Candidate: candidate, ObservedHead: candidate, PushedHead: candidate, PrivateOrigin: p.context.OriginIdentity, ExpectedOrigin: p.context.OriginIdentity, SuiteHead: candidate, SuiteReceiptID: receiptID, SuiteLogDigest: suite.Record.LogSHA256, SuitePassed: true, FullSuite: true, CompletionHead: p.binding.ReviewedHEAD, ReviewedHead: p.binding.ReviewedHEAD, CompletionEvidence: p.binding.EvidenceCommit, CompletionToken: p.binding.Token, ReviewID: fmt.Sprintf("%s:%s#round-%d", p.binding.EvidenceCommit, p.gatePath, p.review.Round), ReviewVerdict: p.review.Verdict, CompletionBound: true}
	if err = bench.ValidateWouldLand(facts); err != nil {
		return bench.WouldLandFacts{}, err
	}
	return facts, ctx.Err()
}

// loadBenchFocusedInputs reuses the immutable finalized measurement snapshot
// and explicit captured originals. It never discovers a live transcript store.
func loadBenchFocusedInputs(ctx context.Context, s *benchSession, sources []benchCapturedSource) (*ownedActualInputs, error) {
	p, err := readBenchCloseProof(ctx, s)
	if err != nil {
		return nil, err
	}
	return benchFocusedFromProof(ctx, p, sources)
}

func benchFocusedFromProof(ctx context.Context, p benchCloseProof, sources []benchCapturedSource) (*ownedActualInputs, error) {
	saved, err := ownedBoundarySnapshot(p.ledger, "", p.context.OriginIdentity)
	if err != nil || saved == nil {
		return nil, err
	}
	if saved.CodeHead != p.binding.ReviewedHEAD {
		return nil, errors.New("retained bench focused inputs name another reviewed candidate")
	}
	if saved.LegacyPrefix != nil {
		return nil, errors.New("bench focused inputs do not discover legacy source prefixes")
	}
	for _, pin := range nonEmpty(saved.CodeHead, saved.TrackerHead, saved.BranchPoint) {
		if _, err := p.env.git("cat-file", "-e", pin+"^{commit}"); err != nil {
			return nil, fmt.Errorf("retained bench measurement Git pin unavailable: %w", err)
		}
	}
	var found []transcripts.SessionEvidence
	captured := map[[2]string]benchCapturedSource{}
	for _, source := range sources {
		if !source.Complete || source.OriginalFile == "" || source.NativeSession == "" {
			continue
		}
		key := [2]string{source.Runtime.Agent, source.NativeSession}
		if prior, exists := captured[key]; exists && prior.OriginalFile != source.OriginalFile {
			return nil, errors.New("ambiguous explicitly captured measurement source")
		}
		captured[key] = source
	}
	for _, pin := range saved.Sources {
		source, ok := captured[[2]string{pin.Harness, pin.ID}]
		if ok && filepath.Clean(source.Runtime.Root) == filepath.Clean(pin.CWD) && source.Length >= pin.Bytes {
			found = append(found, transcripts.SessionEvidence{Harness: pin.Harness, ID: pin.ID, CWD: pin.CWD, ParentID: pin.ParentID, Spawned: pin.Spawned, File: source.OriginalFile})
		}
	}
	inputs, err := ownedReplaySources(*saved, found)
	if err != nil {
		return nil, err
	}
	inputs.Scope.Issue = strings.TrimLeft(p.record.Task.ID, "0")
	ledger := p.ledger
	ledger.Rounds = nil
	for i, cut := range saved.Cuts {
		reviewed := ""
		if cut.Finalizing {
			reviewed = saved.CodeHead
		}
		ledger.Rounds = append(ledger.Rounds, gatestate.Round{N: i + 1, Boundary: cut.Boundary, Timestamp: cut.Finalized, Dispatched: cut.Dispatched, Reviewed: reviewed})
	}
	inputs.Ledger = &ledger
	return inputs, ctx.Err()
}
