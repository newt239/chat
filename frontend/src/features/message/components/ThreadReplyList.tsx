import { useCallback } from "react";

import { Text } from "@mantine/core";
import { useSetAtom } from "jotai";

import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

import { MessageItem } from "./MessageItem";

import type { Message } from "#/gen/chat/v1/message_pb";

type ThreadReplyListProps = {
  replies: Message[];
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

  const handleCopyLink = useCopyMessageLink(workspaceId, channelId);

  const handleCreateThread = useCallback(
    (messageId: string) => {
      setRightSidePanelView({ threadId: messageId, type: "thread" });
    },
    [setRightSidePanelView],
  );

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
          onCopyLink={handleCopyLink}
          onCreateThread={handleCreateThread}
        />
      ))}
    </div>
  );
};
