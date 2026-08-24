import { createHash, randomUUID } from "node:crypto";
import { mkdir, readFile, rename, rm, stat, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { repositoryRoot } from "./skill-registry.mjs";
import { getManagedSkillProvenances, managedSkillNames } from "./media-skill-provenance.mjs";

const workBuddySkillName = "puretokens_workbuddy_router";
export async function renderWorkBuddyMediaSkill() {
  const [sourceSkills, derivedFrom] = await Promise.all([
    Promise.all(managedSkillNames.map((name) => readFile(path.join(repositoryRoot, "skills", name, "SKILL.md"), "utf8"))),
    getManagedSkillProvenances()
  ]);
  const entry = `---\nname: ${workBuddySkillName}\ndescription: Apply the three Pure Tokens API Skills through the current configured connection.\nalwaysApply: true\n---\n\n${sourceSkills.map(stripFrontmatter).join("\n\n")}`;
  const manifest = {
    schemaVersion: 1,
    name: workBuddySkillName,
    version: "0.6.0",
    entry: "SKILL.md",
    sourceSha256: sha256(entry),
    derivedFrom,
    displayName: "Pure Tokens Media",
    description: "Generated WorkBuddy delivery of the three Pure Tokens API Skills.",
    rules: { usesCurrentConfiguredConnection: true, noFallbackSubmission: true },
    supportedClients: ["workbuddy"],
    managedBy: "puretokens-desktop",
    distribution: {
      workbuddy: {
        managedByDesktop: true,
        manualInstallationSupported: true,
        alwaysApply: true,
        installRoot: "~/.workbuddy/skills"
      }
    }
  };
  const files = new Map([
    ["SKILL.md", Buffer.from(entry)],
    ["skill.json", Buffer.from(`${JSON.stringify(manifest, null, 2)}\n`)]
  ]);
  return { files, manifest, entry };
}

export async function writeWorkBuddyMediaSkill(outputRoot) {
  const { files, manifest } = await renderWorkBuddyMediaSkill();
  const destination = path.resolve(outputRoot);
  const parent = path.dirname(destination);
  const name = path.basename(destination);
  const staging = path.join(parent, `.${name}.staging-${randomUUID()}`);
  const backup = path.join(parent, `.${name}.backup-${randomUUID()}`);
  await mkdir(parent, { recursive: true });
  try {
    for (const [relativePath, contents] of files) {
      const filePath = path.join(staging, relativePath);
      await mkdir(path.dirname(filePath), { recursive: true });
      await writeFile(filePath, contents);
    }
    if (await exists(destination)) {
      await rename(destination, backup);
      try {
        await rename(staging, destination);
      } catch (error) {
        await rename(backup, destination).catch(() => undefined);
        throw error;
      }
      await rm(backup, { recursive: true, force: false });
    } else {
      await rename(staging, destination);
    }
  } catch (error) {
    await rm(staging, { recursive: true, force: true }).catch(() => undefined);
    throw error;
  }
  return manifest;
}

async function exists(value) {
  try {
    await stat(value);
    return true;
  } catch {
    return false;
  }
}

function stripFrontmatter(skillText) {
  const match = skillText.match(/^---\r?\n[\s\S]*?\r?\n---\r?\n([\s\S]*)$/);
  if (!match) throw new Error("Skill source must have YAML frontmatter");
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
