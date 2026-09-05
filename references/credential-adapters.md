# Credential adapter scope

Only the native executor reads the documented connection records. It verifies the configured Pure Tokens hostname, HTTPS and path before using the matching credential for a fixed API request. Credentials are not arguments, request-file fields, receipts, logs or persistent executor state. The configured origin never replaces the fixed API origin.

Balance uses the same verified credential on the fixed `https://console.puretokensx.com/api/product/console/api-keys/usage` route used by the official CC Switch integration. This is an API-key route, not a browser-session route. Its public `/api/product/console/status` metadata read carries no credential. The console-origin exception applies only to balance; media, models and connection remain on the fixed API origin.

| Host | Explicit record read by this release | Evidence / boundary |
| --- | --- | --- |
| Codex | `CODEX_HOME/config.toml`, otherwise `~/.codex/config.toml`; configured `profile` and `model_provider`; provider `experimental_bearer_token`, explicit `env_key`, or sibling `auth.json` only when that provider declares `requires_openai_auth = true` | Standard TOML / Switch fixture, selected profile, directory override and both declared credential sources covered. CLI-only `-c`/`--profile` overrides not persisted in the record are not discoverable. |
| Claude Code | `CLAUDE_CONFIG_DIR/settings.json`, otherwise `~/.claude/settings.json`; `env.ANTHROPIC_BASE_URL` and `env.ANTHROPIC_AUTH_TOKEN` | Switch-format user settings fixture. Project/session override precedence has not been verified. |
| Gemini CLI | `~/.gemini/.env`; `GOOGLE_GEMINI_BASE_URL` and `GEMINI_API_KEY` | Switch-format fixture; project/session overrides and alternate auth modes not verified. |
| WorkBuddy | `WORKBUDDY_CONFIG_DIR/models.json`, then `CODEBUDDY_CONFIG_DIR/models.json`, otherwise `~/.workbuddy/models.json`; `url` + `apiKey` records | Requires one unique matching credential. Zero matching records, missing credential and multiple matching credentials have distinct outcomes. This is not proof of a separately selected UI model. |
| Grok Build | `~/.grok/config.toml`; `models.default` selects `model.<id>.base_url/api_key` | Standard TOML fixture. Unpersisted session selection not verified. |
| OpenCode | `~/.config/opencode/opencode.json`; root `model` selects `provider.<id>.options.baseURL/apiKey` | Switch-format fixture; JSONC, project overrides and external credential expansion not verified. |
| Trae | No managed credential adapter | Stops before any API request; installing the Skills does not imply media availability. |

The README's fixture-tested status refers only to these exact records. Complete host acceptance also requires the real host to execute the installed binary, identify the current connection, resume a task and deliver its file. Record those checks per OS in `host-acceptance.json`; do not mark pending cells passed from compilation or mock API tests.

`doctor` inspects only the loaded installation and documented Skill directories for the selected host before the existing init check. These local manifest checks neither resolve project/session credential overrides nor prove host attachment delivery. Help and update-version checks never invoke a credential adapter.
