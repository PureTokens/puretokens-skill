---
name: puretokens-update
description: 用户要求检查、安装或升级本机已安装的 Pure Tokens Skills 时使用。
---

# Pure Tokens Update

升级的是本机已安装的官方 Pure Tokens Skills，不涉及账户、模型、余额或媒体生成。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们定义升级的安全边界和用户可见结果。

## 升级规则

- 只在用户明确要求安装、更新、升级、同步或检查 Pure Tokens Skills 时使用。不得在普通图片、视频、模型、余额或连接请求中自行运行升级。
- 先根据当前运行环境识别宿主：Claude Code 使用 `~/.claude/skills`；Codex 使用 `~/.agents/skills`（Windows PowerShell 为 `$env:USERPROFILE\.agents\skills`）；WorkBuddy 使用 `~/.workbuddy/skills`；Gemini CLI 使用 `~/.gemini/skills`；Grok Build 使用 `~/.grok/skills`；OpenCode 使用 `~/.config/opencode/skills`；Trae 使用 `~/.trae/skills`。无法识别时，先问用户使用哪个受支持宿主，不得猜测或修改目录。
- 若当前回合没有可执行终端、PowerShell、exec 或 shell 工具，说明无法代用户完成本机升级，并提供与该宿主匹配的官方可复制命令；不得声称已升级。
- 升级不能要求用户安装 Node、npm、包管理器、Git 或开发环境。macOS/Linux 使用已安装运行时中的 `sh <installation-root>/.puretokens-runtime/puretokens-skill-install.sh sync --target <installation-root>`；Windows 使用 `& "<installation-root>\.puretokens-runtime\puretokens-skill-install.ps1" sync -Target "<installation-root>"`。若当前旧版安装中还没有对应平台安装器，必须先从 `https://raw.githubusercontent.com/PureTokens/puretokens-skill/main/runtime/` 下载该平台脚本到新的私有临时目录，再执行它；不得把远程内容直接管道给 Shell/PowerShell，也不得退回 Node 流程。这两个官方平台安装器只使用操作系统自带的 HTTPS 下载、归档和文件操作能力；它们下载官方 `main` 源码，先完成只读静态载荷校验，再在同一次运行中同步。不得先运行 `node`、`npm`、Git、生成器或安装依赖。
- 平台安装器会安装缺失的官方 Skill，只升级含匹配 `SKILL.md` 与 `skill.json` 的已受管同名 Skill，并将已验证的旧受管 Skill 移到目标根目录内可恢复的隐藏备份。只要任一当前或旧同名目录，或受管运行器目录，不是受管内容，必须在写入前停止并报告冲突，绝不覆盖、删除或改名该目录。
- 升级完成后，逐项报告安装或升级的 Skill、未改动的冲突和失败原因（如有），并提醒用户新开宿主会话后再测试。

## 安全边界

- 绝不读取、扫描、展示、复制、修改或索取 API Key、Base URL、认证文件、AI 客户端配置、模型配置、MCP 配置、环境变量、Shell 配置、系统代理或账户数据。
- 不调用 Images、Videos、Models、余额或连接 API；不使用 MCP、本地代理、sidecar 或任何媒体生成路径。
- 不使用第三方包镜像；不执行卸载；不删除用户文件。只允许 `sync` 对目标根目录内已验证的官方 Skill 目录做安装或原子升级。
- 官方安装载荷只允许用于读取、静态校验和同步；不得运行 `node`、`npm`、Git 写入、`docs:sync-*`、任何 `--write`、生成器、修复器或会改变载荷的命令，也不得安装依赖。安装器绝不读取连接设置或凭据。
