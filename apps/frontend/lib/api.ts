export class ApiError extends Error {
  code = "unavailable";
  status = 0;
}

export const createAPI = (_getToken: () => Promise<string | null>, _options: {
  origin: string;
  fetch?: (url: string, init?: RequestInit) => Promise<Response>;
  timeoutMs?: number;
}) => ({
  request: async (_path: string, _request?: { method?: string; body?: unknown; signal?: AbortSignal }): Promise<unknown> => {
    throw new ApiError("The request could not be completed. Try again.");
  },
});
