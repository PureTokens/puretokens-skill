---
name: puretokens-balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

只处理固定 Pure Tokens API 的只读余额。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`。

调用一次 `GET https://api.puretokensx.com/api/product/desktop/account/balance`。直接通过当前宿主的终端或原生 HTTPS/API 能力发出该完整固定 URL 请求；只可在内存中解析当前活动连接中与固定 Pure Tokens origin 精确匹配的一把凭据，作为本次请求的认证头。第三方 CC Switch 或手工配置只要已把 Pure Tokens 连接写入宿主，均可作为该凭据来源；不得检查 provider 标签，也不得把已保存的 Base URL 当作请求目标。不得显示、复制、保存、索取或输出凭据、Base URL 或完整宿主配置；不得安装或调用 Node 运行器、MCP、本地代理、sidecar、备用服务或另一条认证路径。

不得申请、调用或使用 Computer Use，也不得打开、点击或控制浏览器、Pure Tokens Switch、Pure Tokens Desktop 或其他图形界面来读取余额、寻找凭据或替代该 GET 请求；不得调用其他 Skill 作为回退。不得因为没有单独的“余额接口”而停止；必须先直接发出固定 GET 请求。

仅当实际终端/网络调用或匹配凭据解析失败时，按本地计费前执行不可用处理，明确 API 未执行且不猜测原因。只处理完成请求后得到的结构化结果；不得把响应原文、请求头或配置内容直接展示给用户。只报告返回的 `balance_display`、`balance_minor`、`quota_remaining`、`quota_total`、`quota_per_unit`、`quota_display_type` 等字段；省略的字段必须说明“接口未返回”，不得估算或补全。

该直接读取请求未获认证或无法执行时，如实报告 API 返回的状态或错误，并引导用户到 Pure Tokens 客户端余额入口查看。不得改猜其他余额路径、重试、读取媒体目录或读取、索取、展示凭据和本地配置。

Skill 不检查或判断用户配置的 Base URL、provider 标签、服务归属或凭据；只调用固定 Pure Tokens 余额 API。
