---
name: puretokens-connection
description: 用户询问 Skill 将调用的 Pure Tokens API 是否可确认、API 身份或连接检查时使用。
---

# Pure Tokens Connection

对固定 Pure Tokens API `https://api.puretokensx.com` 做一次只读身份检查。直接通过当前宿主的终端或原生 HTTPS/API 能力发出完整固定 URL 请求；只可在内存中解析当前活动连接中与固定 Pure Tokens origin 精确匹配的一把凭据，作为本次请求的认证头。第三方 CC Switch 或手工配置只要已写入宿主连接即可使用；不得检查 provider 标签，也不得把保存的 Base URL 当作请求目标。不得显示、复制、保存、索取或输出凭据、Base URL 或完整宿主配置；不得安装或调用 Node 运行器、MCP、本地代理、sidecar、备用服务或另一条认证路径。

不得申请、调用或使用 Computer Use，也不得打开、点击或控制浏览器、Pure Tokens Switch、Pure Tokens Desktop 或其他图形界面来读取配置、检查连接或替代该 GET 请求；不得调用其他 Skill 作为回退。不得因为没有单独的“连接接口”而停止；必须先直接发出固定 GET 请求。

只处理完成请求后得到的结构化结果；不得把响应原文、请求头或配置内容直接展示给用户。仅当实际终端/网络调用或匹配凭据解析失败时，按本地执行不可用处理，明确固定 API 未执行且不猜测配置原因。

先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们定义身份检查与用户可见结论的边界。

## 身份检查

- 每次用户查询只调用一次 `GET https://api.puretokensx.com/v1`；不得调用模型目录、Images、Videos、余额、任务或其他路径，也不得重试。
- 只有响应 JSON 同时明确返回 `status: "ok"`、`name: "Pure Tokens API"`、`base_url: "/v1"` 时，才回答：`Skill 使用的固定 API 端点标识为 Pure Tokens API，可以继续使用 Pure Tokens Skill。`
- 响应不可读、字段缺失、值不匹配或请求失败时，只能回答：`固定 Pure Tokens API 未返回可确认的标识，当前无法完成检查。` 不得把该结果归因于用户 Base URL、API Key 或 provider 配置，也不得读取或要求用户粘贴它们。
- 该检查验证的是固定 API 端点的既有公开声明，不是密码学防伪证明；未确认不等于证明用户的其他连接属于某个服务。不得根据模型列表、模型名称、请求是否成功、Base URL 字符串或 provider 标签改写这一结论。
- 用户问当前真实 Base URL、API Key 或宿主 provider 时，明确说明：Skill 只可为固定请求在内存中使用当前活动连接的匹配凭据，但不展示、比较或报告这些配置；引导用户在 CC Switch、Pure Tokens Desktop 或宿主自身的连接设置中查看。
