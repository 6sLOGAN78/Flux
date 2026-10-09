import { expect } from "@playwright/test";
import { describe, test } from "node:test";
import {
  blockedDestination,
  destinationSyntaxOK,
  invalidDestination,
  localLinkTime,
} from "./links";

describe("link presentation uses safe syntax and exact feedback", () => {
  test("syntax helps feedback while public host policy stays on the server", () => {
    for (const value of [
      "",
      "//example.com",
      "javascript:alert(1)",
      "https://user@example.com",
      "https://example.com\\evil",
      "https://example.com/\n",
    ]) {
      expect(destinationSyntaxOK(value)).toBe(false);
    }
    expect(destinationSyntaxOK("https://example.com/a%2Fb?q=a%2Bb")).toBe(true);
    expect(destinationSyntaxOK("http://[fec0::1]")).toBe(true);
    expect(invalidDestination).toBe("Enter a valid HTTP or HTTPS URL.");
    expect(blockedDestination).toBe(
      "This destination isn't allowed. Choose a different public destination.",
    );
  });
  test("timestamps use the local full date and long time with timezone", () => {
    const value = "2026-10-09T11:00:00Z";
    expect(localLinkTime(value)).toBe(
      new Date(value).toLocaleString(undefined, { dateStyle: "full", timeStyle: "long" }),
    );
  });
});
