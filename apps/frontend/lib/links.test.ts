import { expect } from "@playwright/test";
import { describe, test } from "node:test";
import {
  blockedDestination,
  destinationSyntaxOK,
  invalidDestination,
  localLinkTime,
  LinkPages,
  invalidCursor,
  normalizeLinkSearch,
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

test("pagination remembers successful positions and resets on exact scope changes", () => {
  const pages = new LinkPages();
  pages.accept("workspace-a|all|search", undefined, "next-a");
  expect(pages.canPrevious).toBe(false);
  expect(pages.next).toBe("next-a");
  pages.accept("workspace-a|all|search", "next-a", "next-b");
  expect(pages.canPrevious).toBe(true);
  expect(pages.previous).toBeUndefined();
  pages.accept("workspace-a|all|search", "next-b", null);
  expect(pages.previous).toBe("next-a");
  expect(pages.next).toBeNull();
  pages.accept("workspace-a|all|search", "next-a", "next-b");
  expect(pages.previous).toBeUndefined();
  for (const scope of [
    "workspace-b|all|search",
    "workspace-b|active|search",
    "workspace-b|active|changed",
  ]) {
    pages.bind(scope);
    expect(pages.canPrevious).toBe(false);
    expect(pages.next).toBeNull();
  }
  expect(invalidCursor).toBe("This page is no longer available. Return to the first page.");
});

test("search normalization preserves literal and exact interior query text", () => {
  expect(normalizeLinkSearch("\u0085\u2003 50% under_score \\ path \u2028")).toBe(
    "50% under_score \\ path",
  );
  expect(normalizeLinkSearch("a  B")).toBe("a  B");
  expect(normalizeLinkSearch("\ufeffa\ufeff")).toBe("\ufeffa\ufeff");
});
