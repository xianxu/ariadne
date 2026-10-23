package scripttest

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func plenaryEventsRun(t *testing.T, body string, enabled bool, library string) ([]byte, string, error) {
	t.Helper()
	nvim, err := exec.LookPath("nvim")
	if err != nil {
		t.Skip("Neovim not installed")
	}
	producer, err := filepath.Abs("../sdlc-plenary-events.lua")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "tests"), 0700); err != nil {
		t.Fatal(err)
	}
	channel := filepath.Join(root, "result.json")
	script := "local ok, err = pcall(function()\nlocal producer = " + strconv.Quote(producer) + "\n" + body + "\nend)\nif not ok then io.stderr:write(tostring(err)); vim.cmd('cquit 2') end"
	entry := filepath.Join(root, "entry.lua")
	if err := os.WriteFile(entry, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, nvim, "--headless", "-u", "NONE", "-i", "NONE", "-n", "-l", entry)
	// -l is unavailable on older supported Neovim releases.
	cmd.Args = []string{nvim, "--headless", "-u", "NONE", "-i", "NONE", "-n", "-c", "lua dofile(" + strconv.Quote(entry) + ")"}
	cmd.Dir = root
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "SDLC_FLAKE_") && !strings.HasPrefix(env, "SDLC_PLENARY_DIR=") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	if enabled {
		cmd.Env = append(cmd.Env, "SDLC_FLAKE_UNIT=lua:tests/sample_spec.lua", "SDLC_FLAKE_RESULTS="+channel)
	}
	cmd.Env = append(cmd.Env, "SDLC_PLENARY_DIR="+library)
	output, runErr := cmd.CombinedOutput()
	data, readErr := os.ReadFile(channel)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatal(readErr)
	}
	return data, string(output), runErr
}

const plenaryFake = `
local calls = 0
local fake = {format_results=function(r) calls = calls + 1; return 'native' end}
package.loaded['plenary.busted'] = fake
local m = dofile(producer)
local function valid() return {pass={{descriptions={'suite','passes'}}},fail={{descriptions={'suite','fails'}}},errs={}} end
`

func TestPlenaryEventsChannel(t *testing.T) {
	data, out, err := plenaryEventsRun(t, plenaryFake+`
m.install(vim.fn.getcwd() .. '/tests/sample_spec.lua')
assert(fake.format_results(valid()) == 'native')
assert(calls == 1)
vim.cmd('qa!')
`, true, "")
	if err != nil {
		t.Fatalf("nvim: %v %s", err, out)
	}
	var got struct {
		Version  int    `json:"version"`
		Unit     string `json:"unit"`
		Complete bool   `json:"complete"`
		Cases    []struct {
			Name   string `json:"name"`
			Result string `json:"result"`
		} `json:"cases"`
		Errors int `json:"errors"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("result %q: %v; output %s", data, err, out)
	}
	if got.Version != 1 || got.Unit != "lua:tests/sample_spec.lua" || !got.Complete || got.Errors != 0 || len(got.Cases) != 2 || got.Cases[0].Name != "suite passes" || got.Cases[1].Result != "fail" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestPlenaryEventsBoundsBeforeJoining(t *testing.T) {
	data, out, err := plenaryEventsRun(t, plenaryFake+`
local original_concat = table.concat
local oversized_join = false
table.concat = function(parts, separator)
  local bytes = math.max(0, #parts - 1) * #(separator or '')
  for _, part in ipairs(parts) do bytes = bytes + #part end
  if bytes > 2048 then oversized_join = true end
  return original_concat(parts, separator)
end
m.install('tests/sample_spec.lua')
local result = valid()
result.pass[1].descriptions = {string.rep('x', 1024 * 1024)}
assert(fake.format_results(result) == 'native')
assert(calls == 1, 'native formatter did not run')
assert(not oversized_join, 'producer concatenated an oversized label')
vim.cmd('qa!')
`, true, "")
	if err != nil {
		t.Fatalf("nvim: %v %s", err, out)
	}
	if len(data) != 0 {
		t.Fatalf("oversized label left proof: %s", data)
	}
}

func TestPlenaryEventsRejectsInvalidProof(t *testing.T) {
	cases := map[string]string{
		"mismatch":          `m.install('tests/other_spec.lua'); fake.format_results(valid())`,
		"traversal":         `m.install('tests/../other_spec.lua'); fake.format_results(valid())`,
		"duplicate summary": `m.install('tests/sample_spec.lua'); fake.format_results(valid()); fake.format_results(valid())`,
		"duplicate install": `m.install('tests/sample_spec.lua'); fake.format_results(valid()); m.install('tests/sample_spec.lua')`,
		"bad after good":    `m.install('tests/sample_spec.lua'); fake.format_results(valid()); fake.format_results({})`,
		"duplicate label":   `m.install('tests/sample_spec.lua'); local r=valid();r.fail[1].descriptions={'suite','passes'}; fake.format_results(r)`,
		"control":           `m.install('tests/sample_spec.lua'); local r=valid();r.pass[1].descriptions={'bad\nlabel'};fake.format_results(r)`,
		"long label":        `m.install('tests/sample_spec.lua');local r=valid();r.pass[1].descriptions={string.rep('x',2049)};fake.format_results(r)`,
		"invalid utf8":      `m.install('tests/sample_spec.lua');local r=valid();r.pass[1].descriptions={string.char(255)};fake.format_results(r)`,
		"sparse":            `m.install('tests/sample_spec.lua');local r=valid();r.pass[3]=r.pass[1];fake.format_results(r)`,
		"too many":          `m.install('tests/sample_spec.lua');local r=valid();for i=1,10001 do r.pass[i]={descriptions={'case '..i}} end;fake.format_results(r)`,
		"module reload":     `m.install('tests/sample_spec.lua');fake.format_results(valid());dofile(producer).install('tests/sample_spec.lua')`,
		"payload bound":     `m.install('tests/sample_spec.lua');local r=valid();r.fail={};for i=1,10000 do r.pass[i]={descriptions={tostring(i)..string.rep('x',2000)}} end;fake.format_results(r)`,
		"C1 control":        `m.install('tests/sample_spec.lua');local r=valid();r.pass[1].descriptions={'bad'..string.char(194,133)};fake.format_results(r)`,
		"stdout impostor":   `print('{"version":1,"complete":true,"cases":[]}')`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			data, out, err := plenaryEventsRun(t, plenaryFake+body+"\nvim.cmd('qa!')", true, "")
			if err != nil {
				t.Fatalf("nvim: %v %s", err, out)
			}
			if len(data) != 0 {
				t.Fatalf("invalid scenario left proof: %s", data)
			}
		})
	}
	t.Run("no env", func(t *testing.T) {
		data, out, err := plenaryEventsRun(t, plenaryFake+`m.install('tests/sample_spec.lua');assert(fake.format_results(valid())=='native');assert(calls==1);vim.cmd('qa!')`, false, "")
		if err != nil || len(data) != 0 {
			t.Fatalf("data=%s err=%v output=%s", data, err, out)
		}
	})
}

func TestPlenaryEventsRealConformance(t *testing.T) {
	library := os.Getenv("SDLC_PLENARY_DIR")
	if library == "" {
		t.Skip("set SDLC_PLENARY_DIR for installed-library conformance")
	}
	fixture := `describe('outer', function()
 before_each(function() local f=assert(io.open('hooks','a'));f:write('B');f:close() end)
 after_each(function() local f=assert(io.open('hooks','a'));f:write('A');f:close() end)
 it('passes',function() assert.equals(1,1) end)
 describe('nested',function() it('fails',function() assert.equals(1,2) end) end)
 pending('skipped',function() error('must not run') end)
end)`
	for _, scenario := range []string{"native", "producer", "mismatch", "hook error"} {
		t.Run(scenario, func(t *testing.T) {
			spec := fixture
			if scenario == "hook error" {
				spec = `describe('outer',function() before_each(function() error('hook failure') end); it('never',function() end) end)`
			}
			body := `vim.opt.runtimepath:append(vim.env.SDLC_PLENARY_DIR)
local f=assert(io.open('tests/sample_spec.lua','w')); f:write(` + strconv.Quote(spec) + `);f:close()
vim.api.nvim_create_autocmd('VimLeavePre',{callback=function() local f=io.open('hooks','r');if f then io.stdout:write('HOOKS:'..f:read('*a'));f:close() end end})
`
			if scenario == "producer" || scenario == "hook error" {
				body += `dofile(producer).install('tests/sample_spec.lua')` + "\n"
			}
			if scenario == "mismatch" {
				body += `dofile(producer).install('tests/other_spec.lua')` + "\n"
			}
			body += `require('plenary.busted').run('tests/sample_spec.lua')`
			data, out, err := plenaryEventsRun(t, body, true, library)
			wantExit := 1
			if scenario == "hook error" {
				wantExit = 2
			}
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != wantExit {
				t.Fatalf("native exit=%v; want %d; output=%s", err, wantExit, out)
			}
			if scenario != "hook error" && (!strings.Contains(out, "HOOKS:BABA") || !strings.Contains(out, "outer passes") || !strings.Contains(out, "outer nested fails")) {
				t.Fatalf("native behavior changed: %s", out)
			}
			if scenario == "native" || scenario == "mismatch" {
				if len(data) != 0 {
					t.Fatalf("unexpected proof %s", data)
				}
				return
			}
			var got struct {
				Cases []struct {
					Name   string `json:"name"`
					Result string `json:"result"`
				} `json:"cases"`
				Errors int `json:"errors"`
			}
			if e := json.Unmarshal(data, &got); e != nil {
				t.Fatalf("proof=%s err=%v output=%s", data, e, out)
			}
			if scenario == "hook error" {
				if got.Errors != 1 || len(got.Cases) != 0 {
					t.Fatalf("hook proof=%s", data)
				}
			} else if got.Errors != 0 || len(got.Cases) != 2 || got.Cases[0].Name != "outer passes" || got.Cases[0].Result != "pass" || got.Cases[1].Name != "outer nested fails" || got.Cases[1].Result != "fail" {
				t.Fatalf("proof=%s", data)
			}
		})
	}
}
