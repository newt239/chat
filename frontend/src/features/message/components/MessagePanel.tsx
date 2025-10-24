import { useMemo, useState, useEffect, useRef, useCallback } from "react";
import type { FormEvent } from "react";

import { Card, Loader, Stack, Text, Textarea, ActionIcon } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconInfoCircle } from "@tabler/icons-react";
import { useSetAtom } from "jotai";

import { useMessages, useSendMessage } from "../hooks/useMessage";
import { useMessageInputMode } from "../hooks/useMessageInputMode";

import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessageItem } from "./MessageItem";
import { MessagePreview } from "./MessagePreview";

import { useChannels } from "@/features/channel/hooks/useChannel";
import { toggleMemberPanelAtom } from "@/lib/store/ui";

interface MessagePanelProps {
  workspaceId: string | null;
  channelId: string | null;
}

export const MessagePanel = ({ workspaceId, channelId }: MessagePanelProps) => {
  const { data: channels } = useChannels(workspaceId);
  const { data: messageResponse, isLoading, isError, error } = useMessages(channelId);
  const sendMessage = useSendMessage(channelId);
  const [body, setBody] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const toggleMemberPanel = useSetAtom(toggleMemberPanelAtom);
  const { mode, toggleMode, isEditMode } = useMessageInputMode();
  const dateTimeFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat("ja-JP", {
        dateStyle: "short",
        timeStyle: "short",
      }),
    []
  );
  const orderedMessages = useMemo(() => {
    if (!messageResponse) {
      return [];
    }
    return [...messageResponse.messages].sort((first, second) => {
      const firstTime = new Date(first.createdAt).getTime();
      const secondTime = new Date(second.createdAt).getTime();
      return firstTime - secondTime;
    });
  }, [messageResponse]);

  const activeChannel = useMemo(() => {
    if (!channels || channelId === null) {
      return null;
    }
    return channels.find((channel) => channel.id === channelId) ?? null;
  }, [channels, channelId]);

  // メッセージが読み込まれた時と新しいメッセージが送信された時に最新メッセージにスクロール
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (messageResponse && !isLoading) {
      scrollToBottom();
    }
  }, [messageResponse, isLoading]);

  useEffect(() => {
    if (sendMessage.isSuccess) {
      scrollToBottom();
    }
  }, [sendMessage.isSuccess]);

  // アクションハンドラー
  const handleCopyLink = useCallback(
    (messageId: string) => {
      const url = `${window.location.origin}/app/${workspaceId}/${channelId}?message=${messageId}`;
      navigator.clipboard.writeText(url);
      notifications.show({
        title: "コピーしました",
        message: "メッセージリンクをクリップボードにコピーしました",
      });
    },
    [workspaceId, channelId]
  );

  const handleCreateThread = useCallback((messageId: string) => {
    console.log("Create thread for message:", messageId);
    // TODO: スレッド作成機能を実装
  }, []);

  const handleBookmark = useCallback((messageId: string) => {
    console.log("Bookmark message:", messageId);
    // TODO: ブックマーク機能を実装
  }, []);

  if (workspaceId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">ワークスペースを選択してください</Text>
      </Card>
    );
  }

  if (channelId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">チャンネルを選択するとメッセージが表示されます</Text>
      </Card>
    );
  }

  const handleSubmit = (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault();
    if (body.trim().length === 0) {
      return;
    }
    sendMessage.mutate(
      { body: body.trim() },
      {
        onSuccess: () => {
          setBody("");
        },
      }
    );
  };

  return (
    <div className="flex h-full min-h-0 flex-col w-full">
      <Card withBorder padding="lg" radius="md" className="shrink-0">
        <div className="flex items-start justify-between">
          <Stack gap="xs" className="flex-1">
            <Text fw={600} size="lg">
              {activeChannel ? `#${activeChannel.name}` : "チャンネル"}
            </Text>
            {activeChannel?.description && (
              <Text size="sm" c="dimmed">
                {activeChannel.description}
              </Text>
            )}
          </Stack>
          <ActionIcon
            variant="subtle"
            color="gray"
            size="lg"
            onClick={toggleMemberPanel}
            aria-label="メンバーパネルの表示切り替え"
          >
            <IconInfoCircle size={20} />
          </ActionIcon>
        </div>
      </Card>

      <div className="flex-1 overflow-y-auto min-h-0">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader size="sm" />
          </div>
        ) : isError ? (
          <Text c="red" size="sm">
            {error?.message ?? "メッセージの取得に失敗しました"}
          </Text>
        ) : messageResponse && messageResponse.messages.length > 0 ? (
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
                    dateTimeFormatter={dateTimeFormatter}
                    onCopyLink={handleCopyLink}
                    onCreateThread={handleCreateThread}
                    onBookmark={handleBookmark}
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

      <Card withBorder padding="lg" radius="md" className="shrink-0">
        <form onSubmit={handleSubmit}>
          <MessageInputToolbar
            mode={mode}
            onToggleMode={toggleMode}
            onSubmit={() => handleSubmit()}
            disabled={body.trim().length === 0}
            loading={sendMessage.isPending}
            textareaRef={textareaRef}
          />
          {isEditMode ? (
            <Textarea
              ref={textareaRef}
              placeholder="メッセージを入力..."
              minRows={3}
              autosize
              value={body}
              onChange={(event) => setBody(event.currentTarget.value)}
              disabled={sendMessage.isPending}
            />
          ) : (
            <MessagePreview content={body} />
          )}
          {sendMessage.isError && (
            <Text c="red" size="sm" className="mt-2">
              {sendMessage.error?.message ?? "メッセージの送信に失敗しました"}
            </Text>
          )}
        </form>
      </Card>
    </div>
  );
};
