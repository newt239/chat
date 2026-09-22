import { Text } from "@mantine/core";

import { dateTimeFormatter } from "../utils/time";

import type { SystemMessage } from "../schemas";

type Props = {
  message: SystemMessage;
};

const asText = (value: unknown, fallback = "") => (typeof value === "string" ? value : fallback);

export const SystemMessageItem = ({ message }: Props) => {
  const time = dateTimeFormatter().format(new Date(message.createdAt));
  const { payload } = message;

  const renderText = () => {
    switch (message.kind) {
      case "member_joined": {
        return `ユーザー ${asText(payload.userId)} が参加しました`;
      }
      case "member_added": {
        return `ユーザー ${asText(payload.userId)} が ${asText(payload.addedBy)} により追加されました`;
      }
      case "channel_privacy_changed": {
        return `チャンネルの公開設定が ${asText(payload.from, "public")} から ${asText(payload.to, "public")} に変更されました`;
      }
      case "channel_name_changed": {
        return `チャンネル名が "${asText(payload.from)}" から "${asText(payload.to)}" に変更されました`;
      }
      case "channel_description_changed": {
        return "チャンネルの説明が更新されました";
      }
      case "message_pinned": {
        return `メッセージがピン留めされました（by ${asText(payload.pinnedBy)}）`;
      }
      default: {
        return "システムイベントが記録されました";
      }
    }
  };

  return (
    <div className="px-4 py-2">
      <Text size="xs" c="dimmed">
        {time}
      </Text>
      <Text size="sm" c="dimmed">
        {renderText()}
      </Text>
    </div>
  );
};
