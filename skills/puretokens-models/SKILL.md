---
name: puretokens-models
description: 查询当前 Pure Tokens 连接实际可用的图片和视频模型、参数或媒体操作时使用。
---

# Pure Tokens Models

只读查询固定 Pure Tokens API `https://api.puretokensx.com/v1/media/models` 的当前认证媒体目录。使用当前宿主已有的认证 HTTPS/API 执行能力直接发出该完整固定 URL 请求；只可在内存中读取当前活动连接中与固定 Pure Tokens origin 精确匹配的一把凭据，作为本次请求的认证头。第三方 CC Switch 或手工配置只要已写入宿主连接即可使用；不得检查 provider 标签，也不得把保存的 Base URL 当作请求目标。不得显示、复制、保存、索取或输出凭据、Base URL 或完整宿主配置；不得安装或调用 Node 运行器、MCP、本地代理、sidecar、备用服务或另一条认证路径。

只处理完成请求后得到的结构化结果；不得把响应原文、请求头或配置内容直接展示给用户。若当前宿主不能在不暴露凭据的前提下执行该已认证 HTTPS 请求，按本地执行不可用处理，明确目录请求未执行且不猜测配置原因。

先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们定义模型目录、能力筛选和用户可见输出的约束。

## 查询规则

- 每个用户查询只调用一次 `GET https://api.puretokensx.com/v1/media/models`；不得调用 Images/Videos 提交、任务状态、内容、余额或其他路径，也不得重试。
- 当前认证目录是唯一事实来源。只报告实际返回的精确模型 ID、明确返回的 `capabilities`、`input_schema.properties` 参数资料、`input_schema.constraints` 条件限制和 `input_schema.operations`；不得以 README、安装包静态清单、模型名称或过往任务补全、猜测或回退。
- 用户可查询：当前可用图片/视频模型；某个精确模型；支持某个能力或操作的模型；以及某项时长、画幅、分辨率、尺寸、数量或参考媒体要求有哪些兼容模型。
- 需求匹配只基于当前条目明确声明的 capability、`input_schema.properties` 字段和值、适用的 `input_schema.constraints`、以及 `input_schema.operations`。例如图生视频只匹配发布 `image_to_video` 的模型；参考图、参考视频、参考音频和视频编辑分别只匹配 `reference_image_video`、`reference_video`、`reference_audio`、`video_edit`。没有明确声明就不列为兼容。
- 不对模型质量、速度、价格、用量、排队时间、内容效果或未返回的可用性做推荐或排序。多个模型都满足已声明条件时，完整列出并说明它们在目录声明的差异；用户的需求无法映射到明确 capability、参数或 operation 时，请其选择明确约束或显示完整目录。

## 用户可见输出

- 按用户的筛选条件展示；未筛选时先按图片和视频 capability 分组。
- 每个模型仅展示实际返回的：精确 ID、capability、可选参数名称、`required` 标记、类型、默认值、`enum` 值、数值范围、非请求字段的条件限制（例如 `resolution_by_mode`），以及 operation 名称、请求方法、相对路径、content type、必需字段、附件数量和 transport。`constraints` 不是额外请求字段，绝不把它的名称或推断模式写入 API body。缺失字段写“目录未声明”，不得猜测。
- 若用户询问“哪个能做 X”，先给出满足 X 的兼容模型和匹配依据；不要直接提交生成请求。用户选定模型并提出生成需求后，再交由 `puretokens-image` 或 `puretokens-video` 按各自契约执行。
- 目录为空、无法读取或模型不在当前目录时，如实说明固定 Pure Tokens API 未返回对应模型资料或返回了该错误；不切换 endpoint、不尝试静态目录，也不索取凭据或猜测配置原因。若用户预期可用某精确模型，指导其在 Pure Tokens 客户端配置中勾选包含该模型的分组，创建或选择覆盖所选分组的受管 Key，执行“验证并应用”，然后新开当前宿主会话再查询或提交。除非认证 API 明确返回模型到分组的映射，否则不得说出或猜测具体分组名称。
