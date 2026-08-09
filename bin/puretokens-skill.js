#!/usr/bin/env node

import { cp, mkdir, stat } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import process from "node:process";
import { collectSkillRecords, repositoryRoot, skillsRoot, validateRepository } from "../scripts/skill-registry.mjs";

const [command, ...argumentsList] = process.argv.slice(2);

try {
  switch (command) {
    case "list":
      await listSkills();
      break;
    case "install":
      await installSkill(argumentsList);
      break;
    case "validate":
      await validate();
      break;
    case "help":
    case "--help":
    case "-h":
    case undefined:
      printHelp();
      break;
    default:
      throw new Error(`Unknown command: ${command}`);
  }
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = 1;
}

async function listSkills() {
  const { registry, records } = await collectSkillRecords();
  const byName = new Map(records.map((record) => [record.manifest.name, record]));
  for (const entry of registry.skills) {
    const record = byName.get(entry.name);
    if (!record) continue;
    process.stdout.write(`${entry.name}\t${entry.version}\t${entry.summary || record.manifest.description || ""}\n`);
  }
}

async function installSkill(args) {
  const name = args[0];
  if (!name || name.startsWith("-")) throw new Error("Usage: puretokens-skill install <skill-name> [--target <directory>]");
  const targetIndex = args.indexOf("--target");
  if (targetIndex >= 0 && (!args[targetIndex + 1] || args[targetIndex + 2])) {
    throw new Error("--target requires exactly one directory");
  }
  const unknownFlags = args.filter((item, index) => index > 0 && item !== "--target" && index !== targetIndex + 1);
  if (unknownFlags.length) throw new Error(`Unknown install option: ${unknownFlags.join(" ")}`);

  const errors = await validateRepository();
  if (errors.length) throw new Error(`Repository validation failed:\n${errors.map((error) => `- ${error}`).join("\n")}`);
  const source = path.join(skillsRoot, name);
  if (!source.startsWith(`${skillsRoot}${path.sep}`) || !(await existsDirectory(source))) {
    throw new Error(`Unknown skill: ${name}`);
  }
  const root = targetIndex >= 0 ? path.resolve(args[targetIndex + 1]) : path.join(os.homedir(), ".codex", "skills");
  const destination = path.join(root, name);
  if (!destination.startsWith(`${root}${path.sep}`)) throw new Error("Invalid target path");
  if (await exists(destination)) {
    throw new Error(`Refusing to overwrite existing skill: ${destination}`);
  }
  await mkdir(root, { recursive: true });
  await cp(source, destination, { recursive: true, errorOnExist: true, force: false });
  process.stdout.write(`Installed ${name} to ${destination}\n`);
}

async function validate() {
  const errors = await validateRepository();
  if (errors.length) throw new Error(errors.map((error) => `- ${error}`).join("\n"));
  process.stdout.write("Skill repository validation passed.\n");
}

async function exists(value) {
  try {
    await stat(value);
    return true;
  } catch {
    return false;
  }
}

async function existsDirectory(value) {
  try {
    return (await stat(value)).isDirectory();
  } catch {
    return false;
  }
}

function printHelp() {
  process.stdout.write(`Pure Tokens Skill Manager\n\nUsage:\n  puretokens-skill list\n  puretokens-skill validate\n  puretokens-skill install <skill-name> [--target <directory>]\n\nRepository: ${repositoryRoot}\n`);
}
