// event.go由来 WebSocket TypeScript型定義

export const WS_EVENT_TYPE = {
  JOIN_CHANNEL: "join_channel",
  LEAVE_CHANNEL: "leave_channel",
  POST_MESSAGE: "post_message",
  TYPING: "typing",
  UPDATE_READ_STATE: "update_read_state",
  NEW_MESSAGE: "new_message",
  MESSAGE_UPDATED: "message_updated",
  MESSAGE_DELETED: "message_deleted",
  UNREAD_COUNT: "unread_count",
  PIN_CREATED: "pin_created",
  PIN_DELETED: "pin_deleted",
  ACK: "ack",
  ERROR: "error",
} as const;

export type WsEventType = (typeof WS_EVENT_TYPE)[keyof typeof WS_EVENT_TYPE];

// ペイロード型定義
export type JoinChannelPayload = { channel_id: string };
export type LeaveChannelPayload = { channel_id: string };
export type PostMessagePayload = { channel_id: string; body: string };
export type TypingPayload = { channel_id: string };
export type UpdateReadStatePayload = { channel_id: string; message_id: string };
export type NewMessagePayload = { channel_id: string; message: Record<string, unknown> };
export type MessageUpdatedPayload = { channel_id: string; message: Record<string, unknown> };
export type MessageDeletedPayload = { channel_id: string; deleteData: Record<string, unknown> };
export type PinPayload = {
  channel_id: string;
  message: Record<string, unknown>;
  pinned_by: string;
  pinned_at: string;
};
export type UnreadCountPayload = { channel_id: string; unread_count: number; has_mention: boolean };
export type AckPayload = { type: WsEventType; success: boolean; message?: string };
export type ErrorPayload = { code: string; message: string };

// クライアント→サーバーメッセージ
export type ClientToServerMessage =
  | { type: typeof WS_EVENT_TYPE.JOIN_CHANNEL; payload: JoinChannelPayload }
  | { type: typeof WS_EVENT_TYPE.LEAVE_CHANNEL; payload: LeaveChannelPayload }
  | { type: typeof WS_EVENT_TYPE.POST_MESSAGE; payload: PostMessagePayload }
  | { type: typeof WS_EVENT_TYPE.TYPING; payload: TypingPayload }
  | { type: typeof WS_EVENT_TYPE.UPDATE_READ_STATE; payload: UpdateReadStatePayload };

// サーバー→クライアントメッセージ
export type ServerToClientMessage =
  | { type: typeof WS_EVENT_TYPE.NEW_MESSAGE; payload: NewMessagePayload }
  | { type: typeof WS_EVENT_TYPE.MESSAGE_UPDATED; payload: MessageUpdatedPayload }
  | { type: typeof WS_EVENT_TYPE.MESSAGE_DELETED; payload: MessageDeletedPayload }
  | { type: typeof WS_EVENT_TYPE.UNREAD_COUNT; payload: UnreadCountPayload }
  | {
      type: typeof WS_EVENT_TYPE.PIN_CREATED | typeof WS_EVENT_TYPE.PIN_DELETED;
      payload: PinPayload;
    }
  | { type: typeof WS_EVENT_TYPE.ACK; payload: AckPayload }
  | { type: typeof WS_EVENT_TYPE.ERROR; payload: ErrorPayload };

export type WsEventPayloadMap = {
  new_message: NewMessagePayload;
  message_updated: MessageUpdatedPayload;
  message_deleted: MessageDeletedPayload;
  unread_count: UnreadCountPayload;
  pin_created: PinPayload;
  pin_deleted: PinPayload;
  ack: AckPayload;
  error: ErrorPayload;
};
