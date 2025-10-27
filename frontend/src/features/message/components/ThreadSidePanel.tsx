import { useCallback, useEffect, useRef } from "react";

import { ActionIcon, Card, Divider, Drawer, Loader, Stack, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconX } from "@tabler/icons-react";
import { useAtomValue } from "jotai";

import { useThreadReplies, useSendThreadReply } from "../hooks/useThread";

import { MessageItem } from "./MessageItem";
import { ThreadReplyInput } from "./ThreadReplyInput";
import { ThreadReplyList } from "./ThreadReplyList";

import { userAtom } from "@/providers/store/auth";

type ThreadSidePanelProps = {
  opened: boolean;
  onClose: () => void;
  messageId: string | null;
  workspaceId: string;
  channelId: string;
};

export const ThreadSidePanel = ({
  opened,
  onClose,
  messageId,
  workspaceId,
  channelId,
}: ThreadSidePanelProps) => {
  const currentUser = useAtomValue(userAtom);
  const { data: threadData, isLoading, isError, error } = useThreadReplies(messageId);
  const sendReply = useSendThreadReply(messageId, channelId);
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
      const url = `${window.location.origin}/app/${workspaceId}/${channelId}?message=${msgId}`;
      navigator.clipboard.writeText(url);
      notifications.show({
        title: "コピーしました",
        message: "メッセージリンクをクリップボードにコピーしました",
      });
    },
    [workspaceId, channelId]
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

  return (
    <Drawer
      opened={opened}
      onClose={onClose}
      position="right"
      size="xl"
      title={
        <div className="flex items-center justify-between w-full">
          <Text fw={600} size="lg">
            スレッド
          </Text>
          <ActionIcon variant="subtle" onClick={onClose} aria-label="閉じる">
            <IconX size={20} />
          </ActionIcon>
        </div>
      }
      withCloseButton={false}
    >
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
                  workspaceId={workspaceId}
                  channelId={channelId}
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
    </Drawer>
  );
};
