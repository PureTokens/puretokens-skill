<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Skill — 一个 Skill，连接所有模型" width="100%" />
</p>

<p align="center">
  <a href="./README.md">English</a> · <strong>中文</strong> · <a href="./CHANGELOG.zh-CN.md">更新日志</a>
</p>

# Pure Tokens Skills

`puretokens-skill` 是 Pure Tokens Skill 的源仓库。它管理 Skill 指令、版本、兼容性声明、各客户端安装说明和校验工具；不保存用户凭据、Router 配置或模型路由逻辑。

## 快速开始

只选择当前宿主实际能够执行的一条路径：

| 宿主 | 执行路径 | 配置方式 |
| --- | --- | --- |
| 由 Pure Tokens Desktop 受管的客户端 | Skill → 受管 MCP → 本地 Router → 服务 | 选择客户端分组，点击 **验证并应用**，重启客户端后新建会话。 |
| 具备 MCP/Plugin/Connector 的 GUI 宿主 | Skill → MCP/Plugin/Connector → Direct Cloud | 在宿主自己的 MCP/Plugin Secret 机制中配置 `PURETOKENS_ACCESS_TOKEN`；绝不把 Token 粘贴到对话中。 |
| 具备终端能力的代码 Agent | Skill → Direct Cloud → 服务 | 安装 Skill，并通过宿主的 Secret/环境机制注入 `PURETOKENS_ACCESS_TOKEN`。不需要 Desktop、Router、CLI Sidecar 或 MCP。 |

然后新建会话。直接说“生成一只可爱的狗”即可使用默认图片模型；也可以说“使用 Nano Banana Pro 生成……”。

> **使用指定模型前请先确认：** 精确模型必须存在于当前执行路径认证后的 `GET /v1/media/models` 实时目录。Desktop 受管 MCP 只使用当前客户端已选分组；Direct Cloud 只使用 API Key 权限。两条路径都不能调用公开目录里提到的全部模型。

当前 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_media` | 按当前目录精确选择图片/视频模型，提交一次任务并轮询同一任务，再交付原生结果和本地文件。 |

## 图片模型

下面是公开目录当前列出的图片模型。你的客户端或分组可能只显示其中一部分；只有 `puretokens_list_media_models` 实时返回精确 ID 或别名时，模型才可以使用。

用户不需要记完整 ID。像 `image2`、`Nano Banana Pro` 这样的已登记说法由 Skill 负责理解，然后仍会先去实时目录确认，确认成功才发送请求。只说“生成图片”时默认使用 `gpt-image-2`；如果当前分组没有该精确模型，Skill 会停止并展示可用候选，不会偷偷换模型。

| 模型 ID | 也可以这样说 | 适合做什么 | 真实使用示例 |
| --- | --- | --- | --- |
| `gpt-image-2` | `image2`、`gpt image 2`、`openai image 2` | 高质量海报、产品视觉、插画 | `使用 image2 做一张橙色产品发布海报。` |
| `gemini-3.0-pro-image` | `gemini pro image`、`nano banana pro` | 细节丰富的概念图和营销图 | `使用 Nano Banana Pro 做一张高级云计算主视觉。` |
| `gemini-3.1-flash-lite-image` | `gemini flash lite image` | 快速缩略图和社交媒体草稿 | `使用 gemini flash lite image 做三张明亮的社交媒体缩略图。` |
| `gemini-3.1-flash-image` | `nano banana 2` | 更快速的 Gemini 生图和对话式编辑 | `使用 Nano Banana 2 做一张明亮的产品社交海报。` |
| `grok-imagine-1.0` | `grok image`、`grok imagine` | 快速创意和轻松有趣的场景 | `使用 grok-imagine-1.0 画一只在城市公园里的快乐机器人。` |
| `grok-imagine-image` | `grok image`、`grok imagine` | 社交内容和日常生图 | `使用 grok-imagine-image 做一张咖啡店开业宣传图。` |
| `grok-imagine-image-quality` | `grok quality image` | 更精细的品牌主视觉 | `使用 grok quality image 做一张精致的应用商店横幅。` |
| `wan2.7-image` | `wan image`、`wan 2.7 image` | 中文海报和产品宣传图 | `使用 wan 2.7 image 做一张春节促销海报。` |

Skill 只提交一次 `puretokens_generate_image`，然后用 `puretokens_image_result` 查询同一个任务。

只说 `Nano Banana` 时，表示 Gemini Nano Banana 模型家族。当前目录中，Nano Banana Pro 对应 `gemini-3.0-pro-image`，Nano Banana 2 对应 `gemini-3.1-flash-image`。两个模型都可用时，Skill 会让你选择；只有一个可用时才会直接使用。这样不会把已指定的模型悄悄换成另一个。

## 视频模型

下面是公开目录当前列出的视频模型。模型必须在实时目录中声明 `video` 能力，才能生成视频。

| 模型 ID | 也可以这样说 | 适合做什么 | 真实使用示例 |
| --- | --- | --- | --- |
| `grok-imagine-video` | `grok video`、`grok imagine video` | 短视频和快速创意片段 | `使用 grok-imagine-video 做一条 5 秒咖啡广告。` |
| `grok-imagine-video-1.5` | `grok 1.5 video`、`grok video 1.5` | 更完整的短广告 | `使用 grok 1.5 video 做一条 15 秒、16:9 的产品广告。` |

Skill 只提交一次 `puretokens_generate_video`，然后用 `puretokens_video_result` 查询同一个任务。只说“生成视频”时默认使用 `grok-imagine-video-1.5`；如果当前分组没有该精确模型，Skill 会停止并展示可用候选。

如果模型不存在、有多个候选，或者没有对应能力，Skill 会列出实时候选并让你选择，不会偷偷换模型。

## 你可以这样说

| 你想做什么 | 直接这样说 |
| --- | --- |
| 生成图片 | `生成一只可爱的狗。` |
| 生成视频 | `生成一条 15 秒、16:9 的产品广告。` |
| 使用 Nano Banana | `使用 Nano Banana Pro 做一张高级产品主视觉。` |
| 查看可用模型 | `列出我现在能用的图片和视频模型。` |

## 设计边界

```text
用户自然语言 → Skill →（MCP → 本地 Router → 服务 | Direct Cloud → 服务）
```

- Skill 负责理解“用 image2”“用 Grok Video”等表达，先查询媒体目录，唯一匹配后选择工具，并在歧义时询问。
- MCP 只接受精确模型 ID，执行参数校验、单次提交、结果轮询和本机交付；它不做自然语言识别、不猜模型、不静默换模型。
- Desktop 宿主使用受管 MCP 和本地 Router；具备终端能力的 Agent 可以在宿主已注入 `PURETOKENS_ACCESS_TOKEN` 时使用 Direct Cloud，不需要 Pure Tokens Desktop、Router、额外 CLI 或 MCP。
- 实时目录仍是唯一事实来源：Desktop Router 与 Direct Cloud 都读取同一份认证后的 `/v1/media/models`，并使用明确的 `image` / `video` 能力。

## 前置条件

使用受管 MCP 指定图片或视频模型前，请按以下顺序完成：

1. 在 Pure Tokens Desktop 中打开目标客户端的配置。
2. 选择包含目标模型的一个或多个分组。
3. 点击 **验证并应用**。
4. 重启目标客户端，并新建会话。

Skill 只能通过 MCP 使用当前已选分组中的模型。如果实时媒体目录没有目标模型，请回到客户端配置，选择包含该模型的分组后再次应用配置。Desktop 会为支持的客户端配置名为 `puretokens-image` 的 MCP 服务。Skill 不会替代 MCP 配置，也不会携带任何凭据。

Direct Cloud 不使用 Desktop 的分组选择界面。它要求宿主通过自己的 Secret 机制注入 `PURETOKENS_ACCESS_TOKEN`，并以该 API Key 的权限和认证后的 `/v1/media/models` 目录为准。Skill 永远不会索取、打印、持久化或把 Token 写入提示词。

## 从 GitHub 安装和更新

`puretokens_media` 是所有支持客户端共用的唯一媒体行为源。Claude Desktop 通过可上传 ZIP 使用它；WorkBuddy 在点击 **验证并应用** 后由 Desktop 生成常驻交付载荷；具备终端能力的 Agent 可以在凭据已由宿主注入时直接走 Direct Cloud。共享 Skill 的最新安装说明和文件仍以本仓库为准。

### Codex Plugin

这里的 **Codex** 指具备本机终端和本机文件写入权限的独立 Codex Agent，不包括仅因运行时标签显示为 Codex 的普通 ChatGPT 对话。

在本仓库页面点击 **Code → Download ZIP**，或先克隆仓库，然后在仓库目录执行对应命令。需要 Node.js 20 或更高版本。

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
```

如果走受管 MCP，先在 Pure Tokens Desktop 中为 Codex 配置本机 `puretokens-image` MCP；具备终端能力的 Codex Agent 也可以在宿主注入 `PURETOKENS_ACCESS_TOKEN` 时直接走 Direct Cloud：

```bash
codex features enable plugins
codex plugin marketplace add .
codex plugin add puretokens-media@puretokens
```

安装后必须新建 Codex 任务。受管 MCP 路径下，仅复制到 `~/.codex/skills/puretokens_media` 不够：Codex 虽然会发现文字说明，但不会把 MCP 依赖绑定为可调用工具。Direct Cloud 只要求 Agent 具备 HTTPS 执行能力和宿主注入的 Token。

Pure Tokens Desktop 会在 Codex 的“验证并应用”中执行同一套官方 Plugin 安装。上面的命令仅用于本机恢复或开发。

### Claude Code、Gemini CLI、OpenCode

按客户端安装到各自的用户级 Skill 目录：

```bash

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
请从公开仓库 https://github.com/yanyansay/puretokens-skill 为我当前使用的客户端安装 Pure Tokens Skill。

1. 判断目标客户端前，先确认当前环境同时具备本机终端和本机文件写入权限。
   - 如果当前是普通 ChatGPT 对话，或缺少任一能力，立即停止。不能只因模型或运行时标签显示 Codex 就把它识别为 Codex；不要克隆或下载仓库、不要写入 `~/.codex/skills`、不要声称已经安装。请告诉我改用具备终端能力的本机 Agent，或由本机管理员安装。
2. 只有通过上一步检查后，才判断当前客户端是 Codex、Claude Code、Gemini CLI 还是 OpenCode。
3. 将仓库克隆或下载到临时工作目录。
4. 只安装对应的 Pure Tokens 交付：
   - Codex：执行 `codex features enable plugins`、`codex plugin marketplace add <仓库根目录>`、`codex plugin add puretokens-media@puretokens`。不能用写入 `~/.codex/skills` 代替。
   - Claude Code：~/.claude/skills
   - Gemini CLI：~/.gemini/skills
   - OpenCode：~/.config/opencode/skills
5. 不要覆盖其他 Skill。
6. 不要读取、索取、打印或保存 API Key、Cookie、密码、Router Token 或本地授权地址。
7. 返回实际安装目录和操作结果。

如果当前是 Claude Desktop，不要声称已经自动安装。请按 README 生成 ZIP，并告诉我应该在哪里上传和启用。如果当前是 WorkBuddy，请让我在 Pure Tokens Desktop 点击 **验证并应用**；不要手动创建或替换其生成的 `puretokens_workbuddy_router` 交付载荷。
```

不要在普通 ChatGPT 对话中把这段提示词当作自助安装指令。此类对话可能运行在 Codex 运行时上，但仍无权访问用户的终端或 `~/.codex` 目录；必须明确停止，不能假称本机 Skill 已安装。

更新 Codex 时先拉取仓库，再从同一 Plugin 市场重新安装：

```bash
git pull
codex plugin add puretokens-media@puretokens
```

Claude Code、Gemini CLI、OpenCode 使用上方对应目标目录的 `upgrade` 命令。升级只替换由 Pure Tokens 管理、且包含匹配 `skill.json` 与 `SKILL.md` 的目录，不会覆盖其他 Skill。

### Windows PowerShell

```powershell
git clone https://github.com/yanyansay/puretokens-skill.git
Set-Location puretokens-skill
codex features enable plugins
codex plugin marketplace add .
codex plugin add puretokens-media@puretokens
```

其他客户端使用 `node .\bin\puretokens-skill.js install puretokens_media --target ...`，目标目录为 `$HOME\.claude\skills`、`$HOME\.gemini\skills` 或 `$HOME\.config\opencode\skills`。如果 PowerShell 找不到 `node`，先从 Node.js 官方网站安装 Node.js LTS，再重新打开 PowerShell。

## Claude Desktop 导入与 WorkBuddy 路由

不要把复制到 `~/.codex/skills` 当成 Claude Desktop 安装完成，也不要把它当成可调用的 Codex 媒体 Plugin 安装。

Claude Desktop 使用图形界面上传本地 Skill 包。生成 ZIP：

```bash
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.4.0.zip
```

ZIP 内部结构为：

```text
puretokens_media/
├── SKILL.md
├── skill.json
└── references/
    ├── behavior-scenarios.json
    ├── direct-cloud-contract.md
    ├── model-catalog-contract.md
    └── natural-language-aliases.json
```

在 Claude Desktop 中打开 **Settings → Features → Skills**（部分版本显示为 **Customize → Skills**），选择 **Upload skill**，上传 ZIP 并启用 `Pure Tokens Media`。如果当前版本没有 Skills 入口，则该版本不能导入自定义 Skill，只能使用 MCP 的工具描述；升级或使用支持 Skills 的 Claude 客户端后再导入。

WorkBuddy 不需要手动上传或启用独立 Skill。在 Pure Tokens Desktop 中选择兼容分组并点击 **验证并应用** 后，Desktop 会从共享 `puretokens_media` 源原子化生成并受管常驻的 `puretokens_workbuddy_router` 交付载荷，以及 `puretokens-image` MCP 条目和引用资料。随后重启 WorkBuddy 或新建会话。用户直接说生图、生视频时会先发现延迟加载的 MCP 工具，再通过 `DeferExecuteTool` 实际调用；只发现工具或渲染出组件都不代表已经调用媒体模型。用户明确指定 WorkBuddy 内置 `ImageGen` 或 `VideoGen` 时仍保留该选择。

更新 Claude Desktop 时从 GitHub 获取新版本、重新生成 ZIP、停用旧 Skill、上传新 ZIP 并启用。WorkBuddy 会在下一次点击 **验证并应用** 时重新生成同一份共享媒体行为。不要直接删除 MCP 配置；MCP 由 Pure Tokens Desktop 管理。

## Codex Plugin 安装与更新

Codex 必须使用官方 Plugin 生命周期：

```bash
codex features enable plugins
codex plugin marketplace add /absolute/path/to/puretokens-skill
codex plugin add puretokens-media@puretokens
```

仓库更新后执行：

```bash
codex plugin add puretokens-media@puretokens
```

只有明确不再需要 Codex 媒体 Skill 时才执行 `codex plugin remove puretokens-media@puretokens`。移除 Plugin 不会删除 Pure Tokens Desktop 受管的 MCP 配置。

## 模型选择规则

`puretokens_media` 必须先调用 `puretokens_list_media_models`，只依据本次响应的 `id`、`displayName`、`aliases`、`provider` 和 `capabilities` 匹配。生成工具必须传精确 `model` 和稳定 `request_id`。一次用户请求只提交一次；宿主重试时复用同一 `request_id`，结果工具始终使用同一 `task_id` 和原始模型。

媒体完成后会展示 MCP 返回的实际精确模型、保存文件名和 `Downloads/Pure Tokens`。只有 MCP 返回原生 `image` 内容时，图片才可在支持的宿主内预览。视频在大小受限时会携带原生 MCP 资源，支持该资源的宿主可以预览；较大的视频仍会成功保存为本机 MP4，并从同一下载文件夹打开。

模型歧义、目录为空、MCP 不可用、工具错误和轮询超时的行为测试见 `skills/puretokens_media/references/behavior-scenarios.json`。任何错误都不得自动换模型或重新提交，除非用户明确选择了新的具体模型。

## 安全边界

Skill 不包含、也不会索取：

- 云端凭据、Router Token、Cookie 或密码；
- 用户配置、分组路由或支付数据；
- 本地授权地址或客户端私有文件；
- 图片、视频、任务结果或提示词历史。

Trae 目前不支持该 Skill 的媒体 MCP 流程。
