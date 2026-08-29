---
name: puretokens_balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

只处理固定 Pure Tokens API 的只读余额。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`。

仅当运行环境能为已配置的 Pure Tokens 请求自动附带此账户读取所需认证时，调用一次 `GET https://api.puretokensx.com/api/product/desktop/account/balance`。只报告返回的 `balance_display`、`balance_minor`、`quota_remaining`、`quota_total`、`quota_per_unit`、`quota_display_type` 等字段；省略的字段必须说明“接口未返回”，不得估算或补全。

该直接读取请求未获认证或无法执行时，如实报告 API 返回的状态或错误，并引导用户到 Pure Tokens 客户端余额入口查看。不得改猜其他余额路径、重试、读取媒体目录或读取、索取、展示凭据和本地配置。

Skill 不检查或判断用户配置的 Base URL、provider 标签、服务归属或凭据；只调用固定 Pure Tokens 余额 API。
