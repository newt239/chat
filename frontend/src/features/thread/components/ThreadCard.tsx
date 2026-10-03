import { useMutation } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { Badge } from "#/components/ui/Badge/Badge";
import { Link } from "#/components/ui/Link/Link";
import { BaseMessageInput } from "#/features/message/components/BaseMessageInput";
import { MessageItem } from "#/features/message/components/MessageItem";
import { MessageListCard } from "#/features/message/components/MessageListCard";
import { useUpdateListedThread } from "#/features/thread/hooks/useParticipatingThreads";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

import { ThreadFollowButton } from "./ThreadFollowButton";

import type { Message } from "#/gen/chat/v1/message_pb";
import type { ParticipatingThread } from "#/gen/chat/v1/thread_service_pb";

type ThreadCardProps = {
  workspaceId: string;
  thread: ParticipatingThread & { firstMessage: Message };
};

// 親の投稿・最新の返信・返信の入力欄を並べる。返信は一覧のキャッシュに足してすぐ表示する
export const ThreadCard = ({ workspaceId, thread }: ThreadCardProps) => {
  const { t } = useTranslation();
  const markThreadRead = useMutation(ThreadService.method.markThreadRead);
  const updateThread = useUpdateListedThread();
  const { firstMessage, threadId } = thread;
  const { channelId } = firstMessage;
  const hiddenCount = thread.replyCount - thread.latestReplies.length;

  return (
    <MessageListCard workspaceId={workspaceId} message={firstMessage}>
      <MessageItem message={firstMessage} isHighlighted={false} channelChip={null} />
      <div className="flex items-center gap-2 pr-3 pl-15 max-md:pl-3">
        <Link
          to="/app/$workspaceId/$channelId/thread/$messageId"
          params={{ channelId, messageId: threadId, workspaceId }}
          onPress={() => {
            markThreadRead.mutate(
              { threadId },
              {
                onSuccess: () => {
                  updateThread(threadId, (item) => ({ ...item, unreadCount: 0 }));
                },
              },
            );
          }}
          className="text-xs font-semibold text-accent-text no-underline data-hovered:underline"
        >
          {hiddenCount > 0
            ? t("thread.card.showMore", { count: hiddenCount })
            : t("thread.card.open")}
        </Link>
        {thread.unreadCount > 0 && (
          <Badge tone="accent">{t("thread.card.unread", { count: thread.unreadCount })}</Badge>
        )}
        {/* 解除しても一覧からはすぐに消さず、押し直せるようにする */}
        <span className="ml-auto">
          <ThreadFollowButton threadId={threadId} isFollowing={thread.isFollowing} />
        </span>
      </div>
      {thread.latestReplies.map((reply) => (
        <MessageItem key={reply.id} message={reply} isHighlighted={false} channelChip={null} />
      ))}
      <div className="pt-1">
        <BaseMessageInput
          channelId={channelId}
          parentId={threadId}
          placeholder={t("thread.card.replyPlaceholder")}
          targetPicker={null}
          onSent={(reply) => {
            updateThread(threadId, (item) => ({
              ...item,
              lastActivityAt: reply.createdAt,
              latestReplies: [...item.latestReplies, reply].slice(-2),
              replyCount: item.replyCount + 1,
            }));
          }}
        />
      </div>
    </MessageListCard>
  );
};
