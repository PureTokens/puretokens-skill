# Jev 同步评估

`<执行器> evaluate --host <当前宿主> --request <绝对 UTF-8 JSON 文件>`

兼容 Windows PowerShell 写出的 UTF-8 BOM；不接受 UTF-16 请求文件。

以下三个问题共用同一份文本，只产生一次请求：

```json
{
  "model": "jev-latest",
  "state": "付款连续失败三天，请尽快协助。",
  "questions": {
    "department": {
      "type": "choice",
      "instructions": "由哪个团队处理？",
      "criteria": {
        "billing": "支付、账单或退款",
        "technical": "软件故障或集成问题",
        "other": "不属于以上类别"
      }
    },
    "urgency": {
      "type": "noul",
      "instructions": "是否表达了紧急需求？",
      "criteria": {
        "true": "明确要求尽快处理",
        "false": "没有紧急需求"
      }
    },
    "severity": {
      "type": "score",
      "instructions": "依据所描述的付款影响程度评分。",
      "criteria": ["无付款影响", "偶发付款失败", "连续付款失败"]
    }
  }
}
```

state 和 instructions 支持字符串、JSON 对象或数组。choice 描述可为 null；
score 各档须为字符串／对象／数组，顺序不可反转；noul 的 criteria 可省略，
提供时只接受 true、false 描述。每个问题都要 instructions。
问题 ID 用字母开头的 1–64 位 ASCII 字母、数字、下划线或连字符；选项名最长
128 字符，非空且无控制字符。这些是执行器明确支持的输入范围，不宣称覆盖
上游 SDK 的所有宽松写法。未知字段、重复 JSON 键和过大请求在读取凭据前拒绝。

成功回执为 `command=evaluate`、`submission_outcome=accepted`、
`next_step=done`，`result.answers` 按原问题 ID 返回：

- choice：`type`、`choice`、完整 `probabilities`、`confidence`。
- score：`type`、`score`、完整 `probabilities`、`confidence`、
  `level_count`。档位从 0 起，按原请求解释；执行器不转发服务端任意 legend 文本。
- noul：`type`、`noul`，范围 0–1，不另造 confidence。

执行器检查答案集合／类型、选项、数值范围和概率总和（浮点容差 0.001）；
choice 必须为最大概率项，score 必须与档位概率加权值一致（容差 0.001）。
`result.model` 保留服务端实际 Jev 版本；模型别名可解析为新版本，不把实际
版本替换为请求别名。usage 仅保留返回的合法 token 计数，不据此计算价格。
成功结果可能含用户定义选项，只用于本次结果展示；安全 support 摘要不含
state、问题、选项、评分或文件路径。

本地校验失败是 `not_submitted`；完整收到 HTTP 4xx 响应是 `rejected`；
网络错误、3xx、5xx 或无效成功响应为 `unknown`。它们均为 `await_user`，
没有自动重试。HTTP 拒绝也不作为钱包或退款凭据。客户端进程被取消／没有回执
同样可能已经发出请求；不得用“没有输出”判断未提交。

评估没有异步任务记录及恢复接口，不接受 `--record`、`--index` 或
`--output-dir`。若用户需要保留结果，使用其指定的工作区文件，不能保存为
隐式缓存或把原始响应／输入写进支持摘要。
