import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";

const load = () => import("./scan.ts");
const marker = () => `PRIVATE_${randomBytes(24).toString("hex")}`;
const goConfig = { protocol_version: "v1.0.0", scanner_name: "govulncheck", scanner_version: "v1.8.0", scan_level: "package", scan_mode: "source", go_version: "go1.26.8" };
const goReport = (findings: unknown[] = [], config = goConfig) => [
  { config },
  { SBOM: { go_version: "go1.26.8", modules: [{ path: "example.test/dependency", version: "v1.0.0" }], roots: ["example.test/fixture"] } },
  ...findings.map((finding) => ({ finding })),
].map((event) => JSON.stringify(event)).join("\n");
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
    const code = await runScans({
      root,
      mode: "dependencies",
      verify: async () => {},
      emit: (s) => lines.push(s),
      runner: async (command) => {
        seen.push(command.tool);
        return command.tool === "govulncheck"
          ? {
              code: 0,
              stdout: goReport([{
                  osv: "GO-2026-0001",
                  fixed_version: privateValue,
                  trace: [{ module: privateValue, package: "example.test/dependency" }],
              }]),
              stderr: privateValue,
            }
          : {
              code: 1,
              stdout: JSON.stringify({
                fixture: [
                  {
                    id: 123,
                    url: "https://github.com/advisories/GHSA-cfgh-2345-jmpq",
                    title: privateValue,
                  },
                ],
              }),
              stderr: privateValue,
            };
      },
    });
    assert.equal(code, 1);
    assert.deepEqual(seen, ["govulncheck", "bun"]);
    assert.match(lines.join("\n"), /GO-2026-0001/);
    assert.match(lines.join("\n"), /GHSA-cfgh-2345-jmpq/);
    assert.doesNotMatch(lines.join("\n"), new RegExp(privateValue));
  });
});

test("malformed, empty, inconsistent, failed service and thrown errors all fail closed", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    for (const response of [
      { code: 0, stdout: "", stderr: "" },
      { code: 0, stdout: "not JSON", stderr: marker() },
      { code: 0, stdout: '{"error":"registry failed"}', stderr: "" },
      { code: 1, stdout: "{}", stderr: marker() },
      { code: 2, stdout: "[]", stderr: marker() },
    ]) {
      const lines: string[] = [];
      assert.equal(
        await runScans({
          root,
          mode: "dependencies",
          verify: async () => {},
          emit: (s) => lines.push(s),
          runner: async () => response,
        }),
        1,
      );
      assert.doesNotMatch(lines.join("\n"), /PRIVATE_/);
    }
    assert.equal(
      await runScans({
        root,
        verify: async () => {
          throw new Error(marker());
        },
        emit: () => {},
        runner: async () => {
          throw new Error(marker());
        },
      }),
      1,
    );
  });
});

test("local no-Git worktree runs, CI missing or shallow history fails and full history runs", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    for (const history of ["missing", "shallow", "full"]) {
      for (const ci of [false, true]) {
        const seen: string[][] = [];
        const lines: string[] = [];
        const code = await runScans({
          root,
          ci,
          mode: "secrets",
          verify: async () => {},
          emit: (s) => lines.push(s),
          runner: async (command) => {
            seen.push(command.args);
            if (command.tool === "git")
              return {
                code: history === "missing" ? 128 : 0,
                stdout: command.args.includes("--is-shallow-repository")
                  ? String(history === "shallow")
                  : "true",
                stderr: marker(),
              };
            return { code: 0, stdout: "[]", stderr: "" };
          },
        });
        assert.equal(code, ci && history !== "full" ? 1 : 0);
        assert.equal(
          seen.some((args) => args[0] === "dir"),
          true,
        );
        assert.equal(
          seen.some((args) => args[0] === "git"),
          history !== "missing",
        );
        if (history === "full")
          assert.ok(seen.some((args) => args.includes("--log-opts=--all --full-history")));
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
    const code = await runScans({
      root,
      mode: "secrets",
      verify: async () => {},
      emit: (s) => lines.push(s),
      runner: async (command) =>
        command.tool === "git"
          ? { code: 128, stdout: "", stderr: privateValue }
          : {
              code: 1,
              stderr: privateValue,
              stdout: JSON.stringify([
                {
                  RuleID: "github-pat",
                  File: "credentials.txt",
                  StartLine: 3,
                  StartColumn: 4,
                  Secret: privateValue,
                  Match: privateValue,
                  Author: privateValue,
                },
                {
                  RuleID: privateValue,
                  File: `https://${privateValue}@host`,
                  StartLine: privateValue,
                },
              ]),
            },
    });
    assert.equal(code, 1);
    assert.match(lines.join("\n"), /github-pat.*credentials.txt:3:4/);
    assert.doesNotMatch(lines.join("\n"), new RegExp(privateValue));
  });
});

test("real bounded capture kills timeout and output flood without publishing stderr", async () => {
  const { capture } = await load();
  const privateValue = marker();
  await assert.rejects(
    capture(
      { tool: process.execPath, args: ["-e", "setTimeout(()=>{},10000)"], cwd: process.cwd() },
      30,
      1024,
    ),
    /scanner execution failed/,
  );
  await assert.rejects(
    capture(
      {
        tool: process.execPath,
        args: ["-e", `process.stderr.write(${JSON.stringify(privateValue)}.repeat(1000))`],
        cwd: process.cwd(),
      },
      3000,
      1024,
    ),
    /scanner execution failed/,
  );
});

test("root scan entrypoints call both wrappers and propagate nonzero scanner status", async () => {
  const manifest = JSON.parse(await readFile(new URL("../package.json", import.meta.url), "utf8"));
  for (const [name, flag] of [
    ["scan", ""],
    ["scan:dependencies", " --dependencies"],
    ["scan:secrets", " --secrets"],
  ])
    assert.equal(manifest.scripts[name ?? ""], `bun scripts/scan.ts${flag}`);
  await fixture(async (root) => {
    const { runScans } = await load();
    const seen: string[] = [];
    assert.equal(
      await runScans({
        root,
        verify: async () => {},
        emit: () => {},
        runner: async (command) => {
          seen.push(command.tool);
          if (command.tool === "git") return { code: 128, stdout: "", stderr: "" };
          return { code: 2, stdout: "", stderr: marker() };
        },
      }),
      1,
    );
    for (const tool of ["bun", "govulncheck", "gitleaks"]) assert.ok(seen.includes(tool));
  });
});

test("real pinned Gitleaks catches runtime worktree and deleted history secrets without disclosure", async () => {
  await fixture(async (root) => {
    const { capture, runScans } = await load();
    const secret = ["ghp", randomBytes(30).toString("hex").slice(0, 40)].join("_");
    const lines: string[] = [];
    const scan = () =>
      runScans({ root, mode: "secrets", verify: async () => {}, emit: (s) => lines.push(s) });
    assert.equal(await scan(), 0);
    await writeFile(join(root, "credentials.txt"), `access_token = ${secret}\n`);
    assert.equal(await scan(), 1);
    assert.match(lines.join("\n"), /github-pat.*credentials.txt:1/);
    assert.doesNotMatch(lines.join("\n"), new RegExp(secret));
    for (const args of [
      ["init", "-q"],
      ["add", "credentials.txt"],
      [
        "-c",
        "user.name=Fixture",
        "-c",
        "user.email=fixture@example.test",
        "commit",
        "-qm",
        "runtime fixture",
      ],
    ]) {
      assert.equal((await capture({ tool: "git", args, cwd: root })).code, 0);
    }
    await writeFile(join(root, "credentials.txt"), "removed\n");
    assert.equal(
      (
        await capture({
          tool: "git",
          args: [
            "-c",
            "user.name=Fixture",
            "-c",
            "user.email=fixture@example.test",
            "commit",
            "-qam",
            "remove fixture",
          ],
          cwd: root,
        })
      ).code,
      0,
    );
    lines.length = 0;
    assert.equal(await scan(), 1);
    assert.match(lines.join("\n"), /Secrets worktree: clean/);
    assert.match(lines.join("\n"), /Secrets history: github-pat/);
    assert.doesNotMatch(lines.join("\n"), new RegExp(secret));
  });
});

test("real CLI rejects invalid invocation safely", async () => {
  const { capture } = await load();
  const result = await capture({
    tool: "bun",
    args: ["scripts/scan.ts", "--invalid"],
    cwd: process.cwd(),
  });
  assert.equal(result.code, 1);
});

test("unused module inventory stays visible while every imported package or symbol fails the root scan", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    for (const exposure of ["unused", "package", "symbol"]) {
      const lines: string[] = [];
      const finding = { osv: "GO-2026-0001", fixed_version: "v1.0.1", trace: [{ module: "example.test/dependency", version: "v1.0.0", ...(exposure === "unused" ? {} : exposure === "package" ? { package: "example.test/dependency/unsafe" } : { function: "Unsafe" }) }] };
      const result = await runScans({ root, mode: "dependencies", verify: async () => {}, emit: (line) => lines.push(line), runner: async (command) => {
        if (command.tool !== "govulncheck") return { code: 0, stdout: "{}", stderr: "" };
        assert.deepEqual(command.args, ["-json", "-scan=package", "-test", "./..."]);
        return { code: 0, stdout: goReport([finding]), stderr: "" };
      } });
      assert.equal(result, exposure === "unused" ? 0 : 1);
      assert.match(lines.join("\n"), /GO-2026-0001/);
      if (exposure === "unused") assert.match(lines.join("\n"), /inventory.*no vulnerable package imported/);
    }
  });
});

test("wrong Go protocol, scope, toolchain and incomplete or malformed findings fail closed", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    const responses = [
      goReport([], { ...goConfig, scan_level: "module" }),
      goReport([], { ...goConfig, scan_mode: "binary" }),
      goReport([], { ...goConfig, scanner_version: "v0.0.0" }),
      goReport([], { ...goConfig, protocol_version: "invalid" }),
      goReport([], { ...goConfig, go_version: "go1.25.5" }),
      JSON.stringify({ SBOM: { go_version: "go1.26.8", modules: [] } }),
      goReport([{ osv: "GO-2026-0001", trace: [] }]),
      goReport([{ osv: "GO-2026-0001", trace: [{ module: "example.test/dependency", package: 42 }] }]),
      goReport([]) + JSON.stringify({ unknown: {} }),
    ];
    for (const stdout of responses) assert.equal(await runScans({ root, mode: "dependencies", verify: async () => {}, emit: () => {}, runner: async (command) => ({ code: 0, stdout: command.tool === "govulncheck" ? stdout : "{}", stderr: "" }) }), 1);
    assert.equal(await runScans({ root, mode: "dependencies", verify: async () => {}, emit: () => {}, runner: async (command) => ({ code: 0, stdout: command.tool === "govulncheck" ? goReport() : "{}", stderr: "" }) }), 0);
  });
});

test("metadata that resembles valid identifiers cannot disclose injected secrets", async () => {
  await fixture(async (root) => {
    const { runScans } = await load();
    const secret = `private-${randomBytes(8).toString("hex")}`;
    const lines: string[] = [];
    await runScans({ root, verify: async () => {}, emit: (line) => lines.push(line), runner: async (command) => {
      if (command.tool === "git") return { code: 128, stdout: "", stderr: secret };
      if (command.tool === "govulncheck") return { code: 0, stdout: goReport([{ osv: secret, fixed_version: `v1.0.0-${secret}`, trace: [{ module: secret, version: `v1.0.0-${secret}`, package: secret }] }]), stderr: secret };
      if (command.tool === "bun") return { code: 1, stdout: JSON.stringify({ [secret]: [{ id: 3, url: `https://github.com/advisories/${secret}` }] }), stderr: secret };
      return { code: 1, stdout: JSON.stringify([{ RuleID: secret, File: `${secret}.txt`, StartLine: 1, StartColumn: 2, Secret: secret }]), stderr: secret };
    } });
    assert.doesNotMatch(lines.join("\n"), new RegExp(secret));
  });
});

test("real pinned package scanner fails a temporary legacy OpenPGP import", async () => {
  await fixture(async (root) => {
    const { capture, runScans } = await load();
    for (const file of ["go.mod", "go.sum"]) await writeFile(join(root, "apps/backend", file), await readFile(new URL(`../apps/backend/${file}`, import.meta.url)));
    await writeFile(join(root, "apps/backend/main.go"), 'package main\nimport _ "golang.org/x/crypto/openpgp"\nfunc main() {}\n');
    const lines: string[] = [];
    const result = await runScans({ root, mode: "dependencies", verify: async () => {}, emit: (line) => lines.push(line), runner: async (command) => command.tool === "govulncheck" ? capture({ ...command, tool: join(process.cwd(), "tmp/tools/govulncheck") }) : { code: 0, stdout: "{}", stderr: "" } });
    assert.equal(result, 1);
    assert.match(lines.join("\n"), /GO-2026-5932.*imported package/);
  });
});
