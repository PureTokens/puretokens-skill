import { readSkillRegistry } from "./skill-registry.mjs";
for (const skill of (await readSkillRegistry()).skills) console.log(`${skill.name}\t${skill.version}`);
