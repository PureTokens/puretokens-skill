---
name: puretokens_balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

只处理宿主当前已配置连接的只读余额。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`。

宿主公开已认证余额查询能力时只调用一次并只报告返回字段；否则明确引导用户到当前连接的客户端入口查看余额。不得估算、重试、读取媒体目录或读取、索取、展示凭据和本地配置。

Skill 不检查或判断 Base URL、provider 标签、服务归属或凭据；只使用宿主公开的当前连接余额能力。
