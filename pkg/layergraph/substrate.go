package layergraph

import "path/filepath"

// Probe is what a resolver observed at one candidate path.
type Probe int

const (
	Absent   Probe = iota // nothing on disk
	NotLayer              // present, but no construct/base.manifest
	Layer                 // present and ships construct/base.manifest
)

// Resolution is where one substrate row resolves. Path is lexical; the
// boundary canonicalizes it. Local is the owner-relative candidate. Via is
// the primary checkout the row fell back to, "" when it resolved locally.
type Resolution struct {
	Path, Local, Via string
	Probe            Probe
}

// ChooseSubstrate is #230's selection rule, pure. A relative target resolves
// against owner first; only when nothing is there and owner has a fallback
// anchor (its Git primary checkout) is the same relative path tried from the
// anchor. A present local target wins even when invalid, so a broken local
// layer is reported, never bypassed. Absolute targets never fall back.
func ChooseSubstrate(owner, target, primary string, probe func(string) Probe) Resolution {
	if filepath.IsAbs(target) {
		return Resolution{Path: target, Local: target, Probe: probe(target)}
	}
	local := filepath.Join(owner, target)
	r := Resolution{Path: local, Local: local, Probe: probe(local)}
	if r.Probe != Absent || primary == "" {
		return r
	}
	candidate := filepath.Join(primary, target)
	if p := probe(candidate); p != Absent {
		return Resolution{Path: candidate, Local: local, Via: primary, Probe: p}
	}
	return r
}
