import { initContract } from "@ts-rest/core";
import { healthContract } from "./health.js";
import { identityContract } from "./identity.js";

const c = initContract();

export const apiContract = c.router({
  Health: healthContract,
  Identity: identityContract,
});
