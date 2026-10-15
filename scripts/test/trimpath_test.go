package scripttest

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOwnerMakeCommandsUseTrimpath(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  []string
		args []string
		want []string
	}{
		{"configured", nil, nil, []string{"-p=3", "-trimpath"}},
		{"environment", []string{"GOFLAGS=-race -p=2"}, nil, []string{"-race", "-p=2", "-trimpath"}},
		{"already_trimmed", []string{"GOFLAGS=-p=2 -trimpath"}, nil, []string{"-p=2", "-trimpath"}},
		{"empty_environment", []string{"GOFLAGS="}, nil, []string{"-trimpath"}},
		{"explicit_make_override", nil, []string{"GOFLAGS=-p=1"}, []string{"-p=1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"sdlc-build", "weave-build", "test"}, tc.args...)
			rows := runOwnerMakeTrace(t, args, tc.env)
			if len(rows) != 3 {
				t.Fatalf("want two tool builds and the sharder, got %q", rows)
			}
			for i, want := range []string{"go\tbuild -o bin/sdlc ./cmd/sdlc", "go\tbuild -o bin/weave ./cmd/weave", "python3\tscripts/test-shard.py"} {
				parts := strings.Split(rows[i], "\t")
				if len(parts) != 3 || strings.Join(parts[:2], "\t") != want {
					t.Fatalf("unexpected recipe: %q; want %q", rows[i], want)
				}
				if got := strings.Fields(parts[2]); !reflect.DeepEqual(got, tc.want) {
					t.Errorf("%s flags = %q; want %q", parts[1], got, tc.want)
				}
			}
		})
	}
}

func TestRecordedSuiteUsesOwnerTrimpathDefaults(t *testing.T) {
	var decl struct {
		Version int `json:"version"`
		Checks  []struct {
			Name           string   `json:"name"`
			Command        []string `json:"command"`
			Required       bool     `json:"required"`
			Weight         int      `json:"weight"`
			TimeoutMinutes int      `json:"timeout_minutes"`
		} `json:"checks"`
	}
	data, err := os.ReadFile(filepath.Join("..", "..", ".sdlc", "local-checks.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &decl); err != nil {
		t.Fatal(err)
	}
	if decl.Version != 1 || len(decl.Checks) != 1 {
		t.Fatalf("unexpected suite declaration: %s", data)
	}
	c := decl.Checks[0]
	if c.Name != "suite" || !c.Required || c.Weight != 2 || c.TimeoutMinutes != 60 || !reflect.DeepEqual(c.Command, []string{"make", "test"}) {
		t.Fatalf("suite changed the existing command, enforcement or budgets: %s", data)
	}
	rows := runOwnerMakeTrace(t, c.Command[1:], nil)
	if len(rows) != 1 || rows[0] != "python3\tscripts/test-shard.py\t-p=3 -trimpath" {
		t.Fatalf("recorded suite did not inherit trimmed, capped flags: %q", rows)
	}
}

// Run the tracked owner recipes in a private checkout-shaped fixture. The
// executables retain argv/environment observations, without compiling tools or
// starting another suite. Workflow lock behavior has its own integration tests.
func runOwnerMakeTrace(t *testing.T, args, extraEnv []string) []string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"Makefile", "Makefile.local"} {
		data, err := os.ReadFile(filepath.Join("..", "..", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(dir, "fixture-bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"go": `#!/bin/sh
if [ "$1" = env ] && [ "${2:-}" = GOFLAGS ]; then
    printf '%s\n' "${GOFLAGS-${CONFIGURED_GOFLAGS:-}}"
    exit 0
fi
printf 'go\t%s\t%s\n' "$*" "${GOFLAGS-}" >> "${TRACE:?}"
`,
		"python3": "#!/bin/sh\nprintf 'python3\\t%s\\t%s\\n' \"$*\" \"${GOFLAGS-}\" >> \"${TRACE:?}\"\n",
	} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	trace := filepath.Join(dir, "trace")
	ctx, cancel := context.WithTimeout(context.Background(), runnerDeadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, "make", append([]string{"--no-print-directory"}, args...)...)
	cmd.Dir = dir
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		switch key {
		case "GOFLAGS", "MAKEFLAGS", "MFLAGS", "GNUMAKEFLAGS", "MAKELEVEL", "MAKEOVERRIDES", "MAKEFILES", "ARGS", "PATH", "TRACE", "CONFIGURED_GOFLAGS":
			continue
		}
		cmd.Env = append(cmd.Env, value)
	}
	cmd.Env = append(cmd.Env, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "TRACE="+trace, "CONFIGURED_GOFLAGS=-p=3")
	cmd.Env = append(cmd.Env, extraEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("owner recipes: %v\n%s", err, out)
	}
	data, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}
