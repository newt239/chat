import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { BaseMessageInput } from "#/features/message/components/BaseMessageInput";
import { MessageItem } from "#/features/message/components/MessageItem";
import { MessageListCard } from "#/features/message/components/MessageListCard";
import { messageLocation } from "#/lib/messageLocation";

import type { Message } from "#/gen/chat/v1/message_pb";

type MentionCardProps = {
  workspaceId: string;
  message: Message;
};

// メンションされたメッセージと、そのスレッドへ返信する入力欄。送った返信はスレッドで見てもらう
export const MentionCard = ({ workspaceId, message }: MentionCardProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const displayName = useDisplayName();
  const { channelId } = message;
  // 返信へのメンションには同じスレッドで返す
  const threadId = message.parentId ?? message.id;

  return (
    <MessageListCard workspaceId={workspaceId} message={message}>
      <MessageItem message={message} isHighlighted={false} channelChip={null} />
      <div className="pt-1">
        <BaseMessageInput
          channelId={channelId}
          parentId={threadId}
          targetPicker={null}
          placeholder={t("inbox.mention.replyPlaceholder", {
            name: displayName(message.userId, message.user?.displayName ?? ""),
          })}
          onSent={(reply) => {
            toast(t("inbox.mention.replied"), {
              action: {
                label: t("message.card.showInThread"),
                onAction: () => {
                  void navigate(
                    messageLocation({
                      channelId,
                      messageId: reply.id,
                      parentId: threadId,
                      workspaceId,
                    }),
                  );
                },
              },
              tone: "success",
            });
          }}
        />
      </div>
    </MessageListCard>
  );
};
