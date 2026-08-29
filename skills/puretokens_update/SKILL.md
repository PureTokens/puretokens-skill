---
name: puretokens_update
description: 用户要求检查、安装或升级本机已安装的 Pure Tokens Skills 时使用。
---

# Pure Tokens Update

升级的是本机已安装的官方 Pure Tokens Skills，不涉及账户、模型、余额或媒体生成。先读取已安装的 `references/execution-contract.json` 和 `references/behavior-scenarios.json`；它们定义升级的安全边界和用户可见结果。

## 升级规则

- 只在用户明确要求安装、更新、升级、同步或检查 Pure Tokens Skills 时使用。不得在普通图片、视频、模型、余额或连接请求中自行运行升级。
- 先根据当前运行环境识别宿主：Codex 使用 `~/.agents/skills`（Windows PowerShell 为 `$env:USERPROFILE\.agents\skills`）；Claude Code 使用 `~/.claude/skills`；Gemini CLI 使用 `~/.gemini/skills`。无法识别时，先问用户使用哪个宿主，不得猜测或修改目录。
- 若当前回合没有可执行终端、PowerShell、exec 或 shell 工具，说明无法代用户完成本机升级，并提供与该宿主匹配的官方可复制命令；不得声称已升级。
- 有终端时，新建临时工作目录，不删除、覆盖或复用用户已有目录。克隆 `https://github.com/PureTokens/puretokens-skill.git` 的 `main` 分支，进入仓库后先运行 `npm run check`。校验失败时报告失败并停止，不修改任何已安装 Skill。
- 对 Codex、Claude Code 或 Gemini CLI，使用 `node bin/puretokens-skill.js sync --target <installation-root>`。它会安装缺失的官方 Skill，并只升级含匹配 `SKILL.md` 与 `skill.json` 的已受管同名 Skill；只要任一同名目录不是受管 Skill，必须在写入前停止并报告冲突，绝不覆盖、删除或改名该目录。
- 对 Claude Desktop，运行 `npm run bundle:claude` 生成当前版本的 ZIP 包，报告生成路径，并引导用户在 Claude Desktop 的 Skills 设置中上传并启用新包。不得读取或修改 Claude Desktop 配置、自动点击 UI、删除旧包或假称已在 Desktop 内完成升级。
- 升级完成后，逐项报告安装或升级的 Skill、未改动的冲突和失败原因（如有），并提醒用户新开宿主会话后再测试。

## 安全边界

- 绝不读取、扫描、展示、复制、修改或索取 API Key、Base URL、认证文件、AI 客户端配置、模型配置、MCP 配置、环境变量、Shell 配置、系统代理或账户数据。
- 不调用 Images、Videos、Models、余额或连接 API；不使用 MCP、本地代理、sidecar 或任何媒体生成路径。
- 不使用第三方包镜像；不执行卸载；不删除用户文件。只允许 `sync` 对目标根目录内已验证的官方 Skill 目录做安装或原子升级。
