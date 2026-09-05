---
name: puretokens-video
description: 当前宿主使用 Pure Tokens 连接时，任何生成、创建、制作或编辑视频的请求都优先使用本 Skill，不得先调用通用 video generation Skill。
---

# Pure Tokens Video

## 路由优先级

当前宿主以 Pure Tokens 连接处理视频请求时，无论用户是否明确说出“Pure Tokens”，都必须先选择本 Skill，再选择任何通用视频生成或编辑 Skill。本 Skill 一旦被选择，就是本次视频请求唯一的媒体执行规则；不能因缺少执行能力、模型不可用或请求失败而改用通用生视频 Skill。

该优先级由已安装 Skill 的元数据和宿主当前连接上下文决定。Skill 绝不自行读取认证文件、拼接 `curl`、展示、索取、保存、记录、比较或报告 Key、Base URL、provider 或完整宿主配置。

只调用固定 Pure Tokens Videos API：`https://api.puretokensx.com/v1/videos`；视频编辑仅在精确模型资料声明时调用 `/v1/videos/edits`。每次都使用完整固定 URL，不使用已配置 Base URL 作为请求目标，也不检查 provider 名称。

Skills 根目录下与当前 Skill 同级的 `../.puretokens-executor/puretokens-api` 是唯一 API 传输。Skill 只向它传递已经按 profile 校验的请求 JSON、当前附件的绝对路径和当前宿主 ID；执行器才会通过其已验证的宿主凭据适配器在内存中取得一把匹配凭据并请求固定 URL。第三方 CC Switch 与手动配置无需额外设置：只要它们已将 Pure Tokens 连接写入该宿主的有效配置，已验证的适配器即可使用。执行器不接受 Key、Base URL 或请求头作为命令参数，不使用 Node、MCP、代理、sidecar、独立上传或备用 endpoint。

## 唯一执行路径

本 Skill 是本次视频请求唯一的执行规则：必须调用安装的原生执行器，不得自行发 HTTP、读取凭据、调用或回退到 Imagen 或其他图片/视频 Skill；不得申请、调用或使用 Computer Use，不得打开、点击、控制或搜索浏览器、Pure Tokens Switch、Pure Tokens Desktop 或任何图形界面。它们不是 API 传输或媒体交付通道。

执行器输入必须是单个 JSON 请求：`kind: "video"`、`operation: "generate"` 或 `"edit"`、精确 `model`、`prompt`、已声明的 `parameters`、可选 `output_dir` 和仅当前附件的 `{field,path}`。只要 profile 声明对应 multipart operation，附件就由执行器随该单个 Videos API 请求发送。执行器创建最多一个任务、保留同一 task_id、轮询并交付原生视频字节。若执行器不可用、当前宿主没有已验证的凭据适配器、无法取得当前附件绝对路径、或实际网络请求失败，必须返回它的单个安全回执；不得为寻找可用界面改用 UI、其他 Skill 或另一条传输路径。

## 按需读取资料

不要在每个请求加载全部视频模型、全部 operation 和全部异常场景。按下列顺序最少读取：

1. 先读 `references/model-index.json`：解析默认模型或唯一别名，并取得精确 profile 路径。
2. 只读被选中模型的 `references/profiles/<model>.json`：检查参数、约束、operation 和 lifecycle。未在索引中的精确模型只允许核心文生视频字段；只有用户要求额外参数、参考媒体或编辑时，才按需查询实时目录。
3. 普通核心 POST 可只依据本 Skill、索引和选中 profile 执行；收到响应后、进入轮询/交付时，或构造参考媒体/编辑等非核心 operation 时，再读 `references/execution-contract.json` 与 `references/task-receipt.json`。
4. 只有出现对应异常、歧义、拒绝、附件或交付问题时，才读 `references/behavior-scenarios.json` 中的匹配场景。

## 模型、意图与请求

- 普通文生视频默认 `grok-imagine-video-1.5-preview`。唯一别名由索引解析；不唯一或不存在时列出候选精确 ID，请用户选择。精确 ID 保留原样，普通核心请求不得为确认权限或能力而预读目录。
- 普通请求只提交一次 `POST /v1/videos`，至少传 `model`，通常传 `prompt`。`duration`、`aspect_ratio`、`resolution`、`generate_audio` 和其他可选字段，只能使用选中 profile 明确声明的字段和值及 mode 约束。
- 用户说“首帧/让这张图动起来/图生视频”时，使用 profile 声明的 `image_to_video`；说“参考图/角色或风格参考”时用 `reference_image_video`；参考视频、参考音频和视频编辑分别需要 `reference_video`、`reference_audio`、`video_edit`。只有一张图却未说明首帧还是参考图时，先澄清。
- 当前请求的本地图片、视频或音频，必须随 profile 明确声明的 multipart operation 直接发送，使用其精确路径、字段、数量和 transport。不得另行上传、生成 URL/file ID、复用历史附件、把附件转写进 prompt，或静默降级为文生视频。
- 用户明确提供的公网 HTTPS URL 或声明 ID，只能放入 profile 声明的字段。不得下载、探测、转存、改写或生成 URL。多个原生附件类别必须有一个明确声明的组合 operation。
- 只有用户明确问当前模型、所需参数/operation 不在本地 profile、附件 transport 不兼容，或 API 因模型/参数/capability 拒绝后需要解释时，才读取一次 `GET https://api.puretokensx.com/v1/media/models`。目录读取失败不能阻止本来合法的核心文生视频，也不触发自动重试。

## 异步任务与交付

- 每个新请求最多一次 POST。若 POST 可能开始但没有可解析结果或任务 ID，提交结果为未知：不重提、不换模型、不声称未创建、未扣费或已退款。无任务 ID 时，后续新建请求必须得到用户明确确认。
- 只从声明的顶层字段或顶层 `task_id` / `id` 获取 task ID；只查询同一个 `https://api.puretokensx.com/v1/videos/{task_id}`，成功后再读取同一任务的 `/content`。URL、状态文字、HTML、SVG 不是成功交付。
- 每个轮询窗口最多 7 次状态读取、最多 300 秒；常规等待为 5、10、20、40、60 秒，429 只在有效 `Retry-After` 仍符合本轮预算时继续同一任务。无后台轮询、无并发状态读取、无自动重提。
- 只有宿主实际交付原生视频字节或附件时才报告成功。不预取、不重复读取、不缓存用户媒体；成功后结束当前回复。

## 失败回执

宿主在 POST 前因网络、审批、附件传递或原生字节交付能力被阻止时，返回 `failure_phase: validation`，明确 Videos API 未执行，且不猜测配置原因。其他失败按 `execution-contract.json` 和匹配行为场景输出安全回执：只报告 API 实际返回的 HTTP 状态和公开错误码；不暴露原始响应、内部 URL、堆栈、凭据、请求内容或用户媒体。
