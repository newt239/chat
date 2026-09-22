import { z } from "zod";

// APIのParticipatingThreadsOutputに対応する最小限のzodスキーマ
const participatingThreadSchema = z.object({
  channel_id: z.string().nullable().optional(),
  first_message: z.object({
    attachments: z
      .array(
        z.object({
          createdAt: z.string(),
          fileName: z.string(),
          id: z.string(),
          messageId: z.string(),
          mimeType: z.string(),
          sizeBytes: z.number(),
        }),
      )
      .optional(),
    body: z.string(),
    channelId: z.string(),
    createdAt: z.string(),
    deletedAt: z.string().nullable().optional(),
    deletedBy: z
      .object({
        avatarUrl: z.string().nullable().optional(),
        displayName: z.string().optional(),
        id: z.string().optional(),
      })
      .nullable()
      .optional(),
    editedAt: z.string().nullable().optional(),
    id: z.string(),
    isDeleted: z.boolean(),
    parentId: z.string().nullable().optional(),
    userId: z.string(),
  }),
  last_activity_at: z.string(),
  reply_count: z.number(),
  thread_id: z.string(),
  unread_count: z.number(),
});

export const participatingThreadsResponseSchema = z.object({
  items: z.array(participatingThreadSchema),
  next_cursor: z
    .object({
      last_activity_at: z.string(),
      thread_id: z.string(),
    })
    .optional(),
});

export type ParticipatingThread = z.infer<typeof participatingThreadSchema>;
