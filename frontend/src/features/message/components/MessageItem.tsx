import { useState } from "react";

import { Avatar, Text } from "@mantine/core";

import { MessageActions } from "./MessageActions";
import { MessageContent } from "./MessageContent";
import { ThreadMetadataPreview } from "./ThreadMetadataPreview";

import type { MessageWithUser, ThreadMetadata } from "../types";

import { MessageAttachment } from "@/features/attachment/components/MessageAttachment";
import { ReactionList } from "@/features/reaction/components/ReactionList";

type MessageItemProps = {
  message: MessageWithUser;
  dateTimeFormatter: Intl.DateTimeFormat;
  currentUserId: string | null;
  onCopyLink: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  onBookmark: (messageId: string) => void;
  onEdit?: (messageId: string, currentBody: string) => void;
  onDelete?: (messageId: string) => void;
  threadMetadata?: ThreadMetadata | null;
  onOpenThread?: (messageId: string) => void;
}

export const MessageItem = ({
  message,
  dateTimeFormatter,
  currentUserId,
  onCopyLink,
  onCreateThread,
  onBookmark,
  onEdit,
  onDelete,
  threadMetadata,
  onOpenThread,
}: MessageItemProps) => {
  const [isHovered, setIsHovered] = useState(false);
  const isAuthor = message.userId === currentUserId;

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
            {message.isDeleted ? (
              <Text size="sm" c="dimmed" fs="italic">
                このメッセージは削除されました
                {message.deletedBy && ` (削除者: ${message.deletedBy.displayName})`}
              </Text>
            ) : (
              <MessageContent message={message} />
            )}
          </div>

          {/* 添付ファイル */}
          {message.attachments && message.attachments.length > 0 && (
            <div className="mt-2 space-y-2">
              {message.attachments.map((attachment) => (
                <MessageAttachment key={attachment.id} attachment={attachment} />
              ))}
            </div>
          )}

          {/* リアクション */}
          <ReactionList message={message} />

          {/* スレッドメタデータプレビュー */}
          {threadMetadata && threadMetadata.replyCount > 0 && onOpenThread && (
            <ThreadMetadataPreview
              metadata={threadMetadata}
              onClick={() => onOpenThread(message.id)}
            />
          )}
        </div>
      </div>

      {/* ホバー時のアクションメニュー */}
      {isHovered && (
        <MessageActions
          messageId={message.id}
          isAuthor={isAuthor}
          isDeleted={message.isDeleted}
          onCopyLink={onCopyLink}
          onCreateThread={onCreateThread}
          onBookmark={onBookmark}
          onEdit={onEdit ? (id) => onEdit(id, message.body) : undefined}
          onDelete={onDelete}
        />
      )}
    </div>
  );
};
