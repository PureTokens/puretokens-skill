---
name: puretokens-image
description: 当前宿主使用 Pure Tokens 连接时，任何生成或编辑图片的请求都优先使用本 Skill。
---

# Pure Tokens Image

宿主绑定与本地失败停止：首次调用前根据当前应用或用户明确指定确定一个 host；本次任务所有 Skill 和执行器命令沿用该 host，不从安装目录推断宿主。任何命令返回 `active_connection_*`、`workbuddy_*` 或 `host_credential_adapter_unavailable` 本地失败时，立即停止本次 API 流程并返回脱敏说明；不得自动调用 puretokens-connection、init、doctor、models 或其他 Skill 重复探测，不得枚举、更换 `--host` 或借用其他客户端凭据。保留用户原任务和已有 task ID；本次调用未发请求不代表此前没有提交任务，不自动重提。只有用户明确要求重查或出现可验证的新诊断依据时，才在原 host 做一次相应检查；“继续生成”本身不是切换宿主或重复探测的授权。改变执行宿主必须有用户明确指定，不能把失败恢复当作指定。

## 执行边界

当前宿主上下文选择 Pure Tokens 或用户明确指定时，必须调用安装的原生执行器；它是唯一 API 传输，固定请求 `https://api.puretokensx.com`。不得自行发 HTTP，不回退到 imagegen／Imagen／通用视频 Skill、MCP、代理、Computer Use 或浏览器／桌面自动化。仅执行器在内存中使用当前宿主匹配连接；Skill 不读配置、不传或展示凭据。不需要用户安装 Node、Python、Go 或 Desktop。

从本 SKILL.md 绝对目录解析 `../.puretokens-executor/puretokens-api`，Windows 使用 `puretokens-api.exe`；不依赖 PATH 或工作目录。当前宿主 ID 为 codex、claude-code、workbuddy、gemini-cli、grok-build、opencode、trae、claude-desktop、dsh-desktop、zcode、kimi-code 或 qoder，不借用其他宿主连接。远程／沙箱不能访问执行器、连接或附件时报告实际限制，不复制凭据或换传输。

## 选择与提交

1. 默认模型 `gpt-image-2` 或用户给定精确 ID：直接读选中的 `references/profiles/<model>.json`。需要选模型、解析别名时才读 `references/model-index.json`，别名必须唯一。不读取所有 profile，也不先查余额、init、doctor、preflight 或实时目录。未知精确 ID 的纯文本请求可只传 model/prompt；字段／操作缺口由执行器按需读取一次目录，查询目录不代替提交权限。
2. 只发送 profile 声明的字段、值和 operation；保留用户意图。物理尺寸仅用于比例／输出限制说明，不直接作为 API 尺寸。图片参考／编辑用途不明确时才澄清。本地参考或编辑使用 profile 的 `image_edit`；`gpt-image-2` 使用 `https://api.puretokensx.com/v1/images/edits`、`media_operation: "image_edit"` 和 `image` 字段；公网参考使用 generations 的 `parameters.image`。数量只用 profile 的 `n`，`requested_count` 可省略由 n 推导，提供时必须一致。多个不同设计不能擅自拆成付费任务，先确认本次设计。
3. 当前本地附件仅随声明的 multipart 请求发送；用户给出的公网 HTTPS URL 只放入声明的 JSON 字段。不下载、探测、转存参考媒体，不把附件改成提示词。没有声明的传输方式时停止。
4. 用宿主文件工具创建 UTF-8 请求：kind=`image`、operation=`generate` 或 `edit`、model、prompt、parameters，以及需要的 attachments（field、绝对 path）。执行 `<执行器> submit --host <当前宿主> --request <绝对请求文件>`，完成后清理请求文件。不要把提示词或凭据放进命令行。仅需示例时读 `references/executor-usage.md`。

## 同任务完成交付

多图、跨会话或需要恢复时，在首次 submit 加 `--record <工作区或用户指定的绝对任务文件>`，使用唯一文件名；不能为切换记录模式重新提交。记录只含任务身份、安全参数及进度，不含 prompt、凭据、参考 URL 或媒体字节。单图短会话可不使用记录；无记录时将原 task_id、original_operation、model、确认数量、安全参数、reconciliation_required、retry_not_before 和 wait_windows_completed 原样带入每次续接。

- **提交一次**，立即简短告知返回的 task_id 和状态，然后按 `next_step` 继续。回执缺失／无法解析视为提交未知；仅清理本次临时请求文件后停止，不追加 API 操作、不重提、不猜扣费。
- **wait**：有记录用 `resume --host <host> --record <文件>`，否则用 `wait --host <host> --request <同任务文件>`。图片首次查询等待回执的 `retry_not_before`（默认接受后 5 秒），其后每 3 秒；每窗口最多 40 次读取、120 秒，网络与等待均计入。不要另加 sleep 或重启初始等待。始终遵守 API Retry-After。
- **窗口结束**：`ok=true`、`wait_outcome=window_ended` 表示仍在生成。只有 `next_step=wait` 且原交付授权和当前前台会话仍有效时，再续一个窗口；到累计两个窗口、`retry_deferred` 或 `next_step=await_user` 时暂停并询问是否稍后继续。保留 wait_windows_completed，不重置预算，不后台循环。用户取消、返回失败回执、未知状态或对账立即停止；窗口内短暂 429 仅由执行器按 Retry-After 和剩余预算处理；实际网络超时不是正常窗口结束。
- **content**：只对已完成原任务下载。记录方式用 `content --host <host> --record <文件> --index <索引> --output-dir <现有绝对目录>`；无记录传 completed 状态及确认数量。图片按 0..n-1，视频仅 0；每次一个索引，实际交付后才取下一个。
- **deliver**：把 downloaded_paths 的文件用宿主附件方式交给用户；下载不等于交付，URL、HTML、SVG 或任务号不能代替媒体。有记录时，实际交付后再执行 `delivered --record <文件> --index <索引>`。`done` 后结束，不自动审美检查或生成新任务。

附件交付失败只重交已有文件，不重新生成或下载。已完成且 `reconciliation_required` 不为 true 的记录用 `resume --host <host> --record <文件>` 本地校验并返回待交付文件；此分支不读凭据、不请求 API。按 `next_step=deliver` 交付；缺失有效证明时保留原文件、另选输出目录取同任务索引。跨命令复用须匹配记录中的 SHA-256、字节数和媒体类型。无记录且文件仍是本会话刚下载的原文件，可重交；无法确认则保留文件、另选输出目录取同任务同索引，不声称已交付。宿主无法提供附件时说明“已生成并下载，当前宿主无法交付附件”。

对账不自动续等；用户明确继续时，记录式 resume 只查询原任务一次，按实际结果更新标记。记录不能恢复无 ID 的未知提交，不手改记录或提前标记 delivered。

## 按需说明

失败只说明实际阶段、是否已有任务和下一步，使用执行器的脱敏 next_action；不展示整份 JSON、原始错误、内部 URL 或配置。失败、额度／费用问题才读 `references/failure-guide.md`；复杂异常按 id 查 `references/behavior-scenarios.json`。正式字段见 `references/execution-contract.json`，展示规则见 `references/task-receipt.json`，均非普通生成前置。

只有用户明确检查参数时用 preflight，它不创建任务、不报价、不证明权限。特定宿主安装或交付问题才读 `references/desktop-hosts.md`；本地夹具和 init 成功均不等于实机附件验收。ZCode／Qoder 连接存在不证明当前聊天选择；不据配置存在擅自路由。

当前会话若使用未反映在宿主已声明有效文件中的配置覆盖，停止并说明无法确认有效连接；不读取其他配置或借用默认连接。

本地连接失败按执行器脱敏状态解释：未识别、不可读、不支持或未验证均不等于“尚未配置”。保留现有连接，不据此要求重装、切换模型、重配或更换凭据；`api_request_executed: false` 时说明未请求 API、未认证凭据。具体宿主限制见 `references/desktop-hosts.md`。
