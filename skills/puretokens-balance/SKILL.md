---
name: puretokens-balance
description: 查询当前 Pure Tokens 连接的账户钱包余额或当前 Key 剩余额度时使用。
---

# Pure Tokens Balance

调用安装的单次原生执行器查询余额。从当前 SKILL.md 的绝对目录解析同级 `../.puretokens-executor/puretokens-api`；Windows 使用 `puretokens-api.exe`，不要依赖工作目录或 PATH。调用 `<绝对执行器路径> balance --host <当前宿主 ID>`。仅执行器私密读取文档列出的当前宿主明确连接记录；Skill 不读取配置、不传 Key／Base URL，不自行 HTTP，也不使用其他传输。

执行器复用现有 Pure Tokens API Key，与官方 CC Switch 查询使用同一接口：`GET https://console.puretokensx.com/api/product/console/api-keys/usage`。成功后仅再读取一次公开的 `GET https://console.puretokensx.com/api/product/console/status` 获取金额换算比例，该请求不携带 Key。最多两次 GET，共用 30 秒截止时间，不自动重试。余额是固定 console 域名的特例；其他 API Skill 的固定 `https://api.puretokensx.com` 地址不变。不需要 Desktop、Web 登录态或用户粘贴 Key；不加 init、模型目录、其他 endpoint 或浏览器查询。详细规则按需读 `references/execution-contract.json`，异常时读 `references/behavior-scenarios.json`。

以执行器返回的 `result.remaining` 和 `unit` 为准，默认只回复一行（以下金额为示例）：

- `scope: account_wallet`：**当前账户钱包余额：10.00 USD。**
- `scope: key_allowance`：**当前 Key 剩余额度：10.00 USD（不代表账户钱包余额）。**

金额通常保留两位小数；非零但小于 0.01 的余额显示“低于 0.01 USD”，不要误报为零。负值如实说明欠额，不改为零。仅用户要求详情时补充 `used`／`total`；`total` 是接口返回的总额度，不叫充值总额或硬性上限。不输出 JSON、配置信息、长篇技术说明或无关诊断字段。

`unlimited_quota=true` 的官方接口分支返回账户钱包余额；false 返回该 Key 的额度。Key 不限额不等于钱包无限。这里的余额不含订阅套餐额度；不要把 Key 额度或钱包余额说成包含订阅的统一可用余额。只有用户问总可用、订阅或金额不一致时才解释此范围，并引导到 Pure Tokens 官网控制台查看：https://puretokensx.com。

剩余金额直接来自 `total_available`，按公开 `quota_per_unit` 转成 USD，与官方 CC Switch 脚本一致；不得再减已用额度，不使用旧兼容 billing 接口或其中的 `100000000` 占位值，不猜换算比例或人民币金额。查询不证明一次生成的实际价格、权限或一定可以付款；只有同币种、同范围的实际报价才能比较，最终由提交 API 判定。不要为每次生图／视频自动查余额。

失败时先读结构化回执，即使进程非零退出也不重复执行。余额认证请求的 401／403 简要说明连接认证被拒绝，引导用户在现有配置工具检查当前连接与 Key 状态；429／网络／超时说明稍后可再查；公开换算请求失败不表示 Key 无效；格式或换算信息缺失则说“暂时无法确认余额，请稍后重试或到 Pure Tokens 官网控制台查看”。仅展示实际返回的脱敏原因，未返回余额绝不填零或估算，不索取凭据、不要求安装客户端。
