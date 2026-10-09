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
