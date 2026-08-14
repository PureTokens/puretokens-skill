<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Skill — 一个 Skill，连接所有模型" width="100%" />
</p>

<p align="center">
  <a href="./README.md">English</a> · <strong>中文</strong> · <a href="./CHANGELOG.zh-CN.md">更新日志</a>
</p>

# Pure Tokens Skills

`puretokens-skill` 是 Pure Tokens Skill 的源仓库。它管理 Skill 指令、版本、兼容性声明、各客户端安装说明和校验工具；不保存用户凭据、Router 配置或模型路由逻辑。

## 3 步开始

1. 在 Pure Tokens Desktop 中，为当前客户端选择包含你要使用的图片或视频模型的分组，然后点击 **验证并应用**。
2. WorkBuddy 只需点击 **验证并应用**，Pure Tokens 会自动安装其媒体路由规则；其他客户端按客户端安装表安装 `puretokens_media`，或复制下面的 Agent 提示词。
3. 新建会话。直接说“生成一只可爱的狗”即可使用默认图片模型；也可以说“使用 Nano Banana Pro 生成……”。

> **使用指定模型前请先确认：** 该模型必须位于当前客户端已选择的至少一个分组中。例如，要让 Skill 使用 `image2`，先选择包含 `gpt-image-2` 的分组。每次修改分组后，都要点击 **验证并应用**、重启目标客户端，并新建会话。Skill 只能调用当前已选分组返回的模型，不能调用公开目录中的全部模型。

当前 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_media` | 查询 Pure Tokens 余额或模型价格，或按当前目录精确选择图片/视频模型，提交一次任务并轮询同一任务。 |
| `puretokens_workbuddy_router` | 由 Desktop 受管的 WorkBuddy 路由规则：普通生图、生视频优先走已配置的 Pure Tokens MCP。 |

## 查询余额

直接说“查询我的 Pure Tokens 余额”“我还剩多少余额”或“查一下余额”。Skill 会直接调用 `puretokens_get_balance`，只展示 Pure Tokens 返回的余额和额度字段；不会读取模型目录、API Key、Cookie、密码或本地 Router 凭据。

## 查询模型价格

直接说“gpt-image-2 多少钱？”或“image2 的价格”。Skill 会先读取实时模型目录，只在已登记别名唯一对应到目录中的精确模型 ID 后，调用 `puretokens_get_model_price`。返回结果会展示当前选中分组中的全部价格、分组倍率和更新时间。

价格不会根据模型名称猜测。如果模型不可用、匹配有歧义，或者计费方式是动态计费，Skill 会如实提示，不会估算价格、偷偷选择分组或换成其他模型。

## 图片模型

下面是公开目录当前列出的图片模型。你的客户端或分组可能只显示其中一部分；只有 `puretokens_list_media_models` 实时返回精确 ID 或别名时，模型才可以使用。

用户不需要记完整 ID。像 `image2`、`Nano Banana Pro` 这样的已登记说法由 Skill 负责理解，然后仍会先去实时目录确认，确认成功才发送请求。只说“生成图片”时默认使用 `gpt-image-2`；如果当前分组没有该精确模型，Skill 会停止并展示可用候选，不会偷偷换模型。

| 模型 ID | 也可以这样说 | 适合做什么 | 真实使用示例 |
| --- | --- | --- | --- |
| `gpt-image-2` | `image2`、`gpt image 2`、`openai image 2` | 高质量海报、产品视觉、插画 | `使用 image2 做一张橙色产品发布海报。` |
| `gemini-3.0-pro-image` | `gemini pro image` | 细节丰富的概念图和营销图 | `使用 gemini pro image 做一张高级云计算主视觉。` |
| `gemini-3.1-flash-lite-image` | `gemini flash lite image` | 快速缩略图和社交媒体草稿 | `使用 gemini flash lite image 做三张明亮的社交媒体缩略图。` |
| `gemini-3-pro-image-preview` | `nano banana pro` | 更细致、更专业的 Gemini 图片生成 | `使用 Nano Banana Pro 做一张高级产品主视觉。` |
| `gemini-3.1-flash-image-preview` | `nano banana 2` | 更快速的 Gemini 生图和对话式编辑 | `使用 Nano Banana 2 做一张明亮的产品社交海报。` |
| `grok-imagine-1.0` | `grok image`、`grok imagine` | 快速创意和轻松有趣的场景 | `使用 grok-imagine-1.0 画一只在城市公园里的快乐机器人。` |
| `grok-imagine-image` | `grok image`、`grok imagine` | 社交内容和日常生图 | `使用 grok-imagine-image 做一张咖啡店开业宣传图。` |
| `grok-imagine-image-quality` | `grok quality image` | 更精细的品牌主视觉 | `使用 grok quality image 做一张精致的应用商店横幅。` |
| `wan2.7-image` | `wan image`、`wan 2.7 image` | 中文海报和产品宣传图 | `使用 wan 2.7 image 做一张春节促销海报。` |

Skill 只提交一次 `puretokens_generate_image`，然后用 `puretokens_image_result` 查询同一个任务。

只说 `Nano Banana` 时，表示 Gemini Nano Banana 模型家族。当前目录同时有 `Nano Banana Pro` 和 `Nano Banana 2` 时，Skill 会让你选择；只有一个可用时才会直接使用。这样不会把已指定的模型悄悄换成另一个。

## 视频模型

下面是公开目录当前列出的视频模型。模型必须在实时目录中声明 `video` 能力，才能生成视频。

| 模型 ID | 也可以这样说 | 适合做什么 | 真实使用示例 |
| --- | --- | --- | --- |
| `grok-imagine-video` | `grok video`、`grok imagine video` | 短视频和快速创意片段 | `使用 grok-imagine-video 做一条 5 秒咖啡广告。` |
| `grok-imagine-video-1.5` | `grok 1.5 video`、`grok video 1.5` | 更完整的短广告 | `使用 grok 1.5 video 做一条 15 秒、16:9 的产品广告。` |

Skill 只提交一次 `puretokens_generate_video`，然后用 `puretokens_video_result` 查询同一个任务。只说“生成视频”时默认使用 `grok-imagine-video-1.5`；如果当前分组没有该精确模型，Skill 会停止并展示可用候选。

如果模型不存在、有多个候选，或者没有对应能力，Skill 会列出实时候选并让你选择，不会偷偷换模型。

## 你可以这样说

| 你想做什么 | 直接这样说 |
| --- | --- |
| 生成图片 | `生成一只可爱的狗。` |
| 生成视频 | `生成一条 15 秒、16:9 的产品广告。` |
| 查询模型价格 | `gpt-image-2 多少钱？` |
| 使用 Nano Banana | `使用 Nano Banana Pro 做一张高级产品主视觉。` |
| 查看可用模型 | `列出我现在能用的图片和视频模型。` |

## 设计边界

```text
用户自然语言 → Skill → Pure Tokens MCP → 本地 Router → Pure Tokens 服务
```

- Skill 负责理解“用 image2”“用 Grok Video”等表达，先查询媒体目录，唯一匹配后选择工具，并在歧义时询问。
- MCP 只接受精确模型 ID，执行参数校验、单次提交和结果轮询；它不做自然语言识别、不猜模型、不静默换模型。
- BFF / Router 仍是模型是否可用、分组权限和媒体协议的权威来源。

## 前置条件

使用指定图片或视频模型前，请按以下顺序完成：

1. 在 Pure Tokens Desktop 中打开目标客户端的配置。
2. 选择包含目标模型的一个或多个分组。
3. 点击 **验证并应用**。
4. 重启目标客户端，并新建会话。

Skill 只能使用当前已选分组中的模型。如果实时媒体目录没有目标模型，请回到客户端配置，选择包含该模型的分组后再次应用配置。Desktop 会为支持的客户端配置名为 `puretokens-image` 的 MCP 服务。Skill 不会替代 MCP 配置，也不会携带任何凭据。

## 从 GitHub 安装和更新

Pure Tokens Desktop 不会把共享 Skill 文件写入客户端目录，也不会把共享 Skill 内容绑定在 Desktop 版本中。唯一例外是 WorkBuddy 的小型版本化 `puretokens_workbuddy_router`：在点击 **验证并应用** 后，Desktop 会通过事务方式受管写入它，使普通图片/视频请求进入已配置的 Pure Tokens MCP。共享 Skill 的最新安装说明和文件仍以本仓库为准。安装共享 Skill 前，先在 Pure Tokens Desktop 对目标客户端完成“验证并应用”，然后重启目标客户端并新建会话。

### Codex、Claude Code、Gemini CLI、OpenCode

在本仓库页面点击 **Code → Download ZIP**，或先克隆仓库，然后在仓库目录执行对应命令。需要 Node.js 20 或更高版本。

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
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

### 直接复制给 Agent

把下面提示词复制给能够执行本机命令的 Agent：

```text
请从公开仓库 https://github.com/yanyansay/puretokens-skill 为我当前使用的客户端安装 Pure Tokens Skill。

1. 先判断当前客户端是 Codex、Claude Code、Gemini CLI 还是 OpenCode。
2. 将仓库克隆或下载到临时工作目录。
3. 只用 `install` 或 `upgrade` 安装 `puretokens_media` 到对应的用户 Skill 目录：
   - Codex：~/.codex/skills
   - Claude Code：~/.claude/skills
   - Gemini CLI：~/.gemini/skills
   - OpenCode：~/.config/opencode/skills
5. 不要覆盖其他 Skill。
6. 不要读取、索取、打印或保存 API Key、Cookie、密码、Router Token 或本地授权地址。
7. 返回实际安装目录和操作结果。

如果当前是 Claude Desktop，不要声称已经自动安装。请按 README 生成 ZIP，并告诉我应该在哪里上传和启用。如果当前是 WorkBuddy，请让我在 Pure Tokens Desktop 点击 **验证并应用**；不要手动创建或替换 `puretokens_workbuddy_router`。
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
node .\bin\puretokens-skill.js install puretokens_media --target "$HOME\.codex\skills"
```

其他客户端将目标目录改为 `$HOME\.claude\skills`、`$HOME\.gemini\skills` 或 `$HOME\.config\opencode\skills`。如果 PowerShell 找不到 `node`，先从 Node.js 官方网站安装 Node.js LTS，再重新打开 PowerShell。

## Claude Desktop 导入与 WorkBuddy 路由

不要把复制到 `~/.codex/skills` 当成 Claude Desktop 安装完成。该目录只适用于 Codex 本机 Skill。

Claude Desktop 使用图形界面上传本地 Skill 包。生成 ZIP：

```bash
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.2.7.zip
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

WorkBuddy 不需要手动上传或启用 `puretokens_workbuddy_router`。在 Pure Tokens Desktop 中选择兼容分组并点击 **验证并应用** 后，Desktop 会原子化受管该路由 Skill 和 `puretokens-image` MCP 条目。随后重启 WorkBuddy 或新建会话。用户直接说生图、生视频时会优先走 Pure Tokens；用户明确指定 WorkBuddy 内置 `ImageGen` 或 `VideoGen` 时仍保留该选择。

更新 Claude Desktop 时从 GitHub 获取新版本、重新生成 ZIP、停用旧 Skill、上传新 ZIP 并启用。WorkBuddy 会在下一次点击 **验证并应用** 时更新受管路由规则。不要直接删除 MCP 配置；MCP 由 Pure Tokens Desktop 管理。

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

`puretokens_media` 必须先调用 `puretokens_list_media_models`，只依据本次响应的 `id`、`displayName`、`aliases`、`provider` 和 `capabilities` 匹配。生成工具必须传精确 `model` 和稳定 `request_id`。一次用户请求只提交一次；宿主重试时复用同一 `request_id`，结果工具始终使用同一 `task_id` 和原始模型。

媒体完成后会展示 MCP 返回的实际精确模型、保存文件名和 `Downloads/Pure Tokens`。只有 MCP 返回原生 `image` 内容时，图片才可在支持的宿主内预览。视频在大小受限时会携带原生 MCP 资源，支持该资源的宿主可以预览；较大的视频仍会成功保存为本机 MP4，并从同一下载文件夹打开。

查询价格时，Skill 会使用解析后的精确模型 ID 调用 `puretokens_get_model_price`，展示每个分组的价格结果。不会推测价格、静默选择分组，也不会把动态计费规则当成固定金额。

模型歧义、目录为空、MCP 不可用、工具错误和轮询超时的行为测试见 `skills/puretokens_media/references/behavior-scenarios.json`。任何错误都不得自动换模型或重新提交，除非用户明确选择了新的具体模型。

## 安全边界

Skill 不包含、也不会索取：

- 云端凭据、Router Token、Cookie 或密码；
- 用户配置、分组路由或支付数据；
- 本地授权地址或客户端私有文件；
- 图片、视频、任务结果或提示词历史。

Trae 目前不支持该 Skill 的媒体 MCP 流程。
