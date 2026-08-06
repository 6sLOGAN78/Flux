import { z } from "zod";

// --- Entity Schemas ---

export const ZLink = z.object({
  id: z.string().uuid(),
  shortCode: z.string().min(1).max(20),
  destinationUrl: z.string().url(),
  tenantId: z.string().uuid().optional(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export const ZCreateLinkInput = z.object({
  destinationUrl: z.string().url(),
  customCode: z.string().min(3).max(20).optional(),
});

export const ZUser = z.object({
  id: z.string(),
  email: z.string().email(),
  name: z.string().optional(),
  createdAt: z.string().datetime(),
});

export const ZHealthResponse = z.object({
  status: z.string(),
  database: z.string(),
});

// --- Exported Inferred Types ---
export type Link = z.infer<typeof ZLink>;
export type CreateLinkInput = z.infer<typeof ZCreateLinkInput>;
export type User = z.infer<typeof ZUser>;
export type HealthResponse = z.infer<typeof ZHealthResponse>;
