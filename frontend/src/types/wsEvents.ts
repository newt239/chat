type ClientEventType =
  | "join_channel"
  | "leave_channel"
  | "post_message"
  | "typing"
  | "update_read_state";

type ServerEventType =
  | "new_message"
  | "message_updated"
  | "message_deleted"
  | "unread_count"
  | "pin_created"
  | "pin_deleted"
  | "system_message_created"
  | "ack"
  | "error";

type WsEventType = ClientEventType | ServerEventType;

// ペイロード型定義
type JoinChannelPayload = { channel_id: string };
type LeaveChannelPayload = { channel_id: string };
type PostMessagePayload = { channel_id: string; body: string };
type TypingPayload = { channel_id: string };
type UpdateReadStatePayload = { channel_id: string; message_id: string };
export type NewMessagePayload = { channel_id: string; message: Record<string, unknown> };
type MessageUpdatedPayload = { channel_id: string; message: Record<string, unknown> };
type MessageDeletedPayload = { channel_id: string; deleteData: Record<string, unknown> };
type PinPayload = {
  channel_id: string;
  message: Record<string, unknown>;
  pinned_by: string;
  pinned_at: string;
};
type UnreadCountPayload = { channel_id: string; unread_count: number; has_mention: boolean };
export type SystemMessageCreatedPayload = { channel_id: string; message: Record<string, unknown> };
type AckPayload = { type: WsEventType; success: boolean; message?: string };
type ErrorPayload = { code: string; message: string };

// クライアント→サーバーメッセージ
export type ClientToServerMessage =
  | { type: "join_channel"; payload: JoinChannelPayload }
  | { type: "leave_channel"; payload: LeaveChannelPayload }
  | { type: "post_message"; payload: PostMessagePayload }
  | { type: "typing"; payload: TypingPayload }
  | { type: "update_read_state"; payload: UpdateReadStatePayload };

export type WsEventPayloadMap = {
  new_message: NewMessagePayload;
  message_updated: MessageUpdatedPayload;
  message_deleted: MessageDeletedPayload;
  unread_count: UnreadCountPayload;
  pin_created: PinPayload;
  pin_deleted: PinPayload;
  system_message_created: SystemMessageCreatedPayload;
  ack: AckPayload;
  error: ErrorPayload;
};
