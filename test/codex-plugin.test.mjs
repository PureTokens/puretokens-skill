import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { renderCodexMediaPlugin } from "../scripts/render-codex-media-plugin.mjs";
import { getMediaSkillProvenance, mediaSkillSourceFiles } from "../scripts/media-skill-provenance.mjs";
import { repositoryRoot } from "../scripts/skill-registry.mjs";

const sourceRoot = path.join(repositoryRoot, "skills", "puretokens_media");
const pluginRoot = path.join(repositoryRoot, "plugins", "puretokens-media");
const pluginSkillRoot = path.join(pluginRoot, "skills", "puretokens_media");

test("Codex Plugin is a generated media delivery of the shared media Skill", async () => {
  const [pluginText, marketplaceText, sourceManifestText, pluginManifestText, sourceProvenance] = await Promise.all([
    readFile(path.join(pluginRoot, ".codex-plugin", "plugin.json"), "utf8"),
    readFile(path.join(repositoryRoot, ".agents", "plugins", "marketplace.json"), "utf8"),
    readFile(path.join(sourceRoot, "skill.json"), "utf8"),
    readFile(path.join(pluginSkillRoot, "skill.json"), "utf8"),
    getMediaSkillProvenance()
  ]);
  const plugin = JSON.parse(pluginText);
  const marketplace = JSON.parse(marketplaceText);
  const sourceManifest = JSON.parse(sourceManifestText);
  const pluginSkillManifest = JSON.parse(pluginManifestText);

  assert.equal(plugin.name, "puretokens-media");
  assert.equal(plugin.version, sourceManifest.version);
  assert.equal(plugin.skills, "./skills/");
  assert.equal(plugin.mcpServers, undefined);
  assert.match(plugin.description, /Direct Cloud or the managed media MCP/);
  assert.deepEqual(plugin.derivedFrom, sourceProvenance);
  assert.equal(marketplace.name, "puretokens");
  assert.deepEqual(marketplace.plugins, [{
    name: "puretokens-media",
    source: { source: "local", path: "./plugins/puretokens-media" },
    policy: { installation: "AVAILABLE", authentication: "ON_INSTALL" },
    category: "Productivity"
  }]);
  for (const relativePath of mediaSkillSourceFiles.filter((file) => file !== "skill.json")) {
    const [source, pluginDelivery] = await Promise.all([
      readFile(path.join(sourceRoot, relativePath), "utf8"),
      readFile(path.join(pluginSkillRoot, relativePath), "utf8")
    ]);
    assert.equal(pluginDelivery, source, relativePath);
  }
  const { derivedFrom, ...deliveryWithoutSourceMarker } = pluginSkillManifest;
  assert.deepEqual(deliveryWithoutSourceMarker, sourceManifest);
  assert.deepEqual(derivedFrom, plugin.derivedFrom);
});

test("Codex Plugin renderer has no hand-maintained behavior copy", async () => {
  const { files } = await renderCodexMediaPlugin();
  assert.equal(files.has(".mcp.json"), false);
  for (const [relativePath, generated] of files) {
    const disk = await readFile(path.join(pluginRoot, relativePath));
    assert.deepEqual(disk, generated, relativePath);
  }
});
