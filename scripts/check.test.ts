import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { cp, mkdir, mkdtemp, readFile, rm, symlink, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join, relative } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import type { Stage } from "./check.ts";

const load = () => import("./check.ts");
const commands = ["format:check", "lint", "typecheck", "test", "build"];

test("final secret scan removes generated Next keys while preserving source and refusing unrelated output", async () => {
  const root = await mkdtemp(join(tmpdir(), "flux-next-output-"));
  const { cleanFrontendBuild } = await load();
  const frontend = join(root, "apps/frontend");
  const output = join(frontend, "build");
  try {
    await mkdir(output, { recursive: true });
    await writeFile(join(frontend, "source.ts"), "authored source must remain");
    await writeFile(join(output, "keep.txt"), "unrecognized output");
    await assert.rejects(cleanFrontendBuild(root));
    assert.equal(await readFile(join(output, "keep.txt"), "utf8"), "unrecognized output");
    await writeFile(
      join(output, "prerender-manifest.json"),
      JSON.stringify({
        version: 4,
        routes: {},
        preview: {
          previewModeId: "0".repeat(32),
          previewModeSigningKey: "0".repeat(64),
          previewModeEncryptionKey: "0".repeat(64),
        },
      }),
    );
    await cleanFrontendBuild(root);
    await assert.rejects(readFile(join(output, "prerender-manifest.json")));
    assert.equal(
      await readFile(join(frontend, "source.ts"), "utf8"),
      "authored source must remain",
    );
    await cleanFrontendBuild(root);
    await symlink(frontend, output);
    await assert.rejects(cleanFrontendBuild(root), /Invalid frontend build output/);
    assert.equal(
      await readFile(join(frontend, "source.ts"), "utf8"),
      "authored source must remain",
    );
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});

const browserReport = () => ({
  errors: [],
  stats: { expected: 1, unexpected: 0, skipped: 0, flaky: 0 },
  suites: [
    {
      suites: [],
      specs: [
        {
          ok: true,
          tests: [
            {
              expectedStatus: "passed",
              status: "expected",
              results: [{ status: "passed", retry: 0, errors: [] }],
            },
          ],
        },
      ],
    },
  ],
});

test("browser output is private OS temporary storage rather than a scanned repository directory", async () => {
  const config = await readFile(
    new URL("../apps/frontend/playwright.config.ts", import.meta.url),
    "utf8",
  );
  assert.doesNotMatch(config, /outputDir: "\.\.\/\.\.\/tmp\//);
  assert.match(config, /mkdtempSync/);
  assert.match(config, /tmpdir\(\)/);
});

test("root browser command failure, zero, skip, malformed and inconsistent reports fail closed", async () => {
  for (const failure of ["command", "zero", "skip", "json", "count"]) {
    await fixture(async (root) => {
      const frontend = join(root, "apps/frontend");
      await mkdir(join(frontend, "tests"), { recursive: true });
      await writeFile(
        join(frontend, "package.json"),
        JSON.stringify({
          name: "@flux/frontend",
          scripts: Object.fromEntries(commands.map((command) => [command, "true"])),
        }),
      );
      await writeFile(join(frontend, "tests/auth.spec.ts"), "export {};");
      const { runChecks } = await load();
      await assert.rejects(
        runChecks({
          root,
          group: "test:e2e",
          runner: async (stage) => {
            if (!stage.browserTests) return { code: 0, stdout: successfulOutput(stage) };
            const report = browserReport();
            if (failure === "zero") report.stats.expected = 0;
            if (failure === "skip") report.stats.skipped = 1;
            if (failure === "count") report.stats.expected = 2;
            return {
              code: failure === "command" ? 1 : 0,
              stdout: failure === "json" ? "PRIVATE-REPORT" : JSON.stringify(report),
            };
          },
        }),
        /Check failed: @flux\/frontend:test:e2e|No successful browser tests/,
      );
    });
  }
});

test("browser gate requires positive complete non-skipped nonflaky JSON results", async () => {
  const { assertBrowserReport } = await load();
  assert.equal(assertBrowserReport(JSON.stringify(browserReport())), 1);
  const leafReport = browserReport();
  Reflect.deleteProperty(leafReport.suites[0] ?? {}, "suites");
  assert.equal(assertBrowserReport(JSON.stringify(leafReport)), 1);
  for (const output of [
    "",
    "{}",
    "not-json",
    JSON.stringify({ ...browserReport(), suites: [] }),
    ...["expected", "unexpected", "skipped", "flaky"].map((key) =>
      JSON.stringify({
        ...browserReport(),
        stats: { ...browserReport().stats, [key]: key === "expected" ? 0 : 1 },
      }),
    ),
    JSON.stringify({ ...browserReport(), errors: [{}] }),
    JSON.stringify(browserReport()).replace('"status":"passed"', '"status":"skipped"'),
    JSON.stringify(browserReport()).replace('"retry":0', '"retry":1'),
  ])
    assert.throws(() => assertBrowserReport(output), /No successful browser tests/);
});

test("browser stage forwards selectors without overriding the JSON reporter", async () => {
  await fixture(async (root) => {
    const frontend = join(root, "apps/frontend");
    await mkdir(join(frontend, "tests"), { recursive: true });
    await writeFile(
      join(frontend, "package.json"),
      JSON.stringify({
        name: "@flux/frontend",
        scripts: Object.fromEntries(
          [...commands, "test:e2e", "browser:install"].map((command) => [command, "true"]),
        ),
      }),
    );
    await writeFile(join(frontend, "tests/auth.spec.ts"), "export {};");
    const { runChecks } = await load();
    const seen: Stage[] = [];
    await runChecks({
      root,
      group: "test:e2e",
      browserArgs: ["--project=local", "--grep", "sign-in"],
      runner: async (stage) => {
        seen.push(stage);
        return {
          code: 0,
          stdout: stage.browserTests ? JSON.stringify(browserReport()) : successfulOutput(stage),
        };
      },
    });
    assert.deepEqual(seen.at(-1)?.command, [
      "bun",
      "run",
      "test:e2e",
      "--project=local",
      "--grep",
      "sign-in",
    ]);
  });
});

test("root subprocesses find the pinned Task when the host PATH has none", async () => {
  const { runCommand } = await load();
  const previous = process.env.PATH;
  try {
    process.env.PATH = "/nonexistent-task-host-path";
    const result = await runCommand({
      id: "task:version",
      group: "tools",
      command: ["task", "--version"],
      cwd: fileURLToPath(new URL("../", import.meta.url)),
      timeoutMs: 1000,
    });
    assert.equal(result.code, 0);
    const manifest = JSON.parse(
      await readFile(new URL("../tools.lock.json", import.meta.url), "utf8"),
    );
    assert.equal(result.stdout.trim(), manifest.qualityTools.task.versionOutput);
  } finally {
    process.env.PATH = previous;
  }
});

test("real full and fast root checks reject initial drift before any mutation", {
  timeout: 180000,
}, async () => {
  const repository = fileURLToPath(new URL("../", import.meta.url));
  const root = await mkdtemp(join(tmpdir(), "flux-root-drift-"));
  const { artifactManifest } = await import("./generate.ts");
  try {
    for (const path of ["package.json", "tools.lock.json", "scripts", "packages", "apps"]) {
      await cp(join(repository, path), join(root, path), {
        recursive: true,
        filter: (path) =>
          !/(?:^|\/)(node_modules|dist|build|\.next|tmp|\.env)(?:\/|$)/.test(
            relative(repository, path),
          ),
      });
    }
    await symlink(join(repository, "node_modules"), join(root, "node_modules"));
    for (const name of ["zod", "openapi", "emails"])
      await symlink(
        join(repository, "packages", name, "node_modules"),
        join(root, "packages", name, "node_modules"),
      );
    const original = await Promise.all(artifactManifest.map((path) => readFile(join(root, path))));
    const authored = await readFile(join(root, "packages/zod/src/health.ts"));
    for (const mode of [["--full"], ["--fast"]]) {
      for (const [index, path] of artifactManifest.entries()) {
        for (const mutation of ["changed", "missing"]) {
          if (mutation === "missing") await rm(join(root, path));
          else await writeFile(join(root, path), "private-drift-sentinel");
          const result = spawnSync("bun", [join(root, "scripts/check.ts"), ...mode], {
            cwd: root,
            encoding: "utf8",
            timeout: 60000,
          });
          assert.equal(result.status, 1, `${mode.join(" ")} ${path} ${mutation}`);
          assert.match(result.stderr, /Check failed: generate:check/);
          assert.doesNotMatch(result.stdout + result.stderr, /private-drift-sentinel/);
          assert.doesNotMatch(result.stdout, /tools:verify|:build|:test/);
          if (mutation === "missing") await assert.rejects(readFile(join(root, path)));
          else assert.equal(await readFile(join(root, path), "utf8"), "private-drift-sentinel");
          for (const [otherIndex, other] of artifactManifest.entries())
            if (other !== path)
              assert.deepEqual(await readFile(join(root, other)), original[otherIndex]);
          assert.deepEqual(await readFile(join(root, "packages/zod/src/health.ts")), authored);
          await writeFile(join(root, path), original[index] as Buffer);
        }
      }
    }
    const identityPath = join(root, "packages/zod/src/identity.ts");
    const identity = await readFile(identityPath, "utf8");
    await writeFile(identityPath, identity.replace(".max(320)", ".max(319)"));
    const result = spawnSync("bun", [join(root, "scripts/generate.ts"), "--check"], {
      cwd: root,
      encoding: "utf8",
      timeout: 60000,
    });
    assert.equal(result.status, 1);
    assert.match(result.stderr, /Stale generated artifacts:/);
    assert.match(result.stderr, /packages\/openapi\/openapi.json/);
    for (const [index, path] of artifactManifest.entries())
      assert.deepEqual(await readFile(join(root, path)), original[index]);
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});
const successfulOutput = (stage: Stage) => {
  if (stage.listedTests)
    return Object.entries(stage.listedTests)
      .flatMap(([Package, names]) =>
        names.map((name) => JSON.stringify({ Package, Output: `${name}\n` })),
      )
      .join("\n");
  if (stage.expectedTests)
    return stage.expectedTests
      .map((name) => JSON.stringify({ Action: "pass", Test: name }))
      .join("\n");
  return stage.bunTests ? "1 pass\n0 fail\n" : "";
};

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
      await put(
        `packages/${name}/package.json`,
        JSON.stringify({
          name: `@flux/${name}`,
          scripts: Object.fromEntries(commands.map((command) => [command, "true"])),
        }),
      );
      await put(
        `packages/${name}/src/behavior.test.ts`,
        'import { test } from "node:test"; test("behavior", () => {});',
      );
    }
    const { integrationTests } = await load();
    for (const [pkg, names] of Object.entries(integrationTests)) {
      await put(
        `apps/backend/${pkg}/behavior_test.go`,
        `package fixture\n${[...names, "TestUnitBehavior"].map((name) => `func ${name}(t *testing.T) {}`).join("\n")}`,
      );
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
    await runChecks({
      root,
      mode: "full",
      runner: async (stage) => {
        seen.push(stage.id);
        return {
          code: 0,
          stdout: successfulOutput(stage),
        };
      },
    });
    for (const id of [
      "tools:verify",
      "backend:format",
      "backend:vet",
      "backend:staticcheck",
      "backend:golangci",
      "backend:typecheck",
      "backend:build",
      "migrate:check",
      "generate:check",
      "scan",
    ])
      assert.ok(seen.includes(id), `Missing ${id}`);
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
    assert.deepEqual(
      full
        .filter((stage) => !["test:integration", "migrate:check", "scan"].includes(stage.group))
        .map((stage) => stage.id),
      fast.map((stage) => stage.id),
    );
    assert.ok(
      fast.filter((stage) => stage.expectedTests).every((stage) => stage.command.includes("-race")),
    );
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
      assert.equal(
        new Set([...pkg.unit, ...pkg.integration]).size,
        pkg.unit.length + pkg.integration.length,
      );
      assert.ok(pkg.unit.includes("TestUnitBehavior"));
      assert.ok(pkg.integration.length > 0);
    }
    await writeFile(
      join(root, "apps/backend/internal/app/behavior_test.go"),
      "package fixture\nfunc TestRenamed(t *testing.T) {}\n",
    );
    await assert.rejects(discover(root), /Missing classified integration test/);
  });
});

test("nonzero lint, generated drift and scanner failure stop dispatch", async () => {
  for (const failed of ["backend:golangci", "generate:check", "scan"]) {
    await fixture(async (root) => {
      const { runChecks } = await load();
      const seen: string[] = [];
      await assert.rejects(
        runChecks({
          root,
          mode: "full",
          runner: async (stage) => {
            seen.push(stage.id);
            return {
              code: stage.id === failed ? 1 : 0,
              stdout: successfulOutput(stage),
            };
          },
        }),
        new RegExp(`Check failed: ${failed}`),
      );
      assert.equal(seen.at(-1), failed);
    });
  }
});

test("zero executed or skipped Go tests and zero executed Bun tests fail even with exit zero", async () => {
  for (const failure of ["go-zero", "go-skip", "bun-zero", "list-zero"]) {
    await fixture(async (root) => {
      const { runChecks } = await load();
      await assert.rejects(
        runChecks({
          root,
          mode: "fast",
          runner: async (stage) => {
            let stdout = successfulOutput(stage);
            if (stage.expectedTests && failure === "go-zero") stdout = "";
            if (stage.expectedTests && failure === "go-skip")
              stdout = JSON.stringify({ Action: "skip", Test: stage.expectedTests[0] });
            if (stage.bunTests && failure === "bun-zero") stdout = "0 pass\n0 fail\n";
            if (stage.listedTests && failure === "list-zero") stdout = "";
            return { code: 0, stdout };
          },
        }),
        /No successful tests|Missing successful Go test|Go test manifest mismatch/,
      );
    });
  }
});

test("compiled listing rejects source-signature omissions and build-constrained selections", async () => {
  for (const mismatch of ["additional", "absent"]) {
    await fixture(async (root) => {
      const { runChecks } = await load();
      await assert.rejects(
        runChecks({
          root,
          mode: "fast",
          runner: async (stage) => {
            let stdout = successfulOutput(stage);
            if (stage.listedTests) {
              if (mismatch === "additional")
                stdout +=
                  '\n{"Package":"fixture/internal/app","Output":"TestAlternateSignature\\n"}';
              else stdout = "";
            }
            return { code: 0, stdout };
          },
        }),
        /Go test manifest mismatch/,
      );
    });
  }
});

test("subprocess deadlines and bounded private diagnostics fail safely", async () => {
  const { runCommand } = await load();
  const result = await runCommand({
    id: "timeout",
    group: "lint",
    cwd: process.cwd(),
    command: ["bun", "-e", 'console.error("SECRET-MARKER");setTimeout(()=>{},10000)'],
    timeoutMs: 40,
  });
  assert.notEqual(result.code, 0);
  assert.doesNotMatch(JSON.stringify(result), /SECRET-MARKER/);
});

test("failed Go subprocesses expose only expected test identifiers", async () => {
  const { runCommand, runChecks } = await load();
  await fixture(async (root) => {
    await assert.rejects(
      runChecks({
        root,
        runner: async (stage) => {
          if (stage.id !== "backend:integration:internal/app")
            return { code: 0, stdout: successfulOutput(stage) };
          const events = [
            { Action: "output", Output: "PRIVATE-DIAGNOSTIC-MARKER" },
            { Action: "fail", Test: "TestRoleBinaryStartup/private-subtest" },
            { Action: "fail", Test: "TestRoleBinaryStartup" },
            { Action: "fail", Test: "TestUnknownPRIVATE-DIAGNOSTIC-MARKER" },
            { Action: "fail" },
          ];
          const result = await runCommand({
            ...stage,
            command: [
              "bun",
              "-e",
              `console.log(${JSON.stringify(events.map((event) => JSON.stringify(event)).join("\n"))});console.error("PRIVATE-DIAGNOSTIC-MARKER");process.exit(1)`,
            ],
            timeoutMs: 5000,
          });
          assert.equal(result.code, 1);
          assert.equal(result.stdout, "");
          assert.deepEqual(result.failedTests, ["TestRoleBinaryStartup"]);
          assert.doesNotMatch(JSON.stringify(result), /PRIVATE-DIAGNOSTIC-MARKER/);
          return result;
        },
      }),
      { message: "Check failed: backend:integration:internal/app (TestRoleBinaryStartup)" },
    );
  });
});
