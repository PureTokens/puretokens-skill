import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { repositoryRoot } from "./skill-registry.mjs";

const sourceSkillName = "puretokens_media";
const workBuddySkillName = "puretokens_workbuddy_router";
const referenceFiles = [
  "references/model-catalog-contract.md",
  "references/direct-cloud-contract.md",
  "references/behavior-scenarios.json",
  "references/natural-language-aliases.json"
];

export async function renderWorkBuddyMediaSkill() {
  const sourceRoot = path.join(repositoryRoot, "skills", sourceSkillName);
  const [sourceSkill, manifestText] = await Promise.all([
    readFile(path.join(sourceRoot, "SKILL.md"), "utf8"),
    readFile(path.join(sourceRoot, "skill.json"), "utf8")
  ]);
  const sourceManifest = JSON.parse(manifestText);
  const workBuddyExecutionRules = [
    "## WorkBuddy execution requirement",
    "",
    "ToolSearch only discovers the deferred MCP tools; it neither invokes the MCP nor starts a media-model request.",
    "In WorkBuddy, invoke every discovered Pure Tokens MCP tool through `DeferExecuteTool`, never by merely naming it in text.",
    "Use these exact deferred tool names: `mcp__puretokens-image__puretokens_list_media_models`, `mcp__puretokens-image__puretokens_generate_image`, `mcp__puretokens-image__puretokens_image_result`, `mcp__puretokens-image__puretokens_generate_video`, and `mcp__puretokens-image__puretokens_video_result`.",
    "The required order is: ToolSearch -> DeferExecuteTool(list_media_models) -> choose the exact catalog model -> DeferExecuteTool(generate) -> DeferExecuteTool(result) until the native result arrives.",
    "Do not use `show_widget`, `ImageGen`, `VideoGen`, `WebSearch`, SVG, HTML, or a text-only response as a fallback for a Pure Tokens media request.",
    "A Pure Tokens image/video is successful only after the corresponding `DeferExecuteTool` result contains the MCP-returned exact model and its native result or local delivery metadata.",
    "A discovered tool, a tool name in an answer, a task-less response, or a rendered widget is never proof that a model was called. If deferred invocation is unavailable or fails, report that state and do not fabricate a visual result or claim Pure Tokens generated it."
  ].join("\n");
  const body = [workBuddyExecutionRules, stripFrontmatter(sourceSkill)].join("\n\n");
  const entry = `---\nname: ${workBuddySkillName}\ndescription: Always route eligible WorkBuddy image and video requests through the shared Pure Tokens Media Skill before built-in media tools.\nalwaysApply: true\n---\n\n<system_reminder>\n## WorkBuddy media routing\n\nThis is the WorkBuddy delivery of the shared Pure Tokens Media Skill. Apply these routing rules only when the user asks to generate or create an image or video, asks for a media model, or explicitly mentions Pure Tokens media. Do not affect normal text, coding, analysis, image understanding, or editing requests.\n\n1. Before WorkBuddy built-in \`ImageGen\` or \`VideoGen\`, discover the \`puretokens-image\` MCP tools with \`ToolSearch\` and use the matching Pure Tokens tool. Use the built-in tool only when the user explicitly asks for that built-in tool.\n2. Follow the shared workflow below without creating a second model-selection or retry policy. Its live catalog, exact-model, single-submission, polling, delivery, and failure rules are authoritative.\n3. Do not expose credentials, Router tokens, local authorization URLs, or upstream URLs.\n</system_reminder>\n\n${body}`;
  const derivedFrom = {
    name: sourceManifest.name,
    version: sourceManifest.version,
    sourceSha256: sha256(sourceSkill)
  };
  const manifest = {
    schemaVersion: 1,
    name: workBuddySkillName,
    version: sourceManifest.version,
    entry: "SKILL.md",
    sourceSha256: sha256(entry),
    derivedFrom,
    displayName: "Pure Tokens Media",
    description: "Desktop-managed WorkBuddy delivery of the shared Pure Tokens media behavior.",
    mcp: sourceManifest.mcp,
    rules: sourceManifest.rules,
    supportedClients: ["workbuddy"],
    managedBy: "puretokens-desktop",
    distribution: {
      workbuddy: {
        managedByDesktop: true,
        alwaysApply: true,
        installRoot: "~/.workbuddy/skills"
      }
    }
  };
  const files = new Map([
    ["SKILL.md", Buffer.from(entry)],
    ["skill.json", Buffer.from(`${JSON.stringify(manifest, null, 2)}\n`)]
  ]);
  for (const relativePath of referenceFiles) {
    files.set(relativePath, await readFile(path.join(sourceRoot, relativePath)));
  }
  return { files, manifest, entry };
}

export async function writeWorkBuddyMediaSkill(outputRoot) {
  const { files, manifest } = await renderWorkBuddyMediaSkill();
  const destination = path.resolve(outputRoot);
  for (const [relativePath, contents] of files) {
    const filePath = path.join(destination, relativePath);
    await mkdir(path.dirname(filePath), { recursive: true });
    await writeFile(filePath, contents);
  }
  return manifest;
}

function stripFrontmatter(skillText) {
  const match = skillText.match(/^---\r?\n[\s\S]*?\r?\n---\r?\n([\s\S]*)$/);
  if (!match) throw new Error(`${sourceSkillName}/SKILL.md must have YAML frontmatter`);
  return match[1].replace(/^\s+/, "");
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const outputIndex = process.argv.indexOf("--out");
  const outputRoot = outputIndex >= 0 ? process.argv[outputIndex + 1] : undefined;
  if (!outputRoot || outputRoot.startsWith("--") || process.argv.length !== 4) {
    throw new Error("Usage: node scripts/render-workbuddy-media-skill.mjs --out <directory>");
  }
  await writeWorkBuddyMediaSkill(outputRoot);
  process.stdout.write(`Rendered WorkBuddy media Skill to ${path.resolve(outputRoot)}\n`);
}
