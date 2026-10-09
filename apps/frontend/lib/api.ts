type Code =
  | "invalid_request"
  | "unauthenticated"
  | "forbidden"
  | "cancelled"
  | "unavailable"
  | "request_failed";
const messages: Record<Code, string> = {
  invalid_request: "The request is invalid.",
  unauthenticated: "Sign in to continue.",
  forbidden: "You do not have access to this resource.",
  cancelled: "The request was cancelled.",
  unavailable: "The request could not be completed. Try again.",
  request_failed: "The request could not be completed.",
};

export class ApiError extends Error {
  constructor(
    public readonly code: Code,
    public readonly status = 0,
    public readonly constraint?:
      | "KEY_UNAVAILABLE"
      | "REQUEST_REUSE_CONFLICT"
      | "INVALID_CUSTOM_KEY"
      | "CURSOR_INVALID",
  ) {
    super(messages[code]);
    this.name = "ApiError";
  }
}

type APIOptions = {
  origin: string;
  fetch?: (url: string, init?: RequestInit) => Promise<Response>;
  timeoutMs?: number;
};
type APIRequest = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  signal?: AbortSignal;
  query?: URLSearchParams;
  idempotencyKey?: string;
};

// Called with Clerk useAuth().getToken and window.location.origin. Tokens stay
// in the invocation only; cookies, redirects and arbitrary request URLs cannot
// carry credentials out of the fixed same-origin Go API boundary.
export const createAPI = (getToken: () => Promise<string | null>, options: APIOptions) => {
  const origin = new URL(options.origin);
  if (
    origin.origin !== options.origin ||
    origin.username ||
    origin.password ||
    !(
      origin.protocol === "https:" ||
      (origin.protocol === "http:" && ["127.0.0.1", "localhost", "[::1]"].includes(origin.hostname))
    )
  ) {
    throw new ApiError("invalid_request");
  }
  const timeoutMs = options.timeoutMs ?? 10000;
  if (!Number.isFinite(timeoutMs) || timeoutMs < 1 || timeoutMs > 30000)
    throw new ApiError("invalid_request");
  return {
    async request<T = unknown>(path: string, request: APIRequest = {}): Promise<T> {
      if (
        !/^\/[A-Za-z0-9_-]+(?:\/[A-Za-z0-9_-]+)*$/.test(path) ||
        (request.idempotencyKey && !/^[A-Za-z0-9_-]{16,128}$/.test(request.idempotencyKey))
      )
        throw new ApiError("invalid_request");
      if (request.signal?.aborted) throw new ApiError("cancelled");
      const controller = new AbortController();
      const cancel = () => controller.abort();
      request.signal?.addEventListener("abort", cancel, { once: true });
      const timer = setTimeout(cancel, timeoutMs);
      let abortHandler: (() => void) | undefined;
      const aborted = new Promise<never>((_resolve, reject) => {
        abortHandler = () =>
          reject(new ApiError(request.signal?.aborted ? "cancelled" : "unavailable"));
        controller.signal.addEventListener("abort", abortHandler, { once: true });
      });
      const perform = async (): Promise<T> => {
        const token = await getToken();
        if (controller.signal.aborted) throw new ApiError("cancelled");
        if (!token || /[\r\n]/.test(token)) throw new ApiError("unauthenticated");
        const target = new URL(`/api/v1${path}`, origin);
        if (request.query) target.search = request.query.toString();
        const headers = new Headers({
          Authorization: `Bearer ${token}`,
          Accept: "application/json",
        });
        if (request.body !== undefined) headers.set("Content-Type", "application/json");
        if (request.idempotencyKey) headers.set("Idempotency-Key", request.idempotencyKey);
        const response = await (options.fetch ?? globalThis.fetch)(target.href, {
          method: request.method ?? "GET",
          headers,
          body: request.body === undefined ? undefined : JSON.stringify(request.body),
          signal: controller.signal,
          cache: "no-store",
          credentials: "omit",
          redirect: "error",
          mode: "same-origin",
          referrerPolicy: "no-referrer",
        });
        if (!response.ok) {
          // Only a closed canonical code may cross the error boundary. Never
          // expose server messages, field values or unbounded error bodies.
          let constraint: ApiError["constraint"];
          const reader = response.body?.getReader();
          if (reader) {
            const chunks: Uint8Array[] = [];
            let size = 0;
            try {
              while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                size += value.byteLength;
                if (size > 8192) {
                  await reader.cancel();
                  break;
                }
                chunks.push(value);
              }
              if (size <= 8192) {
                const bytes = new Uint8Array(size);
                let offset = 0;
                for (const chunk of chunks) {
                  bytes.set(chunk, offset);
                  offset += chunk.byteLength;
                }
                const parsed = ZIdentityError.safeParse(
                  JSON.parse(new TextDecoder().decode(bytes)),
                );
                if (parsed.success && parsed.data.status === response.status) {
                  const code = parsed.data.code;
                  if (
                    (response.status === 409 &&
                      (code === "KEY_UNAVAILABLE" || code === "REQUEST_REUSE_CONFLICT")) ||
                    (response.status === 400 &&
                      (code === "INVALID_CUSTOM_KEY" || code === "CURSOR_INVALID"))
                  )
                    constraint = code;
                }
              }
            } catch {
              // Invalid or interrupted error bodies retain fixed status feedback.
            } finally {
              reader.releaseLock();
            }
          }
          throw new ApiError(
            response.status === 401
              ? "unauthenticated"
              : response.status === 403
                ? "forbidden"
                : response.status === 429 || response.status >= 500
                  ? "unavailable"
                  : "request_failed",
            response.status,
            constraint,
          );
        }
        if (response.status === 204) return undefined as T;
        // A 100-row link library may contain 8192-byte destinations. Go's JSON
        // encoder can expand each byte to six bytes; keep this larger allowance
        // confined to the bounded collection GET. Other responses remain 64 KiB.
        const maxResponseBytes =
          (request.method ?? "GET") === "GET" &&
          /^\/workspaces\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\/links$/i.test(
            path,
          )
            ? 8 * 1024 * 1024
            : 65536;
        const reader = response.body?.getReader();
        if (!reader) throw new ApiError("unavailable");
        let size = 0;
        const chunks: Uint8Array[] = [];
        try {
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            size += value.byteLength;
            if (size > maxResponseBytes) {
              await reader.cancel();
              throw new ApiError("unavailable");
            }
            chunks.push(value);
          }
        } finally {
          reader.releaseLock();
        }
        const bytes = new Uint8Array(size);
        let offset = 0;
        for (const chunk of chunks) {
          bytes.set(chunk, offset);
          offset += chunk.byteLength;
        }
        return JSON.parse(new TextDecoder().decode(bytes)) as T;
      };
      try {
        return await Promise.race([perform(), aborted]);
      } catch (error) {
        if (error instanceof ApiError) throw error;
        throw new ApiError(request.signal?.aborted ? "cancelled" : "unavailable");
      } finally {
        clearTimeout(timer);
        request.signal?.removeEventListener("abort", cancel);
        if (abortHandler) controller.signal.removeEventListener("abort", abortHandler);
      }
    },
  };
};
import { ZIdentityError } from "@flux/zod";
