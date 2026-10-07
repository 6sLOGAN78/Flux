import assert from "node:assert/strict";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { test } from "node:test";

const load = () => import("./check.ts");
const commands = ["format:check", "lint", "typecheck", "test", "build"];

const fixture = async (run: (root: string) => Promise<void>) => {
  const root = await mkdtemp(join(tmpdir(), "flux-check-test-"));
  const put = async (path: string, content: string) => {
    await mkdir(dirname(join(root, path)), { recursive: true });
    await writeFile(join(root, path), content);
  };
  try {
    await put("package.json", JSON.stringify({ workspaces: ["apps/*", "packages/*"] }));
    await put("apps/backend/go.mod", "module fixture\n");
    for (const name of ["zod", "openapi", "emails"]) {
      await put(`packages/${name}/package.json`, JSON.stringify({ name: `@flux/${name}`, scripts: Object.fromEntries(commands.map((command) => [command, "true"])) }));
      await put(`packages/${name}/src/behavior.test.ts`, 'import { test } from "node:test"; test("behavior", () => {});');
    }
    const { integrationTests } = await load();
    for (const [pkg, names] of Object.entries(integrationTests)) {
      await put(`apps/backend/${pkg}/behavior_test.go`, `package fixture\n${[...names, "TestUnitBehavior"].map((name) => `func ${name}(t *testing.T) {}`).join("\n")}`);
    }
    await run(root);
  } finally {
    await rm(root, { recursive: true, force: true });
  }
};

test("full manifest dispatches every Go/workspace stage and uncached generation plus scans", async () => {
  await fixture(async (root) => {
    const { runChecks } = await load();
    const seen: string[] = [];
    await runChecks({ root, mode: "full", runner: async (stage) => { seen.push(stage.id); return { code: 0, stdout: stage.expectedTests?.map((name) => JSON.stringify({ Action: "pass", Test: name })).join("\n") ?? (stage.id.includes("test") ? "1 pass\n0 fail\n" : "") }; } });
    for (const id of ["tools:verify", "backend:format", "backend:vet", "backend:staticcheck", "backend:golangci", "backend:typecheck", "backend:build", "migrate:check", "generate:check", "scan"]) assert.ok(seen.includes(id), `Missing ${id}`);
    for (const workspace of ["zod", "openapi", "emails"]) {
      for (const command of commands) assert.ok(seen.includes(`@flux/${workspace}:${command}`));
    }
    assert.ok(seen.some((id) => id.startsWith("backend:unit:")));
    assert.ok(seen.some((id) => id.startsWith("backend:integration:")));
  });
});

test("fast omits only integration migrations and scans", async () => {
  await fixture(async (root) => {
    const { discover, createStages } = await load();
    const context = await discover(root);
    const full = createStages(context, "full");
    const fast = createStages(context, "fast");
    assert.deepEqual(full.filter((stage) => !["test:integration", "migrate:check", "scan"].includes(stage.group)).map((stage) => stage.id), fast.map((stage) => stage.id));
    assert.ok(fast.filter((stage) => stage.expectedTests).every((stage) => stage.command.includes("-race")));
  });
});

test("missing scripts, omitted backend and zero workspace tests fail discovery", async () => {
  for (const failure of ["script", "backend", "tests"]) {
    await fixture(async (root) => {
      const { discover } = await load();
      if (failure === "backend") await rm(join(root, "apps/backend/go.mod"));
      if (failure === "tests") await rm(join(root, "packages/zod/src/behavior.test.ts"));
      if (failure === "script") {
        const file = join(root, "packages/zod/package.json");
        const manifest = JSON.parse(await readFile(file, "utf8"));
        delete manifest.scripts.lint;
        await writeFile(file, JSON.stringify(manifest));
      }
      await assert.rejects(discover(root));
    });
  }
});

test("all Go tests belong to exactly one group and absent or newly renamed integration tests fail", async () => {
  await fixture(async (root) => {
    const { discover } = await load();
    const context = await discover(root);
    for (const pkg of context.goTests) {
      assert.equal(new Set([...pkg.unit, ...pkg.integration]).size, pkg.unit.length + pkg.integration.length);
      assert.ok(pkg.unit.includes("TestUnitBehavior"));
      assert.ok(pkg.integration.length > 0);
    }
    await writeFile(join(root, "apps/backend/internal/app/behavior_test.go"), "package fixture\nfunc TestRenamed(t *testing.T) {}\n");
    await assert.rejects(discover(root), /Missing classified integration test/);
  });
});

test("nonzero lint, generated drift and scanner failure stop dispatch", async () => {
  for (const failed of ["backend:golangci", "generate:check", "scan"]) {
    await fixture(async (root) => {
      const { runChecks } = await load();
      const seen: string[] = [];
      await assert.rejects(runChecks({ root, mode: "full", runner: async (stage) => { seen.push(stage.id); return { code: stage.id === failed ? 1 : 0, stdout: stage.expectedTests?.map((name) => JSON.stringify({ Action: "pass", Test: name })).join("\n") ?? "1 pass\n0 fail\n" }; } }), new RegExp(`Check failed: ${failed}`));
      assert.equal(seen.at(-1), failed);
    });
  }
});

test("zero executed or skipped Go tests and zero executed Bun tests fail even with exit zero", async () => {
  for (const result of ["", JSON.stringify({ Action: "skip", Test: "TestUnitBehavior" }), "0 pass\n0 fail\n"]) {
    await fixture(async (root) => {
      const { runChecks } = await load();
      await assert.rejects(runChecks({ root, mode: "fast", runner: async (stage) => ({ code: 0, stdout: stage.group === "test:unit" ? result : "" }) }), /No successful tests|Missing successful Go test/);
    });
  }
});

test("subprocess deadlines and bounded private diagnostics fail safely", async () => {
  const { runCommand } = await load();
  const result = await runCommand({ id: "timeout", group: "lint", cwd: process.cwd(), command: ["bun", "-e", 'console.error("SECRET-MARKER");setTimeout(()=>{},10000)'], timeoutMs: 40 });
  assert.notEqual(result.code, 0);
  assert.doesNotMatch(JSON.stringify(result), /SECRET-MARKER/);
});
