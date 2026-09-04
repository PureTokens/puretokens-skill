---
name: puretokens-image
description: 通过当前 Pure Tokens 连接生成图片、编辑图片、选择图片模型、数量或尺寸时使用。
---

# Pure Tokens Image

只调用固定 Pure Tokens Images API：`https://api.puretokensx.com/v1/images/generations`，编辑仅在精确模型资料声明时调用 `/v1/images/generations` 或 `/v1/images/edits`。每次都使用完整固定 URL，不使用已配置 Base URL 作为请求目标，也不检查 provider 名称。

当前宿主使用其已有认证 HTTPS/API 能力，以当前匹配 Pure Tokens 连接的凭据完成请求；凭据仅可在内存中用于这一次固定请求。第三方 CC Switch 与手动宿主配置都可作为该连接来源。不得显示、索取、保存、记录或报告 Key、Base URL、请求头、完整宿主配置、原始响应或用户媒体；不得使用 Node、MCP、代理、sidecar、独立上传或备用 endpoint。

## 按需读取资料

不要在每个请求加载全部模型和全部异常场景。按下列顺序最少读取：

1. 先读 `references/model-index.json`：解析默认模型或唯一别名，并取得精确 profile 路径。
2. 只读被选中模型的 `references/profiles/<model>.json`：检查参数、约束、operation 和 lifecycle。未在索引中的精确模型只允许核心文生图字段；只有用户要求额外参数、参考图或编辑时，才按需查询实时目录。
3. 普通核心 POST 可只依据本 Skill、索引和选中 profile 执行；收到响应后、进入轮询/交付时，或构造编辑等非核心 operation 时，再读 `references/execution-contract.json` 与 `references/task-receipt.json`。
4. 只有出现对应异常、歧义、拒绝、附件或交付问题时，才读 `references/behavior-scenarios.json` 中的匹配场景。

## 模型与请求

- 普通文生图默认 `gpt-image-2`；`image2` 是其唯一别名。用户给出唯一别名时使用索引映射；零个或多个候选时列出精确 ID，请用户选择。用户给出精确 ID 时保留该 ID，普通核心请求不得为确认权限或能力而预读目录。
- 普通请求只提交一次 `POST /v1/images/generations`，始终带 `model`、`prompt`、`async: true`。仅当精确 profile 声明时才带 `n`、`size`、`image_size`、`aspect_ratio`、`width`、`height`、`strength` 或其他可选字段及值。
- 一次用户请求只能创建一个付费任务，绝不拆单。`n` 必须被 profile 明确允许；否则说明一次仅支持一张或请用户选支持该数量的模型。
- `200cm × 230cm` 等物理尺寸不能传 API。说明无法精确保证物理尺寸，并列出 profile 实际声明的像素尺寸、语义尺寸或比例选项。`width`、`height` 被声明为必须成对时，缺一即停止；多个尺寸表达同时给出时，只按 profile 的优先级提交一种。
- 只有用户明确问当前模型、所需参数/operation 不在本地 profile、附件 transport 不兼容，或 API 因模型/参数/capability 拒绝后需要解释时，才读取一次 `GET https://api.puretokensx.com/v1/media/models`。目录读取失败不能阻止本来合法的核心文生图，也不触发自动重试。

## 参考图与编辑

- 用户附图但没说明用途时，先问它是视觉参考还是待编辑图片；不得猜测。
- 公网图片只接受用户明确提供的合规 `https://` URL，并且只能放入 profile 声明的字段和 transport。不得下载、探测、转存、改写或生成 URL。
- 本地图片参考或编辑必须由选中 profile 的 `image_edit` multipart operation 明确支持，严格使用其路径、字段、数量和 `multipart_file` transport。只将当前附件随这一次 Images API 请求发送；不得上传到 R2、调用独立上传接口、生成 file ID、把附件内容转写成提示词，或静默降级为文生图。
- 参考图提示词保留“参考”的意图；编辑提示词明确“修改什么、保留什么”。mask、透明背景、未声明的多参考图和其他未声明扩展均在计费前停止并说明限制。

## 异步任务与交付

- 每个新请求最多一次 POST。若 POST 可能开始但没有可解析结果或任务 ID，提交结果为未知：不重提、不换模型、不声称未创建、未扣费或已退款。无任务 ID 时，后续新建请求必须得到用户明确确认。
- 只从声明的顶层字段或顶层 `task_id` / `id` 获取 task ID；只查询同一个 `https://api.puretokensx.com/v1/images/{task_id}`，并在成功后按 `content?index=0..n-1` 顺序交付。URL、状态文字、HTML、SVG 不是成功交付。
- 每个轮询窗口最多 6 次状态读取、最多 120 秒；常规等待为 3、6、12、24、30 秒，429 只在有效 `Retry-After` 仍符合本轮预算时继续同一任务。无后台轮询、无并发状态读取、无自动重提。
- 只有宿主实际交付原生图片字节或附件时才报告成功。逐张交付，不预取、不重复读取、不缓存用户媒体；成功后结束当前回复。

## 失败回执

宿主在 POST 前因网络、审批、附件传递或原生字节交付能力被阻止时，返回 `failure_phase: validation`，明确 Images API 未执行，且不猜测配置原因。其他失败按 `execution-contract.json` 和匹配行为场景输出安全回执：只报告 API 实际返回的 HTTP 状态和公开错误码；不暴露原始响应、内部 URL、堆栈、凭据、请求内容或用户媒体。
