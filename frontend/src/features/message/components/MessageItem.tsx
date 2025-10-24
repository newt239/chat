import { useState } from "react";

import { Avatar, Text } from "@mantine/core";

import { ReactionList } from "../../reaction/components/ReactionList";

import { MessageActions } from "./MessageActions";
import { MessageContent } from "./MessageContent";

import type { MessageWithUser } from "../types";

interface MessageItemProps {
  message: MessageWithUser;
  dateTimeFormatter: Intl.DateTimeFormat;
  onCopyLink: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  onBookmark: (messageId: string) => void;
}

export const MessageItem = ({
  message,
  dateTimeFormatter,
  onCopyLink,
  onCreateThread,
  onBookmark,
}: MessageItemProps) => {
  const [isHovered, setIsHovered] = useState(false);

  return (
    <div
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className="relative group rounded-md px-4 py-2 transition-colors hover:bg-gray-50"
    >
      {/* アバターとメッセージコンテンツ */}
      <div className="flex gap-3">
        <Avatar
          src={message.user.avatarUrl ?? undefined}
          alt={message.user.displayName}
          size="md"
          color="blue"
          radius="xl"
        >
          {message.user.displayName.charAt(0).toUpperCase()}
        </Avatar>

        {/* メッセージコンテンツ */}
        <div className="flex-1 min-w-0">
          {/* ヘッダー: 名前と日時 */}
          <div className="flex items-baseline gap-2">
            <Text fw={600} size="sm">
              {message.user.displayName}
            </Text>
            <Text size="xs" c="dimmed">
              {dateTimeFormatter.format(new Date(message.createdAt))}
            </Text>
          </div>

          {/* メッセージ本文 */}
          <div className="mt-1">
            <MessageContent message={message} />
          </div>

          {/* リアクション */}
          <ReactionList messageId={message.id} />
        </div>
      </div>

      {/* ホバー時のアクションメニュー */}
      {isHovered && (
        <MessageActions
          messageId={message.id}
          onCopyLink={onCopyLink}
          onCreateThread={onCreateThread}
          onBookmark={onBookmark}
        />
      )}
    </div>
  );
};
