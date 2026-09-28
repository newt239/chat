import { useCallback } from "react";

import { useSetAtom } from "jotai";
import { useTranslation } from "react-i18next";

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
  const { t } = useTranslation();
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
      <p className="m-0 py-8 text-center font-sans text-body text-muted">
        {t("message.thread.noReplies")}
      </p>
    );
  }

  return (
    <div className="flex flex-col pt-7">
      {replies.map((reply) => (
        <div key={reply.id} data-message-id={reply.id}>
          <MessageItem
            message={reply}
            currentUserId={currentUserId}
            onCopyLink={handleCopyLink}
            onCreateThread={handleCreateThread}
          />
        </div>
      ))}
    </div>
  );
};
