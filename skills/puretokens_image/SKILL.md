---
name: puretokens_image
description: 通过当前 Pure Tokens 连接生成图片、选择图片模型、数量或尺寸时使用。
---

# Pure Tokens Image

通过宿主当前已配置的连接调用 Pure Tokens Images API。宿主负责 Base URL、认证、路由和原生字节交付；Skill 只请求相对 API 路径和该操作声明的 JSON 或 multipart body，绝不读取、扫描、展示或索取凭据或配置，也不检查 Base URL、provider 标签或服务归属。

先读取已安装的 `references/model-selection.json`、`references/execution-contract.json`、`references/task-receipt.json` 和 `references/behavior-scenarios.json`；它们是模型别名、请求/结果约束、用户回执和特殊场景的绑定规则。

## 模型与提交

- 每次新图片任务在付费提交前都先 `GET /v1/media/models` 一次。未指定模型时选择 `gpt-image-2`；明确 `image2` 也使用该模型。默认模型也必须在当前认证目录中具有 `image` capability；目录没有它或没有权限时，列出当前图片候选，不提交也不替换模型。
- 用户给出精确 ID 时，保留该 ID。用户给出别名时，仅在 `model-selection.json` 中唯一匹配时解析为精确 ID；零个或多个候选时，列出候选精确 ID 并请用户选择，绝不猜测。
- 只有当前认证目录存在同一精确 ID 且明确含 `image` capability 才提交；不存在、无权限或 capability 不匹配时，说明原因、列出当前目录候选，不提交也不换模型。
- 单次 `POST /v1/images/generations` 始终传 `model`、`prompt` 和 `async: true`。图片的 `n`、`size`、`image_size`、`aspect_ratio`、`width`、`height`、`strength` 及任何其他可选字段都必须由该精确模型当前 `input_schema.properties` 明确声明字段和值后才传递；默认值仅可使用目录实际返回的 `default`。未声明、值不兼容或资料缺失时，列出目录已声明的选项并请用户重选，不得按默认模型、README 或模型名猜测。
- 一个用户请求只能有一次付费提交，绝不拆单。只有 `n` 已由当前 profile 声明为允许的整数时才可传递；未指定时不编造 `n`。若用户要求多张但当前 profile 只声明 `n: 1` 或未声明 `n`，说明该模型当前一次只支持一张，请用户改为一张或选择目录明确支持所需数量的模型。
- `200cm × 230cm` 等物理尺寸不能传给 `n`、`size` 或其他 API 字段。说明不能精确保证物理尺寸，并列出当前精确模型实际声明的像素 `size`，或语义 `image_size` 和 `aspect_ratio` 选项；若三者均未声明，明确说目录未声明可选规格并请用户改用纯提示词或选择另一个模型。
- 用户提供的是当前请求中明确给出的**公网 HTTPS 图片 URL**时，先确认它是 `https://` URL；不是时说明当前 profile 只接受公网 HTTPS URL，停止并请用户提供合规 URL。只有当前 profile 同时声明了对应请求属性和该属性的 `constraints.reference_transport` 允许 `public_https_url`（或更宽的明确 URL transport）才可在 JSON body 中传该原始 URL。`string[]` 属性传 URL 数组；`string` 或 `json` 属性只传用户明确给出的原始值。一个请求存在多个可能的 URL 字段、或用户未说明单图/多图输入模式时，列出这些已声明字段并请用户选择；不得依据模型名或附件数量猜测字段，更不能下载、探测、转存或改写 URL。
- 用户提供的是当前请求中明确附带的原生图片附件/宿主媒体对象时，只有 `input_schema.operations.image_edit` 明确声明 `POST`、`multipart/form-data`、必需字段、允许数量和 `multipart_file` transport，且其请求路径为 `/v1/images/generations` 或 `/v1/images/edits` 时才可继续。Skill 严格使用该实时 operation 声明的路径，绝不把编辑路径固定为 generations。Skill 只把附件原生字节随这一次 Images API 请求发送；Pure Tokens 会在网关内部短期 R2 暂存、确认供上游读取的 HTTPS URL 后映射请求，且不会把该 URL 返回给用户。资料缺失、操作不支持或宿主不能交付附件字节时，在计费前停止并说明当前 profile 只支持何种输入；不得生成 URL、file ID 或调用独立上传接口，也不得静默降级为文生图。
- 图片编辑或参考图重绘的 multipart 请求严格使用 operation 声明的路径、字段、数量和 transport，并附带 `model`、`prompt`、`async: true` 和实时 profile 明确允许的可选字段。mask、透明背景、未声明的多参考图或任何其他扩展必须停止并说明未声明；不得伪装成功。

## 任务与交付

- 文生图和已声明的图片编辑都始终传 `async: true`。有 `task_id` 时，只读同一任务的 `/v1/images/{task_id}` 与 `/v1/images/{task_id}/content?index=N`；不得以默认模型猜测、不得查询或提交另一任务。
- 请求 `n` 张时，内容索引从 `0` 开始，必须按 `0` 到 `n-1` **顺序逐个**读取。仅在任务终态成功后读取内容；每项原生字节交付给宿主并完成写入/交付后，才可读取下一项。不得并发预取、重复下载已交付内容、把媒体字节复制进提示词或会话缓存。只有每个请求索引都交付了原生图片字节才算成功；部分成功时报告已交付和缺失的索引，只能继续读取同一任务缺失内容，绝不重提。
- 提交获得 `task_id` 后，才可在**新提交或用户明确继续查询该 `task_id` 的当前用户请求**内开始自动轮询。对同一任务始终最多一个在途状态请求，不创建后台计时器、队列或持续轮询。若状态响应有有效的正数 HTTP `Retry-After` 秒数且不超过剩余自动轮询预算，优先按它等待；否则依次等待 3、6、12、24、30 秒，此后最多每 30 秒读取一次。每个轮询窗口从本次提交或明确继续查询起算，最多 120 秒且最多读取状态 6 次；到期仍为 pending/running 时，报告当前状态并停止自动等待。状态读取遇到限流、5xx、传输错误或超时，也立即停止自动等待并如实报告。用户明确要求继续查询同一 `task_id` 时，才开启一个新的、同样有界的**同任务**轮询窗口；绝不查询或提交其他任务。不得把轮询到期或读取错误当作失败或重提理由。
- pending/running 时报告当前状态并在上述时间表内继续同一任务；失败、取消或过期时报告返回错误，要求用户明确发起新请求后才可再次提交。
- 提交结果未知、状态超时或内容交付失败时，说明不确定/失败状态；只可让用户选择继续查询或重试同一任务内容，绝不自动重提、换模型、换参数或拆分付费请求。
- 只有宿主交付原生图片字节才算成功。任务 ID、状态文字、HTML/SVG、组件或远程 URL 都不算成功。
- 当前连接不能执行 API 或交付原生字节时，在付费提交前停止，引导用户检查已有 Pure Tokens Base URL、认证与路由配置；不得索取凭据或改走备用路径。
- 每次提交、继续查询、完成或失败，都按 `task-receipt.json` 报告：已返回的精确模型 ID、任务 ID、返回状态、请求数量、尺寸/参数、已交付数量（完成时）和下一步。任务仍在处理中时，下一步必须写明下次同任务查询的等待时间；自动轮询到期时，必须说明需用户明确要求继续查询。任务响应未返回模型或任务 ID 时明确写“未返回”，绝不猜测。
