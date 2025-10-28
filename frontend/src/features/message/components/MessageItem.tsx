import { useEffect, useState } from "react";

import { Avatar, Button, Group, Text, Textarea } from "@mantine/core";

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
  onEdit?: (messageId: string, nextBody: string) => Promise<void>;
  onDelete?: (messageId: string) => Promise<void>;
  threadMetadata?: ThreadMetadata | null;
  onOpenThread?: (messageId: string) => void;
};

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
  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState(message.body);
  const [isSaving, setIsSaving] = useState(false);
  const [editError, setEditError] = useState<string | null>(null);
  const isAuthor = message.userId === currentUserId;

  useEffect(() => {
    if (!isEditing) {
      setDraft(message.body);
      setEditError(null);
    }
  }, [isEditing, message.body]);

  const handleStartEdit = () => {
    if (message.isDeleted || !isAuthor) {
      return;
    }
    setDraft(message.body);
    setIsEditing(true);
  };

  const handleCancelEdit = () => {
    setDraft(message.body);
    setEditError(null);
    setIsEditing(false);
  };

  const handleSaveEdit = async () => {
    if (!onEdit) {
      return;
    }

    const trimmed = draft.trim();
    if (trimmed.length === 0) {
      setEditError("メッセージ本文を入力してください");
      return;
    }

    if (trimmed === message.body) {
      setIsEditing(false);
      setEditError(null);
      return;
    }

    try {
      setIsSaving(true);
      await onEdit(message.id, trimmed);
      setIsEditing(false);
      setEditError(null);
    } catch (error) {
      if (error instanceof Error && error.message) {
        setEditError(error.message);
      } else {
        setEditError("メッセージの更新に失敗しました");
      }
    } finally {
      setIsSaving(false);
    }
  };

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
            ) : isEditing ? (
              <div className="space-y-2">
                <Textarea
                  value={draft}
                  onChange={(event) => setDraft(event.currentTarget.value)}
                  minRows={3}
                  autosize
                  disabled={isSaving}
                  data-autofocus
                />
                {editError && (
                  <Text size="xs" c="red">
                    {editError}
                  </Text>
                )}
                <Group gap="xs">
                  <Button size="xs" onClick={handleSaveEdit} loading={isSaving} disabled={isSaving}>
                    保存
                  </Button>
                  <Button size="xs" variant="subtle" onClick={handleCancelEdit} disabled={isSaving}>
                    キャンセル
                  </Button>
                </Group>
              </div>
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
      {isHovered && !isEditing && (
        <MessageActions
          messageId={message.id}
          isAuthor={isAuthor}
          isDeleted={message.isDeleted}
          onCopyLink={onCopyLink}
          onCreateThread={onCreateThread}
          onBookmark={onBookmark}
          onEditRequest={handleStartEdit}
          onDelete={onDelete}
        />
      )}
    </div>
  );
};
