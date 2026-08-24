---
name: puretokens_video
description: 通过当前 Pure Tokens 连接生成视频、选择视频模型或参数时使用。
---

# Pure Tokens Video

通过宿主当前已配置的连接调用 Pure Tokens Videos API。宿主负责 Base URL、认证、路由和原生字节交付；Skill 只请求相对路径和 JSON body，绝不读取、扫描、展示或索取凭据或配置，也不检查 Base URL、provider 标签或服务归属。

先读取已安装的 `references/model-selection.json`、`references/execution-contract.json`、`references/task-receipt.json` 和 `references/behavior-scenarios.json`；它们是模型别名、请求/结果约束、用户回执和特殊场景的绑定规则。

## 模型与提交

- 先 `GET /v1/media/models`。未指定模型时使用 `grok-imagine-video-1.5-preview`；它也必须在当前认证目录中具有 `video` capability。
- 用户给出精确 ID 时，只有当前目录存在同一 ID 且含 `video` capability 才可使用。用户给出别名时，仅在 `model-selection.json` 中唯一匹配时解析；零个或多个候选时，列出候选精确 ID 并请用户选择。例如 `grok video` 不可自动选择两个候选模型。
- 模型不存在、无权限或 capability 不匹配时，说明原因、列出当前目录候选，不提交也不换模型。
- 单次 `POST /v1/videos`。仅有文字提示词时可以提交；用户明确给出的时长、画幅、分辨率、尺寸或其他可选参数，必须先由当前认证目录中该精确模型的 `input_schema` 明确支持。资料缺失、字段不存在或值超出资料声明时，在提交前说明原因、列出已声明值（若有），请用户删除该参数或改选模型；不得根据模型名猜测或静默改写。
- 当前只支持文生视频。图生视频、参考图/参考视频、视频编辑、音频输入、上传媒体或其他用户媒体输入时，明确说明限制并请用户改用文字描述；不得上传、转发或伪装支持。

## 任务与交付

- 视频总是异步任务：只轮询同一 `/v1/videos/{task_id}`，完成后只读取同一 `/v1/videos/{task_id}/content` 并交付原生视频字节；不得查询或提交另一任务。
- pending/running 时报告当前状态并继续同一任务；失败、取消或过期时报告返回错误，要求用户明确发起新请求后才可再次提交。
- 提交结果未知、状态超时或内容交付失败时，说明不确定/失败状态；只可让用户选择继续查询或重试同一任务内容，绝不自动重提、换模型或换参数。
- 只有宿主交付原生视频字节才算成功。任务 ID、状态文字、HTML/SVG、组件或远程 URL 都不算成功。
- 当前连接不能执行 API 或交付原生字节时，在付费提交前停止，引导用户检查已有 Pure Tokens Base URL、认证与路由配置；不得索取凭据或改走备用路径。
- 每次提交、继续查询、完成或失败，都按 `task-receipt.json` 报告：已返回的精确模型 ID、任务 ID、返回状态、请求数量（视频为 1）、尺寸/参数、已交付数量（完成时）和下一步。任务响应未返回模型或任务 ID 时明确写“未返回”，绝不猜测。
