package main

import "encoding/json"

// Version 3 is a capability barrier: a v2 reader must not ignore authoritative
// typed runners and certify a check under Go-only completion semantics.
func isV3LocalChecks(data []byte) bool {
	var header struct {
		Version int `json:"version"`
	}
	return json.Unmarshal(data, &header) == nil && header.Version == 3
}

func parseLocalChecksV3(data []byte) (localCheckDecl, error) {
	return parseModernLocalChecks(data, 3)
}
