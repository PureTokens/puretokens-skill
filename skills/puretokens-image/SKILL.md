---
name: puretokens-image
description: 当前宿主使用 Pure Tokens 连接时，任何生成、创建、绘制或编辑图片的请求都优先使用本 Skill，不得先调用通用 imagegen 或 Imagen。
---

# Pure Tokens Image

## 路由优先级

当前宿主使用 Pure Tokens 连接时，本 Skill 优先处理图片生成和编辑；不先调用通用 imagegen 或 Imagen。必须调用安装的原生执行器；它是唯一 API 传输。不得自行发 HTTP、读取或传递凭据、改用其他媒体 Skill、MCP、Computer Use、浏览器或桌面自动化。

请求固定发送到 `https://api.puretokensx.com`。执行器只在内存中读取当前宿主明确配置的匹配凭据。用户不需要 Node、Python、Go 或 Pure Tokens Desktop。不要以 provider 标签判断连接，也不要假定聊天请求的认证会自动传给其他工具。

## 按需读取

不要在每个请求加载全部模型或全部异常规则。先读 `references/model-index.json`，默认模型为 `gpt-image-2`，唯一别名映射到精确 ID；歧义先让用户选择。只读选中模型的 `references/profiles/<model>.json`。

普通核心 POST 可只依据本 Skill、索引和选中 profile 执行。执行器也会从安装目录读取该 profile 并校验参数。未知精确模型允许只有 model/prompt 的核心请求；额外参数或媒体操作缺少 profile 时，执行器才读取一次实时目录。不得把查余额、init 或拉取最新目录设为每次生成的前置条件。

只有出现对应异常或歧义时，才读 `references/behavior-scenarios.json` 的匹配场景。完整命令和回执见 `references/executor-usage.md`；按需读取 `references/execution-contract.json` 和 `references/task-receipt.json`。

## 请求与附件

从当前 SKILL.md 的绝对位置解析同级 `../.puretokens-executor/puretokens-api`，Windows 使用 `puretokens-api.exe`。不要依赖终端工作目录或 PATH。当前宿主 ID 取 `codex`、`claude-code`、`workbuddy`、`gemini-cli`、`grok-build`、`opencode`、`trae` 中实际运行的一个；不能猜测其他宿主的配置。

用宿主文件工具创建仅本次请求的 UTF-8 JSON 文件，不在命令行放 prompt 或附件内容，不包含 Key、Base URL、headers。调用：

```text
<executor-absolute-path> submit --host <host-id> --request <request-json-absolute-path>
```

收到回执后清理临时请求文件。请求包含 `kind: "image"`、`operation: "generate"` 或 `"edit"`、精确 `model`、`prompt`、声明的 `parameters` 和可选 `attachments: [{"field":"精确字段","path":"当前附件绝对路径"}]`。`requested_count` 可省略，执行器从 `parameters.n` 推导；两个数量字段若同时提供必须一致，绝不能把尺寸传给数量。

只使用 profile 声明的字段和值。物理尺寸如 `200cm × 230cm` 无法精确满足，不传 API；明确说明限制并列出当前模型支持的尺寸或比例。数量不支持时请用户换支持该数量的模型，不能自动拆成多个付费任务。尺寸、时长、分辨率和数量被拒绝时，不自动改参重试。

当前附件要声明精确 `media_operation`，例如 `image_edit`、`image_to_video`、`reference_image_video`、`reference_video`、`reference_audio`、`video_edit`，并使用该 operation 的精确字段和数量。用户没说明首帧、参考或编辑用途时先澄清。只有一个声明的组合 operation 才能混合附件类别。

用户明确给出的公网 HTTPS URL 只能放到 profile 声明的 JSON 参考字段，使用 `generate`；不要下载、探测、转存或改写 URL。本地附件通过声明的 multipart 随一次媒体请求发送；不得单独上传、生成 URL/file ID、读取旧附件或降级为纯文本。执行器决定 profile 声明的固定 API 路径和表示方式。

`gpt-image-2` 本地编辑使用 `POST https://api.puretokensx.com/v1/images/edits`、`media_operation: "image_edit"` 和 `image` 字段；公网参考图使用 generations 的 `parameters.image`，不能混淆。

## 异步与交付

1. 每个新请求只调用一次 submit。创建成功立即输出 task_id 并结束命令；先把 ID 和状态告知用户，再开始等待。没有可解析回执时结果未知，不声称未创建、未扣费或已退款，不自动重提。
2. 用 `wait --host <host-id> --request <same-task-json>` 等待原任务；`status` 只读一次。图片每个窗口最多 6 次、120 秒；视频最多 7 次、300 秒。截止时间包含网络等待，遵守 Retry-After。超时或状态读取失败保留 task_id，询问是否继续该任务，不能新建替代任务或后台轮询。
3. 完成后用 `content`，传相同 task_id、`task_status: "completed"`、已确认的 requested_count、零起点 index 和现有输出目录绝对路径。每次只取一个 index；图片按 0..n-1，视频只取 0。将该文件实际交给用户后，才取下一份。继续下载只取尚未交付的索引。
4. `downloaded_awaiting_host_delivery` 只表示已验证文件落盘。必须用当前宿主附件交付方式把文件交给用户，才能说交付完成。URL、HTML、SVG、空文件、状态文字或 task_id 都不是媒体成功。无法附加时说明“生成完成、已下载，但当前宿主无法交付附件”，保留原任务与文件。

输出目录由用户目标或当前工作区明确选择；执行器不创建隐藏媒体缓存。验证通过的同任务文件可以直接复用，避免重复下载；请求文件和失败下载的临时文件应清理，交付文件按用户保留需求处理。

## 失败引导

始终保留精确模型、task_id、请求数量／非敏感参数、状态、失败阶段和下一步。仅展示 API 实际返回的公开错误码与脱敏错误；没有机器码就不编造。内容审核失败要说明实际原因并建议调整提示词，模型／分组权限不足引导用户检查连接与分组授权，余额不足引导充值或选择已明确支持的较低成本方案，不承诺退款或擅自重试。

连接解析、附件准备或参数校验失败发生在 POST 前时，明确没有提交。网络中断、5xx 或无法读取提交响应时，按“提交结果未知”处理。执行器不可用时说明真实错误并引导更新／初始化当前宿主，不能改走 UI 或其他传输。
