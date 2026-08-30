---
name: puretokens_video
description: 通过当前 Pure Tokens 连接生成视频、选择视频模型或参数时使用。
---

# Pure Tokens Video

直接调用固定的 Pure Tokens Videos API：`https://api.puretokensx.com`。每个请求都使用完整 URL；运行环境会为已配置的 Pure Tokens 请求自动携带认证。Skill 绝不读取、扫描、展示、复制或索取 API Key、Base URL、认证文件或宿主配置，也不构造认证头、调用 MCP、本地代理或备用服务。

先读取已安装的 `references/model-selection.json`、`references/execution-contract.json`、`references/task-receipt.json` 和 `references/behavior-scenarios.json`；它们是模型别名、请求/结果约束、用户回执和特殊场景的绑定规则。

## 模型与提交

- 普通文生视频不得在提交前读取 `GET https://api.puretokensx.com/v1/media/models`。未指定模型时使用已安装选择中的 `grok-imagine-video-1.5-preview`。用户给出别名时，仅在 `model-selection.json` 中唯一匹配时解析；零个或多个候选时，列出候选精确 ID 并请用户选择。当前 `grok video` 唯一解析为 `grok-imagine-video`；`grok 1.5 video` 唯一解析为 `grok-imagine-video-1.5-preview`。
- 用户给出精确 ID 时保留并直接提交；即使该 ID 不在已安装快照中，也不得为了确认模型、权限或 capability 读取目录。对不在快照中的精确 ID，普通文生只能传核心 `model` 与 `prompt`；需要快照未声明的可选参数、参考输入或附件操作时，才可先读取一次实时目录以确定如何满足该明确要求。
- 单次 `POST https://api.puretokensx.com/v1/videos` 至少传 `model`，普通文生必须传 `prompt`。时长、画幅、分辨率、尺寸和其他可选参数由已安装精确模型的 `parameterSchema.properties` 与 `constraints` 决定；只有资料声明字段和值后才可传递。资料没有所需字段、值或“单参考可省略 prompt”的例外时，可读取一次实时目录；仍无资料或不兼容时，说明原因、列出已声明值（若有），请用户删除该参数或改选模型；不得根据模型名猜测或静默改写。
- 当用户指定 `resolution` 且资料声明 `constraints.resolution_by_mode` 时，先按实际操作确定模式：无媒体输入为 `text`，`image_to_video` 为 `image`，`reference_image_video`、`reference_video` 或 `reference_audio` 为 `reference`。该模式的已声明集合会进一步收窄 `resolution`；值不在集合内时，列出该模式允许值并停止。`video_edit` 或无法唯一确定模式的请求没有明确的模式规则时，不得借用别的模式或全局枚举猜测可用分辨率；请用户删除该参数或选择资料明确覆盖该操作的模型。
- 用户提供当前请求中明确给出的**公网 HTTPS 图片、视频或音频 URL**，或资料明确允许的 file-ID/voice-ID 值时，不要求 multipart operation：URL 必须是 `https://`，否则说明需要公网 HTTPS URL 并停止。拒绝含 URL 凭据、`localhost`、`.local`，或显式环回、私网、链路本地 IP 的 URL；Skill 不做 DNS 解析或可访问性探测，公网可读性仍由 API 验证。只有资料声明了对应精确字段且允许该 transport 时，才在 JSON body 中原样传递。`string[]` 字段传数组，`string` 或 `json` 字段只传用户给出的原始值；不得下载、探测、转码、公开、复用、生成 URL 或 file ID。多个候选字段或输入模式都匹配时，先列出资料声明的字段和限制，请用户选择，绝不猜测。
- 用户提供的是当前请求中明确附带的原生图片、视频或音频附件/运行环境媒体对象时，必须从已安装资料或为该明确操作按需读取的目录中读取对应 operation：`image_to_video`、`reference_image_video`、`reference_video`、`reference_audio` 或 `video_edit`。只有 operation 明确声明请求方法、相对路径、`multipart/form-data`、所需字段、数量范围和 `multipart_file` transport 时才可提交。Skill 将相对路径与固定 API origin 组合成完整 URL。`video_edit` 必须声明 `POST /v1/videos/edits` 和 `video` 字段；其他原生附件操作必须声明 `POST /v1/videos` 及其精确字段。资料缺失、操作不支持、路径/字段不一致或附件数量越界时，说明具体限制并停止，绝不猜测字段、路径或换模型。
- 原生附件只以资料声明的 `multipart_file` 随这一次 Videos API 请求发送。Pure Tokens 会在网关内部短期 R2 暂存、确认供上游读取的 HTTPS URL 后映射到上游字段，且不把该 URL 返回给用户。Skill 不下载、转码、公开或复用附件，不生成 URL 或 file ID，也不调用独立上传接口；不得从本地文件路径、提示词、网页或历史任务猜测附件，也不得要求用户提供凭据。
- 多个**公网 URL/ID**字段仅在资料同时声明各字段且没有明示的互斥或模式限制时才可组合；多个原生附件类别仍必须有一个明确声明的组合 multipart operation。资料未说明组合方式时，说明限制并请用户仅保留一个已声明输入，绝不自行拆单、丢弃附件或把附件请求静默改为文生视频。模型元数据中的一般“支持参考媒体”标记本身不构成原生附件操作。
- 只有用户明确问“当前有哪些模型/参数/操作”，请求的可选字段或附件操作不在已安装资料中，或提交被 API 以模型、参数或 capability 问题拒绝后需要解释时，才可读取一次 `GET https://api.puretokensx.com/v1/media/models`。这不是普通生视频的前置条件；目录读取失败绝不能阻止本可直接提交的核心文生视频，也不得触发自动重试或重提。

## 任务与交付

- 文生视频、图生视频和已声明的参考图视频都属于异步任务。回执中的 `task_id` 是统一称呼，不要求响应 JSON 使用同名字段：已安装模型 lifecycle 声明 `create.idField` 时，严格从该顶层字段取值；未声明时只可从顶层 `task_id` 或 `id` 取值。不得从 URL、嵌套对象、提示词或任意其他字段推导任务 ID。两者都未返回时，报告“任务 ID 未返回，无法安全继续查询”；不得轮询、下载、自动重提或宣称未创建任务。
- 有规范化 `task_id` 时，只轮询同一 `https://api.puretokensx.com/v1/videos/{task_id}`，完成后只读取同一 `https://api.puretokensx.com/v1/videos/{task_id}/content` 并交付原生视频字节。将任务 ID 作为不透明值进行 URL path-segment 编码；用户给出 URL、多个 ID 或空值时，请其提供一个任务 ID，不得提取、拼接或猜测。不得查询或提交另一任务。
- 对已安装 lifecycle，严格按其 `pendingStatuses`、`successStatuses`、`failureStatuses` 分类状态；资料未声明 lifecycle 时，仅将 `pending`、`queued`、`running`、`in_progress` 视为处理中，将 `completed`、`succeeded`、`success` 视为成功，将 `failed`、`cancelled`、`canceled`、`expired`、`error` 视为失败。状态字段缺失或值不在上述已知集合时，报告原始状态为未识别并停止自动轮询，只有用户明确要求继续该任务才可再读取一次状态；不得把未知状态当成功、失败或可重提。
- 只有任务终态成功后才读取内容。每个任务只允许一个在途内容读取；将原生视频字节交付给宿主并完成写入/交付后，才可进行任何明确要求的同任务内容重试。不得预取、并发下载、重复下载已交付内容，或把媒体字节复制进提示词或会话缓存。
- 提交获得 `task_id` 后，才可在**新提交或用户明确继续查询该 `task_id` 的当前用户请求**内开始自动轮询。对同一任务始终最多一个在途状态请求，不创建后台计时器、队列或持续轮询。若状态响应有有效的正数 HTTP `Retry-After` 秒数且不超过剩余自动轮询预算，优先按它等待；否则依次等待 5、10、20、40、60 秒，此后最多每 60 秒读取一次。每个轮询窗口从本次提交或明确继续查询起算，最多 300 秒且最多读取状态 7 次；到期仍为 pending/running 时，报告当前状态并停止自动等待。状态读取遇到限流、5xx、传输错误或超时，也立即停止自动等待并如实报告。用户明确要求继续查询同一 `task_id` 时，才开启一个新的、同样有界的**同任务**轮询窗口；绝不查询或提交其他任务。不得把轮询到期或读取错误当作失败或重提理由。
- pending/running 时报告当前状态并在上述时间表内继续同一任务；失败、取消或过期时报告返回错误，要求用户明确发起新请求后才可再次提交。
- 若提交响应明确因模型、参数或 capability 被拒绝且没有任务 ID，可按需读取一次目录，展示返回的当前候选或允许值，并要求用户明确发起修正后的新请求；绝不自动重提。遇到 `429` 时，报告 `Retry-After`（如有）并请用户等待后明确重试；不得自动等待或重试。遇到 5xx、传输错误或超时且未返回任务 ID 时，报告“提交结果未知”，不得断言任务未创建或自动重提；若返回了任务 ID，只能继续查询该任务。
- 状态超时或内容交付失败时，说明不确定/失败状态；只可让用户选择继续查询或重试同一任务内容，绝不自动重提、换模型或换参数。
- 只有运行环境交付原生视频字节才算成功。任务 ID、状态文字、HTML/SVG、组件或远程 URL 都不算成功。
- 直接 API 请求在任务被接受前失败时，如实报告 API 返回的状态或错误，并说明尚未创建任务；不得臆测用户的 Base URL、认证或路由配置，不得索取凭据或改走备用路径。
- 每次提交、继续查询、完成或失败，都按 `task-receipt.json` 报告：已返回的精确模型 ID、任务 ID、返回状态、请求数量（视频为 1）、尺寸/参数、已交付数量（完成时）和下一步。任务仍在处理中时，下一步必须写明下次同任务查询的等待时间；自动轮询到期时，必须说明需用户明确要求继续查询。任务响应未返回模型或任务 ID 时明确写“未返回”，绝不猜测。
