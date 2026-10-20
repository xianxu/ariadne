package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/sidedoc"
	"go.yaml.in/yaml/v3"
)

type moveSource struct {
	Bytes []byte
	Mode  fs.FileMode
	Mtime time.Time
}

type sessionMoveSnapshot struct {
	Root, Old     string
	Files         map[string]moveSource
	Dest          map[string]moveSource
	Existing      map[string]bool
	Directories   map[string]bool
	Ignore        []byte
	IgnoreExists  bool
	Registrations map[string]*syncRegistration
}

type sessionMoveFile struct {
	Old, New      string
	Before, After []byte
	Mode          fs.FileMode
	Mtime         time.Time
}

type sessionMovePlan struct {
	Files       []sessionMoveFile
	Directories []string
	Ignores     []byte
	Notes       []string
}

func moveTarget(old, p string) (string, bool) {
	if p == old {
		return "session.md", true
	}
	if p == path.Base(seenPath(old)) && path.Dir(old) == "." {
		return ".session.md.ss-seen", true
	}
	if p == path.Join(path.Dir(old), "."+path.Base(old)+".ss-seen") {
		return ".session.md.ss-seen", true
	}
	if p == sidedoc.SideDir(old) {
		return "session.side", true
	}
	return sidedoc.MoveUnder(p, sidedoc.SideDir(old), "session.side")
}

func planSessionMove(s sessionMoveSnapshot) (sessionMovePlan, error) {
	if err := rejectMoveIgnoreFiles(s); err != nil {
		return sessionMovePlan{}, err
	}
	p := sessionMovePlan{Ignores: moveIgnores(s.Ignore)}
	for _, dir := range sortedKeys(s.Directories) {
		if to, ok := moveTarget(s.Old, dir); ok {
			if _, occupied := s.Dest[to]; occupied {
				return p, fmt.Errorf("destination directory occupied by file: %s", to)
			}
			p.Directories = append(p.Directories, to)
		}
	}
	for _, old := range sortedKeys(s.Files) {
		src := s.Files[old]
		to, ok := moveTarget(s.Old, old)
		if !ok {
			return p, fmt.Errorf("unexpected source %s", old)
		}
		after := slices.Clone(src.Bytes)
		if strings.HasSuffix(old, ".md") {
			if _, err := moveMetadataFields(string(src.Bytes)); err != nil {
				return p, fmt.Errorf("%s: %w", old, err)
			}
			mapped := func(p string) (string, bool) { return moveTarget(s.Old, p) }
			exists := func(p string) bool { return s.Existing[p] }
			text, dangling := sidedoc.RebaseLinks(string(src.Bytes), old, to, mapped, exists)
			var err error
			text, err = sidedoc.RebaseIndex(text, old, to, mapped)
			if err != nil {
				return p, fmt.Errorf("%s: %w", old, err)
			}
			after = []byte(text)
			for _, target := range dangling {
				p.Notes = append(p.Notes, old+": dangling "+target)
			}
			for _, mention := range mentions(text, s.Old) {
				p.Notes = append(p.Notes, to+": mention "+mention)
			}
		}
		if dst, ok := s.Dest[to]; ok && !bytes.Equal(dst.Bytes, after) {
			return p, fmt.Errorf("destination %s differs; inspect both copies and resolve the conflict", to)
		}
		p.Files = append(p.Files, sessionMoveFile{old, to, src.Bytes, after, src.Mode, src.Mtime})
	}
	// Keep the original session until every side file has been cleaned up.
	slices.SortStableFunc(p.Files, func(a, b sessionMoveFile) int {
		if a.Old == s.Old {
			return 1
		}
		if b.Old == s.Old {
			return -1
		}
		return strings.Compare(a.Old, b.Old)
	})
	return p, nil
}

func moveIgnores(before []byte) []byte {
	s := string(before)
	for _, pattern := range []string{"/session.md", "/session.side/", "/.session.md.ss-seen", "/session.md.lock", "/session.md.clobbered", "/.session.md.shadow", "/.session.md.sync", "/.session.md.flock", "/.session.md.round"} {
		found := false
		for _, line := range strings.Split(s, "\n") {
			if line == pattern {
				found = true
				break
			}
		}
		if !found {
			if s != "" && !strings.HasSuffix(s, "\n") {
				s += "\n"
			}
			s += pattern + "\n"
		}
	}
	return []byte(s)
}

// The invocation-local model grants one effect at a time. A successful
// observation means that effect's postconditions were checked, not merely sent.
// Failed or uncertain observations revoke authority; a retry builds fresh state
// from disk and Couch rather than resuming this model.
type moveEffect uint8

const (
	movePrepare moveEffect = iota
	moveCopy
	moveSwitch
	moveVerify
	moveRemove
	moveDone
	moveStop
)

type moveOutcome uint8

const (
	moveSucceeded moveOutcome = iota
	moveFailed
	moveUnconfirmed
)

type sessionMoveState struct {
	Total, Copies, Bindings, Removed int
	Effect                           moveEffect
	Verified                         bool
}

func transitionMove(s sessionMoveState, observed moveOutcome) (sessionMoveState, error) {
	if s.Total <= 0 || s.Effect >= moveDone {
		return s, fmt.Errorf("terminal or invalid session move state")
	}
	if observed == moveFailed || observed == moveUnconfirmed {
		s.Effect = moveStop
		return s, nil
	}
	if observed != moveSucceeded {
		return s, fmt.Errorf("unknown session move observation")
	}
	switch s.Effect {
	case movePrepare:
		s.Effect = moveCopy
	case moveCopy:
		s.Copies++
		if s.Copies == s.Total {
			s.Effect = moveSwitch
		}
	case moveSwitch:
		if s.Copies != s.Total {
			return s, fmt.Errorf("switch requires all copies")
		}
		s.Bindings++
		if s.Bindings == s.Total {
			s.Effect = moveVerify
		}
	case moveVerify:
		if s.Copies != s.Total || s.Bindings != s.Total {
			return s, fmt.Errorf("cleanup requires verified copies and bindings")
		}
		s.Verified = true
		s.Effect = moveRemove
	case moveRemove:
		if !s.Verified || s.Bindings != s.Total || s.Copies != s.Total {
			return s, fmt.Errorf("deletion requires verified copies and bindings")
		}
		s.Removed++
		if s.Removed == s.Total {
			s.Effect = moveDone
		}
	}
	return s, nil
}

func moveSessionFields(text string) (map[string]string, error) {
	fields, err := moveMetadataFields(text)
	if err != nil {
		return nil, err
	}
	if fields["type"] != "session-sync" {
		return nil, fmt.Errorf("session must have type: session-sync")
	}
	return fields, nil
}

func moveMetadataFields(text string) (map[string]string, error) {
	block, _, ok := sidedoc.FrontBlock(text)
	if !ok {
		if !strings.HasPrefix(text, "---") {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("session frontmatter is not closed")
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(block, "---\n"), "---")), &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("session frontmatter must be a mapping")
	}
	f := map[string]string{}
	root := doc.Content[0]
	read := func(k, v *yaml.Node) error {
		if _, ok := f[k.Value]; ok {
			return fmt.Errorf("duplicate session %s", k.Value)
		}
		if k.Tag != "!!str" || v.Kind != yaml.ScalarNode || v.Tag != "!!str" {
			return fmt.Errorf("session %s must be a string", k.Value)
		}
		f[k.Value] = v.Value
		return nil
	}
	// Type is semantic in every copied document. Other routing fields belong
	// to session-sync only; ordinary documents can use owner for business data.
	for i := 0; i < len(root.Content); i += 2 {
		k, v := root.Content[i], root.Content[i+1]
		if k.Tag == "!!merge" {
			return nil, fmt.Errorf("session routing cannot inherit merged fields")
		}
		if k.Value == "type" {
			if err := read(k, v); err != nil {
				return nil, err
			}
		}
	}
	lines := strings.Split(text, "\n")
	for i := 0; i < len(root.Content); i += 2 {
		k, v := root.Content[i], root.Content[i+1]
		if k.Value == "parent" {
			// RebaseLinks deliberately accepts the exact plain form emitted by
			// tl-lead. Refuse every broader YAML spelling before any move effects.
			if err := read(k, v); err != nil {
				return nil, err
			}
			if k.Style != 0 || v.Style != 0 || k.Line >= len(lines) || lines[k.Line] != "parent: "+v.Value || v.Value == "" || strings.ContainsAny(v.Value, "\"'\\ \t#") {
				return nil, fmt.Errorf("parent must use plain parent: ../file.md without quotes, comments, tags, aliases or block scalars")
			}
		}
		if f["type"] == "session-sync" {
			switch k.Value {
			case "couch-owner", "owner", "operator":
				if err := read(k, v); err != nil {
					return nil, err
				}
			}
		}
	}
	return f, nil
}
