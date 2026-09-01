---
name: puretokens-connection
description: 用户询问 Skill 将调用的 Pure Tokens API 是否可确认、API 身份或连接检查时使用。
---

# Pure Tokens Connection

对固定 Pure Tokens API `https://api.puretokensx.com` 做一次只读身份检查。不得用未认证的通用 Fetch/WebFetch 猜测认证。Claude Code、Codex、WorkBuddy、Gemini CLI、Grok Build 和 OpenCode 都使用受管直连运行器；它只从当前宿主的固定已配置位置狭义匹配 Pure Tokens 凭据，且只在内存中为这一次固定请求构造认证头。WorkBuddy 可以匹配固定 origin 下无 query、无 fragment 的 `/v1` 或 `/v1/...` 单模型资源 URL；其他宿主仍只接受各自规定的精确连接形式。不得显示、复制、保存、索取或输出凭据、Base URL 或完整宿主配置；不得调用 MCP、本地代理、sidecar、备用服务或手工认证请求。

根据当前宿主运行对应命令：Claude Code：`node ~/.claude/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host claude-code --method GET --path /v1`；Codex：`node ~/.agents/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host codex --method GET --path /v1`；WorkBuddy：`node ~/.workbuddy/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host workbuddy --method GET --path /v1`；Gemini CLI：`node ~/.gemini/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host gemini-cli --method GET --path /v1`；Grok Build：`node ~/.grok/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host grok-build --method GET --path /v1`；OpenCode：`node ~/.config/opencode/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host opencode --method GET --path /v1`。只处理其结构化输出；不得把响应原文、请求头或配置内容直接展示给用户。Trae 目前只有手动连接配置，尚无批准的本地凭据读取契约；在 Trae 停止并说明无法安全执行已认证连接检查，绝不改用通用 Fetch、手工读取 Key 或猜测配置。

先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们定义身份检查与用户可见结论的边界。

## 身份检查

- 每次用户查询只调用一次 `GET https://api.puretokensx.com/v1`；不得调用模型目录、Images、Videos、余额、任务或其他路径，也不得重试。
- 只有响应 JSON 同时明确返回 `status: "ok"`、`name: "Pure Tokens API"`、`base_url: "/v1"` 时，才回答：`Skill 使用的固定 API 端点标识为 Pure Tokens API，可以继续使用 Pure Tokens Skill。`
- 响应不可读、字段缺失、值不匹配或请求失败时，只能回答：`固定 Pure Tokens API 未返回可确认的标识，当前无法完成检查。` 不得把该结果归因于用户 Base URL、API Key 或 provider 配置，也不得读取或要求用户粘贴它们。
- 该检查验证的是固定 API 端点的既有公开声明，不是密码学防伪证明；未确认不等于证明用户的其他连接属于某个服务。不得根据模型列表、模型名称、请求是否成功、Base URL 字符串或 provider 标签改写这一结论。
- 用户问当前真实 Base URL、API Key 或宿主 provider 时，明确说明：运行器可能为固定请求在内存中窄范围匹配凭据，但 Skill 不读取、展示、比较或报告这些配置；引导用户在 CC Switch、Pure Tokens Desktop 或宿主自身的连接设置中查看。
