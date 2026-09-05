---
name: puretokens-models
description: 查询当前 Pure Tokens 连接实际可用的图片和视频模型、参数或媒体操作时使用。
---

# Pure Tokens Models

通过安装的单次 Go 执行器查询模型目录。从当前 SKILL.md 的绝对目录解析同级 `../.puretokens-executor/puretokens-api`；Windows 使用 `puretokens-api.exe`，不要依赖工作目录或 PATH。调用 `<绝对执行器路径> models --host <当前宿主 ID>`。仅执行器私密读取文档列出的当前宿主明确连接记录；凭据格式夹具通过不代表项目／会话覆盖或真实宿主端到端验收。Skill 不读取配置、不传 Key／Base URL，不自行 HTTP，也不使用其他传输。

不得申请、调用或使用 Computer Use，也不得打开、点击或控制浏览器、Pure Tokens Switch、Pure Tokens Desktop 或其他图形界面来发现模型、读取配置或替代执行器；不得调用其他 Skill 作为回退。

只处理执行器返回的结构化结果；不得把响应原文、请求头或配置内容直接展示给用户。仅当执行器返回实际本地或 API 失败时，按安全回执处理，明确目录请求是否执行且不猜测配置原因。

详细字段按需读 `references/execution-contract.json`；只有对应异常时才读 `references/behavior-scenarios.json`。

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

用户明确筛选时，可用 `models --host <host-id> --request <UTF-8筛选文件>`，文件可含 `kind`、精确 `model`、`operation` 和 `parameters`，例如 `{"kind":"video","operation":"image_to_video","parameters":{"resolution":"720p"}}`。执行器只读取一次认证目录并按其声明筛选；缺少字段不视为兼容，不提交媒体。无筛选则省略 --request。查询目录不是普通生成的必需前置步骤。
