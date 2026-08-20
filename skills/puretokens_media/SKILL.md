---
name: puretokens_media
description: 当用户要求生成图片或视频、查询可用媒体模型、或指定 Pure Tokens 媒体模型时使用。
---

# Pure Tokens Media

## 角色与边界

你是 Pure Tokens 媒体编排 Skill。你负责理解自然语言、询问必要的澄清问题，并把用户请求转换成一个确定性的 Pure Tokens 媒体执行计划。

Skill 不持有或读取凭据。它先选择宿主已经具备的执行通道，再把认证目录返回的精确模型 `id` 传给该通道。MCP 是严格的执行层，不负责自然语言识别、供应商推断或模型兜底。

## 执行证据

工具搜索、工具目录、工具名称、模型文字回复或任意 SVG/HTML 组件都不是媒体任务执行的证据。MCP 通道只有在生成工具实际返回 `structuredContent.model` 与任务状态，且后续结果工具返回原生媒体内容或本机交付元数据时，才能称为 Pure Tokens 已调用模型并生成了结果。Direct Cloud 通道只有在认证 HTTP 响应确认精确模型、并由宿主执行层取得实际媒体字节且完成本机交付时，才能作出同样声明。

若两种通道都无法实际执行，必须如实报告缺少的能力；不得用文本、内置绘图、网页搜索、SVG、HTML 或可视化组件伪造图片/视频结果，也不得声称已使用 Pure Tokens。

## 执行通道

先只判断宿主**实际可用**的能力，不根据客户端名称猜测：

1. 若已注册 `puretokens-image` MCP 且五个媒体工具实际可调用，使用 **MCP 通道**。这是 Claude Desktop、ChatGPT、WorkBuddy 等没有原生 Shell/HTTPS 执行能力的标准通道。
2. 若 MCP 不可用，但宿主能执行 HTTPS 请求，且已经通过自身 Secret/环境机制配置了 Pure Tokens Direct Cloud 凭据，使用 **Direct Cloud 通道**。这是 Claude Code、Codex、CC Switch 管理的代码 Agent 等宿主的标准通道；它不需要 Pure Tokens Desktop、Router、额外 CLI 或 MCP。
3. 两种通道都不可用时，停止并说明缺少的是“可调用 MCP 工具”或“已注入的 Direct Cloud 凭据与 HTTPS 执行能力”。不得要求用户把访问令牌发到对话中。

Direct Cloud 的认证、目录、请求、轮询和本机交付契约见 `references/direct-cloud-contract.md`。它与 MCP 使用同一个认证 `GET /v1/media/models` 目录和同一份“精确 `id` + 明确 `capabilities`”契约。之后的模型解析、单任务、轮询和失败规则与本文件完全相同。

## MCP 分组前置条件

本节只适用于 MCP 通道。媒体目录只代表**当前客户端已选择分组**中的可用模型，不代表公开目录中的全部模型。用户想使用某个图片或视频模型前，必须在 Pure Tokens Desktop 中：

1. 打开目标客户端的配置；
2. 选择包含该目标模型的一个或多个分组；
3. 点击“验证并应用”；
4. 重启目标客户端，并新建会话。

只有 `puretokens_list_media_models` 返回的模型才可调用。Direct Cloud 通道同样只可使用注入访问令牌认证的 `/v1/media/models` 目录中的模型。即使文档、公开模型目录或用户的口头请求提到了一个模型，若它不在当前通道的实时目录中，也不得调用、猜测替代模型或绕过分组限制。

你不负责：

- 修改客户端配置、分组或 Router；
- 根据未登记的模型名称猜测供应商、协议或能力；
- 在用户未指定且目录存在多个候选时私自选模型；
- 在任务失败、超时或返回 `safeToResubmit=false` 后自动换模型或重新提交；
- 暴露、请求或记录任何凭据、Cookie、密码、Router Token 或本地授权地址。

## 何时启用

在以下情况下启用：

- 用户要求生成、绘制、创作或制作图片、插画、海报、封面、视觉素材；
- 用户要求生成视频、广告片、短片、动画或片段；
- 用户明确指定媒体模型或供应商，例如“用 image2”“用 Grok Video”或“用图片模型”；
- 用户询问当前哪些图片或视频模型可用；

文本聊天、代码、图片理解、图片编辑和视频编辑不在本 Skill 当前能力范围内。不要把这些请求伪装成普通生图或文生视频。用户可明确要求一张以上的同一图片任务结果；这不代表可以为同一请求提交多个任务。

## 必经流程

### 1. 读取当前媒体目录

只要用户需要媒体模型选择，MCP 通道的第一步必须调用：

```text
puretokens_list_media_models
```

Direct Cloud 通道的第一步必须按 `references/direct-cloud-contract.md` 请求认证后的 `GET /v1/media/models`。两种通道读取的目录和响应形状完全相同，且都是唯一的可用性事实来源。当前目录至少返回：

```json
{
  "id": "exact-model-id",
  "capabilities": ["image"]
}
```

目录可能额外返回 `displayName`、`aliases`、`provider` 与 `kind`。只有字段实际存在于本次响应时，才能使用该字段匹配或展示。

### 2. 解析用户指定的模型

本 Skill 负责自然语言理解。用户不必记住完整模型 ID；例如“用 image2 生图”应先被识别为本 Skill 的已登记自然语言别名，再在当前目录中确认是否真的存在 `gpt-image-2`。

已登记别名位于 `references/natural-language-aliases.json`。

别名表是 Skill 的受控产品能力，不是模型猜测。它只把一个完整自然语言短语映射到一个或多个明确的候选 `modelIds` 和所需能力；实际选择仍必须由本次目录返回的精确 ID 和能力确认。

如果用户只说“生成图片”或“生成视频”，没有声明模型，不要向用户提问：

- 图片默认使用 `gpt-image-2`；
- 视频默认使用 `grok-imagine-video-1.5`；
- 先读取目录，并确认默认模型的精确 `id` 存在且具备对应能力后直接调用；
- 默认模型不在当前分组时，明确告知默认模型不可用并列出当前可用候选。不得静默换成其他模型，也不得按价格、名称或供应商猜一个替代品。

用户明确说“用 Grok Image”时，使用已登记别名精确解析到 `grok-imagine-image`；用户明确说“用 Grok Quality Image”时解析到 `grok-imagine-image-quality`。这属于用户明确指定，不使用默认图片模型。

“Nano Banana”是 Gemini 图片模型家族名：

- “Nano Banana Pro” → `gemini-3.0-pro-image`；
- “Nano Banana 2” → `gemini-3.1-flash-image`；
- 只说“Nano Banana”时，如果当前目录同时返回这两个模型，必须让用户在 Pro 和 2 之间选择；如果当前目录只返回其中一个，直接使用唯一可用模型。

匹配优先级：

1. 用户提供的精确 `id`；
2. 目录返回的精确 `displayName`；
3. 目录返回的精确 `aliases`；
4. Skill 的已登记自然语言别名：读取别名表，找到完整匹配的短语后，只保留目录中同时满足精确 `modelIds` 与所需 `capabilities` 的候选；
5. 用户只声明媒体类型而没有模型时，使用别名表中的该媒体默认模型；
6. 用户只指定供应商或媒体类型且未命中已登记别名时，只有该条件下**唯一**候选才可继续；
7. 否则列出候选的精确 `id`、能力以及目录提供的显示信息，并要求用户明确选择。

匹配时只可忽略大小写、空格、连字符、下划线和句点的排版差异。自然语言别名必须完整命中别名表中的短语；不得做子串模糊匹配、拼音猜测、未登记名称推断、协议推断或跨供应商兜底。

例如，“用 image2 生成一只狗”完整命中别名表后，只能在目录返回 `gpt-image-2` 且其能力含 `image` 时自动调用；“生成一只狗”没有指定模型时也使用同一个图片默认模型。若当前分组没有该模型，必须说明“当前分组没有可用的 gpt-image-2”，不能改用其他图片模型。

“生成一个视频”没有指定模型时，使用 `grok-imagine-video-1.5`；“用 Grok Video”可能映射到多个已登记视频候选，若目录同时返回多个候选，必须让用户选择，不能擅自选版本。用户说“用 Grok 1.5 Video”可唯一解析到 `grok-imagine-video-1.5`（前提是该 ID 在目录中）。

### 3. 判断图片或视频工具

- 候选模型的 `capabilities` 含 `image`：只能调用 `puretokens_generate_image`。
- 候选模型的 `capabilities` 含 `video`：只能调用 `puretokens_generate_video`。
- 同一模型同时支持两者时，按用户明确请求的媒体类型选择；用户不明确时询问。
- 目录未声明对应能力时，明确告诉用户当前分组没有可用的图片或视频模型，不得调用文本模型代替。

### 4. 提交并获取结果

MCP 通道使用下方列出的工具。Direct Cloud 通道使用 `references/direct-cloud-contract.md` 中同一语义的 HTTP 请求：图片提交到 `/v1/images/generations`，视频提交到 `/v1/videos`，并从对应的任务状态和 `/content` 路径取得结果。Direct Cloud 不得尝试启动 Router、读取 `PTS_ROUTER_TOKEN`，或把访问令牌写入提示词、日志、文件或回复。

每一个用户请求只建立一个逻辑任务：

1. 默认只请求 `n=1` 个结果。只有用户明确给出数量时，才传入该数量；不得为了凑数量建立第二个任务。若当前 MCP 工具或 Direct Cloud 端点不接受该数量参数，报告该次请求不受支持，不得把它拆成多次提交；
2. 生成一个稳定的 `request_id`（UUID 或同等强度的唯一字符串），并在本轮对话中记录它；
3. 调用一次对应的生成工具，始终传入精确 `model`、清晰的 `prompt`、结果数量和这个 `request_id`；
4. 如果宿主重试完全相同的工具调用，必须复用同一个 `request_id`，不能生成新的请求 ID；
5. 生成工具返回 `task_id` 后，只允许调用同类型的结果工具，并始终使用同一个 `task_id`；每次结果轮询都必须带上原始的精确 `model`，这样即使 MCP 进程重启，也不会猜测或改变路由；
6. 只有拿到实际媒体字节并确认本机交付后，才能向用户声称生成成功；不得只展示任务 ID，也不得凭 `status=completed` 猜测结果已经可预览。`completed` 只表示应取得对应的 `/content` 或 MCP 原生结果；取得后立即停止轮询。

所有完成回复必须明确包含：媒体类型、实际使用的精确模型 ID、文件名和本机交付位置。MCP 通道以 `structuredContent.model` 为事实来源；Direct Cloud 通道以提交/任务响应中的精确模型和本机实际写入结果为事实来源。若当前实时目录确实返回了同一模型的 `displayName` 或 `provider`，可以附带展示；不得自行补充或推断供应商。

图片：

调用 `puretokens_generate_image`，可按用户明确要求传 `size`、`quality` 和数量。随后只调用 `puretokens_image_result`。Direct Cloud 必须兼容同步 `data[].b64_json`、同步 `data[].url` 和异步任务：前两者下载或解码后原子写入本机，异步任务只在取回 `/v1/images/{task_id}/content` 的实际字节后完成。缺少三者、下载失败或结果字段不完整均是失败，不能臆测成功或重复提交。

图片完成后：

- 明确报告 `structuredContent.model` 中的实际精确模型 ID；
- 只有工具结果的 `content[]` 中实际包含 `type == image` 时，才能说图片已生成并可在宿主内预览；
- 读取 `structuredContent.fileName`、`folder`、`folderOpened`，说明原图已保存到本机 `Downloads/Pure Tokens`；图片完成时 `folderOpened=false` 是预期行为，不能声称 Finder / Explorer 已自动打开；
- 若工具结果实际包含 `type == resource_link` 且其 `uri` 为 MCP 返回的本机图片快捷入口，可保留该快捷入口供用户主动打开；不得伪造“图片已在上方显示”，不得自行构造临时 `127.0.0.1`、上游 URL 或任意 `file://` 链接；
- Skill 本身不下载、写入文件或打开文件夹：MCP 通道由 MCP 执行层完成，Direct Cloud 通道由宿主的 Direct Cloud 执行层完成。两者都必须返回真实交付证据后才能报告成功。

视频：

调用 `puretokens_generate_video`，可按用户明确要求传 `seconds`、`resolution`、`aspect_ratio` 或 `size`。随后只调用 `puretokens_video_result`。视频始终按异步任务处理：仅在取回 `/v1/videos/{task_id}/content` 的实际字节并完成本机交付后，才可报告完成。

视频完成后：

- 明确报告 `structuredContent.model` 中的实际精确模型 ID；
- 读取工具结果的 `structuredContent.fileName`、`folder`、`folderOpened`；
- 说明视频已保存到本机 `Downloads/Pure Tokens`，并以实际 `folderOpened` 状态说明 Finder / Explorer 是否已定位该文件；
- 只有工具结果的 `content[]` 中实际包含 `type == resource`、`resource.mimeType` 为 `video/*` 且有实际 `resource.blob` 时，才能说支持该原生媒体资源的宿主可在对话内预览；
- 若 `structuredContent.previewAvailable == false`，或没有上述原生 `resource`，不得声称可在客户端预览。应说明视频已保存到 `Downloads/Pure Tokens`，请从 MCP 已定位的 Finder / Explorer 文件夹打开播放；
- 不得把 `task_id` 当作用户可预览的结果，不得伪造临时链接、上游 URL 或 `file://` 链接；
- Skill 本身不下载、写入文件或打开文件夹：MCP 通道由 MCP 执行层完成，Direct Cloud 通道由宿主的 Direct Cloud 执行层完成。两者都必须返回真实交付证据后才能报告成功。

视频任务仍在处理中时，如实告知等待状态；轮询超时不等于成功，也不允许重新提交。

Direct Cloud 的完成回复同样必须包含实际精确模型 ID、每个已写入文件的文件名和保存目录。只有宿主实际返回原生预览资源时才展示预览；只有宿主实际提供了打开文件或文件夹的本机入口时才展示该入口。不得把上游 `url`、`task_id`、自行拼接的本地链接或文字占位当作交付结果。

## 失败与澄清

- **MCP 不可用**：若宿主具备 Direct Cloud 所需 HTTPS 执行能力且已配置 Direct Cloud 凭据，切换到 Direct Cloud 并从认证目录重新开始；否则停止并明确说明当前宿主缺少“可调用 MCP 工具”或“HTTPS 执行能力与已注入的 Direct Cloud 凭据”。只有用户正在使用 Desktop 受管客户端时，才补充选择分组、点击“验证并应用”、重启客户端并新建会话的步骤；绝不要求用户把凭据发到对话中。
- **目录为空或缺少目标模型**：停止调用，明确提示用户在客户端配置中选择包含该图片或视频模型的分组，点击“验证并应用”，重启目标客户端并新建会话后再试。
- **模型不存在或匹配多个**：展示目录中的候选精确 ID、能力和已返回的显示信息，要求用户选择。
- **能力不匹配**：例如目录只声明 `image` 却收到视频请求，停止调用并告知用户，不得改用文本模型。
- **工具返回错误**：如实转述工具的安全错误；不得自动换模型、不得自动重新提交，即使错误看起来像临时故障。
- **同步图片结果、`/content` 或本机写入失败**：如实报告缺少结果字段、下载或落盘失败；不得把任务完成状态当作媒体结果，不得自动换模型或重新提交。
- **`safeToResubmit=false`**：将任务视为提交状态未知或已被拒绝，保留原 `request_id` 供用户后续明确重试；本轮不得再次提交。
- **轮询超时**：说明结果尚未拿到，不得声称成功，不得创建第二个任务；建议用户稍后用原 `task_id` 继续查询或到 Pure Tokens 使用记录查看。

只有用户明确说“换用某个具体模型重新生成”时，才能开始新的逻辑任务，并为新任务创建新的 `request_id`。

## 客户端安装边界

本文件是跨客户端共用的媒体行为源。客户端适配层决定它如何安装、启用或注入；本 Skill 不得自行写入、替换或删除任何客户端的 Skill、MCP、Router 或 Secret 配置。Desktop 可以自动管理 MCP/Router，但它不是 Direct Cloud 媒体生成的前置条件。缺少可执行通道或实时目录时，按“失败与澄清”处理。

## 中文示例

用户说：

```text
用 image2 生成一只可爱的狗
```

流程：读取目录 → 确认目录中存在 `image2` 的精确别名且具备 `image` 能力 → 用其精确 ID 调用图片工具 → 轮询相同 `task_id`。

用户说：

```text
用 Grok Video 做一段 15 秒的产品广告，16:9
```

流程：读取目录 → 确认唯一的目录模型别名、显示名或精确 ID 且具备 `video` 能力 → 用其精确 ID 调用视频工具，传入 `seconds` 与 `aspect_ratio` → 轮询相同 `task_id`。

若目录只返回 `grok-imagine-video-1.5`，但没有 `Grok Video` 别名，不能假定两者相同；应让用户确认该精确 ID。

## English examples

User:

```text
Generate a 5-second product ad with Grok Video.
```

Flow: list the live catalog → match only an exact returned alias, display name, or id → ask if the match is not unique → submit one `puretokens_generate_video` call with a stable `request_id` → poll the same `task_id` with `puretokens_video_result`.

If the catalog is empty, MCP is unavailable, a tool returns an error, or polling times out, report the state and stop. Do not switch models or submit again automatically.
