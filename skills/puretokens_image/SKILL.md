---
name: puretokens_image
description: 通过当前 Pure Tokens 连接生成图片、选择图片模型、数量或尺寸时使用。
---

# Pure Tokens Image

通过宿主当前已配置的连接调用 Pure Tokens Images API。宿主负责 Base URL、认证、路由和原生字节交付；Skill 只请求相对路径和 JSON body，绝不读取、扫描、展示或索取凭据或配置，也不检查 Base URL、provider 标签或服务归属。

先读取已安装的 `references/model-selection.json`、`references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们是模型别名、请求/结果约束和特殊场景的绑定规则。

## 模型与提交

- 未指定模型时使用 `gpt-image-2`；明确 `image2` 也使用该模型，单次 `POST /v1/images/generations`，传 `model: "gpt-image-2"` 和 `async: true`。
- 用户给出精确 ID 时，保留该 ID。用户给出别名时，仅在 `model-selection.json` 中唯一匹配时解析为精确 ID；零个或多个候选时，列出候选精确 ID 并请用户选择，绝不猜测。
- 非默认模型先 `GET /v1/media/models`。只有当前认证目录存在同一精确 ID 且明确含 `image` capability 才提交；不存在、无权限或 capability 不匹配时，说明原因、列出当前目录候选，不提交也不换模型。
- 默认数量为 `1`；只在用户明确要求时传 `n`，且只能是 `1` 到 `6` 的整数。一个用户请求只能有一次付费提交，绝不拆单。
- `size` 仅支持 `1024x1024`、`1536x1024`、`1024x1536`；`image_size` 仅支持 `1K`、`2K`、`4K`。不支持的像素规格要列出这些可选项并请用户重选。`200cm × 230cm` 等物理尺寸不能传给 `n` 或 `size`：说明不能精确保证物理尺寸，停止并请用户选择支持规格。
- 当前只支持文生图。编辑、蒙版、参考图、上传图片或其他用户媒体输入时，明确说明限制并请用户改用文字描述；不得上传、转发或伪装编辑能力。

## 任务与交付

- 始终传 `async: true`。有 `task_id` 时，只读同一任务的 `/v1/images/{task_id}` 与 `/v1/images/{task_id}/content?index=N`；不得以默认模型猜测、不得查询或提交另一任务。
- pending/running 时报告当前状态并继续同一任务；失败、取消或过期时报告返回错误，要求用户明确发起新请求后才可再次提交。
- 提交结果未知、状态超时或内容交付失败时，说明不确定/失败状态；只可让用户选择继续查询或重试同一任务内容，绝不自动重提、换模型、换参数或拆分付费请求。
- 只有宿主交付原生图片字节才算成功。任务 ID、状态文字、HTML/SVG、组件或远程 URL 都不算成功。
- 当前连接不能执行 API 或交付原生字节时，在付费提交前停止，引导用户检查已有 Pure Tokens Base URL、认证与路由配置；不得索取凭据或改走备用路径。
