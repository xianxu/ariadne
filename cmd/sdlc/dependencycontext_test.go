package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the installed-process seams, rather than replacing weaveRefresh: an
// SDLC invocation must not acquire permission to install machine dependencies.
func TestSDLCWeaveDependenciesCheckOnly(t *testing.T) {
	for _, inherited := range []string{"", "0"} {
		t.Run("inherited="+inherited, func(t *testing.T) {
			const marker = "WEAVE_DEPENDENCIES_CHECK_ONLY"
			t.Setenv(marker, inherited)
			if inherited == "" {
				if err := os.Unsetenv(marker); err != nil {
					t.Fatal(err)
				}
			}
			parent := canonRoot(t.TempDir())
			owner, leaf := filepath.Join(parent, "owner"), filepath.Join(parent, "leaf")
			if err := os.MkdirAll(owner, 0755); err != nil {
				t.Fatal(err)
			}
			propagationRepo(t, owner)
			propagationRepo(t, leaf)
			propagationWrite(t, leaf, "construct/deps", "substrate ../owner\n")
			git(t, leaf, "add", "-A")
			git(t, leaf, "commit", "-qm", "initial")
			bin := filepath.Join(parent, "tools")
			propagationWrite(t, bin, "weave", `#!/bin/sh
printf '%s|%s|%s\n' "$*" "$PWD" "${WEAVE_DEPENDENCIES_CHECK_ONLY-unset}" >> "$DEPENDENCY_EVENTS"
printf 'stdout:%s\n' "$1"
printf 'stderr:%s\n' "$1" >&2
case "$1" in
refresh) test "$2" = --planning ;;
settings) test "$#" = 1 ;;
compile) printf ready > "$DEPENDENCY_READY" ;;
verify-complete) test -f "$DEPENDENCY_READY" ;;
*) exit 10 ;;
esac
`)
			if err := os.Chmod(filepath.Join(bin, "weave"), 0755); err != nil {
				t.Fatal(err)
			}
			events := filepath.Join(parent, "events")
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("DEPENDENCY_EVENTS", events)
			t.Setenv("DEPENDENCY_READY", filepath.Join(parent, "ready"))
			var out bytes.Buffer
			if err := runInstalledWeave(context.Background(), owner, &out); err != nil {
				t.Fatal(err)
			}
			if err := runInstalledWeaveSettings(context.Background(), owner, &out); err != nil {
				t.Fatal(err)
			}
			if err := runPropagateBase(context.Background(), owner, "test#427", false, &out); err != nil {
				t.Fatalf("%v\n%s", err, &out)
			}
			data, err := os.ReadFile(events)
			if err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("refresh --planning|%s|1\nsettings|%s|1\ncompile|%s|1\nverify-complete|%s|1\n", owner, owner, leaf, leaf)
			if string(data) != want {
				t.Errorf("installed weave calls = %q, want %q", data, want)
			}
			for _, command := range []string{"refresh", "settings", "compile", "verify-complete"} {
				for _, stream := range []string{"stdout", "stderr"} {
					if !strings.Contains(out.String(), stream+":"+command+"\n") {
						t.Errorf("lost %s for %s: %s", stream, command, &out)
					}
				}
			}
			if value, present := os.LookupEnv(marker); value != inherited || present != (inherited != "") {
				t.Errorf("parent marker changed: value=%q present=%v", value, present)
			}
		})
	}
}
