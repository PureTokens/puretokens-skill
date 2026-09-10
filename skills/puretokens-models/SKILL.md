---
name: puretokens-models
description: 查询当前 Pure Tokens 连接实际可用的图片和视频模型、参数或媒体操作时使用。
---

# Pure Tokens Models

宿主绑定与本地失败停止：首次调用前根据当前应用或用户明确指定确定一个 host；本次任务所有 Skill 和执行器命令沿用该 host，不从安装目录推断宿主。任何命令返回 `active_connection_*`、`workbuddy_*` 或 `host_credential_adapter_unavailable` 本地失败时，立即停止本次 API 流程并返回脱敏说明；不得自动调用 puretokens-connection、init、doctor、models 或其他 Skill 重复探测，不得枚举、更换 `--host` 或借用其他客户端凭据。保留用户原任务和已有 task ID；本次调用未发请求不代表此前没有提交任务，不自动重提。只有用户明确要求重查或出现可验证的新诊断依据时，才在原 host 做一次相应检查；“继续生成”本身不是切换宿主或重复探测的授权。改变执行宿主必须有用户明确指定，不能把失败恢复当作指定。

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
- 目录为空或精确模型缺失，只说明当前目录没有返回对应资料，不能断言 Key 分组错误。可展示实际返回的模型；用户预期有权限时，建议在自身设置中核对该模型权限或联系 Pure Tokens 支持。只有确认需要调整权限时才由用户选择分组、验证并应用；没有认证 API 明确返回模型到分组的映射，不得猜分组名称。网络／服务失败建议稍后再查询；认证、权限、限流按回执 `next_action` 分别解释。不切换 endpoint、不以静态目录冒充实时结果、不自动重试或提交媒体。

用户明确筛选时，可用 `models --host <host-id> --request <UTF-8筛选文件>`，文件可含 `kind`、精确 `model`、`operation` 和 `parameters`，例如 `{"kind":"video","operation":"image_to_video","parameters":{"resolution":"720p"}}`。执行器只读取一次认证目录并按其声明筛选；缺少字段不视为兼容，不提交媒体。无筛选则省略 --request。查询目录不是普通生成的必需前置步骤。

当前会话若使用未反映在宿主已声明有效文件中的配置覆盖，停止并说明无法确认有效连接；不读取其他配置或借用默认连接。

宿主目录、会话选择限制或特定客户端故障才按需读 `references/desktop-hosts.md`；普通请求不增加诊断前置步骤。

本地连接失败按执行器脱敏状态解释：未识别、不可读、不支持或未验证均不等于“尚未配置”。保留现有连接，不据此要求重装、切换模型、重配或更换凭据；`api_request_executed: false` 时说明未请求 API、未认证凭据。具体宿主限制见 `references/desktop-hosts.md`。
