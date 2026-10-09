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
