import { createHash, randomUUID } from "node:crypto";
import { lstat, readFile, realpath, rm, writeFile } from "node:fs/promises";
import { isAbsolute, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const requiredProviderRows = [
  "email_password_signup",
  "password_login",
  "email_verification",
  "password_recovery",
  "google_oauth",
  "github_oauth",
  "named_workspace_owner",
  "invitation_onboarding",
  "last_used_workspace",
  "membership_loss_tabs",
  "links_empty_state",
  "token_refresh",
  "token_expiry",
  "session_revocation",
  "session_signout",
  "session_maximum",
  "provider_timeout",
  "provider_quota_latency",
  "production_cookie_inspection",
  "provider_key_rotation",
  "invitation_delivery",
  "invitation_key_rotation",
] as const;
export type ProviderRowID = (typeof requiredProviderRows)[number];
export const manualProviderRows = {
  email_verification: {
    reasons: ["authorized_inbox_verification"],
    steps: [
      "Open the authorized signup inbox.",
      "Enter its real verification code in the running signup flow.",
      "Confirm the verified account can enter Flux and the earlier wrong code was denied.",
    ],
  },
  password_recovery: {
    reasons: ["authorized_inbox_verification"],
    steps: [
      "Open the authorized recovery inbox.",
      "Complete the real reset flow with its code and a new private password.",
      "Verify old credentials and prior sessions are denied and the recovered account can sign in.",
    ],
  },
  google_oauth: {
    reasons: ["oauth_account_interaction", "operator_mfa"],
    steps: [
      "Open Sign in to Flux and choose Google.",
      "Complete Google authorization and any MFA with the authorized account.",
      "Confirm return to Flux with a verified provider identity.",
    ],
  },
  github_oauth: {
    reasons: ["oauth_account_interaction", "operator_mfa"],
    steps: [
      "Open Sign in to Flux and choose GitHub.",
      "Complete GitHub authorization and any MFA with the authorized account.",
      "Confirm return to Flux with a verified provider identity.",
    ],
  },
  production_cookie_inspection: {
    reasons: ["https_cookie_inspection"],
    steps: [
      "Open the approved production-equivalent HTTPS deployment.",
      "Inspect actual application JWT and provider long-lived cookies separately.",
      "Verify Secure and SameSite policy and the HttpOnly distinction; record no cookie values.",
    ],
  },
  provider_key_rotation: {
    reasons: ["provider_dashboard_key_rotation"],
    steps: [
      "Rotate signing keys in the authorized Clerk instance dashboard.",
      "Verify a new real session token is accepted by the production Go API.",
      "Verify removed-key tokens are denied after the provider rotation policy takes effect.",
    ],
  },
} as const;
export const providerEnvNames = [
  "NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY",
  "CLERK_SECRET_KEY",
  "FLUX_AUTH.SECRET_KEY",
  "FLUX_AUTH.ISSUER",
  "FLUX_AUTH.AUTHORIZED_PARTIES",
  "FLUX_API_ORIGIN",
  "FLUX_INVITATIONS.ACTIVE_KEY_ID",
  "FLUX_INVITATIONS.ENCRYPTION_KEYS",
  "FLUX_INVITATIONS.SENDER",
  "FLUX_INVITATIONS.PUBLIC_ORIGIN",
  "FLUX_INTEGRATION.RESEND_API_KEY",
  "FLUX_LINKS.MANAGED_HOST",
  "FLUX_LINKS.CURSOR_KEY",
  "FLUX_PROVIDER_TEST_EMAIL",
  "FLUX_PROVIDER_TEST_PASSWORD",
  "FLUX_PROVIDER_SIGNUP_EMAIL",
  "FLUX_PROVIDER_SIGNUP_PASSWORD",
  "FLUX_PROVIDER_INVITEE_EMAIL",
  "FLUX_PROVIDER_INVITEE_PASSWORD",
  "FLUX_PROVIDER_AUTHORIZED_RECIPIENTS",
  "FLUX_PROVIDER_PRIVATE_DIR",
] as const;
export type ProviderEnvironment = Record<string, string | undefined>;
const email = (value: string) =>
  /^[A-Za-z0-9_'+.-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/.test(value) && value.length <= 254;
const host = (value: string) =>
  value.length <= 253 &&
  /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$/.test(value) &&
  !/\.(?:local|internal|localhost)$/.test(value);
const key = (value: string) =>
  /^[A-Za-z0-9+/]{43}=$/.test(value) &&
  Buffer.from(value, "base64").length === 32 &&
  Buffer.from(value, "base64").toString("base64") === value;
const origin = (value: string, https = false) => {
  try {
    const url = new URL(value);
    return (
      url.origin === value &&
      !url.username &&
      !url.password &&
      (url.protocol === "https:" ||
        (!https && url.protocol === "http:" && ["localhost", "127.0.0.1"].includes(url.hostname)))
    );
  } catch {
    return false;
  }
};
export const validateProviderConfig = (env: ProviderEnvironment): string[] => {
  const errors: string[] = [];
  for (const name of providerEnvNames) if (!env[name]?.trim()) errors.push(name);
  const value = (name: string) => env[name] ?? "";
  const reject = (name: string, valid: boolean) => {
    if (!valid && !errors.includes(name)) errors.push(name);
  };
  const encoded =
    value("NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY").match(/^pk_test_([A-Za-z0-9+/]+={0,2})$/)?.[1] ?? "";
  const domain = Buffer.from(encoded, "base64").toString("utf8");
  const canonical = Buffer.from(domain).toString("base64");
  reject(
    "NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY",
    !!encoded &&
      (canonical === encoded || canonical.replace(/=+$/, "") === encoded) &&
      domain.endsWith(".clerk.accounts.dev$"),
  );
  reject(
    "FLUX_AUTH.ISSUER",
    origin(value("FLUX_AUTH.ISSUER"), true) &&
      value("FLUX_AUTH.ISSUER") === `https://${domain.slice(0, -1)}`,
  );
  reject("CLERK_SECRET_KEY", /^sk_test_[A-Za-z0-9]+$/.test(value("CLERK_SECRET_KEY")));
  reject(
    "FLUX_AUTH.SECRET_KEY",
    value("FLUX_AUTH.SECRET_KEY") === value("CLERK_SECRET_KEY") && !!value("CLERK_SECRET_KEY"),
  );
  reject(
    "FLUX_AUTH.AUTHORIZED_PARTIES",
    value("FLUX_AUTH.AUTHORIZED_PARTIES") === "http://localhost:3000",
  );
  reject("FLUX_API_ORIGIN", value("FLUX_API_ORIGIN") === "http://127.0.0.1:8080");
  const publicOrigin = value("FLUX_INVITATIONS.PUBLIC_ORIGIN");
  reject(
    "FLUX_INVITATIONS.PUBLIC_ORIGIN",
    origin(publicOrigin, true) && host(publicOrigin.slice("https://".length)),
  );
  reject("FLUX_INVITATIONS.SENDER", email(value("FLUX_INVITATIONS.SENDER")));
  reject(
    "FLUX_INTEGRATION.RESEND_API_KEY",
    /^re_[A-Za-z0-9_-]{16,}$/.test(value("FLUX_INTEGRATION.RESEND_API_KEY")),
  );
  reject("FLUX_LINKS.MANAGED_HOST", host(value("FLUX_LINKS.MANAGED_HOST")));
  reject("FLUX_LINKS.CURSOR_KEY", key(value("FLUX_LINKS.CURSOR_KEY")));
  reject(
    "FLUX_LINKS.BLOCKED_HOSTS",
    !value("FLUX_LINKS.BLOCKED_HOSTS") ||
      value("FLUX_LINKS.BLOCKED_HOSTS")
        .split(",")
        .every((entry) => host(entry.trim())),
  );
  try {
    const raw = value("FLUX_INVITATIONS.ENCRYPTION_KEYS");
    const keys: unknown = JSON.parse(raw);
    const entries =
      keys && typeof keys === "object" && !Array.isArray(keys) ? Object.entries(keys) : [];
    reject(
      "FLUX_INVITATIONS.ENCRYPTION_KEYS",
      raw.length <= 16384 &&
        entries.length >= 2 &&
        entries.length <= 32 &&
        entries.every(
          ([id, encodedKey]) =>
            /^[A-Za-z0-9_-]{1,64}$/.test(id) && typeof encodedKey === "string" && key(encodedKey),
        ) &&
        new Set(entries.map(([, encodedKey]) => encodedKey)).size === entries.length,
    );
    reject(
      "FLUX_INVITATIONS.ACTIVE_KEY_ID",
      /^[A-Za-z0-9_-]{1,64}$/.test(value("FLUX_INVITATIONS.ACTIVE_KEY_ID")) &&
        entries.some(([id]) => id === value("FLUX_INVITATIONS.ACTIVE_KEY_ID")),
    );
  } catch {
    reject("FLUX_INVITATIONS.ENCRYPTION_KEYS", false);
  }
  const recipients = value("FLUX_PROVIDER_AUTHORIZED_RECIPIENTS")
    .split(",")
    .map((entry) => entry.trim().toLowerCase());
  reject(
    "FLUX_PROVIDER_AUTHORIZED_RECIPIENTS",
    recipients.length >= 3 &&
      recipients.every(email) &&
      new Set(recipients).size === recipients.length,
  );
  for (const name of ["TEST", "SIGNUP", "INVITEE"]) {
    reject(
      `FLUX_PROVIDER_${name}_EMAIL`,
      email(value(`FLUX_PROVIDER_${name}_EMAIL`)) &&
        recipients.includes(value(`FLUX_PROVIDER_${name}_EMAIL`).toLowerCase()),
    );
    reject(`FLUX_PROVIDER_${name}_PASSWORD`, value(`FLUX_PROVIDER_${name}_PASSWORD`).length >= 12);
  }
  reject(
    "FLUX_PROVIDER_SIGNUP_EMAIL",
    new Set(
      ["TEST", "SIGNUP", "INVITEE"].map((name) =>
        value(`FLUX_PROVIDER_${name}_EMAIL`).toLowerCase(),
      ),
    ).size === 3,
  );
  reject("FLUX_PROVIDER_PRIVATE_DIR", isAbsolute(value("FLUX_PROVIDER_PRIVATE_DIR")));
  return errors;
};

// Matching key prefixes are insufficient: compare the real instance's public
// signing keys using fixed BAPI and the validated development issuer only.
export const verifyProviderInstance = async (
  env: ProviderEnvironment,
  fetcher: typeof fetch = fetch,
) => {
  if (validateProviderConfig(env).length) return false;
  try {
    const backend = await fetcher("https://api.clerk.com/v1/jwks", {
      headers: { Authorization: `Bearer ${env.CLERK_SECRET_KEY}` },
      redirect: "error",
      signal: AbortSignal.timeout(10000),
    });
    const frontend = await fetcher(`${env["FLUX_AUTH.ISSUER"]}/.well-known/jwks.json`, {
      redirect: "error",
      signal: AbortSignal.timeout(10000),
    });
    if (!backend.ok || !frontend.ok) return false;
    const keys = (value: unknown) => {
      const item = value as { keys?: { kid?: string; kty?: string; n?: string; e?: string }[] };
      if (
        !Array.isArray(item.keys) ||
        !item.keys.length ||
        item.keys.some((entry) => !entry.kid || entry.kty !== "RSA" || !entry.n || !entry.e)
      )
        throw new Error();
      return item.keys.map(({ kid, kty, n, e }) => JSON.stringify([kid, kty, n, e])).sort();
    };
    return (
      JSON.stringify(keys(await backend.json())) === JSON.stringify(keys(await frontend.json()))
    );
  } catch {
    return false;
  }
};
export type ProviderExecution = {
  id: string;
  project: "provider";
  status: "passed";
  retry: 0;
  rows: { id: ProviderRowID; status: "passed" | "human_pending" }[];
};
export type ProviderProof = {
  schema: 1;
  run_id: string;
  created_at: string;
  revision: string;
  config_fingerprint: string;
  tests: ProviderExecution[];
};
export type ProviderRow = {
  id: ProviderRowID;
  status: "passed" | "human_pending";
  evidence_kind: "automated" | "human_only" | "human_verified";
  tests: string[];
  reason_code?: string;
  manual_steps?: readonly string[];
  verified_at?: string;
};
export type ProviderAcceptance = {
  schema: 1;
  run_id: string;
  proof_sha256: string;
  rows: ProviderRow[];
};
const hash = (value: string) => createHash("sha256").update(value).digest("hex");
export const configFingerprint = (env: ProviderEnvironment) =>
  hash(
    JSON.stringify(
      [
        "FLUX_AUTH.ISSUER",
        "FLUX_AUTH.AUTHORIZED_PARTIES",
        "NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY",
        "FLUX_INVITATIONS.PUBLIC_ORIGIN",
        "FLUX_LINKS.MANAGED_HOST",
      ].map((name) => env[name]),
    ),
  );
export const proofHash = (proof: ProviderProof) => hash(JSON.stringify(proof));
export const validateAcceptance = (
  record: unknown,
  automated = false,
  proof?: ProviderProof,
  fingerprint?: string,
): boolean => {
  try {
    const item = record as ProviderAcceptance;
    if (
      !item ||
      item.schema !== 1 ||
      !proof ||
      proof.schema !== 1 ||
      item.run_id !== proof.run_id ||
      item.proof_sha256 !== proofHash(proof) ||
      !/^[a-f0-9]{40}$/.test(proof.revision) ||
      !/^[a-f0-9]{64}$/.test(proof.config_fingerprint) ||
      (fingerprint && fingerprint !== proof.config_fingerprint) ||
      !/^[a-f0-9-]{36}$/.test(proof.run_id)
    )
      return false;
    const age = Date.now() - Date.parse(proof.created_at);
    if (
      !Number.isFinite(age) ||
      age < 0 ||
      age > 7 * 86400000 ||
      !Array.isArray(proof.tests) ||
      !proof.tests.length
    )
      return false;
    const testIDs = new Set(proof.tests.map((entry) => entry.id));
    const observedRows = proof.tests.flatMap((entry) => entry.rows.map((row) => row.id));
    if (
      observedRows.length !== requiredProviderRows.length ||
      new Set(observedRows).size !== requiredProviderRows.length
    )
      return false;
    if (
      testIDs.size !== proof.tests.length ||
      proof.tests.some(
        (entry) =>
          !entry.id ||
          entry.project !== "provider" ||
          entry.status !== "passed" ||
          entry.retry !== 0 ||
          !Array.isArray(entry.rows) ||
          !entry.rows.length ||
          entry.rows.some(
            (row) =>
              !requiredProviderRows.includes(row.id) ||
              !["passed", "human_pending"].includes(row.status),
          ),
      )
    )
      return false;
    if (
      !Array.isArray(item.rows) ||
      item.rows.length !== requiredProviderRows.length ||
      new Set(item.rows.map((row) => row.id)).size !== requiredProviderRows.length
    )
      return false;
    for (const row of item.rows) {
      if (
        !requiredProviderRows.includes(row.id) ||
        !Array.isArray(row.tests) ||
        !row.tests.length ||
        new Set(row.tests).size !== row.tests.length ||
        row.tests.some((id) => !testIDs.has(id))
      )
        return false;
      const observed = proof.tests.filter((entry) => row.tests.includes(entry.id));
      if (observed.some((entry) => !entry.rows.some((evidence) => evidence.id === row.id)))
        return false;
      const manual = manualProviderRows[row.id as keyof typeof manualProviderRows];
      if (row.status === "human_pending") {
        if (
          !automated ||
          !manual ||
          row.evidence_kind !== "human_only" ||
          !manual.reasons.some((reason) => reason === row.reason_code) ||
          JSON.stringify(row.manual_steps) !== JSON.stringify(manual.steps) ||
          observed.some(
            (entry) =>
              !entry.rows.some(
                (evidence) => evidence.id === row.id && evidence.status === "human_pending",
              ),
          )
        )
          return false;
      } else if (row.status === "passed") {
        if (row.evidence_kind === "automated") {
          if (
            observed.some(
              (entry) =>
                !entry.rows.some(
                  (evidence) => evidence.id === row.id && evidence.status === "passed",
                ),
            )
          )
            return false;
        } else if (row.evidence_kind === "human_verified") {
          const verifiedAge = Date.now() - Date.parse(row.verified_at ?? "");
          if (
            !manual ||
            !Number.isFinite(verifiedAge) ||
            verifiedAge < 0 ||
            verifiedAge > 7 * 86400000 ||
            JSON.stringify(row.manual_steps) !== JSON.stringify(manual.steps)
          )
            return false;
        } else return false;
      } else return false;
    }
    return true;
  } catch {
    return false;
  }
};
export const privateProviderDirectory = async (env: ProviderEnvironment, root: string) => {
  const path = resolve(env.FLUX_PROVIDER_PRIVATE_DIR ?? "");
  const metadata = await lstat(path);
  const actual = await realpath(path);
  const repository = await realpath(root);
  const inside = relative(repository, actual);
  if (
    !metadata.isDirectory() ||
    metadata.isSymbolicLink() ||
    (metadata.mode & 0o077) !== 0 ||
    (process.getuid && metadata.uid !== process.getuid()) ||
    actual !== path ||
    (!inside.startsWith("..") && !isAbsolute(inside))
  )
    throw new Error("Provider private directory must be mode-0700 outside the repository");
  return path;
};
export const readPrivateJSON = async (path: string): Promise<unknown> => {
  const metadata = await lstat(path);
  if (
    !metadata.isFile() ||
    metadata.isSymbolicLink() ||
    (metadata.mode & 0o077) !== 0 ||
    metadata.size > 16 * 1024 * 1024
  )
    throw new Error("Invalid private provider evidence file");
  return JSON.parse(await readFile(path, "utf8"));
};
export const providerConfigCLI = async (
  args: string[],
  env: ProviderEnvironment = process.env,
): Promise<number> => {
  const mode = args[0];
  if (
    args.length !== 1 ||
    !["--check-config", "--check-automated-acceptance", "--check-acceptance"].includes(mode ?? "")
  ) {
    console.error("Use --check-config, --check-automated-acceptance or --check-acceptance");
    return 1;
  }
  const errors = validateProviderConfig(env);
  if (errors.length) {
    console.error(
      `Provider prerequisites missing or invalid: ${errors.join(", ")}. Configure privately using docs/development.md; do not paste values.`,
    );
    return 1;
  }
  const root = fileURLToPath(new URL("../../../", import.meta.url));
  try {
    const directory = await privateProviderDirectory(env, root);
    if (!(await verifyProviderInstance(env)))
      throw new Error("Provider instance mismatch or unavailable");
    if (mode !== "--check-config") {
      const proof = (await readPrivateJSON(
        resolve(directory, "provider-proof.json"),
      )) as ProviderProof;
      const markdown = await readFile(
        resolve(
          root,
          ".planning/phases/02-tenant-safe-link-control-plane/02-PROVIDER-ACCEPTANCE.md",
        ),
        "utf8",
      );
      const blocks = [...markdown.matchAll(/```json\r?\n([\s\S]*?)\r?\n```/g)];
      if (
        blocks.length !== 1 ||
        !validateAcceptance(
          JSON.parse(blocks[0]?.[1] ?? ""),
          mode === "--check-automated-acceptance",
          proof,
          configFingerprint(env),
        )
      )
        throw new Error("Invalid evidence");
    }
    console.log(
      mode === "--check-config"
        ? "Provider prerequisites valid; real acceptance remains required."
        : "Provider evidence validated for the selected mode.",
    );
    return 0;
  } catch {
    console.error(
      "Provider prerequisites or acceptance evidence unavailable/invalid. Follow docs/development.md; pending evidence cannot complete the phase.",
    );
    return 1;
  }
};

export const providerRow = (
  info: import("@playwright/test").TestInfo,
  id: ProviderRowID,
  status: "passed" | "human_pending" = "passed",
) => {
  info.annotations.push({ type: "flux-provider-row", description: JSON.stringify({ id, status }) });
};
export const providerToken = (page: import("@playwright/test").Page) =>
  page.evaluate(async () => {
    const clerk = (
      window as unknown as {
        Clerk?: {
          session?: { getToken(options?: { skipCache: boolean }): Promise<string | null> };
        };
      }
    ).Clerk;
    const token = await clerk?.session?.getToken({ skipCache: true });
    if (!token) throw new Error("Real provider session unavailable");
    return token;
  });
export const validateProviderInboxCode = (
  value: unknown,
  request: string,
  since: number,
): string | null => {
  if (!value || typeof value !== "object") return null;
  const item = value as { request_id?: unknown; received_at?: unknown; code?: unknown };
  const received = typeof item.received_at === "string" ? Date.parse(item.received_at) : NaN;
  return item.request_id === request &&
    typeof item.code === "string" &&
    /^\d{6}$/.test(item.code) &&
    Number.isFinite(received) &&
    received >= since &&
    received <= Date.now()
    ? item.code
    : null;
};
const completeProviderDeviceTrust = async (
  page: import("@playwright/test").Page,
  account: "TEST" | "INVITEE",
) => {
  await page.getByRole("button", { name: /resend/i }).click({ timeout: 35000 });
  const directory = await privateProviderDirectory(
    process.env,
    fileURLToPath(new URL("../../../", import.meta.url)),
  );
  const request = randomUUID();
  const since = Date.now();
  const input = resolve(directory, `verification-request-${request}.json`);
  const output = resolve(directory, `verification-response-${request}.json`);
  await writeFile(
    input,
    JSON.stringify({
      request_id: request,
      email: process.env[`FLUX_PROVIDER_${account}_EMAIL`],
      requested_at: new Date(since).toISOString(),
    }),
    { mode: 0o600, flag: "wx" },
  );
  try {
    const deadline = since + 30000;
    while (Date.now() < deadline) {
      let code: string | null = null;
      try {
        code = validateProviderInboxCode(await readPrivateJSON(output), request, since);
      } catch {
        /* The authorized inbox integration has not responded yet. */
      }
      if (code) {
        await page
          .getByRole("textbox", { name: "Enter verification code", exact: true })
          .fill(code);
        return;
      }
      await new Promise((done) => setTimeout(done, 250));
    }
    throw new Error("Authorized device verification inbox handoff unavailable");
  } finally {
    await rm(input, { force: true });
    await rm(output, { force: true });
  }
};
export const providerLogin = async (
  page: import("@playwright/test").Page,
  account: "TEST" | "INVITEE" = "TEST",
) => {
  await page.goto("/sign-in");
  await page.getByLabel(/email address/i).fill(process.env[`FLUX_PROVIDER_${account}_EMAIL`]!);
  await page.getByRole("button", { name: /continue/i, exact: true }).click();
  await page
    .getByLabel("Password", { exact: true })
    .fill(process.env[`FLUX_PROVIDER_${account}_PASSWORD`]!);
  await page.getByRole("button", { name: /continue/i, exact: true }).click();
  await page.waitForFunction(() => {
    const clerk = (
      window as unknown as {
        Clerk?: { session?: unknown; client?: { signIn?: { status?: string } } };
      }
    ).Clerk;
    return !!clerk?.session || clerk?.client?.signIn?.status === "needs_client_trust";
  });
  const needsTrust = await page.evaluate(
    () =>
      (window as unknown as { Clerk?: { client?: { signIn?: { status?: string } } } }).Clerk?.client
        ?.signIn?.status === "needs_client_trust",
  );
  if (needsTrust) await completeProviderDeviceTrust(page, account);
  await page.getByText("Your account is signed in.", { exact: true }).waitFor();
  return providerToken(page);
};
export const providerAPI = async (
  page: import("@playwright/test").Page,
  path: string,
  method = "GET",
  body?: unknown,
  token?: string,
) => {
  const bearer = token ?? (await providerToken(page));
  return page.evaluate(
    async ({ path, method, body, bearer }) => {
      const response = await fetch(`/api/v1${path}`, {
        method,
        headers: {
          Authorization: `Bearer ${bearer}`,
          "Content-Type": "application/json",
          "Idempotency-Key": crypto.randomUUID(),
        },
        ...(body === undefined ? {} : { body: JSON.stringify(body) }),
        cache: "no-store",
      });
      return { status: response.status, body: (await response.json()) as unknown };
    },
    { path, method, body, bearer },
  );
};
export const providerControl = async <T = Record<string, unknown>>(
  action: string,
  values: Record<string, string> = {},
): Promise<T> => {
  const base = process.env.FLUX_PROVIDER_CONTROL_ORIGIN;
  if (!base || !/^http:\/\/127\.0\.0\.1:\d+$/.test(base) || !process.env.FLUX_PROVIDER_CONTROL_KEY)
    throw new Error("Owned provider runtime unavailable");
  const response = await fetch(base, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${process.env.FLUX_PROVIDER_CONTROL_KEY}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ action, ...values }),
    signal: AbortSignal.timeout(45000),
  });
  if (response.status !== 200) throw new Error("Owned provider runtime assertion failed");
  return response.json() as Promise<T>;
};
export const providerBAPI = async (path: string, method = "GET") => {
  if (
    !/^\/v1\/(?:jwks|sessions\/sess_[A-Za-z0-9]+(?:\/revoke)?)$/.test(path) ||
    !["GET", "POST"].includes(method)
  )
    throw new Error("Invalid provider assertion endpoint");
  const response = await fetch(`https://api.clerk.com${path}`, {
    method,
    headers: { Authorization: `Bearer ${process.env.CLERK_SECRET_KEY}` },
    redirect: "error",
    signal: AbortSignal.timeout(3000),
  });
  if (response.status !== 200) throw new Error("Real provider assertion failed");
  return response.json() as Promise<Record<string, unknown>>;
};
export const readProviderInvitation = async () => {
  const directory = process.env.FLUX_PROVIDER_PRIVATE_DIR!;
  const deadline = Date.now() + 30000;
  while (Date.now() < deadline) {
    try {
      const item = (await readPrivateJSON(resolve(directory, "inbox.json"))) as {
        invitation_url?: string;
      };
      const url = new URL(item.invitation_url ?? "");
      if (
        url.origin !== process.env["FLUX_INVITATIONS.PUBLIC_ORIGIN"] ||
        url.pathname !== "/invitations" ||
        !/^#token=[A-Za-z0-9_-]{43}$/.test(url.hash)
      )
        throw new Error();
      return url.hash;
    } catch {
      /* only a bounded private inbox delivery wait */
    }
    await new Promise((done) => setTimeout(done, 250));
  }
  throw new Error("Authorized invitation inbox evidence unavailable");
};
if (import.meta.main) process.exitCode = await providerConfigCLI(process.argv.slice(2));
