<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens 官方 Skills" width="100%" />
</p>

# Pure Tokens Skills

用于检查 Pure Tokens 连接、余额与模型目录，以及生成图片和视频的官方 Skills。

## 让 Agent 安装

客户端下载页会读取本标题下第一个 `text` 代码块。请保持该标题和代码块不变。

### 复制给可在本机执行命令的 Agent

```text
Install or update the official Pure Tokens Skills from https://github.com/PureTokens/puretokens-skill.
```

## 包含的 Skill

| Skill | 用途 |
| --- | --- |
| `puretokens-connection` | 验证固定 Pure Tokens API 身份，不暴露配置。 |
| `puretokens-balance` | 查询API 展示额度（币种和账户／Key 范围以接口声明为准）。 |
| `puretokens-models` | 查询已认证的当前模型目录和模型声明能力。 |
| `puretokens-image` | 生图和已声明的图片编辑。 |
| `puretokens-video` | 生视频，以及已声明的图片/视频/音频参考和视频编辑。 |
| `puretokens-update` | 初始化、展示使用须知并安全同步官方 Skills。 |

## 直连 API 的工作方式

媒体 Skill 始终请求 `https://api.puretokensx.com` 下的完整固定 URL：生图使用 `/v1/images/generations`，生视频使用 `/v1/videos`。不会将媒体请求发给任意用户配置的 Base URL、MCP、本地代理、sidecar 或第二个 endpoint。

每次请求由随 Skill 安装的单文件原生执行器完成。它只在调用期间运行，直连固定 API 后立即退出；不会启动端口、后台服务、代理、sidecar 或桌面自动化。执行器只读取受支持宿主在其已知配置路径中写入的当前有效连接，先验证该连接指向 Pure Tokens，再仅在内存中使用一把匹配 Key 请求固定 API。Pure Tokens Switch、CC Switch 或手工配置采用 `references/credential-adapters.md` 列出的格式时可使用这些适配器；没有写入这些记录的项目／会话覆盖尚未验证。Skill 和用户都不传递 Key 或 Base URL。不会扫描 Home、检查 provider 标签、索取、展示、保存、记录或报告 API Key、Base URL 或宿主配置。

生成图片和视频不需要 Node、npm、Python、Go、Pure Tokens Desktop、MCP 或上传中转。安装器会校验并放置当前平台的原生执行器；它负责固定请求、multipart 附件、同任务轮询和有边界的原生媒体字节交付。只有执行器、网络、已验证凭据适配器、附件绝对路径、附件字节或 API 失败，才可在计费前停止。

Computer Use、浏览器自动化以及打开或点击 Pure Tokens Switch/Desktop 都不是备用执行路径。Skill 不会用它们寻找可见生成界面、获取凭据、提交媒体或交付结果，也不会调用其他生图或生视频 Skill 作为回退。

## 媒体路由

当前宿主使用 Pure Tokens 连接时，即使用户没有明确说“Pure Tokens”，普通图片请求也必须先选择 `puretokens-image`，再考虑通用 `imagegen`、Imagen 或其他图片 Skill；普通视频请求也必须先选择 `puretokens-video`，再考虑通用视频 Skill。一旦选择了 Pure Tokens 专项 Skill，就只能执行固定 API 路径或安全停止，绝不回退到通用媒体 Skill。

优先级由已安装 Skill 的元数据和宿主当前连接上下文共同提供。对固定请求，Skill 可以私密地在内存中解析一把当前匹配连接的凭据，但绝不展示或报告连接配置；若某个宿主忽略已安装 Skill 的选择元数据，需要在该宿主修正自己的选择策略，Skill 无法在运行时强制第三方宿主或第三方 Skill 改变优先级。

## 支持的宿主

| 宿主 | 全局 Skill 目录 | 直连执行 |
| --- | --- | --- |
| Claude Code | `~/.claude/skills` | 凭据格式通过夹具测试；客户端端到端待验收 |
| Codex | `~/.agents/skills` | 凭据格式通过夹具测试；客户端端到端待验收 |
| WorkBuddy | `~/.workbuddy/skills` | 凭据格式通过夹具测试；客户端端到端待验收 |
| Gemini CLI | `~/.gemini/skills` | 凭据格式通过夹具测试；客户端端到端待验收 |
| Grok Build | `~/.grok/skills` | 凭据格式通过夹具测试；客户端端到端待验收 |
| OpenCode | `~/.config/opencode/skills` | 凭据格式通过夹具测试；客户端端到端待验收 |
| Trae | `~/.trae/skills` | 可安装；Switch 没有受管凭据记录，故请求会安全停止 |

七个宿主以 `references/host-support.json` 为唯一契约；不会依据 provider 名判断。

## 图片、视频与异步任务

普通文生图和文生视频先读取很小的已安装模型索引，再只读取被选中模型的 profile；不会加载全部模型，也不会在每次提交前查询模型目录。`puretokens-image` 默认使用 `gpt-image-2`；`puretokens-video` 使用其已安装默认模型。只有用户明确查询当前模型、请求选中 profile 没有的参数/媒体操作，或需要诊断模型/参数/capability 拒绝时，才读取实时目录。

生图和生视频都是异步任务。每个新请求最多一次 POST；得到顶层任务 ID 后，只轮询和获取同一个任务。若 POST 可能已开始但未取得任务 ID，提交结果即为未知：不会重复提交，也不会声称未扣费。多图按 `0..n-1` 顺序交付；视频仅在任务终态成功后交付。

当前请求中的本地图片、视频或音频附件，只会随模型声明的准确 multipart Images/Videos API 操作发送。Skill 不单独上传、不转存、不把附件改写成提示词，也不会把参考媒体请求静默降级为文生。公网 HTTPS URL 与声明的 ID 只会放入模型资料允许的字段。

模型能力、可选参数、操作、生命周期状态与访问范围以认证后的 `GET https://api.puretokensx.com/v1/media/models` 为准。当前连接未返回的模型不会提交；用户应在 Pure Tokens 选择覆盖该模型的分组/Key，并新开宿主会话后重试。

<!-- media-model-catalog:start -->
## 媒体模型清单

已与基础模型目录同步：2026-09-03T04:00:37.535Z。

这份清单由 Pure Tokens 基础模型目录中的明确图片/视频能力生成。实际执行时，精确模型和所需能力仍必须出现在认证后的 GET https://api.puretokensx.com/v1/media/models 响应中。

README 只从基础目录中带有明确图片/视频能力的模型生成，不通过模型名称推断。已安装模型索引用于选择模型，只有被选中模型的 profile 承载已知参数；实时目录只在明确查询、安装 profile 缺口或提交被拒后的诊断时按需读取。发布前从受控基础目录刷新，并运行 `npm run release:validate`；当快照超过七天时发布校验会失败。

### 图片模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | OpenAI | `image2` | 图片生成 | `用 gpt-image-2 生成一张图片。` |
| `grok-imagine-image` | xAI | `grok image` | 图片生成 | `用 grok-imagine-image 生成一张图片。` |
| `grok-imagine-image-2.0` | xAI | `grok image 2.0` | 图片生成 | `用 grok-imagine-image-2.0 生成一张图片。` |
| `grok-imagine-image-quality` | xAI | 仅精确 ID | 图片生成 | `用 grok-imagine-image-quality 生成一张图片。` |
| `nano-banana-2` | Google | `nano banana 2` | 图片生成 | `用 nano-banana-2 生成一张图片。` |
| `nano-banana-2-lite` | Google | 仅精确 ID | 图片生成 | `用 nano-banana-2-lite 生成一张图片。` |
| `nano-banana-pro` | Google | `nano banana pro` | 图片生成 | `用 nano-banana-pro 生成一张图片。` |
| `seedream-5.0-pro` | ByteDance | 仅精确 ID | 图片生成 | `用 seedream-5.0-pro 生成一张图片。` |

### 视频模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `grok-imagine-video` | xAI | `grok video` | 视频生成 | `用 grok-imagine-video 生成一条视频。` |
| `grok-imagine-video-1.5` | xAI | 仅精确 ID | 视频生成 | `用 grok-imagine-video-1.5 生成一条短视频。` |
| `grok-imagine-video-1.5-preview` | xAI | `grok 1.5 video` | 视频生成 | `用 grok-imagine-video-1.5-preview 生成一条视频。` |
| `minimax-h3` | MiniMax | `minimax h3` | 视频生成 | `用 minimax-h3 生成一条视频。` |
| `seedance-2.0` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.0 生成一条视频。` |
| `seedance-2.0-fast` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.0-fast 生成一条视频。` |
| `seedance-2.0-mini` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.0-mini 生成一条视频。` |
| `seedance-2.5` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.5 生成一条视频。` |
| `wan3.0-video` | Qwen | `wan3 video`, `wan 3 video` | 视频生成 | `用 wan3.0-video 生成一条短视频。` |
| `wan3.0-video-prime` | Qwen | `wan3 video prime`, `wan 3 video prime` | 视频生成 | `用 wan3.0-video-prime 生成一条短视频。` |

<!-- media-model-catalog:end -->

## 失败提示与回执

每次媒体操作都会返回：已返回的精确模型、任务 ID、状态、请求参数和下一步。失败会给出安全的失败阶段、API 明确返回时的公开错误码、API 明确返回时的 HTTP 状态、经清理的提示和下一步操作。Skill 不会暴露原始响应、请求头/体、内部 URL、凭据或用户媒体。

## 更新

`puretokens-update` 会取得官方仓库 `main` 的最新本地检出，并从中执行当前宿主的源码同步脚本。安装器同步六个 Skill 和经 SHA-256 校验的当前平台原生执行器、保护用户自己管理的同名目录、只删除已验证的旧官方 Skill，并在发现时删除已验证的旧 Node 运行器。只有带版本号的成功回执才表示更新完成。

源码同步脚本是 macOS/Linux 的 `runtime/puretokens-skill-install.sh` 和 Windows 的 `runtime/puretokens-skill-install.ps1`。它们只负责安装更新及校验复制平台执行器；用户不需要 Node、npm、Python、Go 或包管理器。

每次安装或更新成功后，安装器都会自动执行 `init`：先做不计费的固定 `/v1` 身份检查，再用一次 `/v1/media/models` 请求验证当前凭据认证，不展示凭据或宿主配置，然后输出当前使用须知和示例。验证未完成时，会给出经过脱敏的原因，例如没有当前匹配连接、缺少凭据、API 拒绝及 HTTP 状态、网络失败或 API 身份未确认；绝不打印配置 URL、provider 或 Key。之后如需再次检查，可让宿主 Agent“初始化 Pure Tokens Skills”或“检查当前 Pure Tokens 连接”；它应调用已安装执行器的 `init`，展示使用须知，但不修改配置。

## 维护者校验

```bash
npm run check
npm run release:validate
```

## 执行回执与验收

0.16 链路为：提交 → 立即返回任务 ID → 有界等待／查询 → 按索引下载 → 宿主交付附件。普通生成不增加余额、init 或目录预检。用有限的请求 JSON 文件输入，避免交互 stdin 一直等待；已下载不等于已交付。

余额使用现有 API Key 认证的 `/v1/dashboard/billing/subscription` 与 `/v1/dashboard/billing/usage`，以总额减用量/100 得到接口展示额度。不推断币种、账户／Key 范围或无限余额。

`references/host-acceptance.json` 分开记录凭据格式测试与真实客户端验收。本地自动化测试通过，不能替代每个客户端在 Windows／macOS 上的附件交付验收。
