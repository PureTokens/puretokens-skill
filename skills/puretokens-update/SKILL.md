---
name: puretokens-update
description: 用户要求检查、安装或升级本机已安装的 Pure Tokens Skills 时使用。
---

# Pure Tokens Update

用于用户主动要求安装、升级、初始化或了解用法。普通媒体请求不自动更新。安装提示来源是官方仓库 `https://github.com/PureTokens/puretokens-skill`。读取 `references/usage-guide.md`，异常时按需读 `references/behavior-scenarios.json`。

## 安装和更新

取得官方 main 的最新本地源码（可用宿主现成 GitHub 下载能力，无需给用户安装 Git、Node、Python 或 Go），从源码执行：

```text
macOS/Linux: sh runtime/puretokens-skill-install.sh sync --host <当前宿主 ID>
Windows: powershell -NoProfile -File runtime/puretokens-skill-install.ps1 sync -Host <当前宿主 ID>
```

宿主 ID 为 claude-code、codex、workbuddy、gemini-cli、grok-build、opencode、trae；识别不了先询问。自定义安装目录可同时传 --target/-Target 的绝对路径，保留 host 供 init 使用。不要以当前工作目录猜宿主。

安装器统一同步六个 Skill、当前平台校验过的执行器，保护非受管同名目录。更新锁阻止并发写入；失败恢复已替换文件，中断留下的受管事务在下次 sync 时恢复。只删除确认受管的旧 Skill／Node runtime，不能清理未知文件或用户配置。

Codex CLI 可用时用官方 plugin list 检查旧 puretokens-media：确实发现后才移除并复核。检查接口不可用不阻挡正常安装，但说明插件检查未完成；若用户仍看到旧 Media 指令，引导在 Codex Plugins 中卸载旧插件后完全重启。已发现插件却移除失败时停止，不称迁移成功。

只在收到 `Pure Tokens Skills <版本> synchronized with the native API executor at <目录>` 后报告安装版本。然后展示 init 结果；初始化失败不等于文件同步失败。超时／中断不能声称完成或自动重试。更新后新开对话；移除旧插件后完全退出并重启 Codex。

## 初始化

按 SKILL.md 位置解析同级 `../.puretokens-executor/puretokens-api init --host <当前宿主 ID>`，Windows 为 `.exe`，不依赖 PATH。执行器先请求固定 `https://api.puretokensx.com/v1` 检查公开身份，再请求一次 `/v1/media/models` 验证当前凭据认证。总共最多两次只读请求，不创建付费任务。

只有 `configuration_status: "verified"`、`api_identity_confirmed: true`、`credential_verified: true` 同时成立时，才说“API 可达且当前凭据认证通过”；不能保证所有模型权限、余额、媒体创建或宿主附件交付都可用。其他情况展示脱敏状态、说明和下一步，不猜配置错误；仍展示版本及 `references/usage-guide.md` 使用须知。

Skill 和安装脚本不读取或修改用户认证文件；仅执行器按当前宿主的明确凭据来源在内存中读取匹配 Key。不扫描 Home、不打印配置、不索取 Key、不安装运行时、不运行开发生成器，不使用 MCP、代理或 Computer Use。
