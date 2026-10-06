import { cp, mkdir, mkdtemp, readFile, readdir, rename, rm, symlink, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { tmpdir } from "node:os";
import { fileURLToPath, pathToFileURL } from "node:url";

export const artifactManifest = [
  "packages/openapi/openapi.json",
  "apps/backend/static/openapi.json",
  "apps/backend/internal/transport/health.gen.go",
  "apps/backend/templates/emails/welcome.html",
] as const;
const repositoryRoot = fileURLToPath(new URL("../", import.meta.url));

// Generator output may contain private template/schema content. Only this stable
// label crosses the CLI diagnostic boundary; subprocess output stays captured.
const run = async (cmd: string[], cwd: string, label: string): Promise<string> => {
  try {
    const child = Bun.spawn(cmd, {
      cwd, env: { ...process.env, TMPDIR: cwd, NODE_DISABLE_COMPILE_CACHE: "1" },
      stdin: "ignore", stdout: "pipe", stderr: "ignore",
    });
    const [stdout, status] = await Promise.all([new Response(child.stdout).text(), child.exited]);
    if (status !== 0) throw new Error();
    return stdout;
  } catch {
    throw new Error(`Generation failed: ${label}`);
  }
};

const copy = async (root: string, stage: string, path: string) => {
  await mkdir(dirname(join(stage, path)), { recursive: true });
  await cp(join(root, path), join(stage, path), { recursive: true });
};

const stagePackage = async (root: string, stage: string, name: string) => {
  const path = `packages/${name}`;
  for (const file of ["src", "package.json", "tsconfig.json"]) await copy(root, stage, `${path}/${file}`);
  const dependencies = join(root, path, "node_modules");
  const target = join(stage, path, "node_modules");
  await mkdir(target);
  for (const entry of await readdir(dependencies)) {
    // Redirect the one workspace dependency to the isolated authored build.
    if (entry === "@flux") continue;
    await symlink(join(dependencies, entry), join(target, entry));
  }
  if (name === "openapi") {
    await mkdir(join(target, "@flux"));
    await symlink(join(stage, "packages/zod"), join(target, "@flux/zod"));
  }
};

const generateArtifacts = async (stage: string, root: string): Promise<void> => {
  await symlink(join(root, "node_modules"), join(stage, "node_modules"));
  for (const name of ["zod", "openapi", "emails"]) await stagePackage(root, stage, name);
  for (const path of ["apps/backend/go.mod", "apps/backend/go.sum", "apps/backend/internal/transport/oapi-codegen.yaml"]) {
    await copy(root, stage, path);
  }
  for (const path of artifactManifest) await mkdir(dirname(join(stage, path)), { recursive: true });

  await run(["bun", "run", "build"], join(stage, "packages/zod"), "packages/zod");
  await run(["bun", "run", "build"], join(stage, "packages/openapi"), "packages/openapi");
  await run(["node", join(stage, "packages/openapi/dist/gen.js"), ...artifactManifest.slice(0, 2).map((path) => join(stage, path))],
    stage, "packages/openapi/src/gen.ts");

  const lock = JSON.parse(await readFile(join(root, "tools.lock.json"), "utf8"))["oapi-codegen"];
  if (lock.module !== "github.com/oapi-codegen/oapi-codegen/v2" || lock.version !== "v2.8.0") {
    throw new Error("Generation failed: tools.lock.json");
  }
  const backend = join(stage, "apps/backend");
  const module = `${lock.module}@${lock.version}`;
  const downloaded = JSON.parse(await run(["go", "mod", "download", "-json", module], backend, "tools.lock.json"));
  if (downloaded.Sum !== lock.sum || downloaded.GoModSum !== lock.goModSum) throw new Error("Generation failed: tools.lock.json");
  const configPath = "apps/backend/internal/transport/oapi-codegen.yaml";
  const config = await readFile(join(stage, configPath), "utf8");
  if (!/^output: .+$/m.test(config)) throw new Error(`Generation failed: ${configPath}`);
  // The YAML output overrides CLI -o. Rewrite only its destination in the copy.
  await writeFile(join(stage, configPath), config.replace(/^output: .+$/m, `output: ${join(stage, artifactManifest[2])}`));
  await run(["go", "run", `${lock.module}/cmd/oapi-codegen@${lock.version}`, "-config", join(stage, configPath), join(stage, artifactManifest[0])],
    backend, configPath);

  const emailPackage = join(root, "packages/emails/node_modules/react-email");
  const email = JSON.parse(await readFile(join(emailPackage, "package.json"), "utf8"));
  if (email.version !== "6.3.3") throw new Error("Generation failed: packages/emails/package.json");
  // Use the already installed, locked CLI directly; never auto-download a tool.
  await run(["node", join(emailPackage, email.bin.email), "export", "--pretty", "--dir", "./src/templates", "--outDir", join(stage, "apps/backend/templates/emails")],
    join(stage, "packages/emails"), "packages/emails/src/templates");
};

type Options = {
  root?: string;
  check?: boolean;
  temporaryParent?: string;
  generateArtifacts?: (stage: string, root: string) => Promise<void>;
};

export const generate = async (options: Options = {}): Promise<void> => {
  const root = options.root ?? repositoryRoot;
  const stage = await mkdtemp(join(options.temporaryParent ?? tmpdir(), "flux-generate-"));
  try {
    try {
      await (options.generateArtifacts ?? generateArtifacts)(stage, root);
    } catch (error) {
      if (error instanceof Error && error.message.startsWith("Generation failed: ")) throw error;
      throw new Error("Generation failed: scripts/generate.ts");
    }
    // Read all outputs before comparison or publication; missing generator output
    // must fail even in write mode, before touching checked artifacts.
    const generated = await Promise.all(artifactManifest.map(async (path) => {
      try {
        return await readFile(join(stage, path));
      } catch {
        throw new Error(`Generation failed: ${path}`);
      }
    }));
    if (options.check) {
      const stale: string[] = [];
      for (const [index, path] of artifactManifest.entries()) {
        let current: Buffer | undefined;
        try {
          current = await readFile(join(root, path));
        } catch {
          // Absent or unreadable checked files are both drift, never repaired.
        }
        if (!current?.equals(generated[index]!)) stale.push(path);
      }
      if (stale.length) throw new Error(`Stale generated artifacts:\n${stale.join("\n")}`);
    } else {
      for (const [index, path] of artifactManifest.entries()) {
        const destination = join(root, path);
        await mkdir(dirname(destination), { recursive: true });
        // Atomic replacement per file, staged on the destination filesystem.
        const publication = await mkdtemp(join(dirname(destination), ".flux-generate-"));
        try {
          await writeFile(join(publication, "artifact"), generated[index]!);
          await rename(join(publication, "artifact"), destination);
        } finally {
          await rm(publication, { recursive: true, force: true });
        }
      }
    }
  } finally {
    await rm(stage, { recursive: true, force: true });
  }
};

if (process.argv[1] && pathToFileURL(resolve(process.argv[1])).href === import.meta.url) {
  try {
    if (process.argv.slice(2).some((argument) => argument !== "--check")) throw new Error("Usage: bun scripts/generate.ts [--check]");
    await generate({ check: process.argv.includes("--check") });
  } catch (error) {
    const message = error instanceof Error && /^(Generation failed: |Stale generated artifacts:|Usage: )/.test(error.message)
      ? error.message : "Generation failed: scripts/generate.ts";
    console.error(message);
    process.exitCode = 1;
  }
}
