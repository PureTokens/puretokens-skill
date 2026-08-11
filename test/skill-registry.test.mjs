import assert from "node:assert/strict";
import test from "node:test";
import { collectSkillRecords, validateRepository } from "../scripts/skill-registry.mjs";

test("the skill registry has one valid media skill", async () => {
  const { registry, records } = await collectSkillRecords();
  assert.equal(registry.schemaVersion, 1);
  assert.deepEqual(registry.skills.map((skill) => skill.name), ["puretokens_media"]);
  assert.equal(records.length, 1);
  assert.equal(records[0].manifest.name, "puretokens_media");
  assert.equal(registry.skills[0].version, records[0].manifest.version);
  assert.equal(records[0].frontmatter.name, "puretokens_media");
});

test("the repository passes its safety and registry validation", async () => {
  assert.deepEqual(await validateRepository(), []);
});
