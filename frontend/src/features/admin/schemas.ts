import { z } from "zod";

export const adminTabValues = ["overview", "members", "permissions", "audit"] as const;

export const adminSearchSchema = z.object({
  tab: z.enum(adminTabValues).default("overview").catch("overview"),
});
