import assert from "node:assert/strict";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { execute } from "./install-tools.ts";
import { capture } from "./scan.ts";

const alive = async (pid: number) => {
  try {
    process.kill(pid, 0);
    // A killed process awaiting reaping cannot execute or hold pipes.
    if (process.platform === "linux")
      return !/\) Z /.test(await readFile(`/proc/${pid}/stat`, "utf8"));
    return true;
  } catch {
    return false;
  }
};

for (const wrapper of ["scanner", "installer"]) {
  for (const failure of ["timeout", "overflow"]) {
    test(`${wrapper} bounds ${failure} and kills nested stdout/stderr holders`, async () => {
      const root = await mkdtemp(join(tmpdir(), "flux-child-proof-"));
      const pidFile = join(root, "child.pid");
      let pid: number | undefined;
      try {
        const child = `require("node:fs").writeFileSync(${JSON.stringify(pidFile)}, String(process.pid));
          ${failure === "overflow" ? 'setTimeout(()=>{process.stdout.write("PRIVATE_MARKER".repeat(200000)); process.stderr.write("PRIVATE_MARKER".repeat(200000));},50);' : ""}
          setTimeout(()=>{},3000);`;
        const parent = `require("node:child_process").spawn(process.execPath,["-e",${JSON.stringify(child)}],{stdio:["ignore","inherit","inherit"]});setTimeout(()=>{},3000);`;
        const command = [process.execPath, "-e", parent];
        const started = performance.now();
        await assert.rejects(
          wrapper === "scanner"
            ? capture(
                { tool: command[0] as string, args: command.slice(1), cwd: root },
                failure === "timeout" ? 500 : 2000,
                2048,
              )
            : execute(command, root, process.env, failure === "timeout" ? 500 : 2000),
          (error: Error) => {
            assert.equal(
              error.message,
              wrapper === "scanner" ? "scanner execution failed" : "tool command failed",
            );
            return true;
          },
        );
        assert.ok(performance.now() - started < 1500, "wrapper exceeded bounded cancellation");
        pid = Number(await readFile(pidFile, "utf8"));
        assert.ok(Number.isSafeInteger(pid) && pid > 0);
        for (let attempt = 0; attempt < 50 && (await alive(pid)); attempt++)
          await new Promise((resolve) => setTimeout(resolve, 10));
        assert.equal(await alive(pid), false, "nested pipe holder survived cancellation");
      } finally {
        if (!pid) pid = Number(await readFile(pidFile, "utf8").catch(() => "0"));
        if (pid > 0 && (await alive(pid))) process.kill(pid, "SIGKILL");
        await rm(root, { recursive: true, force: true });
      }
    });
  }
}

test("cancellation settles after bounded drain even when a pipe holder escapes its group", async () => {
  const root = await mkdtemp(join(tmpdir(), "flux-drain-proof-"));
  const pidFile = join(root, "child.pid");
  try {
    const child = `require("node:fs").writeFileSync(${JSON.stringify(pidFile)},String(process.pid));setTimeout(()=>{},3000);`;
    const parent = `require("node:child_process").spawn(process.execPath,["-e",${JSON.stringify(child)}],{detached:true,stdio:["ignore","inherit","inherit"]});setTimeout(()=>{},3000);`;
    const started = performance.now();
    await assert.rejects(capture({ tool: process.execPath, args: ["-e", parent], cwd: root }, 500));
    assert.ok(performance.now() - started < 1000, "inherited pipes prevented settlement");
  } finally {
    const pid = Number(await readFile(pidFile, "utf8").catch(() => "0"));
    if (pid > 0 && (await alive(pid))) process.kill(pid, "SIGKILL");
    await rm(root, { recursive: true, force: true });
  }
});
