import { extendApi } from "@anatine/zod-openapi";
import { z } from "zod";

const ZHealthState = z.enum(["ready", "not_ready"]);

const ZHealthCheck = z.object({
  name: z.string(),
  state: ZHealthState,
}).strict();

export const ZHealthLiveResponse = extendApi(z.object({
  status: z.literal("alive"),
}).strict(), {
  title: "transport.HealthLiveResponse",
  description: "Process liveness, independent of external dependencies.",
  example: { status: "alive" },
});

export const ZHealthReadyResponse = extendApi(z.object({
  status: ZHealthState,
  checks: z.array(ZHealthCheck),
}).strict(), {
  title: "transport.HealthReadyResponse",
  description: "Readiness of this process and its required components. Only coarse states are exposed.",
  example: { status: "ready", checks: [{ name: "database", state: "ready" }] },
});

export type HealthLiveResponse = z.infer<typeof ZHealthLiveResponse>;
export type HealthReadyResponse = z.infer<typeof ZHealthReadyResponse>;
