import { useMutation } from "@connectrpc/connect-query";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Badge } from "#/components/ui/Badge/Badge";
import { Link } from "#/components/ui/Link/Link";
import { MessageItem } from "#/features/message/components/MessageItem";
import { MessageListCard } from "#/features/message/components/MessageListCard";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { useUpdateListedThread } from "#/features/thread/hooks/useParticipatingThreads";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";
import { userAtom } from "#/providers/store/auth";

import { InlineReplyComposer } from "./InlineReplyComposer";

import type { Message } from "#/gen/chat/v1/message_pb";
import type { ParticipatingThread } from "#/gen/chat/v1/thread_service_pb";

type ThreadCardProps = {
  workspaceId: string;
  thread: ParticipatingThread & { firstMessage: Message };
};

// 親の投稿・最新の返信・返信の入力欄を並べる。返信は一覧のキャッシュに足してすぐ表示する
export const ThreadCard = ({ workspaceId, thread }: ThreadCardProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const currentUserId = useAtomValue(userAtom)?.id ?? null;
  const markThreadRead = useMutation(ThreadService.method.markThreadRead);
  const updateThread = useUpdateListedThread();
  const { firstMessage, threadId } = thread;
  const { channelId } = firstMessage;
  const handleCopyLink = useCopyMessageLink(workspaceId, channelId);
  const openThread = () => {
    void navigate({
      params: { channelId, messageId: threadId, workspaceId },
      to: "/app/$workspaceId/$channelId/thread/$messageId",
    });
  };
  const hiddenCount = thread.replyCount - thread.latestReplies.length;

  return (
    <MessageListCard workspaceId={workspaceId} message={firstMessage}>
      <MessageItem
        message={firstMessage}
        currentUserId={currentUserId}
        onCopyLink={handleCopyLink}
        onCreateThread={openThread}
      />
      <div className="flex items-center gap-2 pr-3 pl-[60px] max-md:pl-3">
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
            ? t("inbox.thread.showMore", { count: hiddenCount })
            : t("inbox.thread.open")}
        </Link>
        {thread.unreadCount > 0 && (
          <Badge tone="accent">{t("inbox.thread.unread", { count: thread.unreadCount })}</Badge>
        )}
      </div>
      {thread.latestReplies.map((reply) => (
        <MessageItem
          key={reply.id}
          message={reply}
          currentUserId={currentUserId}
          onCopyLink={handleCopyLink}
          onCreateThread={openThread}
        />
      ))}
      <div className="pt-1">
        <InlineReplyComposer
          channelId={channelId}
          parentId={threadId}
          placeholder={t("inbox.thread.replyPlaceholder")}
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
