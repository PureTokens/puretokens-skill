# Desktop host execution

Use the current host ID for every executor command, including init, balance, models and media. Installation and credential fixtures do not prove real API or attachment delivery.

## Claude Desktop

Use --host claude-desktop in a local Desktop Code session. Local Code sessions discover ~/.claude/skills (or the explicitly configured CLAUDE_CONFIG_DIR/skills); this directory is shared with Claude Code. The executor uses the active Desktop third-party gateway profile, never Claude Code settings as a fallback. Restart/start a new local session after installation.

Cloud chats and Cowork isolated environments may not have access to the macOS/Windows connection records or executable. Do not copy credentials into them, infer availability from a successful installation, or use a browser/MCP transport. Report the actual local failure. Host attachment handoff must be verified in the selected execution mode.

## DSH Desktop

Use --host dsh-desktop. The local Harness discovers DSH_HOME/skills. Without an explicit DSH_HOME, the installer uses ~/Library/Application Support/dsh-desktop/harness/skills on macOS or %APPDATA%/dsh-desktop/harness/skills on Windows. Quote paths containing spaces. The desktop supplies DSH_HOME to Harness; explicit absolute overrides select that one profile, never a directory search.

Project .dsh/skills, project .agents/skills and custom roots can shadow user skills. Check the loaded SKILL.md location; doctor checks the selected user root and the documented shared Agents root, not arbitrary project/plugin roots. Use the local session's attachment handoff after content; downloaded_awaiting_host_delivery is not a successful handoff.

DSH settings select agent-default-model.provider/model and llm-pi-ai.providers.<selected>. Its apiKeyEnv references exactly one entry in the sibling version-1 .credentials.yaml refs mapping. The executor checks the selected endpoint before reading that file; no process-environment credential fallback.

## ZCode

Use host `zcode` in a local execution environment. Install into `~/.zcode/skills`, or `ZCODE_DATA_BASE_DIR/.zcode/skills` when that explicit base directory is absolute. Refresh and enable the installed Skills in ZCode. The executor reads only the v2 connection store and uses one uniquely enabled matching Pure Tokens connection. This confirms API capability, not the current conversation model. Route media here only when host context explicitly selects Pure Tokens or the user explicitly requests it; connection presence alone is not a routing signal. Multiple matching enabled connections stop before a request. Legacy migration belongs to ZCode; no legacy-store fallback is attempted. Remote workspaces need their own accessible executor and connection; syncing Skills alone does not supply either. Native image/video attachment delivery remains pending real-host acceptance.
