import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { spawn } from "node:child_process";
import {
  chmod,
  lstat,
  mkdir,
  mkdtemp,
  readFile,
  readdir,
  rename,
  rm,
  writeFile,
} from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { gunzipSync } from "node:zlib";

// Bun's existing generation script needs only this subprocess API. Use the
// locked Node declarations already supplied by the OpenAPI workspace, without
// installing a second runtime typings dependency for these scripts.
declare global {
  const Bun: {
    spawn: (
      cmd: string[],
      options: {
        cwd: string;
        env: NodeJS.ProcessEnv;
        stdin: "ignore";
        stdout: "pipe";
        stderr: "ignore";
      },
    ) => { stdout: ReadableStream<Uint8Array>; exited: Promise<number> };
  };
}

const repositoryRoot = fileURLToPath(new URL("../", import.meta.url));
export const defaultInstallRoot = join(repositoryRoot, "tmp/tools");
const names = ["biome", "golangci-lint", "staticcheck", "govulncheck", "gitleaks"] as const;

type Tool = {
  version: string;
  versionArgs: string[];
  versionOutput: string;
  url: string;
  sha256: string;
  binarySha256: string;
  archivePath?: string;
};
type IO = {
  download: (url: string) => Promise<Uint8Array>;
  execute: (
    cmd: string[],
    cwd?: string,
    env?: NodeJS.ProcessEnv,
    timeoutMs?: number,
  ) => Promise<string>;
};
type Build = {
  toolchain: string;
  module: string;
  sum: string;
  goModSum: string;
  goMod: string;
  goSum: string;
};
type Release = Omit<Tool, "url" | "sha256" | "binarySha256" | "archivePath"> & {
  source: string;
  platforms: Record<string, Pick<Tool, "url" | "sha256" | "binarySha256" | "archivePath">>;
  build?: Build;
};
const digest = (bytes: Uint8Array) => createHash("sha256").update(bytes).digest("hex");

const execute: IO["execute"] = (
  cmd,
  cwd = repositoryRoot,
  env = process.env,
  timeoutMs = 300_000,
) =>
  new Promise((accept, reject) => {
    const executable = cmd[0];
    if (!executable) {
      reject(new Error("tool command failed"));
      return;
    }
    const child = spawn(executable, cmd.slice(1), { cwd, env, stdio: ["ignore", "pipe", "pipe"] });
    const chunks: Buffer[] = [];
    let size = 0;
    let timedOut = false;
    const timer = setTimeout(() => {
      timedOut = true;
      child.kill("SIGKILL");
    }, timeoutMs);
    const collect = (chunk: Buffer, capture: boolean) => {
      size += chunk.length;
      if (size > 1024 * 1024) child.kill("SIGKILL");
      else if (capture) chunks.push(chunk);
    };
    child.stdout.on("data", (chunk: Buffer) => collect(chunk, true));
    child.stderr.on("data", (chunk: Buffer) => collect(chunk, false));
    child.on("error", () => {
      clearTimeout(timer);
      reject(new Error("tool command failed"));
    });
    child.on("close", (code) => {
      clearTimeout(timer);
      if (timedOut || size > 1024 * 1024 || code !== 0) reject(new Error("tool command failed"));
      else accept(Buffer.concat(chunks).toString("utf8"));
    });
  });

const download: IO["download"] = async (url) => {
  const response = await fetch(url, { signal: AbortSignal.timeout(120_000) });
  if (!response.ok || !response.body) throw new Error("download failed");
  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  try {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;
      size += value.length;
      if (size > 256 * 1024 * 1024) throw new Error("download exceeds limit");
      chunks.push(value);
    }
    return Buffer.concat(chunks);
  } finally {
    await reader.cancel();
  }
};

// Go makes downloaded module directories read-only. Only this installer-owned
// temporary tree is made writable; symlinks are never followed during cleanup.
const removeStage = async (root: string): Promise<void> => {
  const writable = async (directory: string): Promise<void> => {
    await chmod(directory, 0o700);
    for (const entry of await readdir(directory, { withFileTypes: true })) {
      if (entry.isDirectory()) await writable(join(directory, entry.name));
    }
  };
  await writable(root);
  await rm(root, { recursive: true, force: true });
};

// Extract only one regular file into memory, never archive paths onto disk.
const archiveBinary = (bytes: Uint8Array, member: string): Uint8Array => {
  const tar = gunzipSync(bytes, { maxOutputLength: 256 * 1024 * 1024 });
  let binary: Buffer | undefined;
  for (let offset = 0; offset + 512 <= tar.length; ) {
    const header = tar.subarray(offset, offset + 512);
    if (header.every((byte) => byte === 0)) break;
    const text = (start: number, length: number) =>
      header
        .subarray(start, start + length)
        .toString()
        .replace(/\0.*$/s, "");
    const size = Number.parseInt(text(124, 12).trim(), 8);
    if (!Number.isSafeInteger(size) || size < 0 || offset + 512 + size > tar.length)
      throw new Error("invalid archive");
    const prefix = text(345, 155);
    const path = `${prefix ? `${prefix}/` : ""}${text(0, 100)}`;
    if (path === member) {
      if (binary || (header[156] !== 0 && header[156] !== 48))
        throw new Error("invalid archive member");
      binary = tar.subarray(offset + 512, offset + 512 + size);
    }
    offset += 512 + Math.ceil(size / 512) * 512;
  }
  if (!binary) throw new Error("missing archive member");
  return binary;
};

const checkVersion = async (name: string, tool: Tool, binary: string, io: IO) => {
  const output = await io.execute([binary, ...tool.versionArgs]);
  // Complete version token equality permits upstream build metadata while
  // rejecting e.g. 1.2.30 for the 1.2.3 pin.
  if (
    !output
      .split(/\r?\n/)
      .some(
        (line) =>
          line.trim() === tool.versionOutput || line.trim().startsWith(`${tool.versionOutput} `),
      )
  ) {
    throw new Error(`${name}: version mismatch`);
  }
};

export const selfTest = async () => {
  const root = await mkdtemp(join(tmpdir(), "flux-tools-test-"));
  const bytes = new TextEncoder().encode("verified fixture binary");
  const tool: Tool = {
    version: "1.2.3",
    versionArgs: ["--version"],
    versionOutput: "fixture 1.2.3",
    url: "https://github.com/fixture/tool/releases/download/v1.2.3/tool",
    sha256: digest(bytes),
    binarySha256: digest(bytes),
  };
  let executions = 0;
  const io: IO = {
    download: async () => bytes,
    execute: async () => {
      executions++;
      return "fixture 1.2.3\n";
    },
  };
  try {
    const binary = await installTool("fixture", tool, root, io);
    assert.equal(digest(await readFile(binary)), tool.binarySha256);
    assert.equal(executions, 1);
    await writeFile(binary, "corrupted installed binary");
    executions = 0;
    await assert.rejects(installTool("fixture", tool, root, io), /installed checksum mismatch/);
    assert.equal(executions, 0, "Corrupt installed binary must never execute");
    await rm(binary);
    executions = 0;
    await assert.rejects(
      installTool("fixture", tool, root, {
        ...io,
        download: async () => new TextEncoder().encode("corrupted download"),
      }),
      /checksum mismatch/,
    );
    assert.equal(executions, 0, "Corrupt download must never execute");
    await assert.rejects(
      installTool("fixture", { ...tool, binarySha256: "0".repeat(64) }, root, io),
      /binary checksum mismatch/,
    );
    assert.equal(executions, 0, "Binary digest mismatch must never execute");
    await assert.rejects(
      installTool("fixture", tool, root, {
        ...io,
        execute: async () => "fixture 1.2.30\n",
      }),
      /version mismatch/,
    );
    await assert.rejects(readFile(binary), "Mismatched version must never publish");
    await assert.rejects(
      installTool("fixture", tool, root, {
        ...io,
        download: async () => {
          throw new Error("network failure");
        },
      }),
      /download failed/,
    );
    const readonly = join(root, "readonly-cache");
    await mkdir(readonly);
    await writeFile(join(readonly, "module"), "verified source");
    await chmod(readonly, 0o555);
    await removeStage(readonly);
    await assert.rejects(lstat(readonly), "Read-only module cache must be removed");
    assert.equal(
      await execute([
        process.execPath,
        "-e",
        'console.log("json"); console.error("download progress");',
      ]),
      "json\n",
      "Diagnostic stderr must never corrupt machine-readable stdout",
    );
    await assert.rejects(
      execute([process.execPath, "-e", "setInterval(() => {}, 1000)"], root, process.env, 30),
      /tool command failed/,
      "Subprocess timeout must terminate and fail closed",
    );
    console.log("Tool installer self-test passed");
  } finally {
    await rm(root, { recursive: true, force: true });
  }
};

const installTool = async (
  name: string,
  tool: Tool,
  root: string,
  io: IO,
  build?: Build,
): Promise<string> => {
  const destination = join(root, name);
  await mkdir(root, { recursive: true });
  let exists = false;
  try {
    exists = (await lstat(destination)).isFile();
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw error;
  }
  if (exists) {
    if (digest(await readFile(destination)) !== tool.binarySha256)
      throw new Error(`${name}: installed checksum mismatch`);
    await checkVersion(name, tool, destination, io);
    return destination;
  }
  const stage = await mkdtemp(join(root, ".install-"));
  try {
    let payload: Uint8Array;
    try {
      payload = await io.download(tool.url);
    } catch {
      throw new Error(`${name}: download failed`);
    }
    if (digest(payload) !== tool.sha256) throw new Error(`${name}: checksum mismatch`);
    const binary = join(stage, name);
    if (build) await buildGovulncheck(build, tool, payload, binary, stage, io);
    else
      await writeFile(
        binary,
        tool.archivePath ? archiveBinary(payload, tool.archivePath) : payload,
      );
    if (digest(await readFile(binary)) !== tool.binarySha256)
      throw new Error(`${name}: binary checksum mismatch`);
    await chmod(binary, 0o755);
    await checkVersion(name, tool, binary, io);
    await rename(binary, destination);
    return destination;
  } finally {
    await removeStage(stage);
  }
};

const buildGovulncheck = async (
  build: Build,
  tool: Tool,
  payload: Uint8Array,
  binary: string,
  stage: string,
  io: IO,
) => {
  const env: NodeJS.ProcessEnv = {
    ...process.env,
    GOTOOLCHAIN: build.toolchain,
    CGO_ENABLED: "0",
    GOAMD64: "v1",
    GOARM64: "v8.0",
    GOOS: process.platform,
    GOARCH: process.arch === "x64" ? "amd64" : process.arch,
    GOFLAGS: "",
    GOWORK: "off",
    GOEXPERIMENT: "",
    GOPROXY: "https://proxy.golang.org",
    GOSUMDB: "sum.golang.org",
    GOPRIVATE: "",
    GONOSUMDB: "",
    GONOPROXY: "",
    GOTELEMETRY: "off",
    GOMODCACHE: join(stage, "modules"),
    GOCACHE: join(stage, "cache"),
  };
  if (
    !(await io.execute(["go", "version"], stage, env)).startsWith(`go version ${build.toolchain} `)
  ) {
    throw new Error("govulncheck: compiler version mismatch");
  }
  await writeFile(join(stage, "go.mod"), build.goMod);
  await writeFile(join(stage, "go.sum"), build.goSum);
  const metadata = JSON.parse(
    await io.execute(
      ["go", "mod", "download", "-json", `${build.module}@${tool.version}`],
      stage,
      env,
    ),
  );
  if (
    metadata.Sum !== build.sum ||
    metadata.GoModSum !== build.goModSum ||
    digest(await readFile(metadata.Zip)) !== digest(payload)
  ) {
    throw new Error("govulncheck: source checksum mismatch");
  }
  // Go verifies every transitive download against the exact locked go.sum;
  // no tool source or module is executed before source integrity is established.
  await io.execute(
    [
      "go",
      "build",
      "-mod=readonly",
      "-trimpath",
      "-buildvcs=false",
      "-ldflags=-buildid=",
      "-o",
      binary,
      `${build.module}/cmd/govulncheck`,
    ],
    stage,
    env,
  );
  await io.execute(["go", "mod", "verify"], stage, env);
};

export const installTools = async (
  root = defaultInstallRoot,
): Promise<Record<(typeof names)[number], string>> => {
  const manifest = JSON.parse(await readFile(join(repositoryRoot, "tools.lock.json"), "utf8"));
  const installed = {} as Record<(typeof names)[number], string>;
  for (const name of names) {
    const release: Release = manifest.qualityTools?.[name];
    const pin = release?.platforms?.[`${process.platform}-${process.arch}`];
    if (!release || !pin) throw new Error(`${name}: unsupported platform or missing pin`);
    if (
      !/^[0-9a-f]{64}$/.test(pin.sha256) ||
      !/^[0-9a-f]{64}$/.test(pin.binarySha256) ||
      !/^v?\d+\.\d+\.\d+$/.test(release.version) ||
      !release.versionOutput ||
      !release.versionArgs?.length
    ) {
      throw new Error(`${name}: invalid pin`);
    }
    const official = new URL(pin.url);
    const allowed =
      name === "govulncheck"
        ? "https://proxy.golang.org/golang.org/x/vuln/@v/"
        : {
            biome: "https://github.com/biomejs/biome/releases/download/",
            "golangci-lint": "https://github.com/golangci/golangci-lint/releases/download/",
            staticcheck: "https://github.com/dominikh/go-tools/releases/download/",
            gitleaks: "https://github.com/gitleaks/gitleaks/releases/download/",
          }[name];
    if (
      !official.href.startsWith(allowed) ||
      official.username ||
      official.password ||
      official.search ||
      official.hash ||
      /(?:^|\/)latest(?:\/|$)/.test(official.pathname)
    )
      throw new Error(`${name}: unofficial or floating URL`);
    installed[name] = await installTool(
      name,
      { ...release, ...pin },
      root,
      { download, execute },
      release.build,
    );
    console.log(`Verified ${name} ${release.version}`);
  }
  return installed;
};

const verifyConfig = async (installed: Record<(typeof names)[number], string>) => {
  // This validates configuration only. Repository findings remain the quality
  // runner's responsibility rather than weakening rules to silence them.
  await execute([installed.biome, "format", "biome.json"]);
  await execute([
    installed["golangci-lint"],
    "config",
    "verify",
    "--config",
    "apps/backend/.golangci.yml",
  ]);
  await execute([installed["golangci-lint"], "linters", "--config", "apps/backend/.golangci.yml"]);
  console.log("Quality tool configuration verified");
};

if (process.argv[1] && pathToFileURL(resolve(process.argv[1])).href === import.meta.url) {
  try {
    const args = process.argv.slice(2);
    let root = defaultInstallRoot;
    let mode = "--verify";
    for (let index = 0; index < args.length; index++) {
      const argument = args[index] ?? "";
      const next = args[index + 1];
      if (argument === "--install-root" && next && !next.startsWith("--")) {
        root = resolve(next);
        index++;
      } else if (["--verify", "--verify-config", "--self-test"].includes(argument)) mode = argument;
      else
        throw new Error(
          "Usage: bun scripts/install-tools.ts [--verify|--verify-config|--self-test] [--install-root PATH]",
        );
    }
    if (mode === "--self-test") await selfTest();
    else {
      const installed = await installTools(root);
      if (mode === "--verify-config") await verifyConfig(installed);
    }
  } catch (error) {
    const message =
      error instanceof Error && /^(Usage:|[a-z-]+:)/.test(error.message)
        ? error.message
        : "Tool installation failed";
    console.error(message);
    process.exitCode = 1;
  }
}
