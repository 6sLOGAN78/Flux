import { initClient } from "@ts-rest/core";
import { apiContract } from "@flux/openapi";

export const apiClient = initClient(apiContract, {
  baseUrl: (import.meta.env && import.meta.env.VITE_API_URL) ? import.meta.env.VITE_API_URL : "http://localhost:8080",
  baseHeaders: {},
});
