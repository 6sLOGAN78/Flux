import { spawn } from "node:child_process";
import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import { join, relative, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { installTools } from "./install-tools.ts";

const repositoryRoot = fileURLToPath(new URL("../", import.meta.url));
export type Command = { tool: string; args: string[]; cwd: string };
export type Result = { code: number; stdout: string; stderr: string };
type Runner = (command: Command) => Promise<Result>;
type Mode = "all" | "dependencies" | "secrets";

// Reports stay in bounded memory. Stderr is drained and counted, never retained
// or forwarded. Failures expose a closed diagnostic, not a provider exception.
export const capture = (
  command: Command,
  timeoutMs = 240_000,
  maxBytes = 8_388_608,
): Promise<Result> =>
  new Promise((accept, reject) => {
    const child = spawn(command.tool, command.args, {
      cwd: command.cwd,
      env: {
        ...process.env,
        GOFLAGS: "-mod=readonly",
        GOWORK: "off",
        GOTOOLCHAIN: "go1.26.8",
        NO_COLOR: "1",
      },
      stdio: ["ignore", "pipe", "pipe"],
    });
    const chunks: Buffer[] = [];
    let bytes = 0;
    let failed = false;
    const stop = () => {
      failed = true;
      child.kill("SIGKILL");
    };
    const timer = setTimeout(stop, timeoutMs);
    const collect = (chunk: Buffer, keep: boolean) => {
      bytes += chunk.length;
      if (bytes > maxBytes) stop();
      else if (keep) chunks.push(chunk);
    };
    child.stdout.on("data", (chunk: Buffer) => collect(chunk, true));
    child.stderr.on("data", (chunk: Buffer) => collect(chunk, false));
    child.on("error", () => {
      clearTimeout(timer);
      reject(new Error("scanner execution failed"));
    });
    child.on("close", (code) => {
      clearTimeout(timer);
      if (failed || code === null) reject(new Error("scanner execution failed"));
      else accept({ code, stdout: Buffer.concat(chunks).toString("utf8"), stderr: "" });
    });
  });

const opaque = (value: unknown) =>
  `redacted-${createHash("sha256").update(String(value)).digest("hex").slice(0, 12)}`;
const advisory = (value: unknown) =>
  typeof value === "string" &&
  /^(?:GO-\d{4}-\d{4,}|CVE-\d{4}-\d{4,}|GHSA-[23456789cfghjmpqrvwx]{4}-[23456789cfghjmpqrvwx]{4}-[23456789cfghjmpqrvwx]{4})$/.test(
    value,
  )
    ? value
    : opaque(value);
const rule = (value: unknown) =>
  typeof value === "string" && /^[a-z][a-z0-9-]{1,39}$/.test(value) ? value : opaque(value);
const packageName = (value: unknown) =>
  typeof value === "string" && /^[a-z@][a-z0-9@/._-]{1,99}$/.test(value) ? value : opaque(value);
const version = (value: unknown) =>
  typeof value === "string" &&
  /^v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/.test(value) &&
  value.length < 50
    ? value
    : "not reported";
const location = (value: unknown) =>
  Number.isSafeInteger(value) && Number(value) > 0 ? String(value) : "unknown";
const safePath = (value: unknown, root: string) => {
  if (typeof value !== "string") return opaque(value);
  const path = value.startsWith("/") ? relative(root, value) : value;
  return path.length <= 160 &&
    !path.split("/").some((part) => part === ".." || part.length > 40) &&
    /^(?:[a-zA-Z0-9_.-]+\/)*[a-zA-Z0-9_.-]+$/.test(path)
    ? path
    : opaque(value);
};
const object = (value: unknown): Record<string, unknown> => {
  if (!value || typeof value !== "object" || Array.isArray(value))
    throw new Error("invalid scanner report");
  return value as Record<string, unknown>;
};

// govulncheck emits a sequence of pretty-printed JSON objects, not JSONL.
export const decodeGo = (text: string): Record<string, unknown>[] => {
  const events: Record<string, unknown>[] = [];
  let start = -1;
  let depth = 0;
  let quoted = false;
  let escaped = false;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (start === -1) {
      if (/\s/.test(c ?? "")) continue;
      if (c !== "{") throw new Error("invalid scanner report");
      start = i;
    }
    if (quoted) {
      if (escaped) escaped = false;
      else if (c === "\\") escaped = true;
      else if (c === '"') quoted = false;
    } else if (c === '"') quoted = true;
    else if (c === "{" || c === "[") depth++;
    else if (c === "}" || c === "]") depth--;
    if (depth === 0 && !quoted) {
      events.push(object(JSON.parse(text.slice(start, i + 1))));
      start = -1;
    }
  }
  if (start !== -1 || !events.length) throw new Error("invalid scanner report");
  return events;
};

export const runScans = async (
  options: {
    root?: string;
    mode?: Mode;
    ci?: boolean;
    runner?: Runner;
    verify?: () => Promise<void>;
    emit?: (line: string) => void;
  } = {},
): Promise<number> => {
  const root = resolve(options.root ?? repositoryRoot);
  const mode = options.mode ?? "all";
  const ci = options.ci ?? Boolean(process.env.CI && process.env.CI !== "false");
  const run = options.runner ?? capture;
  const emit = options.emit ?? console.log;
  let failed = false;
  const error = (stage: string) => {
    failed = true;
    emit(`${stage}: scanner execution/report failed; verify tool, network and configuration`);
  };
  try {
    if (options.verify) await options.verify();
    else {
      await installTools();
      const pin = JSON.parse(
        await readFile(join(repositoryRoot, "package.json"), "utf8"),
      ).packageManager;
      const version = await run({ tool: "bun", args: ["--version"], cwd: root });
      if (version.code || pin !== `bun@${version.stdout.trim()}`)
        throw new Error("runtime mismatch");
    }
  } catch {
    error("scanner integrity");
    return 1;
  }
  const scan = async (stage: string, command: Command, parse: (result: Result) => string[]) => {
    try {
      const findings = parse(await run(command));
      if (findings.length) {
        failed = true;
        for (const finding of findings.slice(0, 100)) emit(`${stage}: ${finding}`);
        if (findings.length > 100) emit(`${stage}: additional findings omitted`);
      } else emit(`${stage}: clean`);
    } catch {
      error(stage);
    }
  };
  if (mode !== "secrets") {
    await scan(
      "Go dependencies",
      {
        tool: options.runner ? "govulncheck" : join(repositoryRoot, "tmp/tools/govulncheck"),
        args: ["-json", "-scan=module", "-C", "cmd/api"],
        cwd: join(root, "apps/backend"),
      },
      (result) => {
        if (result.code !== 0) throw new Error("scanner failed");
        const events = decodeGo(result.stdout);
        const findings: string[] = [];
        let complete = false;
        for (const event of events) {
          if (
            Object.keys(event).length !== 1 ||
            !Object.keys(event).every((key) =>
              ["config", "progress", "osv", "SBOM", "finding"].includes(key),
            )
          )
            throw new Error("invalid event");
          if (event.SBOM) {
            const sbom = object(event.SBOM);
            if (
              !Array.isArray(sbom.modules) ||
              !sbom.modules.length ||
              sbom.go_version !== "go1.26.8"
            )
              throw new Error("invalid SBOM/toolchain");
            complete = true;
          }
          if (event.finding) {
            const f = object(event.finding);
            const trace = Array.isArray(f.trace) && f.trace.length ? object(f.trace[0]) : {};
            findings.push(
              `${advisory(f.osv)} ${packageName(trace.module)}@${version(trace.version)} apps/backend/go.mod; fixed=${version(f.fixed_version)}; module-level finding, inspect Go advisory`,
            );
          }
        }
        if (!complete && !findings.length) throw new Error("incomplete report");
        return [...new Set(findings)];
      },
    );
    await scan(
      "Bun dependencies",
      { tool: "bun", args: ["audit", "--json"], cwd: root },
      (result) => {
        if (result.code !== 0 && result.code !== 1) throw new Error("scanner failed");
        const report = object(JSON.parse(result.stdout));
        const findings: string[] = [];
        for (const [name, entries] of Object.entries(report)) {
          if (!Array.isArray(entries) || !entries.length) throw new Error("invalid audit report");
          for (const entry of entries) {
            const f = object(entry);
            if (typeof f.id !== "number" || typeof f.url !== "string")
              throw new Error("invalid audit finding");
            const id = /^https:\/\/github\.com\/advisories\/(GHSA-[a-z0-9-]+)$/.exec(f.url)?.[1];
            findings.push(
              `${id ? advisory(id) : `npm-advisory-${location(f.id)}`} ${packageName(name)} bun.lock; upgrade affected package per advisory`,
            );
          }
        }
        if (result.code === 1 && !findings.length) throw new Error("audit service failed");
        return [...new Set(findings)];
      },
    );
  }
  if (mode !== "dependencies") {
    const parse = (result: Result) => {
      if (result.code !== 0 && result.code !== 1) throw new Error("scanner failed");
      const report: unknown = JSON.parse(result.stdout);
      if (!Array.isArray(report) || (result.code === 1 && !report.length))
        throw new Error("invalid secret report");
      return report.map((entry) => {
        const f = object(entry);
        return `${rule(f.RuleID)} ${safePath(f.File, root)}:${location(f.StartLine)}:${location(f.StartColumn)}; remove secret and rotate credential`;
      });
    };
    const gitleaks = options.runner ? "gitleaks" : join(repositoryRoot, "tmp/tools/gitleaks");
    const flags = [
      "--config",
      join(repositoryRoot, ".gitleaks.toml"),
      "--redact=100",
      "--no-banner",
      "--no-color",
      "--ignore-gitleaks-allow",
      "--gitleaks-ignore-path",
      "/dev/null",
      "--report-format=json",
      "--report-path=-",
      "--timeout=180",
    ];
    await scan(
      "Secrets worktree",
      { tool: gitleaks, args: ["dir", ...flags, root], cwd: root },
      parse,
    );
    try {
      const git = await run({
        tool: "git",
        args: ["rev-parse", "--is-inside-work-tree"],
        cwd: root,
      });
      if (git.code !== 0 || git.stdout.trim() !== "true") {
        emit("Secrets history unavailable locally; worktree only");
        if (ci) error("CI requires full Git history");
      } else {
        const shallow = await run({
          tool: "git",
          args: ["rev-parse", "--is-shallow-repository"],
          cwd: root,
        });
        if (shallow.code || !["true", "false"].includes(shallow.stdout.trim()))
          throw new Error("history check failed");
        if (shallow.stdout.trim() === "true") {
          emit("Secrets history incomplete: shallow checkout");
          if (ci) error("CI requires full Git history");
        }
        await scan(
          "Secrets history",
          {
            tool: gitleaks,
            args: ["git", ...flags, "--log-opts=--all --full-history", root],
            cwd: root,
          },
          parse,
        );
      }
    } catch {
      error("Secrets history");
    }
  }
  return failed ? 1 : 0;
};

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  const args = process.argv.slice(2);
  if (args.length > 1 || (args[0] && !["--dependencies", "--secrets"].includes(args[0]))) {
    console.error("Usage: bun scripts/scan.ts [--dependencies|--secrets]");
    process.exitCode = 1;
  } else
    process.exitCode = await runScans({
      mode:
        args[0] === "--dependencies" ? "dependencies" : args[0] === "--secrets" ? "secrets" : "all",
    });
}
