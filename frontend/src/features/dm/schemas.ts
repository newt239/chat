import { z } from "zod";

export const createDMRequestSchema = z.object({
  userId: z.string().uuid(),
});

export const createGroupDMRequestSchema = z.object({
  userIds: z.array(z.string().uuid()).min(2).max(9),
  name: z.string().optional(),
});

export type CreateDMRequest = z.infer<typeof createDMRequestSchema>;
export type CreateGroupDMRequest = z.infer<typeof createGroupDMRequestSchema>;
