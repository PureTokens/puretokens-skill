<p align="center">
  <a href="./CHANGELOG.md">English</a> · <strong>中文</strong>
</p>

# 更新日志

## 未发布 — 2026-08-12

- 新增 `puretokens_get_balance`，通过 Desktop 发布的只读余额快照查询余额，并提供中英文使用说明。余额查询不会读取媒体模型或任何凭据材料。
- 将 `README.md` 作为默认英文入口，在顶部增加 English/中文切换，并新增中文 README。
- 使用 Image-2 和官方 Pure Tokens 图标生成 Pure Tokens Skill 品牌头图。
- 在中英文文档中增加模型能力表、使用示例和可复制的 Agent 安装提示词。
- 面向首次使用的用户重写 README，分别列出图片/视频模型，并补充直白的使用示例。
- 新增由 Skill 管理的自然语言别名注册表，包括 `image2` → `gpt-image-2`，并在请求前通过实时目录确认，不静默回退。
- 增加固定默认模型：未指定模型的图片请求使用 `gpt-image-2`，未指定模型的视频请求使用 `grok-imagine-video-1.5`。
- 增加 Nano Banana 系列别名：`Nano Banana Pro` 对应 `gemini-3-pro-image-preview`；`Nano Banana 2` 对应 `gemini-3.1-flash-image-preview`；只说 Nano Banana 且两个模型都可用时会先询问用户。

## 0.2.4 — 2026-08-12

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
