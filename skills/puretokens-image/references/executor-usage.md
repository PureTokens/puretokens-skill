# Native executor commands

Resolve the executor absolute path from the loaded Skill directory: sibling `.puretokens-executor/puretokens-api` on macOS/Linux; `.puretokens-executor/puretokens-api.exe` on Windows. The process working directory is irrelevant. Use the current host ID, not a provider label.

macOS/Linux example (substitute actual absolute paths):

```sh
"/absolute/skills/.puretokens-executor/puretokens-api" submit --host codex --request "/absolute/request.json"
```

Windows PowerShell example:

```powershell
& "C:\absolute\skills\.puretokens-executor\puretokens-api.exe" submit --host codex --request "C:\absolute\request.json"
```

Create the UTF-8 request using the host file tool. Do not put credentials in it. Remove the request file after the command completes. JSON input is finite; no interactive stdin, heredoc improvisation, open terminal pipe or shell-escaped prompt is needed.

```json
{"kind":"image","operation":"generate","model":"gpt-image-2","prompt":"A mountain lake at sunrise","parameters":{"image_size":"1K"}}
```

Local image edit:

```json
{"kind":"image","operation":"edit","media_operation":"image_edit","model":"gpt-image-2","prompt":"Change only the sky","attachments":[{"field":"image","path":"/absolute/current.png"}]}
```

Reference-image video:

```json
{"kind":"video","operation":"generate","media_operation":"reference_image_video","model":"seedance-2.0-mini","prompt":"A gentle camera movement preserving the character","attachments":[{"field":"reference_images","path":"/absolute/current.png"}]}
```

`submit` returns one immediate JSON receipt. It never polls or downloads. Legacy `task` is a compatibility alias with the same immediate-submit behavior. Do not issue another submit after a started request with missing output.

`status` reads once; `wait` performs one bounded window. Use the returned task_id:

```json
{"kind":"image","task_id":"RETURNED_ID","model":"gpt-image-2","requested_count":1}
```

`content` downloads one index, only after a same-task completed status:

```json
{"kind":"image","task_id":"RETURNED_ID","task_status":"completed","requested_count":1,"index":0,"output_dir":"/absolute/existing/output-directory"}
```

For Windows paths, use JSON-escaped backslashes or forward slashes. Attach the returned downloaded_paths file to the user before retrieving the next index. It is not delivered merely because it exists locally. A same-task validated file in the same output directory is reused without another content request. Output files are user artifacts, not a hidden cache; preserve or clean them according to the user's chosen output location.

Receipts preserve model/task identity on failure. `submission_outcome: unknown` never authorizes resubmission. `api_error_code` is present only when returned by the API; local classifications use `local_error_code`. `init` checks public identity then one authenticated catalog response without creating a media task. `connection` checks public identity only. `balance` reads two bearer billing endpoints and reports API display units, without inventing a currency, unlimited status or account scope.
