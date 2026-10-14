// Package issueassets defines the bounded, body-indexed evidence owned by an issue.
package issueassets

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const (
	MaxFileBytes   = 1 << 20
	MaxBundleBytes = 8 << 20
	MaxFiles       = 32
)

var issueID = regexp.MustCompile(`^[0-9]{6}$`)
var extension = regexp.MustCompile(`^[a-z0-9]{1,16}$`)
var contentName = regexp.MustCompile(`^[0-9a-f]{64}\.[a-z0-9]{1,16}$`)

// Bundle contains a validated snapshot. The constructor owns copies of its bytes.
type Bundle struct {
	Files map[string][]byte
	Bytes int
}
type Delta struct {
	Writes  map[string][]byte
	Deletes []string
}

func Name(data []byte, ext string) (string, error) {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	if !extension.MatchString(ext) {
		return "", fmt.Errorf("unsafe asset extension %q: rename the source with one alphanumeric extension of 1–16 characters", ext)
	}
	if len(data) > MaxFileBytes {
		return "", fmt.Errorf("asset exceeds 1 MiB: reduce the source file to at most 1 MiB")
	}
	return fmt.Sprintf("%x.%s", sha256.Sum256(data), ext), nil
}

// ParseRefs reads actual Markdown links and images, including reference links.
// Asset destinations must use the exact relative spelling; encoded alternatives
// are refused so every consumer sees the same index.
func ParseRefs(id string, body []byte) ([]string, error) {
	if !issueID.MatchString(id) {
		return nil, fmt.Errorf("invalid issue ID %q: use its six-digit issue ID", id)
	}
	doc := goldmark.New().Parser().Parse(text.NewReader(body))
	seen := map[string]bool{}
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		var raw []byte
		switch n := n.(type) {
		case *ast.Link:
			raw = n.Destination
		case *ast.Image:
			raw = n.Destination
		default:
			return ast.WalkContinue, nil
		}
		s := string(raw)
		rendered := string(util.ResolveEntityNames(util.ResolveNumericReferences(util.UnescapePunctuations(raw))))
		decoded, e := url.PathUnescape(rendered)
		if e != nil {
			if strings.Contains(s, "assets") {
				return ast.WalkStop, fmt.Errorf("invalid asset link %q: replace the link with assets/%s/<content-name>", s, id)
			}
			return ast.WalkContinue, nil
		}
		// External URLs never designate a member of this local bundle.
		if u, e := url.Parse(s); e == nil && (u.Scheme != "" || u.Host != "") {
			return ast.WalkContinue, nil
		}
		if !strings.Contains(strings.ReplaceAll(decoded, "\\", "/"), "assets/") {
			return ast.WalkContinue, nil
		}
		prefix := "assets/" + id + "/"
		name := strings.TrimPrefix(s, prefix)
		if s != decoded || !strings.HasPrefix(s, prefix) || strings.ContainsAny(s, "\\?#\x00\r\n") || path.Clean(s) != s || strings.Contains(name, "/") || !contentName.MatchString(name) {
			return ast.WalkStop, fmt.Errorf("invalid asset link %q: replace the link with assets/%s/<sha256>.<extension> for this issue", s, id)
		}
		seen[s] = true
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}
	refs := make([]string, 0, len(seen))
	for s := range seen {
		refs = append(refs, s)
	}
	sort.Strings(refs)
	return refs, nil
}
func BuildBundle(id string, body []byte, files map[string][]byte) (Bundle, error) {
	refs, e := ParseRefs(id, body)
	if e != nil {
		return Bundle{}, e
	}
	if len(files) > MaxFiles || len(refs) > MaxFiles {
		return Bundle{}, fmt.Errorf("issue %s exceeds 32 asset files: remove an asset and its body link", id)
	}
	wanted := map[string]bool{}
	for _, p := range refs {
		wanted[p] = true
	}
	keys := make([]string, 0, len(files))
	for p := range files {
		keys = append(keys, p)
	}
	sort.Strings(keys)
	for _, p := range keys {
		if !wanted[p] {
			return Bundle{}, fmt.Errorf("unknown file %s not linked from the body: link it in the body or delete it", p)
		}
	}
	b := Bundle{Files: map[string][]byte{}}
	for _, p := range refs {
		data, ok := files[p]
		if !ok {
			return Bundle{}, fmt.Errorf("missing linked asset %s: restore the asset or remove its body link", p)
		}
		n, e := Name(data, path.Ext(p))
		if e != nil {
			return Bundle{}, fmt.Errorf("%s: %w", p, e)
		}
		if path.Base(p) != n {
			return Bundle{}, fmt.Errorf("asset %s bytes do not match its content name: restore the original bytes or reattach the changed file", p)
		}
		b.Bytes += len(data)
		if b.Bytes > MaxBundleBytes {
			return Bundle{}, fmt.Errorf("issue %s assets exceed 8 MiB: remove an asset and its body link", id)
		}
		b.Files[p] = bytes.Clone(data)
	}
	return b, nil
}
func BundleDelta(old, next Bundle) Delta {
	d := Delta{Writes: map[string][]byte{}}
	for p, b := range next.Files {
		if a, ok := old.Files[p]; !ok || !bytes.Equal(a, b) {
			d.Writes[p] = bytes.Clone(b)
		}
	}
	for p := range old.Files {
		if _, ok := next.Files[p]; !ok {
			d.Deletes = append(d.Deletes, p)
		}
	}
	sort.Strings(d.Deletes)
	return d
}
