package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// stampFile names the freshness digest written alongside a materialization.
const stampFile = ".source-sha"

// hashSources is the pure freshness primitive (ARCH-PURE): a deterministic
// digest of the merged source set, keyed name→content and folded in name order
// so it's independent of map iteration order. Any change to any source .cue (or
// the set of nouns) changes the digest.
func hashSources(srcs map[string]string) string {
	h := sha256.New()
	names := make([]string, 0, len(srcs))
	for n := range srcs {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(h, "%s\x00%s\x00", n, srcs[n])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// readSources reads each winning .cue path into a name→content map.
func readSources(paths map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(paths))
	for name, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		out[name] = string(b)
	}
	return out, nil
}

// runCheck fails (non-zero) if <output>'s stamp doesn't match the current merged
// source — i.e. the materialization is stale (the source changed upstream and
// this repo hasn't re-woven) or absent (never materialized). The cross-repo
// freshness gate; wire into `make check` / CI.
func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	output := fs.String("output", "", "materialization dir to check against the current source")
	_ = fs.Parse(args)
	if *output == "" {
		return fmt.Errorf("check needs --output <dir>")
	}
	want, err := os.ReadFile(filepath.Join(*output, stampFile))
	if err != nil {
		return fmt.Errorf("no freshness stamp in %s — run `make weave` to materialize the vocabulary: %w", *output, err)
	}
	paths, err := resolveVocab()
	if err != nil {
		return err
	}
	srcs, err := readSources(paths)
	if err != nil {
		return err
	}
	if got := hashSources(srcs); got != string(trimNL(want)) {
		return fmt.Errorf("STALE: %s reflects an older vocabulary source — run `make weave`", *output)
	}
	return nil
}

func trimNL(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	return b
}
