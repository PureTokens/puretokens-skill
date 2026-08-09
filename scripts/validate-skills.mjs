import { validateRepository } from "./skill-registry.mjs";

const errors = await validateRepository();
if (errors.length) {
  for (const error of errors) process.stderr.write(`error: ${error}\n`);
  process.exitCode = 1;
} else {
  process.stdout.write("Skill repository validation passed.\n");
}
