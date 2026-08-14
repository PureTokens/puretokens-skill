import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { renderWorkBuddyMediaSkill, writeWorkBuddyMediaSkill } from "../scripts/render-workbuddy-media-skill.mjs";

test("WorkBuddy output is generated from the shared media source", async () => {
  const { entry, manifest, files } = await renderWorkBuddyMediaSkill();
  assert.match(entry, /^---\nname: puretokens_workbuddy_router\n[\s\S]*alwaysApply: true/m);
  assert.match(entry, /This is the WorkBuddy delivery of the shared Pure Tokens Media Skill/);
  assert.match(entry, /ToolSearch only discovers the deferred MCP tools/);
  assert.match(entry, /DeferExecuteTool/);
  assert.match(entry, /Do not use `show_widget`/);
  assert.match(entry, /稳定的 `request_id`/);
  assert.equal(manifest.derivedFrom.name, "puretokens_media");
  assert.equal(manifest.derivedFrom.version, "0.3.2");
  assert.equal(manifest.sourceSha256, createHash("sha256").update(entry).digest("hex"));
  assert.equal(manifest.mcp.server, "puretokens-image");
  assert.deepEqual(manifest.mcp.tools, [
    "puretokens_list_media_models",
    "puretokens_generate_image",
    "puretokens_image_result",
    "puretokens_generate_video",
    "puretokens_video_result"
  ]);
  assert.equal(files.size, 5);
  assert.ok(files.has("references/natural-language-aliases.json"));
});

test("WorkBuddy renderer writes a complete managed Skill directory", async (t) => {
  const temporaryRoot = await mkdtemp(path.join(os.tmpdir(), "puretokens-workbuddy-skill-"));
  t.after(() => rm(temporaryRoot, { recursive: true, force: true }));
  await writeWorkBuddyMediaSkill(temporaryRoot);
  const manifest = JSON.parse(await readFile(path.join(temporaryRoot, "skill.json"), "utf8"));
  assert.equal(manifest.name, "puretokens_workbuddy_router");
  assert.equal(manifest.sourceSha256.length, 64);
  await readFile(path.join(temporaryRoot, "references", "model-catalog-contract.md"), "utf8");
});
