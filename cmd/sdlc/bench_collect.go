package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

const benchCaptureLimit int64 = 256 << 20
const benchCaptureLineLimit = 1 << 20

type benchSourceInput struct {
	Role            string                `json:"role"`
	Runtime         bench.RuntimeIdentity `json:"runtime"`
	NativeSession   string                `json:"native_session"`
	File            string                `json:"file"`
	ReviewAttemptID string                `json:"review_attempt_id,omitempty"`
}

type benchCapturedSource struct {
	benchSourceInput
	OriginalFile string `json:"original_file"`
	Digest       string `json:"digest"`
	Length       int64  `json:"length"`
	Device       uint64 `json:"device"`
	Inode        uint64 `json:"inode"`
	Complete     bool   `json:"complete"`
	Issue        string `json:"issue,omitempty"`
}

// captureBenchSources samples each explicit native file once, after runtime
// retirement. It never discovers live stores. The caller owns captureDir under
// observer storage and its normal run-retention lifecycle.
func captureBenchSources(ctx context.Context, captureDir, workerRepo string, inputs []benchSourceInput) (out []benchCapturedSource, retErr error) {
	if len(inputs) == 0 || len(inputs) > 128 {
		return nil, errors.New("capture needs 1..128 explicitly bound sources")
	}
	if !filepath.IsAbs(captureDir) || filepath.Clean(captureDir) != captureDir {
		return nil, errors.New("capture directory must be absolute and clean")
	}
	physical, err := filepath.EvalSymlinks(filepath.Dir(captureDir))
	if err != nil || physical != filepath.Dir(captureDir) {
		return nil, errors.New("capture directory parent is not canonical")
	}
	if err := os.Mkdir(captureDir, 0700); err != nil {
		return nil, err
	}
	var total int64
	seen := map[string]bool{}
	for i, input := range inputs {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		if err := input.Runtime.Validate(); err != nil {
			return out, err
		}
		if input.Role != "worker" && input.Role != "tl" && input.Role != "reviewer" {
			return out, errors.New("unknown captured role")
		}
		if input.NativeSession == "" || len(input.NativeSession) > 256 || strings.ContainsAny(input.NativeSession, "\r\n\x00") {
			return out, errors.New("missing or invalid native session binding")
		}
		if input.Role == "reviewer" {
			if _, err := validateBenchReviewer(ctx, workerRepo, input, "", nil); err != nil {
				excluded := benchCapturedSource{benchSourceInput: input, OriginalFile: input.File, Issue: "unbound reviewer: " + err.Error()}
				excluded.File = ""
				out = append(out, excluded)
				continue
			}
		}
		if !filepath.IsAbs(input.File) || filepath.Clean(input.File) != input.File || seen[input.File] {
			return out, errors.New("invalid or duplicate capture source path")
		}
		seen[input.File] = true
		parent, err := filepath.EvalSymlinks(filepath.Dir(input.File))
		if err != nil || parent != filepath.Dir(input.File) {
			return out, errors.New("source has symlink ancestry")
		}
		source, err := os.OpenFile(input.File, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
		if err != nil {
			return out, err
		}
		info, err := source.Stat()
		if err != nil {
			source.Close()
			return out, err
		}
		if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > benchCaptureLimit-total {
			source.Close()
			return out, errors.New("empty/nonregular source or 256 MiB run capture bound exceeded")
		}
		length := info.Size()
		total += length
		path := filepath.Join(captureDir, fmt.Sprintf("%03d.jsonl", i))
		target, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			source.Close()
			return out, err
		}
		hash := sha256.New()
		sink := &benchCaptureSink{ctx: ctx, writer: io.MultiWriter(target, hash)}
		copied, copyErr := io.Copy(sink, io.LimitReader(source, length))
		after, statErr := source.Stat()
		current, pathErr := os.Lstat(input.File)
		syncErr := target.Sync()
		closeErr := errors.Join(source.Close(), target.Close())
		if err := errors.Join(copyErr, statErr, syncErr, closeErr); err != nil {
			return out, err
		}
		if copied != length {
			return out, errors.New("native transcript shortened during capture")
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return out, errors.New("source lacks native file identity")
		}
		captured := benchCapturedSource{benchSourceInput: input, OriginalFile: input.File, Digest: hex.EncodeToString(hash.Sum(nil)), Length: length, Device: uint64(stat.Dev), Inode: stat.Ino, Complete: true}
		captured.File = path
		if pathErr != nil || !os.SameFile(info, current) || !os.SameFile(info, after) || info.Size() != after.Size() || !info.ModTime().Equal(after.ModTime()) {
			captured.Complete = false
			captured.Issue = "native source changed during capture"
		}
		if sink.lineBytes != 0 {
			captured.Complete = false
			captured.Issue = "captured prefix ends without a complete JSONL line"
		}
		out = append(out, captured)
	}
	return out, nil
}

type benchCaptureSink struct {
	ctx       context.Context
	writer    io.Writer
	lineBytes int
}

func (s *benchCaptureSink) Write(data []byte) (int, error) {
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	for _, b := range data {
		if b == '\n' {
			s.lineBytes = 0
		} else {
			s.lineBytes++
			if s.lineBytes > benchCaptureLineLimit {
				return 0, errors.New("transcript line exceeds 1 MiB")
			}
		}
	}
	return s.writer.Write(data)
}

func readBenchCapturedSource(ctx context.Context, source benchCapturedSource) ([]byte, error) {
	if source.Length <= 0 || source.Length > benchCaptureLimit {
		return nil, errors.New("invalid captured prefix length")
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(source.File))
	if err != nil || parent != filepath.Dir(source.File) {
		return nil, errors.New("captured source ancestry changed")
	}
	f, err := os.OpenFile(source.File, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() != source.Length {
		return nil, errors.New("captured source size/type changed")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(f, source.Length+1))
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	hash := sha256.Sum256(raw)
	if int64(len(raw)) != source.Length || hex.EncodeToString(hash[:]) != source.Digest {
		return nil, errors.New("captured source digest changed")
	}
	return raw, nil
}

type benchSourceDiagnostic struct {
	Role           string            `json:"role"`
	SourceID       string            `json:"source_id"`
	Model          string            `json:"model"`
	Effort         string            `json:"effort"`
	ModelEvidence  string            `json:"model_evidence"`
	EffortEvidence string            `json:"effort_evidence"`
	Valid          bool              `json:"valid"`
	Complete       bool              `json:"complete"`
	Records        int               `json:"records"`
	ToolCalls      int               `json:"tool_calls"`
	Tools          []bench.ToolEvent `json:"-"`
	Issues         []string          `json:"issues"`
}

func (d *benchSourceDiagnostic) note(issue string) {
	if len(d.Issues) < 12 {
		if len(issue) > 512 {
			issue = issue[:512]
		}
		d.Issues = append(d.Issues, issue)
	}
}

func normalizeBenchCapturedSource(ctx context.Context, source benchCapturedSource, windows ...usage.Window) (benchSourceDiagnostic, error) {
	d := benchSourceDiagnostic{Role: source.Role, SourceID: source.Runtime.Agent + ":" + source.NativeSession, Valid: true, Complete: source.Complete, ModelEvidence: "unknown", EffortEvidence: "trusted-provisioning-unverified"}
	raw, err := readBenchCapturedSource(ctx, source)
	if err != nil {
		return d, err
	}
	if len(windows) > 1 {
		return d, errors.New("at most one explicit identity window is allowed")
	}
	cwd := source.Runtime.Root
	header := false
	rootVerified := false
	addTool := func(event bench.ToolEvent) {
		d.ToolCalls++
		if event.Kind == "opaque" || event.Unsupported != "" {
			d.Complete = false
		}
		if len(d.Tools) >= benchEventLimit {
			d.Complete = false
			return
		}
		d.Tools = append(d.Tools, event)
	}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 4096), benchCaptureLineLimit+1)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return d, err
		}
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		d.Records++
		var record struct {
			Type      string          `json:"type"`
			Timestamp string          `json:"timestamp"`
			SessionID string          `json:"sessionId"`
			CWD       string          `json:"cwd"`
			Payload   json.RawMessage `json:"payload"`
			Message   struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if err := jsonstrict.RejectDuplicateKeys(line); err != nil {
			d.Complete = false
			d.Valid = false
			d.note("malformed/duplicate transcript record")
			continue
		}
		if err := json.Unmarshal(line, &record); err != nil {
			d.Valid = false
			d.Complete = false
			continue
		}
		if !header {
			id, root := "", ""
			switch source.Runtime.Agent {
			case "codex":
				if record.Type != "session_meta" {
					d.Complete = false
					continue
				}
				id, err = transcripts.CodexExecutingID(record.Payload)
				var meta struct {
					CWD string `json:"cwd"`
				}
				if e := json.Unmarshal(record.Payload, &meta); e != nil {
					err = e
				}
				root = meta.CWD
			case "claude":
				id, root = record.SessionID, record.CWD
				if id == "" {
					d.Complete = false
					continue
				}
			default:
				d.Valid = false
				d.Complete = false
				d.note("unsupported harness")
				return d, nil
			}
			header = true
			if err != nil || id != source.NativeSession || (root != "" && root != source.Runtime.Root) || (source.Runtime.Agent == "codex" && root == "") {
				d.Valid = false
				d.Complete = false
				d.note("first native session/root differs from runtime binding")
				return d, nil
			}
			rootVerified = root != ""
		} else if source.Runtime.Agent == "claude" && record.SessionID != "" && record.SessionID != source.NativeSession {
			d.Valid = false
			d.Complete = false
			d.note("foreign Claude session record")
			continue
		}
		if source.Runtime.Agent == "claude" && !rootVerified && record.CWD != "" {
			if record.CWD != source.Runtime.Root {
				d.Valid = false
				d.Complete = false
				d.note("first native workspace differs from runtime binding")
				return d, nil
			}
			rootVerified = true
		}
		// Future record kinds may contain new tool protocols. Do not silently
		// claim complete tool coverage merely because JSON decoding succeeded.
		knownRecord := false
		if source.Runtime.Agent == "codex" {
			switch record.Type {
			case "session_meta", "turn_context", "response_item", "event_msg", "token_usage_record", "compacted":
				knownRecord = true
			}
		} else {
			switch record.Type {
			case "user", "assistant", "system", "queue-operation", "file-history-snapshot", "progress", "summary":
				knownRecord = true
			}
		}
		if !knownRecord {
			d.Complete = false
			d.note("unsupported native record kind")
		}
		identity := agentid.Identity{}
		at, stampErr := time.Parse(time.RFC3339Nano, record.Timestamp)
		inWindow := len(windows) == 0 || stampErr == nil && !at.Before(windows[0].Since) && at.Before(windows[0].Until)
		if inWindow {
			identity = agentid.FromTranscript(source.Runtime.Agent, line)
		}
		if identity.Model != "" {
			d.Model = identity.Model
			d.ModelEvidence = "native-transcript"
			if identity.Model != source.Runtime.Model {
				d.Valid = false
				d.note("actual transcript model differs from provisioned profile")
			}
		}
		if identity.Effort != "" {
			d.Effort = identity.Effort
			d.EffortEvidence = "native-transcript"
			if identity.Effort != source.Runtime.Effort {
				d.Valid = false
				d.note("actual transcript effort differs from provisioned profile")
			}
		}
		if source.Runtime.Agent == "codex" {
			if record.Type == "turn_context" {
				var turn struct {
					CWD string `json:"cwd"`
				}
				if json.Unmarshal(record.Payload, &turn) == nil && turn.CWD != "" {
					cwd = turn.CWD
				}
			}
			if record.Type == "response_item" {
				var call struct {
					Type, Name, Arguments, CallID string
					Input                         json.RawMessage
				}
				// Native payload uses snake_case only for the call identity.
				var wire struct {
					Type      string          `json:"type"`
					Name      string          `json:"name"`
					Arguments string          `json:"arguments"`
					CallID    string          `json:"call_id"`
					Input     json.RawMessage `json:"input"`
				}
				if json.Unmarshal(record.Payload, &wire) != nil {
					d.Complete = false
					continue
				}
				call.Type, call.Name, call.Arguments, call.CallID, call.Input = wire.Type, wire.Name, wire.Arguments, wire.CallID, wire.Input
				if call.Type == "function_call" || call.Type == "custom_tool_call" {
					input := json.RawMessage(call.Arguments)
					if call.Type == "custom_tool_call" {
						input = call.Input
					}
					addTool(normalizeBenchTool(call.Name, call.CallID, cwd, input))
				} else if call.Type != "message" && call.Type != "reasoning" && call.Type != "function_call_output" && call.Type != "custom_tool_call_output" {
					addTool(bench.ToolEvent{Kind: "opaque", CWD: cwd, Unsupported: "unsupported native response item: " + call.Type})
				}
			}
		} else if record.Type == "assistant" {
			var blocks []struct {
				Type  string          `json:"type"`
				Name  string          `json:"name"`
				ID    string          `json:"id"`
				Input json.RawMessage `json:"input"`
			}
			if len(record.Message.Content) > 0 && json.Unmarshal(record.Message.Content, &blocks) != nil {
				d.Complete = false
				continue
			}
			for _, block := range blocks {
				if block.Type == "tool_use" {
					addTool(normalizeBenchTool(block.Name, block.ID, cwd, block.Input))
				} else if block.Type != "text" && block.Type != "thinking" && block.Type != "redacted_thinking" {
					addTool(bench.ToolEvent{Kind: "opaque", CWD: cwd, Unsupported: "unsupported assistant content: " + block.Type})
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	if !header || !rootVerified {
		d.Valid = false
		d.Complete = false
		d.note("native session header unavailable")
	}
	if d.Model == "" {
		d.note("actual model unavailable; provisioning remains unverified")
	}
	if d.Effort == "" {
		d.note("actual effort unavailable; trusted provisioning only, unverified")
	}
	return d, nil
}

func normalizeBenchTool(name, id, cwd string, raw json.RawMessage) bench.ToolEvent {
	event := bench.ToolEvent{ID: id, CWD: cwd, Kind: "opaque", Unsupported: "unsupported tool: " + name}
	// A named GitHub connector invocation is itself diagnostic access evidence,
	// even while its provider-specific operand schema remains unsupported.
	if strings.Contains(strings.ToLower(name), "github") {
		event.Kind = "exec"
		event.Command = "github.com tool " + name
		return event
	}
	var args struct {
		Command  json.RawMessage `json:"command"`
		Cmd      string          `json:"cmd"`
		Argv     []string        `json:"argv"`
		Path     string          `json:"path"`
		FilePath string          `json:"file_path"`
		CWD      string          `json:"cwd"`
		Workdir  string          `json:"workdir"`
	}
	if jsonstrict.RejectDuplicateKeys(raw) != nil || json.Unmarshal(raw, &args) != nil {
		event.Unsupported = "malformed tool arguments: " + name
		return event
	}
	if args.CWD != "" {
		event.CWD = args.CWD
	}
	if args.Workdir != "" {
		event.CWD = args.Workdir
	}
	switch strings.TrimPrefix(name, "functions.") {
	case "Read", "read_file":
		path := args.FilePath
		if path == "" {
			path = args.Path
		}
		if path == "" {
			return event
		}
		event.Kind = "read"
		event.Paths = []string{path}
		event.Unsupported = ""
	case "Bash", "exec_command", "shell_command":
		event.Kind = "exec"
		event.Command = args.Cmd
		event.Argv = args.Argv
		if len(args.Command) > 0 {
			if json.Unmarshal(args.Command, &event.Argv) != nil {
				if json.Unmarshal(args.Command, &event.Command) != nil {
					return event
				}
			}
		}
		if event.Command != "" || len(event.Argv) > 0 {
			event.Unsupported = ""
		}
	}
	return event
}

// Types below keep captured evidence separate from its derived measurements.
type benchCollectionInput struct {
	Task         bench.Task
	Run          bench.Run
	Sources      []benchCapturedSource
	Since, Until time.Time
	WorkerRepo   string
	Focused      *ownedActualInputs
}

type benchCollectionResult struct {
	RunID        string                  `json:"run_id"`
	Captured     []benchCapturedSource   `json:"captured_sources"`
	Measurements bench.Measurements      `json:"measurements"`
	Sources      []benchSourceDiagnostic `json:"sources"`
	// Digest hashes this private collection document with Digest empty.
	Digest        string                    `json:"digest"`
	Valid         bool                      `json:"valid"`
	Issues        []string                  `json:"issues"`
	UsageCoverage map[string]usage.Coverage `json:"usage_coverage"`
	Visits        benchCollectionVisits     `json:"visits"`
}

type benchCollectionVisits struct {
	CapturedBytes       int64 `json:"captured_bytes"`
	NormalizedRecords   int   `json:"normalized_records"`
	UsageRecords        int64 `json:"usage_records"`
	UsageResponses      int   `json:"usage_responses"`
	AggregatedResponses int   `json:"aggregated_responses"`
	ActivityPrefixes    int   `json:"activity_prefixes"`
	LeakEvents          int   `json:"leak_events"`
}

func collectBenchMeasurements(ctx context.Context, in benchCollectionInput) (benchCollectionResult, error) {
	out := benchCollectionResult{RunID: in.Run.ID(), Valid: true, Captured: append([]benchCapturedSource(nil), in.Sources...), UsageCoverage: map[string]usage.Coverage{}}
	unknown := bench.RoleUsage{Coverage: "unknown"}
	out.Measurements = bench.Measurements{Autonomous: true, Worker: unknown, TL: unknown, Reviewer: unknown}
	if err := bench.ValidateTask(in.Task); err != nil {
		return out, err
	}
	if !in.Run.RuntimeStopped() || in.Since.IsZero() || !in.Until.After(in.Since) || in.Until.Sub(in.Since) > 24*time.Hour || len(in.Sources) > 128 {
		return out, errors.New("collection requires retired runtime and a bounded explicit run window")
	}
	hub, worker := in.Run.Bindings()
	if in.WorkerRepo != worker.Root {
		return out, errors.New("collection worker repository differs from runtime binding")
	}
	out.Measurements.GrossNanos = in.Until.Sub(in.Since).Nanoseconds()
	window := usage.Window{Since: in.Since, Until: in.Until}
	roles := map[string]bench.AgentProfile{"worker": in.Task.Worker, "tl": in.Task.TL, "reviewer": in.Task.Reviewer}
	sourceRole := map[string]string{}
	sourceSlot := map[string]string{}
	sourceWindows := map[string]usage.Window{}
	roleSources := map[string][]string{}
	roleComplete := map[string]bool{"worker": true, "tl": true, "reviewer": true}
	sourceComplete := true
	var responses []usage.Response
	var tools []bench.ToolEvent
	for _, source := range in.Sources {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		if source.Length > benchCaptureLimit-out.Visits.CapturedBytes {
			return out, errors.New("256 MiB captured run bound exceeded")
		}
		out.Visits.CapturedBytes += source.Length
		expected, knownRole := roles[source.Role]
		key := source.Runtime.Agent + ":" + source.NativeSession
		if !knownRole || sourceRole[key] != "" || source.Runtime.Validate() != nil {
			return out, errors.New("invalid/duplicate captured source role or identity")
		}
		attributionWindow := window
		if source.Role == "reviewer" {
			reviewRecord, err := validateBenchReviewer(ctx, in.WorkerRepo, source.benchSourceInput, in.Task.ID, &window)
			if err != nil || source.File == "" {
				reason := source.Issue
				if err != nil {
					reason = err.Error()
				}
				d := benchSourceDiagnostic{Role: source.Role, SourceID: key, Valid: true, Complete: false, ModelEvidence: "unknown", EffortEvidence: "unbound"}
				d.note("reviewer source excluded: " + reason)
				out.Sources = append(out.Sources, d)
				coverage := out.UsageCoverage["reviewer"]
				coverage.Add("unavailable_reviewer_bindings", reason)
				out.UsageCoverage["reviewer"] = coverage
				sourceComplete = false
				roleComplete["reviewer"] = false
				continue
			}
			requested, _ := time.Parse(time.RFC3339Nano, reviewRecord.Intent.RequestedAt)
			finished, _ := time.Parse(time.RFC3339Nano, reviewRecord.FinishedAt)
			if requested.After(attributionWindow.Since) {
				attributionWindow.Since = requested
			}
			if finished.Before(attributionWindow.Until) {
				attributionWindow.Until = finished
			}
		}
		sourceRole[key] = source.Role
		sourceSlot[key] = source.Runtime.Slot
		sourceWindows[key] = attributionWindow
		roleSources[source.Role] = append(roleSources[source.Role], key)
		if source.Role == "worker" && source.Runtime != worker || source.Role == "tl" && source.Runtime != hub || source.Runtime.Agent != expected.Agent || source.Runtime.Model != expected.Model || source.Runtime.Effort != expected.Effort {
			out.Valid = false
			out.Issues = append(out.Issues, "captured runtime profile/binding differs from immutable task/run: "+key)
		}
		diagnostic, err := normalizeBenchCapturedSource(ctx, source, attributionWindow)
		if err != nil {
			out.Valid = false
			sourceComplete = false
			roleComplete[source.Role] = false
			diagnostic.Valid = false
			diagnostic.Complete = false
			diagnostic.note(err.Error())
			out.Sources = append(out.Sources, diagnostic)
			continue
		}
		out.Sources = append(out.Sources, diagnostic)
		out.Visits.NormalizedRecords += diagnostic.Records
		// Observed tool evidence survives an invalid profile. Bound the run-wide
		// collection too; an omitted tail must remain incomplete coverage.
		available := benchEventLimit - len(tools)
		if len(diagnostic.Tools) > available {
			sourceComplete = false
			tools = append(tools, diagnostic.Tools[:available]...)
		} else {
			tools = append(tools, diagnostic.Tools...)
		}
		if !diagnostic.Valid {
			out.Valid = false
			sourceComplete = false
			roleComplete[source.Role] = false
			continue
		}
		sourceComplete = sourceComplete && diagnostic.Complete
		raw, err := readBenchCapturedSource(ctx, source)
		if err != nil {
			return out, err
		}
		rows, coverage, err := usage.Decode(ctx, source.Runtime.Agent, source.NativeSession, bytes.NewReader(raw))
		if err != nil {
			if ctx.Err() != nil {
				return out, ctx.Err()
			}
			coverage.Add("unavailable_sources", err.Error())
		}
		if !source.Complete {
			coverage.Add("truncated_tails", source.Issue)
		}
		combined := out.UsageCoverage[source.Role]
		combined.Merge(coverage)
		out.UsageCoverage[source.Role] = combined
		out.Visits.UsageRecords += coverage.Counts["records"]
		out.Visits.UsageResponses += len(rows)
		if len(responses)+len(rows) > usage.MaxResponses {
			return out, errors.New("bench response identity bound exceeded")
		}
		for _, response := range rows {
			attributed := !response.At.Before(attributionWindow.Since) && response.At.Before(attributionWindow.Until)
			if attributed && response.Model != "" && response.Model != expected.Model {
				out.Valid = false
				out.Issues = append(out.Issues, "actual response model differs from pinned task: "+key)
			}
			if attributed && response.Model != "" && out.Sources[len(out.Sources)-1].Model == "" {
				out.Sources[len(out.Sources)-1].Model = response.Model
				out.Sources[len(out.Sources)-1].ModelEvidence = "native-response"
			}
		}
		responses = append(responses, rows...)
	}
	normalized, globalCoverage, err := usage.NormalizeResponses(responses)
	if err != nil {
		return out, err
	}
	out.UsageCoverage["all"] = globalCoverage
	partition := map[string][]usage.Response{}
	for _, response := range normalized {
		role := sourceRole[response.Agent+":"+response.SourceID]
		if role == "" {
			out.Valid = false
			out.Issues = append(out.Issues, "unbound normalized usage response")
			continue
		}
		partition[role] = append(partition[role], response)
	}
	for _, role := range []string{"worker", "tl", "reviewer"} {
		rows := partition[role]
		out.Visits.AggregatedResponses += len(rows)
		report, err := usage.AggregateUsage(rows, window, window, func(response usage.Response) usage.Attribution {
			key := response.Agent + ":" + response.SourceID
			bound := sourceWindows[key]
			return usage.Attribution{Include: !response.At.Before(bound.Since) && response.At.Before(bound.Until), Role: role, Slot: sourceSlot[key]}
		})
		if err != nil {
			return out, err
		}
		coverage := out.UsageCoverage[role]
		coverage.Merge(report.Coverage)
		coverage.Merge(globalCoverage)
		out.UsageCoverage[role] = coverage
		roleUsage := bench.RoleUsage{Coverage: "unknown", SourceIDs: roleSources[role]}
		if report.Totals != nil {
			roleUsage.InputTokens = &report.Totals.Input
			roleUsage.OutputTokens = &report.Totals.Output
			roleUsage.CachedTokens = report.Totals.CacheRead
			roleUsage.Coverage = "partial"
			if coverage.Complete() && roleComplete[role] {
				roleUsage.Coverage = "complete"
			}
		}
		if report.Credits != nil {
			credit := json.Number(report.Credits.String())
			roleUsage.Credits = &credit
			if roleUsage.Coverage == "unknown" {
				roleUsage.Coverage = "partial"
			}
		}
		switch role {
		case "worker":
			out.Measurements.Worker = roleUsage
		case "tl":
			out.Measurements.TL = roleUsage
		case "reviewer":
			out.Measurements.Reviewer = roleUsage
		}
	}
	scope := bench.LeakScope{PreparedRoots: []string{worker.Root, hub.Root}, ProtectedRoots: []string{in.Task.SourceRepo}, Complete: sourceComplete && len(roleSources["worker"]) > 0 && len(roleSources["tl"]) > 0}
	out.Measurements.Leaks = bench.ScanLeakEvents(scope, tools)
	out.Visits.LeakEvents = out.Measurements.Leaks.Coverage.Events
	if in.Focused == nil {
		out.Issues = append(out.Issues, "focused hours unavailable without retained ownership measurement inputs")
	} else {
		focused, err := capturedBenchActualInputs(in)
		if err != nil {
			out.Issues = append(out.Issues, "focused hours unavailable: "+err.Error())
		} else {
			out.Visits.ActivityPrefixes = len(focused.Files)
			measured := computeOwnedActualFor(ctx, in.WorkerRepo, focused.Scope.Issue, "", focused)
			if measured.Status == actualMeasured {
				hours := measured.Hours
				out.Measurements.ActiveHours = &hours
			} else {
				out.Issues = append(out.Issues, "focused hours unavailable: "+measured.Detail)
			}
		}
	}
	setBenchCollectionValidity(&out)
	raw, err := json.Marshal(out)
	if err != nil {
		return out, err
	}
	if len(raw) > benchRecordLimit {
		return out, errors.New("bench collection metadata exceeds 1 MiB")
	}
	digest := sha256.Sum256(raw)
	out.Digest = hex.EncodeToString(digest[:])
	return out, nil
}

func validateBenchReviewer(ctx context.Context, workerRepo string, input benchSourceInput, issueID string, window *usage.Window) (reviewusage.Record, error) {
	var absent reviewusage.Record
	if input.ReviewAttemptID == "" {
		return absent, errors.New("reviewer attempt identity unavailable")
	}
	b, err := loadBenchContext(ctx, workerRepo)
	if err != nil {
		return absent, err
	}
	if b == nil || b.WorkRoot != workerRepo {
		return absent, errors.New("reviewer accounting requires the exact bench worker repository")
	}
	store, err := reviewusage.NewStore(workerRepo)
	if err != nil {
		return absent, err
	}
	entry, err := store.Load(ctx, input.ReviewAttemptID)
	if err != nil {
		return absent, err
	}
	r := entry.Record
	if r.Phase != reviewusage.Finished || r.SessionID == "" || r.SessionID != input.NativeSession || r.Intent.Agent != input.Runtime.Agent || r.Intent.Root != input.Runtime.Root || r.Intent.Repository != b.OriginIdentity || r.Intent.Slot != input.Runtime.Slot || (issueID != "" && r.Intent.IssueID != issueID) {
		return absent, errors.New("reviewer accounting does not bind this exact native source/issue")
	}
	if window != nil {
		requested, _ := time.Parse(time.RFC3339Nano, r.Intent.RequestedAt)
		finished, _ := time.Parse(time.RFC3339Nano, r.FinishedAt)
		if !finished.After(window.Since) || !requested.Before(window.Until) {
			return absent, errors.New("reviewer accounting lies outside this run window")
		}
	}
	return r, nil
}

func setBenchCollectionValidity(out *benchCollectionResult) {
	out.Measurements.Validity = "valid"
	add := func(text string) {
		if len(out.Measurements.ValidityIssues) < 128 {
			if len(text) > 512 {
				text = text[:512]
			}
			out.Measurements.ValidityIssues = append(out.Measurements.ValidityIssues, text)
		}
	}
	for _, source := range out.Sources {
		if !source.Valid {
			out.Valid = false
			for _, issue := range source.Issues {
				add(source.Role + " " + source.SourceID + ": " + issue)
			}
		} else if source.Model == "" || source.Effort == "" || !source.Complete {
			if out.Measurements.Validity != "invalid" {
				out.Measurements.Validity = "unverified"
			}
		}
	}
	if !out.Valid {
		out.Measurements.Validity = "invalid"
		for _, issue := range out.Issues {
			add(issue)
		}
		if len(out.Measurements.ValidityIssues) == 0 {
			add("captured actor evidence failed validation")
		}
	}
	if len(out.Sources) == 0 {
		out.Measurements.Validity = "unverified"
	}
	if out.Measurements.Validity == "valid" && (out.Measurements.Worker.Coverage != "complete" || out.Measurements.TL.Coverage != "complete" || out.Measurements.Reviewer.Coverage != "complete") {
		out.Measurements.Validity = "unverified"
	}
}

// The normal owner still computes phases, queue subtraction and attribution.
// Only its explicitly retained source paths move to immutable observer copies.
func capturedBenchActualInputs(in benchCollectionInput) (*ownedActualInputs, error) {
	original := in.Focused
	if err := original.Snapshot.Validate(); err != nil {
		return nil, err
	}
	if !gatestate.IsEngagementMethod(original.Snapshot.Method) || original.LegacyPrefix != nil || original.Windows == nil || len(original.Files) == 0 || len(original.Sessions) == 0 {
		return nil, errors.New("focused inputs lack exact owned source windows")
	}
	issueNumber, err := strconv.Atoi(in.Task.ID)
	if err != nil || original.Scope.Issue != strconv.Itoa(issueNumber) || original.Scope.Head != original.Snapshot.CodeHead {
		return nil, errors.New("retained activity scope differs from the replay issue/head")
	}
	since, err := time.Parse(time.RFC3339Nano, original.Snapshot.Since)
	if err != nil {
		return nil, err
	}
	until, err := time.Parse(time.RFC3339Nano, original.Snapshot.Until)
	if err != nil {
		return nil, err
	}
	if since.Before(in.Since) || until.After(in.Until) {
		return nil, errors.New("retained ownership measurement exceeds the bench run window")
	}
	mapped := map[string]benchCapturedSource{}
	for _, source := range in.Sources {
		mapped[source.OriginalFile] = source
	}
	copy := *original
	copy.Replay = true
	copy.Files = nil
	copy.Sessions = nil
	copy.ExcludedBrain = nil
	copy.Prefixes = map[string]activetime.SourcePrefix{}
	copy.Windows = map[string][]activetime.ActivityWindow{}
	copy.SourceSlots = map[string]string{}
	for _, file := range original.Files {
		source, ok := mapped[file]
		if !ok || !source.Complete {
			return nil, errors.New("retained activity source lacks complete captured bytes")
		}
		pin, ok := original.Prefixes[file]
		if !ok || pin.Bytes > source.Length {
			return nil, errors.New("retained activity prefix is not captured")
		}
		copy.Files = append(copy.Files, source.File)
		copy.Prefixes[source.File] = pin
		copy.Windows[source.File] = append([]activetime.ActivityWindow(nil), original.Windows[file]...)
		copy.SourceSlots[source.File] = original.SourceSlots[file]
	}
	for _, session := range original.Sessions {
		source, ok := mapped[session.File]
		if !ok || source.NativeSession != session.ID || source.Runtime.Agent != session.Harness {
			return nil, errors.New("retained activity session differs from captured native session")
		}
		session.File = source.File
		copy.Sessions = append(copy.Sessions, session)
	}
	return &copy, nil
}
