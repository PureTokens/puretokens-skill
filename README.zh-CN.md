<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Skill — 一个 Skill，连接所有模型" width="100%" />
</p>

<p align="center">
  <a href="./README.md">English</a> · <strong>中文</strong> · <a href="./CHANGELOG.zh-CN.md">更新日志</a>
</p>

# Pure Tokens Skills

`puretokens-skill` 是 Pure Tokens Skill 的源仓库。它管理 Skill 指令、版本、兼容性声明、各客户端安装说明和校验工具；不保存用户凭据、Router 配置或模型路由逻辑。

## 服务提供方范围

本 Skill **仅支持 Pure Tokens**。即使其他服务商 API 兼容 OpenAI，或提供相似名称的模型，也不得通过其 API、MCP 服务或原生媒体工具提交媒体请求。若当前连接无法确认属于 Pure Tokens，Skill 会停止并提示用户切换到 Pure Tokens：https://puretokensx.com/

## 图片尺寸

`n` 仅用于“生成 3 张图”这类明确的图片数量。`200cm × 230cm` 是物理尺寸，绝不会被当作图片数量，也不能原样传给 `size`。

当前支持的图片像素画布为 `1024x1024`、`1536x1024`、`1024x1536`。用户可明确指定 `image_size` 为 `1K`、`2K` 或 `4K`；这是输出分辨率选项，不代表可保证的物理印刷尺寸。用户要求厘米、毫米、米或英寸时，Skill 不提交请求、不猜测 DPI、不自动换算、不擅自选择最接近画布，而是明确告知无法保证该物理尺寸，并列出当前支持的像素画布供用户选择。

## 快速开始

只选择当前宿主实际能够执行的一条路径：

| 宿主 | 执行路径 | 配置方式 |
| --- | --- | --- |
| 已配置 Pure Tokens 连接的 Codex 或 CC Switch | Skill → Pure Tokens Connection API → 服务 | 默认图片和 `gpt-image-2` 请求直接调用 `POST https://api.puretokensx.com/v1/images/generations` 与 `gpt-image-2`。只有宿主把当前连接提供为可调用、已认证的 HTTPS 媒体执行器且能交付实际字节时，视频才通过同一连接调用 `GET /v1/media/models` 与 `POST /v1/videos`。两条路径都不依赖全局指令、MCP 或 `PURETOKENS_API_KEY` 环境变量。 |
| 由 Pure Tokens Desktop 受管的客户端 | Skill → 受管 MCP → 本地 Router → 服务 | 这是可选的便利路径：选择客户端分组，点击 **验证并应用**，重启客户端后新建会话。 |
| 已选择原生 Pure Tokens 媒体模型的宿主 | Skill → 宿主原生媒体操作 → 服务 | 已配置操作必须提供精确、已验证的图片/视频模型和真实媒体交付；只配置通用聊天模型并不够。 |
| 具备可调用 MCP 工具的 GUI 宿主 | Skill → `puretokens-image` MCP → 服务 | 为客户端安装或配置可调用的 MCP 交付；GUI 用户绝不把 Token 粘贴到对话中。 |
| 具备 HTTPS 能力的 Agent | Skill → Direct Cloud → 服务 | 安装 Skill，并通过宿主的 Secret/环境机制注入 `PURETOKENS_API_KEY`。只有确实能执行 HTTPS 并在本机交付媒体字节的 CC Switch 连接宿主才可走此路径；不需要 Desktop、Router、CLI Sidecar 或 MCP。 |

然后新建会话。直接说“生成一只可爱的狗”即可使用默认图片模型；也可以说“使用 Nano Banana Pro 生成……”。

对于 Codex 或 CC Switch，Skill 自身定义默认的 Image-2 图片执行路径：默认生图，或明确要求 `gpt-image-2` / `image2` 时，通过当前 Pure Tokens 连接直接调用 `POST https://api.puretokensx.com/v1/images/generations`，并传入 `model: "gpt-image-2"`。当宿主还明确把当前连接提供为可调用、已认证的 HTTPS 视频执行器且能交付完成后的实际字节时，视频先读取 `GET https://api.puretokensx.com/v1/media/models`，再用返回的精确视频模型调用 `POST https://api.puretokensx.com/v1/videos`，轮询同一任务并取得 `/content`。它不依赖系统、开发者或 AGENTS 指令，也不依赖 `puretokens-image` MCP 或第二份 `PURETOKENS_API_KEY` 环境变量。这是 Pure Tokens 用户入口，绝不能调用上游地址。普通聊天连接仅保存 API Key、却没有可调用的视频执行和交付能力时不满足条件；此时仍需已选择的原生视频执行器、MCP 或 Direct Cloud。

> **使用指定模型前请先确认：** 除 Codex/CC Switch 的默认 `gpt-image-2` Images API 路径外，精确模型必须存在于当前执行路径认证后的 `GET /v1/media/models` 实时目录。已验证的当前连接视频执行器使用其连接范围；Desktop 受管 MCP 只使用当前客户端已选分组；Direct Cloud 只使用 API Key 权限。任何路径都不能调用公开目录里提到的全部模型。

当前 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_media` | 按当前目录精确选择图片/视频模型，提交一次任务并轮询同一任务，再交付实际原生媒体字节和本地文件。 |

<!-- media-model-catalog:start -->
## 媒体模型清单

已与基础模型目录同步：2026-08-21T02:46:19.421Z。

这份清单由 Pure Tokens 基础模型目录中的明确图片/视频能力生成。实际执行时，精确模型和所需能力仍必须出现在当前认证后的 GET /v1/media/models 响应中。

README 只从基础目录中带有明确图片/视频能力的模型生成，不通过模型名称推断。发布前运行 `npm run docs:sync-media-models-from-service`，从受控基础模型目录刷新清单；执行时仍以认证后的 `GET /v1/media/models` 为准。

### 图片模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | OpenAI | `image2`, `image 2`, `gpt image 2`, `openai image 2` | 高质量海报、产品视觉与插画 | `用 image2 做一张简洁的橙色产品发布海报。` |
| `grok-imagine-image` | xAI | `grok image`, `grok imagine` | 社交媒体配图与日常生图 | `用 grok-imagine-image 制作一张写实的咖啡馆开业宣传图。` |
| `grok-imagine-image-quality` | xAI | `grok quality image`, `grok high quality image` | 更锐利的品牌主视觉 | `用 grok quality image 制作一张精致的应用商店横幅。` |
| `nano-banana-2` | Google | `nano banana`, `nano banana 2`, `nano banana two` | 快速视觉探索与社交媒体创意 | `用 Nano Banana 2 制作一张明亮的产品社交媒体配图。` |
| `nano-banana-pro` | Google | `nano banana`, `nano banana pro`, `nano banana professional` | 精致营销视觉与高级主视觉 | `用 Nano Banana Pro 制作一张高级云计算主视觉。` |
| `qwen-image-2.0` | Qwen | `qwen image 2`, `qwen image 2.0` | 通用生图与产品创意视觉 | `用 qwen-image-2.0 制作一张简洁的电商产品场景图。` |
| `qwen-image-2.0-pro` | Qwen | `qwen image 2 pro`, `qwen image 2.0 pro` | 更高保真的营销视觉与产品主视觉 | `用 qwen-image-2.0-pro 制作一张高级产品营销主视觉。` |
| `seedream-5.0-pro` | Doubao | 仅精确 ID | 图片生成 | `用 seedream-5.0-pro 生成一张图片。` |
| `wan2.7-image` | Qwen | `wan image`, `wan 2.7 image` | 中文海报与产品创意视觉 | `用 wan 2.7 image 制作一张春节促销海报。` |
| `wan2.7-image-pro` | Qwen | `wan 2.7 image pro` | 更高保真的中文海报与品牌视觉 | `用 wan2.7-image-pro 制作一张高级中文产品发布海报。` |

### 视频模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `grok-imagine-video` | xAI | `grok video`, `grok imagine video` | 短社交视频与快速概念片 | `用 grok-imagine-video 制作一条 5 秒咖啡广告。` |
| `grok-imagine-video-1.5-preview` | xAI | `grok video`, `grok imagine video`, `grok 1.5 video`, `grok video 1.5`, `grok imagine video 1.5` | 视频生成 | `用 grok-imagine-video-1.5-preview 生成一条短视频。` |
| `minimax-h3` | MiniMax | `minimax h3`, `minimax h3 video` | 电影感产品短片与动态概念视频 | `用 minimax-h3 制作一条 10 秒产品揭幕视频。` |
| `seedance-2.0` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.0 生成一条短视频。` |
| `seedance-2.0-fast` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.0-fast 生成一条短视频。` |
| `seedance-2.0-mini` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.0-mini 生成一条短视频。` |
| `seedance-2.5` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.5 生成一条短视频。` |

<!-- media-model-catalog:end -->

默认只请求 1 个结果。只有用户明确给出数量、且当前执行契约支持该数量时才传入更大的数量；绝不会把一个请求拆成多次提交。MCP 路径只提交一次对应生成工具，再轮询同一个任务。Direct Cloud 图片提交始终传 `async: true`；执行层仍会防御性兼容服务返回的同步 `data[].b64_json`、`data[].url` 和异步图片任务，但只有实际媒体字节已完成本机交付后才报告成功。

## 你可以这样说

| 你想做什么 | 直接这样说 |
| --- | --- |
| 生成图片 | `生成一只可爱的狗。` |
| 生成视频 | `生成一条 15 秒、16:9 的产品广告。` |
| 使用 Nano Banana | `使用 Nano Banana Pro 做一张高级产品主视觉。` |
| 查看可用模型 | `列出我现在能用的图片和视频模型。` |

## 设计边界

```text
用户自然语言 → Skill →（当前 Connection API → 服务 | MCP → 本地 Router → 服务 | Direct Cloud → 服务）
```

- Skill 负责理解“用 image2”“用 Grok Video”等表达，先查询媒体目录，唯一匹配后选择工具，并在歧义时询问。
- MCP 只接受精确模型 ID，执行参数校验、单次提交、结果轮询和本机交付；它不做自然语言识别、不猜模型、不静默换模型。
- 宿主已选择的原生 Pure Tokens 媒体操作在能够报告精确已验证模型和真实媒体交付时优先；只有 Codex/CC Switch 宿主明确提供当前连接的视频执行与交付能力时才使用 Connection Videos API；否则 GUI 客户端优先使用可调用的 `puretokens-image` MCP 工具，具备 HTTPS 能力的 Agent 可以在宿主已注入 `PURETOKENS_API_KEY` 时使用 Direct Cloud。这些独立路径都不需要 Pure Tokens Desktop、Router、额外 CLI 或 MCP。
- 实时目录仍是唯一事实来源：已验证的当前连接、Desktop Router 与 Direct Cloud 都读取认证后的 `/v1/media/models`，并使用明确的 `image` / `video` 能力。

## 前置条件

使用 Desktop 受管 MCP 指定图片或视频模型前，请按以下顺序完成：

1. 在 Pure Tokens Desktop 中打开目标客户端的配置。
2. 选择包含目标模型的一个或多个分组。
3. 点击 **验证并应用**。
4. 重启目标客户端，并新建会话。

这个 Desktop 受管 MCP 路径只能使用当前已选分组中的模型。如果实时媒体目录没有目标模型，请回到客户端配置，选择包含该模型的分组后再次应用配置。Desktop 会为支持的客户端配置名为 `puretokens-image` 的 MCP 服务。通过 CC Switch 或其他提供方配置的自管 MCP 不要求 Desktop；模型可用性由它自己的认证 `puretokens_list_media_models` 响应决定。Skill 不会替代 MCP 配置，也不会携带任何凭据。

对于宿主原生手动配置的 Pure Tokens 模型，已选操作必须提供精确且已验证的 `image` 或 `video` 模型以及真实媒体交付。通用文本/聊天模型连接、不透明的模型名称或渲染组件都不满足该要求。CC Switch 配置 Pure Tokens API Key 也是同样边界：只有宿主已把该连接提供为可调用的已认证 HTTPS 视频执行器并能交付实际字节时，才可不使用第二份凭据走 Connection Videos API。Direct Cloud 不使用 Desktop 的分组选择界面。用户只需在宿主的常规 **API Base URL** 和 **API Key** 配置中填写 `https://api.puretokensx.com` 与 Pure Tokens API Key；使用环境变量的宿主将这两个字段映射为 `PURETOKENS_API_BASE_URL` 和 `PURETOKENS_API_KEY`。它以该 API Key 的权限和认证后的 `/v1/media/models` 目录为准；Skill 永远不会索取、打印、持久化或把 Key 写入提示词。

## 从 GitHub 安装和更新

`puretokens_media` 是所有支持客户端共用的唯一媒体行为源。Claude Desktop 通过可上传 ZIP 使用它；Codex 可以直接安装共享源；WorkBuddy 在需要时使用生成的适配层。Pure Tokens Desktop 可以作为可选的便利路径受管 Codex 和 WorkBuddy。共享 Skill 的最新安装说明和文件仍以本仓库为准。

### Codex

对于 Desktop 受管的 Codex，在 Pure Tokens Desktop 选择目标分组并点击 **验证并应用**；Desktop 会原子化安装生成的共享 Skill 到 `~/.codex/skills/puretokens_media`，并独立配置本机 `puretokens-image` MCP。不需要 Codex Plugin 或插件市场解锁。

对于具备本机终端能力的独立 Codex Agent，可以直接安装相同的共享源，并使用自行配置的 MCP 或 Direct Cloud 能力：

```bash
node bin/puretokens-skill.js install puretokens_media --target ~/.codex/skills
```

两种安装路径完成后都应新建 Codex 任务。Skill 不会打包、启动或替换 Desktop 受管 MCP。

### Claude Code、Gemini CLI、OpenCode

先克隆官方仓库，再按客户端安装到各自的用户级 Skill 目录：

```bash
git clone https://github.com/PureTokens/puretokens-skill.git
cd puretokens-skill

# Claude Code
node bin/puretokens-skill.js install puretokens_media --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_media --target ~/.gemini/skills

# OpenCode
node bin/puretokens-skill.js install puretokens_media --target ~/.config/opencode/skills
```

### 直接复制给具备本机终端的 Agent

只能把下面提示词复制给能够执行本机命令且能够写入本机文件的 Agent：

```text
请从公开仓库 https://github.com/PureTokens/puretokens-skill 为我当前使用的客户端安装 Pure Tokens Skill。

1. 判断目标客户端前，先确认当前环境同时具备本机终端和本机文件写入权限。
   - 如果当前是普通 ChatGPT 对话，或缺少任一能力，立即停止。不要克隆或下载仓库，也不要声称已经安装。请告诉我改用具备终端能力的本机 Agent，或由本机管理员安装。
2. 只有通过上一步检查后，才判断当前客户端是 Codex、Claude Code、Gemini CLI 或 OpenCode。
3. 将仓库克隆或下载到临时工作目录。
4. 只安装对应的 Pure Tokens 交付：
   - Codex：~/.codex/skills
   - Claude Code：~/.claude/skills
   - Gemini CLI：~/.gemini/skills
   - OpenCode：~/.config/opencode/skills
5. 不要覆盖其他 Skill。
6. 不要读取、索取、打印或保存 API Key、Cookie、密码、Router Token 或本地授权地址。
7. 返回实际安装目录和操作结果。

如果当前是 Claude Desktop，不要声称已经自动安装。请按 README 生成 ZIP，并告诉我应该在哪里上传和启用。如果当前是 WorkBuddy，只有在已知本机 WorkBuddy Skill 目录时才使用 README 中的生成适配层命令；否则请让我在 Pure Tokens Desktop 点击 **验证并应用**。
```

不要在普通 ChatGPT 对话中把这段提示词当作自助安装指令；必须明确停止，不能假称本机 Skill 已安装。

更新手动安装的 Codex、Claude Code、Gemini CLI 或 OpenCode Skill 时，先拉取仓库，再执行对应升级命令：

```bash
git pull
node bin/puretokens-skill.js upgrade puretokens_media --target ~/.codex/skills
```

Codex、Claude Code、Gemini CLI、OpenCode 使用上方对应目标目录的 `upgrade` 命令。升级只替换由 Pure Tokens 管理、且包含匹配 `skill.json` 与 `SKILL.md` 的目录，不会覆盖其他 Skill。

### Windows PowerShell

```powershell
git clone https://github.com/PureTokens/puretokens-skill.git
Set-Location puretokens-skill
node .\bin\puretokens-skill.js install puretokens_media --target $HOME\.claude\skills
```

其他客户端使用 `node .\bin\puretokens-skill.js install puretokens_media --target ...`，目标目录为 `$HOME\.claude\skills`、`$HOME\.gemini\skills` 或 `$HOME\.config\opencode\skills`。如果 PowerShell 找不到 `node`，先从 Node.js 官方网站安装 Node.js LTS，再重新打开 PowerShell。

## Claude Desktop 导入与 WorkBuddy 路由

Claude Desktop 使用图形界面上传本地 Skill 包。生成 ZIP：

```bash
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.4.7.zip
```

ZIP 内部结构为：

```text
puretokens_media/
├── SKILL.md
├── skill.json
├── source-delivery.json
└── references/
    ├── behavior-scenarios.json
    ├── direct-cloud-contract.md
    ├── model-catalog-contract.md
    └── natural-language-aliases.json
```

在 Claude Desktop 中打开 **Settings → Features → Skills**（部分版本显示为 **Customize → Skills**），选择 **Upload skill**，上传 ZIP 并启用 `Pure Tokens Media`。通过 CC Switch 连接的 Claude Desktop 也可以使用同一个 ZIP：通过 CC Switch 或其他本机工具提供方独立配置可调用的 `puretokens-image` MCP。若宿主本身提供认证 HTTPS 执行和本机媒体交付能力，也可以改走 Direct Cloud。ZIP 会刻意排除仅供 WorkBuddy 使用的适配层。

WorkBuddy 有两种安装路径。Pure Tokens Desktop 可以从共享 `puretokens_media` 源原子化生成并受管常驻的 `puretokens_workbuddy_router` 交付载荷，以及 `puretokens-image` MCP 条目和引用资料：选择兼容分组，点击 **验证并应用**，然后新建会话。若已知本机 WorkBuddy Skill 目录，也可由本仓库生成同一份交付物：

```bash
node scripts/render-workbuddy-media-skill.mjs --out ~/.workbuddy/skills/puretokens_workbuddy_router
```

自管路径仍需在 WorkBuddy、CC Switch 或其他工具提供方中配置可调用的 `puretokens-image` MCP，或者要求宿主实际具备 Direct Cloud 能力。普通生图、生视频请求会先发现延迟加载的 MCP 工具，再通过 `DeferExecuteTool` 实际调用；只发现工具或渲染出组件都不代表已经调用媒体模型。在 WorkBuddy UI 或工具上下文中选择的 `ImageGen`、`VideoGen` 或手动配置模型必须被保留。该选择若指向已验证的 Pure Tokens 图片/视频操作，由 WorkBuddy 已配置的原生执行器运行，不得再重复提交 MCP 任务。仅配置通用聊天模型，或只在消息文字中写出模型名，都不能绕过目录优先的选择规则。

更新 Claude Desktop 时从 GitHub 获取新版本、重新生成 ZIP、停用旧 Skill、上传新 ZIP 并启用。Desktop 受管 WorkBuddy 会在下一次点击 **验证并应用** 时重新生成同一份共享媒体行为；自管 WorkBuddy 重新运行生成命令。不得手工编辑生成后的交付物。

## Codex 安装与更新

Pure Tokens Desktop 的 **验证并应用** 只会原子化替换自己的生成式 `puretokens_media` 目录，并只配置自己的 `puretokens-image` MCP 条目；不会启用 Codex Plugin、注册 Marketplace，也不会修改其他 Skill。独立 Codex 安装使用上方共享源命令，并自行提供可调用 MCP 或 Direct Cloud 执行能力。

## 模型选择规则

在 Codex 或 CC Switch 中，默认生图和明确的 `gpt-image-2` / `image2` 请求绕过媒体目录，直接按 Skill 定义调用一次 Pure Tokens Connection Images API：`POST https://api.puretokensx.com/v1/images/generations`，传入 `model: "gpt-image-2"`；不使用 MCP、Direct Cloud、上游地址或轮询。视频仅在宿主明确将当前 Pure Tokens 连接作为可调用的已认证 HTTPS 执行器、并能交付真实字节时才使用 Connection Videos API：读取该连接的 `/v1/media/models`，精确模型单次提交到 `/v1/videos`，随后轮询并取得同一任务的 `/content`。仅保存 API Key 却没有这一能力的聊天连接必须保留已选原生执行器或使用 MCP/Direct Cloud。其他宿主和其他模型必须先调用 `puretokens_list_media_models`，只依据本次响应的 `id`、`displayName`、`aliases`、`provider` 和 `capabilities` 匹配。MCP 生成工具必须传精确 `model` 和稳定 `request_id`；Direct Cloud 仅在宿主任务状态中保留该请求 ID，因为公共端点没有已声明的幂等字段。一次用户请求只提交一次，默认只请求 1 个结果，只有用户明确给出数量时才增加数量；MCP 宿主重试时复用同一 `request_id`。`gpt-image-2` 在生成调用中直接返回原生 MCP 图片，绝不能再调用 `puretokens_image_result`；任务型图片模型只轮询其返回的同一 `task_id` 和原始模型。

媒体完成后会展示实际精确模型、保存文件名和 `Downloads/Pure Tokens`。只有 MCP 或宿主返回原生 `image` 内容时，图片才可在支持的宿主内预览；Direct Cloud 始终请求异步图片生成，并会以兼容兜底方式下载返回的 `b64_json`、返回 URL 或完成后的 `/content` 字节，再报告本机交付。完成的多图任务会通过同一个 `/content` 端点和零基 `index` 依次取回每个已声明结果。视频在大小受限时会携带原生 MCP 资源，支持该资源的宿主可以预览；较大的视频仍会成功保存为本机 MP4。只有执行层实际返回本机打开文件/文件夹入口时才展示该入口。

模型歧义、目录为空、MCP 不可用、工具错误和轮询超时的行为测试见 `skills/puretokens_media/references/behavior-scenarios.json`。任何错误都不得自动换模型或重新提交，除非用户明确选择了新的具体模型。

## 安全边界

Skill 不包含、也不会索取：

- 云端凭据、Router Token、Cookie 或密码；
- 用户配置、分组路由或支付数据；
- 本地授权地址或客户端私有文件；
- 图片、视频、任务结果或提示词历史。

Trae 目前不支持该 Skill 的媒体 MCP 流程。
