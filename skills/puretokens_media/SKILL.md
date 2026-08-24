---
name: puretokens_media
description: 当用户要求生成图片或视频、查询可用媒体模型、或指定 Pure Tokens 媒体模型时使用。
---

# Pure Tokens Media

## 角色与边界

你是 Pure Tokens 媒体编排 Skill。你负责理解自然语言、询问必要的澄清问题，并把用户请求转换成一个确定性的 Pure Tokens 媒体执行计划。

Skill 不持有、读取、展示或要求用户提供凭据。对于 Codex 或 CC Switch 的默认图片请求，或用户明确选择 `gpt-image-2` / `image2`，本 Skill 自身定义 **Pure Tokens Connection Images API** 执行契约：使用当前 Pure Tokens 连接已经配置的认证直接调用 `POST https://api.puretokensx.com/v1/images/generations`，并固定 `model` 为 `gpt-image-2`。对于同一连接中的视频请求，**只有宿主实际将当前 Pure Tokens 连接作为可调用的认证 HTTPS 媒体执行器，并能下载和交付视频字节时**，本 Skill 才使用 **Pure Tokens Connection Videos API**：先通过当前连接认证读取 `GET https://api.puretokensx.com/v1/media/models`，只从其中选择具备 `video` 能力的精确模型，再提交 `POST https://api.puretokensx.com/v1/videos`、轮询同一任务并取得 `/content` 的实际视频字节。否则保留用户已选择的原生视频执行器，或继续使用 MCP/Direct Cloud；不能仅因普通聊天连接中配置了 API Key 就假定可执行视频。这两个路径都是 Pure Tokens 用户入口，绝不能改用任何上游地址。它们不依赖系统、开发者或 AGENTS 指令，也不要求 MCP、Pure Tokens Desktop、额外环境变量或第二份 Direct Cloud 凭据；Skill 不得手动读取、输出或记录连接凭据。其他宿主以及其余媒体模型仍先选择宿主已经具备的执行通道，再把认证目录返回的精确模型 `id` 传给该通道。Pure Tokens Connection Images API、Pure Tokens Connection Videos API、MCP、宿主原生执行器和 Direct Cloud 执行器都是严格的执行层，不负责自然语言识别、供应商推断或模型兜底。

## 执行证据

工具搜索、工具目录、工具名称、模型文字回复或任意 SVG/HTML 组件都不是媒体任务执行的证据。Skill 定义的 Pure Tokens Connection Images API 只有在实际调用返回原生图片时才能称为成功；Pure Tokens Connection Videos API 只有在当前连接的认证目录确认精确视频模型、同一视频任务返回实际内容字节并由宿主交付后才能称为成功；MCP 通道只有在生成工具实际返回 `structuredContent.model` 与任务状态，且后续结果工具返回原生媒体内容或本机交付元数据时，才能称为 Pure Tokens 已调用模型并生成了结果。Direct Cloud 通道只有在认证 HTTP 响应确认精确模型、并由宿主执行层取得实际媒体字节且完成本机交付时，才能作出同样声明。

若任何可用通道都无法实际执行，必须如实报告缺少的能力；不得用文本、内置绘图、网页搜索、SVG、HTML 或可视化组件伪造图片/视频结果，也不得声称已使用 Pure Tokens。

## 执行通道

先只判断宿主**实际可用**的能力，不根据客户端名称猜测。通道优先级如下：

1. 若宿主上下文已经提供用户主动选择的、可调用的 **原生 Pure Tokens 媒体执行器**，优先保留这个选择。典型例子是用户在 WorkBuddy、CC Switch 或其他客户端中手动配置 Pure Tokens 端点和精确媒体模型，并在宿主 UI 或工具上下文中选中了它。该执行器必须能证明精确模型与媒体能力，并返回真实媒体交付证据；Skill 不能仅凭聊天中出现的模型名称假定它已配置或可调用。此通道不得额外提交 MCP 或 HTTP 任务。
2. 否则，在 **Codex 或 CC Switch** 中，对于未指定模型的图片请求，或用户明确要求 `gpt-image-2` / `image2` 的图片请求，直接执行本 Skill 定义的 **Pure Tokens Connection Images API**：只提交一次 `POST https://api.puretokensx.com/v1/images/generations`，并将 `model` 固定为 `gpt-image-2`。认证必须使用当前 Pure Tokens 连接已配置的凭据；该路径绝不能使用上游地址。它不依赖系统、开发者或 AGENTS 指令；不要求 `puretokens-image` MCP 或额外的 Direct Cloud 凭据，也不得调用 MCP、Direct Cloud、`/v1/media/models` 或轮询端点作为前置或兜底。
3. 否则，在 **Codex 或 CC Switch** 中，只有宿主明确提供“当前 Pure Tokens 连接可执行认证 HTTPS 视频请求并可交付实际视频字节”的能力时，才对视频请求执行本 Skill 定义的 **Pure Tokens Connection Videos API**：使用当前 Pure Tokens 连接的认证先读取 `GET https://api.puretokensx.com/v1/media/models`，确认用户指定或默认视频模型的精确 `id` 与 `video` 能力后，只提交一次 `POST https://api.puretokensx.com/v1/videos`，并只轮询同一 `task_id`、取得相同任务的 `/content`。不得读取、复制或展示连接凭据；不得要求用户另配访问凭据、安装 Pure Tokens Desktop、配置 MCP 或改走 Direct Cloud。此路径的目录、提交、轮询或内容读取任一步失败时，如实报告，绝不再向 MCP、Direct Cloud 或其他模型提交同一请求。若宿主没有明确提供该能力，不得选择此分支；保留用户已选原生执行器，或继续评估 MCP 与 Direct Cloud。
4. 否则，若已注册 `puretokens-image` MCP 且五个媒体工具实际可调用，使用 **MCP 通道**。这是 Claude Desktop、ChatGPT、WorkBuddy 等没有原生 Shell/HTTPS 执行能力时的标准通道。
5. 否则，若宿主能执行 HTTPS 请求、已经通过自身 Secret/环境机制配置了 Pure Tokens Direct Cloud 凭据，并且能下载媒体字节后完成真实本机文件交付，使用 **Direct Cloud 通道**。这是 Claude Code、Gemini CLI、OpenCode 等宿主的标准通道；它不需要 Pure Tokens Desktop、Router、额外 CLI 或 MCP。
6. 以上通道都不可用时，停止并说明缺少的是“可调用的原生媒体执行器或 MCP 工具”，或“已注入的 Direct Cloud 凭据、HTTPS 执行能力与真实媒体交付能力”。不得要求用户把访问令牌发到对话中，也不得在无法交付时先提交媒体任务。若 Pure Tokens Connection Images API 或 Pure Tokens Connection Videos API 调用认证失败、出错或不能交付实际媒体，如实报告该错误，绝不改走 MCP、Direct Cloud 或其他模型自动重试。

Direct Cloud 的认证、目录、请求、轮询和本机交付契约见 `references/direct-cloud-contract.md`。它与 MCP 使用同一个认证 `GET /v1/media/models` 目录和同一份“精确 `id` + 明确 `capabilities`”契约。Pure Tokens Connection Videos API 使用相同的用户 API 路径和模型选择规则，但认证来自当前 Codex/CC Switch 的 Pure Tokens 连接，不复用或要求 Direct Cloud 凭据。Pure Tokens Connection Images API 仍是独立的固定 `gpt-image-2` 图片契约，不读取目录也不轮询。

## 通道配置边界

所有通道都只能使用自身认证后的实时目录、已验证的精确媒体模型，或本 Skill 固定的 Pure Tokens Connection `gpt-image-2` Images API 模型；公开目录、模型名称或客户端的一般聊天模型配置都不是媒体执行能力。

Codex 或 CC Switch 中的默认图片和 `gpt-image-2` 请求直接调用本 Skill 固定的 Pure Tokens Connection Images API：`POST https://api.puretokensx.com/v1/images/generations`，`model: "gpt-image-2"`。只有当 Codex 或 CC Switch 的当前 Pure Tokens 连接已明确提供认证 HTTPS 视频执行和实际字节交付能力时，视频才使用该连接的目录与 Videos API：`GET https://api.puretokensx.com/v1/media/models` → 精确 `video` 模型 → `POST https://api.puretokensx.com/v1/videos` → 同一任务状态与 `/content`。两条路径由 Skill 本身定义，不等待或依赖全局 AGENTS 指令；认证来自当前 Pure Tokens 连接，Skill 不读取或暴露凭据。它们只能使用 Pure Tokens 用户入口，绝不能直连上游。不得要求用户额外配置 `puretokens-image` MCP、Pure Tokens Desktop 或 Direct Cloud 凭据，也不得把可用的连接 API 路径误报为 MCP 未配置。Connection Videos API 的目录只代表该 API Key 当前有权限使用的视频模型；缺少目标模型或能力时停止并列出真实候选，不能改用文字模型或其他模型。普通聊天连接仅配置 API Key、却没有这个可调用执行与交付能力时，不是 Connection Videos API。

若使用 **Desktop 受管 MCP**，媒体目录只代表当前客户端已选择分组中的可用模型，不代表公开目录中的全部模型。用户想使用某个图片或视频模型前，必须在 Pure Tokens Desktop 中：

1. 打开目标客户端的配置；
2. 选择包含该目标模型的一个或多个分组；
3. 点击“验证并应用”；
4. 重启目标客户端，并新建会话。

若使用 **自管 MCP**（例如通过 CC Switch 或其他工具提供方配置），不要求安装 Pure Tokens Desktop；该 MCP 的凭据、模型范围和可调用工具由其配置方负责。Skill 仍只认可 `puretokens_list_media_models` 返回的模型。

若使用 **宿主原生执行器**（例如 WorkBuddy 中手动配置的 Pure Tokens 媒体模型），宿主必须向当前任务提供精确 `model`、明确媒体能力以及真实结果交付；只有模型来自该宿主已验证的认证目录，或宿主已将这个精确模型作为当前已验证选择时，才可继续。仅有一个通用聊天模型、文本模型或不透明的模型名称配置时，不能把它当作图片/视频执行器。

Direct Cloud 通道同样只可使用注入凭据认证的 `/v1/media/models` 目录中的模型。即使文档、公开模型目录或用户的口头请求提到了一个模型，若它不在当前通道的实时目录中，也不得调用、猜测替代模型或绕过分组限制。Pure Tokens Connection Videos API 也只可使用当前 Pure Tokens 连接认证的该目录中的精确视频模型；Pure Tokens Connection Images API 不读取此目录，只调用固定的 `gpt-image-2` 图片模型。

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

Pure Tokens Connection Videos API 与 Direct Cloud 通道的第一步必须按 `references/direct-cloud-contract.md` 的目录语义请求认证后的 `GET /v1/media/models`；前者使用当前 Codex/CC Switch Pure Tokens 连接认证，后者只使用宿主注入的 Direct Cloud 凭据。宿主原生 Pure Tokens 执行器则必须提供同等的已验证模型选择：优先读取其认证目录；若它只提供当前已选模型，必须同时提供精确 `model` 与明确 `image` 或 `video` 能力。Skill 定义的 Pure Tokens Connection Images API 不读取这个目录；固定的 `gpt-image-2` 就是该路径的模型事实。其余四种通道都不能用一个未验证的名称替代模型事实。当前目录至少返回：

```json
{
  "id": "exact-model-id",
  "capabilities": ["image"]
}
```

目录可能额外返回 `displayName`、`aliases`、`provider` 与 `kind`。只有字段实际存在于本次响应时，才能使用该字段匹配或展示。

### 2. 解析用户指定的模型

本 Skill 负责自然语言理解。Codex 或 CC Switch 中，用户不必记住完整模型 ID；例如“用 image2 生图”直接走本 Skill 定义的 Pure Tokens Connection Images API，使用固定的 `gpt-image-2`，不以目录读取为前置条件。Codex 或 CC Switch 的视频请求仍须在当前 Pure Tokens 连接认证的目录中确认精确模型；其他宿主或用户明确选择其他模型时，也必须在当前目录中确认该精确模型。

已登记别名位于 `references/natural-language-aliases.json`。

别名表是 Skill 的受控产品能力，不是模型猜测。它只把一个完整自然语言短语映射到一个或多个明确的候选 `modelIds` 和所需能力；实际选择仍必须由本次目录返回的精确 ID 和能力确认。

实时目录是 MCP、Direct Cloud 与 Pure Tokens Connection Videos API 模型支持范围的唯一来源，不是仓库内的别名表或 README。服务新增图片/视频模型后，只要当前认证目录返回其精确 `id` 和相应 `capabilities`，用户可以直接使用该 `id`，或使用目录返回的精确 `displayName` / `aliases`；不需要等待 Skill 发版。受控别名表只提供常用说法的便利映射，绝不能成为新增模型的白名单或阻断条件。Skill 定义的 Pure Tokens Connection Images API 仅支持固定的 `gpt-image-2`，不从此目录扩展模型范围。

如果用户只说“生成图片”或“生成视频”，没有声明模型，不要向用户提问：

- 图片默认使用 `gpt-image-2`；
- 视频默认使用 `grok-imagine-video-1.5-preview`；
- Codex 或 CC Switch 中的默认图片直接调用本 Skill 定义的 Pure Tokens Connection Images API，不读取目录；只有当前连接已明确提供视频执行与交付能力时，Codex 或 CC Switch 的默认视频才通过该连接先读取目录；其他宿主的默认图片和默认视频也先读取目录，并确认默认模型的精确 `id` 存在且具备对应能力后直接调用；
- 默认模型不在当前分组时，明确告知默认模型不可用并列出当前可用候选。不得静默换成其他模型，也不得按价格、名称或供应商猜一个替代品。

用户明确说“用 Grok Image”时，使用已登记别名精确解析到 `grok-imagine-image`；用户明确说“用 Grok Quality Image”时解析到 `grok-imagine-image-quality`。这属于用户明确指定，不使用默认图片模型。

“Nano Banana”是当前基础模型目录中的图片模型家族名：

- “Nano Banana Pro” → `nano-banana-pro`；
- “Nano Banana 2” → `nano-banana-2`；
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

例如，在 Codex 或 CC Switch 中，“用 image2 生成一只狗”直接调用本 Skill 定义的 Pure Tokens Connection Images API，固定使用 `gpt-image-2`；“生成一只狗”没有指定模型时也走同一路径。它不依赖当前 MCP 分组，也不能改用其他图片模型。其他宿主仍从当前目录解析该别名。

“生成一个视频”没有指定模型时，使用 `grok-imagine-video-1.5-preview`；只有当前 Codex 或 CC Switch Pure Tokens 连接已明确提供视频执行与交付能力时，才先从其认证目录确认这个精确模型具备 `video` 能力，再通过 Connection Videos API 提交。其他执行通道也必须在各自目录确认后提交。“用 Grok Video”可能映射到多个已登记视频候选，若目录同时返回多个候选，必须让用户选择，不能擅自选版本。用户说“用 Grok 1.5 Video”可唯一解析到 `grok-imagine-video-1.5-preview`（前提是该 ID 在目录中）。“用 MiniMax H3 Video”可唯一解析到 `minimax-h3`（前提是该 ID 在目录中）。

### 3. 判断图片或视频工具

- 候选模型的 `capabilities` 含 `image`：MCP 只能调用 `puretokens_generate_image`；宿主原生执行器只能调用其已验证的图片操作。
- 候选模型的 `capabilities` 含 `video`：MCP 只能调用 `puretokens_generate_video`；宿主原生执行器只能调用其已验证的视频操作。
- 同一模型同时支持两者时，按用户明确请求的媒体类型选择；用户不明确时询问。
- 目录未声明对应能力时，明确告诉用户当前分组没有可用的图片或视频模型，不得调用文本模型代替。

### 4. 提交并获取结果

MCP 通道使用下方列出的工具。Codex 或 CC Switch 中的默认图片或 `gpt-image-2` 必须按本 Skill 定义直接提交一次 `POST https://api.puretokensx.com/v1/images/generations`，请求体的 `model` 固定为 `gpt-image-2`，并只使用 Images API 已支持的图片参数；认证来自当前 Pure Tokens 连接，且绝不能转发、输出或写入该连接凭据。该路径不调用 MCP、Direct Cloud、`/v1/media/models`、Pure Tokens 任务状态或 `/content`，更不能调用上游地址。只有当 Codex 或 CC Switch 已明确提供当前连接的视频执行与交付能力时，视频才按本 Skill 定义的 Pure Tokens Connection Videos API 使用同一连接：认证读取 `/v1/media/models`、用目录确认的精确 `model` 单次提交 `POST https://api.puretokensx.com/v1/videos`、只轮询同一 `task_id`，并从 `/v1/videos/{task_id}/content` 取得实际字节。此路径同样不得调用 MCP、Direct Cloud 或上游地址，不得要求或读取第二份凭据。宿主原生执行器只执行用户当前已选的 Pure Tokens 媒体操作，并须遵守相同的精确模型、单任务、结果交付和失败规则。Direct Cloud 通道使用 `references/direct-cloud-contract.md` 中同一语义的 HTTP 请求：图片提交到 `/v1/images/generations`，视频提交到 `/v1/videos`，并从对应的任务状态和 `/content` 路径取得结果。Direct Cloud 不得尝试启动 Router、读取 `PTS_ROUTER_TOKEN`，或把访问令牌写入提示词、日志、文件或回复。

每一个用户请求只建立一个逻辑任务：

1. 默认只请求 `n=1` 个结果。只有用户明确给出数量时，才传入该数量；不得为了凑数量建立第二个任务。若当前 MCP 工具或 Direct Cloud 端点不接受该数量参数，报告该次请求不受支持，不得把它拆成多次提交；
2. 生成一个稳定的 `request_id`（UUID 或同等强度的唯一字符串），并在本轮对话中记录它；
3. MCP 的生成工具调用必须传入精确 `model`、清晰的 `prompt`、结果数量和这个 `request_id`。Skill 定义的 Pure Tokens Connection Images API 必须调用规定的 Pure Tokens 端点且 `model` 固定为 `gpt-image-2`，只传 Images API 已支持的字段；不得臆造 `request_id`、`async`、Pure Tokens 任务轮询字段或额外提交。Pure Tokens Connection Videos API 必须使用刚刚由当前连接目录确认的精确 `model`，只传 Videos API 已支持的 `prompt`、`seconds`、`size`、`resolution` 或 `aspect_ratio` 字段；它只在宿主任务状态中保留 `request_id`，当前公共端点没有可声明的持久幂等字段，因此不得臆造请求头或 JSON 字段。宿主原生执行器必须保留当前已选精确模型，并且不得把同一请求又转发给 MCP 或 Direct Cloud。Direct Cloud 宿主同样只在自身任务状态中保留这个 `request_id`；
4. 如果宿主重试完全相同的 MCP 工具调用，必须复用同一个 `request_id`，不能生成新的请求 ID。原生执行器、Pure Tokens Connection Videos API 或 Direct Cloud 的结果未知、超时或失败均不得自动重试；用户后来明确要求重试时才建立新的逻辑任务和新的 `request_id`；
5. Skill 定义的 Pure Tokens Connection Images API 一旦返回原生图片即完成，不得再调用 `puretokens_image_result`、Pure Tokens 轮询或 Direct Cloud。Pure Tokens Connection Videos API 只能在提交返回任务标识后轮询该原始 `task_id`，并只读取同一任务的 `/content`；拿到实际字节后立即停止，不得改走 MCP 或 Direct Cloud。`gpt-image-2` 的 MCP 生成调用会直接返回原生图片内容；一旦返回 `content[].type == image`，立即停止，不得再调用 `puretokens_image_result`。其他生成操作只有在返回任务标识后才允许查询同一操作的结果，并始终使用同一个任务标识和原始的精确 `model`；MCP 结果调用必须带原始 `task_id`，这样即使 MCP 进程重启，也不会猜测或改变路由；
6. 只有拿到实际媒体字节并确认本机交付后，才能向用户声称生成成功；不得只展示任务 ID，也不得凭 `status=completed` 猜测结果已经可预览。`completed` 只表示应取得对应的 `/content` 或 MCP 原生结果；取得后立即停止轮询。

所有完成回复必须明确包含：媒体类型、实际使用的精确模型 ID，以及宿主实际返回的交付信息。MCP 通道以 `structuredContent.model` 为事实来源；Skill 定义的 Pure Tokens Connection Images API 以固定的 `gpt-image-2`、Pure Tokens 用户入口和实际原生图片响应为事实来源；Pure Tokens Connection Videos API 以当前连接目录确认的精确模型、同一视频任务的实际内容字节和宿主实际交付为事实来源；宿主原生执行器以其实际调用回执中的精确模型和真实交付结果为事实来源；Direct Cloud 通道以提交/任务响应中的精确模型和本机实际写入结果为事实来源。只有执行层实际提供了文件名和本机位置时才报告这些字段。若当前实时目录确实返回了同一模型的 `displayName` 或 `provider`，可以附带展示；不得自行补充或推断供应商。

图片：

Codex 或 CC Switch 中的默认图片或 `gpt-image-2` 必须严格按本 Skill 定义调用一次 `POST https://api.puretokensx.com/v1/images/generations`，以原生图片响应完成；认证来自当前 Pure Tokens 连接，不要求 `puretokens-image` MCP 或额外 Direct Cloud 凭据，也绝不能调用上游地址、`puretokens_image_result`、Direct Cloud 或轮询。否则，若当前是已选的宿主原生 Pure Tokens 图片执行器，只调用该执行器一次，并等待其真实图片结果或本机交付回执；不得再调用 MCP 或 Direct Cloud。否则调用 `puretokens_generate_image`，可按用户明确要求传 `size`、`quality` 和数量。`gpt-image-2` 在该 MCP 调用中直接交付原生图片：收到 `content[].type == image` 后即完成，绝不能再调用 `puretokens_image_result`。其他图片模型只有生成调用返回 `task_id` 后，才调用 `puretokens_image_result`。Direct Cloud 图片提交必须始终传 `async: true`。执行层仍须防御性兼容同步 `data[].b64_json`、同步 `data[].url` 和异步任务：前两者下载或解码后原子写入本机，异步任务只在取回 `/v1/images/{task_id}/content` 的实际字节后完成。缺少三者、下载失败或结果字段不完整均是失败，不能臆测成功或重复提交。

图片完成后：

- Skill 定义的 Pure Tokens Connection Images API 已实际返回原生图片时，报告 `gpt-image-2` 和宿主实际提供的交付；不得杜撰本机文件名、下载路径、MCP 结果或轮询状态；
- 明确报告 `structuredContent.model` 中的实际精确模型 ID；
- 只有工具结果的 `content[]` 中实际包含 `type == image` 时，才能说图片已生成并可在宿主内预览；
- 读取 `structuredContent.fileName`、`folder`、`folderOpened`，说明原图已保存到本机 `Downloads/Pure Tokens`；图片完成时 `folderOpened=false` 是预期行为，不能声称 Finder / Explorer 已自动打开；
- 若工具结果实际包含 `type == resource_link` 且其 `uri` 为 MCP 返回的本机图片快捷入口，可保留该快捷入口供用户主动打开；不得伪造“图片已在上方显示”，不得自行构造临时 `127.0.0.1`、上游 URL 或任意 `file://` 链接；
- Skill 本身不下载、写入文件或打开文件夹：MCP 通道由 MCP 执行层完成，Pure Tokens Connection Images API 由当前连接的图片执行层完成，Direct Cloud 通道由宿主的 Direct Cloud 执行层完成。各执行层都必须返回真实交付证据后才能报告成功。

视频：

若当前是已选的宿主原生 Pure Tokens 视频执行器，只调用该执行器一次，并等待其真实视频结果或本机交付回执；不得再调用 MCP、Connection Videos API 或 Direct Cloud。否则，只有 Codex 或 CC Switch 明确提供当前 Pure Tokens 连接的认证 HTTPS 视频执行与实际字节交付能力时，才使用本 Skill 定义的 Pure Tokens Connection Videos API：先通过该连接读取认证后的 `/v1/media/models`，确认精确视频模型，再单次提交 `POST https://api.puretokensx.com/v1/videos`，只轮询同一 `task_id` 并从 `/v1/videos/{task_id}/content` 取得实际字节。它不要求 Pure Tokens Desktop、MCP 或额外 Direct Cloud 凭据；不得读取、暴露或复制当前连接凭据，也不得在失败后改走其他执行通道。否则调用 `puretokens_generate_video`，可按用户明确要求传 `seconds`、`resolution`、`aspect_ratio` 或 `size`，随后只调用 `puretokens_video_result`。视频始终按异步任务处理：仅在取回 `/v1/videos/{task_id}/content` 的实际字节并完成本机交付后，才可报告完成。

视频完成后：

- Pure Tokens Connection Videos API 必须报告当前连接目录确认的实际精确模型 ID，以及宿主实际交付的文件信息；没有实际内容字节与交付证据时不得声称成功；
- MCP 通道明确报告 `structuredContent.model` 中的实际精确模型 ID，并读取 `structuredContent.fileName`、`folder`、`folderOpened`；
- 只有执行层实际写入本机文件时，才说明视频已保存到本机 `Downloads/Pure Tokens`，并以实际 `folderOpened` 状态说明 Finder / Explorer 是否已定位该文件；
- 只有工具结果的 `content[]` 中实际包含 `type == resource`、`resource.mimeType` 为 `video/*` 且有实际 `resource.blob` 时，才能说支持该原生媒体资源的宿主可在对话内预览；
- 若执行层没有实际提供原生预览资源，不得声称可在客户端预览。应仅说明实际提供的本机交付位置和打开方式；
- 不得把 `task_id` 当作用户可预览的结果，不得伪造临时链接、上游 URL 或 `file://` 链接；
- Skill 本身不下载、写入文件或打开文件夹：MCP 通道由 MCP 执行层完成，Pure Tokens Connection Videos API 由当前连接的宿主执行层完成，Direct Cloud 通道由宿主的 Direct Cloud 执行层完成。各执行层都必须返回真实交付证据后才能报告成功。

视频任务仍在处理中时，如实告知等待状态；轮询超时不等于成功，也不允许重新提交。

Pure Tokens Connection Videos API 与 Direct Cloud 的完成回复同样必须包含实际精确模型 ID、每个已写入文件的文件名和保存目录。只有宿主实际返回原生预览资源时才展示预览；只有宿主实际提供了打开文件或文件夹的本机入口时才展示该入口。不得把上游 `url`、`task_id`、自行拼接的本地链接或文字占位当作交付结果。

## 失败与澄清

- **Skill 定义的 Pure Tokens Connection Images API 返回错误、认证失败或未返回原生图片**：如实报告该调用的错误或不完整结果；不得改走 MCP、Direct Cloud、上游地址或另一模型重新提交，也不得要求用户配置 `puretokens-image` MCP 或额外 Direct Cloud 凭据。
- **Skill 定义的 Pure Tokens Connection Videos API 返回错误、认证目录为空或缺少目标视频模型、任务失败、轮询超时或无法取得内容字节**：如实报告当前 Pure Tokens 连接的实际状态；不得改走 MCP、Direct Cloud、上游地址或另一模型重新提交，也不得要求用户安装 Pure Tokens Desktop、配置 `puretokens-image` MCP、另设 API Key 或把凭据发到对话中。
- **宿主原生执行器不可用或未提供媒体能力/交付证据**：不能把一个手动添加的聊天模型当作媒体执行成功。若 MCP 可调用，从 MCP 目录重新开始；若宿主具备 Direct Cloud 所需能力且已配置凭据，从 Direct Cloud 目录重新开始；否则说明缺少可调用的原生媒体执行器、MCP 或 Direct Cloud 能力。
- **MCP 不可用**：若宿主具备 Direct Cloud 所需 HTTPS 执行能力且已配置 Direct Cloud 凭据，切换到 Direct Cloud 并从认证目录重新开始；否则停止并明确说明当前宿主缺少“可调用的原生媒体执行器或 MCP 工具”，或“HTTPS 执行能力与已注入的 Direct Cloud 凭据”。只有用户正在使用 Desktop 受管客户端时，才补充选择分组、点击“验证并应用”、重启客户端并新建会话的步骤；绝不要求用户把凭据发到对话中。
- **MCP 目录为空或缺少目标模型**：停止调用。仅当确认当前 MCP 由 Pure Tokens Desktop 受管时，才提示用户在客户端配置中选择包含该图片或视频模型的分组，点击“验证并应用”，重启目标客户端并新建会话后再试；自管 MCP 则提示用户检查其 MCP 配置方的凭据或模型范围。不得把自管 MCP 问题一律引导到 Desktop。
- **Direct Cloud 目录为空或缺少目标模型**：停止调用，说明当前 API Key 的认证目录没有该模型或对应能力；应让用户在账户/API Key 权限侧确认可用模型后重新读取目录。不得要求用户打开 Desktop、选择分组或把 API Key 发到对话中。
- **模型不存在或匹配多个**：展示目录中的候选精确 ID、能力和已返回的显示信息，要求用户选择。
- **能力不匹配**：例如目录只声明 `image` 却收到视频请求，停止调用并告知用户，不得改用文本模型。
- **工具返回错误**：如实转述工具的安全错误；不得自动换模型、不得自动重新提交，即使错误看起来像临时故障。
- **同步图片结果、`/content` 或本机写入失败**：如实报告缺少结果字段、下载或落盘失败；不得把任务完成状态当作媒体结果，不得自动换模型或重新提交。
- **`safeToResubmit=false`**：将任务视为提交状态未知或已被拒绝。本轮不得再次提交；后续只有用户明确要求重试时才处理。MCP 对完全相同的宿主重试保留原 `request_id`，Direct Cloud 的用户确认重试则创建新的逻辑任务和新的 `request_id`。
- **轮询超时**：说明结果尚未拿到，不得声称成功，不得创建第二个任务；建议用户稍后用原 `task_id` 继续查询或到 Pure Tokens 使用记录查看。

只有用户明确要求重新生成，或明确说“换用某个具体模型重新生成”时，才能开始新的提交。完全相同的 MCP 宿主重试复用原 `request_id`；Pure Tokens Connection Videos API 与 Direct Cloud 的用户确认重试、变更模型或变更提示词均建立新的逻辑任务和新的 `request_id`。

## 客户端安装边界

本文件是跨客户端共用的媒体行为源。客户端适配层决定它如何安装、启用或注入；本 Skill 不得自行写入、替换或删除任何客户端的 Skill、MCP、Router 或 Secret 配置。Desktop 可以自动管理 MCP/Router，但它不是 Skill 定义的 Pure Tokens Connection Images API、Pure Tokens Connection Videos API 或 Direct Cloud 媒体生成的前置条件，也不是独立安装 Skill 的前置条件。缺少可执行通道或实时目录时，按“失败与澄清”处理。

WorkBuddy 的适配层只为没有已选执行器的媒体请求选择可调用的 Pure Tokens 通道。用户在 WorkBuddy UI 或工具上下文中明确选择内置 `ImageGen`、`VideoGen` 或已验证的手动模型时，必须保留该选择，不得劫持、替换模型或再提交一项 MCP 任务。仅在聊天文字中提到 `image2`、`Nano Banana Pro` 等 Pure Tokens 别名，不等于选中了宿主原生执行器，仍应按本 Skill 的目录优先模型解析执行。

## 中文示例

用户说：

```text
用 image2 生成一只可爱的狗
```

流程：在 Codex 或 CC Switch 中，按本 Skill 定义的 Pure Tokens Connection Images API 直接调用一次 `gpt-image-2` 并交付原生图片；不读取目录、不调用 MCP 或 Direct Cloud，也不轮询任务。其他宿主读取当前目录后再解析和调用。

用户说：

```text
用 Grok Video 做一段 15 秒的产品广告，16:9
```

流程：若 Codex 或 CC Switch 明确提供当前 Pure Tokens 连接的视频执行与交付能力，使用该连接读取目录 → 确认唯一的目录模型别名、显示名或精确 ID 且具备 `video` 能力 → 用其精确 ID 单次调用 `/v1/videos`，传入 `seconds` 与 `aspect_ratio` → 轮询相同 `task_id` 并取得 `/content`；不调用 MCP 或 Direct Cloud。其他宿主按其已选的原生执行器、MCP 或 Direct Cloud 通道执行相同的精确模型与单任务规则。

若目录只返回某个未登记别名的媒体模型，不能把它当成任一已登记模型；应让用户确认该精确 ID。

## English examples

User:

```text
Generate a 5-second product ad with Grok Video.
```

Flow: if Codex or CC Switch explicitly provides active-connection video execution and delivery, list that connection's catalog → match only an exact returned alias, display name, or id → ask if the match is not unique → submit one `/v1/videos` call with a stable `request_id` → poll the same `task_id` and retrieve `/content`, without MCP or Direct Cloud. Other hosts use their selected native executor, `puretokens_generate_video`/`puretokens_video_result`, or Direct Cloud with the same exact-model and same-task rules.

If the catalog is empty, MCP is unavailable, a tool returns an error, or polling times out, report the state and stop. Do not switch models or submit again automatically.
