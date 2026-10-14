// Package reviewprepare defines review preparation declarations and diagnostics.
// It has no filesystem or process dependencies.
package reviewprepare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Config struct {
	Version        int      `json:"version"`
	Command        []string `json:"command"`
	TimeoutMinutes int      `json:"timeout_minutes"`
}

// Parse admits exactly the version-one declaration, with a ten-minute default.
func Parse(raw []byte) (Config, error) {
	c := Config{TimeoutMinutes: 10}
	d := json.NewDecoder(bytes.NewReader(raw))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return Config{}, fmt.Errorf("preparation declaration must be a JSON object")
	}
	seen := map[string]bool{}
	for d.More() {
		token, err = d.Token()
		if err != nil {
			return Config{}, err
		}
		key, ok := token.(string)
		if !ok {
			return Config{}, fmt.Errorf("invalid field")
		}
		if seen[key] {
			return Config{}, fmt.Errorf("duplicate field %q", key)
		}
		seen[key] = true
		var value json.RawMessage
		if err = d.Decode(&value); err != nil {
			return Config{}, err
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return Config{}, fmt.Errorf("%s must not be null", key)
		}
		switch key {
		case "version":
			err = json.Unmarshal(value, &c.Version)
		case "command":
			err = json.Unmarshal(value, &c.Command)
		case "timeout_minutes":
			err = json.Unmarshal(value, &c.TimeoutMinutes)
		default:
			return Config{}, fmt.Errorf("unknown field %q", key)
		}
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", key, err)
		}
	}
	if _, err = d.Token(); err != nil {
		return Config{}, err
	}
	if _, err = d.Token(); err != io.EOF {
		return Config{}, fmt.Errorf("trailing JSON data")
	}
	if c.Version != 1 {
		return Config{}, fmt.Errorf("version must be 1")
	}
	if len(c.Command) == 0 {
		return Config{}, fmt.Errorf("command must contain an executable")
	}
	for i, arg := range c.Command {
		if arg == "" {
			return Config{}, fmt.Errorf("command argument %d must not be empty", i)
		}
	}
	if c.TimeoutMinutes < 1 || c.TimeoutMinutes > 30 {
		return Config{}, fmt.Errorf("timeout_minutes must be between 1 and 30")
	}
	return c, nil
}
