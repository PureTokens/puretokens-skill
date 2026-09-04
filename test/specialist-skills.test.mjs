import assert from "node:assert/strict";
import { execFile as execFileCallback } from "node:child_process";
import { mkdir, mkdtemp, readFile, readdir, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { promisify } from "node:util";

import { collectSkillRecords, repositoryRoot, validateRepository } from "../scripts/skill-registry.mjs";

const execFile = promisify(execFileCallback);
const skillNames = [
  "puretokens-balance",
  "puretokens-connection",
  "puretokens-models",
  "puretokens-image",
  "puretokens-video",
  "puretokens-update"
];
const apiSkillNames = skillNames.filter((name) => name !== "puretokens-update");
const apiOrigin = "https://api.puretokensx.com";

test("repository validates its host-native direct API contract", async () => {
  assert.deepEqual(await validateRepository(), []);
});

test("all specialist Skills are versioned together and use the expected source files", async () => {
  const packageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "package.json"), "utf8"));
  const { registry, records } = await collectSkillRecords();
  assert.equal(packageManifest.version, "0.14.1");
  assert.deepEqual(registry.skills.map((entry) => entry.name), skillNames);
  assert.equal(records.length, skillNames.length);
  for (const record of records) {
    assert.equal(record.manifest.version, packageManifest.version);
    assert.equal(record.frontmatter.name, record.manifest.name);
  }
});

test("API Skills use host-native authenticated HTTPS without a shipped execution runtime", async () => {
  for (const name of apiSkillNames) {
    const base = path.join(repositoryRoot, "skills", name);
    const [skill, manifest, contract] = await Promise.all([
      readFile(path.join(base, "SKILL.md"), "utf8"),
      readFile(path.join(base, "skill.json"), "utf8").then(JSON.parse),
      readFile(path.join(base, "references", "execution-contract.json"), "utf8").then(JSON.parse)
    ]);
    assert.match(skill, /https:\/\/api\.puretokensx\.com/);
    assert.equal(manifest.rules.usesHostNativeAuthenticatedHttpsExecution, true);
    assert.equal(manifest.rules.usesCurrentHostConfiguredCredential, true);
    assert.equal(manifest.rules.doesNotUseNodeRuntime, true);
    assert.equal(contract.transport.fixedApiOrigin, apiOrigin);
    assert.equal(contract.transport.usesFullApiUrls, true);
    assert.equal(contract.transport.usesHostNativeAuthenticatedHttpsExecution, true);
    assert.equal(contract.transport.usesCurrentHostConfiguredCredential, true);
    assert.equal(contract.transport.credentialSourceNeverBecomesRequestTarget, true);
    assert.equal(contract.transport.neverExposesCredentialsOrHostConfiguration, true);
    assert.equal(contract.transport.doesNotUseNodeRuntime, true);
    assert.equal(contract.transport.doesNotUseMcpOrFallbackTransport, true);
    assert.doesNotMatch(skill, /puretokens-direct-api\.mjs|\.puretokens-runtime|--json-base64|--multipart-base64/);
  }
});

test("media contracts retain fixed endpoints, asynchronous same-task handling, and native delivery", async () => {
  const image = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-image", "references", "execution-contract.json"), "utf8"));
  const video = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-video", "references", "execution-contract.json"), "utf8"));
  assert.equal(image.operations.submit.url, `${apiOrigin}/v1/images/generations`);
  assert.equal(image.operations.status.urlTemplate, `${apiOrigin}/v1/images/{task_id}`);
  assert.equal(image.operations.content.urlTemplate, `${apiOrigin}/v1/images/{task_id}/content?index={index}`);
  assert.equal(image.parameterValidation.defaultModel, "gpt-image-2");
  assert.equal(image.contentRetrieval.indexBase, 0);
  assert.equal(image.contentRetrieval.completeDeliveryRequiresEveryRequestedIndex, true);
  assert.equal(image.submissionFailure.hostExecutionPolicyBlockedBeforeRequestStarts, "report_validation_no_api_request_or_task_id_and_require_host_session_with_external_network_permission");
  assert.equal(video.operations.submit.url, `${apiOrigin}/v1/videos`);
  assert.equal(video.operations.status.urlTemplate, `${apiOrigin}/v1/videos/{task_id}`);
  assert.equal(video.operations.content.urlTemplate, `${apiOrigin}/v1/videos/{task_id}/content`);
  assert.equal(video.inputMediaValidation.nativeAttachmentBodyMode, "host_native_multipart_only");
  assert.equal(video.submissionFailure.hostExecutionPolicyBlockedBeforeRequestStarts, "report_validation_no_api_request_or_task_id_and_require_host_session_with_external_network_permission");
  for (const contract of [image, video]) {
    assert.equal(contract.result.sameTaskOnly, true);
    assert.equal(contract.result.neverAutoResubmit, true);
    assert.equal(contract.transport.requiresNativeMediaByteDelivery, true);
  }
});

test("media Skills load a compact index and only the selected model profile", async () => {
  for (const [name, capability, defaultModel] of [
    ["puretokens-image", "image", "gpt-image-2"],
    ["puretokens-video", "video", "grok-imagine-video-1.5-preview"]
  ]) {
    const base = path.join(repositoryRoot, "skills", name);
    const [skill, manifest, index] = await Promise.all([
      readFile(path.join(base, "SKILL.md"), "utf8"),
      readFile(path.join(base, "skill.json"), "utf8").then(JSON.parse),
      readFile(path.join(base, "references", "model-index.json"), "utf8").then(JSON.parse)
    ]);
    assert.equal(manifest.modelIndex, "references/model-index.json");
    assert.equal(manifest.rules.usesLazyModelProfileLoading, true);
    assert.equal(manifest.rules.loadsBehaviorScenariosOnDemandOnly, true);
    assert.equal(index.capability, capability);
    assert.equal(index.defaultModel, defaultModel);
    assert.equal(index.models.some((model) => model.id === defaultModel), true);
    for (const model of index.models) {
      const profile = JSON.parse(await readFile(path.join(base, "references", model.profile), "utf8"));
      assert.equal(profile.id, model.id);
      assert.equal(profile.capability, capability);
      assert.equal(typeof profile.parameterSchema, "object");
    }
    assert.match(skill, /不要在每个请求加载全部/);
    assert.match(skill, /普通核心 POST 可只依据本 Skill、索引和选中 profile 执行/);
    assert.match(skill, /只有出现对应异常.*才读 `references\/behavior-scenarios\.json`/);
    await assert.rejects(readFile(path.join(base, "references", "model-selection.json"), "utf8"));
  }
});

test("legacy migration archive contains only the current source-only installer and Skills", async (t) => {
  const archive = path.join(repositoryRoot, "dist", "puretokens-skill-install.zip");
  const packageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "package.json"), "utf8"));
  const { stdout: listing } = await execFile("unzip", ["-Z1", archive], { cwd: repositoryRoot });
  const entries = new Set(listing.trim().split("\n"));
  const prefix = "puretokens-skill-main/";
  for (const required of [
    `${prefix}README.md`,
    `${prefix}package.json`,
    `${prefix}runtime/puretokens-skill-install.sh`,
    `${prefix}runtime/puretokens-skill-install.ps1`,
    `${prefix}skills/index.json`,
    `${prefix}skills/puretokens-image/references/model-index.json`,
    `${prefix}skills/puretokens-video/references/model-index.json`
  ]) assert.equal(entries.has(required), true, `missing ${required}`);
  for (const retired of [
    `${prefix}runtime/runtime.json`,
    `${prefix}runtime/puretokens-direct-api.mjs`,
    `${prefix}skills/puretokens-image/references/model-selection.json`,
    `${prefix}skills/puretokens-video/references/model-selection.json`
  ]) assert.equal(entries.has(retired), false, `retired archive entry ${retired}`);
  const { stdout: index } = await execFile("unzip", ["-p", archive, `${prefix}skills/index.json`], { cwd: repositoryRoot });
  assert.equal(JSON.parse(index).skills.every((skill) => skill.version === packageManifest.version), true);

  const unpacked = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-legacy-archive-"));
  const target = path.join(unpacked, "target");
  t.after(() => rm(unpacked, { recursive: true, force: true }));
  await execFile("unzip", ["-q", archive, "-d", unpacked], { cwd: repositoryRoot });
  const source = path.join(unpacked, "puretokens-skill-main");
  const installer = path.join(source, "runtime", "puretokens-skill-install.sh");
  const { stdout } = await execFile("sh", [installer, "sync", "--target", target, "--source", source], { cwd: source });
  assert.match(stdout, /Pure Tokens Skills 0\.14\.1 synchronized at/);
  await assert.rejects(readFile(path.join(target, ".puretokens-runtime", "runtime.json"), "utf8"));
});

test("the public install prompt remains extractable in both README files", async () => {
  const expected = "Install or update the official Pure Tokens Skills from https://github.com/PureTokens/puretokens-skill.";
  for (const file of ["README.md", "README.zh-CN.md"]) {
    const readme = await readFile(path.join(repositoryRoot, file), "utf8");
    const match = readme.match(/### (?:Copy this to a terminal-capable local agent|复制给可在本机执行命令的 Agent)\n\n```text\n([^\n]+)\n```/);
    assert.equal(match?.[1], expected);
  }
});

test("host matrix lists exactly seven host-native direct API hosts", async () => {
  const support = JSON.parse(await readFile(path.join(repositoryRoot, "references", "host-support.json"), "utf8"));
  assert.deepEqual(support.supported.map((host) => host.id), ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode", "trae"]);
  for (const host of support.supported) {
    assert.equal(host.delivery, "manual-source");
    assert.equal(host.directMediaExecution, "host-native-direct-api");
  }
});

test("source installer synchronizes Skills without copying an API runtime", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-host-native-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const installer = path.join(repositoryRoot, "runtime", "puretokens-skill-install.sh");
  const { stdout } = await execFile("sh", [installer, "sync", "--target", target, "--source", repositoryRoot], { cwd: repositoryRoot });
  assert.match(stdout, /Pure Tokens Skills 0\.14\.1 synchronized at/);
  assert.deepEqual((await readdir(target)).filter((entry) => skillNames.includes(entry)).sort(), [...skillNames].sort());
  await assert.rejects(readFile(path.join(target, ".puretokens-runtime", "runtime.json"), "utf8"));
});

test("source installer removes only the verified retired Node runtime", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-legacy-runtime-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const verified = path.join(target, ".puretokens-runtime");
  await mkdir(verified);
  await writeFile(path.join(verified, "runtime.json"), JSON.stringify({ name: "puretokens-direct-api-runtime" }));
  await writeFile(path.join(verified, "puretokens-direct-api.mjs"), "retired runtime\n");
  const installer = path.join(repositoryRoot, "runtime", "puretokens-skill-install.sh");
  const { stdout } = await execFile("sh", [installer, "sync", "--target", target, "--source", repositoryRoot], { cwd: repositoryRoot });
  assert.match(stdout, /Removed retired managed Node runtime/);
  await assert.rejects(readFile(path.join(verified, "runtime.json"), "utf8"));

  const unknown = path.join(target, ".puretokens-runtime");
  await mkdir(unknown);
  await writeFile(path.join(unknown, "runtime.json"), JSON.stringify({ name: "someone-elses-runtime" }));
  await execFile("sh", [installer, "sync", "--target", target, "--source", repositoryRoot], { cwd: repositoryRoot });
  assert.equal(JSON.parse(await readFile(path.join(unknown, "runtime.json"), "utf8")).name, "someone-elses-runtime");
});

test("catalog freshness gate remains present for releases", async () => {
  const policy = JSON.parse(await readFile(path.join(repositoryRoot, "references", "catalog-freshness.json"), "utf8"));
  assert.equal(policy.maxAgeDays, 7);
  assert.equal(policy.releaseCommand, "npm run release:validate");
});
