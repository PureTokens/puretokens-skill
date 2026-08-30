---
name: puretokens_image
description: 通过当前 Pure Tokens 连接生成图片、选择图片模型、数量或尺寸时使用。
---

# Pure Tokens Image

直接调用固定的 Pure Tokens Images API：`https://api.puretokensx.com`。每个请求都使用完整 URL；运行环境会为已配置的 Pure Tokens 请求自动携带认证。Skill 绝不读取、扫描、展示、复制或索取 API Key、Base URL、认证文件或宿主配置，也不构造认证头、调用 MCP、本地代理或备用服务。

先读取已安装的 `references/model-selection.json`、`references/execution-contract.json`、`references/task-receipt.json` 和 `references/behavior-scenarios.json`；它们是模型别名、请求/结果约束、用户回执和特殊场景的绑定规则。

## 模型与提交

- 普通文生图不得在提交前读取 `GET https://api.puretokensx.com/v1/media/models`。未指定模型时用已安装选择中的 `gpt-image-2`；明确 `image2` 也使用该模型。用户给出别名时，仅在 `model-selection.json` 中唯一匹配时解析为精确 ID；零个或多个候选时，列出候选精确 ID 并请用户选择，绝不猜测。
- 用户给出精确模型 ID 时，保留并直接提交；即使该 ID 不在已安装快照中，也不得为了确认模型、权限或 capability 而进行目录预检。对未在快照中的精确 ID，只能提交核心 `model`、`prompt`、`async: true` 字段；用户同时要求快照未声明的可选参数或媒体操作时，才可在提交前读取一次实时目录以确认如何满足该明确要求。
- 单次 `POST https://api.puretokensx.com/v1/images/generations` 始终传 `model`、`prompt` 和 `async: true`。图片的 `n`、`size`、`image_size`、`aspect_ratio`、`width`、`height`、`strength` 及其他可选字段，只有已安装的该精确模型 `parameterSchema.properties` 明确声明字段和值后才传递；只在该快照明确给出时使用 `default`。快照未声明所需字段、值、参考输入或编辑 operation 时，可读取一次实时目录；仍无资料或不兼容时，列出已声明选项并请用户重选，绝不猜测、删改用户参数或拆单。
- 一个用户请求只能有一次付费提交，绝不拆单。只有 `n` 已由已安装模型资料或按需读取的实时目录声明为允许的整数时才可传递；未指定时不编造 `n`。若用户要求多张但资料只声明 `n: 1` 或未声明 `n`，说明该模型当前一次只支持一张，请用户改为一张或明确选择支持所需数量的模型。
- `200cm × 230cm` 等物理尺寸不能传给 `n`、`size` 或其他 API 字段。说明不能精确保证物理尺寸，并列出已安装资料或按需目录查询实际声明的像素 `size`、语义 `image_size` 和 `aspect_ratio` 选项；若三者均未声明，明确说明没有可用规格资料并请用户改用纯提示词或选择另一个模型。
- 用户提供的是当前请求中明确给出的**公网 HTTPS 图片 URL**时，先确认它是 `https://` URL；不是时说明只接受公网 HTTPS URL，停止并请用户提供合规 URL。拒绝含 URL 凭据、`localhost`、`.local`，或显式环回、私网、链路本地 IP 的 URL；Skill 不做 DNS 解析或可访问性探测，公网可读性仍由 API 验证。只有已安装模型资料或为该明确参考需求按需读取的目录同时声明对应请求属性和允许的 URL transport 时，才可在 JSON body 中传原始 URL。`string[]` 属性传 URL 数组；`string` 或 `json` 属性只传用户明确给出的原始值。多个可能 URL 字段、或用户未说明单图/多图输入模式时，列出已声明字段并请用户选择；不得依据模型名或附件数量猜测字段，更不能下载、探测、转存或改写 URL。
- 用户提供的是当前请求中明确附带的原生图片附件/运行环境媒体对象时，只有已安装模型资料或为该明确编辑需求按需读取的目录中 `image_edit` operation 明确声明 `POST`、`multipart/form-data`、必需字段、允许数量和 `multipart_file` transport，且路径为 `/v1/images/generations` 或 `/v1/images/edits` 时才可继续。Skill 将该 operation 的相对路径与固定 API origin 组合成完整 URL，绝不把编辑路径固定为 generations。Skill 只把附件原生字节随这一次 Images API 请求发送；Pure Tokens 会在网关内部短期 R2 暂存、确认供上游读取的 HTTPS URL 后映射请求，且不会把该 URL 返回给用户。资料缺失、操作不支持或运行环境无法随请求传递附件字节时，在计费前停止并说明已声明输入；不得生成 URL、file ID 或调用独立上传接口，也不得静默降级为文生图。
- 图片编辑或参考图重绘的 multipart 请求严格使用资料声明的路径、字段、数量和 transport，并附带 `model`、`prompt`、`async: true` 和声明允许的可选字段。mask、透明背景、未声明的多参考图或任何其他扩展必须停止并说明未声明；不得伪装成功。
- 只有用户明确问“当前有哪些模型/参数/操作”，请求的可选字段或附件操作不在已安装资料中，或提交被 API 以模型、参数或 capability 问题拒绝后需要解释时，才可读取一次 `GET https://api.puretokensx.com/v1/media/models`。这不是普通生图的前置条件；目录读取失败绝不能阻止本可直接提交的核心文生图，也不得触发自动重试或重提。

## 任务与交付

- 文生图和已声明的图片编辑都始终传 `async: true`。回执中的 `task_id` 是统一称呼，不要求响应 JSON 使用同名字段：已安装模型 lifecycle 声明 `create.idField` 时，严格从该顶层字段取值；未声明时只可从顶层 `task_id` 或 `id` 取值。不得从 URL、嵌套对象、提示词或任意其他字段推导任务 ID。两者都未返回时，报告“任务 ID 未返回，无法安全继续查询”；不得轮询、下载、自动重提或宣称未创建任务。
- 有规范化 `task_id` 时，只读同一任务的 `https://api.puretokensx.com/v1/images/{task_id}` 与 `https://api.puretokensx.com/v1/images/{task_id}/content?index=N`。将任务 ID 作为不透明值进行 URL path-segment 编码；用户给出 URL、多个 ID 或空值时，请其提供一个任务 ID，不得提取、拼接或猜测。不得查询或提交另一任务。
- 对已安装 lifecycle，严格按其 `pendingStatuses`、`successStatuses`、`failureStatuses` 分类状态；资料未声明 lifecycle 时，仅将 `pending`、`queued`、`running`、`in_progress` 视为处理中，将 `completed`、`succeeded`、`success` 视为成功，将 `failed`、`cancelled`、`canceled`、`expired`、`error` 视为失败。状态字段缺失或值不在上述已知集合时，报告原始状态为未识别并停止自动轮询，只有用户明确要求继续该任务才可再读取一次状态；不得把未知状态当成功、失败或可重提。
- 请求 `n` 张时，内容索引从 `0` 开始，必须按 `0` 到 `n-1` **顺序逐个**读取。仅在任务终态成功后读取内容；每项原生字节交付给宿主并完成写入/交付后，才可读取下一项。不得并发预取、重复下载已交付内容、把媒体字节复制进提示词或会话缓存。只有每个请求索引都交付了原生图片字节才算成功；部分成功时报告已交付和缺失的索引，只能继续读取同一任务缺失内容，绝不重提。只知道任务 ID 而无法从当前会话确定原始 `n` 时，先请用户说明原始请求数量；不得猜测索引范围或只交付第一张后声称完成。
- 提交获得 `task_id` 后，才可在**新提交或用户明确继续查询该 `task_id` 的当前用户请求**内开始自动轮询。对同一任务始终最多一个在途状态请求，不创建后台计时器、队列或持续轮询。若状态响应有有效的正数 HTTP `Retry-After` 秒数且不超过剩余自动轮询预算，优先按它等待；否则依次等待 3、6、12、24、30 秒，此后最多每 30 秒读取一次。每个轮询窗口从本次提交或明确继续查询起算，最多 120 秒且最多读取状态 6 次；到期仍为 pending/running 时，报告当前状态并停止自动等待。状态读取遇到限流、5xx、传输错误或超时，也立即停止自动等待并如实报告。用户明确要求继续查询同一 `task_id` 时，才开启一个新的、同样有界的**同任务**轮询窗口；绝不查询或提交其他任务。不得把轮询到期或读取错误当作失败或重提理由。
- pending/running 时报告当前状态并在上述时间表内继续同一任务；失败、取消或过期时报告返回错误，要求用户明确发起新请求后才可再次提交。
- 若提交响应明确因模型、参数或 capability 被拒绝且没有任务 ID，可按需读取一次目录，展示返回的当前候选或允许值，并要求用户明确发起修正后的新请求；绝不自动重提。遇到 `429` 时，报告 `Retry-After`（如有）并请用户等待后明确重试；不得自动等待或重试。遇到 5xx、传输错误或超时且未返回任务 ID 时，报告“提交结果未知”，不得断言任务未创建或自动重提；若返回了任务 ID，只能继续查询该任务。
- 状态超时或内容交付失败时，说明不确定/失败状态；只可让用户选择继续查询或重试同一任务内容，绝不自动重提、换模型、换参数或拆分付费请求。
- 只有运行环境交付原生图片字节才算成功。任务 ID、状态文字、HTML/SVG、组件或远程 URL 都不算成功。
- 直接 API 请求在任务被接受前失败时，如实报告 API 返回的状态或错误，并说明尚未创建任务；不得臆测用户的 Base URL、认证或路由配置，不得索取凭据或改走备用路径。
- 每次提交、继续查询、完成或失败，都按 `task-receipt.json` 报告：已返回的精确模型 ID、任务 ID、返回状态、请求数量、尺寸/参数、已交付数量（完成时）和下一步。任务仍在处理中时，下一步必须写明下次同任务查询的等待时间；自动轮询到期时，必须说明需用户明确要求继续查询。任务响应未返回模型或任务 ID 时明确写“未返回”，绝不猜测。
