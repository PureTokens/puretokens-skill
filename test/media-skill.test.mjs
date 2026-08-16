import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { repositoryRoot } from "../scripts/skill-registry.mjs";

const skillRoot = path.join(repositoryRoot, "skills", "puretokens_media");

async function readSkill() {
  return readFile(path.join(skillRoot, "SKILL.md"), "utf8");
}

test("media Skill publishes the complete MCP tool contract", async () => {
  const [skillText, manifestText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "skill.json"), "utf8")
  ]);
  const manifest = JSON.parse(manifestText);
  assert.equal(manifest.mcp.server, "puretokens-image");
  assert.deepEqual(manifest.mcp.tools, [
    "puretokens_list_media_models",
    "puretokens_generate_image",
    "puretokens_image_result",
    "puretokens_generate_video",
    "puretokens_video_result"
  ]);
  assert.equal(manifest.rules.stableRequestIdRequired, true);
  assert.equal(manifest.rules.reuseRequestIdOnHostRetry, true);
  assert.equal(manifest.rules.neverAutoResubmitAfterError, true);
  assert.equal(manifest.rules.neverAutoSwitchModelAfterError, true);
  assert.equal(manifest.rules.successRequiresNativeMediaResult, true);
  assert.equal(manifest.rules.completedMediaRequiresLocalFileDelivery, true);
  assert.equal(manifest.rules.completedImageRequiresNativeImageContent, true);
  assert.equal(manifest.rules.completedMediaReportsExactModel, true);
  assert.equal(manifest.rules.completedMediaReportsLocalDelivery, true);
  assert.equal(manifest.rules.completedVideoUsesBoundedNativeResource, true);
  assert.equal(manifest.rules.usesCuratedNaturalLanguageAliasRegistry, true);
  assert.equal(manifest.rules.usesDeterministicMediaDefaults, true);
  assert.equal(manifest.naturalLanguageAliases, "skills/puretokens_media/references/natural-language-aliases.json");
  assert.match(skillText, /第一步必须调用[：:][\s\S]*puretokens_list_media_models/);
  assert.match(skillText, /稳定的 `request_id`/);
  assert.match(skillText, /同一个 `task_id`/);
  assert.match(skillText, /`structuredContent\.model` 返回的实际使用精确模型 ID/);
  assert.match(skillText, /`type == resource`/);
});

test("media Skill does not advertise tools absent from the media MCP", async () => {
  const skillText = await readSkill();
  const manifest = JSON.parse(await readFile(path.join(skillRoot, "skill.json"), "utf8"));
  assert.doesNotMatch(skillText, /puretokens_get_balance|puretokens_get_model_price/);
  assert.deepEqual(manifest.mcp.tools, [
    "puretokens_list_media_models",
    "puretokens_generate_image",
    "puretokens_image_result",
    "puretokens_generate_video",
    "puretokens_video_result"
  ]);
});

test("natural language aliases resolve only to explicit catalog model IDs", async () => {
  const aliases = JSON.parse(await readFile(
    path.join(skillRoot, "references", "natural-language-aliases.json"),
    "utf8"
  ));
  const image2 = aliases.aliases.find((entry) => entry.phrases.includes("image2"));
  assert.deepEqual(image2, {
    phrases: ["image2", "image 2", "gpt image 2", "openai image 2"],
    capability: "image",
    modelIds: ["gpt-image-2"]
  });
  const aliasesByPhrase = new Map(aliases.aliases.flatMap((entry) => entry.phrases.map((phrase) => [phrase, entry])));
  assert.deepEqual(aliasesByPhrase.get("nano banana pro").modelIds, ["gemini-3-pro-image-preview"]);
  assert.deepEqual(aliasesByPhrase.get("nano banana 2").modelIds, ["gemini-3.1-flash-image-preview"]);
  assert.deepEqual(aliasesByPhrase.get("nano banana").modelIds, [
    "gemini-3-pro-image-preview",
    "gemini-3.1-flash-image-preview"
  ]);
  assert.deepEqual(aliases.defaults, {
    image: {
      modelId: "gpt-image-2",
      when: "The user asks to generate an image without naming a model."
    },
    video: {
      modelId: "grok-imagine-video-1.5",
      when: "The user asks to generate a video without naming a model."
    }
  });
  assert.match(aliases.normalization, /complete registered phrases/i);
});

test("media Skill behavior scenarios cover ambiguity, empty catalog, unavailable MCP, failure, and timeout", async () => {
  const [skillText, scenariosText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "references", "behavior-scenarios.json"), "utf8")
  ]);
  const scenarios = JSON.parse(scenariosText).scenarios;
  assert.deepEqual(scenarios.map((scenario) => scenario.id), [
    "ambiguous-model",
    "no-media-model",
    "mcp-unavailable",
    "task-failure",
    "task-timeout"
  ]);
  assert.deepEqual(scenarios.slice(0, 1).map((scenario) => scenario.firstTool), [
    "puretokens_list_media_models",
  ]);
  assert.match(skillText, /模型不存在或匹配多个/);
  assert.match(skillText, /目录为空/);
  assert.match(skillText, /MCP 不可用/);
  assert.match(skillText, /safeToResubmit=false/);
  assert.match(skillText, /轮询超时/);
  assert.match(skillText, /不得自动换模型/);
  assert.match(skillText, /不得自动重新提交/);
});

test("the shared media source has target-specific Claude and WorkBuddy deliveries", async () => {
  const [skillText, manifestText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "skill.json"), "utf8")
  ]);
  const manifest = JSON.parse(manifestText);
  assert.equal(manifest.distribution.claudeDesktop.format, "zip");
  assert.equal(manifest.distribution.claudeDesktop.archiveRoot, "puretokens_media");
  assert.equal(manifest.distribution.claudeDesktop.enableAfterImport, true);
  assert.equal(manifest.distribution.workbuddy.generatedSkillName, "puretokens_workbuddy_router");
  assert.equal(manifest.distribution.workbuddy.alwaysApply, true);
  assert.match(skillText, /跨客户端共用的媒体行为源/);
  assert.match(skillText, /不得自行写入、替换或删除任何客户端的 Skill、MCP 或 Router 配置/);
});
