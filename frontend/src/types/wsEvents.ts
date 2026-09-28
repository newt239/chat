import { z } from "zod";

import type { MessageWithThread, SystemMessage } from "#/features/message/types";

const clientEventTypes = [
  "join_channel",
  "leave_channel",
  "post_message",
  "typing",
  "update_read_state",
] as const;

const serverEventTypes = [
  "new_message",
  "message_updated",
  "message_deleted",
  "unread_count",
  "pin_created",
  "pin_deleted",
  "system_message_created",
  "reaction_added",
  "reaction_removed",
  "typing",
  "stop_typing",
  "ack",
  "error",
] as const;

const wsEventTypeSchema = z.enum([...clientEventTypes, ...serverEventTypes]);

const messageSchema = z.custom<MessageWithThread>(
  (value) => typeof value === "object" && value !== null,
);
const systemMessageSchema = z.custom<SystemMessage>(
  (value) => typeof value === "object" && value !== null,
);

const channelMessagePayloadSchema = z.object({
  channel_id: z.string(),
  message: messageSchema,
});

// ピン通知はメッセージ本体ではなくピン情報を送る
const pinPayloadSchema = z.object({
  channel_id: z.string(),
  message: z.object({
    message: z.string(),
    pinnedAt: z.string(),
    pinnedBy: z.string(),
  }),
  pinned_at: z.string().optional(),
  pinned_by: z.string().optional(),
});

const reactionPayloadSchema = z.object({
  channel_id: z.string(),
  emoji: z.string(),
  message_id: z.string(),
  user_id: z.string(),
});

const typingPayloadSchema = z.object({
  channel_id: z.string(),
  user_id: z.string(),
});

const serverEventSchema = z.discriminatedUnion("type", [
  z.object({ payload: channelMessagePayloadSchema, type: z.literal("new_message") }),
  z.object({ payload: channelMessagePayloadSchema, type: z.literal("message_updated") }),
  z.object({
    payload: z.object({
      channel_id: z.string(),
      deleteData: z.object({ deleted_at: z.string(), id: z.string() }),
    }),
    type: z.literal("message_deleted"),
  }),
  z.object({
    payload: z.object({
      channel_id: z.string(),
      has_mention: z.boolean(),
      unread_count: z.number(),
    }),
    type: z.literal("unread_count"),
  }),
  z.object({ payload: pinPayloadSchema, type: z.literal("pin_created") }),
  z.object({ payload: pinPayloadSchema, type: z.literal("pin_deleted") }),
  z.object({
    payload: z.object({ channel_id: z.string(), message: systemMessageSchema }),
    type: z.literal("system_message_created"),
  }),
  z.object({ payload: reactionPayloadSchema, type: z.literal("reaction_added") }),
  z.object({ payload: reactionPayloadSchema, type: z.literal("reaction_removed") }),
  z.object({ payload: typingPayloadSchema, type: z.literal("typing") }),
  z.object({ payload: typingPayloadSchema, type: z.literal("stop_typing") }),
  z.object({
    payload: z.object({
      message: z.string().optional(),
      success: z.boolean(),
      type: wsEventTypeSchema,
    }),
    type: z.literal("ack"),
  }),
  z.object({
    payload: z.object({ code: z.string(), message: z.string() }),
    type: z.literal("error"),
  }),
]);

export const parseServerEvent = (data: string) => serverEventSchema.safeParse(JSON.parse(data));

type ServerEvent = z.infer<typeof serverEventSchema>;

export type WsEventPayloadMap = {
  [K in ServerEvent["type"]]: Extract<ServerEvent, { type: K }>["payload"];
};

export type WsEventType = ServerEvent["type"];

// クライアント→サーバーメッセージ
export type ClientToServerMessage =
  | { type: "join_channel"; payload: { channel_id: string } }
  | { type: "leave_channel"; payload: { channel_id: string } }
  | { type: "post_message"; payload: { channel_id: string; body: string } }
  | { type: "typing"; payload: { channel_id: string } }
  | { type: "stop_typing"; payload: { channel_id: string } }
  | { type: "update_read_state"; payload: { channel_id: string; message_id: string } };
