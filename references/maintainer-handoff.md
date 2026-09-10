# Pure Tokens Skills 当前开发状态

更新日期：2026-09-11。此文件只维护当前任务与未解决事项；历史实现、发布记录及审查证据见 Git、CHANGELOG 和 references/audits，不将历史状态作为当前事实。

## 目标与授权

当前任务：按用户要求减少历史兼容包袱、统一安装与使用路径。基线main为01fed3d，已推送。用户已授权删除旧Codex插件迁移和不再需要的历史安装入口；用户已授权提交推送本轮全部修改；未授权真实客户端配置/安装或付费任务。

已删除：Codex插件检查/迁移；旧Node安装、升级、卸载CLI；两份旧迁移ZIP及0.14.3兼容标记；退役Skill/Node runtime扫描清理；嵌入历史目录快照及生成器；未公开的task提交别名。开发list命令仍可用。当前原生fetch→sync→同宿主init为唯一用户安装链，平台包或同提交源归档仅为同一入口内部取包策略。

保留当前受管目录哈希校验、锁、失败回滚、当前官方源码完全匹配的无标记目录接管、任务记录缺少证明时安全续查、当前客户端配置格式和重试规则。它们不增加普通生成前置步骤，不扫描用户历史安装。

npm run check通过（Go、go vet、56项Node）；移除已废弃功能测试并增加退役目录保持不动、旧入口不得重现、同宿主init与不调用插件的回归。六平台执行器重建及executor:verify通过；六份草稿平台包重建。两份迁移ZIP已删除，不再构建。git diff --check通过。已于北京时间2026-09-11 00:55重新采集官方目录，23个模型；release:validate现已通过。按用户授权提交推送；未发布Release。

## 有效决定与实现

- 六个独立Skill入口与单次原生执行器不变；固定API请求、精确连接读取、无凭据持久化或传输回退。安装不增加用户运行时依赖。
- 默认或精确模型只读选中profile；索引仅选择/别名解析时读，实时目录按需。普通生成不运行init、doctor、余额或preflight。
- 异步submit只发一次并立即回执；同任务status/wait/content独立。图片默认5秒首次查询、3秒间隔、40次/120秒窗口；视频7次/300秒窗口。最多两个自动前台窗口，预算跨续接保存。错误、对账和无法容纳的重试停止自动续等。
- 下载时计算摘要；记录中的文件证明匹配后可本地恢复交付，实际附件handoff成功后才标记delivered。逐索引下载交付，不后台预取或重提。
- Kimi Code按默认模型关联连接解析；Qoder按真实端点选择唯一连接，不按名称识别、不比较多个凭据。安装、doctor和凭据适配共享相同目录优先级；Kimi doctor还检查已声明共享Skill目录。
- 新宿主依据：Switch 0ccc470两客户端适配器、已安装Kimi Code程序静态skillRoots/data-home代码与Qoder官方文档，详见credential-adapters.md。旧Python Kimi CLI不是新Kimi Code的路径依据。未读取真实连接。
- 本轮清理：宿主排查段移至按需desktop-hosts，纠正init不依赖配置选择的过度概括；校验器从host-support推导适配器清单；共享严格JSON解析器去除ZCode专属命名；未知宿主路径解析明确拒绝；清除本文件旧候选/旧计时等冲突叙述。历史迁移资产已按用户新决定删除。

## 验证与剩余事项

Windows脚本覆盖已增加，未Windows实机运行。当前清理阶段验证结果见上节；旧阶段测试数不作为当前计数。

- 官方目录已重新采集并同步至23模型，采集时间2026-09-10T16:55:34.848Z；新增三个Grok模型。Nano Banana画幅与Seedance首尾帧仍为单独来源的补充声明，保留出处并不声称来自本次公开响应。npm run check（56项Node及Go/go vet）和release:validate均通过。
- 真实宿主API、原生附件交付及端到端耗时仍待指定环境验收；host-acceptance.json是唯一验收结果。夹具、静态源码和安装成功不能代替实机。
- 远程WorkBuddy诊断及Windows安装体验的实机结果仍未确认。
- 精确报价、服务端幂等和未知任务找回未实现；不从任务记录或本地校验推导保证。
- 发布前重建六平台执行器与平台候选；脏工作区候选sourceCommit=null不是正式发布资源。提交时包含新源码/测试/说明。

## 维护入口

AGENTS.md与CONTRIBUTING.md定义长期约束；skills/为安装指令及profile，schemas/为字段契约，runtime/executor/为行为实现，references/host-support.json为宿主声明，references/host-acceptance.json为实机证据。工程门禁npm run check；发布门禁npm run release:validate；executor:verify独立重建比较六平台产物。
