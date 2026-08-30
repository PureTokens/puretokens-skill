import { createHash } from "node:crypto";
import { readFile, readdir, stat } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

export const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
export const skillsRoot = path.join(repositoryRoot, "skills");

const forbiddenPattern = /(BEGIN [A-Z ]*PRIVATE|api[_-]?key|authorization:|bearer\s+|pts-router-token|127\.0\.0\.1:|\/Users\/)/i;
const directApiOrigin = "https://api.puretokensx.com";
const directAcceptanceScenarioIds = ["api-identity-read", "catalog-read", "media-submit", "same-task-status", "native-media-delivery"];
const receiptCoreFields = ["exact_model_id", "task_id", "returned_state", "requested_operation", "requested_count", "requested_size_or_parameters", "next_action"];

export async function readSkillRegistry() {
  const registryPath = path.join(skillsRoot, "index.json");
  const registry = JSON.parse(await readFile(registryPath, "utf8"));
  if (registry?.schemaVersion !== 1 || !Array.isArray(registry.skills)) {
    throw new Error("skills/index.json must contain schemaVersion=1 and a skills array");
  }
  return registry;
}

export async function collectSkillRecords() {
  const registry = await readSkillRegistry();
  const directoryEntries = await readdir(skillsRoot, { withFileTypes: true });
  const directories = directoryEntries.filter((entry) => entry.isDirectory()).map((entry) => entry.name).sort();
  const records = [];

  for (const directory of directories) {
    const skillDir = path.join(skillsRoot, directory);
    const [skillText, manifestText] = await Promise.all([
      readFile(path.join(skillDir, "SKILL.md"), "utf8"),
      readFile(path.join(skillDir, "skill.json"), "utf8")
    ]);
    const manifest = JSON.parse(manifestText);
    const frontmatter = readFrontmatter(skillText);
    records.push({ directory, skillDir, skillText, manifest, frontmatter });
  }

  return { registry, records };
}

export async function validateRepository() {
  const errors = [];
  const { registry, records } = await collectSkillRecords();
  const packageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "package.json"), "utf8"));
  if (!/^\d+\.\d+\.\d+$/.test(packageManifest?.version || "")) {
    errors.push("package.json must contain a semver version");
  }
  await validateSchemaDocuments(errors);
  const hostSupport = await readHostSupport(errors);
  const directApiExecutionContract = await readDirectApiExecutionContract(errors);
  await validateCatalogFreshnessPolicy(errors);
  validateDirectApiExecutionContract(errors, directApiExecutionContract);
  const seen = new Set();
  const registryByName = new Map(registry.skills.map((skill) => [skill?.name, skill]));

  for (const record of records) {
    const { directory, skillText, manifest, frontmatter } = record;
    const name = manifest?.name;
    if (!/^[a-z][a-z0-9_]*$/.test(directory)) errors.push(`${directory}: directory must use snake_case`);
    if (!/^[a-z][a-z0-9_]*$/.test(name || "")) errors.push(`${directory}: manifest name must use snake_case`);
    if (name !== directory) errors.push(`${directory}: directory and manifest name must match`);
    if (frontmatter.name !== name) errors.push(`${directory}: SKILL.md frontmatter name must match manifest`);
    if (!frontmatter.description) errors.push(`${directory}: SKILL.md needs a non-empty frontmatter description`);
    if (!manifest?.version || !/^\d+\.\d+\.\d+$/.test(manifest.version)) errors.push(`${directory}: manifest version must be semver`);
    if (manifest?.version !== packageManifest?.version) errors.push(`${directory}: manifest version must match package.json`);
    if (manifest?.sourceSha256 !== undefined) {
      const sourceSha256 = String(manifest.sourceSha256);
      const actualSourceSha256 = createHash("sha256").update(skillText).digest("hex");
      if (!/^[a-f0-9]{64}$/.test(sourceSha256)) {
        errors.push(`${directory}: sourceSha256 must be a lowercase SHA-256 digest`);
      } else if (sourceSha256 !== actualSourceSha256) {
        errors.push(`${directory}: sourceSha256 does not match SKILL.md`);
      }
    }
    if (seen.has(name)) errors.push(`${directory}: duplicate skill name ${name}`);
    seen.add(name);
    const registryEntry = registryByName.get(name);
    if (!registryEntry) errors.push(`${directory}: missing skills/index.json entry`);
    if (registryEntry && registryEntry.version !== manifest.version) {
      errors.push(`${directory}: manifest and skills/index.json versions differ`);
    }
    if (manifest?.behaviorTests) await verifyFile(errors, manifest.behaviorTests);
    if (manifest?.naturalLanguageAliases) await verifyFile(errors, manifest.naturalLanguageAliases);
    await validateSpecialistReferences(errors, record);
    const declaredClients = Array.isArray(manifest?.supportedClients) ? [...manifest.supportedClients].sort() : [];
    const supportedHostIds = hostSupport.supported.map((host) => host.id).sort();
    if (supportedHostIds.length && JSON.stringify(declaredClients) !== JSON.stringify(supportedHostIds)) {
      errors.push(`${directory}: supportedClients must match references/host-support.json`);
    }
    const excludedClients = Array.isArray(manifest?.excludedClients) ? [...manifest.excludedClients].sort() : [];
    const notDistributedHostIds = hostSupport.notDistributed.map((host) => host.id).sort();
    if (notDistributedHostIds.length && JSON.stringify(excludedClients) !== JSON.stringify(notDistributedHostIds)) {
      errors.push(`${directory}: excludedClients must match references/host-support.json`);
    }
    const claudeDesktop = manifest?.distribution?.claudeDesktop;
    if (!claudeDesktop) {
      errors.push(`${directory}: Claude Desktop distribution is required`);
    } else {
      if (claudeDesktop.format !== "zip") errors.push(`${directory}: Claude Desktop distribution must use zip format`);
      if (claudeDesktop.archiveRoot !== name) errors.push(`${directory}: Claude Desktop archive root must match the skill name`);
      if (!Array.isArray(claudeDesktop.requiredFiles) || !claudeDesktop.requiredFiles.includes("SKILL.md")) {
        errors.push(`${directory}: Claude Desktop bundle must include SKILL.md`);
      }
      if (claudeDesktop.enableAfterImport !== true) errors.push(`${directory}: Claude Desktop bundle must require explicit enablement`);
    }
    validateHostDistributions(errors, directory, manifest, hostSupport);
    if (forbiddenPattern.test(skillText) || forbiddenPattern.test(JSON.stringify(manifest))) {
      errors.push(`${directory}: skill content contains a forbidden credential or local-runtime marker`);
    }
  }

  const expected = ["puretokens_balance", "puretokens_connection", "puretokens_models", "puretokens_image", "puretokens_video", "puretokens_update"];
  if (registry.skills.length !== expected.length || expected.some((name, index) => registry.skills[index]?.name !== name)) {
    errors.push("skills/index.json must list the six specialist Skills in order");
  }

  for (const entry of registry.skills) {
    if (!entry || !seen.has(entry.name)) errors.push(`skills/index.json: ${entry?.name || "unnamed entry"} has no skill directory`);
    if (!entry?.entry || !entry?.manifest) errors.push(`skills/index.json: ${entry?.name || "unnamed entry"} needs entry and manifest`);
    if (entry?.entry) await verifyFile(errors, entry.entry);
    if (entry?.manifest) await verifyFile(errors, entry.manifest);
  }

  return errors;
}

async function readHostSupport(errors) {
  const file = path.join(repositoryRoot, "references", "host-support.json");
  try {
    const support = JSON.parse(await readFile(file, "utf8"));
    if (support?.$schema !== "https://puretokensx.com/schemas/host-support.schema.json") {
      errors.push("references/host-support.json must declare the host-support schema");
    }
    if (support?.schemaVersion !== 1 || !Array.isArray(support.supported) || !Array.isArray(support.notDistributed)) {
      errors.push("references/host-support.json must be schemaVersion=1 with supported and notDistributed arrays");
      return { supported: [], notDistributed: [] };
    }
    if (support.directApiOrigin !== directApiOrigin) {
      errors.push("references/host-support.json must declare the fixed direct API origin");
    }
    const records = [...support.supported, ...support.notDistributed];
    const ids = records.map((host) => host?.id).filter((id) => typeof id === "string" && /^[a-z][a-z0-9-]*$/.test(id));
    if (ids.length !== records.length || new Set(ids).size !== ids.length) {
      errors.push("references/host-support.json must contain unique, kebab-case host IDs");
    }
    for (const host of support.supported) {
      if (!host || typeof host.guidance !== "string" || !host.guidance) {
        errors.push(`references/host-support.json: ${host?.id || "unnamed host"} needs guidance`);
      }
      if (host?.delivery === "manual-source") {
        if (typeof host.globalSkillDirectory !== "string" || !/^~\/\.[a-z-]+\/skills$/.test(host.globalSkillDirectory)) {
          errors.push(`references/host-support.json: ${host?.id || "unnamed host"} needs a global Skill directory`);
        }
      } else if (host?.delivery !== "bundle") {
        errors.push(`references/host-support.json: ${host?.id || "unnamed host"} has an unsupported delivery type`);
      }
    }
    for (const host of support.notDistributed) {
      if (!host || typeof host.reason !== "string" || !host.reason) {
        errors.push(`references/host-support.json: ${host?.id || "unnamed host"} needs a not-distributed reason`);
      }
    }
    return { supported: support.supported, notDistributed: support.notDistributed };
  } catch {
    errors.push("references/host-support.json is missing or invalid JSON");
    return { supported: [], notDistributed: [] };
  }
}

async function readDirectApiExecutionContract(errors) {
  const file = path.join(repositoryRoot, "references", "direct-api-execution-contract.json");
  try {
    return JSON.parse(await readFile(file, "utf8"));
  } catch {
    errors.push("references/direct-api-execution-contract.json is missing or invalid JSON");
    return undefined;
  }
}

async function validateCatalogFreshnessPolicy(errors) {
  const file = path.join(repositoryRoot, "references", "catalog-freshness.json");
  try {
    const policy = JSON.parse(await readFile(file, "utf8"));
    if (policy?.$schema !== "https://puretokensx.com/schemas/catalog-freshness.schema.json" || policy?.schemaVersion !== 1 ||
      !Number.isInteger(policy?.maxAgeDays) || policy.maxAgeDays < 1 || policy.maxAgeDays > 31 ||
      policy.releaseCommand !== "npm run release:validate" || typeof policy.message !== "string" || !policy.message) {
      errors.push("references/catalog-freshness.json must declare a valid release freshness policy");
    }
  } catch {
    errors.push("references/catalog-freshness.json is missing or invalid JSON");
  }
}

function validateDirectApiExecutionContract(errors, contract) {
  const label = "references/direct-api-execution-contract.json";
  if (!contract || typeof contract !== "object") return;
  if (contract.$schema !== "https://puretokensx.com/schemas/direct-api-execution-contract.schema.json" || contract.schemaVersion !== 1) {
    errors.push(`${label} must declare schemaVersion=1 direct API execution contract`);
  }
  const transport = contract.transport;
  if (contract.apiOrigin !== directApiOrigin || !contract.authentication ||
    contract.authentication.usesRuntimeManagedExistingAuthentication !== true ||
    contract.authentication.skillNeverReadsOrConstructsCredentials !== true ||
    !transport || transport.usesFullApiUrls !== true || transport.doesNotUseMcp !== true ||
    transport.doesNotUseLocalProxyOrSidecar !== true || transport.doesNotUseFallbackEndpoint !== true) {
    errors.push(`${label} must use the fixed full API origin with runtime-managed authentication and no MCP, proxy, sidecar, or fallback`);
  }
  if (!Array.isArray(contract.acceptanceScenarios) || !sameArray(contract.acceptanceScenarios.map((scenario) => scenario?.id), directAcceptanceScenarioIds)) {
    errors.push(`${label} must define the ordered direct-API acceptance scenarios`);
  }
  const balance = contract.balance;
  if (!balance || balance.method !== "GET" || balance.url !== `${directApiOrigin}/api/product/desktop/account/balance` ||
    balance.requiresExistingAuthenticatedAccountSession !== true || balance.responseSchema !== "schemas/balance-snapshot.schema.json" ||
    typeof balance.fallback !== "string" || !balance.fallback) {
    errors.push(`${label} must define the fixed direct balance endpoint and fallback`);
  }
  const userMediaInput = contract.userMediaInput;
  if (!userMediaInput || userMediaInput.onlyCurrentUserExplicitMedia !== true ||
    userMediaInput.requiresInstalledOrOnDemandDeclaredTransport !== true ||
    userMediaInput.skillNeverDownloadsOrRehosts !== true ||
    userMediaInput.pureTokensGatewayStagesMultipartAttachments !== true ||
    typeof userMediaInput.fallback !== "string" || !userMediaInput.fallback) {
    errors.push(`${label} must define the fail-closed direct user-media input contract`);
  }
  const mediaDeliveryResourceBounds = contract.mediaDeliveryResourceBounds;
  if (!mediaDeliveryResourceBounds || mediaDeliveryResourceBounds.oneInFlightStatusReadPerTask !== true ||
    mediaDeliveryResourceBounds.noBackgroundPollingOrQueue !== true ||
    mediaDeliveryResourceBounds.readContentOnlyAfterTerminalSuccess !== true ||
    mediaDeliveryResourceBounds.oneInFlightContentReadPerTask !== true ||
    mediaDeliveryResourceBounds.sequentialImageContentIndexes !== true ||
    mediaDeliveryResourceBounds.neverPrefetchOrRefetchDeliveredContent !== true ||
    mediaDeliveryResourceBounds.handoffNativeBytesBeforeNextRead !== true ||
    mediaDeliveryResourceBounds.doesNotPersistMediaBytesInSkillStateOrLogs !== true ||
    typeof mediaDeliveryResourceBounds.fallback !== "string" || !mediaDeliveryResourceBounds.fallback) {
    errors.push(`${label} must define bounded same-task polling and native-media delivery without skill-state or log persistence`);
  }
}

async function validateSpecialistReferences(errors, record) {
  const { directory, skillDir, manifest } = record;
  const required = ["executionContract", "behaviorScenarios"];
  if (directory === "puretokens_image" || directory === "puretokens_video") required.unshift("modelSelection", "taskReceipt");
  const references = {};
  for (const field of required) {
    const relativePath = manifest?.[field];
    if (typeof relativePath !== "string" || !relativePath) {
      errors.push(`${directory}: ${field} must name an installed reference file`);
      continue;
    }
    references[field] = await readSkillReference(errors, skillDir, relativePath, `${directory}: ${field}`);
  }
  const requiredFiles = manifest?.distribution?.claudeDesktop?.requiredFiles;
  for (const field of required) {
    const relativePath = manifest?.[field];
    if (typeof relativePath === "string" && !requiredFiles?.includes(relativePath)) {
      errors.push(`${directory}: Claude Desktop bundle must include ${relativePath}`);
    }
  }
  validateExecutionContract(errors, directory, references.executionContract);
  validateBehaviorScenarios(errors, directory, references.behaviorScenarios);
  if (references.modelSelection) validateModelSelection(errors, directory, references.modelSelection);
  if (references.taskReceipt) validateTaskReceipt(errors, directory, references.taskReceipt);
}

function validateHostDistributions(errors, directory, manifest, hostSupport) {
  const distribution = manifest?.distribution;
  if (!distribution || typeof distribution !== "object") {
    errors.push(`${directory}: distribution is required`);
    return;
  }
  const expectedKeys = new Set(["claudeDesktop"]);
  for (const host of hostSupport.supported) {
    if (host.delivery !== "manual-source") continue;
    const key = distributionKeyForHost(host.id);
    expectedKeys.add(key);
    const delivery = distribution[key];
    if (!delivery || typeof delivery !== "object") {
      errors.push(`${directory}: ${host.id} manual source distribution is required`);
      continue;
    }
    if (delivery.manualInstallationSupported !== true) {
      errors.push(`${directory}: ${host.id} must support manual source installation`);
    }
    if (delivery.globalSkillDirectory !== host.globalSkillDirectory) {
      errors.push(`${directory}: ${host.id} global Skill directory must match references/host-support.json`);
    }
    if (host.id === "codex") {
      if (delivery.requiresPluginFeature !== false) errors.push(`${directory}: Codex delivery must not require the Plugin feature`);
      for (const retiredField of ["managedByDesktop", "managedSkillDirectory", "plugin"]) {
        if (Object.hasOwn(delivery, retiredField)) errors.push(`${directory}: Codex delivery must not declare ${retiredField}`);
      }
    }
  }
  for (const key of Object.keys(distribution)) {
    if (!expectedKeys.has(key)) errors.push(`${directory}: distribution has unsupported host entry ${key}`);
  }
}

function distributionKeyForHost(id) {
  return id.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
}

function validateExecutionContract(errors, directory, contract) {
  const label = `${directory}: executionContract`;
  if (!contract || typeof contract !== "object") return;
  if (contract.$schema !== "https://puretokensx.com/schemas/media-execution-contract.schema.json") errors.push(`${label} must declare the execution-contract schema`);
  const kind = skillKind(directory);
  if (contract.schemaVersion !== 1 || contract.kind !== kind) errors.push(`${label} must be schemaVersion=1 for ${kind}`);
  if (kind === "update") {
    validateUpdateContract(errors, label, contract);
    return;
  }
  validateDirectTransport(errors, label, contract.transport, kind === "image" || kind === "video");
  const operations = contract.operations;
  if (!operations || typeof operations !== "object") {
    errors.push(`${label} must define operations`);
    return;
  }
  if (kind === "balance") {
    validateRequest(errors, `${label} read`, operations.read, "GET", `${directApiOrigin}/api/product/desktop/account/balance`);
    if (operations.read?.requiresExistingAuthenticatedAccountSession !== true || operations.read?.responseSchema !== "https://puretokensx.com/schemas/balance-snapshot.schema.json") {
      errors.push(`${label} read must require an existing account session and the balance snapshot schema`);
    }
    if (contract.result?.reportOnlyReturnedFields !== true || contract.result?.neverEstimate !== true || contract.result?.neverRetry !== true ||
      contract.result?.fallbackWhenAccountSessionUnavailable !== "report_direct_result_and_direct_to_puretokens_client_balance_view") {
      errors.push(`${label} must report only returned balance fields and use the direct-request fallback`);
    }
    return;
  }
  if (kind === "connection") {
    validateRequest(errors, `${label} identity`, operations.identity, "GET", `${directApiOrigin}/v1`);
    if (contract.result?.identitySource !== "api_declared_public_health" || contract.result?.expectedStatus !== "ok" ||
      contract.result?.expectedName !== "Pure Tokens API" || contract.result?.expectedBasePath !== "/v1" ||
      contract.result?.neverRetry !== true || contract.result?.doesNotExposeHostConfiguration !== true ||
      contract.result?.unconfirmedDoesNotProveOtherService !== true) {
      errors.push(`${label} must verify only the fixed /v1 API identity without retrying, exposing host configuration, or inferring another service`);
    }
    return;
  }
  if (kind === "models") {
    validateRequest(errors, `${label} catalog`, operations.catalog, "GET", `${directApiOrigin}/v1/media/models`);
    if (contract.result?.reportOnlyAuthenticatedCatalog !== true || contract.result?.doesNotSubmitMediaTasks !== true ||
      contract.result?.neverRetry !== true || contract.result?.noStaticCatalogFallback !== true ||
      contract.result?.compatibilityShortlistsRequireDeclaredCapabilityAndInputSchema !== true) {
      errors.push(`${label} must report only the live catalog, never submit media work or retry, and derive compatibility only from declared profile data`);
    }
    return;
  }
  if (kind === "image") {
    validateRequest(errors, `${label} catalog`, operations.catalog, "GET", `${directApiOrigin}/v1/media/models`);
    validateRequest(errors, `${label} submit`, operations.submit, "POST", `${directApiOrigin}/v1/images/generations`);
    validateRequiredBodyFields(errors, `${label} submit`, operations.submit, ["model", "prompt", "async"]);
    validateRequest(errors, `${label} status`, operations.status, "GET", `${directApiOrigin}/v1/images/{task_id}`, true);
    validateRequest(errors, `${label} content`, operations.content, "GET", `${directApiOrigin}/v1/images/{task_id}/content?index={index}`, true);
    const editPaths = Array.isArray(operations.edit?.allowedPaths) ? operations.edit.allowedPaths : [];
    if (operations.edit?.method !== "POST" || editPaths.length !== 2 || !editPaths.includes("/v1/images/generations") || !editPaths.includes("/v1/images/edits") ||
      operations.edit?.pathSource !== "installed_model_selection_or_on_demand_live_profile_operation_path" || operations.edit?.contentType !== "multipart/form-data" ||
      operations.edit?.requiresDeclaredProfileOperation !== "image_edit") {
      errors.push(`${label} must define the fixed-origin declared image-edit operation`);
    }
    validateRequiredBodyFields(errors, `${label} edit`, operations.edit, ["model", "prompt", "image", "async"]);
    if (operations.submit?.fixedBody?.async !== true || operations.edit?.fixedBody?.async !== true ||
      operations.edit?.inputSource !== "current_user_explicit_native_media_only" || operations.edit?.transport !== "profile_declared_multipart_file") {
      errors.push(`${label} must fix async and declare the multipart image-edit input`);
    }
    if (contract.parameterValidation?.defaultModel !== "gpt-image-2" || contract.parameterValidation?.normalSubmissionUsesInstalledModelSelection !== true ||
      contract.parameterValidation?.exactModelCoreSubmissionDoesNotRequireCatalogPreflight !== true ||
      contract.parameterValidation?.onDemandLiveCatalogRead !== "only_for_explicit_discovery_installed_profile_gap_or_post_rejection_diagnosis" ||
      contract.parameterValidation?.allOptionalParametersRequire !== "installed_model_selection_parameter_schema_or_on_demand_live_input_schema") {
      errors.push(`${label} must use its installed selection for normal image submissions and limit live catalog reads to on-demand cases`);
    }
    validateTaskIdentityStateAndSubmissionFailure(errors, label, contract, true);
    validateImageRetrieval(errors, label, contract);
    validatePolling(errors, `${label} polling`, contract.polling, { deadlineSeconds: 120, fallbackDelaysSeconds: [3, 6, 12, 24, 30], steadyDelaySeconds: 30, maxAutomaticStatusReads: 6 });
    if (contract.result?.successRequires !== "native_image_bytes") errors.push(`${label} must require native image bytes for success`);
  } else {
    validateRequest(errors, `${label} catalog`, operations.catalog, "GET", `${directApiOrigin}/v1/media/models`);
    validateRequest(errors, `${label} submit`, operations.submit, "POST", `${directApiOrigin}/v1/videos`);
    validateRequest(errors, `${label} edit`, operations.edit, "POST", `${directApiOrigin}/v1/videos/edits`);
    validateRequest(errors, `${label} status`, operations.status, "GET", `${directApiOrigin}/v1/videos/{task_id}`, true);
    validateRequest(errors, `${label} content`, operations.content, "GET", `${directApiOrigin}/v1/videos/{task_id}/content`, true);
    validateRequiredBodyFields(errors, `${label} submit`, operations.submit, ["model"]);
    if (contract.parameterValidation?.normalSubmissionUsesInstalledModelSelection !== true ||
      contract.parameterValidation?.exactModelCoreSubmissionDoesNotRequireCatalogPreflight !== true ||
      contract.parameterValidation?.onDemandLiveCatalogRead !== "only_for_explicit_discovery_installed_profile_gap_or_post_rejection_diagnosis" ||
      contract.parameterValidation?.allModelsOptionalParametersRequire !== "installed_model_selection_parameter_schema_or_on_demand_live_input_schema" ||
      contract.parameterValidation?.promptRequirementUsesInstalledOrOnDemandPropertiesAndConstraints !== true ||
      contract.parameterValidation?.resolutionUsesInstalledOrOnDemandModeConstraint !== true || contract.result?.successRequires !== "native_video_bytes") {
      errors.push(`${label} must use its installed selection for normal video submissions, limit live catalog reads to on-demand cases, and require native video bytes`);
    }
    validateTaskIdentityStateAndSubmissionFailure(errors, label, contract, false);
    const retrieval = contract.contentRetrieval;
    if (!retrieval || retrieval.fetchOnlyAfterTerminalSuccess !== true || retrieval.oneInFlightContentReadPerTask !== true || retrieval.neverPrefetchOrRefetchDeliveredContent !== true || retrieval.handoffBeforeNextContentRead !== true) {
      errors.push(`${label} must retrieve video content only after terminal success, one response at a time, without prefetching or duplicate delivery`);
    }
    validatePolling(errors, `${label} polling`, contract.polling, { deadlineSeconds: 300, fallbackDelaysSeconds: [5, 10, 20, 40, 60], steadyDelaySeconds: 60, maxAutomaticStatusReads: 7 });
  }
  if (contract.result?.sameTaskOnly !== true || contract.result?.neverAutoResubmit !== true || !Array.isArray(contract.unsupportedInput) || !contract.unsupportedInput.length) {
    errors.push(`${label} must stay on the same task without automatic resubmission and declare unsupported input`);
  }
}

function validateUpdateContract(errors, label, contract) {
  const transport = contract.transport;
  if (!transport || transport.localSkillManager !== true || transport.usesOfficialMainBranch !== true ||
    transport.doesNotReadCredentialsOrHostConfiguration !== true || transport.doesNotUseMediaApiOrMcp !== true) {
    errors.push(`${label} must use the local official Skill manager without credentials, media APIs, or MCP`);
  }
  const sync = contract.operations?.sync;
  if (!sync || sync.commandTemplate !== "node bin/puretokens-skill.js sync --target <installation-root>" ||
    sync.sourceRepository !== "https://github.com/PureTokens/puretokens-skill.git" || sync.sourceBranch !== "main" || sync.validationCommand !== "npm run check") {
    errors.push(`${label} must define the validated official-main sync operation`);
  }
  const result = contract.result;
  if (!result || result.installsMissingOfficialSkills !== true || result.upgradesOnlyManagedMatchingSkills !== true ||
    result.neverOverwritesUnmanagedDirectories !== true || result.requiresNewHostConversationAfterSuccess !== true) {
    errors.push(`${label} must preserve unmanaged directories and report a new-conversation requirement`);
  }
}

function validateDirectTransport(errors, label, transport, requiresNativeMediaByteDelivery) {
  if (!transport || transport.fixedApiOrigin !== directApiOrigin || transport.usesFullApiUrls !== true ||
    transport.usesRuntimeManagedAuthentication !== true || transport.doesNotReadCredentialsOrHostConfiguration !== true ||
    transport.doesNotUseMcpOrFallbackTransport !== true || (requiresNativeMediaByteDelivery && transport.requiresNativeMediaByteDelivery !== true)) {
    errors.push(`${label} must use the fixed full API origin, runtime-managed authentication, and no MCP or fallback transport`);
  }
}

function validateImageRetrieval(errors, label, contract) {
  const retrieval = contract.contentRetrieval;
  if (!retrieval || retrieval.indexBase !== 0 || retrieval.requestedCount !== "explicit_n_or_default_1" || retrieval.allowedIndexes !== "0..requestedCount-1" ||
    retrieval.completeDeliveryRequiresEveryRequestedIndex !== true || retrieval.partialOrMissingContent !== "report_delivered_and_missing_indexes_without_resubmission" ||
    retrieval.fetchOnlyAfterTerminalSuccess !== true || retrieval.sequentialByIndex !== true || retrieval.oneInFlightContentReadPerTask !== true ||
    retrieval.neverPrefetchOrRefetchDeliveredContent !== true || retrieval.handoffBeforeNextContentRead !== true) {
    errors.push(`${label} must retrieve every zero-based requested image content index sequentially after terminal success without prefetching or duplicate delivery`);
  }
}

function validateTaskIdentityStateAndSubmissionFailure(errors, label, contract, image) {
  const identity = contract.taskIdentity;
  if (!identity || identity.receiptField !== "task_id" || identity.preferredResponseFieldSource !== "installed_model_lifecycle_create_idField" ||
    !sameArray(identity.fallbackTopLevelResponseFields, ["task_id", "id"]) || identity.neverDerivesIdFromUrlsNestedObjectsOrPrompt !== true ||
    identity.pathEncoding !== "percent_encode_opaque_id_as_one_path_segment" || identity.whenMissing !== "report_task_id_not_returned_and_do_not_poll_download_or_resubmit") {
    errors.push(`${label} must normalize only a declared top-level task ID and fail closed when it is missing`);
  }
  const state = contract.taskState;
  if (!state || state.declaredStateSource !== "installed_model_lifecycle_poll" ||
    !sameArray(state.fallbackPendingStates, ["pending", "queued", "running", "in_progress"]) ||
    !sameArray(state.fallbackSuccessStates, ["completed", "succeeded", "success"]) ||
    !sameArray(state.fallbackFailureStates, ["failed", "cancelled", "canceled", "expired", "error"]) ||
    state.whenUnrecognized !== "report_raw_state_stop_automatic_polling_and_require_explicit_same_task_continuation") {
    errors.push(`${label} must use declared or bounded fallback task states and stop on an unrecognized state`);
  }
  const failure = contract.submissionFailure;
  if (!failure || failure.modelParameterOrCapabilityRejection !== "one_on_demand_catalog_read_then_require_explicit_corrected_new_request" ||
    failure.rateLimit !== "report_retry_after_when_returned_and_require_explicit_retry" ||
    failure.serverTransportOrTimeoutWithoutTaskId !== "report_submission_outcome_unknown_without_declaring_task_absent_or_resubmitting") {
    errors.push(`${label} must guide rejection, rate limit, and uncertain-submission cases without automatic resubmission`);
  }
  if (image && contract.contentRetrieval?.requestedCount !== "explicit_n_or_default_1") {
    errors.push(`${label} must preserve an explicit image count for same-task delivery`);
  }
}

function validatePolling(errors, label, polling, expected) {
  if (!polling || polling.serverDelay !== "honor_valid_positive_retry_after_before_fallback" || polling.serverDelayMustFitRemainingAutomaticDeadline !== true ||
    !sameArray(polling.fallbackDelaysSeconds, expected.fallbackDelaysSeconds) || polling.steadyDelaySeconds !== expected.steadyDelaySeconds ||
    polling.automaticDeadlineSeconds !== expected.deadlineSeconds || polling.maxAutomaticStatusReads !== expected.maxAutomaticStatusReads ||
    polling.oneInFlightStatusReadPerTask !== true || polling.automaticPollingScope !== "submission_or_explicit_same_task_continuation_turn_only_no_background_timer_or_queue" ||
    polling.afterStatusReadError !== "stop_automatic_polling_and_require_explicit_same_task_continuation" ||
    polling.explicitContinuation !== "new_bounded_same_task_polling_window_only_when_explicitly_requested" ||
    polling.afterDeadline !== "report_pending_and_require_explicit_same_task_continuation") {
    errors.push(`${label} must use the bounded adaptive same-task polling contract without background work or overlapping reads`);
  }
}

function validateRequest(errors, label, request, method, expectedUrl, usesTemplate = false) {
  if (!request || request.method !== method || request[usesTemplate ? "urlTemplate" : "url"] !== expectedUrl) {
    errors.push(`${label} must be ${method} ${expectedUrl}`);
  }
}

function validateRequiredBodyFields(errors, label, request, required) {
  if (!Array.isArray(request?.requiredBodyFields) || required.some((field) => !request.requiredBodyFields.includes(field))) {
    errors.push(`${label} must require ${required.join(", ")}`);
  }
}

function validateBehaviorScenarios(errors, directory, scenarios) {
  const label = `${directory}: behaviorScenarios`;
  if (!scenarios || typeof scenarios !== "object") return;
  if (scenarios.$schema !== "https://puretokensx.com/schemas/media-behavior-scenarios.schema.json") errors.push(`${label} must declare the behavior-scenarios schema`);
  if (scenarios.schemaVersion !== 1 || scenarios.kind !== skillKind(directory) || !Array.isArray(scenarios.scenarios) || !scenarios.scenarios.length) {
    errors.push(`${label} must be a non-empty schemaVersion=1 scenario list for this Skill`);
    return;
  }
  const ids = new Set();
  for (const scenario of scenarios.scenarios) {
    if (!scenario || !/^[a-z][a-z0-9-]*$/.test(scenario.id || "") || typeof scenario.when !== "string" || !scenario.when || typeof scenario.then !== "string" || !scenario.then) {
      errors.push(`${label} scenarios must have a kebab-case id plus non-empty when and then`);
      continue;
    }
    if (ids.has(scenario.id)) errors.push(`${label} duplicate scenario ${scenario.id}`);
    ids.add(scenario.id);
  }
}

function validateModelSelection(errors, directory, selection) {
  const label = `${directory}: modelSelection`;
  if (!selection || typeof selection !== "object") return;
  const capability = skillKind(directory);
  if (selection.$schema !== "https://puretokensx.com/schemas/model-selection.schema.json") errors.push(`${label} must declare the model-selection schema`);
  if (selection.schemaVersion !== 1 || selection.capability !== capability || typeof selection.catalogUpdatedAt !== "string" || typeof selection.catalogCapturedAt !== "string" || !Array.isArray(selection.models) || !selection.models.length) {
    errors.push(`${label} must be a non-empty schemaVersion=1 ${capability} selection`);
    return;
  }
  const ids = new Set();
  for (const model of selection.models) {
    if (!model || typeof model.id !== "string" || !model.id || !Array.isArray(model.aliases) || model.aliases.some((alias) => typeof alias !== "string" || !alias)) {
      errors.push(`${label} models must have a non-empty id and string aliases`);
      continue;
    }
    if (ids.has(model.id)) errors.push(`${label} duplicate model id ${model.id}`);
    ids.add(model.id);
    if (new Set(model.aliases).size !== model.aliases.length) errors.push(`${label} duplicate aliases for ${model.id}`);
    if (Object.hasOwn(model, "parameterSchema") && (!model.parameterSchema || typeof model.parameterSchema !== "object" || Array.isArray(model.parameterSchema))) {
      errors.push(`${label} parameterSchema for ${model.id} must be an object`);
    }
  }
}

function validateTaskReceipt(errors, directory, receipt) {
  const label = `${directory}: taskReceipt`;
  if (!receipt || typeof receipt !== "object") return;
  const kind = skillKind(directory);
  if (receipt.$schema !== "https://puretokensx.com/schemas/task-receipt.schema.json" || receipt.schemaVersion !== 1 || receipt.kind !== kind) {
    errors.push(`${label} must be a schemaVersion=1 ${kind} task receipt`);
    return;
  }
  for (const phase of ["submission", "continuation", "completion", "failure"]) {
    const fields = receipt[phase]?.requiredFields;
    if (!Array.isArray(fields) || receiptCoreFields.some((field) => !fields.includes(field)) || typeof receipt[phase]?.nextAction !== "string" || !receipt[phase].nextAction) {
      errors.push(`${label} ${phase} must include the core user-visible receipt fields`);
    }
    if (phase === "completion" && !fields?.includes("delivered_count")) {
      errors.push(`${label} completion must include delivered_count`);
    }
  }
}

function skillKind(directory) {
  return directory.replace("puretokens_", "");
}

function sameArray(actual, expected) {
  return Array.isArray(actual) && actual.length === expected.length && actual.every((value, index) => value === expected[index]);
}

async function validateSchemaDocuments(errors) {
  const schemas = [
    "schemas/media-execution-contract.schema.json",
    "schemas/media-behavior-scenarios.schema.json",
    "schemas/model-selection.schema.json",
    "schemas/host-support.schema.json",
    "schemas/direct-api-execution-contract.schema.json",
    "schemas/catalog-freshness.schema.json",
    "schemas/balance-snapshot.schema.json",
    "schemas/task-receipt.schema.json"
  ];
  for (const schema of schemas) await verifyJsonFile(errors, schema, `schema ${schema}`);
}

function readFrontmatter(text) {
  const match = text.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n/);
  if (!match) return {};
  const result = {};
  for (const line of match[1].split(/\r?\n/)) {
    const separator = line.indexOf(":");
    if (separator <= 0) continue;
    result[line.slice(0, separator).trim()] = line.slice(separator + 1).trim();
  }
  return result;
}

async function verifyFile(errors, relativePath) {
  const absolutePath = path.resolve(repositoryRoot, relativePath);
  if (!absolutePath.startsWith(`${repositoryRoot}${path.sep}`)) {
    errors.push(`skills/index.json: path escapes repository: ${relativePath}`);
    return;
  }
  try {
    if (!(await stat(absolutePath)).isFile()) errors.push(`skills/index.json: missing file ${relativePath}`);
  } catch {
    errors.push(`skills/index.json: missing file ${relativePath}`);
  }
}

async function verifySkillFile(errors, skillDir, relativePath, label) {
  const absolutePath = path.resolve(skillDir, relativePath);
  if (!absolutePath.startsWith(`${skillDir}${path.sep}`)) {
    errors.push(`${label}: path escapes Skill directory`);
    return undefined;
  }
  try {
    if (!(await stat(absolutePath)).isFile()) {
      errors.push(`${label}: missing file ${relativePath}`);
      return undefined;
    }
    return absolutePath;
  } catch {
    errors.push(`${label}: missing file ${relativePath}`);
    return undefined;
  }
}

async function readSkillReference(errors, skillDir, relativePath, label) {
  const absolutePath = await verifySkillFile(errors, skillDir, relativePath, label);
  if (!absolutePath) return undefined;
  try {
    return JSON.parse(await readFile(absolutePath, "utf8"));
  } catch {
    errors.push(`${label}: must be valid JSON`);
    return undefined;
  }
}

async function verifyJsonFile(errors, relativePath, label) {
  const absolutePath = path.resolve(repositoryRoot, relativePath);
  try {
    JSON.parse(await readFile(absolutePath, "utf8"));
  } catch {
    errors.push(`${label}: missing or invalid JSON`);
  }
}
