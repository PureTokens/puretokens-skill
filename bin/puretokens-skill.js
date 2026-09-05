#!/usr/bin/env node
// Compatibility CLI: all installation writes delegate to the current platform installer.
import { execFileSync } from "node:child_process";
import { readFile, rm } from "node:fs/promises";
import path from "node:path";
import { collectSkillRecords, repositoryRoot, validateRepository } from "../scripts/skill-registry.mjs";

const [command, ...args] = process.argv.slice(2);
try {
  if (["sync", "install", "upgrade"].includes(command)) {
    const targetIndex = args.indexOf("--target");
    const hostIndex = args.indexOf("--host");
    const target = targetIndex >= 0 ? args[targetIndex + 1] : undefined;
    const host = hostIndex >= 0 ? args[hostIndex + 1] : undefined;
    if (!target && !host) throw new Error("Supply --host or --target.");
    if (target && !path.isAbsolute(target)) throw new Error("--target must be absolute.");
    const windows = process.platform === "win32";
    const script = path.join(repositoryRoot, "runtime", `puretokens-skill-install.${windows ? "ps1" : "sh"}`);
    const invocation = windows ? ["-NoProfile", "-File", script, "sync"] : [script, "sync"];
    if (target) invocation.push(windows ? "-Target" : "--target", target);
    if (host) invocation.push(windows ? "-Host" : "--host", host);
    execFileSync(windows ? "powershell.exe" : "sh", invocation, { stdio: "inherit" });
  } else if (command === "validate") {
    const errors = await validateRepository();
    if (errors.length) throw new Error(errors.join("\n"));
    console.log("Skill repository validation passed.");
  } else if (command === "list") {
    const { registry } = await collectSkillRecords();
    for (const entry of registry.skills) console.log(`${entry.name}\t${entry.version}`);
  } else if (command === "uninstall") {
    const name = args[0];
    const { registry } = await collectSkillRecords();
    const at = args.indexOf("--target");
    if (!registry.skills.some(s => s.name === name) || at < 0 || !args.includes("--yes") || !path.isAbsolute(args[at + 1] || "")) throw new Error("Use uninstall <official-skill> --target <absolute-directory> --yes.");
    const directory = path.join(args[at + 1], name);
    const manifest = JSON.parse(await readFile(path.join(directory, "skill.json"), "utf8"));
    if (manifest.name !== name) throw new Error("Unmanaged directory left untouched.");
    await rm(directory, { recursive: true });
    console.log(`Removed ${name}. Other Skills and their shared executor remain installed.`);
  } else {
    console.log("Pure Tokens Skill Manager\nUse sync --host <host> or sync --target <absolute-directory>.\nLegacy install/upgrade commands synchronize all six Skills and the platform executor through the same installer.\nRepository development only: list, validate. Users do not need Node.");
  }
} catch (error) {
  if (!error.status) console.error(error.message);
  process.exitCode = 1;
}
