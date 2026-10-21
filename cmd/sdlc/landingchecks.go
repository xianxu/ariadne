package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
)

type landingPaths struct {
	Version int                        `json:"version"`
	Paths   []string                   `json:"paths"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func (p landingPaths) MarshalJSON() ([]byte, error) {
	type plain landingPaths
	return marshalLocalExtras(plain(p), p.Extra)
}

func parseLandingPaths(data []byte) (landingPaths, error) {
	var p landingPaths
	var err error
	p.Extra, err = decodeLocalObject(data, map[string]any{"version": &p.Version, "paths": &p.Paths}, "version", "paths")
	if err != nil {
		return p, err
	}
	if p.Version != 1 || len(p.Paths) == 0 {
		return p, fmt.Errorf("path policy needs version 1 and nonempty paths")
	}
	include := false
	for _, pattern := range p.Paths {
		negative := strings.HasPrefix(pattern, "!")
		glob := strings.TrimPrefix(pattern, "!")
		if !validLandingPath(glob) {
			return p, fmt.Errorf("invalid glob %q", pattern)
		}
		if _, err := affected.Match(glob, ""); err != nil {
			return p, fmt.Errorf("invalid glob %q: %w", pattern, err)
		}
		if negative && !include {
			return p, fmt.Errorf("exclusion before include")
		}
		if !negative {
			include = true
		}
	}
	return p, nil
}

// selectLandingChecks consumes each car's paths, never a net train diff: a
// reverting car cannot erase an earlier car's trigger. The caller supplies only
// the prefix actually landing. Declaration order determines execution order.
func selectLandingChecks(d localCheckDecl, policies map[string][]string, cars [][]string) ([]localCheck, error) {
	return selectLandingChecksWithMatch(d, policies, cars, affected.Match)
}

func selectLandingChecksWithMatch(d localCheckDecl, policies map[string][]string, cars [][]string, match func(string, string) (bool, error)) ([]localCheck, error) {
	paths := map[string]bool{}
	for _, car := range cars {
		for _, p := range car {
			paths[p] = true
		}
	}
	var selected []localCheck
	for _, c := range d.Checks {
		if c.RunScope != "landing" || !c.Required {
			continue
		}
		if c.PathsFile == "" {
			selected = append(selected, c)
			continue
		}
		patterns, ok := policies[c.PathsFile]
		if !ok {
			return nil, fmt.Errorf("check %s: missing path policy %s", c.Name, c.PathsFile)
		}
		matched := false
		for p := range paths {
			included := false
			for _, pattern := range patterns {
				yes, err := match(strings.TrimPrefix(pattern, "!"), p)
				if err != nil {
					return nil, err
				}
				if yes {
					included = !strings.HasPrefix(pattern, "!")
				}
			}
			if included {
				matched = true
				break
			}
		}
		if matched {
			selected = append(selected, c)
		}
	}
	return selected, nil
}

// LandingCheckProof is authority only within the invocation that observed the
// successful run. A stored status without the policy/car binding is not proof.
type LandingCheckProof struct {
	Head, Policy, Cars string
	Run                localRunRecord
}

func landingProofMatches(p LandingCheckProof, name, head, policy, cars string) bool {
	return p.Head == head && p.Policy == policy && p.Cars == cars && p.Run.Name == name && p.Run.SHA == head && p.Run.passed() && p.Run.LogPath != "" && p.Run.LogSHA256 != ""
}
