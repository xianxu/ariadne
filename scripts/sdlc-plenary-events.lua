-- A private completion channel for a single native Plenary spec process.
-- The formatter remains the owner of diagnostics; assertions and hooks are untouched.
local M = {}
local state_key = 'sdlc.plenary-events.state'
local max_bytes = 16 * 1024 * 1024

local function array_length(value, limit)
  assert(type(value) == 'table', 'expected array')
  local count = 0
  for key in pairs(value) do
    assert(type(key) == 'number' and key >= 1 and key % 1 == 0, 'invalid array index')
    count = count + 1
    assert(count <= limit, 'array too large')
  end
  for i = 1, count do assert(value[i] ~= nil, 'sparse array') end
  return count
end

local function text(value, limit)
  assert(type(value) == 'string' and #value > 0 and #value <= limit, 'invalid text length')
  -- Validate UTF-8 explicitly: LuaJIT strings and JSON encoders can accept invalid bytes.
  local i = 1
  while i <= #value do
    local b = value:byte(i)
    local length, codepoint
    if b < 128 then length, codepoint = 1, b
    elseif b >= 194 and b <= 223 then length, codepoint = 2, b - 192
    elseif b >= 224 and b <= 239 then length, codepoint = 3, b - 224
    elseif b >= 240 and b <= 244 then length, codepoint = 4, b - 240
    else error('invalid UTF-8') end
    for j = 1, length - 1 do
      local next_byte = value:byte(i + j)
      assert(next_byte and next_byte >= 128 and next_byte <= 191, 'invalid UTF-8 continuation')
      codepoint = codepoint * 64 + next_byte - 128
    end
    assert((length < 2 or codepoint >= 128) and (length < 3 or codepoint >= 2048)
      and (length < 4 or codepoint >= 65536) and codepoint <= 1114111
      and not (codepoint >= 55296 and codepoint <= 57343), 'invalid UTF-8 scalar')
    assert(codepoint >= 32 and not (codepoint >= 127 and codepoint <= 159), 'control in text')
    i = i + length
  end
  return value
end

local function clear(state)
  state.invalid = true
  -- Truncate rather than append: a second invocation must revoke the first proof.
  local file = io.open(state.path, 'wb')
  if file then file:close() else os.remove(state.path) end
end

local function summary(state, result)
  assert(type(result) == 'table', 'expected result table')
  local cases, names = {}, {}
  for _, outcome in ipairs({ 'pass', 'fail' }) do
    local count = array_length(result[outcome], 10000)
    assert(#cases + count <= 10000, 'too many cases')
    for i = 1, count do
      local entry = result[outcome][i]
      assert(type(entry) == 'table', 'expected case table')
      local n = array_length(entry.descriptions, 2049)
      assert(n > 0, 'empty descriptions')
      local bytes = n - 1 -- Joining separators count toward the label byte limit.
      for j = 1, n do
        local description = entry.descriptions[j]
        assert(type(description) == 'string', 'expected description string')
        bytes = bytes + #description
        assert(bytes <= 2048, 'label too large')
      end
      local name = text(table.concat(entry.descriptions, ' '), 2048)
      assert(not names[name], 'duplicate case')
      names[name] = true
      cases[#cases + 1] = { name = name, result = outcome }
    end
  end
  local errors = array_length(result.errs, 10000)
  for i = 1, errors do assert(type(result.errs[i]) == 'table', 'expected error table') end
  local encoded = vim.fn.json_encode({ version = 1, unit = state.unit, complete = true,
    cases = cases, errors = errors })
  assert(#encoded <= max_bytes, 'result channel too large')
  local file = assert(io.open(state.path, 'wb'))
  local wrote = file:write(encoded)
  local closed = file:close()
  assert(wrote and closed, 'result channel write failed')
end

function M.install(spec_path)
  local unit, path = vim.env.SDLC_FLAKE_UNIT, vim.env.SDLC_FLAKE_RESULTS
  if not unit and not path then return end
  local previous = package.loaded[state_key]
  if previous then clear(previous); return end
  if not path or path == '' then return end
  local state = { path = path, unit = unit, invalid = false, seen = false }
  package.loaded[state_key] = state
  clear(state)
  local ok = pcall(function()
    text(unit, 516)
    assert(unit:sub(1, 4) == 'lua:', 'expected Lua unit')
    local relative = text(unit:sub(5), 512)
    assert(relative:sub(1, 1) ~= '/' and not relative:find('\\') and not relative:find('//'), 'invalid unit path')
    for component in relative:gmatch('[^/]+') do
      assert(component ~= '.' and component ~= '..', 'noncanonical unit path')
    end
    text(spec_path, 4096)
    assert(vim.fn.fnamemodify(spec_path, ':p') == vim.fn.fnamemodify(relative, ':p'), 'spec path mismatch')
  end)
  if not ok then return end
  state.invalid = false
  local busted = require('plenary.busted')
  local original = busted.format_results
  if type(original) ~= 'function' then clear(state); return end
  busted.format_results = function(result, ...)
    if state.seen or state.invalid then
      clear(state)
    else
      state.seen = true
      local accepted = pcall(summary, state, result)
      if not accepted then clear(state) end
    end
    return original(result, ...)
  end
end

return M
