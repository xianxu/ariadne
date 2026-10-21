package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

const localCheckMaxJSON = 1 << 20

func isV2LocalChecks(data []byte) bool {
	var v struct {
		Version int `json:"version"`
	}
	return json.Unmarshal(data, &v) == nil && v.Version == 2
}

// decodeLocalObject preserves opaque additive values; known names, types and
// duplicate keys remain strict. Shared by the v2 envelope, entries and paths.
func decodeLocalObject(data []byte, fields map[string]any, required ...string) (map[string]json.RawMessage, error) {
	if len(data) > localCheckMaxJSON {
		return nil, fmt.Errorf("JSON exceeds 1 MiB")
	}
	if err := jsonstrict.RejectDuplicateKeys(data); err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("expected object")
	}
	for _, k := range required {
		if _, ok := raw[k]; !ok {
			return nil, fmt.Errorf("missing %s", k)
		}
	}
	extra := map[string]json.RawMessage{}
	for k, v := range raw {
		dst, known := fields[k]
		if !known {
			for name := range fields {
				if strings.EqualFold(k, name) {
					return nil, fmt.Errorf("noncanonical field %s", k)
				}
			}
			extra[k] = append(json.RawMessage(nil), v...)
			continue
		}
		if bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, fmt.Errorf("null %s", k)
		}
		if err := json.Unmarshal(v, dst); err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
	}
	return extra, nil
}

func parseLocalChecksV2(data []byte) (localCheckDecl, error) {
	d := localCheckDecl{Checks: []localCheck{}}
	var entries []json.RawMessage
	var err error
	d.Extra, err = decodeLocalObject(data, map[string]any{"version": &d.Version, "checks": &entries}, "version", "checks")
	if err != nil {
		return d, fmt.Errorf("%s: %w", localChecksPath, err)
	}
	for _, raw := range entries {
		var c localCheck
		fields := map[string]any{"name": &c.Name, "command": &c.Command, "substitutes": &c.Substitutes, "required": &c.Required, "timeout_minutes": &c.TimeoutMinutes, "weight": &c.Weight, "known_flakes": &c.KnownFlakes, "scope": &c.RunScope, "resource": &c.Resource, "paths_file": &c.PathsFile}
		c.Extra, err = decodeLocalObject(raw, fields, "name", "command")
		if err != nil {
			return d, fmt.Errorf("%s: %w", localChecksPath, err)
		}
		if c.RunScope != "" && c.RunScope != "head" && c.RunScope != "landing" {
			return d, fmt.Errorf("invalid scope %q", c.RunScope)
		}
		if c.Resource != "" && c.Resource != "suite" && c.Resource != "runner" {
			return d, fmt.Errorf("invalid resource %q", c.Resource)
		}
		var members map[string]json.RawMessage
		if err := json.Unmarshal(raw, &members); err != nil {
			return d, err
		}
		if c.RunScope == "landing" {
			if _, ok := members["required"]; !ok {
				return d, fmt.Errorf("landing check %s must name required", c.Name)
			}
			if _, ok := members["resource"]; !ok || c.Resource == "" {
				return d, fmt.Errorf("landing check %s must name resource", c.Name)
			}
			if c.Name == suiteCheckName || len(c.Substitutes) > 0 {
				return d, fmt.Errorf("landing check cannot be suite or substitute CI")
			}
		} else if c.PathsFile != "" || c.Resource == "runner" {
			return d, fmt.Errorf("paths_file/runner requires landing scope")
		}
		if c.Resource == "runner" {
			if _, ok := members["weight"]; ok {
				return d, fmt.Errorf("runner check cannot declare weight")
			}
		}
		if c.PathsFile != "" && !validLandingPath(c.PathsFile) {
			return d, fmt.Errorf("invalid paths_file %q", c.PathsFile)
		}
		d.Checks = append(d.Checks, c)
	}
	return validateLocalChecks(d)
}

func validLandingPath(p string) bool {
	return p != "" && p != "." && p != ".." && !strings.HasPrefix(p, "/") && !strings.Contains(p, "\\") && path.Clean(p) == p && !strings.HasPrefix(p, "../")
}

func marshalLocalExtras(value any, extra map[string]json.RawMessage) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	for k, v := range extra {
		for name := range fields {
			if strings.EqualFold(k, name) {
				return nil, fmt.Errorf("extra field conflicts with %s", name)
			}
		}
		fields[k] = v
	}
	return json.Marshal(fields)
}

func (d localCheckDecl) MarshalJSON() ([]byte, error) {
	type plain localCheckDecl
	return marshalLocalExtras(plain(d), d.Extra)
}
func (c localCheck) MarshalJSON() ([]byte, error) {
	type plain localCheck
	raw, err := marshalLocalExtras(plain(c), c.Extra)
	if err != nil || c.RunScope != "landing" {
		return raw, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	fields["required"], err = json.Marshal(c.Required)
	if err != nil {
		return nil, err
	}
	if c.Resource == "runner" {
		delete(fields, "weight")
	}
	return json.Marshal(fields)
}
