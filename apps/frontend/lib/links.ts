export const invalidDestination = "Enter a valid HTTP or HTTPS URL.";
export const blockedDestination =
  "This destination isn't allowed. Choose a different public destination.";

// Syntax feedback only. The Go policy remains authoritative for public hosts.
export const destinationSyntaxOK = (value: string): boolean => {
  if (!value || /[\s\\\u0000-\u001f\u007f]/u.test(value)) return false;
  try {
    const url = new URL(value);
    return (
      ["http:", "https:"].includes(url.protocol) && !!url.hostname && !url.username && !url.password
    );
  } catch {
    return false;
  }
};

export const localLinkTime = (value: string): string =>
  new Date(value).toLocaleString(undefined, { dateStyle: "full", timeStyle: "long" });

export const invalidCursor = "This page is no longer available. Return to the first page.";

// Page positions stay local to one exact workspace/query scope. Only successful
// responses enter history; a failed navigation cannot invent a previous page.
export class LinkPages {
  private scope = "";
  private cursors: (string | undefined)[] = [undefined];
  private index = 0;
  next: string | null = null;

  bind(scope: string) {
    if (this.scope === scope) return;
    this.scope = scope;
    this.cursors = [undefined];
    this.index = 0;
    this.next = null;
  }

  accept(scope: string, cursor: string | undefined, next: string | null) {
    this.bind(scope);
    if (cursor === undefined) {
      this.cursors = [undefined];
      this.index = 0;
    } else {
      const existing = this.cursors.indexOf(cursor);
      if (existing >= 0) this.index = existing;
      else {
        this.cursors = [...this.cursors.slice(0, this.index + 1), cursor];
        this.index++;
      }
    }
    this.next = next;
  }

  get canPrevious() {
    return this.index > 0;
  }
  get previous() {
    return this.cursors[this.index - 1];
  }
}

// Match Go strings.TrimSpace: Unicode White_Space, preserving exact inner text.
export const normalizeLinkSearch = (value: string): string =>
  value.replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");

export type LifecycleAction = "enable" | "disable" | "archive" | "restore" | "delete";
type LifecycleRequest = Readonly<{
  key: string;
  body: Readonly<{ action: LifecycleAction; expectedVersion: string }>;
}>;

// One scoped attempt survives uncertainty unchanged. Only an explicit new
// intention replaces it; no retry can silently adopt a newer displayed version.
export class LifecycleRequests {
  private pending?: LifecycleRequest;
  constructor(private key: () => string = () => crypto.randomUUID()) {}
  start(action: LifecycleAction, expectedVersion: string) {
    this.pending = Object.freeze({
      key: this.key(),
      body: Object.freeze({ action, expectedVersion }),
    });
    return this.pending;
  }
  retry() {
    return this.pending;
  }
  clear() {
    this.pending = undefined;
  }
}
