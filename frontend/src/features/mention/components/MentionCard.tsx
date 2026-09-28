import { useState } from "react";

import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { MessageItem } from "#/features/message/components/MessageItem";
import { MessageListCard } from "#/features/message/components/MessageListCard";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { InlineReplyComposer } from "#/features/thread/components/InlineReplyComposer";
import { userAtom } from "#/providers/store/auth";

import type { Message } from "#/gen/chat/v1/message_pb";

type MentionCardProps = {
  workspaceId: string;
  message: Message;
};

// メンションされたメッセージと、そのスレッドへ返信する入力欄。送った返信はカードの中に足していく
export const MentionCard = ({ workspaceId, message }: MentionCardProps) => {
  const { t } = useTranslation();
  const displayName = useDisplayName();
  const navigate = useNavigate();
  const currentUserId = useAtomValue(userAtom)?.id ?? null;
  const [replies, setReplies] = useState<Message[]>([]);
  const { channelId } = message;
  // 返信へのメンションには同じスレッドで返す
  const threadId = message.parentId ?? message.id;
  const handleCopyLink = useCopyMessageLink(workspaceId, channelId);
  const openThread = () => {
    void navigate({
      params: { channelId, messageId: threadId, workspaceId },
      to: "/app/$workspaceId/$channelId/thread/$messageId",
    });
  };

  return (
    <MessageListCard workspaceId={workspaceId} message={message}>
      {[message, ...replies].map((item) => (
        <MessageItem
          key={item.id}
          message={item}
          currentUserId={currentUserId}
          onCopyLink={handleCopyLink}
          onCreateThread={openThread}
        />
      ))}
      <div className="pt-1">
        <InlineReplyComposer
          channelId={channelId}
          parentId={threadId}
          placeholder={t("inbox.mention.replyPlaceholder", {
            name: displayName(message.userId, message.user?.displayName ?? ""),
          })}
          onSent={(reply) => {
            setReplies((prev) => [...prev, reply]);
          }}
        />
      </div>
    </MessageListCard>
  );
};
