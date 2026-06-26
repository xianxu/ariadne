package transcripts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// codexHarness reads Codex's date-sharded session store at ~/.codex/sessions
// (YYYY/MM/DD/*.jsonl). Unlike Claude, the cwd is not in the path — it lives in a
// session_meta record inside each file — so selection is file-based: walk the
// tree and keep files whose session_meta.cwd is one of the repo cwds.
type codexHarness struct{ root string }

// CodexHarness returns the Codex transcript harness rooted at root
// (~/.codex/sessions in production; a temp dir in tests).
func CodexHarness(root string) Harness { return codexHarness{root: root} }

func (codexHarness) Name() string { return "codex" }

// Sources walks root for *.jsonl session files and returns those whose
// session_meta.cwd is in cwds, sorted. A malformed/empty/no-meta file is skipped,
// never aborting the walk.
func (h codexHarness) Sources(cwds []string) Sources {
	if h.root == "" {
		return Sources{}
	}
	allowed := map[string]bool{}
	for _, cwd := range cwds {
		if cwd != "" {
			allowed[cwd] = true
		}
	}
	if len(allowed) == 0 {
		return Sources{}
	}
	var files []string
	_ = filepath.WalkDir(h.root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".jsonl" {
			return nil
		}
		if allowed[codexSessionCWD(path)] {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	if len(files) == 0 {
		return Sources{}
	}
	return Sources{Files: files}
}

// codexSessionCWD is the thin IO seam: read the file, hand the bytes to the pure
// parser. An unreadable file yields "" (excluded).
func codexSessionCWD(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return codexCWDFromBytes(data)
}

// codexCWDFromBytes extracts session_meta.payload.cwd from a Codex session file's
// bytes. Pure — tolerant of malformed/blank lines and a missing session_meta or
// cwd (returns ""). The robustness contract is asserted directly on bytes.
func codexCWDFromBytes(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec struct {
			Type    string `json:"type"`
			Payload struct {
				CWD string `json:"cwd"`
			} `json:"payload"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if rec.Type == "session_meta" {
			return rec.Payload.CWD
		}
	}
	return ""
}
