---
name: puretokens-image
description: 通过当前 Pure Tokens 连接生成图片、选择图片模型、数量或尺寸时使用。
---

# Pure Tokens Image

直接调用固定的 Pure Tokens Images API：`https://api.puretokensx.com`。每个请求都使用完整 URL；不得用未认证的通用 Fetch/WebFetch 猜测认证。Claude Code、Codex、WorkBuddy、Gemini CLI、Grok Build 和 OpenCode 都使用受管直连运行器；它只从当前宿主的固定已配置位置狭义匹配 Pure Tokens 凭据，且只在内存中为这一次固定 Images API 请求构造认证头。该配置来源只用于绑定凭据，绝不作为请求目标复用。WorkBuddy 可以匹配固定 origin 下无 query、无 fragment 的 `/v1` 或 `/v1/...` 单模型资源 URL；其他宿主仍只接受各自规定的精确连接形式。不得显示、复制、保存、索取或输出 API Key、Base URL 或完整宿主配置；不得调用 MCP、本地代理、sidecar、备用服务或手工认证请求。

根据当前宿主运行对应命令：Claude Code：`node ~/.claude/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host claude-code --method <GET|POST> --path <允许的固定路径>`；Codex：`node ~/.agents/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host codex --method <GET|POST> --path <允许的固定路径>`；WorkBuddy：`node ~/.workbuddy/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host workbuddy --method <GET|POST> --path <允许的固定路径>`；Gemini CLI：`node ~/.gemini/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host gemini-cli --method <GET|POST> --path <允许的固定路径>`；Grok Build：`node ~/.grok/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host grok-build --method <GET|POST> --path <允许的固定路径>`；OpenCode：`node ~/.config/opencode/skills/.puretokens-runtime/puretokens-direct-api.mjs request --host opencode --method <GET|POST> --path <允许的固定路径>`。所有受管宿主的 JSON POST 都必须把唯一 JSON body 编码为有界、规范的 UTF-8 Base64 参数并使用 `--json-base64 <值>`；已声明的原生附件 multipart POST 也必须把唯一受限 JSON 文件描述对象编码为 Base64 并使用 `--multipart-base64 <值>`，其中只允许当前请求明确附带的绝对文件路径、声明字段与普通文本字段。不得对任何宿主使用 `--json-stdin`、`--multipart-stdin`、管道、重定向或 here-document。Base64 只承载当前这一次请求的 JSON，不含 Key；运行器仅在内存中解码。完成后读取内容时，传 `--output-file <新的绝对本地路径>` 取得原生字节；仅在运行器实际交付该文件时才报告为已交付。Trae 目前只有手动连接配置，尚无批准的本地凭据读取契约；在 Trae 计费前停止并说明无法安全执行已认证 Images API 请求，绝不改用通用 Fetch、手工读取 Key 或猜测配置。不得打印运行器输出中的请求头、凭据、Base64 请求值或未清理响应正文。

先读取已安装的 `references/model-selection.json`、`references/execution-contract.json`、`references/task-receipt.json` 和 `references/behavior-scenarios.json`；它们是模型别名、请求/结果约束、用户回执和特殊场景的绑定规则。

## 请求理解与提示词

- 当前用户请求就是完整工作范围。海报、封面、插画等视觉设计请求不等于要求地图、地理资料、合规研究或历史作品检索；除非用户明确要求真实地图、地理信息、文件检索或质量复查，否则不得加载无关 Skill、搜索工作区/历史会话/旧生成文件、读取旧图片，或以“先参考一下”“再确认构图”为由扩展任务。只可读取本 Skill 的规定参考资料和当前用户明确提供的输入。
- 先区分用户要的是：没有输入图片的文生图、以图片作为风格/构图参考，还是修改一张现有图片。用户明确说“参考”“按这个风格/构图”时按参考输入处理；明确说“修改”“替换”“删除”“保留主体”等时按编辑处理。用户给了图片、URL 或附件却未说明角色时，在计费前只问“它是参考图还是待编辑图片？”，不得猜测。
- 当前请求附带的**原生本地图片附件**被用户指定为视觉参考时，优先检查已安装资料或为该明确需求按需读取的目录是否声明可用的 `image_edit` multipart operation。若已声明，使用该 operation 的精确路径、字段、数量和 `multipart_file` transport 直接提交当前附件；提示词仍明确该图片是风格、构图或主体参考，而不是擅自改写用户目标。`image_edit` 是当前 Images API 为本地图片参考/修改声明的传输操作名，不改变用户所表达的参考或编辑意图。只有该 operation 不存在、附件数量越界或宿主不能交付当前附件字节时，才在 `POST` 前停止，并请用户提供该图片的公网 HTTPS URL，或在**新的用户请求**中明确表示忽略附件、改为纯文生图。不得查看、描述、概括、转写、复刻或从无法传输的附件推断任何内容后再把它塞进文生图提示词；不得生成 URL、上传、搜索历史图片或静默降级。
- `n` 只表示**同一完整提示词的变体**，绝不代表多项不同资产。用户一次列出海报、头像和横幅等不同简报时，不得把它们塞进 `n`、拆成多次提交或静默丢弃项目；请用户先确认本次要生成的第一项，再由后续明确请求分别生成其余项。
- 将用户的自然语言整理成简洁、可执行的图片提示词：保留已给出的用途、主体/场景、风格、构图、光线或色彩、材质、逐字文本和限制。提示词已具体时只整理，不添加人物、物体、品牌、标语、色彩或叙事；提示词笼统时可补足不改变意图的构图或成片要求。除非缺少的信息会影响“文生 / 参考 / 编辑”这一操作选择，否则不要为了创作细节反复追问。
- 对参考图，在提示词中说明它是风格、构图或主体参考；对编辑，明确“只改什么、必须保留什么”。用户给出的文字须逐字保留并用引号标出；不得承诺模型一定能无误渲染文字。
- 提示词中的“高清”“透明背景”“局部蒙版”“精确尺寸”等愿望，不等于 API 参数可用。仍只按精确模型已声明的字段、值和 operation 传参；未声明时按本 Skill 的计费前引导处理。

## 模型与提交

- 普通文生图不得在提交前读取 `GET https://api.puretokensx.com/v1/media/models`。未指定模型时，普通文生图、当前请求明确附带的本地图片参考和本地图片编辑都优先使用已安装选择中的 `gpt-image-2`；有本地图片时仍必须使用该模型已声明的 `image_edit` multipart operation。明确 `image2` 也使用该模型。用户给出别名时，仅在 `model-selection.json` 中唯一匹配时解析为精确 ID；零个或多个候选时，列出候选精确 ID 并请用户选择，绝不猜测或自动切换模型。
- 用户给出精确模型 ID 时，保留并直接提交；即使该 ID 不在已安装快照中，也不得为了确认模型、权限或 capability 而进行目录预检。对未在快照中的精确 ID，只能提交核心 `model`、`prompt`、`async: true` 字段；用户同时要求快照未声明的可选参数或媒体操作时，才可在提交前读取一次实时目录以确认如何满足该明确要求。
- 单次 `POST https://api.puretokensx.com/v1/images/generations` 始终传 `model`、`prompt` 和 `async: true`。图片的 `n`、`size`、`image_size`、`aspect_ratio`、`width`、`height`、`strength` 及其他可选字段，只有已安装的该精确模型 `parameterSchema.properties` 明确声明字段和值后才传递；只在该快照明确给出时使用 `default`。同时严格执行该模型 `constraints`：`requires_together` 声明 `width` 与 `height` 必须成对时，用户只给一边就停止并请其补齐；`size_expression_precedence` 声明多个尺寸表达的优先顺序时，只发送用户已给表达中优先级最高的一种，不混传较低优先级表达，并在回执中说明实际提交的尺寸表达。快照未声明所需字段、值、参考输入或编辑 operation 时，可读取一次实时目录；仍无资料或不兼容时，列出已声明选项并请用户重选，绝不猜测、删改用户参数或拆单。
- 一个用户请求只能有一次付费提交，绝不拆单。只有 `n` 已由已安装模型资料或按需读取的实时目录声明为允许的整数时才可传递；未指定时不编造 `n`。若用户要求多张但资料只声明 `n: 1` 或未声明 `n`，说明该模型当前一次只支持一张，请用户改为一张或明确选择支持所需数量的模型。
- `200cm × 230cm` 等物理尺寸不能传给 `n`、`size` 或其他 API 字段。说明不能精确保证物理尺寸，并列出已安装资料或按需目录查询实际声明的像素 `size`、语义 `image_size` 和 `aspect_ratio` 选项；若三者均未声明，明确说明没有可用规格资料并请用户改用纯提示词或选择另一个模型。
- 用户提供的是当前请求中明确给出的**公网 HTTPS 图片 URL**并说明是视觉参考时，先确认它是 `https://` URL；不是时说明只接受公网 HTTPS URL，停止并请用户提供合规 URL。拒绝含 URL 凭据、`localhost`、`.local`，或显式环回、私网、链路本地 IP 的 URL；Skill 不做 DNS 解析或可访问性探测，公网可读性仍由 API 验证。只有已安装模型资料或为该明确参考需求按需读取的目录同时声明对应请求属性和允许的 URL transport 时，才可在 JSON body 的 `POST https://api.puretokensx.com/v1/images/generations` 中传原始 URL。`string[]` 属性传 URL 数组；`string` 或 `json` 属性只传用户明确给出的原始值。多个可能 URL 字段、或用户未说明单图/多图输入模式时，列出已声明字段并请用户选择；不得依据模型名或附件数量猜测字段，更不能下载、探测、转存或改写 URL。
- 用户明确要求**编辑**公网 URL 图片时，不能把一般参考字段当作编辑能力。只有已安装资料或按需目录中的 `image_edit` operation 明确声明 `application/json`、精确请求路径、精确 URL 字段及公网 URL transport 时，才可按该声明发起一次 JSON 编辑请求；路径只可为 `/v1/images/generations` 或 `/v1/images/edits`，并附带 `model`、`prompt`、`async: true`。若只声明参考字段而未声明 JSON 编辑 operation，说明该模型当前只能按参考图使用，询问用户是否改为参考图生成；未经确认不得降级或提交。
- 用户提供的是当前请求中明确附带的原生图片附件/运行环境媒体对象，且明确其为参考图或待编辑图片时，只有已安装模型资料或为该明确需求按需读取的目录中 `image_edit` operation 明确声明 `POST`、`multipart/form-data`、必需字段、允许数量和 `multipart_file` transport，且路径为 `/v1/images/generations` 或 `/v1/images/edits` 时才可继续。Skill 将该 operation 的相对路径与固定 API origin 组合成完整 URL，绝不把编辑路径固定为 generations。Skill 只把附件原生字节随这一次 Images API 请求发送。资料缺失、操作不支持或运行环境无法随请求传递附件字节时，在计费前停止并说明已声明输入；不得生成 URL、file ID 或调用独立上传接口，也不得静默降级为文生图。
- 原生附件图片参考或编辑的 multipart 请求严格使用资料声明的路径、字段、数量和 transport，并附带 `model`、`prompt`、`async: true` 和声明允许的可选字段；参考请求的 prompt 必须保留“参考”语义，编辑请求必须保留“只改什么、保留什么”语义。公网 URL JSON 编辑则严格使用上条所述的 `image_edit` operation。mask、透明背景、未声明的多参考图或任何其他扩展必须停止并说明未声明；不得伪装成功。
- 只有用户明确问“当前有哪些模型/参数/操作”，请求的可选字段或图片输入 operation 不在已安装资料中、已安装 operation 与用户明确输入的 transport 不兼容，或提交被 API 以模型、参数或 capability 问题拒绝后需要解释时，才可读取一次 `GET https://api.puretokensx.com/v1/media/models`。这不是普通生图的前置条件；目录读取失败绝不能阻止本可直接提交的核心文生图，也不得触发自动重试或重提。

## 任务与交付

- 运行器返回可解析的 `runtime_error` 时，必须按其 `phase` 处理，不得把它当作“无输出”。`validation` 说明 POST 尚未开始：使用安全失败回执报告该本地错误，`task_id` 为未返回，且不得说任务已创建或收费；`submission` 说明 POST 可能已开始：按提交结果未知处理，绝不重提；`read` 或 `response` 只代表当前固定读取或响应阶段失败，严格保留任何已验证的同一 `task_id`，不得创建替代任务。运行器总执行期限为 90 秒；收到其超时结构化回执时同样按上述 phase 处理。
- 若宿主在运行器启动前因系统策略、审批策略或网络执行权限而拦截命令，按本地计费前失败处理：`failure_phase` 必须为 `validation`，`task_id`、`http_status` 和 `api_error_code` 均为“未返回”。明确说明“当前会话未允许外部网络执行，Images API 未执行，未创建图片任务”，不得把它写成 `submission`、不得声称扣费、退款或 API 拒绝，也不得重复 POST。`next_action` 只能引导用户在允许外部网络请求且可批准网络操作的宿主会话中重试；Skill 不能请求、提升或绕过宿主系统权限。
- 每个新的图片请求在当前用户回合中只允许一次 `POST` 运行器调用。运行器已开始执行后，若宿主工具只返回执行回显异常、空输出、截断输出或无法解析的输出，必须把提交结果视为未知：不得重复该 `POST`、不得提交替代任务、不得为了“拿到可查询的 ID”再试一次。若这一次实际返回了可验证的 `task_id`，只继续该任务；否则只输出一次“提交结果未知、任务 ID 未返回”的安全失败回执，然后**立即结束当前回复**；不得再调用工具、查询、轮询、重新评估或延长本轮。后续用户只说“继续”“再试”或类似表述但未给出任务 ID 时，说明无法继续原任务，并请其明确确认要创建一次新的付费图片任务；只有该确认后的新用户回合才可按原需求提交一次新的 `POST`。不得声称任务未创建、未收费或已退款。
- 文生图和已声明的图片编辑都始终传 `async: true`。回执中的 `task_id` 是统一称呼，不要求响应 JSON 使用同名字段：已安装模型 lifecycle 声明 `create.idField` 时，严格从该顶层字段取值；未声明时只可从顶层 `task_id` 或 `id` 取值。不得从 URL、嵌套对象、提示词或任意其他字段推导任务 ID。两者都未返回时，报告“任务 ID 未返回，无法安全继续查询”；不得轮询、下载、自动重提或宣称未创建任务。
- 有规范化 `task_id` 时，只读同一任务的 `https://api.puretokensx.com/v1/images/{task_id}` 与 `https://api.puretokensx.com/v1/images/{task_id}/content?index=N`。将任务 ID 作为不透明值进行 URL path-segment 编码；用户给出 URL、多个 ID 或空值时，请其提供一个任务 ID，不得提取、拼接或猜测。不得查询或提交另一任务。
- 每次状态响应都先检查顶层 `reconciliation_required`。它为 `true` 时优先于任何 `pending`、`queued`、`running` 或 `in_progress` 状态：立即停止常规自动轮询，按 reconciliation 回执保留同一 `task_id`，不得提交替代任务、切换模型、推断退款或把它说成普通失败。用户之后明确要求查看该任务时，最多读取一次同一任务状态，不再开启新的自动轮询窗口。
- 对已安装 lifecycle，严格按其 `pendingStatuses`、`successStatuses`、`failureStatuses` 分类状态；资料未声明 lifecycle 时，仅将 `pending`、`queued`、`running`、`in_progress` 视为处理中，将 `completed`、`succeeded`、`success` 视为成功，将 `failed`、`cancelled`、`canceled`、`expired`、`error` 视为失败。状态字段缺失或值不在上述已知集合时，报告原始状态为未识别并停止自动轮询，只有用户明确要求继续该任务才可再读取一次状态；不得把未知状态当成功、失败或可重提。
- 请求 `n` 张时，内容索引从 `0` 开始，必须按 `0` 到 `n-1` **顺序逐个**读取。仅在任务终态成功后读取内容；每项原生字节交付给宿主并完成写入/交付后，才可读取下一项。不得并发预取、重复下载已交付内容、把媒体字节复制进提示词或会话缓存。只有每个请求索引都交付了原生图片字节才算成功；部分成功时报告已交付和缺失的索引，只能继续读取同一任务缺失内容，绝不重提。只知道任务 ID 而无法从当前会话确定原始 `n` 时，先请用户说明原始请求数量；不得猜测索引范围或只交付第一张后声称完成。
- 提交获得 `task_id` 后，才可在**新提交或用户明确继续查询该 `task_id` 的当前用户请求**内开始自动轮询。对同一任务始终最多一个在途状态请求，不创建后台计时器、队列或持续轮询。状态读取返回 HTTP `429` 时，若有有效正数 `Retry-After` 且仍在本轮 120 秒预算内，则等待该时长后继续读取**同一任务**；不得为此重提。没有有效 `Retry-After` 或等待会超出预算时，报告限流与 `task_id` 并停止本轮。非 429 的 5xx、传输错误或超时才立即停止自动等待。正常处理中，若无可用 `Retry-After`，依次等待 3、6、12、24、30 秒，此后最多每 30 秒读取一次。每个轮询窗口最多读取状态 6 次；到期仍为 pending/running 时，报告当前状态并停止自动等待。用户明确要求继续查询同一 `task_id` 时，才开启一个新的、同样有界的**同任务**轮询窗口；绝不查询或提交其他任务。不得把轮询到期或读取错误当作失败或重提理由。
- pending/running 时报告当前状态并在上述时间表内继续同一任务；失败、取消或过期时使用下方的安全失败回执，要求用户明确发起新请求后才可再次提交。
- 若提交响应明确因模型、参数或 capability 被拒绝且没有任务 ID，可按需读取一次目录，展示当前候选或允许值，并要求用户明确发起修正后的新请求；绝不自动重提。若用户指定的精确模型不在该认证目录中，说明“当前连接未返回该模型，无法提交”。若用户预期有该模型权限，引导其在 Pure Tokens 客户端配置中勾选包含该精确模型的分组，创建或选择覆盖所选分组的受管 Key，执行“验证并应用”，然后新开当前宿主会话后明确重试。不得猜测分组名称、Base URL、Key 或归属；除非认证 API 明确返回模型到分组的映射，否则不得声称哪个分组包含该模型。遇到 `429` 时，在安全失败回执中报告有效的 `Retry-After` 秒数（如有）并请用户等待后明确重试；不得自动等待或重试。遇到 5xx、传输错误或超时且未返回任务 ID 时，报告“提交结果未知”，不得断言任务未创建或自动重提；若返回了任务 ID，只能继续查询该任务。
- 状态超时或内容交付失败时，使用安全失败回执说明不确定/失败状态；只可让用户选择继续查询或重试同一任务内容，绝不自动重提、换模型、换参数或拆分付费请求。
- 只有运行环境交付原生图片字节才算成功。任务 ID、状态文字、HTML/SVG、组件或远程 URL 都不算成功。
- 直接 API 请求在任务被接受前失败时，使用安全失败回执报告 API HTTP 状态（如有）和安全错误信息，并说明尚未创建任务；不得臆测用户的 Base URL、认证或路由配置，不得索取凭据或改走备用路径。
- 每次提交、继续查询、完成或失败，都按 `task-receipt.json` 报告：已返回的精确模型 ID、任务 ID、返回状态、请求数量、尺寸/参数、已交付数量（完成时）和下一步。任务仍在处理中时，下一步必须写明下次同任务查询的等待时间；自动轮询到期时，必须说明需用户明确要求继续查询。失败还必须包含下方所述字段。任务响应未返回模型或任务 ID 时明确写“未返回”，绝不猜测。
- 任务完成时，只有宿主实际确认的原生字节、附件或本地文件路径可以报告为已交付；若宿主没有返回持久路径，只说明已通过当前宿主交付原生图片，不编造下载目录、预览或文件名。
- 当前请求的图片内容已成功交付后，立即给出一次完成回执并结束当前回复。不得自动检查成片质量、读取或打开成片、搜索工作区或历史会话中的参考图、调用其他 Skill，或创建另一笔生成任务；只有用户在后续新请求中明确要求这些动作时才可处理，且任何新的付费生成仍须作为一次新的明确请求。

## 安全失败回执

- 所有计费前校验、提交、状态读取和内容交付失败，使用 `failure_phase`（`validation`、`submission`、`status` 或 `content`）、`api_error_code`、`http_status`（API 未返回时写“未返回”）、简洁的 `error_message` 和可执行的 `next_action`。只有 API 明确返回公开机器码时才原样写入 `api_error_code`；否则写“未返回”，绝不由英文报错、模型名或上游现象推断代码。HTTP `429` 且 API 返回有效正数 `Retry-After` 时，额外给出 `retry_after_seconds`。
- `error_message` 只能使用经清理的公开 API 提示，或固定的本地说明。绝不展示原始响应正文、上游/提供方标识、内部主机或 URL、堆栈、请求头、请求体、凭据/会话数据、用户参考 URL 或附件字节。不得把错误归因于用户的认证、Base URL、路由或客户端配置。
