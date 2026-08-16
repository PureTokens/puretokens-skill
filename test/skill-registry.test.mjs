import assert from "node:assert/strict";
import test from "node:test";
import { collectSkillRecords, validateRepository } from "../scripts/skill-registry.mjs";

test("the skill registry has one shared media source for Claude Desktop and WorkBuddy", async () => {
  const { registry, records } = await collectSkillRecords();
  assert.equal(registry.schemaVersion, 1);
  assert.deepEqual(registry.skills.map((skill) => skill.name), ["puretokens_media"]);
  assert.equal(records.length, 1);
  const media = records.find((record) => record.manifest.name === "puretokens_media");
  assert.equal(registry.skills[0].version, media.manifest.version);
  assert.equal(media.frontmatter.name, "puretokens_media");
  assert.equal(media.manifest.distribution.workbuddy.generatedSkillName, "puretokens_workbuddy_router");
  assert.equal(media.manifest.distribution.workbuddy.alwaysApply, true);
});

test("the repository passes its safety and registry validation", async () => {
  assert.deepEqual(await validateRepository(), []);
});
