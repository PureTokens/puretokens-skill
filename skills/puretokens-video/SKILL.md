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

从当前 SKILL.md 的绝对位置解析同级 `../.puretokens-executor/puretokens-api`；Windows 使用 `puretokens-api.exe`，不能依赖工作目录或 PATH。宿主 ID 使用当前的 codex、claude-code、workbuddy、gemini-cli、grok-build、opencode 或 trae，不猜其他宿主配置。

用宿主文件工具创建 UTF-8 请求 JSON，然后调用 `<绝对执行器路径> submit --host <host-id> --request <绝对请求文件>`。请求包括 `kind: "video"`、`operation: "generate"` 或 `"edit"`、精确 model、prompt、parameters，以及可选当前附件 `attachments: [{"field":"声明字段","path":"附件绝对路径"}]`。命令行不携带提示词或凭据；命令完成后清理请求文件。完整示例只在需要时读 `references/executor-usage.md`。

首帧／让图动起来用 `image_to_video`；角色或风格参考用 `reference_image_video`；参考视频、参考音频、编辑视频分别用 `reference_video`、`reference_audio`、`video_edit`。附件用途不明确时先澄清。只能通过一个明确声明的组合 operation 混合附件类别。

本地附件必须随该次声明的 multipart 请求发送。用户明确给出的公网 HTTPS URL 只能放入声明的 JSON 参考字段；不下载、不探测、不转存、不生成 URL／文件 ID，不使用旧附件或把附件改写成提示词。没有声明的表示方式时停止并解释实际限制。

## 同任务交付

1. 新请求只提交一次。立即把返回的 task_id 和状态简短告知用户；没有可解析回执时按提交结果未知处理，不自动重提。
2. `wait --host <host-id> --request <同任务文件>` 只等待原任务，一个窗口最多7 次状态读取、300 秒；`status` 只读一次。完成即可进入 content；超时、未知状态或状态失败保留原 ID，询问是否继续，不能后台轮询或创建替代任务。
3. `content` 必须针对已完成的原任务，传 task_id、`task_status: "completed"`、已确认数量和现有输出目录绝对路径；只下载索引 0，每份实际交付给用户后才下载下一份。已验证的同任务输出可复用。
4. `downloaded_awaiting_host_delivery` 只表示下载完成。用当前宿主的附件交付方式交给用户后才称交付成功；做不到时报告“已生成并下载，当前宿主无法交付附件”，保留任务与文件。URL、HTML、SVG、状态文字或 task_id 不是媒体交付。

需要跨会话续接时，可在用户指定位置或当前工作区用 `submit --record <绝对任务记录>` 留下最小记录。记录仅保留任务身份、原 operation、模型、数量、安全参数与下载／交付进度，不保留凭据、prompt、参考 URL 或媒体字节。用 `resume --host <host-id> --record <文件>` 有界等待原任务，`content --host <host-id> --record <文件> --index <索引> --output-dir <现有绝对输出目录>` 取内容；实际交付后才运行 `delivered --record <文件> --index <索引>`。续接命令使用 --record 时不能同时传 --request。不手改记录、不标记未交付的文件。

## 失败与按需检查

优先使用执行器的阶段、`submission_outcome` 和脱敏说明。只有确认 POST 前的失败才能说未提交；网络或附件传输可能已经开始时，不能凭“附件失败”推断未创建任务。保留已知 task_id、`original_operation`、模型、已确认数量、安全参数及 `retry_not_before`；缺失数量表示未知，图片下载前必须确认原数量。未知提交不自动重试、不承诺未扣费或退款；状态／内容失败只继续原任务。

面向用户只说当前状态、产物或可操作的失败，不重复整份 JSON，不填“未返回”占位，不暴露内部 URL、原始响应、请求数据或配置。完整机器字段由执行器／可选任务记录保留。异常时才读 `references/behavior-scenarios.json` 匹配场景；详细约束在 `references/execution-contract.json`，简短展示规则在 `references/task-receipt.json`。

用户明确要求试参数／检查请求时才用 `preflight --host <host-id> --request <文件>`；它不 POST，不创建任务，也不证明价格、余额或权限。普通生成不增加这一步。精确报价、服务端幂等提交和未知任务查找尚未提供；不能用本地校验或记录代替这些能力。
