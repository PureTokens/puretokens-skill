---
name: puretokens_video
description: 通过当前 Pure Tokens 连接生成视频、选择视频模型或参数时使用。
---

# Pure Tokens Video

通过宿主当前已配置的连接调用 Pure Tokens Videos API。宿主负责 Base URL、认证、路由和原生字节交付；Skill 只请求相对 API 路径和该操作声明的 JSON 或 multipart body，绝不读取、扫描、展示或索取凭据或配置，也不检查 Base URL、provider 标签或服务归属。

先读取已安装的 `references/model-selection.json`、`references/execution-contract.json`、`references/task-receipt.json` 和 `references/behavior-scenarios.json`；它们是模型别名、请求/结果约束、用户回执和特殊场景的绑定规则。

## 模型与提交

- 先 `GET /v1/media/models`。未指定模型时使用 `grok-imagine-video-1.5-preview`；它也必须在当前认证目录中具有 `video` capability。
- 用户给出精确 ID 时，只有当前目录存在同一 ID 且含 `video` capability 才可使用。用户给出别名时，仅在 `model-selection.json` 中唯一匹配时解析；零个或多个候选时，列出候选精确 ID 并请用户选择。当前 `grok video` 唯一解析为 `grok-imagine-video`；`grok 1.5 video` 唯一解析为 `grok-imagine-video-1.5-preview`。
- 模型不存在、无权限或 capability 不匹配时，说明原因、列出当前目录候选，不提交也不换模型。
- 单次 `POST /v1/videos` 至少传 `model`。`prompt` 是否必需、以及时长、画幅、分辨率、尺寸或其他可选参数，必须由当前认证目录中该精确模型的 `input_schema.properties` 和 `constraints` 决定：属性明确 `required` 时必须传；属性未声明但约束明确给出“单一输入参考可省略 prompt”时，只能在该约束的精确单一参考模式省略；其余情况要求提示词。资料缺失、字段不存在或值超出资料声明时，在提交前说明原因、列出已声明值（若有），请用户删除该参数或改选模型；不得根据模型名猜测或静默改写。
- 当用户指定 `resolution` 且当前 profile 的 `constraints.resolution_by_mode` 存在时，先按实际操作确定模式：无媒体输入为 `text`，`image_to_video` 为 `image`，`reference_image_video`、`reference_video` 或 `reference_audio` 为 `reference`。该模式的已声明集合会进一步收窄 `resolution`；值不在集合内时，列出该模式允许值并停止。`video_edit` 或无法唯一确定模式的请求没有明确的模式规则时，不得借用别的模式或全局枚举猜测可用分辨率；请用户删除该参数或选择资料明确覆盖该操作的模型。
- 用户提供当前请求中明确给出的**公网 HTTPS 图片、视频或音频 URL**，或 profile 明确允许的 file-ID/voice-ID 值时，不要求 multipart operation：URL 必须是 `https://`，否则说明当前 profile 需要公网 HTTPS URL 并停止。只有当前 `input_schema.properties` 声明了对应精确字段，且 `constraints.reference_transport` 明确允许该 transport 时，才在 JSON body 中原样传递。`string[]` 字段传数组，`string` 或 `json` 字段只传用户给出的原始值；不得下载、探测、转码、公开、复用、生成 URL 或 file ID。多个候选字段或输入模式都匹配时，先列出目录声明的字段和限制，请用户选择，绝不猜测。
- 用户提供的是当前请求中明确附带的原生图片、视频或音频附件/宿主媒体对象时，必须从当前 `input_schema.operations` 读取对应操作：`image_to_video`、`reference_image_video`、`reference_video`、`reference_audio` 或 `video_edit`。只有操作明确声明请求方法、相对路径、`multipart/form-data`、所需字段、数量范围和 `multipart_file` transport 时才可提交。`video_edit` 必须声明 `POST /v1/videos/edits` 和 `video` 字段；其他原生附件操作必须声明 `POST /v1/videos` 及其精确字段。资料缺失、操作不支持、路径/字段不一致或附件数量越界时，说明具体限制并停止，绝不猜测字段、路径或换模型。
- 原生附件只以 profile 声明的 `multipart_file` 随这一次 Videos API 请求发送。Pure Tokens 会在网关内部短期 R2 暂存、确认供上游读取的 HTTPS URL 后映射到上游字段，且不把该 URL 返回给用户。Skill 不下载、转码、公开或复用附件，不生成 URL 或 file ID，也不调用独立上传接口；不得从本地文件路径、提示词、网页或历史任务猜测附件，也不得要求用户提供凭据。
- 多个**公网 URL/ID**字段仅在当前 profile 同时声明各字段且没有明示的互斥或模式限制时才可组合；多个原生附件类别仍必须有一个明确声明的组合 multipart operation。当前 profile 未说明组合方式时，说明限制并请用户仅保留一个已声明输入，绝不自行拆单、丢弃附件或把附件请求静默改为文生视频。模型元数据中的一般“支持参考媒体”标记本身不构成原生附件操作。

## 任务与交付

- 文生视频、图生视频和已声明的参考图视频都属于异步任务：只轮询同一 `/v1/videos/{task_id}`，完成后只读取同一 `/v1/videos/{task_id}/content` 并交付原生视频字节；不得查询或提交另一任务。
- 只有任务终态成功后才读取内容。每个任务只允许一个在途内容读取；将原生视频字节交付给宿主并完成写入/交付后，才可进行任何明确要求的同任务内容重试。不得预取、并发下载、重复下载已交付内容，或把媒体字节复制进提示词或会话缓存。
- 提交获得 `task_id` 后，才可在**新提交或用户明确继续查询该 `task_id` 的当前用户请求**内开始自动轮询。对同一任务始终最多一个在途状态请求，不创建后台计时器、队列或持续轮询。若状态响应有有效的正数 HTTP `Retry-After` 秒数且不超过剩余自动轮询预算，优先按它等待；否则依次等待 5、10、20、40、60 秒，此后最多每 60 秒读取一次。每个轮询窗口从本次提交或明确继续查询起算，最多 300 秒且最多读取状态 7 次；到期仍为 pending/running 时，报告当前状态并停止自动等待。状态读取遇到限流、5xx、传输错误或超时，也立即停止自动等待并如实报告。用户明确要求继续查询同一 `task_id` 时，才开启一个新的、同样有界的**同任务**轮询窗口；绝不查询或提交其他任务。不得把轮询到期或读取错误当作失败或重提理由。
- pending/running 时报告当前状态并在上述时间表内继续同一任务；失败、取消或过期时报告返回错误，要求用户明确发起新请求后才可再次提交。
- 提交结果未知、状态超时或内容交付失败时，说明不确定/失败状态；只可让用户选择继续查询或重试同一任务内容，绝不自动重提、换模型或换参数。
- 只有宿主交付原生视频字节才算成功。任务 ID、状态文字、HTML/SVG、组件或远程 URL 都不算成功。
- 当前连接不能执行 API 或交付原生字节时，在付费提交前停止，引导用户检查已有 Pure Tokens Base URL、认证与路由配置；不得索取凭据或改走备用路径。
- 每次提交、继续查询、完成或失败，都按 `task-receipt.json` 报告：已返回的精确模型 ID、任务 ID、返回状态、请求数量（视频为 1）、尺寸/参数、已交付数量（完成时）和下一步。任务仍在处理中时，下一步必须写明下次同任务查询的等待时间；自动轮询到期时，必须说明需用户明确要求继续查询。任务响应未返回模型或任务 ID 时明确写“未返回”，绝不猜测。
