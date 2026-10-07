import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

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
  execute: (cmd: string[], cwd?: string) => Promise<string>;
};
const digest = (bytes: Uint8Array) => createHash("sha256").update(bytes).digest("hex");

export const selfTest = async () => {
  const root = await mkdtemp(join(tmpdir(), "flux-tools-test-"));
  const bytes = new TextEncoder().encode("verified fixture binary");
  const tool: Tool = {
    version: "1.2.3", versionArgs: ["--version"], versionOutput: "fixture 1.2.3",
    url: "https://github.com/fixture/tool/releases/download/v1.2.3/tool",
    sha256: digest(bytes), binarySha256: digest(bytes),
  };
  let executions = 0;
  const io: IO = {
    download: async () => bytes,
    execute: async () => { executions++; return "fixture 1.2.3\n"; },
  };
  try {
    const binary = await installTool("fixture", tool, root, io);
    assert.equal(digest(await readFile(binary)), tool.binarySha256);
    assert.equal(executions, 1);
    await rm(binary);
    executions = 0;
    await assert.rejects(installTool("fixture", tool, root, {
      ...io, download: async () => new TextEncoder().encode("corrupted download"),
    }), /checksum mismatch/);
    assert.equal(executions, 0, "Corrupt download must never execute");
    await assert.rejects(installTool("fixture", tool, root, {
      ...io, execute: async () => "fixture 1.2.30\n",
    }), /version mismatch/);
    await assert.rejects(readFile(binary), "Mismatched version must never publish");
    await assert.rejects(installTool("fixture", tool, root, {
      ...io, download: async () => { throw new Error("network failure"); },
    }), /download failed/);
    console.log("Tool installer self-test passed");
  } finally {
    await rm(root, { recursive: true, force: true });
  }
};

const installTool = async (_name: string, _tool: Tool, _root: string, _io: IO): Promise<string> => {
  throw new Error("Installation not implemented");
};

if (import.meta.main) {
  try {
    if (process.argv.includes("--self-test")) await selfTest();
    else throw new Error("Installation not implemented");
  } catch (error) {
    console.error(error instanceof Error ? error.message : "Tool installation failed");
    process.exitCode = 1;
  }
}
