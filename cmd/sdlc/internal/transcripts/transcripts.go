// Package transcripts is the harness-agnostic transcript-source layer behind
// `sdlc actual` (#134). Active-time events come ONLY from agent-CLI transcript
// `.jsonl` files, and each harness (Claude Code, Codex, …) stores them under a
// different on-disk convention. Rather than special-casing each harness inline in
// actual.go (the #134 diagnosis — Codex work read as "no activity" because the
// selection was Claude-only), every harness implements one uniform Harness
// interface, a Registry lists them, and a pure Select aggregates their
// contributions for a set of repo cwds.
//
// Adding a future harness is one new file implementing Harness + one Registry
// entry — actual.go and the activetime engine never change. That extensibility
// IS the issue's stated purpose: transcript discovery as a harness abstraction,
// not a Claude-only path convention.
//
// ARCH-PURE: Select is pure over the injected harness slice; each harness's
// Sources method is the thin IO seam (Stat / WalkDir / ReadFile). The encoders
// and parsers (cwdToClaudeDir, codexCWDFromBytes) are pure and unit-tested
// directly.
package transcripts

// Sources is the engine input a set of harnesses resolves to — exactly the shape
// activetime.Compute consumes (Options.Dirs / Options.Files). Dirs are globbed
// for *.jsonl (Claude's one-dir-per-cwd layout); Files are individual session
// files pre-selected by cwd (Codex's date-sharded layout, where cwd lives inside
// each file rather than in the path).
type Sources struct {
	Dirs  []string
	Files []string
}

// Harness resolves one agent CLI's on-disk transcripts for a set of repo cwds.
// Implementations are the IO seam; the aggregation over them (Select) is pure.
type Harness interface {
	Name() string
	Sources(cwds []string) Sources
}

// Select runs every harness for the given cwds and merges their contributions
// into one Sources, deduping while preserving first-seen order so the engine's
// inputs are stable. Pure over the injected harness slice — production passes
// DefaultHarnesses(); tests pass fakes or temp-rooted real harnesses.
func Select(cwds []string, hs []Harness) Sources {
	var out Sources
	seenDir := map[string]bool{}
	seenFile := map[string]bool{}
	for _, h := range hs {
		s := h.Sources(cwds)
		for _, d := range s.Dirs {
			if !seenDir[d] {
				seenDir[d] = true
				out.Dirs = append(out.Dirs, d)
			}
		}
		for _, f := range s.Files {
			if !seenFile[f] {
				seenFile[f] = true
				out.Files = append(out.Files, f)
			}
		}
	}
	return out
}
