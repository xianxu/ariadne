// Package repoenum is the pure core of repo-defined enums (#348): the shared
// model (pkg/vocab RepoEnumModel) declares a field's syntax, and each repository
// declares the field's allowed values in its own declaration file.
//
// A value is allowed when it is declared or is an ancestor of a declared value
// (by "/" segments), so declaring couch/broadcast/web also allows couch and
// couch/broadcast. Queries match by segment prefix.
package repoenum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// Declaration is a repository's parsed declaration: field → allowed values.
type Declaration struct {
	Version int                 `json:"version"`
	Enums   map[string][]string `json:"enums"`
}

// pathPattern builds a field's value pattern from its spec (the model's segment
// rule and depth), so Go restates no syntax of its own.
func pathPattern(spec vocab.RepoEnumSpec) (*regexp.Regexp, error) {
	if spec.Syntax != "path" || spec.Segment == "" || spec.MaxDepth < 1 {
		return nil, fmt.Errorf("unsupported repo enum syntax %+v", spec)
	}
	return regexp.Compile(fmt.Sprintf(`^%s(/%s){0,%d}$`, spec.Segment, spec.Segment, spec.MaxDepth-1))
}

// ValidValue reports whether v has the field's syntax.
func ValidValue(spec vocab.RepoEnumSpec, v string) bool {
	re, err := pathPattern(spec)
	return err == nil && re.MatchString(v)
}

// Parse reads a declaration strictly: unknown fields, another version, a field
// the model does not declare, a value with the wrong syntax, and a duplicate
// are errors. Empty input (no file) is the empty declaration. Errors name the
// declaration path, so a refusal says which file to fix.
func Parse(data []byte, m vocab.RepoEnumModel) (Declaration, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return Declaration{Version: 1, Enums: map[string][]string{}}, nil
	}
	bad := func(format string, a ...any) (Declaration, error) {
		return Declaration{}, fmt.Errorf("%s: "+format, append([]any{m.DeclarationPath}, a...)...)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var d Declaration
	if err := dec.Decode(&d); err != nil {
		return bad("%v", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return bad("trailing data after the declaration")
	}
	if d.Version != 1 {
		return bad("version %d, want 1", d.Version)
	}
	if d.Enums == nil {
		d.Enums = map[string][]string{}
	}
	for field, values := range d.Enums {
		spec, ok := m.Fields[field]
		if !ok {
			return bad("%q is not a repo enum field in the shared model", field)
		}
		re, err := pathPattern(spec)
		if err != nil {
			return bad("%s: %v", field, err)
		}
		seen := map[string]bool{}
		for _, v := range values {
			if !re.MatchString(v) {
				return bad("%s value %q does not match %s", field, v, re)
			}
			if seen[v] {
				return bad("%s value %q declared twice", field, v)
			}
			seen[v] = true
		}
	}
	return d, nil
}

// Allowed reports whether v is declared for field or is an ancestor of a
// declared value.
func (d Declaration) Allowed(field, v string) bool {
	for _, declared := range d.Enums[field] {
		if MatchPrefix(v, declared) {
			return true
		}
	}
	return false
}

// Undeclared returns the values not allowed for field, in order.
func (d Declaration) Undeclared(field string, values []string) []string {
	var out []string
	for _, v := range values {
		if !d.Allowed(field, v) {
			out = append(out, v)
		}
	}
	return out
}

// MatchPrefix reports whether v is prefix or lies under it, by whole segments:
// couch/broadcast matches couch/broadcast/web but not couch/broadcaster.
func MatchPrefix(prefix, v string) bool {
	return v == prefix || strings.HasPrefix(v, prefix+"/")
}

// Refusal is the next-action message for undeclared values: which values,
// which file declares them, and that the file is governed.
func Refusal(m vocab.RepoEnumModel, field string, undeclared []string) error {
	return fmt.Errorf("%s %s not declared in %s on main; declare it there (a local-gates change: it needs the operator's local-gates-verified label)",
		field, strings.Join(quoteAll(undeclared), ", "), m.DeclarationPath)
}

func quoteAll(vs []string) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = fmt.Sprintf("%q", v)
	}
	return out
}
