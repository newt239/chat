import { formatDateTime } from "@chat/i18n";
import { useMutation } from "@connectrpc/connect-query";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { Badge } from "#/components/ui/Badge";
import { Link } from "#/components/ui/Link";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";
import { toDate } from "#/lib/timestamp";
import { preferencesAtom } from "#/providers/store/preferences";

import type { Message } from "#/gen/chat/v1/message_pb";
import type { ParticipatingThread } from "#/gen/chat/v1/thread_service_pb";

type ThreadCardProps = {
  workspaceId: string;
  thread: ParticipatingThread & { firstMessage: Message };
  onMarkedRead: (threadId: string) => void;
};

export const ThreadCard = ({ workspaceId, thread, onMarkedRead }: ThreadCardProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const markThreadRead = useMutation(ThreadService.method.markThreadRead);
  const { firstMessage } = thread;
  const author = useDisplayName()(firstMessage.userId, firstMessage.user?.displayName ?? "");

  return (
    <Link
      to="/app/$workspaceId/$channelId/thread/$messageId"
      params={{ channelId: firstMessage.channelId, messageId: thread.threadId, workspaceId }}
      onPress={() => {
        markThreadRead.mutate(
          { threadId: thread.threadId },
          {
            onSuccess: () => {
              onMarkedRead(thread.threadId);
            },
          },
        );
      }}
      className="flex gap-3 rounded-lg border border-border bg-surface p-3 text-text no-underline data-hovered:bg-hover"
    >
      <Avatar name={author} src={firstMessage.user?.avatarUrl} size={32} />
      <span className="flex min-w-0 flex-1 flex-col gap-1">
        <span className="flex items-center gap-2">
          <b className="truncate text-body-strong">{author}</b>
          <span className="flex-1 truncate text-caption text-muted">
            {formatDateTime(toDate(thread.lastActivityAt), locale)}
          </span>
          {thread.unreadCount > 0 && <Badge>{thread.unreadCount}</Badge>}
        </span>
        <span className="line-clamp-2 text-body">{firstMessage.body}</span>
        <span className="text-caption text-accent-text">
          {t("shell.thread.replyCount", { count: thread.replyCount })}
        </span>
      </span>
    </Link>
  );
};
