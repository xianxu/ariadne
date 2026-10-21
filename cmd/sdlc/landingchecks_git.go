package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

type landingCheckInputs struct {
	Decl     localCheckDecl
	Policies map[string][]string
	Paths    [][]string
	Policy   string
	Cars     string
}

// Cache ownership is one landing invocation. Keys pin immutable base/head
// commits, so polling and repeated prefix selection never rescan Git histories.
type landingInputCache struct {
	bases map[string]landingCheckInputs
	paths map[string][]string
}

func readLandingBlob(git func(...string) (string, error), base, file string, optional bool) (string, error) {
	if !validLandingPath(file) {
		return "", fmt.Errorf("invalid tracked policy path %q", file)
	}
	tree, err := git("ls-tree", "--full-tree", "-z", base, "--", file)
	if err != nil {
		return "", err
	}
	if tree == "" && optional {
		return "", nil
	}
	entry := strings.TrimSuffix(tree, "\x00")
	parts := strings.SplitN(entry, "\t", 2)
	if len(parts) != 2 || parts[1] != file {
		return "", fmt.Errorf("missing or ambiguous policy %s on %s", file, base)
	}
	fields := strings.Fields(parts[0])
	if len(fields) != 3 || fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
		return "", fmt.Errorf("policy %s is not a regular tracked blob", file)
	}
	raw, err := git("show", base+":"+file)
	if err != nil {
		return "", err
	}
	if len(raw) > localCheckMaxJSON {
		return "", fmt.Errorf("policy %s exceeds 1 MiB", file)
	}
	return raw, nil
}

func readLandingCheckInputs(git func(...string) (string, error), base string, heads []string, cache *landingInputCache) (landingCheckInputs, error) {
	if cache.bases == nil {
		cache.bases = map[string]landingCheckInputs{}
		cache.paths = map[string][]string{}
	}
	in, ok := cache.bases[base]
	if !ok {
		raw, err := readLandingBlob(git, base, localChecksPath, true)
		if err != nil {
			return in, err
		}
		in.Decl, err = parseLocalChecks([]byte(raw))
		if err != nil {
			return in, err
		}
		in.Policies = map[string][]string{}
		digest := sha256.New()
		_, _ = digest.Write([]byte(raw))
		for _, c := range in.Decl.Checks {
			if c.RunScope != "landing" || !c.Required || c.PathsFile == "" {
				continue
			}
			if _, seen := in.Policies[c.PathsFile]; seen {
				continue
			}
			blob, err := readLandingBlob(git, base, c.PathsFile, false)
			if err != nil {
				return in, err
			}
			p, err := parseLandingPaths([]byte(blob))
			if err != nil {
				return in, fmt.Errorf("%s: %w", c.PathsFile, err)
			}
			in.Policies[c.PathsFile] = p.Paths
			_, _ = digest.Write([]byte("\x00" + c.PathsFile + "\x00" + blob))
		}
		in.Policy = hex.EncodeToString(digest.Sum(nil))
		cache.bases[base] = in
	}
	in.Paths = nil
	in.Cars = strings.Join(heads, "\x00")
	enabled := false
	for _, c := range in.Decl.Checks {
		enabled = enabled || c.RunScope == "landing" && c.Required
	}
	if !enabled {
		return in, nil
	}
	if len(heads) == 0 {
		return in, fmt.Errorf("landing checks require source car evidence")
	}
	for _, head := range heads {
		key := base + ":" + head
		paths, ok := cache.paths[key]
		if !ok {
			ancestor, err := git("merge-base", base, head)
			if err != nil || strings.TrimSpace(ancestor) == "" {
				return in, fmt.Errorf("car merge-base unavailable: %s: %v", head, err)
			}
			raw, err := git(gitx.ChangedPathArgs(strings.TrimSpace(ancestor), head)...)
			if err != nil {
				return in, err
			}
			paths, err = landingChangedPaths(raw)
			if err != nil {
				return in, err
			}
			cache.paths[key] = paths
		}
		in.Paths = append(in.Paths, paths)
	}
	return in, nil
}

// Never trim filename whitespace; -z is the only record delimiter.
func landingChangedPaths(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	if !strings.HasSuffix(raw, "\x00") || len(raw) > 16<<20 {
		return nil, fmt.Errorf("changed-path evidence is incomplete or oversized")
	}
	paths := strings.Split(strings.TrimSuffix(raw, "\x00"), "\x00")
	for _, p := range paths {
		if !validLandingPath(p) {
			return nil, fmt.Errorf("invalid changed-path evidence %q", p)
		}
	}
	return paths, nil
}
