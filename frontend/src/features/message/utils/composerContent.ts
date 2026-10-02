import type { MessageLocation, PollInput } from "#/gen/chat/v1/message_pb";

// 入力欄から送る内容。CreateMessage の入力にそのまま広げて使う
export type ComposerContent = {
  body: string;
  attachmentIds: string[];
  location: MessageLocation | undefined;
  poll: PollInput | undefined;
};
