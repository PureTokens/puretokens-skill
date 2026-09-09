# Native image executor commands

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
{"kind":"image","operation":"generate","model":"gpt-image-2","prompt":"A mountain lake at sunrise","parameters":{"image_size":"1K"}}
```

Current local image edit:

```json
{"kind":"image","operation":"edit","media_operation":"image_edit","model":"gpt-image-2","prompt":"Change only the sky","attachments":[{"field":"image","path":"/absolute/current.png"}]}
```

`submit` returns the task ID immediately; it never polls or downloads. Do not repeat a submission with unknown output. `status` reads once and `wait` performs one bounded window. Carry the original operation as `original_operation`, model, confirmed count, safe parameters and the exact returned `retry_not_before` into a same-task request. The timestamp below is illustrative; copy the actual receipt value without recalculating it, or use `--record`:

```json
{"kind":"image","original_operation":"generate","task_id":"RETURNED_ID","model":"gpt-image-2","requested_count":1,"parameters":{"image_size":"1K"},"retry_not_before":"2026-09-06T00:00:20Z"}
```

For a newly accepted pending image, `submit` sets `retry_not_before` to 20 seconds after receipt of the accepted response, unless the API supplied a valid positive `Retry-After`. This timestamp survives tool handoff and task records, so time already spent processing the receipt is not waited again. Do not add a separate sleep. `wait` checks when the timestamp is reached, then waits 3 seconds between reads. Expired or absent timestamps (including older task records) allow an immediate first read; continuation never invents another initial wait. API `Retry-After` remains authoritative. Each window allows at most 40 status reads and 120 seconds including network time, whichever ends first; stop when the next wait cannot fit. Optional `poll.max_status_reads` and `poll.deadline_seconds` may shorten these limits. Waiting stops immediately on completion, reconciliation, unknown state or an unrecoverable status error; it never submits another task.

After a completed same-task receipt, `content --host codex --request <file>` downloads one index:

```json
{"kind":"image","original_operation":"generate","task_id":"RETURNED_ID","model":"gpt-image-2","requested_count":1,"parameters":{"image_size":"1K"},"task_status":"completed","index":0,"output_dir":"/absolute/existing/output-directory"}
```

For multiple images, preserve the accepted count and retrieve indexes 0..n-1. Attach the returned `downloaded_paths` file before requesting another index. Downloading is not delivery. Reuse requires a matching SHA-256, byte count and media type in an explicit task record. Without that proof, preserve any existing file and select another output directory for this same task index.

## Choose the next command

| Latest result | Next action |
| --- | --- |
| Accepted and pending | Return the receipt, then use `wait` (`resume` with a record); preserve the returned wait timestamp. |
| Completed | Use `content` for the confirmed index and an existing absolute output directory. |
| Downloaded | Hand the file to the user; with a record, acknowledge using `delivered` only after actual handoff. |
| Wait window ended or status failed | Keep the same ID and ask whether to continue; do not POST again. |
| Reconciliation required | Stop automatic waiting; an explicit recorded `resume` reads status once. |
| Output conflict or missing proof | Preserve files and fetch the same index into another output directory. |
| Submission unknown with no ID | Stop; records cannot discover the missing task and must not trigger another POST. |

## Optional continuation record

Choose one mode before submitting: use `--request` alone for a single-session flow, or add `--record <absolute-task-json>` for durable recovery and verified file reuse. Both modes submit exactly once; do not create a second task to switch modes. Choose the output location with `content --output-dir <existing-absolute-directory>`; it can also be retained from an explicit submission output_dir. The record contains task ID, kind, model, original operation, requested count, validated safe parameters, progress and per-index download proofs (`sha256`, `bytes`, `media_type`); no prompt, credentials, reference URL or media bytes. Do not edit it by hand.

```text
<executor> submit --host codex --request <request-file> --record <absolute-task-json>
<executor> resume --host codex --record <absolute-task-json>
<executor> content --host codex --record <absolute-task-json> --index 0 --output-dir <existing-absolute-output-directory>
<executor> delivered --record <absolute-task-json> --index 0
```

`status`, `wait` and `resume` accept an existing `--record` instead of `--request`; never combine both. `resume` performs a bounded wait for only the recorded task. Call `delivered` only after the host actually hands off that downloaded file. Records and output files are explicit user artifacts, retained or cleaned according to the user's chosen location. A record cannot recover an unknown submission that returned no task ID.

## Explicit validation

`preflight --host codex --request <file>` checks the requested model parameters and attachment representation without a POST. Use it only for an explicit check; normal generation validates during submit. It is not a price quote or a guarantee of permission, balance or media delivery. A profile gap may require one catalog GET.

Machine receipts preserve available context with `original_operation` separate from the invoked command. `retry_not_before` is an RFC3339 lower bound for the next same-task read, from the initial image delay or the API; preserve it across continuation. `retry_after_seconds` is emitted only for an actual API header, never for the executor's local initial delay. Existing `puretokens-task-v1` records remain readable. If they lack download proofs, resume the same task but fetch its content into another output directory before handoff. Never edit a digest to bypass verification. An omitted continuation count is unknown, not one; image content requires the confirmed original count. Report `submission_outcome: unknown` as uncertainty; never automatically resubmit. Show API codes only if actually returned, and retain only sanitized error detail. Keep user-facing updates to task ID/status, actual artifact delivery or the needed corrective action.

A record requiring reconciliation is an exception to bounded waiting: an explicit `resume` reads its status once, keeping the same task ID and honoring Retry-After. Continue from the new receipt; never clear the flag by hand or resubmit.

Always retain `reconciliation_required` and the exact `retry_not_before` when present. Without a record, carry these fields from the latest receipt into each same-task request. A request file is temporary; the receipt or optional task record supplies continuation metadata after that file is removed.
