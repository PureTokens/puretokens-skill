---
name: puretokens-video
description: 当前宿主使用 Pure Tokens 连接时，任何生成或编辑视频的请求都优先使用本 Skill。
---

# Pure Tokens Video

## 路由优先级

当前宿主使用 Pure Tokens 时，优先选择本 Skill，不得回退到通用视频 Skill。必须调用安装的原生执行器；单次 Go 执行器是唯一 API 传输，固定请求 `https://api.puretokensx.com`；Skill 不读取或传递凭据、不得自行发 HTTP、不调用其他媒体 Skill、MCP、代理、Computer Use、浏览器或桌面自动化。执行器只使用文档列出的当前宿主连接记录；凭据格式测试不等于真实宿主端到端验收。不需要用户安装 Node、Python、Go 或 Desktop。

## 选择与请求

先读小型 `references/model-index.json`，默认模型为 `grok-imagine-video-1.5-preview`；只读选中精确模型的 `references/profiles/<model>.json`。别名必须唯一。普通生成不先查余额、init、doctor、preflight 或实时目录。未知精确模型可仅传 model/prompt；只有明确查询、请求的字段／操作缺口或拒绝后的诊断，才按需查询实时目录。目录缺少一个模型不等于已证明它不可提交，权限以实际 API 返回为准。

只使用 profile 声明的字段、值和 operation，保留用户的目的、主体、构图、文字和限制。物理尺寸如 `200cm × 230cm` 仅用于说明比例／输出限制，不能直接作为 API 尺寸。一次任务只生成一个视频。仅在 profile 声明 `generate_audio` 时，将生成声音／静音映射为 true／false；当前音频附件仍使用参考音频 operation。

从当前 SKILL.md 的绝对位置解析同级 `../.puretokens-executor/puretokens-api`；Windows 使用 `puretokens-api.exe`，不能依赖工作目录或 PATH。宿主 ID 使用当前的 codex、claude-code、workbuddy、gemini-cli、grok-build、opencode、trae、claude-desktop、dsh-desktop 或 zcode，不猜其他宿主配置。

用宿主文件工具创建 UTF-8 请求 JSON，然后调用 `<绝对执行器路径> submit --host <host-id> --request <绝对请求文件>`。请求包括 `kind: "video"`、`operation: "generate"` 或 `"edit"`、精确 model、prompt、parameters，以及可选当前附件 `attachments: [{"field":"声明字段","path":"附件绝对路径"}]`。命令行不携带提示词或凭据；命令完成后清理请求文件。完整示例只在需要时读 `references/executor-usage.md`。

首帧／让图动起来用 `image_to_video`；角色或风格参考用 `reference_image_video`；参考视频、参考音频、编辑视频分别用 `reference_video`、`reference_audio`、`video_edit`。附件用途不明确时先澄清。只能通过一个明确声明的组合 operation 混合附件类别。

本地附件必须随该次声明的 multipart 请求发送。用户明确给出的公网 HTTPS URL 只能放入声明的 JSON 参考字段；不下载、不探测、不转存、不生成 URL／文件 ID，不使用旧附件或把附件改写成提示词。没有声明的表示方式时停止并解释实际限制。

## 同任务交付

1. 新请求只提交一次。立即把返回的 task_id 和状态简短告知用户；没有可解析回执时按提交结果未知处理，不自动重提。
2. `wait --host <host-id> --request <同任务文件>` 只等待原任务，一个窗口最多7 次状态读取、300 秒；`status` 只读一次。完成即可进入 content；超时、未知状态或状态失败保留原 ID，询问是否继续，不能后台轮询或创建替代任务。
3. `content` 必须针对已完成的原任务，传 task_id、`task_status: "completed"`、已确认数量和现有输出目录绝对路径；只下载索引 0，再实际交付该文件；一个视频任务没有下一索引。只有显式任务记录中的下载摘要与当前文件匹配时，才能复用同任务输出；无摘要、旧记录或文件已被改写时保留原文件，选择另一个输出目录获取同一索引。
4. `downloaded_awaiting_host_delivery` 只表示下载完成。用当前宿主的附件交付方式交给用户后才称交付成功；做不到时报告“已生成并下载，当前宿主无法交付附件”，保留任务与文件。URL、HTML、SVG、状态文字或 task_id 不是媒体交付。

服务端要求对账时停止本轮自动等待。用户明确继续带对账标记的记录时，`resume` 只查询原任务一次，更新标记；仍需对账则保留任务，已恢复后再按新状态等待或取文件。不清除标记来绕过服务端确认。

提交前选择一种方式：单次会话可只用 `--request`；需要跨会话续接或文件复用时，在同一次 `submit --host <host-id> --request <请求文件>` 加上 `--record <绝对任务记录>`，将最小记录留在用户指定位置或当前工作区。不能为切换方式重新提交。记录仅保留任务身份、原 operation、模型、数量、安全参数与下载／交付进度，不保留凭据、prompt、参考 URL 或媒体字节；下载进度包含文件 SHA-256、字节数和媒体类型，用于核对复用与交付。用 `resume --host <host-id> --record <文件>` 有界等待原任务，`content --host <host-id> --record <文件> --index <索引> --output-dir <现有绝对输出目录>` 取内容；实际交付后才运行 `delivered --record <文件> --index <索引>`。续接命令使用 --record 时不能同时传 --request。不手改记录、不标记未交付的文件。

## 失败与按需检查

失败答复应包含“本次发生了什么、任务是否已提交／是否保留原任务、下一步可以做什么”，将执行器 `next_action` 翻译成简洁自然语言。不要只给错误码或让用户自行排查。

- `insufficient_quota`／`insufficient_balance`：明确提示本次可用额度／余额不足，提供 [Pure Tokens 官网 · 钱包充值](https://console.puretokensx.com/wallet)，引导用户登录查看钱包并按需充值。额度不足还应提醒检查当前 Key 限额，不能断言哪项耗尽或保证充值解决 Key 限额。不猜缺多少钱、实际扣费或充值后一定成功；用户要求时才查余额。此链接只供用户自行访问，不作为执行器请求或浏览器代查入口。
- 认证拒绝与权限不足分别解释，建议在宿主自身设置检查连接有效性或对应模型权限，不索取凭据、不自动更换连接。模型不存在时可建议明确查询模型；参数不支持时按所选 profile 修正字段并保留用户意图。
- 限流遵守回执中的等待要求；没有等待时间就不编造秒数。服务故障／网络失败只说明当前请求无法完成，不推断余额或凭据问题。内容拒绝请用户调整需求，不自动换模型绕过。
- 已有任务遇到状态／下载错误时保留原 ID、原操作及数量；终态失败不再轮询；下载成功但附件交付失败只处理已有文件，不能说已交付。记录或输出路径错误按脱敏建议检查访问权限／绝对路径，不删除未知文件。

优先使用执行器的阶段、`submission_outcome` 和脱敏说明。只有确认 POST 前的失败才能说未提交；网络或附件传输可能已经开始时，不能凭“附件失败”推断未创建任务。保留已知 task_id、`reconciliation_required`、`original_operation`、模型、已确认数量、安全参数及 `retry_not_before`；缺失数量表示未知，视频任务只允许一个输出。未知提交不自动重试、不承诺未扣费或退款；状态／内容失败只继续原任务。

面向用户只说当前状态、产物或可操作的失败，不重复整份 JSON，不填“未返回”占位，不暴露内部 URL、原始响应、请求数据或配置。完整机器字段由执行器／可选任务记录保留。异常时才读 `references/behavior-scenarios.json` 匹配场景；详细约束在 `references/execution-contract.json`，简短展示规则在 `references/task-receipt.json`。

错误文字使用执行器的固定分类提示；只有服务端实际返回且执行器识别为公开分类的错误码才展示，未知码省略。不得根据被省略的原始错误猜测原因。容器校验通过不等于视频可播放或附件已交付；文件损坏时按回执继续获取同一任务索引，不复用坏文件。

用户明确要求试参数／检查请求时才用 `preflight --host <host-id> --request <文件>`；它不 POST，不创建任务，也不证明价格、余额或权限。普通生成不增加这一步。精确报价、服务端幂等提交和未知任务查找尚未提供；不能用本地校验或记录代替这些能力。

Claude Desktop 使用 `claude-desktop`，仅在能访问本机执行器和 Desktop 当前连接的本地会话执行；不能用 Claude Code 的连接代替。DSH Desktop 使用 `dsh-desktop`，采用当前本地 Harness 的连接。云端或隔离会话不能访问宿主记录时按实际失败停止，不复制凭据进入沙箱。宿主安装和交付说明按需读 `references/desktop-hosts.md`。

ZCode 本地执行使用宿主 ID `zcode`。宿主上下文已明确选择 Pure Tokens，或用户明确指定通过 Pure Tokens 执行时才路由到本 Skill；配置存在本身不能作为当前会话选择的证据。只使用唯一启用且端点匹配的 Pure Tokens 连接；不能据此声称识别当前聊天模型。目录、远程工作区及交付限制按需读 `references/desktop-hosts.md`。

官网引导只适用于当前对话模型仍能正常回复，且已收到独立 API 请求的结构化结果或用户主动询问的情况。当前对话模型本身因认证或服务故障无法返回时，Skill 无法执行提示，应由宿主处理；不因此新增 Key 管理或渠道状态的官网跳转。官网链接仅供用户自行访问，不调用浏览器代查，不新增 API 请求。

用户主动询问消费明细、历史扣费或某次生成的实际费用，而现有结果无法确认时，提供 [Pure Tokens 官网 · 使用记录](https://console.puretokensx.com/usage-logs)，建议按发生时间及模型核对。不要用余额差、任务状态或暂未找到记录推断是否扣费，不承诺退款，也不把使用记录入口说成自动找回任务的能力。不在普通生成或所有失败回执中自动附加此入口。
