import { spawn } from "node:child_process";
import { readdir, readFile } from "node:fs/promises";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const repositoryRoot = fileURLToPath(new URL("../", import.meta.url));
// Validate the complete JSON tree as well as aggregate stats. Retries, skips,
// empty selection and infrastructure errors must never produce green evidence.
export const assertBrowserReport = (output: string): number => {
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
      if (!suite || !Array.isArray(suite.specs) || !Array.isArray(suite.suites)) fail();
      for (const spec of suite.specs as Record<string, unknown>[]) {
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
      visit(suite.suites);
    }
  };
  visit(report.suites);
  if (completed !== stats.expected) return fail();
  return completed;
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
    "TestRoleRealDependenciesRemainIndependent",
    "TestRoleBinaryStartup",
    "TestPartialStartupRealLifecycleResources",
    "TestSIGTERMActiveHTTPSubprocess",
    "TestSIGTERMRoleProcesses",
    "TestTracestateHTTPRedisOTLP",
    "TestCollectorRedactsAllSignals",
    "TestTelemetryOutageReadinessAndCleanup",
    "TestTelemetryOutageMigrator",
  ],
  "internal/database": [
    "TestDatabaseTelemetryParameterizedPostgres",
    "TestMigrationEmptyDatabaseAndBinary",
  ],
  "internal/handler": ["TestRoleHealthActualHTTP"],
  "internal/lib/job": ["TestCorrelationRetryLegacyRedis"],
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
  for (const path of ["packages/zod", "packages/openapi", "packages/emails"]) {
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
    add("scripts:test", "test:unit", ["bun", "test", ...scriptTests], root, { bunTests: true });
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

export const runChecks = async (
  options: {
    root?: string;
    mode?: "fast" | "full";
    group?: Group;
    runner?: Runner;
    progress?: (id: string) => void;
    browserArgs?: string[];
  } = {},
) => {
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
  for (const stage of stages) {
    if (stage.browserTests) stage.command.push(...(options.browserArgs ?? []));
    options.progress?.(stage.id);
    const result = await (options.runner ?? runCommand)(stage);
    if (result.code !== 0 || (stage.emptyOutput && result.stdout.trim())) {
      const failedTests = (stage.expectedTests ?? []).filter((name) =>
        result.failedTests?.includes(name),
      );
      const detail = failedTests.length ? ` (${failedTests.join(", ")})` : "";
      throw new Error(`Check failed: ${stage.id}${detail}`);
    }
    if (stage.bunTests && !/[1-9]\d* pass\b/.test(result.stdout))
      throw new Error(`No successful tests: ${stage.id}`);
    if (stage.browserTests) assertBrowserReport(result.stdout);
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
          if (event.Action === "pass" && typeof event.Test === "string") passed.add(event.Test);
        } catch {
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
      progress: (id) => console.log(`Checking ${id}`),
    });
    console.log("Checks passed");
  } catch (error) {
    console.error(
      error instanceof Error &&
        /^(Check failed:|Missing |No |Invalid Go test output:|Go test manifest mismatch:|Omitted |Unsupported |Workspace |Duplicate |Usage:)/.test(
          error.message,
        )
        ? error.message
        : "Quality check failed",
    );
    process.exitCode = 1;
  }
}
