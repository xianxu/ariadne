package transcripts

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const metadataLimit = 1 << 20

// DiscoverEvidence streams source metadata, retaining only exact requested
// paths. Built-in Codex discovery avoids the legacy whole-file cwd parser.
// Digests cover the stat-sized prefix, so later appends cannot change replay.
func DiscoverEvidence(ctx context.Context, cwds []string, hs []Harness) ([]SessionEvidence, []string, error) {
	var out []SessionEvidence
	var warnings []string
	seen := map[string]bool{}
	allowed := map[string]bool{}
	for _, cwd := range cwds {
		if cwd != "" {
			allowed[normalizedPath(cwd)] = true
		}
	}
	collect := func(path, harness, cwd string, known bool) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		path = normalizedPath(path)
		if seen[path] {
			return nil
		}
		seen[path] = true
		s, err := readEvidenceMetadata(ctx, path, harness, cwd)
		if err != nil {
			if known || allowed[normalizedPath(s.CWD)] || ctx.Err() != nil {
				return fmt.Errorf("source %s: %w", path, err)
			}
			warnings = append(warnings, fmt.Sprintf("skipped source %s: %v", path, err))
			return nil
		}
		if !allowed[normalizedPath(s.CWD)] {
			return nil
		}
		if err := pinEvidence(ctx, &s); err != nil {
			return fmt.Errorf("source %s: %w", path, err)
		}
		out = append(out, s)
		return nil
	}
	for _, h := range hs {
		if err := ctx.Err(); err != nil {
			return nil, warnings, err
		}
		if codex, ok := h.(codexHarness); ok {
			if codex.root == "" || len(allowed) == 0 {
				continue
			}
			err := filepath.WalkDir(codex.root, func(path string, d os.DirEntry, err error) error {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if err != nil {
					if os.IsNotExist(err) {
						return nil
					}
					return err
				}
				if d.IsDir() || filepath.Ext(path) != ".jsonl" {
					return nil
				}
				return collect(path, h.Name(), "", false)
			})
			if err != nil {
				return nil, warnings, err
			}
			continue
		}
		sources := Select(cwds, []Harness{h})
		for _, dir := range sources.Dirs {
			entries, err := os.ReadDir(dir)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, warnings, err
			}
			inferred := ""
			for _, cwd := range cwds {
				if filepath.Base(dir) == cwdToClaudeDir(cwd) {
					inferred = cwd
					break
				}
			}
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".jsonl" {
					if err := collect(filepath.Join(dir, entry.Name()), h.Name(), inferred, true); err != nil {
						return nil, warnings, err
					}
				}
			}
		}
		for _, file := range sources.Files {
			if err := collect(file, h.Name(), "", true); err != nil {
				return nil, warnings, err
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].File < out[j].File })
	sort.Strings(warnings)
	return out, warnings, nil
}

func readEvidenceMetadata(ctx context.Context, path, harness, inferredCWD string) (SessionEvidence, error) {
	s := SessionEvidence{Harness: harness, File: path, CWD: inferredCWD}
	f, err := os.Open(path)
	if err != nil {
		return s, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(io.LimitReader(f, metadataLimit))
	scanner.Buffer(make([]byte, 4096), metadataLimit)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return s, err
		}
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var rec struct {
			Type      string `json:"type"`
			Timestamp string `json:"timestamp"`
			SessionID string `json:"sessionId"`
			CWD       string `json:"cwd"`
			Payload   struct {
				ID        string          `json:"id"`
				CWD       string          `json:"cwd"`
				Timestamp string          `json:"timestamp"`
				ParentID  string          `json:"parent_thread_id"`
				Source    json.RawMessage `json:"source"`
			} `json:"payload"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return s, fmt.Errorf("malformed session metadata: %w", err)
		}
		if harness == "codex" {
			if rec.Type != "session_meta" {
				continue
			}
			s.ID, s.CWD = rec.Payload.ID, rec.Payload.CWD
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
					return s, fmt.Errorf("malformed source lineage: %w", err)
				}
			}
			if source.Subagent.ThreadSpawn != nil {
				s.Spawned = true
				s.ParentID = source.Subagent.ThreadSpawn.ParentID
			}
			if source.ThreadSpawn != nil {
				s.Spawned = true
				s.ParentID = source.ThreadSpawn.ParentID
			}
			if s.Spawned && s.ParentID == "" {
				s.ParentID = rec.Payload.ParentID
			}
		} else {
			if rec.SessionID == "" || rec.Timestamp == "" {
				continue
			}
			s.ID = rec.SessionID
			if rec.CWD != "" {
				s.CWD = rec.CWD
			}
		}
		stamp := rec.Timestamp
		if stamp == "" {
			stamp = rec.Payload.Timestamp
		}
		s.Started, err = time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			return s, fmt.Errorf("invalid session start timestamp: %w", err)
		}
		if s.ID == "" || s.CWD == "" {
			return s, fmt.Errorf("missing session identity or cwd")
		}
		return s, nil
	}
	if err := scanner.Err(); err != nil {
		return s, err
	}
	return s, fmt.Errorf("session metadata absent within %d bytes", metadataLimit)
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func pinEvidence(ctx context.Context, s *SessionEvidence) error {
	f, err := os.Open(s.File)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	s.Length = stat.Size()
	hash := sha256.New()
	n, err := io.Copy(hash, contextReader{ctx: ctx, r: io.LimitReader(f, s.Length)})
	if err != nil {
		return err
	}
	if n != s.Length {
		return fmt.Errorf("source truncated during discovery")
	}
	s.Digest = fmt.Sprintf("%x", hash.Sum(nil))
	return nil
}
