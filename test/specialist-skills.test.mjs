import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdir, mkdtemp, readFile, readdir, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { promisify } from "node:util";
import { checkMediaCatalogFreshness } from "../scripts/check-media-catalog-freshness.mjs";
import { collectSkillRecords, repositoryRoot, validateRepository } from "../scripts/skill-registry.mjs";
import { sourceRows, buildPublishedCatalog } from "../scripts/sync-media-model-catalog-from-service.mjs";
import { buildSelection } from "../scripts/sync-skill-model-selection.mjs";
import {
  parseCodexConnection,
  parseDotEnv,
  parseGrokModelEntries,
  parseArguments,
  parseWorkBuddyModelEntries,
  readJsonBase64Argument,
  resolveClaudeCodeCredential,
  resolveCodexCredential,
  resolveConfiguredCredential,
  resolveGeminiCliCredential,
  resolveOpenCodeCredential,
  resolveWorkBuddyCredential
} from "../runtime/puretokens-direct-api.mjs";

const names = ["puretokens-balance", "puretokens-connection", "puretokens-models", "puretokens-image", "puretokens-video", "puretokens-update"];

test("registry exposes exactly the six specialist Skills", async () => {
  const { registry, records } = await collectSkillRecords();
  assert.deepEqual(registry.skills.map((skill) => skill.name), names);
  assert.equal(records.length, 6);
  assert.deepEqual([...new Set(records.map((record) => record.manifest.version))], ["0.13.15"]);
  assert.deepEqual(await validateRepository(), []);
});

test("specialist policies use the fixed direct API without MCP fallback", async () => {
  const [balance, connection, models, image, video] = await Promise.all(names.map((name) => readFile(path.join(repositoryRoot, "skills", name, "SKILL.md"), "utf8")));
  assert.match(connection, /GET https:\/\/api\.puretokensx\.com\/v1/);
  assert.match(connection, /Claude Code、Codex、WorkBuddy、Gemini CLI、Grok Build 和 OpenCode/);
  assert.match(connection, /--host claude-code/);
  assert.match(connection, /--host opencode/);
  assert.match(models, /GET https:\/\/api\.puretokensx\.com\/v1\/media\/models/);
  assert.match(models, /不得调用 Images\/Videos 提交、任务状态、内容、余额或其他路径/);
  assert.match(models, /`input_schema\.constraints`/);
  assert.match(models, /勾选包含该模型的分组/);
  assert.match(image, /POST https:\/\/api\.puretokensx\.com\/v1\/images\/generations/);
  assert.match(image, /未指定模型时，普通文生图、当前请求明确附带的本地图片参考和本地图片编辑都优先使用已安装选择中的 `gpt-image-2`/);
  assert.match(image, /本地图片附件.*`image_edit` multipart operation/s);
  assert.match(image, /`requires_together`/);
  assert.match(image, /`size_expression_precedence`/);
  assert.match(image, /`200cm × 230cm`/);
  assert.match(video, /POST https:\/\/api\.puretokensx\.com\/v1\/videos/);
  assert.match(video, /`grok-imagine-video-1.5-preview`/);
  assert.match(video, /`minimax-h3`/);
  assert.match(video, /勾选包含该精确模型的分组/);
  assert.match(video, /首帧\/第一帧.*`image_to_video`/s);
  assert.match(video, /`generate_audio`/);
  const policy = balance + connection + models + image + video;
  assert.match(policy, /https:\/\/api\.puretokensx\.com/);
  assert.doesNotMatch(policy, /internal-upstream|当前连接不是 Pure Tokens|无法确认归属/i);
  assert.match(policy, /\.puretokens-runtime\/puretokens-direct-api\.mjs/);
  assert.match(policy, /--multipart-stdin/);
  assert.match(image, /WorkBuddy 的 Bash 不可靠地传递并关闭标准输入/);
  assert.match(image, /--json-base64 <值>/);
  assert.match(video, /--multipart-base64 <值>/);
  assert.match(video, /原生媒体请求绝不能退化为只含 `model`、`prompt` 的 JSON 文生视频/);
  assert.match(video, /seedance-2\.0-mini.*reference_image_video/s);
  assert.doesNotMatch(policy, /自动携带认证/);
});

test("bilingual READMEs include safe copyable Agent installation prompts", async () => {
  const [english, chinese] = await Promise.all([
    readFile(path.join(repositoryRoot, "README.md"), "utf8"),
    readFile(path.join(repositoryRoot, "README.zh-CN.md"), "utf8")
  ]);
  assert.match(english, /^<p align="center">\n  <img src="\.\/assets\/brand\/puretokens-skill-hero\.png" alt="Pure Tokens Official Skills" width="100%" \/>\n<\/p>/);
  assert.match(english, /## Agent-assisted installation/);
  assert.match(english, /^### Copy this to a terminal-capable local agent\n\n```text$/m);
  assert.equal((english.match(/^### Copy this to a terminal-capable local agent$/gm) ?? []).length, 1);
  const englishPrompt = english.match(/^### Copy this to a terminal-capable local agent\n\n```text\n([\s\S]*?)\n```$/m);
  assert.ok(englishPrompt, "English client-download prompt must be the first bounded text block under the fixed heading");
  assert.equal(englishPrompt[1], "Install or update the official Pure Tokens Skills from https://github.com/PureTokens/puretokens-skill.");
  assert.match(english, /## Model access groups/);
  assert.match(english, /select a group containing that exact model/);
  assert.match(chinese, /## 让 Agent 协助安装/);
  assert.match(chinese, /^### 直接复制给具备本机终端的 Agent\n\n```text$/m);
  assert.equal((chinese.match(/^### 直接复制给具备本机终端的 Agent$/gm) ?? []).length, 1);
  const chinesePrompt = chinese.match(/^### 直接复制给具备本机终端的 Agent\n\n```text\n([\s\S]*?)\n```$/m);
  assert.ok(chinesePrompt, "Chinese client-download prompt must be the first bounded text block under the fixed heading");
  assert.equal(chinesePrompt[1], "请从 https://github.com/PureTokens/puretokens-skill 安装或更新官方 Pure Tokens Skills。");
  assert.match(chinese, /^<p align="center">\n  <img src="\.\/assets\/brand\/puretokens-skill-hero\.png" alt="Pure Tokens 官方 Skills" width="100%" \/>\n<\/p>/);
  assert.match(chinese, /## 模型访问分组/);
  assert.match(chinese, /勾选包含该精确模型的分组/);
});

test("specialist manifests use the fixed direct API with a narrow configured-credential resolver", async () => {
  for (const name of names) {
    const manifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", name, "skill.json"), "utf8"));
    assert.equal(manifest.rules.doesNotUseMcpOrFallbackTransport, true);
    if (name === "puretokens-update") {
      assert.equal(manifest.rules.usesOfficialMainBranch, true);
      assert.equal(manifest.rules.usesManagedSkillSync, true);
      assert.equal(manifest.rules.neverOverwritesUnmanagedDirectories, true);
      assert.equal(manifest.rules.removesVerifiedRetiredSkills, true);
      assert.equal(manifest.rules.reportsSynchronizedVersion, true);
      assert.equal(manifest.rules.doesNotReadCredentialsOrHostConfiguration, true);
    } else {
      assert.equal(manifest.rules.usesFixedPureTokensApiOrigin, true);
      assert.equal(manifest.rules.usesFullApiUrls, true);
      assert.equal(manifest.rules.usesRuntimeManagedAuthentication, true);
      assert.equal(manifest.rules.usesNarrowConfiguredCredentialResolver, true);
      assert.equal(manifest.rules.neverExposesCredentialsOrHostConfiguration, true);
    }
  }
});

test("installed model selections are generated from the published media catalog", async () => {
  const catalog = JSON.parse(await readFile(path.join(repositoryRoot, "references", "media-model-catalog.json"), "utf8"));
  for (const [skill, capability] of [["puretokens-image", "image"], ["puretokens-video", "video"]]) {
    const selection = JSON.parse(await readFile(path.join(repositoryRoot, "skills", skill, "references", "model-selection.json"), "utf8"));
    const expected = catalog.models
      .filter((model) => model.capabilities.includes(capability))
      .map((model) => ({ id: model.id, aliases: model.aliases, ...(model.parameterSchema ? { parameterSchema: model.parameterSchema } : {}) }));
    assert.equal(selection.schemaVersion, 1);
    assert.equal(selection.$schema, "https://puretokensx.com/schemas/model-selection.schema.json");
    assert.equal(selection.capability, capability);
    assert.equal(selection.catalogCapturedAt, catalog.serviceCatalog.capturedAt);
    assert.deepEqual(selection.models, expected);
  }
});

test("gpt-image-2 publishes its declared native image-edit operation", async () => {
  const selection = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-image", "references", "model-selection.json"), "utf8"));
  const gptImage2 = selection.models.find((model) => model.id === "gpt-image-2");
  assert.ok(gptImage2, "gpt-image-2 must remain in the installed image selection");
  assert.deepEqual(gptImage2.parameterSchema.operations.image_edit, {
    request: { method: "POST", path: "/v1/images/edits", contentType: "multipart/form-data" },
    requiredBodyFields: ["model", "prompt", "image"],
    inputs: {
      image: {
        field: "image",
        required: true,
        minItems: 1,
        maxItems: 6,
        transports: ["multipart_file"]
      }
    }
  });
});

test("installed contracts cover bounded requests, task recovery, and user-facing failure guidance", async () => {
  const requiredScenarios = {
    "puretokens-balance": ["balance-account-session-unavailable", "balance-response"],
    "puretokens-connection": ["connection-identity-confirmed", "connection-identity-unconfirmed", "connection-identity-unavailable", "connection-base-url-request", "connection-identity-assurance"],
    "puretokens-models": ["models-catalog-unavailable", "models-catalog-empty", "models-exact-id-unavailable", "models-capability-filter-empty", "models-parameter-profile-absent", "models-operation-profile-absent", "models-requirement-ambiguous", "models-catalog-response"],
    "puretokens-image": ["image-model-alias-ambiguous", "image-normal-submission-skips-catalog", "image-failure-receipt-safe", "image-content-safety-failure", "image-input-role-ambiguous", "image-request-scope-boundary", "image-distinct-assets", "image-prompt-shaping", "image-submission-model-parameter-rejected", "image-submission-rate-limited", "image-submission-outcome-unknown", "image-submission-runtime-output-unknown", "image-unknown-submission-continuation-without-id", "image-model-unavailable", "image-model-parameter-profile-unavailable", "image-model-parameter-unsupported", "image-size-expression-constraints", "image-catalog-on-demand-unavailable", "image-execution-unavailable", "image-count-invalid", "image-pixel-size-invalid", "image-physical-size", "image-edit-profile-unavailable", "image-public-url-reference-profile-unavailable", "image-public-url-edit-profile-unavailable", "image-public-url-reference-invalid", "image-public-url-reference-not-public", "image-public-url-reference-ambiguous", "image-native-reference-attachment-direct-api", "image-native-reference-attachment-unavailable", "image-edit-attachment-unavailable", "image-edit-attachment-direct-api", "image-edit-input-unsupported", "image-task-id-normalization", "image-task-id-missing", "image-task-id-invalid", "image-task-state-unrecognized", "image-task-reconciliation-required", "image-task-request-count-unknown", "image-task-pending", "image-task-poll-delay", "image-task-poll-rate-limited", "image-task-poll-resource-bound", "image-task-polling-deadline", "image-task-explicit-continuation", "image-task-terminal-failure", "image-task-timeout-or-unknown", "image-content-delivery-failure", "image-content-delivery-location", "image-delivery-terminal-boundary", "image-content-index-missing", "image-content-resource-bound"],
    "puretokens-video": ["video-model-alias-ambiguous", "video-normal-submission-skips-catalog", "video-failure-receipt-safe", "video-content-safety-failure", "video-submission-model-parameter-rejected", "video-submission-rate-limited", "video-submission-outcome-unknown", "video-submission-runtime-output-unknown", "video-unknown-submission-continuation-without-id", "video-model-unavailable", "video-model-parameter-profile-unavailable", "video-model-parameter-unsupported", "video-catalog-on-demand-unavailable", "video-execution-unavailable", "video-parameter-unsupported", "video-first-frame-operation-mapping", "video-generate-audio-intent-mapping", "video-resolution-mode-unsupported", "video-image-operation-profile-unavailable", "video-media-operation-profile-unavailable", "video-image-transport-unavailable", "video-native-attachment-route-locked", "video-native-attachment-path-unavailable", "video-public-url-reference-profile-unavailable", "video-public-url-reference-invalid", "video-public-url-reference-not-public", "video-public-url-reference-ambiguous", "video-prompt-requirement", "video-attachment-direct-api", "video-image-count-unsupported", "video-media-combination-unsupported", "video-input-media-unsupported", "video-task-id-normalization", "video-task-id-missing", "video-task-id-invalid", "video-task-state-unrecognized", "video-task-reconciliation-required", "video-task-pending", "video-task-poll-delay", "video-task-poll-rate-limited", "video-task-poll-resource-bound", "video-task-polling-deadline", "video-task-explicit-continuation", "video-task-terminal-failure", "video-task-timeout-or-unknown", "video-content-delivery-failure", "video-content-resource-bound"],
    "puretokens-update": ["update-host-unknown", "update-terminal-unavailable", "update-native-installer-bootstrap", "update-validation-failed", "update-unmanaged-conflict", "update-managed-retired-skill", "update-managed-sync"]
  };
  for (const name of names) {
    const root = path.join(repositoryRoot, "skills", name, "references");
    const contract = JSON.parse(await readFile(path.join(root, "execution-contract.json"), "utf8"));
    const scenarios = JSON.parse(await readFile(path.join(root, "behavior-scenarios.json"), "utf8"));
    assert.equal(contract.schemaVersion, 1);
    assert.equal(contract.$schema, "https://puretokensx.com/schemas/media-execution-contract.schema.json");
    assert.equal(scenarios.schemaVersion, 1);
    assert.equal(scenarios.$schema, "https://puretokensx.com/schemas/media-behavior-scenarios.schema.json");
    assert.equal(contract.result.neverAutoResubmit ?? false, ["puretokens-image", "puretokens-video"].includes(name));
    const ids = new Set(scenarios.scenarios.map((scenario) => scenario.id));
    for (const id of requiredScenarios[name]) assert.ok(ids.has(id), `${name} missing ${id}`);
    if (["puretokens-image", "puretokens-video"].includes(name)) {
      const unavailable = scenarios.scenarios.find((scenario) => scenario.id === `${name.replace("puretokens-", "")}-model-unavailable`);
      assert.match(unavailable.then, /select a group that contains that exact model/);
      assert.match(unavailable.then, /create or select a managed key covering the selected groups/);
      assert.match(unavailable.then, /Do not claim which group contains the model/);
      const runtimeOutputUnknown = scenarios.scenarios.find((scenario) => scenario.id === `${name.replace("puretokens-", "")}-submission-runtime-output-unknown`);
      assert.match(runtimeOutputUnknown.then, /do not invoke POST again, submit a replacement task/);
      assert.match(runtimeOutputUnknown.then, /claim the task was not created, charged, or refunded/);
      assert.match(runtimeOutputUnknown.then, /end the current response without further tool, status, polling, or deliberation work/);
      const noIdContinuation = scenarios.scenarios.find((scenario) => scenario.id === `${name.replace("puretokens-", "")}-unknown-submission-continuation-without-id`);
      assert.match(noIdContinuation.then, /explicitly confirm a new paid/);
      assert.match(noIdContinuation.then, /do not submit until that confirmation arrives in a new user turn/);
    }
    if (name === "puretokens-models") {
      const unavailable = scenarios.scenarios.find((scenario) => scenario.id === "models-exact-id-unavailable");
      assert.match(unavailable.then, /select a group containing that exact model/);
      assert.match(unavailable.then, /Do not infer an alias, substitute another model, or claim which group contains the model/);
    }
  }
  const image = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-image", "references", "execution-contract.json"), "utf8"));
  assert.equal(image.operations.submit.conditionalBodyFields.all_user_optional_fields, "only_when_the_installed_exact_model_selection_or_an_on_demand_live_input_schema_declares_the_field_and_value");
  assert.equal(image.operations.submit.fixedBody.async, true);
  assert.equal(image.operations.catalog.url, "https://api.puretokensx.com/v1/media/models");
  assert.deepEqual(image.operations.edit.allowedPaths, ["/v1/images/generations", "/v1/images/edits"]);
  assert.equal(image.operations.edit.pathSource, "installed_model_selection_or_on_demand_live_profile_operation_path");
  assert.equal(image.operations.edit.contentType, "profile_declared_json_or_multipart");
  assert.equal(image.operations.edit.requiresDeclaredProfileOperation, "image_edit");
  assert.deepEqual(image.operations.edit.requiredBodyFields, ["model", "prompt", "async"]);
  assert.equal(image.inputMediaValidation.publicUrlReferenceRequires, "installed_model_selection_or_on_demand_live_profile_property_and_declared_reference_transport");
  assert.equal(image.inputMediaValidation.publicUrlEditRequires, "installed_model_selection_or_on_demand_live_profile_image_edit_operation_with_application_json_exact_field_and_public_url_transport");
  assert.equal(image.inputMediaValidation.nativeAttachmentRequires, "installed_model_selection_or_on_demand_live_profile_image_edit_operation_for_current_visual_reference_or_explicit_edit");
  assert.equal(image.inputMediaValidation.imageEditOperation, "image_edit");
  assert.equal(image.inputMediaValidation.nativeVisualReferenceUsesImageEditOperation, true);
  assert.equal(image.inputMediaValidation.nativeVisualReferencePromptSemantics, "preserve_reference_intent_while_using_the_declared_image_edit_multipart_transport");
  assert.equal(image.inputMediaValidation.onlyCurrentUserExplicitMedia, true);
  assert.equal(image.inputMediaValidation.skillPassesUserProvidedPublicUrlsWithoutDownloadingOrRehosting, true);
  assert.equal(image.inputMediaValidation.skillNeverDownloadsOrRehostsMedia, true);
  assert.equal(image.inputMediaValidation.gatewayStagesCurrentRequestAttachment, true);
  assert.equal(image.inputMediaValidation.requiresRuntimeNativeAttachmentByteDelivery, true);
  assert.equal(image.inputMediaValidation.nativeReferenceAttachmentRequires, "installed_model_selection_or_on_demand_profile_image_edit_operation_with_multipart_file_transport");
  assert.equal(image.inputMediaValidation.untransportableNativeReferenceMustNeverBeConvertedToTextPrompt, true);
  assert.equal(image.inputMediaValidation.whenNativeReferenceTransportUnavailable, "stop_before_post_when_no_declared_image_edit_multipart_operation_or_no_native_byte_delivery_and_require_public_https_url_or_explicit_new_text_only_request");
  assert.equal(image.inputMediaValidation.requestScope, "current_user_request_only_no_unrelated_skill_history_workspace_search_or_quality_review");
  assert.equal(image.result.sameTaskOnly, true);
  assert.equal(image.result.neverAutoResubmit, true);
  assert.equal(image.result.afterConfirmedDelivery, "return_one_completion_receipt_and_end_turn_without_quality_review_history_search_or_new_media_submission");
  assert.equal(image.result.deliveryReceipt, "report_a_local_path_or_attachment_only_when_confirmed_by_the_runtime_otherwise_report_native_byte_handoff_without_inventing_a_location");
  assert.deepEqual(image.failureReceipt.requiredFields, ["failure_phase", "api_error_code", "http_status", "error_message", "next_action"]);
  assert.deepEqual(image.failureReceipt.allowedFailurePhases, ["validation", "submission", "status", "content"]);
  assert.equal(image.failureReceipt.httpStatus, "numeric_when_returned_otherwise_not_returned");
  assert.equal(image.failureReceipt.apiErrorCode, "exact_explicit_public_api_error_code_otherwise_not_returned");
  assert.equal(image.failureReceipt.errorMessage, "safe_user_facing_summary_using_only_sanitized_public_api_detail_or_a_fixed_local_explanation");
  assert.equal(image.failureReceipt.retryAfterSeconds, "include_only_for_http_429_when_a_valid_positive_retry_after_is_returned");
  assert.deepEqual(image.failureReceipt.forbiddenDisclosure, ["raw_response_body", "upstream_or_provider_identifier", "internal_hostname_or_url", "stack_trace", "request_headers", "request_body", "credentials_or_session_data", "user_reference_url_or_attachment_bytes"]);
  assert.deepEqual(image.transport, {
    fixedApiOrigin: "https://api.puretokensx.com",
    usesFullApiUrls: true,
    usesRuntimeManagedAuthentication: true,
    usesNarrowConfiguredCredentialResolver: true,
    verifiedManagedRuntimeHosts: ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode"],
    neverExposesCredentialsOrHostConfiguration: true,
    requiresNativeMediaByteDelivery: true,
    doesNotUseMcpOrFallbackTransport: true,
    workBuddyPostBodyTransport: "bounded_base64_argument_never_stdin"
  });
  assert.equal(image.parameterValidation.normalSubmissionUsesInstalledModelSelection, true);
  assert.equal(image.parameterValidation.exactModelCoreSubmissionDoesNotRequireCatalogPreflight, true);
  assert.equal(image.parameterValidation.onDemandLiveCatalogRead, "only_for_explicit_discovery_installed_profile_gap_or_post_rejection_diagnosis");
  assert.equal(image.parameterValidation.allOptionalParametersRequire, "installed_model_selection_parameter_schema_or_on_demand_live_input_schema");
  assert.equal(image.parameterValidation.enforcesDeclaredRequiresTogether, true);
  assert.equal(image.parameterValidation.sizeExpressionPrecedence, "send_only_the_highest_precedence_user_supplied_declared_size_expression");
  assert.deepEqual(image.taskIdentity.fallbackTopLevelResponseFields, ["task_id", "id"]);
  assert.equal(image.taskIdentity.neverDerivesIdFromUrlsNestedObjectsOrPrompt, true);
  assert.deepEqual(image.taskState.fallbackPendingStates, ["pending", "queued", "running", "in_progress"]);
  assert.equal(image.taskState.reconciliationRequiredField, "reconciliation_required");
  assert.equal(image.taskState.reconciliationPrecedesLifecycle, true);
  assert.equal(image.taskState.whenReconciliationRequired, "stop_ordinary_polling_retain_same_task_id_do_not_submit_replacement_or_infer_refund");
  assert.equal(image.submissionFailure.rateLimit, "report_retry_after_when_returned_and_require_explicit_retry");
  assert.equal(image.submissionFailure.runtimeInvocationOutputUnknown, "treat_submission_outcome_as_unknown_do_not_repeat_post_or_submit_replacement_task");
  assert.equal(image.submissionFailure.runtimeUnknownOutcomeResponse, "one_terminal_receipt_then_end_turn_without_followup_tool_status_or_poll_work");
  assert.equal(image.submissionFailure.noTaskIdLaterContinuation, "require_explicit_confirmation_of_a_new_billable_request_before_one_new_post");
  assert.equal(image.contentRetrieval.indexBase, 0);
  assert.equal(image.contentRetrieval.allowedIndexes, "0..requestedCount-1");
  assert.equal(image.contentRetrieval.completeDeliveryRequiresEveryRequestedIndex, true);
  assert.equal(image.contentRetrieval.fetchOnlyAfterTerminalSuccess, true);
  assert.equal(image.contentRetrieval.sequentialByIndex, true);
  assert.equal(image.contentRetrieval.oneInFlightContentReadPerTask, true);
  assert.equal(image.contentRetrieval.neverPrefetchOrRefetchDeliveredContent, true);
  assert.equal(image.contentRetrieval.handoffBeforeNextContentRead, true);
  assert.deepEqual(image.polling.fallbackDelaysSeconds, [3, 6, 12, 24, 30]);
  assert.equal(image.polling.serverDelay, "honor_valid_positive_retry_after_before_fallback");
  assert.equal(image.polling.serverDelayMustFitRemainingAutomaticDeadline, true);
  assert.equal(image.polling.afterHttp429, "honor_valid_positive_retry_after_then_continue_same_task_if_remaining_budget");
  assert.equal(image.polling.steadyDelaySeconds, 30);
  assert.equal(image.polling.automaticDeadlineSeconds, 120);
  assert.equal(image.polling.maxAutomaticStatusReads, 6);
  assert.equal(image.polling.oneInFlightStatusReadPerTask, true);
  assert.equal(image.polling.automaticPollingScope, "submission_or_explicit_same_task_continuation_turn_only_no_background_timer_or_queue");
  assert.equal(image.polling.afterStatusReadError, "stop_on_5xx_transport_or_timeout_and_require_explicit_same_task_continuation");
  assert.equal(image.polling.explicitContinuation, "new_bounded_same_task_polling_window_only_when_explicitly_requested");
  assert.equal(image.polling.afterDeadline, "report_pending_and_require_explicit_same_task_continuation");
  const video = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-video", "references", "execution-contract.json"), "utf8"));
  assert.equal(video.operations.catalog.url, "https://api.puretokensx.com/v1/media/models");
  assert.equal(video.operations.submit.url, "https://api.puretokensx.com/v1/videos");
  assert.deepEqual(video.operations.submit.requiredBodyFields, ["model"]);
  assert.equal(video.operations.edit.url, "https://api.puretokensx.com/v1/videos/edits");
  assert.equal(video.operations.edit.requiresDeclaredProfileOperation, "video_edit");
  assert.equal(video.result.sameTaskOnly, true);
  assert.equal(video.result.neverAutoResubmit, true);
  assert.deepEqual(video.failureReceipt, image.failureReceipt);
  assert.equal(video.parameterValidation.normalSubmissionUsesInstalledModelSelection, true);
  assert.equal(video.parameterValidation.exactModelCoreSubmissionDoesNotRequireCatalogPreflight, true);
  assert.equal(video.parameterValidation.onDemandLiveCatalogRead, "only_for_explicit_discovery_installed_profile_gap_or_post_rejection_diagnosis");
  assert.equal(video.parameterValidation.allModelsOptionalParametersRequire, "installed_model_selection_parameter_schema_or_on_demand_live_input_schema");
  assert.equal(video.parameterValidation.promptRequirementUsesInstalledOrOnDemandPropertiesAndConstraints, true);
  assert.equal(video.parameterValidation.resolutionUsesInstalledOrOnDemandModeConstraint, true);
  assert.deepEqual(video.taskIdentity.fallbackTopLevelResponseFields, ["task_id", "id"]);
  assert.equal(video.taskIdentity.neverDerivesIdFromUrlsNestedObjectsOrPrompt, true);
  assert.deepEqual(video.taskState.fallbackPendingStates, ["pending", "queued", "running", "in_progress"]);
  assert.equal(video.taskState.reconciliationRequiredField, "reconciliation_required");
  assert.equal(video.taskState.reconciliationPrecedesLifecycle, true);
  assert.equal(video.taskState.whenReconciliationRequired, "stop_ordinary_polling_retain_same_task_id_do_not_submit_replacement_or_infer_refund");
  assert.equal(video.submissionFailure.rateLimit, "report_retry_after_when_returned_and_require_explicit_retry");
  assert.equal(video.submissionFailure.runtimeInvocationOutputUnknown, "treat_submission_outcome_as_unknown_do_not_repeat_post_or_submit_replacement_task");
  assert.equal(video.submissionFailure.runtimeUnknownOutcomeResponse, "one_terminal_receipt_then_end_turn_without_followup_tool_status_or_poll_work");
  assert.equal(video.submissionFailure.noTaskIdLaterContinuation, "require_explicit_confirmation_of_a_new_billable_request_before_one_new_post");
  assert.equal(video.inputMediaValidation.imageToVideoOperation, "image_to_video");
  assert.equal(video.inputMediaValidation.explicitFirstFrameIntentUses, "image_to_video");
  assert.equal(video.inputMediaValidation.declaredGenerateAudioIntent, "map_explicit_generated_sound_to_true_and_silence_to_false_only_for_declared_boolean_generate_audio");
  assert.equal(video.inputMediaValidation.referenceImageVideoOperation, "reference_image_video");
  assert.equal(video.inputMediaValidation.referenceVideoOperation, "reference_video");
  assert.equal(video.inputMediaValidation.referenceAudioOperation, "reference_audio");
  assert.equal(video.inputMediaValidation.videoEditOperation, "video_edit");
  assert.equal(video.inputMediaValidation.publicUrlOrDeclaredIdReferenceRequires, "installed_model_selection_or_on_demand_live_profile_property_and_declared_reference_transport");
  assert.deepEqual(video.inputMediaValidation.allowedTransports, ["multipart_file"]);
  assert.equal(video.inputMediaValidation.onlyCurrentUserExplicitMedia, true);
  assert.equal(video.inputMediaValidation.skillPassesUserProvidedPublicUrlsOrDeclaredIdsWithoutDownloadingOrRehosting, true);
  assert.equal(video.inputMediaValidation.skillNeverDownloadsOrRehostsMedia, true);
  assert.equal(video.inputMediaValidation.gatewayStagesCurrentRequestAttachment, true);
  assert.equal(video.inputMediaValidation.nativeAttachmentRouteMustMatchDeclaredOperation, true);
  assert.equal(video.inputMediaValidation.nativeAttachmentMustNeverFallBackToJsonTextSubmission, true);
  assert.equal(video.inputMediaValidation.workBuddyNativeAttachmentBodyMode, "multipart_base64_descriptor_only");
  assert.equal(video.inputMediaValidation.whenCurrentAttachmentPathUnavailable, "stop_before_post_and_explain_declared_multipart_requirement");
  assert.equal(video.inputMediaValidation.multipleNativeAttachmentTypesRequireOneDeclaredCombinedOperation, true);
  assert.equal(video.inputMediaValidation.multiplePublicUrlOrIdFieldsRequireNoDeclaredConflict, true);
  assert.equal(video.contentRetrieval.fetchOnlyAfterTerminalSuccess, true);
  assert.equal(video.contentRetrieval.oneInFlightContentReadPerTask, true);
  assert.equal(video.contentRetrieval.neverPrefetchOrRefetchDeliveredContent, true);
  assert.equal(video.contentRetrieval.handoffBeforeNextContentRead, true);
  assert.deepEqual(video.polling.fallbackDelaysSeconds, [5, 10, 20, 40, 60]);
  assert.equal(video.polling.serverDelay, "honor_valid_positive_retry_after_before_fallback");
  assert.equal(video.polling.serverDelayMustFitRemainingAutomaticDeadline, true);
  assert.equal(video.polling.afterHttp429, "honor_valid_positive_retry_after_then_continue_same_task_if_remaining_budget");
  assert.equal(video.polling.steadyDelaySeconds, 60);
  assert.equal(video.polling.automaticDeadlineSeconds, 300);
  assert.equal(video.polling.maxAutomaticStatusReads, 7);
  assert.equal(video.polling.oneInFlightStatusReadPerTask, true);
  assert.equal(video.polling.automaticPollingScope, "submission_or_explicit_same_task_continuation_turn_only_no_background_timer_or_queue");
  assert.equal(video.polling.afterStatusReadError, "stop_on_5xx_transport_or_timeout_and_require_explicit_same_task_continuation");
  assert.equal(video.polling.explicitContinuation, "new_bounded_same_task_polling_window_only_when_explicitly_requested");
  assert.equal(video.polling.afterDeadline, "report_pending_and_require_explicit_same_task_continuation");

  const balance = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-balance", "references", "execution-contract.json"), "utf8"));
  assert.equal(balance.operations.read.url, "https://api.puretokensx.com/api/product/desktop/account/balance");
  assert.equal(balance.operations.read.requiresExistingAuthenticatedAccountSession, true);
  assert.equal(balance.operations.read.responseSchema, "https://puretokensx.com/schemas/balance-snapshot.schema.json");

  const connection = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-connection", "references", "execution-contract.json"), "utf8"));
  assert.equal(connection.kind, "connection");
  assert.equal(connection.operations.identity.url, "https://api.puretokensx.com/v1");
  assert.deepEqual(connection.result, {
    identitySource: "api_declared_public_health",
    expectedStatus: "ok",
    expectedName: "Pure Tokens API",
    expectedBasePath: "/v1",
    neverRetry: true,
    doesNotExposeHostConfiguration: true,
    unconfirmedDoesNotProveOtherService: true
  });

  const models = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-models", "references", "execution-contract.json"), "utf8"));
  assert.equal(models.kind, "models");
  assert.equal(models.operations.catalog.url, "https://api.puretokensx.com/v1/media/models");
  assert.equal(models.result.reportOnlyAuthenticatedCatalog, true);
  assert.equal(models.result.doesNotSubmitMediaTasks, true);
  assert.equal(models.result.neverRetry, true);
  assert.equal(models.result.noStaticCatalogFallback, true);
  assert.equal(models.result.compatibilityShortlistsRequireDeclaredCapabilityAndInputSchema, true);

  const update = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-update", "references", "execution-contract.json"), "utf8"));
  assert.equal(update.kind, "update");
  assert.equal(update.operations.sync.commandTemplate, "native_platform_installer sync --target <installation-root>");
  assert.equal(update.operations.sync.validationCommand, "native_installer_downloads_official_main_and_performs_static_validation_before_sync");
  assert.equal(update.operations.sync.sourceBranch, "main");
  assert.equal(update.result.neverOverwritesUnmanagedDirectories, true);
  assert.equal(update.result.removesVerifiedRetiredManagedSkills, true);
  assert.equal(update.result.reportsSynchronizedVersion, true);
  assert.equal(update.result.neverModifiesOfficialCheckout, true);
  const updateScenarios = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-update", "references", "behavior-scenarios.json"), "utf8"));
  assert.match(updateScenarios.scenarios.find((scenario) => scenario.id === "update-validation-failed").then, /Do not install Node, npm packages, Git, generators, docs sync, formatters, repair commands, Git writes/);
  const updateSkill = await readFile(path.join(repositoryRoot, "skills", "puretokens-update", "SKILL.md"), "utf8");
  assert.match(updateSkill, /不能要求用户安装 Node、npm、包管理器、Git 或开发环境/);
  assert.match(updateSkill, /puretokens-skill-install\.sh/);
  assert.match(updateSkill, /puretokens-skill-install\.ps1/);
  assert.match(updateSkill, /不得把远程内容直接管道给 Shell\/PowerShell/);
});

test("distribution matrix and fixed direct API contract match every specialist manifest", async () => {
  const hostSupport = JSON.parse(await readFile(path.join(repositoryRoot, "references", "host-support.json"), "utf8"));
  assert.equal(hostSupport.$schema, "https://puretokensx.com/schemas/host-support.schema.json");
  const supported = hostSupport.supported.map((host) => host.id).sort();
  assert.deepEqual(supported, ["claude-code", "codex", "gemini-cli", "grok-build", "opencode", "trae", "workbuddy"]);
  assert.deepEqual(Object.fromEntries(hostSupport.supported.map((host) => [host.id, host.globalSkillDirectory])), {
    "claude-code": "~/.claude/skills",
    codex: "~/.agents/skills",
    workbuddy: "~/.workbuddy/skills",
    "gemini-cli": "~/.gemini/skills",
    "grok-build": "~/.grok/skills",
    opencode: "~/.config/opencode/skills",
    trae: "~/.trae/skills"
  });
  assert.deepEqual(Object.fromEntries(hostSupport.supported.map((host) => [host.id, host.directMediaExecution])), {
    "claude-code": "verified-managed-local-runtime",
    codex: "verified-managed-local-runtime",
    workbuddy: "verified-managed-local-runtime",
    "gemini-cli": "verified-managed-local-runtime",
    "grok-build": "verified-managed-local-runtime",
    opencode: "verified-managed-local-runtime",
    trae: "manual-credential-setup"
  });
  assert.equal(hostSupport.directApiOrigin, "https://api.puretokensx.com");
  const directContract = JSON.parse(await readFile(path.join(repositoryRoot, "references", "direct-api-execution-contract.json"), "utf8"));
  assert.equal(directContract.apiOrigin, "https://api.puretokensx.com");
  assert.deepEqual(directContract.authentication, {
    usesConfiguredConnectionCredential: true,
    managedRuntimeHosts: ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode"],
    credentialUse: "in_memory_only_for_fixed_puretokens_api_requests",
    neverDisplaysCopiesStoresOrRequestsCredentials: true
  });
  assert.deepEqual(directContract.transport, {
    usesFullApiUrls: true,
    doesNotUseMcp: true,
    doesNotUseLocalProxyOrSidecar: true,
    doesNotUseFallbackEndpoint: true,
    workBuddyPostBodyTransport: "bounded_base64_argument_never_stdin"
  });
  assert.deepEqual(directContract.installableHosts, ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode", "trae"]);
  assert.deepEqual(directContract.verifiedManagedRuntimeHosts, ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode"]);
  assert.deepEqual(directContract.requiredRuntimeCapabilities, ["authenticated_full_url_request", "json_task_response", "same_task_status_read", "native_media_byte_delivery"]);
  assert.deepEqual(directContract.acceptanceScenarios.map((scenario) => scenario.id), ["api-identity-read", "catalog-read", "media-submit", "same-task-status", "native-media-delivery"]);
  assert.equal(directContract.acceptanceScenarios.find((scenario) => scenario.id === "api-identity-read").request, "GET https://api.puretokensx.com/v1");
  assert.match(directContract.acceptanceScenarios.find((scenario) => scenario.id === "media-submit").request, /fixed Images or Videos API URL/);
  assert.deepEqual(directContract.userMediaInput, {
    onlyCurrentUserExplicitMedia: true,
    publicUrlOrDeclaredIdRequiresInstalledOrOnDemandProfilePropertyAndTransport: true,
    publicUrlImageEditRequiresExactDeclaredJsonOperation: true,
    requiresInstalledOrOnDemandDeclaredTransport: true,
    skillNeverDownloadsOrRehosts: true,
    apiHandlesDeclaredMultipartAttachments: true,
    nativeAttachmentRouteMustMatchDeclaredOperation: true,
    nativeAttachmentMustNeverFallBackToJsonTextSubmission: true,
    workBuddyNativeAttachmentBodyMode: "multipart_base64_descriptor_only",
    untransportableNativeReferenceMustNeverBeConvertedToTextPrompt: true,
    currentRequestScopeOnly: true,
    fallback: "If the active runtime cannot send an explicitly attached file as the declared multipart representation, stop before submission, explain the declared transport, and do not invent a URL, file ID, or upload path. A public HTTPS URL, file ID, or voice ID explicitly supplied by the user may be sent only in the exact installed or on-demand-profile field whose declared transport permits it; the Skill never downloads, validates, or rehosts it."
  });
  assert.deepEqual(directContract.asyncTaskHandling, {
    reconciliationRequiredPrecedesLifecycle: true,
    reconciliationOutcome: "stop_ordinary_polling_retain_same_task_id_do_not_submit_replacement_or_infer_refund",
    submissionRuntimeOutputUnknown: "treat_submission_outcome_as_unknown_do_not_repeat_post_or_submit_replacement_task",
    statusHttp429: "honor_valid_positive_retry_after_then_continue_same_task_if_remaining_budget",
    non429StatusReadFailure: "stop_bounded_window_without_replacement_task",
    doesNotInventApiErrorCode: true
  });
  assert.deepEqual(directContract.mediaDeliveryResourceBounds, {
    oneInFlightStatusReadPerTask: true,
    noBackgroundPollingOrQueue: true,
    readContentOnlyAfterTerminalSuccess: true,
    oneInFlightContentReadPerTask: true,
    sequentialImageContentIndexes: true,
    neverPrefetchOrRefetchDeliveredContent: true,
    handoffNativeBytesBeforeNextRead: true,
    doesNotPersistMediaBytesInSkillStateOrLogs: true,
    confirmedDeliveryEndsCurrentResponse: true,
    fallback: "If native media bytes cannot be handed off, report same-task delivery as unavailable. Do not substitute a URL, HTML, SVG, status text, or a re-submitted task."
  });
  for (const name of names) {
    const manifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", name, "skill.json"), "utf8"));
    assert.deepEqual([...manifest.supportedClients].sort(), supported);
    assert.deepEqual(manifest.distribution.codex, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.agents/skills",
      requiresPluginFeature: false
    });
    assert.deepEqual(manifest.distribution.claudeCode, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.claude/skills"
    });
    assert.deepEqual(manifest.distribution.geminiCli, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.gemini/skills"
    });
    assert.deepEqual(manifest.distribution.workbuddy, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.workbuddy/skills"
    });
    assert.deepEqual(manifest.distribution.grokBuild, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.grok/skills"
    });
    assert.deepEqual(manifest.distribution.opencode, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.config/opencode/skills"
    });
    assert.deepEqual(manifest.distribution.trae, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.trae/skills"
    });
  }
  const imageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-image", "skill.json"), "utf8"));
  assert.equal(imageManifest.rules.profiledPublicUrlReferenceSupported, true);
  assert.equal(imageManifest.rules.profiledImageEditSupported, true);
  assert.equal(imageManifest.rules.profiledPublicUrlJsonImageEditSupported, true);
  assert.equal(imageManifest.rules.untransportableNativeReferenceAttachmentsFailClosed, true);
  assert.equal(imageManifest.rules.neverConvertsUntransportableAttachmentsToPrompt, true);
  assert.equal(imageManifest.rules.currentRequestScopeOnly, true);
  assert.equal(imageManifest.rules.endsAfterConfirmedMediaDelivery, true);
  assert.equal(imageManifest.rules.recognizesReconciliationRequired, true);
  assert.equal(imageManifest.rules.honorsStatusRetryAfterForSameTask, true);
  assert.equal(imageManifest.rules.reportsOnlyExplicitPublicApiErrorCodes, true);
  const videoManifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-video", "skill.json"), "utf8"));
  assert.equal(videoManifest.rules.recognizesReconciliationRequired, true);
  assert.equal(videoManifest.rules.honorsStatusRetryAfterForSameTask, true);
  assert.equal(videoManifest.rules.reportsOnlyExplicitPublicApiErrorCodes, true);
  assert.equal(imageManifest.rules.usesIntentAwarePromptPreparation, true);
  assert.equal(imageManifest.rules.doesNotSplitDistinctImageAssets, true);
  assert.equal(imageManifest.rules.usesSafeStructuredFailureReceipts, true);
  assert.equal(imageManifest.rules.usesResourceBoundedPolling, true);
  assert.equal(imageManifest.rules.doesNotCreateBackgroundMediaWork, true);
  assert.equal(imageManifest.rules.usesSequentialNativeContentDelivery, true);
  assert.equal(Object.hasOwn(imageManifest.rules, "textToImageOnly"), false);
  assert.equal(videoManifest.rules.profiledPublicUrlOrDeclaredIdReferenceSupported, true);
  assert.equal(videoManifest.rules.profiledImageToVideoSupported, true);
  assert.equal(videoManifest.rules.profiledReferenceImageVideoSupported, true);
  assert.equal(videoManifest.rules.profiledReferenceVideoSupported, true);
  assert.equal(videoManifest.rules.profiledReferenceAudioSupported, true);
  assert.equal(videoManifest.rules.profiledVideoEditSupported, true);
  assert.equal(videoManifest.rules.nativeAttachmentsMustUseExactDeclaredMultipartOperation, true);
  assert.equal(videoManifest.rules.workBuddyNativeAttachmentsUseMultipartBase64, true);
  assert.equal(videoManifest.rules.nativeAttachmentNeverFallsBackToTextSubmission, true);
  assert.equal(videoManifest.rules.usesSafeStructuredFailureReceipts, true);
  assert.equal(videoManifest.rules.modeSpecificResolutionUsesInstalledOrOnDemandConstraint, true);
  assert.equal(videoManifest.rules.usesResourceBoundedPolling, true);
  assert.equal(videoManifest.rules.doesNotCreateBackgroundMediaWork, true);
  assert.equal(videoManifest.rules.usesSequentialNativeContentDelivery, true);
  assert.equal(Object.hasOwn(videoManifest.rules, "textToVideoOnly"), false);
});

test("formal reference schemas ship with the package", async () => {
  for (const schema of [
    "media-execution-contract.schema.json",
    "media-behavior-scenarios.schema.json",
    "model-selection.schema.json",
    "host-support.schema.json",
    "direct-api-execution-contract.schema.json",
    "catalog-freshness.schema.json",
    "balance-snapshot.schema.json",
    "task-receipt.schema.json"
  ]) {
    const document = JSON.parse(await readFile(path.join(repositoryRoot, "schemas", schema), "utf8"));
    assert.equal(document.$schema, "https://json-schema.org/draft/2020-12/schema");
    assert.match(document.$id, /^https:\/\/puretokensx\.com\/schemas\//);
  }
});

test("native platform Skill installers ship with the managed runtime", async () => {
  const shellInstaller = path.join(repositoryRoot, "runtime", "puretokens-skill-install.sh");
  const powerShellInstaller = path.join(repositoryRoot, "runtime", "puretokens-skill-install.ps1");
  const run = promisify(execFile);
  await run("sh", ["-n", shellInstaller], { cwd: repositoryRoot });
  const [shellSource, powerShellSource] = await Promise.all([readFile(shellInstaller, "utf8"), readFile(powerShellInstaller, "utf8")]);
  assert.match(shellSource, /repository_archive_url="https:\/\/github\.com\/PureTokens\/puretokens-skill\/archive\/refs\/heads\/main\.zip"/);
  assert.match(shellSource, /curl --fail --location --proto '=https'/);
  assert.match(shellSource, /unzip -q/);
  assert.match(shellSource, /unmanaged Skill conflicts/);
  assert.match(shellSource, /Pure Tokens Skills \$release_version synchronized at \$target_root/);
  assert.doesNotMatch(shellSource, /node|npm|git clone/i);
  assert.match(powerShellSource, /Invoke-WebRequest -Uri \$archiveUrl/);
  assert.match(powerShellSource, /Expand-Archive/);
  assert.match(powerShellSource, /unmanaged Skill conflicts/);
  assert.match(powerShellSource, /Pure Tokens Skills \$releaseVersion synchronized at \$targetRoot/);
  assert.doesNotMatch(powerShellSource, /node|npm|git clone/i);
});

test("Grok Build direct runtime resolves only matching configured model credentials", () => {
  const config = [
    '[model."puretokens-a"]',
    'base_url = "https://api.puretokensx.com/v1"',
    'api_key = "test-direct-key"',
    "",
    '[model."other"]',
    'base_url = "https://example.com/v1"',
    'api_key = "other-key"',
    "",
    '[model."puretokens-env"]',
    'base_url = "https://api.puretokensx.com/v1"',
    'env_key = "PURETOKENS_TEST_KEY"'
  ].join("\n");
  const records = parseGrokModelEntries(config, { PURETOKENS_TEST_KEY: "test-direct-key" });
  assert.deepEqual(records.map((record) => ({ baseUrl: record.baseUrl, hasCredential: Boolean(record.credential) })), [
    { baseUrl: "https://api.puretokensx.com/v1", hasCredential: true },
    { baseUrl: "https://example.com/v1", hasCredential: true },
    { baseUrl: "https://api.puretokensx.com/v1", hasCredential: true }
  ]);
});

test("direct runtime accepts bounded base64 request bodies without standard input", () => {
  const jsonBody = { model: "gpt-image-2", prompt: "test image", async: true };
  const jsonBase64 = Buffer.from(JSON.stringify(jsonBody), "utf8").toString("base64");
  assert.deepEqual(readJsonBase64Argument(jsonBase64), jsonBody);
  assert.deepEqual(
    parseArguments(["request", "--host", "workbuddy", "--method", "POST", "--path", "/v1/images/generations", "--json-base64", jsonBase64]),
    {
      host: "workbuddy",
      method: "POST",
      path: "/v1/images/generations",
      jsonStdin: false,
      jsonBase64,
      multipartStdin: false,
      multipartBase64: undefined,
      outputFile: undefined
    }
  );
  const multipartBase64 = Buffer.from(JSON.stringify({ fields: { model: "seedance-2.0-mini", prompt: "test" }, files: [{ field: "reference_images", path: "/tmp/reference.png" }] }), "utf8").toString("base64");
  assert.deepEqual(
    parseArguments(["request", "--host", "workbuddy", "--method", "POST", "--path", "/v1/videos", "--multipart-base64", multipartBase64]).multipartBase64,
    multipartBase64
  );
  assert.throws(
    () => parseArguments(["request", "--host", "codex", "--method", "POST", "--path", "/v1/images/generations", "--json-stdin", "--json-base64", jsonBase64]),
    /exactly one JSON or multipart request-body mode/
  );
  assert.throws(
    () => parseArguments(["request", "--host", "workbuddy", "--method", "POST", "--path", "/v1/videos", "--json-stdin"]),
    /WorkBuddy POST request bodies must use --json-base64 or --multipart-base64/
  );
  assert.throws(
    () => parseArguments(["request", "--host", "workbuddy", "--method", "POST", "--path", "/v1/videos", "--multipart-stdin"]),
    /WorkBuddy POST request bodies must use --json-base64 or --multipart-base64/
  );
  assert.throws(() => readJsonBase64Argument("eyJ4IjoxfQ"), /canonical UTF-8 JSON/);
  assert.throws(() => readJsonBase64Argument(Buffer.from("not json", "utf8").toString("base64")), /valid JSON/);
});

test("Claude Code direct runtime accepts only an exact configured Pure Tokens origin", async (t) => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "puretokens-claude-runtime-"));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const settingsPath = path.join(directory, "settings.json");
  await writeFile(settingsPath, JSON.stringify({
    env: {
      ANTHROPIC_BASE_URL: "https://api.puretokensx.com/",
      ANTHROPIC_AUTH_TOKEN: "test-claude-key"
    }
  }));
  assert.equal(await resolveClaudeCodeCredential(settingsPath), "test-claude-key");
  await writeFile(settingsPath, JSON.stringify({
    env: {
      ANTHROPIC_BASE_URL: "https://example.com",
      ANTHROPIC_AUTH_TOKEN: "test-claude-key"
    }
  }));
  await assert.rejects(resolveClaudeCodeCredential(settingsPath), /No usable Pure Tokens API credential/);
});

test("Codex direct runtime resolves only the active exact Pure Tokens provider", async (t) => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "puretokens-codex-runtime-"));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const configPath = path.join(directory, "config.toml");
  const authPath = path.join(directory, "auth.json");
  const activeConfig = [
    'model_provider = "puretokens"',
    "",
    "[model_providers.puretokens]",
    'base_url = "https://api.puretokensx.com/v1"',
    'experimental_bearer_token = "test-codex-key"',
    "",
    "[model_providers.other]",
    'base_url = "https://example.com/v1"',
    'experimental_bearer_token = "other-key"'
  ].join("\n");
  assert.deepEqual(parseCodexConnection(activeConfig), {
    baseUrl: "https://api.puretokensx.com/v1",
    credential: "test-codex-key"
  });
  await writeFile(configPath, activeConfig);
  await writeFile(authPath, JSON.stringify({ OPENAI_API_KEY: "fallback-key" }));
  assert.equal(await resolveCodexCredential(configPath, authPath), "test-codex-key");
  await writeFile(configPath, activeConfig.replace('experimental_bearer_token = "test-codex-key"', ""));
  assert.equal(await resolveCodexCredential(configPath, authPath), "fallback-key");
  await writeFile(configPath, activeConfig.replace("https://api.puretokensx.com/v1", "https://example.com/v1"));
  await assert.rejects(resolveCodexCredential(configPath, authPath), /No usable Pure Tokens API credential/);
});

test("Gemini CLI direct runtime requires the selected exact Pure Tokens API-key connection", async (t) => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "puretokens-gemini-runtime-"));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const envPath = path.join(directory, ".env");
  const settingsPath = path.join(directory, "settings.json");
  const source = [
    "# current managed connection",
    "GOOGLE_GEMINI_BASE_URL=https://api.puretokensx.com",
    "GEMINI_API_KEY=test-gemini-key"
  ].join("\n");
  assert.equal(parseDotEnv(source).get("GEMINI_API_KEY"), "test-gemini-key");
  await writeFile(envPath, source);
  await writeFile(settingsPath, JSON.stringify({ security: { auth: { selectedType: "gemini-api-key" } } }));
  assert.equal(await resolveGeminiCliCredential(envPath, settingsPath), "test-gemini-key");
  await writeFile(settingsPath, JSON.stringify({ security: { auth: { selectedType: "oauth" } } }));
  await assert.rejects(resolveGeminiCliCredential(envPath, settingsPath), /No usable Pure Tokens API-key connection/);
});

test("OpenCode direct runtime accepts only its exact Pure Tokens provider entry", async (t) => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "puretokens-opencode-runtime-"));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const configPath = path.join(directory, "opencode.json");
  await writeFile(configPath, JSON.stringify({
    provider: {
      puretokens: { options: { baseURL: "https://api.puretokensx.com/v1/", apiKey: "test-opencode-key" } },
      other: { options: { baseURL: "https://example.com/v1", apiKey: "other-key" } }
    }
  }));
  assert.equal(await resolveOpenCodeCredential(configPath), "test-opencode-key");
  await writeFile(configPath, JSON.stringify({ provider: { puretokens: { options: { baseURL: "https://example.com/v1", apiKey: "test-opencode-key" } } } }));
  await assert.rejects(resolveOpenCodeCredential(configPath), /No usable Pure Tokens API credential/);
});

test("Trae has no local credential resolver", async () => {
  await assert.rejects(resolveConfiguredCredential("trae"), /manual connection setup/);
});

test("WorkBuddy direct runtime parses only model URL and credential fields", () => {
  const entries = parseWorkBuddyModelEntries([
    { id: "puretokens", url: "https://api.puretokensx.com/v1", apiKey: "test-direct-key", vendor: "openai" },
    { id: "other", url: "https://example.com/v1", apiKey: "other-key" },
    { id: "missing" }
  ]);
  assert.deepEqual(entries.map((entry) => ({ baseUrl: entry.baseUrl, hasCredential: Boolean(entry.credential) })), [
    { baseUrl: "https://api.puretokensx.com/v1", hasCredential: true },
    { baseUrl: "https://example.com/v1", hasCredential: true },
    { baseUrl: undefined, hasCredential: false }
  ]);
});

test("WorkBuddy direct runtime resolves one credential from fixed API base or resource URLs only", async (t) => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "puretokens-workbuddy-runtime-"));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const configPath = path.join(directory, "models.json");
  await writeFile(configPath, JSON.stringify([
    { url: "https://example.com/v1", apiKey: "other-key" },
    { url: "https://api.puretokensx.com/v1?unexpected=query", apiKey: "not-a-match" },
    { url: "https://api.puretokensx.com/v2/chat/completions", apiKey: "not-a-match" },
    { url: "https://api.puretokensx.com/v1/chat/completions", apiKey: "test-direct-key" },
    { url: "https://api.puretokensx.com/v1/responses", apiKey: "test-direct-key" }
  ]));
  assert.equal(await resolveWorkBuddyCredential(configPath), "test-direct-key");
  await writeFile(configPath, JSON.stringify([
    { url: "https://api.puretokensx.com/v1/chat/completions", apiKey: "first-key" },
    { url: "https://api.puretokensx.com/v1/responses", apiKey: "second-key" }
  ]));
  await assert.rejects(resolveWorkBuddyCredential(configPath), /Multiple different Pure Tokens API credentials/);
});

test("CLI installs each specialist Skill", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-cli-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const run = promisify(execFile);
  for (const name of names) {
    await run(process.execPath, [path.join(repositoryRoot, "bin", "puretokens-skill.js"), "install", name, "--target", target], { cwd: repositoryRoot });
    assert.equal(JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8")).name, name);
    if (["puretokens-image", "puretokens-video"].includes(name)) {
      const manifest = JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8"));
      assert.equal(manifest.taskReceipt, "references/task-receipt.json");
      assert.equal(JSON.parse(await readFile(path.join(target, name, manifest.taskReceipt), "utf8")).kind, name.replace("puretokens-", ""));
    }
  }
  assert.equal(JSON.parse(await readFile(path.join(target, ".puretokens-runtime", "runtime.json"), "utf8")).name, "puretokens-direct-api-runtime");
});

test("CLI sync installs missing Skills and refuses unmanaged conflicts before writing", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-sync-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const run = promisify(execFile);
  const conflict = path.join(target, "puretokens-image");
  await mkdir(conflict);
  await writeFile(path.join(conflict, "SKILL.md"), "unmanaged\n");
  await assert.rejects(
    run(process.execPath, [path.join(repositoryRoot, "bin", "puretokens-skill.js"), "sync", "--target", target], { cwd: repositoryRoot }),
    /unmanaged Skill conflicts/
  );
  await assert.rejects(readFile(path.join(target, "puretokens-balance", "skill.json"), "utf8"));
  await assert.rejects(readFile(path.join(target, ".puretokens-runtime", "runtime.json"), "utf8"));
  await rm(conflict, { recursive: true, force: true });
  await run(process.execPath, [path.join(repositoryRoot, "bin", "puretokens-skill.js"), "sync", "--target", target], { cwd: repositoryRoot });
  for (const name of names) {
    assert.equal(JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8")).name, name);
  }
  assert.equal(JSON.parse(await readFile(path.join(target, ".puretokens-runtime", "runtime.json"), "utf8")).version, "0.13.15");
});

test("CLI refuses an unmanaged direct runtime before changing Skills", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-runtime-conflict-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const run = promisify(execFile);
  const runtime = path.join(target, ".puretokens-runtime");
  await mkdir(runtime);
  await writeFile(path.join(runtime, "runtime.json"), JSON.stringify({ name: "another-runtime" }));
  await assert.rejects(
    run(process.execPath, [path.join(repositoryRoot, "bin", "puretokens-skill.js"), "sync", "--target", target], { cwd: repositoryRoot }),
    /unmanaged Pure Tokens runtime conflicts/
  );
  await assert.rejects(readFile(path.join(target, "puretokens-image", "skill.json"), "utf8"));
  assert.equal(JSON.parse(await readFile(path.join(runtime, "runtime.json"), "utf8")).name, "another-runtime");
});

test("CLI sync removes verified retired managed Skills and stale retired backups", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-retired-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const run = promisify(execFile);
  const retired = path.join(target, "puretokens_image");
  await mkdir(retired);
  await writeFile(path.join(retired, "SKILL.md"), "retired managed Skill\n");
  await writeFile(path.join(retired, "skill.json"), JSON.stringify({ name: "puretokens_image" }));
  const staleBackup = path.join(target, ".puretokens_image.retired-prior-update");
  await mkdir(staleBackup);
  await writeFile(path.join(staleBackup, "SKILL.md"), "retired managed Skill\n");
  await writeFile(path.join(staleBackup, "skill.json"), JSON.stringify({ name: "puretokens_image" }));

  const { stdout } = await run(process.execPath, [path.join(repositoryRoot, "bin", "puretokens-skill.js"), "sync", "--target", target], { cwd: repositoryRoot });
  assert.equal((stdout.match(/Removed retired managed puretokens_image from/g) ?? []).length, 2);
  assert.match(stdout, /Pure Tokens Skills 0\.13\.15 synchronized at/);
  await assert.rejects(readFile(path.join(retired, "skill.json"), "utf8"));
  await assert.rejects(readFile(path.join(staleBackup, "skill.json"), "utf8"));
  assert.equal((await readdir(target)).some((entry) => entry.startsWith(".puretokens_image.retired-")), false);
  assert.equal(JSON.parse(await readFile(path.join(target, "puretokens-image", "skill.json"), "utf8")).name, "puretokens-image");
});

test("task receipts always expose the same core task fields", async () => {
  const core = ["exact_model_id", "task_id", "returned_state", "requested_operation", "requested_count", "requested_size_or_parameters", "next_action"];
  const failure = ["failure_phase", "api_error_code", "http_status", "error_message"];
  for (const name of ["puretokens-image", "puretokens-video"]) {
    const receipt = JSON.parse(await readFile(path.join(repositoryRoot, "skills", name, "references", "task-receipt.json"), "utf8"));
    for (const phase of ["submission", "continuation", "reconciliation", "completion", "failure"]) {
      for (const field of core) assert.ok(receipt[phase].requiredFields.includes(field), `${name} ${phase} missing ${field}`);
    }
    assert.ok(receipt.reconciliation.requiredFields.includes("reconciliation_required"));
    assert.ok(receipt.completion.requiredFields.includes("delivered_count"));
    for (const field of failure) assert.ok(receipt.failure.requiredFields.includes(field), `${name} failure missing ${field}`);
    assert.equal(receipt.failure.conditionalFields.retry_after_seconds, "Include only when the API returned HTTP 429 and a valid positive Retry-After value.");
    assert.equal(receipt.failure.conditionalFields.api_error_code, "Use only the exact public API error code explicitly returned by the API; otherwise write not returned.");
  }
});

test("catalog parameter profiles propagate into installed model selections", () => {
  const previous = {
    models: [{ id: "example-image", aliases: ["example"], goodFor: { en: "Image generation", zh: "图片生成" }, example: { en: "Example", zh: "示例" } }]
  };
  const rows = sourceRows({
    vendors: [{ id: 9, name: "Example vendor" }],
    data: [{ model_name: "example-image", vendor_id: 9, capabilities: ["image"], media_input_schema: { properties: { size: { enum: ["1024x1024"] } } } }]
  });
  const catalog = buildPublishedCatalog(previous, rows, { sourcePath: "/controlled/catalog", capturedAt: "2026-08-21T02:46:19.421Z" });
  const selection = buildSelection(catalog, "image");
  assert.deepEqual(selection.models, [{ id: "example-image", aliases: ["example"], parameterSchema: { properties: { size: { enum: ["1024x1024"] } } } }]);
});

test("catalog freshness gate is deterministic", async () => {
  const catalog = JSON.parse(await readFile(path.join(repositoryRoot, "references", "media-model-catalog.json"), "utf8"));
  const capturedAt = Date.parse(catalog.serviceCatalog.capturedAt);
  const passed = await checkMediaCatalogFreshness({ maxAgeDays: 7, now: capturedAt + 3 * 24 * 60 * 60 * 1000 });
  assert.equal(passed.maxAgeDays, 7);
  await assert.rejects(
    checkMediaCatalogFreshness({ maxAgeDays: 1, now: capturedAt + 3 * 24 * 60 * 60 * 1000 }),
    /days old/
  );
});
