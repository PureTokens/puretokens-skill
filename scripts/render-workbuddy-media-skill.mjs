import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { repositoryRoot } from "./skill-registry.mjs";
import { getMediaSkillProvenance, mediaSkillName } from "./media-skill-provenance.mjs";

const workBuddySkillName = "puretokens_workbuddy_router";
const referenceFiles = [
  "references/model-catalog-contract.md",
  "references/direct-cloud-contract.md",
  "references/behavior-scenarios.json",
  "references/natural-language-aliases.json"
];

export async function renderWorkBuddyMediaSkill() {
  const sourceRoot = path.join(repositoryRoot, "skills", mediaSkillName);
  const [sourceSkill, manifestText, workBuddyExecutionRules, derivedFrom] = await Promise.all([
    readFile(path.join(sourceRoot, "SKILL.md"), "utf8"),
    readFile(path.join(sourceRoot, "skill.json"), "utf8"),
    readFile(path.join(sourceRoot, "adapters", "workbuddy-execution.md"), "utf8"),
    getMediaSkillProvenance()
  ]);
  const sourceManifest = JSON.parse(manifestText);
  const body = [workBuddyExecutionRules.trim(), stripFrontmatter(sourceSkill)].join("\n\n");
  const entry = `---\nname: ${workBuddySkillName}\ndescription: Always route eligible WorkBuddy image and video requests through the shared Pure Tokens Media Skill before built-in media tools.\nalwaysApply: true\n---\n\n${body}`;
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
  if (!match) throw new Error(`${mediaSkillName}/SKILL.md must have YAML frontmatter`);
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
