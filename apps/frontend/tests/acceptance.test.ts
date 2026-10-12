import { describe, expect, test } from "bun:test";
import { requiredProviderRows, validateAcceptance } from "./provider.config";

const automatedProof = { evidence_kind: "automated", tests: ["provider:real-test"], assertions: 1 };
const record = () => ({ schema: 1, rows: requiredProviderRows.map((id) => ({ id, status: "passed", ...automatedProof })) });

describe("provider evidence cannot conceal incomplete required boundaries", () => {
  test("missing automated rows fail even with positive remaining results", () => {
    const evidence = record();
    evidence.rows.pop();
    expect(validateAcceptance(evidence, true)).toBe(false);
  });
  test("unknown rows and duplicate identities fail", () => {
    const evidence = record();
    evidence.rows.push({ ...evidence.rows[0]!, id: "unknown" as never });
    expect(validateAcceptance(evidence, true)).toBe(false);
    evidence.rows.pop();
    evidence.rows.push(evidence.rows[0]!);
    expect(validateAcceptance(evidence, true)).toBe(false);
  });
  test("pending automated rows and arbitrary manual exemptions fail", () => {
    const evidence = record();
    Object.assign(evidence.rows.find((row) => row.id === "token_expiry")!, { status: "human_pending", evidence_kind: "human_only", reason_code: "operator_mfa", manual_steps: ["Verify expiry"] });
    expect(validateAcceptance(evidence, true)).toBe(false);
  });
  test("zero executions, SDK factor shortcuts and failed rows fail", () => {
    for (const override of [{ tests: [], assertions: 0 }, { evidence_kind: "sdk_helper" }, { status: "failed" }]) {
      const evidence = record();
      Object.assign(evidence.rows[0]!, override);
      expect(validateAcceptance(evidence, true)).toBe(false);
    }
  });
  test("strict final mode rejects pending human verification", () => {
    const evidence = record();
    Object.assign(evidence.rows.find((row) => row.id === "email_verification")!, { status: "human_pending", evidence_kind: "human_only", reason_code: "authorized_inbox_verification", manual_steps: ["Verify the authorized inbox code"] });
    expect(validateAcceptance(evidence)).toBe(false);
  });
});
