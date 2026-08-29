<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens 官方 Skills" width="100%" />
</p>

# Pure Tokens Skills

本仓库提供五个独立 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_balance` | 仅在宿主公开只读能力时读取当前余额快照。 |
| `puretokens_connection` | 不读取连接配置，检查当前 API 端点是否声明自己为 Pure Tokens API。 |
| `puretokens_models` | 查询当前认证模型目录，并说明已声明的模型能力、参数和媒体操作。 |
| `puretokens_image` | 通过当前已配置的 Pure Tokens Images API 生图，并按 profile 支持图片编辑。 |
| `puretokens_video` | 通过当前已配置的 Pure Tokens Videos API 生视频，并按 profile 支持图生视频、参考图/视频/音频视频和视频编辑。 |

按需安装到受支持宿主已声明的全局 Skill 目录：

```bash
# Codex
node bin/puretokens-skill.js install puretokens_balance --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_connection --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_models --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.agents/skills

# Claude Code
node bin/puretokens-skill.js install puretokens_balance --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_connection --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_models --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_balance --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_connection --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_models --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.gemini/skills
```

## 让 Agent 协助安装

如果 Agent 具备终端和本机文件写入能力，可复制对应提示词给它。Agent 只能更新下列五个 Pure Tokens Skill 目录；不得读取、打印或修改 API Key、Base URL、模型配置或 MCP 配置。

### Windows 上的 Codex：当前 `develop` 测试版本

```text
请在这台 Windows 机器上为 Codex 安装或升级当前的 Pure Tokens Skills。请直接完成操作，不要只告诉我命令。

1. 在新建的临时工作目录中克隆 https://github.com/PureTokens/puretokens-skill.git 的 develop 分支，然后进入该仓库。
2. 执行 npm run check；如失败，立即停止并报告错误。
3. 只使用 $env:USERPROFILE\.agents\skills 作为安装根目录。
4. 依次处理 puretokens_balance、puretokens_connection、puretokens_models、puretokens_image、puretokens_video：若目标目录已经是同名的受管 Pure Tokens Skill，就执行对应的 upgrade；若目录不存在，就执行 install。若已有目录不是同名受管 Skill，绝不覆盖，报告冲突。
5. 不得读取、展示、复制、修改或索取任何 API Key、Base URL、认证文件、模型配置或 MCP 配置。
6. 报告已安装的 Skill 名称和路径，并提醒我新开一个 Codex 会话后再测试。
```

### Claude Code 或 Gemini CLI

```text
请为当前宿主安装或升级 Pure Tokens Skills。请直接完成操作，不要只告诉我命令。

1. 在临时工作目录中克隆或更新 https://github.com/PureTokens/puretokens-skill.git 的 develop 分支，进入仓库后执行 npm run check；如校验失败，立即停止并报告错误。
2. Claude Code 只使用 ~/.claude/skills；Gemini CLI 只使用 ~/.gemini/skills 作为安装根目录。
3. 只安装或升级 puretokens_balance、puretokens_connection、puretokens_models、puretokens_image、puretokens_video。除非现有目录是同名受管 Skill，否则绝不覆盖；出现冲突时报告即可。
4. 不得读取、展示、复制、修改或索取任何 API Key、Base URL、认证文件、模型配置或 MCP 配置。
5. 报告已安装的 Skill 名称和路径，并提醒我新开一个宿主会话后再测试。
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

`puretokens_connection` 只通过当前连接调用一次 `GET /v1`。只有端点明确返回 `status: "ok"`、`name: "Pure Tokens API"` 与 `base_url: "/v1"` 时，才确认当前 API 标识为 Pure Tokens API。它不读取或展示真实 Base URL 和宿主配置。声明缺失或请求失败时，只会说“无法确认是 Pure Tokens”，不据此断言它属于其他服务。这是端点公开声明检查，不是密码学防伪证明。

`puretokens_models` 只调用一次 `GET /v1/media/models`。它将当前认证目录转换为用户可读的信息：精确模型 ID、实际返回的能力、已声明可选参数和媒体操作。用户可以问图生视频、参考媒体、时长、画幅或分辨率等明确技术要求有哪些兼容模型；Skill 只会根据当前 profile 实际声明的字段和值筛选，不会提交媒体任务、重试目录请求、回退到 README 静态目录，或编造质量、价格、速度和可用性排序。

`puretokens_image` 每个新任务（包括默认 `gpt-image-2`）都会先读取 `GET /v1/media/models`，再通过 `POST /v1/images/generations` 传 `async: true`。数量、像素尺寸、语义尺寸、画幅、参考字段、强度等所有非核心图片字段，都必须由该精确模型认证后的 profile 明确声明。用户明确提供的公网 HTTPS 参考图 URL 只能写入 profile 声明且 transport 允许的字段；原生图片附件则只有在 `input_schema.operations.image_edit` 明确声明 Images 路径、multipart、字段、数量和 transport 时才可发送。

`puretokens_video` 先使用 `GET /v1/media/models`，验证精确 `video` 模型 ID，再使用 `POST /v1/videos`。默认模型为 `grok-imagine-video-1.5-preview`；只轮询并交付同一任务的原生字节。提示词是否必填和所有可选字段都由精确实时 profile 决定。profile 声明 `constraints.resolution_by_mode` 时，文生、图生或参考模式必须使用该模式自己的分辨率集合，不能只按更宽的 `resolution` 属性判断。用户明确提供的公网 HTTPS 媒体 URL、file ID 或 voice ID 可以使用已声明字段和 transport；原生附件则必须使用对应的 multipart operation（`image_to_video`、`reference_image_video`、`reference_video`、`reference_audio` 或 `video_edit`）。视频编辑仍需明确声明 `POST /v1/videos/edits`、`video` 附件字段和 multipart transport。

每个已支持宿主都必须满足同一份原生执行契约：已认证的相对路径 HTTP、JSON 任务响应、原生媒体字节交付、按同一任务 ID 继续查询。验收矩阵在 `references/host-native-execution-contract.json`；它不会让 Skill 获取 Base URL、API Key 或宿主配置。

当前连接必须能执行这些请求并交付原生图片或视频字节。不能时，Skill 会在付费提交前停止，并提示用户检查已有 Pure Tokens Base URL、认证和路由配置；不会切换到其他执行路径，也不识别或分支处理其他中转服务。

## 余额

只有宿主能复用当前连接中已存在的已认证账户会话时，`puretokens_balance` 才会执行一次只读 `GET /api/product/desktop/account/balance`。它只报告接口返回的字段。若该会话未被宿主公开，Skill 会引导用户到当前连接的客户端余额入口；绝不会猜余额、尝试其他路径或索取凭据。

## 图片尺寸和数量

一个生图请求绝不会拆成多次付费提交。数量和所有尺寸控制都是模型级的：`n`、`size`、`image_size`、`aspect_ratio`、`width`、`height` 只有当前认证 profile 明确声明字段和值时才会发送。未声明 `n` 时，Skill 不会自行补充。

`200cm × 230cm` 这类物理尺寸无法精确保证，也绝不会传给 `n`、`size` 或其他 API 字段。Skill 会说明限制，并列出该精确模型当前声明的像素或语义尺寸选项。

请求 `n` 张图时，交付会在任务成功后从同一任务严格按顺序读取零基索引 `0` 到 `n-1`，一次只读取一个索引。只有每个请求索引都拿到原生字节才算成功。部分结果会明确列出已交付和缺失索引，并且只允许继续读取该任务缺失的内容。Skill 不会预取或重复下载已交付内容；每一项原生结果交付给宿主后才读取下一项。

## 模型参数资料与任务回执

每个新图片和视频任务都会使用所选模型认证后的实时 `input_schema`；静态模型清单只用于解析别名。任何可选字段都必须存在且值兼容。视频提示词在 profile 明确要求时必须提供；只有 profile 明确声明的单参考例外才可省略。资料缺失或值不兼容时，Skill 会在提交前请用户移除该选项或选择已发布参数资料的模型。

媒体输入同样由实时 profile 控制。用户明确提供的公网 HTTPS URL、file ID 或 voice ID 只能写入该 profile 的精确字段和允许 transport；Skill 不会下载、探测、检查可访问性、转存或改写它。用户当前请求附带的原生媒体则只能走已声明的 `multipart_file` operation；Skill 只会将字节随这一条 Images 或 Videos API 请求发送，Pure Tokens 网关会内部短期 R2 暂存、验证上游可读 URL，且不返回 URL。多个原生附件类型需要明确的组合 operation；多个公网 URL/ID 字段则不能与已声明的互斥或模式限制冲突。Skill 不会伪造 URL 或 file ID、调用独立上传 API，或把媒体请求静默改为纯文生。

媒体 Skill 在提交、继续查询、完成和失败时统一返回回执：已返回的精确模型 ID、已返回的任务 ID、当前状态、请求操作、请求数量、尺寸/参数、完成时的已交付数量和下一步。任务元数据未返回时会明确写“未返回”，绝不猜测。

## 异步轮询

只有提交返回 `task_id` 后才开始轮询；自动轮询只在本次提交所在的用户回合，或用户明确继续查询同一 `task_id` 的用户回合内运行。同一任务最多一个状态请求在途，绝不会创建后台计时器、队列或工作器。状态响应有有效的正数 HTTP `Retry-After` 且自动轮询预算尚有剩余时优先遵循；否则生图依次等待 `3、6、12、24、30、30` 秒，最多读取同一任务状态 6 次；生视频依次等待 `5、10、20、40、60、60` 秒，最多读取 7 次。每个有界轮询窗口最多持续：生图 120 秒，生视频 300 秒。遇到限流、5xx、传输错误或超时时立即停止本轮。如仍在处理中会连同任务 ID 如实报告；用户明确继续查询时，才会为**同一任务**开启一个新的有界轮询窗口。Skill 不会把到期或读取错误当失败，也不会提交替代任务。

媒体字节不会缓存到 Skill 状态、提示词或日志中。只有任务终态成功后才读取内容，每个任务最多一个内容读取在途；宿主交付原生字节后才会进行下一次读取。如果宿主无法在不创建无界后台工作、重复读取或缓存副本的前提下交付，Skill 会报告同任务交付不可用，不会用 URL 代替或重提任务。

## 使用示例

- 连接：`我当前连接的是 Pure Tokens 吗？` Skill 只检查当前端点的 `GET /v1` 声明，不会展示配置。
- 模型：`查看我当前可用的视频模型、它们已声明的时长和画幅选项，以及哪些支持图生视频。` 此查询只读，不会提交任务。
- 生图：`使用 gpt-image-2 生成一张 2K、16:9 的雪后黎明小镇插画。`
- 其他图片模型：`用 nano banana pro 生成一张简洁的产品海报。` Skill 只会解析唯一的已安装别名，再在当前认证目录中确认精确 ID 和图片 capability。
- 参考图 URL：`使用 gpt-image-2，并以此公网参考图 URL 生图：https://example.com/reference.png`。Skill 会先确认匹配的 profile 字段和 URL transport。
- 图片编辑：`用 grok-imagine-image 编辑我附上的图片：把阴天改成晴朗的日落。` Skill 会先确认认证后的图片编辑 profile 以及宿主可交付 multipart 附件。
- 生视频：`用 grok 1.5 video 生成一段六秒钟的电影感海上日出。`
- 图生视频：`用 grok 1.5 video 把这个公网图片 URL 制作成六秒视频：https://example.com/reference.png`。只有认证后的 profile 声明匹配 URL 字段和 transport 时才会提交。
- 参考视频：`用 seedance-2.5 根据我附上的视频制作一段六秒视频。` 只有认证目录发布 `reference_video` 时才会提交。
- 参考音频：`用 minimax h3 根据我附上的音频生成一段视频。` 只有认证目录发布 `reference_audio` 时才会提交。
- 视频编辑：`编辑我附上的视频：将白天改为夜景。` 只有认证目录发布 `video_edit` 时，才会向 `/v1/videos/edits` 提交。
- 继续已有任务：`继续查询任务 <task_id>。` Skill 只读取该任务，绝不会自动提交替代任务。

## 模型发现

用户想知道当前连接实际可用哪些模型、哪些模型支持某项媒体操作，或哪些模型接受某个已声明参数时，应使用 `puretokens_models`。它读取认证后的 `GET /v1/media/models`，只展示精确模型 ID、能力、可选参数资料和 `input_schema.operations`，不补全缺失字段。兼容模型清单只是技术匹配：只根据已声明 capability、字段/值和 operation 筛选，不做主观质量或价格推荐。

README 仅用于发现能力。模型能力只来自基础模型目录明确声明的图片/视频能力，绝不通过名称推断。每个安装后的图片/视频 Skill 都携带从同一目录生成的、按能力拆分的 `references/model-selection.json`；别名只有唯一对应一个精确模型 ID 时才可使用。

<!-- media-model-catalog:start -->
## 媒体模型清单

已与基础模型目录同步：2026-08-29T05:03:13.833Z。

这份清单由 Pure Tokens 基础模型目录中的明确图片/视频能力生成。实际执行时，精确模型和所需能力仍必须出现在当前认证后的 GET /v1/media/models 响应中。

README 只从基础目录中带有明确图片/视频能力的模型生成，不通过模型名称推断。当前目录快照只用于发现能力；实际执行时以认证后的实时模型和其 `input_schema` 为准。发布前从受控基础目录刷新，并运行 `npm run release:validate`；当快照超过七天时发布校验会失败。

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
