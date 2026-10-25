// Package jsonstrict preserves SDLC's import path while commands share one
// strict JSON authority in cmd/internal/jsonstrict.
package jsonstrict

import (
	"encoding/json"
	shared "github.com/xianxu/ariadne/cmd/internal/jsonstrict"
)

func RejectDuplicateKeys(raw []byte) error       { return shared.RejectDuplicateKeys(raw) }
func Scan(dec *json.Decoder, maxDepth int) error { return shared.Scan(dec, maxDepth) }
