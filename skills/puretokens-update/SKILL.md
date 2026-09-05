---
name: puretokens-update
description: 用户要求检查、安装或升级本机已安装的 Pure Tokens Skills 时使用。
---

# Pure Tokens Update

升级的是本机已安装的官方 Pure Tokens Skills。`init` 用于检查当前宿主是否能验证固定 Pure Tokens API，并向用户展示 `references/usage-guide.md` 使用须知；它不修改账户、模型、余额或媒体任务。先读取已安装的 `references/execution-contract.json`、`references/behavior-scenarios.json` 和 `references/usage-guide.md`；它们定义初始化、升级的安全边界和用户可见结果。

## 初始化

- 用户要求初始化、检查当前配置是否可用、或询问如何使用 Pure Tokens Skills 时，调用 Skills 根目录下与当前 Skill 同级的 `../.puretokens-executor/puretokens-api init --host <当前宿主 ID>`。执行器只请求一次固定 `GET https://api.puretokensx.com/v1`；它只能私密读取当前宿主已写入的有效连接记录中的匹配 Key，绝不展示、索取、复制、保存或报告 Key、Base URL、provider 或完整宿主配置。
- 只在回执 `configuration_status: "verified"` 且 `api_identity_confirmed: true` 时说明当前宿主已能访问固定 Pure Tokens API；其他结果只能说明未完成验证，并给出下一步，不得猜测连接归属或声称媒体可用。
- 初始化回执包含使用示例；同时展示 `references/usage-guide.md` 中的使用须知。初始化不调用余额、模型、Images 或 Videos API，不创建付费任务，也不自动修复或修改宿主配置。

## 升级规则

- 只在用户明确要求安装、更新、升级、同步或检查 Pure Tokens Skills 时使用。不得在普通图片、视频、模型、余额或连接请求中自行运行升级。
- 先根据当前运行环境识别宿主：Claude Code 使用 `~/.claude/skills`；Codex 使用 `~/.agents/skills`（Windows PowerShell 为 `$env:USERPROFILE\.agents\skills`）；WorkBuddy 使用 `~/.workbuddy/skills`；Gemini CLI 使用 `~/.gemini/skills`；Grok Build 使用 `~/.grok/skills`；OpenCode 使用 `~/.config/opencode/skills`；Trae 使用 `~/.trae/skills`。无法识别时，先问用户使用哪个受支持宿主，不得猜测或修改目录。
- 若当前回合没有可执行终端、PowerShell、exec 或 shell 工具，说明无法代用户完成本机升级，并请用户把官方仓库安装提示交给具备本机终端的 Agent；不得声称已升级，也不得编造手工复制命令。
- 升级时，先通过当前宿主可用的官方 GitHub 仓库获取方式取得本仓库 `main` 的**最新本地检出**；不得从已安装的 Skill 目录、旧缓存、第三方镜像或自定义安装 ZIP 取源。随后从该检出直接执行匹配平台的源码同步脚本：macOS/Linux 使用 `runtime/puretokens-skill-install.sh sync --host <当前宿主 ID>`，Windows 使用 `runtime/puretokens-skill-install.ps1 sync -Host <当前宿主 ID>`。脚本自行推导全局 Skill 目录，校验当前平台原生执行器的 SHA-256 后同步它。用户不需要安装 Node、npm、包管理器、Git、Go、Python 或开发环境；若当前 Agent 无法以其已有工具取得官方仓库或执行本机脚本，应说明无法完成，绝不下载或要求用户安装依赖。不得运行已安装的旧更新器、复用用户连接配置、把远程内容直接管道给 Shell/PowerShell，或再次下载自定义载荷。最新检出只允许读取、校验和同步；不得运行 `node`、`npm`、生成器、`docs:sync-*`、任何 `--write`、修复器或会改写检出的命令。
- 平台安装器会安装缺失的官方 Skill 和一个同级的 `.puretokens-executor`。它只升级含匹配 `SKILL.md` 与 `skill.json` 的已受管同名 Skill，并且只替换含有受管 `runtime.json` 与平台二进制的旧执行器。当前 Skill 全部成功同步后，会删除目标根目录内已验证的旧受管 Skill，以及旧版更新遗留的同名隐藏备份，避免宿主继续发现废弃的 `puretokens_media` 等目录。若存在 `.puretokens-runtime`，只有它同时含有旧官方 `runtime.json` 与 `puretokens-direct-api.mjs`，且 `runtime.json.name` 精确为 `puretokens-direct-api-runtime` 时才会删除；任何其他同名目录均不视为冲突，也绝不删除。当前或旧同名 Skill/执行器目录不是受管内容时，必须在写入前停止并报告冲突，绝不覆盖、删除或改名该目录。
- 当目标是 Codex 的 `~/.agents/skills` 时，安装器会在写入任何 Skill 前只通过官方 `codex plugin` 接口检查已安装插件；若精确发现旧 `puretokens-media`，会使用其精确插件 selector 移除、重新列出并确认已不存在，才允许继续同步，防止旧插件继续注入 `Puretokens Media` 指令。绝不读取插件配置、目录、凭据或其他插件。若 Codex CLI 不可用、插件受工作区管理、移除失败或无法复核，必须停止且不输出同步成功回执；提示用户在 Codex Plugins 卸载，或请工作区管理员处理后重新安装。移除确认后，必须完全退出并重新启动 Codex，再新开对话测试；当前已打开的对话会保留已加载的旧 Skill 指令。
- 仅当终端输出 `Pure Tokens Skills <版本> synchronized with the native API executor at <目录>` 时，才报告升级成功，并逐项报告安装或升级的 Skill 与版本号，再提醒用户新开宿主会话后测试。随后安装器会输出 `init` 结果与使用须知；`init` 未验证并不回滚已完成的同步。若未收到同步完成回执，包括宿主执行超时或中断，只能说明“未确认同步完成”，不得称已更新、不得猜测版本、不得自动重试或创建后台任务；用户明确要求后才可再次执行一次。

## 安全边界

- Skill 指令、安装器和源码同步绝不读取、扫描、展示、复制、修改或索取 API Key、Base URL、认证文件、AI 客户端配置、模型配置、MCP 配置、环境变量、Shell 配置、系统代理或账户数据。仅已安装执行器在 `init` 或 API 请求时，可按当前宿主的已验证字段契约私密读取一个有效连接记录；它绝不扫描其他位置，也不展示、复制、保存或报告所读内容。
- 不调用 Images、Videos、Models、余额或连接 API；不使用 MCP、本地代理、sidecar 或任何媒体生成路径。
- 不使用第三方包镜像；不删除用户文件。`sync` 只会删除目标根目录内已验证的废弃官方 Skill 与其旧隐藏备份，以及可被精确识别的旧官方 Node 运行器；当且仅当精确匹配到 Codex 的旧官方插件时，才通过官方 `codex plugin` 接口将其移除并重新列出确认。无法确认移除时停止，不输出成功回执。
- 官方最新检出只允许用于读取、静态校验和同步；不得运行 `node`、`npm`、Git 写入、`docs:sync-*`、任何 `--write`、生成器、修复器或会改变检出的命令，也不得安装依赖。源码同步脚本绝不读取连接设置或凭据。
