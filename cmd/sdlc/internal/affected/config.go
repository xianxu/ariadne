// Package affected plans the affected suite (#473): which Go packages and declared test groups a change can break.
// It does no IO; LoadGoGraph (golist.go) is its one seam.
package affected

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// Group is a declared non-Go test group, selected by changed paths or by its Go roots' dependencies.
type Group struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths,omitempty"`
	Go    []string `json:"go,omitempty"` // package dirs relative to the module root, e.g. "./cmd/pair"
}

// Config is the parsed .sdlc/affected.json.
type Config struct {
	Version        int      `json:"version"`
	Command        []string `json:"command"`
	Weight         int      `json:"weight,omitempty"`
	TimeoutMinutes int      `json:"timeout_minutes,omitempty"`
	Full           []string `json:"full,omitempty"`
	Groups         []Group  `json:"groups,omitempty"`
}

// groupNameRE is the local-check name rule, so group names need no quoting in make.
var groupNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)

// ParseConfig parses affected.json strictly: a misread selection would skip tests, so anything unexpected refuses.
func ParseConfig(data []byte) (Config, error) {
	if err := jsonstrict.RejectDuplicateKeys(data); err != nil {
		return Config{}, fmt.Errorf("affected.json: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("affected.json: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Config{}, errors.New("affected.json: trailing data after the object")
	}
	if cfg.Version != 1 {
		return Config{}, fmt.Errorf("affected.json: version %d, want 1", cfg.Version)
	}
	if len(cfg.Command) == 0 {
		return Config{}, errors.New("affected.json: command is empty")
	}
	for _, a := range cfg.Command {
		if a == "" {
			return Config{}, errors.New("affected.json: command has an empty argument")
		}
	}
	if cfg.Weight == 0 {
		cfg.Weight = 1
	}
	if cfg.Weight < 1 || cfg.Weight > 8 {
		return Config{}, fmt.Errorf("affected.json: weight %d outside 1..8", cfg.Weight)
	}
	if cfg.TimeoutMinutes == 0 {
		cfg.TimeoutMinutes = 30
	}
	if cfg.TimeoutMinutes < 1 || cfg.TimeoutMinutes > 120 {
		return Config{}, fmt.Errorf("affected.json: timeout_minutes %d outside 1..120", cfg.TimeoutMinutes)
	}
	for _, g := range cfg.Full {
		if !validGlob(g) {
			return Config{}, fmt.Errorf("affected.json: full: bad glob %q", g)
		}
	}
	seen := map[string]bool{}
	for _, gr := range cfg.Groups {
		if !groupNameRE.MatchString(gr.Name) {
			return Config{}, fmt.Errorf("affected.json: group name %q must match %s", gr.Name, groupNameRE)
		}
		if seen[gr.Name] {
			return Config{}, fmt.Errorf("affected.json: duplicate group %q", gr.Name)
		}
		seen[gr.Name] = true
		if len(gr.Paths) == 0 && len(gr.Go) == 0 {
			return Config{}, fmt.Errorf("affected.json: group %q needs paths or go", gr.Name)
		}
		for _, g := range gr.Paths {
			if !validGlob(g) {
				return Config{}, fmt.Errorf("affected.json: group %q: bad glob %q", gr.Name, g)
			}
		}
		for _, r := range gr.Go {
			if !validGoRoot(r) {
				return Config{}, fmt.Errorf("affected.json: group %q: go root %q must be \".\" or \"./dir\"", gr.Name, r)
			}
		}
	}
	return cfg, nil
}

func validGoRoot(r string) bool {
	if r == "." {
		return true
	}
	rest, ok := strings.CutPrefix(r, "./")
	return ok && rest != "" && path.Clean(rest) == rest && !strings.HasPrefix(rest, "..")
}

// goRootDir maps a validated go root to a Graph.Dir value ("./a/b" → "a/b", "." → ".").
func goRootDir(r string) string {
	if rest := strings.TrimPrefix(r, "./"); rest != "" && rest != "." {
		return rest
	}
	return "."
}
