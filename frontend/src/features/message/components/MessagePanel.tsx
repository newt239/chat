import { useMemo, useEffect, useRef, useCallback } from "react";

import { Card, Loader, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useAtom, useSetAtom, useAtomValue } from "jotai";

import { useMessages, useUpdateMessage, useDeleteMessage } from "../hooks/useMessage";

import { MessageItem } from "./MessageItem";

import { userAtom } from "@/providers/store/auth";
import { setRightSidePanelViewAtom } from "@/providers/store/ui";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "@/providers/store/workspace";
import { useReadStateEvents } from "@/providers/websocket/useReadStateEvents";
import { useWebSocketEvents } from "@/providers/websocket/useWebSocketEvents";

export const MessagePanel = () => {
  const [currentWorkspaceId] = useAtom(currentWorkspaceIdAtom);
  const [currentChannelId] = useAtom(currentChannelIdAtom);
  const currentUser = useAtomValue(userAtom);
  const { data: messageResponse, isLoading, isError, error } = useMessages(currentChannelId);
  const updateMessage = useUpdateMessage(currentChannelId);
  const deleteMessage = useDeleteMessage(currentChannelId);

  // WebSocketイベントを処理
  useWebSocketEvents();
  const { updateReadState } = useReadStateEvents();
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const setRightSidebarView = useSetAtom(setRightSidePanelViewAtom);

  const dateTimeFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat("ja-JP", {
        dateStyle: "short",
        timeStyle: "short",
      }),
    []
  );
  const orderedMessages = useMemo(() => {
    if (!messageResponse || !currentChannelId) {
      return [];
    }
    return [...messageResponse.messages].sort((first, second) => {
      const firstTime = new Date(first.createdAt).getTime();
      const secondTime = new Date(second.createdAt).getTime();
      return firstTime - secondTime;
    });
  }, [messageResponse, currentChannelId]);

  // メッセージが読み込まれた時と新しいメッセージが送信された時に最新メッセージにスクロール
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (messageResponse && !isLoading) {
      scrollToBottom();
    }
  }, [messageResponse, isLoading]);

  // チャンネルが変更された時にスクロールをリセット
  useEffect(() => {
    if (currentChannelId) {
      scrollToBottom();
    }
  }, [currentChannelId]);

  // チャンネルが変更された時にスレッドパネルを閉じる
  useEffect(() => {
    setRightSidebarView({ type: "hidden" });
  }, [currentChannelId, setRightSidebarView]);

  // メッセージが表示されたときに既読状態を更新
  useEffect(() => {
    if (currentChannelId && messageResponse && !isLoading) {
      const lastMessage = orderedMessages[orderedMessages.length - 1];
      if (lastMessage) {
        updateReadState(currentChannelId, lastMessage.id);
      }
    }
  }, [currentChannelId, messageResponse, isLoading, orderedMessages, updateReadState]);

  // アクションハンドラー
  const handleCopyLink = useCallback(
    (messageId: string) => {
      const url = `${window.location.origin}/app/${currentWorkspaceId}/${currentChannelId}?message=${messageId}`;
      navigator.clipboard.writeText(url);
      notifications.show({
        title: "コピーしました",
        message: "メッセージリンクをクリップボードにコピーしました",
      });
    },
    [currentWorkspaceId, currentChannelId]
  );

  const handleCreateThread = useCallback(
    (messageId: string) => {
      setRightSidebarView({ type: "thread", threadId: messageId });
    },
    [setRightSidebarView]
  );

  const handleOpenThread = useCallback(
    (messageId: string) => {
      setRightSidebarView({ type: "thread", threadId: messageId });
    },
    [setRightSidebarView]
  );

  const handleBookmark = useCallback((messageId: string) => {
    console.log("Bookmark message:", messageId);
    // TODO: ブックマーク機能を実装
  }, []);

  const handleEdit = useCallback(
    (messageId: string, currentBody: string) => {
      updateMessage.mutate(
        { messageId, body: currentBody },
        {
          onSuccess: () => {
            notifications.show({
              title: "更新しました",
              message: "メッセージを更新しました",
            });
          },
          onError: (err) => {
            notifications.show({
              title: "エラー",
              message: err.message ?? "メッセージの更新に失敗しました",
              color: "red",
            });
          },
        }
      );
    },
    [updateMessage]
  );

  const handleDelete = useCallback(
    (messageId: string) => {
      deleteMessage.mutate(
        { messageId },
        {
          onSuccess: () => {
            notifications.show({
              title: "削除しました",
              message: "メッセージを削除しました",
            });
          },
          onError: (err) => {
            notifications.show({
              title: "エラー",
              message: err.message ?? "メッセージの削除に失敗しました",
              color: "red",
            });
          },
        }
      );
    },
    [deleteMessage]
  );

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
            {error?.message ?? "メッセージの取得に失敗しました"}
          </Text>
        ) : messageResponse && messageResponse.messages.length > 0 && currentChannelId ? (
          <div className="flex h-full flex-col">
            {messageResponse.hasMore && (
              <Text size="xs" c="dimmed" className="px-4 py-2 text-center">
                さらに過去のメッセージがあります
              </Text>
            )}
            <div className="flex flex-1 flex-col justify-end">
              <div className="space-y-1 px-4 pb-6">
                {orderedMessages.map((message) => (
                  <MessageItem
                    key={message.id}
                    message={message}
                    currentUserId={currentUser?.id ?? null}
                    dateTimeFormatter={dateTimeFormatter}
                    onCopyLink={handleCopyLink}
                    onCreateThread={handleCreateThread}
                    onBookmark={handleBookmark}
                    onOpenThread={handleOpenThread}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                  />
                ))}
                <div ref={messagesEndRef} />
              </div>
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
