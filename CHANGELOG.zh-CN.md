<p align="center">
  <a href="./CHANGELOG.md">English</a> · <strong>中文</strong>
</p>

# 更新日志

## 0.3.2 — 2026-08-14

- 强化 WorkBuddy 媒体路由：明确 `ToolSearch` 只负责发现工具，所有 Pure Tokens 的目录、生成和结果调用都必须通过 `DeferExecuteTool` 实际执行。
- 禁止用 SVG/HTML 组件、内置媒体工具、搜索或文字声明冒充 Pure Tokens 媒体任务成功；只有 MCP 返回实际模型及原生结果或本机交付信息时才能称为生成成功。

## 0.3.1 — 2026-08-14

- 将 `puretokens_media` 收敛为 Claude Desktop 与 WorkBuddy 共用的唯一媒体行为源。WorkBuddy 的常驻路由载荷现在由该源自动生成，目录、精确模型、单次提交、轮询、交付和失败规则完全一致。
- 移除独立的 WorkBuddy 路由 Skill 源。Pure Tokens Desktop 仍会在点击 **验证并应用** 后自动安装生成后的适配载荷，用户无需手动上传。
- 移除媒体 Skill 中已过期的余额和模型价格工具声明；现在只暴露本地 Sidecar 实际提供的五个媒体 MCP 工具。

## 0.3.0 — 2026-08-14

- 新增 `puretokens_workbuddy_router`：轻量常驻的 WorkBuddy Skill。普通生图、生视频请求会优先进入已配置的 Pure Tokens MCP，而不是 WorkBuddy 内置媒体工具；用户明确指定 WorkBuddy 内置工具时保持该选择。文本、代码等普通请求不受影响，并保持目录优先、精确模型与不静默回退的约束。
- WorkBuddy 路由 Skill 由 Pure Tokens Desktop 作为受管集成自动安装和升级，用户无需上传或手动启用。

## 0.2.7 — 2026-08-14

- 媒体完成结果现在会展示 MCP 实际返回的精确模型 ID、文件名和持久 `Downloads/Pure Tokens` 交付状态。
- 为支持渲染的宿主增加有大小上限的原生 MCP 视频资源；较大视频仍成功保存为本机 MP4，不会把超大内容塞进 stdio 响应。
- MCP 重启后的结果轮询会携带原始精确模型，不再猜测默认路由。
- 明确指定媒体模型前，必须为目标客户端选择包含该模型的分组，再点击“验证并应用”、重启客户端并新建会话。

## 0.2.6 — 2026-08-13

- 图片和视频完成后统一以 MCP 已落盘的本机 `Downloads/Pure Tokens` 文件作为持久交付，不再依赖临时预览链接。
- 图片只有在 MCP 实际返回原生图片内容时，Skill 才能说可在对话内预览；不再伪造“图片已在上方显示”。

## 0.2.4 — 2026-08-12

- 新增 `puretokens_get_balance`，通过 Desktop 发布的只读余额快照查询余额，并提供中英文使用说明。余额查询不会读取媒体模型或任何凭据材料。
- 将 `README.md` 作为默认英文入口，在顶部增加 English/中文切换，并新增中文 README。
- 使用 Image-2 和官方 Pure Tokens 图标生成 Pure Tokens Skill 品牌头图。
- 在中英文文档中增加模型能力表、使用示例和可复制的 Agent 安装提示词。
- 面向首次使用的用户重写 README，分别列出图片/视频模型，并补充直白的使用示例。
- 新增由 Skill 管理的自然语言别名注册表，包括 `image2` → `gpt-image-2`，并在请求前通过实时目录确认，不静默回退。
- 增加固定默认模型：未指定模型的图片请求使用 `gpt-image-2`，未指定模型的视频请求使用 `grok-imagine-video-1.5`。
- 增加 Nano Banana 系列别名：`Nano Banana Pro` 对应 `gemini-3-pro-image-preview`；`Nano Banana 2` 对应 `gemini-3.1-flash-image-preview`；只说 Nano Banana 且两个模型都可用时会先询问用户。
- 新增精确模型价格查询工具 `puretokens_get_model_price`。
- 展示每个选中分组对应的真实价格，不推测、不替换模型。

## 0.2.0 — 2026-08-09

- 让媒体 Skill 强制先读目录、精确匹配能力、使用稳定请求 ID、单次提交、同任务轮询，并在错误或超时后禁止自动回退。
- 增加 Claude Desktop ZIP 分发说明和便携打包命令。
- 增加安全升级和显式确认卸载命令。
- 增加模型歧义、目录为空、MCP 不可用、任务失败和轮询超时等行为场景。

## 0.1.0 — 2026-08-09

- 创建 Pure Tokens Skill 管理仓库。
- 新增 `puretokens_media`，用于选择并调用 Pure Tokens 图片和视频 MCP 工具。
- 新增版本化 Skill 注册表、机器可读清单、媒体模型目录 schema、安全本地安装器和仓库校验。
