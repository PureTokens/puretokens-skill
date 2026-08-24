# Pure Tokens Skills

本仓库提供三个独立 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_balance` | 仅在宿主公开只读能力时读取当前余额快照。 |
| `puretokens_image` | 通过当前已配置的 Pure Tokens Images API 生成图片。 |
| `puretokens_video` | 通过当前已配置的 Pure Tokens Videos API 生成视频。 |

按需安装到受支持宿主已声明的全局 Skill 目录：

```bash
# Codex
node bin/puretokens-skill.js install puretokens_balance --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.agents/skills

# Claude Code
node bin/puretokens-skill.js install puretokens_balance --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_balance --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.gemini/skills
```

## 宿主支持

CC Switch 是连接配置工具，不是 Skill 宿主。受支持的宿主会使用 CC Switch、Pure Tokens Desktop 或用户已经配置好的当前连接。

| 宿主 | 当前专项 Skill 交付方式 | 用户操作 |
| --- | --- | --- |
| Codex | 手动安装源文件 | 将所需 Skill 安装到 `~/.agents/skills`。 |
| Claude Code | 手动安装源文件 | 将所需 Skill 安装到 `~/.claude/skills`。 |
| Claude Desktop | ZIP 包 | 打包所需专项 Skill，在 Claude Desktop 的 Skills 设置中上传并启用。 |
| Gemini CLI | 手动安装源文件 | 将所需 Skill 安装到 `~/.gemini/skills`。 |
| WorkBuddy、Grok Build、OpenCode、Trae | 当前不提供交付 | 它们的 Desktop Router/连接配置 Adapter 不代表已提供兼容的专项 Skill 交付。 |

唯一事实来源是 `references/host-support.json`。CLI 刻意不会猜测宿主目录。

## 连接契约

宿主当前已配置的连接负责 Base URL、认证和路由。CC Switch、Pure Tokens Desktop 或用户手动配置的宿主连接都可以提供该连接。Skill 不会读取、扫描、索取、打印或保存凭据和宿主配置，也不检查 provider 标签、Base URL 或服务归属。

`puretokens_image` 使用 `POST /v1/images/generations`；默认模型为 `gpt-image-2`，每次图片请求都传 `async: true`。选择其他图片模型时，先通过 `GET /v1/media/models` 验证精确 ID 和 `image` capability。

`puretokens_video` 先使用 `GET /v1/media/models`，验证精确 `video` 模型 ID，再使用 `POST /v1/videos`。默认模型为 `grok-imagine-video-1.5-preview`；只轮询并交付同一任务的原生字节。

每个已支持宿主都必须满足同一份原生执行契约：已认证的相对路径 HTTP、JSON 任务响应、原生媒体字节交付、按同一任务 ID 继续查询。验收矩阵在 `references/host-native-execution-contract.json`；它不会让 Skill 获取 Base URL、API Key 或宿主配置。

当前连接必须能执行这些请求并交付原生图片或视频字节。不能时，Skill 会在付费提交前停止，并提示用户检查已有 Pure Tokens Base URL、认证和路由配置；不会切换到其他执行路径，也不识别或分支处理其他中转服务。

## 余额

只有宿主能复用当前连接中已存在的已认证账户会话时，`puretokens_balance` 才会执行一次只读 `GET /api/product/desktop/account/balance`。它只报告接口返回的字段。若该会话未被宿主公开，Skill 会引导用户到当前连接的客户端余额入口；绝不会猜余额、尝试其他路径或索取凭据。

## 图片尺寸和数量

默认生成一张图片。明确的 `n` 必须是 1 到 6 的整数；一个请求绝不会拆成多次付费提交。支持的 `size` 为 `1024x1024`、`1536x1024`、`1024x1536`；支持的 `image_size` 为 `1K`、`2K`、`4K`。

`200cm × 230cm` 这类物理尺寸无法精确保证，也绝不会传给 `n` 或 `size`。Skill 会说明限制并请用户选择支持规格。

请求 `n` 张图时，交付会从同一任务严格读取零基索引 `0` 到 `n-1`。只有每个请求索引都拿到原生字节才算成功。部分结果会明确列出已交付和缺失索引，并且只允许继续读取该任务缺失的内容。

## 模型参数资料与任务回执

默认 `gpt-image-2` 使用上述数量和尺寸取值。其他图片模型若要求 `n`、`size`、`image_size` 等可选字段，必须由该精确模型当前认证目录中的 `input_schema` 明确声明字段和值；只给文字提示词的生图不需要参数资料。视频的时长、画幅、分辨率、尺寸及其他可选字段同样必须由所选模型的实时 `input_schema` 明确声明；只给文字提示词的生视频在没有资料时仍可请求。资料缺失或值不兼容时，Skill 会在提交前请用户移除该选项或选择已发布参数资料的模型。

媒体 Skill 在提交、继续查询、完成和失败时统一返回回执：已返回的精确模型 ID、已返回的任务 ID、当前状态、请求数量、尺寸/参数、完成时的已交付数量和下一步。任务元数据未返回时会明确写“未返回”，绝不猜测。

## 异步轮询

只有提交返回 `task_id` 后才开始轮询。如果任务状态响应带有有效的 HTTP `Retry-After` 等待时间，Skill 优先遵循；否则前四次同任务状态查询依次等待 2、3、5、8 秒，之后每次间隔 15 秒。自动轮询从提交响应起算：生图最多 120 秒，生视频最多 300 秒。到期仍未完成会如实报告任务仍在处理中；用户可以明确要求继续查询同一任务，但 Skill 不会把到期当失败，也不会提交替代任务。

## 使用示例

- 生图：`使用 gpt-image-2 生成一张 1024x1024 的雪后黎明小镇插画。`
- 其他图片模型：`用 nano banana pro 生成一张简洁的产品海报。` Skill 只会解析唯一的已安装别名，再在当前认证目录中确认精确 ID 和图片 capability。
- 生视频：`用 grok 1.5 video 生成一段六秒钟的电影感海上日出。`
- 继续已有任务：`继续查询任务 <task_id>。` Skill 只读取该任务，绝不会自动提交替代任务。

## 模型发现

README 仅用于发现能力。实际执行时，当前认证后的 `GET /v1/media/models` 仍是非默认图片和所有视频的唯一事实来源。模型能力只来自基础模型目录明确声明的图片/视频能力，绝不通过名称推断。每个安装后的图片/视频 Skill 都携带从同一目录生成的、按能力拆分的 `references/model-selection.json`；别名只有唯一对应一个精确模型 ID 时才可使用。

<!-- media-model-catalog:start -->
## 媒体模型清单

已与基础模型目录同步：2026-08-21T02:46:19.421Z。

这份清单由 Pure Tokens 基础模型目录中的明确图片/视频能力生成。实际执行时，精确模型和所需能力仍必须出现在当前认证后的 GET /v1/media/models 响应中。

README 只从基础目录中带有明确图片/视频能力的模型生成，不通过模型名称推断。当前目录快照只用于发现能力；实际执行时以认证后的实时模型和其 `input_schema` 为准。发布前从受控基础目录刷新，并运行 `npm run release:validate`；当快照超过七天时发布校验会失败。

### 图片模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | OpenAI | `image2` | 图片生成 | `用 image2 生成一张图片。` |
| `grok-imagine-image` | xAI | `grok image` | 图片生成 | `用 grok-imagine-image 生成一张图片。` |
| `grok-imagine-image-quality` | xAI | 仅精确 ID | 图片生成 | `用 grok-imagine-image-quality 生成一张图片。` |
| `nano-banana-2` | Google | `nano banana 2` | 图片生成 | `用 nano-banana-2 生成一张图片。` |
| `nano-banana-pro` | Google | `nano banana pro` | 图片生成 | `用 nano-banana-pro 生成一张图片。` |
| `qwen-image-2.0` | Qwen | 仅精确 ID | 图片生成 | `用 qwen-image-2.0 生成一张图片。` |
| `qwen-image-2.0-pro` | Qwen | 仅精确 ID | 图片生成 | `用 qwen-image-2.0-pro 生成一张图片。` |
| `seedream-5.0-pro` | Doubao | 仅精确 ID | 图片生成 | `用 seedream-5.0-pro 生成一张图片。` |
| `wan2.7-image` | Qwen | 仅精确 ID | 图片生成 | `用 wan2.7-image 生成一张图片。` |
| `wan2.7-image-pro` | Qwen | 仅精确 ID | 图片生成 | `用 wan2.7-image-pro 生成一张图片。` |

### 视频模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `grok-imagine-video` | xAI | `grok video` | 视频生成 | `用 grok-imagine-video 生成一条视频。` |
| `grok-imagine-video-1.5-preview` | xAI | `grok 1.5 video` | 视频生成 | `用 grok-imagine-video-1.5-preview 生成一条视频。` |
| `minimax-h3` | MiniMax | `minimax h3` | 视频生成 | `用 minimax-h3 生成一条视频。` |
| `seedance-2.0` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.0 生成一条视频。` |
| `seedance-2.0-fast` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.0-fast 生成一条视频。` |
| `seedance-2.0-mini` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.0-mini 生成一条视频。` |
| `seedance-2.5` | Doubao | 仅精确 ID | 视频生成 | `用 seedance-2.5 生成一条视频。` |

<!-- media-model-catalog:end -->

## 更新

拉取最新仓库后，分别升级已经安装的 Skill：

```bash
node bin/puretokens-skill.js upgrade puretokens_image --target ~/.agents/skills
```

Claude Desktop 需要打包并上传对应的专项 Skill：

```bash
node bin/puretokens-skill.js bundle puretokens_image --format claude-desktop --out ./puretokens_image.zip
```

发布前运行：

```bash
npm run docs:sync-media-models-from-service
npm run release:validate
```
