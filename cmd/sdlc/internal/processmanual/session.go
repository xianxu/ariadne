// session.go — the DYNAMIC counterpart to the static process manual (#157). Given
// a Claude session transcript (JSONL), reconstruct which catalogued injection
// points actually FIRED, in timestamp order, segmented on the 60-min-gap /
// away_summary boundary, each matched to its M1 `Kind`.
//
// It lives beside the M1 catalog on purpose (ARCH-DRY): reconstruction matches
// against the in-process `InjectionSource` set rather than serializing across a
// shell boundary to Python. The pure core (parseEvents / classifyToolUse /
// segmentEvents / renderSessionReport) does no IO; only locateSessionJSONL + the
// file read touch the filesystem.
//
// Two hard limits are documented, not fought (verified against 68 local sessions):
//  1. agents-chain (AGENTS/CLAUDE.md) + memory are session-start SYSTEM-PROMPT
//     injections that never appear in the transcript — only their explicit
//     mid-session Reads would show, so we can assert availability, never firing.
//  2. Forked review PROMPTS aren't in the main JSONL — only their OUTPUT, streamed
//     back through the `sdlc close`/`milestone-close` Bash stdout (where we recover
//     the verdict via judge.ParseVerdict).
package processmanual

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// SessionReport is the public composition for the dynamic pass (mirrors Manual for
// the static one): locate + read the transcript (the IO shell), then the pure core
// (parse → segment → render) matched against the M1 catalog via Collect (ARCH-DRY).
func SessionReport(opts CollectOptions, sessionArg, linkPrefix string) (string, error) {
	jsonlPath, err := locateSessionJSONL(opts.HomeDir, opts.RepoRoot, sessionArg)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(jsonlPath)
	if err != nil {
		return "", err
	}
	catalog := Collect(opts)
	events, allTimes, awaySummaryTimes, err := parseEvents(data, helpTextVerbs(catalog))
	if err != nil {
		return "", err
	}
	segments := segmentEvents(events, allTimes, awaySummaryTimes)
	return renderSessionReport(segments, catalog, linkPrefix), nil
}

// helpTextVerbs is the real, linkable verb set: the titles of the catalog's
// help-text entries (one per `sdlc … --help` contract). Deriving it from the
// catalog keeps "a verb classifies" and "a verb links" as a single source of truth
// (ARCH-DRY) — a fired `sdlc <verb>` is kept iff it's a documented verb.
func helpTextVerbs(catalog []InjectionSource) map[string]bool {
	verbs := map[string]bool{}
	for _, s := range catalog {
		if s.Kind == KindHelpText {
			verbs[s.Title] = true
		}
	}
	return verbs
}

// locateSessionJSONL resolves the transcript path. An explicit sessionArg (any
// value other than "current") is returned as-is. "current" resolves to
// <projDir>/<$CLAUDE_CODE_SESSION_ID>.jsonl when that env var is set AND the file
// exists (the authoritative signal — the harness sets it to the running session),
// else the newest *.jsonl by mtime in the repo's Claude project dir (a Bash call
// appends to the current session, so newest ≈ current; guessy only under concurrent
// same-repo sessions). Reuses claudeProjectSlug (ARCH-DRY).
func locateSessionJSONL(homeDir, absRepoRoot, sessionArg string) (string, error) {
	if sessionArg != "current" {
		return sessionArg, nil
	}
	projDir := filepath.Join(homeDir, ".claude", "projects", claudeProjectSlug(absRepoRoot))
	if sid := os.Getenv("CLAUDE_CODE_SESSION_ID"); sid != "" {
		p := filepath.Join(projDir, sid+".jsonl")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		// env set but file absent → fall through to newest-mtime.
	}
	entries, err := os.ReadDir(projDir)
	if err != nil {
		return "", fmt.Errorf("no Claude session dir for this repo (%s): %w", projDir, err)
	}
	var newest string
	var newestMod time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		if newest == "" || info.ModTime().After(newestMod) {
			newest = filepath.Join(projDir, e.Name())
			newestMod = info.ModTime()
		}
	}
	if newest == "" {
		return "", fmt.Errorf("no session transcripts (*.jsonl) in %s", projDir)
	}
	return newest, nil
}

// FiredEvent is one injection that actually fired in a session — the dynamic
// record that REFERENCES the M1 catalog (via Kind + Detail), rather than mutating
// the static InjectionSource. Its link is resolved against the catalog at render
// time (renderSessionReport), so this stays pure parse output.
type FiredEvent struct {
	Time    time.Time
	Kind    Kind
	Detail  string // verb / skill name / file basename
	Verdict string // optional — review verdict for close/milestone-close
}

// rec is the tolerant JSONL record. Only the fields we consume are named; unknown
// record types (newer sessions add bookkeeping types) unmarshal into a rec whose
// Type we don't switch on and are skipped. encoding/json matches keys
// case-insensitively (Type↔type, ID↔id, …); ToolUseID needs its tag because
// "ToolUseID" won't match the underscored "tool_use_id".
type rec struct {
	Type      string    `json:"type"`
	Subtype   string    `json:"subtype"`
	Timestamp time.Time `json:"timestamp"`
	Message   struct {
		Content []struct {
			Type      string          `json:"type"` // "tool_use" | "tool_result"
			Name      string          `json:"name"`
			ID        string          `json:"id"`
			Input     json.RawMessage `json:"input"`
			ToolUseID string          `json:"tool_use_id"`      // tool_result → its tool_use's id
			Result    json.RawMessage `json:"content"`          // tool_result → its output (string | [{text}])
		} `json:"content"`
	} `json:"message"`
	ToolUseResult json.RawMessage `json:"toolUseResult"` // polymorphic: {stdout,…} | string | null
}

// parseEvents is pure over bytes. It scans the JSONL, keeps the fired injections
// (classifyToolUse), recovers close/milestone-close verdicts from the following
// tool_result's stdout (linked by tool_use_id), and reports segmentation inputs:
// allTimes (every record's timestamp — so non-injection work between fired events
// doesn't trigger a false gap split) and awaySummaryTimes.
func parseEvents(data []byte, validVerbs map[string]bool) (events []FiredEvent, allTimes []time.Time, awaySummaryTimes []time.Time, err error) {
	// A fired event plus the tool_use id it needs for verdict recovery — dropped
	// once the verdict is linked, so FiredEvent stays free of transcript plumbing.
	type pending struct {
		ev FiredEvent
		id string
	}
	var pend []pending
	stdoutByID := map[string]string{}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // transcript lines can be large
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var r rec
		if uerr := json.Unmarshal(line, &r); uerr != nil {
			continue // tolerant: a malformed line is skipped, not fatal
		}
		if !r.Timestamp.IsZero() {
			allTimes = append(allTimes, r.Timestamp)
		}
		switch r.Type {
		case "assistant":
			for _, c := range r.Message.Content {
				if c.Type != "tool_use" {
					continue
				}
				kind, detail, ok := classifyToolUse(c.Name, c.Input, validVerbs)
				if !ok {
					continue
				}
				pend = append(pend, pending{
					ev: FiredEvent{Time: r.Timestamp, Kind: kind, Detail: detail},
					id: c.ID,
				})
			}
		case "user":
			// The tool_result record carries BOTH its tool_use_id (in content) and the
			// top-level toolUseResult.stdout — so a single record links id → stdout.
			for _, c := range r.Message.Content {
				if c.Type == "tool_result" && c.ToolUseID != "" {
					if sd, ok := extractStdout(r.ToolUseResult); ok {
						stdoutByID[c.ToolUseID] = sd
					}
				}
			}
		case "system":
			if r.Subtype == "away_summary" && !r.Timestamp.IsZero() {
				awaySummaryTimes = append(awaySummaryTimes, r.Timestamp)
			}
		}
	}
	if serr := sc.Err(); serr != nil {
		return nil, nil, nil, serr
	}

	// Resolve verdicts after the full scan (the tool_result follows its tool_use, so
	// the map must be complete first). judge.ParseVerdict is the exact fn `sdlc close`
	// uses (ARCH-DRY): it reads the reviewer's `VERDICT:`/block output, not the trailer.
	for _, p := range pend {
		ev := p.ev
		if ev.Kind == KindSDLCPrompt && (ev.Detail == "close" || ev.Detail == "milestone-close") {
			if sd, ok := stdoutByID[p.id]; ok {
				// A fresh close streams the reviewer body (VERDICT: line); a re-close
				// streams only the Review-Verdict git-trailer — fall back to it so the
				// ~20% trailer-only closes don't lose their verdict.
				v := judge.ParseVerdict(sd)
				if v == judge.VerdictUnknown {
					v = judge.ParseVerdictTrailer(sd)
				}
				if v != judge.VerdictUnknown {
					ev.Verdict = string(v)
				}
			}
		}
		events = append(events, ev)
	}
	return events, allTimes, awaySummaryTimes, nil
}

// gapBoundary is the inactivity threshold that opens a new segment — ported from
// introspect's normalize.py (GAP_BOUNDARY_SECONDS = 60*60), so the dynamic manual
// segments a session the same way introspect does (ARCH-DRY on the algorithm).
const gapBoundary = 60 * time.Minute

// segmentEvents buckets the fired stream into segments. A boundary is EITHER an
// away_summary instant OR a >gapBoundary lull between consecutive allTimes (all
// activity, not just fired events — so quiet non-injection work resuming after a
// long break splits, but a burst of unclassified tools between two fired events
// does not). Pure over its inputs.
func segmentEvents(events []FiredEvent, allTimes []time.Time, awaySummaryTimes []time.Time) [][]FiredEvent {
	if len(events) == 0 {
		return nil
	}
	// The resumed timestamp after each long lull — a new segment opens there.
	var gapResumes []time.Time
	for i := 1; i < len(allTimes); i++ {
		if allTimes[i].Sub(allTimes[i-1]) > gapBoundary {
			gapResumes = append(gapResumes, allTimes[i])
		}
	}
	// A boundary falls between the previous and current fired event when a gap
	// resumed (prev < g ≤ cur) or an away_summary was emitted (prev ≤ s < cur; the
	// recap closes the prior segment, so events strictly after it start the next).
	boundaryBetween := func(prev, cur time.Time) bool {
		for _, g := range gapResumes {
			if g.After(prev) && !g.After(cur) {
				return true
			}
		}
		for _, s := range awaySummaryTimes {
			if !s.Before(prev) && s.Before(cur) {
				return true
			}
		}
		return false
	}

	var segments [][]FiredEvent
	cur := []FiredEvent{events[0]}
	for i := 1; i < len(events); i++ {
		if boundaryBetween(events[i-1].Time, events[i].Time) {
			segments = append(segments, cur)
			cur = nil
		}
		cur = append(cur, events[i])
	}
	return append(segments, cur)
}

// hardLimitsHeader states the two things the transcript structurally cannot show,
// rendered INTO the report (ARCH-PURPOSE) rather than silently omitted.
const hardLimitsHeader = "# Session process reconstruction\n\n" +
	"_Generated by `sdlc process-manual --session`. The injection points that actually **fired** in this session, in order, matched to the process-manual catalog._\n\n" +
	"**Two hard limits** (stated, not fought — verified against 68 local sessions):\n" +
	"1. **agents-chain (AGENTS/CLAUDE.md) + memory** are session-start *system-prompt* injections — they never appear in the transcript, so this report can only assert they were *available* (from the static catalog), never that they fired. Only an explicit mid-session Read would surface them.\n" +
	"2. **Forked review *prompts*** aren't in the transcript — only their *output* is, streamed back through the `sdlc close`/`milestone-close` Bash stdout (from which the **verdict** below is recovered).\n\n"

// renderSessionReport is pure markdown: the hard-limits header, then one
// `## Segment N` per segment with a chronological line per fired event —
// `HH:MM:SS · Kind · detail`, the detail linked to its matched catalog source
// (Skill/lessons by exact match; a fired sdlc verb via the help-text fallback),
// verdict inline for close/milestone-close. linkPrefix mirrors M1: prepended to
// repo-relative links so the doc resolves from wherever it is written.
func renderSessionReport(segments [][]FiredEvent, catalog []InjectionSource, linkPrefix string) string {
	var b strings.Builder
	b.WriteString(hardLimitsHeader)
	if len(segments) == 0 {
		b.WriteString("_No catalogued injection points fired in this session._\n")
		return b.String()
	}
	for i, seg := range segments {
		fmt.Fprintf(&b, "## Segment %d\n\n", i+1)
		for _, ev := range seg {
			detail := ev.Detail
			if link := resolveLink(catalog, ev.Kind, ev.Detail); link != "" {
				if !strings.HasPrefix(link, "/") {
					link = linkPrefix + link
				}
				detail = fmt.Sprintf("[%s](%s)", ev.Detail, link)
			}
			fmt.Fprintf(&b, "- `%s` · **%s** · %s", ev.Time.Format("15:04:05"), ev.Kind, detail)
			if ev.Verdict != "" {
				fmt.Fprintf(&b, " — verdict: **%s**", ev.Verdict)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// resolveLink matches a fired event to its catalog source. Skill and lessons
// resolve by exact catalog match (skill Title == dir name; lessons is the single
// entry). A fired sdlc verb has NO sdlc-prompt catalog entry (those are titled by
// judge category, never by verb), so it falls back to the help-text entry for that
// verb — else "" (rendered unlinked, never silently dropped).
func resolveLink(catalog []InjectionSource, kind Kind, detail string) string {
	switch kind {
	case KindSkill:
		for _, s := range catalog {
			if s.Kind == KindSkill && s.Title == detail {
				return s.Link
			}
		}
	case KindLessons:
		for _, s := range catalog {
			if s.Kind == KindLessons {
				return s.Link
			}
		}
	case KindSDLCPrompt, KindHelpText:
		for _, s := range catalog {
			if s.Kind == KindHelpText && s.Title == detail {
				return s.Link
			}
		}
	}
	return ""
}

// extractStdout pulls .stdout from the polymorphic toolUseResult. It is a dict
// {stdout,…} for Bash; a string or null for other tools — both of which fail the
// struct unmarshal, which we swallow to stay tolerant.
func extractStdout(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var obj struct {
		Stdout string `json:"stdout"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", false // string / null / non-object → no stdout
	}
	if obj.Stdout == "" {
		return "", false
	}
	return obj.Stdout, true
}

// sdlcVerbRE matches a real `sdlc <verb>` INVOCATION — `sdlc` at a command
// boundary (start of the command, or right after a shell separator `;|&(){}` /
// newline / backtick, with optional whitespace), then a letter-initial verb. This
// is deliberately precise (precision over recall): it rejects `sdlc` mentioned
// mid-string in a grep pattern, a commit message, or a `--flag` (`./cmd/sdlc
// --include=*.go`, `git commit -m "…sdlc matcher…"`), which the naive substring
// match wrongly counted as fired verbs. The verb is validated against the real
// verb set on top of this (see classifyToolUse), so a real verb name appearing in
// prose right after a separator is also dropped. Known accepted miss: an env-var
// prefix (`VAR=1 sdlc close`) isn't a boundary, so it's not matched — rare, and
// dropping it is the safe side of the precision/recall trade.
var sdlcVerbRE = regexp.MustCompile(`(?:^\s*|[;|&(){}\n\x60]\s*)sdlc ([a-z][a-z-]*)`)

// classifyToolUse is the pure match table (ports the IDEA of introspect's
// segment_text.py summarize_tool_input, not its code): the three injection-bearing
// tool calls we can see in a transcript. A Bash `sdlc <verb>` only classifies when
// the verb is in validVerbs (the real, linkable verb set — so "classified" implies
// "in the catalog"). Anything else → ok=false.
func classifyToolUse(name string, input json.RawMessage, validVerbs map[string]bool) (Kind, string, bool) {
	switch name {
	case "Skill":
		var in struct {
			Skill string `json:"skill"`
		}
		if json.Unmarshal(input, &in) == nil && in.Skill != "" {
			return KindSkill, in.Skill, true
		}
	case "Bash":
		var in struct {
			Command string `json:"command"`
		}
		if json.Unmarshal(input, &in) != nil {
			return "", "", false
		}
		m := sdlcVerbRE.FindStringSubmatch(in.Command)
		if m == nil {
			return "", "", false
		}
		verb := m[1]
		if !validVerbs[verb] {
			return "", "", false // "sdlc <word>" where <word> isn't a real verb (prose mention)
		}
		// `sdlc <verb> --help` prints embedded help text (a distinct Kind from the
		// injected review/gate prompts the bare verb fires).
		if bytes.Contains([]byte(in.Command), []byte("--help")) {
			return KindHelpText, verb, true
		}
		return KindSDLCPrompt, verb, true
	case "Read":
		var in struct {
			FilePath string `json:"file_path"`
		}
		if json.Unmarshal(input, &in) != nil {
			return "", "", false
		}
		base := path.Base(in.FilePath)
		if base == "lessons.md" {
			return KindLessons, base, true
		}
	}
	return "", "", false
}
