import { z } from "zod";

import { messageWithUserSchema } from "#/features/message/schemas";

export const searchFilterValues = ["all", "messages", "channels", "users"] as const;
export type SearchFilter = (typeof searchFilterValues)[number];

const paginationBaseSchema = z.object({
  hasMore: z.boolean(),
  page: z.number().int().min(1),
  perPage: z.number().int().min(1),
  total: z.number().int().min(0),
});

const channelSearchItemSchema = z.object({
  createdAt: z.string(),
  createdBy: z.string(),
  description: z.string().nullable().optional(),
  hasMention: z.boolean(),
  id: z.string(),
  isPrivate: z.boolean(),
  name: z.string(),
  unreadCount: z.number().int().min(0),
  updatedAt: z.string(),
  workspaceId: z.string(),
});

const memberInfoSchema = z.object({
  avatarUrl: z.string().nullable().optional(),
  displayName: z.string(),
  email: z.email(),
  joinedAt: z.string(),
  role: z.enum(["owner", "admin", "member"]),
  userId: z.string(),
});

const paginatedMessagesSchema = paginationBaseSchema.extend({
  items: z.array(messageWithUserSchema),
});

const paginatedChannelsSchema = paginationBaseSchema.extend({
  items: z.array(channelSearchItemSchema),
});

const paginatedUsersSchema = paginationBaseSchema.extend({
  items: z.array(memberInfoSchema),
});

export const workspaceSearchResponseSchema = z.object({
  channels: paginatedChannelsSchema,
  messages: paginatedMessagesSchema,
  users: paginatedUsersSchema,
});

export type WorkspaceSearchResponse = z.infer<typeof workspaceSearchResponseSchema>;
