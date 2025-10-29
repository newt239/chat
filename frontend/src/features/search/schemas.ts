import { z } from "zod";

import { messageWithUserSchema } from "@/features/message/schemas";

export const searchFilterValues = ["all", "messages", "channels", "users"] as const;
export type SearchFilter = (typeof searchFilterValues)[number];

const paginationBaseSchema = z.object({
	total: z.number().int().min(0),
	page: z.number().int().min(1),
	perPage: z.number().int().min(1),
	hasMore: z.boolean(),
});

const channelSearchItemSchema = z.object({
	id: z.string(),
	workspaceId: z.string(),
	name: z.string(),
	description: z.string().nullable().optional(),
	isPrivate: z.boolean(),
	createdBy: z.string(),
	createdAt: z.string(),
	updatedAt: z.string(),
	unreadCount: z.number().int().min(0),
	hasMention: z.boolean(),
});

const memberInfoSchema = z.object({
	userId: z.string(),
	email: z.string().email(),
	displayName: z.string(),
	avatarUrl: z.string().nullable().optional(),
	role: z.enum(["owner", "admin", "member"]),
	joinedAt: z.string(),
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
	messages: paginatedMessagesSchema,
	channels: paginatedChannelsSchema,
	users: paginatedUsersSchema,
});

export type WorkspaceSearchResponse = z.infer<typeof workspaceSearchResponseSchema>;
export type PaginatedMessages = z.infer<typeof paginatedMessagesSchema>;
export type PaginatedChannels = z.infer<typeof paginatedChannelsSchema>;
export type PaginatedUsers = z.infer<typeof paginatedUsersSchema>;
