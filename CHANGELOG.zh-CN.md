<p align="center">
  <a href="./CHANGELOG.md">English</a> · <strong>中文</strong>
</p>

# 更新日志

## Unreleased

- 补充 CC Switch/Codex 的 Connection Images API 缺失能力提示：用户明确选择非 `gpt-image-2` 图片模型，而当前连接不能直连执行时，Skill 必须在转用原生执行器、MCP 或 Direct Cloud 前告知用户该连接不能直接执行这个精确模型，并说明已验证的同模型备用通道。没有同模型备用通道就停止，绝不静默换模型。
- 新增受保护的 Codex/CC Switch Connection Images API 分支，用于用户明确选择的非 `gpt-image-2` 图片模型：只有宿主明确将当前 Pure Tokens 连接提供为可调用、已认证的 HTTPS Images API 执行器并能交付原生图片时，Skill 才读取该连接的 `/v1/media/models`、确认精确 `image` 模型，并单次提交到 `/v1/images/generations`，不使用 MCP、Direct Cloud 或第二份凭据。仅保存 API Key 或仅能读取目录不构成图片执行能力。
- 新增物理图片尺寸的 fail-closed 规则：`200cm × 230cm` 等带物理单位的值既不是图片数量，也不是可传入的 `size` 参数；Skill 不提交、不猜测 DPI、不自动换算，并明确列出支持的 `1024x1024`、`1536x1024`、`1024x1536` 画布及可选 `1K`/`2K`/`4K` 输出分辨率。
- 将 `puretokens_media` 限定为仅支持 Pure Tokens：当前连接、MCP 服务、原生媒体执行器或 Direct Cloud 凭据属于其他服务商、或无法确认归属时，Skill 会 fail-closed，不提交媒体请求，并明确告知用户本 Skill 仅支持 Pure Tokens，同时提供官网 `https://puretokensx.com/`。
- 新增受保护的 Codex/CC Switch **Pure Tokens Connection Videos API** 路径：只有宿主明确把当前 Pure Tokens 连接提供为可调用、已认证的 HTTPS 视频执行器并可交付实际字节时，Skill 才读取该连接的 `/v1/media/models`，将精确 `video` 模型单次提交至 `/v1/videos`，再轮询并取回同一任务的 `/content`。它复用已配置连接，不要求 Desktop、MCP 或第二份 Direct Cloud 凭据；仅保存 API Key 的聊天连接绝不会被误认为具备视频执行能力。
- 纠正 Codex/CC Switch 的 Image-2 路由为面向用户的 Pure Tokens 连接：默认生图和明确 `gpt-image-2` / `image2` 调用 `POST https://api.puretokensx.com/v1/images/generations` 并传入 `model: "gpt-image-2"`。Skill 绝不调用上游地址，不依赖全局宿主指令或 MCP，且该 API 路径失败后绝不回退到 MCP 或 Direct Cloud。
- 修复 `gpt-image-2` 的 MCP 分支：生成调用一旦已返回原生图片就结束，WorkBuddy 不会在已完成的 Image-2 响应后错误继续调用 `puretokens_image_result`。
- 重建中英文 README 的媒体模型清单：它现在从全局基础模型目录生成，不再读取本机路由缓存或某个 API Key 范围内的响应。`npm run docs:sync-media-models-from-base-catalog` 只接受明确的图片/视频能力，并将每个已配置模型 ID 写入发布清单；执行时仍以 `GET /v1/media/models` 的认证结果为准。

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
