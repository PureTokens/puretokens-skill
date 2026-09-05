# Native video executor commands

Resolve the executable from this Skill's absolute directory: sibling `.puretokens-executor/puretokens-api` on macOS/Linux; `.puretokens-executor/puretokens-api.exe` on Windows. Use the current host ID. All commands are one-shot; no user runtime is required.

```sh
"/absolute/skills/.puretokens-executor/puretokens-api" submit --host codex --request "/absolute/request.json"
```

```powershell
& "C:\absolute\skills\.puretokens-executor\puretokens-api.exe" submit --host codex --request "C:\absolute\request.json"
```

Create UTF-8 JSON using the host file tool; use escaped backslashes or forward slashes for Windows paths. Remove the request file after the command. Never put credentials in requests or prompts on the command line.

Generation:

```json
{"kind":"video","operation":"generate","model":"grok-imagine-video-1.5-preview","prompt":"A slow camera movement over a mountain lake","parameters":{"duration":5}}
```

Current local first-frame image:

```json
{"kind":"video","operation":"generate","media_operation":"image_to_video","model":"grok-imagine-video-1.5-preview","prompt":"A gentle camera movement","attachments":[{"field":"image","path":"/absolute/current.png"}]}
```

`submit` returns the task ID immediately; it never polls or downloads. Do not repeat a submission with unknown output. `status` reads once and `wait` performs one bounded window. Carry the original operation as `original_operation`, model, confirmed count and safe parameters into a same-task request:

```json
{"kind":"video","original_operation":"generate","task_id":"RETURNED_ID","model":"grok-imagine-video-1.5-preview","requested_count":1,"parameters":{"duration":5}}
```

After a completed same-task receipt, `content --host codex --request <file>` downloads one index:

```json
{"kind":"video","original_operation":"generate","task_id":"RETURNED_ID","model":"grok-imagine-video-1.5-preview","requested_count":1,"parameters":{"duration":5},"task_status":"completed","index":0,"output_dir":"/absolute/existing/output-directory"}
```

A video task has one output, index 0. Attach the returned `downloaded_paths` file before requesting another index. Downloading is not delivery. Existing validated same-task files in the same directory are reused.

## Optional continuation record

Set `--record <absolute-task-json>` on submit only when an explicit user/workspace artifact is useful. Choose the output location with `content --output-dir <existing-absolute-directory>`; it can also be retained from an explicit submission output_dir. The record contains task ID, kind, model, original operation, requested count, safe parameters and progress; no prompt, credentials, reference URL or media bytes. Do not edit it by hand.

```text
<executor> submit --host codex --request <request-file> --record <absolute-task-json>
<executor> resume --host codex --record <absolute-task-json>
<executor> content --host codex --record <absolute-task-json> --index 0 --output-dir <existing-absolute-output-directory>
<executor> delivered --record <absolute-task-json> --index 0
```

`status`, `wait` and `resume` accept an existing `--record` instead of `--request`; never combine both. `resume` performs a bounded wait for only the recorded task. Call `delivered` only after the host actually hands off that downloaded file. Records and output files are explicit user artifacts, retained or cleaned according to the user's chosen location. A record cannot recover an unknown submission that returned no task ID.

## Explicit validation

`preflight --host codex --request <file>` checks the requested model parameters and attachment representation without a POST. Use it only for an explicit check; normal generation validates during submit. It is not a price quote or a guarantee of permission, balance or media delivery. A profile gap may require one catalog GET.

Machine receipts preserve available context with `original_operation` separate from the invoked command. `retry_not_before` is an RFC3339 lower bound for the next same-task read; preserve it across continuation. An omitted continuation count is unknown, not one; image content requires the confirmed original count. Report `submission_outcome: unknown` as uncertainty; never automatically resubmit. Show API codes only if actually returned, and retain only sanitized error detail. Keep user-facing updates to task ID/status, actual artifact delivery or the needed corrective action.
