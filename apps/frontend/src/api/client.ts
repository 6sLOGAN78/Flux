import { initClient } from "@ts-rest/core";
import { apiContract } from "@flux/openapi";

export const apiClient = initClient(apiContract, {
  baseUrl: "http://localhost:8080",
  baseHeaders: {},
});
