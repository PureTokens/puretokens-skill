<p align="center">
  <a href="./CHANGELOG.md">English</a> · <strong>中文</strong>
</p>

# 更新日志

## 0.12.0 — 2026-08-30

- 普通文生图和文生视频现在使用已安装的版本化模型选择资料并直接提交，不再在每个任务前执行一次 `GET /v1/media/models` 预检。
- 实时媒体目录改为按需读取：仅用于用户明确查询当前模型、安装资料没有所需参数或媒体操作，或在模型/参数/capability 被拒后解释一次。目录读取失败不会阻止原本有效的核心生成请求，也不会触发自动重提。
- 强化异步任务处理：只规范化已声明的顶层任务 ID，按 lifecycle 分类状态（含 `queued`、`in_progress`），未识别状态即停止自动轮询；提交结果未知、限流、拒绝、无效续查 ID 及多图数量缺失均会给出明确引导，绝不猜测或自动重提。
- 公网参考 URL 现会拒绝带凭据、本地主机及显式环回/私网/链路本地 IP；Skill 仍不会探测、下载或转存用户媒体。

## 0.11.0 — 2026-08-29

- 客户端下载页的 Agent 安装提示词现已在第一句明确唯一 GitHub 官方仓库及 `main` 分支，用户复制后在 Agent 开始执行前就能得到无歧义的安装来源。

- 新增 `puretokens_update` 与 `puretokens-skill sync --target <directory>` 命令。用户明确要求更新时会先校验新克隆的官方 `main`，安装缺失的官方 Skill，只原子升级同名受管 Skill；任一非受管同名目录都会让同步在改动前停止。
- 增加独立的更新行为场景、本机更新执行契约、Claude Desktop 打包引导，以及同步成功和冲突安全拒绝的测试覆盖。

- 移除相对路径和宿主能力前置契约，改为固定直连 API 契约。所有面向 API 的专项 Skill 现在都调用完整的 `https://api.puretokensx.com` URL；不依赖 MCP、本地代理、sidecar、用户配置的 Base URL 或宿主专用的相对路径执行器。
- 认证仍由当前运行环境管理。Skill 绝不读取、扫描、展示、复制、索取、保存或构造 API Key、Base URL、认证头或客户端配置。
- 生图、生视频、模型查询、API 身份检查和余额查询统一改为固定 API origin；图片和视频的状态/内容读取也使用完整固定 URL，profile 声明的 multipart operation 路径只会与该 origin 组合。
- 用 `references/direct-api-execution-contract.json` 替代旧的宿主原生执行矩阵，并同步更新校验、manifest、行为场景、双语 README 与测试。

- 已在两份 README 顶部恢复仓库内受版本控制的官方 Skills Hero 图。
- 将客户端下载页的 Agent 安装提示词重写为简洁、分行的执行清单：自动识别当前受支持宿主、选择目录、同步六个 Skill、复验结果，同时保留既有安全边界。
- 已将公开媒体目录同步到当前包含 18 个模型的基础目录，新增 `grok-imagine-image-2.0`、`grok-imagine-video-1.5`、`wan3.0-video`、`wan3.0-video-prime`。为 Wan3 增加了无歧义别名；可能混淆的 Grok 1.5 名称仍要求精确 ID。
- 已收紧媒体资源使用：轮询只在提交所在回合或用户明确继续同一任务的回合内进行，同一任务只有一个状态请求在途，不创建后台计时器或队列，使用有上限的自适应退避，并在限流、服务端错误、传输错误或超时时立即停止。任务仍在处理时，只有用户明确继续同一 `task_id` 才会开启下一轮有界轮询。内容交付只在终态成功后进行，一次一个，多图按索引顺序读取，不预取、不重复读取，也不写入 Skill 状态或日志。
- 视频分辨率会遵循实时 profile 的 `resolution_by_mode`；图片编辑示例改为当前实际声明 `image_edit` 的模型。
- 图片和视频 Skill 已与当前认证媒体 profile 契约对齐。每个新图片任务（包括默认模型）都会读取实时目录；所有模型级图片参数都改为由 profile 控制，不再使用共享数量或画布清单。
- 增加按 profile 控制的公网参考输入：用户明确提供的公网 HTTPS URL、file ID、voice ID 只能写入精确声明的字段和 transport。原生附件仍是单次请求中的 `multipart_file`，网关内部完成私有 R2 暂存。
- 视频提示词是否必填改为遵循精确实时 profile 及其声明的单参考例外。原生附件组合仍需明确组合 operation；公网 URL/ID 字段必须遵守已声明的互斥和模式限制。

## 0.9.0 — 2026-08-25

- 新增 `puretokens_connection`：只读检查当前连接。它只调用 `GET /v1`，核对现有 API 声明的 `status: "ok"`、`name: "Pure Tokens API"` 和 `base_url: "/v1"`，不读取或展示真实 Base URL、凭据和宿主配置。
- 新增 `puretokens_models`：只读查询当前连接的认证模型目录，只展示 `GET /v1/media/models` 实际返回的精确模型 ID、能力、公开参数属性、非请求字段的条件限制和媒体操作。
- 新增按需求给出 profile 兼容模型清单的能力，例如图生视频、参考媒体、指定时长或画幅；只匹配实际返回的 capability 与 `input_schema`，不编造质量、价格、速度或可用性排序。

## 0.8.4 — 2026-08-25

- 为视频 Skill 增加按认证后实时 profile 开启的参考视频、参考音频和视频编辑。`reference_video`、`reference_audio` 仍走 `POST /v1/videos`；`video_edit` 只在目录明确声明时走 `POST /v1/videos/edits`。
- 所有新增输入只接收用户在当前请求附带的原生媒体字节，且仅接受 profile 声明的 `multipart_file`；网关内部短期 R2 暂存与 URL 映射不会暴露给用户。组合媒体没有明确 operation 时会在计费前停止，不拆单、不丢弃附件。

## 0.8.3 — 2026-08-25

- 修正按 profile 开启的图片编辑：与文生图一样使用 `POST /v1/images/generations` 和 `multipart/form-data`。
- 所有当前请求的已支持图片输入统一使用单次 Images/Videos API 请求中的 `multipart_file`。Pure Tokens 网关在内部负责短期 R2 暂存和公网可读校验；Skill 不暴露或伪造供上游使用的 URL。

## 0.8.2 — 2026-08-24

- 新增认证后按 profile 开启的图片编辑：精确模型发布对应输入操作时，使用 Images API；同时支持通过 `POST /v1/videos` 的图生视频和参考图视频。
- 媒体附件改为失败即停止：只使用当前请求中用户明确提供的媒体和实时 profile 声明的精确 transport；绝不下载、转存、伪造 URL 或 file ID、静默回退为文生，也不读取凭据。
- 为上述媒体输入流程补齐请求操作回执、随安装提供的契约、行为场景、清单和双语文档。

## 0.8.1 — 2026-08-24

- 明确媒体异步任务的有界轮询：优先遵循有效的 `Retry-After`；否则同一任务状态查询依次等待 2、3、5、8 秒，之后每次等待 15 秒。生图自动轮询最多 120 秒，生视频最多 300 秒；到期会提示用户明确继续同一任务，不会被视为失败或重提理由。

## 0.8.0 — 2026-08-24

- 让多图交付可确定：请求 `n` 张时只读取零基内容索引 `0..n-1`；只有每个请求索引都交付原生字节才成功；部分结果会报告已交付/缺失索引，且不会重提任务。
- 加入逐模型参数资料规则。非默认图片模型的可选参数、以及视频的全部可选参数，都必须由精确模型当前认证后的实时 `input_schema` 明确声明；仅提示词的请求在契约允许时仍可使用；资料缺失会在计费前停止，绝不猜测。
- 加入宿主原生执行契约和验收矩阵，覆盖每个支持宿主的已认证相对 HTTP、JSON 任务响应、原生媒体字节和同任务继续查询。
- 为余额提供具体的只读契约：使用既有已认证账户会话调用 `GET /api/product/desktop/account/balance`，只报告返回字段；会话不可用时引导至客户端余额视图。
- 加入结构化媒体任务回执、七天模型目录新鲜度发布门槛、schema 校验和覆盖新契约的测试；三个专项 Skill 统一升至 0.8.0。

## 0.7.0 — 2026-08-24

- 为余额、图片和视频加入随 Skill 安装的版本化执行契约与面向用户的行为场景，覆盖别名歧义、目录授权、请求边界、不支持的媒体输入、任务状态、未知提交、超时和原生内容交付，且绝不自动重提。
- 将公开模型目录收敛为唯一别名来源；按图片/视频能力生成的模型选择引用会随对应 Skill 安装，并与双语 README 一起校验。
- 发布诚实的宿主支持矩阵，并移除旧的 WorkBuddy 合并 always-apply 媒体渲染器。Desktop Router Adapter 与专项 Skill 交付现在明确是两个独立问题。

## 0.6.1 — 2026-08-24

- 将 `puretokens_balance`、`puretokens_image`、`puretokens_video` 收敛到当前已配置连接：宿主负责 Base URL、认证、路由、HTTP 执行与原生媒体交付；Skill 只使用相对 API 路径，绝不检查凭据、Base URL、provider 标签或服务归属。
- 移除过时的连接归属拒绝与官网跳转。已配置连接要么执行 Images/Videos API 请求并交付原生字节，要么在计费提交前停止；不切换执行路径，也不做回退。

## 已被 API-first 契约取代的历史记录

以下记录仅描述过去版本。其中出现的 `puretokens_media`、MCP、Direct Cloud、Desktop 受管 Codex 路径、旧环境变量或 WorkBuddy 路由，均不属于当前专项 Skill 的执行契约。

## 0.4.6 — 2026-08-20

- 保留 Pure Tokens Desktop 作为 Codex 和 WorkBuddy 的可选受管交付，同时恢复独立安装 Skill 与 Direct Cloud 执行路径。
- 明确手动配置 Pure Tokens 媒体模型时的第三种宿主原生执行情形：只有宿主能证明精确媒体能力和真实交付时才保留 UI/工具上下文中的显式选择；否则回到正常 MCP 或 Direct Cloud 决策树，且不得重复提交。
- 保持新增媒体模型只要出现在认证后的实时目录即可直接使用、无需等待 Skill 更新，同时发布已登记模型清单和便利别名，让用户能发现 Skill 的能力范围。
- 明确 Direct Cloud 的交付能力前置检查、API Key 范围内的缺模型恢复、仅宿主保存的请求 ID，以及多图通过零基 `/content?index=N` 取回的语义。
- WorkBuddy 生成改为原子替换受管交付，避免已删除源文件留下旧行为；同时保留用户显式选择的内置或手动配置模型，并修正官方仓库地址。

## 0.4.5 — 2026-08-20

- 恢复 Direct Cloud 图片提交策略：图片请求始终传 `async: true`。执行层仍会防御性兼容服务返回的同步 `b64_json` 或 `url`，但 Skill 不会主动请求同步图片生成。

## 0.4.4 — 2026-08-20

- 确认 Codex 只交付由 Desktop 受管的生成式 Skill，已彻底移除已废弃的 Plugin/Marketplace 交付。
- 修复 Direct Cloud 结果模式契约：图片请求不再强制 `async: true`，执行层根据服务真实响应处理并交付同步 `b64_json`/`url` 或异步任务内容。
- 明确 Skill 不亲自执行 I/O：MCP 与宿主的 Direct Cloud 执行层分别负责下载、落盘、预览和真实本机交付证据。两条执行通道都不可用时，Skill 只说明缺少的能力，不把 Desktop 或聊天中粘贴凭据当成前提。
- 强化生成交付验证：Claude Desktop ZIP 会解包并逐文件与共享源做字节级比对，和已有 Codex、WorkBuddy 来源校验保持一致。

## 0.4.3 — 2026-08-20

- 统一 Direct Cloud 凭据为常规的 **API Base URL** 与 **API Key**：使用环境变量的宿主固定映射为 `PURETOKENS_API_BASE_URL` 和 `PURETOKENS_API_KEY`，不再使用容易与登录态混淆的 `PURETOKENS_ACCESS_TOKEN`。

## 0.4.2 — 2026-08-20

- 将 Pure Tokens 专用的 Codex Plugin 交付改为生成出的受管 Skill 交付。Pure Tokens Desktop 现在负责 `~/.codex/skills/puretokens_media` 与独立的 `puretokens-image` MCP 配置；媒体生成不再依赖 Codex Plugin Feature、Marketplace 配置或插件市场解锁。
- 保持 Direct Cloud 不依赖 Desktop，并明确异步多图片交付：完成图片任务声明多个内容项时，按响应顺序先取 `/content`，再使用有界的 `/content?index=N` 获取后续结果。

## 0.4.1 — 2026-08-20

- `puretokens_media` 默认只请求 1 个结果；只有用户明确给出数量且当前执行契约支持时才请求更多结果。一个请求绝不会变成多次生成提交。
- 补全 Direct Cloud 交付契约：同步图片 `data[].b64_json` 和 `data[].url` 会被解码或下载为本地文件，异步图片和所有视频都必须取回 `/content`，只有媒体字节已经写入本机后才可将完成状态报告为成功。预览和“打开文件/文件夹”入口也只能在执行层实际返回时展示。
- 新增覆盖共享 Skill、清单、Agent 入口和引用资料的确定性来源哈希。生成的 Codex Plugin 与 WorkBuddy 交付都携带该哈希；Claude Desktop ZIP 新增 `source-delivery.json`，测试会验证每种交付与共享源一致。

## 0.4.0 — 2026-08-20

- 明确 `puretokens_media` 的执行边界：Skill 负责自然语言策略，MCP 负责类型化本机执行，Desktop Router 是受管传输；具备终端能力的 Agent 可以按 Direct Cloud 契约生成媒体，不需要 Desktop、Router、额外 CLI 或 MCP。Desktop Router 与 Direct Cloud 现在都使用认证 `GET /v1/media/models` 的同一响应形状，再调用同一组图片/视频提交、状态和内容端点。Codex Plugin 是生成出的 Skill 交付物，不会打包或启动 Desktop 受管 MCP。
- 新增官方 Codex Plugin 交付：`puretokens_media` 会在 `puretokens-image` MCP 实际可调用时使用 Desktop 受管 MCP，同时仍可在没有该 MCP 时走 Direct Cloud。
- 收紧可复制的本机 Agent 安装提示词：普通 ChatGPT 对话即使显示 Codex 运行时，也不能被当作具备本机安装权限的 Codex。Agent 必须先确认同时具备终端和本机文件写入权限，否则明确停止。
- 按当前目录更新 Nano Banana 别名：`gemini-3.0-pro-image` 和 `gemini-3.1-flash-image`。

## 0.3.2 — 2026-08-14

- 强化 WorkBuddy 媒体路由：明确 `ToolSearch` 只负责发现工具，所有 Pure Tokens 的目录、生成和结果调用都必须通过 `DeferExecuteTool` 实际执行。
- 禁止用 SVG/HTML 组件、内置媒体工具、搜索或文字声明冒充 Pure Tokens 媒体任务成功；只有 MCP 返回实际模型及原生结果或本机交付信息时才能称为生成成功。

## 0.3.1 — 2026-08-14

- 将 `puretokens_media` 收敛为 Claude Desktop 与 WorkBuddy 共用的唯一媒体行为源。WorkBuddy 的常驻路由载荷现在由该源自动生成，目录、精确模型、单次提交、轮询、交付和失败规则完全一致。
- 移除独立的 WorkBuddy 路由 Skill 源。Pure Tokens Desktop 仍会在点击 **验证并应用** 后自动安装生成后的适配载荷，用户无需手动上传。
- 移除媒体 Skill 中已过期的余额和模型价格工具声明；现在只暴露本地 Sidecar 实际提供的五个媒体 MCP 工具。

## 0.3.0 — 2026-08-14

- 新增 `puretokens_workbuddy_router`：轻量常驻的 WorkBuddy Skill。普通生图、生视频请求会优先进入已配置的 Pure Tokens MCP，而不是 WorkBuddy 内置媒体工具；用户明确指定 WorkBuddy 内置工具时保持该选择。文本、代码等普通请求不受影响，并保持目录优先、精确模型与不静默回退的约束。
- WorkBuddy 路由 Skill 由 Pure Tokens Desktop 作为受管集成自动安装和升级，用户无需上传或手动启用。

## 0.2.7 — 2026-08-14

- 媒体完成结果现在会展示 MCP 实际返回的精确模型 ID、文件名和持久 `Downloads/Pure Tokens` 交付状态。
- 为支持渲染的宿主增加有大小上限的原生 MCP 视频资源；较大视频仍成功保存为本机 MP4，不会把超大内容塞进 stdio 响应。
- MCP 重启后的结果轮询会携带原始精确模型，不再猜测默认路由。
- 明确指定媒体模型前，必须为目标客户端选择包含该模型的分组，再点击“验证并应用”、重启客户端并新建会话。

## 0.2.6 — 2026-08-13

- 图片和视频完成后统一以 MCP 已落盘的本机 `Downloads/Pure Tokens` 文件作为持久交付，不再依赖临时预览链接。
- 图片只有在 MCP 实际返回原生图片内容时，Skill 才能说可在对话内预览；不再伪造“图片已在上方显示”。

## 0.2.4 — 2026-08-12

- 新增 `puretokens_get_balance`，通过 Desktop 发布的只读余额快照查询余额，并提供中英文使用说明。余额查询不会读取媒体模型或任何凭据材料。
- 将 `README.md` 作为默认英文入口，在顶部增加 English/中文切换，并新增中文 README。
- 使用 Image-2 和官方 Pure Tokens 图标生成 Pure Tokens Skill 品牌头图。
- 在中英文文档中增加模型能力表、使用示例和可复制的 Agent 安装提示词。
- 面向首次使用的用户重写 README，分别列出图片/视频模型，并补充直白的使用示例。
- 新增由 Skill 管理的自然语言别名注册表，包括 `image2` → `gpt-image-2`，并在请求前通过实时目录确认，不静默回退。
- 增加固定默认模型：未指定模型的图片请求使用 `gpt-image-2`，未指定模型的视频请求使用 `grok-imagine-video-1.5`。
- 增加 Nano Banana 系列别名：`Nano Banana Pro` 对应 `gemini-3-pro-image-preview`；`Nano Banana 2` 对应 `gemini-3.1-flash-image-preview`；只说 Nano Banana 且两个模型都可用时会先询问用户。
- 新增精确模型价格查询工具 `puretokens_get_model_price`。
- 展示每个选中分组对应的真实价格，不推测、不替换模型。

## 0.2.0 — 2026-08-09

- 让媒体 Skill 强制先读目录、精确匹配能力、使用稳定请求 ID、单次提交、同任务轮询，并在错误或超时后禁止自动回退。
- 增加 Claude Desktop ZIP 分发说明和便携打包命令。
- 增加安全升级和显式确认卸载命令。
- 增加模型歧义、目录为空、MCP 不可用、任务失败和轮询超时等行为场景。

## 0.1.0 — 2026-08-09

- 创建 Pure Tokens Skill 管理仓库。
- 新增 `puretokens_media`，用于选择并调用 Pure Tokens 图片和视频 MCP 工具。
- 新增版本化 Skill 注册表、机器可读清单、媒体模型目录 schema、安全本地安装器和仓库校验。
