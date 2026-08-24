---
name: puretokens_balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

只处理当前已配置的 Pure Tokens 连接的只读余额。宿主公开已认证余额查询能力时只调用一次并只报告返回字段；否则引导用户到当前连接的客户端入口查看余额。不得估算、重试、读取媒体目录或读取、索取、展示凭据和本地配置。

当前连接不是 Pure Tokens 或无法确认归属时停止，并说明“当前 Pure Tokens Skill 仅支持 Pure Tokens API”，提供 https://puretokensx.com/。
