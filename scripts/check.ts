import { spawn } from "node:child_process";
import { randomBytes, randomUUID } from "node:crypto";
import type { Stats } from "node:fs";
import { lstat, mkdtemp, readdir, readFile, realpath, rm, writeFile } from "node:fs/promises";
import { createServer } from "node:http";
import { connect, createServer as createTCPServer, isIP, type Socket } from "node:net";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const repositoryRoot = fileURLToPath(new URL("../", import.meta.url));

// Next generates deployment-local signing/encryption keys in its build tree.
// Remove the disposable output after verification, before scanning the complete
// worktree. Authored source and full Git history keep every secret rule enabled.
export const cleanFrontendBuild = async (root: string): Promise<void> => {
  const output = join(root, "apps/frontend/build");
  let metadata: Stats;
  try {
    metadata = await lstat(output);
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") return;
    throw new Error("Frontend build cleanup failed");
  }
  if (!metadata.isDirectory() || metadata.isSymbolicLink())
    throw new Error("Invalid frontend build output");
  if ((await realpath(output)) !== join(await realpath(root), "apps/frontend/build"))
    throw new Error("Invalid frontend build output");
  let manifest: { version?: number; routes?: unknown; preview?: Record<string, unknown> };
  try {
    manifest = JSON.parse(await readFile(join(output, "prerender-manifest.json"), "utf8"));
  } catch (cause) {
    throw new Error("Invalid frontend build output", { cause });
  }
  const preview = manifest.preview ?? {};
  if (
    manifest?.version !== 4 ||
    !manifest.routes ||
    typeof preview.previewModeId !== "string" ||
    !/^[a-f0-9]{32}$/.test(preview.previewModeId) ||
    typeof preview.previewModeSigningKey !== "string" ||
    !/^[a-f0-9]{64}$/.test(preview.previewModeSigningKey) ||
    typeof preview.previewModeEncryptionKey !== "string" ||
    !/^[a-f0-9]{64}$/.test(preview.previewModeEncryptionKey)
  )
    throw new Error("Invalid frontend build output");
  await rm(output, { recursive: true });
};
// Validate the complete JSON tree as well as aggregate stats. Retries, skips,
// empty selection and infrastructure errors must never produce green evidence.
export const assertBrowserReport = (output: string, expected?: string[]): number => {
  const fail = (): never => {
    throw new Error("No successful browser tests");
  };
  let report: Record<string, unknown>;
  try {
    report = JSON.parse(output);
  } catch {
    return fail();
  }
  if (!report || typeof report !== "object") return fail();
  const stats = report.stats as Record<string, unknown> | undefined;
  if (
    !stats ||
    !Number.isSafeInteger(stats.expected) ||
    Number(stats.expected) < 1 ||
    stats.unexpected !== 0 ||
    stats.skipped !== 0 ||
    stats.flaky !== 0 ||
    !Array.isArray(report.errors) ||
    report.errors.length !== 0
  )
    return fail();
  let completed = 0;
  const visit = (suites: unknown): void => {
    if (!Array.isArray(suites)) fail();
    for (const suite of suites as Record<string, unknown>[]) {
      if (
        !suite ||
        (suite.specs !== undefined && !Array.isArray(suite.specs)) ||
        (suite.suites !== undefined && !Array.isArray(suite.suites))
      )
        fail();
      for (const spec of (suite.specs ?? []) as Record<string, unknown>[]) {
        if (spec.ok !== true || !Array.isArray(spec.tests) || spec.tests.length === 0) fail();
        for (const test of spec.tests as Record<string, unknown>[]) {
          if (
            test.expectedStatus !== "passed" ||
            test.status !== "expected" ||
            !Array.isArray(test.results) ||
            test.results.length !== 1
          )
            fail();
          const result = (test.results as Record<string, unknown>[])[0];
          if (
            result?.status !== "passed" ||
            result.retry !== 0 ||
            !Array.isArray(result.errors) ||
            result.errors.length !== 0
          )
            fail();
          completed++;
        }
      }
      visit(suite.suites ?? []);
    }
  };
  visit(report.suites);
  if (completed !== stats.expected) return fail();
  if (expected && JSON.stringify(browserTestIds(output)) !== JSON.stringify(expected))
    throw new Error("Browser test manifest mismatch");
  return completed;
};

// Native --list and execution reports share spec IDs and project IDs. A report
// with consistent totals is insufficient if it omits or substitutes a test.
export const browserTestIds = (output: string, listing = false): string[] => {
  const fail = (): never => {
    throw new Error("Invalid browser test manifest");
  };
  let report: Record<string, unknown>;
  try {
    report = JSON.parse(output);
  } catch {
    return fail();
  }
  if (
    !report ||
    typeof report !== "object" ||
    !Array.isArray(report.errors) ||
    report.errors.length
  )
    return fail();
  const ids: string[] = [];
  const visit = (suites: unknown): void => {
    if (!Array.isArray(suites)) fail();
    for (const suite of suites as Record<string, unknown>[]) {
      if (
        !suite ||
        typeof suite !== "object" ||
        (suite.specs !== undefined && !Array.isArray(suite.specs))
      )
        fail();
      for (const spec of (suite.specs ?? []) as Record<string, unknown>[]) {
        if (
          !spec ||
          typeof spec.id !== "string" ||
          !spec.id ||
          !Array.isArray(spec.tests) ||
          !spec.tests.length
        )
          fail();
        for (const test of spec.tests as Record<string, unknown>[]) {
          if (
            !test ||
            typeof test.projectId !== "string" ||
            test.expectedStatus !== "passed" ||
            (listing && (!Array.isArray(test.results) || test.results.length))
          )
            fail();
          ids.push(JSON.stringify([spec.id, test.projectId]));
        }
      }
      visit(suite.suites ?? []);
    }
  };
  visit(report.suites);
  if (!ids.length || new Set(ids).size !== ids.length) return fail();
  if (listing) {
    const stats = report.stats as Record<string, unknown> | undefined;
    if (
      stats?.expected !== 0 ||
      stats.skipped !== ids.length ||
      stats.unexpected !== 0 ||
      stats.flaky !== 0
    )
      return fail();
  }
  return ids.sort();
};
const requiredCommands = ["format:check", "lint", "typecheck", "test", "build"] as const;
type Group =
  | "tools"
  | "format:check"
  | "lint"
  | "typecheck"
  | "test:unit"
  | "test:integration"
  | "test:e2e"
  | "build"
  | "migrate:check"
  | "generate:check"
  | "scan";
export type Stage = {
  id: string;
  group: Group;
  command: string[];
  cwd: string;
  timeoutMs: number;
  expectedTests?: string[];
  listedTests?: Record<string, string[]>;
  bunTests?: boolean;
  browserTests?: boolean;
  browserList?: boolean;
  env?: Record<string, string>;
  emptyOutput?: boolean;
};
type Result = { code: number; stdout: string; failedTests?: string[] };
type Runner = (stage: Stage) => Promise<Result>;
type Workspace = { path: string; name: string; dependencies: string[] };
type GoTests = { path: string; unit: string[]; integration: string[] };
type Context = {
  root: string;
  workspaces: Workspace[];
  goTests: GoTests[];
  goFiles: string[];
  scriptTests: string[];
  goModule: string;
};

// Exact container-backed groups, independent of file naming. All remaining
// discovered top-level tests run as units; both groups always use the race detector.
// A renamed/deleted classified test fails discovery instead of silently dropping coverage.
export const integrationTests: Record<string, string[]> = {
  "internal/app": [
    "TestAPIMigratorSpecialCredentials",
    "TestInvitationWorkerActualRedisPostgres",
    "TestInvitationWorkerRoleOwnership",
    "TestRoleRealDependenciesRemainIndependent",
    "TestRoleBinaryStartup",
    "TestPartialStartupRealLifecycleResources",
    "TestSIGTERMActiveHTTPSubprocess",
    "TestSIGTERMRoleProcesses",
    "TestTracestateHTTPRedisOTLP",
    "TestCollectorExportCompletenessBarrier",
    "TestCollectorRedactsAllSignals",
    "TestTelemetryOutageReadinessAndCleanup",
    "TestTelemetryOutageMigrator",
  ],
  "internal/database": [
    "TestDatabaseTelemetryParameterizedPostgres",
    "TestMigrationEmptyDatabaseAndBinary",
  ],
  "internal/handler": [
    "TestRoleHealthActualHTTP",
    "TestProductActualHTTP",
    "TestControlPlaneTenantDenials",
  ],
  "internal/lib/job": ["TestCorrelationRetryLegacyRedis"],
  "internal/service": ["TestBearerBrowserFixture"],
  "internal/testing": ["TestBrowserProductFixture"],
};

const filesUnder = async (directory: string): Promise<string[]> => {
  const files: string[] = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    if (["node_modules", "dist", ".git", "build", ".next", "tmp", ".turbo"].includes(entry.name))
      continue;
    const path = join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await filesUnder(path)));
    else if (entry.isFile()) files.push(path);
  }
  return files.sort();
};

export const discover = async (root = repositoryRoot): Promise<Context> => {
  const manifest = JSON.parse(await readFile(join(root, "package.json"), "utf8"));
  if (!Array.isArray(manifest.workspaces) || manifest.workspaces.length === 0)
    throw new Error("Missing workspace manifest");
  const workspaces: Workspace[] = [];
  for (const pattern of manifest.workspaces) {
    if (typeof pattern !== "string" || !/^(apps|packages)\/\*$/.test(pattern))
      throw new Error("Unsupported workspace pattern");
    const parent = pattern.slice(0, -2);
    for (const entry of await readdir(join(root, parent), { withFileTypes: true })) {
      if (!entry.isDirectory()) continue;
      const path = `${parent}/${entry.name}`;
      const allFiles = await filesUnder(join(root, path));
      if (!allFiles.some((file) => /\.(ts|tsx|js|jsx)$/.test(file))) continue;
      const pkg = JSON.parse(await readFile(join(root, path, "package.json"), "utf8"));
      if (typeof pkg.name !== "string" || !pkg.name.trim())
        throw new Error(`Missing workspace name: ${path}`);
      for (const command of requiredCommands) {
        if (typeof pkg.scripts?.[command] !== "string" || !pkg.scripts[command].trim())
          throw new Error(`Missing ${command}: ${path}`);
      }
      if (!allFiles.some((file) => /\.(test|spec)\.[cm]?[jt]sx?$/.test(file)))
        throw new Error(`No tests discovered: ${path}`);
      workspaces.push({ path, name: pkg.name, dependencies: Object.keys(pkg.dependencies ?? {}) });
    }
  }
  for (const path of ["packages/zod", "packages/openapi", "packages/emails", "apps/frontend"]) {
    if (!workspaces.some((workspace) => workspace.path === path))
      throw new Error(`Omitted workspace: ${path}`);
  }
  // Build dependencies before consumers, even when invoked without Turbo/cache.
  const ordered: Workspace[] = [];
  const pending = [...workspaces];
  while (pending.length) {
    const index = pending.findIndex(
      (workspace) =>
        !workspace.dependencies.some((name) =>
          pending.some((candidate) => candidate.name === name),
        ),
    );
    if (index < 0) throw new Error("Workspace dependency cycle");
    const workspace = pending.splice(index, 1)[0];
    if (workspace) ordered.push(workspace);
  }
  const backend = join(root, "apps/backend");
  const moduleFile = await readFile(join(backend, "go.mod"), "utf8");
  const goModule = /^module\s+(\S+)$/m.exec(moduleFile)?.[1];
  if (!goModule) throw new Error("Missing Go module");
  const goFiles = (await filesUnder(backend)).filter((file) => file.endsWith(".go"));
  const packages = new Map<string, string[]>();
  for (const file of goFiles.filter((file) => file.endsWith("_test.go"))) {
    const source = await readFile(file, "utf8");
    const names = [...source.matchAll(/^func (Test\w+)\(\s*\w+\s+\*\w+\.T\s*\)/gm)]
      .map((match) => match[1])
      .filter((name): name is string => !!name);
    const path = relative(backend, dirname(file));
    packages.set(path, [...(packages.get(path) ?? []), ...names]);
  }
  if (![...packages.values()].some((tests) => tests.length > 0))
    throw new Error("No Go tests discovered");
  for (const [path, names] of Object.entries(integrationTests)) {
    for (const name of names)
      if (!packages.get(path)?.includes(name))
        throw new Error(`Missing classified integration test: ${path}/${name}`);
  }
  const goTests = [...packages.entries()]
    .map(([path, names]) => {
      if (new Set(names).size !== names.length) throw new Error(`Duplicate Go tests: ${path}`);
      const integration = integrationTests[path] ?? [];
      return {
        path,
        unit: names.filter((name) => !integration.includes(name)).sort(),
        integration: [...integration].sort(),
      };
    })
    .sort((a, b) => a.path.localeCompare(b.path));
  const scriptTests = (await filesUnder(join(root, "scripts")).catch(() => [])).filter((file) =>
    /\.(test|spec)\.ts$/.test(file),
  );
  return { root, workspaces: ordered, goTests, goFiles, scriptTests, goModule };
};

export const createStages = (context: Context, mode: "fast" | "full"): Stage[] => {
  const { root, workspaces, goTests, goFiles, scriptTests, goModule } = context;
  const backend = join(root, "apps/backend");
  const stages: Stage[] = [];
  const add = (
    id: string,
    group: Group,
    command: string[],
    cwd = root,
    extra: Partial<Stage> = {},
  ) => {
    stages.push({ id, group, command, cwd, timeoutMs: 300000, ...extra });
  };
  // Reject the initial checked bytes before any build or test can publish assets.
  add("generate:check", "generate:check", ["bun", "run", "generate:check"], root, {
    timeoutMs: 600000,
  });
  add("tools:verify", "tools", ["bun", "scripts/install-tools.ts", "--verify"]);
  add("backend:modules", "tools", ["go", "mod", "verify"], backend);
  add("backend:format", "format:check", ["gofmt", "-l", ...goFiles], backend, {
    emptyOutput: true,
  });
  add("scripts:format", "format:check", [join(root, "tmp/tools/biome"), "format", "scripts"]);
  for (const workspace of workspaces)
    add(
      `${workspace.name}:format:check`,
      "format:check",
      ["bun", "run", "format:check"],
      join(root, workspace.path),
    );
  // Go's compiler is the authority for discoverable tests. Compare the real
  // listing before execution so signatures, build constraints or TestMain
  // cannot make a source-only selection silently omit or invent a test.
  add(
    "backend:test-discovery",
    "test:unit",
    ["go", "test", "-json", "-list", "^Test", "./..."],
    backend,
    {
      listedTests: Object.fromEntries(
        goTests
          .filter((pkg) => pkg.unit.length + pkg.integration.length > 0)
          .map((pkg) => [`${goModule}/${pkg.path}`, [...pkg.unit, ...pkg.integration].sort()]),
      ),
    },
  );
  add("backend:vet", "lint", ["go", "vet", "./..."], backend);
  add(
    "backend:staticcheck",
    "lint",
    [join(root, "tmp/tools/staticcheck"), "-checks=all", "./..."],
    backend,
  );
  add(
    "backend:golangci",
    "lint",
    [
      join(root, "tmp/tools/golangci-lint"),
      "run",
      "--max-issues-per-linter=0",
      "--max-same-issues=0",
    ],
    backend,
  );
  add("scripts:lint", "lint", [
    join(root, "tmp/tools/biome"),
    "lint",
    "scripts",
    "--error-on-warnings",
  ]);
  for (const workspace of workspaces)
    add(`${workspace.name}:lint`, "lint", ["bun", "run", "lint"], join(root, workspace.path));
  // Authored package builds precede type checks/tests that import workspace exports.
  for (const workspace of workspaces)
    add(`${workspace.name}:build`, "build", ["bun", "run", "build"], join(root, workspace.path));
  add("backend:typecheck", "typecheck", ["go", "test", "-run", "^$", "./..."], backend);
  add("scripts:typecheck", "typecheck", [
    join(root, "node_modules/.bin/tsc"),
    "-p",
    "scripts/tsconfig.json",
  ]);
  for (const workspace of workspaces)
    add(
      `${workspace.name}:typecheck`,
      "typecheck",
      ["bun", "run", "typecheck"],
      join(root, workspace.path),
    );
  for (const workspace of workspaces)
    add(`${workspace.name}:test`, "test:unit", ["bun", "run", "test"], join(root, workspace.path), {
      bunTests: true,
    });
  for (const workspace of workspaces.filter((workspace) => workspace.path === "apps/frontend")) {
    add(
      `${workspace.name}:browser-install`,
      "test:e2e",
      ["bun", "run", "browser:install"],
      join(root, workspace.path),
    );
    add(
      `${workspace.name}:browser-discovery`,
      "test:e2e",
      ["bun", "run", "test:e2e", "--list"],
      join(root, workspace.path),
      { browserList: true },
    );
    add(
      `${workspace.name}:test:e2e`,
      "test:e2e",
      ["bun", "run", "test:e2e"],
      join(root, workspace.path),
      {
        browserTests: true,
        timeoutMs: 600000,
      },
    );
  }
  if (scriptTests.length)
    add("scripts:test", "test:unit", ["bun", "test", ...scriptTests], root, {
      bunTests: true,
      timeoutMs: 600000,
    });
  add("tools:self-test", "test:unit", ["bun", "scripts/install-tools.ts", "--self-test"]);
  for (const pkg of goTests) {
    if (pkg.unit.length)
      add(
        `backend:unit:${pkg.path}`,
        "test:unit",
        [
          "go",
          "test",
          "-race",
          "-count=1",
          "-timeout=8m",
          "-json",
          "-run",
          `^(${pkg.unit.join("|")})$`,
          `./${pkg.path}`,
        ],
        backend,
        { expectedTests: pkg.unit, timeoutMs: 540000 },
      );
  }
  if (mode === "full") {
    for (const pkg of goTests) {
      if (pkg.integration.length)
        add(
          `backend:integration:${pkg.path}`,
          "test:integration",
          [
            "go",
            "test",
            "-race",
            "-count=1",
            "-timeout=12m",
            "-json",
            "-run",
            `^(${pkg.integration.join("|")})$`,
            `./${pkg.path}`,
          ],
          backend,
          { expectedTests: pkg.integration, timeoutMs: 780000 },
        );
    }
    add(
      "migrate:check",
      "migrate:check",
      [
        "go",
        "test",
        "-race",
        "-count=1",
        "-timeout=5m",
        "-json",
        "-run",
        "^TestMigrationEmptyDatabaseAndBinary$",
        "./internal/database",
      ],
      backend,
      { expectedTests: ["TestMigrationEmptyDatabaseAndBinary"], timeoutMs: 360000 },
    );
  }
  add(
    "backend:build",
    "build",
    ["go", "build", "-o", join(root, "tmp/check-bin/"), "./cmd/..."],
    backend,
  );
  if (mode === "full") add("scan", "scan", ["bun", "run", "scan"], root, { timeoutMs: 900000 });
  return stages;
};

// Extract only repository-discovered test identifiers; subtest labels, output,
// package names and malformed events remain private.
const failedGoTests = (stage: Stage, stdout: string): string[] => {
  const failed = new Set<string>();
  for (const line of stdout.split("\n")) {
    try {
      const event = JSON.parse(line);
      if (event?.Action !== "fail" || typeof event.Test !== "string") continue;
      const name = event.Test.split("/")[0];
      if (stage.expectedTests?.includes(name)) failed.add(name);
    } catch {
      // Incomplete or non-JSON diagnostics cannot contribute public metadata.
    }
  }
  return (stage.expectedTests ?? []).filter((name) => failed.has(name));
};

// Capture at most 16 MiB and never print subprocess diagnostics. On timeout or
// overflow terminate the whole subprocess group, including compiler/container helpers.
export const runCommand: Runner = async (stage) =>
  new Promise((complete) => {
    let stdout = "";
    let bytes = 0;
    let failed = false;
    const child = spawn(stage.command[0] ?? "", stage.command.slice(1), {
      cwd: stage.cwd,
      env: {
        ...process.env,
        ...stage.env,
        PATH: `${join(repositoryRoot, "tmp/tools")}:${process.env.PATH ?? ""}`,
      },
      stdio: ["ignore", "pipe", "pipe"],
      detached: true,
    });
    const terminate = () => {
      failed = true;
      if (child.pid) {
        try {
          process.kill(-child.pid, "SIGKILL");
        } catch {
          child.kill("SIGKILL");
        }
      }
    };
    const capture = (chunk: Buffer, retain: boolean) => {
      bytes += chunk.length;
      if (bytes > 16 * 1024 * 1024) terminate();
      else if (retain) stdout += chunk.toString("utf8");
    };
    child.stdout.on("data", (chunk: Buffer) => capture(chunk, true));
    child.stderr.on("data", (chunk: Buffer) => capture(chunk, !!stage.bunTests));
    const timer = setTimeout(terminate, stage.timeoutMs);
    child.once("error", () => {
      failed = true;
    });
    child.once("close", (code) => {
      clearTimeout(timer);
      const failedTests = stage.expectedTests ? failedGoTests(stage, stdout) : [];
      complete({
        code: failed ? 1 : (code ?? 1),
        stdout: failed || code !== 0 ? "" : stdout,
        ...(failedTests.length ? { failedTests } : {}),
      });
    });
  });

// Provider diagnostics and signed tokens stay private; readiness exposes only a
// validated loopback protocol URL owned by the test binary.

export const selectedBrowserProject = (args: string[] = []): "local" | "provider" => {
  const projects: string[] = [];
  for (let i = 0; i < args.length; i++) {
    const argument = args[i] ?? "";
    if (argument.startsWith("--project=")) projects.push(argument.slice(10));
    else if (args[i] === "--project") projects.push(args[++i] ?? "");
  }
  if (
    projects.length > 1 ||
    (projects.length === 1 && !["local", "provider"].includes(projects[0] ?? ""))
  )
    throw new Error("Select exactly one supported browser project");
  return projects[0] === "provider" ? "provider" : "local";
};

// Every resource below belongs to this one acceptance run. No provider transport,
// auth implementation, production endpoint or host network is replaced.
export const startProviderIngress = async (
  target: () => { host: string; port: number } | undefined,
  port = 8080,
) => {
  const sockets = new Set<Socket>();
  const server = createTCPServer((client) => {
    const endpoint = target();
    if (!endpoint) {
      client.destroy();
      return;
    }
    const upstream = connect(endpoint);
    for (const socket of [client, upstream]) {
      sockets.add(socket);
      socket.on("error", () => socket.destroy());
      socket.once("close", () => {
        sockets.delete(socket);
        client.destroy();
        upstream.destroy();
      });
    }
    client.pipe(upstream).pipe(client);
  });
  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(port, "127.0.0.1", resolve);
  });
  let stopping: Promise<void> | undefined;
  return {
    port: (server.address() as import("node:net").AddressInfo).port,
    stop: () =>
      (stopping ??= new Promise<void>((resolve) => {
        for (const socket of sockets) socket.destroy();
        server.close(() => resolve());
      })),
  };
};

const startProviderFixture = async (root: string) => {
  const { validateProviderConfig, privateProviderDirectory } = await import(
    "../apps/frontend/tests/provider.config.ts"
  );
  if (validateProviderConfig(process.env).length)
    throw new Error("Provider prerequisites invalid; run provider:check");
  const directory = await privateProviderDirectory(process.env, root);
  const owned = await mkdtemp(join(directory, "runtime-"));
  const runID = randomUUID();
  const containers: string[] = [];
  const networks: string[] = [];
  let api = "";
  let apiAddress: { host: string; port: number } | undefined;
  let ingress: Awaited<ReturnType<typeof startProviderIngress>> | undefined;
  let worker: ReturnType<typeof spawn> | undefined;
  let control: ReturnType<typeof createServer> | undefined;
  const controlKey = randomBytes(32).toString("hex");
  const base: Record<string, string> = Object.fromEntries(
    Object.entries(process.env).filter(
      (entry): entry is [string, string] =>
        !!entry[1] &&
        (entry[0].startsWith("FLUX_") ||
          entry[0].startsWith("CLERK_") ||
          entry[0] === "NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY"),
    ),
  );
  const run = async (command: string[], env: Record<string, string> = {}, cwd = root) => {
    const result = await runCommand({
      id: "provider:private-operation",
      group: "test:e2e",
      command,
      cwd,
      timeoutMs: 90000,
      env,
    });
    if (result.code) throw new Error("Provider runtime operation failed");
    return result.stdout.trim();
  };
  const docker = (...args: string[]) => run(["docker", ...args], base);
  const removeContainer = async (id: string) => {
    await docker("rm", "--force", "--volumes", id);
    containers.splice(containers.indexOf(id), 1);
  };
  const stopWorker = async () => {
    if (!worker) return;
    const current = worker;
    worker = undefined;
    await new Promise<void>((complete, reject) => {
      if (current.exitCode !== null) {
        current.exitCode === 0 ? complete() : reject(new Error("Provider worker exited"));
        return;
      }
      const timer = setTimeout(() => {
        current.kill("SIGKILL");
        reject(new Error("Provider worker drain failed"));
      }, 35000);
      current.once("close", (code) => {
        clearTimeout(timer);
        code === 0 ? complete() : reject(new Error("Provider worker drain failed"));
      });
      current.kill("SIGTERM");
    });
  };
  const stop = async () => {
    const errors: unknown[] = [];
    try {
      await stopWorker();
    } catch (error) {
      errors.push(error);
    }
    await new Promise<void>((done) => (control ? control.close(() => done()) : done()));
    await ingress?.stop();
    for (const id of [...containers].reverse()) {
      try {
        await removeContainer(id);
      } catch (error) {
        errors.push(error);
      }
    }
    for (const id of [...networks].reverse()) {
      try {
        await docker("network", "rm", id);
      } catch (error) {
        errors.push(error);
      }
    }
    await rm(owned, { recursive: true, force: true });
    if (errors.length) throw new Error("Provider runtime cleanup failed");
  };
  const ready = async (url: string) => {
    const deadline = Date.now() + 30000;
    while (Date.now() < deadline) {
      try {
        if ((await fetch(url, { signal: AbortSignal.timeout(2000) })).status === 200) return;
      } catch {
        /* bounded startup */
      }
      await new Promise((done) => setTimeout(done, 200));
    }
    throw new Error("Provider runtime readiness failed");
  };
  try {
    const pinSource = await readFile(
      join(root, "apps/backend/internal/testing/container.go"),
      "utf8",
    );
    const postgresImage = pinSource.match(/PostgresImage\s*=\s*"([^"]+@sha256:[a-f0-9]{64})"/)?.[1];
    const redisImage = pinSource.match(/RedisImage\s*=\s*"([^"]+@sha256:[a-f0-9]{64})"/)?.[1];
    if (!postgresImage || !redisImage) throw new Error("Provider runtime image pins missing");
    const internal = await docker(
      "network",
      "create",
      "--internal",
      "--label",
      `flux.provider-run=${runID}`,
      `flux-provider-internal-${runID}`,
    );
    networks.push(internal);
    const egress = await docker(
      "network",
      "create",
      "--label",
      `flux.provider-run=${runID}`,
      `flux-provider-egress-${runID}`,
    );
    networks.push(egress);
    const postgresName = `flux-provider-pg-${runID}`;
    const password = randomBytes(32).toString("hex");
    const postgres = await run(
      [
        "docker",
        "run",
        "--detach",
        "--name",
        postgresName,
        "--label",
        `flux.provider-run=${runID}`,
        "--network",
        egress,
        "--publish",
        "127.0.0.1::5432",
        "--tmpfs",
        "/var/lib/postgresql/data",
        "--env",
        "POSTGRES_PASSWORD",
        "--env",
        "POSTGRES_USER=flux_provider",
        "--env",
        "POSTGRES_DB=flux_provider",
        postgresImage,
      ],
      { POSTGRES_PASSWORD: password },
    );
    containers.push(postgres);
    await docker("network", "connect", "--alias", postgresName, internal, postgres);
    const redisName = `flux-provider-redis-${runID}`;
    const redis = await docker(
      "run",
      "--detach",
      "--name",
      redisName,
      "--label",
      `flux.provider-run=${runID}`,
      "--network",
      egress,
      "--publish",
      "127.0.0.1::6379",
      redisImage,
    );
    containers.push(redis);
    await docker("network", "connect", "--alias", redisName, internal, redis);
    const mapped = (output: string) => {
      const match = output.match(/^127\.0\.0\.1:(\d+)$/);
      if (!match?.[1]) throw new Error("Invalid owned container port");
      return match[1];
    };
    const postgresPort = mapped(await docker("port", postgres, "5432/tcp"));
    const redisPort = mapped(await docker("port", redis, "6379/tcp"));
    let readyDependencies = false;
    for (let attempt = 0; attempt < 100; attempt++) {
      const pg = await runCommand({
        id: "provider:postgres-readiness",
        group: "test:e2e",
        command: [
          "docker",
          "exec",
          postgres,
          "pg_isready",
          "-U",
          "flux_provider",
          "-d",
          "flux_provider",
        ],
        cwd: root,
        timeoutMs: 2000,
      });
      if (!pg.code && (await docker("exec", redis, "redis-cli", "ping")) === "PONG") {
        readyDependencies = true;
        break;
      }
      await new Promise((done) => setTimeout(done, 200));
    }
    if (!readyDependencies) throw new Error("Provider dependencies unavailable");
    Object.assign(base, {
      "FLUX_PRIMARY.ENV": "development",
      "FLUX_DATABASE.HOST": "127.0.0.1",
      "FLUX_DATABASE.PORT": postgresPort,
      "FLUX_DATABASE.USER": "flux_provider",
      "FLUX_DATABASE.PASSWORD": password,
      "FLUX_DATABASE.NAME": "flux_provider",
      "FLUX_DATABASE.SSL_MODE": "disable",
      "FLUX_DATABASE.MAX_OPEN_CONNS": "5",
      "FLUX_DATABASE.MAX_IDLE_CONNS": "2",
      "FLUX_DATABASE.CONN_MAX_LIFETIME": "60",
      "FLUX_DATABASE.CONN_MAX_IDLE_TIME": "30",
      "FLUX_REDIS.ADDRESS": `127.0.0.1:${redisPort}`,
      "FLUX_API.PRODUCER_ENABLED": "false",
      "FLUX_SERVER.CORS_ALLOWED_ORIGINS": "http://localhost:3000",
      "FLUX_API.LISTEN_ADDRESS": "0.0.0.0:8080",
      "FLUX_WORKER.LISTEN_ADDRESS": "127.0.0.1:8082",
    });
    await run(
      ["go", "build", "-o", `${owned}/`, "./cmd/api", "./cmd/migrator", "./cmd/worker"],
      { CGO_ENABLED: "0" },
      join(root, "apps/backend"),
    );
    await run([join(owned, "migrator")], base, owned);
    const originalKey = base["FLUX_INVITATIONS.ACTIVE_KEY_ID"] ?? "";
    const ring = JSON.parse(base["FLUX_INVITATIONS.ENCRYPTION_KEYS"] ?? "{}") as Record<
      string,
      string
    >;
    const rotatedKey = Object.keys(ring).find((id) => id !== originalKey);
    if (!originalKey || !rotatedKey) throw new Error("Provider rotation keys unavailable");
    const startAPI = async (active: string) => {
      apiAddress = undefined;
      if (api) await removeContainer(api);
      base["FLUX_INVITATIONS.ACTIVE_KEY_ID"] = active;
      const containerEnv = {
        ...base,
        "FLUX_DATABASE.HOST": postgresName,
        "FLUX_DATABASE.PORT": "5432",
        "FLUX_REDIS.ADDRESS": `${redisName}:6379`,
      };
      const args = [
        "docker",
        "create",
        "--label",
        `flux.provider-run=${runID}`,
        "--network",
        internal,
        "--workdir",
        "/app",
        "--user",
        "65534:65534",
        "--mount",
        `type=bind,source=${join(owned, "api")},target=/run/flux-api,readonly`,
        "--mount",
        `type=bind,source=${join(root, "apps/backend/static")},target=/app/static,readonly`,
        "--entrypoint",
        "/run/flux-api",
      ];
      for (const name of Object.keys(containerEnv).filter(
        (name) => name.startsWith("FLUX_") && !name.startsWith("FLUX_PROVIDER_"),
      ))
        args.push("--env", name);
      args.push(postgresImage);
      api = await run(args, containerEnv);
      containers.push(api);
      await docker("network", "connect", egress, api);
      await docker("start", api);
      const attachments = JSON.parse(
        await docker("inspect", "--format", "{{json .NetworkSettings.Networks}}", api),
      ) as Record<string, { NetworkID: string; IPAddress: string }>;
      const address = Object.values(attachments).find((item) => item.NetworkID === internal);
      if (!address || isIP(address.IPAddress) !== 4)
        throw new Error("Owned provider internal address unavailable");
      apiAddress = { host: address.IPAddress, port: 8080 };
      // Docker removes published ports when the last external network is
      // detached. Keep loopback ingress on the owned internal address instead.
      ingress ??= await startProviderIngress(() => apiAddress);
      await ready("http://127.0.0.1:8080/ready");
    };
    await startAPI(originalKey);
    const sql = async (statement: string) =>
      JSON.parse(
        await docker(
          "exec",
          postgres,
          "psql",
          "-U",
          "flux_provider",
          "-d",
          "flux_provider",
          "-At",
          "-v",
          "ON_ERROR_STOP=1",
          "-c",
          statement,
        ),
      ) as unknown;
    const uuid = (value: unknown): string => {
      if (
        typeof value !== "string" ||
        !/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/.test(value)
      )
        throw new Error("Invalid provider assertion scope");
      return value;
    };
    control = createServer(async (request, response) => {
      try {
        if (request.method !== "POST" || request.headers.authorization !== `Bearer ${controlKey}`)
          throw new Error();
        let body = "";
        for await (const chunk of request) {
          body += chunk;
          if (body.length > 4096) throw new Error();
        }
        const input = JSON.parse(body) as {
          action: string;
          workspace?: string;
          invitation?: string;
          user?: string;
        };
        let result: unknown;
        switch (input.action) {
          case "owner-proof": {
            const workspace = uuid(input.workspace);
            result = await sql(
              `SELECT json_build_object('owners',(SELECT count(*) FROM memberships WHERE workspace_id='${workspace}' AND role='owner'),'workspaces',(SELECT count(*) FROM workspaces WHERE id='${workspace}'))`,
            );
            break;
          }
          case "delivery-proof": {
            const workspace = uuid(input.workspace),
              invitation = uuid(input.invitation);
            result = await sql(
              `SELECT json_build_object('queued',state='queued','delivered',state='delivered','erased',ciphertext IS NULL,'old_key',key_id='${originalKey}','new_key',key_id='${rotatedKey}') FROM invitation_delivery_intents WHERE workspace_id='${workspace}' AND invitation_id='${invitation}' AND current`,
            );
            break;
          }
          case "remove-member": {
            const workspace = uuid(input.workspace),
              user = uuid(input.user);
            result = await sql(
              `WITH removed AS (DELETE FROM memberships WHERE workspace_id='${workspace}' AND user_id='${user}' AND role<>'owner' RETURNING id) SELECT json_build_object('removed',(SELECT count(*) FROM removed))`,
            );
            break;
          }
          case "rotate-start-worker":
            await startAPI(rotatedKey);
            worker = spawn(join(owned, "worker"), [], {
              cwd: owned,
              env: { ...process.env, ...base },
              stdio: "ignore",
            });
            await ready("http://127.0.0.1:8082/ready");
            result = { ready: true };
            break;
          case "stop-worker":
            await stopWorker();
            result = { stopped: true };
            break;
          case "disconnect-provider":
            await docker("network", "disconnect", egress, api);
            await ready("http://127.0.0.1:8080/ready");
            if ((await docker("exec", redis, "redis-cli", "ping")) !== "PONG") throw new Error();
            result = { database_ready: true, redis_ready: true };
            break;
          case "reconnect-provider":
            await docker("network", "connect", egress, api);
            result = { connected: true };
            break;
          default:
            throw new Error();
        }
        response.writeHead(200, {
          "Content-Type": "application/json",
          "Cache-Control": "no-store",
        });
        response.end(JSON.stringify(result));
      } catch {
        response.writeHead(500);
        response.end();
      }
    });
    const server = control;
    await new Promise<void>((complete) => server.listen(0, "127.0.0.1", complete));
    const address = control.address();
    if (!address || typeof address === "string") throw new Error();
    return {
      env: {
        FLUX_PROVIDER_RUN: "1",
        FLUX_PROVIDER_CONTROL_ORIGIN: `http://127.0.0.1:${address.port}`,
        FLUX_PROVIDER_CONTROL_KEY: controlKey,
      },
      stop,
    };
  } catch {
    await stop();
    throw new Error("Provider runtime startup failed");
  }
};
const startBearerFixture = async (root: string) => {
  const child = spawn(
    "go",
    [
      "test",
      "-run",
      "^TestBrowserProductFixture$",
      "-v",
      "./internal/testing",
      "-args",
      "-browser-fixture",
    ],
    {
      cwd: join(root, "apps/backend"),
      env: process.env,
      stdio: ["ignore", "pipe", "ignore"],
      detached: true,
    },
  );
  const kill = () => {
    if (child.pid) {
      try {
        process.kill(-child.pid, "SIGKILL");
      } catch {
        child.kill("SIGKILL");
      }
    }
  };
  try {
    const url = await new Promise<string>((complete, reject) => {
      let output = "";
      const timer = setTimeout(() => reject(new Error("Bearer fixture readiness failed")), 90000);
      child.once("error", () => {
        clearTimeout(timer);
        reject(new Error("Bearer fixture startup failed"));
      });
      child.once("close", () => {
        clearTimeout(timer);
        reject(new Error("Bearer fixture exited"));
      });
      child.stdout.on("data", (chunk: Buffer) => {
        output += chunk.toString("utf8");
        if (output.length > 65536) {
          clearTimeout(timer);
          reject(new Error("Bearer fixture output limit"));
        }
        const ready = output.match(/(?:^|\n)FLUX_BROWSER_FIXTURE (http:\/\/127\.0\.0\.1:\d+)\r?\n/);
        if (ready?.[1]) {
          clearTimeout(timer);
          complete(ready[1]);
        }
      });
    });
    return {
      url,
      stop: async () => {
        try {
          const response = await fetch(`${url}/stop`, {
            method: "POST",
            signal: AbortSignal.timeout(3000),
          });
          if (response.status !== 204) throw new Error();
          await new Promise<void>((complete, reject) => {
            if (child.exitCode !== null) {
              if (child.exitCode === 0) complete();
              else reject(new Error());
              return;
            }
            const timer = setTimeout(() => {
              kill();
              reject(new Error());
            }, 5000);
            child.once("close", (code) => {
              clearTimeout(timer);
              if (code === 0) complete();
              else reject(new Error());
            });
          });
        } catch {
          kill();
          throw new Error("Browser fixture teardown failed");
        }
      },
    };
  } catch {
    kill();
    throw new Error("Bearer fixture readiness failed");
  }
};

export const saveProviderEvidence = async (
  root: string,
  output: string,
  env: NodeJS.ProcessEnv = process.env,
) => {
  assertBrowserReport(output);
  const {
    configFingerprint,
    manualProviderRows,
    privateProviderDirectory,
    proofHash,
    requiredProviderRows,
    validateAcceptance,
  } = await import("../apps/frontend/tests/provider.config.ts");
  type Proof = import("../apps/frontend/tests/provider.config.ts").ProviderProof;
  type Row = import("../apps/frontend/tests/provider.config.ts").ProviderRow;
  const tests: Proof["tests"] = [];
  const visit = (suites: Record<string, unknown>[]) => {
    for (const suite of suites) {
      for (const spec of (suite.specs ?? []) as Record<string, unknown>[]) {
        for (const test of spec.tests as Record<string, unknown>[]) {
          if (test.projectId !== "provider")
            throw new Error("Provider proof contains another project");
          const rows = (test.annotations as { type: string; description?: string }[])
            .filter((entry) => entry.type === "flux-provider-row")
            .map((entry) => JSON.parse(entry.description ?? ""));
          tests.push({
            id: JSON.stringify([spec.id, test.projectId]),
            project: "provider",
            status: "passed",
            retry: 0,
            rows,
          });
        }
      }
      visit((suite.suites ?? []) as Record<string, unknown>[]);
    }
  };
  visit(JSON.parse(output).suites);
  const revision = await runCommand({
    id: "provider:revision",
    group: "test:e2e",
    command: ["git", "rev-parse", "HEAD"],
    cwd: root,
    timeoutMs: 3000,
  });
  const proof: Proof = {
    schema: 1,
    run_id: randomUUID(),
    created_at: new Date().toISOString(),
    revision: revision.stdout.trim(),
    config_fingerprint: configFingerprint(env),
    tests,
  };
  const rows: Row[] = requiredProviderRows.map((id) => {
    const observed = tests.filter((entry) => entry.rows.some((row) => row.id === id));
    const pending = observed.some((entry) =>
      entry.rows.some((row) => row.id === id && row.status === "human_pending"),
    );
    const manual = manualProviderRows[id as keyof typeof manualProviderRows];
    return {
      id,
      status: pending ? "human_pending" : "passed",
      evidence_kind: pending ? "human_only" : "automated",
      tests: observed.map((entry) => entry.id),
      ...(pending && manual ? { reason_code: manual.reasons[0], manual_steps: manual.steps } : {}),
    };
  });
  const record = { schema: 1, run_id: proof.run_id, proof_sha256: proofHash(proof), rows };
  if (!validateAcceptance(record, true, proof, configFingerprint(env)))
    throw new Error("Incomplete provider execution evidence");
  const directory = await privateProviderDirectory(env, root);
  // Unique run files are never overwritten; a stable proof reference is written
  // only inside the operator-owned private directory after native acceptance.
  const proofPath = join(directory, `provider-proof-${proof.run_id}.json`);
  await writeFile(proofPath, JSON.stringify(proof), { mode: 0o600, flag: "wx" });
  const recordPath = join(directory, `provider-acceptance-${proof.run_id}.md`);
  await writeFile(
    recordPath,
    `# Real-provider acceptance\n\nAutomated execution passed; human_pending rows remain incomplete.\n\n\`\`\`json\n${JSON.stringify(record, null, 2)}\n\`\`\`\n`,
    { mode: 0o600, flag: "wx" },
  );
  console.log(
    "Provider evidence prepared privately; promote the sanitized record and matching proof during plan 35.",
  );
};
export const runChecks = async (
  options: {
    root?: string;
    mode?: "fast" | "full";
    group?: Group;
    runner?: Runner;
    progress?: (id: string) => void;
    browserArgs?: string[];
    browserCompleted?: (count: number) => void;
  } = {},
) => {
  const project = selectedBrowserProject(options.browserArgs);
  if (project === "provider" && options.group !== "test:e2e")
    throw new Error("Provider selection requires the explicit test:e2e stage");
  const context = await discover(options.root);
  const all = createStages(context, options.mode ?? "full");
  const stages = options.group
    ? all.filter(
        (stage) =>
          stage.group === "tools" ||
          stage.group === options.group ||
          (["typecheck", "test:unit"].includes(options.group ?? "") &&
            stage.group === "build" &&
            stage.id !== "backend:build"),
      )
    : all;
  if (!stages.some((stage) => stage.group !== "tools")) throw new Error("No stages selected");
  if (project === "provider")
    stages.unshift({
      id: "provider:check",
      group: "test:e2e",
      command: ["bun", "apps/frontend/tests/provider.config.ts", "--check-config"],
      cwd: context.root,
      timeoutMs: 30000,
    });
  let browserManifest: string[] | undefined;
  for (const stage of stages) {
    if (stage.browserTests || stage.browserList) {
      const selectors = options.browserArgs ?? [];
      const explicit = selectors.some((arg) => arg === "--project" || arg.startsWith("--project="));
      stage.command.push(...(explicit ? selectors : [`--project=${project}`, ...selectors]));
      stage.env = { ...stage.env, FLUX_PROVIDER_RUN: project === "provider" ? "1" : "0" };
    }
    options.progress?.(stage.id);
    if (stage.id === "scan") await cleanFrontendBuild(context.root);
    const fixture =
      stage.browserTests && !options.runner && project === "local"
        ? await startBearerFixture(context.root)
        : undefined;
    const provider =
      stage.browserTests && !options.runner && project === "provider"
        ? await startProviderFixture(context.root)
        : undefined;
    if (fixture) stage.env = { ...stage.env, FLUX_BROWSER_FIXTURE: fixture.url };
    if (provider) stage.env = { ...stage.env, ...provider.env };
    let result: Result;
    try {
      result = await (options.runner ?? runCommand)(stage);
    } finally {
      await fixture?.stop();
      await provider?.stop();
    }
    if (result.code !== 0 || (stage.emptyOutput && result.stdout.trim())) {
      const failedTests = (stage.expectedTests ?? []).filter((name) =>
        result.failedTests?.includes(name),
      );
      const detail = failedTests.length ? ` (${failedTests.join(", ")})` : "";
      throw new Error(`Check failed: ${stage.id}${detail}`);
    }
    if (stage.bunTests && !/[1-9]\d* pass\b/.test(result.stdout))
      throw new Error(`No successful tests: ${stage.id}`);
    if (stage.browserList) browserManifest = browserTestIds(result.stdout, true);
    if (stage.browserTests) {
      if (!browserManifest) throw new Error("Invalid browser test manifest");
      const completed = assertBrowserReport(result.stdout, browserManifest);
      if (project === "provider" && !options.runner)
        await saveProviderEvidence(context.root, result.stdout);
      options.browserCompleted?.(completed);
    }
    if (stage.listedTests) {
      const listed: Record<string, string[]> = {};
      for (const line of result.stdout.split("\n")) {
        if (!line.trim()) continue;
        let event: { Package?: string; Output?: string };
        try {
          event = JSON.parse(line);
        } catch {
          throw new Error(`Invalid Go test output: ${stage.id}`);
        }
        const name = event.Output?.trim();
        if (event.Package && name && /^Test\w+$/.test(name)) {
          const packageTests = listed[event.Package] ?? [];
          packageTests.push(name);
          listed[event.Package] = packageTests;
        }
      }
      const packages = new Set([...Object.keys(listed), ...Object.keys(stage.listedTests)]);
      for (const pkg of packages) {
        if (
          JSON.stringify((listed[pkg] ?? []).sort()) !==
          JSON.stringify(stage.listedTests[pkg] ?? [])
        )
          throw new Error(`Go test manifest mismatch: ${pkg}`);
      }
    }
    if (stage.expectedTests) {
      const passed = new Set<string>();
      for (const line of result.stdout.split("\n")) {
        if (!line.trim()) continue;
        try {
          const event = JSON.parse(line);
          if (event.Action === "skip" || event.Action === "fail")
            throw new Error(`Unsuccessful Go test: ${stage.id}`);
          if (event.Action === "pass" && typeof event.Test === "string") passed.add(event.Test);
        } catch (error) {
          if (error instanceof Error && error.message.startsWith("Unsuccessful Go test:"))
            throw error;
          throw new Error(`Invalid Go test output: ${stage.id}`);
        }
      }
      for (const name of stage.expectedTests)
        if (!passed.has(name)) throw new Error(`Missing successful Go test: ${stage.id}/${name}`);
    }
  }
};

if (process.argv[1] && pathToFileURL(resolve(process.argv[1])).href === import.meta.url) {
  try {
    const args = process.argv.slice(2);
    const group = args[0] === "--stage" ? (args[1] as Group) : undefined;
    const browserArgs =
      group === "test:e2e"
        ? args.slice(2).filter((arg, index) => !(index === 0 && arg === "--"))
        : [];
    const groups: Group[] = [
      "format:check",
      "lint",
      "typecheck",
      "test:unit",
      "test:integration",
      "test:e2e",
      "build",
      "migrate:check",
      "generate:check",
    ];
    if (
      !(args.length === 1 && ["--fast", "--full"].includes(args[0] ?? "")) &&
      !(
        args.length >= 2 &&
        group &&
        groups.includes(group) &&
        (args.length === 2 || group === "test:e2e")
      )
    )
      throw new Error("Usage: bun scripts/check.ts --fast|--full|--stage <group>");
    await runChecks({
      mode: args[0] === "--fast" ? "fast" : "full",
      group,
      browserArgs,
      browserCompleted: (count) => console.log(`Browser tests passed: ${count}`),
      progress: (id) => console.log(`Checking ${id}`),
    });
    console.log("Checks passed");
  } catch (error) {
    console.error(
      error instanceof Error &&
        /^(Check failed:|Missing |No |Invalid browser test manifest|Browser test manifest mismatch|Unsuccessful Go test:|Invalid Go test output:|Go test manifest mismatch:|Omitted |Unsupported |Workspace |Duplicate |Usage:)/.test(
          error.message,
        )
        ? error.message
        : "Quality check failed",
    );
    process.exitCode = 1;
  }
}
