# Native reviewer scratch conformance for #571

Text-only host evidence, retained for the boundary reviewer. Metadata names the exact production and probe blobs; the raw combined output below is unchanged.

```text
Ariadne #571 native Parley review scratch conformance
Ariadne production base: 41aa096de14b015ac506d5152c12ca45b4e9e458
confine.go blob: 6d3f463a8ada2b345a85c74bd5a24b9ff83038d5
checkout.go blob: 8d2a3a1f95732bce3cd7091dde3fca5953ddfb0c
reviewtmp_live_test.go blob: e31b8389a494aefbf1e0d7ad81c408ed2371801f
Parley pinned HEAD: 92bf8cbca1be44942bbfc078011cee7b89084812
Go: go1.27.3 darwin/arm64
Command: env WF_LIVE_PARLEY_REPO=/Volumes/TestScratch/tmp/r422 WF_LIVE_PARLEY_ROOT=/Volumes/TestScratch/p.qhfkKX WF_LIVE_PLENARY=/Users/xianxu/.local/share/nvim/lazy/plenary.nvim PARLEY_TEST_NVIM=/opt/homebrew/opt/neovim/bin/nvim go test ./cmd/sdlc -run ^TestReviewerTmpLiveParley$ -count=1 -v
Raw output SHA256: bb16d91b3f5376c9c23f7352bc47035f9aea04700441e324d0f17519dae4d0ff
Exit: 0

Raw combined output follows:
=== RUN   TestReviewerTmpLiveParley
=== RUN   TestReviewerTmpLiveParley/claude
    reviewtmp_live_test.go:94: Parley HEAD=92bf8cbca1be44942bbfc078011cee7b89084812 fence=seatbelt agent=claude
            … still working — 30s elapsed via claude (pid 19680; inspect: ps -p 19680)
            … still working — 1m0s elapsed via claude (pid 19680; inspect: ps -p 19680)
            … still working — 1m30s elapsed via claude (pid 19680; inspect: ps -p 19680)
        effective TMPDIR=/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp checkout=/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0
        effective tempname=/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/nvim.xianxu/izYbvw/0effective tempname=/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/p/tmp/nvim.xianxu/ahExBs/0owned process census verified
        Running tests/integration/zero_config_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/zero_config_spec.lua	
        [32mSuccess[0m	||	zero-config setup boots with setup({}) and no keys, without warnings, and creates a chat	
        
        [32mSuccess[0m	||	zero-config setup keeps keyless provider defaults when the user supplies other keys	
        
        [32mSuccess[0m	||	zero-config setup drops a default key set to false	
        
        [32mSuccess[0m	||	zero-config setup keeps a provider disabled by an empty table even when a default key exists	
        
        [32mSuccess[0m	||	zero-config setup reports a missing key when it is first used, not at startup	
        
        [32mSuccess[0m	||	zero-config setup detects the nearest marked project when chat_dir is set	
        
        [32mSuccess[0m	||	zero-config setup stays out of repo mode with repo_root = false	
        
        [32mSuccess[0m	||	zero-config setup stays out of repo mode with PARLEY_REPO_MODE=0	
        
        [32mSuccess[0m	||	zero-config setup lets an explicit repo_root win over PARLEY_REPO_MODE=0	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/chat_dirs_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/chat_dirs_spec.lua	
        [32mSuccess[0m	||	chat roots: derivation and persistence contract get_chat_roots returns config.chat_dir as the only root in plain mode	
        
        [32mSuccess[0m	||	chat roots: derivation and persistence contract does not write chat_roots / chat_dirs to state.json	
        
        [32mSuccess[0m	||	chat roots: derivation and persistence contract ignores chat_roots / chat_dirs when loading from state.json	
        
        [32mSuccess[0m	||	chat roots: derivation and persistence contract removes chat_roots / chat_dirs from state.json on next persist	
        
        	
        [32mSuccess: [0m	4	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/chat_finder_logic_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/chat_finder_logic_spec.lua	
        [32mSuccess[0m	||	ChatFinder logic does not borrow a newer picker's insertion target for an older callback	
        
        Parley.nvim: Inserted chat reference: 2026-09-15.12-00-00.000_choice.md[32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) config <C-d> confirmation y	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) config <C-d> confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) config <C-g>D confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) registry fallback <C-d> confirmation y	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) registry fallback <C-d> confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) registry fallback <C-g>D confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) starter <C-d> confirmation y	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) starter <C-d> confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) starter <C-g>D confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic Group A: Timestamp parsing from filename A1: parses timestamp from valid YYYY-MM-DD-HH-MM-SS filename	
        
        [32mSuccess[0m	||	ChatFinder logic Group A: Timestamp parsing from filename A2: returns nil for non-timestamp filename	
        
        [32mSuccess[0m	||	ChatFinder logic Group A: Timestamp parsing from filename A3: handles YYYY-MM-DD format (without time)	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B1: resolves configured presets and keeps all as the final state	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B2: uses configured default month when opening in filtered mode	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B3: cycles left and right across presets and all	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B4: files within recency cutoff are included	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B5: files older than recency cutoff are excluded	
        
        [32mSuccess[0m	||	ChatFinder logic Group C: Entry sorting by timestamp C1: entries are sorted newest first	
        
        [32mSuccess[0m	||	ChatFinder logic Group C: Entry sorting by timestamp C2: equal timestamps maintain stable order	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D1: formats display with filename, topic, date, no tags	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D2: formats display with tags	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D3: formats searchable ordinal with filename, topic, tags	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D4: ordinal with no tags has no trailing space	
        
        [32mSuccess[0m	||	ChatFinder logic Group E: Delete confirmation reopen behavior reopens ChatFinder on Esc during delete confirmation	
        
        [32mSuccess[0m	||	ChatFinder logic Group E: Delete confirmation reopen behavior reopens on the item that moves into the deleted visual row	
        
        [32mSuccess[0m	||	ChatFinder logic Group E: Delete confirmation reopen behavior opens delete confirmation from the source window	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings opens a loading picker before asynchronous acquisition starts	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings animates the real picker while Chat acquisition remains delayed	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings cancels loading acquisition before a recency mapping reopens Chat	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings joins a matching retained prewarm instead of starting a second scan	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings cancels picker-owned acquisition and ignores late root delivery	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps recency outside the exact discovery fingerprint	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings settles all-absent optional roots as a successful empty picker	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps a bounded error status when every Chat root fails	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps successful Chat records and warns once for record failures	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings replaces a mismatched retained prewarm instead of joining it	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings reuses unchanged cached metadata without requesting another read	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings prunes cache only for roots that enumerate successfully	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings opens with the active recency label and all four cycle mappings	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings <Tab> cycles left (aliases <C-a>), <S-Tab> right (aliases <C-s>) (#159)	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings prefers restoring selection by item value when reopening after delete	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings includes chat files from secondary chat roots	
        
        Parley.nvim: Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-secondary-19571/2026-02-04-10-00-00-move-me.md
        Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-secondary-19571/2026-02-04-10-00-00-move-me.md[32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps a moved chat visible when moving it from ChatFinder	
        
        
        Parley.nvim: Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-[archive]-2a3e52/2026-02-04-10-00-00-escaped-root.md
        Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-[archive]-2a3e52/2026-02-04-10-00-00-escaped-root.md[32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings finds moved chats in roots whose paths contain glob characters	
        
        
        Parley.nvim: Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-tilde-a54086/2026-02-04-10-00-00-tilde-root.md
        Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/541c515baf0670b0.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-tilde-a54086/2026-02-04-10-00-00-tilde-root.md[32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings reopens on moved chats when the destination root was configured with a tilde path	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings passes filename, tags, and topic as search text for matcher ranking	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps canonical facet behavior across toggles and rediscovery	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings includes extra-root labels in finder search text	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings includes empty braces in primary-root search text	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings passes preserved tag filters back into the picker prompt	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings preserves only bracket-tag and brace-label fragments from finder queries	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings matches extra-root labels when the query uses brace syntax	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings preserves empty brace filters for the primary root	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings falls back to the newer visual neighbor when deleting the oldest item	
        
        [32mSuccess[0m	||	ChatFinder logic default_sticky_query_for_repo_mode returns {} in plain repo mode	
        
        [32mSuccess[0m	||	ChatFinder logic default_sticky_query_for_repo_mode returns nil when not in repo mode	
        
        [32mSuccess[0m	||	ChatFinder logic default_sticky_query_for_repo_mode returns nil in super-repo mode (don't override aggregation intent)	
        
        	
        [32mSuccess: [0m	58	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        
        Parley.nvim: Chat finder: partial scan (1 roots, 0 files failed)Running tests/unit/helper_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/helper_spec.lua	
        [32mSuccess[0m	||	helper.uuid returns a string	
        
        [32mSuccess[0m	||	helper.uuid matches the expected xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx pattern	
        
        [32mSuccess[0m	||	helper.uuid returns unique values on successive calls	
        
        [32mSuccess[0m	||	helper.starts_with returns true when string starts with prefix	
        
        [32mSuccess[0m	||	helper.starts_with returns true for empty prefix (always matches)	
        
        [32mSuccess[0m	||	helper.starts_with returns false when string does not start with prefix	
        
        [32mSuccess[0m	||	helper.starts_with returns false for partial prefix mismatch	
        
        [32mSuccess[0m	||	helper.starts_with returns true when string equals prefix exactly	
        
        [32mSuccess[0m	||	helper.starts_with returns false when prefix is longer than string	
        
        [32mSuccess[0m	||	helper.ends_with returns true when string ends with suffix	
        
        [32mSuccess[0m	||	helper.ends_with returns true for empty suffix (always matches)	
        
        [32mSuccess[0m	||	helper.ends_with returns false when string does not end with suffix	
        
        [32mSuccess[0m	||	helper.ends_with returns true when string equals suffix exactly	
        
        [32mSuccess[0m	||	helper.ends_with returns false when suffix is longer than string	
        
        [32mSuccess[0m	||	helper.ends_with works with file extension patterns	
        
        [32mSuccess[0m	||	helper.last_content_line returns 0 when buffer has only whitespace	
        
        [32mSuccess[0m	||	helper.last_content_line returns the last non-whitespace line with trailing blanks	
        
        [32mSuccess[0m	||	helper.last_content_line returns the last line when the tail has content	
        
        [32mSuccess[0m	||	helper.last_content_line invalidates cache when buffer content changes	
        
        	
        [32mSuccess: [0m	19	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/provider_params_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/provider_params_spec.lua	
        [32mSuccess[0m	||	get_schema returns base schema for known provider	
        
        [32mSuccess[0m	||	get_schema returns empty params for unknown provider	
        
        [32mSuccess[0m	||	get_schema returns openai-compatible schema for cliproxyapi provider	
        
        [32mSuccess[0m	||	get_schema applies o-series override: removes temperature/top_p/max_tokens, adds reasoning_effort	
        
        [32mSuccess[0m	||	get_schema applies gpt-5 override: max_tokens gets api_name=max_completion_tokens	
        
        [32mSuccess[0m	||	get_schema applies gpt-5 override to gpt-5.2	
        
        [32mSuccess[0m	||	get_schema applies Claude Sonnet 4.6 exclusive group override	
        
        [32mSuccess[0m	||	get_schema does NOT apply Claude 4.6 override to Claude Sonnet 4 (non-4.6)	
        
        [32mSuccess[0m	||	resolve_params returns empty table for string model_config	
        
        [32mSuccess[0m	||	resolve_params uses user-provided values	
        
        [32mSuccess[0m	||	resolve_params applies defaults only when user omits param	
        
        [32mSuccess[0m	||	resolve_params does not send params with nil default when user doesn't set them	
        
        [32mSuccess[0m	||	resolve_params clamps values to declared range	
        
        [32mSuccess[0m	||	resolve_params uses api_name for Gemini params	
        
        [32mSuccess[0m	||	resolve_params omits unsupported params for o-series models	
        
        [32mSuccess[0m	||	resolve_params uses max_completion_tokens api_name for gpt-5	
        
        [32mSuccess[0m	||	resolve_params uses max_completion_tokens api_name for gpt-5.2	
        
        [32mSuccess[0m	||	resolve_params applies gpt-5 param mapping for cliproxyapi provider	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns allowed_callers={direct} for anthropic haiku 4.5 web_search	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns allowed_callers={direct} for anthropic haiku 4.5 web_fetch	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns empty overrides for non-haiku models	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns empty overrides for unknown provider/tool	
        
        [32mSuccess[0m	||	validate_agent returns no errors for valid openai agent	
        
        [32mSuccess[0m	||	validate_agent errors when provider is missing	
        
        [32mSuccess[0m	||	validate_agent skips validation for string model	
        
        [32mSuccess[0m	||	validate_agent warns about unknown parameters	
        
        [32mSuccess[0m	||	validate_agent errors on exclusive group at_most_one violation	
        
        [32mSuccess[0m	||	validate_agent no error when only one of exclusive group is set	
        
        [32mSuccess[0m	||	validate_agent no error when neither of exclusive group is set	
        
        [32mSuccess[0m	||	validate_agent errors on require_one when none set	
        
        [32mSuccess[0m	||	validate_agent warns on range violation	
        
        [32mSuccess[0m	||	validate_agent does not warn when value is within range	
        
        [32mSuccess[0m	||	max_tokens default for Claude models (#228) is 64000 for Claude, on every transport that carries it	
        
        [32mSuccess[0m	||	max_tokens default for Claude models (#228) leaves non-Claude models alone, including behind the same proxy	
        
        [32mSuccess[0m	||	max_tokens default for Claude models (#228) does not disturb the gpt-5 rename to max_completion_tokens	
        
        	
        [32mSuccess: [0m	35	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/pure_functions_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/pure_functions_spec.lua	
        [32mSuccess[0m	||	find_exchange_at_line returns (index, 'question') when line is in question range	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (index, 'answer') when line is in answer range	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (nil, nil) when line is in gap between exchanges	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (nil, nil) when line is before all exchanges	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (nil, nil) when line is after all exchanges	
        
        [32mSuccess[0m	||	find_exchange_at_line finds correct index in multi-exchange chat (not always first)	
        
        [32mSuccess[0m	||	find_exchange_at_line returns exchange when line is in margin after question with no answer	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact question boundary (line_start)	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact question boundary (line_end)	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact answer boundary (line_start)	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact answer boundary (line_end)	
        
        [32mSuccess[0m	||	simple_markdown_to_html escapes HTML special characters	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts fenced code block with language to <pre><code> with class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts fenced code block without language to <pre><code>	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts inline code to <code> with inline-code class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts # H1 to <h1> with main-header class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts ## H2 to <h2> with section-header class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts ### H3 to <h3> with sub-header class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts **bold** to <strong> with bold-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts __bold__ to <strong> with bold-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts *italic* to <em> with italic-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts _italic_ to <em> with italic-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts - list item to <li> with list-item class	
        
        [32mSuccess[0m	||	simple_markdown_to_html wraps list items in <ul> with bullet-list class	
        
        [32mSuccess[0m	||	simple_markdown_to_html blockquote conversion is broken (escapes > before matching)	
        
        [32mSuccess[0m	||	simple_markdown_to_html wraps content in paragraph tags	
        
        [32mSuccess[0m	||	simple_markdown_to_html splits multiple paragraphs on double newlines	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns plain agent name when web_search is disabled	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎] for openai model with search_model when web_search is enabled	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎?] for openai model without search_model when web_search is enabled	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎?] for cliproxyapi when strategy is none	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎] for cliproxyapi claude model in anthropic_tools_route	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search uses per-model strategy override for cliproxyapi labels	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎?] for cliproxyapi openai_search_model without search_model	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎] for cliproxyapi openai_tools_route without search_model	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🔧🌎] when tool-enabled agent also has web_search supported	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🔧] when tool-enabled agent has web_search disabled	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) closes an INDENTED fence — the shape the prompt convention produces	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) an unterminated fence does not consume the following turns	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) a code block after prose is NOT nested inside a paragraph	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) still converts a well-formed flush-left block	
        
        	
        [32mSuccess: [0m	41	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        owned process census verified
        clean: no surviving test processes
        owned process census verified
        Running tests/arch/arrival_budget_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/arch/arrival_budget_spec.lua	
        [32mSuccess[0m	||	arch: predicate waits use the arrival budget or say why not (#374) finds the waits it checks	
        
        [32mSuccess[0m	||	arch: predicate waits use the arrival budget or say why not (#374) bounds each by the budget, a named deadline, or a stated reason	
        
        	
        [32mSuccess: [0m	2	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/destructive_recipe_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/arch/destructive_recipe_spec.lua	
        [32mSuccess[0m	||	arch: test-clean-env only deletes what the harness built removes exactly the harness-built leaves, never the caller-supplied root	
        
        [32mSuccess[0m	||	arch: test-clean-env only deletes what the harness built stays word-tight when the checkout path contains a space	
        
        [32mSuccess[0m	||	arch: test-clean-env only deletes what the harness built wipes before the suite starts, not partway through it	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/embedded_rpc_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/arch/embedded_rpc_spec.lua	
        [32mSuccess[0m	||	arch: embedded nvim requests go through the bounded helper (#401) scans the suite, including the helper's own users	
        
        [32mSuccess[0m	||	arch: embedded nvim requests go through the bounded helper (#401) starts and drives no child nvim outside the helper	
        
        	
        [32mSuccess: [0m	2	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/fixture_lifecycle_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/arch/fixture_lifecycle_spec.lua	
        [32mSuccess[0m	||	arch: fixture process lifecycle no spec or helper spawns outside the fixture seam	
        
        [32mSuccess[0m	||	arch: fixture process lifecycle every executable fixture either watches its parent or declares why it cannot block	
        
        [32mSuccess[0m	||	arch: fixture process lifecycle both watchdogs recognize a parent already reparented to init	
        
        [32mSuccess[0m	||	arch: fixture process lifecycle TOOLING.md's remedy uses ps, and warns about pgrep	
        
        	
        [32mSuccess: [0m	4	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/scratch_placement_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/arch/scratch_placement_spec.lua	
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree nvim's temp directory is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree nvim's swap directory is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $XDG_DATA_HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $XDG_STATE_HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $XDG_CACHE_HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree scratch has no project-marked ancestor	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree the ancestor guard sees a marker in $HOME itself	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree no script or test roots a temp dir at a fixed /tmp	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree the /tmp guard sees every spelling it refuses, and passes the fallbacks	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard refuses a root under a marked tree and creates nothing there	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard refuses a root that carries the marker itself	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard refuses a root inside the checkout even without a marker	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard accepts an unmarked root outside the checkout, existing or not	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard is wired into PREP_TEST_ENV with the configured marker	
        
        	
        [32mSuccess: [0m	15	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/superseded_comment_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/arch/superseded_comment_spec.lua	
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces fires on a stacked restatement (planted false negative)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces stays silent when an annotation restates its own prose (planted false positive)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces stays silent on two paragraphs that merely share a topic	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces fires on an @param the signature does not take (planted false negative)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces fires on a blank line severing the block from its function	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces accepts an underscore-marked unused argument (planted false positive)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces accepts a block that documents its own function	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces every @param block in the tree documents the function it precedes	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces no comment run in the tree stacks a paragraph on its replacement	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/chat_progress_process_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/chat_progress_process_spec.lua	
        [32mSuccess[0m	||	chat progress managed process lifecycle shows pending decoration then writes admitted text without a minimum-visible delay	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle commits admitted partial output before the broken failure notification	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle commits admitted partial output before the unauthorized failure notification	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle commits admitted partial output before the http500 failure notification	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle cleans the admitted session when its provider secret is missing	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle cleans the admitted session when its exact transport admission key is occupied	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle cleans the admitted session when process spawn is rejected	
        
        	
        [32mSuccess: [0m	7	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/discovery_registry_fixture_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/discovery_registry_fixture_spec.lua	
        [32mSuccess[0m	||	default discovery registry fixture ordinary Markdown discovery gets positive absence without reaching the sentinel	
        
        [32mSuccess[0m	||	default discovery registry fixture keeps direct calls and unmocked mutations behind the strict sentinel	
        
        [32mSuccess[0m	||	default discovery registry fixture lets an explicit stateful fake win over the empty discovery fixture	
        
        [32mSuccess[0m	||	default discovery registry fixture installs the same empty registry boundary in clean embedded editors	
        
        [32mSuccess[0m	||	default discovery registry fixture installs once without replacing the first fixture or wrapping it twice	
        
        	
        [32mSuccess: [0m	5	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/embedded_nvim_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/embedded_nvim_spec.lua	
        [32mSuccess[0m	||	embedded nvim helper runs Lua in the child and returns its value	
        
        [32mSuccess[0m	||	embedded nvim helper fails a request the child never answers, naming it, and kills the child	
        
        [32mSuccess[0m	||	embedded nvim helper joins an intentionally killed child before removing its private fixture	
        
        [32mSuccess[0m	||	embedded nvim helper retains and names its private fixture when the child exit is not confirmed	
        
        [32mSuccess[0m	||	embedded nvim helper retains a replaced fixture using the identity captured before child actions	
        
        [32mSuccess[0m	||	embedded nvim helper bounds a request at half the spec's idle budget, which the runner passes	
        
        [32mSuccess[0m	||	embedded nvim helper stops the child when its setup fails, and raises	
        
        [32mSuccess[0m	||	embedded nvim helper names a wedged child before the runner's idle kill	
        
        	
        [32mSuccess: [0m	8	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/fixture_ready_publish_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/fixture_ready_publish_spec.lua	
        [32mSuccess[0m	||	fake SSE server readiness publish never exposes the ready path in an incomplete state	
        
        	
        [32mSuccess: [0m	1	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/fixture_reaping_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/fixture_reaping_spec.lua	
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone a spec-child Neovim does	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone removes its query directory from the watchdog callback without following links	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_cliproxy fixture does	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_github_releases fixture does	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_sse_server fixture does, even with no request to answer	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_cliproxy login path does, though it never binds a port	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_sips slow mode does, though it never binds a port	
        
        [32mSuccess[0m	||	#220 census conformance, against the real ps reads a real ps, or says why it cannot	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts reap() kills what spawn() started, and forgets it	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts reap({since = mark}) spares what was started before the mark	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts forgets a process that exited on its own, so live() cannot over-count	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts reaps with the signal the caller asks for	
        
        	
        [32mSuccess: [0m	12	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/fresh_clone_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/fresh_clone_spec.lua	
        [32mSuccess[0m	||	standalone tracked archive loads and creates chats with intact and unavailable optional vocabulary	
        
        [32mSuccess[0m	||	standalone tracked archive rejects an escaping archive symlink before loading plugin code	
        
        [32mSuccess[0m	||	standalone tracked archive accepts a self-contained archive symlink	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/portable_make_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/portable_make_spec.lua	
        [32mSuccess[0m	||	portable test dependency configuration shows product test commands without a maintainer overlay	
        
        [32mSuccess[0m	||	portable test dependency configuration rejects a missing Plenary dependency with actionable advice	
        
        [32mSuccess[0m	||	portable test dependency configuration accepts an explicit dependency outside the default profile	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/spec_runner_native_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/spec_runner_native_spec.lua	
        [32mSuccess[0m	||	owned headless spec execution executes real Plenary assertions and preserves ordinary output	
        
        [32mSuccess[0m	||	owned headless spec execution retains an actual failed assertion and nonzero exit	
        
        [32mSuccess[0m	||	owned headless spec execution kills and reaps a stalled case after completed work	
        
        [32mSuccess[0m	||	owned headless spec execution enforces the absolute ceiling even while cases keep completing	
        
        [32mSuccess[0m	||	owned headless spec execution does not let arbitrary bootstrap output keep a child alive	
        
        [32mSuccess[0m	||	owned headless spec execution gives the absolute deadline precedence at controlled capped load	
        
        [32mSuccess[0m	||	owned headless spec execution includes after_each teardown in a case's progress boundary	
        
        [32mSuccess[0m	||	owned headless spec execution ends the spec's whole process group at the idle deadline	
        
        [32mSuccess[0m	||	owned headless spec execution ends a spec's vim.system grandchild at the idle deadline	
        
        [32mSuccess[0m	||	owned headless spec execution takes its spec's process group with it when the runner gets sigterm	
        
        [32mSuccess[0m	||	owned headless spec execution takes its spec's process group with it when the runner gets sigint	
        
        [32mSuccess[0m	||	owned headless spec execution signals no recorded job whose pid now names another process	
        
        	
        [32mSuccess: [0m	12	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/test_isolation_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/test_isolation_spec.lua	
        [32mSuccess[0m	||	test isolation guard refuses to run before any command when the env carries no isolation marker	
        
        [32mSuccess[0m	||	test isolation guard refuses when HOME or an XDG root lies outside the marked root	
        
        [32mSuccess[0m	||	test isolation guard fails closed when the isolation check itself cannot load	
        
        [32mSuccess[0m	||	test isolation guard runs normally when every root is inside the marked root	
        
        [32mSuccess[0m	||	the allowlist environment (#365) this spec process sees no ambient Couch, Pair or terminal-owner name	
        
        [32mSuccess[0m	||	the allowlist environment (#365) resolves a bare couch and pair to the sentinel, which records the call and refuses	
        
        [32mSuccess[0m	||	the allowlist environment (#365) the launcher passes exactly the allowlist, whatever the caller exports	
        
        [32mSuccess[0m	||	the allowlist environment (#365) lets each live flag carry only its own check's names	
        
        [32mSuccess[0m	||	the allowlist environment (#365) passes a credential only when the target names it in PARLEY_TEST_ENV_PASS	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/test_process_census_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/integration/test_process_census_spec.lua	
        [32mSuccess[0m	||	retained embedded child census receives its retained checkout context and a recorded native spec birth	
        
        [32mSuccess[0m	||	retained embedded child census retains unknown when the launcher dies inside birth publication	
        
        [32mSuccess[0m	||	retained embedded child census installs one job owner and preserves on_exit for children already reaped at publication	
        
        [32mSuccess[0m	||	retained embedded child census records jobs from clean embedded children that never load minimal_init	
        
        [32mSuccess[0m	||	retained embedded child census cannot call an ordinary spec exit reaped while its recorded child lives	
        
        [32mSuccess[0m	||	retained embedded child census the independent terminal census detects and joins a TERM-resistant orphan	
        
        	
        [32mSuccess: [0m	6	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/await_helper_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/await_helper_spec.lua	
        [32mSuccess[0m	||	await.until_progress (#293) settles past the stall window while progress keeps moving	
        
        [32mSuccess[0m	||	await.until_progress (#293) fails as stalled one window after progress stops	
        
        [32mSuccess[0m	||	await.until_progress (#293) fails at the ceiling even while progress keeps moving	
        
        [32mSuccess[0m	||	await.until_progress (#293) a wait with no progress at all gets exactly the stall window	
        
        	
        [32mSuccess: [0m	4	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/golden_fixture_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/golden_fixture_spec.lua	
        [32mSuccess[0m	||	golden fixture normalization ignores only ripgrep tool description versions on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization compares captures with backend-only descriptions on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only the dedicated ls backend metadata on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only find backend metadata on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization preserves meaningful description changes on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization leaves the input and non-description values unchanged on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization ignores only ripgrep tool description versions on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization compares captures with backend-only descriptions on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only the dedicated ls backend metadata on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only find backend metadata on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization preserves meaningful description changes on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization leaves the input and non-description values unchanged on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization preserves unrelated tools and payloads without tools	
        
        	
        [32mSuccess: [0m	13	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/jit_oracle_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/jit_oracle_spec.lua	
        [32mSuccess[0m	||	jit trace oracle names LuaJIT's own mcode-allocation error text	
        
        [32mSuccess[0m	||	jit trace oracle words an abort as LuaJIT does, filling in its detail	
        
        [32mSuccess[0m	||	jit trace oracle blames the mcode defect only for a zero with at least one mcode abort	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/jit_tuning_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/jit_tuning_spec.lua	
        [32mSuccess[0m	||	jit tuning for arm64 macOS mcode placement (#294) knows which LuaJIT builds carry the mcode placement fix	
        
        [32mSuccess[0m	||	jit tuning for arm64 macOS mcode placement (#294) tunes only arm64 macOS with the fix, where a large area cannot thrash	
        
        [32mSuccess[0m	||	allocator safety policy (#267) identifies old and unknown allocators only on arm64 macOS	
        
        [32mSuccess[0m	||	allocator safety policy (#267) protects only the supplied module and its nested functions	
        
        [32mSuccess[0m	||	allocator safety policy (#267) rejects a missing chunk before it can disable the global engine	
        
        [32mSuccess[0m	||	allocator safety policy (#267) leaves fixed and other-platform editor compilation alone	
        
        [32mSuccess[0m	||	allocator safety policy (#267) disables only the isolated test VM on the unsafe allocator	
        
        [32mSuccess[0m	||	allocator safety policy (#267) keeps the existing fixed-build harness tuning without enabling JIT	
        
        [32mSuccess[0m	||	module-scoped compiler protection in a real VM keeps unrelated JIT traces active while protecting the entire document pipeline	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/ready_port_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/ready_port_spec.lua	
        [32mSuccess[0m	||	ready_port.wait_for_port times out by name when the signal file never appears	
        
        [32mSuccess[0m	||	ready_port.wait_for_port does not accept a zero-byte file as a published port	
        
        [32mSuccess[0m	||	ready_port.wait_for_port returns the port when digits land after the file is created empty	
        
        [32mSuccess[0m	||	ready_port.wait_for_port tolerates the trailing newline a shell writer would add	
        
        [32mSuccess[0m	||	ready_port.wait_for_port fails by name on non-numeric content instead of returning nil	
        
        [32mSuccess[0m	||	ready_port.wait_for_port rejects a number outside the TCP port range	
        
        [32mSuccess[0m	||	ready_port.is_listening rejects a connect that completed with nothing listening	
        
        [32mSuccess[0m	||	ready_port.is_listening accepts a real listener	
        
        [32mSuccess[0m	||	ready_port.is_listening never runs lsof while the port still refuses	
        
        [32mSuccess[0m	||	ready_port.is_listening lets the connect stand alone where lsof is absent	
        
        [32mSuccess[0m	||	ready_port.is_listening treats an lsof that is present but cannot answer as unconfirmed	
        
        [32mSuccess[0m	||	ready_port.is_listening require_listening fails by name when the fake never comes up (#361)	
        
        [32mSuccess[0m	||	ready_port.is_listening require_listening returns once a real listener is up	
        
        [32mSuccess[0m	||	ready_port.is_listening reads each lsof exit shape	
        
        	
        [32mSuccess: [0m	14	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/reap_test_orphans_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/reap_test_orphans_spec.lua	
        [32mSuccess[0m	||	#220 orphan census the Lua orphan predicate takes observations as data	
        
        [32mSuccess[0m	||	#220 orphan census runs the direct Python predicate tests through the normal suite	
        
        [32mSuccess[0m	||	#220 orphan census serializes real targeted runs and retains orphan cleanup (#409)	
        
        [32mSuccess[0m	||	#220 orphan census selects exactly this checkout's leaked harness processes	
        
        [32mSuccess[0m	||	#220 orphan census leaves another checkout's processes alone	
        
        [32mSuccess[0m	||	#220 orphan census leaves an editor open on a spec file alone	
        
        [32mSuccess[0m	||	#220 orphan census never selects itself or the recipe that invoked it	
        
        [32mSuccess[0m	||	#220 orphan census passes when the table holds none of this checkout's processes	
        
        [32mSuccess[0m	||	#220 orphan census reports but never fails the run in the before phase	
        
        [32mSuccess[0m	||	#220 orphan census fails loudly when ps ran but its columns cannot be read	
        
        [32mSuccess[0m	||	#220 orphan census skips visibly, and passes, where ps is refused	
        
        	
        [32mSuccess: [0m	11	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_deadline_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/spec_deadline_spec.lua	
        [32mSuccess[0m	||	spec deadline policy retains bounded machine-load scaling	
        
        [32mSuccess[0m	||	spec deadline policy starts one inactivity budget and an unscaled four-base absolute ceiling	
        
        [32mSuccess[0m	||	spec deadline policy allows bounded sequential work beyond the former whole-file deadline	
        
        [32mSuccess[0m	||	spec deadline policy applies inactivity to bootstrap, a running case and teardown	
        
        [32mSuccess[0m	||	spec deadline policy checks expiry before progress and gives the absolute ceiling precedence	
        
        [32mSuccess[0m	||	spec deadline policy does not double-scale the absolute ceiling under contention	
        
        [32mSuccess[0m	||	spec deadline policy only exact-next positive integer sequences refresh inactivity	
        
        [32mSuccess[0m	||	spec deadline policy rejects stale frames without accepting an old clock as fresh work	
        
        [32mSuccess[0m	||	spec deadline policy ignores malformed clocks and unrecognized events without refreshing	
        
        [32mSuccess[0m	||	spec deadline policy never revives any terminal outcome	
        
        [32mSuccess[0m	||	spec deadline policy keeps predecessor state and detached observations immutable	
        
        	
        [32mSuccess: [0m	11	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_jobs_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/spec_jobs_spec.lua	
        [32mSuccess[0m	||	spec job registry signals a job that is the same process or gone, and skips a reused pid	
        
        [32mSuccess[0m	||	spec job registry reads only well-formed ledger lines	
        
        [32mSuccess[0m	||	spec job registry knows this process's start time, and none for a missing pid	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_progress_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/spec_progress_spec.lua	
        [32mSuccess[0m	||	completed-case framing reassembles split frames and preserves the original surrounding output	
        
        [32mSuccess[0m	||	completed-case framing does not treat arbitrary, foreign, malformed or unterminated output as progress	
        
        [32mSuccess[0m	||	completed-case framing streams an oversized line without retaining it or admitting a suffix as a frame	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_runner_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/spec_runner_spec.lua	
        [32mSuccess[0m	||	scoped test timeouts extends only the real large conformance corpora	
        
        [32mSuccess[0m	||	scoped test timeouts stretches the deadline by oversubscription, never below 1x or above the cap	
        
        	
        [32mSuccess: [0m	2	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_isolation_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/test_isolation_spec.lua	
        [32mSuccess[0m	||	test_isolation.check accepts every root inside the marked root	
        
        [32mSuccess[0m	||	test_isolation.check refuses an environment with no marker	
        
        [32mSuccess[0m	||	test_isolation.check names each root outside the marked one	
        
        [32mSuccess[0m	||	test_isolation.check refuses a root that is unset, and a prefix that is not a parent	
        
        [32mSuccess[0m	||	test_isolation.check resolves links before comparing	
        
        [32mSuccess[0m	||	the harness pins every guarded root assigns each test_isolation root inside TEST_ENV	
        
        [32mSuccess[0m	||	cliproxy data root under test mode runs under the harness's marker	
        
        [32mSuccess[0m	||	cliproxy data root under test mode refuses a root outside the isolated env instead of writing there	
        
        [32mSuccess[0m	||	cliproxy data root under test mode allows the isolated default and a tempname override	
        
        [32mSuccess[0m	||	test_isolation.ambient names every COUCH_*, PAIR_* and terminal owner, sorted, and nothing else	
        
        [32mSuccess[0m	||	test_isolation.ambient is empty for the allowlist	
        
        [32mSuccess[0m	||	every harness nvim starts through the allowlist launcher has no nvim launch in Makefile.parley outside scripts/test-env.sh	
        
        [32mSuccess[0m	||	every harness nvim starts through the allowlist launcher strips PARLEY_LIVE_COUCH from every launch but check-couch-actors-live, and logs per target	
        
        [32mSuccess[0m	||	the sentinel gate and the launcher's argv runs SENTINEL_CHECK in every target that runs the orphan gate	
        
        [32mSuccess[0m	||	the sentinel gate and the launcher's argv builds no NAME=value argument except PATH, SHELL and the sentinel log	
        
        	
        [32mSuccess: [0m	15	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_process_model_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/test_process_model_spec.lua	
        [32mSuccess[0m	||	retained test process policy requires exact start identity, preserving opaque numeric precision	
        
        [32mSuccess[0m	||	retained test process policy accepts absence only after a complete birth record	
        
        [32mSuccess[0m	||	retained test process policy rejects malformed record envelopes regardless of observation	
        
        [32mSuccess[0m	||	retained test process policy rejects invalid pids even when identity matches or appears absent	
        
        [32mSuccess[0m	||	retained test process policy rejects malformed starts in either evidence source	
        
        [32mSuccess[0m	||	retained test process policy keeps missing, unavailable and malformed observations unknown	
        
        [32mSuccess[0m	||	retained test process policy accepts future fields without changing either input	
        
        	
        [32mSuccess: [0m	7	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_process_scope_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/test_process_scope_spec.lua	
        [32mSuccess[0m	||	retained process scope admission atomically admits one of two callers whose snapshots both have one free slot	
        
        [32mSuccess[0m	||	retained process scope admission retains legacy random scopes and refuses further admission visibly	
        
        [32mSuccess[0m	||	retained process scope admission admits after a sibling retires an empty scope from its completed snapshot	
        
        [32mSuccess[0m	||	retained process scope admission admits only one stale caller into the slot retired after both full snapshots	
        
        [32mSuccess[0m	||	retained process scope admission never returns a slot reused after its old directory identity was validated	
        
        [32mSuccess[0m	||	retained process scope admission refuses unreadable or unknown metadata rather than treating every nil as absence	
        
        [32mSuccess[0m	||	retained process scope admission refuses invalid present evidence even though atomic reservation has spare slots	
        
        [32mSuccess[0m	||	retained process scope admission uses a bounded slot set and allows reuse only after empty scope retirement	
        
        [32mSuccess[0m	||	retained process scope admission refuses malformed slot entries without following or deleting them	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_processes_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/test_processes_spec.lua	
        [32mSuccess[0m	||	retained birth registry persists unknown before any spawn and leaves interrupted publication intact	
        
        [32mSuccess[0m	||	retained birth registry preserves future nested fields through atomic pending to born publication	
        
        [32mSuccess[0m	||	retained birth registry preserves future JSON null members through publication	
        
        [32mSuccess[0m	||	retained birth registry round trips generated future JSON structures and preserves empty container kinds	
        
        [32mSuccess[0m	||	retained birth registry preserves unknown JSON numeric tokens without a lossy decoder oracle	
        
        [32mSuccess[0m	||	retained birth registry refuses malformed, duplicated or unresolved storage without effects	
        
        [32mSuccess[0m	||	retained birth registry refuses persisted identifiers outside native pid_t before observation	
        
        [32mSuccess[0m	||	retained birth registry bounds writers before creating the child and never drops old records	
        
        [32mSuccess[0m	||	retained birth registry observes joined death after TERM resistance instead of trusting signal delivery	
        
        [32mSuccess[0m	||	retained birth registry retains evidence after denied cleanup and allows a proven later retirement	
        
        [32mSuccess[0m	||	retained birth registry rechecks immediately before signals and never signals a reused pid	
        
        [32mSuccess[0m	||	retained birth registry does not treat unreadable native identity as an absent process	
        
        [32mSuccess[0m	||	retained birth registry keeps a legacy shared ledger until every original incarnation is gone	
        
        [32mSuccess[0m	||	retained birth registry scales native observation count linearly at N and 2N	
        
        [32mSuccess[0m	||	retained birth registry rejects symlink ledgers without reading or removing the external file	
        
        [32mSuccess[0m	||	retained birth registry keeps evidence replaced during observation instead of unlinking a newer birth	
        
        [32mSuccess[0m	||	retained birth registry refuses a registry writable by another principal	
        
        [32mSuccess[0m	||	retained birth registry uses explicit native absence and refuses values outside pid_t	
        
        [32mSuccess[0m	||	retained birth registry rescans new scopes created during cleanup before allowing a clean census	
        
        	
        [32mSuccess: [0m	19	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/version_support_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/541c515baf0670b0/tests/unit/version_support_spec.lua	
        [32mSuccess[0m	||	version_support supports the minimum and anything newer	
        
        [32mSuccess[0m	||	version_support refuses anything older	
        
        [32mSuccess[0m	||	version_support names the minimum and the running version in its refusal	
        
        [32mSuccess[0m	||	version_support supports the editor running this spec	
        
        [32mSuccess[0m	||	version_support makes setup() refuse an older editor before installing anything	
        
        	
        [32mSuccess: [0m	5	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        owned process census verified
        clean: no surviving test processes
=== RUN   TestReviewerTmpLiveParley/codex
    reviewtmp_live_test.go:94: Parley HEAD=92bf8cbca1be44942bbfc078011cee7b89084812 fence=seatbelt agent=codex
            … still working — 30s elapsed via codex (pid 89186; inspect: ps -p 89186)
            … still working — 1m0s elapsed via codex (pid 89186; inspect: ps -p 89186)
            … still working — 1m30s elapsed via codex (pid 89186; inspect: ps -p 89186)
            … still working — 2m0s elapsed via codex (pid 89186; inspect: ps -p 89186)
            … still working — 2m30s elapsed via codex (pid 89186; inspect: ps -p 89186)
        effective TMPDIR=/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp checkout=/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2
        effective tempname=/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/nvim.xianxu/OrseqJ/0effective tempname=/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/p/tmp/nvim.xianxu/ayKAT3/0owned process census verified
        Running tests/integration/zero_config_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/zero_config_spec.lua	
        [32mSuccess[0m	||	zero-config setup boots with setup({}) and no keys, without warnings, and creates a chat	
        
        [32mSuccess[0m	||	zero-config setup keeps keyless provider defaults when the user supplies other keys	
        
        [32mSuccess[0m	||	zero-config setup drops a default key set to false	
        
        [32mSuccess[0m	||	zero-config setup keeps a provider disabled by an empty table even when a default key exists	
        
        [32mSuccess[0m	||	zero-config setup reports a missing key when it is first used, not at startup	
        
        [32mSuccess[0m	||	zero-config setup detects the nearest marked project when chat_dir is set	
        
        [32mSuccess[0m	||	zero-config setup stays out of repo mode with repo_root = false	
        
        [32mSuccess[0m	||	zero-config setup stays out of repo mode with PARLEY_REPO_MODE=0	
        
        [32mSuccess[0m	||	zero-config setup lets an explicit repo_root win over PARLEY_REPO_MODE=0	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/chat_dirs_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/chat_dirs_spec.lua	
        [32mSuccess[0m	||	chat roots: derivation and persistence contract get_chat_roots returns config.chat_dir as the only root in plain mode	
        
        [32mSuccess[0m	||	chat roots: derivation and persistence contract does not write chat_roots / chat_dirs to state.json	
        
        [32mSuccess[0m	||	chat roots: derivation and persistence contract ignores chat_roots / chat_dirs when loading from state.json	
        
        [32mSuccess[0m	||	chat roots: derivation and persistence contract removes chat_roots / chat_dirs from state.json on next persist	
        
        	
        [32mSuccess: [0m	4	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/chat_finder_logic_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/chat_finder_logic_spec.lua	
        [32mSuccess[0m	||	ChatFinder logic does not borrow a newer picker's insertion target for an older callback	
        
        Parley.nvim: Inserted chat reference: 2026-09-15.12-00-00.000_choice.md[32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) config <C-d> confirmation y	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) config <C-d> confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) config <C-g>D confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) registry fallback <C-d> confirmation y	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) registry fallback <C-d> confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) registry fallback <C-g>D confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) starter <C-d> confirmation y	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) starter <C-d> confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic finder delete keys (#249) starter <C-g>D confirmation n	
        
        [32mSuccess[0m	||	ChatFinder logic Group A: Timestamp parsing from filename A1: parses timestamp from valid YYYY-MM-DD-HH-MM-SS filename	
        
        [32mSuccess[0m	||	ChatFinder logic Group A: Timestamp parsing from filename A2: returns nil for non-timestamp filename	
        
        [32mSuccess[0m	||	ChatFinder logic Group A: Timestamp parsing from filename A3: handles YYYY-MM-DD format (without time)	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B1: resolves configured presets and keeps all as the final state	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B2: uses configured default month when opening in filtered mode	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B3: cycles left and right across presets and all	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B4: files within recency cutoff are included	
        
        [32mSuccess[0m	||	ChatFinder logic Group B: Recency cycling B5: files older than recency cutoff are excluded	
        
        [32mSuccess[0m	||	ChatFinder logic Group C: Entry sorting by timestamp C1: entries are sorted newest first	
        
        [32mSuccess[0m	||	ChatFinder logic Group C: Entry sorting by timestamp C2: equal timestamps maintain stable order	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D1: formats display with filename, topic, date, no tags	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D2: formats display with tags	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D3: formats searchable ordinal with filename, topic, tags	
        
        [32mSuccess[0m	||	ChatFinder logic Group D: Display formatting D4: ordinal with no tags has no trailing space	
        
        [32mSuccess[0m	||	ChatFinder logic Group E: Delete confirmation reopen behavior reopens ChatFinder on Esc during delete confirmation	
        
        [32mSuccess[0m	||	ChatFinder logic Group E: Delete confirmation reopen behavior reopens on the item that moves into the deleted visual row	
        
        [32mSuccess[0m	||	ChatFinder logic Group E: Delete confirmation reopen behavior opens delete confirmation from the source window	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings opens a loading picker before asynchronous acquisition starts	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings animates the real picker while Chat acquisition remains delayed	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings cancels loading acquisition before a recency mapping reopens Chat	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings joins a matching retained prewarm instead of starting a second scan	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings cancels picker-owned acquisition and ignores late root delivery	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps recency outside the exact discovery fingerprint	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings settles all-absent optional roots as a successful empty picker	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps a bounded error status when every Chat root fails	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps successful Chat records and warns once for record failures	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings replaces a mismatched retained prewarm instead of joining it	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings reuses unchanged cached metadata without requesting another read	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings prunes cache only for roots that enumerate successfully	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings opens with the active recency label and all four cycle mappings	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings <Tab> cycles left (aliases <C-a>), <S-Tab> right (aliases <C-s>) (#159)	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings prefers restoring selection by item value when reopening after delete	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings includes chat files from secondary chat roots	
        
        Parley.nvim: Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-secondary-6c01c3/2026-02-04-10-00-00-move-me.md
        Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-secondary-6c01c3/2026-02-04-10-00-00-move-me.md[32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps a moved chat visible when moving it from ChatFinder	
        
        
        Parley.nvim: Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-[archive]-6cb774/2026-02-04-10-00-00-escaped-root.md
        Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-[archive]-6cb774/2026-02-04-10-00-00-escaped-root.md[32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings finds moved chats in roots whose paths contain glob characters	
        
        
        Parley.nvim: Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-tilde-a56c51/2026-02-04-10-00-00-tilde-root.md
        Moved chat tree to: /Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2.sdlc-review/tmp/p/tmp/claude/parley-test-chatfinder-tilde-a56c51/2026-02-04-10-00-00-tilde-root.md[32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings reopens on moved chats when the destination root was configured with a tilde path	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings passes filename, tags, and topic as search text for matcher ranking	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings keeps canonical facet behavior across toggles and rediscovery	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings includes extra-root labels in finder search text	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings includes empty braces in primary-root search text	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings passes preserved tag filters back into the picker prompt	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings preserves only bracket-tag and brace-label fragments from finder queries	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings matches extra-root labels when the query uses brace syntax	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings preserves empty brace filters for the primary root	
        
        [32mSuccess[0m	||	ChatFinder logic Group F: Chat finder picker mappings falls back to the newer visual neighbor when deleting the oldest item	
        
        [32mSuccess[0m	||	ChatFinder logic default_sticky_query_for_repo_mode returns {} in plain repo mode	
        
        [32mSuccess[0m	||	ChatFinder logic default_sticky_query_for_repo_mode returns nil when not in repo mode	
        
        [32mSuccess[0m	||	ChatFinder logic default_sticky_query_for_repo_mode returns nil in super-repo mode (don't override aggregation intent)	
        
        	
        [32mSuccess: [0m	58	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        
        Parley.nvim: Chat finder: partial scan (1 roots, 0 files failed)Running tests/unit/helper_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/helper_spec.lua	
        [32mSuccess[0m	||	helper.uuid returns a string	
        
        [32mSuccess[0m	||	helper.uuid matches the expected xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx pattern	
        
        [32mSuccess[0m	||	helper.uuid returns unique values on successive calls	
        
        [32mSuccess[0m	||	helper.starts_with returns true when string starts with prefix	
        
        [32mSuccess[0m	||	helper.starts_with returns true for empty prefix (always matches)	
        
        [32mSuccess[0m	||	helper.starts_with returns false when string does not start with prefix	
        
        [32mSuccess[0m	||	helper.starts_with returns false for partial prefix mismatch	
        
        [32mSuccess[0m	||	helper.starts_with returns true when string equals prefix exactly	
        
        [32mSuccess[0m	||	helper.starts_with returns false when prefix is longer than string	
        
        [32mSuccess[0m	||	helper.ends_with returns true when string ends with suffix	
        
        [32mSuccess[0m	||	helper.ends_with returns true for empty suffix (always matches)	
        
        [32mSuccess[0m	||	helper.ends_with returns false when string does not end with suffix	
        
        [32mSuccess[0m	||	helper.ends_with returns true when string equals suffix exactly	
        
        [32mSuccess[0m	||	helper.ends_with returns false when suffix is longer than string	
        
        [32mSuccess[0m	||	helper.ends_with works with file extension patterns	
        
        [32mSuccess[0m	||	helper.last_content_line returns 0 when buffer has only whitespace	
        
        [32mSuccess[0m	||	helper.last_content_line returns the last non-whitespace line with trailing blanks	
        
        [32mSuccess[0m	||	helper.last_content_line returns the last line when the tail has content	
        
        [32mSuccess[0m	||	helper.last_content_line invalidates cache when buffer content changes	
        
        	
        [32mSuccess: [0m	19	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/provider_params_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/provider_params_spec.lua	
        [32mSuccess[0m	||	get_schema returns base schema for known provider	
        
        [32mSuccess[0m	||	get_schema returns empty params for unknown provider	
        
        [32mSuccess[0m	||	get_schema returns openai-compatible schema for cliproxyapi provider	
        
        [32mSuccess[0m	||	get_schema applies o-series override: removes temperature/top_p/max_tokens, adds reasoning_effort	
        
        [32mSuccess[0m	||	get_schema applies gpt-5 override: max_tokens gets api_name=max_completion_tokens	
        
        [32mSuccess[0m	||	get_schema applies gpt-5 override to gpt-5.2	
        
        [32mSuccess[0m	||	get_schema applies Claude Sonnet 4.6 exclusive group override	
        
        [32mSuccess[0m	||	get_schema does NOT apply Claude 4.6 override to Claude Sonnet 4 (non-4.6)	
        
        [32mSuccess[0m	||	resolve_params returns empty table for string model_config	
        
        [32mSuccess[0m	||	resolve_params uses user-provided values	
        
        [32mSuccess[0m	||	resolve_params applies defaults only when user omits param	
        
        [32mSuccess[0m	||	resolve_params does not send params with nil default when user doesn't set them	
        
        [32mSuccess[0m	||	resolve_params clamps values to declared range	
        
        [32mSuccess[0m	||	resolve_params uses api_name for Gemini params	
        
        [32mSuccess[0m	||	resolve_params omits unsupported params for o-series models	
        
        [32mSuccess[0m	||	resolve_params uses max_completion_tokens api_name for gpt-5	
        
        [32mSuccess[0m	||	resolve_params uses max_completion_tokens api_name for gpt-5.2	
        
        [32mSuccess[0m	||	resolve_params applies gpt-5 param mapping for cliproxyapi provider	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns allowed_callers={direct} for anthropic haiku 4.5 web_search	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns allowed_callers={direct} for anthropic haiku 4.5 web_fetch	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns empty overrides for non-haiku models	
        
        [32mSuccess[0m	||	resolve_tool_overrides returns empty overrides for unknown provider/tool	
        
        [32mSuccess[0m	||	validate_agent returns no errors for valid openai agent	
        
        [32mSuccess[0m	||	validate_agent errors when provider is missing	
        
        [32mSuccess[0m	||	validate_agent skips validation for string model	
        
        [32mSuccess[0m	||	validate_agent warns about unknown parameters	
        
        [32mSuccess[0m	||	validate_agent errors on exclusive group at_most_one violation	
        
        [32mSuccess[0m	||	validate_agent no error when only one of exclusive group is set	
        
        [32mSuccess[0m	||	validate_agent no error when neither of exclusive group is set	
        
        [32mSuccess[0m	||	validate_agent errors on require_one when none set	
        
        [32mSuccess[0m	||	validate_agent warns on range violation	
        
        [32mSuccess[0m	||	validate_agent does not warn when value is within range	
        
        [32mSuccess[0m	||	max_tokens default for Claude models (#228) is 64000 for Claude, on every transport that carries it	
        
        [32mSuccess[0m	||	max_tokens default for Claude models (#228) leaves non-Claude models alone, including behind the same proxy	
        
        [32mSuccess[0m	||	max_tokens default for Claude models (#228) does not disturb the gpt-5 rename to max_completion_tokens	
        
        	
        [32mSuccess: [0m	35	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/pure_functions_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/pure_functions_spec.lua	
        [32mSuccess[0m	||	find_exchange_at_line returns (index, 'question') when line is in question range	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (index, 'answer') when line is in answer range	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (nil, nil) when line is in gap between exchanges	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (nil, nil) when line is before all exchanges	
        
        [32mSuccess[0m	||	find_exchange_at_line returns (nil, nil) when line is after all exchanges	
        
        [32mSuccess[0m	||	find_exchange_at_line finds correct index in multi-exchange chat (not always first)	
        
        [32mSuccess[0m	||	find_exchange_at_line returns exchange when line is in margin after question with no answer	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact question boundary (line_start)	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact question boundary (line_end)	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact answer boundary (line_start)	
        
        [32mSuccess[0m	||	find_exchange_at_line includes line at exact answer boundary (line_end)	
        
        [32mSuccess[0m	||	simple_markdown_to_html escapes HTML special characters	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts fenced code block with language to <pre><code> with class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts fenced code block without language to <pre><code>	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts inline code to <code> with inline-code class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts # H1 to <h1> with main-header class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts ## H2 to <h2> with section-header class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts ### H3 to <h3> with sub-header class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts **bold** to <strong> with bold-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts __bold__ to <strong> with bold-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts *italic* to <em> with italic-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts _italic_ to <em> with italic-text class	
        
        [32mSuccess[0m	||	simple_markdown_to_html converts - list item to <li> with list-item class	
        
        [32mSuccess[0m	||	simple_markdown_to_html wraps list items in <ul> with bullet-list class	
        
        [32mSuccess[0m	||	simple_markdown_to_html blockquote conversion is broken (escapes > before matching)	
        
        [32mSuccess[0m	||	simple_markdown_to_html wraps content in paragraph tags	
        
        [32mSuccess[0m	||	simple_markdown_to_html splits multiple paragraphs on double newlines	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns plain agent name when web_search is disabled	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎] for openai model with search_model when web_search is enabled	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎?] for openai model without search_model when web_search is enabled	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎?] for cliproxyapi when strategy is none	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎] for cliproxyapi claude model in anthropic_tools_route	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search uses per-model strategy override for cliproxyapi labels	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎?] for cliproxyapi openai_search_model without search_model	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🌎] for cliproxyapi openai_tools_route without search_model	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🔧🌎] when tool-enabled agent also has web_search supported	
        
        [32mSuccess[0m	||	agent_display_name_with_web_search returns [🔧] when tool-enabled agent has web_search disabled	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) closes an INDENTED fence — the shape the prompt convention produces	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) an unterminated fence does not consume the following turns	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) a code block after prose is NOT nested inside a paragraph	
        
        [32mSuccess[0m	||	simple_markdown_to_html fence handling (#218) still converts a well-formed flush-left block	
        
        	
        [32mSuccess: [0m	41	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        owned process census verified
        clean: no surviving test processes
        owned process census verified
        Running tests/arch/arrival_budget_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/arch/arrival_budget_spec.lua	
        [32mSuccess[0m	||	arch: predicate waits use the arrival budget or say why not (#374) finds the waits it checks	
        
        [32mSuccess[0m	||	arch: predicate waits use the arrival budget or say why not (#374) bounds each by the budget, a named deadline, or a stated reason	
        
        	
        [32mSuccess: [0m	2	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/destructive_recipe_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/arch/destructive_recipe_spec.lua	
        [32mSuccess[0m	||	arch: test-clean-env only deletes what the harness built removes exactly the harness-built leaves, never the caller-supplied root	
        
        [32mSuccess[0m	||	arch: test-clean-env only deletes what the harness built stays word-tight when the checkout path contains a space	
        
        [32mSuccess[0m	||	arch: test-clean-env only deletes what the harness built wipes before the suite starts, not partway through it	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/embedded_rpc_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/arch/embedded_rpc_spec.lua	
        [32mSuccess[0m	||	arch: embedded nvim requests go through the bounded helper (#401) scans the suite, including the helper's own users	
        
        [32mSuccess[0m	||	arch: embedded nvim requests go through the bounded helper (#401) starts and drives no child nvim outside the helper	
        
        	
        [32mSuccess: [0m	2	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/fixture_lifecycle_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/arch/fixture_lifecycle_spec.lua	
        [32mSuccess[0m	||	arch: fixture process lifecycle no spec or helper spawns outside the fixture seam	
        
        [32mSuccess[0m	||	arch: fixture process lifecycle every executable fixture either watches its parent or declares why it cannot block	
        
        [32mSuccess[0m	||	arch: fixture process lifecycle both watchdogs recognize a parent already reparented to init	
        
        [32mSuccess[0m	||	arch: fixture process lifecycle TOOLING.md's remedy uses ps, and warns about pgrep	
        
        	
        [32mSuccess: [0m	4	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/scratch_placement_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/arch/scratch_placement_spec.lua	
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree nvim's temp directory is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree nvim's swap directory is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $XDG_DATA_HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $XDG_STATE_HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree $XDG_CACHE_HOME is outside the repo	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree scratch has no project-marked ancestor	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree the ancestor guard sees a marker in $HOME itself	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree no script or test roots a temp dir at a fixed /tmp	
        
        [32mSuccess[0m	||	arch: harness scratch stays out of the repo tree the /tmp guard sees every spelling it refuses, and passes the fallbacks	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard refuses a root under a marked tree and creates nothing there	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard refuses a root that carries the marker itself	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard refuses a root inside the checkout even without a marker	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard accepts an unmarked root outside the checkout, existing or not	
        
        [32mSuccess[0m	||	arch: the pre-write placement guard is wired into PREP_TEST_ENV with the configured marker	
        
        	
        [32mSuccess: [0m	15	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/arch/superseded_comment_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/arch/superseded_comment_spec.lua	
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces fires on a stacked restatement (planted false negative)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces stays silent when an annotation restates its own prose (planted false positive)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces stays silent on two paragraphs that merely share a topic	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces fires on an @param the signature does not take (planted false negative)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces fires on a blank line severing the block from its function	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces accepts an underscore-marked unused argument (planted false positive)	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces accepts a block that documents its own function	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces every @param block in the tree documents the function it precedes	
        
        [32mSuccess[0m	||	arch: a rewritten comment deletes the paragraph it replaces no comment run in the tree stacks a paragraph on its replacement	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/chat_progress_process_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/chat_progress_process_spec.lua	
        [32mSuccess[0m	||	chat progress managed process lifecycle shows pending decoration then writes admitted text without a minimum-visible delay	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle commits admitted partial output before the broken failure notification	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle commits admitted partial output before the unauthorized failure notification	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle commits admitted partial output before the http500 failure notification	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle cleans the admitted session when its provider secret is missing	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle cleans the admitted session when its exact transport admission key is occupied	
        
        [32mSuccess[0m	||	chat progress managed process lifecycle cleans the admitted session when process spawn is rejected	
        
        	
        [32mSuccess: [0m	7	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/discovery_registry_fixture_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/discovery_registry_fixture_spec.lua	
        [32mSuccess[0m	||	default discovery registry fixture ordinary Markdown discovery gets positive absence without reaching the sentinel	
        
        [32mSuccess[0m	||	default discovery registry fixture keeps direct calls and unmocked mutations behind the strict sentinel	
        
        [32mSuccess[0m	||	default discovery registry fixture lets an explicit stateful fake win over the empty discovery fixture	
        
        [32mSuccess[0m	||	default discovery registry fixture installs the same empty registry boundary in clean embedded editors	
        
        [32mSuccess[0m	||	default discovery registry fixture installs once without replacing the first fixture or wrapping it twice	
        
        	
        [32mSuccess: [0m	5	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/embedded_nvim_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/embedded_nvim_spec.lua	
        [32mSuccess[0m	||	embedded nvim helper runs Lua in the child and returns its value	
        
        [32mSuccess[0m	||	embedded nvim helper fails a request the child never answers, naming it, and kills the child	
        
        [32mSuccess[0m	||	embedded nvim helper joins an intentionally killed child before removing its private fixture	
        
        [32mSuccess[0m	||	embedded nvim helper retains and names its private fixture when the child exit is not confirmed	
        
        [32mSuccess[0m	||	embedded nvim helper retains a replaced fixture using the identity captured before child actions	
        
        [32mSuccess[0m	||	embedded nvim helper bounds a request at half the spec's idle budget, which the runner passes	
        
        [32mSuccess[0m	||	embedded nvim helper stops the child when its setup fails, and raises	
        
        [32mSuccess[0m	||	embedded nvim helper names a wedged child before the runner's idle kill	
        
        	
        [32mSuccess: [0m	8	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/fixture_ready_publish_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/fixture_ready_publish_spec.lua	
        [32mSuccess[0m	||	fake SSE server readiness publish never exposes the ready path in an incomplete state	
        
        	
        [32mSuccess: [0m	1	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/fixture_reaping_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/fixture_reaping_spec.lua	
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone a spec-child Neovim does	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone removes its query directory from the watchdog callback without following links	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_cliproxy fixture does	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_github_releases fixture does	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_sse_server fixture does, even with no request to answer	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_cliproxy login path does, though it never binds a port	
        
        [32mSuccess[0m	||	#220 a harness process exits when its parent is gone the fake_sips slow mode does, though it never binds a port	
        
        [32mSuccess[0m	||	#220 census conformance, against the real ps reads a real ps, or says why it cannot	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts reap() kills what spawn() started, and forgets it	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts reap({since = mark}) spares what was started before the mark	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts forgets a process that exited on its own, so live() cannot over-count	
        
        [32mSuccess[0m	||	#220 the fixture seam owns every process it starts reaps with the signal the caller asks for	
        
        	
        [32mSuccess: [0m	12	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/fresh_clone_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/fresh_clone_spec.lua	
        [32mSuccess[0m	||	standalone tracked archive loads and creates chats with intact and unavailable optional vocabulary	
        
        [32mSuccess[0m	||	standalone tracked archive rejects an escaping archive symlink before loading plugin code	
        
        [32mSuccess[0m	||	standalone tracked archive accepts a self-contained archive symlink	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/portable_make_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/portable_make_spec.lua	
        [32mSuccess[0m	||	portable test dependency configuration shows product test commands without a maintainer overlay	
        
        [32mSuccess[0m	||	portable test dependency configuration rejects a missing Plenary dependency with actionable advice	
        
        [32mSuccess[0m	||	portable test dependency configuration accepts an explicit dependency outside the default profile	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/spec_runner_native_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/spec_runner_native_spec.lua	
        [32mSuccess[0m	||	owned headless spec execution executes real Plenary assertions and preserves ordinary output	
        
        [32mSuccess[0m	||	owned headless spec execution retains an actual failed assertion and nonzero exit	
        
        [32mSuccess[0m	||	owned headless spec execution kills and reaps a stalled case after completed work	
        
        [32mSuccess[0m	||	owned headless spec execution enforces the absolute ceiling even while cases keep completing	
        
        [32mSuccess[0m	||	owned headless spec execution does not let arbitrary bootstrap output keep a child alive	
        
        [32mSuccess[0m	||	owned headless spec execution gives the absolute deadline precedence at controlled capped load	
        
        [32mSuccess[0m	||	owned headless spec execution includes after_each teardown in a case's progress boundary	
        
        [32mSuccess[0m	||	owned headless spec execution ends the spec's whole process group at the idle deadline	
        
        [32mSuccess[0m	||	owned headless spec execution ends a spec's vim.system grandchild at the idle deadline	
        
        [32mSuccess[0m	||	owned headless spec execution takes its spec's process group with it when the runner gets sigterm	
        
        [32mSuccess[0m	||	owned headless spec execution takes its spec's process group with it when the runner gets sigint	
        
        [32mSuccess[0m	||	owned headless spec execution signals no recorded job whose pid now names another process	
        
        	
        [32mSuccess: [0m	12	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/test_isolation_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/test_isolation_spec.lua	
        [32mSuccess[0m	||	test isolation guard refuses to run before any command when the env carries no isolation marker	
        
        [32mSuccess[0m	||	test isolation guard refuses when HOME or an XDG root lies outside the marked root	
        
        [32mSuccess[0m	||	test isolation guard fails closed when the isolation check itself cannot load	
        
        [32mSuccess[0m	||	test isolation guard runs normally when every root is inside the marked root	
        
        [32mSuccess[0m	||	the allowlist environment (#365) this spec process sees no ambient Couch, Pair or terminal-owner name	
        
        [32mSuccess[0m	||	the allowlist environment (#365) resolves a bare couch and pair to the sentinel, which records the call and refuses	
        
        [32mSuccess[0m	||	the allowlist environment (#365) the launcher passes exactly the allowlist, whatever the caller exports	
        
        [32mSuccess[0m	||	the allowlist environment (#365) lets each live flag carry only its own check's names	
        
        [32mSuccess[0m	||	the allowlist environment (#365) passes a credential only when the target names it in PARLEY_TEST_ENV_PASS	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/integration/test_process_census_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/integration/test_process_census_spec.lua	
        [32mSuccess[0m	||	retained embedded child census receives its retained checkout context and a recorded native spec birth	
        
        [32mSuccess[0m	||	retained embedded child census retains unknown when the launcher dies inside birth publication	
        
        [32mSuccess[0m	||	retained embedded child census installs one job owner and preserves on_exit for children already reaped at publication	
        
        [32mSuccess[0m	||	retained embedded child census records jobs from clean embedded children that never load minimal_init	
        
        [32mSuccess[0m	||	retained embedded child census cannot call an ordinary spec exit reaped while its recorded child lives	
        
        [32mSuccess[0m	||	retained embedded child census the independent terminal census detects and joins a TERM-resistant orphan	
        
        	
        [32mSuccess: [0m	6	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/await_helper_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/await_helper_spec.lua	
        [32mSuccess[0m	||	await.until_progress (#293) settles past the stall window while progress keeps moving	
        
        [32mSuccess[0m	||	await.until_progress (#293) fails as stalled one window after progress stops	
        
        [32mSuccess[0m	||	await.until_progress (#293) fails at the ceiling even while progress keeps moving	
        
        [32mSuccess[0m	||	await.until_progress (#293) a wait with no progress at all gets exactly the stall window	
        
        	
        [32mSuccess: [0m	4	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/golden_fixture_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/golden_fixture_spec.lua	
        [32mSuccess[0m	||	golden fixture normalization ignores only ripgrep tool description versions on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization compares captures with backend-only descriptions on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only the dedicated ls backend metadata on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only find backend metadata on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization preserves meaningful description changes on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization leaves the input and non-description values unchanged on Anthropic	
        
        [32mSuccess[0m	||	golden fixture normalization ignores only ripgrep tool description versions on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization compares captures with backend-only descriptions on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only the dedicated ls backend metadata on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization normalizes only find backend metadata on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization preserves meaningful description changes on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization leaves the input and non-description values unchanged on OpenAI	
        
        [32mSuccess[0m	||	golden fixture normalization preserves unrelated tools and payloads without tools	
        
        	
        [32mSuccess: [0m	13	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/jit_oracle_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/jit_oracle_spec.lua	
        [32mSuccess[0m	||	jit trace oracle names LuaJIT's own mcode-allocation error text	
        
        [32mSuccess[0m	||	jit trace oracle words an abort as LuaJIT does, filling in its detail	
        
        [32mSuccess[0m	||	jit trace oracle blames the mcode defect only for a zero with at least one mcode abort	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/jit_tuning_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/jit_tuning_spec.lua	
        [32mSuccess[0m	||	jit tuning for arm64 macOS mcode placement (#294) knows which LuaJIT builds carry the mcode placement fix	
        
        [32mSuccess[0m	||	jit tuning for arm64 macOS mcode placement (#294) tunes only arm64 macOS with the fix, where a large area cannot thrash	
        
        [32mSuccess[0m	||	allocator safety policy (#267) identifies old and unknown allocators only on arm64 macOS	
        
        [32mSuccess[0m	||	allocator safety policy (#267) protects only the supplied module and its nested functions	
        
        [32mSuccess[0m	||	allocator safety policy (#267) rejects a missing chunk before it can disable the global engine	
        
        [32mSuccess[0m	||	allocator safety policy (#267) leaves fixed and other-platform editor compilation alone	
        
        [32mSuccess[0m	||	allocator safety policy (#267) disables only the isolated test VM on the unsafe allocator	
        
        [32mSuccess[0m	||	allocator safety policy (#267) keeps the existing fixed-build harness tuning without enabling JIT	
        
        [32mSuccess[0m	||	module-scoped compiler protection in a real VM keeps unrelated JIT traces active while protecting the entire document pipeline	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/ready_port_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/ready_port_spec.lua	
        [32mSuccess[0m	||	ready_port.wait_for_port times out by name when the signal file never appears	
        
        [32mSuccess[0m	||	ready_port.wait_for_port does not accept a zero-byte file as a published port	
        
        [32mSuccess[0m	||	ready_port.wait_for_port returns the port when digits land after the file is created empty	
        
        [32mSuccess[0m	||	ready_port.wait_for_port tolerates the trailing newline a shell writer would add	
        
        [32mSuccess[0m	||	ready_port.wait_for_port fails by name on non-numeric content instead of returning nil	
        
        [32mSuccess[0m	||	ready_port.wait_for_port rejects a number outside the TCP port range	
        
        [32mSuccess[0m	||	ready_port.is_listening rejects a connect that completed with nothing listening	
        
        [32mSuccess[0m	||	ready_port.is_listening accepts a real listener	
        
        [32mSuccess[0m	||	ready_port.is_listening never runs lsof while the port still refuses	
        
        [32mSuccess[0m	||	ready_port.is_listening lets the connect stand alone where lsof is absent	
        
        [32mSuccess[0m	||	ready_port.is_listening treats an lsof that is present but cannot answer as unconfirmed	
        
        [32mSuccess[0m	||	ready_port.is_listening require_listening fails by name when the fake never comes up (#361)	
        
        [32mSuccess[0m	||	ready_port.is_listening require_listening returns once a real listener is up	
        
        [32mSuccess[0m	||	ready_port.is_listening reads each lsof exit shape	
        
        	
        [32mSuccess: [0m	14	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/reap_test_orphans_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/reap_test_orphans_spec.lua	
        [32mSuccess[0m	||	#220 orphan census the Lua orphan predicate takes observations as data	
        
        [32mSuccess[0m	||	#220 orphan census runs the direct Python predicate tests through the normal suite	
        
        [32mSuccess[0m	||	#220 orphan census serializes real targeted runs and retains orphan cleanup (#409)	
        
        [32mSuccess[0m	||	#220 orphan census selects exactly this checkout's leaked harness processes	
        
        [32mSuccess[0m	||	#220 orphan census leaves another checkout's processes alone	
        
        [32mSuccess[0m	||	#220 orphan census leaves an editor open on a spec file alone	
        
        [32mSuccess[0m	||	#220 orphan census never selects itself or the recipe that invoked it	
        
        [32mSuccess[0m	||	#220 orphan census passes when the table holds none of this checkout's processes	
        
        [32mSuccess[0m	||	#220 orphan census reports but never fails the run in the before phase	
        
        [32mSuccess[0m	||	#220 orphan census fails loudly when ps ran but its columns cannot be read	
        
        [32mSuccess[0m	||	#220 orphan census skips visibly, and passes, where ps is refused	
        
        	
        [32mSuccess: [0m	11	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_deadline_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/spec_deadline_spec.lua	
        [32mSuccess[0m	||	spec deadline policy retains bounded machine-load scaling	
        
        [32mSuccess[0m	||	spec deadline policy starts one inactivity budget and an unscaled four-base absolute ceiling	
        
        [32mSuccess[0m	||	spec deadline policy allows bounded sequential work beyond the former whole-file deadline	
        
        [32mSuccess[0m	||	spec deadline policy applies inactivity to bootstrap, a running case and teardown	
        
        [32mSuccess[0m	||	spec deadline policy checks expiry before progress and gives the absolute ceiling precedence	
        
        [32mSuccess[0m	||	spec deadline policy does not double-scale the absolute ceiling under contention	
        
        [32mSuccess[0m	||	spec deadline policy only exact-next positive integer sequences refresh inactivity	
        
        [32mSuccess[0m	||	spec deadline policy rejects stale frames without accepting an old clock as fresh work	
        
        [32mSuccess[0m	||	spec deadline policy ignores malformed clocks and unrecognized events without refreshing	
        
        [32mSuccess[0m	||	spec deadline policy never revives any terminal outcome	
        
        [32mSuccess[0m	||	spec deadline policy keeps predecessor state and detached observations immutable	
        
        	
        [32mSuccess: [0m	11	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_jobs_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/spec_jobs_spec.lua	
        [32mSuccess[0m	||	spec job registry signals a job that is the same process or gone, and skips a reused pid	
        
        [32mSuccess[0m	||	spec job registry reads only well-formed ledger lines	
        
        [32mSuccess[0m	||	spec job registry knows this process's start time, and none for a missing pid	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_progress_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/spec_progress_spec.lua	
        [32mSuccess[0m	||	completed-case framing reassembles split frames and preserves the original surrounding output	
        
        [32mSuccess[0m	||	completed-case framing does not treat arbitrary, foreign, malformed or unterminated output as progress	
        
        [32mSuccess[0m	||	completed-case framing streams an oversized line without retaining it or admitting a suffix as a frame	
        
        	
        [32mSuccess: [0m	3	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/spec_runner_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/spec_runner_spec.lua	
        [32mSuccess[0m	||	scoped test timeouts extends only the real large conformance corpora	
        
        [32mSuccess[0m	||	scoped test timeouts stretches the deadline by oversubscription, never below 1x or above the cap	
        
        	
        [32mSuccess: [0m	2	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_isolation_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/test_isolation_spec.lua	
        [32mSuccess[0m	||	test_isolation.check accepts every root inside the marked root	
        
        [32mSuccess[0m	||	test_isolation.check refuses an environment with no marker	
        
        [32mSuccess[0m	||	test_isolation.check names each root outside the marked one	
        
        [32mSuccess[0m	||	test_isolation.check refuses a root that is unset, and a prefix that is not a parent	
        
        [32mSuccess[0m	||	test_isolation.check resolves links before comparing	
        
        [32mSuccess[0m	||	the harness pins every guarded root assigns each test_isolation root inside TEST_ENV	
        
        [32mSuccess[0m	||	cliproxy data root under test mode runs under the harness's marker	
        
        [32mSuccess[0m	||	cliproxy data root under test mode refuses a root outside the isolated env instead of writing there	
        
        [32mSuccess[0m	||	cliproxy data root under test mode allows the isolated default and a tempname override	
        
        [32mSuccess[0m	||	test_isolation.ambient names every COUCH_*, PAIR_* and terminal owner, sorted, and nothing else	
        
        [32mSuccess[0m	||	test_isolation.ambient is empty for the allowlist	
        
        [32mSuccess[0m	||	every harness nvim starts through the allowlist launcher has no nvim launch in Makefile.parley outside scripts/test-env.sh	
        
        [32mSuccess[0m	||	every harness nvim starts through the allowlist launcher strips PARLEY_LIVE_COUCH from every launch but check-couch-actors-live, and logs per target	
        
        [32mSuccess[0m	||	the sentinel gate and the launcher's argv runs SENTINEL_CHECK in every target that runs the orphan gate	
        
        [32mSuccess[0m	||	the sentinel gate and the launcher's argv builds no NAME=value argument except PATH, SHELL and the sentinel log	
        
        	
        [32mSuccess: [0m	15	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_process_model_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/test_process_model_spec.lua	
        [32mSuccess[0m	||	retained test process policy requires exact start identity, preserving opaque numeric precision	
        
        [32mSuccess[0m	||	retained test process policy accepts absence only after a complete birth record	
        
        [32mSuccess[0m	||	retained test process policy rejects malformed record envelopes regardless of observation	
        
        [32mSuccess[0m	||	retained test process policy rejects invalid pids even when identity matches or appears absent	
        
        [32mSuccess[0m	||	retained test process policy rejects malformed starts in either evidence source	
        
        [32mSuccess[0m	||	retained test process policy keeps missing, unavailable and malformed observations unknown	
        
        [32mSuccess[0m	||	retained test process policy accepts future fields without changing either input	
        
        	
        [32mSuccess: [0m	7	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_process_scope_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/test_process_scope_spec.lua	
        [32mSuccess[0m	||	retained process scope admission atomically admits one of two callers whose snapshots both have one free slot	
        
        [32mSuccess[0m	||	retained process scope admission retains legacy random scopes and refuses further admission visibly	
        
        [32mSuccess[0m	||	retained process scope admission admits after a sibling retires an empty scope from its completed snapshot	
        
        [32mSuccess[0m	||	retained process scope admission admits only one stale caller into the slot retired after both full snapshots	
        
        [32mSuccess[0m	||	retained process scope admission never returns a slot reused after its old directory identity was validated	
        
        [32mSuccess[0m	||	retained process scope admission refuses unreadable or unknown metadata rather than treating every nil as absence	
        
        [32mSuccess[0m	||	retained process scope admission refuses invalid present evidence even though atomic reservation has spare slots	
        
        [32mSuccess[0m	||	retained process scope admission uses a bounded slot set and allows reuse only after empty scope retirement	
        
        [32mSuccess[0m	||	retained process scope admission refuses malformed slot entries without following or deleting them	
        
        	
        [32mSuccess: [0m	9	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/test_processes_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/test_processes_spec.lua	
        [32mSuccess[0m	||	retained birth registry persists unknown before any spawn and leaves interrupted publication intact	
        
        [32mSuccess[0m	||	retained birth registry preserves future nested fields through atomic pending to born publication	
        
        [32mSuccess[0m	||	retained birth registry preserves future JSON null members through publication	
        
        [32mSuccess[0m	||	retained birth registry round trips generated future JSON structures and preserves empty container kinds	
        
        [32mSuccess[0m	||	retained birth registry preserves unknown JSON numeric tokens without a lossy decoder oracle	
        
        [32mSuccess[0m	||	retained birth registry refuses malformed, duplicated or unresolved storage without effects	
        
        [32mSuccess[0m	||	retained birth registry refuses persisted identifiers outside native pid_t before observation	
        
        [32mSuccess[0m	||	retained birth registry bounds writers before creating the child and never drops old records	
        
        [32mSuccess[0m	||	retained birth registry observes joined death after TERM resistance instead of trusting signal delivery	
        
        [32mSuccess[0m	||	retained birth registry retains evidence after denied cleanup and allows a proven later retirement	
        
        [32mSuccess[0m	||	retained birth registry rechecks immediately before signals and never signals a reused pid	
        
        [32mSuccess[0m	||	retained birth registry does not treat unreadable native identity as an absent process	
        
        [32mSuccess[0m	||	retained birth registry keeps a legacy shared ledger until every original incarnation is gone	
        
        [32mSuccess[0m	||	retained birth registry scales native observation count linearly at N and 2N	
        
        [32mSuccess[0m	||	retained birth registry rejects symlink ledgers without reading or removing the external file	
        
        [32mSuccess[0m	||	retained birth registry keeps evidence replaced during observation instead of unlinking a newer birth	
        
        [32mSuccess[0m	||	retained birth registry refuses a registry writable by another principal	
        
        [32mSuccess[0m	||	retained birth registry uses explicit native absence and refuses values outside pid_t	
        
        [32mSuccess[0m	||	retained birth registry rescans new scopes created during cleanup before allowing a clean census	
        
        	
        [32mSuccess: [0m	19	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        Running tests/unit/version_support_spec.lua
        
        ========================================	
        Testing: 	/Volumes/TestScratch/p.qhfkKX/2038b4292bede1b2/tests/unit/version_support_spec.lua	
        [32mSuccess[0m	||	version_support supports the minimum and anything newer	
        
        [32mSuccess[0m	||	version_support refuses anything older	
        
        [32mSuccess[0m	||	version_support names the minimum and the running version in its refusal	
        
        [32mSuccess[0m	||	version_support supports the editor running this spec	
        
        [32mSuccess[0m	||	version_support makes setup() refuse an older editor before installing anything	
        
        	
        [32mSuccess: [0m	5	
        [31mFailed : [0m	0	
        [31mErrors : [0m	0	
        ========================================	
        owned process census verified
        clean: no surviving test processes
--- PASS: TestReviewerTmpLiveParley (260.27s)
    --- PASS: TestReviewerTmpLiveParley/claude (107.29s)
    --- PASS: TestReviewerTmpLiveParley/codex (152.75s)
PASS
ok  	github.com/xianxu/ariadne/cmd/sdlc	260.728s

```
