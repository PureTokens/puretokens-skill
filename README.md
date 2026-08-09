# Pure Tokens Skills

`puretokens-skill` 是 Pure Tokens 的私有 Skill 管理仓库。它管理可安装的 Skill 指令、版本、兼容性声明与校验工具；它不保存 API Key、Router Token、用户配置或模型路由逻辑。

当前 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens_media` | 让支持 MCP 的客户端按用户明确指定的图片或视频模型，通过 Pure Tokens MCP 生成媒体。 |

## 设计边界

```text
用户自然语言 → Skill → Pure Tokens MCP → 本地 Router → Pure Tokens 服务
```

- Skill 负责理解“用 image2”“用 Grok Video”等用户表达，先查询媒体目录、做确定性匹配，并在歧义时询问用户。
- MCP 只接受精确模型 ID、执行严格校验、提交任务和轮询结果；它不做自然语言识别、不猜模型、不静默换模型。
- BFF / Router 决定模型是否在当前分组可用，以及它是否明确支持图片或视频协议。

## 本地安装

先克隆本仓库，再把 Skill 安装到 Codex 的本机 Skill 目录：

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
node bin/puretokens-skill.js list
node bin/puretokens-skill.js install puretokens_media
```

默认安装目标是：

```text
~/.codex/skills/puretokens_media/
```

也可以安装到当前项目：

```bash
node bin/puretokens-skill.js install puretokens_media --target .codex/skills
```

安装器不会覆盖同名目录。若目标已存在，请先检查并备份已有 Skill，再重新安装。

## 使用前提

`puretokens_media` 需要用户已在 Pure Tokens Desktop 中完成对应客户端的“验证并应用”。Desktop 会为 Claude Code、Claude Desktop、ChatGPT/Codex、WorkBuddy、Gemini CLI、Grok Build 与 OpenCode 配置 `puretokens-image` MCP 服务。

Trae 目前不由该 Skill 自动配置或调用媒体 MCP。

Skill 不包含、也不会索取：

- Pure Tokens API Key
- 本地 Router Token
- Cookie、密码或任意云端凭证
- 用户的原始图片、视频或生成结果

## 模型选择规则

当用户明确说出模型时，例如“用 image2 生成一只可爱的狗”或“用 Grok Video 生成 15 秒广告”，Skill 必须：

1. 调用 `puretokens_list_media_models`；
2. 仅依据目录返回的精确模型 ID、显示名、别名、供应商与媒体类型进行匹配；
3. 匹配唯一时，将**精确模型 ID**传给生成工具；
4. 未找到或匹配多个模型时，请用户选择；
5. 不从模型名称、供应商名称或历史默认值猜测模型。

当前 Router 目录至少返回 `id` 与 `capabilities`。显示名称、别名和供应商字段是目录的向后兼容可选扩展；字段缺失时，Skill 只能使用精确 ID 或向用户询问，不能虚构匹配结果。

## 开发与校验

```bash
npm run check
node bin/puretokens-skill.js list
node bin/puretokens-skill.js install puretokens_media --target /private/tmp/puretokens-skill-test
```

本仓库暂不配置 GitHub Actions，也不发布 npm 包。后续是否公开、发布 npm 或由 Desktop 自动同步，必须单独确认。
