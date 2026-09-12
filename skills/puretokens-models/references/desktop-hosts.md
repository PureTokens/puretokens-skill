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

连接验证失败必须按脱敏状态解释，适用于所有宿主。`active_connection_record_missing` 只说明执行环境中未找到适配器声明的记录；`active_connection_unavailable` 表示读取或解释未完成；`active_connection_selection_unconfirmed` 表示无法确认有效选择；`active_connection_format_unsupported`、`active_connection_endpoint_missing`、`active_connection_endpoint_unsupported` 表示适配器支持范围或字段识别问题；历史码 `active_connection_not_puretokens` 仅表示未通过固定端点识别，不证明用户从未配置。`active_connection_credential_missing` 表示适配器未取得可用凭据，不等于服务端判定凭据无效。`active_connection_ambiguous` 表示无法安全选择，不要求删除其他连接。保留现有可用连接，反馈宿主版本和诊断码；不展示配置，不要求重装、切换聊天模型、重配或更换凭据。`api_request_executed: false` 时明确本次未发送 API 请求、未认证凭据。TRAE 的 `host_credential_adapter_unavailable` 表示缺少受支持适配器，不是用户没配置。身份或网络检查失败也不能归因于未配置。

宿主绑定与本地失败停止：首次调用前根据当前应用或用户明确指定确定一个 host；本次任务所有 Skill 和执行器命令沿用该 host，不从安装目录推断宿主。任何命令返回 `active_connection_*`、`workbuddy_*` 或 `host_credential_adapter_unavailable` 本地失败时，立即停止本次 API 流程并返回脱敏说明；不得自动调用 puretokens-connection、init、doctor、models 或其他 Skill 重复探测，不得枚举、更换 `--host` 或借用其他客户端凭据。保留用户原任务和已有 task ID；本次调用未发请求不代表此前没有提交任务，不自动重提。只有用户明确要求重查或出现可验证的新诊断依据时，才在原 host 做一次相应检查；“继续生成”本身不是切换宿主或重复探测的授权。改变执行宿主必须有用户明确指定，不能把失败恢复当作指定。

## Pi

Use `--host pi`. Skills use `~/.pi/agent/skills`, or absolute `PI_CODING_AGENT_DIR/skills` without parent traversal. The executor first verifies one matching `models.json` endpoint using `openai-completions`, then checks only sibling `auth.json` for the same provider ID. Its inline `api_key` entry takes precedence; unsupported or unreadable active authentication stops without fallback. With no matching auth entry, only custom providers without built-in environment authentication may use the saved inline value. Commands, `$` interpolation/escapes and unrepresented CLI, extension or session overrides are unsupported. Never remove active authentication to force fallback. Saved connection verification does not prove current chat selection. Real API and attachment delivery remain unverified.
