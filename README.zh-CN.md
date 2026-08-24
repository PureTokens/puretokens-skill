# Pure Tokens Skills

本仓库提供三个独立 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_balance` | 仅在宿主公开只读能力时读取当前余额快照。 |
| `puretokens_image` | 通过当前已配置的 Pure Tokens Images API 生成图片。 |
| `puretokens_video` | 通过当前已配置的 Pure Tokens Videos API 生成视频。 |

按需安装到宿主的 Skill 目录：

```bash
node bin/puretokens-skill.js install puretokens_balance --target ~/.codex/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.codex/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.codex/skills
```

## 连接契约

宿主当前已配置的 Pure Tokens 连接负责 Base URL、认证和路由。CC Switch、Pure Tokens Desktop 或用户手动配置的宿主连接都可以提供该连接。Skill 不会读取、扫描、索取、打印或保存凭据和宿主配置。

`puretokens_image` 使用 `POST /v1/images/generations`；默认模型为 `gpt-image-2`，每次图片请求都传 `async: true`。选择其他图片模型时，先通过 `GET /v1/media/models` 验证精确 ID 和 `image` capability。

`puretokens_video` 先使用 `GET /v1/media/models`，验证精确 `video` 模型 ID，再使用 `POST /v1/videos`。默认模型为 `grok-imagine-video-1.5-preview`；只轮询并交付同一任务的原生字节。

当前连接必须能执行这些请求并交付原生图片或视频字节。不能时，Skill 会在付费提交前停止，并提示用户检查已有 Pure Tokens Base URL、认证和路由配置；不会切换到其他提供方或执行路径。当前连接不是 Pure Tokens 时，请使用 https://puretokensx.com/。

## 图片尺寸和数量

默认生成一张图片。明确的 `n` 必须是 1 到 6 的整数；一个请求绝不会拆成多次付费提交。支持的 `size` 为 `1024x1024`、`1536x1024`、`1024x1536`；支持的 `image_size` 为 `1K`、`2K`、`4K`。

`200cm × 230cm` 这类物理尺寸无法精确保证，也绝不会传给 `n` 或 `size`。Skill 会说明限制并请用户选择支持规格。

## 模型发现

README 仅用于发现能力。实际执行时，当前认证后的 `GET /v1/media/models` 仍是非默认图片和所有视频的唯一事实来源。模型能力只来自基础模型目录明确声明的图片/视频能力，绝不通过名称推断。

<!-- media-model-catalog:start -->
## 媒体模型清单

已与基础模型目录同步：2026-08-21T02:46:19.421Z。

这份清单由 Pure Tokens 基础模型目录中的明确图片/视频能力生成。实际执行时，精确模型和所需能力仍必须出现在当前认证后的 GET /v1/media/models 响应中。

README 只从基础目录中带有明确图片/视频能力的模型生成，不通过模型名称推断。发布前运行 `npm run docs:sync-media-models-from-service`，从受控基础模型目录刷新清单；执行时仍以认证后的 `GET /v1/media/models` 为准。

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
node bin/puretokens-skill.js upgrade puretokens_image --target ~/.codex/skills
```

Claude Desktop 需要打包并上传对应的专项 Skill：

```bash
node bin/puretokens-skill.js bundle puretokens_image --format claude-desktop --out ./puretokens_image.zip
```
