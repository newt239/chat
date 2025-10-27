import { useMemo, useCallback } from "react";

import { Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useSetAtom } from "jotai";

import { MessageItem } from "./MessageItem";

import type { MessageWithUser } from "../types";

import { setRightSidePanelViewAtom } from "@/providers/store/ui";

type ThreadReplyListProps = {
  replies: MessageWithUser[];
  currentUserId: string | null;
  workspaceId: string;
  channelId: string;
};

export const ThreadReplyList = ({
  replies,
  currentUserId,
  workspaceId,
  channelId,
}: ThreadReplyListProps) => {
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);
  const dateTimeFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat("ja-JP", {
        dateStyle: "short",
        timeStyle: "short",
      }),
    []
  );

  const handleCopyLink = useCallback(
    (messageId: string) => {
      const url = `${window.location.origin}/app/${workspaceId}/${channelId}?message=${messageId}`;
      navigator.clipboard.writeText(url);
      notifications.show({
        title: "コピーしました",
        message: "メッセージリンクをクリップボードにコピーしました",
      });
    },
    [workspaceId, channelId]
  );

  const handleCreateThread = useCallback(
    (messageId: string) => {
      setRightSidePanelView({ type: "thread", threadId: messageId });
    },
    [setRightSidePanelView]
  );

  const handleBookmark = useCallback((messageId: string) => {
    if (!messageId) {
      return;
    }
    notifications.show({
      title: "ブックマーク",
      message: "ブックマークの状態を更新しました",
    });
  }, []);

  if (replies.length === 0) {
    return (
      <div className="flex items-center justify-center py-8">
        <Text c="dimmed" size="sm">
          まだ返信がありません
        </Text>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      {replies.map((reply) => (
        <MessageItem
          key={reply.id}
          message={reply}
          currentUserId={currentUserId}
          dateTimeFormatter={dateTimeFormatter}
          onCopyLink={handleCopyLink}
          onCreateThread={handleCreateThread}
          onBookmark={handleBookmark}
        />
      ))}
    </div>
  );
};
