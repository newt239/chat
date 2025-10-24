import { z } from "zod";

const messageUserSchema = z
  .object({
    id: z.string(),
    displayName: z.string(),
    avatarUrl: z.string().nullable().optional(),
    email: z.string().email().optional(),
  })
  .passthrough();

const userMentionSchema = z.object({
  userId: z.string(),
  displayName: z.string(),
});

const groupMentionSchema = z.object({
  groupId: z.string(),
  name: z.string(),
});

const linkInfoSchema = z.object({
  id: z.string(),
  url: z.string(),
  title: z.string().nullable().optional(),
  description: z.string().nullable().optional(),
  imageUrl: z.string().nullable().optional(),
  siteName: z.string().nullable().optional(),
  cardType: z.string().nullable().optional(),
});

const baseMessageSchema = z
  .object({
    id: z.string(),
    channelId: z.string(),
    userId: z.string(),
    parentId: z.string().nullable().optional(),
    body: z.string(),
    mentions: z.array(userMentionSchema),
    groups: z.array(groupMentionSchema),
    links: z.array(linkInfoSchema),
    createdAt: z.string(),
    editedAt: z.string().nullable().optional(),
    deletedAt: z.string().nullable().optional(),
  })
  .passthrough();

export const messageWithUserSchema = baseMessageSchema.extend({
  user: messageUserSchema,
});

export const messagesResponseSchema = z.object({
  messages: z.array(messageWithUserSchema),
  hasMore: z.boolean(),
});

export type MessageWithUser = z.infer<typeof messageWithUserSchema>;

export type MessagesResponse = z.infer<typeof messagesResponseSchema>;
