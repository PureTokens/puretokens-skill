import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { renderCodexMediaPlugin } from "../scripts/render-codex-media-plugin.mjs";
import { repositoryRoot } from "../scripts/skill-registry.mjs";

const sourceRoot = path.join(repositoryRoot, "skills", "puretokens_media");
const pluginRoot = path.join(repositoryRoot, "plugins", "puretokens-media");
const pluginSkillRoot = path.join(pluginRoot, "skills", "puretokens_media");

test("Codex Plugin is a generated media delivery of the shared media Skill", async () => {
  const [pluginText, marketplaceText, agentManifest, mcpConfigText, sourceSkillText, sourceManifestText, pluginManifestText] = await Promise.all([
    readFile(path.join(pluginRoot, ".codex-plugin", "plugin.json"), "utf8"),
    readFile(path.join(repositoryRoot, ".agents", "plugins", "marketplace.json"), "utf8"),
    readFile(path.join(pluginSkillRoot, "agents", "openai.yaml"), "utf8"),
    readFile(path.join(pluginRoot, ".mcp.json"), "utf8"),
    readFile(path.join(sourceRoot, "SKILL.md"), "utf8"),
    readFile(path.join(sourceRoot, "skill.json"), "utf8"),
    readFile(path.join(pluginSkillRoot, "skill.json"), "utf8")
  ]);
  const plugin = JSON.parse(pluginText);
  const marketplace = JSON.parse(marketplaceText);
  const sourceManifest = JSON.parse(sourceManifestText);
  const pluginSkillManifest = JSON.parse(pluginManifestText);
  const mcpConfig = JSON.parse(mcpConfigText);

  assert.equal(plugin.name, "puretokens-media");
  assert.equal(plugin.version, "0.4.0");
  assert.equal(plugin.skills, "./skills/");
  assert.equal(plugin.mcpServers, "./.mcp.json");
  assert.match(plugin.description, /Direct Cloud or the managed media MCP/);
  assert.deepEqual(plugin.derivedFrom, {
    name: "puretokens_media",
    version: sourceManifest.version,
    sourceSha256: createHash("sha256").update(sourceSkillText).digest("hex")
  });
  assert.equal(marketplace.name, "puretokens");
  assert.deepEqual(marketplace.plugins, [{
    name: "puretokens-media",
    source: { source: "local", path: "./plugins/puretokens-media" },
    policy: { installation: "AVAILABLE", authentication: "ON_INSTALL" },
    category: "Productivity"
  }]);
  assert.match(agentManifest, /value: "puretokens-image"/);
  assert.match(agentManifest, /allow_implicit_invocation: true/);
  assert.deepEqual(mcpConfig, {
    mcpServers: {
      "puretokens-image": {
        command: "puretokens-mcp",
        args: [],
        env_vars: ["PURETOKENS_API_BASE_URL", "PURETOKENS_ACCESS_TOKEN"]
      }
    }
  });

  for (const relativePath of [
    "SKILL.md",
    "references/behavior-scenarios.json",
    "references/direct-cloud-contract.md",
    "references/model-catalog-contract.md",
    "references/natural-language-aliases.json"
  ]) {
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
  for (const [relativePath, generated] of files) {
    const disk = await readFile(path.join(pluginRoot, relativePath));
    assert.deepEqual(disk, generated, relativePath);
  }
});
