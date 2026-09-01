---
name: puretokens-balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

只处理固定 Pure Tokens API 的只读余额。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`。

调用一次 `GET https://api.puretokensx.com/api/product/desktop/account/balance`。不得用未认证的通用 Fetch/WebFetch 猜测认证。Claude Code、Codex、WorkBuddy、Gemini CLI、Grok Build 和 OpenCode 都使用受管直连运行器；它只从当前宿主的固定已配置位置狭义匹配 Pure Tokens 凭据，且只在内存中为这一次固定请求构造认证头。WorkBuddy 可以匹配固定 origin 下无 query、无 fragment 的 `/v1` 或 `/v1/...` 单模型资源 URL；其他宿主仍只接受各自规定的精确连接形式。不得显示、复制、保存、索取或输出凭据、Base URL 或完整宿主配置；不得调用 MCP、本地代理、sidecar、备用服务或手工认证请求。

根据当前宿主运行对应命令：Claude Code：`node ~/.claude/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host claude-code --method GET --path /api/product/desktop/account/balance`；Codex：`node ~/.agents/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host codex --method GET --path /api/product/desktop/account/balance`；WorkBuddy：`node ~/.workbuddy/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host workbuddy --method GET --path /api/product/desktop/account/balance`；Gemini CLI：`node ~/.gemini/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host gemini-cli --method GET --path /api/product/desktop/account/balance`；Grok Build：`node ~/.grok/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host grok-build --method GET --path /api/product/desktop/account/balance`；OpenCode：`node ~/.config/opencode/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host opencode --method GET --path /api/product/desktop/account/balance`。只处理其结构化输出；不得把响应原文、请求头或配置内容直接展示给用户。Trae 目前只有手动连接配置，尚无批准的本地凭据读取契约；在 Trae 停止并说明无法安全执行已认证余额请求，绝不改用通用 Fetch、手工读取 Key 或猜测配置。只报告返回的 `balance_display`、`balance_minor`、`quota_remaining`、`quota_total`、`quota_per_unit`、`quota_display_type` 等字段；省略的字段必须说明“接口未返回”，不得估算或补全。

该直接读取请求未获认证或无法执行时，如实报告 API 返回的状态或错误，并引导用户到 Pure Tokens 客户端余额入口查看。不得改猜其他余额路径、重试、读取媒体目录或读取、索取、展示凭据和本地配置。

Skill 不检查或判断用户配置的 Base URL、provider 标签、服务归属或凭据；只调用固定 Pure Tokens 余额 API。
