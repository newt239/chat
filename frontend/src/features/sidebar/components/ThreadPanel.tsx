import { useCallback, useEffect, useRef } from "react";

import { Card, Divider, Loader, Stack, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useAtom, useAtomValue } from "jotai";

import { MessageItem } from "@/features/message/components/MessageItem";
import { ThreadReplyInput } from "@/features/message/components/ThreadReplyInput";
import { ThreadReplyList } from "@/features/message/components/ThreadReplyList";
import { useThreadReplies, useSendThreadReply } from "@/features/message/hooks/useThread";
import { userAtom } from "@/providers/store/auth";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "@/providers/store/workspace";

const SIDEBAR_CONTAINER_CLASS = "border-l border-gray-200 bg-gray-50 p-4 h-full overflow-y-auto";

type ThreadPanelProps = {
  threadId: string;
};

export const ThreadPanel = ({ threadId }: ThreadPanelProps) => {
  const currentUser = useAtomValue(userAtom);
  const [currentWorkspaceId] = useAtom(currentWorkspaceIdAtom);
  const [currentChannelId] = useAtom(currentChannelIdAtom);
  const { data: threadData, isLoading, isError, error } = useThreadReplies(threadId);
  const sendReply = useSendThreadReply(threadId, currentChannelId);
  const repliesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    repliesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (threadData && !isLoading) {
      scrollToBottom();
    }
  }, [threadData, isLoading]);

  useEffect(() => {
    if (sendReply.isSuccess) {
      scrollToBottom();
    }
  }, [sendReply.isSuccess]);

  const dateTimeFormatter = new Intl.DateTimeFormat("ja-JP", {
    dateStyle: "short",
    timeStyle: "short",
  });

  const handleCopyLink = useCallback(
    (msgId: string) => {
      const url = `${window.location.origin}/app/${currentWorkspaceId}/${currentChannelId}?message=${msgId}`;
      navigator.clipboard.writeText(url);
      notifications.show({
        title: "コピーしました",
        message: "メッセージリンクをクリップボードにコピーしました",
      });
    },
    [currentWorkspaceId, currentChannelId]
  );

  const handleCreateThread = useCallback((msgId: string) => {
    console.log("Create thread for message:", msgId);
  }, []);

  const handleBookmark = useCallback((msgId: string) => {
    console.log("Bookmark message:", msgId);
  }, []);

  const handleSendReply = useCallback(
    (body: string) => {
      sendReply.mutate({ body });
    },
    [sendReply]
  );

  if (!currentWorkspaceId || !currentChannelId) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text c="dimmed" size="sm">
          ワークスペースまたはチャンネルが選択されていません
        </Text>
      </div>
    );
  }

  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <div className="flex h-full flex-col">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader size="sm" />
          </div>
        ) : isError ? (
          <Text c="red" size="sm">
            {error?.message ?? "スレッドの取得に失敗しました"}
          </Text>
        ) : threadData ? (
          <>
            <div className="flex-1 overflow-y-auto">
              <Stack gap="md">
                {/* 親メッセージ */}
                <Card withBorder padding="md" radius="md">
                  <MessageItem
                    message={threadData.parentMessage}
                    currentUserId={currentUser?.id ?? null}
                    dateTimeFormatter={dateTimeFormatter}
                    onCopyLink={handleCopyLink}
                    onCreateThread={handleCreateThread}
                    onBookmark={handleBookmark}
                  />
                </Card>

                <Divider label={`${threadData.replies.length}件の返信`} labelPosition="center" />

                {/* 返信一覧 */}
                <ThreadReplyList
                  replies={threadData.replies}
                  currentUserId={currentUser?.id ?? null}
                  workspaceId={currentWorkspaceId}
                  channelId={currentChannelId}
                />

                <div ref={repliesEndRef} />
              </Stack>
            </div>

            {/* 返信入力 */}
            <div className="shrink-0">
              <ThreadReplyInput
                onSubmit={handleSendReply}
                isPending={sendReply.isPending}
                isError={sendReply.isError}
                errorMessage={sendReply.error?.message}
              />
            </div>
          </>
        ) : (
          <Text c="dimmed" size="sm">
            スレッドが見つかりません
          </Text>
        )}
      </div>
    </div>
  );
};
