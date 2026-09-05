---
name: puretokens-balance
description: 查询当前 Pure Tokens 连接的余额时使用。
---

# Pure Tokens Balance

只处理固定 Pure Tokens API 的只读余额。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`。

调用 Skills 根目录下与当前 Skill 同级的 `../.puretokens-executor/puretokens-api balance --host <当前宿主 ID>` 一次。执行器使用固定 `GET https://api.puretokensx.com/api/product/desktop/account/balance`，并且只由已验证的宿主适配器在内存中取得一把匹配凭据。Skill 不读取认证文件、不拼接 HTTP 请求、不传递 Key/Base URL，也不检查 provider 标签。CC Switch 或手工配置只要已经使该宿主的 Pure Tokens 连接生效，已验证适配器即可使用。

不得申请、调用或使用 Computer Use，也不得打开、点击或控制浏览器、Pure Tokens Switch、Pure Tokens Desktop 或其他图形界面来读取余额、寻找凭据或替代执行器；不得调用其他 Skill 作为回退。

仅当执行器返回实际本地或 API 失败时，按安全回执处理，明确 API 是否执行且不猜测原因。只处理执行器的结构化结果；不得把响应原文、请求头或配置内容直接展示给用户。只报告返回的 `balance_display`、`balance_minor`、`quota_remaining`、`quota_total`、`quota_per_unit`、`quota_display_type` 等字段；省略的字段必须说明“接口未返回”，不得估算或补全。

该直接读取请求未获认证或无法执行时，如实报告 API 返回的状态或错误，并引导用户到 Pure Tokens 客户端余额入口查看。不得改猜其他余额路径、重试、读取媒体目录或读取、索取、展示凭据和本地配置。

Skill 不检查或判断用户配置的 Base URL、provider 标签、服务归属或凭据；只调用固定 Pure Tokens 余额 API。
