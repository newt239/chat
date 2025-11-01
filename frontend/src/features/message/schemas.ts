import { z } from "zod";

const messageUserSchema = z.object({
  id: z.string(),
  displayName: z.string(),
  avatarUrl: z.string().nullable().optional(),
  email: z.string().email().optional(),
});

const threadMetadataSchema = z.object({
  messageId: z.string(),
  replyCount: z.number(),
  lastReplyAt: z.string().nullable().optional(),
  lastReplyUser: messageUserSchema.nullable().optional(),
  participantUserIds: z.array(z.string()),
});

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

const reactionInfoSchema = z.object({
  user: messageUserSchema,
  emoji: z.string(),
  createdAt: z.string(),
});

const attachmentSchema = z.object({
  id: z.string(),
  messageId: z.string(),
  fileName: z.string(),
  mimeType: z.string(),
  sizeBytes: z.number(),
  createdAt: z.string(),
});

const baseMessageSchema = z.object({
  id: z.string(),
  channelId: z.string(),
  userId: z.string(),
  parentId: z.string().nullable().optional(),
  body: z.string(),
  mentions: z.array(userMentionSchema).optional(),
  groups: z.array(groupMentionSchema).optional(),
  links: z.array(linkInfoSchema).optional(),
  reactions: z.array(reactionInfoSchema).optional(),
  attachments: z.array(attachmentSchema).optional(),
  createdAt: z.string(),
  editedAt: z.string().nullable().optional(),
  deletedAt: z.string().nullable().optional(),
  isDeleted: z.boolean(),
  deletedBy: messageUserSchema.nullable().optional(),
});

export const messageWithUserSchema = baseMessageSchema.extend({
	user: messageUserSchema,
});

const messageWithThreadSchema = messageWithUserSchema.extend({
  threadMetadata: threadMetadataSchema.nullable().optional(),
});

export { messageWithThreadSchema };

// スレッド返信一覧レスポンススキーマ
export const threadRepliesResponseSchema = z.object({
  parentMessage: messageWithUserSchema,
  replies: z.array(messageWithUserSchema),
  hasMore: z.boolean(),
});

export type MessageWithUser = z.infer<typeof messageWithUserSchema>;
export type ThreadMetadata = z.infer<typeof threadMetadataSchema>;
// System message and timeline unified schema
const systemMessageSchema = z.object({
  id: z.string(),
  channelId: z.string(),
  kind: z.string(),
  payload: z.record(z.string(), z.unknown()),
  actorId: z.string().nullable().optional(),
  createdAt: z.string(),
});

export const timelineItemSchema = z.object({
  type: z.enum(["user", "system"]),
  userMessage: messageWithUserSchema.optional(),
  systemMessage: systemMessageSchema.optional(),
  createdAt: z.string(),
});

export const messagesTimelineResponseSchema = z.object({
  messages: z.array(timelineItemSchema),
  hasMore: z.boolean(),
});

export type SystemMessage = z.infer<typeof systemMessageSchema>;
export type TimelineItem = z.infer<typeof timelineItemSchema>;
