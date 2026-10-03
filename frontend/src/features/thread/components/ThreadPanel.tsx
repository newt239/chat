import { create } from "@bufbuild/protobuf";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { BaseMessageInput } from "#/features/message/components/BaseMessageInput";
import { MessageItem } from "#/features/message/components/MessageItem";
import { MessageList } from "#/features/message/components/MessageList";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { useSendMessage } from "#/features/message/hooks/useMessage";
import { ThreadPanelContext } from "#/features/message/hooks/useOwnsMessageOverlay";
import { useThreadReplies } from "#/features/message/hooks/useThread";
import { toDateKey } from "#/features/message/utils/dateJump";
import { buildTimelineRows } from "#/features/message/utils/timelineRows";
import { TimelineItemSchema } from "#/gen/chat/v1/message_pb";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";

import type { TimelineRow } from "#/features/message/utils/timelineRows";
import type { Message } from "#/gen/chat/v1/message_pb";

type ThreadPanelProps = {
  workspaceId: string;
  channelId: string;
  threadId: string;
};

const noopRef = () => undefined;

export const ThreadPanel = ({ workspaceId, channelId, threadId }: ThreadPanelProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  // ?message= で返信を指しているときはその返信の前後を読み、そこまでスクロールする
  const targetReplyId = useSearch({ select: (search) => search.message, strict: false }) ?? null;
  const {
    parentMessage,
    replies,
    replyCount,
    hasOlder,
    hasNewer,
    isLoading,
    isError,
    error,
    load,
    loading,
  } = useThreadReplies(threadId, targetReplyId);
  const sendReply = useSendMessage();
  const { timeZone } = useDateFormat();
  // 親チャンネルの集約表示から開いたスレッドは子孫チャンネルのものなので、返信先は親メッセージのチャンネルにする
  const threadChannelId = parentMessage?.channelId ?? channelId;
  const handleCopyLink = useCopyMessageLink(workspaceId, threadChannelId);

  const handleCreateThread = (messageId: string) => {
    void navigate({
      params: { channelId, messageId, workspaceId },
      to: "/app/$workspaceId/$channelId/thread/$messageId",
    });
  };

  // 親メッセージは最初の返信まで読み込んだときだけ先頭に置く
  const replyRows = buildTimelineRows(
    (replies ?? []).map((reply) =>
      create(TimelineItemSchema, {
        content: { case: "userMessage", value: reply },
        createdAt: reply.createdAt,
      }),
    ),
    false,
    timeZone,
  );
  const rows: TimelineRow[] =
    replies === undefined || parentMessage === undefined
      ? []
      : hasOlder
        ? replyRows
        : [
            {
              dateKey: toDateKey(toDate(parentMessage.createdAt), timeZone),
              key: "header",
              kind: "header",
            },
            ...replyRows,
          ];

  const renderMessage = (message: Message, isHighlighted: boolean) => (
    <MessageItem
      message={message}
      onCopyLink={handleCopyLink}
      onCreateThread={handleCreateThread}
      isHighlighted={isHighlighted}
    />
  );

  const renderBody = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col gap-2 p-4">
          <Skeleton className="h-4 w-40" />
          <Skeleton className="h-4 w-60" />
        </div>
      );
    }
    if (isError) {
      return <p className="m-0 p-4 text-caption text-danger">{error?.message}</p>;
    }
    if (parentMessage === undefined || replies === undefined) {
      return <p className="m-0 p-4 text-caption text-muted">{t("shell.thread.notFound")}</p>;
    }
    return (
      <MessageList
        key={`${threadId}:${targetReplyId ?? ""}`}
        rows={rows}
        targetMessageId={targetReplyId}
        hasOlder={hasOlder}
        hasNewer={hasNewer}
        loading={loading}
        onLoad={load}
        onJumpToLatest={() => {
          void navigate({ search: (prev) => ({ ...prev, message: undefined }), to: "." });
        }}
        latestMessageRef={noopRef}
        latestUserMessageId={null}
        renderMessage={renderMessage}
        header={
          <>
            <MessageItem
              message={parentMessage}
              onCopyLink={handleCopyLink}
              onCreateThread={handleCreateThread}
            />
            <div className="mx-4 my-2 flex items-center gap-2 text-caption text-muted">
              {replyCount === 0
                ? t("message.thread.noReplies")
                : t("shell.thread.replyCount", { count: replyCount })}
              <span className="h-px flex-1 bg-border" />
            </div>
          </>
        }
      />
    );
  };

  return (
    <ThreadPanelContext value>
      <div className="flex min-h-0 flex-1 flex-col">
        {renderBody()}
        {parentMessage && (
          <BaseMessageInput
            channelId={threadChannelId}
            parentId={threadId}
            onSubmit={(content) => {
              sendReply.mutate({ ...content, channelId: threadChannelId, parentId: threadId });
            }}
            placeholder={t("message.thread.replyPlaceholder")}
            isPending={sendReply.isPending}
            error={
              sendReply.isError
                ? sendReply.error.message || t("message.thread.sendFailed")
                : undefined
            }
          />
        )}
      </div>
    </ThreadPanelContext>
  );
};
