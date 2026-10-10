import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import { render } from "@react-email/components";
import WelcomeEmail from "./welcome.js";

test("email and OpenAPI workspaces expose every quality command", async () => {
  for (const path of ["../../package.json", "../../../openapi/package.json"]) {
    const manifest = JSON.parse(await readFile(new URL(path, import.meta.url), "utf8"));
    for (const command of ["format:check", "lint", "typecheck", "test", "build"]) {
      assert.equal(
        typeof manifest.scripts[command],
        "string",
        `${manifest.name}: missing ${command}`,
      );
      assert.ok(manifest.scripts[command].trim());
    }
  }
});

test("welcome renders useful content and escapes untrusted names", async () => {
  const html = await render(<WelcomeEmail userFirstName={'<script>alert("x")</script>&'} />);
  assert.match(html, /Welcome to Flux!/);
  assert.match(html, /Thank you for joining!/);
  assert.match(html, /&lt;script&gt;alert\(&quot;x&quot;\)&lt;\/script&gt;&amp;/);
  assert.doesNotMatch(html, /<script>/);
  assert.match(html, /href="\/dashboard"/);
});

test("export rendering is deterministic and preserves the Go substitution token", async () => {
  const first = await render(<WelcomeEmail userFirstName="{{.UserFirstName}}" />, { pretty: true });
  const second = await render(<WelcomeEmail userFirstName="{{.UserFirstName}}" />, {
    pretty: true,
  });
  assert.equal(first, second);
  assert.match(first, /\{\{\.UserFirstName\}\}/);
  assert.doesNotMatch(first, /Hi John,/);
});

// Two bounded CLI exports need their own finite budget on slower CI runners.
test("locked CLI exports repeatable bytes matching the embedded template", {
  timeout: 130000,
}, async () => {
  const directory = await mkdtemp(join(tmpdir(), "flux-email-test-"));
  const cwd = fileURLToPath(new URL("../../", import.meta.url));
  try {
    for (const output of ["first", "second"]) {
      const result = spawnSync(
        "node",
        [
          "node_modules/react-email/dist/cli/index.mjs",
          "export",
          "--pretty",
          "--dir",
          "src/templates",
          "--outDir",
          join(directory, output),
        ],
        { cwd, timeout: 60000, stdio: "pipe" },
      );
      assert.equal(result.status, 0, "Locked email export failed");
    }
    const first = await readFile(join(directory, "first/welcome.html"), "utf8");
    assert.equal(first, await readFile(join(directory, "second/welcome.html"), "utf8"));
    assert.equal(
      first,
      await readFile(
        new URL("../../../../apps/backend/templates/emails/welcome.html", import.meta.url),
        "utf8",
      ),
    );
    assert.match(first, /\{\{\.UserFirstName\}\}/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
