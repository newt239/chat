import { Text } from "@mantine/core";

import type { SystemMessage } from "../schemas";

type Props = {
  message: SystemMessage;
  dateTimeFormatter: Intl.DateTimeFormat;
};

export const SystemMessageItem = ({ message, dateTimeFormatter }: Props) => {
  const time = dateTimeFormatter.format(new Date(message.createdAt));
  const payload = message.payload as Record<string, unknown>;

  const renderText = () => {
    switch (message.kind) {
      case "member_joined": {
        const userId = String(payload.userId ?? "");
        return `ユーザー ${userId} が参加しました`;
      }
      case "member_added": {
        const userId = String(payload.userId ?? "");
        const addedBy = String(payload.addedBy ?? "");
        return `ユーザー ${userId} が ${addedBy} により追加されました`;
      }
      case "channel_privacy_changed": {
        const from = String(payload.from ?? "public");
        const to = String(payload.to ?? "public");
        return `チャンネルの公開設定が ${from} から ${to} に変更されました`;
      }
      case "channel_name_changed": {
        const from = String(payload.from ?? "");
        const to = String(payload.to ?? "");
        return `チャンネル名が "${from}" から "${to}" に変更されました`;
      }
      case "channel_description_changed": {
        return `チャンネルの説明が更新されました`;
      }
      case "message_pinned": {
        const pinnedBy = typeof payload.pinnedBy === "string" ? payload.pinnedBy : "";
        return `メッセージがピン留めされました（by ${pinnedBy}）`;
      }
      default:
        return "システムイベントが記録されました";
    }
  };

  return (
    <div className="px-4 py-2">
      <Text size="xs" c="dimmed">{time}</Text>
      <Text size="sm" c="dimmed">{renderText()}</Text>
    </div>
  );
};


