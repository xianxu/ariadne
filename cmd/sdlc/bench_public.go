package main

// The context is public, root-bound authority for this replay's local backend.
// Keep provenance and acceptance inputs out of this type, including extensions.
type benchPublicContext struct {
	Version        int    `json:"version"`
	RunID          string `json:"run_id"`
	WorkRoot       string `json:"work_root"`
	HubRoot        string `json:"hub_root"`
	OriginPath     string `json:"origin_path"`
	OriginIdentity string `json:"origin_identity"`
}

const benchPublicContextPath = ".sdlc/bench.json"
