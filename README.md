# Pure Tokens Skills

`puretokens-skill` 是 Pure Tokens Skill 的源仓库。它管理 Skill 指令、版本、兼容性声明、各客户端安装说明和校验工具；不保存用户凭据、Router 配置或模型路由逻辑。

当前 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_media` | 按当前目录精确选择图片或视频模型，通过 `puretokens-image` MCP 提交一次任务并轮询同一任务。 |

## 设计边界

```text
用户自然语言 → Skill → Pure Tokens MCP → 本地 Router → Pure Tokens 服务
```

- Skill 负责理解“用 image2”“用 Grok Video”等表达，先查询媒体目录，唯一匹配后选择工具，并在歧义时询问。
- MCP 只接受精确模型 ID，执行参数校验、单次提交和结果轮询；它不做自然语言识别、不猜模型、不静默换模型。
- BFF / Router 仍是模型是否可用、分组权限和媒体协议的权威来源。

## 前置条件

用户必须先在 Pure Tokens Desktop 中对当前客户端完成“验证并应用”，并重启目标客户端。Desktop 会为支持的客户端配置名为 `puretokens-image` 的 MCP 服务。Skill 不会替代 MCP 配置，也不会携带任何凭据。

## 从 GitHub 安装和更新

Pure Tokens Desktop 不会把 Skill 文件写入客户端目录，也不会把 Skill 内容绑定在 Desktop 版本中。请始终从本仓库获取最新安装说明和 Skill 文件。安装前先在 Pure Tokens Desktop 对目标客户端完成“验证并应用”，然后重启目标客户端并新建会话。

### Codex、Claude Code、Gemini CLI、OpenCode

在本仓库页面点击 **Code → Download ZIP**，或先克隆仓库，然后在仓库目录执行对应命令。需要 Node.js 20 或更高版本。

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
node bin/puretokens-skill.js validate
```

按客户端安装到各自的用户级 Skill 目录：

```bash
# Codex
node bin/puretokens-skill.js install puretokens_media --target ~/.codex/skills

# Claude Code
node bin/puretokens-skill.js install puretokens_media --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_media --target ~/.gemini/skills

# OpenCode
node bin/puretokens-skill.js install puretokens_media --target ~/.config/opencode/skills
```

更新时从 GitHub 重新下载或拉取仓库，然后执行对应的 `upgrade` 命令：

```bash
git pull
node bin/puretokens-skill.js upgrade puretokens_media --target ~/.codex/skills
```

把 `~/.codex/skills` 换成当前客户端的目标目录即可。升级只替换由 Pure Tokens 管理、且包含匹配 `skill.json` 与 `SKILL.md` 的目录，不会覆盖其他 Skill。

### Windows PowerShell

```powershell
git clone https://github.com/yanyansay/puretokens-skill.git
Set-Location puretokens-skill
node .\bin\puretokens-skill.js validate
node .\bin\puretokens-skill.js install puretokens_media --target "$HOME\.codex\skills"
```

其他客户端将目标目录改为 `$HOME\.claude\skills`、`$HOME\.gemini\skills` 或 `$HOME\.config\opencode\skills`。如果 PowerShell 找不到 `node`，先从 Node.js 官方网站安装 Node.js LTS，再重新打开 PowerShell。

## Claude Desktop 和 WorkBuddy 导入

不要把复制到 `~/.codex/skills` 当成 Claude Desktop 安装完成。该目录只适用于 Codex 本机 Skill。

Claude Desktop 和 WorkBuddy 使用图形界面上传本地 Skill 包。生成 ZIP：

```bash
node bin/puretokens-skill.js validate
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.2.0.zip
```

ZIP 内部结构为：

```text
puretokens_media/
├── SKILL.md
├── skill.json
└── references/
    ├── behavior-scenarios.json
    └── model-catalog-contract.md
```

在 Claude Desktop 中打开 **Settings → Features → Skills**（部分版本显示为 **Customize → Skills**），选择 **Upload skill**，上传 ZIP 并启用 `Pure Tokens Media`。如果当前版本没有 Skills 入口，则该版本不能导入自定义 Skill，只能使用 MCP 的工具描述；升级或使用支持 Skills 的 Claude 客户端后再导入。

在 WorkBuddy 中打开 **Skills → 添加技能 → 上传技能**，选择同一个 ZIP，确认 `Pure Tokens Media` 出现在已安装列表并启用，然后新建会话。WorkBuddy 负责导入和本地配置，Pure Tokens Desktop 不会尝试写入 WorkBuddy 未公开的本地目录。

更新时从 GitHub 获取新版本、重新生成 ZIP，在 Claude Desktop 或 WorkBuddy 中停用旧 Skill、上传新 ZIP 并启用。卸载时在对应 Skills 页面关闭并删除 Skill。不要直接删除 MCP 配置；MCP 由 Pure Tokens Desktop 管理。

## Codex 本机安装、升级和卸载

默认目录是 Codex 的本机 Skill 目录：

```bash
node bin/puretokens-skill.js list
node bin/puretokens-skill.js install puretokens_media
node bin/puretokens-skill.js upgrade puretokens_media
node bin/puretokens-skill.js uninstall puretokens_media --yes
```

也可以指定项目目录：

```bash
node bin/puretokens-skill.js install puretokens_media --target .codex/skills
node bin/puretokens-skill.js upgrade puretokens_media --target .codex/skills
node bin/puretokens-skill.js uninstall puretokens_media --target .codex/skills --yes
```

升级会先把旧目录原子移到临时备份位置，替换成功后再清理备份。卸载要求显式 `--yes`，并且只删除包含匹配 `skill.json` 和 `SKILL.md` 的受管 Skill 目录。

## 模型选择规则

`puretokens_media` 必须先调用 `puretokens_list_media_models`，只依据本次响应的 `id`、`displayName`、`aliases`、`provider` 和 `capabilities` 匹配。生成工具必须传精确 `model` 和稳定 `request_id`。一次用户请求只提交一次；宿主重试时复用同一 `request_id`，结果工具始终使用同一 `task_id`。

模型歧义、目录为空、MCP 不可用、工具错误和轮询超时的行为测试见 `skills/puretokens_media/references/behavior-scenarios.json`。任何错误都不得自动换模型或重新提交，除非用户明确选择了新的具体模型。

## 安全边界

Skill 不包含、也不会索取：

- 云端凭据、Router Token、Cookie 或密码；
- 用户配置、分组路由或支付数据；
- 本地授权地址或客户端私有文件；
- 图片、视频、任务结果或提示词历史。

Trae 目前不支持该 Skill 的媒体 MCP 流程。

## 校验

```bash
npm run check
node bin/puretokens-skill.js validate
npm test
```

本仓库不启用 GitHub Actions，不发布 npm 包，也不自动发布 Claude Desktop Skill；只有明确执行 bundle 并在 Claude Desktop 中上传后才会生效。
