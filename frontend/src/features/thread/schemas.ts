import { z } from "zod";

// APIのParticipatingThreadsOutputに対応する最小限のzodスキーマ
export const participatingThreadSchema = z.object({
  thread_id: z.string(),
  channel_id: z.string().nullable().optional(),
  first_message: z.object({
    id: z.string(),
    channelId: z.string(),
    userId: z.string(),
    parentId: z.string().nullable().optional(),
    body: z.string(),
    createdAt: z.string(),
    editedAt: z.string().nullable().optional(),
    deletedAt: z.string().nullable().optional(),
    isDeleted: z.boolean(),
    attachments: z
      .array(
        z.object({
          id: z.string(),
          messageId: z.string(),
          fileName: z.string(),
          mimeType: z.string(),
          sizeBytes: z.number(),
          createdAt: z.string(),
        })
      )
      .optional(),
    deletedBy: z
      .object({
        id: z.string().optional(),
        displayName: z.string().optional(),
        avatarUrl: z.string().nullable().optional(),
      })
      .nullable()
      .optional(),
  }),
  reply_count: z.number(),
  last_activity_at: z.string(),
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
export type ParticipatingThreadsResponse = z.infer<typeof participatingThreadsResponseSchema>;
