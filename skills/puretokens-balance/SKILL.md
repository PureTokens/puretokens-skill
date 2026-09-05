---
name: puretokens-balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

调用当前 SKILL.md 同级 `../.puretokens-executor/puretokens-api balance --host <当前宿主 ID>`；Windows 使用绝对路径和 `.exe`。执行器是唯一 API 传输，只在内存中使用当前匹配连接凭据。Skill 不读取、索取或传递 Key、Base URL，不拼接 HTTP，也不使用 Computer Use、MCP、浏览器、桌面自动化或其他 Skill 回退。

固定请求 `https://api.puretokensx.com/v1/dashboard/billing/subscription` 和 `/v1/dashboard/billing/usage`，各读一次。它们使用 API Key 认证，无需 Pure Tokens Desktop 或 Web 登录会话。完整响应规则见 `references/execution-contract.json`，异常时读 `references/behavior-scenarios.json`。

只展示执行器返回的 `reported_limit`、`reported_usage`、`reported_remaining`。remaining 为接口总额减用量，usage 按接口契约除以 100；这两次顺序读取不是原子快照。明确“接口展示单位，未声明币种及账户／Key 范围”，不得因为字段名含 usd 就称美元，也不得把大额哨兵解释成真实无限余额。不转换人民币、不补充账户余额、不据此保证一次生图／视频足够付款。

任一接口不可达、无权限、返回逻辑错误或缺少数字时，余额未确认。展示脱敏状态与下一步，引导用户到 Pure Tokens 官网控制台核对；不猜 endpoint、不自动重试、不索取凭据或要求安装客户端。
