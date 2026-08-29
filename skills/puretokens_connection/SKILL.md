---
name: puretokens_connection
description: 用户询问当前连接是否为 Pure Tokens API、连接身份或连接检查时使用。
---

# Pure Tokens Connection

通过宿主当前已配置的连接做一次只读 API 身份检查。宿主负责 Base URL、认证、路由和 HTTP 执行；Skill 只请求相对路径，绝不读取、扫描、展示或索取真实 Base URL、凭据、provider 标签或宿主配置。

先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们定义身份检查与用户可见结论的边界。

## 身份检查

- 每次用户查询只调用一次 `GET /v1`；不得调用模型目录、Images、Videos、余额、任务或其他路径，也不得重试。
- 只有响应 JSON 同时明确返回 `status: "ok"`、`name: "Pure Tokens API"`、`base_url: "/v1"` 时，才回答：`当前连接的 API 标识为 Pure Tokens API，可以按 Pure Tokens 连接继续使用。`
- 响应不可读、字段缺失、值不匹配或请求失败时，只能回答：`当前连接未返回可确认的 Pure Tokens API 标识，无法确认它是否为 Pure Tokens。` 可建议用户检查当前连接的配置来源，但不得读取或要求用户粘贴 Base URL、API Key 或 provider 配置。
- 该检查验证的是当前 API 端点的既有公开声明，不是密码学防伪证明；未确认不等于证明该连接属于其他服务。不得根据模型列表、模型名称、请求是否成功、Base URL 字符串或 provider 标签改写这一结论。
- 用户问当前真实 Base URL、API Key 或宿主 provider 时，明确说明 Skill 不读取这些配置，并引导用户在 CC Switch、Pure Tokens Desktop 或宿主自身的连接设置中查看。
