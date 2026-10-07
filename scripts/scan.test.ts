import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";

const load = () => import("./scan.ts");
const marker = () => `PRIVATE_${randomBytes(24).toString("hex")}`;
const fixture = async (run: (root: string) => Promise<void>) => {
  const root = await mkdtemp(join(tmpdir(), "flux-scan-test-"));
  try {
    await mkdir(join(root, "apps/backend"), { recursive: true });
    await writeFile(join(root, "bun.lock"), "{}");
    await writeFile(join(root, "apps/backend/go.mod"), "module example.test/fixture\n");
    await run(root);
  } finally {
    await rm(root, { recursive: true, force: true });
  }
};

test("both dependency scanners run and findings fail with safe advisory identifiers", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    const privateValue = marker();
    const lines: string[] = [];
    const seen: string[] = [];
    const code = await runScans({ root, mode: "dependencies", verify: async () => {}, emit: (s) => lines.push(s), runner: async (command) => {
      seen.push(command.tool);
      return command.tool === "govulncheck" ? { code: 0, stdout: JSON.stringify({ finding: { osv: "GO-2026-0001", fixed_version: privateValue, trace: [{ module: privateValue }] } }), stderr: privateValue } : { code: 1, stdout: JSON.stringify({ fixture: [{ id: 123, url: "https://github.com/advisories/GHSA-abcd-1234-abcd", title: privateValue }] }), stderr: privateValue };
    } });
    assert.equal(code, 1);
    assert.deepEqual(seen, ["govulncheck", "bun"]);
    assert.match(lines.join("\n"), /GO-2026-0001/);
    assert.match(lines.join("\n"), /GHSA-abcd-1234-abcd/);
    assert.doesNotMatch(lines.join("\n"), new RegExp(privateValue));
  });
});

test("malformed, empty, inconsistent, failed service and thrown errors all fail closed", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    for (const response of [{ code: 0, stdout: "", stderr: "" }, { code: 0, stdout: "not JSON", stderr: marker() }, { code: 0, stdout: '{"error":"registry failed"}', stderr: "" }, { code: 1, stdout: "{}", stderr: marker() }, { code: 2, stdout: "[]", stderr: marker() }]) {
      const lines: string[] = [];
      assert.equal(await runScans({ root, mode: "dependencies", verify: async () => {}, emit: (s) => lines.push(s), runner: async () => response }), 1);
      assert.doesNotMatch(lines.join("\n"), /PRIVATE_/);
    }
    assert.equal(await runScans({ root, verify: async () => { throw new Error(marker()); }, emit: () => {}, runner: async () => { throw new Error(marker()); } }), 1);
  });
});

test("local no-Git worktree runs, CI missing or shallow history fails and full history runs", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    for (const history of ["missing", "shallow", "full"]) {
      for (const ci of [false, true]) {
        const seen: string[][] = [];
        const lines: string[] = [];
        const code = await runScans({ root, ci, mode: "secrets", verify: async () => {}, emit: (s) => lines.push(s), runner: async (command) => {
          seen.push(command.args);
          if (command.tool === "git") return { code: history === "missing" ? 128 : 0, stdout: command.args.includes("--is-shallow-repository") ? String(history === "shallow") : "true", stderr: marker() };
          return { code: 0, stdout: "[]", stderr: "" };
        } });
        assert.equal(code, ci && history !== "full" ? 1 : 0);
        assert.equal(seen.some((args) => args[0] === "dir"), true);
        assert.equal(seen.some((args) => args[0] === "git"), history !== "missing");
        if (history === "full") assert.ok(seen.some((args) => args.includes("--log-opts=--all --full-history")));
        if (history === "missing") assert.match(lines.join("\n"), /history unavailable/);
      }
    }
  });
});

test("secret reports publish rule, path and location but discard raw match, authors and credentials", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    const privateValue = marker();
    const lines: string[] = [];
    const code = await runScans({ root, mode: "secrets", verify: async () => {}, emit: (s) => lines.push(s), runner: async (command) => command.tool === "git" ? { code: 128, stdout: "", stderr: privateValue } : { code: 1, stderr: privateValue, stdout: JSON.stringify([{ RuleID: "github-pat", File: "credentials.txt", StartLine: 3, StartColumn: 4, Secret: privateValue, Match: privateValue, Author: privateValue }, { RuleID: privateValue, File: `https://${privateValue}@host`, StartLine: privateValue }]) } });
    assert.equal(code, 1);
    assert.match(lines.join("\n"), /github-pat.*credentials.txt:3:4/);
    assert.doesNotMatch(lines.join("\n"), new RegExp(privateValue));
  });
});

test("real bounded capture kills timeout and output flood without publishing stderr", async () => {
  const { capture } = await load();
  const privateValue = marker();
  await assert.rejects(capture({ tool: process.execPath, args: ["-e", "setTimeout(()=>{},10000)"], cwd: process.cwd() }, 30, 1024), /scanner execution failed/);
  await assert.rejects(capture({ tool: process.execPath, args: ["-e", `process.stderr.write(${JSON.stringify(privateValue)}.repeat(1000))`], cwd: process.cwd() }, 3000, 1024), /scanner execution failed/);
});

test("root scan entrypoints call both wrappers and propagate nonzero scanner status", async () => {
  const manifest = JSON.parse(await readFile(new URL("../package.json", import.meta.url), "utf8"));
  for (const [name, flag] of [["scan", ""], ["scan:dependencies", " --dependencies"], ["scan:secrets", " --secrets"]]) assert.equal(manifest.scripts[name ?? ""], `bun scripts/scan.ts${flag}`);
  await fixture(async (root) => {
    const { runScans } = await load();
    const seen: string[] = [];
    assert.equal(await runScans({ root, verify: async () => {}, emit: () => {}, runner: async (command) => {
      seen.push(command.tool);
      if (command.tool === "git") return { code: 128, stdout: "", stderr: "" };
      return { code: 2, stdout: "", stderr: marker() };
    } }), 1);
    for (const tool of ["bun", "govulncheck", "gitleaks"]) assert.ok(seen.includes(tool));
  });
});
