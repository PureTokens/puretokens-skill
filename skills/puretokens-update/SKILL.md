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

宿主 ID 为 claude-code、codex、workbuddy、gemini-cli、grok-build、opencode、trae、claude-desktop、dsh-desktop、zcode；识别不了先询问。自定义路径可同时传 --target／-Target 绝对目录，保留 host 供 init 使用。fetch 会解析官方 main 的精确提交和版本，并取得同一提交的目录选择器，不复用旧版同目录选择器；check-update 只报告版本差异，不修改安装或执行 init。安装／更新优先选择版本及提交匹配、SHA-256 校验通过的平台包；没有匹配已发布包时，取得相同提交的官方源码归档，再执行其中的 sync。下载或校验失败不重复安装、不改用镜像。已有明确官方检出也可运行它的 `runtime/puretokens-skill-install.sh/.ps1 sync`。

只由 sync 写入六个 Skill 和一个当前平台原生执行器及下载／同步脚本。它保护非受管目录；更新锁阻止并发，中断的受管事务在下次显式 sync 时恢复。Gemini 如已有较高优先级的共享 `.agents/skills`，更新其有效目录并报告重复的低优先级副本，不擅自删除。只删除确认受管的旧官方 Skill／Node runtime；未知文件和用户配置保持不动。

受管识别核对 `.puretokens-managed.json` 的完整文件清单和校验值；没有记录的旧安装只能与随候选执行器提供的历史清单或当前官方源文件完全匹配后迁移，不能仅凭同名或目录中的自报版本判定归属。存在新增文件、本地修改、缺失文件或符号链接时停止覆盖；向用户说明需保留并处理该冲突，不自动删除、重建清单或声称更新完成。中断恢复遇到改动同样保留原目录和备份。

Codex 只用官方 plugin 接口检查旧 puretokens-media；确实发现后才移除并复核。检查不可用要说明但不阻挡正常安装；已发现却无法移除时停止。移除后完全重启 Codex，再新开对话。

只在收到 `Pure Tokens Skills <版本> synchronized with the native API executor at <目录>` 后报告同步完成；随后 init 失败不回滚已完成的文件同步。超时／中断不能声称完成或自动重试。成功后提醒新开宿主对话，不重复列出六份相同版本信息。

## 初始化与诊断

从当前 SKILL.md 的绝对目录解析同级 `../.puretokens-executor/puretokens-api`；Windows 使用 `puretokens-api.exe`，不依赖 PATH。`init --host <host-id>` 先请求固定 `/v1` 公开身份，再请求一次 `/v1/media/models` 验证认证；最多两次只读请求，无付费任务。`doctor --host <host-id>` 额外检查本次加载位置和当前宿主已知 Skill 目录中的安装版本／重复项，再执行同样的只读检查；不扫描 Home、不读取其他宿主配置，也不自动修复。

只有 init 的 configuration_status 为 verified 且 api_identity_confirmed、credential_verified 均为 true，才说身份与凭据认证通过。doctor 不能证明安装二进制校验、任意会话配置覆盖或宿主附件交付成功；按返回字段说明已检查项目与未验证项，不承诺模型权限、余额或媒体可用性。

Skill 和安装脚本不读取、显示或修改认证文件。只有原生执行器按当前宿主明确记录在内存中读取匹配凭据；不使用其他认证、MCP、代理或 Computer Use。

Claude Desktop 使用 `claude-desktop`，仅在能访问本机执行器和 Desktop 当前连接的本地会话执行；不能用 Claude Code 的连接代替。DSH Desktop 使用 `dsh-desktop`，采用当前本地 Harness 的连接。云端或隔离会话不能访问宿主记录时按实际失败停止，不复制凭据进入沙箱。宿主安装和交付说明按需读 `references/desktop-hosts.md`。

ZCode 本地执行使用宿主 ID `zcode`。只使用唯一启用且端点匹配的 Pure Tokens 连接；不能据此声称识别当前聊天模型。目录、远程工作区及交付限制按需读 `references/desktop-hosts.md`。

失败时区分版本检查、下载／校验、文件同步和 init 阶段，说明哪些步骤已确认完成及下一步。网络失败可建议稍后重新检查；校验失败停止使用该包并反馈官方维护方；写入受限可建议检查目标目录权限，不提权或删除未知目录。sync 已成功但 init 失败时明确“文件已同步，连接验证未完成”，按 init 脱敏建议处理，不重复安装。未知中断先核对已安装状态，不自动重跑。

WorkBuddy 适配器无法匹配连接时，不据此断言当前聊天未使用 Pure Tokens；若用户确认已配置并成功使用 Pure Tokens，保留可用连接，按具体诊断检查适配；不能把自带模型聊天正常当作 Pure Tokens 配置有效的证据。排查仅提供客户端版本及脱敏诊断码，不读取或索取配置原文。Windows 的 `cleanup_status: pending` 表示已完成事务的暂存清理被拒绝，不等于文件同步失败；按独立 init 结果说明连接状态，不绕过删除守卫。锁文件存在不等于被占用；缺少管理清单不能归因于清理失败，也不能称为无害。

`init` 验证已保存的 Pure Tokens 连接，与当前聊天模型无关，不要求切换聊天模型。WorkBuddy 使用 `workbuddy_record_missing`、`workbuddy_record_unreadable`、`workbuddy_record_format_unsupported`、`workbuddy_connection_not_found`、`workbuddy_credential_missing`、`workbuddy_connection_ambiguous` 区分本地原因；这些状态不代表安装失败，也不证明用户从未配置过。只反馈脱敏状态，不读取配置原文。

Windows 安装只运行一次官方 fetch／sync 流程并汇总同步、清理和 init 结果；不要自行拆成逐目录命令或重复执行已通过的校验。安装器捕获原生执行器输出并禁止创建额外控制台窗口；宿主自身的命令窗口仍由宿主控制。执行被策略拒绝时停止并说明，不改用备用启动方式或提权。

安装执行边界：确认宿主后，将官方 fetch 下载为本地文件并执行一次，由脚本完成提交固定、下载、校验、同步和 init。不要在正常安装中额外审计整份脚本、逐项探测 PowerShell 能力、重复核对清单或再跑 doctor。必要的宿主审批照常遵守。遇到启动被拒绝、输出不可用、超时、下载或校验失败，报告已完成阶段与脱敏失败并停止；不得创建 probe、shim、Python 补丁、修改官方脚本、换启动路径或自动重试。只有用户另行要求开发排查时才进入调试流程。
