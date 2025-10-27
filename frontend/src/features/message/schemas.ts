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
    reactions: z.array(reactionInfoSchema),
    attachments: z.array(attachmentSchema).optional(),
    createdAt: z.string(),
    editedAt: z.string().nullable().optional(),
    deletedAt: z.string().nullable().optional(),
    isDeleted: z.boolean(),
    deletedBy: messageUserSchema.nullable().optional(),
  })
  .passthrough();

const messageWithUserSchema = baseMessageSchema.extend({
  user: messageUserSchema,
});

export const messagesResponseSchema = z.object({
  messages: z.array(messageWithUserSchema),
  hasMore: z.boolean(),
});

// スレッド返信一覧レスポンススキーマ
export const threadRepliesResponseSchema = z.object({
  parentMessage: messageWithUserSchema,
  replies: z.array(messageWithUserSchema),
  hasMore: z.boolean(),
});

export type MessageWithUser = z.infer<typeof messageWithUserSchema>;

export type ThreadMetadata = {
  messageId: string;
  replyCount: number;
  lastReplyAt?: string | null;
  lastReplyUser?: z.infer<typeof messageUserSchema> | null;
  participantUserIds: string[];
};
