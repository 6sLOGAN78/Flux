import { mkdtemp, rename, rm, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { OpenAPI } from "./index.js";

const normalize = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(normalize);
  if (value === null || typeof value !== "object") return value;
  const object = value as Record<string, unknown>;
  const normalized = Object.fromEntries(
    Object.keys(object)
      .sort()
      .map((key) => [key, normalize(object[key])]),
  );
  // Preserve the existing ts-rest custom file representation as OpenAPI binary.
  if (
    JSON.stringify(normalized) ===
    JSON.stringify({
      properties: { type: { enum: ["file"], type: "string" } },
      required: ["type"],
      type: "object",
    })
  ) {
    return { format: "binary", type: "string" };
  }
  return normalized;
};

export const serializeOpenAPI = (document: unknown = OpenAPI): string => {
  return `${JSON.stringify(normalize(document), null, 2)}\n`;
};

const defaultOutputs = [
  fileURLToPath(new URL("../openapi.json", import.meta.url)),
  fileURLToPath(new URL("../../../apps/backend/static/openapi.json", import.meta.url)),
];

type WriteArtifact = (path: string, bytes: string) => Promise<void>;

const writeArtifact: WriteArtifact = async (path, bytes) => {
  await writeFile(path, bytes, { encoding: "utf8", flag: "wx" });
};

export const generateOpenAPI = async (
  outputs: string[] = defaultOutputs,
  write: WriteArtifact = writeArtifact,
): Promise<void> => {
  const destinations = outputs.map((output) => resolve(output));
  if (destinations.length === 0 || new Set(destinations).size !== destinations.length) {
    throw new Error("OpenAPI outputs must be nonempty and distinct");
  }
  const bytes = serializeOpenAPI();
  const staged: { directory: string; path: string; destination: string }[] = [];
  const failures: unknown[] = [];
  try {
    // Complete every staging write before publishing any destination. Each stage
    // is on the destination filesystem so an individual rename is atomic.
    for (const destination of destinations) {
      const directory = await mkdtemp(join(dirname(destination), ".flux-openapi-"));
      const path = join(directory, "openapi.json");
      staged.push({ directory, path, destination });
      await write(path, bytes);
    }
    for (const { path, destination } of staged) await rename(path, destination);
  } catch (error) {
    failures.push(error);
  }
  const cleanup = await Promise.allSettled(
    staged.map(({ directory }) => rm(directory, { recursive: true, force: true })),
  );
  for (const result of cleanup) {
    if (result.status === "rejected") failures.push(result.reason);
  }
  if (failures.length === 1) throw failures[0];
  if (failures.length > 1) throw new AggregateError(failures, "OpenAPI generation failed");
};

if (process.argv[1] && pathToFileURL(resolve(process.argv[1])).href === import.meta.url) {
  try {
    await generateOpenAPI(process.argv.length > 2 ? process.argv.slice(2) : undefined);
  } catch (error) {
    console.error("OpenAPI generation failed:", error);
    process.exitCode = 1;
  }
}
