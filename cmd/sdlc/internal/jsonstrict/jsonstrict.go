// Package jsonstrict holds the strict-JSON checks a persisted format needs
// beyond encoding/json (ARCH-EVOLVE: duplicate keys stay strict).
package jsonstrict

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// foldKey maps every rune to the least member of its Unicode simple-folding
// orbit, so foldKey(x) == foldKey(y) exactly when strings.EqualFold(x, y): the
// rule encoding/json matches field names by ("ſ" and "s" fill one field).
func foldKey(s string) string {
	return strings.Map(func(r rune) rune {
		least := r
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			least = min(least, f)
		}
		return least
	}, s)
}

// RejectDuplicateKeys refuses a JSON document in which any object, at any
// depth, names a key twice: encoding/json silently keeps the last value. Keys
// equal under case folding are duplicates too, since encoding/json matches
// field names case-insensitively ("command" and "COMMAND" fill one field).
// Only one top-level value is allowed.
func RejectDuplicateKeys(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber() // Opaque future numeric fields need no float64 interpretation.
	if err := scan(dec); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func scan(dec *json.Decoder) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]string{} // case-folded key → as first written
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if prior, ok := seen[foldKey(key)]; ok {
				return fmt.Errorf("duplicate object key %q (as %q)", key, prior)
			}
			seen[foldKey(key)] = key
			if err := scan(dec); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := scan(dec); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	_, err = dec.Token()
	return err
}
