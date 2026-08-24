import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, readFile, rm, stat, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { renderWorkBuddyMediaSkill, writeWorkBuddyMediaSkill } from "../scripts/render-workbuddy-media-skill.mjs";
import { getMediaSkillProvenance } from "../scripts/media-skill-provenance.mjs";
import { repositoryRoot } from "../scripts/skill-registry.mjs";

const sourceRoot = path.join(repositoryRoot, "skills", "puretokens_media");
const workBuddyReferenceFiles = [
  "references/model-catalog-contract.md",
  "references/direct-cloud-contract.md",
  "references/behavior-scenarios.json",
  "references/natural-language-aliases.json",
  "references/balance.md",
  "references/image.md",
  "references/video.md"
];

test("WorkBuddy output is generated from the shared media source", async () => {
  const [rendered, sourceProvenance] = await Promise.all([
    renderWorkBuddyMediaSkill(),
    getMediaSkillProvenance()
  ]);
  const { entry, manifest, files } = rendered;
  assert.match(entry, /^---\nname: puretokens_workbuddy_router\n[\s\S]*alwaysApply: true/m);
  assert.match(entry, /This is the WorkBuddy delivery of the shared Pure Tokens Media Skill/);
  assert.match(entry, /Preserve an explicit WorkBuddy choice/);
  assert.match(entry, /ToolSearch only discovers the deferred MCP tools/);
  assert.match(entry, /DeferExecuteTool/);
  assert.match(entry, /Do not use `show_widget`/);
  assert.match(entry, /这是一个路由 Skill/);
  assert.match(entry, /`references\/balance\.md`/);
  assert.match(entry, /`references\/image\.md`/);
  assert.match(entry, /`references\/video\.md`/);
  assert.deepEqual(manifest.derivedFrom, sourceProvenance);
  assert.equal(manifest.sourceSha256, createHash("sha256").update(entry).digest("hex"));
  assert.equal(manifest.mcp.server, "puretokens-image");
  assert.deepEqual(manifest.mcp.tools, [
    "puretokens_list_media_models",
    "puretokens_generate_image",
    "puretokens_image_result",
    "puretokens_generate_video",
    "puretokens_video_result"
  ]);
  assert.equal(files.size, 9);
  assert.ok(files.has("references/direct-cloud-contract.md"));
  assert.ok(files.has("references/natural-language-aliases.json"));
  assert.ok(files.has("references/balance.md"));
  assert.ok(files.has("references/image.md"));
  assert.ok(files.has("references/video.md"));
  const sourceSkill = await readFile(path.join(sourceRoot, "SKILL.md"), "utf8");
  const sharedBody = sourceSkill.replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n\s*/, "");
  assert.ok(entry.endsWith(sharedBody));
  for (const relativePath of workBuddyReferenceFiles) {
    assert.deepEqual(files.get(relativePath), await readFile(path.join(sourceRoot, relativePath)), relativePath);
  }
});

test("WorkBuddy renderer writes a complete managed Skill directory", async (t) => {
  const temporaryRoot = await mkdtemp(path.join(os.tmpdir(), "puretokens-workbuddy-skill-"));
  t.after(() => rm(temporaryRoot, { recursive: true, force: true }));
  const staleFile = path.join(temporaryRoot, "obsolete-rule.md");
  await writeFile(staleFile, "stale behavior\n");
  await writeWorkBuddyMediaSkill(temporaryRoot);
  const manifest = JSON.parse(await readFile(path.join(temporaryRoot, "skill.json"), "utf8"));
  assert.equal(manifest.name, "puretokens_workbuddy_router");
  assert.equal(manifest.distribution.workbuddy.manualInstallationSupported, true);
  assert.equal(manifest.sourceSha256.length, 64);
  await readFile(path.join(temporaryRoot, "references", "model-catalog-contract.md"), "utf8");
  await assert.rejects(stat(staleFile));
});
