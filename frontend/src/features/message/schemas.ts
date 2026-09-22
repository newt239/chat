import { z } from "zod";

const messageUserSchema = z.object({
  avatarUrl: z.string().nullable().optional(),
  displayName: z.string(),
  email: z.email().optional(),
  id: z.string(),
});

const threadMetadataSchema = z.object({
  lastReplyAt: z.string().nullable().optional(),
  lastReplyUser: messageUserSchema.nullable().optional(),
  messageId: z.string(),
  participantUserIds: z.array(z.string()),
  replyCount: z.number(),
});

const userMentionSchema = z.object({
  displayName: z.string(),
  userId: z.string(),
});

const groupMentionSchema = z.object({
  groupId: z.string(),
  name: z.string(),
});

const linkInfoSchema = z.object({
  cardType: z.string().nullable().optional(),
  description: z.string().nullable().optional(),
  id: z.string(),
  imageUrl: z.string().nullable().optional(),
  siteName: z.string().nullable().optional(),
  title: z.string().nullable().optional(),
  url: z.string(),
});

const reactionInfoSchema = z.object({
  createdAt: z.string(),
  emoji: z.string(),
  user: messageUserSchema,
});

const attachmentSchema = z.object({
  createdAt: z.string(),
  fileName: z.string(),
  id: z.string(),
  messageId: z.string(),
  mimeType: z.string(),
  sizeBytes: z.number(),
});

const baseMessageSchema = z.object({
  attachments: z.array(attachmentSchema).optional(),
  body: z.string(),
  channelId: z.string(),
  createdAt: z.string(),
  deletedAt: z.string().nullable().optional(),
  deletedBy: messageUserSchema.nullable().optional(),
  editedAt: z.string().nullable().optional(),
  groups: z.array(groupMentionSchema).optional(),
  id: z.string(),
  isDeleted: z.boolean(),
  links: z.array(linkInfoSchema).optional(),
  mentions: z.array(userMentionSchema).optional(),
  parentId: z.string().nullable().optional(),
  reactions: z.array(reactionInfoSchema).optional(),
  userId: z.string(),
});

export const messageWithUserSchema = baseMessageSchema.extend({
  user: messageUserSchema,
});

export const messageWithThreadSchema = messageWithUserSchema.extend({
  threadMetadata: threadMetadataSchema.nullable().optional(),
});

// スレッド返信一覧レスポンススキーマ
export const threadRepliesResponseSchema = z.object({
  hasMore: z.boolean(),
  parentMessage: messageWithUserSchema,
  replies: z.array(messageWithUserSchema),
});

export type MessageWithUser = z.infer<typeof messageWithUserSchema>;
export type ThreadMetadata = z.infer<typeof threadMetadataSchema>;
// System message and timeline unified schema
export const systemMessageSchema = z.object({
  actorId: z.string().nullable().optional(),
  channelId: z.string(),
  createdAt: z.string(),
  id: z.string(),
  kind: z.string(),
  payload: z.record(z.string(), z.unknown()),
});

export const timelineItemSchema = z.object({
  createdAt: z.string(),
  systemMessage: systemMessageSchema.optional(),
  type: z.enum(["user", "system"]),
  userMessage: messageWithUserSchema.optional(),
});

export const messagesTimelineResponseSchema = z.object({
  hasMore: z.boolean(),
  messages: z.array(timelineItemSchema),
});

export type SystemMessage = z.infer<typeof systemMessageSchema>;
export type TimelineItem = z.infer<typeof timelineItemSchema>;
export type MessagesTimelineResponse = z.infer<typeof messagesTimelineResponseSchema>;
