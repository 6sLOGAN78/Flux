import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { test } from "node:test";
import { ZHealthLiveResponse, ZHealthReadyResponse } from "./health.js";

test("implemented schema workspace exposes every quality command", async () => {
  const manifest = JSON.parse(await readFile(new URL("../package.json", import.meta.url), "utf8"));
  for (const command of ["format:check", "lint", "typecheck", "test", "build"]) {
    assert.equal(typeof manifest.scripts[command], "string", `Missing ${command}`);
    assert.ok(manifest.scripts[command].trim());
  }
});

test("health accepts documented liveness and both readiness states", () => {
  assert.deepEqual(ZHealthLiveResponse.parse({ status: "alive" }), { status: "alive" });
  for (const state of ["ready", "not_ready"]) {
    const response = { status: state, checks: [{ name: "database", state }] };
    assert.deepEqual(ZHealthReadyResponse.parse(response), response);
  }
  assert.deepEqual(ZHealthReadyResponse.parse({ status: "ready", checks: [] }), {
    status: "ready",
    checks: [],
  });
});

test("health rejects unknown states, missing fields and private diagnostics", () => {
  for (const response of [{}, { status: "ready" }, { status: "alive", error: "private" }]) {
    assert.equal(ZHealthLiveResponse.safeParse(response).success, false);
  }
  for (const response of [
    { status: "unknown", checks: [] },
    { status: "ready" },
    { status: "ready", checks: [{ name: "database", state: "healthy" }] },
    { status: "ready", checks: [{ state: "ready" }] },
    { status: "ready", checks: [{ name: "database", state: "ready", error: "secret" }] },
    { status: "ready", checks: [], credentials: "secret" },
  ]) {
    assert.equal(ZHealthReadyResponse.safeParse(response).success, false);
  }
});
