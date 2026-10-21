package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFleetResumeBoundedRecheck(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	c.resumeErr["pair:6"] = fmt.Errorf("%w: request req-6", errUncertain)
	dir := t.TempDir()
	seen := filepath.Join(dir, "seen")
	bin := filepath.Join(dir, "couch")
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	script := "#!/bin/sh\nif test ! -e " + quote(seen) + "; then\n : > " + quote(seen) + "\n printf '%s\\n' '{\"Code\":\"ok\",\"Actors\":[]}'\nelse\n exec sleep 30\nfi\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	d.actors = execActors{bin: bin, timeout: 500 * time.Millisecond}
	var out strings.Builder
	began := time.Now()
	err := fleetResume(&out, d, cwd, true)
	if !errors.Is(err, errUncertain) || time.Since(began) > 5*time.Second {
		t.Fatalf("unbounded/wrong result: %v\n%s", err, out.String())
	}
	stopped := strings.Index(out.String(), "not attempted: pair:8")
	observed := strings.Index(out.String(), "observation unavailable:")
	guidance := strings.Index(out.String(), "couch --peek pair:6")
	if stopped < 0 || observed < stopped || guidance < 0 || guidance > observed {
		t.Fatalf("recovery output followed the optional stalled probe: %s", out.String())
	}
	if strings.Join(c.attempts(), ",") != "pair:2,pair:6" {
		t.Fatal(c.attempts())
	}
}
