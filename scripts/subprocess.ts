import { spawn } from "node:child_process";

// Both supported platforms (Linux/macOS) provide POSIX process groups. Raw
// diagnostics stay private; cancellation also bounds settlement if a pipe holder
// escapes the group or the runtime never emits close after termination.
export const captureBounded = (
  command: string[],
  options: { cwd: string; env: NodeJS.ProcessEnv; timeoutMs: number; maxBytes: number },
): Promise<{ code: number; stdout: string; stderr: string }> =>
  new Promise((accept, reject) => {
    if (!command[0] || !["linux", "darwin"].includes(process.platform)) {
      reject(new Error("subprocess failed"));
      return;
    }
    const child = spawn(command[0], command.slice(1), {
      cwd: options.cwd,
      env: options.env,
      detached: true,
      stdio: ["ignore", "pipe", "pipe"],
    });
    const chunks: Buffer[] = [];
    let bytes = 0;
    let failed = false;
    let settled = false;
    let drain: ReturnType<typeof setTimeout> | undefined;
    const finish = (code: number | null) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      clearTimeout(drain);
      if (failed || code === null) reject(new Error("subprocess failed"));
      else accept({ code, stdout: Buffer.concat(chunks).toString("utf8"), stderr: "" });
    };
    const stop = () => {
      if (failed || settled) return;
      failed = true;
      if (child.pid) {
        try {
          process.kill(-child.pid, "SIGKILL");
        } catch {
          child.kill("SIGKILL");
        }
      }
      drain = setTimeout(() => {
        child.stdout.destroy();
        child.stderr.destroy();
        child.unref();
        finish(null);
      }, 100);
    };
    const timer = setTimeout(stop, options.timeoutMs);
    const collect = (chunk: Buffer, keep: boolean) => {
      if (failed || settled) return;
      bytes += chunk.length;
      if (bytes > options.maxBytes) stop();
      else if (keep) chunks.push(chunk);
    };
    child.stdout.on("data", (chunk: Buffer) => collect(chunk, true));
    child.stderr.on("data", (chunk: Buffer) => collect(chunk, false));
    child.once("error", stop);
    child.once("close", finish);
  });
