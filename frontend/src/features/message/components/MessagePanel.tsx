import { useEffect, useRef, useCallback } from "react";

import { Card, Loader, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useAtom, useSetAtom, useAtomValue } from "jotai";

import { useAutoScrollToBottom } from "#/features/message/hooks/useAutoScrollToBottom";
import { useChannelTimeline } from "#/features/message/hooks/useChannelTimeline";
import { useMessageActions } from "#/features/message/hooks/useMessageActions";
import { useMessageViewportDetection } from "#/features/message/hooks/useMessageViewportDetection";
import { paths } from "#/lib/paths";
import { userAtom } from "#/providers/store/auth";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "#/providers/store/workspace";
import { useWsClient } from "#/providers/ws/useWsClient";

import { useMessages } from "../hooks/useMessage";
import { MessageItem } from "./MessageItem";
import { SystemMessageItem } from "./SystemMessageItem";

export const MessagePanel = () => {
  const [currentWorkspaceId] = useAtom(currentWorkspaceIdAtom);
  const [currentChannelId] = useAtom(currentChannelIdAtom);
  const currentUser = useAtomValue(userAtom);
  const { data: messageResponse, isLoading, isError, error } = useMessages(currentChannelId);
  const { wsClient } = useWsClient();

  const { orderedItems } = useChannelTimeline({
    currentChannelId,
    initialMessages: messageResponse?.messages,
    wsClient: wsClient ?? null,
  });

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const setRightSidebarView = useSetAtom(setRightSidePanelViewAtom);

  // 最新メッセージのIDを取得（ユーザーメッセージのみ）
  const latestUserMessageId =
    orderedItems.length > 0
      ? (() => {
          for (let i = orderedItems.length - 1; i >= 0; i--) {
            const item = orderedItems[i];
            if (item?.type === "user" && item.userMessage) {
              return item.userMessage.id;
            }
          }
          return null;
        })()
      : null;

  const { latestMessageRef } = useMessageViewportDetection({
    channelId: currentChannelId,
    latestMessageId: latestUserMessageId,
    workspaceId: currentWorkspaceId,
  });

  const scrollToBottom = useAutoScrollToBottom(messagesEndRef);

  useEffect(() => {
    scrollToBottom();
  }, [messageResponse, isLoading, currentChannelId, scrollToBottom]);

  useEffect(() => {
    setRightSidebarView({ type: "hidden" });
  }, [currentChannelId, setRightSidebarView]);

  const handleCopyLink = useCallback(
    (messageId: string) => {
      if (!currentWorkspaceId || !currentChannelId) {
        return;
      }
      void navigator.clipboard.writeText(
        paths.channel(currentWorkspaceId, currentChannelId, messageId),
      );
      notifications.show({
        message: "メッセージリンクをクリップボードにコピーしました",
        title: "コピーしました",
      });
    },
    [currentWorkspaceId, currentChannelId],
  );

  const handleCreateThread = useCallback(
    (messageId: string) => {
      setRightSidebarView({ threadId: messageId, type: "thread" });
    },
    [setRightSidebarView],
  );

  const handleOpenThread = useCallback(
    (messageId: string) => {
      setRightSidebarView({ threadId: messageId, type: "thread" });
    },
    [setRightSidebarView],
  );

  const { handleEdit, handleDelete } = useMessageActions(currentChannelId);

  if (currentWorkspaceId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">ワークスペースを選択してください</Text>
      </Card>
    );
  }

  if (currentChannelId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">チャンネルを選択するとメッセージが表示されます</Text>
      </Card>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col w-full">
      <div className="flex-1 overflow-y-auto min-h-0">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader size="sm" />
          </div>
        ) : isError ? (
          <Text c="red" size="sm">
            {error.message}
          </Text>
        ) : messageResponse && messageResponse.messages.length > 0 && currentChannelId ? (
          <div className="flex h-full flex-col">
            {messageResponse.hasMore && (
              <Text size="xs" c="dimmed" className="px-4 py-2 text-center">
                さらに過去のメッセージがあります
              </Text>
            )}
            <div className="flex flex-1 flex-col justify-end">
              {orderedItems.map((item) => {
                if (item.type === "user" && item.userMessage) {
                  const msg = item.userMessage;
                  const isLatestMessage = msg.id === latestUserMessageId;
                  return (
                    <div key={`u-${msg.id}`} ref={isLatestMessage ? latestMessageRef : undefined}>
                      <MessageItem
                        message={msg}
                        currentUserId={currentUser?.id ?? null}
                        onCopyLink={handleCopyLink}
                        onCreateThread={handleCreateThread}
                        onOpenThread={handleOpenThread}
                        onEdit={handleEdit}
                        onDelete={handleDelete}
                      />
                    </div>
                  );
                }
                if (item.type === "system" && item.systemMessage) {
                  return (
                    <SystemMessageItem
                      key={`s-${item.systemMessage.id}`}
                      message={item.systemMessage}
                    />
                  );
                }
                return null;
              })}
              <div ref={messagesEndRef} />
            </div>
          </div>
        ) : (
          <div className="flex h-full items-center justify-center">
            <Text c="dimmed" size="sm">
              メッセージはまだありません
            </Text>
          </div>
        )}
      </div>
    </div>
  );
};
