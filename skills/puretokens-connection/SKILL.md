---
name: puretokens-connection
description: 用户询问 Skill 将调用的 Pure Tokens API 是否可确认、API 身份或连接检查时使用。
---

# Pure Tokens Connection

通过安装的单次 Go 执行器查询固定 API 公开身份。从当前 SKILL.md 的绝对目录解析同级 `../.puretokens-executor/puretokens-api`；Windows 使用 `puretokens-api.exe`，不要依赖工作目录或 PATH。调用 `<绝对执行器路径> connection --host <当前宿主 ID>`。仅执行器私密读取文档列出的当前宿主明确连接记录；凭据格式夹具通过不代表项目／会话覆盖或真实宿主端到端验收。Skill 不读取配置、不传 Key／Base URL，不自行 HTTP，也不使用其他传输。

不得申请、调用或使用 Computer Use，也不得打开、点击或控制浏览器、Pure Tokens Switch、Pure Tokens Desktop 或其他图形界面来读取配置、检查连接或替代执行器；不得调用其他 Skill 作为回退。

只处理执行器返回的结构化结果；不得把响应原文、请求头或配置内容直接展示给用户。仅当执行器返回实际本地或 API 失败时，按安全回执处理，明确固定 API 是否执行且不猜测配置原因。

本页足以执行普通身份检查；只有对应异常时才读 `references/behavior-scenarios.json`，详细约束见 `references/execution-contract.json`。

## 身份检查

- 每次用户查询只调用一次 `GET https://api.puretokensx.com/v1`；不得调用模型目录、Images、Videos、余额、任务或其他路径，也不得重试。
- 只有响应 JSON 同时明确返回 `status: "ok"`、`name: "Pure Tokens API"`、`base_url: "/v1"` 时，才回答：`固定 API 返回了 Pure Tokens API 的公开身份标识；此检查未验证凭据、余额、模型权限或媒体可用性。`
- 响应不可读、字段缺失、值不匹配或请求失败时，只能回答：`固定 Pure Tokens API 未返回可确认的标识，当前无法完成检查。` 不得把该结果归因于用户 Base URL、API Key 或 provider 配置，也不得读取或要求用户粘贴它们。
- 该检查验证的是固定 API 端点的既有公开声明，不是密码学防伪证明；未确认不等于证明用户的其他连接属于某个服务。不得根据模型列表、模型名称、请求是否成功、Base URL 字符串或 provider 标签改写这一结论。
- 用户问当前真实 Base URL、API Key 或宿主 provider 时，明确说明：只有原生执行器可为固定请求在内存中使用文档列出的匹配连接凭据，但不展示、比较或报告这些配置；引导用户在 CC Switch、Pure Tokens Desktop 或宿主自身的连接设置中查看。

ZCode 本地执行使用宿主 ID `zcode`。只使用唯一启用且端点匹配的 Pure Tokens 连接；不能据此声称识别当前聊天模型。目录、远程工作区及交付限制按需读 `references/desktop-hosts.md`。

失败结论后可补充下一步：网络／服务异常可由用户稍后重查或联系 Pure Tokens 支持；标识不符可反馈身份检查结果。公开身份检查的 401／403 不验证凭据，不能据此要求更换 Key。执行器无法运行则说明检查尚未执行，并按回执处理本地可访问性。

WorkBuddy 适配器无法匹配连接时，不据此断言当前聊天未使用 Pure Tokens；若用户确认已配置并成功使用 Pure Tokens，保留可用连接，按具体诊断检查适配；不能把自带模型聊天正常当作 Pure Tokens 配置有效的证据。排查仅提供客户端版本及脱敏诊断码，不读取或索取配置原文。Windows 的 `cleanup_status: pending` 表示已完成事务的暂存清理被拒绝，不等于文件同步失败；按独立 init 结果说明连接状态，不绕过删除守卫。锁文件存在不等于被占用；缺少管理清单不能归因于清理失败，也不能称为无害。

`init` 验证已保存的 Pure Tokens 连接，与当前聊天模型无关，不要求切换聊天模型。WorkBuddy 使用 `workbuddy_record_missing`、`workbuddy_record_unreadable`、`workbuddy_record_format_unsupported`、`workbuddy_connection_not_found`、`workbuddy_credential_missing`、`workbuddy_connection_ambiguous` 区分本地原因；这些状态不代表安装失败，也不证明用户从未配置过。只反馈脱敏状态，不读取配置原文。
