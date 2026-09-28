import { useCallback, useEffect, useRef } from "react";

import { useNavigate, useSearch } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton";
import { MessageItem } from "#/features/message/components/MessageItem";
import { ThreadReplyInput } from "#/features/message/components/ThreadReplyInput";
import { ThreadReplyList } from "#/features/message/components/ThreadReplyList";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { ThreadPanelContext } from "#/features/message/hooks/useOwnsMessageOverlay";
import { useSendThreadReply, useThreadReplies } from "#/features/message/hooks/useThread";
import { userAtom } from "#/providers/store/auth";

type ThreadPanelProps = {
  workspaceId: string;
  channelId: string;
  threadId: string;
};

export const ThreadPanel = ({ workspaceId, channelId, threadId }: ThreadPanelProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const currentUserId = useAtomValue(userAtom)?.id ?? null;
  // ?message= で返信を指しているときはその返信までスクロールする
  const targetReplyId = useSearch({ select: (search) => search.message, strict: false });
  const { data, isLoading, isError, error } = useThreadReplies(threadId);
  const sendReply = useSendThreadReply();
  const bodyRef = useRef<HTMLDivElement>(null);
  // 親チャンネルの集約表示から開いたスレッドは子孫チャンネルのものなので、返信先は親メッセージのチャンネルにする
  const threadChannelId = data?.parentMessage?.channelId ?? channelId;
  const handleCopyLink = useCopyMessageLink(workspaceId, threadChannelId);

  const replyCount = data?.replies.length;
  useEffect(() => {
    if (replyCount === undefined) {
      return;
    }
    const target =
      targetReplyId === undefined
        ? null
        : bodyRef.current?.querySelector(`[data-message-id="${targetReplyId}"]`);
    if (target) {
      target.scrollIntoView({ block: "center" });
    } else {
      bodyRef.current?.scrollTo({ top: bodyRef.current.scrollHeight });
    }
  }, [replyCount, targetReplyId]);

  const handleCreateThread = useCallback(
    (messageId: string) => {
      void navigate({
        params: { channelId, messageId, workspaceId },
        to: "/app/$workspaceId/$channelId/thread/$messageId",
      });
    },
    [navigate, channelId, workspaceId],
  );

  return (
    <ThreadPanelContext value>
      <div className="flex min-h-0 flex-1 flex-col">
        <div ref={bodyRef} className="min-h-0 flex-1 overflow-y-auto">
          {isLoading ? (
            <div className="flex flex-col gap-2 p-4">
              <Skeleton className="h-4 w-40" />
              <Skeleton className="h-4 w-60" />
            </div>
          ) : isError ? (
            <p className="m-0 p-4 text-caption text-danger">{error.message}</p>
          ) : data?.parentMessage ? (
            <>
              <MessageItem
                message={data.parentMessage}
                currentUserId={currentUserId}
                onCopyLink={handleCopyLink}
                onCreateThread={handleCreateThread}
              />
              <div className="mx-4 my-2 flex items-center gap-2 text-caption text-muted">
                {t("shell.thread.replyCount", { count: data.replies.length })}
                <span className="h-px flex-1 bg-border" />
              </div>
              <ThreadReplyList
                replies={data.replies}
                currentUserId={currentUserId}
                workspaceId={workspaceId}
                channelId={threadChannelId}
              />
            </>
          ) : (
            <p className="m-0 p-4 text-caption text-muted">{t("shell.thread.notFound")}</p>
          )}
        </div>
        {data?.parentMessage && (
          <ThreadReplyInput
            channelId={threadChannelId}
            onSubmit={(body, attachmentIds) => {
              sendReply.mutate({
                attachmentIds,
                body,
                channelId: threadChannelId,
                parentId: threadId,
              });
            }}
            isPending={sendReply.isPending}
            isError={sendReply.isError}
            errorMessage={sendReply.error?.message}
          />
        )}
      </div>
    </ThreadPanelContext>
  );
};
