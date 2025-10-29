import { z } from "zod";

const createDMRequestSchema = z.object({
  userId: z.string().uuid(),
});

export type CreateDMRequest = z.infer<typeof createDMRequestSchema>;