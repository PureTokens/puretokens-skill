---
name: puretokens_connection
description: 用户询问 Skill 将调用的 Pure Tokens API 是否可确认、API 身份或连接检查时使用。
---

# Pure Tokens Connection

对固定 Pure Tokens API `https://api.puretokensx.com` 做一次只读身份检查。运行环境自动携带已有的 Pure Tokens 请求认证；Skill 绝不读取、扫描、展示、复制或索取真实 Base URL、凭据、provider 标签或宿主配置，也不构造认证头或调用 MCP。

先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们定义身份检查与用户可见结论的边界。

## 身份检查

- 每次用户查询只调用一次 `GET https://api.puretokensx.com/v1`；不得调用模型目录、Images、Videos、余额、任务或其他路径，也不得重试。
- 只有响应 JSON 同时明确返回 `status: "ok"`、`name: "Pure Tokens API"`、`base_url: "/v1"` 时，才回答：`Skill 使用的固定 API 端点标识为 Pure Tokens API，可以继续使用 Pure Tokens Skill。`
- 响应不可读、字段缺失、值不匹配或请求失败时，只能回答：`固定 Pure Tokens API 未返回可确认的标识，当前无法完成检查。` 不得把该结果归因于用户 Base URL、API Key 或 provider 配置，也不得读取或要求用户粘贴它们。
- 该检查验证的是固定 API 端点的既有公开声明，不是密码学防伪证明；未确认不等于证明用户的其他连接属于某个服务。不得根据模型列表、模型名称、请求是否成功、Base URL 字符串或 provider 标签改写这一结论。
- 用户问当前真实 Base URL、API Key 或宿主 provider 时，明确说明 Skill 不读取这些配置，并引导用户在 CC Switch、Pure Tokens Desktop 或宿主自身的连接设置中查看。
