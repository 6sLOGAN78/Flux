import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdir, mkdtemp, readdir, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import * as schemas from "@flux/zod";
import type { ZodType } from "zod";

const root = fileURLToPath(new URL("../../../", import.meta.url));
const generatorPath = join(root, "packages/openapi/dist/gen.js");
const canonicalPath = join(root, "packages/openapi/openapi.json");
const servedPath = join(root, "apps/backend/static/openapi.json");

type Generator = {
  serializeOpenAPI: (document?: unknown) => string;
  generateOpenAPI: (
    outputs?: string[],
    write?: (path: string, bytes: string) => Promise<void>,
  ) => Promise<void>;
};

const loadGenerator = async (): Promise<Generator> => {
  // Assert the public API before importing: the legacy script writes on import.
  const source = await readFile(generatorPath, "utf8");
  assert.match(source, /export (const|function) serializeOpenAPI/);
  assert.match(source, /export (const|async function|function) generateOpenAPI/);
  return (await import(generatorPath)) as unknown as Generator;
};

const withDirectory = async (run: (directory: string) => Promise<void>) => {
  const directory = await mkdtemp(join(tmpdir(), "flux-contract-"));
  try {
    await run(directory);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
};

test("health schemas accept coarse states and reject diagnostic fields", () => {
  const exported = schemas as unknown as Record<string, ZodType>;
  const live = exported.ZHealthLiveResponse;
  const ready = exported.ZHealthReadyResponse;
  assert.ok(live, "liveness schema is exported");
  assert.ok(ready, "readiness schema is exported");
  assert.deepEqual(live.parse({ status: "alive" }), { status: "alive" });
  assert.equal(live.safeParse({ status: "ready" }).success, false);
  for (const state of ["ready", "not_ready"]) {
    const payload = { status: state, checks: [{ name: "database", state }] };
    assert.deepEqual(ready.parse(payload), payload);
  }
  assert.equal(ready.safeParse({ status: "ready", checks: [] }).success, true);
  assert.equal(ready.safeParse({ status: "healthy", checks: [] }).success, false);
  assert.equal(live.safeParse({ status: "alive", environment: "secret" }).success, false);
  assert.equal(
    ready.safeParse({
      status: "ready",
      checks: [{ name: "database", state: "ready", error: "secret" }],
    }).success,
    false,
  );
});

test("canonical OpenAPI documents live 200 and ready 200/503 without legacy diagnostics", async () => {
  const { serializeOpenAPI } = await loadGenerator();
  const document = JSON.parse(serializeOpenAPI());
  assert.equal(document.openapi, "3.0.2");
  assert.equal(document.info.version, "1.0.0");
  assert.deepEqual(Object.keys(document.paths).sort(), ["/live", "/ready"]);
  assert.deepEqual(Object.keys(document.paths["/live"].get.responses), ["200"]);
  assert.deepEqual(Object.keys(document.paths["/ready"].get.responses), ["200", "503"]);
  const responseSchema = (path: string, status: string) => {
    const reference =
      document.paths[path].get.responses[status].content["application/json"].schema.$ref;
    assert.ok(reference, "health schemas have reusable named components");
    return document.components.schemas[reference.split("/").at(-1)];
  };
  assert.equal(responseSchema("/live", "200").title, "transport.HealthLiveResponse");
  assert.deepEqual(responseSchema("/live", "200").properties.status.enum, ["alive"]);
  for (const status of ["200", "503"]) {
    const schema = responseSchema("/ready", status);
    assert.equal(schema.title, "transport.HealthReadyResponse");
    assert.deepEqual(schema.properties.status.enum, ["ready", "not_ready"]);
    assert.deepEqual(Object.keys(schema.properties.checks.items.properties).sort(), [
      "name",
      "state",
    ]);
  }
  assert.doesNotMatch(serializeOpenAPI(), /"(error|environment|timestamp|response_time)"/);
});

test("serialization ignores object insertion order and preserves binary file conversion", async () => {
  const { serializeOpenAPI } = await loadGenerator();
  assert.equal(
    serializeOpenAPI({ z: { b: 2, a: 1 }, a: [2, 1] }),
    serializeOpenAPI({ a: [2, 1], z: { a: 1, b: 2 } }),
  );
  const file = {
    type: "object",
    properties: { type: { type: "string", enum: ["file"] } },
    required: ["type"],
  };
  assert.deepEqual(JSON.parse(serializeOpenAPI(file)), { format: "binary", type: "string" });
  assert.equal(
    serializeOpenAPI(file),
    serializeOpenAPI({
      required: ["type"],
      properties: { type: { enum: ["file"], type: "string" } },
      type: "object",
    }),
  );
});

test("generation awaits writes and produces matching repeatable bytes", async () => {
  const { generateOpenAPI, serializeOpenAPI } = await loadGenerator();
  await withDirectory(async (directory) => {
    const canonical = join(directory, "canonical.json");
    const served = join(directory, "served.json");
    const outputs = [canonical, served];
    await generateOpenAPI(outputs);
    const first = await readFile(canonical, "utf8");
    assert.equal(first, serializeOpenAPI());
    assert.equal(first, await readFile(served, "utf8"));
    await generateOpenAPI(outputs);
    assert.equal(first, await readFile(canonical, "utf8"));
    assert.deepEqual((await readdir(directory)).sort(), ["canonical.json", "served.json"]);
  });
});

test("each failed staging write rejects without publishing either output", async () => {
  const { generateOpenAPI } = await loadGenerator();
  for (const failedIndex of [0, 1]) {
    await withDirectory(async (directory) => {
      const good = join(directory, "good.json");
      await writeFile(good, "original");
      const bad = join(directory, "missing", "bad.json");
      const outputs = failedIndex === 0 ? [bad, good] : [good, bad];
      await assert.rejects(generateOpenAPI(outputs));
      assert.equal(await readFile(good, "utf8"), "original");
      assert.deepEqual(await readdir(directory), ["good.json"]);
    });
  }
});

test("rename failure rejects and temporary files are removed", async () => {
  const { generateOpenAPI } = await loadGenerator();
  await withDirectory(async (directory) => {
    const blocked = join(directory, "blocked.json");
    await mkdir(blocked);
    await assert.rejects(generateOpenAPI([blocked]));
    assert.deepEqual(await readdir(directory), ["blocked.json"]);
  });
});

test("each asynchronous write failure is awaited, preserved, and cleaned up", async () => {
  const { generateOpenAPI } = await loadGenerator();
  for (const failedIndex of [0, 1]) {
    await withDirectory(async (directory) => {
      const outputs = [join(directory, "first.json"), join(directory, "second.json")];
      for (const output of outputs) await writeFile(output, "original");
      const failure = new Error(`write failure ${failedIndex}`);
      let writes = 0;
      let settled = false;
      await assert.rejects(
        generateOpenAPI(outputs, async (path, bytes) => {
          const index = writes++;
          await new Promise((resolve) => setTimeout(resolve, 5));
          if (index === failedIndex) {
            settled = true;
            throw failure;
          }
          await writeFile(path, bytes);
        }),
        (error) => error === failure,
      );
      assert.equal(settled, true);
      for (const output of outputs) assert.equal(await readFile(output, "utf8"), "original");
      assert.deepEqual((await readdir(directory)).sort(), ["first.json", "second.json"]);
    });
  }
});

test("CLI works from another directory and both tracked artifacts match", async () => {
  await withDirectory(async (directory) => {
    const result = spawnSync("node", [generatorPath], { cwd: directory, encoding: "utf8" });
    assert.equal(result.status, 0, result.stderr);
    assert.deepEqual(await readdir(directory), []);
    assert.equal(await readFile(canonicalPath, "utf8"), await readFile(servedPath, "utf8"));
    const { serializeOpenAPI } = await loadGenerator();
    assert.equal(await readFile(canonicalPath, "utf8"), serializeOpenAPI());
  });
});

test("CLI propagates every output write failure to a nonzero process exit", async () => {
  const { serializeOpenAPI } = await loadGenerator();
  for (const failedIndex of [0, 1]) {
    await withDirectory(async (directory) => {
      const outputs = [join(directory, "first.json"), join(directory, "second.json")];
      outputs[failedIndex] = join(directory, "missing", "failed.json");
      const result = spawnSync("node", [generatorPath, ...outputs], {
        cwd: directory,
        encoding: "utf8",
      });
      assert.notEqual(result.status, 0);
      assert.match(result.stderr, /ENOENT/);
      assert.deepEqual(await readdir(directory), []);
    });
  }
  assert.ok(serializeOpenAPI());
});

test("built contract export resolves under Node and Bun", () => {
  for (const runtime of ["node", "bun"]) {
    const result = spawnSync(
      runtime,
      [
        "--eval",
        "import('@flux/openapi/contracts').then(({ apiContract }) => { if (apiContract.Health.getLive.path !== '/live' || apiContract.Health.getReady.path !== '/ready') throw new Error('contract mismatch'); }).catch(error => { console.error(error); process.exitCode = 1; });",
      ],
      { cwd: join(root, "packages/openapi"), encoding: "utf8" },
    );
    assert.equal(result.status, 0, `${runtime}: ${result.stderr}`);
  }
});
