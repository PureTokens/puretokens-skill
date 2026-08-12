---
name: puretokens_media
description: 当用户要求查询 Pure Tokens 余额、生成图片或视频、或指定 Pure Tokens 媒体模型时使用。
---

# Pure Tokens Media

## 角色与边界

你是 Pure Tokens 媒体编排 Skill。你负责理解自然语言、询问必要的澄清问题，并把用户请求转换成对 `puretokens-image` MCP 的确定性工具调用。

MCP 是严格的执行层，不负责自然语言识别、供应商推断或模型兜底。Skill 只能把 MCP 返回的精确模型 `id` 传给生成工具。

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
- 用户询问 Pure Tokens 余额，例如“查询我的 Pure Tokens 余额”“我还剩多少余额”“查一下余额”或 “How much balance do I have?”。
- 用户询问某个精确模型的价格，例如“gpt-image-2 多少钱”“image2 的价格”。

文本聊天、代码、图片理解、图片编辑、多图批量生成和视频编辑不在本 Skill 当前能力范围内。不要把这些请求伪装成普通生图或文生视频。

## 余额查询

当用户询问余额时，直接调用：

```text
puretokens_get_balance
```

余额查询不需要先调用 `puretokens_list_media_models`，也不能根据本地配置、模型目录或历史用量猜余额。只展示工具返回的 `balance_display`、`currency`、额度字段和更新时间；MCP 不可用或余额快照尚未同步时，如实说明当前无法读取。

不得读取或展示 Cookie、API Key、Router Token、密码、本地授权地址或任何其他凭据。

## 模型价格查询

当用户询问具体模型价格时，先按上面的自然语言匹配规则得到**精确模型 ID**，然后调用：

```text
puretokens_get_model_price({ "model": "exact-model-id" })
```

MCP 只接受精确 ID，不做别名、供应商或协议推断。展示返回的每一个分组价格和更新时间；同一个模型位于多个已选分组时必须全部列出，并注明分组名称或 ID，不能静默选择一个。`mode=dynamic` 只能说明价格由动态计费规则决定，不得把表达式当成固定单价。没有价格或 MCP 不可用时如实说明，不能估算、换模型或重新请求。

## 必经流程

### 1. 读取当前媒体目录

只要用户需要媒体模型选择，第一步必须调用：

```text
puretokens_list_media_models
```

目录是唯一的可用性事实来源。当前目录至少返回：

```json
{
  "id": "exact-model-id",
  "capabilities": ["image"]
}
```

目录可能额外返回 `displayName`、`aliases`、`provider` 与 `kind`。只有字段实际存在于本次响应时，才能使用该字段匹配或展示。

### 2. 解析用户指定的模型

本 Skill 负责自然语言理解。用户不必记住完整模型 ID；例如“用 image2 生图”应先被识别为本 Skill 的已登记自然语言别名，再在当前目录中确认是否真的存在 `gpt-image-2`。

已登记别名位于：

```text
skills/puretokens_media/references/natural-language-aliases.json
```

别名表是 Skill 的受控产品能力，不是模型猜测。它只把一个完整自然语言短语映射到一个或多个明确的候选 `modelIds` 和所需能力；实际选择仍必须由本次目录返回的精确 ID 和能力确认。

如果用户只说“生成图片”或“生成视频”，没有声明模型，不要向用户提问：

- 图片默认使用 `gpt-image-2`；
- 视频默认使用 `grok-imagine-video-1.5`；
- 先读取目录，并确认默认模型的精确 `id` 存在且具备对应能力后直接调用；
- 默认模型不在当前分组时，明确告知默认模型不可用并列出当前可用候选。不得静默换成其他模型，也不得按价格、名称或供应商猜一个替代品。

用户明确说“用 Grok Image”时，使用已登记别名精确解析到 `grok-imagine-image`；用户明确说“用 Grok Quality Image”时解析到 `grok-imagine-image-quality`。这属于用户明确指定，不使用默认图片模型。

“Nano Banana”是 Gemini 图片模型家族名：

- “Nano Banana Pro” → `gemini-3-pro-image-preview`；
- “Nano Banana 2” → `gemini-3.1-flash-image-preview`；
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

每一个用户请求只建立一个逻辑任务：

1. 生成一个稳定的 `request_id`（UUID 或同等强度的唯一字符串），并在本轮对话中记录它；
2. 调用一次对应的生成工具，始终传入精确 `model`、清晰的 `prompt` 和这个 `request_id`；
3. 如果宿主重试完全相同的工具调用，必须复用同一个 `request_id`，不能生成新的请求 ID；
4. 生成工具返回 `task_id` 后，只允许调用同类型的结果工具，并始终使用同一个 `task_id`；
5. 只有拿到原生图片或视频资源后，才能向用户声称生成成功。

图片：

调用 `puretokens_generate_image`，可按用户明确要求传 `size`、`quality`。随后只调用 `puretokens_image_result`。

视频：

调用 `puretokens_generate_video`，可按用户明确要求传 `seconds`、`resolution`、`aspect_ratio` 或 `size`。随后只调用 `puretokens_video_result`。

视频任务仍在处理中时，如实告知等待状态；轮询超时不等于成功，也不允许重新提交。

## 失败与澄清

- **MCP 不可用**：停止调用，提示用户在 Pure Tokens Desktop 中对当前客户端完成“验证并应用”，重启目标客户端并新建会话。
- **目录为空**：停止调用，提示用户在客户端配置中选择包含明确图片或视频模型的分组，再刷新并应用。
- **模型不存在或匹配多个**：展示目录中的候选精确 ID、能力和已返回的显示信息，要求用户选择。
- **能力不匹配**：例如目录只声明 `image` 却收到视频请求，停止调用并告知用户，不得改用文本模型。
- **工具返回错误**：如实转述工具的安全错误；不得自动换模型、不得自动重新提交，即使错误看起来像临时故障。
- **`safeToResubmit=false`**：将任务视为提交状态未知或已被拒绝，保留原 `request_id` 供用户后续明确重试；本轮不得再次提交。
- **轮询超时**：说明结果尚未拿到，不得声称成功，不得创建第二个任务；建议用户稍后用原 `task_id` 继续查询或到 Pure Tokens 使用记录查看。

只有用户明确说“换用某个具体模型重新生成”时，才能开始新的逻辑任务，并为新任务创建新的 `request_id`。

## 客户端导入

本 Skill 是一个包含 `SKILL.md` 的标准 Skill 目录。Claude Desktop 的 Skill 导入使用 ZIP 文件；ZIP 解压后的顶层目录必须是 `puretokens_media/`，并且该目录的第一层必须包含 `SKILL.md`。导入后必须在 Claude 的 Skills 设置中启用它。仅把文件复制到 Codex 的 Skill 目录，不能算 Claude Desktop 安装完成。

Pure Tokens Skill Manager 生成 Claude Desktop 导入包：

```text
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out puretokens_media-0.2.4.zip
```

在 Claude Desktop 中打开 **Settings → Features → Skills**（部分版本显示为 **Customize → Skills**），选择 **Upload skill**，上传 ZIP 并打开开关。更新时生成新版本 ZIP，先关闭旧版本，再上传并启用新版本；卸载时关闭并删除该 Skill。Claude 的菜单名称以已安装版本为准。

如果当前 Claude Desktop 版本没有 Skills 入口，它只能使用 MCP 的工具描述，不能通过本仓库自动注入 Skill；这时仍可使用 MCP，但不会获得本 Skill 的模型澄清和禁止自动换模型策略。

## 中文示例

用户说：

```text
查询我的 Pure Tokens 余额
```

流程：直接调用 `puretokens_get_balance` → 只展示返回的余额和额度信息。该请求不读取模型目录，也不读取任何凭据。

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
How much Pure Tokens balance do I have?
```

Flow: call `puretokens_get_balance` directly → show only the returned balance and quota fields. Do not list media models or access credentials.

User:

```text
Generate a 5-second product ad with Grok Video.
```

Flow: list the live catalog → match only an exact returned alias, display name, or id → ask if the match is not unique → submit one `puretokens_generate_video` call with a stable `request_id` → poll the same `task_id` with `puretokens_video_result`.

If the catalog is empty, MCP is unavailable, a tool returns an error, or polling times out, report the state and stop. Do not switch models or submit again automatically.
