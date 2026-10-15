package transcripts

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const UsageCandidateLimit = 16384
const UsageLineLimit = 4 << 20

// UsageSource retains unknown sources so reported spend cannot disappear merely
// because slot/session metadata is unavailable. Discovery pins length, not a hash.
type UsageSource struct {
	SessionEvidence
	// Info pins sampled inode and modification metadata for the streaming caller.
	Info             os.FileInfo `json:"-"`
	MetadataComplete bool
	LineageStatus    string
}

// UsageDiscovery separates readable sources from proven lineage inputs. Old
// parents can appear in Sessions without entering the recent usage source set.
type UsageDiscovery struct {
	Sources      []UsageSource
	Sessions     []SessionEvidence
	Diagnostics  map[string]int
	Examples     map[string][]string
	UnknownFiles []string
}

func (d *UsageDiscovery) note(kind, example string) {
	d.Diagnostics[kind]++
	if len(d.Examples[kind]) < 3 {
		if len(example) > 200 {
			example = example[:200]
		}
		d.Examples[kind] = append(d.Examples[kind], example)
	}
}

type usageCandidate struct {
	UsageSource
	recent             bool
	claudeRoot, toolID string
	err                error
}

// DiscoverUsageEvidence enumerates the complete local stores, including unknown
// cwds and opt-in Claude children. It never changes actual's flat discovery.
func DiscoverUsageEvidence(ctx context.Context, hs []Harness, since time.Time) (UsageDiscovery, error) {
	return discoverUsageEvidence(ctx, hs, since, UsageCandidateLimit)
}

func discoverUsageEvidence(ctx context.Context, hs []Harness, since time.Time, limit int) (UsageDiscovery, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	out := UsageDiscovery{Diagnostics: map[string]int{}, Examples: map[string][]string{}}
	var candidates []*usageCandidate
	seen := map[string]bool{}
	stop := fmt.Errorf("usage candidate limit")
	for _, h := range hs {
		root := ""
		switch h := h.(type) {
		case codexHarness:
			root = h.root
		case claudeHarness:
			root = h.root
		default:
			out.note("unsupported_harness", h.Name())
			continue
		}
		if root == "" {
			continue
		}
		err := filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				if os.IsNotExist(err) && path == root {
					return nil
				}
				out.note("unavailable_source", path)
				return nil
			}
			if e.IsDir() || filepath.Ext(path) != ".jsonl" {
				return nil
			}
			path = filepath.Clean(path)
			if seen[path] {
				return nil
			}
			seen[path] = true
			if out.Diagnostics["candidates"] >= limit {
				return stop
			}
			out.Diagnostics["candidates"]++
			info, err := e.Info()
			if err != nil {
				out.note("unavailable_source", path)
				return nil
			}
			if !info.Mode().IsRegular() {
				out.note("unavailable_source", path)
				return nil
			}
			c := &usageCandidate{UsageSource: UsageSource{SessionEvidence: SessionEvidence{Harness: h.Name(), File: path, ID: "file:" + path, Length: info.Size()}, LineageStatus: "unknown", Info: info}, recent: since.IsZero() || !info.ModTime().Before(since)}
			c.err = readUsageMetadata(ctx, c)
			if !c.MetadataComplete {
				c.ID = "file:" + path
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			candidates = append(candidates, c)
			return nil
		})
		if err == stop {
			out.note("candidate_limit", root)
			break
		}
		if err != nil {
			return out, err
		}
	}
	byID := map[string][]*usageCandidate{}
	for _, c := range candidates {
		if c.MetadataComplete {
			key := c.Harness + ":" + c.ID
			byID[key] = append(byID[key], c)
		}
	}
	// Bind Claude tools by explicit ID across the root and its sidechain files.
	// A tool found in a child makes that child the immediate parent; duplicate
	// dispatch evidence is ambiguous rather than resolved by filesystem order.
	type dispatch struct {
		parent string
		at     time.Time
	}
	dispatches := map[string][]dispatch{}
	needed := map[string]bool{}
	neededRoots := map[string]bool{}
	for _, c := range candidates {
		if c.recent && c.claudeRoot != "" {
			neededRoots[c.claudeRoot] = true
		}
	}
	for _, c := range candidates {
		if c.Harness == "claude" && c.toolID != "" && neededRoots[c.claudeRoot] {
			needed[c.claudeRoot+"\x00"+c.toolID] = true
		}
	}
	for _, c := range candidates {
		if c.Harness != "claude" || !c.MetadataComplete {
			continue
		}
		root := c.claudeRoot
		if root == "" {
			root = c.ID
		}
		hasNeeded := false
		for key := range needed {
			if strings.HasPrefix(key, root+"\x00") {
				hasNeeded = true
				break
			}
		}
		if !hasNeeded {
			continue
		}
		err := scanUsagePrefix(ctx, c.File, c.Length, func(raw []byte) error {
			var rec struct {
				Type, Timestamp string
				Message         struct{ Content json.RawMessage }
			}
			if json.Unmarshal(raw, &rec) != nil || rec.Type != "assistant" {
				return nil
			}
			var blocks []struct{ Type, Name, ID string }
			if json.Unmarshal(rec.Message.Content, &blocks) != nil {
				return nil
			}
			for _, b := range blocks {
				key := root + "\x00" + b.ID
				if b.Type != "tool_use" || (b.Name != "Task" && b.Name != "Agent") || !needed[key] {
					continue
				}
				at, err := time.Parse(time.RFC3339Nano, rec.Timestamp)
				if err != nil {
					return fmt.Errorf("invalid dispatch timestamp")
				}
				duplicate := false
				for _, prior := range dispatches[key] {
					if prior.parent == c.ID && prior.at.Equal(at) {
						duplicate = true
						break
					}
				}
				if !duplicate {
					dispatches[key] = append(dispatches[key], dispatch{parent: c.ID, at: at})
				}
			}
			return nil
		})
		if err != nil {
			if ctx.Err() != nil {
				return out, ctx.Err()
			}
			out.note("dispatch_scan_incomplete", c.File)
			c.err = err
		}
	}
	for _, c := range candidates {
		if c.claudeRoot == "" {
			continue
		}
		matches := dispatches[c.claudeRoot+"\x00"+c.toolID]
		if c.toolID == "" || len(matches) != 1 {
			c.err = fmt.Errorf("Claude child lacks unique explicit dispatch")
			continue
		}
		if matches[0].at.After(c.Started) {
			c.err = fmt.Errorf("Claude dispatch occurs after child execution")
			continue
		}
		c.ParentID = matches[0].parent
		c.Started = matches[0].at
	}
	states := map[*usageCandidate]int{}
	var verify func(*usageCandidate) error
	verify = func(c *usageCandidate) error {
		if states[c] == 1 {
			return fmt.Errorf("session lineage cycle")
		}
		if states[c] == 2 {
			return c.err
		}
		states[c] = 1
		if c.err == nil && !c.MetadataComplete {
			c.err = fmt.Errorf("unknown session metadata")
		}
		if c.err == nil && len(byID[c.Harness+":"+c.ID]) != 1 {
			c.err = fmt.Errorf("ambiguous session identity")
		}
		if c.err == nil && c.Spawned {
			parents := byID[c.Harness+":"+c.ParentID]
			if c.ParentID == "" || len(parents) != 1 {
				c.err = fmt.Errorf("missing or ambiguous parent")
			} else if err := verify(parents[0]); err != nil {
				c.err = err
			} else {
				p := parents[0]
				if c.Started.Before(p.Started) {
					c.err = fmt.Errorf("child dispatch predates parent")
				} else if c.Harness == "claude" {
					c.CWD = p.CWD
				}
			}
		}
		states[c] = 2
		return c.err
	}
	for _, c := range candidates {
		err := verify(c)
		if err == nil {
			c.LineageStatus = "verified"
			out.Sessions = append(out.Sessions, c.SessionEvidence)
		}
		if !c.recent {
			out.Diagnostics["pruned_old_source"]++
			continue
		}
		out.Diagnostics["sources"]++
		if err != nil {
			kind := "invalid_lineage"
			if !c.MetadataComplete {
				kind = "unknown_metadata"
			}
			out.note(kind, c.File)
			out.UnknownFiles = append(out.UnknownFiles, c.File)
		}
		out.Sources = append(out.Sources, c.UsageSource)
	}
	sort.Slice(out.Sources, func(i, j int) bool { return out.Sources[i].File < out.Sources[j].File })
	sort.Slice(out.Sessions, func(i, j int) bool { return out.Sessions[i].File < out.Sessions[j].File })
	sort.Strings(out.UnknownFiles)
	if ctx.Err() != nil {
		return out, ctx.Err()
	}
	return out, nil
}

func readUsageMetadata(ctx context.Context, c *usageCandidate) error {
	found := false
	stop := fmt.Errorf("metadata found")
	limit := c.Length
	if limit > UsageLineLimit {
		limit = UsageLineLimit
	}
	err := scanUsagePrefix(ctx, c.File, limit, func(raw []byte) error {
		// These protocol names do not map through encoding/json's case folding.
		var rec struct {
			Type      string `json:"type"`
			Timestamp string `json:"timestamp"`
			SessionID string `json:"sessionId"`
			CWD       string `json:"cwd"`
			AgentID   string `json:"agentId"`
			Payload   struct {
				CWD       string          `json:"cwd"`
				Timestamp string          `json:"timestamp"`
				ParentID  string          `json:"parent_thread_id"`
				Source    json.RawMessage `json:"source"`
			} `json:"payload"`
		}
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("malformed metadata")
		}
		if c.Harness == "codex" {
			if rec.Type != "session_meta" {
				return nil
			}
			var envelope struct {
				Payload json.RawMessage `json:"payload"`
			}
			if err := json.Unmarshal(raw, &envelope); err != nil {
				return fmt.Errorf("malformed metadata")
			}
			id, err := CodexExecutingID(envelope.Payload)
			if err != nil {
				return err
			}
			c.ID = id
			c.CWD = rec.Payload.CWD
			var source struct {
				Subagent struct {
					ThreadSpawn *struct {
						ParentID string `json:"parent_thread_id"`
					} `json:"thread_spawn"`
				} `json:"subagent"`
				ThreadSpawn *struct {
					ParentID string `json:"parent_thread_id"`
				} `json:"thread_spawn"`
			}
			if len(rec.Payload.Source) > 0 && rec.Payload.Source[0] == '{' {
				if err := json.Unmarshal(rec.Payload.Source, &source); err != nil {
					return fmt.Errorf("malformed source lineage")
				}
			}
			if source.Subagent.ThreadSpawn != nil {
				c.Spawned = true
				c.ParentID = source.Subagent.ThreadSpawn.ParentID
			}
			if source.ThreadSpawn != nil {
				c.Spawned = true
				c.ParentID = source.ThreadSpawn.ParentID
			}
			if c.Spawned && c.ParentID == "" {
				c.ParentID = rec.Payload.ParentID
			}
		} else {
			// Queue lifecycle records can carry session/time before any cwd exists;
			// they are not source metadata. The first actual metadata still wins.
			if rec.Type == "queue-operation" {
				return nil
			}
			if rec.SessionID == "" || rec.Timestamp == "" {
				return nil
			}
			c.ID, c.CWD = rec.SessionID, rec.CWD
			if filepath.Base(filepath.Dir(c.File)) == "subagents" {
				parent := filepath.Base(filepath.Dir(filepath.Dir(c.File)))
				agent := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(c.File), ".jsonl"), "agent-")
				if rec.AgentID == "" || rec.AgentID != agent || rec.SessionID != parent {
					return fmt.Errorf("conflicting Claude child identity")
				}
				c.ID = parent + "/agent/" + agent
				c.ParentID = parent
				c.claudeRoot = parent
				c.Spawned = true
				c.toolID, _ = readUsageToolID(strings.TrimSuffix(c.File, ".jsonl") + ".meta.json")
			}
		}
		stamp := rec.Timestamp
		if stamp == "" {
			stamp = rec.Payload.Timestamp
		}
		var err error
		c.Started, err = time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			return fmt.Errorf("invalid metadata timestamp")
		}
		if c.ID == "" || (c.CWD == "" && c.claudeRoot == "") {
			return fmt.Errorf("missing session identity or cwd")
		}
		c.MetadataComplete = true
		found = true
		return stop
	})
	if err == stop {
		return nil
	}
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("session metadata absent within bounded prefix")
	}
	return nil
}

func scanUsagePrefix(ctx context.Context, path string, length int64, visit func([]byte) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(io.LimitReader(contextReader{ctx: ctx, r: f}, length))
	sc.Buffer(make([]byte, 4096), UsageLineLimit)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if len(strings.TrimSpace(string(sc.Bytes()))) == 0 {
			continue
		}
		if err := visit(sc.Bytes()); err != nil {
			return err
		}
	}
	return sc.Err()
}

func readUsageToolID(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > 64<<10 {
		return "", fmt.Errorf("invalid or oversized Claude meta file")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if err != nil {
		return "", err
	}
	if len(raw) > 64<<10 {
		return "", fmt.Errorf("oversized Claude meta file")
	}
	var meta struct {
		ToolUseID string `json:"toolUseId"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return "", err
	}
	if meta.ToolUseID == "" || strings.ContainsAny(meta.ToolUseID, "\r\n") {
		return "", fmt.Errorf("missing or malformed dispatch identity")
	}
	return meta.ToolUseID, nil
}
