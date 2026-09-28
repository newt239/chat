import { Text } from "@mantine/core";

import { SystemMessageKind } from "#/gen/chat/v1/message_pb";
import { toDate } from "#/lib/timestamp";

import { dateTimeFormatter } from "../utils/time";

import type { SystemMessage } from "#/gen/chat/v1/message_pb";

import type { JsonObject } from "@bufbuild/protobuf";

type Props = {
  message: SystemMessage;
};

const asText = (value: unknown, fallback = "") => (typeof value === "string" ? value : fallback);

const systemMessageTexts: Record<SystemMessageKind, (payload: JsonObject) => string> = {
  [SystemMessageKind.UNSPECIFIED]: () => "システムイベントが記録されました",
  [SystemMessageKind.MEMBER_JOINED]: (payload) =>
    `ユーザー ${asText(payload.userId)} が参加しました`,
  [SystemMessageKind.MEMBER_ADDED]: (payload) =>
    `ユーザー ${asText(payload.userId)} が ${asText(payload.addedBy)} により追加されました`,
  [SystemMessageKind.MEMBER_REMOVED]: (payload) =>
    `ユーザー ${asText(payload.userId)} がチャンネルから外されました`,
  [SystemMessageKind.MEMBER_LEFT]: (payload) => `ユーザー ${asText(payload.userId)} が退出しました`,
  [SystemMessageKind.CHANNEL_PRIVACY_CHANGED]: (payload) =>
    `チャンネルの公開設定が ${asText(payload.from, "public")} から ${asText(payload.to, "public")} に変更されました`,
  [SystemMessageKind.CHANNEL_NAME_CHANGED]: (payload) =>
    `チャンネル名が "${asText(payload.from)}" から "${asText(payload.to)}" に変更されました`,
  [SystemMessageKind.CHANNEL_DESCRIPTION_CHANGED]: () => "チャンネルの説明が更新されました",
  [SystemMessageKind.MESSAGE_PINNED]: (payload) =>
    `メッセージがピン留めされました（by ${asText(payload.pinnedBy)}）`,
};

export const SystemMessageItem = ({ message }: Props) => {
  const time = dateTimeFormatter().format(toDate(message.createdAt));
  const payload = message.payload ?? {};

  return (
    <div className="px-4 py-2">
      <Text size="xs" c="dimmed">
        {time}
      </Text>
      <Text size="sm" c="dimmed">
        {systemMessageTexts[message.kind](payload)}
      </Text>
    </div>
  );
};
