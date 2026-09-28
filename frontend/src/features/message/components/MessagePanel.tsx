import { useEffect, useMemo, useRef, useCallback } from "react";

import { Button, Card, Loader, Text } from "@mantine/core";
import { useAtom, useSetAtom, useAtomValue } from "jotai";

import { useAutoScrollToBottom } from "#/features/message/hooks/useAutoScrollToBottom";
import { useChannelThreadMetadata } from "#/features/message/hooks/useChannelThreadMetadata";
import { useChannelTimeline } from "#/features/message/hooks/useChannelTimeline";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { useHighlightedMessage } from "#/features/message/hooks/useHighlightedMessage";
import { useMessageActions } from "#/features/message/hooks/useMessageActions";
import { useMessageViewportDetection } from "#/features/message/hooks/useMessageViewportDetection";
import { useOlderMessages } from "#/features/message/hooks/useOlderMessages";
import { userAtom } from "#/providers/store/auth";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "#/providers/store/workspace";
import { useWsClient } from "#/providers/ws/useWsClient";

import { useMessages } from "../hooks/useMessage";
import { MessageItem } from "./MessageItem";
import { SystemMessageItem } from "./SystemMessageItem";
import { TypingIndicator } from "./TypingIndicator";

export const MessagePanel = () => {
  const [currentWorkspaceId] = useAtom(currentWorkspaceIdAtom);
  const [currentChannelId] = useAtom(currentChannelIdAtom);
  const currentUser = useAtomValue(userAtom);
  const { data: messageResponse, isLoading, isError, error } = useMessages(currentChannelId);
  const { wsClient } = useWsClient();
  const threadMetadataById = useChannelThreadMetadata(currentChannelId);

  const {
    olderItems,
    hasMore: hasOlderMessages,
    isLoading: isLoadingOlder,
    loadOlder,
  } = useOlderMessages(currentChannelId, messageResponse?.hasMore ?? false);

  const initialMessages = useMemo(
    () => (messageResponse ? [...olderItems, ...messageResponse.messages] : undefined),
    [olderItems, messageResponse],
  );

  const { orderedItems, typingUserIds } = useChannelTimeline({
    currentChannelId,
    initialMessages,
    wsClient: wsClient ?? null,
  });

  const handleLoadOlder = () => {
    const oldest = orderedItems.at(0);
    if (oldest) {
      void loadOlder(oldest.createdAt);
    }
  };

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const setRightSidebarView = useSetAtom(setRightSidePanelViewAtom);

  // 最新メッセージのIDを取得（ユーザーメッセージのみ）
  const latestUserMessageId =
    orderedItems.length > 0
      ? (() => {
          for (let i = orderedItems.length - 1; i >= 0; i--) {
            const item = orderedItems[i];
            if (item?.content.case === "userMessage") {
              return item.content.value.id;
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
  const { highlightedId, highlightedMessageRef, targetMessageId } =
    useHighlightedMessage(!isLoading);

  useEffect(() => {
    // 特定メッセージへのリンクで開いたときは最下部へ飛ばさない
    if (targetMessageId === null) {
      scrollToBottom();
    }
  }, [messageResponse, isLoading, currentChannelId, scrollToBottom, targetMessageId]);

  useEffect(() => {
    setRightSidebarView({ type: "hidden" });
  }, [currentChannelId, setRightSidebarView]);

  const handleCopyLink = useCopyMessageLink(currentWorkspaceId, currentChannelId);

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

  const { handleEdit, handleDelete } = useMessageActions();

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
            {hasOlderMessages && (
              <div className="flex justify-center py-2">
                <Button
                  variant="subtle"
                  size="xs"
                  loading={isLoadingOlder}
                  onClick={handleLoadOlder}
                >
                  さらに過去のメッセージを読み込む
                </Button>
              </div>
            )}
            <div className="flex flex-1 flex-col justify-end">
              {orderedItems.map((item) => {
                if (item.content.case === "userMessage") {
                  const msg = item.content.value;
                  const isLatestMessage = msg.id === latestUserMessageId;
                  const isHighlighted = msg.id === highlightedId;
                  return (
                    <div
                      key={`u-${msg.id}`}
                      ref={
                        msg.id === targetMessageId
                          ? highlightedMessageRef
                          : isLatestMessage
                            ? latestMessageRef
                            : undefined
                      }
                      className={isHighlighted ? "bg-yellow-50 transition-colors" : undefined}
                    >
                      <MessageItem
                        message={msg}
                        currentUserId={currentUser?.id ?? null}
                        onCopyLink={handleCopyLink}
                        onCreateThread={handleCreateThread}
                        onOpenThread={handleOpenThread}
                        onEdit={handleEdit}
                        onDelete={handleDelete}
                        threadMetadata={threadMetadataById.get(msg.id)}
                      />
                    </div>
                  );
                }
                if (item.content.case === "systemMessage") {
                  return (
                    <SystemMessageItem
                      key={`s-${item.content.value.id}`}
                      message={item.content.value}
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
      <TypingIndicator userIds={typingUserIds} />
    </div>
  );
};
