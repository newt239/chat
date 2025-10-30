import { useMemo, useEffect, useRef, useCallback, useState } from "react";

import { Card, Loader, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useAtom, useSetAtom, useAtomValue } from "jotai";

import { useMessages, useUpdateMessage, useDeleteMessage } from "../hooks/useMessage";
import { messageWithThreadSchema } from "../schemas";

import { MessageItem } from "./MessageItem";

import type { MessageWithThread } from "../schemas";
import type { NewMessagePayload } from "@/types/wsEvents";

import { userAtom } from "@/providers/store/auth";
import { setRightSidePanelViewAtom } from "@/providers/store/ui";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "@/providers/store/workspace";
import { useWsClient } from "@/providers/ws/WsProvider";

const resolveErrorMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return fallback;
};

export const MessagePanel = () => {
  const [currentWorkspaceId] = useAtom(currentWorkspaceIdAtom);
  const [currentChannelId] = useAtom(currentChannelIdAtom);
  const currentUser = useAtomValue(userAtom);
  const { data: messageResponse, isLoading, isError, error } = useMessages(currentChannelId);
  const updateMessage = useUpdateMessage(currentChannelId);
  const deleteMessage = useDeleteMessage(currentChannelId);
  const { wsClient } = useWsClient();

  const [messages, setMessages] = useState<MessageWithThread[]>([]);
  // メッセージロード・チャンネル変更時に初期ロード
  useEffect(() => {
    setMessages((messageResponse?.messages ?? []) as MessageWithThread[]);
  }, [messageResponse, currentChannelId]);

  // new_message(Ws)購読とjoin/leave管理
  useEffect(() => {
    if (!wsClient || !currentChannelId) return;
    wsClient.joinChannel(currentChannelId);
    // new_message購読
    const handleNewMessage = (payload: NewMessagePayload) => {
      const result = messageWithThreadSchema.safeParse(payload.message);
      if (!result.success) return;
      setMessages((prev: MessageWithThread[]): MessageWithThread[] => {
        // 同一ID重複は排除
        if (prev.some((m) => m.id === result.data.id)) return prev;
        return [...prev, result.data];
      });
    };
    wsClient.onNewMessage(handleNewMessage);
    return () => {
      wsClient.leaveChannel(currentChannelId);
    };
  }, [wsClient, currentChannelId]);

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
    if (!messages || !currentChannelId) {
      return [];
    }
    const uniqueMessages = messages.filter(
      (message: MessageWithThread, index: number, self: MessageWithThread[]) =>
        index === self.findIndex((m) => m.id === message.id)
    );
    return uniqueMessages.sort((first: MessageWithThread, second: MessageWithThread) => {
      const firstTime = new Date(first.createdAt).getTime();
      const secondTime = new Date(second.createdAt).getTime();
      return firstTime - secondTime;
    });
  }, [messages, currentChannelId]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (messageResponse && !isLoading) {
      scrollToBottom();
    }
  }, [messageResponse, isLoading]);

  useEffect(() => {
    if (currentChannelId) {
      scrollToBottom();
    }
  }, [currentChannelId]);

  useEffect(() => {
    setRightSidebarView({ type: "hidden" });
  }, [currentChannelId, setRightSidebarView]);

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

  const handleEdit = useCallback(
    async (messageId: string, nextBody: string) => {
      try {
        await updateMessage.mutateAsync({ messageId, body: nextBody });
        notifications.show({
          title: "更新しました",
          message: "メッセージを更新しました",
        });
      } catch (error) {
        notifications.show({
          title: "エラー",
          message: resolveErrorMessage(error, "メッセージの更新に失敗しました"),
          color: "red",
        });
        throw error;
      }
    },
    [updateMessage]
  );

  const handleDelete = useCallback(
    async (messageId: string) => {
      try {
        await deleteMessage.mutateAsync({ messageId });
        notifications.show({
          title: "削除しました",
          message: "メッセージを削除しました",
        });
      } catch (error) {
        notifications.show({
          title: "エラー",
          message: resolveErrorMessage(error, "メッセージの削除に失敗しました"),
          color: "red",
        });
      }
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
              {orderedMessages.map((message) => (
                <MessageItem
                  key={message.id}
                  message={message}
                  currentUserId={currentUser?.id ?? null}
                  dateTimeFormatter={dateTimeFormatter}
                  onCopyLink={handleCopyLink}
                  onCreateThread={handleCreateThread}
                  onOpenThread={handleOpenThread}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                  threadMetadata={message.threadMetadata ?? null}
                />
              ))}
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
