package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

// Inputs belong to Metis's immutable experiment. Actors receive only the
// sanitized public task; with.json and hidden material stay observer-private.
type benchStepInput struct {
	Version       int        `json:"version"`
	Task          bench.Task `json:"task"`
	ToolSource    string     `json:"tool_source"`
	ToolSHA256    string     `json:"tool_sha256"`
	ToolRevision  string     `json:"tool_revision"`
	PrepareStep   string     `json:"prepare_step"`
	SessionStep   string     `json:"session_step"`
	Hub           string     `json:"hub"`
	Worker        string     `json:"worker"`
	Provision     string     `json:"provision"`
	Retirement    string     `json:"retirement"`
	BudgetSeconds int        `json:"budget_seconds"`
}

type benchStepFolder struct{ Dir, Run, ID, Identity string }
type benchStepDeps struct {
	Provenance func(context.Context, benchStepFolder, string) error
}
type benchPreparation struct {
	Version        int    `json:"version"`
	TaskDigest     string `json:"task_sha256"`
	TemplateDigest string `json:"template_sha256"`
}
type benchSessionEvidence struct {
	Version        int                  `json:"version"`
	MetisRun       string               `json:"metis_run"`
	Identity       string               `json:"identity"`
	TaskDigest     string               `json:"task_sha256"`
	TemplateDigest string               `json:"template_sha256"`
	Candidate      string               `json:"candidate"`
	BundleDigest   string               `json:"bundle_sha256"`
	Outcome        string               `json:"outcome"`
	Reason         string               `json:"reason"`
	Runtime        benchRuntimeState    `json:"runtime"`
	Retirement     benchRetirement      `json:"retirement"`
	Facts          bench.WouldLandFacts `json:"would_land"`
	Measurements   bench.Measurements   `json:"measurements"`
}

func loadBenchStepFolder() (benchStepFolder, error) {
	var f benchStepFolder
	f.Dir, f.Run, f.ID = os.Getenv("METIS_STEP_DIR"), os.Getenv("METIS_RUN_DIR"), os.Getenv("METIS_STEP_ID")
	if f.Dir == "" || f.Run == "" || !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$`).MatchString(f.ID) {
		return f, errors.New("bench step requires the Metis step environment")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return f, err
	}
	for _, p := range []string{f.Dir, f.Run} {
		physical, err := filepath.EvalSymlinks(p)
		if err != nil || !filepath.IsAbs(p) || filepath.Clean(p) != p || physical != p || p == string(filepath.Separator) {
			return f, errors.New("Metis folders must be canonical physical directories")
		}
	}
	if cwd != f.Dir || filepath.Join(f.Run, f.ID) != f.Dir {
		return f, errors.New("Metis step cwd or run identity changed")
	}
	sum := sha256.Sum256([]byte(f.Run))
	f.Identity = hex.EncodeToString(sum[:16])
	return f, nil
}

func benchTaskDigest(t bench.Task) (string, error) {
	raw, err := bench.EncodeTask(t)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func runBenchStep(ctx context.Context, kind string, out io.Writer) error {
	return runBenchStepWith(ctx, kind, out, benchStepDeps{})
}

func runBenchStepWith(ctx context.Context, kind string, out io.Writer, d benchStepDeps) error {
	f, err := loadBenchStepFolder()
	if err != nil {
		return err
	}
	if err := requireBenchObserver(ctx); err != nil {
		return err
	}
	raw, err := readBenchInput(filepath.Join(f.Dir, "with.json"), benchRecordLimit)
	if err != nil {
		return err
	}
	var input benchStepInput
	if err := decodeBenchRequired(raw, &input, "version", "task", "tool_source", "tool_sha256", "tool_revision"); err != nil {
		return err
	}
	if input.Version != 1 {
		return errors.New("unsupported bench step version")
	}
	if err := bench.ValidateTask(input.Task); err != nil {
		return err
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(input.ToolSHA256) || !benchCandidatePattern.MatchString(input.ToolRevision) {
		return errors.New("step inputs must pin the backend executable SHA256 and source revision")
	}
	if kind != "prepare" && kind != "session" && kind != "score" {
		return errors.New("unsupported Metis bench step")
	}
	lock, err := os.OpenFile(filepath.Join(f.Dir, ".bench-step.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	info, err := lock.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("Metis step lock is not a regular file")
	}
	if err := flock.Exclusive(lock, true); err != nil {
		return fmt.Errorf("Metis step owner is active: %w", err)
	}
	provenance := d.Provenance
	if provenance == nil {
		provenance = func(ctx context.Context, f benchStepFolder, source string) error {
			return writeBenchStepProvenance(ctx, f, source, input.ToolSHA256, input.ToolRevision)
		}
	}
	if err := provenance(ctx, f, input.ToolSource); err != nil {
		return err
	}
	if err := writeBenchJSONExclusive(filepath.Join(f.Dir, "step-intent.json"), map[string]any{"version": 1, "metis_run": f.Run, "step": f.ID, "backend": kind}); err != nil {
		return fmt.Errorf("step was already started; reconcile retained custody, then start a fresh Metis repeat: %w", err)
	}
	started := time.Now()
	switch kind {
	case "prepare":
		path := filepath.Join(f.Dir, "template.tar")
		if err := prepareBenchTemplate(ctx, input.Task, path); err != nil {
			return err
		}
		taskDigest, err := benchTaskDigest(input.Task)
		if err != nil {
			return err
		}
		templateDigest, err := hashBenchStepFile(path, benchTemplateArchiveLimit)
		if err != nil {
			return err
		}
		if err := writeBenchJSONExclusive(filepath.Join(f.Dir, "preparation.json"), benchPreparation{1, taskDigest, templateDigest}); err != nil {
			return err
		}
		return writeBenchJSONExclusive(filepath.Join(f.Dir, "metrics.json"), map[string]float64{"wall_seconds": time.Since(started).Seconds()})
	case "session":
		return driveBenchSessionStep(ctx, f, input, out)
	default:
		return scoreBenchStep(ctx, f, input)
	}
}

func benchUpstreamFolder(f benchStepFolder, id string) (string, error) {
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$`).MatchString(id) || id == f.ID {
		return "", errors.New("upstream must name a different Metis step")
	}
	p := filepath.Join(f.Run, id)
	physical, err := filepath.EvalSymlinks(p)
	if err != nil || physical != p {
		return "", errors.New("upstream step is missing or aliased")
	}
	return p, nil
}

func loadBenchPreparation(f benchStepFolder, input benchStepInput) (string, benchPreparation, error) {
	var p benchPreparation
	root, err := benchUpstreamFolder(f, input.PrepareStep)
	if err != nil {
		return "", p, err
	}
	raw, err := readBenchInput(filepath.Join(root, "preparation.json"), benchRecordLimit)
	if err != nil {
		return "", p, err
	}
	if err := decodeBenchRequired(raw, &p, "version", "task_sha256", "template_sha256"); err != nil {
		return "", p, err
	}
	digest, err := benchTaskDigest(input.Task)
	if err != nil {
		return "", p, err
	}
	template := filepath.Join(root, "template.tar")
	actual, err := hashBenchStepFile(template, benchTemplateArchiveLimit)
	if err != nil || p.Version != 1 || p.TaskDigest != digest || p.TemplateDigest != actual {
		return "", p, errors.New("prepared task or template digest differs")
	}
	return template, p, nil
}

func waitBenchStepFile(ctx context.Context, path string, deadline time.Time) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("provision/retirement evidence requires an absolute path")
	}
	for {
		if _, err := os.Lstat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("waiting for %s exceeded the step budget; retained custody needs reconciliation", path)
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func driveBenchSessionStep(ctx context.Context, f benchStepFolder, input benchStepInput, out io.Writer) error {
	template, prepared, err := loadBenchPreparation(f, input)
	if err != nil {
		return err
	}
	if input.Hub == "" || input.Worker == "" || input.Hub == input.Worker || input.BudgetSeconds <= 0 || input.BudgetSeconds > 86400 || input.Provision == "" || input.Retirement == "" {
		return errors.New("session needs exact hub/worker, provisioning/retirement paths and a budget within 24h")
	}
	deadline := time.Now().Add(time.Duration(input.BudgetSeconds) * time.Second)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	s, err := createBenchSession(filepath.Join(f.Dir, "runtime"), input.Task, f.Identity)
	if err != nil {
		return err
	}
	helper, err := captureBenchSuiteHelper(input.ToolSource)
	if err != nil {
		return err
	}
	if err := instantiateBenchSession(ctx, s, template, benchPrepareDeps{SuiteHelper: helper}); err != nil {
		return err
	}
	// M1 provisioning is deliberately manual. Roots exist before the operator's
	// supported launch/profile actions; no task dispatch precedes trusted proof.
	fmt.Fprintf(out, "bench roots ready: %s; Metis identity %s; waiting for provisioning %s\n", filepath.Join(s.root, "public"), f.Identity, input.Provision)
	if err := waitBenchStepFile(ctx, input.Provision, deadline); err != nil {
		return err
	}
	runErr := runBenchObserver(ctx, s, benchRunFlags{Hub: input.Hub, Worker: input.Worker, Provision: input.Provision, Budget: time.Until(deadline)}, benchObserverDeps{})
	facts, state, err := s.read()
	if err != nil {
		return err
	}
	if facts.Data["runtime"] == nil {
		return runErr
	}
	fmt.Fprintf(out, "bench task stopped: %v; retire exact actors and supply %s\n", runErr, input.Retirement)
	if err := waitBenchStepFile(ctx, input.Retirement, deadline); err != nil {
		return errors.Join(runErr, err)
	}
	if err := acceptBenchRetirement(ctx, s, input.Retirement, benchCouchDeps{}); err != nil {
		return err
	}
	release, err := s.acquireRun()
	if err != nil {
		return err
	}
	defer release()
	measurements, err := measureBenchRun(ctx, s, benchCouchDeps{})
	if err != nil {
		return err
	}
	facts, state, err = s.read()
	if err != nil {
		return err
	}
	e := benchSessionEvidence{Version: 1, MetisRun: f.Run, Identity: f.Identity, TaskDigest: prepared.TaskDigest, TemplateDigest: prepared.TemplateDigest, Candidate: state.Candidate(), Outcome: "stopped", Measurements: measurements}
	if runErr != nil {
		e.Reason = runErr.Error()
	}
	if err := json.Unmarshal(facts.Data["runtime"], &e.Runtime); err != nil {
		return err
	}
	if err := json.Unmarshal(facts.Data["retirement"], &e.Retirement); err != nil {
		return err
	}
	if e.Candidate != "" {
		if err := json.Unmarshal(facts.Data["would_land_facts"], &e.Facts); err != nil {
			return err
		}
		if err := bench.ValidateWouldLand(e.Facts); err != nil {
			return err
		}
		work := filepath.Join(s.root, "public", "work")
		if _, err := runBenchPrepareCommand(ctx, work, nil, "git", "update-ref", "refs/bench/candidate", e.Candidate); err != nil {
			return err
		}
		bundle := filepath.Join(f.Dir, "candidate.bundle")
		if _, err := runBenchPrepareCommand(ctx, work, nil, "git", "bundle", "create", bundle, "refs/bench/candidate"); err != nil {
			return err
		}
		e.BundleDigest, err = hashBenchStepFile(bundle, 128<<20)
		if err != nil {
			return err
		}
		e.Outcome = "would-land"
	}
	if err := verifyBenchRetirement(ctx, s, benchCouchDeps{}); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	// The immutable evidence remains in runtime/observer. Delete only the retired
	// public mutable trees; Metis then inventories regular evidence artifacts.
	if err := os.RemoveAll(filepath.Join(s.root, "public")); err != nil {
		return err
	}
	if err := writeBenchJSONExclusive(filepath.Join(f.Dir, "session-evidence.json"), e); err != nil {
		return err
	}
	return writeBenchJSONExclusive(filepath.Join(f.Dir, "metrics.json"), bench.ProjectBenchMetrics(measurements, nil))
}

func scoreBenchStep(ctx context.Context, f benchStepFolder, input benchStepInput) error {
	started := time.Now()
	_, prepared, err := loadBenchPreparation(f, input)
	if err != nil {
		return err
	}
	upstream, err := benchUpstreamFolder(f, input.SessionStep)
	if err != nil {
		return err
	}
	raw, err := readBenchInput(filepath.Join(upstream, "session-evidence.json"), benchRecordLimit)
	if err != nil {
		return err
	}
	var e benchSessionEvidence
	if err := decodeBenchRequired(raw, &e, "version", "metis_run", "identity", "task_sha256", "template_sha256", "candidate", "bundle_sha256", "outcome", "runtime", "retirement", "would_land", "measurements"); err != nil {
		return err
	}
	if err := validateBenchSessionEvidence(f, prepared, e, upstream, input.Task); err != nil {
		return err
	}
	if err := verifyBenchRootsVacant(ctx, benchCouchDeps{}, e.Runtime.Hub.Identity.Root, e.Runtime.Worker.Identity.Root); err != nil {
		return err
	}
	work := filepath.Join(f.Dir, "candidate")
	if e.Candidate != "" {
		bundle := filepath.Join(upstream, "candidate.bundle")
		digest, err := hashBenchStepFile(bundle, 128<<20)
		if err != nil || digest != e.BundleDigest {
			return errors.New("candidate bundle digest differs")
		}
		if err := os.Mkdir(work, 0700); err != nil {
			return err
		}
		for _, args := range [][]string{{"init", "-q"}, {"-c", "protocol.file.allow=always", "fetch", "-q", bundle, "refs/bench/candidate"}} {
			if _, err := runBenchPrepareCommand(ctx, work, nil, "git", args...); err != nil {
				return err
			}
		}
	}
	helper, err := captureBenchSuiteHelper(input.ToolSource)
	if err != nil {
		return err
	}
	acceptance, err := scoreBenchEvidence(ctx, input.Task, e, work, f.Dir, benchScoringDeps{SuiteHelper: helper})
	if err != nil {
		return err
	}
	// A confirmed result means the scorer has joined all owned children. Keep
	// only immutable output artifacts, not a live repo for generic Metis walking.
	if e.Candidate != "" {
		if err := os.RemoveAll(work); err != nil {
			return err
		}
	}
	metrics := bench.ProjectBenchMetrics(bench.Measurements{}, &acceptance)
	metrics["wall_seconds"] = time.Since(started).Seconds()
	if acceptance.CommandPassed != nil {
		metrics["hidden_command_passed"] = 0
		if *acceptance.CommandPassed {
			metrics["hidden_command_passed"] = 1
		}
	}
	return writeBenchJSONExclusive(filepath.Join(f.Dir, "metrics.json"), metrics)
}

func validateBenchSessionEvidence(f benchStepFolder, p benchPreparation, e benchSessionEvidence, sessionFolder string, task bench.Task) error {
	if e.Version != 1 || e.MetisRun != f.Run || e.Identity != f.Identity || e.Runtime.RunID != f.Identity || e.TaskDigest != p.TaskDigest || e.TemplateDigest != p.TemplateDigest {
		return errors.New("session evidence belongs to another Metis execution or preparation")
	}
	public := filepath.Join(sessionFolder, "runtime", "public")
	h, w := e.Runtime.Hub.Identity, e.Runtime.Worker.Identity
	if e.Runtime.Version != 1 || h.Validate() != nil || w.Validate() != nil || h.Slot == w.Slot || h.Root != filepath.Join(public, "hub") || w.Root != filepath.Join(public, "work") {
		return errors.New("session evidence does not bind the exact valid Metis runtime roots")
	}
	if !benchProfileEqual(bench.AgentProfile{Agent: h.Agent, Model: h.Model, Effort: h.Effort}, task.TL) || !benchProfileEqual(bench.AgentProfile{Agent: w.Agent, Model: w.Model, Effort: w.Effort}, task.Worker) {
		return errors.New("session evidence differs from the task's approved launch profiles")
	}
	if err := validateBenchRetirement(e.Runtime, e.Retirement); err != nil {
		return err
	}
	if e.Candidate != "" {
		origin := "file:" + filepath.Join(public, "origin.git")
		if e.Outcome != "would-land" || e.Facts.Candidate != e.Candidate || len(e.BundleDigest) != 64 || e.Facts.PrivateOrigin != origin || e.Facts.ExpectedOrigin != origin {
			return errors.New("candidate evidence is incomplete")
		}
		return bench.ValidateWouldLand(e.Facts)
	}
	if e.Outcome != "stopped" || e.BundleDigest != "" {
		return errors.New("stopped session has conflicting candidate evidence")
	}
	return nil
}

func hashBenchStepFile(path string, limit int64) (string, error) {
	physical, err := filepath.EvalSymlinks(path)
	if err != nil || physical != path {
		return "", errors.New("artifact path is missing or aliased")
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("artifact is not regular")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, limit+1))
	if err != nil {
		return "", err
	}
	if n > limit {
		return "", errors.New("artifact exceeds byte bound")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
