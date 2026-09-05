---
name: puretokens-update
description: 用户要求检查版本、安装、升级、诊断或了解 Pure Tokens Skills 用法时使用。
---

# Pure Tokens Update

只响应用户主动的安装、升级、检查或帮助请求。普通媒体生成不自动更新、初始化或诊断。仅问用法时直接读 `references/usage-guide.md` 回答，不运行 init／doctor，不查凭据或访问网络。异常时才读 `references/behavior-scenarios.json`。

## 检查、安装和更新

已安装时，从本 SKILL.md 的绝对目录解析同级 `.puretokens-executor/puretokens-skill-fetch.sh`（Windows 为 `.ps1`）。首次安装由宿主现成 HTTPS 下载能力把官方仓库 `runtime/puretokens-skill-fetch.sh`／`.ps1` 下载为本地文件，再运行；不把远程响应管道进 shell，不安装 Git、Node、Python 或 Go。

```text
macOS/Linux: sh <fetch脚本绝对路径> check-update --host <当前宿主ID>
macOS/Linux: sh <fetch脚本绝对路径> install --host <当前宿主ID>
macOS/Linux: sh <fetch脚本绝对路径> update --host <当前宿主ID>
Windows: powershell.exe -NoProfile -ExecutionPolicy Bypass -File <fetch脚本绝对路径> check-update -Host <当前宿主ID>
```

Windows 的 install／update 使用相同参数形式替换子命令。ExecutionPolicy 只作用于该进程，不修改系统策略；宿主明确阻止执行时停止，不绕过审批。

宿主 ID 为 claude-code、codex、workbuddy、gemini-cli、grok-build、opencode、trae；识别不了先询问。自定义路径可同时传 --target／-Target 绝对目录，保留 host 供 init 使用。fetch 会解析官方 main 的精确提交和版本；check-update 只报告版本差异，不修改安装或执行 init。安装／更新优先选择版本及提交匹配、SHA-256 校验通过的平台包；没有匹配已发布包时，取得相同提交的官方源码归档，再执行其中的 sync。下载或校验失败不重复安装、不改用镜像。已有明确官方检出也可运行它的 `runtime/puretokens-skill-install.sh/.ps1 sync`。

只由 sync 写入六个 Skill 和一个当前平台原生执行器及下载／同步脚本。它保护非受管目录；更新锁阻止并发，中断的受管事务在下次显式 sync 时恢复。Gemini 如已有较高优先级的共享 `.agents/skills`，更新其有效目录并报告重复的低优先级副本，不擅自删除。只删除确认受管的旧官方 Skill／Node runtime；未知文件和用户配置保持不动。

Codex 只用官方 plugin 接口检查旧 puretokens-media；确实发现后才移除并复核。检查不可用要说明但不阻挡正常安装；已发现却无法移除时停止。移除后完全重启 Codex，再新开对话。

只在收到 `Pure Tokens Skills <版本> synchronized with the native API executor at <目录>` 后报告同步完成；随后 init 失败不回滚已完成的文件同步。超时／中断不能声称完成或自动重试。成功后提醒新开宿主对话，不重复列出六份相同版本信息。

## 初始化与诊断

从当前 SKILL.md 的绝对目录解析同级 `../.puretokens-executor/puretokens-api`；Windows 使用 `puretokens-api.exe`，不依赖 PATH。`init --host <host-id>` 先请求固定 `/v1` 公开身份，再请求一次 `/v1/media/models` 验证认证；最多两次只读请求，无付费任务。`doctor --host <host-id>` 额外检查本次加载位置和当前宿主已知 Skill 目录中的安装版本／重复项，再执行同样的只读检查；不扫描 Home、不读取其他宿主配置，也不自动修复。

只有 init 的 configuration_status 为 verified 且 api_identity_confirmed、credential_verified 均为 true，才说身份与凭据认证通过。doctor 不能证明安装二进制校验、任意会话配置覆盖或宿主附件交付成功；按返回字段说明已检查项目与未验证项，不承诺模型权限、余额或媒体可用性。

Skill 和安装脚本不读取、显示或修改认证文件。只有原生执行器按当前宿主明确记录在内存中读取匹配凭据；不使用其他认证、MCP、代理或 Computer Use。
