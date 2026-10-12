import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { randomUUID } from "node:crypto";
const expect = (actual: unknown) => ({
  toBe: (expected: unknown) => assert.equal(actual, expected),
  toEqual: (expected: unknown) => assert.deepEqual(actual, expected),
  toBeGreaterThan: (expected: number) => assert.ok(typeof actual === "number" && actual > expected),
});
import {
  manualProviderRows,
  proofHash,
  requiredProviderRows,
  validateAcceptance,
  validateProviderConfig,
  verifyProviderInstance,
  type ProviderAcceptance,
  type ProviderProof,
} from "./provider.config";

const fixture = () => {
  const proof: ProviderProof = {
    schema: 1,
    run_id: randomUUID(),
    created_at: new Date().toISOString(),
    revision: "a".repeat(40),
    config_fingerprint: "b".repeat(64),
    tests: requiredProviderRows.map((id) => ({
      id: `provider:${id}`,
      project: "provider",
      status: "passed",
      retry: 0,
      rows: [{ id, status: "passed" }],
    })),
  };
  const record: ProviderAcceptance = {
    schema: 1,
    run_id: proof.run_id,
    proof_sha256: proofHash(proof),
    rows: requiredProviderRows.map((id) => ({
      id,
      status: "passed",
      evidence_kind: "automated",
      tests: [`provider:${id}`],
    })),
  };
  return { proof, record };
};
const check = (item: ReturnType<typeof fixture>, automated = true) =>
  validateAcceptance(item.record, automated, item.proof);
describe("provider evidence cannot conceal incomplete required boundaries", () => {
  test("complete positive real-execution shape passes both modes", () => {
    const item = fixture();
    expect(check(item)).toBe(true);
    expect(check(item, false)).toBe(true);
  });
  test("missing rows fail even with positive remaining results", () => {
    const item = fixture();
    item.record.rows.pop();
    expect(check(item)).toBe(false);
  });
  test("unknown rows and duplicate identities fail", () => {
    for (const id of ["unknown", requiredProviderRows[0]]) {
      const item = fixture();
      item.record.rows[item.record.rows.length - 1]!.id = id as never;
      expect(check(item)).toBe(false);
    }
  });
  test("failed, missing, unknown and pending automated results fail", () => {
    for (const status of ["failed", undefined, "unknown", "human_pending"]) {
      const item = fixture();
      Object.assign(item.record.rows.find((row) => row.id === "token_expiry")!, {
        status,
        evidence_kind: "human_only",
        reason_code: "operator_mfa",
        manual_steps: ["Verify expiry"],
      });
      expect(check(item)).toBe(false);
    }
  });
  test("zero executions, skipped/retried/local project proof and SDK factors fail", () => {
    for (const override of [
      { tests: [] },
      { tests: [{ ...fixture().proof.tests[0], status: "skipped" }] },
      { tests: [{ ...fixture().proof.tests[0], retry: 1 }] },
      { tests: [{ ...fixture().proof.tests[0], project: "local" }] },
    ]) {
      const item = fixture();
      Object.assign(item.proof, override);
      item.record.proof_sha256 = proofHash(item.proof);
      expect(check(item)).toBe(false);
    }
    const item = fixture();
    Object.assign(item.record.rows[0]!, { evidence_kind: "sdk_helper" });
    expect(check(item)).toBe(false);
  });
  test("only exact enumerated human rows can remain pending with positive automated controls", () => {
    const item = fixture();
    for (const [id, manual] of Object.entries(manualProviderRows)) {
      Object.assign(item.record.rows.find((row) => row.id === id)!, {
        status: "human_pending",
        evidence_kind: "human_only",
        reason_code: manual.reasons[0],
        manual_steps: manual.steps,
      });
      item.proof.tests.find((entry) => entry.rows[0]?.id === id)!.rows[0]!.status = "human_pending";
    }
    item.record.proof_sha256 = proofHash(item.proof);
    expect(check(item)).toBe(true);
    expect(check(item, false)).toBe(false);
    for (const override of [
      { reason_code: "other" },
      { reason_code: undefined },
      { manual_steps: [] },
      { evidence_kind: "automated" },
    ]) {
      const copy = structuredClone(item);
      Object.assign(copy.record.rows.find((row) => row.id === "email_verification")!, override);
      expect(check(copy)).toBe(false);
    }
  });
  test("proof binds run, fresh timestamp, instance fingerprint, native rows and digest", () => {
    for (const override of [
      { run_id: randomUUID() },
      { created_at: "2000-01-01T00:00:00Z" },
      { revision: "unknown" },
    ]) {
      const item = fixture();
      Object.assign(item.proof, override);
      item.record.proof_sha256 = proofHash(item.proof);
      expect(check(item)).toBe(false);
    }
    const item = fixture();
    expect(validateAcceptance(item.record, true, item.proof, "c".repeat(64))).toBe(false);
    item.proof.tests[0]!.rows = [];
    expect(check(item)).toBe(false);
    expect(validateAcceptance(fixture().record, true)).toBe(false);
  });
});
const configuration = () => {
  const secret = ["sk", "test", "syntheticOnly"].join("_");
  return {
    NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY:
      "pk_test_" + Buffer.from("unit.clerk.accounts.dev$").toString("base64"),
    CLERK_SECRET_KEY: secret,
    "FLUX_AUTH.SECRET_KEY": secret,
    "FLUX_AUTH.ISSUER": "https://unit.clerk.accounts.dev",
    "FLUX_AUTH.AUTHORIZED_PARTIES": "http://localhost:3000",
    FLUX_API_ORIGIN: "http://127.0.0.1:8080",
    "FLUX_INVITATIONS.ACTIVE_KEY_ID": "current",
    "FLUX_INVITATIONS.ENCRYPTION_KEYS": JSON.stringify({
      current: Buffer.alloc(32, 1).toString("base64"),
      previous: Buffer.alloc(32, 2).toString("base64"),
    }),
    "FLUX_INVITATIONS.SENDER": "sender@example.test",
    "FLUX_INVITATIONS.PUBLIC_ORIGIN": "https://app.example.test",
    "FLUX_INTEGRATION.RESEND_API_KEY": ["re", "syntheticOnly0123456789"].join("_"),
    "FLUX_LINKS.MANAGED_HOST": "links.example.test",
    "FLUX_LINKS.CURSOR_KEY": Buffer.alloc(32, 3).toString("base64"),
    FLUX_PROVIDER_TEST_EMAIL: "owner@example.test",
    FLUX_PROVIDER_TEST_PASSWORD: "synthetic-only-password",
    FLUX_PROVIDER_SIGNUP_EMAIL: "signup@example.test",
    FLUX_PROVIDER_SIGNUP_PASSWORD: "synthetic-only-password",
    FLUX_PROVIDER_INVITEE_EMAIL: "invitee@example.test",
    FLUX_PROVIDER_INVITEE_PASSWORD: "synthetic-only-password",
    FLUX_PROVIDER_AUTHORIZED_RECIPIENTS:
      "owner@example.test,signup@example.test,invitee@example.test",
    FLUX_PROVIDER_PRIVATE_DIR: "/tmp/provider-unit-only",
  };
};
describe("private provider configuration", () => {
  test("real instance lookup must match BAPI and configured issuer keys", async () => {
    const keys = { keys: [{ kid: "unit-key", kty: "RSA", n: "unit-modulus", e: "AQAB" }] };
    const fetcher = (async () => Response.json(keys)) as typeof fetch;
    expect(await verifyProviderInstance(configuration(), fetcher)).toBe(true);
    let calls = 0;
    const mismatched = (async () =>
      Response.json(
        ++calls === 1 ? keys : { keys: [{ ...keys.keys[0], kid: "other-instance" }] },
      )) as typeof fetch;
    expect(await verifyProviderInstance(configuration(), mismatched)).toBe(false);
    const unavailable = (async () => new Response(null, { status: 503 })) as typeof fetch;
    expect(await verifyProviderInstance(configuration(), unavailable)).toBe(false);
  });
  test("complete coherent development configuration passes shape validation", () =>
    expect(validateProviderConfig(configuration())).toEqual([]));
  test("missing configuration yields safe variable names only", () => {
    const errors = validateProviderConfig({});
    expect(errors.length).toBeGreaterThan(10);
    expect(errors.every((name) => /^[A-Z_.]+$/.test(name))).toBe(true);
  });
  test("mismatched instances/secrets, unauthorized recipients and malformed AES maps fail", () => {
    for (const change of [
      { "FLUX_AUTH.ISSUER": "https://other.clerk.accounts.dev" },
      { "FLUX_AUTH.SECRET_KEY": "other" },
      { "FLUX_AUTH.AUTHORIZED_PARTIES": "*" },
      { FLUX_API_ORIGIN: "https://other.example.test" },
      { "FLUX_INVITATIONS.ENCRYPTION_KEYS": "{}" },
      {
        "FLUX_INVITATIONS.ENCRYPTION_KEYS": JSON.stringify({ current: "short", previous: "short" }),
      },
      { "FLUX_INVITATIONS.ACTIVE_KEY_ID": "missing" },
      { "FLUX_INVITATIONS.PUBLIC_ORIGIN": "https://app.example.test/path" },
      { FLUX_PROVIDER_AUTHORIZED_RECIPIENTS: "other@example.test" },
      { FLUX_PROVIDER_PRIVATE_DIR: "relative" },
    ])
      expect(validateProviderConfig({ ...configuration(), ...change }).length).toBeGreaterThan(0);
  });
});
