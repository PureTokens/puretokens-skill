import { createHash } from "node:crypto";
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { repositoryRoot } from "./skill-registry.mjs";

const sourceSkillName = "puretokens_media";
const pluginName = "puretokens-media";
const pluginRoot = path.join(repositoryRoot, "plugins", pluginName);
const sourceFiles = [
  "SKILL.md",
  "skill.json",
  "references/behavior-scenarios.json",
  "references/direct-cloud-contract.md",
  "references/model-catalog-contract.md",
  "references/natural-language-aliases.json"
];

export async function renderCodexMediaPlugin() {
  const sourceRoot = path.join(repositoryRoot, "skills", sourceSkillName);
  const [skillText, manifestText] = await Promise.all([
    readFile(path.join(sourceRoot, "SKILL.md"), "utf8"),
    readFile(path.join(sourceRoot, "skill.json"), "utf8")
  ]);
  const sourceManifest = JSON.parse(manifestText);
  const derivedFrom = {
    name: sourceManifest.name,
    version: sourceManifest.version,
    sourceSha256: sha256(skillText)
  };
  const pluginManifest = {
    name: pluginName,
    version: sourceManifest.version,
    description: "Generate images and videos with Pure Tokens through Direct Cloud or the managed media MCP.",
    author: { name: "Pure Tokens" },
    license: "Proprietary",
    keywords: ["image-generation", "video-generation", "mcp", "puretokens"],
    skills: "./skills/",
    mcpServers: "./.mcp.json",
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
  const mcpConfig = {
    mcpServers: {
      "puretokens-image": {
        command: "puretokens-mcp",
        args: [],
        env_vars: [
          "PURETOKENS_API_BASE_URL",
          "PURETOKENS_ACCESS_TOKEN"
        ]
      }
    }
  };
  const desktopAgent = [
    "interface:",
    "  display_name: \"Pure Tokens Media\"",
    "  short_description: \"Generate images and videos with the media models available through Pure Tokens.\"",
    "  default_prompt: \"Use $puretokens_media to generate an image or video through Pure Tokens.\"",
    "",
    "dependencies:",
    "  tools:",
    "    - type: \"mcp\"",
    "      value: \"puretokens-image\"",
    "      description: \"Pure Tokens managed media MCP server\"",
    "      transport: \"stdio\"",
    "",
    "policy:",
    "  allow_implicit_invocation: true",
    ""
  ].join("\n");
  const files = new Map();
  for (const relativePath of sourceFiles) {
    files.set(
      path.join("skills", sourceSkillName, relativePath),
      await readFile(path.join(sourceRoot, relativePath))
    );
  }
  files.set(
    path.join(".codex-plugin", "plugin.json"),
    Buffer.from(`${JSON.stringify(pluginManifest, null, 2)}\n`)
  );
  files.set(".mcp.json", Buffer.from(`${JSON.stringify(mcpConfig, null, 2)}\n`));
  files.set(
    path.join("skills", sourceSkillName, "skill.json"),
    Buffer.from(`${JSON.stringify(pluginSkillManifest, null, 2)}\n`)
  );
  files.set(
    path.join("skills", sourceSkillName, "agents", "openai.yaml"),
    Buffer.from(desktopAgent)
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

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
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
