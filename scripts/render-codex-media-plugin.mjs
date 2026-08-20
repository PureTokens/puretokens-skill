import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { repositoryRoot } from "./skill-registry.mjs";
import { getMediaSkillProvenance, mediaSkillName, mediaSkillSourceFiles } from "./media-skill-provenance.mjs";

const pluginName = "puretokens-media";
const pluginRoot = path.join(repositoryRoot, "plugins", pluginName);

export async function renderCodexMediaPlugin() {
  const sourceRoot = path.join(repositoryRoot, "skills", mediaSkillName);
  const [manifestText, derivedFrom] = await Promise.all([
    readFile(path.join(sourceRoot, "skill.json"), "utf8"),
    getMediaSkillProvenance()
  ]);
  const sourceManifest = JSON.parse(manifestText);
  const pluginManifest = {
    name: pluginName,
    version: sourceManifest.version,
    description: "Generate images and videos with Pure Tokens through Direct Cloud or the managed media MCP.",
    author: { name: "Pure Tokens" },
    license: "Proprietary",
    keywords: ["image-generation", "video-generation", "mcp", "puretokens"],
    skills: "./skills/",
    interface: {
      displayName: "Pure Tokens Media",
      shortDescription: "Generate images and videos with Pure Tokens.",
      longDescription: "Uses the shared Pure Tokens media Skill to select an exact model and complete one image or video task through Direct Cloud or a managed MCP.",
      developerName: "Pure Tokens",
      category: "Productivity",
      capabilities: ["Interactive"],
      defaultPrompt: [
        "Generate an image with Pure Tokens.",
        "Generate a video with Pure Tokens."
      ],
      brandColor: "#F97316",
      screenshots: []
    },
    derivedFrom
  };
  const pluginSkillManifest = { ...sourceManifest, derivedFrom };
  const files = new Map();
  for (const relativePath of mediaSkillSourceFiles) {
    files.set(
      path.join("skills", mediaSkillName, relativePath),
      await readFile(path.join(sourceRoot, relativePath))
    );
  }
  files.set(
    path.join(".codex-plugin", "plugin.json"),
    Buffer.from(`${JSON.stringify(pluginManifest, null, 2)}\n`)
  );
  files.set(
    path.join("skills", mediaSkillName, "skill.json"),
    Buffer.from(`${JSON.stringify(pluginSkillManifest, null, 2)}\n`)
  );
  return { files, pluginManifest, pluginSkillManifest };
}

export async function writeCodexMediaPlugin(outputRoot = pluginRoot) {
  const destination = path.resolve(outputRoot);
  const { files, pluginManifest } = await renderCodexMediaPlugin();
  await rm(destination, { recursive: true, force: true });
  for (const [relativePath, contents] of files) {
    const target = path.join(destination, relativePath);
    await mkdir(path.dirname(target), { recursive: true });
    await writeFile(target, contents);
  }
  return pluginManifest;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const outputIndex = process.argv.indexOf("--out");
  const outputRoot = outputIndex >= 0 ? process.argv[outputIndex + 1] : undefined;
  if (outputIndex >= 0 && (!outputRoot || outputRoot.startsWith("--"))) {
    throw new Error("--out requires a directory");
  }
  const manifest = await writeCodexMediaPlugin(outputRoot);
  process.stdout.write(`Rendered ${pluginName} ${manifest.version} to ${path.resolve(outputRoot || pluginRoot)}\n`);
}
