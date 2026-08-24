---
name: puretokens_balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

只处理宿主当前已配置连接的只读余额。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`。

仅当宿主能复用当前连接中已经存在的已认证账户会话时，调用一次 `GET /api/product/desktop/account/balance`。只报告返回的 `balance_display`、`balance_minor`、`quota_remaining`、`quota_total`、`quota_per_unit`、`quota_display_type` 等字段；省略的字段必须说明“接口未返回”，不得估算或补全。

宿主不能公开这一个已认证账户会话时，明确引导用户到当前连接的客户端余额入口查看。不得改猜其他余额路径、重试、读取媒体目录或读取、索取、展示凭据和本地配置。

Skill 不检查或判断 Base URL、provider 标签、服务归属或凭据；只使用宿主公开的当前连接余额能力。
