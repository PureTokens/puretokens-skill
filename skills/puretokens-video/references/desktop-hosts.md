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

## Kimi Code and Qoder

Use host IDs `kimi-code` and `qoder` in the local execution environment. Kimi Code uses its declared data directory for Skills; Qoder uses its declared user directory. Explicit supported directory overrides must be absolute. These are separate from legacy Kimi CLI and QoderWork. Qoder saved connection verification does not identify the current chat selection. A session using an unrepresented command-line configuration override must stop instead of reading a different connection. Missing native attachment capability reports downloaded but not delivered; do not open a browser or substitute another transport. Real API and native attachment delivery remain separate acceptance checks.

## 按宿主排查

WorkBuddy 适配器无法匹配连接时，不据此断言当前聊天未使用 Pure Tokens；若用户确认已配置并成功使用 Pure Tokens，保留可用连接，按具体诊断检查适配；不能把自带模型聊天正常当作 Pure Tokens 配置有效的证据。排查仅提供客户端版本及脱敏诊断码，不读取或索取配置原文。Windows 的 `cleanup_status: pending` 表示已完成事务的暂存清理被拒绝，不等于文件同步失败；按独立 init 结果说明连接状态，不绕过删除守卫。锁文件存在不等于被占用；缺少管理清单不能归因于清理失败，也不能称为无害。

`init` 按当前宿主适配器验证已声明的有效连接，不要求用户为检查而切换聊天模型；配置选择规则依宿主而定，不能将此结果当作当前会话选择证明。WorkBuddy 使用 `workbuddy_record_missing`、`workbuddy_record_unreadable`、`workbuddy_record_format_unsupported`、`workbuddy_connection_not_found`、`workbuddy_credential_missing`、`workbuddy_connection_ambiguous` 区分本地原因；这些状态不代表安装失败，也不证明用户从未配置过。只反馈脱敏状态，不读取配置原文。
