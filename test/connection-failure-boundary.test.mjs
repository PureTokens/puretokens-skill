import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

test("every installed Skill stops the screenshot's local-failure to cross-host probe chain", async () => {
  for (const kind of ["image", "video", "connection", "balance", "models", "update"]) {
    const root = new URL(`../skills/puretokens-${kind}/`, import.meta.url);
    const skill = await readFile(new URL("SKILL.md", root), "utf8");
    const scenarios = JSON.parse(await readFile(new URL("references/behavior-scenarios.json", root), "utf8"));
    // The guard must be visible before command instructions, not hidden in a
    // reference that an agent would read only after starting recovery probes.
    const guard = skill.indexOf("宿主绑定与本地失败停止");
    const firstSection = skill.indexOf("## ");
    assert.ok(guard >= 0 && guard < 500 && (firstSection < 0 || guard < firstSection), kind);
    for (const text of ["不得自动调用 puretokens-connection、init、doctor、models", "不得枚举、更换 `--host`", "保留用户原任务和已有 task ID", "“继续生成”本身不是"]) {
      assert.ok(skill.includes(text), `${kind}: ${text}`);
    }
    const scenario = scenarios.scenarios.find(s => s.id === `${kind}-local-connection-failure-stops-host-probing`);
    assert.ok(scenario, kind);
    assert.match(scenario.when, /continue generation/);
    assert.match(scenario.then, /Do not invoke connection, init, doctor, models, another Skill, or another --host/);
    assert.match(scenario.then, /changing host requires explicit user selection/);
  }
});
